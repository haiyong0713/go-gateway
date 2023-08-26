package feed

import (
	"context"
	"strconv"
	"time"

	"go-common/library/log"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	errgroup "go-common/library/sync/errgroup.v2"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

// Actives2 return actives
func (s *Service) Actives2(c context.Context, id, mid int64, mobiApp, platform, device string, plat int8, build int, now time.Time, accessKey, actionKey, appkey, statistics, buvid, network string) (items []card.Handler, cover string, isBnj bool, bnjDays int, err error) {
	if id == s.c.Bnj.TabID {
		isBnj = true
		nt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		bt, _ := time.Parse("2006-01-02", s.c.Bnj.BeginTime)
		bnjDays = int(bt.Sub(nt).Hours() / 24)
		if bnjDays < 0 {
			bnjDays = 0
		}
	}
	var rs []*operate.Active
	if rs, cover, err = s.rsc.AppActive(c, id); err != nil {
		log.Error("%+v", err)
		// 未获取到直接返回空数组
		return []card.Handler{}, "", isBnj, bnjDays, nil
	}
	if items, err = s.dealTab2(c, rs, mid, mobiApp, platform, device, plat, build, now, accessKey, actionKey, appkey, statistics, buvid, network); err != nil {
		log.Error("s.dealTab(%v) error(%v)", rs, err)
		return
	}
	return
}

// nolint: gocognit
func (s *Service) dealTab2(c context.Context, rs []*operate.Active, mid int64, mobiApp, platform, device string, plat int8, build int, _ time.Time,
	accessKey, actionKey, appkey, statistics, buvid, network string) (is []card.Handler, err error) {
	if len(rs) == 0 {
		is = []card.Handler{}
		return
	}
	var (
		paids, aids, tids, roomIDs, metaIDs, picIDs []int64
		sids, epIDs                                 []int32
		amplayer, am                                map[int64]*arcgrpc.ArcPlayer
		rm                                          map[int64]*live.Room
		sm                                          map[int32]*episodegrpc.EpisodeCardsProto
		metam                                       map[int64]*article.Meta
		tagm                                        map[int64]*taggrpc.Tag
		picm                                        map[int64]*bplus.Picture
		eppm                                        map[int32]*pgcinline.EpisodeCard
		// haslike                                           map[int64]int8
	)
	convergem := map[int64]*operate.Card{}
	specialm := map[int64]*operate.Card{}
	downloadm := map[int64]*operate.Download{}
	for _, r := range rs {
		switch r.Type {
		case model.GotoPlayer:
			if r.Pid != 0 {
				paids = append(paids, r.Pid)
			}
		case model.GotoPlayerLive:
			if r.Pid != 0 {
				roomIDs = append(roomIDs, r.Pid)
			}
		case model.GotoPlayerOGV:
			//pgc播放大卡 从540开始展示
			if r.Pid != 0 && ((model.IsAndroid(plat) && build > 5395000) || (model.IsIOSNormal(plat) && build > 8430)) {
				epIDs = append(epIDs, int32(r.Pid))
			}
		case model.GotoTabTagRcmd:
			if r.Pid != 0 {
				var taids []int64
				if taids, err = s.rcmd.TagTop(c, mid, r.Pid, r.Limit); err != nil {
					log.Error("%+v", err)
					err = nil
					continue
				}
				tids = append(tids, r.Pid)
				r.Items = make([]*operate.Active, 0, len(taids))
				for _, aid := range taids {
					item := &operate.Active{Pid: aid, Goto: model.GotoAv, Param: strconv.FormatInt(aid, 10)}
					r.Items = append(r.Items, item)
					aids = append(aids, aid)
				}
			}
		case model.GotoConverge:
			cardm, aid, roomID, metaID := s.convergeCard(c, 3, r.Pid)
			for id, card := range cardm {
				if !cdm.ShowLive(mobiApp, device, build) && card.Goto == model.GotoLive {
					continue
				}
				convergem[id] = card
			}
			aids = append(aids, aid...)
			roomIDs = append(roomIDs, roomID...)
			metaIDs = append(metaIDs, metaID...)
		case model.GotoTabContentRcmd:
			for _, item := range r.Items {
				if item.Pid == 0 {
					continue
				}
				switch item.Goto {
				case cdm.GotoAv:
					aids = append(aids, item.Pid)
				case cdm.GotoLive:
					roomIDs = append(roomIDs, item.Pid)
				case cdm.GotoBangumi:
					sids = append(sids, int32(item.Pid))
				case cdm.GotoGame:
					if card, ok := s.downloadCache[item.Pid]; ok {
						downloadm[item.Pid] = card
					}
				case cdm.GotoArticle:
					metaIDs = append(metaIDs, item.Pid)
				case cdm.GotoSpecial:
					cardm, _, _, _, _, _ := s.specialCard(c, item.Pid)
					for id, card := range cardm {
						if !cdm.ShowLive(mobiApp, device, build) && card.Goto == model.GotoLive {
							continue
						}
						specialm[id] = card
					}
				case cdm.GotoPicture:
					// 版本过滤5.37为新卡片
					if (plat == model.PlatIPhone && build > 8300) || (plat == model.PlatAndroid && build > 5365000) {
						picIDs = append(picIDs, item.Pid)
					}
				default:
				}
			}
		case model.GotoSpecial:
			cardm, _, _, _, _, _ := s.specialCard(c, r.Pid)
			for id, card := range cardm {
				if !cdm.ShowLive(mobiApp, device, build) && card.Goto == model.GotoLive {
					continue
				}
				specialm[id] = card
			}
		default:
		}
	}
	g := errgroup.WithContext(c)
	if len(tids) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if tagm, err = s.tg.TagsInfoByIDs(c, 0, tids); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(aids) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			var (
				arcs map[int64]*api.Arc
			)
			if arcs, err = s.arc.Archives(ctx, aids, mid, mobiApp, device); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			am = make(map[int64]*arcgrpc.ArcPlayer, len(arcs))
			for aid, a := range arcs {
				am[aid] = &arcgrpc.ArcPlayer{Arc: a}
			}
			return
		})
	}
	if len(paids) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if amplayer, err = s.ArcsPlayer(ctx, paids); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(roomIDs) != 0 && cdm.ShowLive(mobiApp, device, build) {
		g.Go(func(ctx context.Context) (err error) {
			if rm, err = s.lv.AppMRoom(ctx, roomIDs, mid, platform, "", accessKey, actionKey, appkey, device, mobiApp, statistics, buvid, network, build, 0, 0, 0, 0, 0); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(sids) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if sm, err = s.bgm.CardsByAids(ctx, sids); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(metaIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if metam, err = s.art.Articles(ctx, metaIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(picIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if picm, err = s.bplus.DynamicDetail(ctx, platform, mobiApp, device, build, picIDs...); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(epIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if eppm, err = s.bgm.InlineCards(ctx, epIDs, mobiApp, platform, device, build, mid, false, false, false, buvid, nil); err != nil {
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
	is = make([]card.Handler, 0, len(rs))
	for _, r := range rs {
		var (
			main     interface{}
			cardType cdm.CardType
		)
		cardGoto := cdm.CardGt(r.Type)
		op := &operate.Card{}
		op.From(cardGoto, r.Pid, 0, plat, build, mobiApp)
		op.SwitchLargeCoverShow = cdm.SwitchLargeCoverHideAll
		// 版本过滤
		hasThreePoint := (plat == model.PlatIPhone && build >= 8240) || (plat == model.PlatAndroid && build > 5341000)
		if hasThreePoint {
			op.FromSwitch(cdm.SwitchFeedIndexTabThreePoint)
		}
		switch r.Type {
		case model.GotoPlayer:
			main = amplayer
		case model.GotoPlayerLive:
			main = rm
		case model.GotoPlayerOGV:
			main = eppm
		case model.GotoSpecial:
			op = specialm[r.Pid]
		case model.GotoConverge:
			main = map[cdm.Gt]interface{}{cdm.GotoAv: am, cdm.GotoLive: rm, cdm.GotoArticle: metam}
			op = convergem[r.Pid]
		case model.GotoBanner:
			var isNewBanner bool
			// feature TagNewBanner
			if (mobiApp == "iphone" && build > 8940) || (mobiApp == "ipad" && build > 12350) || (mobiApp == "android" && build > 5499999) || mobiApp == "iphone_i" || (mobiApp == "android_i" && build > 2042030) {
				isNewBanner = true
			}
			var canOff bool
			for _, v := range r.Items {
				if !cdm.ShowLive(mobiApp, device, build) && v.Goto == model.GotoLive {
					canOff = true
					break
				}
			}
			if canOff {
				continue
			}
			cardType = op.FromActiveBanner(r.Items, "", isNewBanner)
		case model.GotoTabNews:
			op.FromActive(r)
		case model.GotoTabContentRcmd:
			main = map[cdm.Gt]interface{}{cdm.GotoAv: am, cdm.GotoGame: downloadm, cdm.GotoBangumi: sm, cdm.GotoLive: rm, cdm.GotoArticle: metam, cdm.GotoSpecial: specialm, cdm.GotoPicture: picm}
			op.FromActive(r)
		case model.GotoTabEntrance:
			op.FromActive(r)
		case model.GotoTabTagRcmd:
			main = map[cdm.Gt]interface{}{cdm.GotoAv: am}
			op.FromActive(r)
			op.Items = make([]*operate.Card, 0, len(r.Items))
			for _, item := range r.Items {
				if item != nil {
					op.Items = append(op.Items, &operate.Card{ID: item.Pid, Goto: item.Goto})
				}
			}
		case model.GotoVip:
			op.FromActive(r)
		}
		if op != nil {
			var isOff bool
			if !cdm.ShowLive(mobiApp, device, build) && (op.CardGoto == model.GotoLive || op.CardGoto == model.GotoLiveUpRcmd || op.CardGoto == model.GotoPlayerLive) {
				continue
			}
			for _, v := range op.Items {
				if !cdm.ShowLive(mobiApp, device, build) && v.Goto == model.GotoLive {
					isOff = true
					break
				}
			}
			if isOff {
				continue
			}
		}
		h := card.Handle(plat, cardGoto, cardType, cdm.ColumnSvrDouble, nil, tagm, nil, nil, nil, nil, nil)
		if h == nil {
			continue
		}
		op.FromDev(mobiApp, plat, build)
		if err := h.From(main, op); err != nil {
			log.Error("Failed to From: %+v", err)
		}
		if !h.Get().Right {
			continue
		}
		if hasThreePoint {
			h.Get().TabThreePointWatchLater()
		}
		is = append(is, h)
	}
	return
}
