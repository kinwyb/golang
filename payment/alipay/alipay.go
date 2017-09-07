package alipay

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"git.oschina.net/kinwyb/golang/payment"

	"git.oschina.net/kinwyb/golang/utils"

	"fmt"

	"sort"

	"bytes"

	"crypto/rand"
	"crypto/x509"

	"crypto/rsa"

	"crypto/sha1"

	"crypto"

	"encoding/base64"
	"encoding/pem"

	"strconv"
)

type alipay struct {
	payment.PayInfo
	config       *PayConfig
	gateway      string //支付宝提供给商户的服务接入网关URL(新)
	verifyURL    string //支付宝消息验证地址
	signType     string //签名方式
	inputCharset string //字符编码
}

//支付,返回支付代码
func (a *alipay) Pay(req *payment.PayRequest) (string, error) {
	sParams := map[string]string{
		"service":        "create_direct_pay_by_user",
		"partner":        a.config.Partner,
		"seller_id":      a.config.Partner,
		"_input_charset": a.inputCharset,
		"payment_type":   "1",
		"notify_url":     a.config.NotifyURL,
		"return_url":     a.config.ReturnURL,
		"subject":        req.Desc,
		"total_fee":      fmt.Sprintf("%.2f", req.Money),
		"out_trade_no":   req.No,
	}
	a.sign(sParams)
	buf := bytes.NewBufferString("<form id=\"alipaysubmit\" name=\"alipaysubmit\" action=\"")
	buf.WriteString(a.gateway)
	buf.WriteString("_input_charset=\"")
	buf.WriteString(a.inputCharset)
	buf.WriteString("\" method=\"get\">")
	for k, v := range sParams {
		buf.WriteString("<input type=\"hidden\" name=\"")
		buf.WriteString(k)
		buf.WriteString("\" value=\"")
		buf.WriteString(v)
		buf.WriteString("\"/>")
	}
	buf.WriteString("<input type=\"submit\" value=\"提交\" style=\"display:none;\"></form>")
	buf.WriteString("<script>document.forms['alipaysubmit'].submit();</script>")
	return buf.String(), nil
}

//异步结果通知处理,返回支付结果
func (a *alipay) Notify(params map[string]string) *payment.PayResult {
	if params["trade_status"] == "WAIT_BUYER_PAY" {
		return nil
	}
	delete(params, "request_post_body")
	result := &payment.PayResult{
		PayCode: a.Code(),
		Navite:  params,
	}
	if _, ok := params["total_fee"]; !ok {
		result.ErrMsg = "支付宝回调数据错误"
		result.Succ = false
		return result
	}
	var err error
	result.TradeNo = params["out_trade_no"]  //商户订单号
	result.No = params["out_trade_no"]       //原始订单号
	result.ThirdTradeNo = params["trade_no"] //支付宝交易号
	result.Money, err = strconv.ParseFloat(params["total_fee"], 64)
	if err != nil {
		result.Succ = false
		result.ErrMsg = "支付宝回调数据错误"
	} else if a.verify(params) {
		status := params["trade_status"]
		if status == "TRADE_FINISHED" || status == "TRADE_SUCCESS" {
			result.Succ = true
			result.ThirdAccount = params["buyer_email"]
		} else {
			result.Succ = false
		}
	} else {
		result.Succ = false
		result.ErrMsg = "支付宝回调数据验证失败"
	}
	return result
}

//同步结果跳转处理,返回支付结果
func (a *alipay) Result(params map[string]string) *payment.PayResult {
	return a.Notify(params)
}

//NotifyResult 通知结果返回内容
func (a *alipay) NotifyResult(payResult *payment.PayResult) string {
	if payResult.Succ {
		return "success"
	}
	return "fail"
}

