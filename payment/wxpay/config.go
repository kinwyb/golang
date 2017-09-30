package wxpay

import "github.com/kinwyb/golang/payment"

//PayConfig 支付配置信息
type PayConfig struct {
	payment.Config
	AppID     string //微信应用ID
	MchID     string //微信商户ID
	Key       string //微信交易密钥
	NotifyURL string //交易结果通知地址
}

//WithdrawConfig 提现配置信息
type WithdrawConfig struct {
	payment.Config
	AppID        string //微信应用ID
	MchID        string //微信商户ID
	Key          string //微信交易密钥
	CertKey      []byte //提现密钥
	CertPassword string //提现密钥密码
}
