package workshop

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	"github.com/pkg/errors"
)

const _favCount = "/mall-up-search/items/fav/itemsCount/gateway"

type Dao struct {
	client   *bm.Client
	favCount string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:   bm.NewClient(c.HTTPClient),
		favCount: c.Host.Workshop + _favCount,
	}
	return
}

func (d *Dao) FavCount(c context.Context, mid int64) (int64, error) {

	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Count int64 `json:"count"`
		} `json:"data"`
	}
	if err := d.client.Get(c, d.favCount, ip, params, &res); err != nil {
		return 0, err
	}
	if res.Code != ecode.OK.Code() {
		return 0, errors.Wrap(ecode.Int(res.Code), d.favCount+"?"+params.Encode())
	}
	return res.Data.Count, nil
}
