package ssltool

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func Test_crt(t *testing.T) {
	convey.Convey("证书生成", t, func() {
		baseinfo := CertInformation{IsCA: true, CommonName: "zhifangw.cn",
			CrtName: "test_root.crt", KeyName: "test_root.key"}
		err := CreateCRT(nil, nil, baseinfo)
		if err != nil {
			convey.Printf("Create crt error,Error info:%s\n", err)
		}
		convey.So(err, convey.ShouldBeNil)
		crtinfo := baseinfo
		crtinfo.IsCA = false
		crtinfo.CrtName = "test_server.crt"
		crtinfo.KeyName = "test_server.key"
		crtinfo.Names = []pkix.AttributeTypeAndValue{{asn1.ObjectIdentifier{2, 1, 3}, "MAC_ADDR"}} //添加扩展字段用来做自定义使用
		crt, pri, err := Parse(baseinfo.CrtName, baseinfo.KeyName)
		if err != nil {
			convey.Printf("Parse crt error,Error info:%s\n", err)
		}
		convey.So(err, convey.ShouldBeNil)
		err = CreateCRT(crt, pri, crtinfo)
		if err != nil {
			convey.Printf("Create crt error,Error info:%s\n", err)
		}
		convey.So(err, convey.ShouldBeNil)
		os.Remove(baseinfo.CrtName)
		os.Remove(baseinfo.KeyName)
		os.Remove(crtinfo.CrtName)
		os.Remove(crtinfo.KeyName)
	})
}
