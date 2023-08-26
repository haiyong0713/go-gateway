package up_archive

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	uparcapi "git.bilibili.co/bapis/bapis-go/up-archive/service"

	"github.com/pkg/errors"
)

type Dao struct {
	upArcClient uparcapi.UpArchiveClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.upArcClient, err = uparcapi.NewClient(c.UpArcClient); err != nil {
		panic(fmt.Sprintf("uparcapi.NewClient error(%v)", err))
	}
	return
}

func (d *Dao) ArcPassedByAid(ctx context.Context, args *uparcapi.ArcPassedByAidReq) (*uparcapi.ArcPassedByAidReply, error) {
	return d.upArcClient.ArcPassedByAid(ctx, args)
}

func (d *Dao) ArcPassedExist(ctx context.Context, args *uparcapi.ArcPassedExistReq) (*uparcapi.ArcPassedExistReply, error) {
	return d.upArcClient.ArcPassedExist(ctx, args)
}

func (d *Dao) ArcPassed(ctx context.Context, mid, pn, ps int64, order string, without []uparcapi.Without) (arcs []*uparcapi.Arc, total int64, err error) {
	var searchOrder uparcapi.SearchOrder
	switch order {
	case "click":
		searchOrder = uparcapi.SearchOrder_click
	default:
		searchOrder = uparcapi.SearchOrder_pubtime
	}
	arg := &uparcapi.ArcPassedReq{Mid: mid, Pn: pn, Ps: ps, Without: without, Order: searchOrder}
	reply, err := d.upArcClient.ArcPassed(ctx, arg)
	if err != nil {
		return nil, 0, errors.Wrapf(err, "ArcPassed arg:%+v", arg)
	}
	if reply == nil {
		return nil, 0, errors.Wrapf(ecode.NothingFound, "ArcPassed arg:%+v", arg)
	}
	return reply.Archives, reply.Total, nil
}
