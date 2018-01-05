package chinapay

import (
	"bytes"
	"errors"

	"github.com/kinwyb/golang/payment"

	"github.com/kinwyb/golang/utils"

	"strconv"
	"time"
)

type chinapay struct {
	payment.PayInfo
	apiURL string
	config *PayConfig
	sess   *NetPaySecssUtil
}

//支付,返回支付代码
func (c *chinapay) Pay(req *payment.PayRequest) (string, error) {
	t := time.Now()
	params := map[string]string{
		"Version":    "20140728",
		"MerId":      c.config.MerID,
		"MerResv":    req.No,
		"TranDate":   t.Format("20060102"),
		"TranTime":   t.Format("150405"),
		"OrderAmt":   strconv.FormatInt(int64(req.Money*100), 10),
		"BusiType":   "0001",
		"MerPageUrl": c.config.ReturnURL,
		"MerBgUrl":   c.config.NotifyURL,
		"MerOrderNo": t.Format("150405") + req.No,
	}
	err := c.sign(params)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBufferString("<form id=\"chinapaysubmit\" name=\"chinapaysubmit\" action=\"")
	buf.WriteString(c.apiURL)
	buf.WriteString("\" method=\"POST\">")
	for k, v := range params {
		buf.WriteString("<input type=\"hidden\" name=\"")
		buf.WriteString(k)
		buf.WriteString("\" value=\"")
		buf.WriteString(v)
		buf.WriteString("\"/>")
	}
	buf.WriteString("<input type=\"submit\" value=\"提交\" style=\"display:none;\"></form>")
	buf.WriteString("<script>document.forms['chinapaysubmit'].submit();</script>")
	return buf.String(), nil
}

//异步结果通知处理,返回支付结果
func (c *chinapay) Notify(params map[string]string) *payment.PayResult {
	ret := &payment.PayResult{
		PayCode:      c.Code(),
		Navite:       params,
		TradeNo:      params["MerOrderNo"],
		No:           params["MerResv"],
		ThirdTradeNo: params["AcqSeqId"],
	}
	if !c.verify(params) {
		ret.Succ = false
		ret.ErrMsg = "签名验证失败"
		return ret
	}
	status := params["OrderStatus"]
	if status == "0000" {
		ret.Succ = true
		money, err := strconv.ParseFloat(params["OrderAmt"], 64)
		if err != nil {
			ret.Succ = false
			ret.ErrMsg = "交易金额异常"
		} else {
			ret.Money = money / 100
		}
	} else {
		ret.Succ = false
	}
	return ret
}

//异步通知处理结果返回内容
func (c *chinapay) NotifyResult(payResult *payment.PayResult) string {
	if payResult.Succ {
		return "success"
	}
	return "fail"
}

//同步结果跳转处理,返回支付结果
func (c *chinapay) Result(params map[string]string) *payment.PayResult {
	return c.Notify(params)
}

//GetPayment 生成一个支付对象
func (c *chinapay) GetPayment(cfg interface{}) payment.Payment {
	var conf *PayConfig
	ok := false
	if conf, ok = cfg.(*PayConfig); !ok || c == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的银联支付配置")
		return nil
	}
	if conf.SignInvalidFields == "" {
		conf.SignInvalidFields = "Signature,CertId"
	}
	if conf.SignatureField == "" {
		conf.SignatureField = "Signature"
	}
	obj := &chinapay{
		apiURL: "https://payment.chinapay.com/CTITS/service/rest/page/nref/000000000017/0/0/0/0/0",
		config: conf,
	}
	obj.sess = &NetPaySecssUtil{}
	err := obj.sess.Init(obj.config.PrivateKey, obj.config.PrivateKeyPassword, obj.config.PublicKey, obj.config.SignInvalidFields, obj.config.SignatureField)
	if err != nil {
		log(utils.LogLevelError, "密钥初始化失败:%s", err.Error())
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}

//Driver 驱动编码
func (c *chinapay) Driver() string {
	return "chinapay"
}

//交验签名
func (c *chinapay) verify(params map[string]string) bool {
	ret, err := c.sess.Verify(params)
	if err != nil {
		log(utils.LogLevelError, "签名验证失败:"+err.Error())
	}
	return ret
}

//签名
func (c *chinapay) sign(params map[string]string) error {
	signer, err := c.sess.Sign(params)
	if err != nil {
		return errors.New("签名生成失败:" + err.Error())
	}
	params[c.config.SignatureField] = signer
	return nil
}

//无需确认支付
func (c *chinapay) PayConfirm(req *payment.PayConfirmRequest) *payment.PayResult {
	return payment.NoPayConfirmResult
}
