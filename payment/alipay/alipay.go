package alipay

import (
	"io/ioutil"
	"net/http"

	"github.com/kinwyb/golang/payment"

	"github.com/kinwyb/golang/utils"

	"fmt"

	"encoding/json"
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
	service := "alipay.trade.page.pay"
	sParams := map[string]string{
		"subject":      req.Desc,
		"total_amount": fmt.Sprintf("%.2f", req.Money),
		"out_trade_no": req.No,
		"product_code": "FAST_INSTANT_TRADE_PAY",
	}
	requestbytes, err := json.Marshal(sParams)
	if err != nil {
		return "", fmt.Errorf("参数序列化错误")
	}
	if req.IsApp { //app支付
		sParams["product_code"] = "QUICK_MSECURITY_PAY"
		service = "alipay.trade.app.pay"
		respdata, err := request(service, a.config, string(requestbytes), a.gateway)
		if err != nil {
			return "", err
		}
		return string(respdata), nil
	}
	return buildForm(service, a.config, string(requestbytes), a.gateway), nil
}

//异步结果通知处理,返回支付结果
func (a *alipay) Notify(params map[string]string) *payment.PayResult {
	if params["trade_status"] == "WAIT_BUYER_PAY" {
		return nil
	}
	delete(params, "request_post_body")
	result := &payment.PayResult{
		PayCode:      a.Code(),
		Navite:       params,
		TradeNo:      params["out_trade_no"], //商户订单号
		No:           params["out_trade_no"], //原始订单号
		ThirdTradeNo: params["trade_no"],     //支付宝交易号
	}
	if _, ok := params["total_amount"]; !ok {
		result.ErrMsg = "支付宝回调数据错误"
		result.Succ = false
		return result
	}
	var err error
	result.Money, err = strconv.ParseFloat(params["total_amount"], 64)
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
	//支付宝回调数据不存在支付结果字段，咨询客服后回答只有成功才会同步跳转，所以同步跳转结果只要验证签名即可，默认都是成功的
	result := &payment.PayResult{
		PayCode:      a.Code(),
		Navite:       params,
		TradeNo:      params["out_trade_no"], //商户订单号
		No:           params["out_trade_no"], //原始订单号
		ThirdTradeNo: params["trade_no"],     //支付宝交易号
		Succ:         true,
	}
	result.Money,_ = strconv.ParseFloat(params["total_amount"],64)
	if !a.verify(params) {
		result.Succ = false
		result.ErrMsg = "支付宝回调数据验证失败"
	}
	return result
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
		gateway:      "https://openapi.alipay.com/gateway.do",
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

//verify 支付结果校验
func (a *alipay) verify(params map[string]string) bool {
	//if v, ok := params["notify_id"]; ok {
	//	if !a.verifyResponse(v) {
	//		return false
	//	}
	//}
	sign := params["sign"]
	delete(params, "sign_type")
	keys := paraFilter(params)
	signStr := createLinkString(keys, params)
	return verify(signStr, sign, a.config.PublicKey)
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
