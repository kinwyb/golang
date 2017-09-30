package wxpay

import (
	"crypto"
	"fmt"

	"net/http"
	"strings"

	"crypto/tls"

	"io/ioutil"

	"errors"

	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/utils"
	"golang.org/x/crypto/pkcs12"
)

//微信提现

type wxwithdraw struct {
	payment.PayInfo
	config    *WithdrawConfig
	transport *http.Transport
}

//获取驱动编码
func (w *wxwithdraw) Driver() string {
	return "wxwithdraw"
}

//生成一个提现对象
func (w *wxwithdraw) GetWithdraw(cfg interface{}) payment.Withdraw {
	var c *WithdrawConfig
	ok := false
	if c, ok = cfg.(*WithdrawConfig); !ok || c == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的微信支付配置")
		return nil
	}
	if c.Name == "" || c.Code == "" {
		return nil
	} else if c.CertPassword == "" { //证书密码就是商户号
		c.CertPassword = c.MchID
	}
	privkey, certificate, err := pkcs12.Decode(c.CertKey, c.CertPassword)
	if err != nil {
		log(utils.LogLevelError, "微信提现证书解析失败:%s", err.Error())
		return nil
	}
	cliCrt := tls.Certificate{
		PrivateKey: privkey.(crypto.PrivateKey),
	}
	cliCrt.Certificate = append(cliCrt.Certificate, certificate.Raw)
	obj := &wxwithdraw{
		config: c,
		transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{cliCrt},
			},
		},
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}

//提现操作,成功返回第三方交易流水,失败返回错误
func (w *wxwithdraw) Withdraw(info *payment.WithdrawInfo) (*payment.WithdrawResult, error) {
	params := map[string]string{
		"mch_appid":        w.config.AppID,
		"mchid":            w.config.MchID,
		"nonce_str":        nonceStr(),
		"partner_trade_no": info.TradeNo,
		"check_name":       "FORCE_CHECK",
		"openid":           info.CardNo,
		"re_user_name":     info.UserName,
		"amount":           fmt.Sprintf("%.0f", info.Money*100),
		"desc":             info.Desc,
		"spbill_create_ip": info.IP,
	}
	result, err := w.request(params, "https://api.mch.weixin.qq.com/mmpaymkttransfers/promotion/transfers")
	if err != nil {
		if err.Error() == "请求结果读取失败" { //结果读取失败的，确认一下是否交易完成
			return w.withdrawCheckResult(info)
		}
		return nil, err
	}
	if result["return_code"] == "SUCCESS" {
		if result["result_code"] == "SUCCESS" {
			return &payment.WithdrawResult{
				TradeNo:     info.TradeNo,
				ThridFlowNo: result["payment_no"],
				CardNo:      info.CardNo,
				CertID:      info.CertID,
				Money:       info.Money,
				Status:      payment.SUCCESS,
			}, nil
		} else if result["err_code"] == "SYSTEMERROR" { //请求结果提示业务繁忙的,调用查询接口确认一下业务是否真实失败
			return w.withdrawCheckResult(info)
		}
		log(utils.LogLevelError, "微信提现失败:%s", result["err_code_des"])
		return nil, errors.New(result["err_code_des"])
	}
	log(utils.LogLevelError, "微信提现失败:%s", result["return_msg"])
	return nil, errors.New(result["return_msg"])
}

//提现检测是否完成
func (w *wxwithdraw) withdrawCheckResult(info *payment.WithdrawInfo) (*payment.WithdrawResult, error) {
	res := w.QueryWithdraw(info.TradeNo)
	if res.Status == payment.SUCCESS {
		return &payment.WithdrawResult{
			TradeNo:     info.TradeNo,
			ThridFlowNo: res.ThridFlowNo,
			CardNo:      info.CardNo,
			CertID:      info.CertID,
			Money:       info.Money,
			Status:      payment.SUCCESS,
		}, nil
	} else if res.Status == payment.DEALING {
		return &payment.WithdrawResult{
			TradeNo: info.TradeNo,
			CardNo:  info.CardNo,
			CertID:  info.CertID,
			Money:   info.Money,
			Status:  payment.DEALING,
		}, nil
	}
	log(utils.LogLevelError, "微信提现失败[%s]:%s", res.FailCode, res.FailMsg)
	return nil, errors.New(res.FailMsg)
}

//请求
//@param params:map[string]string 请求参数
//@param apiURL:string 请求地址
func (w *wxwithdraw) request(params map[string]string, apiURL string) (map[string]string, error) {
	sign(params, w.config.Key)
	xmlstr := buildXML(params)
	log(utils.LogLevelInfo, "微信提现请求:%s", xmlstr)
	request, err := http.NewRequest("POST", apiURL, strings.NewReader(xmlstr.String()))
	if err != nil {
		log(utils.LogLevelError, "微信提现请求创建失败:%s", err.Error())
		return nil, fmt.Errorf("请求创建失败")
	}
	client := &http.Client{Transport: w.transport}
	response, err := client.Do(request)
	if err != nil {
		log(utils.LogLevelError, "微信提现请求失败:%s", err.Error())
		return nil, fmt.Errorf("请求失败")
	}
	responsedata, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log(utils.LogLevelError, "微信提现请求结果读取失败:%s", err.Error())
		return nil, fmt.Errorf("请求结果读取失败")
	}
	log(utils.LogLevelInfo, "微信提现结果:%s", responsedata)
	result, err := decodeXMLToMap(responsedata)
	if err != nil {
		log(utils.LogLevelError, "微信提现请求结果解析失败:%s", err.Error())
		return nil, fmt.Errorf("请求结果解析失败")
	}
	return result, nil
}

//查询提现交易
func (w *wxwithdraw) QueryWithdraw(tradeno string) *payment.WithdrawQueryResult {
	params := map[string]string{
		"mch_appid":        w.config.AppID,
		"mchid":            w.config.MchID,
		"nonce_str":        nonceStr(),
		"partner_trade_no": tradeno,
	}
	result, err := w.request(params, "https://api.mch.weixin.qq.com/mmpaymkttransfers/gettransferinfo")
	if err != nil {
		return &payment.WithdrawQueryResult{
			Status:  payment.DEALING,
			TradeNo: tradeno,
		}
	}
	ret := &payment.WithdrawQueryResult{
		Status:  payment.DEALING,
		TradeNo: tradeno,
	}
	if result["return_code"] == "SUCCESS" {
		if result["result_code"] == "SUCCESS" {
			if result["status"] == "SUCCESS" {
				ret.Status = payment.SUCCESS
				ret.PayTime = result["transfer_time"]
				ret.ThridFlowNo = result["detail_id"]
			} else if result["status"] == "FAILED" {
				ret.Status = payment.FAIL
				ret.FailMsg = result["reason"]
			}
		}
	}
	return ret
}
