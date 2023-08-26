package show

import (
	"context"
	"hash/crc32"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-car/ecode"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/popular"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) popularGroup(mid int64, buvid string) int {
	if mid > 0 {
		// nolint:gomnd
		return int((mid / 1000) % 10)
	}
	// nolint:gomnd
	return int((crc32.ChecksumIEEE([]byte(buvid)) / 1000) % 10)
}

func (s *Service) Index(c context.Context, mid int64, plat int8, buvid string, param *popular.PopularParam) ([]cardm.Handler, *cardm.Page, error) {
	var (
		ps = 20
	)
	key := s.popularGroup(mid, buvid)
	cards := s.PopularCardTenList(c, key, param.Pos, ps)
	// 插入逻辑
	if param.Pos == 0 && param.ParamStr != "" {
		pcard, ok := cardm.FromPopular(param.ParamStr)
		if ok {
			var isok bool
			for _, card := range cards {
				if card.Value == pcard.Value {
					// 有相同数据
					isok = true
					break
				}
			}
			// 不相同直接插入第一位
			if !isok {
				tmp := []*popular.PopularCard{pcard}
				tmp = append(tmp, cards...)
				cards = tmp
			}
		}
	}
	if len(cards) == 0 {
		return []cardm.Handler{}, nil, nil
	}
	is := s.dealItem(c, plat, ps, param, cards)
	if len(is) == 0 {
		return []cardm.Handler{}, nil, nil
	}
	itemPage := &cardm.Page{
		Position: is[len(is)-1].Get().Pos,
	}
	return is, itemPage, nil
}

func (s *Service) dealItem(c context.Context, plat int8, ps int, param *popular.PopularParam, cards []*popular.PopularCard) []cardm.Handler {
	var (
		max             = 100
		_fTypeOperation = "operation"
		aids            []int64
		arcs            map[int64]*arcgrpc.Arc
		feedcards       []*popular.PopularCard
		seams           map[int32]*episodegrpc.EpisodeCardsProto
	)
	for p, ca := range cards {
		cardIdx := param.Pos + (p + 1)
		if cardIdx > max && ca.FromType != _fTypeOperation {
			continue
		}
		tmp := &popular.PopularCard{}
		*tmp = *ca
		tmp.Idx = cardIdx
		feedcards = append(feedcards, tmp)
		switch ca.Type {
		case model.GotoAv:
			if ca.Value == 0 {
				continue
			}
			aids = append(aids, ca.Value)
		}
		if len(feedcards) == ps {
			break
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) != 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsByAids(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	is := []cardm.Handler{}
	for _, ca := range feedcards {
		var (
			r        = ca.PopularCardToAiChange()
			main     interface{}
			cardType model.CardType
			op       = &operate.Card{TrackID: ca.TrackID}
		)
		switch param.FromType {
		case model.FromList:
			cardType = model.SmallCoverV1
		default:
			cardType = model.SmallCoverV4
		}
		op.From(model.CardGt(r.Goto), model.EntrancePopular, r.ID, plat, param.Build, param.MobiApp)
		switch r.Goto {
		case model.GotoAv:
			// 如果当前稿件是pgc视频，转成pgc卡片处理，否则还是稿件卡片
			main = arcs
			if _, ok := seams[int32(r.ID)]; ok {
				main = seams
				r.Goto = model.GotoPGC
			}
		}
		materials := &cardm.Materials{
			Prune: cardm.PopularPrune(ca),
		}
		h := cardm.Handle(plat, model.CardGt(r.Goto), cardType, r, materials)
		if h == nil || !h.From(main, op) {
			continue
		}
		// 过滤互动视频
		if h.Get().Filter == model.FilterAttrBitSteinsGate {
			continue
		}
		h.Get().FromType = ca.FromType
		h.Get().Pos = ca.Idx
		is = append(is, h)
	}
	if len(is) == 0 {
		return []cardm.Handler{}
	}
	return is
}

// PopularCardList cards
func (s *Service) PopularCardTenList(c context.Context, i, index, ps int) (res []*popular.PopularCard) {
	var err error
	if res, err = s.dao.PopularCardTenCache(c, i, index, ps); err != nil {
		log.Error("%+v", err)
		return
	}
	return
}

func (s *Service) MediaPopular(c context.Context, param *popular.MediaPopularParam) ([]*cardm.MediaItem, error) {
	start := (param.Pn - 1) * param.Ps
	cards := s.PopularCardTenList(c, 0, start, param.Ps)
	var (
		aids  []int64
		arcs  map[int64]*arcgrpc.Arc
		seams map[int32]*episodegrpc.EpisodeCardsProto
	)
	for _, v := range cards {
		if v.Value == 0 {
			continue
		}
		switch v.Type {
		case model.GotoAv:
			aids = append(aids, v.Value)
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) != 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsByAids(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	is := []*cardm.MediaItem{}
	for _, v := range cards {
		var (
			main interface{}
			gt   string
		)
		if v.Value == 0 {
			continue
		}
		switch v.Type {
		case model.GotoAv:
			// 如果当前稿件是pgc视频，转成pgc卡片处理，否则还是稿件卡片
			main = arcs
			gt = model.GotoAv
			if _, ok := seams[int32(v.Value)]; ok {
				main = seams
				gt = model.GotoPGC
			}
		}
		materials := &cardm.Materials{Prune: cardm.GtPrune(model.GotoAv, v.Value)}
		i := &cardm.MediaItem{}
		ok := i.FromItem(v.Value, gt, main, materials)
		if !ok {
			continue
		}
		is = append(is, i)
	}
	if len(is) == 0 {
		return nil, xecode.AppMediaNotData
	}
	return is, nil
}
