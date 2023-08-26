package videoup

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"

	videogrpc "git.bilibili.co/bapis/bapis-go/videoup/open/service"
)

type Dao struct {
	c           *conf.Config
	videoupGRPC videogrpc.VideoUpOpenClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.videoupGRPC, err = videogrpc.NewClient(c.VideoupGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) ExtBiliCut(c context.Context, aids []int64) (map[int64]*videogrpc.DynamicView, error) {
	resTmp, err := d.videoupGRPC.ArcsDynamicView(c, &videogrpc.ArcsViewReq{Aids: aids})
	if err != nil {
		xmetric.DyanmicItemAPI.Inc("/videoup.open.service.v1.VideoUpOpen/ArcsDynamicView", "request_error")
		log.Error("%v", err)
		return nil, err
	}
	if resTmp == nil {
		xmetric.DyanmicItemAPI.Inc("/videoup.open.service.v1.VideoUpOpen/ArcsDynamicView", "reply_date_error")
		return nil, ecode.NothingFound
	}
	var res = make(map[int64]*videogrpc.DynamicView)
	for aid, re := range resTmp.BCut {
		if re == nil {
			continue
		}
		res[aid] = re
	}
	return res, nil
}
