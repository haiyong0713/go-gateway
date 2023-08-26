package recommend

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-job/job/conf"
	"go-gateway/app/app-svr/app-job/job/model/recommend"

	"github.com/pkg/errors"
)

const (
	_rcmd       = "/recommand"
	_schoolRcmd = "/data/rank/reco-campus-nearby-zb.json"
)

// Dao is recommend dao.
type Dao struct {
	client     *httpx.Client
	rcmd       string
	schoolRcmd string
	c          *conf.Config
}

// New recommend dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		client:     httpx.NewClient(c.HTTPClient),
		rcmd:       c.Host.Data + _rcmd,
		schoolRcmd: c.Host.Data + _schoolRcmd,
	}
	return
}

// Recommend list
func (d *Dao) Recommend(c context.Context) (rs map[int64]string, err error) {
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
		err = errors.Wrapf(ecode.Int(res.Code), "recommend url(%s) code(%d)", d.rcmd, res.Code)
		return
	}
	rs = make(map[int64]string)
	for _, l := range res.Data {
		if l.Id > 0 {
			rs[l.Id] = ""
		}
	}
	return
}

func (d *Dao) SchoolRcmd(c context.Context) ([]*recommend.Item, error) {
	var res struct {
		Code int               `json:"code"`
		Data []*recommend.Item `json:"data"`
	}
	if err := d.client.Get(c, d.schoolRcmd, "", nil, &res); err != nil {
		return nil, err
	}
	if code := ecode.Int(res.Code); !code.Equal(ecode.OK) {
		return nil, errors.Wrap(code, d.schoolRcmd)
	}
	return res.Data, nil
}
