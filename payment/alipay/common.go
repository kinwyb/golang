package alipay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"sort"
	"strings"

	"encoding/pem"

	"time"

	"net/url"

	"github.com/kinwyb/golang/utils"
)

//签名
func sign(args map[string]string, privatekey string) {
	keys := paraFilter(args)
	signStr := createLinkString(keys, args)
	data, err := decodeRSAKey(privatekey)
	if err != nil {
		log(utils.LogLevelError, "支付宝私钥解析失败")
		return
	}
	priv, err := x509.ParsePKCS8PrivateKey(data)
	if err != nil {
		log(utils.LogLevelError, "支付宝签名RSA私钥初始化失败:"+err.Error())
		return
	}
	log(utils.LogLevelDebug, "支付宝签名字符串:%s", signStr)
	dt := sha256.Sum256([]byte(signStr))
	data, err = rsa.SignPKCS1v15(rand.Reader, priv.(*rsa.PrivateKey), crypto.SHA256, dt[:])
	if err != nil {
		log(utils.LogLevelError, "支付宝签名失败:"+err.Error())
		return
	}
	args["sign"] = base64.StdEncoding.EncodeToString(data)
	args["sign_type"] = "RSA2"
	log(utils.LogLevelError, "签名结果:%s", args["sign"])
}

//过滤
func paraFilter(params map[string]string) []string {
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
func createLinkString(keys []string, args map[string]string) string {
	buf := bytes.NewBufferString("")
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(args[k])
		buf.WriteString("&")
	}
	if buf.Len() > 0 {
		buf.Truncate(buf.Len() - 1)
	}
	return buf.String()
}

//decodeRSAKey 解析RSA密钥
func decodeRSAKey(key string) ([]byte, error) {
	if key[0] == '-' {
		block, _ := pem.Decode([]byte(key))
		if block == nil {
			return nil, errors.New("支付宝提现签名私钥解析失败")
		}
		return block.Bytes, nil
	}
	return base64.StdEncoding.DecodeString(key)
}

//request请求
func request(service string, config *PayConfig, bizContent string, getway string) ([]byte, error) {
	args := buildParams(service, config, bizContent)
	params := url.Values{}
	for k, v := range args {
		params.Add(k, v)
	}
	log(utils.LogLevelDebug, "支付宝接口请求参数:%s", params.Encode())
	resp, err := http.Post(getway,
		"application/x-www-form-urlencoded;charset=utf-8", strings.NewReader(params.Encode()))
	if err != nil {
		log(utils.LogLevelError, "支付宝接口请求异常:%s", err.Error())
		return nil, fmt.Errorf("请求失败")
	}
	respdata, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log(utils.LogLevelError, "支付宝接口结果读取异常:%s", err.Error())
		return nil, fmt.Errorf("结果读取失败")
	}
	return respdata, nil
}

func buildForm(service string, config *PayConfig, bizContent string, getway string) string {
	sParams := buildParams(service, config, bizContent)
	buf := bytes.NewBufferString("<form id=\"alipaysubmit\" name=\"alipaysubmit\" action=\"")
	buf.WriteString(getway)
	buf.WriteString("?charset=UTF-8\" method=\"POST\">\n")
	for k, v := range sParams {
		buf.WriteString("<input type=\"hidden\" name=\"")
		buf.WriteString(k)
		buf.WriteString("\" value='")
		buf.WriteString(v)
		buf.WriteString("' />\n")
	}
	buf.WriteString("<input type=\"submit\" value=\"提交\" style=\"display:none;\"></form>")
	buf.WriteString("<script>document.forms['alipaysubmit'].submit();</script>")
	return buf.String()
}

//生成请求参数
func buildParams(service string, config *PayConfig, bizContent string) map[string]string {
	args := map[string]string{
		"app_id":      config.Partner,
		"method":      service,
		"format":      "json",
		"charset":     "UTF-8",
		"sign_type":   "RSA2",
		"timestamp":   time.Now().Format("2006-01-02 15:04:05"),
		"version":     "1.0",
		"biz_content": bizContent,
	}
	if config.NotifyURL != "" {
		args["notify_url"] = config.NotifyURL
	}
	if config.ReturnURL != "" {
		args["return_url"] = config.ReturnURL
	}
	sign(args, config.PrivateKey)
	return args
}

//verify 支付结果校验
func verify(response string, signString string, publicKey string) bool {
	sign, _ := base64.StdEncoding.DecodeString(signString)
	data, err := decodeRSAKey(publicKey)
	if err != nil {
		log(utils.LogLevelError, "支付宝公钥解析失败")
		return false
	}
	pubi, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		log(utils.LogLevelError, "支付宝结果校验RSA公钥初始化错误:"+err.Error())
		return false
	}
	dt := sha256.Sum256([]byte(response))
	err = rsa.VerifyPKCS1v15(pubi.(*rsa.PublicKey), crypto.SHA256, dt[:], sign)
	if err != nil {
		log(utils.LogLevelError, "支付宝结果校验失败:"+err.Error())
		log(utils.LogLevelError, "支付宝校验签名的字符串:%s", response)
		log(utils.LogLevelError, "支付宝校验的签名:%s", signString)
		return false
	}
	return true
}
