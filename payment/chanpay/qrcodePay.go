package chanpay

import (
	"errors"
	"strconv"
	"time"

	"fmt"

	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/payment/wxpay"
	"github.com/kinwyb/golang/utils"
)

//畅捷二维码扫描支付

type qrcodePay struct {
	payment.PayInfo
	config *QRPayConfig
	apiURL string
}

//支付,返回支付代码
func (q *qrcodePay) Pay(req *payment.PayRequest) (string, error) {
	if req.Ext == "" {
		req.Ext = "ALIPAY"
	}
	req.No = encodeNo(req.No)
	t := time.Now()
	params := map[string]string{
		"Service":        "mag_init_code_pay",
		"Version":        "1.0",
		"PartnerId":      q.config.PartnerID,
		"InputCharset":   "utf-8",
		"TradeDate":      t.Format("20060102"),
		"TradeTime":      t.Format("150405"),
		"OutTradeNo":     req.No,
		"MchId":          q.config.MchID,
		"TradeType":      "11",
		"BankCode":       req.Ext,
		"TradeAmount":    fmt.Sprintf("%.2f", req.Money),
		"GoodsName":      req.Desc,
		"Subject":        req.Desc,
		"OrderStartTime": t.Format("20060102150405"),
		"SpbillCreateIp": req.IP,
		"NotifyUrl":      q.config.NotifyURL,
	}
	result, err := request(q.apiURL, params, q.config.PrivateKey, q.config.PublicKey)
	if err != nil {
		return "", err
	}
	img, err := wxpay.QRCode(result["CodeUrl"], 150)
	if err != nil {
		log(utils.LogLevelError, "二维码创建失败:%s", err.Error())
		return result["CodeUrl"], errors.New("二维码创建失败")
	}
	return img, nil
}

//异步结果通知处理,返回支付结果
func (q *qrcodePay) Notify(params map[string]string) *payment.PayResult {
	delete(params, "request_post_body")
	result := &payment.PayResult{
		PayCode: q.Code(),
		Navite:  params,
	}
	var err error
	No := decodeNo(params["outer_trade_no"])
	result.TradeNo = No                            //商户订单号
	result.No = No                                 //原始订单号
	result.ThirdTradeNo = params["inner_trade_no"] //畅捷平台订单号
	result.Money, err = strconv.ParseFloat(params["trade_amount"], 64)
	if err != nil {
		log(utils.LogLevelError, err.Error())
		result.Succ = false
		result.ErrMsg = "畅捷支付回调数据错误"
	} else if verify(params, q.config.PublicKey) {
		status := params["trade_status"]
		if status == "TRADE_SUCCESS" || status == "TRADE_FINISHED" {
			result.Succ = true
		} else {
			result.Succ = false
		}
	} else {
		result.Succ = false
		result.ErrMsg = "畅捷支付回调数据验证失败"
	}
	return result
}

//异步通知处理结果返回内容
func (q *qrcodePay) NotifyResult(payResult *payment.PayResult) string {
	if payResult.Succ {
		return "success"
	}
	return "fail"
}

//同步结果跳转处理,返回支付结果
func (q *qrcodePay) Result(params map[string]string) *payment.PayResult {
	return nil
}

//获取驱动编码
func (q *qrcodePay) Driver() string {
	return "chanpayqrcode"
}

//生成一个支付对象
func (q *qrcodePay) GetPayment(cfg interface{}) payment.Payment {
	var c *QRPayConfig
	ok := false
	if c, ok = cfg.(*QRPayConfig); !ok || c == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的畅捷扫码支付配置")
		return nil
	}
	if c.Name == "" || c.Code == "" {
		return nil
	}
	obj := &qrcodePay{
		apiURL: "https://pay.chanpay.com/mag-unify/gateway/receiveOrder.do",
		config: c,
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}

//无需确认支付
func (q *qrcodePay) PayConfirm(req *payment.PayConfirmRequest) *payment.PayResult {
	return payment.NoPayConfirmResult
}
