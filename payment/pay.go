package payment

//NoPayConfirmResult 无需确认支付步骤结果提示
var NoPayConfirmResult = &PayResult{
	Succ:   false,
	ErrMsg: "该支付方式无需确认支付步骤",
}

//Payment 付款接口
type Payment interface {
	Pay(req *PayRequest) (string, error)          //支付,返回支付代码
	PayConfirm(req *PayConfirmRequest) *PayResult //确认支付.部分支付方式需要支付确认才能完成支付过程
	Notify(params map[string]string) *PayResult   //异步结果通知处理,返回支付结果
	NotifyResult(payResult *PayResult) string     //异步通知处理结果返回内容
	Result(params map[string]string) *PayResult   //同步结果跳转处理,返回支付结果
	Code() string                                 //返回支付编码
	Name() string                                 //返回支付方式名称
	Start() bool                                  //启用状态
}

//Driver 支付方式驱动接口
type Driver interface {
	Driver() string                 //获取驱动编码
	GetPayment(interface{}) Payment //生成一个支付对象
}

//Withdraw 提现接口
type Withdraw interface {
	withdraw(info *WithdrawInfo) (string, error) //提现操作,成功返回第三方交易流水,失败返回错误
	Code() string                                //返回提现编码
	Name() string                                //返回提现方式名称
	Start() bool                                 //启用状态
}
