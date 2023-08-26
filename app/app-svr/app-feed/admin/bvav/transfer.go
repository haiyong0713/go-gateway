package bvav

import (
	"fmt"
	"strconv"
	"strings"

	"go-gateway/pkg/idsafe/bvid"
)

// AvStrToBvStr .
func AvStrToBvStr(value string) (res string, err error) {
	var (
		avid int64
	)
	if avid, err = strconv.ParseInt(value, 10, 64); err != nil {
		return
	}
	if res, err = bvid.AvToBv(avid); err != nil {
		return
	}
	return
}

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

// ToAvInt .
func ToAvInt(value string) (res int64, err error) {
	if strings.HasPrefix(value, "bv") {
		return 0, fmt.Errorf("bv id 需要全部大写")
	}
	if strings.HasPrefix(value, "BV") {
		return bvid.BvToAv(value)
	}
	return strconv.ParseInt(value, 10, 64)
}

// ToAvBvStr .
func ToBvStr(value string) (res string, err error) {
	var (
		avid      int64
		avidStr   string
		valuesTmp []string
	)
	values := strings.Split(value, ",")
	for _, v := range values {
		if avid, err = strconv.ParseInt(v, 10, 64); err != nil {
			return
		}
		if avidStr, err = bvid.AvToBv(avid); err != nil {
			return
		}
		valuesTmp = append(valuesTmp, avidStr)
	}
	return strings.Join(valuesTmp, ","), nil
}
