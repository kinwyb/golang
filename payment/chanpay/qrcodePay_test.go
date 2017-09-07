package chanpay

import (
	"fmt"
	"testing"
	"time"

	"git.oschina.net/kinwyb/golang/payment"
	"github.com/astaxie/beego/logs"
	"github.com/smartystreets/goconvey/convey"
)

func Test_QrcodePay(t *testing.T) {
	SetLogger(logs.NewLogger())
	convey.Convey("畅捷扫码支付", t, func() {
		q := &qrcodePay{
			config: &QRPayConfig{
				PartnerID: "200001220035",
				MchID:     "200001220035",
				PrivateKey: []byte(`-----BEGIN -----
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
-----END -----`),
				PublicKey: []byte(`-----BEGIN -----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDPq3oXX5aFeBQGf3Ag/86zNu0V
ICXmkof85r+DDL46w3vHcTnkEWVbp9DaDurcF7DMctzJngO0u9OG1cb4mn+Pn/uN
C1fp7S4JH4xtwST6jFgHtXcTG9uewWFYWKw/8b3zf4fXyRuI/2ekeLSstftqnMQd
enVP7XCxMuEnnmM1RwIDAQAB
-----END -----`),
			},
			apiURL: "https://pay.chanpay.com/mag-unify/gateway/receiveOrder.do",
		}
		convey.Convey("通知结果处理", func() {
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
				"sign_type":      "RSA",
				"sign":           "uERyn9W/b8I88bAVyaXUXXpyw0Ir5D3da6WiO5qrpDrvpgBmDzrYWt2wjZsu6CZdgxZ3+VSdRszrCKJM0UxUGqqKkf0gg90DFlGPMqloZHBzemXSoU2Zz/XYc7/CXWoi3+ZYk43dMhbh/S++RQFBOq+abkiGeD6QNlm4TUiJ7os=",
			}
			ret := q.Notify(params)
			result := q.NotifyResult(ret)
			if !ret.Succ {
				t.Errorf("结果处理错误:%s", ret.ErrMsg)
			}
			convey.So(result, convey.ShouldEqual, "success")
		})
		convey.Convey("测试交易", func() {
			req := &payment.PayRequest{
				No:    fmt.Sprintf("%d", time.Now().UnixNano()),
				Money: 0.2,
				Desc:  "交易测试",
				IP:    "127.0.0.1",
			}
			result, err := q.Pay(req)
			if err != nil {
				t.Error(err.Error())
				return
			}
			convey.So(result, convey.ShouldNotBeEmpty)
		})
	})
}
