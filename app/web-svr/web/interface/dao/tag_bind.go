package dao

import (
	"context"
	"math"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"

	"github.com/pkg/errors"
)

func (d *Dao) TagBind(ctx context.Context, oidList []int64) ([]int64, error) {
	var bindOidList []int64
	onceLimit := 30
	getLen := int(math.Ceil(float64(len(oidList)) / float64(onceLimit)))
	for i := 0; i < getLen; i++ {
		startIndex := i * onceLimit
		endIndex := (i + 1) * onceLimit
		if endIndex > len(oidList) {
			endIndex = len(oidList)
		}
		splitOidList := oidList[startIndex:endIndex]
		res, err := d.tagBindOnce(ctx, splitOidList)
		if err != nil {
			return nil, err
		}
		if res != nil {
			bindOidList = append(bindOidList, res...)
		}
	}
	return bindOidList, nil
}

func (d *Dao) tagBindOnce(ctx context.Context, oidList []int64) ([]int64, error) {
	oidListS := ""
	for i, n := range oidList {
		if i > 0 {
			oidListS += ","
		}
		oidListS += strconv.FormatInt(n, 10)
	}
	params := url.Values{}
	params.Set("from", "WechatApplet")
	params.Set("oids", oidListS)
	params.Set("format", "4")
	req, err := d.httpR.NewRequest("GET", d.tagBindURL, "", params)
	if err != nil {
		return nil, err
	}
	var res struct {
		Code int             `json:"code"`
		Data map[string]bool `json:"data"`
	}
	if err := d.httpR.Do(ctx, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.tagBindURL+"?"+params.Encode())
		log.Warn("tag_bind tagBindOnce error(%v)", err)
		return nil, err
	}
	var bindOidList []int64
	if res.Data != nil {
		for k, v := range res.Data {
			if i, err := strconv.ParseInt(k, 10, 64); err == nil && v {
				bindOidList = append(bindOidList, i)
			}
		}
	}
	return bindOidList, nil
}
