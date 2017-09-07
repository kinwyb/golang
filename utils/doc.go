/*
Package utils 简单工具包

包含: AES加密..JS模版解析..Proxy HTTP代理..idWork唯一id生成工具
*/
package utils

import (
	"bytes"
	"io/ioutil"
	"strconv"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

//Unicode2String Unicode编码转utf8字符串
func Unicode2String(str string) string {
	buf := bytes.NewBuffer(nil)
	i, j := 0, len(str)
	for i < j {
		x := i + 6
		if x > j {
			buf.WriteString(str[i:])
			break
		}
		if str[i] == '\\' && str[i+1] == 'u' {
			hex := str[i+2 : x]
			r, err := strconv.ParseUint(hex, 16, 64)
			if err == nil {
				buf.WriteRune(rune(r))
			} else {
				buf.WriteString(str[i:x])
			}
			i = x
		} else {
			buf.WriteByte(str[i])
			i++
		}
	}
	return buf.String()
}

//GBK2UTF8 gbk转utf8
func GBK2UTF8(str string) string {
	ret, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(str)), simplifiedchinese.GB18030.NewDecoder()))
	if err != nil {
		return str
	}
	return string(ret)
}
