package model

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	_xor      = 46852038695161
	_alphabet = "QWERTYUIOPASDFGHJKLZXCVBNM"
	_base     = 26
)

func TokenDecode(int int64) (str string) {
	var (
		tmpbv string
		ids   []string
	)
	tmp := ((1 << 51) | int) ^ _xor
	for tmp != 0 {
		tmpbv = fmt.Sprintf("%s%s", string(_alphabet[tmp%_base]), tmpbv)
		tmp = tmp / _base
	}
	ids = strings.Split(tmpbv, "")
	// 位置转换
	ids[6], ids[0] = ids[0], ids[6]
	ids[1], ids[4] = ids[4], ids[1]
	str = strings.Join(ids, "")
	return
}

func TokenEncode(str string) (int int64, err error) {
	// 位置转换
	tmpStr := strings.Split(str, "")
	tmpStr[4], tmpStr[1] = tmpStr[1], tmpStr[4]
	tmpStr[0], tmpStr[6] = tmpStr[6], tmpStr[0]
	str = strings.Join(tmpStr, "")
	var tmpNum int64
	for _, bs := range str {
		index := strings.Index(_alphabet, string(bs))
		if index == -1 {
			err = fmt.Errorf("str(%s) is illegal, invalid char:%c", str, bs)
			return
		}
		tmpNum = tmpNum*_base + int64(index)
	}
	if len(strconv.FormatInt(tmpNum, 2)) > 52 {
		err = fmt.Errorf("str is too big")
		return
	} else if len(strconv.FormatInt(tmpNum, 2)) < 52 {
		err = fmt.Errorf("str is too small")
		return
	}
	int = (tmpNum & (1<<51 - 1)) ^ _xor
	return
}
