package ai

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	"github.com/pkg/errors"
)

type Dao struct {
	c      *conf.Config
	client *bm.Client
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: bm.NewClient(c.HTTPClient),
	}
	return
}

const (
	_rcmd = "/recommand"
)

// Recommend list
func (d *Dao) Recommend(ctx context.Context) (map[int64]struct{}, error) {
	params := url.Values{}
	params.Set("cmd", "hot")
	params.Set("from", "10")
	timeout := time.Duration(d.c.RecommendTimeout) / time.Millisecond
	params.Set("timeout", strconv.FormatInt(int64(timeout), 10))
	params.Set("ignore_custom", "1")
	var res struct {
		Code int `json:"code"`
		Data []struct {
			Id int64 `json:"id"`
		} `json:"data"`
	}
	if err := d.client.Get(ctx, d.c.Hosts.Data+_rcmd, "", params, &res); err != nil {
		return nil, errors.WithStack(err)
	}
	if res.Code != 0 {
		err := errors.Wrapf(ecode.Int(res.Code), "recommend url: %v, code: %v", d.c.Hosts.Data+_rcmd, res.Code)
		return nil, err
	}
	ret := map[int64]struct{}{}
	for _, l := range res.Data {
		if l.Id > 0 {
			ret[l.Id] = struct{}{}
		}
	}
	return ret, nil
}
