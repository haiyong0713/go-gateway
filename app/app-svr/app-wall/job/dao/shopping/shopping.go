package shopping

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"go-common/library/ecode"
	"go-common/library/naming/discovery"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	unicomdl "go-gateway/app/app-svr/app-wall/interface/model/unicom"
	"go-gateway/app/app-svr/app-wall/job/conf"

	"github.com/pkg/errors"
)

const (
	_couponURL   = "/mall-marketing/coupon_code/create"
	_couponV2URL = "/asset/release_asset"
)

// Dao is shopping dao
type Dao struct {
	client      *httpx.Client
	couponURL   string
	couponV2URL string
	// redis
}

// New shopping dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client: httpx.NewClient(c.HTTPClient, httpx.SetResolver(resolver.New(nil, discovery.Builder()))),
		// url
		couponURL:   c.Host.Mall + _couponURL,
		couponV2URL: c.Host.MallDiscovery + _couponV2URL,
	}
	return
}

// Coupon user vip
func (d *Dao) Coupon(c context.Context, couponID string, mid int64, uname string) (msg string, err error) {
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	data := map[string]interface{}{
		"couponId": couponID,
		"mid":      mid,
		"uname":    uname,
	}
	var (
		bytesData []byte
		req       *http.Request
	)
	if bytesData, err = json.Marshal(data); err != nil {
		err = errors.Wrapf(err, "%v", data)
		return "", err
	}
	if req, err = http.NewRequest("POST", d.couponURL, bytes.NewReader(bytesData)); err != nil {
		err = errors.Wrap(err, d.couponURL+"?"+string(bytesData))
		return "", err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	req.Header.Set("X-BACKEND-BILI-REAL-IP", "")
	if err = d.client.Do(c, req, &res); err != nil {
		err = errors.Wrap(err, d.couponURL+"?"+string(bytesData))
		return "", err
	}
	if res.Code != 0 {
		//   "code": 83110020,
		//   "message": "优惠券领取时间已超过有效期"
		err = errors.Wrap(ecode.Int(res.Code), d.couponURL+"?"+string(bytesData))
		return res.Msg, err
	}
	return res.Msg, nil
}

// Coupon user vip
// nolint:gomnd
func (d *Dao) CouponV2(c context.Context, data *unicomdl.CouponParam) (string, error) {
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	bytesData, err := json.Marshal(data)
	if err != nil {
		return "", errors.Wrapf(err, "%+v", data)
	}
	req, err := http.NewRequest("POST", d.couponV2URL, bytes.NewReader(bytesData))
	if err != nil {
		return "", errors.Wrap(err, d.couponV2URL+"?"+string(bytesData))
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if err = d.client.Do(c, req, &res); err != nil {
		return "", errors.Wrap(err, d.couponV2URL+"?"+string(bytesData))
	}
	if res.Code != 0 {
		return res.Msg, errors.Wrap(ecode.Int(res.Code), d.couponV2URL+"?"+string(bytesData))
	}
	return res.Msg, nil
}
