package alipay

import (
	"time"

	"fmt"

	"encoding/json"

	"strings"

	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/utils"
)

//withdraw 支付宝提现
type withdraw struct {
	payment.PayInfo
	config    *PayConfig
	gateway   string
	verifyURL string
	signType  string
}

//提现操作,成功返回第三方交易流水,失败返回错误
func (w *withdraw) Withdraw(info *payment.WithdrawInfo) *payment.WithdrawResult {
	arg := &withdrawAPIRequest{
		OutBizNo: info.TradeNo,
		Type:     "ALIPAY_LOGONID",
		Account:  info.CardNo,
		RealName: info.UserName,
		Amount:   fmt.Sprintf("%.2f", info.Money),
		Remark:   info.Desc,
	}
	requestbytes, err := arg.MarshalJSON()
	if err != nil {
		return payment.WithdrawParamsSerializeFail
	}
	respdata, err := request("alipay.fund.trans.toaccount.transfer", w.config, string(requestbytes), w.gateway)
	if err != nil {
		return payment.WithdrawResponseReadFail
	}
	log(utils.LogLevelInfo, "支付宝提现结果:%s", respdata)
	vmap := &withdrawAPIResp{}
	err = json.Unmarshal(respdata, &vmap)
	if err != nil {
		log(utils.LogLevelError, "结果解析错误:%s", err.Error())
	}
	if vmap.Sign != "" && !w.verify(string(respdata), vmap.Sign, 49) {
		log(utils.LogLevelError, "支付宝提现请求结果签名验证异常")
		return &payment.WithdrawResult{
			TradeNo:      info.TradeNo,
			CardNo:       info.CardNo,
			CertID:       info.CertID,
			Money:        info.Money,
			UserName:     info.UserName,
			WithdrawCode: w.Code(),
			Status:       payment.DEALING,
		}
	}
	response := vmap.Method
	if response.Code == "10000" {
		return &payment.WithdrawResult{
			TradeNo:      response.OutBizNo,
			ThridFlowNo:  response.OrderID,
			CardNo:       info.CardNo,
			CertID:       info.CertID,
			Money:        info.Money,
			PayTime:      response.PayDate,
			UserName:     info.UserName,
			WithdrawCode: w.Code(),
			WithdrawName: w.Name(),
			Status:       payment.SUCCESS,
		}
	} else if response.SubCode == "SYSTEM_ERROR" { //请求结果提示业务繁忙的,调用查询接口确认一下业务是否真实失败
		qret := w.QueryWithdraw(info.TradeNo)
		if qret.Status == payment.FAIL { //提现失败
			return &payment.WithdrawResult{
				Status:   payment.FAIL,
				FailCode: qret.FailCode,
				FailMsg:  qret.FailMsg,
			}
		} else if qret.Status == payment.SUCCESS { //查询出来是成功的返回成功
			return &payment.WithdrawResult{
				TradeNo:      response.OutBizNo,
				ThridFlowNo:  qret.ThridFlowNo,
				CardNo:       info.CardNo,
				CertID:       info.CertID,
				Money:        info.Money,
				PayTime:      response.PayDate,
				UserName:     info.UserName,
				WithdrawCode: w.Code(),
				WithdrawName: w.Name(),
				Status:       payment.SUCCESS,
			}
		} else { //处理中的返回处理中
			return &payment.WithdrawResult{
				TradeNo:      response.OutBizNo,
				ThridFlowNo:  qret.ThridFlowNo,
				CardNo:       info.CardNo,
				CertID:       info.CertID,
				Money:        info.Money,
				PayTime:      response.PayDate,
				UserName:     info.UserName,
				WithdrawCode: w.Code(),
				WithdrawName: w.Name(),
				Status:       payment.DEALING,
			}
		}
	}
	return &payment.WithdrawResult{
		Status:   payment.FAIL,
		FailCode: response.SubCode,
		FailMsg:  response.SubMsg,
	}
}

//根据交易单号查询提现信息
func (w *withdraw) QueryWithdraw(tradeno string, tradeDate ...time.Time) *payment.WithdrawQueryResult {
	respdata, err := request("alipay.fund.trans.order.query", w.config,
		`{"out_biz_no":"`+tradeno+`"}`, w.gateway)
	if err != nil {
		return &payment.WithdrawQueryResult{
			Status:  payment.DEALING,
			TradeNo: tradeno,
		}
	}
	log(utils.LogLevelInfo, "支付宝提现查询结果:%s", respdata)
	vmap := &withdrawQueryAPIResp{}
	err = json.Unmarshal(respdata, &vmap)
	if err != nil {
		log(utils.LogLevelError, "支付宝提现查询结果解析错误:%s", err.Error())
	}
	if vmap.Sign != "" && !w.verify(string(respdata), vmap.Sign, 42) {
		log(utils.LogLevelError, "支付宝提现查询请求结果签名验证异常:%v", vmap)
		return &payment.WithdrawQueryResult{
			Status:  payment.DEALING,
			TradeNo: tradeno,
		}
	}
	response := vmap.Method
	ret := &payment.WithdrawQueryResult{
		Status:  payment.DEALING, //默认处理中
		TradeNo: tradeno,
	}
	if response.Code == "10000" { //业务请求成功
		log(utils.LogLevelError, "请求成功")
		switch response.Status {
		case "SUCCESS":
			ret.Status = payment.SUCCESS
			ret.PayTime = response.PayDate
			ret.ThridFlowNo = response.OrderID
		case "FAIL", "REFUND":
			ret.Status = payment.FAIL
			ret.FailCode = response.ErrorCode
			ret.FailMsg = response.FailReason
		}
	}
	return ret
}

//获取驱动编码
func (w *withdraw) Driver() string {
	return "alipay"
}

//提现验证签名
func (w *withdraw) verify(response string, signString string, start int) bool {
	response = response[start:strings.LastIndex(response, ",\"sign\":")]
	return verify(response, signString, w.config.PublicKey)
}

//生成一个提现对象
func (w *withdraw) GetWithdraw(cfg interface{}) payment.Withdraw {
	var c *PayConfig
	ok := false
	if c, ok = cfg.(*PayConfig); !ok || c == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的支付宝配置")
		return nil
	}
	if c.Name == "" || c.Code == "" {
		return nil
	}
	obj := &withdraw{
		gateway:   "https://openapi.alipay.com/gateway.do",
		verifyURL: "https://mapi.alipay.com/gateway.do?service=notify_verify&",
		signType:  "RSA2",
		config:    c,
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}
