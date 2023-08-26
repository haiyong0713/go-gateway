package like

import (
	"context"

	accmdl "git.bilibili.co/bapis/bapis-go/account/service"
	relmdl "git.bilibili.co/bapis/bapis-go/account/service/relation"
	"go-common/library/log"

	lmdl "go-gateway/app/web-svr/activity/interface/model/like"

	"go-common/library/sync/errgroup.v2"
)

// RcmdData return recommend Information
func (s *Service) RcmdData(c context.Context, mids []int64, mid int64) (rsp *lmdl.RcmdRsp, err error) {
	var (
		accReq = &accmdl.MidsReq{Mids: mids}
		accRsp *accmdl.CardsReply
		relReq = &relmdl.RelationsReq{
			Mid: mid,
			Fid: mids,
		}
		relRsp *relmdl.FollowingMapReply
	)
	eGroup := errgroup.WithContext(c)
	eGroup.Go(func(ctx context.Context) (e error) {
		if accRsp, e = s.accClient.Cards3(c, accReq); e != nil {
			log.Errorc(c, "grpc.account.Card3(%+v) failed. error(%v)", accReq, e)
		}
		return
	})
	if mid != 0 {
		eGroup.Go(func(ctx context.Context) (e error) {
			if relRsp, e = s.relClient.Relations(c, relReq); e != nil {
				log.Errorc(c, "grpc.relation.Relations(%v) failed. error(%v)", relReq, e)
				e = nil
				relRsp = nil
			}
			return
		})
	}
	if err = eGroup.Wait(); err != nil {
		return
	}
	rsp = &lmdl.RcmdRsp{
		Infos: make(map[int64]*lmdl.RcmdInfo),
	}
	for _, m := range mids {
		tmp := &lmdl.RcmdInfo{}
		uInfo, ok := accRsp.Cards[m]
		if !ok || uInfo == nil {
			continue
		}
		var isFav = 0
		if relRsp != nil {
			if favInfo, ok := relRsp.FollowingMap[m]; ok && favInfo != nil && favInfo.Attribute < 128 {
				isFav = 1
			}
		}
		tmp.Mid = m
		tmp.Face = uInfo.Face
		tmp.Name = uInfo.Name
		tmp.Vip = uInfo.Vip
		tmp.Official = uInfo.Official
		tmp.IsFav = isFav
		rsp.Infos[m] = tmp
	}
	return
}
