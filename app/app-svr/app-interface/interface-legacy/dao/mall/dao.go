package mall

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	mallmdl "go-gateway/app/app-svr/app-interface/interface-legacy/model/mall"

	"github.com/pkg/errors"
)

const (
	_favCount = "/mall-ugc/ugc/vote/user/wishcount"
	_shop     = "/merchant/enter/service/shop/get"
)

type Dao struct {
	client   *bm.Client
	favCount string
	shop     string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:   bm.NewClient(c.HTTPClient, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		favCount: c.Host.Mall + _favCount,
		shop:     c.HostDiscovery.Mall + _shop,
	}
	return
}

func (d *Dao) FavCount(c context.Context, mid int64) (count int32, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int   `json:"code"`
		Data int32 `json:"data"`
	}
	if err = d.client.Get(context.Background(), d.favCount, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.favCount+"?"+params.Encode())
		return
	}
	count = res.Data
	return
}

// Mall for space mall
func (d *Dao) Mall(c context.Context, mid int64) (st *mallmdl.Mall, err error) {
	var (
		req *http.Request
		ip  = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("type", "1")
	// new request
	if req, err = d.client.NewRequest("GET", d.shop, ip, params); err != nil {
		return
	}
	var res struct {
		Code int           `json:"code"`
		Data *mallmdl.Mall `json:"data"`
	}
	if err = d.client.Do(context.Background(), req, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.shop+"?"+params.Encode())
		return
	}
	st = res.Data
	return
}
