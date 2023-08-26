package s10

import (
	"context"

	"go-gateway/app/web-svr/activity/interface/model/s10"
	"go-gateway/app/web-svr/activity/interface/tool"

	"go-common/library/log"
)

func (d *Dao) FreeFlowPubDataBus(ctx context.Context, typ, source int32, tel string) (err error) {
	pack := &s10.FreeFlowPub{Message: tel, Type: typ, Source: source}
	if err = d.freeFlowDataBusPub.Send(ctx, tel, pack); err != nil {
		tool.Metric4FreeFlowPubDatabus.WithLabelValues("FreeFlowPubDataBus").Inc()
		log.Errorc(ctx, "s10 d.dao.SignedPubDataBus(tel:%s,type:%d) error(%v)", tel, typ, err)
	}
	return
}
