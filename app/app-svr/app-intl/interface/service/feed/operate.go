package feed

import (
	"context"

	"go-gateway/app/app-svr/app-card/interface/model"
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

// nolint:unparam
func (s *Service) specialCard(_ context.Context, ids ...int64) (cardm map[int64]*operate.Card, aids, roomids, metaids []int64, epids []int32) {
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
			default:
			}
		}
	}
	return
}
