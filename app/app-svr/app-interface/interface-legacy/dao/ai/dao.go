package ai

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	infocv2 "go-common/library/log/infoc.v2"
	"go-common/library/naming/discovery"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	"github.com/pkg/errors"
)

const (
	_rcmd = "/recommand"
)

// Dao is recommend dao.
type Dao struct {
	client  *httpx.Client
	rcmd    string
	rcmdTag string
	c       *conf.Config
	infocv2 infocv2.Infoc
}

// New recommend dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:       c,
		client:  httpx.NewClient(c.HTTPClient, httpx.SetResolver(resolver.New(nil, discovery.Builder()))),
		rcmd:    c.HostDiscovery.Data + _rcmd,
		rcmdTag: c.Host.Data + _rcmd,
		infocv2: c.Infocv2,
	}
	return
}

// Recommend list
func (d *Dao) Recommend(c context.Context) (rs map[int64]struct{}, err error) {
	params := url.Values{}
	params.Set("cmd", "hot")
	params.Set("from", "10")
	timeout := time.Duration(d.c.Custom.RecommendTimeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("ignore_custom", "1")
	var res struct {
		Code int `json:"code"`
		Data []struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.rcmd, "", params, &res); err != nil {
		return
	}
	if res.Code != 0 {
		// err = errors.Wrap(err, fmt.Sprintf("code(%d)", res.Code))
		err = errors.Wrapf(ecode.Int(res.Code), "recommend url(%s) code(%d)", d.rcmd, res.Code)
		return
	}
	rs = map[int64]struct{}{}
	for _, l := range res.Data {
		if l.Id > 0 {
			rs[l.Id] = struct{}{}
		}
	}
	return
}
