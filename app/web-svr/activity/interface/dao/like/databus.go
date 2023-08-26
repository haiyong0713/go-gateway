package like

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/like"
)

const _retry = 3

func (d *Dao) ReportGaia(c context.Context, scene, eventCtx string) (err error) {
	nowTime := time.Now()
	millisecond := nowTime.UnixNano() / 1000 / 1000
	data := &like.GaiaReport{
		Scene:    scene,
		TraceId:  strconv.FormatInt(millisecond, 10),
		EventTs:  nowTime.Unix(),
		EventCtx: eventCtx,
	}
	d.retry(func() error {
		return component.GaiaRiskProducer.Send(c, data.TraceId, data)
	})
	log.Infoc(c, "reportGaia gaiaRiskPub info data(%+v) ", data)
	return
}

func (d *Dao) retry(f func() error) {
	for i := 0; i < _retry; i++ {
		err := f()
		if err == nil {
			return
		}
		log.Error("retry error:%v", err)
		time.Sleep(100 * time.Millisecond)
	}
}
