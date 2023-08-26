package dao

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-common/library/ecode"

	"github.com/pkg/errors"
)

const (
	_bnjConfURI  = "/activity/v0/bainian/config"
	_bnj2020Conf = "/activity/v0/bainian2020/config"
)

// Bnj2019Conf .
func (d *Dao) Bnj2019Conf(c context.Context) (mids []int64, err error) {
	var res struct {
		Code int `json:"code"`
		Data struct {
			GreyStatus int    `json:"grey_status"`
			GreyUids   string `json:"grey_uids"`
		} `json:"data"`
	}
	if err = d.httpR.Get(c, d.bnjConfURL, "", nil, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.bnjConfURL)
		return
	}
	if res.Data.GreyStatus == 1 {
		midsStr := strings.Split(res.Data.GreyUids, ",")
		for _, midStr := range midsStr {
			var mid int64
			if mid, err = strconv.ParseInt(midStr, 10, 64); err != nil {
				err = fmt.Errorf("live grey_uids(%s)", res.Data.GreyUids)
				return
			}
			mids = append(mids, mid)
		}
	}
	return
}

// Bnj2019Conf .
func (d *Dao) Bnj2020Conf(c context.Context) (mids []int64, giftCount int64, err error) {
	var res struct {
		Code int `json:"code"`
		Data struct {
			GreyStatus    int    `json:"grey_status"`
			GreyUids      string `json:"grey_uids"`
			GiftUserCount int64  `json:"gift_user_count"`
		} `json:"data"`
	}
	if err = d.httpR.Get(c, d.bnj20ConfURL, "", nil, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.bnj20ConfURL)
		return
	}
	if res.Data.GreyStatus == 1 {
		midsStr := strings.Split(res.Data.GreyUids, ",")
		for _, midStr := range midsStr {
			var mid int64
			if mid, err = strconv.ParseInt(midStr, 10, 64); err != nil {
				err = fmt.Errorf("live grey_uids(%s)", res.Data.GreyUids)
				return
			}
			mids = append(mids, mid)
		}
	}
	giftCount = res.Data.GiftUserCount
	return
}
