package chinapay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"sort"
	"strings"

	"github.com/kinwyb/golang/crypto/rsautil"
	"github.com/kinwyb/golang/gosql"
	"github.com/kinwyb/golang/utils"
	"github.com/astaxie/beego/config"

	"golang.org/x/crypto/pkcs12"
)

//银联Secss工具包,根据php版本转换而来

const (
	netPaySecss_CP_SIGN_FILE           = "sign.file"
	netPaySecss_CP_SIGN_FILE_PASSWORD  = "sign.file.password"
	netPaySecss_CP_SIGN_CERT_TYPE      = "sign.cert.type"
	netPaySecss_CP_SIGN_INVALID_FIELDS = "sign.invalid.fields"
	netPaySecss_CP_VERIFY_FILE         = "verify.file"
	netPaySecss_CP_SIGNATURE_FIELD     = "signature.field"
	netPaySecss_CP_KEY_VALUE_CONNECT   = "="
	netPaySecss_CP_MESSAGE_CONNECT     = "&"
)

var (
	netPaySecss_CP_LOAD_CONFIG_ERROR      = gosql.NewError(1, "加载security.properties配置文件出错，请检查文件路径！")
	netPaySecss_CP_SIGN_CERT_ERROR        = gosql.NewError(2, "签名文件路径配置错误！")
	netPaySecss_CP_SIGN_CERT_TYPE_ERROR   = gosql.NewError(4, "签名文件密钥容器类型配置错误，需为PKCS12！")
	netPaySecss_CP_VERIFY_CERT_ERROR      = gosql.NewError(6, "验签证书路径配置错误！")
	netPaySecss_CP_GET_PRI_KEY_ERROR      = gosql.NewError(8, "获取签名私钥出错！")
	netPaySecss_CP_INIT_VERIFY_CERT_ERROR = gosql.NewError(7, "初始化验签证书出错！")
	netPaySecss_CP_NO_INIT                = gosql.NewError(23, "未初化安全控件")
	netPaySecss_CP_GET_SIGN_STRING_ERROR  = gosql.NewError(10, "获取签名字符串出错！")
	netPaySecss_CP_SIGN_GOES_WRONG        = gosql.NewError(11, "签名过程发生错误！")
	netPaySecss_CP_SIGN_VALUE_NULL        = gosql.NewError(15, "报文中签名为空！")
	netPaySecss_CP_VERIFY_GOES_WRONG      = gosql.NewError(12, "验签过程发生错误！")
	netPaySecss_CP_VERIFY_FAILED          = gosql.NewError(13, "验签失败！")
	netPaySecss_CP_ENCPIN_GOES_WRONG      = gosql.NewError(17, "Pin加密过程发生错误！")
	netPaySecss_CP_DECDATA_GOES_WRONG     = gosql.NewError(19, "数据解密过程发生错误！")
)

//NetPaySecssUtil 银联Secss工具包
type NetPaySecssUtil struct {
	version           string
	signFile          string
	signFilePassword  string
	signCertType      string
	signInvalidFields map[string]int
	verifyFile        string
	MerPrivateKey     *rsa.PrivateKey
	CPPublicKey       *rsa.PublicKey
	privatePFXCertID  string
	publicCERCertID   string
	signatureField    string
	initFalg          bool
}

