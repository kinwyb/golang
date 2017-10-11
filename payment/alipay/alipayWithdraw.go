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
	"crypto/sha1"
	"crypto/x509"

	"bytes"
	"sort"

	"encoding/pem"

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
func (w *withdraw) Withdraw(info *payment.WithdrawInfo) (*payment.WithdrawResult, error) {
	request := &withdrawAPIRequest{
		OutBizNo: info.TradeNo,
		Type:     "ALIPAY_LOGONID",
		Account:  info.CardNo,
		RealName: info.UserName,
		Amount:   fmt.Sprint("%.2f", info.Money),
		Remark:   info.Desc,
	}
	requestbytes, err := request.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("请求参数编码错误:%s", err.Error())
	}
	args := map[string]string{
		"appid":       w.config.Partner,
		"method":      "alipay.fund.trans.toaccount.transfer",
		"format":      "JSON",
		"charset":     "utf-8",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": string(requestbytes),
	}
	w.sign(args)
	params := url.Values{}
	for k, v := range args {
		params.Add(k, v)
	}
	resp, err := http.PostForm(w.gateway, params)
	if err != nil {
		return nil, fmt.Errorf("请求异常:%s", err.Error())
	}
	respdata, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("请求结果读取错误:%s", err.Error())
	}
	vmap := map[string]string{}
	json.Unmarshal(respdata, &vmap)
	if w.verify(vmap) {
		log(utils.LogLevelError, "支付宝提现请求结果签名验证异常:%v", vmap)
		return &payment.WithdrawResult{
			TradeNo:      info.TradeNo,
			CardNo:       info.CardNo,
			CertID:       info.CertID,
			Money:        info.Money,
			UserName:     info.UserName,
			WithdrawCode: w.Code(),
			Status:       payment.DEALING,
		}, nil
	}
	response := &withdrawAPIResponse{}
	err = response.UnmarshalJSON(respdata)
	if err != nil {
		return nil, fmt.Errorf("请求结果解析失败:%s", err.Error())
	}
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
		}, nil
	} else if response.SubCode == "SYSTEM_ERROR" { //请求结果提示业务繁忙的,调用查询接口确认一下业务是否真实失败
		qret := w.QueryWithdraw(info.TradeNo)
		if qret.Status == payment.FAIL { //提现失败
			return nil, fmt.Errorf("提现失败[%s]:%s", qret.FailCode, qret.FailMsg)
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
			}, nil
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
			}, nil
		}
	}
	return nil, fmt.Errorf("提现失败[%s]:%s", response.SubCode, response.SubMsg)
}

//根据交易单号查询提现信息
func (w *withdraw) QueryWithdraw(tradeno string, tradeDate ...time.Time) *payment.WithdrawQueryResult {
	args := map[string]string{
		"appid":       w.config.Partner,
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
	resp, err := http.PostForm(w.gateway, params)
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
	vmap := map[string]string{}
	json.Unmarshal(respdata, &vmap)
	if w.verify(vmap) {
		log(utils.LogLevelError, "支付宝提现查询请求结果签名验证异常:%v", vmap)
		return &payment.WithdrawQueryResult{
			Status:  payment.DEALING,
			TradeNo: tradeno,
		}
	}
	response := &withdrawQueryAPIResponse{}
	err = json.Unmarshal(respdata, &response)
	if err != nil {
		log(utils.LogLevelError, "支付宝提现查询请求结果解析失败:%s", err.Error())
		return &payment.WithdrawQueryResult{
			Status:  payment.DEALING,
			TradeNo: tradeno,
		}
	}
	ret := &payment.WithdrawQueryResult{
		Status:  payment.DEALING, //默认处理中
		TradeNo: tradeno,
	}
	if response.Code == "10000" { //业务请求成功
		switch resp.Status {
		case "SUCCESS":
			ret.Status = payment.SUCCESS
			ret.PayTime = response.PayDate
			ret.ThridFlowNo = response.OrderID
		case "FAIL", "REFUND":
			ret.Status = payment.FAIL
			ret.FailCode = response.ErrorCode
			ret.FailMsg = response.FailReason
		}
		return ret
	} else if response.SubCode == "SYSTEM_ERROR" { //返回业务繁忙的默认是处理中
		return ret
	}
	//其他的异常查询结果失败
	return &payment.WithdrawQueryResult{
		Status:   payment.FAIL,
		FailCode: response.ErrorCode,
		FailMsg:  response.FailReason,
	}
}

//签名
func (w *withdraw) sign(args map[string]string) {
	keys := w.paraFilter(args)
	signStr := w.createLinkString(keys, args)
	data, err := w.decodeRSAKey(w.config.PrivateKey)
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
	args["sign_type"] = w.signType
}

//过滤
func (w *withdraw) paraFilter(params map[string]string) []string {
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
func (w *withdraw) createLinkString(keys []string, args map[string]string) string {
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

//verify 支付结果校验
func (w *withdraw) verify(params map[string]string) bool {
	if v, ok := params["notify_id"]; ok {
		if !w.verifyResponse(v) {
			return false
		}
	}
	sign, _ := base64.StdEncoding.DecodeString(params["sign"])
	keys := w.paraFilter(params)
	signStr := w.createLinkString(keys, params)
	data, err := w.decodeRSAKey(w.config.PublicKey)
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

//decodeRSAKey 解析RSA密钥
func (w *withdraw) decodeRSAKey(key string) ([]byte, error) {
	if key[0] == '-' {
		block, _ := pem.Decode([]byte(key))
		if block == nil {
			return nil, errors.New("alipay签名私钥解析失败")
		}
		return block.Bytes, nil
	}
	return base64.StdEncoding.DecodeString(key)
}

//获取远程服务器ATN结果,验证返回URL
func (w *withdraw) verifyResponse(notifyID string) bool {
	verifyURL := w.verifyURL + "partner=" + w.config.Partner + "&notify_id=" + notifyID
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
