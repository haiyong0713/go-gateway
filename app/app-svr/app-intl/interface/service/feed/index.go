package feed

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	"go-common/library/text/translate/chinese.v2"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-intl/interface/model"
	"go-gateway/app/app-svr/app-intl/interface/model/feed"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

// Index is.
func (s *Service) Index(c context.Context, buvid string, mid int64, plat int8, param *feed.IndexParam, now time.Time, style int) (is []card.Handler, userFeature json.RawMessage, isRcmd, newUser bool, code int, autoPlay, clean int8, autoPlayInfoc string, err error) {
	var (
		rs        []*ai.Item
		adm       map[int]*cm.AdInfo
		adAidm    map[int64]struct{}
		banners   []*banner.Banner
		version   string
		blackAidm map[int64]struct{}
		adInfom   map[int]*cm.AdInfo
		follow    *operate.Card
		ip        = metadata.String(c, metadata.RemoteIP)
		info      *locgrpc.InfoReply
		isTW      = model.TWLocale(param.Locale)
	)
	// 国际版不做abtest
	clean = 0
	autoPlay = 2
	group := s.group(mid, buvid)
	if info, err = s.loc.Info(c, ip); err != nil {
		log.Warn("s.loc.Info(%v) error(%v)", ip, err)
		err = nil
	}
	if !s.c.Feed.Index.Abnormal {
		g, ctx := errgroup.WithContext(c)
		g.Go(func() error {
			rs, userFeature, isRcmd, newUser, code = s.indexRcmd(ctx, plat, param.Build, buvid, mid, group, param.LoginEvent, info, param.Interest, param.Network, style, param.Column, param.Flush, autoPlayInfoc, now)
			if isTW {
				for _, r := range rs {
					if r.RcmdReason != nil {
						r.RcmdReason.Content = chinese.Convert(ctx, r.RcmdReason.Content)
					}
				}
			}
			return nil
		})
		g.Go(func() (err error) {
			if banners, version, err = s.indexBanner2(ctx, plat, buvid, mid, param); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		g.Go(func() (err error) {
			if blackAidm, err = s.BlackList(ctx, mid); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		if err = g.Wait(); err != nil {
			return
		}
		rs, adInfom = s.mergeItem(c, mid, rs, adm, adAidm, banners, version, blackAidm, plat, follow)
	} else {
		count := s.indexCount(plat)
		rs = s.recommendCache(count)
		log.Warn("feed index show disaster recovery data len(%d)", len(rs))
	}
	is, isRcmd, err = s.dealItem(c, param.Column, mid, plat, param.Build, rs, isRcmd, param.MobiApp, isTW, param)
	s.dealAdLoc(is, param, adInfom, now)
	return
}

// indexRcmd is.
func (s *Service) indexRcmd(c context.Context, plat int8, build int, buvid string, mid int64, group int, loginEvent int, info *locgrpc.InfoReply, interest, network string, style int, column cdm.ColumnStatus, flush int, autoPlay string, now time.Time) (is []*ai.Item, userFeature json.RawMessage, isRcmd, newUser bool, code int) {
	count := s.indexCount(plat)
	if buvid != "" || mid != 0 {
		var (
			err    error
			zoneID int64
		)
		if info != nil {
			zoneID = info.ZoneId
		}
		if is, userFeature, code, newUser, err = s.rcmd.Recommend(c, plat, buvid, mid, build, loginEvent, zoneID, group, interest, network, style, column, flush, autoPlay, now); err != nil {
			log.Error("%+v", err)
		} else if len(is) != 0 {
			isRcmd = true
		}
		var fromCache bool
		if len(is) == 0 && mid != 0 && !ecode.ServiceUnavailable.Equal(err) {
			if is, err = s.indexCache(c, mid, count); err != nil {
				log.Error("%+v", err)
			}
			if len(is) != 0 {
				s.pHit.Incr("index_cache")
			} else {
				s.pMiss.Incr("index_cache")
			}
			fromCache = true
		}
		if len(is) == 0 || (fromCache && len(is) < count) {
			is = s.recommendCache(count)
		}
	} else {
		is = s.recommendCache(count)
	}
	return
}

// mergeItem is.
func (s *Service) mergeItem(_ context.Context, mid int64, rs []*ai.Item, adm map[int]*cm.AdInfo, adAidm map[int64]struct{}, banners []*banner.Banner, version string, blackAids map[int64]struct{}, _ int8, _ *operate.Card) (is []*ai.Item, adInfom map[int]*cm.AdInfo) {
	if len(rs) == 0 {
		return
	}
	if len(banners) != 0 {
		rs = append([]*ai.Item{{Goto: model.GotoBanner, Banners: banners, Version: version}}, rs...)
	}
	is = make([]*ai.Item, 0, len(rs)+len(adm))
	adInfom = make(map[int]*cm.AdInfo, len(adm))
	for _, r := range rs {
		if r.Goto == model.GotoAv {
			if _, ok := blackAids[r.ID]; ok {
				continue
			} else if _, ok := s.blackCache[r.ID]; ok {
				continue
			}
			if _, ok := adAidm[r.ID]; ok {
				continue
			}
		} else if r.Goto == model.GotoBanner && len(is) != 0 {
			// banner 必须在第一位
			continue
		} else if r.Goto == model.GotoLogin && mid != 0 {
			continue
		}
		is = append(is, r)
	}
	return
}

// dealAdLoc is.
func (*Service) dealAdLoc(is []card.Handler, param *feed.IndexParam, adInfom map[int]*cm.AdInfo, now time.Time) {
	il := len(is)
	if il == 0 {
		return
	}
	if param.Idx < 1 {
		param.Idx = now.Unix()
	}
	for i, h := range is {
		if param.Pull {
			h.Get().Idx = param.Idx + int64(il-i)
		} else {
			h.Get().Idx = param.Idx - int64(i+1)
		}
		if ad, ok := adInfom[i]; ok {
			h.Get().AdInfo = ad
		} else if h.Get().AdInfo != nil {
			h.Get().AdInfo.CardIndex = int32(i)
		}
	}
}

// nolint:gocognit
func (s *Service) dealItem(c context.Context, column cdm.ColumnStatus, mid int64, plat int8, build int, rs []*ai.Item, isRcmd bool, mobiApp string, isTW bool, param *feed.IndexParam) (is []card.Handler, isAI bool, err error) {
	if len(rs) == 0 {
		is = []card.Handler{}
		return
	}
	var (
		tids                             []int64
		aids                             []*arcgrpc.PlayAv
		seasonIDs, sids                  []int32
		upIDs, avUpIDs, rmUpIDs, mtUpIDs []int64
		amplayer                         map[int64]*arcgrpc.ArcPlayer
		tagm                             map[int64]*taggrpc.Tag
		cardm                            map[int64]*account.Card
		statm                            map[int64]*relationgrpc.StatReply
		isAtten                          map[int64]int8
		arcOK                            bool
		banners                          []*banner.Banner
		version                          string
		seasonm, sm                      map[int32]*episodegrpc.EpisodeCardsProto
		flowInfosV2Reply                 *cfcgrpc.FlowCtlInfosV2Reply
	)
	isAI = isRcmd
	for _, r := range rs {
		if r == nil {
			continue
		}
		if isTW && r.RcmdReason != nil {
			r.RcmdReason.Content = chinese.Convert(c, r.RcmdReason.Content)
		}
		switch r.Goto {
		case model.GotoAv, model.GotoPlayer:
			if r.ID != 0 {
				aids = append(aids, &arcgrpc.PlayAv{Aid: r.ID})
			}
			if r.Tid != 0 {
				tids = append(tids, r.Tid)
			}
		case model.GotoBanner:
			if len(r.Banners) != 0 {
				banners = r.Banners
				version = r.Version
			}
		case model.GotoBangumi:
			if r.ID != 0 {
				sids = append(sids, int32(r.ID))
				aids = append(aids, &arcgrpc.PlayAv{Aid: r.ID})
			}
			if r.Tid != 0 {
				tids = append(tids, r.Tid)
			}
		case model.GotoPGC:
			if r.ID != 0 {
				seasonIDs = append(seasonIDs, int32(r.ID))
			}
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(aids) != 0 {
		g.Go(func() (err error) {
			if amplayer, err = s.arc.ArcsPlayer(ctx, aids); err != nil {
				return
			}
			for _, a := range amplayer {
				if a == nil || a.Arc == nil {
					continue
				}
				avUpIDs = append(avUpIDs, a.Arc.Author.Mid)
				if isTW {
					out := chinese.Converts(ctx, a.Arc.Title, a.Arc.Desc, a.Arc.TypeName)
					a.Arc.Title = out[a.Arc.Title]
					a.Arc.Desc = out[a.Arc.Desc]
					a.Arc.TypeName = out[a.Arc.TypeName]
				}
			}
			arcOK = true
			return
		})
		g.Go(func() error {
			//此aids是结构体数组
			var oids []int64
			for _, arc := range aids {
				oids = append(oids, arc.Aid)
			}
			flowInfosV2Reply, err = s.cfc.ContentFlowControlInfosV2(ctx, oids)
			if err != nil {
				log.Error("s.cfc.ContentFlowControlInfosV2 err=%+v", err)
				return nil
			}
			return nil
		})
	}
	if len(tids) != 0 {
		g.Go(func() (err error) {
			if tagm, err = s.tg.Tags(ctx, mid, tids); err != nil {
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
	if len(seasonIDs) != 0 {
		g.Go(func() (err error) {
			if seasonm, err = s.bgm.CardsByEpisodeIds(ctx, seasonIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		if isRcmd {
			count := s.indexCount(plat)
			rs = s.recommendCache(count)
		}
	} else {
		upIDs = append(upIDs, avUpIDs...)
		upIDs = append(upIDs, rmUpIDs...)
		upIDs = append(upIDs, mtUpIDs...)
		g, ctx = errgroup.WithContext(c)
		if len(upIDs) != 0 {
			g.Go(func() (err error) {
				if cardm, err = s.acc.Cards3(ctx, upIDs); err != nil {
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
					isAtten = s.acc.IsAttention(ctx, upIDs, mid)
					return nil
				})
			}
		}
		if err = g.Wait(); err != nil {
			log.Error("dealItem errGroup err(%+v)", err)
		}
	}
	isAI = isAI && arcOK
	var cardTotal int
	is = make([]card.Handler, 0, len(rs))
	for _, r := range rs {
		if r == nil {
			continue
		}
		var (
			main     interface{}
			cardType cdm.CardType
		)
		op := &operate.Card{}
		op.From(cdm.CardGt(r.Goto), r.ID, r.Tid, plat, build, mobiApp)
		switch r.Goto {
		case model.GotoBanner:
			if mobiApp == "iphone_i" || mobiApp == "android_i" && build > 2042030 {
				if model.IsIPad(plat) {
					cardType = cdm.BannerV6
				} else {
					switch cdm.Columnm[param.Column] {
					case cdm.ColumnSvrSingle:
						cardType = cdm.BannerV4
					case cdm.ColumnSvrDouble:
						cardType = cdm.BannerV5
					default:
					}
				}
			}
		}
		h := card.Handle(plat, cdm.CardGt(r.Goto), cardType, column, r, tagm, isAtten, nil, statm, cardm, nil)
		if h == nil {
			continue
		}
		switch r.Goto {
		case model.GotoAv, model.GotoPlayer:
			if !arcOK {
				if r.Archive != nil {
					if isTW {
						out := chinese.Converts(c, r.Archive.Title, r.Archive.Desc, r.Archive.TypeName, r.Archive.Author.Name)
						r.Archive.Title = out[r.Archive.Title]
						r.Archive.Desc = out[r.Archive.Desc]
						r.Archive.TypeName = out[r.Archive.TypeName]
					}
					amplayer = map[int64]*arcgrpc.ArcPlayer{r.Archive.Aid: {Arc: r.Archive}}
				}
				if r.Tag != nil {
					tagm = map[int64]*taggrpc.Tag{r.Tag.Id: r.Tag}
					op.Tid = r.Tag.Id
				}
			}
			if a, ok := amplayer[r.ID]; ok && a != nil && a.Arc != nil && (!getAttrBitValueFromInfosV2(flowInfosV2Reply, a.Arc.GetAid(), model.OverseaBlockKey) || !model.IsOverseas(plat)) {
				main = amplayer
				op.TrackID = r.TrackID
			}
		case model.GotoBanner:
			op.FromBanner(banners, version)
		case model.GotoBangumi:
			main = sm
			if r.Tag != nil {
				tagm = map[int64]*taggrpc.Tag{r.Tag.Id: r.Tag}
				op.Tid = r.Tag.Id
			}
			if a, ok := amplayer[r.ID]; ok && a != nil && a.Arc != nil {
				op.Desc = a.Arc.TypeName
			}
		case model.GotoPGC:
			main = seasonm
		default:
			log.Warn("unexpected goto(%s) %+v", r.Goto, r)
			continue
		}
		op.FromDev(mobiApp, plat, build)
		if err = h.From(main, op); err != nil {
			log.Error("Failed to From: %+v", err)
		}
		// 卡片不正常要continue
		if !h.Get().Right {
			continue
		}
		is, cardTotal = s.appendItem(plat, is, h, column, cardTotal, build, mobiApp)
	}
	// 双列末尾卡片去空窗
	if !model.IsIPad(plat) {
		if cdm.Columnm[column] == cdm.ColumnSvrDouble {
			is = is[:len(is)-cardTotal%2]
		}
	} else {
		// 复杂的ipad去空窗逻辑
		// nolint: gomnd
		if cardTotal%4 == 3 {
			if is[len(is)-2].Get().CardLen == 2 {
				is = is[:len(is)-2]
			} else {
				is = is[:len(is)-3]
			}
		} else if cardTotal%4 == 2 {
			if is[len(is)-1].Get().CardLen == 2 {
				is = is[:len(is)-1]
			} else {
				is = is[:len(is)-2]
			}
		} else if cardTotal%4 == 1 {
			is = is[:len(is)-1]
		}
	}
	if len(is) == 0 {
		is = []card.Handler{}
		return
	}
	return
}

// appendItem is.
// nolint:gomnd
func (s *Service) appendItem(plat int8, rs []card.Handler, h card.Handler, column cdm.ColumnStatus, cardTotal, build int, mobiApp string) (is []card.Handler, total int) {
	const _oldDislikeExp = 0 // 老样式
	h.Get().ThreePointFrom(mobiApp, build, _oldDislikeExp, nil, 0, 0)
	// 国际版暂不支持稿件反馈
	if h.Get().ThreePoint != nil {
		h.Get().ThreePoint.Feedbacks = nil
	}
	if !model.IsIPad(plat) {
		// 双列大小卡换位去空窗
		if cdm.Columnm[column] == cdm.ColumnSvrDouble {
			// 通栏卡
			if h.Get().CardLen == 0 {
				if cardTotal%2 == 1 {
					is = card.SwapTwoItem(rs, h)
				} else {
					is = append(rs, h)
				}
			} else {
				is = append(rs, h)
			}
		} else {
			is = append(rs, h)
		}
	} else {
		// ipad卡片不展示标签
		h.Get().DescButton = nil
		// ipad大小卡换位去空窗
		if h.Get().CardLen == 0 {
			// 通栏卡
			if cardTotal%4 == 3 {
				is = card.SwapFourItem(rs, h)
			} else if cardTotal%4 == 2 {
				is = card.SwapThreeItem(rs, h)
			} else if cardTotal%4 == 1 {
				is = card.SwapTwoItem(rs, h)
			} else {
				is = append(rs, h)
			}
		} else if h.Get().CardLen == 2 {
			// 半栏卡
			if cardTotal%4 == 3 {
				is = card.SwapTwoItem(rs, h)
			} else if cardTotal%4 == 2 {
				is = append(rs, h)
			} else if cardTotal%4 == 1 {
				is = card.SwapTwoItem(rs, h)
			} else {
				is = append(rs, h)
			}
		} else {
			is = append(rs, h)
		}
	}
	total = cardTotal + h.Get().CardLen
	return
}

// indexCount is.
func (s *Service) indexCount(plat int8) (count int) {
	if plat == model.PlatIPad {
		count = s.c.Feed.Index.IPadCount
	} else {
		count = s.c.Feed.Index.Count
	}
	return
}

func (s *Service) indexBanner2(c context.Context, plat int8, buvid string, mid int64, param *feed.IndexParam) (banners []*banner.Banner, version string, err error) {
	hash := param.BannerHash
	if param.LoginEvent != 0 {
		hash = ""
	}
	banners, version, err = s.banners(c, plat, param.Build, mid, buvid, param.Network, param.MobiApp, param.Device, param.OpenEvent, param.AdExtra, hash)
	return
}

func getAttrBitValueFromInfosV2(reply *cfcgrpc.FlowCtlInfosV2Reply, aid int64, arcsAttrKey string) bool {
	//获取不到禁止项，默认为空
	if reply == nil {
		return false
	}
	val, ok := reply.ItemsMap[aid]
	if !ok {
		return false
	}
	for _, item := range val.Items {
		//处理reply
		if item.Key == arcsAttrKey {
			return item.Value == 1
		}
	}
	return false
}
