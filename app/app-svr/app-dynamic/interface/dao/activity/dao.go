package activity

import (
	"context"
	"sync"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	natpagegrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
)

type Dao struct {
	c                 *conf.Config
	natPageGrpcClient natpagegrpc.NaPageClient
	grpcClient        activitygrpc.ActivityClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.natPageGrpcClient, err = natpagegrpc.NewClient(c.NatPageGRPC); err != nil {
		panic(err)
	}
	if d.grpcClient, err = activitygrpc.NewClient(c.ActivityClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) NatInfoFromForeign(c context.Context, tids []int64, pageType int64) (map[int64]*natpagegrpc.NativePage, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*natpagegrpc.NativePage)
	for i := 0; i < len(tids); i += max50 {
		var partTids []int64
		if i+max50 > len(tids) {
			partTids = tids[i:]
		} else {
			partTids = tids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			nfs, err := d.NatInfoFromForeignSlice(ctx, partTids, 1)
			if err != nil {
				return err
			}
			mu.Lock()
			for tid, nf := range nfs {
				res[tid] = nf
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("ChannelInfos tids(%+v) eg.wait(%+v)", tids, err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) NatInfoFromForeignSlice(c context.Context, tids []int64, pageType int64) (res map[int64]*natpagegrpc.NativePage, err error) {
	var (
		args   = &natpagegrpc.NatInfoFromForeignReq{Fids: tids, PageType: pageType}
		resTmp *natpagegrpc.NatInfoFromForeignReply
	)
	if resTmp, err = d.natPageGrpcClient.NatInfoFromForeign(c, args); err != nil {
		return
	}
	// 木有getList方法
	if resTmp != nil {
		res = resTmp.List
	}
	return
}

func (d *Dao) ActivityRelation(c context.Context, mid int64, actIDs []int64) (map[int64]*activitygrpc.ActRelationInfoReply, error) {
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	var res = make(map[int64]*activitygrpc.ActRelationInfoReply)
	for _, id := range actIDs {
		var actID = id
		g.Go(func(ctx context.Context) (err error) {
			resTmp, err := d.grpcClient.ActRelationInfo(ctx, &activitygrpc.ActRelationInfoReq{
				Id:       actID,
				Mid:      mid,
				Specific: "reserve,native",
			})
			if err != nil {
				return err
			}
			mu.Lock()
			res[actID] = resTmp
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("ActivityRelation actIDs(%+v) eg.wait(%+v)", actIDs, err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) NativePageCards(c context.Context, ids []int64, general *mdlv2.GeneralParam) (map[int64]*natpagegrpc.NativePageCard, error) {
	var arg = &natpagegrpc.NativePageCardsReq{
		Pids:     ids,
		Device:   general.GetDevice(),
		MobiApp:  general.GetMobiApp(),
		Build:    int32(general.GetBuild()),
		Buvid:    general.GetBuvid(),
		Platform: general.GetPlatform(),
	}
	res, err := d.natPageGrpcClient.NativePageCards(c, arg)
	if err != nil {
		log.Error("NativePageCards arg %+v,err %v", arg, err)
		return nil, err
	}
	return res.List, err
}

func (d *Dao) UpActReserveRelationInfo(c context.Context, ids []int64, mid int64) (*activitygrpc.UpActReserveRelationInfoReply, error) {
	arg := &activitygrpc.UpActReserveRelationInfoReq{
		Mid:  mid,
		Sids: ids,
		From: activitygrpc.UpCreateActReserveFrom_FromDynamic,
	}
	reply, err := d.grpcClient.UpActReserveRelationInfo(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) CheckReserveDoveAct(c context.Context, mid int64, relation *activitygrpc.UpActReserveRelationInfoReply) (map[int64]*activitygrpc.ReserveDoveActRelationInfo, error) {
	arg := &activitygrpc.CheckReserveDoveActReq{
		Mid:       mid,
		Relations: relation,
		Source:    2,
	}
	reply, err := d.grpcClient.CheckReserveDoveAct(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.List, nil
}

func (d *Dao) UpActUserSpaceCard(c context.Context, upmid, mid int64) ([]*activitygrpc.UpActReserveRelationInfo, error) {
	arg := &activitygrpc.UpActUserSpaceCardReq{
		Upmid: upmid,
		Mid:   mid,
	}
	reply, err := d.grpcClient.UpActUserSpaceCard(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.List, nil
}

func (d *Dao) NativeAllPageCards(c context.Context, ids []int64) (map[int64]*natpagegrpc.NativePageCard, error) {
	var arg = &natpagegrpc.NativeAllPageCardsReq{
		Pids: ids,
	}
	res, err := d.natPageGrpcClient.NativeAllPageCards(c, arg)
	if err != nil {
		log.Error("NativeAllPageCards arg %+v,err %v", arg, err)
		return nil, err
	}
	return res.List, nil
}
