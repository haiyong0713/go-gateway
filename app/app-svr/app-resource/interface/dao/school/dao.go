package school

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/app-resource/interface/conf"
)

const (
	_school = "/api/v1/home/tab"
)

type Dao struct {
	c          *conf.Config
	httpClient *bm.Client
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:          c,
		httpClient: bm.NewClient(c.HTTPClient),
	}
	return
}

func (d *Dao) ChangeSchoolTabPosition(ctx context.Context, mid int64) bool {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(mid, 10))
	params.Set("just_position", "1")
	var res = struct {
		Code int64 `json:"code"`
		Data struct {
			PositionType int64 `json:"position_type"`
		} `json:"data"`
	}{}
	schoolUrl := d.c.Host.School + _school
	if err := d.httpClient.Get(ctx, schoolUrl, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		log.Error("ChangeSchoolTabPosition error(%+v) url(%s), mid(%d)", err, schoolUrl, mid)
		return false
	}
	if res.Code != 0 {
		log.Error("ChangeSchoolTabPosition code(%+v) url(%s), mid(%d)", res.Code, schoolUrl, mid)
		return false
	}
	return res.Data.PositionType == 1
}
