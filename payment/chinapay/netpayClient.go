package chinapay

import (
	"crypto/cipher"
	"crypto/des"
	"crypto/sha1"
	"errors"
	"strings"

	"github.com/kinwyb/golang/utils"

	"math/big"

	"bytes"
	"sort"

	"github.com/astaxie/beego/config"
)

const (
	desKey  = "SCUBEPGW"
	hashPAD = "0001ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff003021300906052b0e03021a05000414"
)

//netPayKey 银联密钥
type netPayKey struct {
	merID          string
	pgID           string
	modulus        string
	prime1         string
	prime2         string
	primeExponent1 string
	primeExponent2 string
	coefficient    string
}

//NetPayClient
type NetPayClient struct {
	key *netPayKey
}

func (n *NetPayClient) sha1_128(str string) string {
	hash := sha1.Sum([]byte(str))
	return hex2bin(hashPAD) + string(hash[:])
}

func BuildNetPayClientKeyFile(keypath string) (*NetPayClient, error) {
	client := &NetPayClient{}
	cfg, err := config.NewConfig("ini", keypath)
	if err != nil {
		return nil, errors.New("配置数据解析失败:" + err.Error())
	}
	if client.BuildKey(cfg) {
		return client, nil
	}
	return nil, nil
}

func BuildNetPayClientKey(key []byte) (*NetPayClient, error) {
	client := &NetPayClient{}
	cfg, err := config.NewConfigData("ini", key)
	if err != nil {
		return nil, errors.New("配置数据解析失败:" + err.Error())
	}
	if client.BuildKey(cfg) {
		return client, nil
	}
	return nil, nil
}

func (n *NetPayClient) rsaEncrypt(input string) string {
	p := bin2int(n.key.prime1)
	q := bin2int(n.key.prime2)
	u := bin2int(n.key.coefficient)
	dP := bin2int(n.key.primeExponent1)
	dQ := bin2int(n.key.primeExponent2)
	c := bin2int(input)
	cp := big.NewInt(0)
	cq := big.NewInt(0)
	cp = cp.Mod(c, p)
	cq = cq.Mod(c, q)
	a := bcpowmod(cp, dP, p)
	b := bcpowmod(cq, dQ, q)
	result := big.NewInt(0)
	if a.Cmp(b) >= 0 {
		result = result.Sub(a, b)
	} else {
		result = result.Sub(b, a)
		result = result.Sub(p, result)
	}
	result = result.Mod(result, p)
	result = result.Mul(result, u)
	result = result.Mod(result, p)
	result = result.Mul(result, q)
	result = result.Add(result, b)
	ret := bcdechex(result)
	ret = strings.ToUpper(padstr(ret))
	if len(ret) == 256 {
		return ret
	}
	return ""
}

func (n *NetPayClient) rsaDecrypt(input string) string {
	check := bchexdec(input)
	modulus := bin2int(n.key.modulus)
	exponent := bchexdec("010001")
	result := check.Exp(check, exponent, modulus)
	ret := bcdechex(result)
	return strings.ToUpper(padstr(ret))
}

func (n *NetPayClient) BuildKey(cfg config.Configer) bool {
	merID := cfg.String("NetPayClient::MERID")
	pgID := cfg.String("NetPayClient::PGID")
	if merID != "" {
		n.key = &netPayKey{}
		n.key.merID = merID
		hex := cfg.String("NetPayClient::prikeyS")
		if hex == "" || len(hex) < 704 {
			log(utils.LogLevelError, "prikeyS长度异常")
			return false
		}
		bin := hex2bin(hex[80:])
		n.key.modulus = bin[:128]
		cpr, err := des.NewCipher([]byte(desKey))
		if err != nil {
			log(utils.LogLevelError, "DES加密初始化失败:%s", err.Error())
			return false
		}
		vi := strings.Repeat("\x00", 8)
		//prime1
		prime1 := bin[384 : 384+64]
		cp1 := cipher.NewCBCDecrypter(cpr, []byte(vi))
		origData := make([]byte, 64)
		cp1.CryptBlocks(origData, []byte(prime1))
		n.key.prime1 = string(origData)
		//prime2
		prime2 := bin[448 : 448+64]
		prime2origData := make([]byte, 64)
		cp2 := cipher.NewCBCDecrypter(cpr, []byte(vi))
		cp2.CryptBlocks(prime2origData, []byte(prime2))
		n.key.prime2 = string(prime2origData)
		//prime_exponent1
		primeExponent1 := bin[512 : 512+64]
		primeExponent1OrigData := make([]byte, 64)
		cpe1 := cipher.NewCBCDecrypter(cpr, []byte(vi))
		cpe1.CryptBlocks(primeExponent1OrigData, []byte(primeExponent1))
		n.key.primeExponent1 = string(primeExponent1OrigData)
		//prime_exponent2
		primeExponent2 := bin[576 : 576+64]
		primeExponent2origData := make([]byte, 64)
		cpe2 := cipher.NewCBCDecrypter(cpr, []byte(vi))
		cpe2.CryptBlocks(primeExponent2origData, []byte(primeExponent2))
		n.key.primeExponent2 = string(primeExponent2origData)
		//coefficient
		coefficient := bin[640 : 640+64]
		coefficientOrigData := make([]byte, 64)
		cpcoefficient := cipher.NewCBCDecrypter(cpr, []byte(vi))
		cpcoefficient.CryptBlocks(coefficientOrigData, []byte(coefficient))
		n.key.coefficient = string(coefficientOrigData)
	} else if pgID != "" {
		n.key = &netPayKey{
			pgID: pgID,
		}
		hex := cfg.String("NetPayClient::pubkeyS")
		if hex == "" || len(hex) < 48 {
			log(utils.LogLevelError, "pubkeyS长度异常")
			return false
		}
		bin := hex2bin(hex[48:])
		n.key.modulus = bin[:128]
	} else {
		log(utils.LogLevelError, "配置文件错误不存在MERID和PGID")
		return false
	}
	return true
}

func (n *NetPayClient) Sign(msg string) string {
	if n.key == nil || n.key.merID == "" {
		return ""
	}
	hb := n.sha1_128(msg)
	return n.rsaEncrypt(hb)
}

func (n *NetPayClient) Verify(plain, check string) bool {
	if n.key == nil || n.key.pgID == "" {
		return false
	} else if len(check) != 256 {
		return false
	}
	hb := n.sha1_128(plain)
	hbhex := strings.ToUpper(bin2hex(hb))
	rbhex := n.rsaDecrypt(check)
	return rbhex == hbhex
}

func GetSignStr(params map[string]string) string {
	keys := []string{}
	for k := range params {
		keys = append(keys, k)
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
