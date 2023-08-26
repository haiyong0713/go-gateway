package feed

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	article "git.bilibili.co/bapis/bapis-go/article/model"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

func (s *Service) Menus(c context.Context, plat int8, build int, now time.Time) []*operate.Menu {
	menus, err := s.rsc.Menus(c, plat, build)
	if err != nil {
		log.Error("%+v", err)
		return []*operate.Menu{}
	}
	return menus
}

// Actives return actives
func (s *Service) Actives(c context.Context, id, mid int64, platform string, now time.Time, mobiApp, buvid, device, accessKey, actionKey, appkey, statistics, network string, build int) (items []*feed.Item, cover string, isBnj bool, bnjDays int, err error) {
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
		return []*feed.Item{}, "", isBnj, bnjDays, nil
	}
	if items, err = s.dealTab(c, rs, mid, platform, now, mobiApp, buvid, device, accessKey, actionKey, appkey, statistics, network, build); err != nil {
		log.Error("%+v", err)
		return
	}
	return
}

// nolint: gocognit
func (s *Service) dealTab(c context.Context, rs []*operate.Active, mid int64, platform string, _ time.Time, mobiApp, buvid, device, accessKey, actionKey, appkey, statistics, network string, build int) (is []*feed.Item, err error) {
	if len(rs) == 0 {
		is = _emptyItem
		return
	}
	var (
		aids, tids, roomIDs, metaIDs []int64
		sids                         []int32
		am                           map[int64]*arcgrpc.Arc
		rm                           map[int64]*live.Room
		sm                           map[int32]*episodegrpc.EpisodeCardsProto
		metam                        map[int64]*article.Meta
		tagm                         map[int64]*taggrpc.Tag
	)
	convergem := map[int64]*operate.Converge{}
	downloadm := map[int64]*operate.Download{}
	for _, r := range rs {
		switch r.Type {
		case model.GotoPlayer:
			if r.Pid != 0 {
				aids = append(aids, r.Pid)
			}
		case model.GotoPlayerLive:
			if r.Pid != 0 {
				roomIDs = append(roomIDs, r.Pid)
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
					item := &operate.Active{Pid: aid, Goto: model.GotoAv}
					r.Items = append(r.Items, item)
					aids = append(aids, aid)
				}
			}
		case model.GotoConverge:
			if card, ok := s.convergeCache[r.Pid]; ok {
				for _, item := range card.Items {
					switch item.Goto {
					case model.GotoAv:
						if item.Pid != 0 {
							aids = append(aids, item.Pid)
						}
					case model.GotoLive:
						if item.Pid != 0 {
							roomIDs = append(roomIDs, item.Pid)
						}
					case model.GotoArticle:
						if item.Pid != 0 {
							metaIDs = append(metaIDs, item.Pid)
						}
					default:
					}
				}
				convergem[r.Pid] = card
			}
		case model.GotoTabEntrance, model.GotoTabContentRcmd:
			for _, item := range r.Items {
				switch item.Goto {
				case model.GotoAv:
					if item.Pid != 0 {
						aids = append(aids, item.Pid)
					}
				case model.GotoLive:
					if item.Pid != 0 {
						roomIDs = append(roomIDs, item.Pid)
					}
				case model.GotoBangumi:
					if item.Pid != 0 {
						sids = append(sids, int32(item.Pid))
					}
				case model.GotoGame:
					if card, ok := s.downloadCache[item.Pid]; ok {
						downloadm[item.Pid] = card
					}
				case model.GotoArticle:
					if item.Pid != 0 {
						metaIDs = append(metaIDs, item.Pid)
					}
				default:
				}
			}
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(tids) != 0 {
		g.Go(func() (err error) {
			if tagm, err = s.tg.TagsInfoByIDs(c, 0, tids); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(aids) != 0 {
		g.Go(func() (err error) {
			if am, err = s.arc.Archives(ctx, aids, 0, "", ""); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(roomIDs) != 0 {
		g.Go(func() (err error) {
			if rm, err = s.lv.AppMRoom(ctx, roomIDs, mid, platform, "", accessKey, actionKey, appkey, device, mobiApp, statistics, buvid, network, build, 0, 0, 0, 0, 0); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(sids) != 0 {
		g.Go(func() (err error) {
			if sm, err = s.bgm.CardsByAids(ctx, sids); err != nil {
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
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	is = make([]*feed.Item, 0, len(rs))
	for _, r := range rs {
		i := &feed.Item{}
		switch r.Type {
		case model.GotoPlayer:
			if a, ok := am[r.Pid]; ok {
				i.FromPlayer(a)
				is = append(is, i)
			}
		case model.GotoPlayerLive:
			if room, ok := rm[r.Pid]; ok {
				i.FromPlayerLive(room)
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoSpecial:
			if sc, ok := s.specialCache[r.Pid]; ok {
				i.FromSpecial(sc.ID, sc.Title, sc.Cover, sc.Desc, sc.ReValue, sc.ReType, sc.Badge, sc.Size)
			}
			if i.Goto != "" {
				is = append(is, i)
			}
		case model.GotoConverge:
			if cc, ok := convergem[r.Pid]; ok {
				i.FromConverge(cc, am, rm, metam)
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoTabTagRcmd:
			i.FromTabTags(r, am, tagm)
			if i.Goto != "" {
				is = append(is, i)
			}
		case model.GotoTabEntrance, model.GotoTabContentRcmd:
			i.FromTabCards(r, am, downloadm, sm, rm, metam, s.specialCache)
			if i.Goto != "" {
				is = append(is, i)
			}
		case model.GotoBanner:
			i.FromTabBanner(r)
			if i.Goto != "" {
				is = append(is, i)
			}
		case model.GotoTabNews:
			i.FromNews(r)
			if i.Goto != "" {
				is = append(is, i)
			}
		}
	}
	if len(is) == 0 {
		is = _emptyItem
	}
	return
}
