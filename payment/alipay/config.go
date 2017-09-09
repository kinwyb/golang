package alipay

import "github.com/kinwyb/golang/payment"

//PayConfig 支付配置信息
type PayConfig struct {
	payment.Config
	Partner    string //商户号
	PrivateKey string //交易私钥
	PublicKey  string //交易公钥
	ReturnURL  string //同步跳转地址
	NotifyURL  string //异步跳转地址
}
