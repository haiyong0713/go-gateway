package feed

import (
	"context"
	"sort"
	"strconv"
	"time"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-intl/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/model/feed"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/text/translate/chinese.v2"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) Menus(c context.Context, plat int8, build int, now time.Time) (menus []*operate.Menu) {
	memuCache := s.menuCache
	menus = make([]*operate.Menu, 0, len(memuCache))
LOOP:
	for _, m := range memuCache {
		if vs, ok := m.Versions[plat]; ok {
			for _, v := range vs {
				if model.InvalidBuild(build, v.Build, v.Condition) {
					continue LOOP
				}
			}
			if m.Status == 1 && (m.STime == 0 || now.After(m.STime.Time())) && (m.ETime == 0 || now.Before(m.ETime.Time())) {
				menus = append(menus, m)
			}
		}
	}
	return
}

func (s *Service) loadTabCache() {
	c := context.TODO()
	menus, err := s.rsc.Menus(c)
	if err != nil {
		log.Error("s.rsc.Menus err is %+v", err)
	} else {
		s.menuCache = menus
	}
	acs, err := s.rsc.Actives(c)
	if err != nil {
		log.Error("%+v", err)
	} else {
		s.tabCache, s.coverCache = mergeTab(acs)
	}
}

func mergeTab(acs []*operate.Active) (tabm map[int64][]*operate.Active, coverm map[int64]string) {
	coverm = make(map[int64]string, len(acs))
	parentm := make(map[int64]struct{}, len(acs))
	for _, ac := range acs {
		if ac.Type == feed.GotoTabBackground {
			parentm[ac.ID] = struct{}{}
			coverm[ac.ID] = ac.Cover
		}
	}
	sort.Sort(operate.Actives(acs))
	tabm = make(map[int64][]*operate.Active, len(acs))
	for parentID := range parentm {
		for _, ac := range acs {
			if ac.ParentID == parentID {
				tabm[ac.ParentID] = append(tabm[ac.ParentID], ac)
			}
		}
	}
	return
}

// Actives2 return actives
func (s *Service) Actives(c context.Context, id, mid int64, mobiApp, device string, plat int8, build int, now time.Time) (items []card.Handler, cover string, isBnj bool, bnjDays int, err error) {
	rs := s.tabCache[id]
	if items, err = s.dealTab(c, rs, mid, mobiApp, device, plat, build, now); err != nil {
		log.Error("s.dealTab(%v) error(%v)", rs, err)
		return
	}
	cover = s.coverCache[id]
	return
}

// nolint:gocognit, staticcheck
func (s *Service) dealTab(c context.Context, rs []*operate.Active, mid int64, mobiApp, device string, plat int8, build int, _ time.Time) (is []card.Handler, err error) {
	if len(rs) == 0 {
		is = []card.Handler{}
		return
	}
	var (
		aids, tids, roomIDs, metaIDs, picIDs []int64
		paids                                []*arcgrpc.PlayAv
		sids, epIDs                          []int32
		am, amplayer                         map[int64]*arcgrpc.ArcPlayer
		rm                                   map[int64]*live.Room
		sm                                   map[int32]*episodegrpc.EpisodeCardsProto
		metam                                map[int64]*article.Meta
		tagm                                 map[int64]*taggrpc.Tag
		picm                                 map[int64]*bplus.Picture
		eppm                                 map[int32]*pgcinline.EpisodeCard
		// haslike                                           map[int64]int8
	)
	convergem := map[int64]*operate.Card{}
	specialm := map[int64]*operate.Card{}
	downloadm := map[int64]*operate.Download{}
	for _, r := range rs {
		switch r.Type {
		case model.GotoPlayer:
			if r.Pid != 0 {
				paids = append(paids, &arcgrpc.PlayAv{Aid: r.Pid})
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
					log.Error("s.rcmd.TagTop err is %+v", err)
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
				case cdm.GotoArticle:
					metaIDs = append(metaIDs, item.Pid)
				case cdm.GotoSpecial:
					cardm, _, _, _, _ := s.specialCard(c, item.Pid)
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
			cardm, _, _, _, _ := s.specialCard(c, r.Pid)
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
			if tagm, err = s.tg.Tags(c, 0, tids); err != nil {
				log.Error("s.tg.InfoByIDs err is %+v", err)
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
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("s.arc.Archives err is %+v", err)
				err = nil
			}
			am = make(map[int64]*arcgrpc.ArcPlayer, len(arcs))
			for aid, a := range arcs {
				a.TypeName = chinese.Convert(c, a.TypeName)
				a.Title = chinese.Convert(c, a.Title)
				am[aid] = &arcgrpc.ArcPlayer{Arc: a}
			}
			return
		})
	}
	if len(paids) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if amplayer, err = s.ArcsPlayer(ctx, paids); err != nil {
				log.Error("s.ArcsWithPlayurl err is %+v", err)
				err = nil
			}
			for _, v := range amplayer {
				if v == nil || v.Arc == nil {
					continue
				}
				v.Arc.Title = chinese.Convert(c, v.Arc.Title)
			}
			return
		})
	}
	if len(sids) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if sm, err = s.bgm.CardsByAids(ctx, sids); err != nil {
				log.Error("s.bgm.CardsByAids err is %+v", err)
				err = nil
			}
			for _, v := range sm {
				v.Title = chinese.Convert(c, v.Title)
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
		if err = h.From(main, op); err != nil {
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
