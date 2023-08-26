package ai

import (
	"context"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-view/interface/conf"
)

const (
	_av2GameURL = "/avid2gameid"
)

type Dao struct {
	client     *bm.Client
	av2GameURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:     bm.NewClient(c.HTTPGameAsync),
		av2GameURL: c.Host.AI + _av2GameURL,
	}
	return
}

func (d *Dao) Av2Game(c context.Context) (res map[int64]int64, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	if err = d.client.Get(c, d.av2GameURL, ip, nil, &res); err != nil {
		return
	}
	return
}
