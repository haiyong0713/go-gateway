package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"

	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/card/rank"
	"go-gateway/app/app-svr/app-card/interface/model/card/show"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	stat2 "go-gateway/app/app-svr/app-feed/interface/model/stat"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/pkg/adresource"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
)

const (
	_cardAdAv    = 1
	_cardAdWeb   = 2
	_cardAdWebS  = 3
	_cardAdLarge = 7
	_feedgroups  = "tianma2.0_autoplay_card"
)

// Index is
func (s *Service) Index(c context.Context, mid int64, plat int8, build int, buvid, network, mobiApp, device, platform, openEvent string, loginEvent int, idx int64, pull bool, now time.Time, bannerHash, adExtra, interest string, style, flush, autoplayCard int,
	accessKey, actionKey, appkey, statistics string) (is []*feed.Item, userFeature json.RawMessage, isRcmd, newUser bool, code, clean int, autoPlayInfoc string, info *locgrpc.InfoReply, err error) {
	var (
		rs        *feed.AIResponse
		adm       map[int32]*cm.AdInfo
		adAidm    map[int64]struct{}
		hasBanner bool
		bs        []*banner.Banner
		version   string
		adInfom   map[int]*cm.AdInfo
		autoPlay  int
		ip        = metadata.String(c, metadata.RemoteIP)
	)
	//abtest================
	// if mid > 0 && mid%20 == 19 {
	// 	clean = 1
	// } else {
	// 	clean = 0
	// }
	clean = 0
	// ipad 不允许自动播放、不在实验里面也不允许自动播放
	autoPlay = 2
	//nolint:gomnd
	if crc32.ChecksumIEEE([]byte(buvid+_feedgroups))%100 < 5 {
		//nolint:gomnd
		switch autoplayCard {
		case 0, 1, 2, 3:
			autoPlay = 1
		}
	}
	autoPlayInfoc = fmt.Sprintf("%d|%d", autoPlay, autoplayCard)
	if info, err = s.loc.InfoGRPC(c, ip); err != nil {
		log.Warn("s.loc.InfoGRPC(%v) error(%v)", ip, err)
		err = nil
	}
	//abtest================
	group := s.group(mid, buvid)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() error {
		rs = s.indexRcmd(ctx, plat, build, buvid, mid, group, loginEvent, 0, info, interest, network, style, -1, flush, autoPlayInfoc, openEvent, bannerHash, now, mobiApp)
		userFeature = rs.UserFeature
		isRcmd = rs.IsRcmd
		newUser = rs.NewUser
		code = rs.RespCode
		return nil
	})
	// 暂停实验
	// if !((group == 18 || group == 19) && style == 3) {
	g.Go(func() (err error) {
		if adm, adAidm, err = s.indexAd(ctx, plat, build, buvid, mid, network, mobiApp, device, openEvent, info, now, adExtra, style); err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	// }
	g.Go(func() (err error) {
		if hasBanner, bs, version, err = s.indexBanner(ctx, plat, build, buvid, mid, loginEvent, bannerHash, network, mobiApp, device, "", adExtra); err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	if err = g.Wait(); err != nil {
		return
	}
	rs.Items, adInfom = s.mergeItem(c, mid, rs.Items, adm, adAidm, hasBanner, plat)
	is, isRcmd, err = s.dealItem(c, mid, plat, build, buvid, platform, rs.Items, bs, version, isRcmd, network, mobiApp, device, openEvent, idx, pull, now, adExtra, adInfom, autoPlay, accessKey, actionKey, appkey, statistics)
	return
}

// Dislike is.
func (s *Service) Dislike(c context.Context, mid, id int64, buvid, gt string, reasonID, cmreasonID, feedbackID, upperID, rid, tagID int64,
	adcb, fromspmid, frommodule string, now time.Time, disableRcmd, fromAvid, fromType, materialId int64, reportData string) (err error) {
	switch gt {
	case model.GotoAv:
		s.blk.AddBlacklist(mid, id)
	case model.GotoLive:
		rid = tagID
		tagID = 0
	}
	return s.rcmd.PubDislike(c, buvid, gt, id, mid, reasonID, cmreasonID, feedbackID, upperID, rid, tagID, adcb,
		fromspmid, frommodule, now, disableRcmd, fromAvid, fromType, materialId, reportData)
}

// DislikeCancel is.
func (s *Service) DislikeCancel(c context.Context, mid, id int64, buvid, gt string, reasonID, cmreasonID, feedbackID,
	upperID, rid, tagID int64, adcb, fromspmid, frommodule string, now time.Time, closeRcmd, fromAvid, fromType,
	materialId int64, reportData string) (err error) {
	switch gt {
	case model.GotoAv:
		s.blk.DelBlacklist(mid, id)
	case model.GotoLive:
		rid = tagID
		tagID = 0
	}
	return s.rcmd.PubDislikeCancel(c, buvid, gt, id, mid, reasonID, cmreasonID, feedbackID, upperID, rid, tagID, adcb,
		fromspmid, frommodule, now, closeRcmd, fromAvid, fromType, materialId, reportData)
}

func (s *Service) indexRcmd(c context.Context, plat int8, build int, buvid string, mid int64, group int, loginEvent, parentMode int, zone *locgrpc.InfoReply, interest, network string, style int, column cdm.ColumnStatus, flush int, autoPlay, openEvent, bannerHash string, now time.Time, mobiApp string) (res *feed.AIResponse) {
	count := s.indexCount(plat, nil)
	if buvid != "" || mid != 0 {
		var (
			err    error
			zoneID int64
		)
		if zone != nil {
			zoneID = zone.ZoneId
		}
		if res, err = s.rcmd.Recommend(c, plat, buvid, mid, build, loginEvent, parentMode, 0, 0, 0, zoneID, group, interest,
			network, style, column, flush, count, 0, 0, 0, autoPlay, "", openEvent, bannerHash, "", "", "", 0, 0, 0, mobiApp, "", false, 0, 0, 0, now, 0, 0, "", "", 0, 0, "", 0); err != nil {
			log.Error("%+v", err)
		} else if len(res.Items) != 0 {
			res.IsRcmd = true
		}
		var fromCache bool
		if len(res.Items) == 0 && mid != 0 && !ecode.ServiceUnavailable.Equal(err) {
			res.Items = s.recommendCache(count)
			if len(res.Items) != 0 {
				s.pHit.Incr("index_cache")
			} else {
				s.pMiss.Incr("index_cache")
			}
			fromCache = true
		}
		if len(res.Items) == 0 || (fromCache && len(res.Items) < count) {
			res.Items = s.recommendCache(count)
		}
	} else {
		res = &feed.AIResponse{Items: s.recommendCache(count)}
		s.errProm.Incr("Buvid_empty")
		log.Warn("[BuvidEmpty] Plat %d, Build %d, Buvid %s, Mid %d, Group %d, LoginEvent %d, ParentMode %d, Interest %s, Network %s, Style %d, Column %d, Flush %d, Autoplay %s",
			plat, build, buvid, mid, group, loginEvent, parentMode, interest, network, style, column, flush, autoPlay)
	}
	return
}

func (s *Service) indexAd(c context.Context, plat int8, build int, buvid string, mid int64, network, mobiApp, device, openEvent string, zone *locgrpc.InfoReply, now time.Time, adExtra string, style int) (adm map[int32]*cm.AdInfo, adAidm map[int64]struct{}, err error) {
	var advert *cm.Ad
	//resource := s.adResource(c, plat, build)
	resource := s.adResource(c, plat, build)
	if resource == 0 {
		return
	}
	stat2.MetricCMResource.Inc(strconv.FormatInt(resource, 10), strconv.FormatInt(int64(plat), 10))
	//  兼容老的style逻辑，3为新单列，上报给商业产品的参数定义为：1 单列 2双列
	//nolint:gomnd
	if style == 3 {
		style = 1
	}
	var country, province, city string
	if zone != nil {
		country = zone.Country
		province = zone.Province
		city = zone.City
	}
	if advert, err = s.ad.Ad(c, mid, build, buvid, []int64{resource}, country, province, city, network, mobiApp, device, openEvent, adExtra, style, now); err != nil {
		return
	}
	if advert == nil || len(advert.AdsInfo) == 0 {
		return
	}
	if adsInfo, ok := advert.AdsInfo[resource]; ok {
		adm = make(map[int32]*cm.AdInfo, len(adsInfo))
		adAidm = make(map[int64]struct{}, len(adsInfo))
		for source, info := range adsInfo {
			if info == nil {
				continue
			}
			var adInfo *cm.AdInfo
			if info.AdInfo != nil {
				adInfo = info.AdInfo
				adInfo.RequestID = advert.RequestID
				adInfo.Resource = resource
				adInfo.Source = source
				adInfo.IsAd = info.IsAd
				adInfo.IsAdLoc = true
				adInfo.CmMark = info.CmMark
				adInfo.Index = info.Index
				adInfo.CardIndex = info.CardIndex
				adInfo.ClientIP = advert.ClientIP
				if adInfo.CreativeID != 0 && adInfo.CardType == _cardAdAv {
					adAidm[adInfo.CreativeContent.VideoID] = struct{}{}
				}
			} else {
				adInfo = &cm.AdInfo{RequestID: advert.RequestID, Resource: resource, Source: source, IsAdLoc: true, IsAd: info.IsAd, CmMark: info.CmMark, Index: info.Index, CardIndex: info.CardIndex, ClientIP: advert.ClientIP}
			}
			adm[adInfo.CardIndex-1] = adInfo
		}
	}
	return
}

func (s *Service) indexBanner(c context.Context, plat int8, build int, buvid string, mid int64, loginEvent int, hash, network, mobiApp, device, openEvent, adExtra string) (has bool, bs []*banner.Banner, version string, err error) {
	const (
		_androidBanBannerHash = 515009
		_iphoneBanBannerHash  = 6120
		_ipadBanBannerHash    = 6160
	)
	if s.c.Custom.ResourceDegradeSwitch {
		return
	}
	if (plat == model.PlatAndroid && build > _androidBanBannerHash) || (plat == model.PlatIPhone && build > _iphoneBanBannerHash) || (plat == model.PlatIPad && build > _ipadBanBannerHash) || loginEvent != 0 {
		if bs, version, err = s.banners(c, plat, build, mid, buvid, network, mobiApp, device, openEvent, adExtra, "", 0, nil, 0, nil, 0); err != nil {
			return
		} else if loginEvent != 0 {
			has = true
		} else if version != "" {
			has = hash != version
		}
	}
	return
}

// nolint: gocognit
func (s *Service) mergeItem(_ context.Context, mid int64, rs []*ai.Item, adm map[int32]*cm.AdInfo, adAidm map[int64]struct{}, hasBanner bool, plat int8) (is []*ai.Item, adInfom map[int]*cm.AdInfo) {
	if len(rs) == 0 {
		return
	}
	const (
		cardIndex     = 7
		cardIndexIPad = 17
		cardOffset    = 2
	)
	if hasBanner {
		rs = append([]*ai.Item{{Goto: model.GotoBanner}}, rs...)
		for index, ad := range adm {
			if ((model.IsPad(plat) && index <= cardIndexIPad) || index <= cardIndex) && (ad.CardType == _cardAdWeb || ad.CardType == _cardAdLarge) {
				ad.CardIndex = ad.CardIndex + cardOffset
			}
		}
	}
	is = make([]*ai.Item, 0, len(rs)+len(adm))
	adInfom = make(map[int]*cm.AdInfo, len(adm))
	var existsBanner, existsAdWeb bool
	for _, r := range rs {
		for {
			if ad, ok := adm[int32(len(is))]; ok {
				if ad.CreativeID != 0 {
					var item *ai.Item
					if ad.CardType == _cardAdAv {
						item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdAv, Ad: ad}
					} else if ad.CardType == _cardAdWeb {
						item = &ai.Item{Goto: model.GotoAdWeb, Ad: ad}
						existsAdWeb = true
					} else if ad.CardType == _cardAdWebS {
						item = &ai.Item{Goto: model.GotoAdWebS, Ad: ad}
					} else if ad.CardType == _cardAdLarge {
						item = &ai.Item{Goto: model.GotoAdLarge, Ad: ad}
					} else {
						b, _ := json.Marshal(ad)
						log.Error("ad---%s", b)
						break
					}
					is = append(is, item)
					continue
				} else {
					adInfom[len(is)] = ad
				}
			}
			break
		}
		if r.Goto == model.GotoAv {
			if _, ok := adAidm[r.ID]; ok {
				continue
			}
		} else if r.Goto == model.GotoBanner {
			if existsBanner {
				continue
			} else {
				existsBanner = true
			}
		} else if r.Goto == model.GotoRank && existsAdWeb {
			continue
		} else if r.Goto == model.GotoLogin && mid != 0 {
			continue
		}
		is = append(is, r)
	}
	return
}

// nolint:gocognit
func (s *Service) dealItem(c context.Context, mid int64, plat int8, build int, buvid, platform string, rs []*ai.Item, bs []*banner.Banner, version string, isRcmd bool, network, mobiApp, device, openEvent string, idx int64, pull bool, now time.Time, adExtra string, adInfom map[int]*cm.AdInfo, autoPlay int,
	accessKey, actionKey, appkey, statistics string) (is []*feed.Item, isAI bool, err error) {
	if len(rs) == 0 {
		is = _emptyItem
		return
	}
	var (
		aids, tids, roomIDs, metaIDs, shopIDs, audioIDs []int64
		upIDs, avUpIDs, rmUpIDs, mtUpIDs                []int64
		seasonIDs, sids                                 []int32
		ranks                                           []*rank.Rank
		am                                              map[int64]*arcgrpc.Arc
		tagm                                            map[int64]*taggrpc.Tag
		follows                                         map[int64]bool
		rm                                              map[int64]*live.Room
		hasBangumiRcmd                                  bool
		update                                          *bangumi.Update
		atm                                             map[int64]*article.Meta
		scm                                             map[int64]*show.Shopping
		aum                                             map[int64]*audio.Audio
		hasBanner                                       bool
		card                                            map[int64]*accountgrpc.Card
		upStatm                                         map[int64]*relationgrpc.StatReply
		arcOK                                           bool
		seasonCards, sm                                 map[int32]*episodegrpc.EpisodeCardsProto
	)
	isAI = isRcmd
	convergem := map[int64]*operate.Converge{}
	downloadm := map[int64]*operate.Download{}
	liveUpm := map[int64][]*live.Card{}
	followm := map[int64]*operate.Follow{}
	for _, r := range rs {
		switch r.Goto {
		case model.GotoAv, model.GotoAdAv, model.GotoPlayer, model.GotoUpRcmdAv:
			if r.ID != 0 {
				aids = append(aids, r.ID)
			}
			if r.Tid != 0 {
				tids = append(tids, r.Tid)
			}
		case model.GotoLive, model.GotoPlayerLive:
			if r.ID != 0 {
				roomIDs = append(roomIDs, r.ID)
			}
		case model.GotoBangumi:
			if r.ID != 0 {
				sids = append(sids, int32(r.ID))
			}
		case model.GotoPGC:
			if r.ID != 0 {
				seasonIDs = append(seasonIDs, int32(r.ID))
			}
		case model.GotoRank:
			card, aid := s.RankCard(model.PlatIPhone)
			ranks = card
			aids = append(aids, aid...)
		case model.GotoBangumiRcmd:
			hasBangumiRcmd = true
		case model.GotoBanner:
			hasBanner = true
		case model.GotoConverge:
			if card, ok := s.convergeCache[r.ID]; ok {
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
				convergem[r.ID] = card
			}
		case model.GotoGameDownloadS:
			if card, ok := s.downloadCache[r.ID]; ok {
				downloadm[r.ID] = card
			}
		case model.GotoArticleS:
			if r.ID != 0 {
				metaIDs = append(metaIDs, r.ID)
			}
		case model.GotoShoppingS:
			if r.ID != 0 {
				shopIDs = append(shopIDs, r.ID)
			}
		case model.GotoAudio:
			if r.ID != 0 {
				audioIDs = append(audioIDs, r.ID)
			}
		case model.GotoLiveUpRcmd:
			if r.ID != 0 {
				if cs, ok := s.liveCardCache[r.ID]; ok {
					for _, c := range cs {
						upIDs = append(upIDs, c.UID)
					}
				}
			}
		case model.GotoSubscribe:
			if r.ID != 0 {
				if card, ok := s.followCache[r.ID]; ok {
					for _, item := range card.Items {
						switch item.Goto {
						case cdm.GotoMid:
							if item.Pid != 0 {
								upIDs = append(upIDs, item.Pid)
							}
						case cdm.GotoTag:
							if item.Pid != 0 {
								tids = append(tids, item.Pid)
							}
						default:
						}
					}
					followm[r.ID] = card
				}
			}
		case model.GotoChannelRcmd:
			if r.ID != 0 {
				if card, ok := s.followCache[r.ID]; ok {
					if card.Pid != 0 {
						aids = append(aids, card.Pid)
					}
					if card.Tid != 0 {
						tids = append(tids, card.Tid)
					}
					followm[r.ID] = card
				}
			}
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(aids) != 0 {
		g.Go(func() (err error) {
			if am, err = s.arc.Archives(ctx, aids, 0, "", ""); err != nil {
				return
			}
			arcOK = true
			for _, a := range am {
				avUpIDs = append(avUpIDs, a.Author.Mid)
			}
			return
		})
	}
	if len(tids) != 0 {
		g.Go(func() (err error) {
			if tagm, err = s.tg.TagsInfoByIDs(ctx, mid, tids); err != nil {
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
			for _, r := range rm {
				rmUpIDs = append(rmUpIDs, r.UID)
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
			if seasonCards, err = s.bgm.CardsInfoReply(ctx, seasonIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	// TODO DEL
	// if hasBangumiRcmd && mid != 0 {
	if hasBangumiRcmd {
		g.Go(func() (err error) {
			if update, err = s.bgm.Updates(ctx, mid, now); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if hasBanner && version == "" && !s.c.Custom.ResourceDegradeSwitch {
		g.Go(func() (err error) {
			if bs, version, err = s.banners(ctx, plat, build, mid, buvid, network, mobiApp, device, openEvent, adExtra, "", 0, nil, 0, nil, 0); err != nil {
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
			if scm, err = s.show.Card(ctx, shopIDs); err != nil {
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
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		if isRcmd {
			count := s.indexCount(plat, nil)
			rs = s.recommendCache(count)
		}
	} else {
		upIDs = append(upIDs, avUpIDs...)
		upIDs = append(upIDs, rmUpIDs...)
		upIDs = append(upIDs, mtUpIDs...)
		g, ctx = errgroup.WithContext(c)
		if len(upIDs) != 0 {
			g.Go(func() (err error) {
				if card, err = s.acc.Cards3GRPC(ctx, upIDs); err != nil {
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
		if err := g.Wait(); err != nil {
			log.Error("Failed to wait: %+v", err)
		}
	}
	isAI = isAI && arcOK
	// init feed items
	is = make([]*feed.Item, 0, len(rs))
	var (
		smallCardCnt  int
		middleCardCnt int
	)
	ip := metadata.String(c, metadata.RemoteIP)
	adm := map[int32]*feed.Item{}
	isIpad := plat == model.PlatIPad
	for _, r := range rs {
		il := int32(len(is))
		i := &feed.Item{AI: r}
		i.FromRcmd(r)
		switch r.Goto {
		case model.GotoAv, model.GotoUpRcmdAv:
			a, ok := am[r.ID]
			if !ok && !arcOK {
				a = r.Archive
			}
			if a != nil && a.IsNormal() {
				i.FromPlayerAv(a)
				if arcOK {
					if info, ok := tagm[r.Tid]; ok {
						i.Tag = &feed.Tag{TagID: info.Id, TagName: info.Name, IsAtten: int8(info.Attention), Count: &feed.TagCount{Atten: int(info.Sub)}}
					}
				} else if r.Tag != nil {
					i.Tag = &feed.Tag{TagID: r.Tag.Id, TagName: r.Tag.Name}
				}
				i.FromDislikeReason(plat, build)
				i.FromRcmdReason(r.RcmdReason)
				if follows[i.Mid] {
					i.IsAtten = 1
				}
				if card, ok := card[i.Mid]; ok {
					if card.Official.Role != 0 {
						role := card.Official.Role
						//nolint:gomnd
						if card.Official.Role == 7 {
							role = 1
						}
						i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
					}
				}
				// for GotoUpRcmdAv
				i.Goto = r.Goto
				if i.Goto == model.GotoUpRcmdAv {
					// TODO 等待开启
					// percent := i.Like / (i.Like + i.Dislike) * 100
					// if percent != 0 {
					// 	i.Desc = strconv.Itoa(percent) + "%的人推荐"
					// }
					i.Desc = ""
				}
				is = append(is, i)
				smallCardCnt++
			}
		case model.GotoLive:
			if r, ok := rm[r.ID]; ok {
				i.FromLive(r)
				if card, ok := card[i.Mid]; ok {
					if card.Official.Role != 0 {
						role := card.Official.Role
						//nolint:gomnd
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
					smallCardCnt++
				}
			}
		case model.GotoBangumi:
			if s, ok := sm[int32(r.ID)]; ok && s.Season != nil {
				i.FromSeason(s)
				is = append(is, i)
				smallCardCnt++
			}
		case model.GotoPGC:
			if s, ok := seasonCards[int32(r.ID)]; ok && s.Season != nil {
				i.FromPGCSeason(s)
				is = append(is, i)
				smallCardCnt++
			}
		case model.GotoLogin:
			i.FromLogin()
			is = append(is, i)
			smallCardCnt++
		case model.GotoAdAv:
			if r.Ad != nil {
				if a, ok := am[r.ID]; ok && model.AdAvIsNormal(a) {
					i.FromAdAv(r.Ad, a)
					if follows[i.Mid] {
						i.IsAtten = 1
					}
					if card, ok := card[i.Mid]; ok {
						if card.Official.Role != 0 {
							role := card.Official.Role
							//nolint:gomnd
							if card.Official.Role == 7 {
								role = 1
							}
							i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
						}
					}
					i.ClientIP = ip
					adm[i.CardIndex-1] = i
				}
			}
		case model.GotoAdWebS:
			if r.Ad != nil {
				i.FromAdWebS(r.Ad)
				i.ClientIP = ip
				adm[i.CardIndex-1] = i
			}
		case model.GotoAdWeb:
			if r.Ad != nil {
				i.FromAdWeb(r.Ad)
				i.ClientIP = ip
				adm[i.CardIndex-1] = i
			}
		case model.GotoAdLarge:
			if r.Ad != nil {
				i.FromAdLarge(r.Ad)
				i.ClientIP = ip
				adm[i.CardIndex-1] = i
			}
		case model.GotoSpecial:
			if sc, ok := s.specialCache[r.ID]; ok {
				i.FromSpecial(sc.ID, sc.Title, sc.Cover, sc.Desc, sc.ReValue, sc.ReType, sc.Badge, sc.Size)
			}
			if i.Goto != "" {
				if !isIpad {
					if smallCardCnt%2 != 0 {
						is = swapTwoItem(is, i)
					} else {
						is = append(is, i)
					}
				} else {
					if (smallCardCnt+middleCardCnt*2)%2 != 0 {
						is = swapTwoItem(is, i)
					} else {
						is = append(is, i)
					}
					middleCardCnt++
				}
			}
		case model.GotoSpecialS:
			if sc, ok := s.specialCache[r.ID]; ok {
				i.FromSpecialS(sc.ID, sc.Title, sc.Cover, sc.SingleCover, sc.Desc, sc.ReValue, sc.ReType, sc.Badge)
			}
			if i.Goto != "" {
				if !isIpad {
					is = append(is, i)
					smallCardCnt++
				}
			}
		case model.GotoRank:
			i.FromRank(ranks, am)
			if i.Goto != "" {
				if !isIpad {
					if smallCardCnt%2 != 0 {
						is = swapTwoItem(is, i)
					} else {
						is = append(is, i)
					}
				} else {
					if (smallCardCnt+middleCardCnt*2)%2 != 0 {
						is = swapTwoItem(is, i)
					} else {
						is = append(is, i)
					}
					middleCardCnt++
				}
			}
		case model.GotoBangumiRcmd:
			if mid != 0 && update != nil && update.Updates != 0 {
				i.FromBangumiRcmd(update)
				if !isIpad {
					if smallCardCnt%2 != 0 {
						is = swapTwoItem(is, i)
					} else {
						is = append(is, i)
					}
				} else {
					is = append(is, i)
					smallCardCnt++
				}
			}
		case model.GotoBanner:
			if len(bs) != 0 {
				i.FromBanner(bs, version)
				if !isIpad {
					if smallCardCnt%2 != 0 {
						is = swapTwoItem(is, i)
					} else {
						is = append(is, i)
					}
				} else {
					//nolint:gomnd
					switch (smallCardCnt + middleCardCnt*2) % 4 {
					case 0:
						is = append(is, i)
					case 1:
						is = swapTwoItem(is, i)
					case 2:
						//nolint:gomnd
						switch is[len(is)-1].Goto {
						case model.GotoRank, model.GotoAdWeb, model.GotoAdLarge:
							is = swapTwoItem(is, i)
						default:
							is = swapThreeItem(is, i)
						}
					case 3:
						is = swapThreeItem(is, i)
					}
				}
			}
		case model.GotoConverge:
			if cc, ok := convergem[r.ID]; ok {
				i.FromConverge(cc, am, rm, atm)
				if i.Goto != "" {
					if !isIpad {
						if smallCardCnt%2 != 0 {
							is = swapTwoItem(is, i)
						} else {
							is = append(is, i)
						}
					}
				}
			}
		case model.GotoGameDownloadS:
			if gd, ok := downloadm[r.ID]; ok {
				i.FromGameDownloadS(gd, plat, build)
				if i.Goto != "" {
					if !isIpad {
						is = append(is, i)
						smallCardCnt++
					}
				}
			}
		case model.GotoArticleS:
			if m, ok := atm[r.ID]; ok {
				i.FromArticleS(m)
				if card, ok := card[i.Mid]; ok {
					if card.Official.Role != 0 {
						role := card.Official.Role
						//nolint:gomnd
						if card.Official.Role == 7 {
							role = 1
						}
						i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
					}
				}
				if i.Goto != "" {
					if !isIpad {
						is = append(is, i)
						smallCardCnt++
					}
				}
			}
		case model.GotoShoppingS:
			if c, ok := scm[r.ID]; ok {
				i.FromShoppingS(c)
				if i.Goto != "" {
					if !isIpad {
						is = append(is, i)
						smallCardCnt++
					}
				}
			}
		case model.GotoAudio:
			if au, ok := aum[r.ID]; ok {
				i.FromAudio(au)
				is = append(is, i)
				smallCardCnt++
			}
		case model.GotoPlayer:
			if a, ok := am[r.ID]; ok {
				i.FromPlayer(a)
				if i.Goto != "" {
					if info, ok := tagm[r.Tid]; ok {
						i.Tag = &feed.Tag{TagID: info.Id, TagName: info.Name, IsAtten: int8(info.Attention), Count: &feed.TagCount{Atten: int(info.Sub)}}
					}
					if follows[i.Mid] {
						i.IsAtten = 1
					}
					if card, ok := card[i.Mid]; ok {
						if card.Official.Role != 0 {
							role := card.Official.Role
							//nolint:gomnd
							if card.Official.Role == 7 {
								role = 1
							}
							i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
						}
					}
					i.FromDislikeReason(plat, build)
					if !isIpad {
						if smallCardCnt%2 != 0 {
							is = swapTwoItem(is, i)
						} else {
							is = append(is, i)
						}
					}
				}
			}
		case model.GotoPlayerLive:
			if r, ok := rm[r.ID]; ok {
				i.FromPlayerLive(r)
				if i.Goto != "" {
					if follows[i.Mid] {
						i.IsAtten = 1
					}
					if card, ok := card[i.Mid]; ok {
						if card.Official.Role != 0 {
							role := card.Official.Role
							//nolint:gomnd
							if card.Official.Role == 7 {
								role = 1
							}
							i.Official = &feed.OfficialInfo{Role: role, Title: card.Official.Title, Desc: card.Official.Desc}
						}
					}
					if stat, ok := upStatm[i.Mid]; ok {
						i.Fans = stat.Follower
					}
					if !isIpad {
						if smallCardCnt%2 != 0 {
							is = swapTwoItem(is, i)
						} else {
							is = append(is, i)
						}
					}
				}
			}
		case model.GotoSubscribe:
			if c, ok := followm[r.ID]; ok {
				if !isIpad {
					i.FromSubscribe(c, card, follows, upStatm, tagm)
					if i.Goto != "" {
						if smallCardCnt%2 != 0 {
							is = swapTwoItem(is, i)
						} else {
							is = append(is, i)
						}
					}
				}
			}
		case model.GotoChannelRcmd:
			if c, ok := followm[r.ID]; ok {
				if !isIpad {
					i.FromChannelRcmd(c, am, tagm)
					if i.Goto != "" {
						if !isIpad {
							is = append(is, i)
							smallCardCnt++
						}
					}
				}
			}
		case model.GotoLiveUpRcmd:
			if c, ok := liveUpm[r.ID]; ok {
				if !isIpad {
					i.FromLiveUpRcmd(r.ID, c, card)
					if i.Goto != "" {
						if smallCardCnt%2 != 0 {
							is = swapTwoItem(is, i)
						} else {
							is = append(is, i)
						}
					}
				}
			}
		default:
			log.Warn("v1 unexpected goto(%s) %+v", r.Goto, r)
			continue
		}
		if ad, ok := adm[il]; ok {
			switch ad.Goto {
			case model.GotoAdAv, model.GotoAdWebS:
				is = append(is, ad)
				smallCardCnt++
			case model.GotoAdWeb, model.GotoAdLarge:
				if !isIpad {
					if smallCardCnt%2 != 0 {
						is = swapTwoItem(is, ad)
					} else {
						is = append(is, ad)
					}
				} else {
					if (smallCardCnt+middleCardCnt*2)%2 != 0 {
						is = swapTwoItem(is, ad)
					} else {
						is = append(is, ad)
					}
					middleCardCnt++
				}
			}
		}
	}
	if !isIpad {
		is = is[:len(is)-smallCardCnt%2]
	} else {
		//nolint:gomnd
		switch (smallCardCnt + middleCardCnt*2) % 4 {
		case 1:
			is = is[:len(is)-1]
		case 2:
			if isMiddleCard(is[len(is)-1].Goto) {
				is = is[:len(is)-1]
			} else {
				is = is[:len(is)-2]
			}
		case 3:
			if isMiddleCard(is[len(is)-1].Goto) {
				is = is[:len(is)-2]
			} else {
				is = is[:len(is)-3]
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
		if ad, ok := adInfom[i]; ok {
			r.SrcID = ad.Source
			r.RequestID = ad.RequestID
			r.IsAdLoc = ad.IsAdLoc
			r.IsAd = ad.IsAd
			r.CmMark = ad.CmMark
			r.AdIndex = ad.Index
			r.ClientIP = ip
			r.CardIndex = int32(i + 1)
		} else if r.IsAd {
			r.CardIndex = int32(i + 1)
		}
		if i == 0 {
			r.AutoplayCard = autoPlay
		}
	}
	return
}

func (s *Service) adResource(ctx context.Context, plat int8, build int) (resource int64) {
	const (
		_androidBanAd = 500001
	)
	scene := adresource.EmptyScene
	switch plat {
	case model.PlatIPhone, model.PlatIPhoneB:
		scene = adresource.PegasusIOS
	case model.PlatIPadHD, model.PlatIPad:
		scene = adresource.PegasusIPad
	case model.PlatAndroid:
		if build >= _androidBanAd {
			scene = adresource.PegasusAndroid
		}
	case model.PlatAndroidB:
		scene = adresource.PegasusAndroid
	default:
		log.Info("Failed to match scene by plat: %d", plat)
	}
	resourceId, ok := adresource.CalcResourceID(ctx, scene)
	if !ok {
		return 0
	}
	return int64(resourceId)
}

func swapTwoItem(rs []*feed.Item, i *feed.Item) (is []*feed.Item) {
	rs[len(rs)-1].Idx, i.Idx = i.Idx, rs[len(rs)-1].Idx
	is = append(rs, rs[len(rs)-1])
	is[len(is)-2] = i
	return
}

func swapThreeItem(rs []*feed.Item, i *feed.Item) (is []*feed.Item) {
	rs[len(rs)-1].Idx, i.Idx = i.Idx, rs[len(rs)-1].Idx
	rs[len(rs)-2].Idx, rs[len(is)-1].Idx = rs[len(rs)-1].Idx, rs[len(rs)-2].Idx
	is = append(rs, rs[len(rs)-1])
	is[len(is)-2] = i
	is[len(is)-3], is[len(is)-2] = is[len(is)-2], is[len(is)-3]
	return
}

func isMiddleCard(gt string) bool {
	return gt == model.GotoRank || gt == model.GotoAdWeb || gt == model.GotoPlayer ||
		gt == model.GotoPlayerLive || gt == model.GotoConverge || gt == model.GotoSpecial || gt == model.GotoAdLarge || gt == model.GotoLiveUpRcmd
}

func (s *Service) indexCount(plat int8, abtest *feed.Abtest) (count int) {
	// ai侧目前未使用网关传递的count值，ai依据plat自行转换
	if plat == model.PlatIPad || plat == model.PlatIPadHD || plat == model.PlatWPhone {
		count = s.c.Feed.Index.IPadCount
	} else {
		count = s.c.Feed.Index.Count
	}
	// 命中3列实验
	if abtest != nil && abtest.IpadHDThreeColumn == 1 {
		count = int(s.c.Feed.Index.IpadHDThreeColumnCount)
	}
	return
}
