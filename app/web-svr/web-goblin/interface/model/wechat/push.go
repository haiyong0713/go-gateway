package wechat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// Msg wechat send data struct.
type Msg struct {
	ToUserName   string `json:"ToUserName"`
	FromUserName string `json:"FromUserName"`
	CreateTime   int64  `json:"CreateTime"`
	MsgType      string `json:"MsgType"`
	Content      string `json:"Content"`
	Encrypt      string `json:"Encrypt"`
}

// SendMsg .
type SendMsg struct {
	Touser  string   `json:"touser"`
	Msgtype string   `json:"msgtype"`
	Link    *LinkMsg `json:"link"`
}

// LinkMsg .
type LinkMsg struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	ThumbURL    string `json:"thumb_url"`
}

// PushArg push argument.
type PushArg struct {
	EncryptType  string `form:"encrypt_type"`
	MsgSignature string `form:"msg_signature"`
	Nonce        string `form:"nonce"`
	Openid       string `form:"openid"`
	Signature    string `form:"signature" validate:"min=1"`
	Timestamp    int64  `form:"timestamp" validate:"min=1"`
	Echostr      string `form:"echostr"`
}

// Encrypt .
func Encrypt(origData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// PKCS7Padding padding.
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
