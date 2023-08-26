package shop

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/shop"

	"github.com/pkg/errors"
)

const (
	_newShop = "/mall-shop/merchant/enter/service/shop/info"
)

type Dao struct {
	client  *httpx.Client
	newShop string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:  httpx.NewClient(c.HTTPClient),
		newShop: c.Host.Mall + _newShop,
	}
	return
}

func (d *Dao) Info(c context.Context, mid int64, mobiApp, device string, build int) (info *shop.Info, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("mobi_app", mobiApp)
	params.Set("device", device)
	params.Set("build", strconv.Itoa(build))
	params.Set("type", "1")
	var res struct {
		Code int        `json:"code"`
		Data *shop.Info `json:"data"`
	}
	if err = d.client.Get(context.Background(), d.newShop, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		//nolint:gomnd
		if res.Code == 130000 {
			return
		}
		err = errors.Wrap(ecode.Int(res.Code), d.newShop+"?"+params.Encode())
		return
	}
	info = res.Data
	return
}
