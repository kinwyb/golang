package alipay

import (
	"crypto"
	"errors"
	"io/ioutil"
	"strings"
	"time"

	"fmt"

	"encoding/base64"
	"encoding/json"

	"net/http"

	"net/url"

	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"bytes"
	"sort"

	"encoding/pem"

	"crypto/sha256"

	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/utils"
)

//withdraw 支付宝提现
type withdraw struct {
	payment.PayInfo
	config    *PayConfig
	gateway   string
	verifyURL string
	signType  string
}

//提现操作,成功返回第三方交易流水,失败返回错误
func (w *withdraw) Withdraw(info *payment.WithdrawInfo) *payment.WithdrawResult {
	request := &withdrawAPIRequest{
		OutBizNo: info.TradeNo,
		Type:     "ALIPAY_LOGONID",
		Account:  info.CardNo,
		RealName: info.UserName,
		Amount:   fmt.Sprintf("%.2f", info.Money),
		Remark:   info.Desc,
	}
	requestbytes, err := request.MarshalJSON()
	if err != nil {
		return payment.WithdrawParamsSerializeFail
	}
	args := map[string]string{
		"app_id":      w.config.Partner,
		"method":      "alipay.fund.trans.toaccount.transfer",
		"format":      "json",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(requestbytes),
	}
	w.sign(args)
	params := url.Values{}
	for k, v := range args {
		params.Add(k, v)
	}
	log(utils.LogLevelDebug, "支付宝提现请求参数:%s", params.Encode())
	resp, err := http.Post(w.gateway, "application/x-www-form-urlencoded;charset=utf-8", strings.NewReader(params.Encode()))
	if err != nil {
		log(utils.LogLevelError, "支付宝提现请求异常:%s", err.Error())
		return payment.WithdrawRequestFail
	}
	respdata, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log(utils.LogLevelError, "支付宝提现请求结果读取异常:%s", err.Error())
		return payment.WithdrawResponseReadFail
	}
	log(utils.LogLevelInfo, "支付宝提现结果:%s", respdata)
	vmap := &withdrawAPIResp{}
	err = json.Unmarshal(respdata, &vmap)
	if err != nil {
		log(utils.LogLevelError, "结果解析错误:%s", err.Error())
	}
	if vmap.Sign != "" && !w.verify(string(respdata), vmap.Sign, 49) {
		log(utils.LogLevelError, "支付宝提现请求结果签名验证异常")
		return &payment.WithdrawResult{
			TradeNo:      info.TradeNo,
			CardNo:       info.CardNo,
			CertID:       info.CertID,
			Money:        info.Money,
			UserName:     info.UserName,
			WithdrawCode: w.Code(),
			Status:       payment.DEALING,
		}
	}
	response := vmap.Method
	if response.Code == "10000" {
		return &payment.WithdrawResult{
			TradeNo:      response.OutBizNo,
			ThridFlowNo:  response.OrderID,
			CardNo:       info.CardNo,
			CertID:       info.CertID,
			Money:        info.Money,
			PayTime:      response.PayDate,
			UserName:     info.UserName,
			WithdrawCode: w.Code(),
			WithdrawName: w.Name(),
			Status:       payment.SUCCESS,
		}
	} else if response.SubCode == "SYSTEM_ERROR" { //请求结果提示业务繁忙的,调用查询接口确认一下业务是否真实失败
		qret := w.QueryWithdraw(info.TradeNo)
		if qret.Status == payment.FAIL { //提现失败
			return &payment.WithdrawResult{
				Status:   payment.FAIL,
				FailCode: qret.FailCode,
				FailMsg:  qret.FailMsg,
			}
		} else if qret.Status == payment.SUCCESS { //查询出来是成功的返回成功
			return &payment.WithdrawResult{
				TradeNo:      response.OutBizNo,
				ThridFlowNo:  qret.ThridFlowNo,
				CardNo:       info.CardNo,
				CertID:       info.CertID,
				Money:        info.Money,
				PayTime:      response.PayDate,
				UserName:     info.UserName,
				WithdrawCode: w.Code(),
				WithdrawName: w.Name(),
				Status:       payment.SUCCESS,
			}
		} else { //处理中的返回处理中
			return &payment.WithdrawResult{
				TradeNo:      response.OutBizNo,
				ThridFlowNo:  qret.ThridFlowNo,
				CardNo:       info.CardNo,
				CertID:       info.CertID,
				Money:        info.Money,
				PayTime:      response.PayDate,
				UserName:     info.UserName,
				WithdrawCode: w.Code(),
				WithdrawName: w.Name(),
				Status:       payment.DEALING,
			}
		}
	}
	return &payment.WithdrawResult{
		Status:   payment.FAIL,
		FailCode: response.SubCode,
		FailMsg:  response.SubMsg,
	}
}

