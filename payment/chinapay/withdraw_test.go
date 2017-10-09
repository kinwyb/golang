package chinapay

import (
	"testing"

	"io/ioutil"
	"os"

	"time"

	"github.com/astaxie/beego/logs"
	"github.com/kinwyb/golang/payment"
)

func TestWithdraw_Withdraw(t *testing.T) {
	SetLogger(logs.NewLogger())
	driver := &withdraw{}
	cfg := &WithdrawConfig{
		Config: payment.Config{
			Code:  "chinapay", //支付编码
			Name:  "银联提现",     //支付名称
			State: true,       //是否启用
		},
		TestMode: true,              //是否测试
		MerID:    "808080211881045", //商户号
	}
	//公钥
	filepath := "/Users/heldiam/Developer/文档/银联/提现接口/测试/PgPubk.key"
	file, err := os.Open(filepath)
	if err != nil {
		t.Fatalf("公钥文件打开失败：%s", err)
		return
	}
	cfg.PublicKey, err = ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		t.Fatalf("公钥文件读取失败：%s", err)
		return
	}
	//私钥
	filepath = "/Users/heldiam/Developer/文档/银联/提现接口/测试/808080211881045/MerPrK_808080211881045_20170607144054.key"
	file, err = os.Open(filepath)
	if err != nil {
		t.Fatalf("私钥文件打开失败：%s", err)
		return
	}
	cfg.PrivateKey, err = ioutil.ReadAll(file)
	file.Close()
	if err != nil {
		t.Fatalf("私钥文件读取失败：%s", err)
		return
	}
	w := driver.GetWithdraw(cfg).(*withdraw)
	withdrawInfo := &payment.WithdrawInfo{
		//TradeNo: "10" + time.Now().Format("060102150405"), //交易流水号
		TradeNo:  "2365417527855114",
		UserName: "张三",                  //收款人姓名
		CardNo:   "6217856200025173163", //收款账户
		CertID:   "330782199901017788",  //收款人身份证号
		OpenBank: "中国银行",                //开户银行名称
		Prov:     "浙江省",                 //开户银行所在省份
		City:     "温州",                  //开户银行所在地区
		Money:    1,                     //提现金额
		Desc:     "银行卡提现",               //描述
		IP:       "127.0.0.1",           //提现的IP地址
		People:   true,                  //是个人，否企业
	}
	//成功
	res, err := w.Withdraw(withdrawInfo)
	if err != nil {
		t.Fatalf("提现失败=>%s", err)
	}
	if res.Status != payment.SUCCESS {
		t.Fatalf("提现结果异常=>%s", payment.WithdrawStatusMsg(res.Status))
	}
	//失败
	withdrawInfo.TradeNo = "11" + time.Now().Format("060102150405")
	withdrawInfo.Money = 1
	res, err = w.Withdraw(withdrawInfo)
	if err != nil {
		t.Fatalf("提现失败=>%s", err)
		return
	}
	if res.Status != payment.FAIL {
		t.Fatalf("提现异常=>%s", payment.WithdrawStatusMsg(res.Status))
		return
	}
	//等待中
	withdrawInfo.TradeNo = "12" + time.Now().Format("060102150405")
	withdrawInfo.Money = 1
	withdrawInfo.People = false
	res, err = w.Withdraw(withdrawInfo)
	if err != nil {
		t.Fatalf("提现失败=>%s", err)
		return
	}
	if res.Status != payment.DEALING {
		t.Fatalf("提现异常=>%s", payment.WithdrawStatusMsg(res.Status))
		return
	}
}
