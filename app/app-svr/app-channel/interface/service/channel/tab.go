package channel

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	shopping "go-gateway/app/app-svr/app-card/interface/model/card/show"
	"go-gateway/app/app-svr/app-channel/interface/model"
	"go-gateway/app/app-svr/app-channel/interface/model/card"
	"go-gateway/app/app-svr/app-channel/interface/model/feed"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"

	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"

	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

func (s *Service) TabList(c context.Context, _, mid int64, _, mobiApp, platform string, build int, plat int8, _ time.Time) (res *feed.Tab, err error) {
	var (
		cards []*card.Card
		item  []cardm.Handler
	)
	item, err = s.dealTab2(c, plat, build, mobiApp, platform, mid, cards)
	res = &feed.Tab{
		Items: item,
	}
	return
}

// nolint:gocognit
func (s *Service) dealTab2(c context.Context, plat int8, build int, mobiApp, platform string, mid int64, cards []*card.Card) (is []cardm.Handler, err error) {
	if len(cards) == 0 {
		is = []cardm.Handler{}
		return
	}
	var (
		shopIDs, roomIDs, metaIDs []int64
		rmUpIDs, mtUpIDs, upIDs   []int64
		seasonIDs                 []int32
		rm                        map[int64]*live.Room
		metam                     map[int64]*article.Meta
		shopm                     map[int64]*shopping.Shopping
		seasonm                   map[int32]*seasongrpc.CardInfoProto
		ac                        map[int64]*accountgrpc.Card
		statm                     map[int64]*relationgrpc.StatReply
		isAtten                   map[int64]int8
	)
	for _, card := range cards {
		switch card.Type {
		case model.GotoPGC:
			if card.Value != 0 {
				seasonIDs = append(seasonIDs, int32(card.Value))
			}
		case model.GotoLive:
			if card.Value != 0 {
				roomIDs = append(roomIDs, card.Value)
			}
		case model.GotoArticle:
			if card.Value != 0 {
				metaIDs = append(metaIDs, card.Value)
			}
		case model.GotoShoppingS:
			if card.Value != 0 {
				shopIDs = append(shopIDs, card.Value)
			}
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(roomIDs) != 0 {
		g.Go(func() (err error) {
			if rm, err = s.lv.AppMRoom(ctx, roomIDs, platform); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			for _, r := range rm {
				rmUpIDs = append(rmUpIDs, r.UID)
			}
			return
		})
	}
	if len(seasonIDs) != 0 {
		g.Go(func() (err error) {
			if seasonm, err = s.bgm.CardsInfoReply(ctx, seasonIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(metaIDs) != 0 {
		g.Go(func() (err error) {
			if metam, err = s.art.Articles(ctx, metaIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			for _, meta := range metam {
				if meta.Author != nil {
					mtUpIDs = append(mtUpIDs, meta.Author.Mid)
				}
			}
			return
		})
	}
	if len(shopIDs) != 0 {
		g.Go(func() (err error) {
			if shopm, err = s.sp.Card(ctx, shopIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	upIDs = append(upIDs, rmUpIDs...)
	upIDs = append(upIDs, mtUpIDs...)
	g, ctx = errgroup.WithContext(c)
	if len(upIDs) != 0 {
		g.Go(func() (err error) {
			if ac, err = s.acc.Cards3GRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		g.Go(func() (err error) {
			if statm, err = s.rel.StatsGRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		if mid != 0 {
			g.Go(func() error {
				isAtten = s.acc.IsAttentionGRPC(ctx, upIDs, mid)
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, card := range cards {
		var (
			r    = card.CardToAiChange()
			main interface{}
		)
		h := cardm.Handle(plat, cdm.CardGt(r.Goto), "", cdm.ColumnSvrSingle, r, nil, isAtten, nil, statm, ac, nil)
		if h == nil {
			continue
		}
		op := &operate.Card{}
		op.From(cdm.CardGt(r.Goto), r.ID, 0, plat, build, mobiApp)
		switch r.Goto {
		case model.GotoLive:
			main = rm
		case model.GotoPGC:
			main = seasonm
		case model.GotoArticle:
			main = metam
		case model.GotoShoppingS:
			main = shopm
		}
		_ = h.From(main, op)
		if h.Get() == nil {
			continue
		}
		if h.Get().Right {
			is = append(is, h)
		}
	}
	if rl := len(is); rl == 0 {
		is = []cardm.Handler{}
		return
	}
	return
}
