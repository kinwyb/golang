package wxpay

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/astaxie/beego/logs"
	"github.com/kinwyb/golang/payment"
)

func Test_withdraw(t *testing.T) {
	lg = logs.NewLogger()
	cfg := &WithdrawConfig{
		Config: payment.Config{
			Code:  "wxwithdraw",
			Name:  "微信提现",
			State: true,
		},
		AppID:        "wx6273246ec533ad53",
		MchID:        "1481486922",
		Key:          "2dR6vljSqYKzKsK4Ge72gQOENLn4oHtI",
		CertPassword: "1481486922",
	}
	c, err := os.Open("/Users/heldiam/Desktop/apiclient_cert.p12")
	if err != nil {
		t.Errorf("证书打开失败:%s", err.Error())
		return
	}
	cfg.CertKey, err = ioutil.ReadAll(c)
	if err != nil {
		t.Error("证书读取失败:%s", err.Error())
		return
	}
	ww := &wxwithdraw{}
	w := ww.GetWithdraw(cfg)
	winfo := &payment.WithdrawInfo{
		TradeNo:  "1234567890",
		UserName: "王迎宾",
		CardNo:   "okBJH0Tnnq6BlUJL761XicI8B8F8", //收款账户
		Money:    0.1,
		Desc:     "提现测试",
		IP:       "127.0.0.1",
	}
	result, e := w.Withdraw(winfo)
	if e != nil {
		t.Errorf("提现失败:%s", e.Error())
		return
	}
	t.Logf("提现单号:%s", result.ThridFlowNo)
}
