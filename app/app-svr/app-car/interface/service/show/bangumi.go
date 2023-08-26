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

const (
	_followTypeBangumi   = "bangumi"
	_followTypeCinema    = "cinema"
	_followTypeDomestic  = "domestic"
	_followTypeCinemaDoc = "cinema_doc"
)

func (s *Service) MyAnime(c context.Context, plat int8, mid int64, param *bangumi.MyAnimeParam) ([]cardm.Handler, error) {
	var followType int
	switch param.FollowType {
	case _followTypeBangumi:
		followType = 1
	case _followTypeCinema:
		followType = 2
	}
	follows, err := s.bgm.MyFollows(c, mid, followType, param.Pn, param.Ps)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil
	}
	// 插入逻辑
	if param.Pn == 1 && param.ParamStr != "" {
		anime, ok := cardm.FromPGCFollow(param.ParamStr)
		if ok {
			var isok bool
			for _, f := range follows {
				if f.SeasonId == anime.SeasonId {
					isok = true
					break
				}
			}
			// 不相同直接插入第一位
			if !isok {
				cards := []*cardappgrpc.CardSeasonProto{anime}
				cards = append(cards, follows...)
				follows = cards
			}
		}
	}
	var (
		ssids   []int32
		seasonm map[int32]*seasongrpc.CardInfoProto
	)
	for _, f := range follows {
		if f.SeasonId == 0 {
			continue
		}
		ssids = append(ssids, f.SeasonId)
	}
	if len(ssids) > 0 {
		var err error
		if seasonm, err = s.bgm.CardsAll(c, ssids); err != nil {
			log.Error("%+v", err)
		}
	}
	is := []cardm.Handler{}
	for _, f := range follows {
		var (
			r        = &ai.Item{Goto: model.GotoPGC, ID: int64(f.SeasonId)}
			op       = &operate.Card{FollowType: param.FollowType}
			main     interface{}
			cardType model.CardType
		)
		op.From(model.CardGt(r.Goto), model.EntranceMyAnmie, r.ID, plat, param.Build, param.MobiApp)
		if f.Progress != nil {
			op.Cid = int64(f.Progress.EpisodeId)
		}
		materials := &cardm.Materials{
			Seams: seasonm,
			Prune: cardm.PGCFollowPrune(f),
		}
		switch param.FromType {
		case model.FromList:
			main = f
			cardType = model.SmallCoverV2
		default:
			main = seasonm
			cardType = model.SmallCoverV4
		}
		h := cardm.Handle(plat, model.CardGt(r.Goto), cardType, r, materials)
		if h == nil || !h.From(main, op) {
			continue
		}
		is = append(is, h)
	}
	if len(is) == 0 {
		return []cardm.Handler{}, nil
	}
	return is, nil
}

func (s *Service) List(c context.Context, plat int8, mid int64, buvid string, param *bangumi.ListParam) ([]cardm.Handler, error) {
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
	}
	// 实际PGC这接口一期没有分页，第一页就返回了所有数据
	if param.Pn > 1 {
		return []cardm.Handler{}, nil
	}
	list, err := s.bgm.Module(c, followType, param.MobiApp, buvid)
	if err != nil {
		log.Warn("%+v", err)
		return []cardm.Handler{}, nil
	}
	// 插入逻辑
	if param.Pn == 1 && param.ParamStr != "" {
		mlist, ok := cardm.FromPGCModule(param.ParamStr)
		if ok {
			var isok bool
			for _, f := range list {
				if f.SeasonID == mlist.SeasonID {
					isok = true
					break
				}
			}
			// 不相同直接插入第一位
			if !isok {
				cards := []*bangumi.Module{mlist}
				cards = append(cards, list...)
				list = cards
			}
		}
	}
	var (
		ssids   []int32
		seasonm map[int32]*seasongrpc.CardInfoProto
	)
	for _, l := range list {
		if l.SeasonID == 0 {
			continue
		}
		ssids = append(ssids, l.SeasonID)
	}
	if len(ssids) > 0 {
		var err error
		if seasonm, err = s.bgm.CardsAll(c, ssids); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
	}
	is := []cardm.Handler{}
	for _, l := range list {
		var (
			r        = &ai.Item{Goto: model.GotoPGC, ID: int64(l.SeasonID)}
			op       = &operate.Card{FollowType: param.FollowType}
			main     interface{}
			cardType model.CardType
		)
		op.From(model.CardGt(r.Goto), model.EntrancePgcList, r.ID, plat, param.Build, param.MobiApp)
		materials := &cardm.Materials{
			Seams: seasonm,
			Prune: cardm.PGCModulePrune(l),
		}
		main = seasonm
		if param.FromType != model.FromList {
			cardType = model.SmallCoverV4
		}
		h := cardm.Handle(plat, model.CardGt(r.Goto), cardType, r, materials)
		if h == nil || !h.From(main, op) {
			continue
		}
		is = append(is, h)
	}
	if len(is) == 0 {
		return []cardm.Handler{}, nil
	}
	if param.FromType == model.FromList && len(is) > 20 {
		is = is[:20]
	}
	return is, nil
}

