package show

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/show"

	cardappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

// nolint: gocognit
func (s *Service) PGCShowWeb(c context.Context, mid int64, plat int8, buvid string, param *show.ShowParam) ([]*show.ItemWeb, error) {
	const (
		_bangumi = "bangumi"
		_movie   = "movie"
	)
	var (
		cardsTypes = []string{_myBangumi, _bangumi, _domestic, _myCinema, _cinema, _cinemaDoc}
		mutex      sync.Mutex
		ssids      []int32
		seasonm    map[int32]*seasongrpc.CardInfoProto
	)
	listm := map[string][]*ai.Item{}
	animem := map[int32]*cardappgrpc.CardSeasonProto{}
	bangumim := map[int32]*bangumi.Module{}
	// 第一次批量
	group := errgroup.WithContext(c)
	switch param.FollowType {
	case _bangumi:
		if mid > 0 {
			// 我的追番
			group.Go(func(ctx context.Context) error {
				const followType = 1
				list, err := s.bgm.MyFollows(ctx, mid, followType, _defaultPn, _defaultPs)
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				items := []*ai.Item{}
				for _, v := range list {
					item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonId), Entrance: model.EntranceMyAnmie}
					if v.Progress != nil {
						item.ChildID = int64(v.Progress.EpisodeId)
					}
					items = append(items, item)
					mutex.Lock()
					animem[v.SeasonId] = v
					mutex.Unlock()
				}
				mutex.Lock()
				listm[_myBangumi] = items
				mutex.Unlock()
				return nil
			})
		}
		// 18 番剧推荐
		group.Go(func(ctx context.Context) error {
			const followType = 18
			list, err := s.bgm.Module(ctx, followType, model.AndroidBilithings, buvid)
			if err != nil {
				log.Error("%+v", err)
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID), Entrance: model.EntrancePgcList}
				items = append(items, item)
				mutex.Lock()
				bangumim[v.SeasonID] = v
				mutex.Unlock()
			}
			mutex.Lock()
			listm[_bangumi] = items
			mutex.Unlock()
			return nil
		})
		// 19 国创推荐
		group.Go(func(ctx context.Context) error {
			const followType = 19
			list, err := s.bgm.Module(ctx, followType, model.AndroidBilithings, buvid)
			if err != nil {
				log.Error("%+v", err)
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID), Entrance: model.EntrancePgcList}
				items = append(items, item)
				mutex.Lock()
				bangumim[v.SeasonID] = v
				mutex.Unlock()
			}
			mutex.Lock()
			listm[_domestic] = items
			mutex.Unlock()
			return nil
		})
	case _movie:
		if mid > 0 {
			// 我的追剧
			group.Go(func(ctx context.Context) error {
				const followType = 2
				list, err := s.bgm.MyFollows(ctx, mid, followType, _defaultPn, _defaultPs)
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				items := []*ai.Item{}
				for _, v := range list {
					item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonId), Entrance: model.EntranceMyAnmie}
					if v.Progress != nil {
						item.ChildID = int64(v.Progress.EpisodeId)
					}
					items = append(items, item)
					mutex.Lock()
					animem[v.SeasonId] = v
					mutex.Unlock()
				}
				mutex.Lock()
				listm[_myCinema] = items
				mutex.Unlock()
				return nil
			})
		}
		// 88 电影热播
		group.Go(func(ctx context.Context) error {
			const followType = 88
			list, err := s.bgm.Module(ctx, followType, model.AndroidBilithings, buvid)
			if err != nil {
				log.Error("%+v", err)
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID), Entrance: model.EntrancePgcList}
				items = append(items, item)
				mutex.Lock()
				bangumim[v.SeasonID] = v
				mutex.Unlock()
			}
			mutex.Lock()
			listm[_cinema] = items
			mutex.Unlock()
			return nil
		})
		// 87 纪录片热播
		group.Go(func(ctx context.Context) error {
			const followType = 87
			list, err := s.bgm.Module(ctx, followType, model.AndroidBilithings, buvid)
			if err != nil {
				log.Error("%+v", err)
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID), Entrance: model.EntrancePgcList}
				items = append(items, item)
				mutex.Lock()
				bangumim[v.SeasonID] = v
				mutex.Unlock()
			}
			mutex.Lock()
			listm[_cinemaDoc] = items
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
			case model.GotoPGC:
				ssids = append(ssids, int32(v.ID))
			}
		}
	}
	// 第二次批量
	group = errgroup.WithContext(c)
	// 获取稿件物料
	if len(ssids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seasonm, err = s.bgm.CardsAll(ctx, ssids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	materials := &cardm.Materials{
		Animem:   animem,
		Bangumim: bangumim,
		Seams:    seasonm,
	}
	list := cardsTypes
	items := []*show.ItemWeb{}
	for _, ct := range list {
		var (
			entrance string
			cardType model.CardType
		)
		op := &operate.Card{FollowType: pgcType[ct]}
		cards, ok := listm[ct]
		if !ok {
			continue
		}
		cardParam := &cardm.CardParam{
			Plat:     plat,
			Mid:      mid,
			FromType: model.FromList,
		}
		switch ct {
		case _myBangumi, _myCinema:
			entrance = model.EntranceMyAnmie
			cardType = model.SmallCoverV2
		case _domestic, _cinema, _cinemaDoc, _bangumi:
			entrance = model.EntrancePgcList
			cardType = model.VerticalCoverV1
		}
		list := s.cardDealWebItem(cardParam, cards, entrance, cardType, materials, op)
		if len(list) == 0 {
			continue
		}
		if len(list) > _max {
			list = list[:_max]
		}
		item := &show.ItemWeb{Type: ct, Items: list}
		item.FromItemWeb(entrance, op.Rid)
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) MyAnimeWeb(c context.Context, plat int8, mid int64, param *bangumi.MyAnimeParam) ([]cardm.Handler, error) {
	var followType int
	switch param.FollowType {
	case _followTypeBangumi:
		followType = 1
	case _followTypeCinema:
		followType = 2
	default:
		return []cardm.Handler{}, nil
	}
	follows, err := s.bgm.MyFollows(c, mid, followType, param.Pn, param.Ps)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil
	}
	var (
		ssids    []int32
		cardItem []*ai.Item
		seasonm  map[int32]*seasongrpc.CardInfoProto
	)
	for _, v := range follows {
		if v.SeasonId == 0 {
			continue
		}
		item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonId)}
		if v.Progress != nil {
			item.ChildID = int64(v.Progress.EpisodeId)
		}
		cardItem = append(cardItem, item)
	}
	// 插入逻辑
	cardItem = s.listInsert(cardItem, param.Pn, param.ParamStr)
	for _, v := range cardItem {
		ssids = append(ssids, int32(v.ID))
	}
	if len(ssids) > 0 {
		var err error
		if seasonm, err = s.bgm.CardsAll(c, ssids); err != nil {
			log.Error("%+v", err)
		}
	}
	cardParam := &cardm.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromView,
	}
	materials := &cardm.Materials{
		Seams: seasonm,
	}
	op := &operate.Card{FollowType: param.FollowType}
	list := s.cardDealWebItem(cardParam, cardItem, model.EntranceMyAnmie, model.SmallCoverV4, materials, op)
	return list, nil
}

