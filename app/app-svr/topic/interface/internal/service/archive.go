package service

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	coingrpc "git.bilibili.co/bapis/bapis-go/community/service/coin"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

func (s *Service) arcsPlayer(ctx context.Context, aids []*archivegrpc.PlayAv, showPgcPlayurl bool, from string) (map[int64]*archivegrpc.ArcPlayer, error) {
	var max50 = 50
	g, mu := errgroup.WithContext(ctx), sync.Mutex{}
	res := make(map[int64]*archivegrpc.ArcPlayer)
	batchPlayArg := constructBatchPlayArgFromCtx(ctx, showPgcPlayurl, from)
	for i := 0; i < len(aids); i += max50 {
		partAids := aids[i:]
		if i+max50 < len(aids) {
			partAids = aids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			arg := &archivegrpc.ArcsPlayerRequest{
				PlayAvs:      partAids,
				BatchPlayArg: batchPlayArg,
			}
			archives, err := s.archiveGRPC.ArcsPlayer(ctx, arg)
			if err != nil {
				log.Error("ArcsPlayer partAids(%+v) err(%v)", partAids, err)
				return err
			}
			mu.Lock()
			for aid, arc := range archives.GetArcsPlayer() {
				if arc == nil || !arc.Arc.IsNormal() {
					continue
				}
				// 判断卡片是否为付费卡片，跳过付费卡片
				if arc.Arc.Pay != nil && arc.Arc.Pay.PayAttr > 0 {
					continue
				}
				res[aid] = arc
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

func constructBatchPlayArgFromCtx(ctx context.Context, showPgcPlayurl bool, from string) *archivegrpc.BatchPlayArg {
	batchArg, _ := arcmid.FromContext(ctx)
	tmpBatchArg := &archivegrpc.BatchPlayArg{}
	if batchArg != nil {
		*tmpBatchArg = *batchArg
		tmpBatchArg.ShowPgcPlayurl = showPgcPlayurl
		tmpBatchArg.From = from
	}
	return tmpBatchArg
}

func (s *Service) isFavVideos(ctx context.Context, mid int64, aids []int64) (map[int64]int8, error) {
	reply, err := s.favGRPC.IsFavoreds(ctx, &favgrpc.IsFavoredsReq{Typ: 2, Mid: mid, Oids: aids})
	if err != nil {
		return nil, err
	}
	res := make(map[int64]int8)
	for k, v := range reply.Faveds {
		if v {
			res[k] = 1
		}
	}
	return res, nil
}

func (s *Service) archiveUserCoins(ctx context.Context, aids []int64, mid int64) (map[int64]int64, error) {
	arg := &coingrpc.ItemsUserCoinsReq{
		Mid:      mid,
		Aids:     aids,
		Business: "archive",
	}
	reply, err := s.coinGRPC.ItemsUserCoins(ctx, arg)
	if err != nil {
		return nil, err
	}
	return reply.Numbers, nil
}

func (s *Service) multiStats(c context.Context, mid int64, business map[string][]*dynmdlV2.ThumbsRecord) (*thumgrpc.MultiStatsReply, error) {
	var busiParams = map[string]*thumgrpc.MultiStatsReq_Business{}
	for busType, item := range business {
		busiTmp, ok := busiParams[busType]
		if !ok {
			busiTmp = &thumgrpc.MultiStatsReq_Business{}
			busiParams[busType] = busiTmp
		}
		for _, v := range item {
			recTmp := &thumgrpc.MultiStatsReq_Record{
				OriginID:  v.OrigID,
				MessageID: v.MsgID,
			}
			busiTmp.Records = append(busiTmp.Records, recTmp)
		}
	}
	in := &thumgrpc.MultiStatsReq{
		Mid:      mid,
		Business: busiParams,
	}
	likeStats, err := s.thumbGRPC.MultiStats(c, in)
	if err != nil {
		return nil, err
	}
	return likeStats, nil
}

func (s *Service) hasLike(ctx context.Context, buvid string, mid int64, messageIDs []int64) (map[int64]int8, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	out := make(map[int64]int8)
	if mid > 0 {
		arg := &thumgrpc.HasLikeReq{
			Business:   "archive",
			MessageIds: messageIDs,
			Mid:        mid,
			IP:         ip,
		}
		reply, err := s.thumbGRPC.HasLike(ctx, arg)
		if err != nil {
			return nil, err
		}
		for k, v := range reply.States {
			out[k] = int8(v.State)
		}
		return out, nil
	}
	arg := &thumgrpc.BuvidHasLikeReq{
		Business:   "archive",
		MessageIds: messageIDs,
		Buvid:      buvid,
		IP:         ip,
	}
	reply, err := s.thumbGRPC.BuvidHasLike(ctx, arg)
	if err != nil {
		return nil, err
	}
	for k, v := range reply.States {
		out[k] = int8(v.State)
	}
	return out, nil
}
