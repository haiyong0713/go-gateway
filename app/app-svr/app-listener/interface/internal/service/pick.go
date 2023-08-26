package service

import (
	"context"
	"fmt"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) PickFeed(ctx context.Context, req *v1.PickFeedReq) (resp *v1.PickFeedResp, err error) {
	dev, net, au := DevNetAuthFromCtx(ctx)
	picks, offset, err := s.dao.PickCards(ctx, dao.PickCardsOpt{Mid: au.Mid, Buvid: dev.Buvid, Offset: req.Offset})
	if err != nil {
		return nil, err
	}
	collections := make([]*listenerSvc.Collection, 0, len(picks)*3)
	for _, p := range picks {
		collections = append(collections, p.Pick.GetCollections()...)
	}
	pCtx := s.writeMaterialIds(model.PickFeed, collections)
	err = s.getMaterials(ctx, getMaterialOpt{Dev: dev, Net: net, Auth: au}, pCtx)
	if err != nil {
		return
	}
	resp = &v1.PickFeedResp{
		Offset: offset,
	}
	sc := model.SingleCollection{}
	for _, p := range picks {
		sc.PickId = p.Pick.Id
		sc.PickTitle = p.Pick.Title
		for i, c := range p.Pick.GetCollections() {
			sc.Collection = c
			// 只有pick组中的第一张卡才带标题
			if i != 0 {
				sc.PickTitle = ""
			}
			resp.Cards = append(resp.Cards, pCtx.GetPickFeedCard(sc))
		}
	}

	return
}

func (s *Service) PickCardDetail(ctx context.Context, req *v1.PickCardDetailReq) (resp *v1.PickCardDetailResp, err error) {
	dev, net, au := DevNetAuthFromCtx(ctx)
	cDetail, err := s.dao.CardDetail(ctx, dao.CardDetailsOpt{
		PickId: req.PickId, CardId: req.CardId,
	})
	if err != nil {
		return
	}
	pCtx := s.writeMaterialIds(model.PickDetail, cDetail.Collection)
	err = s.getMaterials(ctx, getMaterialOpt{Dev: dev, Net: net, Auth: au}, pCtx)
	if err != nil {
		return
	}
	resp = &v1.PickCardDetailResp{
		PickId: req.PickId, CardId: req.CardId,
	}
	resp.Modules = pCtx.GetPickDetailModules(cDetail)
	return
}

func (s *Service) writeMaterialIds(From string, objs ...interface{}) *model.PickContext {
	pCtx := &model.PickContext{C: s.C, From: From}
	pCtx.Init()
	for i := range objs {
		switch o := objs[i].(type) {
		case []*listenerSvc.Collection:
			for _, c := range o {
				if From == model.PickFeed && int64(len(c.GetArchives())) > c.GetDisplayNum() {
					c.Archives = c.Archives[0:c.DisplayNum]
				}
				for _, ak := range c.GetArchives() {
					pCtx.ArcM[ak.Aid] = nil
				}
			}
		case *listenerSvc.Collection:
			if From == model.PickFeed && int64(len(o.GetArchives())) > o.GetDisplayNum() {
				o.Archives = o.Archives[0:o.DisplayNum]
			}
			for _, ak := range o.GetArchives() {
				pCtx.ArcM[ak.Aid] = nil
			}
		default:
			panic(fmt.Sprintf("writeMaterialIds: programmer error unknown type %T", objs[i]))
		}
	}
	return pCtx
}

type getMaterialOpt struct {
	Dev  *device.Device
	Net  *network.Network
	Auth *auth.Auth
}

func (s *Service) getMaterials(ctx context.Context, opt getMaterialOpt, pCtx *model.PickContext) (err error) {
	aids := make([]int64, 0, len(pCtx.ArcM))
	for k := range pCtx.ArcM {
		aids = append(aids, k)
	}
	if len(aids) <= 0 {
		return nil
	}
	eg1 := errgroup.WithContext(ctx)

	eg1.Go(func(c context.Context) error {
		res, err := s.dao.ArchiveInfos(c, dao.ArchiveInfoOpt{Aids: aids, Mid: opt.Auth.Mid, Device: opt.Dev})
		if err != nil {
			return err
		}
		for k := range pCtx.ArcM {
			if res[k].Arc != nil {
				pCtx.ArcM[k] = &model.ArchiveInfo{Arc: res[k].Arc}
			}
		}
		return nil
	})

	// 一级获取
	err = eg1.Wait()
	if err != nil {
		return
	}

	return
}
