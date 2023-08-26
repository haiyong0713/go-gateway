package v1

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	"go-common/library/text/translate/chinese.v2"

	errgroupv2 "go-common/library/sync/errgroup.v2"

	cardmdl "go-gateway/app/app-svr/app-card/interface/model"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	searchadm "go-gateway/app/app-svr/app-feed/admin/model/search"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	midint64 "go-gateway/app/app-svr/app-interface/interface-legacy/middleware/midInt64"
	"go-gateway/app/app-svr/app-interface/interface-legacy/middleware/stat"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model"
	"go-gateway/app/app-svr/app-search/configs"
	"go-gateway/app/app-svr/app-search/internal/model/search"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	resmdl "go-gateway/app/app-svr/resource/service/model"
	siriext "go-gateway/app/app-svr/siri-ext/service/api"
	ugcSeasonGrpc "go-gateway/app/app-svr/ugc-season/service/api"

	account "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	managersearch "git.bilibili.co/bapis/bapis-go/ai/search/mgr/interface"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	esportGRPC "git.bilibili.co/bapis/bapis-go/esports/service"
	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	gameentry "git.bilibili.co/bapis/bapis-go/manager/operation/game-entry"
	esportsservice "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	gallerygrpc "git.bilibili.co/bapis/bapis-go/pangu/platform/gallery-service"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	pgcsearch "git.bilibili.co/bapis/bapis-go/pgc/service/card/search/v1"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	pgcstat "git.bilibili.co/bapis/bapis-go/pgc/service/stat/v1"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	"github.com/pkg/errors"
)

// const for search
const (
	_oldAndroid = 514000
	_oldIOS     = 6090

	IPhoneSearchResourceID  = 2447
	AndroidSearchResourceID = 2450
	IPadSearchResourceID    = 2811

	_liveRoomWordType = 7

	// 默认zoneid 中国
	_defaultZoneID = 4194304

	_rankingConfigResTypeVideo = "video"
	_rankingConfigResTypeLive  = "live"
)

var (
	_emptyItem   []*search.Item
	_emptyResult = &search.Result{
		NavInfo: []*search.NavInfo{},
		Page:    0,
		Items: search.ResultItems{
			Season:   _emptyItem,
			Upper:    _emptyItem,
			Movie:    _emptyItem,
			Archive:  _emptyItem,
			LiveRoom: _emptyItem,
			LiveUser: _emptyItem,
		},
	}
)

func (s *Service) loadSpecialCache() {
	special, err := s.dao.SpecialCards(context.Background())
	if err != nil {
		log.Error("日志告警 搜索特殊小卡加载失败：%+v", err)
		return
	}
	if len(special) > 0 {
		s.specialCache = special
	}
}

func (s *Service) loadSystemNotice() {
	notices, err := s.dao.ALLSearchSystemNotice(context.Background())
	if err != nil {
		log.Error("日志告警 搜索系统提示加载失败：%+v", err)
		return
	}
	if len(notices) > 0 {
		s.systemNotice = notices
	}
}

func (s *Service) initCron() {
	s.loadHotCache()
	s.loadSearchTipsCache()
	s.loadSpecialCache()
	s.loadSystemNotice()
	if err := s.cron.AddFunc(s.c.Cron.LoadHotCache, s.loadHotCache); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc(s.c.Cron.LoadSearchTipsCache, s.loadSearchTipsCache); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc(s.c.Cron.LoadSpecialCache, s.loadSpecialCache); err != nil {
		panic(err)
	}
	if err := s.cron.AddFunc(s.c.Cron.LoadSystemNotice, s.loadSystemNotice); err != nil {
		panic(err)
	}
}