//根据交易单号查询提现信息
func (w *withdraw) QueryWithdraw(tradeno string, tradeDate ...time.Time) *payment.WithdrawQueryResult {
	args := map[string]string{
		"app_id":      w.config.Partner,
		"method":      "alipay.fund.trans.order.query",
		"format":      "JSON",
		"charset":     "utf-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": `{"out_biz_no":"` + tradeno + `"}`,
	}
	w.sign(args)
	params := url.Values{}
	for k, v := range args {
		params.Add(k, v)
	}
	resp, err := http.Post(w.gateway, "application/x-www-form-urlencoded;charset=utf-8", strings.NewReader(params.Encode()))
	if err != nil {
		log(utils.LogLevelError, "支付宝提现查询请求异常:%s", err.Error())
		return &payment.WithdrawQueryResult{
			Status:  payment.DEALING,
			TradeNo: tradeno,
		}
	}
	respdata, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log(utils.LogLevelError, "支付宝提现查询请求结果读取错误:%s", err.Error())
		return &payment.WithdrawQueryResult{
			Status:  payment.DEALING,
			TradeNo: tradeno,
		}
	}
	log(utils.LogLevelInfo, "支付宝提现查询结果:%s", respdata)
	vmap := &withdrawQueryAPIResp{}
	err = json.Unmarshal(respdata, &vmap)
	if err != nil {
		log(utils.LogLevelError, "支付宝提现查询结果解析错误:%s", err.Error())
	}
	if vmap.Sign != "" && !w.verify(string(respdata), vmap.Sign, 42) {
		log(utils.LogLevelError, "支付宝提现查询请求结果签名验证异常:%v", vmap)
		return &payment.WithdrawQueryResult{
			Status:  payment.DEALING,
			TradeNo: tradeno,
		}
	}
	response := vmap.Method
	ret := &payment.WithdrawQueryResult{
		Status:  payment.DEALING, //默认处理中
		TradeNo: tradeno,
	}
	if response.Code == "10000" { //业务请求成功
		log(utils.LogLevelError, "请求成功")
		switch response.Status {
		case "SUCCESS":
			ret.Status = payment.SUCCESS
			ret.PayTime = response.PayDate
			ret.ThridFlowNo = response.OrderID
		case "FAIL", "REFUND":
			ret.Status = payment.FAIL
			ret.FailCode = response.ErrorCode
			ret.FailMsg = response.FailReason
		}
	}
	return ret
}

//签名
func (w *withdraw) sign(args map[string]string) {
	keys := w.paraFilter(args)
	signStr := w.createLinkString(keys, args)
	data, err := w.decodeRSAKey(w.config.PrivateKey)
	if err != nil {
		log(utils.LogLevelError, "支付宝提现私钥解析失败")
		return
	}
	priv, err := x509.ParsePKCS8PrivateKey(data)
	if err != nil {
		log(utils.LogLevelError, "支付宝提现签名RSA私钥初始化失败:"+err.Error())
		return
	}
	log(utils.LogLevelDebug, "支付宝提现签名字符串:%s", signStr)
	dt := sha256.Sum256([]byte(signStr))
	data, err = rsa.SignPKCS1v15(rand.Reader, priv.(*rsa.PrivateKey), crypto.SHA256, dt[:])
	if err != nil {
		log(utils.LogLevelError, "支付宝提现签名失败:"+err.Error())
		return
	}
	args["sign"] = base64.StdEncoding.EncodeToString(data)
	log(utils.LogLevelError, "签名结果:%s", args["sign"])
}

//过滤
func (w *withdraw) paraFilter(params map[string]string) []string {
	keys := make([]string, 0)
	for k, v := range params {
		if k == "sign" || strings.TrimSpace(v) == "" {
			delete(params, k)
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

//拼接字符串 按照“参数=参数值”的模式用“&”字符拼接成字符串
func (w *withdraw) createLinkString(keys []string, args map[string]string) string {
	buf := bytes.NewBufferString("")
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(args[k])
		buf.WriteString("&")
	}
	if buf.Len() > 0 {
		buf.Truncate(buf.Len() - 1)
	}
	return buf.String()
}

//verify 支付结果校验
func (w *withdraw) verify(response string, signString string, start int) bool {
	response = response[start:strings.LastIndex(response, ",\"sign\":")]
	sign, _ := base64.StdEncoding.DecodeString(signString)
	data, err := w.decodeRSAKey(w.config.PublicKey)
	if err != nil {
		log(utils.LogLevelError, "支付宝提现公钥解析失败")
		return false
	}
	pubi, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		log(utils.LogLevelError, "支付宝提现结果校验RSA公钥初始化错误:"+err.Error())
		return false
	}
	dt := sha256.Sum256([]byte(response))
	err = rsa.VerifyPKCS1v15(pubi.(*rsa.PublicKey), crypto.SHA256, dt[:], sign)
	if err != nil {
		log(utils.LogLevelError, "支付宝提现结果校验失败:"+err.Error())
		log(utils.LogLevelError, "支付宝提现校验签名的字符串:%s", response)
		log(utils.LogLevelError, "支付宝提现校验的签名:%s", signString)
		return false
	}
	return true
}

//decodeRSAKey 解析RSA密钥
func (w *withdraw) decodeRSAKey(key string) ([]byte, error) {
	if key[0] == '-' {
		block, _ := pem.Decode([]byte(key))
		if block == nil {
			return nil, errors.New("支付宝提现签名私钥解析失败")
		}
		return block.Bytes, nil
	}
	return base64.StdEncoding.DecodeString(key)
}

//获取驱动编码
func (w *withdraw) Driver() string {
	return "alipay"
}

//生成一个提现对象
func (w *withdraw) GetWithdraw(cfg interface{}) payment.Withdraw {
	var c *PayConfig
	ok := false
	if c, ok = cfg.(*PayConfig); !ok || c == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的支付宝配置")
		return nil
	}
	if c.Name == "" || c.Code == "" {
		return nil
	}
	obj := &withdraw{
		gateway:   "https://openapi.alipay.com/gateway.do",
		verifyURL: "https://mapi.alipay.com/gateway.do?service=notify_verify&",
		signType:  "RSA2",
		config:    c,
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}