//InitFromFile 从文件初始化
// @param securityPropFile:string 配置文件地址
func (n *NetPaySecssUtil) InitFromFile(conf config.Configer) gosql.Error {
	n.signFile = conf.DefaultString(netPaySecss_CP_SIGN_FILE, "")
	if n.signFile == "" {
		return netPaySecss_CP_SIGN_CERT_ERROR
	}
	signFile, err := os.OpenFile(n.signFile, os.O_RDONLY, 666)
	if os.IsNotExist(err) {
		return netPaySecss_CP_SIGN_CERT_ERROR
	}
	defer signFile.Close()
	n.signFilePassword = conf.DefaultString(netPaySecss_CP_SIGN_FILE_PASSWORD, "")
	n.signCertType = conf.DefaultString(netPaySecss_CP_SIGN_CERT_TYPE, "")
	if "PKCS12" != n.signCertType {
		return netPaySecss_CP_SIGN_CERT_TYPE_ERROR
	}
	signInvalidFields := conf.DefaultString(netPaySecss_CP_SIGN_INVALID_FIELDS, "")
	if signInvalidFields != "" {
		sts := strings.Split(signInvalidFields, ",")
		n.signInvalidFields = map[string]int{}
		for i, v := range sts {
			n.signInvalidFields[v] = i
		}
	}
	n.verifyFile = conf.DefaultString(netPaySecss_CP_VERIFY_FILE, "")
	if n.verifyFile == "" {
		return netPaySecss_CP_VERIFY_CERT_ERROR
	}
	verifyFile, err := os.OpenFile(n.verifyFile, os.O_RDONLY, 666)
	if os.IsNotExist(err) {
		return netPaySecss_CP_VERIFY_CERT_ERROR
	}
	defer verifyFile.Close()
	n.signatureField = conf.DefaultString(netPaySecss_CP_SIGNATURE_FIELD, "Signature")
	merPkcs12, err := ioutil.ReadAll(signFile)
	if err != nil {
		log(utils.LogLevelError, "netpaySecssUtil获取私钥错误:%s", err.Error())
		return netPaySecss_CP_GET_PRI_KEY_ERROR
	}
	priv, cert, err := pkcs12.Decode(merPkcs12, n.signFilePassword)
	if err != nil {
		log(utils.LogLevelError, "netpaySecssUtil解析私钥错误:%s", err.Error())
		return netPaySecss_CP_GET_PRI_KEY_ERROR
	}
	n.MerPrivateKey = priv.(*rsa.PrivateKey)
	n.privatePFXCertID = cert.SerialNumber.String()
	merPkcs12, err = ioutil.ReadAll(verifyFile)
	if err != nil {
		log(utils.LogLevelError, "netpaySecssUtil初始化验签证书错误:%s", err.Error())
		return netPaySecss_CP_INIT_VERIFY_CERT_ERROR
	}
	block, _ := pem.Decode(merPkcs12)
	if block == nil {
		return netPaySecss_CP_INIT_VERIFY_CERT_ERROR
	}
	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		log(utils.LogLevelError, "netpaySecssUtil初始化验签证书错误:%s", err.Error())
		return netPaySecss_CP_INIT_VERIFY_CERT_ERROR
	}
	n.publicCERCertID = cert.SerialNumber.String()
	n.CPPublicKey = cert.PublicKey.(*rsa.PublicKey)
	n.initFalg = true
	n.version = "1.0"
	return nil
}

//Init 初始化
func (n *NetPaySecssUtil) Init(signKey []byte, signKeyPassword string, verifyKey []byte, signInvalidFields string, signatureField string) gosql.Error {
	n.signatureField = signatureField
	if n.signatureField == "" {
		n.signatureField = "Signature"
	}
	n.signFilePassword = signKeyPassword
	if signInvalidFields != "" {
		sts := strings.Split(signInvalidFields, ",")
		n.signInvalidFields = map[string]int{}
		for i, v := range sts {
			n.signInvalidFields[v] = i
		}
	}
	priv, cert, err := pkcs12.Decode(signKey, n.signFilePassword)
	if err != nil {
		log(utils.LogLevelError, "netpaySecssUtil解析私钥错误:%s", err.Error())
		return netPaySecss_CP_GET_PRI_KEY_ERROR
	}
	n.MerPrivateKey = priv.(*rsa.PrivateKey)
	n.privatePFXCertID = cert.SerialNumber.String()
	block, _ := pem.Decode(verifyKey)
	if block == nil {
		return netPaySecss_CP_INIT_VERIFY_CERT_ERROR
	}
	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		log(utils.LogLevelError, "netpaySecssUtil初始化验签证书错误:%s", err.Error())
		return netPaySecss_CP_INIT_VERIFY_CERT_ERROR
	}
	n.publicCERCertID = cert.SerialNumber.String()
	n.CPPublicKey = cert.PublicKey.(*rsa.PublicKey)
	n.initFalg = true
	n.version = "1.0"
	return nil
}

