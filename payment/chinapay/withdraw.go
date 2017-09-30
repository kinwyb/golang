package chinapay

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"net/http"
	"net/url"

	"io/ioutil"

	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/utils"
)

type withdraw struct {
	payment.PayInfo
	config *PayConfig
	apiURL string
	sess   *NetPaySecssUtil
}

func (w *withdraw) Withdraw(info *payment.WithdrawInfo) (*payment.WithdrawResult, error) {
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
	request, err := http.NewRequest("http://sfj.chinapay.com/dac/SinPayServletUTF8", "application/x-www-form-urlencoded", strings.NewReader(params.Encode()))
	if err != nil {
		log(utils.LogLevelError, "银联提现请求创建失败:%s", err.Error())
		return nil, errors.New("请求创建异常")
	}
	client := http.Client{
		Timeout: 1 * time.Minute,
	}
	response, err := client.Do(request)
	if err != nil {
		log(utils.LogLevelError, "银联提现请求失败:%s", err.Error())
		return nil, errors.New("请求异常")
	}
	responseData, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log(utils.LogLevelError, "银联提现请求结果读取失败:%s", err.Error())
		return nil, errors.New("请求结果读取异常")
	}
	log(utils.LogLevelInfo, "银联提现请求结果:%s", responseData)
	res, err := url.ParseQuery(string(responseData))
	if err != nil { //要检测提现是否完成
		return nil, errors.New("银联提现结果读取异常")
	}
	if res["responseCode"][0] == "000" { //表示请求成功

	}
	//TODO: .....
	return nil, nil
}

//签名
func (w *withdraw) sign(params map[string]string) error {
	signer, err := w.sess.Sign(params)
	if err != nil {
		return errors.New("签名生成失败:" + err.Error())
	}
	params[w.config.SignatureField] = signer
	return nil
}

func (w *withdraw) QueryWithdraw(tradeno string) *payment.WithdrawQueryResult {
	return nil
}

func (w *withdraw) Driver() string {
	return "chinapay"
}

func (w *withdraw) GetWithdraw(interface{}) payment.Withdraw {
	return nil
}