func (s *Service) BangumiListWeb(c context.Context, plat int8, mid int64, buvid string, param *bangumi.ListParam) ([]cardm.Handler, error) {
	// 18 番剧推荐
	// 19 国创推荐
	// 88 电影热播
	// 87 纪录片热播
	var followType int
	switch param.FollowType {
	case _followTypeBangumi:
		followType = 18
	case _followTypeDomestic:
		followType = 19
	case _followTypeCinema:
		followType = 88
	case _followTypeCinemaDoc:
		followType = 87
	default:
		return []cardm.Handler{}, nil
	}
	// 实际PGC这接口一期没有分页，第一页就返回了所有数据
	if param.Pn > 1 {
		return []cardm.Handler{}, nil
	}
	bgms, err := s.bgm.Module(c, followType, model.AndroidBilithings, buvid)
	if err != nil {
		log.Warn("%+v", err)
		return []cardm.Handler{}, nil
	}
	var (
		ssids    []int32
		cardItem []*ai.Item
		seasonm  map[int32]*seasongrpc.CardInfoProto
	)
	for _, v := range bgms {
		if v.SeasonID == 0 {
			continue
		}
		item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)}
		cardItem = append(cardItem, item)
	}
	// 插入逻辑
	cardItem = s.listInsert(cardItem, param.Pn, param.ParamStr)
	for _, v := range cardItem {
		ssids = append(ssids, int32(v.ID))
	}
	if len(ssids) > 0 {
		var err error
		if seasonm, err = s.bgm.CardsAll(c, ssids); err != nil {
			log.Error("%+v", err)
		}
	}
	cardParam := &cardm.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromView,
	}
	materials := &cardm.Materials{
		Seams: seasonm,
	}
	op := &operate.Card{FollowType: param.FollowType}
	list := s.cardDealWebItem(cardParam, cardItem, model.EntrancePgcList, model.SmallCoverV4, materials, op)
	return list, nil
}