//Sign 签名
// @param params:map[string]string 待签名的数据
// @return string 签名结果
//         gosql.Error 签名错误内容
func (n *NetPaySecssUtil) Sign(params map[string]string) (string, gosql.Error) {
	if !n.initFalg {
		return "", netPaySecss_CP_NO_INIT
	}
	signRawData := n.getSignStr(params)
	if signRawData == "" {
		return "", netPaySecss_CP_GET_SIGN_STRING_ERROR
	}
	hashed := sha512.Sum512([]byte(signRawData))
	signer, err := rsa.SignPKCS1v15(rand.Reader, n.MerPrivateKey, crypto.SHA512, hashed[:])
	if err != nil {
		log(utils.LogLevelError, "签名发生错误:%s", err.Error())
		return "", netPaySecss_CP_SIGN_GOES_WRONG
	}
	return base64.StdEncoding.EncodeToString(signer), nil
}

//Verify 验签
// @param params:map[string]string 待验签数据
// @param signatureField:string 签名所在字段
// @return bool 验证是否成功
func (n *NetPaySecssUtil) Verify(params map[string]string, signatureField ...string) (bool, gosql.Error) {
	if !n.initFalg {
		return false, netPaySecss_CP_NO_INIT
	}
	signField := n.signatureField
	if signatureField != nil && len(signatureField) > 0 {
		signField = signatureField[0]
	}
	orgSignMsg := params[signField]
	if orgSignMsg == "" {
		return false, netPaySecss_CP_SIGN_VALUE_NULL
	}
	log(utils.LogLevelTrace, "NetPaySecssUtil::Verify => 待验证签名:%s", orgSignMsg)
	delete(params, signField)
	verifySignData := n.getSignStr(params)
	hashed := sha512.Sum512([]byte(verifySignData))
	signByte, err := base64.StdEncoding.DecodeString(orgSignMsg)
	if err != nil {
		log(utils.LogLevelError, "NetPaySecssUtil::Verify => 签名数据格式异常,正确签名应该为base64编码:%s", err.Error())
		return false, netPaySecss_CP_VERIFY_GOES_WRONG
	}
	err = rsa.VerifyPKCS1v15(n.CPPublicKey, crypto.SHA512, hashed[:], signByte)
	if err != nil {
		log(utils.LogLevelWarn, "NetPaySecssUtil::Verify => 验签失败:%s", err.Error())
		return false, netPaySecss_CP_VERIFY_FAILED
	}
	return true, nil
}

//EncryptPin PIN加密
// @param pin:string pin
// @param card:string card
func (n *NetPaySecssUtil) EncryptPin(pin, card string) (string, gosql.Error) {
	if !n.initFalg {
		return "", netPaySecss_CP_NO_INIT
	}
	pinBlock := n.pin2PinBlockWithCardNO(pin, card)
	if pinBlock == "" {
		return "", netPaySecss_CP_ENCPIN_GOES_WRONG
	}
	crypted := bcpowmod(bin2int(pinBlock), big.NewInt(int64(n.CPPublicKey.E)), n.CPPublicKey.N)
	rb := bcdechex(crypted)
	rb = padstr(rb)
	ret, _ := hex.DecodeString(rb)
	return base64.StdEncoding.EncodeToString(ret), nil
}

//EncryptData 数据加密
// @param data:string 需要加密的内容
func (n *NetPaySecssUtil) EncryptData(data string) (string, gosql.Error) {
	if !n.initFalg {
		return "", netPaySecss_CP_NO_INIT
	}
	crypted := bcpowmod(bin2int(data), big.NewInt(int64(n.CPPublicKey.E)), n.CPPublicKey.N)
	rb := bcdechex(crypted)
	rb = padstr(rb)
	ret, _ := hex.DecodeString(rb)
	return base64.StdEncoding.EncodeToString(ret), nil
}

