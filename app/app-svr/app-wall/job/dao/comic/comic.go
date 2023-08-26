package comic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-wall/job/conf"

	"github.com/pkg/errors"
)

const (
	_comicUser   = "/twirp/manage.v0.Manage/IsComicUser"
	_comicCoupon = "/twirp/manage.v0.Manage/SendUniformCoupon"
)

// Dao is shopping dao
type Dao struct {
	client         *httpx.Client
	comicUserURL   string
	comicCouponURL string
}

// New shopping dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:         httpx.NewClient(c.HTTPClient),
		comicUserURL:   c.Host.Comic + _comicUser,
		comicCouponURL: c.Host.Comic + _comicCoupon,
	}
	return
}

func (d *Dao) Coupon(c context.Context, mid int64, amount int) (msg string, err error) {
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	data := map[string]interface{}{
		"uid":    mid,
		"amount": amount, // 1 coupon for 50 points
	}
	bytesData, err := json.Marshal(data)
	if err != nil {
		err = errors.Wrap(err, string(bytesData))
		return "", err
	}
	req, err := http.NewRequest("POST", d.comicCouponURL, bytes.NewReader(bytesData))
	if err != nil {
		err = errors.Wrap(err, d.comicCouponURL+"?"+string(bytesData))
		return "", err
	}
	req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if err := d.client.Do(c, req, &res); err != nil {
		err = errors.Wrap(err, d.comicCouponURL+"?"+string(bytesData))
		return "", err
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.comicCouponURL+"?"+string(bytesData))
		return res.Msg, err
	}
	return res.Msg, nil
}
