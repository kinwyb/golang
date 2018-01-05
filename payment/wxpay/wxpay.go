package wxpay

import (
	"errors"
	"time"

	"github.com/kinwyb/golang/payment"

	"github.com/kinwyb/golang/utils"

	"net/http"

	"io/ioutil"

	"strconv"
)

type wxpay struct {
	payment.PayInfo
	config *PayConfig
	apiURL string
}

//支付,返回支付代码
func (w *wxpay) Pay(req *payment.PayRequest) (string, error) {
	params := map[string]string{
		"appid":        w.config.AppID,                              //微信分配的公众账号ID
		"mch_id":       w.config.MchID,                              //微信支付分配的商户号
		"nonce_str":    nonceStr(),                                  //随机字符串
		"body":         req.Desc,                                    //商品名称
		"attach":       req.No,                                      //由于统一订单号无法重复发起支付所以订单号只能存放在附加字段,交易单号重新生成
		"total_fee":    strconv.FormatInt(int64(req.Money*100), 10), //交易金额,单位分
		"notify_url":   w.config.NotifyURL,
		"trade_type":   "NATIVE",
		"product_id":   "0",
		"out_trade_no": time.Now().Format("150405") + req.No,
	}
	sign(params, w.config.Key)
	xmlBuf := buildXML(params)
	resp, err := http.Post(w.apiURL, "application/xml;charset=utf-8", xmlBuf)
	if err != nil {
		return "", errors.New("微信请求失败:" + err.Error())
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", errors.New("微信请求失败:" + err.Error())
	} else if req.IsApp {
		return string(data), nil //app支付无需处理直接返回结果
	}
	return w.decodeResp(data)
}

//异步结果通知处理,返回支付结果
func (w *wxpay) Notify(params map[string]string) *payment.PayResult {
	data := params["request_post_body"]
	ret := &payment.PayResult{
		PayCode: w.Code(),
	}
	args, err := decodeXMLToMap([]byte(data))
	if err != nil {
		ret.Succ = false
		ret.ErrMsg = "微信交易请求返回数据解析失败"
		return ret
	}
	ret.Navite = args
	ret.No = args["attach"]
	ret.TradeNo = args["out_trade_no"]
	ret.ThirdAccount = args["openid"]
	ret.ThirdTradeNo = args["transaction_id"]
	if args["return_code"] != "SUCCESS" || args["result_code"] != "SUCCESS" {
		ret.Succ = false
		ret.ErrMsg = "微信支付失败:" + args["return_msg"] + ":" + args["err_code_des"]
		return ret
	}
	signSrc := args["sign"]
	sign(args, w.config.Key)
	if args["sign"] != signSrc {
		ret.Succ = false
		ret.ErrMsg = "微信支付签名验证失败"
		return ret
	}
	money, err := strconv.ParseFloat(args["total_fee"], 64)
	if err != nil {
		ret.Succ = false
		ret.ErrMsg = "交易金额异常:" + err.Error()
		return ret
	}
	ret.Succ = true
	ret.Money = money / 100
	return ret
}

//异步通知处理结果返回内容
func (w *wxpay) NotifyResult(payResult *payment.PayResult) string {
	if payResult.Succ {
		return "<xml><return_code>SUCCESS</return_code><return_msg>OK</return_msg></xml>"
	}
	return "<xml><return_code>FAIL</return_code><return_msg>处理失败</return_msg></xml>"
}

//同步结果跳转处理,返回支付结果
func (w *wxpay) Result(params map[string]string) *payment.PayResult { //微信支付没有同步返回内容
	return nil
}

//GetPayment 生成一个支付对象
func (w *wxpay) GetPayment(cfg interface{}) payment.Payment {
	var c *PayConfig
	ok := false
	if c, ok = cfg.(*PayConfig); !ok || c == nil {
		log(utils.LogLevelWarn, "传递的配置信息不是一个有效的微信支付配置")
		return nil
	}
	if c.Name == "" || c.Code == "" {
		return nil
	}
	obj := &wxpay{
		apiURL: "https://api.mch.weixin.qq.com/pay/unifiedorder",
		config: c,
	}
	obj.Init(obj.config.Code, obj.config.Name, obj.config.State)
	return obj
}

func (w *wxpay) Driver() string {
	return "wxpay"
}

//解析请求返回结果
func (w *wxpay) decodeResp(data []byte) (string, error) {
	wxRes, err := decodeXMLToMap(data)
	if err != nil {
		return "", errors.New("微信请求结果解析失败:" + err.Error())
	}
	signSrc := wxRes["sign"]
	if wxRes["return_code"] != "SUCCESS" {
		return "", errors.New("微信通讯失败:" + wxRes["return_msg"])
	} else if wxRes["result_code"] != "SUCCESS" {
		return "", errors.New("微信请求失败:" + wxRes["err_code_des"])
	}
	sign(wxRes, w.config.Key)
	if wxRes["sign"] != signSrc {
		return "", errors.New("微信签名验证失败")
	}
	return wxRes["code_url"], nil
	//编码成二维码图片返回
	//img, err := QRCode(wxRes["code_url"], 150)
	//if err != nil {
	//	log(utils.LogLevelError, "二维码创建失败:%s", err.Error())
	//	return wxRes["code_url"], nil
	//}
	//return img, nil
}

//无需确认支付
func (w *wxpay) PayConfirm(req *payment.PayConfirmRequest) *payment.PayResult {
	return payment.NoPayConfirmResult
}
