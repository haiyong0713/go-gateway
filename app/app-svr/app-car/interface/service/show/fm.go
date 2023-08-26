package show

import (
	"context"

	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/popular"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) FmList(c context.Context, mid int64, plat int8, buvid string, param *popular.PopularParam) ([]cardm.Handler, *cardm.Page, error) {
	// 热门
	const (
		ps = 20
	)
	key := s.popularGroup(mid, buvid)
	cards := s.PopularCardTenList(c, key, param.Pos, ps)
	if len(cards) == 0 {
		return []cardm.Handler{}, nil, nil
	}
	var (
		feedCards []*ai.Item
	)
	// 转换成统一格式
	for _, v := range cards {
		feedCard := v.PopularCardToAiChange()
		feedCard.Entrance = model.EntrancePopular
		feedCards = append(feedCards, feedCard)
	}
	var (
		aids  []int64
		seams map[int32]*episodegrpc.EpisodeCardsProto
	)
	arcm := map[int64]*arcgrpc.ArcPlayer{}
	for _, v := range feedCards {
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		}
	}
	group := errgroup.WithContext(c)
	// 获取稿件物料
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) error {
			// 不要秒开信息
			arcs, err := s.arc.Archives(ctx, aids)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			for _, v := range arcs {
				arcm[v.Aid] = &arcgrpc.ArcPlayer{Arc: v}
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsByAidsAll(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	cardParam := &card.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromList,
		MobiApp:  param.MobiApp,
		Build:    param.Build,
	}
	materials := &card.Materials{
		ArcPlayers:         arcm,
		EpisodeCardsProtom: seams,
	}
	op := &operate.Card{}
	list := s.cardDealItem(cardParam, feedCards, model.EntrancePopular, model.FmV1, materials, op)
	if len(list) == 0 {
		return []cardm.Handler{}, nil, nil
	}
	itemPage := &cardm.Page{
		Position: list[len(list)-1].Get().Pos,
	}
	return list, itemPage, nil
}
