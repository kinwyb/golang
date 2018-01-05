package chinapay

import (
	"fmt"
	"strings"
	"time"

	"net/http"
	"net/url"

	"io/ioutil"

	"encoding/base64"

	"regexp"

	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/utils"
)

type withdraw struct {
	payment.PayInfo
	config           *WithdrawConfig
	privKey          *NetPayClient
	pubKey           *NetPayClient
	withdrawURL      string
	queryWithdrawURL string
}

var timeFormat = "2006-01-02 15:04:05"
var tradeFormatError = &payment.WithdrawResult{
	Status:   payment.FAIL,
	FailCode: "TRADENO_FORMAT_ERROR",
	FailMsg:  "交易单号必须市小于16位的纯数字",
}

var regExpTradeNo *regexp.Regexp

func init() {
	regExpTradeNo, _ = regexp.Compile("\\d{1,16}")
}

func (w *withdraw) Withdraw(info *payment.WithdrawInfo) *payment.WithdrawResult {
	if regExpTradeNo != nil && !regExpTradeNo.MatchString(info.TradeNo) {
		return tradeFormatError
	}
	args := map[string]string{
		"merId":    w.config.MerID,                //商户号
		"merDate":  time.Now().Format("20060102"), //商户日期
		"merSeqId": info.TradeNo,                  //商户流水号
		//"certType", "01", //证件类型 01=身份证
		//"certId":info.CertID //身份证号
		"cardNo":   info.CardNo,   //收款账户
		"openBank": info.OpenBank, //开户银行
		"prov":     info.Prov,     ///开户省份
		"city":     info.City,     //开户地区
		"usrName":  info.UserName, //收款人姓名
		"transAmt": fmt.Sprintf("%.0f", info.Money*100),
		"purpose":  info.Desc,
		"termType": "07",
		"signFlag": "1",
		"version":  "20160530",
		"payMode":  "1",
	}
	if info.People {
		args["flag"] = "00" //对私
	} else {
		args["flag"] = "01" //对公
	}
	w.sign(args)
	params := url.Values{}
	for k, v := range args {
		params.Add(k, v)
	}
	request, err := http.NewRequest("POST", w.withdrawURL, strings.NewReader(params.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log(utils.LogLevelError, "银联提现请求创建失败:%s", err.Error())
		return payment.WithdrawRequestFail
	}
	client := http.Client{
		Timeout: 1 * time.Minute,
	}
	response, err := client.Do(request)
	if err != nil {
		log(utils.LogLevelError, "银联提现请求失败:%s", err.Error())
		return payment.WithdrawRequestFail
	}
	responseData, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log(utils.LogLevelError, "银联提现请求结果读取失败:%s", err.Error())
		return payment.WithdrawResponseReadFail
	}
	log(utils.LogLevelInfo, "银联提现请求结果:%s", responseData)
	responseString := string(responseData)
	idex := strings.LastIndex(responseString, "&")
	res, err := url.ParseQuery(responseString)
	if err != nil { //要检测提现是否完成
		return payment.WithdrawResponseUnserializeFail
	}
	result := map[string]string{}
	for k, v := range res {
		result[k] = v[0]
	}
	if result["responseCode"] == "000" { //表示请求成功 应答失败时候检测会发生异常
		v := w.pubKey.Verify(base64.StdEncoding.EncodeToString([]byte(responseString[:idex])), responseString[idex+10:])
		if !v {
			log(utils.LogLevelError, "银联结果签名异常=>[%s]\n等待签名base64结果:%s\n签名:%s",
				responseString[:idex], base64.StdEncoding.EncodeToString([]byte(responseString[:idex])), responseString[idex+10:])
			return payment.WithdrawResponseVerifyFail
		}
		switch result["stat"] {
		case "s":
			//交易成功
			return &payment.WithdrawResult{
				TradeNo:      info.TradeNo,                  //交易流水号
				ThridFlowNo:  result["cpSeqId"],             //第三方交易流水号
				CardNo:       info.CardNo,                   //收款账户
				CertID:       info.CertID,                   //收款人身份证号
				Money:        info.Money,                    //提现金额
				PayTime:      time.Now().Format(timeFormat), //完成时间
				UserName:     info.UserName,
				WithdrawCode: w.Code(),
				WithdrawName: w.Name(),
				Status:       payment.SUCCESS, //提现状态
			}
		case "2", "3", "4", "5", "7", "8":
			return &payment.WithdrawResult{
				TradeNo:      info.TradeNo,                  //交易流水号
				ThridFlowNo:  result["cpSeqId"],             //第三方交易流水号
				CardNo:       info.CardNo,                   //收款账户
				CertID:       info.CertID,                   //收款人身份证号
				Money:        info.Money,                    //提现金额
				PayTime:      time.Now().Format(timeFormat), //完成时间
				UserName:     info.UserName,
				WithdrawCode: w.Code(),
				WithdrawName: w.Name(),
				Status:       payment.DEALING, //提现状态
			}
			//处理中
		case "6", "9":
			//交易失败
			return &payment.WithdrawResult{
				TradeNo:      info.TradeNo,                  //交易流水号
				ThridFlowNo:  result["cpSeqId"],             //第三方交易流水号
				CardNo:       info.CardNo,                   //收款账户
				CertID:       info.CertID,                   //收款人身份证号
				Money:        info.Money,                    //提现金额
				PayTime:      time.Now().Format(timeFormat), //完成时间
				UserName:     info.UserName,
				WithdrawCode: w.Code(),
				WithdrawName: w.Name(),
				Status:       payment.FAIL, //提现状态
				FailCode:     result["stat"],
			}
		}
	}
	//否则查询下交易状态返回查询的状态结果
	qresult := w.QueryWithdraw(info.TradeNo)
	return &payment.WithdrawResult{
		TradeNo:      info.TradeNo,                  //交易流水号
		ThridFlowNo:  result["cpSeqId"],             //第三方交易流水号
		CardNo:       info.CardNo,                   //收款账户
		CertID:       info.CertID,                   //收款人身份证号
		Money:        info.Money,                    //提现金额
		PayTime:      time.Now().Format(timeFormat), //完成时间
		UserName:     info.UserName,
		WithdrawCode: w.Code(),
		WithdrawName: w.Name(),
		Status:       qresult.Status, //提现状态
		FailCode:     qresult.FailCode,
		FailMsg:      qresult.FailMsg,
	}
}

