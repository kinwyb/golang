package crypto

//加密解密工具

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
)

//AESCrypto AES加密
type AESCrypto struct{}

//CBCEncrypt AesCBC加密PKCS5
//
//@param origData []byte 加密的字节数组
//
//@param key []byte 密钥字节数组
func (a *AESCrypto) CBCEncrypt(origData []byte, key []byte) ([]byte, error) {
	srckey := MD5(key)
	block, err := aes.NewCipher(srckey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = a.PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, srckey[:blockSize])
	cypted := make([]byte, len(origData))
	blockMode.CryptBlocks(cypted, origData)
	return []byte(hex.EncodeToString(cypted)), nil
}

//ECBEncrypt Aes ECB加密PKCS5
//
//@param origData []byte 加密的字节数组
//
//@param key []byte 密钥字节数组
func (a *AESCrypto) ECBEncrypt(origData []byte, key []byte) ([]byte, error) {
	srckey := MD5(key)
	block, err := aes.NewCipher(srckey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = a.PKCS5Padding(origData, blockSize)
	blockMode := newECBEncrypter(block)
	cypted := make([]byte, len(origData))
	blockMode.CryptBlocks(cypted, origData)
	return []byte(hex.EncodeToString(cypted)), nil
}

//CBCDecrypt Aes CBC解密PKCS5
//
//@param origData []byte 解密的字节数组
//
//@param key []byte 密钥字节数组
func (a *AESCrypto) CBCDecrypt(cypted []byte, key []byte) ([]byte, error) {
	cypted, err := hex.DecodeString(string(cypted))
	if err != nil {
		return nil, err
	}
	srckey := MD5(key)
	block, err := aes.NewCipher(srckey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, srckey[:blockSize])
	origData := make([]byte, len(cypted))
	blockMode.CryptBlocks(origData, cypted)
	origData = a.PKCS5UnPadding(origData)
	return origData, nil
}

//ECBDecrypt Aes ECB解密PKCS5
//
//@param origData []byte 解密的字节数组
//
//@param key []byte 密钥字节数组
func (a *AESCrypto) ECBDecrypt(cypted []byte, key []byte) ([]byte, error) {
	cypted, err := hex.DecodeString(string(cypted))
	if err != nil {
		return nil, err
	}
	srckey := MD5(key)
	block, err := aes.NewCipher(srckey)
	if err != nil {
		return nil, err
	}
	blockMode := newECBDecrypter(block)
	origData := make([]byte, len(cypted))
	blockMode.CryptBlocks(origData, cypted)
	origData = a.PKCS5UnPadding(origData)
	return origData, nil
}

//PKCS5Padding PKCS5密钥填充方式
func (a *AESCrypto) PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//PKCS5UnPadding PKCS5密钥填充方式返填充
func (a *AESCrypto) PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
