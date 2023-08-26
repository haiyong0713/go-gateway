package util

import (
	"fmt"
	"strconv"
	"strings"
)

// avid/bvid编解码方法，将数字类型编码为定长字符串
// 为便于呈现效果设计，避免目标区间增大造成ID过长问题

const (
	_xorcode          = 23442827791579
	_maskcode         = 2251799813685247
	_maxAid           = int64(1 << 51)
	_maxAidBeforeMask = int64(1 << 52)
	_minAid           = int64(1)
	_base             = 58
	_prefix           = "BV1"
	_alphabet         = "FcwAPNKTMug3GV5Lj7EJnHpWsx4tb8haYeviqBz6rkCy12mUSDQX9RdoZf"
)

// AvToBv avid to bvid
func AvToBv(avid int64) (bvid string, err error) {
	if avid < _minAid {
		err = fmt.Errorf("avid lte zero")
		return
	}
	if avid >= _maxAid {
		err = fmt.Errorf("avid is too big")
		return
	}
	var (
		tmpbv string
		ids   []string
	)
	tmp := ((1 << 51) | avid) ^ _xorcode
	for tmp != 0 {
		tmpbv = fmt.Sprintf("%s%s", string(_alphabet[tmp%_base]), tmpbv)
		tmp = tmp / _base
	}
	ids = strings.Split(tmpbv, "")
	// 位置转换
	tmpstr := ids[0]
	ids[0] = ids[6]
	ids[6] = tmpstr
	tmpstr = ids[1]
	ids[1] = ids[4]
	ids[4] = tmpstr
	bvid = _prefix + strings.Join(ids, "")
	return
}

// BvToAv bvid to avid
func BvToAv(bvid string) (avid int64, err error) {
	if bvid == "" {
		err = fmt.Errorf("bvid is empty")
		return
	}
	if len([]rune(bvid)) == 10 && strings.Index(bvid, "1") == 0 {
		bvid = bvid[1:]
	} else if len([]rune(bvid)) > 3 && strings.Index(strings.ToUpper(bvid[0:3]), _prefix) == 0 {
		bvid = bvid[3:]
	}
	if len([]rune(bvid)) < 9 {
		err = fmt.Errorf("bvid is too small")
		return
	}
	// 位置转换
	tmpStr := strings.Split(bvid, "")
	tmp := tmpStr[1]
	tmpStr[1] = tmpStr[4]
	tmpStr[4] = tmp
	tmp = tmpStr[0]
	tmpStr[0] = tmpStr[6]
	tmpStr[6] = tmp
	bvid = strings.Join(tmpStr, "")
	var tmpNum int64
	for _, bs := range bvid {
		index := strings.Index(_alphabet, string(bs))
		if index == -1 {
			err = fmt.Errorf("bvid(%s) is illegal, invalid char:%c", bvid, bs)
			return
		}
		tmpNum = tmpNum*_base + int64(index)
	}
	if len(strconv.FormatInt(tmpNum, 2)) > 52 {
		err = fmt.Errorf("bvid is too big")
		return
	} else if len(strconv.FormatInt(tmpNum, 2)) < 52 {
		err = fmt.Errorf("bvid is too small")
		return
	}
	avid = (tmpNum & _maskcode) ^ _xorcode
	if avid < _minAid {
		err = fmt.Errorf("bvid is too small")
		return
	}
	return
}
