package show

import (
	"context"
	"fmt"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/dynamic"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	followgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

// DynVideo 动态视频列表
// nolint: gocognit
func (s *Service) DynVideo(c context.Context, plat int8, mid int64, buvid string, param *dynamic.DynamicParam) (res []cardm.Handler, page *cardm.Page, err error) {
	// Step 1. 获取用户关注链信息(关注的up、追番、购买的课程）
	following, pgcFollowing, err := s.followings(c, mid)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil, nil
	}
	// TODO 没有关系链的情况不返回
	if len(following) == 0 && len(pgcFollowing) == 0 {
		return []cardm.Handler{}, nil, nil
	}
	attentions := dynamic.GetAttentionsParams(mid, following, pgcFollowing)
	var (
		dynList     *dynamic.DynVideoListRes
		dynTypeList = []string{"8", "512", "4097", "4098", "4099", "4100", "4101"}
	)
	switch param.RefreshType {
	case dynamic.RefreshHistory:
		dynList, err = s.dyn.DynVideoHistory(c, mid, param.Offset, param.Page, dynTypeList, attentions, param.Build, param.Platform, param.MobiApp, buvid, param.Device)
		if err != nil {
			log.Error("%+v", err)
			return nil, nil, err
		}
	default:
		dynList, err = s.dyn.DynVideoList(c, mid, param.UpdateBaseline, param.AssistBaseline, dynTypeList, attentions, param.Build, param.Platform, param.MobiApp, buvid, param.Device)
		if err != nil {
			log.Error("%+v", err)
			return nil, nil, err
		}
	}
	if dynList == nil || len(dynList.Dynamics) == 0 {
		return []cardm.Handler{}, nil, nil
	}
	// 额外插入逻辑
	if param.RefreshType == dynamic.RefreshNew && param.TopParamStr != "" {
		var (
			topItem *cardm.Prune
			topDyns []*dynamic.Dynamic
		)
		if param.ParamStr != "" {
			if item, ok := card.FromGtPruneItem(param.TopParamStr); ok {
				topItem = item
				topDyns = append(topDyns, item.FromAiItemToDyn())
			}
		}
		if reply, ok := card.FromGtPrunes(param.TopParamStr); ok {
			for _, v := range reply {
				if topItem != nil && v.ID == topItem.ID && v.Goto == topItem.Goto {
					continue
				}
				topDyns = append(topDyns, v.FromAiItemToDyn())
			}
		}
		// 插入数据去重复
		if len(topDyns) > 0 {
		LOOP:
			for _, v := range dynList.Dynamics {
				for _, t := range topDyns {
					// 相同则跳过
					if t.Rid == v.Rid && t.Type == v.Type {
						continue LOOP
					}
				}
				topDyns = append(topDyns, v)
			}
			dynList.Dynamics = topDyns
		}
	} else if param.RefreshType == dynamic.RefreshNew && param.ParamStr != "" {
		if dyn, ok := card.FromDynamicPrune(param.ParamStr); ok {
			var isok bool
			for _, l := range dynList.Dynamics {
				if l.DynamicID == dyn.DynamicID {
					isok = true
					break
				}
			}
			if !isok {
				cards := []*dynamic.Dynamic{dyn}
				cards = append(cards, dynList.Dynamics...)
				dynList.Dynamics = cards
			}
		}
	}
	items := s.dealItemDynamic(c, plat, param, dynList.Dynamics)
	page = &cardm.Page{
		UpdateNum:      dynList.UpdateNum,
		HistoryOffset:  dynList.HistoryOffset,
		UpdateBaseline: dynList.UpdateBaseline,
		Pn:             int(param.Page) + 1,
	}
	return items, page, nil
}

func (s *Service) dealItemDynamic(c context.Context, plat int8, param *dynamic.DynamicParam, dynList []*dynamic.Dynamic) []cardm.Handler {
	var (
		aids     []int64
		epids    []int32
		arcs     map[int64]*arcgrpc.Arc
		seams    map[int32]*episodegrpc.EpisodeCardsProto
		seamAids map[int32]*episodegrpc.EpisodeCardsProto
	)
	for _, l := range dynList {
		r := l.DynamicCardChange()
		switch r.Goto {
		case model.GotoAv:
			if r.ID == 0 {
				continue
			}
			aids = append(aids, r.ID)
		case model.GotoPGC:
			if r.ID == 0 {
				continue
			}
			epids = append(epids, int32(r.ID))
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
	items := []cardm.Handler{}
	// 兜底卡片
	backupCard := []cardm.Handler{}
	// 去掉当前一刷内的重复数据
	deduplication := map[string]struct{}{}
	for _, l := range dynList {
		var (
			r        = l.DynamicCardChange()
			main     interface{}
			cardType model.CardType
			op       = &operate.Card{DynCtime: l.Ctime}
		)
		dkey := fmt.Sprintf("%s_%d", r.Goto, r.ID)
		if _, ok := deduplication[dkey]; ok {
			// 如果有重复数据跳出本次循环
			continue
		}
		deduplication[dkey] = struct{}{}
		op.From(model.CardGt(r.Goto), model.EntranceDynamicVideo, r.ID, plat, param.Build, param.MobiApp)
		switch r.Goto {
		case model.GotoAv:
			main = arcs
			if _, ok := seamAids[int32(r.ID)]; ok {
				main = seamAids
				r.Goto = model.GotoPGC
			}
		case model.GotoPGC:
			main = seams
		}
		materials := &cardm.Materials{
			EpisodeCardsProtom: seams,
			Prune:              cardm.DynamicPrune(l),
		}
		switch param.FromType {
		case model.FromList:
			cardType = model.SmallCoverV1
		default:
			cardType = model.SmallCoverV4
		}
		h := cardm.Handle(plat, model.CardGt(r.Goto), cardType, r, materials)
		if h == nil || !h.From(main, op) {
			continue
		}
		// 找出互动视频卡片，并把第一张放入兜底卡片里面，并且长度为0的时候放入
		if h.Get().Filter == model.FilterAttrBitSteinsGate {
			if len(backupCard) == 0 {
				backupCard = append(backupCard, h)
			}
			// 互动视频不放入列表里面
			continue
		}
		items = append(items, h)
	}
	if len(items) == 0 {
		return backupCard
	}
	return items
}

func (s *Service) followings(c context.Context, mid int64) ([]*relationgrpc.FollowingReply, []*followgrpc.FollowSeasonProto, error) {
	eg := errgroup.WithCancel(c)
	var (
		follow []*relationgrpc.FollowingReply
		pgc    []*followgrpc.FollowSeasonProto
	)
	eg.Go(func(ctx context.Context) error {
		var err error
		follow, err = s.reldao.Followings(ctx, mid)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var err error
		pgc, err = s.bgm.MyRelations(ctx, mid)
		if err != nil {
			return nil
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}
	return follow, pgc, nil
}
