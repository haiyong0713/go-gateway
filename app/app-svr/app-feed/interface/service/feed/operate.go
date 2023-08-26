package feed

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
)

func (s *Service) convergeCard(_ context.Context, limit int, ids ...int64) (cardm map[int64]*operate.Card, aids, roomIDs, metaIDs []int64) {
	if len(ids) == 0 {
		return
	}
	cardm = make(map[int64]*operate.Card, len(ids))
	for i, id := range ids {
		if o, ok := s.convergeCache[id]; ok {
			card := &operate.Card{}
			card.FromConverge(o)
			cardm[id] = card
			for _, item := range card.Items {
				switch item.Goto {
				case model.GotoAv:
					if item.ID != 0 {
						aids = append(aids, item.ID)
					}
				case model.GotoLive:
					if item.ID != 0 {
						roomIDs = append(roomIDs, item.ID)
					}
				case model.GotoArticle:
					if item.ID != 0 {
						metaIDs = append(metaIDs, item.ID)
					}
				default:
				}
				if i == limit-1 {
					break
				}
			}
		}
	}
	return
}

func (s *Service) channelRcmdCard(_ context.Context, ids ...int64) (cardm map[int64]*operate.Card, aids, tids []int64) {
	if len(ids) == 0 {
		return
	}
	cardm = make(map[int64]*operate.Card, len(ids))
	for _, id := range ids {
		if o, ok := s.followCache[id]; ok {
			card := &operate.Card{}
			card.FromFollow(o)
			cardm[id] = card
			switch card.Goto {
			case model.GotoAv:
				if card.ID != 0 {
					aids = append(aids, card.ID)
				}
				if card.Tid != 0 {
					tids = append(tids, card.Tid)
				}
			default:
			}
		}
	}
	return
}

//nolint:unparam
func (s *Service) specialCard(_ context.Context, ids ...int64) (cardm map[int64]*operate.Card, aids, roomids, metaids []int64, epids, seasonids []int32) {
	if len(ids) == 0 {
		return
	}
	cardm = make(map[int64]*operate.Card, len(ids))
	for _, id := range ids {
		if o, ok := s.specialCache[id]; ok {
			card := &operate.Card{}
			card.FromSpecial(o)
			cardm[id] = card
			switch card.Goto {
			case model.GotoAv:
				aids = append(aids, card.ID)
			case model.GotoLive:
				roomids = append(roomids, card.ID)
			case model.GotoBangumi:
				epids = append(epids, int32(card.ID))
			case model.GotoArticle:
				metaids = append(metaids, card.ID)
			case model.GotoPGC:
				// it is pgc-season
				seasonids = append(seasonids, int32(card.ID))
			default:
			}
		}
	}
	return
}

func (s *Service) teenagersSpecialCard(_ context.Context) (cardm map[int64]*operate.Card) {
	cardm = map[int64]*operate.Card{
		s.c.Feed.Index.TeenagersSpecialCard.ID: {
			ID:       s.c.Feed.Index.TeenagersSpecialCard.ID,
			CardGoto: model.CardGotoSpecial,
			Param:    strconv.FormatInt(s.c.Feed.Index.TeenagersSpecialCard.ID, 10),
			Coverm:   map[model.ColumnStatus]string{model.ColumnSvrSingle: s.c.Feed.Index.TeenagersSpecialCard.Cover, model.ColumnSvrDouble: s.c.Feed.Index.TeenagersSpecialCard.Cover},
			Title:    s.c.Feed.Index.TeenagersSpecialCard.Title,
			Goto:     model.GotoWeb,
			URI:      s.c.Feed.Index.TeenagersSpecialCard.URL,
		},
	}
	return
}

func (s *Service) convergeCardAi(_ context.Context, r *ai.ConvergeInfo, id int64) (card *operate.Card, aids []int64) {
	var limit = 10
	if r == nil || len(r.Items) == 0 {
		return
	}
	card = new(operate.Card)
	card.FromConvergeAi(r, id)
	for _, item := range card.Items {
		switch item.Goto {
		case model.GotoAv:
			if item.ID != 0 {
				aids = append(aids, item.ID)
			}
		default:
		}
	}
	if len(aids) > limit {
		aids = aids[:limit]
	}
	return
}

func (s *Service) avConvergeCard(_ context.Context, r *ai.Item) (card *operate.Card, aids []int64) {
	if r == nil || r.ConvergeInfo == nil || len(r.ConvergeInfo.Items) == 0 {
		return
	}
	var aid int64
LOOP:
	for _, item := range r.ConvergeInfo.Items {
		switch model.Gt(item.Goto) {
		case model.GotoAv:
			if item.ID != 0 {
				aid = item.ID
				aids = append(aids, item.ID)
				break LOOP
			}
		default:
		}
	}
	card = new(operate.Card)
	card.FromAvConverge(r, aid)
	switch model.Gt(r.JumpGoto) {
	case model.GotoAv:
		if r.JumpID != 0 {
			aids = append(aids, r.JumpID)
		}
	case model.GotoMultilayerConverge, model.GotoAvConverge:
		card.Goto = model.GotoAvConverge
	default:
	}
	return
}

// 后台已屏蔽
//func (s *Service) specialCardB(_ context.Context, id int64) (card *operate.Card, aid, roomid, epid, metaid int64) {
//	if o, ok := s.specialCache[id]; ok {
//		card = &operate.Card{}
//		card.FromSpecialB(o)
//		switch card.Goto {
//		case model.GotoAv:
//			aid = card.ID
//		case model.GotoLive:
//			roomid = card.ID
//		case model.GotoBangumi:
//			epid = card.ID
//		case model.GotoArticle:
//			metaid = card.ID
//		default:
//		}
//	}
//	return
//}

func (s *Service) specialCardChannel(_ context.Context, id int64) (card *operate.Card, channelID int64) {
	if o, ok := s.specialCache[id]; ok {
		card = &operate.Card{}
		card.FromSpecialChannel(o)
		channelID = card.ID
	}
	return
}

func (s *Service) storyCard(_ context.Context, story *ai.StoryInfo) (aids, tids []int64) {
	if story == nil {
		return
	}
	for _, v := range story.Items {
		switch model.Gt(v.Goto) {
		case model.GotoVerticalAv:
			aids = append(aids, v.ID)
			tids = append(tids, v.Tid)
		default:
		}
	}
	return
}
