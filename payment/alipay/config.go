package alipay

import (
	"bytes"

	"fmt"

	"github.com/kinwyb/golang/payment"
)

//PayConfig 支付配置信息
type PayConfig struct {
	payment.Config
	Partner    string //商户号
	PrivateKey string //交易私钥
	PublicKey  string //交易公钥
	ReturnURL  string //同步跳转地址
	NotifyURL  string //异步跳转地址
}

type withdrawAPIResp struct {
	Method *withdrawAPIResponse `json:"alipay_fund_trans_toaccount_transfer_response"`
	Sign   string               `json:"sign"`
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
type withdrawQueryAPIResp struct {
	Method *withdrawQueryAPIResponse `json:"alipay_fund_trans_order_query_response"`
	Sign   string                    `json:"sign"`
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
	Amount   string `json:"amount"`          //转账金额
	RealName string `json:"payee_real_name"` //收款方真实姓名
	Remark   string `json:"remark"`          //转账备注
}

//app支付返回结果
type appPayReturn struct {
	Result []byte `json:"result"`       //处理结果
	Status string `json:"resultStatus"` //结果码
	Memo   string `json:"memo"`         //描述信息
}

//app支付结果
type appPayResult struct {
	Response *appPayResponse `json:"alipay_trade_app_pay_response"`
	Sign     string          `json:"sign"`
	SignType string          `json:"sign_type"`
}

type appPayResponse struct {
	Code        string  `json:"code"`         //结果码
	Msg         string  `json:"msg"`          //处理结果的描述
	AppID       string  `json:"app_id"`       //支付宝分配给开发者的应用Id
	OutTradeNo  string  `json:"out_trade_no"` //商户网站唯一订单号
	TradeNo     string  `json:"trade_no"`     //该交易在支付宝系统中的交易流水号
	TotalAmount float64 `json:"total_amount"` //该笔订单的资金总额
	SellerID    string  `json:"seller_id"`    //收款支付宝账号对应的支付宝唯一用户号
	Charset     string  `json:"charset"`      //编码格式
	Timestamp   string  `json:"timestamp"`    //时间
}

func (a *appPayResponse) Xml() string {
	buf := bytes.NewBufferString("<xml>")
	buf.WriteString("<code>")
	buf.WriteString(a.Code)
	buf.WriteString("</code>")
	buf.WriteString("<msg>")
	buf.WriteString(a.Msg)
	buf.WriteString("</msg>")
	buf.WriteString("<app_id>")
	buf.WriteString(a.AppID)
	buf.WriteString("</app_id>")
	buf.WriteString("<out_trade_no>")
	buf.WriteString(a.OutTradeNo)
	buf.WriteString("</out_trade_no>")
	buf.WriteString("<trade_no>")
	buf.WriteString(a.TradeNo)
	buf.WriteString("</trade_no>")
	buf.WriteString("<total_amount>")
	buf.WriteString(fmt.Sprintf("%.2f", a.TotalAmount))
	buf.WriteString("</total_amount>")
	buf.WriteString("<seller_id>")
	buf.WriteString(a.SellerID)
	buf.WriteString("</seller_id>")
	buf.WriteString("<charset>")
	buf.WriteString(a.Charset)
	buf.WriteString("</charset>")
	buf.WriteString("<timestamp>")
	buf.WriteString(a.Timestamp)
	buf.WriteString("</timestamp>")
	buf.WriteString("</xml>")
	return buf.String()
}
