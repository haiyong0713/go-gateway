package show

import (
	"context"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/popular"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) PopularListWeb(c context.Context, mid int64, plat int8, buvid string, param *popular.PopularParam) ([]cardm.Handler, *cardm.Page, error) {
	key := s.popularGroup(mid, buvid)
	cards := s.PopularCardTenList(c, key, param.Pos, _max)
	const (
		_listMax        = 100
		_fTypeOperation = "operation"
	)
	var (
		cardItem []*ai.Item
		aids     []int64
		arcs     map[int64]*arcgrpc.Arc
		seams    map[int32]*episodegrpc.EpisodeCardsProto
	)
	for index, v := range cards {
		if v.Value == 0 || v.Type != model.GotoAv {
			continue
		}
		cardIdx := param.Pos + (index + 1)
		if cardIdx > _listMax && v.FromType != _fTypeOperation {
			continue
		}
		item := &ai.Item{Goto: model.GotoAv, ID: v.Value, TrackID: v.TrackID, Position: param.Pos + (index + 1)}
		cardItem = append(cardItem, item)
	}
	// 插入逻辑
	var pn int
	if param.Pos == 0 {
		pn = 1
	}
	cardItem = s.listInsert(cardItem, pn, param.ParamStr)
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
	list := s.cardDealWebItem(cardParam, cardItem, model.EntrancePopular, model.SmallCoverV4, materials, op)
	itemPage := &cardm.Page{}
	if len(list) > 0 {
		itemPage.Position = list[len(list)-1].Get().Pos
	}
	return list, itemPage, nil
}
