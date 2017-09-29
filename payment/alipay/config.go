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

//withdrawAPIResponse 提现接口返回结果对象
type withdrawAPIResponse struct {
	Code     string `json:"code"`       //网关返回码
	Msg      string `json:"msg"`        //网关返回码描述
	SubCode  string `json:"sub_code"`   //业务返回码
	SubMsg   string `json:"sub_msg"`    //业务返回码描述
	Sign     string `json:"sign"`       //签名
	OutBizNo string `json:"out_biz_no"` //商户转账唯一单号
	OrderID  string `json:"order_id"`   //支付宝转账单据
	PayDate  string `json:"pay_date"`   //支付时间
}

//withdrawQueryAPIResponse 提现查询接口返回结果对象
type withdrawQueryAPIResponse struct {
	Code       string `json:"code"`        //网关返回码
	Msg        string `json:"msg"`         //网关返回码描述
	SubCode    string `json:"sub_code"`    //业务返回码
	SubMsg     string `json:"sub_msg"`     //业务返回码描述
	Sign       string `json:"sign"`        //签名
	OrderID    string `json:"order_id"`    //支付宝转账单据
	PayDate    string `json:"pay_date"`    //支付时间
	Status     string `json:"status"`      //转账单据状态
	OutBizNo   string `json:"out_biz_no"`  //商户转账唯一单号
	FailReason string `json:"fail_reason"` //失败原因
	ErrorCode  string `json:"error_code"`  //错误代码
}

//withdrawAPIRequest 提现接口请求参数
type withdrawAPIRequest struct {
	OutBizNo string `json:"out_biz_no"`      //商户转账唯一订单号
	Type     string `json:"payee_type"`      //收款方账户类型
	Account  string `json:"payee_account"`   //收款方账户
	RealName string `json:"payee_real_name"` //收款方真实姓名
	Amount   string `json:"amount"`          //转账金额
	Remark   string `json:"remark"`          //转账备注
}
