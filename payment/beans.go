package payment

//RegDriverFun 驱动注入函数
type RegDriverFun func(Driver) error

//RegWithdrawDriverFun 提现驱动注入函数
type RegWithdrawDriverFun func(WithdrawDriver) error

type Status string

const (
	SUCCESS Status = "SUCCESS" //成功
	FAIL    Status = "FAIL"    //失败
	DEALING Status = "DEALING" //处理中
	UNKNOW  Status = "UNKNOW"  //未知
)

//交易状态说明
func StatusMsg(status Status) string {
	switch status {
	case SUCCESS:
		return "成功"
	case FAIL:
		return "失败"
	case DEALING:
		return "处理中"
	}
	return "未知"
}

//WithdrawInfo 提现基本信息
type WithdrawInfo struct {
	TradeNo  string  `description:"交易流水号"`
	UserName string  `description:"收款人姓名"`
	CardNo   string  `description:"收款账户"`
	CertID   string  `description:"收款人身份证号"`
	OpenBank string  `description:"开户银行名称"`
	Prov     string  `description:"开户银行所在省份"`
	City     string  `description:"开户银行所在地区"`
	Money    float64 `description:"提现金额"`
	Desc     string  `description:"描述"`
	IP       string  `description:"提现的IP地址"`
	People   bool    `description:"是个人，否企业"`
}

//提现结果
type WithdrawResult struct {
	AppID        int     `description:"发起提现的应用编码"`
	WithdrawCode string  `description:"提现方式编码"`
	WithdrawName string  `description:"提现方式名称"`
	TradeNo      string  `description:"交易流水号"`
	ThridFlowNo  string  `description:"第三方交易流水号"`
	CardNo       string  `description:"收款账户"`
	UserName     string  `description:"收款人姓名"`
	CertID       string  `description:"收款人身份证号"`
	Money        float64 `description:"提现金额"`
	PayTime      string  `description:"完成时间"`
	Status       Status  `description:"提现状态"`
	FailCode     string  `description:"错误编码"`
	FailMsg      string  `description:"错误消息"`
}

//提现查询结果
type WithdrawQueryResult struct {
	Status      Status //提现状态
	PayTime     string //完成时间
	TradeNo     string //交易流水号
	ThridFlowNo string //第三方交易流水号
	FailCode    string //错误代码
	FailMsg     string //错误原因
}

//PayRequest 支付请求
type PayRequest struct {
	No       string  `description:"交易单号"`
	Desc     string  `description:"交易描述"`
	Money    float64 `description:"交易金额"`
	IsApp    bool    `description:"是否是APP支付"`
	PayCode  string  `description:"支付方式"`
	IP       string  `description:"交易发起端IP"`
	MemberID string  `description:"商户网站用户唯一标识[部分支付方式必填]"`
	Ext      string  `description:"支付方式扩展内容[部分支付方式需填写,json字符串]"`
}

//PayConfirmRequest 支付确认请求参数
type PayConfirmRequest struct {
	No         string `description:"交易单号"`
	ThirdNo    string `description:"第三方交易流水号"`
	VerifyCode string `description:"验证信息"`
}

//PayResult 支付结果
type PayResult struct {
	Succ         bool              //是否成功
	ErrMsg       string            //错误消息
	No           string            //订单号
	TradeNo      string            //交易单号
	Money        float64           //交易金额
	PayCode      string            //交易方式编码
	ThirdAccount string            //第三方交易帐号
	ThirdTradeNo string            //第三方交易流水号
	Navite       map[string]string //原始数据
}

//PayInfo 支付方式基础信息
type PayInfo struct {
	code  string
	name  string
	start bool
	drive string
}

//Init 初始化基本信息
func (p *PayInfo) Init(code, name string, start bool) {
	p.code = code
	p.name = name
	p.start = start
}

//Code 支付方式编码
func (p *PayInfo) Code() string {
	return p.code
}

//Name 支付方式名称
func (p *PayInfo) Name() string {
	return p.name
}

//Start 支付方式启用状态
func (p *PayInfo) Start() bool {
	return p.start
}

//Config 支付方式配置基础字段
type Config struct {
	Code  string //支付编码
	Name  string //支付名称
	State bool   //是否启用
}
