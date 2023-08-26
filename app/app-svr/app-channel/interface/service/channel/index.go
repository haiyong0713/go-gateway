package channel

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	shopping "go-gateway/app/app-svr/app-card/interface/model/card/show"
	"go-gateway/app/app-svr/app-channel/interface/model"
	"go-gateway/app/app-svr/app-channel/interface/model/activity"
	"go-gateway/app/app-svr/app-channel/interface/model/card"
	"go-gateway/app/app-svr/app-channel/interface/model/feed"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	errgroupv2 "go-common/library/sync/errgroup.v2"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

var (
	_emptyItem = []*feed.Item{}
)

// Index channel index
func (s *Service) Index(c context.Context, mid, channelID, idx int64, plat int8, mobiApp, device, buvid, platform, channelName string, build, loginEvent, displayID int, pull bool, now time.Time) (res *feed.Show, err error) {
	var (
		aids              []int64
		requestCnt        = 10
		isIpad            = plat == model.PlatIPad
		topic             *feed.Item
		item              []*feed.Item
		channelResource   *taggrpc.ChannelResourcesReply
		topChannel, isRec int
	)
	if isIpad {
		requestCnt = 20
	}
	if channelID > 0 {
		channelName = ""
	}
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if channelResource, err = s.tg.Resources(ctx, plat, channelID, mid, channelName, buvid, build, requestCnt, loginEvent, displayID, 0); err != nil {
			log.Error("index s.tg.Resources error(%v)", err)
			return
		}
		if channelResource != nil {
			aids = channelResource.Oids
			if channelResource.Failover {
				isRec = 0
			} else {
				isRec = 1
			}
			if channelResource.WhetherChannel {
				topChannel = 1
			} else {
				topChannel = 0
			}
		}
		return
	})
	g.Go(func() (err error) {
		var t *taggrpc.ChannelReply
		if t, err = s.tg.ChannelDetail(c, mid, channelID, channelName, s.isOverseas(plat)); err != nil {
			log.Error("s.tag.ChannelDetail(%d, %d, %s) error(%v)", mid, channelID, channelName, err)
			return
		}
		channelID = t.GetChannel().Id
		channelName = t.GetChannel().Name
		return
	})
	if err = g.Wait(); err != nil {
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.ChannelIndex, &feature.OriginResutl{
			BuildLimit: (mobiApp == "iphone" && build > 8050) || (mobiApp == "android" && build > 5305000),
		}) {
			log.Error("%+v", err)
			res = &feed.Show{
				Feed: _emptyItem,
			}
			return
		}
		err = nil
	}
	if loginEvent == 1 || loginEvent == 2 {
		if cards, ok := s.cardCache[channelID]; ok {
			topic, item, err = s.dealItem(c, mid, idx, plat, build, platform, pull, now, cards, aids)
		} else {
			item, err = s.feedItem(c, plat, aids)
		}
	} else {
		item, err = s.feedItem(c, plat, aids)
	}
	res = &feed.Show{
		Topic: topic,
		Feed:  item,
	}
	//infoc
	infoc := &feedInfoc{
		mobiApp:     mobiApp,
		device:      device,
		build:       strconv.Itoa(build),
		now:         now.Format("2006-01-02 15:04:05"),
		pull:        strconv.FormatBool(pull),
		loginEvent:  strconv.Itoa(loginEvent),
		channelID:   strconv.FormatInt(channelID, 10),
		channelName: channelName,
		mid:         strconv.FormatInt(mid, 10),
		buvid:       buvid,
		displayID:   strconv.Itoa(displayID),
		feed:        res,
		isRec:       strconv.Itoa(isRec),
		topChannel:  strconv.Itoa(topChannel),
		ServerCode:  "0",
	}
	s.infoc(infoc)
	return
}