// Search get all type search data.
//
//nolint:gocognit
func (s *Service) Search(c context.Context, mid int64, mobiApp, device, platform, buvid, keyword, duration, order,
	filtered, lang, fromSource, recommend, parent, adExtra, extraWord, tidList, durationList, qvid string, plat int8, rid, highlight, build, pn, ps, isQuery, teenagersMode, lessonsMode int,
	qn, fnver, fnval, fourk int64, old bool, now time.Time, localTime, autoPlayCard int64) (res *search.Result, code int, err error) {
	const (
		_newIPhonePGC      = 6500
		_newAndroidPGC     = 519010
		_newIPhoneSearch   = 6500
		_newIPhoneISearch  = 63600000
		_newAndroidSearch  = 5215000
		_newAndroidBSearch = 591200
		_newAndroidISearch = 3000000
	)
	var (
		newPGC, flow, isNewTwitter, isBlue bool
		avids                              []int64
		inlineAvids                        []int64
		owners                             []int64
		follows                            map[int64]bool
		roomIDs                            []int64
		entryRoom                          map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
		seasonIDs                          []int64
		seasonStatIDs                      []int32
		seasonstat                         map[int32]*pgcstat.SeasonStatProto
		sepReqs                            []*pgcsearch.SeasonEpReq
		bangumis                           map[string]*search.Card
		seasonEps                          map[int32]*pgcsearch.SearchCardProto
		medisas                            map[int32]*pgcsearch.SearchMediaProto
		comicIDs                           []int64
		// tagSeasonIDs []int32
		seasonInfoBangumi                  map[int32]*seasongrpc.CardInfoProto
		tags                               []int64
		tagMyInfos                         []*search.Tag
		dynamicIDs                         []int64
		dynamicDetails                     map[int64]*search.Detail
		dynamicTopic                       map[int64]*search.DynamicTopics
		accProfiles                        map[int64]*account.ProfileWithoutPrivacy
		nftRegion                          map[int64]*gallerygrpc.NFTRegion
		accCards                           map[int64]*account.Card
		cooperation, isNewOrder, newPlayer bool
		// 赛事卡
		matchIDs           []int64
		esportIds          []int64
		matchLiveEntryRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
		matchm             map[int64]*esportGRPC.Contest
		esportConfigs      map[int64]*managersearch.EsportConfigInfo
		isNewDuration      bool
		isNewOGVURL        bool
		// 体育卡
		sportsIds       []int64                //赛程id
		sportsSeasonIds []int64                //赛季id
		sportsMaterials *search.SportsMaterial // 体育卡聚合信息
		// 新频道卡
		chanIDs         []int64
		newChannelCards *channelgrpc.SearchChannelInHomeReply
		newChannelm     map[int64]*channelgrpc.SearchChannelInHome
		// ogv频道卡
		ogvChanIDs      []int64
		ogvChannelCards *channelgrpc.SearchChannelInHomeReply
		ogvChannelm     map[int64]*search.OgvChannelMaterial
		// 新订阅关系
		relationm map[int64]*relationgrpc.InterrelationReply
		// 推荐理由新色值
		isNewColor bool
		// 温馨推荐卡 id 提示卡片只有一张
		tipsID int64
		// 广告卡 brand_ad
		adOwners          []int64
		adOwnersEntryRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
		// 	特殊卡物料 id
		specialIDs   []int64
		specialCards map[int64]*searchadm.SpreadConfig
		playAvs      []*arcgrpc.PlayAv
		apm          map[int64]*arcgrpc.ArcPlayer
		// inline 卡用
		hasLike         map[int64]thumbupgrpc.State
		hasFav          map[int64]int8
		hasCoin         map[int64]int64
		inlineEPIDs     []int32
		inlineEPCards   map[int32]*pgcinline.EpisodeCard
		inlineEPHasLike map[int64]thumbupgrpc.State
		// 强化游戏卡用
		topGameIDs           []int64
		topGameMaterials     map[int64]*search.TopGameMaterial
		topGameConfigs       *search.TopGameConfig
		topGameInlineConfigs *search.TopGameInlineInfo
		topGameCardIds       []int64
		topGameCardIdMs      = make(map[int64]int64, 1)
		// 游戏小卡
		gameIDs        []int64
		multiGameInfos map[int64]*search.NewGame
		cloudGameReply *gameentry.MultiShowResp
		// ogv新人实验
		isOgvExpNewUser bool
		// 合集卡
		collectionCardIds []int64
		collectionViews   map[int64]*ugcSeasonGrpc.ViewReply
		// 漫画信息
		comicInfo map[int64]*search.ComicInfo
	)
	// android 概念版 591205
	if (plat == model.PlatAndroid && build >= _newAndroidPGC && build != 591205) ||
		(plat == model.PlatIPhone && build >= _newIPhonePGC && build != 7140) ||
		(plat == model.PlatIPhoneI && build >= _newIPhoneISearch) ||
		(plat == model.PlatAndroidB && build >= _newAndroidBSearch) ||
		(plat == model.PlatIPad && build >= search.SearchNewIPad) ||
		(plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) ||
		(plat == model.PlatAndroidI) ||
		model.IsIPhoneB(plat) ||
		model.IsAndroidHD(plat) {
		newPGC = true
	}
	// 处理一个ios概念版是 7140，是否需要过滤
	if (plat == model.PlatAndroid && build >= _newAndroidSearch) ||
		(plat == model.PlatIPhone && build >= _newIPhoneSearch && build != 7140) ||
		(plat == model.PlatAndroidB && build >= _newAndroidBSearch) ||
		(plat == model.PlatIPhoneI && build >= _newIPhoneISearch) ||
		model.IsIPhoneB(plat) ||
		(plat == model.PlatAndroidI && build >= _newAndroidISearch) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		flow = true
	}
	// player build limit
	if limitBuild, ok := s.c.PlayerBuildLimit[mobiApp]; ok && build > limitBuild {
		newPlayer = true
	}
	if !cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
		isBlue = true
	}
	var (
		seasonNum int
		movieNum  int
	)
	if (plat == model.PlatIPad && build >= search.SearchNewIPad) ||
		(plat == model.PlatIpadHD && build >= search.SearchNewIPadHD) ||
		(model.IsAndroidHD(plat)) {
		seasonNum = s.iPadSearchBangumi
		movieNum = s.iPadSearchFt
	} else {
		seasonNum = s.seasonNum
		movieNum = s.movieNum
	}
	if (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.NewOrderAndroid) ||
		(model.IsIPhone(plat) && build > s.c.SearchBuildLimit.NewOrderIOS) {
		if rid != 0 || duration != "0" || order != "totalrank" {
			isNewOrder = true
		}
	}
	isOgvExpNewUser = s.dao.CheckNewDeviceAndUser(c, mid, buvid, model.NewUserOgvExperimentPeriod)
	all, code, err := s.dao.Search(c, mid, mobiApp, device, platform, buvid, keyword, duration, order, filtered, fromSource, recommend, parent, adExtra, extraWord, tidList, durationList, qvid, plat, seasonNum, movieNum,
		s.upUserNum, s.uvLimit, s.userNum, s.userVideoLimitMix, s.biliUserNum, s.biliUserVideoLimitMix, rid, highlight, build, pn, ps, isQuery, teenagersMode, lessonsMode, old, isOgvExpNewUser, now, newPGC, flow,
		isNewOrder, autoPlayCard)
	if err != nil {
		stat.MetricSearchAiMainFailed.Inc("/main/search", strconv.Itoa(all.Code))
		log.Error("%+v", err)
		return
	}
	if (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.NewTwitterAndroid) ||
		(model.IsIPhone(plat) && build > s.c.SearchBuildLimit.NewTwitterIOS) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		isNewTwitter = true
	}
	if (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.VideoDurationAndroid) ||
		(model.IsIPhone(plat) && build > s.c.SearchBuildLimit.VideoDurationIOS) ||
		(model.PlatIpadHD == plat && build > 33700000) ||
		(model.PlatIPad == plat && build > 66000000) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		isNewDuration = true
	}
	if (plat == model.PlatAndroid && build > s.c.SearchBuildLimit.OGVURLAndroid) ||
		(plat == model.PlatIPhone && build > s.c.SearchBuildLimit.OGVURLIOS) ||
		(model.IsAndroidHD(plat) && build <= 1000000) {
		isNewOGVURL = true
	}
	if (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.CardOptimizeAndroid) ||
		(model.IsIPhone(plat) && build > s.c.SearchBuildLimit.CardOptimizeIPhone) ||
		(model.PlatIpadHD == plat && build > s.c.SearchBuildLimit.CardOptimizeIpadHD) ||
		model.IsAndroidHD(plat) {
		isNewColor = true
	}
	if code == model.ForbidCode || code == model.NoResultCode {
		res = _emptyResult
		err = nil
		return
	}
	res = &search.Result{}
	res.Trackid = all.Trackid
	res.QvId = all.QvId
	res.ExpStr = all.ExpStr
	res.ExtraWordList = all.ExtraWordList
	res.OriginExtraWord = all.OriginExtraWord
	res.SelectBarType = all.SelectBarType
	res.NewSearchExpNum = all.NewSearchExpNum
	res.AppDisplayOption = all.AppDisplayOption
	res.KeyWord = keyword
	res.Page = all.Page
	res.Array = all.FlowPlaceholder
	res.Attribute = all.Attribute
	if pn < all.NumPages {
		res.Next = strconv.FormatInt(int64(pn+1), 10)
	} else {
		res.Next = ""
	}

	if teenagersMode == 0 {
		res.NavInfo = s.convertNav(c, all, plat, build, lang, mobiApp, device, old, newPGC)
	}
	if len(all.FlowResult) != 0 {
		var item []*search.Item
		for _, v := range all.FlowResult {
			switch v.Type {
			case search.TypeUser, search.TypeBiliUser:
				owners = append(owners, v.User.Mid)
				for _, vr := range v.User.Res {
					avids = append(avids, vr.Aid)
					playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: vr.Aid})
				}
				// if !model.IsBlue(plat) {
				if cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
					roomIDs = append(roomIDs, v.User.RoomID)
				}
			case search.TypeVideo:
				avids = append(avids, v.Video.ID)
				playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: v.Video.ID})
				owners = append(owners, v.Video.Mid)
				if v.Video.IsUGCInline > 0 {
					inlineAvids = append(inlineAvids, v.Video.ID)
				}
			case search.TypeLive:
				if cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
					roomIDs = append(roomIDs, v.Live.RoomID)
					owners = append(owners, v.Live.UID)
				}
			case search.TypeMediaBangumi, search.TypeMediaFt:
				seasonIDs = append(seasonIDs, v.Media.SeasonID)
				if v.Media.Canplay() {
					sepReqs = append(sepReqs, v.Media.BuildPgcReq())
				}
				if v.Media.IsOGVInline > 0 {
					inlineEPIDs = append(inlineEPIDs, v.Media.EPID)
				}
			case search.TypeStar:
				if v.Star.MID != 0 {
					owners = append(owners, v.Star.MID)
				}
				if v.Star.TagID != 0 {
					tags = append(tags, v.Star.TagID)
				}
			case search.TypeArticle:
				owners = append(owners, v.Article.Mid)
			case search.TypeChannel:
				tags = append(tags, v.Channel.TagID)
				if len(v.Channel.Values) > 0 {
					for _, vc := range v.Channel.Values {
						switch vc.Type {
						case search.TypeVideo:
							if vc.Video != nil {
								avids = append(avids, vc.Video.ID)
								playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: vc.Video.ID})
							}
							// case search.TypeLive:
							//	if vc.Live != nil {
							//		roomIDs = append(roomIDs, vc.Live.RoomID)
							//	}
							// case search.TypeMediaBangumi, search.TypeMediaFt:
							//	if vc.Media != nil {
							//		tagSeasonIDs = append(tagSeasonIDs, int32(vc.Media.SeasonID))
							//	}
						}
					}
				}
			case search.TypeTwitter:
				dynamicIDs = append(dynamicIDs, v.Twitter.ID)
			case search.TypeOGVCard:
				if ogvCard := v.SearchOGVCard; ogvCard != nil {
					for _, module := range ogvCard.Modules {
						switch module.Type {
						case search.OGVCardTypePGC, search.OGVCardTypeOGVCluster:
							for _, v := range module.Values {
								for _, ssidInt32 := range v.SeasonIDList {
									seasonStatIDs = append(seasonStatIDs, int32(ssidInt32))
								}
							}
						case search.OGVCardTypeComicCluster:
							for _, v := range module.Values {
								for _, comicID := range v.ComicIDList {
									comicIDs = append(comicIDs, comicID)
								}
							}
						}
					}
				}
			case search.TypeESports:
				if v.ESport != nil {
					for _, match := range v.ESport.MatchList {
						if match != nil {
							if match.ID == 0 {
								continue
							}
							matchIDs = append(matchIDs, match.ID)
						}
					}
					esportIds = append(esportIds, v.ESport.ID)
				}
			case search.TypeNewChannel:
				if v.NewChannel != nil {
					chanIDs = append(chanIDs, v.NewChannel.ID)
				}
			case search.TypeOgvChannel:
				if v.NewChannel != nil {
					ogvChanIDs = append(ogvChanIDs, v.NewChannel.ID)
				}
			case search.TypeTips:
				if v.Tips != nil {
					tipsID = v.Tips.ID
				}
			case search.TypeBrandAD, search.TypeBrandAdGiant, search.TypeBrandAdGiantTriple, search.TypeVideoAd, search.TypePictureAd:
				adContent := v.BrandAD.GetADContent()
				if adContent != nil {
					if adContent.UPMid != 0 {
						owners = append(owners, adContent.UPMid)
						adOwners = append(adOwners, adContent.UPMid)
					}
					if len(adContent.Aids) > 0 {
						avids = append(avids, adContent.Aids...)
						for _, aid := range adContent.Aids {
							playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: aid})
						}
					}
				}
			case search.TypeSpecial, search.TypeSpecialS:
				if v.Operate != nil && v.Operate.ID > 0 {
					specialIDs = append(specialIDs, v.Operate.ID)
				}
			case search.TypeTopGame:
				if v.TopGame != nil {
					topGameIDs = append(topGameIDs, v.TopGame.ID)
					topGameCardIds = append(topGameCardIds, v.TopGame.CardId)
					topGameCardIdMs[v.TopGame.ID] = v.TopGame.CardId
				}
			case search.TypeGame:
				if v.Game != nil {
					gameIDs = append(gameIDs, v.Game.ID)
				}
			case search.TypeBrandAdAv, search.TypeBrandAdLive, search.TypeBrandAdLocalAv:
				adContent := v.BrandADInline.GetADContent()
				if adContent != nil {
					if adContent.UPMid != 0 {
						owners = append(owners, adContent.UPMid)
						adOwners = append(adOwners, adContent.UPMid)
					}
					if len(adContent.Aids) > 0 {
						// 只拿首位
						avids = append(avids, adContent.Aids[0])
						playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: adContent.Aids[0]})
						inlineAvids = append(inlineAvids, adContent.Aids[0])
					}
				}
			case search.TypeSports, search.TypeSportsVersus:
				if v.Sports != nil {
					sportsIds = append(sportsIds, v.Sports.ID)
					sportsSeasonIds = append(sportsSeasonIds, v.Sports.SeasonId)
				}
			case search.TypeCollectionCard:
				if v.CollectionCard != nil {
					collectionCardIds = append(collectionCardIds, v.CollectionCard.ID)
					owners = append(owners, v.CollectionCard.Uid)
				}
			case search.TypePediaInlineCard:
				if v.PediaCard != nil && v.PediaCard.NavigationCard.Avid > 0 {
					avids = append(avids, v.PediaCard.NavigationCard.Avid)
					playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: v.PediaCard.NavigationCard.Avid})
					inlineAvids = append(inlineAvids, v.PediaCard.NavigationCard.Avid)
				}
			}
		}
		// 新频道卡需要获取关联数据，相对其他卡片多一步
		if len(chanIDs) > 0 {
			newChannelCards, _ = s.dao.SearchChannelInHome(c, chanIDs)
			newChannelm = make(map[int64]*channelgrpc.SearchChannelInHome)
			for _, newChannelCard := range newChannelCards.GetCards() {
				if newChannelCard.GetCid() == 0 {
					continue
				}
				newChannelm[newChannelCard.GetCid()] = newChannelCard
				switch newChannelCard.GetResourceType() {
				case search.NewChannelResourceTypeArchive:
					for _, arc := range newChannelCard.GetVideoCards() {
						if arc.GetRid() == 0 {
							continue
						}
						avids = append(avids, arc.GetRid())
						playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: arc.GetRid()})
					}
				case search.NewChannelResourceTypeChildChannel:
					// 暂时不做额外逻辑，预留
				default:
					log.Error("SearchChannelInHome get unknown type %v", newChannelCard.GetResourceType())
				}
			}
		}
		g, ctx := errgroup.WithContext(c)
		if len(owners) != 0 {
			if mid > 0 {
				g.Go(func() error {
					follows = s.dao.Relations3(ctx, owners, mid)
					return nil
				})
				g.Go(func() error {
					relationm, err = s.dao.Interrelations(ctx, mid, owners)
					return nil
				})
			}
			g.Go(func() (err error) {
				if accProfiles, err = s.dao.ProfilesWithoutPrivacy3(ctx, owners); err != nil {
					log.Error("s.accDao.ProfilesWithoutPrivacy3 owners=%+v, err=%+v", owners, err)
					err = nil
				}
				return
			})
			g.Go(func() (err error) {
				if nftRegion, err = s.getNFTIconInfo(ctx, owners); err != nil {
					log.Error("getNFTIconInfo owners=%+v, err=%+v", owners, err)
					err = nil
				}
				return
			})
			g.Go(func() (err error) {
				if accCards, err = s.dao.Cards3(ctx, owners); err != nil {
					log.Error("accDao.Cards owners=%+v, err=%+v", owners, err)
					err = nil
				}
				return
			})
		}
		if len(playAvs) != 0 {
			if newPlayer {
				g.Go(func() (err error) {
					if apm, err = s.dao.ArcsPlayer(ctx, playAvs, true); err != nil {
						log.Error("%+v", err)
						err = nil
					}
					return
				})
			} else {
				g.Go(func() (err error) {
					if apm, err = s.dao.Arcs(ctx, avids, mobiApp, device, mid); err != nil {
						log.Error("%+v", err)
						err = nil
					}
					return
				})
			}
		}
		if len(inlineAvids) != 0 {
			g.Go(func() (err error) {
				hasLike, err = s.dao.HasLike(ctx, buvid, mid, inlineAvids)
				if err != nil {
					log.Error("Failed to get has like state: %+v", err)
					return nil
				}
				return nil
			})
			g.Go(func() (err error) {
				hasFav, err = s.dao.IsFavVideos(ctx, mid, inlineAvids)
				if err != nil {
					log.Error("Failed to get is fav videos: %+v", err)
					return nil
				}
				return nil
			})
			g.Go(func() (err error) {
				hasCoin, err = s.dao.ArchiveUserCoins(ctx, inlineAvids, mid)
				if err != nil {
					log.Error("Failed to get archive user coins: %+v", err)
					return nil
				}
				return nil
			})
		}
		if len(roomIDs) != 0 {
			g.Go(func() (err error) {
				req := &livexroomgate.EntryRoomInfoReq{
					EntryFrom: []string{model.DefaultLiveEntry, model.SearchInlineCard, model.SearchLiveInlineCard},
					RoomIds:   roomIDs,
					Uid:       mid,
					Uipstr:    metadata.String(ctx, metadata.RemoteIP),
					Platform:  platform,
					Build:     int64(build),
					Network:   "other",
				}
				if entryRoom, err = s.dao.EntryRoomInfo(ctx, req); err != nil {
					log.Error("Failed to get entry room info: %+v: %+v", req, err)
					err = nil
					return
				}
				return
			})
		}
		if len(adOwners) != 0 {
			g.Go(func() (err error) {
				req := &livexroomgate.EntryRoomInfoReq{
					EntryFrom: []string{model.BrandADLiveEntry},
					Uids:      adOwners,
					Uid:       mid,
					Uipstr:    metadata.String(ctx, metadata.RemoteIP),
					Platform:  platform,
					Build:     int64(build),
					Network:   "other",
					ReqBiz:    "/x/v2/search",
				}
				if adOwnersEntryRoom, err = s.dao.EntryRoomInfo(ctx, req); err != nil {
					log.Error("Failed to get ad entry room info: %+v: %+v", req, err)
					err = nil
					return
				}
				return
			})
		}
		if len(seasonIDs) != 0 {
			g.Go(func() (err error) {
				if bangumis, err = s.dao.BangumiCard(ctx, mid, seasonIDs); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				return
			})
		}
		if len(sepReqs) != 0 {
			g.Go(func() (err error) {
				if seasonEps, medisas, err = s.dao.SearchPGCCards(ctx, sepReqs, keyword, mobiApp, device, platform, mid, fnver, fnval, qn, fourk, int64(build), true); err != nil {
					log.Error("bangumiDao SearchPGCCards %v", err)
					err = nil
				}
				return
			})
		}
		if len(inlineEPIDs) != 0 {
			g.Go(func() (err error) {
				inlineEPCards, err = s.dao.InlineCards(ctx, inlineEPIDs, mobiApp, platform, device, build, mid)
				if err != nil {
					log.Error("Failed to get inline ep card: %+v", err)
					return nil
				}
				inlineEPAids := []int64{}
				for _, ep := range inlineEPCards {
					inlineEPAids = append(inlineEPAids, ep.Aid)
				}
				inlineEPHasLike, err = s.dao.HasLike(ctx, buvid, mid, inlineEPAids)
				if err != nil {
					log.Error("Failed to get has inline ep like state: %+v", err)
					return nil
				}
				return nil
			})
		}
		if len(seasonStatIDs) != 0 {
			g.Go(func() (err error) {
				if seasonstat, err = s.dao.SeasonsStatGRPC(ctx, seasonStatIDs); err != nil {
					log.Error("bangumiDao SeasonsStatGRPC error(%v)", err)
					err = nil
				}
				return
			})
			g.Go(func() (err error) {
				if seasonInfoBangumi, err = s.dao.SeasonCards(ctx, seasonStatIDs); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				return
			})
		}
		if len(tags) != 0 {
			g.Go(func() (err error) {
				if tagMyInfos, err = s.dao.TagInfos(ctx, tags, mid); err != nil {
					log.Error("%v \n", err)
					err = nil
				}
				return
			})
		}
		if len(dynamicIDs) != 0 {
			g.Go(func() (err error) {
				if dynamicDetails, err = s.dao.DynamicDetails(ctx, dynamicIDs, "search"); err != nil {
					log.Errorc(ctx, "s.dao.DynamicDetails %+v", err)
					err = nil
				}
				return
			})
			g.Go(func() (err error) {
				if dynamicTopic, err = s.dao.DynamicTopics(ctx, dynamicIDs, platform, mobiApp, build); err != nil {
					log.Errorc(ctx, "s.dao.DynamicTopics %+v", err)
					err = nil
				}
				return
			})
		}
		if len(matchIDs) > 0 {
			g.Go(func() (err error) {
				if matchm, err = s.dao.Matchs(ctx, mid, matchIDs); err != nil {
					log.Error("%v", err)
					err = nil
					return
				}
				matchLiveRooms := []int64{}
				for _, v := range matchm {
					matchLiveRooms = append(matchLiveRooms, v.LiveRoom)
				}
				entryReq := &livexroomgate.EntryRoomInfoReq{
					EntryFrom: []string{model.DefaultLiveEntry, model.SearchEsCard},
					RoomIds:   matchLiveRooms,
					Uid:       mid,
					Uipstr:    metadata.String(c, metadata.RemoteIP),
					Platform:  platform,
					Build:     int64(build),
					Network:   "other",
				}
				if matchLiveEntryRoom, err = s.dao.EntryRoomInfo(ctx, entryReq); err != nil {
					log.Error("Failed to get entry room info: %+v: %+v", entryReq, err)
					err = nil
					return
				}
				return
			})
			g.Go(func() (err error) {
				req := &managersearch.GetEsportConfigsReq{
					EsportIds:  esportIds,
					Plat:       int32(plat),
					EsportType: 1, // 电竞
				}
				// ipad HD 的 plat 应该是 20，本应用暂时未修正
				if req.Plat == int32(model.PlatIpadHD) {
					req.Plat = 20
				}
				reply, err := s.dao.GetEsportConfigs(ctx, req)
				if err != nil {
					log.Error("Failed to get esport configs: %+v: %+v", req, err)
					return nil
				}
				esportConfigs = reply.Configs
				return nil
			})
		}
		if len(specialIDs) > 0 {
			g.Go(func() (err error) {
				specialCards = map[int64]*searchadm.SpreadConfig{}
				for _, id := range specialIDs {
					sc, ok := s.specialCache[id]
					if !ok {
						log.Warn("Failed to find special card: %d: %+v", id, s.specialCache)
						continue
					}
					specialCards[id] = sc
				}
				return nil
			})
		}
		if len(topGameIDs) > 0 {
			g.Go(func() (err error) {
				reply, err := s.dao.FetchTopGameConfigs(ctx, topGameIDs)
				if err != nil {
					log.Error("s.gameDao.FetchTopGameConfigs err=%+v", err)
					return nil
				}
				topGameConfigs = reply
				return nil
			})
			g.Go(func() (err error) {
				topGameData, err := s.dao.TopGame(ctx, mid, topGameIDs, makeSdkType(platform))
				if err != nil {
					log.Error("s.gameDao.TopGame err=%+v", err)
					return nil
				}
				topGameInlineConfigs, err = s.dao.FetchTopGameInlineConfigs(ctx, topGameCardIds)
				if err != nil {
					log.Error("s.dao.FetchTopGameInlineConfigs err=%+v", err)
					err = nil
				}
				topGameMaterials, err = s.makeTopGameMaterials(ctx, topGameData, topGameInlineConfigs, topGameCardIdMs)
				if err != nil {
					log.Error("s.makeTopGameMaterials err=%+v", err)
					return nil
				}
				return
			})
		}
		if len(sportsIds) > 0 {
			sportsMaterials = &search.SportsMaterial{}
			g.Go(func() (err error) {
				reply, err := s.dao.GetSportsEventMatches(ctx, &esportsservice.GetSportsEventMatchesReq{Ids: sportsIds, Mid: mid})
				if err != nil {
					log.Error("s.matchDao.GetSportsEventMatches err=%+v", err)
					return nil
				}
				sportsMaterials.SportsEventMatches = reply.Matches
				sportsMaterials.InlineFns, sportsMaterials.MatchVersusLiveEntryRoom, err = s.makeSportsInlineFns(ctx, sportsIds, reply.Matches)
				if err != nil {
					log.Error("s.makeTopGameMaterials err=%+v", err)
					return nil
				}
				return nil
			})
			if len(sportsSeasonIds) > 0 {
				g.Go(func() (err error) {
					req := &managersearch.GetEsportConfigsReq{
						EsportIds:  sportsSeasonIds,
						Plat:       int32(plat),
						EsportType: 2, // 体育
					}
					// ipad HD 的 plat 应该是 20，本应用暂时未修正
					if req.Plat == int32(model.PlatIpadHD) {
						req.Plat = 20
					}
					reply, err := s.dao.GetEsportConfigs(ctx, req)
					if err != nil {
						log.Error("Failed to get esport configs: %+v: %+v", req, err)
						return nil
					}
					sportsMaterials.Configs = reply.Configs
					return nil
				})
			}
		}
		if len(gameIDs) > 0 {
			g.Go(func() (err error) {
				multiGameInfos, err = s.dao.MultiGameInfos(ctx, mid, gameIDs, build, makeSdkType(platform))
				if err != nil {
					log.Error("Failed to get MultiGameInfos gameIDs=%+v, err=%+v", gameIDs, err)
				}
				return nil
			})
			g.Go(func() (err error) {
				cloudGameReply, err = s.dao.CloudGameEntry(ctx, &gameentry.MultiShowReq{Ids: gameIDs})
				if err != nil {
					log.Error("Failed to get CloudGameEntry gameIDs=%+v, err=%+v", gameIDs, err)
				}
				return nil
			})
		}
		if len(collectionCardIds) > 0 {
			collectionViews = make(map[int64]*ugcSeasonGrpc.ViewReply, len(collectionCardIds))
			mu := sync.Mutex{}
			for _, v := range collectionCardIds {
				collectionId := v
				g.Go(func() (err error) {
					reply, err := s.dao.SeasonView(ctx, &ugcSeasonGrpc.ViewRequest{SeasonID: collectionId})
					if err != nil {
						log.Error("s.ugcSeasonDao.SeasonViews, collectionId=%+v, err=%+v", collectionId, err)
						return nil
					}
					mu.Lock()
					collectionViews[collectionId] = reply
					mu.Unlock()
					return nil
				})
			}
		}
		if len(ogvChanIDs) > 0 {
			g.Go(func() (err error) {
				ogvChannelCards, err = s.dao.SearchChannelInHome(ctx, ogvChanIDs)
				if err != nil {
					log.Error("s.channelDao.SearchChannelInHome err=%+v", err)
					return nil
				}
				ogvChannelm = make(map[int64]*search.OgvChannelMaterial, len(chanIDs))
				for _, card := range ogvChannelCards.GetCards() {
					if card.GetCid() == 0 || card.GetBizId() == 0 {
						continue
					}
					reply, err := s.fetchOgvChannelMaterial(ctx, card.BizId, card.BizType)
					if err != nil {
						log.Error("s.fetchOgvChannelMaterial err=%+v", err)
						continue
					}
					ogvChannelm[card.GetCid()] = reply
				}
				return nil
			})
		}
		if len(comicIDs) > 0 {
			g.Go(func() error {
				var err error
				comicInfo, err = s.dao.GetComicInfos(ctx, comicIDs)
				if err != nil {
					log.Error("s.dao.GetComicInfos err:%+v", err)
					return nil
				}
				return nil
			})
		}
		if err = g.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		if all.SuggestKeyword != "" && pn == 1 {
			i := &search.Item{Title: all.SuggestKeyword, Goto: model.GotoSuggestKeyWord, SugKeyWordType: 1}
			item = append(item, i)
		} else if all.CrrQuery != "" && pn == 1 && !isNewOrder {
			if (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.QueryCorAndroid) ||
				(model.IsIPhone(plat) && build > s.c.SearchBuildLimit.QueryCorIOS) {
				i := &search.Item{Title: fmt.Sprintf("已匹配%q的搜索结果", all.CrrQuery), Goto: model.GotoSuggestKeyWord, SugKeyWordType: 2}
				item = append(item, i)
			}
		}
		for _, v := range all.FlowResult {
			i := &search.Item{TrackID: v.TrackID, LinkType: v.LinkType, Position: v.Position}
			switch v.Type {
			case search.TypeTips:
				// 缓存命中，缓存中获取物料
				cacheResult, ok := s.searchTipsCache[tipsID]
				if !ok {
					continue
				}
				i.FromTips(v.Tips, cacheResult)
			case search.TypeVideo:
				if (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.CooperationAndroid) ||
					(model.IsIPhone(plat) && build > s.c.SearchBuildLimit.CooperationIOS) ||
					(model.IsIPadHD(plat) && build > s.c.SearchBuildLimit.CooperationIPadHD) ||
					model.IsAndroidHD(plat) {
					cooperation = true
				}
				// 安卓 6.27 和 6.28 要修个 bug，填充一个空白角标
				defaultEmptyBizBadge := false
				if model.IsAndroid(plat) && (build >= 6270000 && build < 6290000) {
					defaultEmptyBizBadge = true
				}
				var ishot bool
				for hotaid := range s.hotAids {
					if hotaid == v.Video.ID {
						ishot = true
						break
					}
				}
				ugcInlineParams := &search.OptUGCInlineFnParams{
					SearchMeta: v.Video,
					Archive:    apm[v.Video.ID],
					UserInfo:   accCards[v.Video.Mid],
					Follow:     follows,
					HasLike:    hasLike,
					HotAids:    castAsHotAidSet(s.hotAids),
					HasFav:     hasFav,
					HasCoin:    hasCoin,
					NftRegion:  nftRegion,
				}
				optUGCInline := search.OptUGCInlineFn(c, ugcInlineParams, model.GotoUGCInline)
				i.FromVideo(v.Video, apm[v.Video.ID], cooperation, isNewDuration, isNewOGVURL, isNewColor, ishot, defaultEmptyBizBadge, order, optUGCInline)
				// UGC三点
				if s.c.Switch.SearchThreePoint {
					i.ThreePoint = append(i.ThreePoint, &search.ThreePoint{Type: "wait", Icon: s.c.Resource.SearchThreePoint.WaitIcon, Title: s.c.Resource.SearchThreePoint.WaitTitle})
					i.ThreePoint = append(i.ThreePoint, &search.ThreePoint{Type: "share", Icon: s.c.Resource.SearchThreePoint.ShareIcon, Title: s.c.Resource.SearchThreePoint.ShareTitle})
				}
			case search.TypeLive:
				if cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
					optLiveRoom := search.OptLiveRoomInlineFn(c, entryRoom[v.Live.RoomID], accCards[v.Live.UID], follows, v.Live, model.GotoLiveInline, model.SearchLiveInlineCard, nftRegion)
					i.FromLive(v.Live, entryRoom[v.Live.RoomID], optLiveRoom)
				}
			case search.TypeMediaBangumi:
				//i.FromMediaPgcCard(v.Media, "", model.GotoBangumi, bangumis, seasonEps, medisas, s.c.Cfg.PgcSearchCard, false) // flow result, not ipad
				var extFunc []func(*search.Item)
				if isOgvExpNewUser {
					extFunc = append(extFunc, search.WithOgvNewUserUpdateBadges(ctx, v.Media, seasonEps))
				}
				extFunc = append(extFunc, search.OptOGVInlineFn(ctx, inlineEPCards[v.Media.EPID], inlineEPHasLike, follows, v.Media))
				if err := i.FromMediaPgcCardPureRPC(v.Media, "", model.GotoBangumi, seasonEps, s.c.Cfg.PgcSearchCard, false, extFunc...); err != nil { // flow result, not ipad
					log.Error("Failed to build pgc card by pgc RPC: %+v: %+v", v.Media, err)
					continue
				}
			case search.TypeMediaFt:
				//i.FromMediaPgcCard(v.Media, "", model.GotoMovie, bangumis, seasonEps, medisas, s.c.Cfg.PgcSearchCard, false)
				var extFunc []func(*search.Item)
				if isOgvExpNewUser {
					extFunc = append(extFunc, search.WithOgvNewUserUpdateBadges(ctx, v.Media, seasonEps))
				}
				extFunc = append(extFunc, search.OptOGVInlineFn(ctx, inlineEPCards[v.Media.EPID], inlineEPHasLike, follows, v.Media))
				if err := i.FromMediaPgcCardPureRPC(v.Media, "", model.GotoMovie, seasonEps, s.c.Cfg.PgcSearchCard, false, extFunc...); err != nil {
					log.Error("Failed to build pgc card by pgc RPC: %+v: %+v", v.Media, err)
					continue
				}
			case search.TypeArticle:
				i.FromArticle(v.Article, accProfiles[v.Article.Mid])
			case search.TypeSpecial:
				// i.FromOperate(v.Operate, model.GotoSpecial, false)
				if err := i.FromCardSpecial(v.Operate, specialCards[v.Operate.ID], model.GotoSpecial, false); err != nil {
					log.Warn("Failed to build from card special: %+v", err)
					continue
				}
			case search.TypeBanner:
				i.FromOperate(v.Operate, model.GotoBanner, false)
			case search.TypeUser:
				if follows[v.User.Mid] {
					i.Attentions = 1
				}
				var extFunc []func(*search.Item)
				extFunc = append(extFunc, search.WithUserCardGetNftRegion(nftRegion))
				i.Relation = cardmdl.RelationChange(v.User.Mid, relationm)
				version := upCardVersion(v.User.Version, build, plat, s.c.SearchBuildLimit)
				switch version {
				case 1:
					i.FromUpUserNew(v.User, accCards[v.User.Mid], apm, entryRoom[v.User.RoomID], isBlue, isNewDuration, s.c.Search, nil, s.systemNotice[v.User.Mid], accProfiles[v.User.Mid], extFunc...)
				default:
					i.FromUserVip(v.User, apm, entryRoom[v.User.RoomID], accCards[v.User.Mid], isBlue)
				}
			case search.TypeBiliUser:
				if follows[v.User.Mid] {
					i.Attentions = 1
				}
				var extFunc []func(*search.Item)
				extFunc = append(extFunc, search.WithUserCardGetNftRegion(nftRegion))
				// 只有 bili_user 需要 inline 直播
				optInlineLive := search.OptInlineLiveFn(c, entryRoom[v.User.RoomID], accCards[v.User.Mid], follows)
				i.Relation = cardmdl.RelationChange(v.User.Mid, relationm)
				version := upCardVersion(v.User.Version, build, plat, s.c.SearchBuildLimit)
				switch version {
				case 1:
					i.FromUpUserNew(v.User, accCards[v.User.Mid], apm, entryRoom[v.User.RoomID], isBlue, isNewDuration, s.c.Search, optInlineLive, s.systemNotice[v.User.Mid], accProfiles[v.User.Mid], extFunc...)
				default:
					i.FromUpUserVip(v.User, apm, entryRoom[v.User.RoomID], accCards[v.User.Mid], isBlue, isNewDuration)
				}
				// UGC三点
				if s.c.Switch.SearchThreePoint {
					if (model.IsAndroid(plat) && build >= 6410000) || (model.IsIOS(plat) && build >= 64100000) {
						i.ThreePoint = append(i.ThreePoint, &search.ThreePoint{Type: "share", Icon: s.c.Resource.SearchThreePoint.ShareIcon, Title: s.c.Resource.SearchThreePoint.ShareTitle})
					}
					i.SharePlane = search.ConstructUserSharePlane(accCards[v.User.Mid])
				}
			case search.TypeSpecialS:
				// i.FromOperate(v.Operate, model.GotoSpecialS, isNewColor)
				if err := i.FromCardSpecial(v.Operate, specialCards[v.Operate.ID], model.GotoSpecialS, isNewColor); err != nil {
					log.Warn("Failed to build from card special: %+v", err)
					continue
				}
			case search.TypeGame:
				if ok := i.FromGameBasedOnMultiGameInfos(v.Game.ID, multiGameInfos); !ok {
					log.Warn("Failed to make FromGameBasedOnMultiGameInfos v.Game=%+v multiGameInfos=%+v", v.Game, multiGameInfos)
					continue
				}
				if cloudGameReply != nil && cloudGameReply.Data != nil {
					if showEntry, ok := cloudGameReply.Data[v.Game.ID]; ok && showEntry == 1 {
						i.FromCloudGameConfigs()
					}
				}
			case search.TypeQuery:
				i.Title = v.TypeName
				i.FromQuery(v.Query)
			case search.TypeComic:
				i.FromComic(c, v.Comic)
			case search.TypeConverge:
				var (
					aids, rids, artids []int64
					am                 map[int64]*arcgrpc.Arc
					rm                 map[int64]*search.Room
					artm               map[int64]*article.Meta
				)
				for _, c := range v.Operate.ContentList {
					//nolint:gomnd
					switch c.Type {
					case 0:
						aids = append(aids, c.ID)
					case 1:
						rids = append(rids, c.ID)
					case 2:
						artids = append(artids, c.ID)
					}
				}
				g, ctx := errgroup.WithContext(c)
				if len(aids) != 0 {
					g.Go(func() (err error) {
						if am, err = s.dao.Archives(ctx, aids, mobiApp, device, mid); err != nil {
							log.Error("%+v", err)
							err = nil
						}
						return
					})
				}
				if len(rids) != 0 && cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
					g.Go(func() (err error) {
						if rm, err = s.dao.AppMRoom(ctx, rids, platform); err != nil {
							log.Error("%+v", err)
							err = nil
						}
						return
					})
				}
				if len(artids) != 0 {
					g.Go(func() (err error) {
						if artm, err = s.dao.Articles(ctx, artids); err != nil {
							log.Error("%+v", err)
							err = nil
						}
						return
					})
				}
				if err = g.Wait(); err != nil {
					log.Error("%+v", err)
					continue
				}
				i.FromConverge(v.Operate, am, rm, artm)
			case search.TypeTwitter:
				i.FromTwitter(v.Twitter, dynamicDetails, dynamicTopic, s.c.SearchDynamicSwitch.IsUP, s.c.SearchDynamicSwitch.IsCount, isNewTwitter)
			case search.TypeStar:
				if v.Star.TagID != 0 {
					i.URIType = search.StarChannel
					for _, myInfo := range tagMyInfos {
						if myInfo != nil && myInfo.TagID == v.Star.TagID {
							i.IsAttention = myInfo.IsAtten
							break
						}
					}
				} else if v.Star.MID != 0 {
					i.URIType = search.StarSpace
					if follows[v.Star.MID] {
						i.IsAttention = 1
					}
					i.Relation = cardmdl.RelationChange(v.Star.MID, relationm)
				}
				i.FromStar(v.Star, order)
			case search.TypeTicket:
				i.FromTicket(v.Ticket)
			case search.TypeProduct:
				i.FromProduct(v.Product)
			case search.TypeSpecialerGuide:
				i.FromSpecialerGuide(v.SpecialerGuide)
			case search.TypeChannel:
				i.FromChannel(v.Channel, apm, tagMyInfos, "all_search", order)
			case search.TypeOGVCard:
				var (
					tmps     []*search.Item
					isBroken bool // 相关推荐模块0或者1为true
				)
				res.OGVCard, tmps, isBroken = i.FromOGVCard(v.SearchOGVCard, seasonstat, seasonInfoBangumi, plat, comicInfo, keyword)
				if isBroken {
					res.OGVCard = nil
					continue
				}
				if isOgvExpNewUser {
					for _, tmp := range tmps {
						if badgeStyle, ok := search.UpdateBadgeStyleForOgvPgcCard(c, v.SearchOGVCard, seasonInfoBangumi); ok {
							tmp.BadgeStyle = badgeStyle
						}
					}
				}
				item = append(item, tmps...)
			case search.TypeESports:
				extFunc := []func(*search.Item){}
				if enableWithESportSearchConfig(plat, int64(build)) {
					extFunc = append(extFunc, search.WithESportConfig(v.ESport.ID, "全部赛程", v.ESport.UrlBottom, esportConfigs, plat))
				}
				if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
					pd.IsPlatAndroid().And().Build(">=", 6890000)
				}).OrWhere(func(pd *pd.PDContext) {
					pd.IsPlatIPhone().And().Build(">=", 68900000)
				}).MustFinish() {
					if esRoomMid, ok := checkESportSearchInline(esportConfigs, v.ESport, matchm, matchLiveEntryRoom); ok {
						var (
							esAccCards map[int64]*account.Card
							esFollows  map[int64]bool
						)
						esGroup := errgroupv2.WithContext(c)
						esGroup.Go(func(ctx context.Context) error {
							var err error
							if esAccCards, err = s.dao.Cards3(ctx, []int64{esRoomMid}); err != nil {
								log.Error("s.dao.Cards3 ESport esRoomMid:%d, err:%+v", esRoomMid, err)
								return err
							}
							return nil
						})
						if mid > 0 {
							esGroup.Go(func(ctx context.Context) error {
								esFollows = s.dao.Relations3(ctx, []int64{esRoomMid}, mid)
								return nil
							})
						}
						if err := esGroup.Wait(); err != nil {
							log.Error("s.search ESport inline error:%+v", err)
						}
						if esAccCards != nil {
							extFunc = append(extFunc, search.OptEsportInlineFn(c, v.ESport.ID, esportConfigs, matchLiveEntryRoom, esAccCards[esRoomMid], esFollows, model.GotoEsportsInline, model.SearchEsCard))
						}
					}
				}
				i.FormESport(v.ESport, localTime, matchm, matchLiveEntryRoom, extFunc...)
				if !i.Right {
					continue
				}
			case search.TypeNewChannel:
				i.FormNewChannel(ctx, v.NewChannel, newChannelm, apm)
				if !i.Right {
					continue
				}
			case search.TypeOgvChannel:
				if ok := i.FromOgvChannel(v.NewChannel, ogvChannelm); !ok {
					continue
				}
			case search.TypeBrandAD, search.TypeBrandAdGiant, search.TypeBrandAdGiantTriple:
				if err := i.FromBrandAD(v.BrandAD, apm, accCards, adOwnersEntryRoom, relationm, v.Type); err != nil {
					log.Warn("Failed to build from brand ad: %+v: %+v", v.BrandAD, err)
					continue
				}
			case search.TypeVideoAd:
				if err := i.FromBrandVideoAd(v.BrandAD, apm, accCards, adOwnersEntryRoom, relationm, v.Type); err != nil {
					log.Warn("Failed to build from video ad: %+v: %+v", v.BrandAD, err)
					continue
				}
			case search.TypePictureAd:
				if err := i.FromBrandSimpleAD(v.BrandAD, accCards, adOwnersEntryRoom, relationm, v.Type); err != nil {
					log.Warn("Failed to build from brand simple ad: %+v: %+v", v.BrandAD, err)
					continue
				}
			case search.TypeGameAD:
				if err := i.FromBrandGameAD(v.GameAD); err != nil {
					log.Warn("Failed to build from brand game ad: %+v: %+v", v.GameAD, err)
					continue
				}
			case search.TypePediaCard:
				if err := i.FromPediaCard(v.PediaCard, model.GotoPediaCard); err != nil {
					log.Warn("Failed to build from pedia card: %+v: %+v", v.PediaCard, err)
					continue
				}
			case search.TypeTopGame:
				if meta, ok := topGameMaterials[v.TopGame.ID]; ok {
					var extFunc []func(*search.Item)
					if meta.InlineFn != nil {
						extFunc = append(extFunc, meta.InlineFn)
					}
					if f, ok := OptTopGameTabInfoFn(v.TopGame.ID, v.TopGame.CardId, topGameConfigs); ok {
						extFunc = append(extFunc, f)
					}
					i.FromTopGame(meta.TopGameData, extFunc...)
				}
			case search.TypeBrandAdAv:
				var extFunc []func(*search.Item)
				if inlineFn, ok := buildBrandAdAvInlineProcess(ctx, v.BrandADInline.GetADContent(), apm, accCards, follows, hasLike, hasFav, hasCoin, nftRegion); ok {
					extFunc = append(extFunc, inlineFn)
				}
				if err := i.FromBrandADAv(v.BrandADInline, accCards, adOwnersEntryRoom, relationm, extFunc...); err != nil {
					log.Warn("Failed to build from brand ad av inline: %+v: %+v", v.BrandADInline, err)
					continue
				}
			case search.TypeBrandAdLocalAv:
				if err := i.FromBrandADLocalAv(v.BrandADInline, accCards, adOwnersEntryRoom, relationm); err != nil {
					log.Warn("Failed to build from brand ad local av inline: %+v: %+v", v.BrandADInline, err)
					continue
				}
			case search.TypeBrandAdLive:
				var extFunc []func(*search.Item)
				if inlineFn, ok := buildBrandAdLiveInlineProcess(ctx, v.BrandADInline.GetADContent(), accCards, adOwnersEntryRoom, follows, nftRegion); ok {
					extFunc = append(extFunc, inlineFn)
				}
				if err := i.FromBrandADLive(v.BrandADInline, adOwnersEntryRoom, accCards, relationm, extFunc...); err != nil {
					log.Warn("Failed to build from brand ad live inline: %+v: %+v", v.BrandADInline, err)
					continue
				}
			case search.TypeSportsVersus:
				if sportsMaterials == nil {
					continue
				}
				var extFunc []func(*search.Item)
				if enableWithESportSearchConfig(plat, int64(build)) {
					extFunc = append(extFunc, search.WithESportConfig(v.Sports.SeasonId, "热门赛程", v.Sports.Url, sportsMaterials.Configs, plat))
				}
				if match, ok := sportsMaterials.SportsEventMatches[v.Sports.ID]; ok {
					if err := i.FromSportsVersus(v.Sports, match, localTime, sportsMaterials.MatchVersusLiveEntryRoom, extFunc...); err != nil {
						log.Warn("Failed to build from sports versus, sports=%+v error=%+v", v.Sports, err)
						continue
					}
				}
			case search.TypeSports:
				if sportsMaterials == nil {
					continue
				}
				var extFunc []func(*search.Item)
				if inlineFn, ok := sportsMaterials.InlineFns[v.Sports.ID]; ok {
					extFunc = append(extFunc, inlineFn)
				}
				if enableWithESportSearchConfig(plat, int64(build)) {
					extFunc = append(extFunc, search.WithESportConfig(v.Sports.SeasonId, "热门赛程", v.Sports.Url, sportsMaterials.Configs, plat))
				}
				if match, ok := sportsMaterials.SportsEventMatches[v.Sports.ID]; ok {
					if err := i.FromSports(v.Sports, match, localTime, extFunc...); err != nil {
						log.Warn("Failed to build from sports card, sports=%+v error=%+v", v.Sports, err)
						continue
					}
				}
			case search.TypeCollectionCard:
				if collectionViews == nil || v.CollectionCard == nil {
					continue
				}
				if card, ok := collectionViews[v.CollectionCard.ID]; ok && card.View != nil {
					if err := i.FromCollectionCard(v.CollectionCard, card.View); err != nil {
						log.Warn("Failed to build from collection card: %+v: %+v", v.CollectionCard, err)
						continue
					}
				}
			case search.TypePediaInlineCard:
				var extFunc []func(*search.Item)
				ugcInlineParams := &search.OptUGCInlineFnParams{
					Archive:   apm[v.PediaCard.NavigationCard.Avid],
					Follow:    follows,
					HasLike:   hasLike,
					HotAids:   castAsHotAidSet(s.hotAids),
					HasFav:    hasFav,
					HasCoin:   hasCoin,
					NftRegion: nftRegion,
				}
				extFunc = append(extFunc, search.OptUGCInlineFn(c, ugcInlineParams, model.GotoPediaInlineCard))
				if err := i.FromPediaCard(v.PediaCard, search.TypePediaInlineCard, extFunc...); err != nil {
					log.Warn("Failed to build from pedia inline card: %+v: %+v", v.PediaCard, err)
					continue
				}
			default:
			}
			if i.Goto != "" && v.Type != search.TypeOGVCard {
				stat.MetricSearchAppCardTotal.Inc(i.LinkType, i.Goto)
				item = append(item, i)
			}
		}
		res.Item = item
		if plat == model.PlatAndroid && build < search.SearchEggInfoAndroid {
			return
		}
		if all.EggInfo != nil && teenagersMode == 0 {
			res.EasterEgg = &search.EasterEgg{ID: all.EggInfo.ID, ShowCount: all.EggInfo.ShowCount, EggType: all.EggInfo.EggType,
				CloseCount: s.c.Search.EggCloseCount, MaskTransparency: all.EggInfo.MaskTransparency, MaskColor: all.EggInfo.MaskColor}
			var gotoType string
			//nolint:gomnd
			switch all.EggInfo.ReType {
			case 1:
				gotoType = model.GotoWeb
			case 2:
				gotoType = model.GotoAv
			case 3:
				gotoType = model.GotoPGC
			case 4:
				gotoType = model.GotoArticle
			case 5:
				gotoType = model.GotoDynamic
			case 6:
				gotoType = model.GotoLive
			case 7:
				gotoType = model.GotoWeb
			}
			if all.EggInfo.ReValue != "" && all.EggInfo.ReValue != "0" {
				res.EasterEgg.URL = model.FillURI(gotoType, all.EggInfo.ReValue, nil)
			}
			switch all.EggInfo.EggType {
			case search.EggTypeVideo:
				res.EasterEgg.SourceURL = all.EggInfo.URL
				res.EasterEgg.SourceMd5 = all.EggInfo.Md5
				res.EasterEgg.SourceSize = all.EggInfo.Size
			case search.EggTypeURL:
				res.EasterEgg.URL = all.EggInfo.ReURL
			case search.EggTypePIC:
				res.EasterEgg.PicType = all.EggInfo.PicType
				res.EasterEgg.ShowTime = all.EggInfo.PicShowTime
				res.EasterEgg.SourceURL = all.EggInfo.URL
				res.EasterEgg.SourceMd5 = all.EggInfo.Md5
				res.EasterEgg.SourceSize = all.EggInfo.Size
			}
		}
		// mid int64过滤Items
		res.Item = filterMidInt64OnItem(c, res.Item)
		return
	}
	var items []*search.Item
	if all.SuggestKeyword != "" && pn == 1 {
		res.Items.SuggestKeyWord = &search.Item{Title: all.SuggestKeyword, Goto: model.GotoSuggestKeyWord}
	}
	// archive
	for _, v := range all.Result.Video {
		switch v.Type {
		case "special_card":
			continue
		default:
			playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: v.ID})
			avids = append(avids, v.ID)
		}
	}
	if duration == "0" && order == "totalrank" && rid == 0 {
		for _, v := range all.Result.Movie {
			if v.Type == "movie" {
				playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: v.Aid})
				avids = append(avids, v.Aid)
			}
		}
		for _, v := range all.Result.MediaBangumi {
			seasonIDs = append(seasonIDs, v.SeasonID)
			if v.Canplay() {
				sepReqs = append(sepReqs, v.BuildPgcReq())
			}
		}
		for _, v := range all.Result.MediaFt {
			seasonIDs = append(seasonIDs, v.SeasonID)
			if v.Canplay() {
				sepReqs = append(sepReqs, v.BuildPgcReq())
			}
		}
	}
	for _, v := range all.Result.ESports {
		for _, match := range v.MatchList {
			if match != nil {
				if match.ID == 0 {
					continue
				}
				matchIDs = append(matchIDs, match.ID)
			}
		}
		esportIds = append(esportIds, v.ID)
	}
	if pn == 1 {
		for _, v := range all.Result.User {
			owners = append(owners, v.Mid)
			for _, vr := range v.Res {
				playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: vr.Aid})
				avids = append(avids, vr.Aid)
			}
		}
		if old {
			for _, v := range all.Result.UpUser {
				for _, vr := range v.Res {
					playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: vr.Aid})
					avids = append(avids, vr.Aid)
				}
				owners = append(owners, v.Mid)
				if cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
					roomIDs = append(roomIDs, v.RoomID)
				}
			}
		} else {
			for _, v := range all.Result.BiliUser {
				for _, vr := range v.Res {
					playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: vr.Aid})
					avids = append(avids, vr.Aid)
				}
				owners = append(owners, v.Mid)
				if cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
					roomIDs = append(roomIDs, v.RoomID)
				}
			}
		}
	}
	if model.IsOverseas(plat) {
		for _, v := range all.Result.LiveRoom {
			if !cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
				break
			}
			roomIDs = append(roomIDs, v.RoomID)
		}
		for _, v := range all.Result.LiveUser {
			if !cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
				break
			}
			roomIDs = append(roomIDs, v.RoomID)
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(owners) != 0 {
		g.Go(func() (err error) {
			if accCards, err = s.dao.Cards3(ctx, owners); err != nil {
				log.Error("accDao.Cards Owners %+v, Err %+v", owners, err)
				err = nil
			}
			return
		})
		g.Go(func() (err error) {
			if accProfiles, err = s.dao.ProfilesWithoutPrivacy3(ctx, owners); err != nil {
				log.Error("s.accDao.ProfilesWithoutPrivacy3 owners=%+v, err=%+v", owners, err)
				err = nil
			}
			return
		})
		g.Go(func() (err error) {
			if nftRegion, err = s.getNFTIconInfo(ctx, owners); err != nil {
				log.Error("getNFTIconInfo owners=%+v, err=%+v", owners, err)
				err = nil
			}
			return
		})
		if mid > 0 {
			g.Go(func() error {
				follows = s.dao.Relations3(ctx, owners, mid)
				return nil
			})
			g.Go(func() (err error) {
				relationm, err = s.dao.Interrelations(ctx, mid, owners)
				if err != nil {
					log.Error("%+v", err)
					err = nil
					return
				}
				return
			})
		}
	}
	if len(avids) != 0 {
		if newPlayer {
			g.Go(func() (err error) {
				if apm, err = s.dao.ArcsPlayer(ctx, playAvs, false); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				return
			})
		} else {
			g.Go(func() (err error) {
				if apm, err = s.dao.Arcs(ctx, avids, mobiApp, device, mid); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				return
			})
		}
	}
	if len(matchIDs) != 0 {
		g.Go(func() (err error) {
			if matchm, err = s.dao.Matchs(ctx, mid, matchIDs); err != nil {
				log.Error("%v", err)
				err = nil
				return
			}
			matchLiveRooms := []int64{}
			for _, v := range matchm {
				matchLiveRooms = append(matchLiveRooms, v.LiveRoom)
			}
			entryReq := &livexroomgate.EntryRoomInfoReq{
				EntryFrom: []string{model.DefaultLiveEntry},
				RoomIds:   matchLiveRooms,
				Uid:       mid,
				Uipstr:    metadata.String(c, metadata.RemoteIP),
				Platform:  platform,
				Build:     int64(build),
				Network:   "other",
			}
			if matchLiveEntryRoom, err = s.dao.EntryRoomInfo(ctx, entryReq); err != nil {
				log.Error("Failed to get entry room info: %+v: %+v", entryReq, err)
				err = nil
				return
			}
			return
		})
		g.Go(func() (err error) {
			req := &managersearch.GetEsportConfigsReq{
				EsportIds:  esportIds,
				Plat:       int32(plat),
				EsportType: 1, // 电竞
			}
			// ipad HD 的 plat 应该是 20，本应用暂时未修正
			if req.Plat == int32(model.PlatIpadHD) {
				req.Plat = 20
			}
			reply, err := s.dao.GetEsportConfigs(ctx, req)
			if err != nil {
				log.Error("Failed to get esport configs: %+v: %+v", req, err)
				return nil
			}
			esportConfigs = reply.Configs
			return nil
		})
	}
	if len(roomIDs) != 0 {
		g.Go(func() (err error) {
			req := &livexroomgate.EntryRoomInfoReq{
				EntryFrom: []string{model.DefaultLiveEntry},
				RoomIds:   roomIDs,
				Uid:       mid,
				Uipstr:    metadata.String(ctx, metadata.RemoteIP),
				Platform:  platform,
				Build:     int64(build),
				Network:   "other",
			}
			entryRoom, err = s.dao.EntryRoomInfo(ctx, req)
			if err != nil {
				log.Error("Failed to get entry room info: %+v: %+v", req, err)
				err = nil
				return
			}
			return
		})
	}
	if len(seasonIDs) != 0 {
		g.Go(func() (err error) {
			if bangumis, err = s.dao.BangumiCard(ctx, mid, seasonIDs); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(sepReqs) != 0 {
		g.Go(func() (err error) {
			if seasonEps, medisas, err = s.dao.SearchPGCCards(ctx, sepReqs, keyword, mobiApp, device, platform, mid, fnver, fnval, qn, fourk, int64(build), true); err != nil {
				log.Error("bangumiDao SearchPGCCards %v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	if duration == "0" && order == "totalrank" && rid == 0 {
		var promptBangumi, promptFt string
		// season
		bangumi := all.Result.Bangumi
		items = make([]*search.Item, 0, len(bangumi))
		for _, v := range bangumi {
			si := &search.Item{}
			if (model.IsAndroid(plat) && build <= _oldAndroid) || (model.IsIPhone(plat) && build <= _oldIOS) {
				si.FromSeason(v, model.GotoBangumi)
			} else {
				si.FromSeason(v, model.GotoBangumiWeb)
			}
			items = append(items, si)
		}
		res.Items.Season = items
		// movie
		movie := all.Result.Movie
		items = make([]*search.Item, 0, len(movie))
		for _, v := range movie {
			si := &search.Item{}
			si.FromMovie(v, apm)
			items = append(items, si)
		}
		res.Items.Movie = items
		// season2
		mb := all.Result.MediaBangumi
		items = make([]*search.Item, 0, len(mb))
		for k, v := range mb {
			si := &search.Item{}
			if model.IsAndroidHD(plat) || ((plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD)) && (k == len(mb)-1) && all.PageInfo.MediaBangumi.NumResults > s.iPadSearchBangumi {
				promptBangumi = fmt.Sprintf("查看全部番剧 ( %d ) >", all.PageInfo.MediaBangumi.NumResults)
			}
			var extFunc []func(*search.Item)
			if isOgvExpNewUser {
				extFunc = append(extFunc, search.WithOgvNewUserUpdateBadges(ctx, v, seasonEps))
			}
			si.FromMediaPgcCard(v, promptBangumi, model.GotoBangumi, bangumis, seasonEps, medisas, s.c.Cfg.PgcSearchCard, false, extFunc...) // non-flow result, not direct
			// si.FromMedia(v, promptBangumi, model.GotoBangumi, bangumis)
			items = append(items, si)
		}
		res.Items.Season2 = items
		// movie2
		mf := all.Result.MediaFt
		items = make([]*search.Item, 0, len(mf))
		for k, v := range mf {
			si := &search.Item{}
			if model.IsAndroidHD(plat) || ((plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD)) && (k == len(mf)-1) && all.PageInfo.MediaFt.NumResults > s.iPadSearchFt {
				promptFt = fmt.Sprintf("查看全部影视 ( %d ) >", all.PageInfo.MediaFt.NumResults)
			}
			// si.FromMedia(v, promptFt, model.GotoMovie, bangumis)
			var extFunc []func(*search.Item)
			if isOgvExpNewUser {
				extFunc = append(extFunc, search.WithOgvNewUserUpdateBadges(ctx, v, seasonEps))
			}
			si.FromMediaPgcCard(v, promptFt, model.GotoMovie, bangumis, seasonEps, medisas, s.c.Cfg.PgcSearchCard, false, extFunc...)
			si.Goto = model.GotoAv
			items = append(items, si)
		}
		res.Items.Movie2 = items
	}
	if pn == 1 {
		// upper + user
		var tmp []*search.User
		if old {
			tmp = all.Result.UpUser
		} else {
			tmp = all.Result.BiliUser
		}
		items = make([]*search.Item, 0, len(tmp)+len(all.Result.User))
		for _, v := range all.Result.User {
			si := &search.Item{}
			var extFunc []func(*search.Item)
			extFunc = append(extFunc, search.WithUserCardGetNftRegion(nftRegion))
			version := upCardVersion(v.Version, build, plat, s.c.SearchBuildLimit)
			//nolint:gomnd
			switch version {
			case 1:
				// 这里应该不需要 inline 直播
				si.FromUpUserNew(v, accCards[v.Mid], apm, entryRoom[v.RoomID], isBlue, isNewDuration, s.c.Search, nil, s.systemNotice[v.Mid], accProfiles[v.Mid], extFunc...)
			case 2:
				si.FromUpUserNewIPadHD(v, accCards[v.Mid], apm, entryRoom[v.RoomID], isBlue, isNewDuration, s.c.Search, accProfiles[v.Mid], extFunc...)
				si.Relation = cardmdl.RelationChange(v.Mid, relationm) // 目前只有 ipadhd 需要 relation 字段
			default:
				si.FromUser(v, accCards[v.Mid], apm, entryRoom[v.RoomID], isBlue)
			}
			if follows[v.Mid] {
				si.Attentions = 1
			}
			items = append(items, si)
		}
		if len(items) == 0 {
			for _, v := range tmp {
				si := &search.Item{}
				var extFunc []func(*search.Item)
				extFunc = append(extFunc, search.WithUserCardGetNftRegion(nftRegion))
				version := upCardVersion(v.Version, build, plat, s.c.SearchBuildLimit)
				//nolint:gomnd
				switch version {
				case 1:
					// 这里应该不需要 inline 直播
					si.FromUpUserNew(v, accCards[v.Mid], apm, entryRoom[v.RoomID], isBlue, isNewDuration, s.c.Search, nil, s.systemNotice[v.Mid], accProfiles[v.Mid], extFunc...)
				case 2:
					si.FromUpUserNewIPadHD(v, accCards[v.Mid], apm, entryRoom[v.RoomID], isBlue, isNewDuration, s.c.Search, accProfiles[v.Mid], extFunc...)
					si.Relation = cardmdl.RelationChange(v.Mid, relationm) // 目前只有 ipadhd 需要 relation 字段
				default:
					si.FromUpUser(v, accCards[v.Mid], apm, entryRoom[v.RoomID], isBlue, isNewDuration, s.systemNotice)
				}
				if follows[v.Mid] {
					si.Attentions = 1
				}
				if old {
					si.IsUp = true
				}
				items = append(items, si)
			}
		}
		res.Items.Upper = items
	}
	items = make([]*search.Item, 0, len(all.Result.Video))
	for _, v := range all.Result.Video {
		switch v.Type {
		case "special_card":
			si := &search.Item{}
			if err := si.FromVideoSpecial(v); err != nil {
				log.Warn("Failed to build from card special: %+v", err)
				continue
			}
			items = append(items, si)
		default:
			isNewColor := false
			if (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.CardOptimizeAndroid) ||
				(model.IsIPhone(plat) && build > s.c.SearchBuildLimit.CardOptimizeIPhone) ||
				(model.PlatIpadHD == plat && build > s.c.SearchBuildLimit.CardOptimizeIpadHD) ||
				model.IsAndroidHD(plat) {
				isNewColor = true
			}
			si := &search.Item{}
			si.FromVideo(v, apm[v.ID], cooperation, isNewDuration, isNewOGVURL, isNewColor, false, false, order, nil)
			items = append(items, si)
		}
	}
	res.Items.Archive = items
	// live room
	if model.IsOverseas(plat) {
		if cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
			items = make([]*search.Item, 0, len(all.Result.LiveRoom))
			for _, v := range all.Result.LiveRoom {
				si := &search.Item{}
				si.FromLive(v, entryRoom[v.RoomID], nil)
				items = append(items, si)
			}
			res.Items.LiveRoom = items
			// live user
			items = make([]*search.Item, 0, len(all.Result.LiveUser))
			for _, v := range all.Result.LiveUser {
				si := &search.Item{}
				si.FromLive(v, entryRoom[v.RoomID], nil)
				items = append(items, si)
			}
			res.Items.LiveUser = items
		}
	}
	// esport
	items = make([]*search.Item, 0, len(all.Result.ESports))
	for _, v := range all.Result.ESports {
		extFunc := []func(*search.Item){}
		if enableWithESportSearchConfig(plat, int64(build)) {
			extFunc = append(extFunc, search.WithESportConfig(v.ID, "全部赛程", v.UrlBottom, esportConfigs, plat))
		}
		si := &search.Item{}
		si.FormESport(v, localTime, matchm, matchLiveEntryRoom, extFunc...)
		if !si.Right {
			continue
		}
		items = append(items, si)
	}
	res.Items.ESport = items
	// mid int64过滤Items
	res.Items = filterMidInt64OnItems(c, res.Items)
	return
}

// nolint:gomnd
func makeSdkType(platform string) int {
	switch platform {
	case "android":
		return 1
	case "ios":
		return 2
	default:
	}
	return 0
}

func filterMidInt64OnItems(ctx context.Context, items search.ResultItems) search.ResultItems {
	if !midint64.IsDisableInt64MidVersion(ctx) {
		return items
	}
	items.Upper = filterMidInt64OnItem(ctx, items.Upper)
	items.Archive = filterMidInt64OnItem(ctx, items.Archive)
	items.ESport = filterMidInt64OnItem(ctx, items.ESport)
	items.Operation = filterMidInt64OnItem(ctx, items.Operation)
	items.Season2 = filterMidInt64OnItem(ctx, items.Season2)
	items.Season = filterMidInt64OnItem(ctx, items.Season)
	items.Movie2 = filterMidInt64OnItem(ctx, items.Movie2)
	items.Movie = filterMidInt64OnItem(ctx, items.Movie)
	items.LiveRoom = filterMidInt64OnItem(ctx, items.LiveRoom)
	items.LiveUser = filterMidInt64OnItem(ctx, items.LiveUser)
	return items
}

func filterMidInt64OnItem(ctx context.Context, items []*search.Item) []*search.Item {
	if !midint64.IsDisableInt64MidVersion(ctx) {
		return items
	}
	var res []*search.Item
	for _, v := range items {
		var midArrays []int64
		midArrays = append(midArrays, v.Mid)
		if v.Upper != nil {
			midArrays = append(midArrays, v.Upper.Mid)
		}
		if v.BrandADAccount != nil {
			midArrays = append(midArrays, v.BrandADAccount.Mid)
		}
		if midint64.CheckHasInt64InMids(midArrays...) {
			continue
		}
		res = append(res, v)
	}
	return res
}

/*
*
0：旧版up主
1：新版up主卡
2: ipad hd up 主卡
*/
func upCardVersion(version, build int, plat int8, config *configs.SearchBuildLimit) int {
	v := 0
	//nolint:gomnd
	if version == 60200 {
		v = 1
	}
	if (model.IsAndroid(plat) && build < config.UpNewAndroid) || (model.IsIOS(plat) && build < config.UpNewIOS) {
		v = 0
	}
	// ipad hd 31600000 后有写额外逻辑
	if model.IsIPadHD(plat) && build >= 31600000 ||
		model.IsAndroidHD(plat) ||
		model.IsIPadPink(plat) && build >= 64000000 {
		v = 2
	}
	return v
}

// SearchByType is tag bangumi movie upuser video search
func (s *Service) SearchByType(c context.Context, mid int64, mobiApp, device, platform, buvid, sType, keyword, filtered, order, qvid string, plat int8, build, highlight, categoryID, userType, orderSort, pn, ps int, fnver, fnval, qn, fourk int64, old bool, now time.Time) (res *search.TypeSearch, code int, err error) {
	switch sType {
	case "season":
		if res, code, err = s.dao.Season(c, mid, keyword, mobiApp, device, platform, buvid, filtered, plat, build, pn, ps, now); err != nil {
			return
		}
	case "upper":
		if res, code, err = s.upper(c, mid, keyword, mobiApp, device, platform, buvid, filtered, order, qvid, s.biliUserVideoLimit, highlight, build, userType, orderSort, pn, ps, old, now); err != nil {
			return
		}
	case "movie":
		if !model.IsOverseas(plat) {
			if res, code, err = s.dao.MovieByType(c, mid, keyword, mobiApp, device, platform, buvid, filtered, plat, build, pn, ps, now); err != nil {
				return
			}
		}
	case "live_room", "live_user":
		if !cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
			return
		}
		if res, code, err = s.dao.LiveByType(c, mid, keyword, mobiApp, device, platform, buvid, filtered, order, sType, qvid, plat, build, pn, ps, now); err != nil {
			return
		}
	case "article":
		if res, code, err = s.article(c, mid, highlight, keyword, mobiApp, device, platform, buvid, filtered, order, sType, qvid, plat, categoryID, build, pn, ps, now); err != nil {
			return
		}
	case "season2":
		if (mobiApp == "android" && build <= s.c.SearchBuildLimit.PGCHighLightAndroid) || (model.IsIOS(plat) && build <= s.c.SearchBuildLimit.PGCHighLightIOS) ||
			(mobiApp == "android_i" && build <= 2033000) {
			highlight = 0
		}
		if res, code, err = s.dao.Season2(c, mid, keyword, mobiApp, device, platform, buvid, qvid, highlight, build, pn, ps, fnver, fnval, qn, fourk); err != nil {
			return
		}
	case "movie2":
		if (mobiApp == "android" && build <= s.c.SearchBuildLimit.PGCHighLightAndroid) || (model.IsIOS(plat) && build <= s.c.SearchBuildLimit.PGCHighLightIOS) ||
			(mobiApp == "android_i" && build <= 2033000) {
			highlight = 0
		}
		if res, code, err = s.dao.MovieByType2(c, mid, keyword, mobiApp, device, platform, buvid, qvid, highlight, build, pn, ps, fnver, fnval, qn, fourk); err != nil {
			return
		}
	case "tag":
		if res, code, err = s.channel(c, mid, keyword, mobiApp, platform, buvid, device, order, sType, build, pn, ps, highlight, plat); err != nil {
			return
		}
	case "video":
		if res, code, err = s.dao.Video(c, mid, keyword, mobiApp, device, platform, buvid, order, highlight, build, pn, ps); err != nil {
			return
		}
	}
	if res == nil {
		res = &search.TypeSearch{Items: []*search.Item{}}
	}
	res.Items = filterMidInt64OnItem(c, res.Items)
	return
}

// SearchLive is search live
func (s *Service) SearchLive(c context.Context, mid int64, mobiApp, platform, buvid, device, sType, keyword, order, qvid string, build, pn, ps int) (res *search.TypeSearch, err error) {
	if res, err = s.dao.Live(c, mid, keyword, mobiApp, platform, buvid, device, order, sType, qvid, build, pn, ps); err != nil {
		return
	}
	if res == nil {
		res = &search.TypeSearch{Items: []*search.Item{}}
	}
	res.Items = filterMidInt64OnItem(c, res.Items)
	return
}

// SearchLiveAll is search live
func (s *Service) SearchLiveAll(c context.Context, mid int64, mobiApp, platform, buvid, device, sType, keyword, order string, build, pn, ps int) (res *search.TypeSearchLiveAll, err error) {
	var (
		g         *errgroup.Group
		ctx       context.Context
		uid       int64
		owners    []int64
		glorys    []*search.LiveGlory
		follows   map[int64]bool
		userInfos map[int64]map[string]*search.Exp
	)
	if res, err = s.dao.LiveAll(c, mid, keyword, mobiApp, platform, buvid, device, order, sType, build, pn, ps); err != nil {
		return
	}
	if res.Master != nil {
		for _, item := range res.Master.Items {
			uid = item.Mid
			owners = append(owners, uid)
			break
		}
	}
	if len(owners) != 0 {
		g, ctx = errgroup.WithContext(c)
		if uid > 0 {
			g.Go(func() error {
				follows = s.dao.Relations3(ctx, owners, mid)
				return nil
			})
			g.Go(func() error {
				glorys, _ = s.dao.LiveGlory(ctx, uid)
				return nil
			})
			g.Go(func() error {
				userInfos, _ = s.dao.UserInfo(ctx, owners)
				return nil
			})
		}
		if err = g.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, m := range res.Master.Items {
			if follows[uid] {
				m.IsAttention = 1
			}
			m.Glory = &search.Glory{
				Title: "主播荣誉",
				Total: len(glorys),
				Items: make([]*search.Item, 0, len(glorys)),
			}
			if userInfo, ok := userInfos[m.Mid]; ok {
				if u, ok := userInfo["exp"]; ok {
					if u != nil || u.Master != nil {
						m.Level = u.Master.Level
						m.LevelColor = u.Master.Color
					}
				}
			}
			for _, glory := range glorys {
				if glory.GloryInfo != nil {
					item := &search.Item{
						Title: glory.GloryInfo.Name,
						Cover: glory.GloryInfo.Cover,
					}
					m.Glory.Items = append(m.Glory.Items, item)
				}
			}
		}
	}
	if res == nil {
		res = &search.TypeSearchLiveAll{Master: &search.TypeSearch{Items: []*search.Item{}}, Room: &search.TypeSearch{Items: []*search.Item{}}}
	}
	return
}

// channel search for channel
//
//nolint:gocognit
func (s *Service) channel(c context.Context, mid int64, keyword, mobiApp, platform, buvid, device, order, sType string, build, pn, ps, highlight int, plat int8) (res *search.TypeSearch, code int, err error) {
	var (
		g              *errgroup.Group
		ctx            context.Context
		tags           []int64
		tagMyInfos     []*search.Tag
		channelDetails map[int64]*channelgrpc.ChannelCard
	)
	if res, code, err = s.dao.Channel(c, mid, keyword, mobiApp, platform, buvid, device, order, sType, build, pn, ps, highlight); err != nil {
		return
	}
	if res == nil || len(res.Items) == 0 {
		return
	}
	tags = make([]int64, 0, len(res.Items))
	for _, item := range res.Items {
		tags = append(tags, item.ID)
	}
	if len(tags) != 0 {
		g, ctx = errgroup.WithContext(c)
		if mid > 0 {
			g.Go(func() error {
				tagMyInfos, _ = s.dao.TagInfos(ctx, tags, mid)
				return nil
			})
		}
		if (model.IsIPhone(plat) && build > s.c.SearchBuildLimit.NewChannelIOS) || (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.NewChannelAndroid) {
			g.Go(func() (err error) {
				if channelDetails, err = s.dao.Details(ctx, tags); err != nil {
					log.Error("err=%+v", err)
					err = nil
				}
				return
			})
		}
		if err = g.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, item := range res.Items {
			for _, myInfo := range tagMyInfos {
				if myInfo != nil && myInfo.TagID == item.ID {
					item.IsAttention = myInfo.IsAtten
					break
				}
			}
			if (model.IsIPhone(plat) && build > s.c.SearchBuildLimit.NewChannelIOS) || (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.NewChannelAndroid) {
				if channelDetails != nil {
					if channelDetail, ok := channelDetails[item.ID]; ok {
						item.Cover = channelDetail.Icon
						item.URI = model.FillURI(model.GotoChannelNew, item.Param, model.ChannelHandler("tab=select"))
					}
				}
			}
		}
	}
	return
}

// upper search for upper
func (s *Service) upper(c context.Context, mid int64, keyword, mobiApp, device, platform, buvid, filtered, order, qvid string, biliUserVL, highlight, build, userType, orderSort, pn, ps int, old bool, now time.Time) (res *search.TypeSearch, code int, err error) {
	var (
		g         *errgroup.Group
		ctx       context.Context
		owners    []int64
		follows   map[int64]bool
		accCards  map[int64]*account.Card
		nftRegion map[int64]*gallerygrpc.NFTRegion
		// 新订阅关系
		relationm map[int64]*relationgrpc.InterrelationReply
	)
	if res, code, err = s.dao.Upper(c, mid, keyword, mobiApp, device, platform, buvid, filtered, order, qvid, biliUserVL, highlight, build, userType, orderSort, pn, ps, old, now, s.systemNotice); err != nil {
		return
	}
	if res == nil || len(res.Items) == 0 {
		return
	}
	owners = make([]int64, 0, len(res.Items))
	for _, item := range res.Items {
		owners = append(owners, item.Mid)
	}
	if len(owners) != 0 {
		g, ctx = errgroup.WithContext(c)
		if mid > 0 {
			g.Go(func() error {
				follows = s.dao.Relations3(ctx, owners, mid)
				return nil
			})
			g.Go(func() error {
				if relationm, err = s.dao.Interrelations(ctx, mid, owners); err != nil {
					log.Error("%+v", err)
				}
				return nil
			})
		}
		g.Go(func() error {
			var err error
			if accCards, err = s.dao.Cards3(ctx, owners); err != nil {
				log.Error("accDao.Cards Owners %v, Err %v", owners, err)
			}
			return nil
		})
		g.Go(func() error {
			nftRegion, err = s.getNFTIconInfo(ctx, owners)
			if err != nil {
				log.Error("s.getNFTIconInfo err=%+v", err)
				return nil
			}
			return nil
		})
		if err = g.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, item := range res.Items {
			if follows[item.Mid] {
				item.Attentions = 1
			}
			item.Relation = cardmdl.RelationChange(item.Mid, relationm)
			if card, ok := accCards[item.Mid]; ok {
				item.Vip = &card.Vip
				item.FaceNftNew = card.FaceNftNew
				if nftRegion != nil && card.FaceNftNew == 1 {
					if v, ok := nftRegion[item.Mid]; ok {
						item.NftFaceIcon = &search.NftFaceIcon{
							RegionType: int32(v.Type),
							Icon:       v.Icon,
							ShowStatus: int32(v.ShowStatus),
						}
					}
				}
				item.IsSeniorMember = card.IsSeniorMember
			}
		}
		res.Items = filterMidInt64OnItem(ctx, res.Items)
	}
	return
}

// article search for article
func (s *Service) article(c context.Context, mid int64, highlight int, keyword, mobiApp, device, platform, buvid, filtered, order, sType, qvid string, plat int8, categoryID, build, pn, ps int, now time.Time) (res *search.TypeSearch, code int, err error) {
	if res, code, err = s.dao.ArticleByType(c, mid, keyword, mobiApp, device, platform, buvid, filtered, order, sType, qvid, plat, categoryID, build, highlight, pn, ps, now); err != nil {
		log.Error("%+v", err)
		return
	}
	if res != nil && len(res.Items) > 0 {
		var mids []int64
		for _, v := range res.Items {
			mids = append(mids, v.Mid)
		}
		var infom map[int64]*account.Info
		if infom, err = s.dao.Infos3(c, mids); err != nil {
			log.Error("%+v", err)
			err = nil
			return
		}
		for _, item := range res.Items {
			if info, ok := infom[item.Mid]; ok {
				item.Name = info.Name
			}
		}
	}
	return
}

// HotSearch is hot word search
func (s *Service) HotSearch(c context.Context, buvid string, mid int64, build, limit int, mobiApp, device, platform string, now time.Time) (res *search.Hot) {
	zoneId := _defaultZoneID
	if zone, err := s.dao.LocationInfo(c, metadata.String(c, metadata.RemoteIP)); err == nil && zone != nil {
		zoneId = int(zone.ZoneId)
	}
	var err error
	if res, err = s.dao.HotSearch(c, buvid, mid, build, limit, zoneId, mobiApp, device, platform, now); err != nil {
		log.Error("%+v", err)
	}
	if res != nil {
		res.TrackID = res.SeID
		res.SeID = ""
		res.Code = 0
		for _, re := range res.List {
			switch re.GotoType {
			case search.HotTypeArchive:
				re.Goto = model.GotoAv
				re.Param = re.GotoValue
				re.URI = model.FillURI(re.Goto, re.Param, nil)
			case search.HotTypeArticle:
				re.Goto = model.GotoArticle
				re.Param = re.GotoValue
				re.URI = model.FillURI(re.Goto, re.Param, nil)
			case search.HotTypePGC:
				re.Goto = model.GotoEP
				re.Param = re.GotoValue
				re.URI = model.FillURI(re.Goto, re.Param, nil)
			case search.HotTypeURL:
				re.Goto = model.GotoWeb
				re.URI = model.FillURI(re.Goto, re.GotoValue, nil)
			}
			re.GotoType = 0
			re.GotoValue = ""
			re.ModuleID = re.ID
			re.ID = 0
			re.Position = re.Pos
			re.Pos = 0
		}
	} else {
		res = &search.Hot{}
	}
	return
}

func (s *Service) Trending(c context.Context, buvid string, mid int64, build, limit int, mobiApp, device, platform string, now time.Time) (res *search.Hot) {
	zoneId := _defaultZoneID
	if zone, err := s.dao.LocationInfo(c, metadata.String(c, metadata.RemoteIP)); err == nil && zone != nil {
		zoneId = int(zone.ZoneId)
	}
	var err error
	if res, err = s.dao.Trending(c, buvid, mid, build, limit, zoneId, mobiApp, device, platform, now, false); err != nil {
		log.Error("%+v", err)
	}
	if res == nil {
		res = &search.Hot{}
		return
	}
	res.TrackID = res.SeID
	res.SeID = ""
	for _, re := range res.List {
		switch re.GotoType {
		case search.HotTypeArchive:
			re.Goto = model.GotoAv
			re.Param = re.GotoValue
			re.URI = model.FillURI(re.Goto, re.Param, nil)
		case search.HotTypeArticle:
			re.Goto = model.GotoArticle
			re.Param = re.GotoValue
			re.URI = model.FillURI(re.Goto, re.Param, nil)
		case search.HotTypePGC:
			re.Goto = model.GotoEP
			re.Param = re.GotoValue
			re.URI = model.FillURI(re.Goto, re.Param, nil)
		case search.HotTypeURL:
			re.Goto = model.GotoWeb
			re.URI = model.FillURI(re.Goto, re.GotoValue, nil)
		}
		re.GotoType = 0
		re.GotoValue = ""
		re.ModuleID = re.ID
		re.ID = 0
		re.Position = re.Pos
		re.Pos = 0
	}
	if !checkTrendingRevisionSupportVersion(c) {
		trendingLimit := s.c.Search.TrendingLimit
		if trendingLimit > 0 && len(res.List) > trendingLimit {
			res.List = res.List[:trendingLimit]
		}
	}
	return
}

func checkTrendingRevisionSupportVersion(ctx context.Context) bool {
	return pd.WithContext(ctx).Where(func(pdContext *pd.PDContext) {
		pdContext.IsPlatAndroid().And().Build(">=", 6870000)
	}).OrWhere(func(pdContext *pd.PDContext) {
		pdContext.IsPlatIPhone().And().Build(">=", 68700000)
	}).MustFinish()
}

// Suggest for search suggest
func (s *Service) Suggest(c context.Context, mid int64, buvid, keyword string, build int, mobiApp, device string, now time.Time) (res *search.Suggestion) {
	var (
		suggest *search.Suggest
		err     error
	)
	res = &search.Suggestion{}
	if s.c.Switch.SearchSuggest {
		return
	}
	if suggest, err = s.dao.Suggest(c, mid, buvid, keyword, build, mobiApp, device, now); err != nil {
		log.Error("%+v", err)
		return
	}
	if suggest != nil {
		res.UpUser = suggest.Result.Accurate.UpUser
		res.Bangumi = suggest.Result.Accurate.Bangumi
		for _, v := range suggest.Result.Tag {
			res.Suggest = append(res.Suggest, v.Value)
		}
		res.TrackID = suggest.Stoken
	}
	return
}

// Suggest2 for search suggest
func (s *Service) Suggest2(c context.Context, mid int64, platform, buvid, keyword string, build int, mobiApp string, now time.Time, device string) (res *search.Suggestion2) {
	var (
		suggest *search.Suggest2
		err     error
		avids   []int64
		avm     map[int64]*arcgrpc.Arc
		roomIDs []int64
		lm      map[int64]*livexroom.Infos
	)
	res = &search.Suggestion2{}
	if s.c.Switch.SearchSuggest {
		return
	}
	if suggest, err = s.dao.Suggest2(c, mid, platform, buvid, keyword, build, mobiApp, now); err != nil {
		log.Error("%+v", err)
		return
	}
	plat := model.Plat(mobiApp, device)
	if suggest.Result != nil {
		for _, v := range suggest.Result.Tag {
			if v.SpID == search.SuggestionJump {
				if v.Type == search.SuggestionAV {
					avids = append(avids, v.Ref)
				}
				if v.Type == search.SuggestionLive && !model.IsBlue(plat) {
					roomIDs = append(roomIDs, v.Ref)
				}
			}
		}
		g, ctx := errgroup.WithContext(c)
		if len(avids) != 0 {
			g.Go(func() (err error) {
				if avm, err = s.dao.Archives(ctx, avids, mobiApp, device, mid); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				return
			})
		}
		if len(roomIDs) != 0 {
			g.Go(func() (err error) {
				if lm, err = s.dao.GetMultiple(ctx, roomIDs); err != nil {
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
		for _, v := range suggest.Result.Tag {
			if v.Type == search.SuggestionLive && model.IsBlue(plat) {
				continue
			}
			si := &search.Item{}
			si.FromSuggest2(v, avm, lm)
			res.List = append(res.List, si)
		}
		res.TrackID = suggest.Stoken
	}
	return
}

// Suggest3 for search suggest
func (s *Service) Suggest3Json(c context.Context, mid int64, platform, buvid, keyword, device string, build, highlight int, mobiApp string, now time.Time) (res *search.SuggestionResult3) {
	var (
		suggest   *search.Suggest3
		err       error
		avids     []int64
		avm       map[int64]*arcgrpc.Arc
		roomIDs   []int64
		entryRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList
		ssids     []int32
		seasonm   map[int32]*pgcsearch.SearchCardProto
		mids      []int64
		nftRegion map[int64]*gallerygrpc.NFTRegion
	)
	res = &search.SuggestionResult3{}
	if s.c.Switch.SearchSuggest {
		return
	}
	if suggest, err = s.dao.Suggest3(c, mid, platform, buvid, keyword, device, build, highlight, mobiApp, now); err != nil {
		log.Error("%+v", err)
		return
	}
	plat := model.Plat(mobiApp, device)
	var buildLimit bool
	buildLimit = cdm.ShowLive(mobiApp, device, build)
	if s.c.Feature.FeatureBuildLimit.Switch {
		buildLimit = cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, &feature.OriginResutl{
			BuildLimit: !cdm.ShowLive(mobiApp, device, build),
		})
	}
	for _, v := range suggest.Result {
		if v.TermType == search.SuggestionJump {
			if v.SubType == search.SuggestionAV {
				avids = append(avids, v.Ref)
			}
			if v.SubType == search.SuggestionLive && buildLimit && !model.IsOverseas(plat) {
				roomIDs = append(roomIDs, v.Ref)
			}
		} else if v.TermType == search.SuggestionJumpPGC && !model.IsOverseas(plat) {
			if v.PGC == nil && v.PGC.SeasonID != 0 {
				continue
			}
			ssids = append(ssids, int32(v.PGC.SeasonID))
		} else if v.TermType == search.SuggestionJumpUser {
			mids = append(mids, v.User.Mid)
		}
	}
	g, ctx := errgroup.WithContext(c)
	if len(mids) != 0 {
		g.Go(func() (err error) {
			nftRegion, err = s.getNFTIconInfo(ctx, mids)
			if err != nil {
				log.Error("s.getNFTIconInfo mids=%+v, err=%+v", mids, err)
				return nil
			}
			return
		})
	}
	if len(avids) != 0 {
		g.Go(func() (err error) {
			if avm, err = s.dao.Archives(ctx, avids, mobiApp, device, mid); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
	}
	if len(roomIDs) != 0 {
		g.Go(func() (err error) {
			req := &livexroomgate.EntryRoomInfoReq{
				EntryFrom: []string{model.DefaultLiveEntry},
				RoomIds:   roomIDs,
				Uid:       mid,
				Uipstr:    metadata.String(ctx, metadata.RemoteIP),
				Platform:  platform,
				Build:     int64(build),
				Network:   "other",
			}
			if entryRoom, err = s.dao.EntryRoomInfo(ctx, req); err != nil {
				log.Error("Failed to get entry room info: %+v: %+v", req, err)
				err = nil
				return
			}
			return
		})
	}
	if len(ssids) != 0 {
		g.Go(func() (err error) {
			if seasonm, err = s.dao.SugOGV(ctx, ssids); err != nil {
				log.Error("%v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	for _, v := range suggest.Result {
		if v.SubType == search.SuggestionLive && (!buildLimit || model.IsOverseas(plat)) {
			continue
		}
		if v.TermType == search.SuggestionJumpPGC && model.IsOverseas(plat) {
			continue
		}
		si := &search.Item{}
		si.FromSuggest3(v, avm, entryRoom, seasonm, nftRegion)
		res.List = append(res.List, si)
	}
	res.TrackID = suggest.TrackID
	res.ExpStr = suggest.ExpStr
	return
}

func (s *Service) getNFTIconInfo(ctx context.Context, mids []int64) (map[int64]*gallerygrpc.NFTRegion, error) {
	req := &memberAPI.NFTBatchInfoReq{
		Mids:   mids,
		Status: "inUsing",
		Source: "face",
	}
	reply, err := s.dao.NFTBatchInfo(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "s.accDao.NFTBatchInfo req=%+v", req)
	}
	var (
		nftIDs        []string
		nftRegionInfo *gallerygrpc.GetNFTRegionReply
	)
	for _, v := range reply.GetNftInfos() {
		nftIDs = append(nftIDs, v.NftId)
	}
	if len(nftIDs) == 0 {
		return nil, err
	}
	if nftRegionInfo, err = s.dao.GetNFTRegionBatch(ctx, nftIDs); err != nil {
		return nil, errors.Wrapf(err, "s.galleryDao.GetNFTRegion nftIDs=%+v", nftIDs)
	}
	res := make(map[int64]*gallerygrpc.NFTRegion, len(nftIDs))
	for _, info := range reply.GetNftInfos() {
		if v, ok := nftRegionInfo.Region[info.NftId]; ok {
			res[info.Mid] = v
		}
	}
	return res, nil
}

// User for search uer
func (s *Service) User(c context.Context, mid int64, buvid, mobiApp, device, platform, keyword, filtered, order, fromSource string, highlight, build, userType, orderSort, pn, ps int, now time.Time) (res *search.UserResult) {
	res = &search.UserResult{}
	user, err := s.dao.User(c, mid, keyword, mobiApp, device, platform, buvid, filtered, order, fromSource, highlight, build, userType, orderSort, pn, ps, now)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if len(user) == 0 {
		return
	}
	res.Items = make([]*search.Item, 0, len(user))
	for _, u := range user {
		res.Items = append(res.Items, &search.Item{Mid: u.Mid, Name: u.Name, Face: u.Pic})
	}
	return
}

// convertNav deal with old search pageinfo to new.
//
//nolint:gocognit
func (s *Service) convertNav(c context.Context, all *search.Search, plat int8, build int, lang, _, _ string, old, newPGC bool) (nis []*search.NavInfo) {
	const (
		_showHide          = 0
		_oldAndroidArticle = 515009
	)
	var (
		season  = "番剧"
		live    = "直播"
		upper   = "用户"
		movie   = "影视"
		article = "专栏"
	)
	if old {
		upper = "UP主"
	}
	if lang == model.Hant {
		season = "番劇"
		live = "直播"
		upper = "UP主"
		movie = "影視"
		article = "專欄"
	}
	nis = make([]*search.NavInfo, 0, 5)
	// season
	if !newPGC && all.PageInfo.Bangumi != nil {
		var nav = &search.NavInfo{
			Name:  season,
			Total: all.PageInfo.Bangumi.NumResults,
			Pages: all.PageInfo.Bangumi.Pages,
			Type:  1,
		}
		if all.PageInfo.Bangumi.NumResults > s.seasonNum {
			nav.Show = s.seasonShowMore
		} else {
			nav.Show = _showHide
		}
		nis = append(nis, nav)
	}
	// media season
	if newPGC && all.PageInfo.MediaBangumi != nil {
		var nav = &search.NavInfo{
			Name:  season,
			Total: all.PageInfo.MediaBangumi.NumResults,
			Pages: all.PageInfo.MediaBangumi.Pages,
			Type:  7,
		}
		if all.PageInfo.MediaBangumi.NumResults > s.seasonNum {
			nav.Show = s.seasonShowMore
		} else {
			nav.Show = _showHide
		}
		nis = append(nis, nav)
	}
	// live
	if cdm.ShowLiveV2(c, s.c.Feature.FeatureBuildLimit.ShowLive, nil) {
		if (model.IsAndroid(plat) && build > search.SearchLiveAllAndroid) || (model.IsIPhone(plat) && build > search.SearchLiveAllIOS) || ((plat == model.PlatIPad && build >= search.SearchNewIPad) || (plat == model.PlatIpadHD && build >= search.SearchNewIPadHD)) {
			if all.PageInfo.LiveAll != nil {
				var nav = &search.NavInfo{
					Name:  live,
					Total: all.PageInfo.LiveAll.NumResults,
					Pages: all.PageInfo.LiveAll.Pages,
					Type:  4,
				}
				nis = append(nis, nav)
			}
		} else {
			if all.PageInfo.LiveRoom != nil {
				var nav = &search.NavInfo{
					Name:  live,
					Total: all.PageInfo.LiveRoom.NumResults,
					Pages: all.PageInfo.LiveRoom.Pages,
					Type:  4,
				}
				nis = append(nis, nav)
			}
		}
	}
	// upper
	if old {
		if all.PageInfo.UpUser != nil {
			var nav = &search.NavInfo{
				Name:  upper,
				Total: all.PageInfo.UpUser.NumResults,
				Pages: all.PageInfo.UpUser.Pages,
				Type:  2,
			}
			nis = append(nis, nav)
		}
	} else {
		if all.PageInfo.BiliUser != nil {
			var nav = &search.NavInfo{
				Name:  upper,
				Total: all.PageInfo.BiliUser.NumResults,
				Pages: all.PageInfo.BiliUser.Pages,
				Type:  2,
			}
			nis = append(nis, nav)
		}
	}
	// movie
	if !newPGC && all.PageInfo.Film != nil {
		var nav = &search.NavInfo{
			Name:  movie,
			Total: all.PageInfo.Film.NumResults,
			Pages: all.PageInfo.Film.Pages,
			Type:  3,
		}
		if all.PageInfo.Movie != nil && all.PageInfo.Movie.NumResults > s.movieNum {
			nav.Show = s.movieShowMore
		} else {
			nav.Show = _showHide
		}
		nis = append(nis, nav)
	}
	// media movie
	if newPGC && all.PageInfo.MediaFt != nil {
		var nav = &search.NavInfo{
			Name:  movie,
			Total: all.PageInfo.MediaFt.NumResults,
			Pages: all.PageInfo.MediaFt.Pages,
			Type:  8,
		}
		if all.PageInfo.MediaFt.NumResults > s.movieNum {
			nav.Show = s.movieShowMore
		} else {
			nav.Show = _showHide
		}
		nis = append(nis, nav)
	}
	if all.PageInfo.Article != nil {
		if (model.IsIPhone(plat) && build > _oldIOS) || (model.IsAndroid(plat) && build > _oldAndroidArticle) || model.IsIPhoneB(plat) {
			var nav = &search.NavInfo{
				Name:  article,
				Total: all.PageInfo.Article.NumResults,
				Pages: all.PageInfo.Article.Pages,
				Type:  6,
			}
			nis = append(nis, nav)
		}
	}
	return
}

// RecommendNoResult search when no result
func (s *Service) RecommendNoResult(c context.Context, platform, mobiApp, device, buvid, keyword string, build, pn, ps int, mid int64) (res *search.NoResultRcndResult, err error) {
	if res, err = s.dao.RecommendNoResult(c, platform, mobiApp, device, buvid, keyword, build, pn, ps, mid); err != nil {
		log.Error("%+v", err)
	}
	return
}

// Recommend search recommend
func (s *Service) Recommend(c context.Context, mid int64, build, from, show, disableRcmd int, buvid, platform, mobiApp, device string) (res *search.RecommendResult, err error) {
	if s.c.Switch.SearchRecommend {
		return
	}
	if res, err = s.dao.Recommend(c, mid, build, from, show, disableRcmd, buvid, platform, mobiApp, device); err != nil {
		log.Error("%+v", err)
	}
	return
}

// DefaultWords search for default words
func (s *Service) DefaultWordsJson(c context.Context, mid int64, build, from int, buvid, platform, mobiApp, device string, loginEvent int64, extParam *search.DefaultWordsExtParam) (res *search.DefaultWords, err error) {
	if res, err = s.dao.DefaultWords(c, mid, build, from, buvid, platform, mobiApp, device, loginEvent, extParam); err != nil {
		log.Error("%+v", err)
	}
	return
}

// Resource for rsource
func (s *Service) Resource(c context.Context, mobiApp, device, network, buvid, adExtra string, build int, plat int8, mid int64) (res []*search.Banner, err error) {
	var (
		bnsm  map[int][]*resmdl.Banner
		resID int
	)
	if model.IsAndroid(plat) {
		resID = AndroidSearchResourceID
	} else if model.IsIPhone(plat) {
		resID = IPhoneSearchResourceID
	} else if model.IsPad(plat) {
		resID = IPadSearchResourceID
	}
	if bnsm, err = s.dao.Banner(c, mobiApp, device, network, "", buvid, adExtra, strconv.Itoa(resID), build, plat, mid); err != nil {
		return
	}
	// only one position
	for _, rb := range bnsm[resID] {
		b := &search.Banner{}
		b.ChangeBanner(rb)
		res = append(res, b)
		break
	}
	return
}

// RecommendPre search at pre-page.
func (s *Service) RecommendPre(c context.Context, platform, mobiApp, device, buvid string, build, ps int, mid int64) (res *search.RecommendPreResult, err error) {
	if res, err = s.dao.RecommendPre(c, platform, mobiApp, device, buvid, build, ps, mid); err != nil {
		log.Error("%+v", err)
	}
	return
}

// SearchEpisodes search PGC episodes
func (s *Service) SearchEpisodes(c context.Context, mid, ssID int64) (res []*search.Item, err error) {
	var (
		seasonIDs []int64
		bangumis  map[string]*search.Card
	)
	seasonIDs = []int64{ssID}
	if bangumis, err = s.dao.BangumiCard(c, mid, seasonIDs); err != nil {
		log.Error("%+v", err)
		return
	}
	if bangumi, ok := bangumis[strconv.FormatInt(ssID, 10)]; ok {
		for pos, v := range bangumi.Episodes {
			tmp := &search.Item{
				Param:    strconv.Itoa(int(v.ID)),
				Index:    v.Index,
				Badges:   v.Badges,
				Position: pos + 1,
				URI:      v.URL,
			}
			// tmp.URI = model.FillURI(model.GotoEP, tmp.Param, nil)
			res = append(res, tmp)
		}
	}
	return
}

// SearchEpsNew search PGC episodes, with the info from pgc grpc
func (s *Service) SearchEpsNew(c context.Context, req *search.EpisodesNewReq) (result *search.EpsNewResult, err error) {
	var reply *pgcsearch.SearchEpReply
	if reply, err = s.dao.SearchEpsGrpc(c, req); err != nil {
		log.Error("%+v", err)
		return
	}
	result = &search.EpsNewResult{
		Title: reply.Title,
		Total: reply.Total,
	}
	for pos, v := range reply.List {
		item := new(search.Item)
		item.Position = pos + 1
		item.FromPgcEp(v, s.c.Cfg.PgcSearchCard)
		result.Episodes = append(result.Episodes, item)
	}
	return
}

// SearchConverge search converge
func (s *Service) SearchConverge(c context.Context, mid, cid int64, trackID, platform, mobiApp, device, buvid, order, sort string, plat int8, build, pn, ps int) (res *search.ResultConverge, err error) {
	if res, err = s.dao.Converge(c, mid, cid, trackID, platform, mobiApp, device, buvid, order, sort, plat, build, pn, ps); err != nil {
		log.Error("%+v", err)
		return
	}
	var (
		owners  []int64
		follows map[int64]bool
	)
	owners = make([]int64, 0, len(res.UserItems))
	for _, item := range res.UserItems {
		owners = append(owners, item.Mid)
	}
	if len(owners) > 0 {
		if mid > 0 {
			follows = s.dao.Relations3(c, owners, mid)
		}
		for _, item := range res.UserItems {
			if follows[item.Mid] {
				item.IsAttention = 1
			}
		}
	}
	return
}

func (s *Service) loadHotCache() {
	log.Info("cronLog start loadHotCache")
	tmp, err := s.dao.AiRecommend(context.TODO())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.hotAids = tmp
}

func (s *Service) loadSearchTipsCache() {
	log.Info("cronLog start loadSearchTipsCache")
	tmp, err := s.dao.SearchTips(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.searchTipsCache = tmp
}

func (s *Service) SearchChannel(c context.Context, keyword, platform, mobiApp, device, buvid string, plat int8, build, pn, ps, highlight int, mid int64) (res *search.ChannelResult, err error) {
	var tids []int64
	if (model.IsAndroid(plat) && build > s.c.SearchBuildLimit.TypeSearchChannelESAndroid) || (model.IsIPhone(plat) && build > s.c.SearchBuildLimit.TypeSearchChannelESIOS) {
		if res, tids, err = s.dao.EsSearchChannel(c, mid, keyword, pn, ps, search.ChannelOK); err != nil {
			log.Error("%v", err)
			return
		}
	} else {
		if res, tids, err = s.dao.ChannelNew(c, mid, keyword, mobiApp, platform, buvid, device, build, pn, ps, highlight); err != nil {
			log.Error("%v", err)
			return
		}
	}
	if res == nil || len(tids) == 0 {
		return
	}
	var channels map[int64]*channelgrpc.SearchChannel
	if channels, err = s.dao.SearchChannel(c, mid, tids); err != nil {
		log.Error("%v", err)
		return
	}
	var items []*search.ChannleItem
	for _, tid := range tids {
		if channel, ok := channels[tid]; ok && channel != nil {
			i := &search.ChannleItem{}
			i.FormChannelNew(channel)
			if i.Right {
				items = append(items, i)
			}
		}
	}
	res.Items = items
	return
}

func (s *Service) inStreamingRoom(ctx context.Context, roomID []int64) map[int64]struct{} {
	reply, err := s.dao.EntryRoomInfo(ctx, &livexroomgate.EntryRoomInfoReq{
		EntryFrom:     []string{"NONE"},
		RoomIds:       roomID,
		NotPlayurl:    1,
		FilterOffline: 1,
	})
	if err != nil {
		return nil
	}
	out := map[int64]struct{}{}
	for _, r := range reply {
		// 直播中
		if r.LiveStatus == 1 {
			out[r.RoomId] = struct{}{}
		}
	}
	return out
}

func liveTrendingRoomID(in *search.Hot) []int64 {
	var out []int64
	for _, v := range in.List {
		if v.WordType == _liveRoomWordType {
			if len(v.LiveIds) > 0 {
				out = append(out, v.LiveIds...)
			}
		}
	}
	return out
}

func setShowLiveIcon(in *search.Hot, streamingRoom map[int64]struct{}) {
	for _, v := range in.List {
		if v.WordType == _liveRoomWordType {
			v.Icon = ""
			for _, liveId := range v.LiveIds {
				if liveId <= 0 {
					continue
				}
				_, inStreaming := streamingRoom[liveId]
				if inStreaming {
					v.ShowLiveIcon = true
					break
				}
			}
		}
	}
}

func (s *Service) Square(c context.Context, mid int64, mobiApp, device, network, platform, adExtra, buvid string, build, limit, from, show, disableRcmd int, now time.Time, isHant bool) (res []*search.IterationConverge, err error) {
	var (
		trending    *search.Hot
		recomResult *search.RecommendResult
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		trending = s.Trending(ctx, buvid, mid, build, limit, mobiApp, device, platform, now)
		if trending == nil || len(trending.List) == 0 || trending.Code != 0 {
			return errors.Errorf("搜索发现模块热搜降级 (%+v)", trending)
		}
		liveTrending := liveTrendingRoomID(trending)
		if len(liveTrending) > 0 {
			inStreamingRoom := s.inStreamingRoom(ctx, liveTrending)
			setShowLiveIcon(trending, inStreamingRoom)
		}
		return
	})
	g.Go(func() (err error) {
		if recomResult, err = s.Recommend(ctx, mid, build, from, show, disableRcmd, buvid, platform, mobiApp, device); err != nil {
			log.Error("Square s.Recommend() mid(%d) error(%v)", mid, err)
			err = nil
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	res = []*search.IterationConverge{
		{Type: "trending", Title: s.makeTrendingTitle(ctx), Data: trending, SearchRankingMeta: s.makeSearchRankingMeta(ctx, isHant)},
		{Type: "history", Title: s.c.SearchPageTitle.HistoryTitle},
		{Type: "recommend", Title: s.c.SearchPageTitle.FindTitle, Data: recomResult}}
	if checkTrendingRevisionSupportVersion(ctx) {
		switch trending.SearchHotwordRevision {
		case 1, 2:
			res = []*search.IterationConverge{
				{Type: "history", Title: s.c.SearchPageTitle.HistoryTitle, SearchHotWordRevision: trending.SearchHotwordRevision},
				{Type: "recommend", Title: s.c.SearchPageTitle.FindTitle, Data: recomResult, SearchHotWordRevision: trending.SearchHotwordRevision},
				{Type: "trending", Title: s.makeTrendingTitle(ctx), Data: trending, SearchRankingMeta: s.makeSearchRankingMeta(ctx, isHant), SearchHotWordRevision: trending.SearchHotwordRevision}}
		case 3:
			res = []*search.IterationConverge{
				{Type: "history", Title: s.c.SearchPageTitle.HistoryTitle, SearchHotWordRevision: 1},
				{Type: "trending", Title: s.makeTrendingTitle(ctx), Data: trending, SearchRankingMeta: s.makeSearchRankingMeta(ctx, isHant), SearchHotWordRevision: 1},
				{Type: "recommend", Title: s.c.SearchPageTitle.FindTitle, Data: recomResult, SearchHotWordRevision: 1}}
		default:
		}
	}
	// 标题简繁体转换处理
	if isHant {
		for _, v := range res {
			out := chinese.Converts(c, v.Title)
			v.Title = out[v.Title]
		}
	}
	return
}

func (s *Service) makeTrendingTitle(ctx context.Context) string {
	if !s.checkSearchRankingSwitchOpen(ctx) {
		return "热搜"
	}
	return "B站热搜"
}

func (s *Service) checkSearchRankingSwitchOpen(ctx context.Context) bool {
	// 配合客户端改动做版本控制
	if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().Or().IsPlatAndroidI().Or().IsPlatAndroidB().And().Build(">=", int64(6670000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().Or().IsPlatIPhoneI().Or().IsPlatIPhoneB().And().Build(">=", int64(66700000))
	}).MustFinish() {
		return s.c.Search.SearchRankingSwitch
	}
	return false
}

func (s *Service) makeSearchRankingMeta(ctx context.Context, isHant bool) *search.RankingMeta {
	if !s.checkSearchRankingSwitchOpen(ctx) {
		return nil
	}
	rankingMetaText := "完整榜单"
	if isHant {
		out := chinese.Converts(ctx, rankingMetaText)
		rankingMetaText = out[rankingMetaText]
	}
	return &search.RankingMeta{
		OpenSearchRanking: true,
		Text:              rankingMetaText,
		Link:              "https://www.bilibili.com/blackboard/activity-trending-topic.html?navhide=1",
	}
}

func (s *Service) SearchChannel2(c context.Context, params *search.Param) (res *search.ChannelResult, err error) {
	var tids, hideTids []int64
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if res, tids, err = s.dao.EsSearchChannel(ctx, params.MID, params.Keyword, params.PN, params.PS, search.ChannelOK); err != nil {
			log.Error("%+v", err)
		}
		return
	})
	g.Go(func() (err error) {
		if _, hideTids, err = s.dao.EsSearchChannel(ctx, params.MID, params.Keyword, 1, 50, search.ChannelHide); err != nil {
			log.Error("%+v", err)
			return nil
		}
		return
	})
	if err = g.Wait(); err != nil {
		return
	}
	var (
		channels    map[int64]*channelgrpc.SearchChannelCard
		more        []*channelgrpc.RelativeChannel
		hot         *channelgrpc.ChannelListReply
		allTids     = hideTids
		isFirstPage bool
	)
	if params.PN == 1 {
		isFirstPage = true
		allTids = append(allTids, tids...)
	}
	g2, ctx2 := errgroup.WithContext(c)
	if len(tids) > 0 {
		g2.Go(func() (err error) {
			if channels, err = s.dao.SearchChannelsInfo(ctx2, params.MID, tids); err != nil {
				log.Error("%v", err)
				res = nil
			}
			return
		})
	}
	if isFirstPage && len(allTids) > 0 {
		g2.Go(func() (err error) {
			if more, err = s.dao.RelativeChannel(ctx2, params.MID, allTids); err != nil {
				log.Error("%+v", err)
				return nil
			}
			return
		})
	}
	g2.Go(func() (err error) {
		if hot, err = s.dao.ChannelList(ctx2, params.MID, 100, ""); err != nil {
			log.Error("%+v", err)
			return nil
		}
		return
	})
	if err = g2.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	var (
		aids    []int64
		playAvs []*arcgrpc.PlayAv
	)
	for _, channel := range channels {
		if channel == nil {
			continue
		}
		for _, video := range channel.GetVideoCards() {
			if video.GetRid() == 0 {
				continue
			}
			aids = append(aids, video.GetRid())
			playAvs = append(playAvs, &arcgrpc.PlayAv{Aid: video.GetRid()})
		}
	}
	var apm map[int64]*arcgrpc.ArcPlayer
	if len(aids) > 0 {
		apm, _ = s.dao.ArcsPlayer(c, playAvs, false)
	}
	//版本判断
	var isHightBuild bool
	if pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsMobiAppIPhone().And().Build(">=", s.c.BuildLimit.OGVChanIOSBuild)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", s.c.BuildLimit.OGVChanAndroidBuild)
	}).FinishOr(false) {
		isHightBuild = true
	}
	for _, tid := range tids {
		if channel, ok := channels[tid]; ok && channel != nil {
			i := &search.ChannleItem{}
			i.FormChannel2(channel, apm, params.Plat, params.Build, isHightBuild, params.Spmid)
			//nolint:gomnd
			if len(i.Items) != 3 {
				res.FaildNum++
				log.Error("search channel(%d,%s) archives no enough three", channel.GetCid(), channel.GetCname())
				continue
			}
			res.Items = append(res.Items, i)
		}
	}
	res.NoMoreLabel = "没有更多结果啦~"
	if !isFirstPage || len(res.Items) > 1 {
		return
	}
	var items []*search.ChannleItem
	for _, m := range more {
		i := &search.ChannleItem{}
		i.FormChannelMore(m, params.MobiApp, params.Spmid, params.Build, isHightBuild)
		items = append(items, i)
	}
	if len(items) > 0 {
		res.NoSearchLabel = "没有找到相关频道，看看更多频道吧~"
		res.Extend = &search.ChannleItem2{
			Label:     "更多频道",
			ModelType: "more",
			Items:     items,
		}
		res.NoMoreLabel = "到底啦~"
		return
	}
	if len(res.Items) == 0 {
		for _, h := range hot.GetCards() {
			i := &search.ChannleItem{}
			i.FormChannelHot(c, h, isHightBuild, params.Spmid)
			items = append(items, i)
		}
		if len(items) > 0 {
			res.NoSearchLabel = "没有找到相关频道，看看有哪些热门频道吧~"
			res.Extend = &search.ChannleItem2{
				Label:     "热门频道",
				ModelType: "hot",
				Items:     items,
			}
			res.NoMoreLabel = "到底啦~"
		}
	}
	return
}

func (s *Service) ResolveCommand(ctx context.Context, req *search.SiriCommandReq) (*siriext.ResolveCommandReply, error) {
	//nolint:gomnd
	if len([]rune(req.Command)) > 50 {
		return nil, ecode.RequestErr
	}
	return s.dao.ResolveCommand(ctx, &siriext.ResolveCommandReq{
		Mid:     req.Mid,
		Command: req.Command,
		Debug:   req.Debug,
		Device: siriext.DeviceMeta{
			MobiApp:  req.MobiApp,
			Device:   req.Device,
			Build:    req.Build,
			Channel:  req.Channel,
			Buvid:    req.Buvid,
			Platform: req.Platform,
		},
	})

}

// nolint:unparam
func (s *Service) makeSportsInlineFns(ctx context.Context, ids []int64, matches map[int64]*esportsservice.SportsEventMatchItem) (map[int64]func(i *search.Item), map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, error) {
	loader := NewInlineCardFanoutLoader{General: constructGeneralParamFromCtx(ctx), Service: s}
	for _, v := range matches {
		if v.QueryCard == nil {
			continue
		}
		if v.QueryCard.AvId > 0 {
			loader.Archive.Aids = append(loader.Archive.Aids, v.QueryCard.AvId)
		}
		if v.QueryCard.UpMid > 0 {
			loader.Live.UpMids = append(loader.Live.UpMids, v.QueryCard.UpMid)
			loader.Live.LiveEntryFrom = []string{model.SearchEsInlineCard}
		}
	}
	fanout, err := loader.doSearchCardFanoutLoad(ctx)
	if err != nil {
		log.Error("makeTopGameMaterials doSearchCardFanoutLoad loader=%+v, error=%+v", loader, err)
	}
	inline := make(map[int64]func(i *search.Item), len(ids))
	for k, v := range matches {
		if inlineFn, ok := buildSportsInlineProcess(ctx, fanout, v.QueryCard); ok {
			inline[k] = inlineFn
		}
	}
	return inline, fanout.Live.InlineRoom, nil
}

func (s *Service) makeTopGameMaterials(ctx context.Context, data []*search.TopGameData, inlineInfo *search.TopGameInlineInfo, topGameCardIds map[int64]int64) (map[int64]*search.TopGameMaterial, error) {
	loader := NewInlineCardFanoutLoader{General: constructGeneralParamFromCtx(ctx), Service: s}
	for _, v := range data {
		if cardId, ok := topGameCardIds[v.GameBaseId]; ok {
			v.Avid = convertGameAvidForConfig(inlineInfo, v.GameBaseId, cardId, v.Avid)
		}
		if v.Avid > 0 {
			loader.Archive.Aids = append(loader.Archive.Aids, v.Avid)
		}
		if v.GameOfficialAccount > 0 {
			loader.Account.AccountUIDs = append(loader.Account.AccountUIDs, v.GameOfficialAccount)
		}
	}
	fanout, err := loader.doSearchCardFanoutLoad(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "makeTopGameMaterials doSearchCardFanoutLoad loader=%+v", loader)
	}
	res := make(map[int64]*search.TopGameMaterial, len(data))
	for _, v := range data {
		topGameMaterial := &search.TopGameMaterial{TopGameData: v}
		if inlineFn, ok := buildTopGameInlineProcess(ctx, fanout, v); ok {
			topGameMaterial.InlineFn = inlineFn
		}
		res[v.GameBaseId] = topGameMaterial
	}
	return res, nil
}

func convertGameAvidForConfig(info *search.TopGameInlineInfo, gameId, cardId, rawAvid int64) int64 {
	if info == nil {
		return rawAvid
	}
	for _, inline := range info.InlineInfos {
		if inline.GameId == gameId && inline.CardId == cardId && inline.Avid > 0 {
			return inline.Avid
		}
	}
	return rawAvid
}

func (s *Service) fetchOgvChannelMaterial(ctx context.Context, bizId int64, bizType channelgrpc.ChannelBizlType) (*search.OgvChannelMaterial, error) {
	res := &search.OgvChannelMaterial{BizId: bizId, BizType: int64(bizType)}
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		reply, err := s.dao.GetMediaBizInfoByMediaBizId(ctx, bizId)
		if err != nil {
			return err
		}
		res.MediaBizInfo = reply
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		reply, err := s.dao.GetMediaReviewInfo(ctx, bizId)
		if err != nil {
			return err
		}
		res.ReviewInfo = reply
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		rep, e := s.dao.GetMediaAllowReview(ctx, int32(bizId))
		if e != nil {
			return nil
		}
		res.AllowReview = rep
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}

func OptTopGameTabInfoFn(gameId, gameCardId int64, config *search.TopGameConfig) (func(i *search.Item), bool) {
	if config == nil || len(config.ButtonInfos) == 0 {
		return nil, false
	}
	for _, v := range config.ButtonInfos {
		if v.GameId == gameId && v.CardId == gameCardId {
			return func(i *search.Item) { i.TabInfo = makeTopGameConfigButtons(v.Infos) }, true
		}
	}
	return nil, false
}

func makeTopGameConfigButtons(buttons []*search.TopGameConfigButton) []*search.TabInfo {
	var tabInfos []*search.TabInfo
	for _, v := range buttons {
		tabInfos = append(tabInfos, &search.TabInfo{TabName: v.Content, TabUrl: v.Url})
	}
	return tabInfos
}

func castAsHotAidSet(in map[int64]struct{}) sets.Int64 {
	out := sets.Int64{}
	for aid := range in {
		out.Insert(aid)
	}
	return out
}

func enableWithESportSearchConfig(plat int8, build int64) bool {
	if (model.IsIOSPick(plat) && build >= 64100000) ||
		(model.IsIPadHD(plat) && build >= 32500000) ||
		(model.IsAndroid(plat) && build >= 6410000) {
		return true
	}
	return false
}

func checkESportSearchInline(esportConfigs map[int64]*managersearch.EsportConfigInfo, esport *search.ESport, matchm map[int64]*esportGRPC.Contest, matchLiveEntryRoom map[int64]*livexroomgate.EntryRoomInfoResp_EntryList) (int64, bool) {
	if ec, ok := esportConfigs[esport.ID]; !ok || ec.IsInline != 1 || len(esport.MatchList) < 1 {
		return 0, false
	}
	if mc, ok := matchm[esport.MatchList[0].ID]; ok {
		if room, ok := matchLiveEntryRoom[mc.LiveRoom]; ok {
			return room.Uid, ok
		}
	}
	return 0, false
}
