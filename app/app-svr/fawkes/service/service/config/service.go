package config

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"

	//paladinConf "go-common/library/conf"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

var (
// confClient *paladinConf.Client
// feConfig   string
)

// Service service struct info.
type Service struct {
	c      *conf.Config
	fkDao  *fawkes.Dao
	aesKey []byte
}

//func init() {
//var err error
//confClient, err = paladinConf.New()
//if err != nil {
//	panic(err)
//}
//if err = load(); err != nil {
//	panic(err)
//}
//reload()
//}

// New service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		fkDao: fawkes.New(c),
	}
	s.aesKey = []byte(s.c.Keys.AesKey)
	return
}

// AesEncrypt aes encrypt
func (s *Service) AesEncrypt(origData []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, s.aesKey[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// AesDecrypt aes encrypt
func (s *Service) AesDecrypt(crypted []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.aesKey)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, s.aesKey[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

// PKCS5Padding padding
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS5UnPadding unpadding
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// Ping dao.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.fkDao.Ping(c); err != nil {
		log.Error("s.dao error(%v)", err)
	}
	return
}

// Close dao.
func (s *Service) Close() {
	s.fkDao.Close()
}
