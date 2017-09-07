package crypto

import "crypto/md5"

//MD5 MD5编码
func MD5(str []byte) []byte {
	md5code := md5.New()
	md5code.Write(str)
	return md5code.Sum(nil)[:]
}
