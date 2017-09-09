package chanpay

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/kinwyb/golang/crypto/rsautil"
	"github.com/kinwyb/golang/utils"
)

//签名
func sign(params map[string]string, privateKey []byte) error {
	if params != nil && len(params) > 0 {
		keys := make([]string, 0)
		for k, v := range params {
			if k == "Sign" || k == "SignType" || strings.TrimSpace(v) == "" {
				delete(params, k)
			} else {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		bf := &bytes.Buffer{}
		for _, k := range keys {
			bf.WriteString(k)
			bf.WriteString("=")
			bf.WriteString(params[k])
			bf.WriteString("&")
		}
		bf.Truncate(bf.Len() - 1)
		log(utils.LogLevelTrace, "畅捷支付待签名字符串:%s", bf.String())
		sign, err := rsautil.SignSHA1(privateKey, bf.Bytes(), rsautil.PKCS8)
		if err != nil {
			log(utils.LogLevelError, "畅捷支付签名失败:%s", err.Error())
			return err
		}
		params["Sign"] = base64.StdEncoding.EncodeToString(sign)
		log(utils.LogLevelTrace, "畅捷支付签名结果:%s", params["Sign"])
		params["SignType"] = "RSA"
	}
	return nil
}

//验证签名
func verify(params map[string]string, publicKey []byte) bool {
	var sign []byte
	if params["Sign"] != "" {
		sign, _ = base64.StdEncoding.DecodeString(params["Sign"])
	} else {
		sign, _ = base64.StdEncoding.DecodeString(params["sign"])
	}
	keys := make([]string, 0)
	for k, v := range params {
		if k == "Sign" || k == "SignType" || k == "sign" || k == "sign_type" || strings.TrimSpace(v) == "" {
			delete(params, k)
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	bf := &bytes.Buffer{}
	for _, k := range keys {
		bf.WriteString(k)
		bf.WriteString("=")
		bf.WriteString(params[k])
		bf.WriteString("&")
	}
	bf.Truncate(bf.Len() - 1)
	ret, err := rsautil.VerifySHA1(publicKey, bf.Bytes(), sign)
	if err != nil {
		log(utils.LogLevelError, "畅捷支付验签错误:%s", err.Error())
		return ret
	}
	return ret
}

//拼接字符串 按照“参数=参数值”的模式用“&”字符拼接成字符串
func buildRequestQueryString(args map[string]string) string {
	buf := bytes.NewBufferString("")
	for k, v := range args {
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(url.QueryEscape(url.QueryEscape(v)))
		buf.WriteString("&")
	}
	buf.Truncate(buf.Len() - 1)
	return buf.String()
}

//请求
func request(apiURL string, params map[string]string, privateKey []byte, publicKey []byte) (map[string]string, error) {
	err := sign(params, privateKey)
	if err != nil {
		return nil, errors.New("签名失败")
	}
	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(buildRequestQueryString(params)))
	if err != nil {
		return nil, errors.New("畅捷支付请求失败:" + err.Error())
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, errors.New("畅捷支付请求失败:" + err.Error())
	}
	log(utils.LogLevelInfo, "畅捷支付请求结果:%s", data)
	result := map[string]string{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log(utils.LogLevelInfo, "畅捷支付结果解析失败:%s", data)
		return nil, errors.New("畅捷支付结果解析失败")
	}
	if !verify(result, publicKey) {
		log(utils.LogLevelWarn, "畅捷支付返回结果验签失败")
	} else if result["AcceptStatus"] == "S" && (result["RetCode"] == "SYSTEM_SUCCESS" || result["RetCode"] == "S001" || result["RetCode"] == "P0002") {
		return result, nil
	}
	log(utils.LogLevelError, "畅捷支付请求失败:%s[%s]%s", result["AcceptStatus"], result["RetCode"], result["RetMsg"])
	return nil, errors.New("畅捷支付请求失败:[" + result["RetCode"] + "]" + result["RetMsg"])
}

//编码订单号
func encodeNo(no string) string {
	no = "1" + no + time.Now().Format("150405.999")
	no = strings.Replace(no, ".", "", -1)
	if bi, ok := new(big.Int).SetString(no, 10); ok {
		return bi.Text(32)
	}
	return no
}

//解码订单号
func decodeNo(no string) string {
	if bi, ok := new(big.Int).SetString(no, 32); ok {
		no = bi.Text(10)
		if len(no) < 10 {
			return no
		}
		return no[1 : len(no)-9]
	}
	return no
}
