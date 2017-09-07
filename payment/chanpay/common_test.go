package chanpay

import (
	"testing"

	"github.com/astaxie/beego/logs"
	"github.com/smartystreets/goconvey/convey"
)

func Test_SignAndVerify(t *testing.T) {
	SetLogger(logs.NewLogger())
	convey.Convey("畅捷支付", t, func() {
		PrivateKey := []byte(`-----BEGIN -----
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBANB5cQ5pf+QHF9Z2
+DjrAXstdxQHJDHyrni1PHijKVn5VHy/+ONiEUwSd5nx1d/W+mtYKxyc6HiN+5lg
WSB5DFimyYCiOInh3tGQtN+pN/AtE0dhMh4J9NXad0XEetLPRgmZ795O/sZZTnA3
yo54NBquT19ijYfrvi0JVf3BY9glAgMBAAECgYBFdSCox5eXlpFnn+2lsQ6mRoiV
AKgbiBp/FwsVum7NjleK1L8MqyDOMpzsinlSgaKfXxnGB7UgbVW1TTeErS/iQ06z
x3r4CNMDeIG1lYwiUUuguIDMedIJxzSNXfk65Bhps37lm129AE/VnIecpKxzelaU
uzyGEoFWYGevwc/lQQJBAPO0mGUxOR/0eDzqsf7ehE+Iq9tEr+aztPVacrLsEBAw
qOjUEYABvEasJiBVj4tECnbgGxXeZAwyQAJ5YmgseLUCQQDa/dgviW/4UMrY+cQn
zXVSZewISKg/bv+nW1rsbnk+NNwdVBxR09j7ifxg9DnQNk1Edardpu3z7ipHDTC+
z7exAkAM5llOue1JKLqYlt+3GvYr85MNNzSMZKTGe/QoTmCHStwV/uuyN+VMZF5c
RcskVwSqyDAG10+6aYqD1wMDep8lAkBQBoVS0cmOF5AY/CTXWrht1PsNB+gbzic0
dCjkz3YU6mIpgYwbxuu69/C3SWg7EyznQIyhFRhNlJH0hvhyMhvxAkEAuf7DNrgm
OJjRPcmAXfkbaZUf+F4iK+szpggOZ9XvKAhJ+JGd+3894Y/05uYYRhECmSlPv55C
BAPwd8VUsSb/1w==
-----END -----`)
		PublicKey := []byte(`-----BEGIN -----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDPq3oXX5aFeBQGf3Ag/86zNu0V
ICXmkof85r+DDL46w3vHcTnkEWVbp9DaDurcF7DMctzJngO0u9OG1cb4mn+Pn/uN
C1fp7S4JH4xtwST6jFgHtXcTG9uewWFYWKw/8b3zf4fXyRuI/2ekeLSstftqnMQd
enVP7XCxMuEnnmM1RwIDAQAB
-----END -----`)
		convey.Convey("测试签名", func() {
			params := map[string]string{
				"BankCode":       "ALIPAY",
				"GoodsName":      "交易测试",
				"InputCharset":   "utf-8",
				"MchId":          "200001220035",
				"OrderStartTime": "20170809170737",
				"OutTradeNo":     "123456789",
				"PartnerId":      "200001220035",
				"Service":        "mag_init_code_pay",
				"SpbillCreateIp": "127.0.0.1",
				"Subject":        "交易测试",
				"TradeAmount":    "0.20",
				"TradeDate":      "20170809",
				"TradeTime":      "170737",
				"TradeType":      "11",
				"Version":        "1.0",
			}
			sign(params, PrivateKey)
			convey.So(params["Sign"], convey.ShouldEqual, "IKiWKHjVfDWHygTng1KS+kXXOTY9/adrIMx6gq+atOAyGdp4nyEkLb5SPsOeS4VkGbUJK0hVjycK3u+fZP0YTW4y7P4NLKeH6qQh8feBgkoATNb8V4KfujMW9ud6c825ogk8m68tPmj1AaAPbIQp0ajXqMqlz/3LyOuus4yEA48=")
		})
		convey.Convey("验签", func() {
			params := map[string]string{
				"notify_id":      "448b7e8b93694e958eae50e295617033",
				"notify_type":    "trade_status_sync",
				"notify_time":    "20170705104121",
				"_input_charset": "UTF-8",
				"version":        "1.0",
				"outer_trade_no": "0001149922237380501132458",
				"inner_trade_no": "101149922237409129479",
				"trade_status":   "TRADE_SUCCESS",
				"trade_amount":   "0.01",
				"gmt_create":     "20170705104121",
				"gmt_payment":    "20170705104121",
				"extension":      "{}",
				"Sign":           "uERyn9W/b8I88bAVyaXUXXpyw0Ir5D3da6WiO5qrpDrvpgBmDzrYWt2wjZsu6CZdgxZ3+VSdRszrCKJM0UxUGqqKkf0gg90DFlGPMqloZHBzemXSoU2Zz/XYc7/CXWoi3+ZYk43dMhbh/S++RQFBOq+abkiGeD6QNlm4TUiJ7os=",
			}
			ret := verify(params, PublicKey)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

//订单号编码
func Test_codeNo(t *testing.T) {
	convey.Convey("订单号编码", t, func() {
		no := "00011499222373805011324582413623123578"
		str := encodeNo(no)
		convey.Printf("编码前订单号:%s\n长度:%d\n后的订单号:%s\n长度:%d", no, len(no), str, len(str))
		dno := decodeNo(str)
		convey.So(dno, convey.ShouldEqual, no)
	})
}
