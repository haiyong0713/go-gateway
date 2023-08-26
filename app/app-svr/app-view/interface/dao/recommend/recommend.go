package recommend

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/app-svr/app-view/interface/conf"

	"github.com/pkg/errors"
)

const (
	_rcmd = "/recommand"
)

// Dao is recommend dao.
type Dao struct {
	client *httpx.Client
	rcmd   string
	c      *conf.Config
}

// New recommend dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: httpx.NewClient(c.HTTPClient),
		rcmd:   c.Host.Data + _rcmd,
	}
	return
}

// Recommend list
func (d *Dao) Recommend(c context.Context) (rs map[int64]struct{}, unsorted []int64, err error) {
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
	unsorted = make([]int64, 0, len(res.Data))
	for _, l := range res.Data {
		if l.Id > 0 {
			rs[l.Id] = struct{}{}
			unsorted = append(unsorted, l.Id)
		}
	}
	return
}
