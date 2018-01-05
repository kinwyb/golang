package chanpay

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/utils"
)

//畅捷网关支付

type bankPay struct {
	payment.PayInfo
	config *BankPayConfig
	apiURL string
}

//支付,返回支付代码
func (b *bankPay) Pay(req *payment.PayRequest) (string, error) {
	req.No = encodeNo(req.No)
	if req.Ext == "" {
		req.Ext = "{}"
	}
	ext := &BankPayRequestExt{}
	if req.IsApp {
		ext.ChannelType = "01"
	} else {
		ext.ChannelType = "02"
	}
	err := json.Unmarshal([]byte(req.Ext), &ext)
	if err != nil {
		return "", errors.New("支付扩展信息[Ext]解析错误:" + err.Error())
	} else if ext.BankCode == "" {
		return "", errors.New("扩展信息银行编码[BankCode]不能为空")
	} else if ext.BizType == "" {
		return "", errors.New("扩展信息账户类型[BizType]不能为空")
	} else if ext.ChannelType == "" {
		ext.ChannelType = "02"
	}
	t := time.Now()
	params := map[string]string{
		"Service":      "nmg_ebank_pay",
		"Version":      "1.0",
		"PartnerId":    b.config.PartnerID,
		"InputCharset": "utf-8",
		"TradeDate":    t.Format("20060102"),
		"TradeTime":    t.Format("150405"),
		//"ReturnUrl":
		"OutTradeNo":     req.No,
		"MchId":          b.config.MchID,
		"ChannelType":    ext.ChannelType, //请求渠道 01:WAP、02:WEB
		"BizType":        ext.BizType,     //业务类型 01：B2C 个人网银 02：B2B 企业网银 API接口直联银行时用来判断跳转到个人/企业网银 收银台时用来判断跳转到个人/企业收银台
		"CardFlag":       "01",            //借贷标识 01：DC借记卡; 02：CC贷记卡;
		"PayFlag":        "00",            //支付标识 00：API接口直联 01：畅捷收银台
		"ServiceType":    "01",            //服务类型 01：及时; 02：担保
		"TradeType":      "00",            //交易类型 00：充值;01：转账;02：还款;03：缴费;04：理财;05：消费;06：其他
		"GoodsName":      "智纺平台交易",        //交易名称
		"GoodsType":      "00",            //商品类别 00：虚拟 01：实体
		"BankCode":       ext.BankCode,    //银行编码 API接口直联时必须输入
		"Currency":       "00",            //货币类型 默认00：CNY，暂只支持人民币
		"OrderAmt":       fmt.Sprintf("%.2f", req.Money),
		"OrderStartTime": t.Format("20060102150405"),
		"UserIp":         req.IP,
		"NotifyUrl":      b.config.NotifyURL,
	}
	err = sign(params, b.config.PrivateKey)
	if err != nil {
		return "", errors.New("签名失败")
	}
	buf := bytes.NewBufferString("<form id=\"chanpaysubmit\" name=\"chanpaysubmit\" action=\"")
	buf.WriteString(b.apiURL)
	buf.WriteString("\" method=\"POST\">")
	for k, v := range params {
		buf.WriteString("<input type=\"hidden\" name=\"")
		buf.WriteString(k)
		buf.WriteString("\" value=\"")
		buf.WriteString(v)
		buf.WriteString("\"/>")
	}
	buf.WriteString("<input type=\"submit\" value=\"提交\" style=\"display:none;\"></form>")
	buf.WriteString("<script>document.forms['chanpaysubmit'].submit();</script>")
	return buf.String(), nil
}

//异步结果通知处理,返回支付结果
func (b *bankPay) Notify(params map[string]string) *payment.PayResult {
	delete(params, "request_post_body")
	result := &payment.PayResult{
		PayCode: b.Code(),
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
	} else if verify(params, b.config.PublicKey) {
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
func (b *bankPay) NotifyResult(payResult *payment.PayResult) string {
	if payResult.Succ {
		return "success"
	}
	return "fail"
}

//同步结果跳转处理,返回支付结果
func (b *bankPay) Result(params map[string]string) *payment.PayResult {
	return b.Notify(params)
}

//获取驱动编码
func (b *bankPay) Driver() string {
	return "chanpaybank"
}

//生成一个支付对象
func (b *bankPay) GetPayment(cfg interface{}) payment.Payment {
	var c *BankPayConfig
	ok := false
	if c, ok = cfg.(*BankPayConfig); !ok || c == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的畅捷扫码支付配置")
		return nil
	}
	if c.Name == "" || c.Code == "" {
		return nil
	}
	obj := &bankPay{
		apiURL: "https://pay.chanpay.com/mag-unify/gateway/receiveOrder.do",
		config: c,
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}

//无需确认支付
func (b *bankPay) PayConfirm(req *payment.PayConfirmRequest) *payment.PayResult {
	return payment.NoPayConfirmResult
}
