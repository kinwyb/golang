package chanpay

import "github.com/kinwyb/golang/payment"

//QRPayConfig 畅捷二维码扫码支付配置
//	支付请求参数PayRequest中： Ext  可空  用作表示支付方式=[WXPAY:微信渠道,ALIPAY:支付宝渠道,UNIONPAY:银联渠道]
//									   默认：ALIPAY
type QRPayConfig struct {
	payment.Config
	PartnerID  string //签约合作方的唯一用户号
	MchID      string //商户标识id
	PrivateKey []byte //签名私钥
	PublicKey  []byte //验签公钥
	NotifyURL  string //结果通知地址
}

//QuickPayConfig 畅捷快捷支付配置
//	支付请求PayRequest中：MemberID 必填
//    					Ext      必填 结构为QuickPayRequestExt
type QuickPayConfig struct {
	payment.Config
	PartnerID   string //签约合作方的唯一用户号
	MchID       string //商户标识id
	PrivateKey  []byte //私钥
	PublicKey   []byte //公钥
	ExpiredTime string //交易有效时间,取值范围：1m～48h。单位为分，如1.5h，可转换为90m。如果超过该有效期进行确认则提示订单已超时。不允许确认
	NotifyURL   string //结果通知地址
}

//QuickPayRequestExt 支付请求扩张信息
type QuickPayRequestExt struct {
	BkAcctNo     string //银行卡账号
	IDNo         string //身份证号
	CstmrNm      string //持卡人姓名
	MobNo        string //持卡人预留手机号
	IsCreditCard bool   //是否是信用卡
	CardExprDt   string //有效期[当是信用卡时必填]
	CardCvn2     string //cvv2码[信用卡时必填]
}
