package bvav

import (
	"fmt"
	"strconv"
	"strings"

	"go-gateway/pkg/idsafe/bvid"
)

// AvsStrToAvsIntSlice .
func AvsStrToAvsIntSlice(values []string) (res []int64, err error) {
	for _, v := range values {
		if strings.HasPrefix(v, "bv") {
			return nil, fmt.Errorf("bv id 需要全部大写")
		}
		var (
			avid int64
		)
		if strings.HasPrefix(v, "BV") {
			if avid, err = bvid.BvToAv(v); err != nil {
				return
			}
			res = append(res, avid)
		} else {
			if avid, err = strconv.ParseInt(v, 10, 64); err != nil {
				return
			}
			res = append(res, avid)
		}
	}
	return res, nil
}

// ToAvStr .
func ToAvStr(value string) (res string, err error) {
	if strings.HasPrefix(value, "bv") {
		return "", fmt.Errorf("bv id 需要全部大写")
	}
	if strings.HasPrefix(value, "BV") {
		var (
			id int64
		)
		if id, err = bvid.BvToAv(value); err != nil {
			return
		}
		return strconv.FormatInt(id, 10), nil
	}
	return value, nil
}

// ToAvsStr .
func ToAvsStr(value string) (res string, err error) {
	var tmpRes []string
	values := strings.Split(value, ",")
	for _, v := range values {
		if strings.HasPrefix(v, "bv") {
			return "", fmt.Errorf("bv id 需要全部大写")
		}
		var (
			avid int64
		)
		if strings.HasPrefix(v, "BV") {
			if avid, err = bvid.BvToAv(v); err != nil {
				return
			}
			avidStr := strconv.FormatInt(avid, 10)
			tmpRes = append(tmpRes, avidStr)
		} else {
			tmpRes = append(tmpRes, v)
		}
	}
	return strings.Join(tmpRes, ","), nil
}

func EditToBvsStr(value string) (res string, err error) {
	var (
		tmpRes  []string
		bvidStr string
	)
	values := strings.Split(value, ",")
	for _, v := range values {
		var (
			avid int64
		)
		if avid, err = strconv.ParseInt(v, 10, 64); err != nil {
			return
		}
		if bvidStr, err = bvid.AvToBv(avid); err != nil {
			return
		}
		tmpRes = append(tmpRes, bvidStr)
	}
	return strings.Join(tmpRes, ","), nil
}
