package wxpay

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"image/png"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/kinwyb/golang/payment"

	"github.com/kinwyb/golang/utils"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"

	"math/rand"

	"crypto/md5"

	"net/http"

	"io/ioutil"

	"strconv"

	"io"

	"encoding/hex"
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
		"nonce_str":    w.nonceStr(),                                //随机字符串
		"body":         req.Desc,                                    //商品名称
		"attach":       req.No,                                      //由于统一订单号无法重复发起支付所以订单号只能存放在附加字段,交易单号重新生成
		"total_fee":    strconv.FormatInt(int64(req.Money*100), 10), //交易金额,单位分
		"notify_url":   w.config.NotifyURL,
		"trade_type":   "NATIVE",
		"product_id":   "0",
		"out_trade_no": time.Now().Format("150405") + req.No,
	}
	w.sign(params)
	xmlBuf := w.buildXML(params)
	resp, err := http.Post(w.apiURL, "application/xml;charset=utf-8", xmlBuf)
	if err != nil {
		return "", errors.New("微信请求失败:" + err.Error())
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", errors.New("微信请求失败:" + err.Error())
	}
	return w.decodeResp(data)
}

//异步结果通知处理,返回支付结果
func (w *wxpay) Notify(params map[string]string) *payment.PayResult {
	data := params["request_post_body"]
	ret := &payment.PayResult{
		PayCode: w.Code(),
	}
	args, err := w.decodeXMLToMap([]byte(data))
	if err != nil {
		ret.Succ = false
		ret.ErrMsg = "微信交易请求返回数据解析失败"
		return ret
	}
	ret.Navite = args
	if args["return_code"] != "SUCCESS" || args["result_code"] != "SUCCESS" {
		ret.Succ = false
		ret.ErrMsg = "微信支付失败:" + args["return_msg"] + ":" + args["err_code_des"]
		return ret
	}
	sign := args["sign"]
	w.sign(args)
	if args["sign"] != sign {
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
	ret.No = args["attach"]
	ret.TradeNo = args["out_trade_no"]
	ret.ThirdAccount = args["openid"]
	ret.ThirdTradeNo = args["transaction_id"]
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

//随机32位字符串
func (w *wxpay) nonceStr() string {
	chars := []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	rand.Seed(time.Now().UnixNano())
	result := [32]byte{}
	for i := 0; i < 32; i++ {
		result[i] = chars[rand.Int31n(35)]
	}
	return string(result[:])
}

//过滤
func (w *wxpay) paraFilter(params map[string]string) []string {
	keys := make([]string, 0)
	for k, v := range params {
		if k == "sign" || strings.TrimSpace(v) == "" {
			delete(params, k)
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

//拼接字符串 按照“参数=参数值”的模式用“&”字符拼接成字符串
func (w *wxpay) createLinkString(keys []string, args map[string]string) string {
	buf := bytes.NewBufferString("")
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(args[k])
		buf.WriteString("&")
	}
	buf.Truncate(buf.Len() - 1)
	return buf.String()
}

//签名
func (w *wxpay) sign(args map[string]string) {
	keys := w.paraFilter(args)
	signStr := w.createLinkString(keys, args) + "&key=" + w.config.Key
	sign := md5.Sum([]byte(signStr))
	args["sign"] = strings.ToUpper(hex.EncodeToString(sign[:]))
}

//创建XML
func (w *wxpay) buildXML(args map[string]string) *bytes.Buffer {
	buf := bytes.NewBufferString("<xml>")
	for k, v := range args {
		buf.WriteString("<")
		buf.WriteString(k)
		buf.WriteString(">")
		buf.WriteString(v)
		buf.WriteString("</")
		buf.WriteString(k)
		buf.WriteString(">")
	}
	buf.WriteString("</xml>")
	return buf
}

//解析请求返回结果
func (w *wxpay) decodeResp(data []byte) (string, error) {
	wxRes, err := w.decodeXMLToMap(data)
	if err != nil {
		return "", errors.New("微信请求结果解析失败:" + err.Error())
	}
	sign := wxRes["sign"]
	if wxRes["return_code"] != "SUCCESS" {
		return "", errors.New("微信通讯失败:" + wxRes["return_msg"])
	} else if wxRes["result_code"] != "SUCCESS" {
		return "", errors.New("微信请求失败:" + wxRes["err_code_des"])
	}
	w.sign(wxRes)
	if wxRes["sign"] != sign {
		return "", errors.New("微信签名验证失败")
	}
	img, err := QRCode(wxRes["code_url"], 150)
	if err != nil {
		log(utils.LogLevelError, "二维码创建失败:%s", err.Error())
		return wxRes["code_url"], nil
	}
	return img, nil
}

//解析XML
func (w *wxpay) decodeXMLToMap(data []byte) (map[string]string, error) {
	buf := bytes.NewBuffer(data)
	x := xml.NewDecoder(buf)
	ret := map[string]string{}
	var key, value string
	var t xml.Token
	var err error
	for t, err = x.Token(); err == nil; t, err = x.Token() {
		switch token := t.(type) {
		case xml.StartElement:
			key = token.Name.Local
		case xml.CharData:
			value = string([]byte(token))
			if key != "" {
				ret[key] = value
				key = ""
				value = ""
			}
		}
	}
	if err == io.EOF {
		err = nil
	}
	return ret, err
}

//QRCode 生成二维码
//@param content string 二维码
//@param size int  大小
//@param filename int 保存文件名称[空时不保存文件]
//@return string base64编码图片数据[如果保存到文件返回空]
func QRCode(content string, size int, filename ...string) (string, error) {
	qrCode, _ := qr.Encode(content, qr.M, qr.Auto)
	qrCode, _ = barcode.Scale(qrCode, size, size)
	if filename != nil && len(filename) > 0 {
		file, err := os.Create(filename[0])
		if err != nil {
			return "", errors.New("文件创建失败:" + err.Error())
		}
		err = png.Encode(file, qrCode)
		file.Close()
		if err != nil {
			return "", errors.New("图像编码失败:" + err.Error())
		}
		return "", nil
	}
	bt := &bytes.Buffer{}
	err := png.Encode(bt, qrCode)
	if err != nil {
		return "", errors.New("图像编码失败:" + err.Error())
	}
	imgstr := base64.StdEncoding.EncodeToString(bt.Bytes())
	return "data:image/png;base64," + imgstr, nil
}

//无需确认支付
func (w *wxpay) PayConfirm(req *payment.PayConfirmRequest) *payment.PayResult {
	return payment.NoPayConfirmResult
}