func (s *Service) PGCShow(c context.Context, mid int64, plat int8, buvid string, param *show.ShowParam) ([]*show.Item, error) {
	items := []*show.Item{}
	const (
		_bangumi = "bangumi"
		_movie   = "movie"
	)
	var (
		cardsTypes = []string{_myBangumi, _bangumi, _domestic, _myCinema, _cinema, _cinemaDoc}
		cards      = map[string][]cardm.Handler{}
		mutex      sync.Mutex
	)
	group := errgroup.WithContext(c)
	if param.FollowType == _bangumi {
		if mid > 0 {
			group.Go(func(ctx context.Context) (err error) {
				items, err := s.MyAnime(ctx, plat, mid, &bangumi.MyAnimeParam{DeviceInfo: param.DeviceInfo, FromType: model.FromList, FollowType: _followTypeBangumi, Pn: _defaultPn, Ps: _defaultPs})
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				if len(items) > 0 {
					mutex.Lock()
					cards[_myBangumi] = items
					mutex.Unlock()
				}
				return nil
			})
		}
		group.Go(func(ctx context.Context) (err error) {
			items, err := s.List(ctx, plat, mid, buvid, &bangumi.ListParam{DeviceInfo: param.DeviceInfo, FromType: model.FromList, FollowType: _followTypeBangumi, Pn: _defaultPn, Ps: _defaultPs})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			if len(items) > 0 {
				mutex.Lock()
				cards[_bangumi] = items
				mutex.Unlock()
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			items, err := s.List(ctx, plat, mid, buvid, &bangumi.ListParam{DeviceInfo: param.DeviceInfo, FromType: model.FromList, FollowType: _followTypeDomestic, Pn: _defaultPn, Ps: _defaultPs})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			if len(items) > 0 {
				mutex.Lock()
				cards[_domestic] = items
				mutex.Unlock()
			}
			return nil
		})
	}
	if param.FollowType == _movie {
		if mid > 0 {
			group.Go(func(ctx context.Context) (err error) {
				items, err := s.MyAnime(ctx, plat, mid, &bangumi.MyAnimeParam{DeviceInfo: param.DeviceInfo, FromType: model.FromList, FollowType: _followTypeCinema, Pn: _defaultPn, Ps: _defaultPs})
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				if len(items) > 0 {
					mutex.Lock()
					cards[_myCinema] = items
					mutex.Unlock()
				}
				return nil
			})
		}
		group.Go(func(ctx context.Context) (err error) {
			items, err := s.List(ctx, plat, mid, buvid, &bangumi.ListParam{DeviceInfo: param.DeviceInfo, FromType: model.FromList, FollowType: _followTypeCinema, Pn: _defaultPn, Ps: _defaultPs})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			if len(items) > 0 {
				mutex.Lock()
				cards[_cinema] = items
				mutex.Unlock()
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			items, err := s.List(ctx, plat, mid, buvid, &bangumi.ListParam{DeviceInfo: param.DeviceInfo, FromType: model.FromList, FollowType: _followTypeCinemaDoc, Pn: _defaultPn, Ps: _defaultPs})
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			if len(items) > 0 {
				mutex.Lock()
				cards[_cinemaDoc] = items
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	for _, ct := range cardsTypes {
		card, ok := cards[ct]
		if !ok {
			continue
		}
		item := &show.Item{Type: ct, Items: card}
		item.FromItem2("")
		items = append(items, item)
	}
	return items, nil
}
