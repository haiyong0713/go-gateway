package show

import (
	"context"
	"strconv"

	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-show/interface/model"
)

func (s *Service) cardSetChange(_ context.Context, ids ...int64) (cardm map[int64]*operate.Card, aids []int64, upid int64) {
	if len(ids) == 0 {
		return
	}
	cardm = make(map[int64]*operate.Card, len(ids))
	for _, id := range ids {
		if cs, ok := s.cardSetCache[id]; ok {
			card := &operate.Card{}
			card.FromCardSet(cs)
			cardm[id] = card
			upid = card.ID
			for _, item := range card.Items {
				switch cs.Type {
				case "up_rcmd_new", "up_rcmd_new_single":
					aids = append(aids, item.ID)
				}
			}
		}
	}
	return
}

func (s *Service) eventTopicChange(_ context.Context, plat int8, ids ...int64) (cardm map[int64]*operate.Card) {
	if len(ids) == 0 {
		return
	}
	cardm = make(map[int64]*operate.Card, len(ids))
	for _, id := range ids {
		if st, ok := s.eventTopicCache[id]; ok {
			if plat == model.PlatH5 && !model.H5Link(st.ReValue) { // h5 仅展示 https:// 开头的链接的卡片
				continue
			}
			card := &operate.Card{}
			card.FromEventTopic(st)
			cardm[id] = card
		}
	}
	return
}

func (s *Service) handleLargeCard(_ context.Context, id int64) (cardm map[int64]*operate.Card, aid int64) {
	if id == 0 {
		return
	}
	cardm = make(map[int64]*operate.Card, 1)
	if cs, ok := s.largeCards[id]; ok {
		card := &operate.Card{}
		aid = cs.RID
		card.ID = cs.ID
		card.Desc = cs.Title
		if cs.Auto == 1 {
			card.CanPlay = true
		}
		card.Param = strconv.FormatInt(cs.ID, 10)
		card.SubParam = strconv.FormatInt(cs.RID, 10)
		card.CardGoto = model.GotoHotPlayerAV
		cardm[id] = card
	}
	return
}

func (s *Service) handleLiveCard(_ context.Context, id int64) (cardm map[int64]*operate.Card, roomID int64) {
	if id == 0 {
		return
	}
	cardm = make(map[int64]*operate.Card, 1)
	if cs, ok := s.liveCards[id]; ok {
		card := &operate.Card{}
		roomID = cs.RID
		card.ID = cs.ID
		card.Param = strconv.FormatInt(cs.ID, 10)
		card.SubParam = strconv.FormatInt(cs.RID, 10)
		card.CardGoto = model.GotoHotPlayerAV
		card.Goto = model.GotoLive
		card.RoomID = cs.RID
		card.Cover = cs.Cover
		cardm[id] = card
	}
	return
}

func (s *Service) handleArticleCard(_ context.Context, id int64) (card *operate.Card, articleID int64) {
	if id == 0 {
		return
	}
	if cs, ok := s.articleCards[id]; ok {
		articleID = cs.ArticleID
		card = &operate.Card{
			ID:    cs.ArticleID,
			Cover: cs.Cover,
			Param: strconv.FormatInt(cs.ArticleID, 10),
		}
	}
	return
}
