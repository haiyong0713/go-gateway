package show

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/show"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) ShowWeb(c context.Context, plat int8, mid int64, buvid string, param *show.ShowParam) ([]*show.ItemWeb, error) {
	var (
		regionType = []int64{3, 129, 4, 36, 202}
		cardsTypes = []string{_feed, _popular, regionKey(3), regionKey(129), regionKey(4), regionKey(36), regionKey(202)}
		mutex      sync.Mutex
		aids       []int64
		seams      map[int32]*episodegrpc.EpisodeCardsProto
	)
	listm := map[string][]*ai.Item{}
	// 物料
	arcs := map[int64]*arcgrpc.Arc{}
	regionArcm := map[int64]*arcgrpc.Arc{}
	// 分区key映射关系
	regionm := map[string]int64{}
	// 第一次批量
	group := errgroup.WithContext(c)
	// 热门
	group.Go(func(ctx context.Context) error {
		key := s.popularGroup(mid, "")
		list := s.PopularCardTenList(ctx, key, 0, _max)
		items := []*ai.Item{}
		for _, v := range list {
			item := v.PopularCardToAiChange()
			item.Entrance = model.EntrancePopular
			items = append(items, item)
		}
		mutex.Lock()
		listm[_popular] = items
		mutex.Unlock()
		return nil
	})
	if mid > 0 {
		// 天马
		group.Go(func(ctx context.Context) error {
			group := s.group(mid, buvid)
			list, err := s.rcmd.FeedRecommend(ctx, plat, model.AndroidBilithings, buvid, mid, param.Build, param.LoginEvent, group, _max, 0)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{}
				*item = *v
				item.Entrance = model.EntranceCommonSearch
				items = append(items, item)
			}
			mutex.Lock()
			listm[_feed] = items
			mutex.Unlock()
			return nil
		})
	}
	// 分区
	for _, rid := range regionType {
		tmp := rid
		group.Go(func(ctx context.Context) error {
			list, err := s.reg.RegionDynamic(ctx, tmp, _defaultPn, _defaultPs)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{Goto: model.GotoAv, ID: v.Aid, Entrance: model.EntranceRegion}
				items = append(items, item)
				mutex.Lock()
				regionArcm[v.Aid] = v
				mutex.Unlock()
			}
			mutex.Lock()
			listm[regionKey(tmp)] = items
			regionm[regionKey(tmp)] = tmp
			mutex.Unlock()
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	for _, list := range listm {
		for _, v := range list {
			switch v.Goto {
			case model.GotoAv:
				aids = append(aids, v.ID)
			}
		}
	}
	// 第二次批量
	group = errgroup.WithContext(c)
	// 获取稿件物料
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.ArcsAll(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsByAidsAll(ctx, aids); err != nil {
				log.Error("%+v", err)
				return nil
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	list := cardsTypes
	// 列表处理
	if arcs == nil {
		arcs = map[int64]*arcgrpc.Arc{}
	}
	for _, v := range regionArcm {
		arcs[v.Aid] = v
	}
	materials := &cardm.Materials{
		Arcs:               arcs,
		EpisodeCardsProtom: seams,
	}
	items := []*show.ItemWeb{}
	for _, ct := range list {
		var (
			entrance string
			cardType model.CardType
		)
		op := &operate.Card{}
		cards, ok := listm[ct]
		if !ok {
			continue
		}
		if regionID, ok := regionm[ct]; ok {
			op.Rid = regionID
		}
		cardParam := &cardm.CardParam{
			Plat:     plat,
			Mid:      mid,
			FromType: model.FromList,
		}
		switch ct {
		case _feed:
			entrance = model.EntranceRelate
		case _popular:
			entrance = model.EntrancePopular
		case _myBangumi, _myCinema:
			entrance = model.EntranceMyAnmie
			cardType = model.SmallCoverV2
		default:
			if _, ok := regionm[ct]; ok {
				entrance = model.EntranceRegion
			}
		}
		list := s.cardDealWebItem(cardParam, cards, entrance, cardType, materials, op)
		if len(list) == 0 {
			continue
		}
		item := &show.ItemWeb{Type: ct, Items: list}
		item.FromItemWeb(entrance, op.Rid)
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) cardDealWebItem(param *cardm.CardParam, feedCard []*ai.Item, entrance string, cardType model.CardType, materials *cardm.Materials, op *operate.Card) []cardm.Handler {
	is := []cardm.Handler{}
	// 兜底卡片
	backupCard := []cardm.Handler{}
	for _, v := range feedCard {
		var (
			main interface{}
		)
		op.TrackID = v.TrackID
		op.Cid = v.ChildID
		switch param.FromType {
		case model.FromView:
			cardType = model.SmallCoverV4
		}
		op.From(model.CardGt(v.Goto), entrance, v.ID, param.Plat, param.Build, param.MobiApp)
		materials.Prune = cardm.GtWebPrune(v.Goto, v.ID, v.ChildID)
		switch v.Goto {
		case model.GotoAv:
			main = materials.Arcs
			if ep, ok := materials.EpisodeCardsProtom[int32(v.ID)]; ok {
				main = materials.EpisodeCardsProtom
				v.Goto = model.GotoPGC
				op.Epid = ep.EpisodeId
			}
		case model.GotoAvHis:
			main = materials.ViewReplym
		case model.GotoPGC:
			main = materials.Seams
			if materials.Epms != nil {
				if _, ok := materials.Epms[int32(v.ChildID)]; !ok {
					op.Cid = 0
				}
			}
			switch entrance {
			case model.EntranceMyAnmie:
				if param.FromType != model.FromView {
					anim, ok := materials.Animem[int32(v.ID)]
					if !ok {
						continue
					}
					main = anim
				}
			}
		case model.GotoPGCEp, model.GotoPGCEpHis:
			main = materials.Epms
		}
		h := cardm.Handle(model.PlatH5, model.CardGt(v.Goto), cardType, v, materials)
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
		h.Get().Pos = v.Position
		is = append(is, h)
	}
	if len(is) == 0 {
		if param.IsBackUpCard {
			return backupCard
		}
		return []cardm.Handler{}
	}
	return is
}

func (s *Service) listInsert(cards []*ai.Item, pn int, param string) []*ai.Item {
	var cardItem []*ai.Item
	// 插入逻辑
	if pn == 1 && param != "" {
		gt, id, cid, ok := cardm.FromGtPrune(param)
		if ok {
			cardItem = append(cardItem, &ai.Item{Goto: gt, ID: id, ChildID: cid})
			for _, v := range cards {
				if v.Goto == gt && v.ID == id {
					continue
				}
				cardItem = append(cardItem, v)
			}
		}
		return cardItem
	}
	return cards
}
