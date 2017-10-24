package chinapay

import (
	"encoding/base64"
	"testing"

	"io/ioutil"
	"os"

	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/utils"
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
		TestMode: true, //是否测试
		//MerID:    "808080211881045", //商户号
		MerID: "808080211305113",
	}
	//公钥
	//filepath := "/Users/heldiam/Developer/文档/银联/提现接口/测试/PgPubk.key"
	filepath := "/Users/heldiam/Developer/Documents/chinapay/提现接口/PgPubk.key"
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
	//filepath = "/Users/heldiam/Developer/文档/银联/提现接口/测试/808080211881045/MerPrK_808080211881045_20170607144054.key"
	filepath = "/Users/heldiam/Developer/Documents/chinapay/提现接口/808080211305113-116142/MerPrK.key"
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
	//验证签名
	responseData := "responseCode=0000&merId=808080211305113&merDate=20171010&merSeqId=129847382931&cpDate=20171010&cpSeqId=364902&transAmt=10&stat=s&cardNo=6222021208007333758&chkValue=06E90B2E6AA357CCFB5F38E0C81EF3C923B84115BC5C26257B965689FE9DB8314D47E7B66D2F97CF52288C4861EF76C54EE87206971E9B0FDD91190DC41B60399BA7AF2054039A22BB9D17FE708C0FA5E904DDD3C2A0F82886F6301D75E8D7DD3C5480EBA17F95712B145D1E1E92B5B1A70983C237AEF727D4890B94D14183D0"
	//responseData := "responseCode=0100&merId=&merDate=&merSeqId=&cpDate=&cpSeqId=&transAmt=&stat=&cardNo=&chkValue=82855B2C538973EA398585B462D310EC106534308A2AD518E8D561BC2F3D5251701D694794359562DF4F11A4A28E0DFC00E39A2EA770727A2261A1E5C3F2B14300C0884E52C62639EA4853CDB0F52501A05574D5CCBF480AD25DF2015FD03503272BF97DB065190F6088F11F2C1A6E53CD553B92F5010C7035CFEA6662015B91"
	idex := strings.LastIndex(responseData, "&")
	v := w.pubKey.Verify(base64.StdEncoding.EncodeToString([]byte(responseData[:idex])), responseData[idex+10:])
	log(utils.LogLevelDebug, "验证签名结果:%t", v)
	log(utils.LogLevelError, "银联结果签名=>[%s]\n等待签名base64结果:%s\n签名:%s",
		responseData[:idex], base64.StdEncoding.EncodeToString([]byte(responseData[:idex])), responseData[idex+10:])
	return
	/*
		withdrawInfo := &payment.WithdrawInfo{
			//TradeNo: "10" + time.Now().Format("060102150405"), //交易流水号
			TradeNo:  "2365417527855114",
			UserName: "王迎宾",                 //收款人姓名
			CardNo:   "6222021208007333758", //收款账户
			CertID:   "330782199110206617",  //收款人身份证号
			OpenBank: "工商银行",                //开户银行名称
			Prov:     "浙江省",                 //开户银行所在省份
			City:     "金华",                  //开户银行所在地区
			Money:    0.1,                   //提现金额
			Desc:     "提现测试",                //描述
			IP:       "127.0.0.1",           //提现的IP地址
			People:   true,                  //是个人，否企业
		}
		//成功
		res, err := w.Withdraw(withdrawInfo)
		if err != nil {
			t.Fatalf("提现失败=>%s", err)
		}
		if res.Status != payment.SUCCESS {
			t.Fatalf("提现结果异常=>%s", payment.StatusMsg(res.Status))
		}
		/*
			//失败
			withdrawInfo.TradeNo = "11" + time.Now().Format("060102150405")
			withdrawInfo.Money = 1
			res, err = w.Withdraw(withdrawInfo)
			if err != nil {
				t.Fatalf("提现失败=>%s", err)
				return
			}
			if res.Status != payment.FAIL {
				t.Fatalf("提现异常=>%s", payment.StatusMsg(res.Status))
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
				t.Fatalf("提现异常=>%s", payment.StatusMsg(res.Status))
				return
			}
	*/
}
