package show

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/relate"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

func (s *Service) RelateWeb(c context.Context, plat int8, mid int64, buvid string, param *relate.RelateParam) ([]cardm.Handler, error) {
	var (
		topItem, feedItem []*ai.Item
		aids              []int64
		ssids, epids      []int32
		arcs              map[int64]*arcgrpc.Arc
		seamAids          map[int32]*episodegrpc.EpisodeCardsProto
		epms              map[int32]*episodegrpc.EpisodeCardsProto
		seams             map[int32]*seasongrpc.CardInfoProto
	)
	gt, id, child, ok := cardm.FromGtPrune(param.ParamStr)
	if !ok {
		return []cardm.Handler{}, nil
	}
	topItem = append(topItem, &ai.Item{Goto: gt, ID: id, ChildID: child})
	switch gt {
	case model.GotoAv:
		relates, err := s.rcmd.Relate(c, mid, id, buvid)
		if err != nil {
			log.Error("%+v", err)
		}
		// 转统一类型
		for _, v := range relates {
			if v.Goto != model.GotoAv {
				continue
			}
			feedItem = append(feedItem, &ai.Item{Goto: model.GotoAv, ID: v.ID})
		}
	}
	// 获取ID
	cards := []*ai.Item{}
	cards = append(cards, topItem...)
	cards = append(cards, feedItem...)
	for _, v := range cards {
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
	// 获取物料
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
		return []cardm.Handler{}, nil
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
		FromType: model.FromView,
		MobiApp:  param.MobiApp,
		Build:    param.Build,
	}
	// 顶部
	op := &operate.Card{}
	cardType := model.SmallCoverV4
	items := s.cardDealWebItem(cardParam, cards, model.EntranceRelate, cardType, materials, op)
	return items, nil
}