//GetPayment 生成一个支付对象
func (a *alipay) GetPayment(cfg interface{}) payment.Payment {
	var c *PayConfig
	ok := false
	if c, ok = cfg.(*PayConfig); !ok || c == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的支付宝配置")
		return nil
	}
	if c.Name == "" || c.Code == "" {
		return nil
	}
	obj := &alipay{
		gateway:      "https://mapi.alipay.com/gateway.do?",
		verifyURL:    "https://mapi.alipay.com/gateway.do?service=notify_verify&",
		signType:     "RSA",
		inputCharset: "UTF-8",
		config:       c,
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}

func (a *alipay) Driver() string {
	return "alipay"
}

//过滤
func (a *alipay) paraFilter(params map[string]string) []string {
	keys := make([]string, 0)
	for k, v := range params {
		if k == "sign" || k == "sign_type" || strings.TrimSpace(v) == "" {
			delete(params, k)
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

//拼接字符串 按照“参数=参数值”的模式用“&”字符拼接成字符串
func (a *alipay) createLinkString(keys []string, args map[string]string) string {
	buf := bytes.NewBufferString("")
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(args[k])
		buf.WriteString("&")
	}
	buf.Truncate(buf.Len() - 1)
	return buf.String()
}

//签名
func (a *alipay) sign(args map[string]string) {
	keys := a.paraFilter(args)
	signStr := a.createLinkString(keys, args)
	data, err := a.decodeRSAKey(a.config.PrivateKey)
	if err != nil {
		log(utils.LogLevelError, "alipay私钥解析失败")
		return
	}
	priv, err := x509.ParsePKCS8PrivateKey(data)
	if err != nil {
		log(utils.LogLevelError, "alipay签名RSA私钥初始化失败:"+err.Error())
		return
	}
	dt := sha1.Sum([]byte(signStr))
	data, err = rsa.SignPKCS1v15(rand.Reader, priv.(*rsa.PrivateKey), crypto.SHA1, dt[:])
	if err != nil {
		log(utils.LogLevelError, "alipay签名失败:"+err.Error())
		return
	}
	args["sign"] = base64.StdEncoding.EncodeToString(data)
	args["sign_type"] = a.signType
}

//decodeRSAKey 解析RSA密钥
func (a *alipay) decodeRSAKey(key string) ([]byte, error) {
	if key[0] == '-' {
		block, _ := pem.Decode([]byte(key))
		if block == nil {
			return nil, errors.New("alipay签名私钥解析失败")
		}
		return block.Bytes, nil
	}
	return base64.StdEncoding.DecodeString(key)
}

//verify 支付结果校验
func (a *alipay) verify(params map[string]string) bool {
	if v, ok := params["notify_id"]; ok {
		if !a.verifyResponse(v) {
			return false
		}
	}
	sign, _ := base64.StdEncoding.DecodeString(params["sign"])
	keys := a.paraFilter(params)
	signStr := a.createLinkString(keys, params)
	data, err := a.decodeRSAKey(a.config.PublicKey)
	if err != nil {
		log(utils.LogLevelError, "alipay公钥解析失败")
		return false
	}
	pubi, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		log(utils.LogLevelError, "alipay结果校验RSA公钥初始化错误:"+err.Error())
		return false
	}
	dt := sha1.Sum([]byte(signStr))
	err = rsa.VerifyPKCS1v15(pubi.(*rsa.PublicKey), crypto.SHA1, dt[:], sign)
	if err != nil {
		log(utils.LogLevelError, "alipay结果校验失败:"+err.Error())
		return false
	}
	return true
}

//获取远程服务器ATN结果,验证返回URL
func (a *alipay) verifyResponse(notifyID string) bool {
	verifyURL := a.verifyURL + "partner=" + a.config.Partner + "&notify_id=" + notifyID
	resp, err := http.Get(verifyURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return string(body) == "true"
}

//支付宝支付无需确认支付
func (a *alipay) PayConfirm(req *payment.PayConfirmRequest) *payment.PayResult {
	return payment.NoPayConfirmResult
}
