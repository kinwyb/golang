package ssltool

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	rd "math/rand"
	"net"
	"os"
	"time"
)

func init() {
	rd.Seed(time.Now().UnixNano())
}

//CertInformation 证书基本信息
type CertInformation struct {
	Country            []string
	Organization       []string
	OrganizationalUnit []string
	EmailAddress       []string
	Province           []string
	Locality           []string
	CommonName         string
	CrtName, KeyName   string
	IsCA               bool
	Names              []pkix.AttributeTypeAndValue
	IPAddresses        []net.IP
	DNSNames           []string
	Year               int
	Months             int
	Days               int
}

//CreateCRT 生成证书
func CreateCRT(RootCa *x509.Certificate, RootKey *rsa.PrivateKey, info CertInformation) error {
	Crt := newCertificate(info)
	Key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	var buf []byte
	if RootCa == nil || RootKey == nil {
		//创建自签名证书
		buf, err = x509.CreateCertificate(rand.Reader, Crt, Crt, &Key.PublicKey, Key)
	} else {
		//使用根证书签名
		buf, err = x509.CreateCertificate(rand.Reader, Crt, RootCa, &Key.PublicKey, RootKey)
	}
	if err != nil {
		return err
	}
	err = write(info.CrtName, "CERTIFICATE", buf)
	if err != nil {
		return err
	}
	buf = x509.MarshalPKCS1PrivateKey(Key)
	return write(info.KeyName, "RSA PRIVATE KEY", buf)
}

//编码写入文件
func write(filename, Type string, p []byte) error {
	File, err := os.Create(filename)
	defer File.Close()
	if err != nil {
		return err
	}
	b := &pem.Block{Bytes: p, Type: Type}
	return pem.Encode(File, b)
}

//Parse 解析证书和密钥
func Parse(crtPath, keyPath string) (rootcertificate *x509.Certificate, rootPrivateKey *rsa.PrivateKey, err error) {
	rootcertificate, err = ParseCrt(crtPath)
	if err != nil {
		return
	}
	rootPrivateKey, err = ParseKey(keyPath)
	return
}

//ParseCrt 解析证书
func ParseCrt(path string) (*x509.Certificate, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p := &pem.Block{}
	p, buf = pem.Decode(buf)
	return x509.ParseCertificate(p.Bytes)
}

//ParseKey 解析密钥
func ParseKey(path string) (*rsa.PrivateKey, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, buf := pem.Decode(buf)
	return x509.ParsePKCS1PrivateKey(p.Bytes)
}

func newCertificate(info CertInformation) *x509.Certificate {
	if info.Year == 0 && info.Months == 0 && info.Days == 0 {
		info.Year = 10
	}
	cert := &x509.Certificate{
		SerialNumber:          big.NewInt(rd.Int63()),
		NotBefore:             time.Now(),                                            //证书的开始时间
		NotAfter:              time.Now().AddDate(info.Year, info.Months, info.Days), //证书的结束时间
		BasicConstraintsValid: true,                                                  //基本的有效性约束
		IsCA:        info.IsCA,
		IPAddresses: info.IPAddresses,
		DNSNames:    info.DNSNames,                                                              //是否是根证书
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}, //证书用途
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	cert.Subject = pkix.Name{}
	if info.Country != nil && len(info.Country) > 0 {
		cert.Subject.Country = info.Country
	}
	if info.Organization != nil && len(info.Organization) > 0 {
		cert.Subject.Organization = info.Organization
	}
	if info.OrganizationalUnit != nil && len(info.OrganizationalUnit) > 0 {
		cert.Subject.OrganizationalUnit = info.OrganizationalUnit
	}
	if info.Province != nil && len(info.Province) > 0 {
		cert.Subject.Province = info.Province
	}
	if info.CommonName != "" {
		cert.Subject.CommonName = info.CommonName
	}
	if info.Locality != nil && len(info.Locality) > 0 {
		cert.Subject.Locality = info.Locality
	}
	if info.Names != nil && len(info.Names) > 0 {
		cert.Subject.ExtraNames = info.Names
	}
	if info.EmailAddress != nil && len(info.EmailAddress) > 0 {
		cert.EmailAddresses = info.EmailAddress
	}
	return cert
}
