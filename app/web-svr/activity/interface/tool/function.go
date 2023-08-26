package tool

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	r "crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"io"
	"math"
	"math/rand"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"
)

func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func RandStringRunes(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func SHA1(data string) string {
	t := sha1.New()
	io.WriteString(t, data)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func CFBDecrypt(str string, secKey string) string {
	key, _ := hex.DecodeString(secKey)
	ciphertext, _ := hex.DecodeString(str)
	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}
	if len(ciphertext) < aes.BlockSize {
		return ""
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return fmt.Sprintf("%s", ciphertext)
}

func CFBEncrypt(str string, secKey string) string {
	key, _ := hex.DecodeString(secKey)
	plaintext := []byte(str)
	block, err := aes.NewCipher(key)
	if err != nil {
		return ""
	}
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(r.Reader, iv); err != nil {
		return ""
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return fmt.Sprintf("%x", ciphertext)
}

func InStrSlice(find string, set []string) bool {
	for _, v := range set {
		if find == v {
			return true
		}
	}
	return false
}

func Now() int64 {
	return time.Now().Unix()
}

func InInt64Slice(find int64, set []int64) bool {
	for _, v := range set {
		if find == v {
			return true
		}
	}
	return false
}

func SeparatorJoin(set []int64, base string, sep string) (res string) {
	res = ""
	isFirst := true
	for i := 0; i < len(set); i++ {
		if isFirst {
			res += base
			isFirst = false
		} else {
			res += sep + base
		}
	}
	return
}

func SimpleCopyProperties(dst, src interface{}, propertys []string, tag string) (err error) {
	// 防止意外panic
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()

	dstType, dstValue := reflect.TypeOf(dst), reflect.ValueOf(dst)
	srcType, srcValue := reflect.TypeOf(src), reflect.ValueOf(src)

	// dst必须结构体指针类型
	if dstType.Kind() != reflect.Ptr || dstType.Elem().Kind() != reflect.Struct {
		return errors.New("dst type should be a struct pointer")
	}

	// src必须为结构体或者结构体指针
	if srcType.Kind() == reflect.Ptr {
		srcType, srcValue = srcType.Elem(), srcValue.Elem()
	}
	if srcType.Kind() != reflect.Struct {
		return errors.New("src type should be a struct or a struct pointer")
	}
	// 取具体内容
	dstType, dstValue = dstType.Elem(), dstValue.Elem()

	// 属性个数
	checkMap := make(map[string]struct{})
	for _, v := range propertys {
		checkMap[v] = struct{}{}
	}
	propertyNums := dstType.NumField()
	for i := 0; i < propertyNums; i++ {
		// 属性
		property := dstType.Field(i)
		if _, ok := checkMap[property.Tag.Get(tag)]; !ok {
			continue
		}
		// 待填充属性值
		propertyValue := srcValue.FieldByName(property.Name)
		// 无效，说明src没有这个属性 || 属性同名但类型不同
		if !propertyValue.IsValid() || property.Type != propertyValue.Type() {
			continue
		}

		if dstValue.Field(i).CanSet() {
			dstValue.Field(i).Set(propertyValue)
		}
	}

	return nil
}

func Decimal(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

func TencentMd5Sign(signKey string, params map[string]string) string {

	keyArr := make([]string, 0)
	for k := range params {
		keyArr = append(keyArr, k)
	}
	sort.Strings(keyArr)
	signStr := ""
	for _, key := range keyArr {
		for k, v := range params {
			if key == k {
				signStr = fmt.Sprintf("%v%v+", signStr, url.QueryEscape(v))
			}
		}
	}
	signStr = fmt.Sprintf("%v%v", signStr, signKey)
	h := md5.New()
	h.Write([]byte(signStr))
	newStr := h.Sum(nil)
	signStr = fmt.Sprintf("%X", newStr)
	return strings.ToLower(signStr)
}

func TencentSignCheck(ctx context.Context, appKey, secret, sign string, params map[string]string) (err error) {
	keyArr := make([]string, 0)
	for k := range params {
		keyArr = append(keyArr, k)
	}

	sort.Strings(keyArr)
	var signStr string
	for _, key := range keyArr {
		for k, v := range params {
			if key == k {
				signStr = fmt.Sprintf("%v%v%v", signStr, k, v)
				params[k] = v
			}
		}
	}
	signStrNoSecret := signStr
	signStr = fmt.Sprintf("%v%v%v", secret, signStr, secret)
	signBorn := sha256Sign(signStr, secret)
	log.Infov(ctx, log.KV("INFO", fmt.Sprintf("[Info][Sign][Check] info, checkPass:%+v", signBorn == sign)),
		log.KVString("signStr", fmt.Sprintf("%v%v%v", "{secret}", signStrNoSecret, "{secret}")),
		log.KVString("externalSign", sign),
		log.KVString("sign", signBorn),
	)
	if signBorn != sign {
		err = xecode.Errorf(xecode.RequestErr, "sign校验未通过")
		return
	}
	return
}

func sha256Sign(signString string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(signString))
	return hex.EncodeToString(h.Sum(nil))
}
