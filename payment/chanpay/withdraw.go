package chanpay

import (
	"fmt"
	"time"

	"strings"

	"github.com/kinwyb/golang/payment"
	"github.com/kinwyb/golang/utils"
)

const timeFormat = "2006-01-02 15:04:05"

//提现接口

type chanpayWithdraw struct {
	payment.PayInfo
	config *WithdrawConfig
	apiURL string
}

//获取驱动编码
func (c *chanpayWithdraw) Driver() string {
	return "chanpay"
}

//生成一个提现对象
func (c *chanpayWithdraw) GetWithdraw(cfg interface{}) payment.Withdraw {
	var conf *WithdrawConfig
	ok := false
	if conf, ok = cfg.(*WithdrawConfig); !ok || conf == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的银联提现配置")
		return nil
	}
	if conf.Name == "" || conf.Code == "" {
		return nil
	}
	obj := &chanpayWithdraw{
		apiURL: "https://pay.chanpay.com/mag-unify/gateway/receiveOrder.do",
		config: conf,
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}

//提现操作,成功返回第三方交易流水,失败返回错误
func (c *chanpayWithdraw) Withdraw(info *payment.WithdrawInfo) *payment.WithdrawResult {
	info.TradeNo = encodeNo(info.TradeNo)
	t := time.Now()
	params := map[string]string{
		"Service":        "cjt_dsf", //同步单笔代付
		"Version":        "1.0",
		"PartnerId":      c.config.PartnerID,
		"InputCharset":   "utf-8",
		"TradeDate":      t.Format("20060102"),
		"TradeTime":      t.Format("150405"),
		"TransCode":      "T10000",
		"OutTradeNo":     info.TradeNo,
		"BankCommonName": info.OpenBank,                              //通用银行名称
		"AcctNo":         encrypt(c.config.PublicKey, info.CardNo),   //收款方银行卡或存折号码。使用平台公钥加密
		"AcctName":       encrypt(c.config.PublicKey, info.UserName), //收款方银行卡或存折上的所有人姓名。使用平台公钥加密
		"TradeAmount":    fmt.Sprintf("%.2f", info.Money),            //交易金额
		"LiceneceType":   "01",                                       //证件类型
		"LiceneceNo":     encrypt(c.config.PublicKey, info.CertID),   //证件号
	}
	if info.People { //业务类型 0=私人，1=公司
		params["BusinessType"] = "0"
	} else {
		params["BusinessType"] = "1"
	}
	result, err := request(c.apiURL, params, c.config.PrivateKey, c.config.PublicKey)
	if err != nil {
		return payment.WithdrawResponseReadFail
	}
	PlatformRetCode := result["PlatformRetCode"] //平台受理码
	OriginalRetCode := result["OriginalRetCode"] //原交易返回代码
	AppRetcode := result["AppRetcode"]           //应用返回码
	switch PlatformRetCode {
	case "1000", "2004", "2009":
		//交易失败
		return &payment.WithdrawResult{
			TradeNo:      info.TradeNo,                  //交易流水号
			ThridFlowNo:  result["FlowNo"],              //第三方交易流水号
			CardNo:       info.CardNo,                   //收款账户
			CertID:       info.CertID,                   //收款人身份证号
			Money:        info.Money,                    //提现金额
			PayTime:      time.Now().Format(timeFormat), //完成时间
			UserName:     info.UserName,
			WithdrawCode: c.Code(),
			WithdrawName: c.Name(),
			Status:       payment.FAIL, //提现状态
			FailCode:     "Pf:" + PlatformRetCode,
			FailMsg:      "畅捷平台未受理",
		}
	}
	if OriginalRetCode == "111111" {
		//交易失败
		return &payment.WithdrawResult{
			TradeNo:      info.TradeNo,                  //交易流水号
			ThridFlowNo:  result["FlowNo"],              //第三方交易流水号
			CardNo:       info.CardNo,                   //收款账户
			CertID:       info.CertID,                   //收款人身份证号
			Money:        info.Money,                    //提现金额
			PayTime:      time.Now().Format(timeFormat), //完成时间
			UserName:     info.UserName,
			WithdrawCode: c.Code(),
			WithdrawName: c.Name(),
			Status:       payment.FAIL, //提现状态
			FailCode:     "Org:" + OriginalRetCode,
			FailMsg:      "原交易返回代码失败",
		}
	}
	if AppRetcode == "000000" || AppRetcode == "00019999" {
		return &payment.WithdrawResult{
			TradeNo:      info.TradeNo,                  //交易流水号
			ThridFlowNo:  result["FlowNo"],              //第三方交易流水号
			CardNo:       info.CardNo,                   //收款账户
			CertID:       info.CertID,                   //收款人身份证号
			Money:        info.Money,                    //提现金额
			PayTime:      time.Now().Format(timeFormat), //完成时间
			UserName:     info.UserName,
			WithdrawCode: c.Code(),
			WithdrawName: c.Name(),
			Status:       payment.SUCCESS, //提现状态
		}
	} else if AppRetcode == "111111" || AppRetcode == "11019999" {
		return &payment.WithdrawResult{
			TradeNo:      info.TradeNo,                  //交易流水号
			ThridFlowNo:  result["FlowNo"],              //第三方交易流水号
			CardNo:       info.CardNo,                   //收款账户
			CertID:       info.CertID,                   //收款人身份证号
			Money:        info.Money,                    //提现金额
			PayTime:      time.Now().Format(timeFormat), //完成时间
			UserName:     info.UserName,
			WithdrawCode: c.Code(),
			WithdrawName: c.Name(),
			Status:       payment.FAIL, //提现状态
			FailCode:     "AP:" + AppRetcode,
			FailMsg:      "应用返回码失败",
		}
	}
	//交易处理中
	return &payment.WithdrawResult{
		TradeNo:      info.TradeNo,                  //交易流水号
		ThridFlowNo:  result["FlowNo"],              //第三方交易流水号
		CardNo:       info.CardNo,                   //收款账户
		CertID:       info.CertID,                   //收款人身份证号
		Money:        info.Money,                    //提现金额
		PayTime:      time.Now().Format(timeFormat), //完成时间
		UserName:     info.UserName,
		WithdrawCode: c.Code(),
		WithdrawName: c.Name(),
		Status:       payment.DEALING, //提现状态
	}
}

