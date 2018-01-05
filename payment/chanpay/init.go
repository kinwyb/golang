package chanpay

import (
	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/utils"
)

var lg utils.Logger

//SetLogger 设置日志
func SetLogger(log utils.Logger) {
	lg = log
}

//日志输出
func log(level utils.LoggerLevel, format string, args ...interface{}) {
	utils.WriteLog(lg, level, format, args...)
}

//DriverQrcode 畅捷扫码支付驱动
func DriverQrcode(fun payment.RegDriverFun, logger utils.Logger) {
	lg = logger
	err := fun(&qrcodePay{})
	if err != nil {
		log(utils.LogLevelError, "畅捷扫码支付驱动注入......[失败]:%s", err.Error())
	} else {
		log(utils.LogLevelInfo, "畅捷扫码支付驱动注入......[成功]")
	}
}

//DriverQuick 畅捷快捷支付驱动
func DriverQuick(fun payment.RegDriverFun, logger utils.Logger) {
	lg = logger
	err := fun(&quickPay{})
	if err != nil {
		log(utils.LogLevelError, "畅捷快捷支付驱动注入......[失败]:%s", err.Error())
	} else {
		log(utils.LogLevelInfo, "畅捷快捷支付驱动注入......[成功]")
	}
}

func DriverBank(fun payment.RegDriverFun, logger utils.Logger) {
	lg = logger
	err := fun(&bankPay{})
	if err != nil {
		log(utils.LogLevelError, "畅捷银行网关支付驱动注入......[失败]:%s", err.Error())
	} else {
		log(utils.LogLevelInfo, "畅捷银行网关支付驱动注入......[成功]")
	}
}

//WithdrawDriver 提现驱动
func WithdrawDriver(fun payment.RegWithdrawDriverFun, logger utils.Logger) {
	lg = logger
	err := fun(&chanpayWithdraw{})
	if err != nil {
		log(utils.LogLevelError, "畅捷提现驱动注入......[失败]:%s", err.Error())
	} else {
		log(utils.LogLevelInfo, "畅捷提现驱动注入......[成功]")
	}
}
