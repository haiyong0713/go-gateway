package show

import (
	"context"
	"strconv"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/search"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

const (
	_suggestionJump     = 99
	_suggestionJumpUser = 81
	_suggestionJumpPGC  = 82
	_defaultUpperPs     = 3
)

var (
	suggestionType = map[int]struct{}{
		_suggestionJump:     {},
		_suggestionJumpUser: {},
		_suggestionJumpPGC:  {},
	}
)

func (s *Service) Suggest(c context.Context, plat int8, mid int64, buvid string, param *search.SearchSuggestParam) (res []*search.SuggestItem, args *search.SearchArgs, err error) {
	suggest, err := s.srch.Suggest(c, plat, mid, param.Platform, buvid, param.Keyword, param.MobiApp, param.Device, param.Build, param.Highlight)
	if err != nil {
		log.Error("%+v", err)
		return []*search.SuggestItem{}, nil, nil
	}
	if suggest == nil || len(suggest.Result) == 0 {
		return []*search.SuggestItem{}, nil, nil
	}
	args = &search.SearchArgs{}
	args.FromSuggestArgs(suggest)
	for _, v := range suggest.Result {
		// 屏蔽所有特殊跳转
		if _, ok := suggestionType[v.TermType]; ok {
			continue
		}
		item := &search.SuggestItem{}
		item.FromSuggest(v)
		res = append(res, item)
	}
	return res, args, nil
}

// nolint: gocognit
func (s *Service) Search(c context.Context, plat int8, mid int64, buvid string, param *search.SearchParam) (res []cardm.Handler, up []*search.UpItem, args *search.SearchArgs, err error) {
	var (
		cardItem, feedItem []*ai.Item
		upItems            []*search.UpItem
		aids, mids         []int64
		ssids, epids       []int32
		arcs               map[int64]*arcgrpc.Arc
		seams              map[int32]*seasongrpc.CardInfoProto
		seamAids           map[int32]*episodegrpc.EpisodeCardsProto
		epms               map[int32]*episodegrpc.EpisodeCardsProto
		accms              map[int64]*accountgrpc.Card
		all                *search.Search
		upReply            []*search.User
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) (err error) {
		all, err = s.srch.Search(ctx, mid, 0, param.Pn, param.Ps, param.Keyword, buvid)
		if err != nil {
			log.Error("%+v", err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) (err error) {
		upReply, err = s.srch.Upper(ctx, mid, _defaultPn, _defaultUpperPs, param.Keyword, buvid)
		if err != nil {
			log.Error("%+v", err)
			return err
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil, nil, nil
	}

	if all == nil || all.Result == nil {
		return []cardm.Handler{}, nil, nil, nil
	}
	args = &search.SearchArgs{}
	args.SearchArgsFrom(all)
	// 转换成统一结构体
	// pgc
	for _, v := range all.Result.MediaBangumi {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	for _, v := range all.Result.MediaFt {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	// archive
	for _, v := range all.Result.Video {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoAv, ID: int64(v.ID)})
	}
	// 拆入逻辑
	if param.Pn == 1 && param.ParamStr != "" {
		gt, id, childID, ok := cardm.FromGtPrune(param.ParamStr)
		if ok {
			feedItem = append(feedItem, &ai.Item{Goto: gt, ID: id, ChildID: childID})
			for _, v := range cardItem {
				if v.Goto == gt && v.ID == id {
					continue
				}
				feedItem = append(feedItem, v)
			}
		}
	} else {
		feedItem = append(feedItem, cardItem...)
	}
	for _, v := range feedItem {
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
	for _, v := range upReply {
		mids = append(mids, v.Mid)
	}
	group = errgroup.WithContext(c)
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
	if len(mids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if accms, err = s.acc.Cards3All(ctx, mids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return []cardm.Handler{}, nil, nil, nil
	}
	materials := &card.Materials{
		Arcs:               arcs,
		EpisodeCardsProtom: seamAids,
		Seams:              seams,
		Epms:               epms,
	}
	cardParam := &card.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: param.FromType,
		MobiApp:  param.MobiApp,
		Build:    param.Build,
	}
	op := &operate.Card{KeyWord: param.Keyword}
	items := s.cardSearchDealItem(cardParam, feedItem, model.EntranceCommonSearch, materials, op)
	// up
	for _, v := range upReply {
		acc, ok := accms[v.Mid]
		if !ok {
			continue
		}
		upItem := &search.UpItem{
			Mid:   v.Mid,
			Name:  v.Name,
			Desc1: model.FanIntString(int32(v.Fans)) + " " + model.VedioIntString(int32(v.Videos)),
			Desc2: v.Usign,
			URI:   model.FillURI(model.GotoSpace, plat, param.Build, strconv.FormatInt(v.Mid, 10), nil),
			Face:  acc.Face,
		}
		upItems = append(upItems, upItem)
	}
	return items, upItems, args, nil
}

func (s *Service) cardSearchDealItem(param *card.CardParam, feedCard []*ai.Item, entrance string, materials *card.Materials, op *operate.Card) []cardm.Handler {
	is := []cardm.Handler{}
	for _, v := range feedCard {
		var (
			main     interface{}
			cardType model.CardType
		)
		op.TrackID = v.TrackID
		switch param.FromType {
		case model.FromView:
			cardType = model.SmallCoverV4
		default:
			if param.FromType == model.FromVoice {
				entrance = model.EntranceVoiceSearch
			}
			switch v.Goto {
			case model.GotoAv:
				cardType = model.SmallCoverV1
			}
		}
		op.From(model.CardGt(v.Goto), entrance, v.ID, param.Plat, param.Build, param.MobiApp)
		materials.Prune = cardm.GtPrune(v.Goto, v.ID)
		switch v.Goto {
		case model.GotoAv:
			main = materials.Arcs
			if _, ok := materials.EpisodeCardsProtom[int32(v.ID)]; ok {
				main = materials.EpisodeCardsProtom
				v.Goto = model.GotoPGC
			}
		case model.GotoPGC:
			main = materials.Seams
			if _, ok := materials.Epms[int32(v.ChildID)]; ok {
				op.Cid = v.ChildID
			}
		}
		h := cardm.Handle(param.Plat, model.CardGt(v.Goto), cardType, v, materials)
		if h == nil || !h.From(main, op) {
			continue
		}
		// 过滤互动视频
		if h.Get().Filter == model.FilterAttrBitSteinsGate {
			continue
		}
		is = append(is, h)
	}
	if len(is) == 0 {
		return []cardm.Handler{}
	}
	return is
}
