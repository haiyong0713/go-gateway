package show

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/search"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

func (s *Service) SuggestWeb(c context.Context, plat int8, mid int64, buvid string, param *search.SearchSuggestParam) (res []*search.SuggestItem, err error) {
	suggest, err := s.srch.Suggest(c, plat, mid, param.Platform, buvid, param.Keyword, model.AndroidBilithings, param.Device, param.Build, param.Highlight)
	if err != nil {
		log.Error("%+v", err)
		return []*search.SuggestItem{}, nil
	}
	if suggest == nil || len(suggest.Result) == 0 {
		return []*search.SuggestItem{}, nil
	}
	for _, v := range suggest.Result {
		// 屏蔽所有特殊跳转
		if _, ok := suggestionType[v.TermType]; ok {
			continue
		}
		item := &search.SuggestItem{}
		item.FromSuggestWeb(v)
		res = append(res, item)
	}
	return res, nil
}

func (s *Service) SearchWeb(c context.Context, plat int8, mid int64, buvid string, param *search.SearchParam) (pgcItem, item []cardm.Handler, err error) {
	var (
		cardItem, bgmItem, ugcItem []*ai.Item
		aids                       []int64
		ssids, epids               []int32
		arcs                       map[int64]*arcgrpc.Arc
		seams                      map[int32]*seasongrpc.CardInfoProto
		seamAids                   map[int32]*episodegrpc.EpisodeCardsProto
		epms                       map[int32]*episodegrpc.EpisodeCardsProto
	)
	all, err := s.srch.Search(c, mid, 0, param.Pn, param.Ps, param.Keyword, buvid)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, []cardm.Handler{}, nil
	}
	if all == nil || all.Result == nil {
		return []cardm.Handler{}, []cardm.Handler{}, nil
	}
	// 转换成统一结构体
	// pgc
	for _, v := range all.Result.MediaBangumi {
		bgmItem = append(bgmItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	for _, v := range all.Result.MediaFt {
		bgmItem = append(bgmItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	// archive
	for _, v := range all.Result.Video {
		ugcItem = append(ugcItem, &ai.Item{Goto: model.GotoAv, ID: int64(v.ID)})
	}
	// 插入逻辑
	bgmItem = s.listInsert(bgmItem, param.Pn, param.ParamStr)
	cardItem = append(cardItem, bgmItem...)
	cardItem = append(cardItem, ugcItem...)
	for _, v := range cardItem {
		if v.ID == 0 {
			continue
		}
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		case model.GotoPGC:
			ssids = append(ssids, int32(v.ID))
			if v.ChildID > 0 {
				epids = append(epids, int32(v.ChildID))
			}
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
	if len(ssids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsAll(ctx, ssids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(epids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if epms, err = s.bgm.EpCards(ctx, epids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, []cardm.Handler{}, nil
	}
	materials := &cardm.Materials{
		Arcs:               arcs,
		EpisodeCardsProtom: seamAids,
		Seams:              seams,
		Epms:               epms,
	}
	cardParam := &cardm.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: param.FromType,
	}
	op := &operate.Card{KeyWord: param.Keyword}
	switch param.FromType {
	case model.FromList:
		pgclist := s.cardDealWebItem(cardParam, bgmItem, model.EntranceRelate, model.VerticalCoverV1, materials, op)
		if len(pgclist) == 0 {
			pgclist = nil
		}
		ugclist := s.cardDealWebItem(cardParam, ugcItem, model.EntranceRelate, model.SmallCoverV1, materials, op)
		return pgclist, ugclist, nil
	}
	// 详情页旁边则聚合在一起
	ugclist := s.cardDealWebItem(cardParam, cardItem, model.EntranceRelate, model.SmallCoverV1, materials, op)
	return nil, ugclist, nil
}