// nolint:gocognit
func (s *Service) dealItem(c context.Context, mid, idx int64, plat int8, build int, platform string, pull bool, now time.Time, cards []*card.Card, listAID []int64) (top *feed.Item, is []*feed.Item, err error) {
	if len(cards) == 0 {
		is = _emptyItem
		return
	}
	var (
		aids, sids, roomIDs, metaIDs, shopIDs, audioIDs []int64
		upIDs, tids, avUpIDs, rmUpIDs, mtUpIDs          []int64
		seasonIDs                                       []int32
		am                                              map[int64]*arcgrpc.ArcPlayer
		tagm                                            map[int64]*taggrpc.Tag
		follows                                         map[int64]bool
		rm                                              map[int64]*live.Room
		sm                                              map[int64]*bangumi.Season
		actIDs, topIDs                                  []int64
		actm, topm                                      map[int64]*activity.Activity
		atm                                             map[int64]*article.Meta
		scm                                             map[int64]*shopping.Shopping
		aum                                             map[int64]*audio.Audio
		infocard                                        map[int64]*accountgrpc.Card
		upStatm                                         map[int64]*relationgrpc.StatReply
		cardAids                                        = map[int64]struct{}{}
		channelCards                                    []*card.Card
		seasonCards                                     map[int32]*episodegrpc.EpisodeCardsProto
		flowInfosV2Reply                                *cfcgrpc.FlowCtlInfosV2Reply
		// key
		_initCardPlatKey = "card_platkey_%d_%d"
		_fTypeOperation  = "operation"
		_fTypeRecommend  = "recommend"
	)
	convergem := map[int64]*operate.Converge{}
	downloadm := map[int64]*operate.Download{}
	liveUpm := map[int64][]*live.Card{}
	followm := map[int64]*operate.Follow{}
LOOP:
	for _, c := range cards {
		key := fmt.Sprintf(_initCardPlatKey, plat, c.ID)
		if cardPlat, ok := s.cardPlatCache[key]; ok {
			for _, l := range cardPlat {
				if model.InvalidBuild(build, l.Build, l.Condition) {
					continue LOOP
				}
			}
		} else {
			continue LOOP
		}
		channelCards = append(channelCards, c)
		switch c.Type {
		case model.GotoAv, model.GotoPlayer, model.GotoUpRcmdAv:
			if c.Value != 0 {
				aids = append(aids, c.Value)
				cardAids[c.Value] = struct{}{}
			}
		case model.GotoLive, model.GotoPlayerLive:
			if c.Value != 0 {
				roomIDs = append(roomIDs, c.Value)
			}
		case model.GotoBangumi:
			if c.Value != 0 {
				sids = append(sids, c.Value)
			}
		case model.GotoPGC:
			if c.Value != 0 {
				seasonIDs = append(seasonIDs, int32(c.Value))
			}
		case model.GotoActivity:
			if c.Value != 0 {
				actIDs = append(actIDs, c.Value)
			}
		case model.GotoTopic:
			if c.Value != 0 {
				topIDs = append(topIDs, c.Value)
			}
		case model.GotoConverge:
			if card, ok := s.convergeCardCache[c.Value]; ok {
				for _, item := range card.Items {
					// nolint:exhaustive
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
					}
				}
				convergem[c.Value] = card
			}
		case model.GotoGameDownload, model.GotoGameDownloadS:
			if card, ok := s.gameDownloadCache[c.Value]; ok {
				downloadm[c.Value] = card
			}
		case model.GotoArticle, model.GotoArticleS:
			if c.Value != 0 {
				metaIDs = append(metaIDs, c.Value)
			}
		case model.GotoShoppingS:
			if c.Value != 0 {
				shopIDs = append(shopIDs, c.Value)
			}
		case model.GotoAudio:
			if c.Value != 0 {
				audioIDs = append(audioIDs, c.Value)
			}
		case model.GotoLiveUpRcmd:
			if c.Value != 0 {
				if cs, ok := s.liveCardCache[c.Value]; ok {
					for _, c := range cs {
						if c == nil {
							continue
						}
						upIDs = append(upIDs, c.UID)
					}
				}
			}
		case model.GotoSubscribe:
			if c.Value != 0 {
				if card, ok := s.upCardCache[c.Value]; ok {
					for _, item := range card.Items {
						// nolint:exhaustive
						switch item.Goto {
						case cdm.GotoMid:
							if item.Pid != 0 {
								upIDs = append(upIDs, item.Pid)
							}
						case cdm.GotoTag:
							if item.Pid != 0 {
								tids = append(tids, item.Pid)
							}
						}
					}
					followm[c.Value] = card
				}
			}
		case model.GotoChannelRcmd:
			if c.Value != 0 {
				if card, ok := s.upCardCache[c.Value]; ok {
					if card.Pid != 0 {
						aids = append(aids, card.Pid)
					}
					if card.Tid != 0 {
						tids = append(tids, card.Tid)
					}
					followm[c.Value] = card
				}
			}
		}
	}
	if len(listAID) != 0 {
		aids = append(aids, listAID...)
	}
	g, ctx := errgroup.WithContext(c)
	if len(aids) != 0 {
		var aidsV2 []*arcgrpc.PlayAv
		for _, aid := range aids {
			if aid != 0 {
				aidsV2 = append(aidsV2, &arcgrpc.PlayAv{Aid: aid})
			}
		}
		g.Go(func() (err error) {
			if am, err = s.Archives(ctx, aidsV2, true); err != nil {
				return
			}
			for _, a := range am {
				if a != nil && a.Arc != nil {
					avUpIDs = append(avUpIDs, a.Arc.Author.Mid)
				}
			}
			return
		})
		g.Go(func() error {
			req := makeContentFlowControlInfosV2Params(s.c.CfcSvrConfig, aids)
			flowInfosV2Reply, err = s.arc.ContentFlowControlInfosV2(ctx, req)
			if err != nil {
				log.Error("s.arcDao.ContentFlowControlInfosV2 err=%+v", err)
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
	if len(sids) != 0 {
		g.Go(func() (err error) {
			if sm, err = s.bgm.Seasons(ctx, sids, now); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(seasonIDs) != 0 {
		g.Go(func() (err error) {
			if seasonCards, err = s.bgm.EpidsCardsInfoReply(ctx, seasonIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(metaIDs) != 0 {
		g.Go(func() (err error) {
			if atm, err = s.art.Articles(ctx, metaIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			for _, at := range atm {
				if at.Author != nil {
					mtUpIDs = append(mtUpIDs, at.Author.Mid)
				}
			}
			return
		})
	}
	if len(shopIDs) != 0 {
		g.Go(func() (err error) {
			if scm, err = s.sp.Card(ctx, shopIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(audioIDs) != 0 {
		g.Go(func() (err error) {
			if aum, err = s.audio.Audios(ctx, audioIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(actIDs) != 0 {
		g.Go(func() (err error) {
			if actm, err = s.act.Activitys(ctx, actIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(topIDs) != 0 {
		g.Go(func() (err error) {
			if topm, err = s.act.Activitys(ctx, topIDs); err != nil {
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
	upIDs = append(upIDs, avUpIDs...)
	upIDs = append(upIDs, rmUpIDs...)
	upIDs = append(upIDs, mtUpIDs...)
	g, ctx = errgroup.WithContext(c)
	if len(upIDs) != 0 {
		g.Go(func() (err error) {
			if infocard, err = s.acc.Cards3GRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		g.Go(func() (err error) {
			if upStatm, err = s.rel.StatsGRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		if mid != 0 {
			g.Go(func() error {
				follows = s.acc.Relations3GRPC(ctx, upIDs, mid)
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, c := range channelCards {
		i := &feed.Item{}
		i.Pos = c.Pos
		i.FromType = _fTypeOperation
		switch c.Type {
		case model.GotoAv, model.GotoUpRcmdAv:
			a := am[c.Value]
			isOsea := model.IsOverseas(plat)
			if a != nil && a.Arc != nil && a.Arc.IsNormal() && (!isOsea || (isOsea && getAttrBitValueFromInfosV2(flowInfosV2Reply, c.Value, model.OverseaBlockKey) == 0)) {
				i.FromPlayerAv(a)
				i.FromDislikeReason()
				i.FromRcmdReason(c)
				if follows[i.Mid] {
					i.IsAtten = 1
					if i.RcmdReason != nil && i.RcmdReason.Content == "已关注" {
						i.RcmdReason.Content = ""
					}
				}
				//for GotoUpRcmdAv
				i.Goto = c.Type
				is = append(is, i)
			}
		case model.GotoLive:
			if r, ok := rm[c.Value]; ok {
				i.FromLive(r)
				if card, ok := infocard[i.Mid]; ok {
					if card.Official.Role != 0 {
						role := card.Official.Role
						// nolint:gomnd
						if card.Official.Role == 7 {
							role = 1
						}
						i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
					}
				}
				if stat, ok := upStatm[i.Mid]; ok {
					i.Fans = stat.Follower
				}
				if follows[i.Mid] {
					i.IsAtten = 1
				}
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoBangumi:
			if s, ok := sm[c.Value]; ok {
				i.FromSeason(s)
				is = append(is, i)
			}
		case model.GotoPGC:
			if s, ok := seasonCards[int32(c.Value)]; ok {
				i.FromPGCSeason(s)
				is = append(is, i)
			}
		case model.GotoActivity:
			if act, ok := actm[c.Value]; ok && act.H5Cover != "" && act.H5URL != "" {
				i.FromActivity(act, now)
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoTopic:
			if top, ok := topm[c.Value]; ok && top.H5Cover != "" && top.H5URL != "" {
				i.FromTopic(top)
				is = append(is, i)
			}
		case model.GotoSpecial:
			if sc, ok := s.specialCardCache[c.Value]; ok {
				i.FromSpecial(sc.ID, sc.Title, sc.Cover, sc.Desc, sc.ReValue, sc.ReType, sc.Badge, sc.Size)
			}
			if i.Goto != "" {
				is = append(is, i)
			}
		case model.GotoSpecialS:
			if sc, ok := s.specialCardCache[c.Value]; ok {
				i.FromSpecialS(sc.ID, sc.Title, sc.Cover, sc.Desc, sc.ReValue, sc.ReType, sc.Badge)
			}
			if i.Goto != "" {
				is = append(is, i)
			}
		case model.GotoTopstick:
			if sc, ok := s.specialCardCache[c.Value]; ok {
				i.FromTopstick(sc.ID, sc.Title, sc.Cover, sc.Desc, sc.ReValue, sc.ReType)
				top = i
			}
		case model.GotoConverge:
			if cc, ok := convergem[c.Value]; ok {
				i.FromConverge(cc, am, rm, atm)
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoGameDownload:
			if gd, ok := downloadm[c.Value]; ok {
				i.FromGameDownload(gd, plat, build)
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoGameDownloadS:
			if gd, ok := downloadm[c.Value]; ok {
				i.FromGameDownloadS(gd, plat, build)
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoArticle:
			if m, ok := atm[c.Value]; ok {
				i.FromArticle(m)
				if card, ok := infocard[i.Mid]; ok {
					// nolint:gomnd
					if card.Official.Role != 0 {
						role := card.Official.Role
						if card.Official.Role == 7 {
							role = 1
						}
						i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
					}
				}
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoArticleS:
			if m, ok := atm[c.Value]; ok {
				i.FromArticleS(m)
				if card, ok := infocard[i.Mid]; ok {
					// nolint:gomnd
					if card.Official.Role != 0 {
						role := card.Official.Role
						if card.Official.Role == 7 {
							role = 1
						}
						i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
					}
				}
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoShoppingS:
			if c, ok := scm[c.Value]; ok {
				i.FromShoppingS(c)
				if i.Goto != "" {
					is = append(is, i)
				}
			}
		case model.GotoAudio:
			if au, ok := aum[c.Value]; ok {
				i.FromAudio(au)
				is = append(is, i)
			}
		case model.GotoPlayer:
			if a, ok := am[c.Value]; ok && a != nil && a.Arc != nil {
				i.FromPlayer(a)
				if i.Goto != "" {
					if follows[i.Mid] {
						i.IsAtten = 1
					}
					if card, ok := infocard[i.Mid]; ok {
						// nolint:gomnd
						if card.Official.Role != 0 {
							role := card.Official.Role
							if card.Official.Role == 7 {
								role = 1
							}
							i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
						}
					}
					i.FromDislikeReason()
					is = append(is, i)
				}
			}
		case model.GotoPlayerLive:
			if r, ok := rm[c.Value]; ok {
				i.FromPlayerLive(r)
				if i.Goto != "" {
					if follows[i.Mid] {
						i.IsAtten = 1
					}
					if card, ok := infocard[i.Mid]; ok {
						if card.Official.Role != 0 {
							role := card.Official.Role
							// nolint:gomnd
							if card.Official.Role == 7 {
								role = 1
							}
							i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
						}
					}
					if stat, ok := upStatm[i.Mid]; ok {
						i.Fans = stat.Follower
					}
					is = append(is, i)
				}
			}
		case model.GotoSubscribe:
			if c, ok := followm[c.Value]; ok {
				i.FromSubscribe(c, infocard, follows, upStatm, tagm)
				is = append(is, i)
			}
		case model.GotoChannelRcmd:
			if c, ok := followm[c.Value]; ok {
				i.FromChannelRcmd(c, am, tagm)
				is = append(is, i)
			}
		case model.GotoLiveUpRcmd:
			if l, ok := liveUpm[c.Value]; ok {
				i.FromLiveUpRcmd(c.Value, l, infocard)
				is = append(is, i)
			}
		}
	}
	if len(listAID) > 0 {
		isOsea := model.IsOverseas(plat)
		for _, aid := range listAID {
			if _, ok := cardAids[aid]; ok {
				continue
			}
			i := &feed.Item{}
			a := am[aid]
			if a != nil && a.Arc != nil && a.Arc.IsNormal() && (!isOsea || (isOsea && getAttrBitValueFromInfosV2(flowInfosV2Reply, aid, model.OverseaBlockKey) == 0)) {
				i.FromType = _fTypeRecommend
				i.FromPlayerAv(a)
				i.FromDislikeReason()
				//for GotoUpRcmdAv
				i.Goto = model.GotoAv
				is = append(is, i)
			}
		}
	}
	rl := len(is)
	if rl == 0 {
		is = _emptyItem
		return
	}
	if idx == 0 {
		idx = now.Unix()
	}
	for i, r := range is {
		if pull {
			r.Idx = idx + int64(rl-i)
		} else {
			r.Idx = idx - int64(i+1)
		}
	}
	return
}

func (s *Service) feedItem(c context.Context, plat int8, aids []int64) (is []*feed.Item, err error) {
	const _fTypeRecommend = "recommend"
	if len(aids) == 0 {
		is = _emptyItem
		return
	}
	var aidsV2 []*arcgrpc.PlayAv
	for _, aid := range aids {
		if aid != 0 {
			aidsV2 = append(aidsV2, &arcgrpc.PlayAv{Aid: aid})
		}
	}
	var (
		flowInfosV2Reply *cfcgrpc.FlowCtlInfosV2Reply
		channelids       = make(map[int64]*arcgrpc.ArcPlayer, len(aids))
	)
	eg := errgroupv2.WithContext(c)
	eg.Go(func(ctx context.Context) error {
		channelids, err = s.Archives(c, aidsV2, true)
		if err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		req := makeContentFlowControlInfosV2Params(s.c.CfcSvrConfig, aids)
		flowInfosV2Reply, err = s.arc.ContentFlowControlInfosV2(ctx, req)
		if err != nil {
			log.Error("s.arcDao.ContentFlowControlInfosV2 err=%+v", err)
			return nil
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	if len(channelids) > 0 {
		isOsea := model.IsOverseas(plat)
		for _, aid := range aids {
			i := &feed.Item{}
			i.FromType = _fTypeRecommend
			a := channelids[aid]
			if a != nil && a.Arc != nil && a.Arc.IsNormal() && (!isOsea || (isOsea && getAttrBitValueFromInfosV2(flowInfosV2Reply, aid, model.OverseaBlockKey) == 0)) {
				i.FromPlayerAv(a)
				i.FromDislikeReason()
				i.Goto = model.GotoAv
				is = append(is, i)
			}
		}
	}
	return
}
