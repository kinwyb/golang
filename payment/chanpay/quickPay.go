package chanpay

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"git.oschina.net/kinwyb/golang/crypto/rsautil"
	"git.oschina.net/kinwyb/golang/payment"
	"git.oschina.net/kinwyb/golang/utils"
)

//quickPay 快捷支付
type quickPay struct {
	payment.PayInfo
	config *QuickPayConfig
	apiURL string
}

//支付,返回支付代码
func (q *quickPay) Pay(req *payment.PayRequest) (string, error) {
	if req.MemberID == "" {
		return "", errors.New("用户唯一标识[MemberID]不能为空")
	} else if req.Ext == nil {
		return "", errors.New("支付扩展信息[Ext]不能为空,而且必须为*QuickPayRequestExt")
	}
	jsondata, _ := json.Marshal(req.Ext)
	ext := &QuickPayRequestExt{}
	err := json.Unmarshal(jsondata, &ext)
	if err != nil {
		return "", errors.New("支付扩展信息[Ext]解析错误:" + err.Error())
	} else if ext.BkAcctNo == "" {
		return "", errors.New("扩展信息银行卡号[BkAcctNo]不能为空")
	} else if ext.IDNo == "" {
		return "", errors.New("扩展信息身份证号[IDNo]不能为空")
	} else if ext.CstmrNm == "" {
		return "", errors.New("扩展信息持卡人姓名[CstmrNm]不能为空")
	} else if ext.MobNo == "" {
		return "", errors.New("扩展信息持卡人预留手机号[MobNo]不能为空")
	}
	t := time.Now()
	params := map[string]string{
		"Service":      "nmg_zft_api_quick_payment", //直接支付接口
		"Version":      "1.0",
		"PartnerId":    q.config.PartnerID,
		"InputCharset": "utf-8",
		"TradeDate":    t.Format("20060102"),
		"TradeTime":    t.Format("150405"),
		"TrxId":        encodeNo(req.No),
		"MerUserId":    req.MemberID,
		"SellerId":     q.config.MchID,
		"ExpiredTime":  q.config.ExpiredTime, //交易有效时间30分钟
		"TradeType":    "11",
		"BkAcctTp":     "01",
		"IDTp":         "01",
		"TrxAmt":       strconv.FormatFloat(req.Money, 'f', -1, 64),
		"OrdrName":     req.Desc,
		"SmsFlag":      "1", //短信发送标识
		"NotifyUrl":    q.config.NotifyURL,
	}
	if ext.IsCreditCard {
		params["BkAcctTp"] = "00"
		if ext.CardCvn2 == "" {
			return "", errors.New("支付扩张信息信用卡[CardCvn2]不能为空")
		} else if ext.CardExprDt == "" {
			return "", errors.New("支付扩张信息信用卡有效期[CardExprDt]不能为空")
		}
		CardCvn2, _ := rsautil.Encrypt(q.config.PublicKey, []byte(ext.CardCvn2))
		params["CardCvn2"] = base64.StdEncoding.EncodeToString(CardCvn2)
		CardExprDt, _ := rsautil.Encrypt(q.config.PublicKey, []byte(ext.CardExprDt))
		params["CardExprDt"] = base64.StdEncoding.EncodeToString(CardExprDt)
	}
	BkAcctNo, _ := rsautil.Encrypt(q.config.PublicKey, []byte(ext.BkAcctNo))
	params["BkAcctNo"] = base64.StdEncoding.EncodeToString(BkAcctNo)
	IDNo, _ := rsautil.Encrypt(q.config.PublicKey, []byte(ext.IDNo))
	params["IDNo"] = base64.StdEncoding.EncodeToString(IDNo)
	CstmrNm, _ := rsautil.Encrypt(q.config.PublicKey, []byte(ext.CstmrNm))
	params["CstmrNm"] = base64.StdEncoding.EncodeToString(CstmrNm)
	MobNo, _ := rsautil.Encrypt(q.config.PublicKey, []byte(ext.MobNo))
	params["MobNo"] = base64.StdEncoding.EncodeToString(MobNo)
	result, err := request(q.apiURL, params, q.config.PrivateKey, q.config.PublicKey)
	if err != nil {
		return "", err
	}
	return result["TrxId"], nil
}

//异步结果通知处理,返回支付结果
func (q *quickPay) Notify(params map[string]string) *payment.PayResult {
	delete(params, "request_post_body")
	result := &payment.PayResult{
		PayCode: q.Code(),
		Navite:  params,
	}
	var err error
	no := decodeNo(params["outer_trade_no"])
	result.TradeNo = no                            //商户订单号
	result.No = no                                 //原始订单号
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
func (q *quickPay) NotifyResult(payResult *payment.PayResult) string {
	if payResult.Succ {
		return "success"
	}
	return "fail"
}

//同步结果跳转处理,返回支付结果
func (q *quickPay) Result(params map[string]string) *payment.PayResult {
	return nil
}

//获取驱动编码
func (q *quickPay) Driver() string {
	return "chanpayquick"
}

//生成一个支付对象
func (q *quickPay) GetPayment(cfg interface{}) payment.Payment {
	var c *QuickPayConfig
	ok := false
	if c, ok = cfg.(*QuickPayConfig); !ok || c == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的畅捷快捷支付配置")
		return nil
	}
	if c.Name == "" || c.Code == "" {
		return nil
	}
	obj := &quickPay{
		apiURL: "https://pay.chanpay.com/mag-unify/gateway/receiveOrder.do",
		config: c,
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}

//确认支付
func (q *quickPay) PayConfirm(req *payment.PayConfirmRequest) *payment.PayResult {
	t := time.Now()
	params := map[string]string{
		"Service":      "nmg_api_quick_payment_smsconfirm", //直接支付接口
		"Version":      "1.0",
		"PartnerId":    q.config.PartnerID,
		"InputCharset": "utf-8",
		"TradeDate":    t.Format("20060102"),
		"TradeTime":    t.Format("150405"),
		"TrxId":        req.No,
		"OriPayTrxId":  req.No,
		"SmsCode":      req.VerifyCode,
	}
	result, err := request(q.apiURL, params, q.config.PrivateKey, q.config.PublicKey)
	if err != nil {
		return &payment.PayResult{
			Succ:   false,
			ErrMsg: err.Error(),
		}
	}
	return &payment.PayResult{
		Succ:         true,
		No:           req.No,
		TradeNo:      req.No,
		PayCode:      q.Code(),
		ThirdTradeNo: result["OrderTrxId"],
		Navite:       result,
	}
}
