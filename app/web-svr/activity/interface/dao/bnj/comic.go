package bnj

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"

	"github.com/pkg/errors"
)

const _comicCouponURI = "/twirp/coupon.v0.Coupon/SendNewYearCoupon"

// ComicCoupon .
func (d *Dao) ComicCoupon(c context.Context, mid, num int64) (err error) {
	var (
		bs  []byte
		req *http.Request
	)
	param := &struct {
		Uid    int64 `json:"uid"`
		Amount int64 `json:"amount"`
	}{
		Uid:    mid,
		Amount: num,
	}
	if bs, err = json.Marshal(param); err != nil {
		return
	}
	params := url.Values{}
	params.Set("ts", strconv.FormatInt(time.Now().Unix(), 10))
	params.Set("appkey", d.c.HTTPClientComic.Key)
	mh := md5.Sum([]byte(params.Encode() + d.c.HTTPClientComic.Secret))
	params.Set("sign", hex.EncodeToString(mh[:]))
	if req, err = http.NewRequest(http.MethodPost, d.comicCouponURL+"?"+params.Encode(), bytes.NewReader(bs)); err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err = d.comicClient.Do(c, req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.comicCouponURL+"?"+string(bs)+"msg:"+res.Msg)
	}
	return
}
