package alipay

import (
	"github.com/kinwyb/golang/payment"

	"github.com/kinwyb/golang/utils"
)

var lg utils.Logger

//Driver 支付宝驱动
func Driver(fun payment.RegDriverFun, logger utils.Logger) {
	lg = logger
	err := fun(&alipay{})
	if err != nil {
		log(utils.LogLevelError, "支付宝驱动注入......[失败]:%s", err.Error())
	} else {
		log(utils.LogLevelInfo, "支付宝驱动注入......[成功]")
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