//签名
func (w *withdraw) sign(params map[string]string) error {
	signer := params["merId"] + params["merDate"] + params["merSeqId"] +
		params["cardNo"] + params["usrName"] + params["certType"] +
		params["certId"] + params["openBank"] + params["prov"] +
		params["city"] + params["transAmt"] + params["purpose"] +
		params["subBank"] + params["flag"] + params["version"] +
		params["termType"] + params["payMode"] + params["userId"] +
		params["userRegisterTime"] + params["userMail"] +
		params["userMobile"] + params["diskSn"] +
		params["mac"] + params["imei"] + params["ip"] +
		params["coordinates"] + params["baseStationSn"] +
		params["codeInputType"] + params["mobileForBank"] + params["desc"]
	log(utils.LogLevelTrace, "待编码的签名字符串：%s", signer)
	signer = base64.StdEncoding.EncodeToString([]byte(signer))
	log(utils.LogLevelTrace, "待签名字符串:%s", signer)
	params[w.config.SignatureField] = w.privKey.Sign(signer)
	log(utils.LogLevelTrace, "签名结果:%s", params[w.config.SignatureField])
	return nil
}

func (w *withdraw) QueryWithdraw(tradeno string, tradeDate ...time.Time) *payment.WithdrawQueryResult {
	if tradeDate == nil || len(tradeDate) < 1 {
		tradeDate = []time.Time{time.Now()}
	}
	returnDealign := &payment.WithdrawQueryResult{
		Status:  payment.DEALING, //提现状态
		TradeNo: tradeno,         //交易流水号
	}
	version := "20090501"
	merDate := tradeDate[0].Format("20060102")
	signValue := w.config.MerID + merDate + tradeno + version
	args := url.Values{
		"merId":    []string{w.config.MerID}, //商户号
		"merDate":  []string{merDate},        //商户日期
		"merSeqId": []string{tradeno},        //流水号
		"version":  []string{version},
		"signFlag": []string{"1"},
		"chkValue": []string{w.privKey.Sign(base64.StdEncoding.EncodeToString([]byte(signValue)))},
	}
	request, err := http.NewRequest("POST", w.queryWithdrawURL, strings.NewReader(args.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		log(utils.LogLevelError, "银联提现查询请求创建失败:%s", err.Error())
		return returnDealign
	}
	client := http.Client{
		Timeout: 1 * time.Minute,
	}
	response, err := client.Do(request)
	if err != nil {
		log(utils.LogLevelError, "银联提现查询请求失败:%s", err.Error())
		return returnDealign
	}
	responseData, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log(utils.LogLevelError, "银联提现查询请求结果读取失败:%s", err.Error())
		return returnDealign
	}
	log(utils.LogLevelInfo, "银联提现查询请求结果:%s", responseData)
	responseString := string(responseData)
	result := strings.Split(responseString, "|")
	v := w.pubKey.Verify(base64.StdEncoding.EncodeToString([]byte(strings.Join(result[:len(result)-1], "|")+"|")), result[len(result)-1])
	if !v {
		log(utils.LogLevelError, "银联提现查询结果验签失败:%s", responseString)
		return returnDealign
	}
	if result[0] == "000" {
		if len(result) > 14 {
			switch result[14] {
			case "s":
				//交易成功
				return &payment.WithdrawQueryResult{
					Status:      payment.SUCCESS,               //提现状态
					PayTime:     time.Now().Format(timeFormat), //完成时间
					TradeNo:     tradeno,                       //交易流水号
					ThridFlowNo: result[5],                     //第三方交易流水号
				}
			case "2", "3", "4", "5", "7", "8":
				return returnDealign
				//处理中
			case "6", "9":
				//交易失败
				ret := &payment.WithdrawQueryResult{
					TradeNo:  tradeno,      //交易流水号
					Status:   payment.FAIL, //提现状态
					FailCode: "-1",         //TODO: 银联查询接口无法获取失败原因
					FailMsg:  "银行退单",
				}
				return ret
			}
		}
	}
	//TODO： result[0] == "001" 表示没有交易记录如何处理？是否直接标记为失败，以便用户可以再次发起申请
	return returnDealign
}

