package alipay

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/astaxie/beego/logs"
	"github.com/kinwyb/golang/payment"
)

func TestWithdraw_Withdraw(t *testing.T) {
	SetLogger(logs.NewLogger())
	publicKeyFilePath := "/Users/heldiam/Desktop/public.key"
	privateKeyFilePath := "/Users/heldiam/Desktop/private.key"
	pubKey, err := os.Open(publicKeyFilePath)
	if err != nil {
		t.Fatalf("公钥文件打开失败:%s", err.Error())
	}
	privKey, err := os.Open(privateKeyFilePath)
	if err != nil {
		t.Fatalf("私钥文件打开失败:%s", err.Error())
	}
	pub, err := ioutil.ReadAll(pubKey)
	pubKey.Close()
	if err != nil {
		t.Fatalf("公钥文件读取失败:%s", err.Error())
	}
	pirv, err := ioutil.ReadAll(privKey)
	privKey.Close()
	if err != nil {
		t.Fatalf("私钥文件读取失败:%s", err.Error())
	}
	pcfg := &PayConfig{
		Config: payment.Config{
			Code:  "alipay",
			Name:  "支付宝提现",
			State: true,
		},
		Partner:    "2017031506224738",
		PrivateKey: string(pirv),
		PublicKey:  string(pub),
	}
	w := withdraw{}
	ww := w.GetWithdraw(pcfg)
	winfo := &payment.WithdrawInfo{
		TradeNo:  "29837491236",
		UserName: "王迎宾",
		CardNo:   "634246706@qq.com",
		CertID:   "330782199110206617",
		Money:    0.1,
		Desc:     "支付宝提现测试",
		People:   true,
	}
	result, err := ww.Withdraw(winfo)
	if err != nil {
		t.Fatalf("提现错误:%s", err.Error())
	} else if result.Status != payment.SUCCESS {
		t.Fatalf("提现失败")
	}
	t.Logf("提现成功")
	res := ww.QueryWithdraw(winfo.TradeNo)
	if res.Status != payment.SUCCESS {
		t.Fatalf("支付宝提现查询异常")
	}
	t.Logf("查询成功")
}
