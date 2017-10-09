package chinapay

import (
	"github.com/kinwyb/golang/payment"

	"github.com/kinwyb/golang/utils"
)

var lg utils.Logger

//Driver 银联支付驱动
func Driver(fun payment.RegDriverFun, logger utils.Logger) {
	lg = logger
	err := fun(&chinapay{})
	if err != nil {
		log(utils.LogLevelError, "银联支付驱动注入......[失败]:%s", err.Error())
	} else {
		log(utils.LogLevelInfo, "银联支付驱动注入......[成功]")
	}
}

//WithdrawDriver 银联提现驱动
func WithdrawDriver(fun payment.RegWithdrawDriverFun, logger utils.Logger) {
	lg = logger
	err := fun(&withdraw{})
	if err != nil {
		log(utils.LogLevelError, "银联提现驱动注入......[失败]:%s", err.Error())
	} else {
		log(utils.LogLevelInfo, "银联提现驱动注入......[成功]")
	}
}

//SetLogger 设置日志
func SetLogger(log utils.Logger) {
	lg = log
}

//日志输出
func log(level utils.LoggerLevel, format string, args ...interface{}) {
	utils.WriteLog(lg, level, format, args...)
}
