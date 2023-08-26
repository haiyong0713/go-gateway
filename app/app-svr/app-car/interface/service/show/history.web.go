package show

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/history"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/sync/errgroup.v2"

	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

var (
	businesToGotoWebMap = map[string]string{
		_arcStr: model.GotoAvHis,
		_pgcStr: model.GotoPGCEpHis,
	}
)

func (s *Service) cursorList(c context.Context, mid int64, buvid string, param *history.HisParam) (res []*ai.Item, page *cardm.Page, err error) {
	var paramMaxBus string
	if _, ok := businessMap[param.MaxTP]; ok {
		paramMaxBus = businessMap[param.MaxTP]
	}
	businesses := []string{_arcStr, _pgcStr}
	hiss, err := s.his.HistoryCursor(c, mid, param.Max, param.Max, _max, paramMaxBus, buvid, businesses)
	if err != nil {
		log.Error("%+v", err)
		return nil, nil, err
	}
	var (
		cards []*ai.Item
	)
	page = &cardm.Page{}
	for _, v := range hiss {
		gt := businesToGotoWebMap[v.Business]
		switch gt {
		case model.GotoAvHis:
			cards = append(cards, &ai.Item{Goto: gt, ID: v.Oid, ChildID: v.Cid, Card: v})
		case model.GotoPGCEpHis:
			cards = append(cards, &ai.Item{Goto: gt, ID: v.Oid, ChildID: v.Epid, Card: v})
		}
		// 分页
		page.Max = v.Unix
		page.MaxTP = v.Tp
	}
	return cards, page, nil
}

func (s *Service) CursorWeb(c context.Context, plat int8, mid int64, buvid string, param *history.HisParam) (res []cardm.Handler, page *cardm.Page, err error) {
	cardItem, page, err := s.cursorList(c, mid, buvid, param)
	if err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil, nil
	}
	// 插入逻辑
	var pn int
	if param.Max == 0 {
		pn = 1
	}
	cardItem = s.listInsert(cardItem, pn, param.ParamStr)
	var (
		aids     []int64
		epids    []int32
		arcs     map[int64]*arcgrpc.ViewReply
		seams    map[int32]*episodegrpc.EpisodeCardsProto
		cardList []*ai.Item
	)
	for _, v := range cardItem {
		switch v.Goto {
		case model.GotoAvHis:
			aids = append(aids, v.ID)
			cardList = append(cardList, &ai.Item{Goto: v.Goto, ID: v.ID})
		case model.GotoPGCEpHis:
			epids = append(epids, int32(v.ChildID))
			cardList = append(cardList, &ai.Item{Goto: v.Goto, ID: v.ChildID})
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Views(ctx, aids); err != nil {
				log.Error("%+v", err)
				return nil
			}
			return nil
		})
	}
	if len(epids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			seams, err = s.bgm.EpCards(ctx, epids)
			if err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	materials := &cardm.Materials{
		ViewReplym: arcs,
		Epms:       seams,
	}
	cardParam := &cardm.CardParam{
		Plat:         plat,
		Mid:          mid,
		FromType:     model.FromView,
		IsBackUpCard: true,
	}
	op := &operate.Card{}
	list := s.cardDealWebItem(cardParam, cardList, model.EntranceHistoryRecord, model.SmallCoverV4, materials, op)
	return list, page, nil
}
