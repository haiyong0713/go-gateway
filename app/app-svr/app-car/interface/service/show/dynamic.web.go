package show

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/dynamic"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) dynVideoList(c context.Context, mid int64, buvid string, param *dynamic.DynamicParam) (cards []*ai.Item, page *cardm.Page, err error) {
	// Step 1. 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, err := s.followings(c, mid)
	if err != nil {
		log.Error("%+v", err)
		return nil, nil, err
	}
	// 没有关系链的情况不返回
	if len(following) == 0 && len(pgcFollowing) == 0 {
		return nil, nil, nil
	}
	attentions := dynamic.GetAttentionsParams(mid, following, pgcFollowing)
	var (
		dynList     *dynamic.DynVideoListRes
		dynTypeList = []string{"8", "512", "4097", "4098", "4099", "4100", "4101"}
	)
	switch param.RefreshType {
	case dynamic.RefreshHistory:
		dynList, err = s.dyn.DynVideoHistory(c, mid, param.Offset, param.Page, dynTypeList, attentions, param.Build, param.Platform, model.AndroidBilithings, buvid, param.Device)
		if err != nil {
			log.Error("%+v", err)
			return nil, nil, err
		}
	default:
		dynList, err = s.dyn.DynVideoList(c, mid, param.UpdateBaseline, param.AssistBaseline, dynTypeList, attentions, param.Build, param.Platform, model.AndroidBilithings, buvid, param.Device)
		if err != nil {
			log.Error("%+v", err)
			return nil, nil, err
		}
	}
	if dynList == nil || len(dynList.Dynamics) == 0 {
		return nil, nil, nil
	}
	page = &cardm.Page{
		UpdateNum:      dynList.UpdateNum,
		HistoryOffset:  dynList.HistoryOffset,
		UpdateBaseline: dynList.UpdateBaseline,
		Pn:             int(param.Page) + 1,
	}
	// 去掉当前一刷内的重复数据
	deduplication := map[string]struct{}{}
	for _, v := range dynList.Dynamics {
		item := v.DynamicCardChangeV2()
		dkey := fmt.Sprintf("%s_%d", item.Goto, item.ID)
		if _, ok := deduplication[dkey]; ok {
			// 如果有重复数据跳出本次循环
			continue
		}
		deduplication[dkey] = struct{}{}
		cards = append(cards, item)
	}
	return cards, page, nil
}

func (s *Service) DynVideoWeb(c context.Context, plat int8, mid int64, buvid string, param *dynamic.DynamicParam) (cards []cardm.Handler, page *cardm.Page, err error) {
	cardItem, page, err := s.dynVideoList(c, mid, buvid, param)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil, nil
	}
	// 插入逻辑
	var pn int
	if param.RefreshType == dynamic.RefreshNew {
		pn = 1
	}
	cardItem = s.listInsert(cardItem, pn, param.ParamStr)
	var (
		aids     []int64
		epids    []int32
		arcs     map[int64]*arcgrpc.Arc
		seams    map[int32]*episodegrpc.EpisodeCardsProto
		seamAids map[int32]*episodegrpc.EpisodeCardsProto
	)
	for _, v := range cardItem {
		if v.ID == 0 {
			continue
		}
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		case model.GotoPGCEp:
			epids = append(epids, int32(v.ID))
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seamAids, err = s.bgm.CardsByAids(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(epids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.EpCards(ctx, epids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	materials := &cardm.Materials{
		Arcs:               arcs,
		EpisodeCardsProtom: seamAids,
		Epms:               seams,
	}
	cardParam := &cardm.CardParam{
		Plat:         plat,
		Mid:          mid,
		FromType:     model.FromView,
		IsBackUpCard: true,
	}
	op := &operate.Card{}
	list := s.cardDealWebItem(cardParam, cardItem, model.EntranceDynamicVideo, model.SmallCoverV4, materials, op)
	return list, page, nil
}
