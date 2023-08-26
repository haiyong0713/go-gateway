package utils

import (
	"crypto/hmac"
	"crypto/md5"

	// nolint:gosec
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// MD5Hash MD5哈希值
func MD5Hash(b []byte) string {
	h := md5.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// MD5HashString MD5哈希值
func MD5HashString(s string) string {
	return MD5Hash([]byte(s))
}

// SHA1Hash SHA1哈希值
func SHA1Hash(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// SHA1HashString SHA1哈希值
func SHA1HashString(s string) string {
	return SHA1Hash([]byte(s))
}

func Sha256Hex(s string) string {
	b := sha256.Sum256([]byte(s))
	return hex.EncodeToString(b[:])
}

func HmacSha256(s, key string) string {
	hashed := hmac.New(sha256.New, []byte(key))
	hashed.Write([]byte(s))
	return string(hashed.Sum(nil))
}

// ---------- File handler ----------
func GetFileMD5(filePath string) (md5S string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	h := md5.New()
	_, _ = io.Copy(h, file)
	return hex.EncodeToString(h.Sum(nil)), err
}

func GetFileSHA1(filePath string) (sha1S string, err error) {
	pFile, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer pFile.Close()
	sha1h := sha1.New()
	_, _ = io.Copy(sha1h, pFile)
	return hex.EncodeToString(sha1h.Sum(nil)), err
}