func (w *withdraw) Driver() string {
	return "chinapay"
}

func (w *withdraw) GetWithdraw(cfg interface{}) payment.Withdraw {
	var conf *WithdrawConfig
	ok := false
	if conf, ok = cfg.(*WithdrawConfig); !ok || w == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的银联提现配置")
		return nil
	}
	if conf.SignatureField == "" {
		conf.SignatureField = "chkValue"
	}
	obj := &withdraw{
		config:           conf,
		withdrawURL:      "http://sfj.chinapay.com/dac/SinPayServletUTF8",
		queryWithdrawURL: "http://sfj.chinapay.com/dac/SinPayQueryServletUTF8",
	}
	if conf.TestMode {
		obj.withdrawURL = "http://sfj-test.chinapay.com/dac/SinPayServletUTF8"
		obj.queryWithdrawURL = "http://sfj-test.chinapay.com/dac/SinPayQueryServletUTF8"
	}
	v, err := BuildNetPayClientKey(conf.PrivateKey)
	if err != nil {
		log(utils.LogLevelError, "私钥初始化失败:%s", err.Error())
	}
	obj.privKey = v
	obj.pubKey, err = BuildNetPayClientKey(conf.PublicKey)
	if err != nil {
		log(utils.LogLevelError, "公钥初始化失败:%s", err.Error())
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}
