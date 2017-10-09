package chinapay

import "github.com/kinwyb/golang/payment"

//PayConfig 支付配置信息
type PayConfig struct {
	payment.Config
	MerID              string //商户号
	PrivateKey         []byte //交易私钥
	PublicKey          []byte //交易公钥
	PrivateKeyPassword string //交易私钥密码
	ReturnURL          string //同步跳转地址
	NotifyURL          string //异步通知地址
	SignInvalidFields  string //忽略签名的字段名称集合按','分割默认:Signature,CertId
	SignatureField     string //签名的字段名称默认:Signature
}

//WithdrawConfig 提现配置信息
type WithdrawConfig struct {
	payment.Config
	TestMode           bool   //是否测试
	MerID              string //商户号
	PrivateKey         []byte //交易私钥
	PublicKey          []byte //交易公钥
	PrivateKeyPassword string //交易私钥密码
	SignInvalidFields  string //忽略签名的字段名称集合按','分割默认:chkValue
	SignatureField     string //签名的字段名称默认:chkValue
}
