package wxpay

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"image/png"
	"io"
	"math/rand"
	"os"
	"strings"
	"time"

	"sort"

	"crypto/md5"
	"encoding/hex"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

//公共函数

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

//解析XML
func decodeXMLToMap(data []byte) (map[string]string, error) {
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

//创建XML
func buildXML(args map[string]string) *bytes.Buffer {
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

//拼接字符串 按照“参数=参数值”的模式用“&”字符拼接成字符串
func createLinkString(keys []string, args map[string]string) string {
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

//随机32位字符串
func nonceStr() string {
	chars := []byte("abcdefghijklmnopqrstuvwxyz0123456789")
	rand.Seed(time.Now().UnixNano())
	result := [32]byte{}
	for i := 0; i < 32; i++ {
		result[i] = chars[rand.Int31n(35)]
	}
	return string(result[:])
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

//签名
func sign(args map[string]string, key string) {
	keys := paraFilter(args)
	signStr := createLinkString(keys, args) + "&key=" + key
	sign := md5.Sum([]byte(signStr))
	args["sign"] = strings.ToUpper(hex.EncodeToString(sign[:]))
}