//查询提现交易
func (c *chanpayWithdraw) QueryWithdraw(tradeno string, tradeDate ...time.Time) *payment.WithdrawQueryResult {
	t := time.Now()
	params := map[string]string{
		"Service":       "cjt_dsf", //同步单笔代付
		"Version":       "1.0",
		"PartnerId":     c.config.PartnerID,
		"InputCharset":  "utf-8",
		"TradeDate":     t.Format("20060102"),
		"TradeTime":     t.Format("150405"),
		"TransCode":     "C00000",
		"OriOutTradeNo": tradeno,
		"OutTradeNo":    strings.Replace(t.Format("20060102150405.99999"), ".", "", -1),
	}
	returnDealign := &payment.WithdrawQueryResult{
		Status:  payment.DEALING, //提现状态
		TradeNo: tradeno,         //交易流水号
	}
	result, err := request(c.apiURL, params, c.config.PrivateKey, c.config.PublicKey)
	if err != nil {
		return returnDealign
	}
	PlatformRetCode := result["PlatformRetCode"] //平台受理码
	OriginalRetCode := result["OriginalRetCode"] //原交易返回代码
	AppRetcode := result["AppRetcode"]           //应用返回码
	switch PlatformRetCode {
	case "1000", "2004", "2009":
		//交易失败
		return &payment.WithdrawQueryResult{
			TradeNo:     tradeno,                       //交易流水号
			ThridFlowNo: result["FlowNo"],              //第三方交易流水号
			PayTime:     time.Now().Format(timeFormat), //完成时间
			Status:      payment.FAIL,                  //提现状态
			FailCode:    "Pf:" + PlatformRetCode,
			FailMsg:     "畅捷平台未受理",
		}
	}
	if OriginalRetCode == "111111" {
		//交易失败
		return &payment.WithdrawQueryResult{
			TradeNo:     tradeno,                       //交易流水号
			ThridFlowNo: result["FlowNo"],              //第三方交易流水号
			PayTime:     time.Now().Format(timeFormat), //完成时间
			Status:      payment.FAIL,                  //提现状态
			FailCode:    "Org:" + OriginalRetCode,
			FailMsg:     "原交易返回代码失败",
		}
	}
	if AppRetcode == "000000" || AppRetcode == "00019999" {
		return &payment.WithdrawQueryResult{
			TradeNo:     tradeno,                       //交易流水号
			ThridFlowNo: result["FlowNo"],              //第三方交易流水号
			PayTime:     time.Now().Format(timeFormat), //完成时间
			Status:      payment.SUCCESS,               //提现状态
		}
	} else if AppRetcode == "111111" || AppRetcode == "11019999" {
		return &payment.WithdrawQueryResult{
			TradeNo:     tradeno,                       //交易流水号
			ThridFlowNo: result["FlowNo"],              //第三方交易流水号
			PayTime:     time.Now().Format(timeFormat), //完成时间
			Status:      payment.FAIL,                  //提现状态
			FailCode:    "AP:" + AppRetcode,
			FailMsg:     "应用返回码失败",
		}
	}
	//交易处理中
	return returnDealign
}
