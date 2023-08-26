package activity

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/space"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	actgrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
)

// Dao is answer dao.
type Dao struct {
	natClient actgrpc.NaPageClient
	actClient activitygrpc.ActivityClient
}

// New answerClient.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.natClient, err = actgrpc.NewClient(c.NatGRPC); err != nil {
		panic(fmt.Sprintf("natClient NewClient error(%v)", err))
	}
	if d.actClient, err = activitygrpc.NewClient(c.ActivityClient); err != nil {
		panic(fmt.Sprintf("activityClient NewClient error(%v)", err))
	}
	return
}

func (d *Dao) NatActInfo(ctx context.Context, ids []int64, pageType int64, content map[string]string) (map[int64]*actgrpc.NativePage, error) {
	resTmp, err := d.natClient.NatInfoFromForeign(ctx, &actgrpc.NatInfoFromForeignReq{Fids: ids, PageType: pageType, Content: content})
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	if resTmp == nil {
		return nil, nil
	}
	return resTmp.List, nil
}

func (d *Dao) UpActUserSpaceCard(c context.Context, mid, vmid int64) ([]*activitygrpc.UpActReserveRelationInfo, error) {
	arg := &activitygrpc.UpActUserSpaceCardReq{
		Upmid: vmid,
		Mid:   mid,
		From:  activitygrpc.UpCreateActReserveFrom_FromSpace,
	}
	reply, err := d.actClient.UpActUserSpaceCard(c, arg)
	if err != nil {
		return nil, err
	}
	return reply.List, nil
}

func (d *Dao) CheckReserveDoveAct(c context.Context, mid int64, source int64, relation map[int64]*activitygrpc.UpActReserveRelationInfo) (map[int64]*activitygrpc.ReserveDoveActRelationInfo, error) {
	arg := &activitygrpc.CheckReserveDoveActReq{
		Mid:    mid,
		Source: source,
		Relations: &activitygrpc.UpActReserveRelationInfoReply{
			List: relation,
		},
	}
	reply, err := d.actClient.CheckReserveDoveAct(c, arg)
	if err != nil {
		return nil, err
	}
	return reply.List, nil
}

func (d *Dao) AddReserve(ctx context.Context, req *space.AddReserveReq) error {
	arg := &activitygrpc.AddReserveReq{
		Sid:      req.Sid,
		Mid:      req.Mid,
		From:     req.From,
		Typ:      req.Type,
		Oid:      req.Oid,
		Ip:       metadata.String(ctx, metadata.RemoteIP),
		Platform: req.Platform,
		Mobiapp:  req.Mobiapp,
		Buvid:    req.Buvid,
		Spmid:    req.Spmid,
	}
	_, err := d.actClient.AddReserve(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) GetReserveCalendarInfo(ctx context.Context, sid int64) (*activitygrpc.GetReserveCalendarInfoReply, error) {
	arg := &activitygrpc.GetReserveCalendarInfoReq{
		Sid: sid,
	}
	return d.actClient.GetReserveCalendarInfo(ctx, arg)
}

func (d *Dao) CancelUpActReserve(ctx context.Context, mid int64, sid int64) error {
	arg := &activitygrpc.CancelUpActReserveReq{
		Mid:  mid,
		Sid:  sid,
		From: activitygrpc.UpCreateActReserveFrom_FromSpace,
	}
	_, err := d.actClient.CancelUpActReserve(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dao) DelReserve(ctx context.Context, mid int64, sid int64) error {
	arg := &activitygrpc.DelReserveReq{
		Sid: sid,
		Mid: mid,
	}
	_, err := d.actClient.DelReserve(ctx, arg)
	if err != nil {
		return err
	}
	return nil
}
