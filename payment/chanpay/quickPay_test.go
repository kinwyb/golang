package chanpay

import (
	"encoding/base64"
	"testing"

	"github.com/kinwyb/golang/crypto/rsautil"
	"github.com/kinwyb/golang/payment"
	"github.com/astaxie/beego/logs"
	"github.com/smartystreets/goconvey/convey"
)

func Test_Encode(t *testing.T) {
	publicKey := []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDEPW04vWG61OUt0QjtlrPZGgWE
ZQQEU+x9qpajBeVT3GLfDETsm1SyQ1hnt7Na9FCLcX1EhU14w5Z+XtLrEq+KkJBU
VKtZYk6nseEb3PBBT4nxcjFeN/wPw4qpYKvMSj8JiI/HDwC5Baa99Pvj9iUGIqh+
7dbIsL0g7DH5DizxoQIDAQAB
-----END PUBLIC KEY-----`)
	privateKey := []byte(`-----BEGIN PRIVATE KEY-----
MIICcwIBADALBgkqhkiG9w0BAQEEggJfMIICWwIBAAKBgQDEPW04vWG61OUt0Qjt
lrPZGgWEZQQEU+x9qpajBeVT3GLfDETsm1SyQ1hnt7Na9FCLcX1EhU14w5Z+XtLr
Eq+KkJBUVKtZYk6nseEb3PBBT4nxcjFeN/wPw4qpYKvMSj8JiI/HDwC5Baa99Pvj
9iUGIqh+7dbIsL0g7DH5DizxoQIDAQABAoGAEb0LtmlIAD9mR/HxQKiysRktDn6j
ElETu3hEDZBm3mG5fjf5svmHemWkBBwS1lHnRfOIQz1Zd2UWoW2o2x7hRxiok3Rm
dZUnEKFBHrb0bPgQtoOyCx7fLOMmbiWFvgazu0Jil7spGtzrCcVjmX6CfLyxWvht
HIPWOy8JyXl00wECQQDZykTs/+3qbnuJWA8Gl7Wt7frIzdXkE3pVgKgKLjuRTbVA
ijRIkl5LQeIpPeI/bN2LUd9xGvDY3X57iJ1hOxuJAkEA5qtDODlsAOwyBYrz8Vk9
eqTXLyfGhDWD3vGTk4N2QZiRugbdX4unCwkg630Pn7r08U5XKAfb82lxd4iqh9qn
WQJAOXNqCzrX/+d1Hx3jmNGcU21bomzp52hb9QIjUcwwWnwtPAE5GYvC5AdVKZvx
etm093N5hdSdhBeprdyz51o4QQJAAriAaXhb6sLecCxMZktcK0codpjsgYC0FnwY
9oN1cJ6hEWWlVMwr4zhvV/e4qHSnEPWQl5tIH93dhcBp6oJMuQJAa3W66V3ThMCC
enfezUu4+VRzUdv+gkUCABCnvGz9Ep8RCeCEUuLMmHD+KhCPMgtwvQLv5hP2Qu+C
2wcDIWRtMA==
-----END PRIVATE KEY-----`)
	convey.Convey("RSA加解密", t, func() {
		waitEncryptData := "等待加密的内容,waitecode"
		data, err := rsautil.Encrypt(publicKey, []byte(waitEncryptData))
		if err != nil {
			t.Errorf("加密失败:%s", err.Error())
			return
		}
		convey.Println(base64.StdEncoding.EncodeToString(data))
		decodeData, err := rsautil.Decrypt(privateKey, data, rsautil.PKCS8)
		if err != nil {
			t.Errorf("解密失败:%s", err.Error())
			return
		}
		d := "H3XKhYacScCnM5XmXSQFenrMB2u39Rx7iqNExtIoW99rB7av6K7GFk+Nkudroh5ANQFM+I9YpRVvEzpsfxIk8QnETsCtQCu41sNOWA0zNpb6ND8NVgCT0YYVhDo7mKi9ss/0/BtbUehdDhahvyo6EQ9NkDPG3NXbAe0Wc0vI11M="
		data, err = base64.StdEncoding.DecodeString(d)
		if err != nil {
			t.Errorf("BASE64解码失败:%s", err.Error())
			return
		}
		decodeData, err = rsautil.Decrypt(privateKey, data, rsautil.PKCS8)
		if err != nil {
			t.Errorf("解密失败2:%s", err.Error())
			return
		}
		convey.So(string(decodeData), convey.ShouldEqual, waitEncryptData)
	})
}

func Test_quickPay(t *testing.T) {
	SetLogger(logs.NewLogger())
	convey.Convey("畅捷快捷支付", t, func() {
		quickPayConfig := &QuickPayConfig{
			PartnerID: "200001160097",
			MchID:     "200001160097",
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
			ExpiredTime: "40m",
		}
		q := &quickPay{
			config: quickPayConfig,
			apiURL: "https://pay.chanpay.com/mag-unify/gateway/receiveOrder.do",
		}
		convey.Convey("支付测试", func() {
			req := &payment.PayRequest{
				No:       "1503563106960768911", //fmt.Sprintf("%d", time.Now().UnixNano()),
				Desc:     "交易测试",
				Money:    0.1,
				IP:       "127.0.0.1",
				MemberID: "9234mf8al2439fj90wrk234190",
				Ext: &QuickPayRequestExt{
					BkAcctNo: "6222021208007333758",
					IDNo:     "330782199110206617",
					CstmrNm:  "王迎宾",
					MobNo:    "15058679668",
				},
			}
			result, err := q.Pay(req)
			if err != nil {
				t.Error(err.Error())
				return
			}
			convey.So(result, convey.ShouldNotBeEmpty)
		})
		convey.Convey("结果通知测试", func() {
			params := map[string]string{
				"notify_id":      "87e8e322cc9b4dba8fc2f0e8010a1382",
				"notify_type":    "trade_status_sync",
				"notify_time":    "20170807233141",
				"_input_charset": "UTF-8",
				"version":        "1.0",
				"outer_trade_no": "17080720052853968889",
				"inner_trade_no": "101150210752881582798",
				"trade_status":   "TRADE_SUCCESS",
				"trade_amount":   "0.01",
				"gmt_create":     "20170807200534",
				"gmt_payment":    "20170807200534",
				"extension":      "{\"BANK_RET_DATA\":\"{'bank_type':'CFT','fee_type':'CNY','is_subscribe':'N','openid':'oMJGHs2wAz41X5GjYp4bPbcuB-EU','out_trade_no':'SG102946308070000040358','out_transaction_id':'4001502001201708075023175339','pay_result':'0','result_code':'0','status':'0','sub_appid':'wxfa2f613ed691411f','sub_is_subscribe':'Y','sub_openid':'os7Olwggpu_x6urLCdMh6uJseiUI','time_end':'20170807200533','total_fee':'1','transaction_id':'299540006994201708072263952157'}\"}",
				"sign_type":      "RSA",
				"sign":           "MBz+rB4byKsyN/BziK15rQNY0fWDQcMHiOpNqEh7+Boah3GowrvkYjy6hN8zLKQIzwtRoB3Bgpgru8epWaPipL46ci6GwqMid4bYs00QSXrHJLhOimf5jAWyj9v2baqHVfnwJ+4ySXxzAs0oFNe+Dy0q02NAhO6i3qocM5VTT3A=",
			}
			ret := q.Notify(params)
			result := q.NotifyResult(ret)
			if !ret.Succ {
				t.Errorf("结果处理错误:%s", ret.ErrMsg)
			}
			convey.So(result, convey.ShouldEqual, "success")
		})
	})
}
