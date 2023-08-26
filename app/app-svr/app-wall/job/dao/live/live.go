package live

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-wall/job/conf"

	"github.com/pkg/errors"
)

const (
	_addVipURL = "/user/v0/Vip/addVip"
)

// Dao is live dao
type Dao struct {
	client    *httpx.Client
	addVipURL string
}

// New live dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:    httpx.NewClient(c.HTTPClient),
		addVipURL: c.Host.APILive + _addVipURL,
	}
	return
}

func (d *Dao) AddVIP(c context.Context, mid int64, day int) (msg string, err error) {
	params := url.Values{}
	params.Set("vip_type", "1")
	params.Set("day", strconv.Itoa(day))
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("platform", "main")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"message"`
	}
	if err = d.client.Post(c, d.addVipURL, "", params, &res); err != nil {
		err = errors.Wrap(err, d.addVipURL+"?"+params.Encode())
		return "", err
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.addVipURL+"?"+params.Encode())
		return res.Msg, err
	}
	return res.Msg, nil
}