//DecryptData 数据解密
// @param data:string 需要解密的内容
func (n *NetPaySecssUtil) DecryptData(data string) (string, gosql.Error) {
	if !n.initFalg {
		return "", netPaySecss_CP_NO_INIT
	}
	raw, _ := base64.StdEncoding.DecodeString(data)
	raw = rsautil.DecryptNoPadding(n.MerPrivateKey, raw)
	return string(raw), nil
}

//SignCertID 签名证书编号
func (n *NetPaySecssUtil) SignCertID() string {
	return n.privatePFXCertID
}

//GetPrivatePFXCertID 私钥证书编号
func (n *NetPaySecssUtil) GetPrivatePFXCertID() string {
	return n.privatePFXCertID
}

//GetPublicCERCertID 公钥证书编号
func (n *NetPaySecssUtil) GetPublicCERCertID() string {
	return n.publicCERCertID
}

//Version 版本号
func (n *NetPaySecssUtil) Version() string {
	return n.version
}

func (n *NetPaySecssUtil) pin2PinBlockWithCardNO(pin, card string) string {
	tPinByte := n.pin2PinBlock(pin)
	if tPinByte == nil || len(tPinByte) < 8 {
		return ""
	}
	if len(card) == 11 {
		card = "00" + card
	} else if len(card) == 12 {
		card = "0" + card
	}
	tPanByte := n.formatPan(card)
	if tPanByte == nil || len(tPanByte) < 8 {
		return ""
	}
	ret := make([]int64, 8)
	for i := 0; i < 8; i++ {
		ret[i] = tPinByte[i] ^ tPanByte[i]
	}
	result := make([]byte, len(ret))
	for i, v := range ret {
		result[i] = byte(v)
	}
	return string(result)
}

func (n *NetPaySecssUtil) formatPan(card string) []int64 {
	tPanLen := len(card)
	tmp := tPanLen - 13
	ret := []int64{0, 0}
	for i := 2; i < 8; i++ {
		a := hexdec(card[tmp : tmp+2])
		ret = append(ret, int64(a))
		tmp += 2
	}
	return ret
}

func (n *NetPaySecssUtil) pin2PinBlock(pin string) []int64 {
	tTemp := 1
	tPinLen := len(pin)
	ret := []int64{int64(tPinLen)}
	if tPinLen%2 == 0 {
		for i := 0; i < tPinLen; {
			a := hexdec(pin[i : i+2])
			ret = append(ret, int64(a))
			if (i == tPinLen-2) && tTemp < 7 {
				for x := tTemp + 1; x < 8; x++ {
					ret = append(ret, -1)
				}
			}
			tTemp++
			i += 2
		}
	} else {
		for i := 0; i < tPinLen-1; {
			a := hexdec(pin[i : i+2])
			ret = append(ret, int64(a))
			if i == tPinLen-3 {
				b := hexdec(pin[tPinLen-1:] + "F")
				ret = append(ret, int64(b))
				if tTemp+1 < 7 {
					for x := tTemp + 2; x < 8; x++ {
						ret = append(ret, -1)
					}
				}
			}
			tTemp++
			i += 2
		}
	}
	return ret
}

func (n *NetPaySecssUtil) getSignStr(params map[string]string) string {
	keys := []string{}
	for k := range params {
		if _, ok := n.signInvalidFields[k]; ok {
			continue
		}
		keys = append(keys, k)
	}
	if len(keys) < 1 { //一个参数都没有直接返回
		return ""
	}
	sort.Strings(keys)
	bf := &bytes.Buffer{}
	for _, k := range keys {
		bf.WriteString(k)
		bf.WriteString(netPaySecss_CP_KEY_VALUE_CONNECT)
		bf.WriteString(params[k])
		bf.WriteString(netPaySecss_CP_MESSAGE_CONNECT)
	}
	if bf.Len() > 0 {
		bf.Truncate(bf.Len() - 1)
	}
	return bf.String()
}
