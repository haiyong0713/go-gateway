package service

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/log"
	"go-common/library/sync/errgroup"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	chmdl "go-gateway/app/web-svr/web/interface/model/channel"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	appCardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
)

func (s *Service) ChannelDetail(c context.Context, params *chmdl.Param) (res *chmdl.Detail, err error) {
	var (
		detail        *channelgrpc.ChannelCard
		selectSort    []*channelgrpc.FeaturedOption
		seasonm       map[int64]*appCardgrpc.SeasonCards
		tabs          []*channelgrpc.ShowTab
		labels        []*channelgrpc.ShowLabel
		defaultTabIdx int64
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		reply, err := s.dao.ChannelDetail(ctx, params.MID, params.ChannelID)
		if err != nil {
			log.Error("%v", err)
			return err
		}
		detail, selectSort, tabs, labels, defaultTabIdx = reply.Channel, reply.FeaturedOptions, reply.Tabs, reply.Labels, reply.DefaultTabIdx
		return
	})
	if s.c.Switch != nil && s.c.Switch.DetailVerify {
		g.Go(func() (err error) {
			if seasonm, err = s.dao.TagOGV(ctx, []int64{params.ChannelID}); err != nil {
				log.Error("%v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	if detail == nil {
		log.Error("detail nil")
		return
	}
	res = &chmdl.Detail{}
	res.FormDetail(detail, selectSort, seasonm, tabs, labels, defaultTabIdx)
	return
}

func (s *Service) ChannelMultiple(c context.Context, params *chmdl.Param) (res *chmdl.ChannelListResult, err error) {
	var (
		cSort   channelgrpc.TotalSortType
		isBadge bool
	)
	switch params.Sort {
	case "hot":
		cSort = channelgrpc.TotalSortType_SORT_BY_HOT
		isBadge = true
	case "view":
		cSort = channelgrpc.TotalSortType_SORT_BY_VIEW_CNT
	case "new":
		cSort = channelgrpc.TotalSortType_SORT_BY_PUB_TIME
	}
	var (
		cards *channelgrpc.ResourceListReply
		arg   = &channelgrpc.ResourceListReq{ChannelId: params.ChannelID, TabType: channelgrpc.TabType_TAB_TYPE_TOTAL, SortType: cSort, Offset: params.Offset, PageSize: 20, Mid: params.MID}
	)
	if cards, err = s.dao.ResourceList(c, arg); err != nil {
		log.Error("%v", err)
		return
	}
	res = &chmdl.ChannelListResult{Offset: cards.GetNextOffset()}
	if cards.GetHasMore() {
		res.HasMore = 1
	}
	if cards.GetUpdateHotCnt() > 0 {
		res.Label = fmt.Sprintf("有%s", chmdl.StatString(cards.GetUpdateHotCnt(), "个视频更新"))
	}
	var hs []card.Handler
	if hs, err = s.dealItems(c, cards, params.ChannelID, params.MID, isBadge, params.Theme, false); err != nil {
		log.Error("%v", err)
		return
	}
	res.Items = hs
	return
}

func (s *Service) ChannelSelected(c context.Context, params *chmdl.Param) (res *chmdl.ChannelListResult, err error) {
	var (
		cFilter, _ = strconv.Atoi(params.Sort)
		cards      *channelgrpc.ResourceListReply
		arg        = &channelgrpc.ResourceListReq{ChannelId: params.ChannelID, TabType: channelgrpc.TabType_TAB_TYPE_FEATURED, FilterType: int32(cFilter), Offset: params.Offset, PageSize: 20, Mid: params.MID}
	)
	if cards, err = s.dao.ResourceList(c, arg); err != nil {
		log.Error("%v", err)
		return
	}
	res = &chmdl.ChannelListResult{Offset: cards.GetNextOffset()}
	if cards.GetHasMore() {
		res.HasMore = 1
	}
	if cards.GetUpdateHotCnt() > 0 {
		res.Label = fmt.Sprintf("有%s", chmdl.StatString(cards.GetUpdateHotCnt(), "个视频更新"))
	}
	var (
		hs    []card.Handler
		isOGV bool
	)
	if (params.Offset == "" || params.Offset == "0") && cFilter == 0 {
		isOGV = true
	}
	if hs, err = s.dealItems(c, cards, params.ChannelID, params.MID, true, params.Theme, isOGV); err != nil {
		log.Error("%v", err)
		return
	}
	res.Items = hs
	return
}

// nolint: gocognit
func (s *Service) dealItems(c context.Context, cards *channelgrpc.ResourceListReply, channelID, mid int64, isBadge bool, theme string, isOGV bool) (is []card.Handler, err error) {
	var (
		cs = cards.GetCards()
		// aids为全量视频卡；oids是包含双列视频卡
		aids, oids   []int64
		videoBadges  = make(map[int64]*operate.ChannelBadge)
		customBadges = make(map[int64]*operate.ChannelBadge)
	)
	for _, card := range cs {
		switch card.GetCardType() {
		case chmdl.CardTypeVideo:
			if card.GetVideoCard().Rid != 0 {
				aids = append(aids, card.GetVideoCard().Rid)
				oids = append(oids, card.GetVideoCard().Rid)
				if isBadge && card.GetVideoCard().GetBadgeTitle() != "" && card.GetVideoCard().GetBadgeBackground() != "" {
					videoBadges[card.GetVideoCard().Rid] = &operate.ChannelBadge{
						Text:  card.GetVideoCard().GetBadgeTitle(),
						Cover: card.GetVideoCard().GetBadgeBackground(),
					}
				}
			}
		case chmdl.CardTypeCustom:
			if customs := card.GetCustomCard(); customs != nil {
				for _, card := range customs.GetCards() {
					if card.GetRid() != 0 {
						aids = append(aids, card.GetRid())
						if isBadge && card.GetBadgeTitle() != "" && card.GetBadgeBackground() != "" {
							customBadges[card.GetRid()] = &operate.ChannelBadge{
								Text:  card.GetBadgeTitle(),
								Cover: card.GetBadgeBackground(),
							}
						}
					}
				}
			}
		case chmdl.CardTypeRank:
			if ranks := card.GetRankCard(); ranks != nil {
				for _, rank := range ranks.GetCards() {
					if rank.GetRid() != 0 {
						aids = append(aids, rank.GetRid())
					}
				}
			}
		default:
			log.Warn("channel dealItem unknown type %+v", card)
		}
	}
	var (
		arcs    map[int64]*archivegrpc.ArcPlayer
		isFav   map[int64]bool
		coins   map[int64]int64
		seasonm map[int64]*appCardgrpc.SeasonCards
	)
	g, ctx := errgroup.WithContext(c)
	if len(aids) > 0 {
		g.Go(func() (err error) {
			if arcs, err = s.arcs(ctx, aids); err != nil {
				log.Error("%v", err)
				err = nil
			}
			return
		})
		if len(oids) > 0 {
			if mid > 0 {
				g.Go(func() (err error) {
					if coins, err = s.dao.IsCoins(ctx, oids, mid); err != nil {
						log.Error("%v", err)
						err = nil
					}
					return
				})
			}
			g.Go(func() (err error) {
				if isFav, err = s.dao.IsFavoreds(ctx, mid, aids); err != nil {
					log.Error("%v", err)
					err = nil
				}
				return
			})
		}
		if s.c.Switch != nil && s.c.Switch.DetailVerify && isOGV && cards.GetPGC() {
			g.Go(func() (err error) {
				if seasonm, err = s.dao.TagOGV(ctx, []int64{channelID}); err != nil {
					log.Error("%v", err)
				}
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		cardTotal int
		position  int64
	)
	is = make([]card.Handler, 0, len(cs))
	for _, season := range seasonm {
		if season == nil || len(season.GetCards()) == 0 {
			continue
		}
		op := &operate.Card{
			ID:      channelID,
			Param:   strconv.FormatInt(channelID, 10),
			Channel: &operate.Channel{Position: position},
		}
		if s.c.Switch != nil {
			if s.c.Switch.ListOGVFold && len(season.GetCards()) > 3 {
				op.Channel.HasFold = true
			}
			op.Channel.HasMore = s.c.Switch.ListOGVMore
		}
		op.From(cdm.CardGt("channel_ogv_large"), channelID, 0, 0, 0, "")
		h := cardm.Handle(0, cdm.CardGt("channel_ogv_large"), cdm.CardType("channel_ogv_large"), cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
		if h == nil {
			continue
		}
		if err := h.From(season, op); err != nil {
			log.Error("%+v", err)
		}
		if h.Get() == nil || !h.Get().Right {
			continue
		}
		position = h.Get().Idx
		is, cardTotal = s.dealAppend(is, h, cardTotal)
	}
	for _, card := range cs {
		switch card.GetCardType() {
		case chmdl.CardTypeVideo:
			if card.GetVideoCard() == nil || card.GetVideoCard().Rid == 0 {
				continue
			}
			op := &operate.Card{
				ID:      card.GetVideoCard().Rid,
				Param:   strconv.FormatInt(card.GetVideoCard().GetRid(), 10),
				Channel: &operate.Channel{Badges: videoBadges, IsFav: isFav, Coins: coins, Position: position},
			}
			h := cardm.Handle(0, cdm.CardGt("channel_new_detail"), "", cdm.ColumnSvrDouble, nil, nil, nil, nil, nil, nil, nil)
			if h == nil {
				continue
			}
			if err := h.From(arcs, op); err != nil {
				log.Error("%+v", err)
			}
			if h.Get() == nil || !h.Get().Right {
				continue
			}
			position = h.Get().Idx
			is, cardTotal = s.dealAppend(is, h, cardTotal)
		case chmdl.CardTypeCustom:
			var (
				customs = card.GetCustomCard()
				cis     []*operate.Card
			)
			if customs == nil {
				continue
			}
			for _, card := range customs.GetCards() {
				if card == nil || card.GetRid() == 0 {
					continue
				}
				arc, ok := arcs[card.GetRid()]
				if !ok || arc == nil {
					continue
				}
				cis = append(cis, &operate.Card{ID: card.GetRid()})
			}
			if len(cis) == 0 {
				continue
			}
			op := &operate.Card{
				ID:      channelID,
				Param:   strconv.FormatInt(channelID, 10),
				Items:   cis,
				Channel: &operate.Channel{Position: position, Badges: customBadges},
			}
			if customs.GetDetail() != nil {
				op.Title = customs.GetDetail().GetName()
				op.Channel.CustomDesc = customs.GetDetail().GetJumpDesc()
				op.Channel.CustomURI = customs.GetDetail().GetJumpUrl()
			}
			op.From(cdm.CardGt("channel_new_detail_custom"), channelID, 0, 0, 0, "")
			h := cardm.Handle(0, cdm.CardGt("channel_new_detail_custom"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
			if h == nil {
				continue
			}
			if err := h.From(arcs, op); err != nil {
				log.Error("%+v", err)
			}
			if h.Get() == nil || !h.Get().Right {
				continue
			}
			position = h.Get().Idx
			is, cardTotal = s.dealAppend(is, h, cardTotal)
		case chmdl.CardTypeRank:
			var (
				ranks = card.GetRankCard()
				cis   []*operate.Card
			)
			if ranks == nil {
				continue
			}
			for _, rank := range ranks.GetCards() {
				if rank == nil || rank.GetRid() == 0 {
					continue
				}
				arc, ok := arcs[rank.GetRid()]
				if !ok || arc == nil {
					continue
				}
				cis = append(cis, &operate.Card{ID: rank.GetRid()})
			}
			if len(cis) == 0 {
				continue
			}
			op := &operate.Card{
				ID:      channelID,
				Param:   strconv.FormatInt(channelID, 10),
				Items:   cis,
				Channel: &operate.Channel{Position: position},
			}
			if ranks.GetDetail() != nil {
				op.Title = ranks.GetDetail().GetTitle()
				op.Channel.CustomDesc = ranks.GetDetail().GetJumpDesc()
				op.Channel.CustomURI = fmt.Sprintf(chmdl.RankURL, channelID, url.QueryEscape(theme))
				op.Channel.RankType = ranks.Detail.SortType
			}
			op.From(cdm.CardGt("channel_new_detail_rank"), channelID, 0, 0, 0, "")
			h := cardm.Handle(0, cdm.CardGt("channel_new_detail_rank"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
			if h == nil {
				continue
			}
			if err := h.From(arcs, op); err != nil {
				log.Error("%+v", err)
			}
			if h.Get() == nil || !h.Get().Right {
				continue
			}
			position = h.Get().Idx
			is, cardTotal = s.dealAppend(is, h, cardTotal)
		default:
		}
	}
	return
}

func (s *Service) dealAppend(rs []card.Handler, h card.Handler, cardTotal int) (is []card.Handler, total int) {
	if h.Get().CardLen == 0 {
		if cardTotal%2 == 1 {
			is = card.SwapTwoItem(rs, h)
		} else {
			is = append(rs, h)
		}
	} else {
		is = append(rs, h)
	}
	total = cardTotal + h.Get().CardLen
	return
}

func (s *Service) arcs(c context.Context, aids []int64) (map[int64]*archivegrpc.ArcPlayer, error) {
	if len(aids) == 0 {
		return nil, nil
	}
	reply, err := s.dao.Arcs(c, aids)
	if err != nil {
		return nil, err
	}
	arcm := map[int64]*archivegrpc.ArcPlayer{}
	for aid, a := range reply {
		arcm[aid] = &archivegrpc.ArcPlayer{Arc: a}
	}
	return arcm, nil
}
