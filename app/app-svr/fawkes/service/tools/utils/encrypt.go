package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

var keyArr = []string{
	"BILIBILIBOY0BOYO",
	"BlL1BILIB0YOBOYO",
	"BILlBILlBOYOBOY0",
	"BlL1BILlBOYOB0YO",
	"BILIBILIBOYOB0Y0",
	"BlL1B1LIB0YOBOYO",
	"BILlBIL1BOYOBOY0",
	"BlLIBILlB0YOBOY0",
	"B1LIBILIBOY0B0YO",
	"BILIBIL1B0Y0B0Y0",
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func aesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func aesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func base64Encode(src []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(src))
}

func base64Decode(src []byte) ([]byte, error) {
	return base64.StdEncoding.DecodeString(string(src))
}

// FawkesEncode encode
func FawkesEncode(org string) string {
	text := []byte(org)
	//nolint:gomnd
	for i := 1; i < 10; i++ {
		mod := (i*i - 1) % 10
		//nolint:gomnd
		if mod == 4 {
			salt := []byte("12")
			text = append(text, salt[0])
			text = append(text, salt[1])
			text = base64Encode(text)
		} else {
			key := []byte(keyArr[mod])
			text, _ = aesEncrypt(text, key)
		}
	}
	text = base64Encode(text)
	return string(text)
}

// FawkesDecode decode
func FawkesDecode(org string) string {
	text := []byte(org)
	text, _ = base64Decode(text)
	//nolint:gomnd
	for i := 1; i < 10; i++ {
		mod := (i*i - 1) % 10
		//nolint:gomnd
		if mod == 4 {
			text, _ = base64Decode(text)
			text = text[0 : len(text)-2]
		} else {
			key := []byte(keyArr[mod])
			text, _ = aesDecrypt(text, key)
		}
	}
	return string(text)
}
