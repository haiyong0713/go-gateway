package filter

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/client"
	"go-gateway/app/web-svr/activity/interface/conf"

	fligrpc "git.bilibili.co/bapis/bapis-go/filter/service"
)

// Dao is filter dao
type Dao struct {
	filterClient fligrpc.FilterClient
}

// New is
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	d.filterClient = client.FilterClient
	return
}

// ActFilter .
func (d *Dao) ActFilter(ctx context.Context, area, msg string, keys []string, level int32) (res bool, err error) {
	var filRly *fligrpc.FilterReply
	arg := &fligrpc.FilterReq{
		Area:    area,
		Keys:    keys,
		Message: msg,
	}
	if filRly, err = d.filterClient.Filter(ctx, arg); err != nil {
		log.Errorc(ctx, "ActFilter s.fliClient.Filter(%+v) error(%+v)", arg, err)
		return
	}
	if filRly != nil && filRly.Level >= level {
		log.Infoc(ctx, "ActFilter s.fliClient.Filter(%+v) level(%d)", arg, filRly.Level)
		return true, nil
	}
	return
}
