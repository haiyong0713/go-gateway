package lottery

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	xhttp "net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	lotmdl "go-gateway/app/web-svr/activity/admin/model/lottery"

	"github.com/pkg/errors"
)

// GetAddressByID http get call /api/basecenter/addr/view, get address information by addrID
func (d *Dao) GetAddressByID(c context.Context, id int64, uid int) (addr *lotmdl.Address, err error) {
	var (
		params = url.Values{}
		res    struct {
			ErrCode int            `json:"errcode"`
			ErrTag  int            `json:"tag"`
			ErrMsg  string         `json:"msg"`
			Data    lotmdl.Address `json:"data"`
		}
	)
	u := d.addrDetailURL
	params.Set("id", strconv.FormatInt(id, 10))
	params.Set("uid", strconv.Itoa(uid))
	params.Set("app_id", d.c.Lottery.AppKey)
	params.Set("app_token", d.c.Lottery.AppToken)

	if err = d.httpClient.Get(c, u, "", params, &res); err != nil {
		log.Errorc(c, "GetAddressByID d.httpClient.Get() failed. error(%v)", err)
		return
	}
	if res.ErrCode != 0 {
		err = errors.Errorf("GetAddressByID: errcode: %d, errmsg: %s", res.ErrCode, res.ErrMsg)
		return
	}
	addr = &res.Data
	return
}

// GetVIPInfo http get call /x/admin/vip/act/info, get vip information by id
func (d *Dao) GetVIPInfo(c context.Context, id, cookie string) (info bool, err error) {
	var (
		params = url.Values{}
		res    struct {
			ErrCode int    `json:"code"`
			ErrMsg  string `json:"message"`
			Data    bool   `json:"data"`
		}
		req   *xhttp.Request
		en, u string
	)
	en = d.sign(&params, "token", id)
	if en != "" {
		u += d.vipInfoURL + "?" + en
	}
	if req, err = xhttp.NewRequest(xhttp.MethodGet, u, nil); err != nil {
		log.Errorc(c, "GetVIPInfo req error(%v)", err)
		return
	}
	req.Body = ioutil.NopCloser(strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookie)

	if err = d.httpClient.Do(c, req, &res); err != nil {
		log.Errorc(c, "GetVIPInfo d.httpClient.Do() failed. error(%v)", err)
		return
	}
	if res.ErrCode != 0 {
		err = errors.Errorf("GetVIPInfo: errcode: %v, errmsg: %v", res.ErrCode, res.ErrMsg)
		return
	}
	info = res.Data
	return
}

// GetCouponInfo .
func (d *Dao) GetCouponInfo(c context.Context, token, cookie string) (couponInfo *lotmdl.CouponInfo, err error) {
	var (
		params = url.Values{}
		res    struct {
			ErrCode int               `json:"code"`
			ErrMsg  string            `json:"message"`
			Data    lotmdl.CouponInfo `json:"data"`
		}
		req   *xhttp.Request
		en, u string
	)
	params.Set("batch_token", token)
	if en = params.Encode(); en != "" {
		u += d.couponInfoURL + "?" + en
	}
	if req, err = xhttp.NewRequest(xhttp.MethodGet, u, nil); err != nil {
		log.Errorc(c, "GetCouponInfo req error(%v)", err)
		return
	}
	req.Body = ioutil.NopCloser(strings.NewReader(params.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookie)
	if err = d.httpClient.Do(c, req, &res); err != nil {
		log.Errorc(c, "GetCouponInfo d.httpClient.Do() failed. error(%v)", err)
		return
	}
	if res.ErrCode != 0 {
		if res.ErrCode == -404 {
			err = nil
			return
		}
		err = errors.Errorf("GetCouponInfo: errcode: %v, errmsg: %v", res.ErrCode, res.ErrMsg)
		return
	}
	couponInfo = &res.Data
	return
}

func (d *Dao) sign(params *url.Values, key, value string) (query string) {

	if params == nil {
		params = &url.Values{}
	}
	params.Set("appkey", d.c.HTTPClient.Key)
	params.Set("ts", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set(key, value)
	tmp := params.Encode()
	var b bytes.Buffer
	b.WriteString(tmp)
	b.WriteString(d.c.HTTPClient.Secret)
	mh := md5.Sum(b.Bytes())
	// query
	var qb bytes.Buffer
	qb.WriteString(tmp)
	qb.WriteString("&sign=")
	qb.WriteString(hex.EncodeToString(mh[:]))
	query = qb.String()
	return
}
