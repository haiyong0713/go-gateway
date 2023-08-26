package show

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/region"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) RegionListWeb(c context.Context, plat int8, mid int64, param *region.RegionParam) ([]cardm.Handler, error) {
	regionlist, err := s.reg.RegionDynamic(c, param.Rid, param.Pn, param.Ps)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil
	}
	var (
		cardItem []*ai.Item
		aids     []int64
		arcs     map[int64]*arcgrpc.Arc
		seams    map[int32]*episodegrpc.EpisodeCardsProto
	)
	for _, v := range regionlist {
		if v.Aid == 0 {
			continue
		}
		item := &ai.Item{Goto: model.GotoAv, ID: v.Aid}
		cardItem = append(cardItem, item)
	}
	cardItem = s.listInsert(cardItem, param.Pn, param.ParamStr)
	for _, v := range cardItem {
		aids = append(aids, v.ID)
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
	materials := &cardm.Materials{
		Arcs:               arcs,
		EpisodeCardsProtom: seams,
	}
	cardParam := &cardm.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromView,
	}
	op := &operate.Card{}
	list := s.cardDealWebItem(cardParam, cardItem, model.EntranceRegion, model.SmallCoverV4, materials, op)
	return list, nil
}
