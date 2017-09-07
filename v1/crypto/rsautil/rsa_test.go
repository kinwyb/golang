package rsautil

import (
	"fmt"
	"testing"
)

func Test_Res(t *testing.T) {
	privatekey, publickey, err := GenRsaKey(2048, PKCS8)
	if err != nil {
		t.Errorf("密钥生成失败:%v", err)
		return
	}
	fmt.Printf("私钥:\r\n%s", string(privatekey))
	fmt.Printf("公钥:\r\n%s", string(publickey))
	data := []byte("加密测试")
	edata, err := Encrypt(publickey, data)
	if err != nil {
		t.Errorf("加密失败:%s", err.Error())
		return
	}
	d, err := Decrypt(privatekey, edata, PKCS8)
	if err != nil {
		t.Errorf("解密失败:%s", err.Error())
		return
	}
	fmt.Printf("解密数据:%s\r", string(d))
}
