package dynamicV2

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	bcgmdl "go-gateway/app/app-svr/app-dynamic/interface/model/bcg"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	"go-gateway/pkg/idsafe/bvid"

	accAPI "git.bilibili.co/bapis/bapis-go/account/service"
	relagrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dynCheeseGrpc "git.bilibili.co/bapis/bapis-go/cheese/service/dynamic"
	shareApi "git.bilibili.co/bapis/bapis-go/community/interface/share"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynactivitygrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/activity"
	dyncampusgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/campus-svr"
	dyngrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	pgcFollowGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
)

type Handler func(context.Context, *mdlv2.DynamicContext, *mdlv2.GeneralParam) error

const (
	_handleTypeVideo             = "vidoe"
	_handleTypeVideoPersonal     = "vidoe_personal"
	_handleTypeDetail            = "detail"
	_handleTypeAll               = "all"
	_handleTypeAllPersonal       = "all_personal"
	_handleTypeAllFilter         = "all_filter" // 综合页筛选器
	_handleTypeShare             = "share"
	_handleTypeForward           = "forward"
	_handleTypeFake              = "fake"
	_handleTypeRepost            = "repost"
	_handleTypeLight             = "light"
	_handleTypeView              = "view"
	_handleTypeSpace             = "space"
	_handleTypeSearch            = "search"
	_handleTypeUnLogin           = "unlogin"
	_handleTypeReservePersonal   = "reserve_personal"
	_handleTypeSchool            = "school"
	_handleTypeSpaceSearchDetail = "space_search"
	_handleTypeServerDetail      = "server_detail"
	_handleTypeSchoolTopicFeed   = "school_topic" // 校园话题讨论
	_handleTypeLBS               = "lbs"
	_handleTypeLegacyTopic       = "legacy_topic" // 老话题feed流

	// ad ext
	// 起飞内容
	_adContentFly = 1
)

var (
	// 竖屏视频直接去story播放器的场景
	_verticalAvToStory = map[string]bool{
		_handleTypeAll: true, _handleTypeAllFilter: true, _handleTypeAllPersonal: true,
		_handleTypeVideo: true, _handleTypeVideoPersonal: true,
		_handleTypeDetail: true, // 折叠展开
		_handleTypeSpace:  true, _handleTypeSpaceSearchDetail: true,
	}
	// 类似于个人空间页的场景
	isDynSpaceLike = map[string]bool{
		_handleTypeSpaceSearchDetail: true, _handleTypeSpace: true,
		_handleTypeAllPersonal: true, _handleTypeVideoPersonal: true,
	}
)

func (s *Service) fromName(from string) string {
	switch from {
	case _handleTypeVideo:
		return "视频页"
	case _handleTypeVideoPersonal:
		return "视频快消页"
	case _handleTypeAll:
		return "综合页"
	case _handleTypeAllFilter:
		return "综合页(筛选器)"
	case _handleTypeAllPersonal:
		return "综合快消页"
	case _handleTypeFake:
		return "假卡"
	case _handleTypeDetail:
		return "折叠展开"
	case _handleTypeRepost:
		return "详情页-转发列表"
	case _handleTypeLight:
		return "轻浏览"
	case _handleTypeView:
		return "详情页"
	case _handleTypeSpace:
		return "空间/直播"
	case _handleTypeSearch:
		return "垂搜页"
	case _handleTypeUnLogin:
		return "未登录推荐页"
	case _handleTypeSchool:
		return "校园页"
	case _handleTypeReservePersonal:
		return "快消页预约"
	case _handleTypeSpaceSearchDetail:
		return "空间搜索页"
	case _handleTypeServerDetail:
		return "服务端批量查询动态"
	case _handleTypeSchoolTopicFeed:
		return "校园话题讨论"
	case _handleTypeLBS:
		return "LBS"
	}
	return "未知"
}

// followings 获取关注链信息
// nolint:unparam
func (s *Service) followings(c context.Context, mid int64, withAnime, withCinema bool, general *mdlv2.GeneralParam) (*relagrpc.FollowingsReply, *pgcFollowGrpc.MyRelationsReply, *dynCheeseGrpc.MyPaidReply, *favgrpc.BatchFavsReply, []int64, error) {
	eg := errgroup.WithCancel(c)
	var (
		follow            *relagrpc.FollowingsReply
		pgc               *pgcFollowGrpc.MyRelationsReply
		cheese            *dynCheeseGrpc.MyPaidReply
		ugcSeason         *favgrpc.BatchFavsReply
		batchListFavorite []int64
	)
	eg.Go(func(ctx context.Context) error {
		var err error
		follow, err = s.accountDao.Followings(ctx, mid)
		if err != nil {
			xmetric.DyanmicRelationAPI.Inc("用户关注", "request_error")
			log.Warn("accountDao.Followings mid %v error %v", mid, err)
			return err
		}
		return nil
	})
	if mid > 0 {
		eg.Go(func(ctx context.Context) error {
			var err error
			pgc, err = s.pgcDao.MyRelations(ctx, mid, withAnime, withCinema)
			if err != nil {
				xmetric.DyanmicRelationAPI.Inc("追番", "request_error")
				log.Warn("pgcDao.MyRelations mid %v error %v", mid, err)
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) error {
		var err error
		cheese, err = s.cheeseDao.MyPaid(ctx, mid)
		if err != nil {
			xmetric.DyanmicRelationAPI.Inc("付费课程", "request_error")
			log.Warn("cheeseDao.MyPaid mid %v error %v", mid, err)
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		var err error
		ugcSeason, err = s.favDao.UGCSeasonRelations(ctx, mid)
		if err != nil {
			xmetric.DyanmicRelationAPI.Inc("合集订阅关系", "request_error")
			log.Warn("favDao.UGCSeasonRelations mid %v error %v", mid, err)
		}
		return nil
	})
	if (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynMatchIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynMatchAndroid) &&
		mid > 0 {
		eg.Go(func(ctx context.Context) error {
			var err error
			batchListFavorite, err = s.comicDao.ListFavorite(ctx, mid)
			if err != nil {
				log.Warn("comicDao.ListFavorite mid %v error %v", mid, err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, nil, nil, nil, nil, err
	}
	return follow, pgc, cheese, ugcSeason, batchListFavorite, nil
}

// getMaterialIDs 聚合物料ID
// nolint:gocognit
func (s *Service) getMaterialIDs(dyn *mdlv2.Dynamic, ret *mdlv2.DynamicContext, general *mdlv2.GeneralParam, adExtra string, aidm, storyAidm map[int64]map[int64]struct{}, midm, epIDm, cheeseBatchIDm, cheeseSeasonIDm,
	dynIDm, drawIDm, articlIDm, musicIDm, commonIDm, liveIDm, medialistIDm, adIDm, appletIDm, subIDm, liveRcmdIDm, ugcSeasonIDm, mangaIDm, pugvIDm, matchIDm, gameIDm, gameActIDm, voteIDm, biliCutIDm,
	_, additionalTopicIDm, bbqExtend, decorationIDm, subNewIDm, officActivityIDm, additionalUpIDm, additionalUpActivityIDm map[int64]struct{}, likeIDm map[string][]*mdlv2.ThumbsRecord, replyIDm map[string]struct{},
	dynamicActivityArgs map[int64]*dynactivitygrpc.DynamicAttachedPromoInfo, additionalTopic map[int64][]*mdlv2.Topic, goods map[int64]*bcgmdl.GoodsParams, entryLiveUidm, creativeIDm map[int64]struct{}, dramaIDm map[int64]struct{},
	batchIDm, batchIDUid map[int64]struct{}, premiereAids map[int64]map[int64]struct{}, shoppingIDs map[int64]struct{}, newTopicSetIDm map[int64]int64, mantianxinIds map[int64]struct{}, ipInfoMap map[string]struct{}, userIPFrequentM map[int64]struct{}) {
	// 评论业务ID聚合
	if dyn.GetReplyID() != "" {
		replyIDm[dyn.GetReplyID()] = struct{}{}
	}
	// 点赞业务ID聚合
	if busParam, busType, isThum := dyn.GetLikeID(); isThum {
		likeIDm[busType] = append(likeIDm[busType], busParam)
	}
	// 基础信息：用户
	if dyn.UID != 0 {
		midm[dyn.UID] = struct{}{}
	}
	// 只在开启IP功能显示时才收集相关IP信息
	if s.appFeatureGate.UserIPDisplay().Enabled() {
		// 动态发布IP
		ipInfoMap[dyn.Property.GetCreateIp()] = struct{}{}
		// 只获取动态卡片上写明是UID的用户id
		if dyn.UIDType == int(dyncommongrpc.DynUidType_DYNAMIC_UID_UP) && dyn.UID > 0 {
			userIPFrequentM[dyn.UID] = struct{}{}
		}
	}
	// 核心物料ID
	if dyn.Rid != 0 {
		// 类型判断
		if dyn.IsAv() {
			// 首映召回
			if dyn.Property != nil && dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_PREMIERE_RESERVE &&
				(general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynPropertyIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynPropertyAndroid) {
				premiereAids[dyn.Rid] = make(map[int64]struct{})
			} else {
				if dyn.SType == mdlv2.VideoStypeDynamicStory {
					if _, ok := aidm[dyn.Rid]; !ok {
						storyAidm[dyn.Rid] = make(map[int64]struct{})
					}
				} else {
					if _, ok := aidm[dyn.Rid]; !ok {
						aidm[dyn.Rid] = make(map[int64]struct{})
					}
				}
			}
			// get cid from dynExt
			// 获取当前UGC稿件已经收入到ogv视频的详细信息
			if dyn.PassThrough != nil && dyn.PassThrough.PgcBadge != nil && dyn.PassThrough.PgcBadge.EpisodeId > 0 {
				epIDm[dyn.PassThrough.PgcBadge.EpisodeId] = struct{}{}
			}
		}
		if dyn.IsPGC() {
			epIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsCheeseBatch() {
			cheeseBatchIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsWord() {
			dynIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsForward() {
			dynIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsDraw() {
			drawIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsArticle() {
			articlIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsMusic() {
			musicIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsCommon() {
			commonIDm[dyn.Rid] = struct{}{}
		}
		// 转发聚合了直播推荐、付费系列、播单
		if dyn.IsLive() {
			liveIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsCheeseSeason() || dyn.IsCourUp() {
			cheeseSeasonIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsMedialist() {
			medialistIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsAD() {
			adIDm[dyn.Rid] = struct{}{}
			creativeIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsApplet() {
			appletIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsSubscription() {
			subIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsLiveRcmd() {
			liveRcmdIDm[dyn.Rid] = struct{}{}
			if dyn.Property != nil {
				if (dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE) || (dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_HISTORY) {
					entryLiveUidm[dyn.UID] = struct{}{}
				}
			}
		}
		if dyn.IsUGCSeason() {
			ugcSeasonIDm[dyn.UID] = struct{}{}
			if _, ok := aidm[dyn.Rid]; !ok {
				aidm[dyn.Rid] = make(map[int64]struct{})
			}
		}
		if dyn.IsUGCSeasonShare() {
			ugcSeasonIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsSubscriptionNew() {
			subNewIDm[dyn.Rid] = struct{}{}
		}
		if dyn.IsAD() {
			if dyn.PassThrough != nil {
				if dyn.PassThrough.AdverMid > 0 {
					midm[dyn.PassThrough.AdverMid] = struct{}{}
				}
				if dyn.PassThrough.AdContentType == _adContentFly && dyn.PassThrough.AdAvid > 0 {
					aidm[dyn.PassThrough.AdAvid] = make(map[int64]struct{})
				}
			}
		}
		if dyn.IsBatch() {
			batchIDm[dyn.Rid] = struct{}{}
			batchIDUid[dyn.UID] = struct{}{}
		}
		if dyn.IsNewTopicSet() && dyn.GetNewTopicSetPushId() != 0 {
			// push id -> topic set id
			newTopicSetIDm[dyn.GetNewTopicSetPushId()] = dyn.UID
		}
	}
	// 帮推
	tmpDynamicAttachedPromoInfo := &dynactivitygrpc.DynamicAttachedPromoInfo{
		DynamicId: dyn.DynamicID,
		Rid:       dyn.Rid,
		Type:      dyn.Type,
	}
	dynamicActivityArgs[dyn.DynamicID] = tmpDynamicAttachedPromoInfo
	// 附加大卡
	for _, v := range dyn.AttachCardInfos {
		if v.Rid == 0 {
			continue
		}
		// nolint:exhaustive
		switch v.CardType {
		case dyncommongrpc.AttachCardType_ATTACH_CARD_VOTE:
			// 投票
			voteIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MATCH:
			// 赛事
			matchIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MANGA:
			// 漫画
			mangaIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_PUGV:
			// 课程
			pugvIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_GAME:
			// 游戏
			gameIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_OGV, dyncommongrpc.AttachCardType_ATTACH_CARD_AUTO_OGV:
			epIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_DECORATION:
			// 商品
			decorationIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_OFFICIAL_ACTIVITY:
			// 普通活动
			officActivityIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UGC:
			if _, ok := aidm[v.Rid]; !ok {
				aidm[v.Rid] = make(map[int64]struct{})
			}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_TOPIC, dyncommongrpc.AttachCardType_ATTACH_CARD_UP_TOPIC:
			additionalTopicIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
			// UP主预约卡
			additionalUpIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UP_ACTIVITY:
			// UP发布的活动
			additionalUpActivityIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_UP_MAOER:
			// 猫儿
			dramaIDm[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MAN_TIAN_XING:
			mantianxinIds[v.Rid] = struct{}{}
		case dyncommongrpc.AttachCardType_ATTACH_CARD_MEMBER_GOODS:
			// 会员购
			shoppingIDs[v.Rid] = struct{}{}
		}
	}
	// 附加小卡
	for _, v := range dyn.Tags {
		// nolint:exhaustive
		switch v.TagType {
		case dyncommongrpc.TagType_TAG_DIVERSION:
			biliCutIDm[v.Rid] = struct{}{}
		case dyncommongrpc.TagType_TAG_BBQ:
			bbqExtend[dyn.DynamicID] = struct{}{}
		case dyncommongrpc.TagType_TAG_AUTOOGV, dyncommongrpc.TagType_TAG_OGV:
			epIDm[v.Rid] = struct{}{}
		case dyncommongrpc.TagType_TAG_GAME, dyncommongrpc.TagType_TAG_GAME_SDK:
			// 需要展示行动点的时候在请求游戏接口
			if v.ActionPoint == 1 {
				gameActIDm[v.Rid] = struct{}{}
			}
		case dyncommongrpc.TagType_TAG_GAME_CARD_CONVERT:
			gameIDm[v.Rid] = struct{}{}
			gameActIDm[v.Rid] = struct{}{}
		}
	}
	ret.ResExtendBBQ = bbqExtend
	// 扩展部分
	if dyn.Extend != nil {
		// 附加内容判断
		if ok, likeUser := dyn.GetLikeUser(); ok { // 好友点赞
			for _, uid := range likeUser {
				if uid != 0 {
					midm[uid] = struct{}{}
				}
			}
		}
		// 话题
		if dyn.Extend.TopicInfo != nil {
			for _, topic := range dyn.Extend.TopicInfo.TopicInfos {
				additionalTopic[topic.TopicID] = append(additionalTopic[topic.TopicID], topic)
			}
			ret.ResAdditionalTopic = additionalTopic
		}
		// 新话题 透传
		if dyn.Extend.NewTopic != nil {
			if ret.ResNewTopic == nil {
				ret.ResNewTopic = make(map[int64]*mdlv2.NewTopicHeader)
			}
			ret.ResNewTopic[dyn.DynamicID] = dyn.Extend.NewTopic
		}
		// 商品卡
		if dyn.Extend.OpenGoods != nil {
			goodsCtx := &bcgmdl.GoodsCtx{
				AdExtra:    adExtra,
				Build:      general.GetBuild(),
				Buvid:      general.GetBuvid(),
				IP:         general.IP,
				PlatformID: bcgmdl.TranPlatformID(general.GetPlatform()),
			}
			goodsCtxStr, _ := json.Marshal(goodsCtx)
			inputStr, _ := json.Marshal(dyn.Extend.OpenGoods)
			tmp := &bcgmdl.GoodsParams{
				Uid:         general.Mid,
				UpUid:       dyn.UID,
				DynamicID:   dyn.DynamicID,
				Ctx:         string(goodsCtxStr),
				InputExtend: string(inputStr),
			}
			goods[dyn.DynamicID] = tmp
		}
		if dyn.Extend.VideoShare != nil && dyn.IsAv() && dyn.Rid != 0 {
			if dyn.Property != nil && dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_PREMIERE_RESERVE {
				premiereAids[dyn.Rid][dyn.Extend.VideoShare.CID] = struct{}{}
			} else {
				if dyn.SType == mdlv2.VideoStypeDynamicStory {
					storyAidm[dyn.Rid][dyn.Extend.VideoShare.CID] = struct{}{}
				} else {
					aidm[dyn.Rid][dyn.Extend.VideoShare.CID] = struct{}{}
				}
			}
		}
		// 校园同学点赞信息加入视频大卡
		if dyn.GetCampusLike() != nil {
			for _, uid := range dyn.GetCampusLike() {
				// 统计所有涉及的点赞人
				midm[uid] = struct{}{}
			}
		}
	}
}

type getMaterialOption struct {
	general           *mdlv2.GeneralParam
	dynamics          []*mdlv2.Dynamic
	rcmdUps           *mdlv2.RcmdUPCard
	upRegionRcmds     *dyngrpc.UnLoginRsp
	mixUpList         *dyngrpc.MixUpListRsp
	reserves          []*activitygrpc.UpActReserveRelationInfo
	playurlParam      *api.PlayurlParam
	fold              *mdlv2.FoldInfo
	channelIDs        []int64
	requestID         string
	adExtra           string
	campusRcmd        *dyncampusgrpc.RcmdCampusInfo
	storyRcmd         *dyngrpc.StoryUPCard
	homeRecommendItem []*dyncampusgrpc.RecommendItem
	homeAiRcmd        []*mdlv2.RcmdItem
}

// getMaterial 并发请求物料
// nolint:gocognit
func (s *Service) getMaterial(c context.Context, opt getMaterialOption) (*mdlv2.DynamicContext, error) {
	var (
		aidm                            = make(map[int64]map[int64]struct{})
		storyAidm                       = make(map[int64]map[int64]struct{})
		midm                            = make(map[int64]struct{})
		epIDm                           = make(map[int64]struct{})
		cheeseBatchIDm                  = make(map[int64]struct{})
		cheeseSeasonIDm                 = make(map[int64]struct{})
		dynIDm                          = make(map[int64]struct{})
		drawIDm                         = make(map[int64]struct{})
		articlIDm                       = make(map[int64]struct{})
		musicIDm                        = make(map[int64]struct{})
		commonIDm                       = make(map[int64]struct{})
		liveIDm                         = make(map[int64]struct{})
		medialistIDm                    = make(map[int64]struct{})
		adIDm                           = make(map[int64]struct{})
		appletIDm                       = make(map[int64]struct{})
		subIDm                          = make(map[int64]struct{})
		liveRcmdIDm                     = make(map[int64]struct{})
		ugcSeasonIDm                    = make(map[int64]struct{})
		mangaIDm                        = make(map[int64]struct{})
		pugvIDm                         = make(map[int64]struct{})
		matchIDm                        = make(map[int64]struct{})
		gameIDm                         = make(map[int64]struct{})
		gameActIDm                      = make(map[int64]struct{})
		voteIDm                         = make(map[int64]struct{})
		biliCutIDm                      = make(map[int64]struct{})
		topicIDm                        = make(map[int64]struct{})
		additionalTopicIDm              = make(map[int64]struct{})
		bbqExtend                       = make(map[int64]struct{})
		decorationIDm                   = make(map[int64]struct{})
		subNewIDm                       = make(map[int64]struct{})
		officActivityIDm                = make(map[int64]struct{})
		additionalActivityIDm           = make(map[int64]struct{})
		additionalUpIDm                 = make(map[int64]struct{})
		additionalUpActivityIDm         = make(map[int64]struct{})
		attachedPromo                   = make(map[int64]int64)
		likeIDm                         = map[string][]*mdlv2.ThumbsRecord{}
		replyIDm                        = make(map[string]struct{})
		dynamicActivityArgs             = make(map[int64]*dynactivitygrpc.DynamicAttachedPromoInfo)
		additionalTopic                 = make(map[int64][]*mdlv2.Topic)
		goods                           = make(map[int64]*bcgmdl.GoodsParams)
		ugcSeasonAids, upAdditionalAids []int64
		shareReq                        *shareApi.BusinessChannelsReq
		uplivemid                       = make(map[int64][]string)
		liveAidm                        = make(map[int64]struct{})
		entryLiveUidm                   = make(map[int64]struct{})
		storyArchive                    map[int64]*archivegrpc.ArcPlayer
		dynArchive                      map[int64]*archivegrpc.ArcPlayer
		arcPartCidM                     map[int64]int64 // cid -> aid
		creativeIDm                     = make(map[int64]struct{})
		reservesDynIDm                  = make(map[int64]struct{})
		dramaIDm                        = make(map[int64]struct{})
		batchIDm                        = make(map[int64]struct{})
		batchIDUid                      = make(map[int64]struct{})
		shoppingIDs                     = make(map[int64]struct{})
		playCountIDm                    = make(map[int64]int64)
		playCountIDs                    = make(map[int64]int64)
		premiereAidm                    = make(map[int64]map[int64]struct{})
		nftIdm                          = make(map[string]struct{})
		mantianxinIds                   = make(map[int64]struct{})
		newTopicSetIDm                  = make(map[int64]int64) // key 是push id， val是topic set id
		ipInfoMap                       = make(map[string]struct{})
		userIPFrequentM                 = make(map[int64]struct{})
		// 鸽子蛋需要的物料
		relationReply *activitygrpc.UpActReserveRelationInfoReply
	)
	ret := &mdlv2.DynamicContext{}
	// 初始化灰度策略
	if s.c.Grayscale != nil {
		grayTmp := make(map[string]int)
		if s.c.Grayscale.UplistMore != nil && s.c.Grayscale.UplistMore.Key != "" {
			grayTmp[s.c.Grayscale.UplistMore.Key] = s.c.Grayscale.UplistMore.GrayCheck(opt.general.Mid, opt.general.GetBuvid())
		}
		if s.c.Grayscale.ShowInPersonal != nil && s.c.Grayscale.ShowInPersonal.Key != "" {
			grayTmp[s.c.Grayscale.ShowInPersonal.Key] = s.c.Grayscale.ShowInPersonal.GrayCheck(opt.general.Mid, opt.general.GetBuvid())
		}
		if s.c.Grayscale.ShowPlayIcon != nil && s.c.Grayscale.ShowPlayIcon.Key != "" {
			grayTmp[s.c.Grayscale.ShowPlayIcon.Key] = s.c.Grayscale.ShowPlayIcon.GrayCheck(opt.general.Mid, opt.general.GetBuvid())
		}
		if s.c.Grayscale.Relation != nil && s.c.Grayscale.Relation.Key != "" {
			grayTmp[s.c.Grayscale.Relation.Key] = s.c.Grayscale.Relation.GrayCheck(opt.general.Mid, opt.general.GetBuvid())
		}
		if len(grayTmp) > 0 {
			ret.Grayscale = grayTmp
		}
	}
	// 聚合物料ID
	for _, dyn := range opt.dynamics {
		s.getMaterialIDs(dyn, ret, opt.general, opt.adExtra, aidm, storyAidm, midm, epIDm, cheeseBatchIDm, cheeseSeasonIDm,
			dynIDm, drawIDm, articlIDm, musicIDm, commonIDm, liveIDm, medialistIDm, adIDm, appletIDm, subIDm, liveRcmdIDm, ugcSeasonIDm, mangaIDm, pugvIDm, matchIDm,
			gameIDm, gameActIDm, voteIDm, biliCutIDm, topicIDm, additionalTopicIDm, bbqExtend, decorationIDm, subNewIDm, officActivityIDm, additionalUpIDm, additionalUpActivityIDm,
			likeIDm, replyIDm, dynamicActivityArgs, additionalTopic, goods, entryLiveUidm, creativeIDm, dramaIDm, batchIDm, batchIDUid, premiereAidm, shoppingIDs, newTopicSetIDm, mantianxinIds, ipInfoMap, userIPFrequentM)
		if dyn.IsForward() {
			s.getMaterialIDs(dyn.Origin, ret, opt.general, opt.adExtra, aidm, storyAidm, midm, epIDm, cheeseBatchIDm, cheeseSeasonIDm,
				dynIDm, drawIDm, articlIDm, musicIDm, commonIDm, liveIDm, medialistIDm, adIDm, appletIDm, subIDm, liveRcmdIDm, ugcSeasonIDm, mangaIDm,
				pugvIDm, matchIDm, gameIDm, gameActIDm, voteIDm, biliCutIDm, topicIDm, additionalTopicIDm, bbqExtend, decorationIDm, subNewIDm, officActivityIDm,
				additionalUpIDm, additionalUpActivityIDm, likeIDm, replyIDm, dynamicActivityArgs, additionalTopic, goods, entryLiveUidm, creativeIDm, dramaIDm, batchIDm, batchIDUid, premiereAidm, shoppingIDs, newTopicSetIDm, mantianxinIds, ipInfoMap, userIPFrequentM)
		}
		// 详情页分享组件
		shareReq = s.shareReqParam(opt.general, dyn)
	}
	// 综合页最常访问
	if opt.mixUpList != nil {
		for _, item := range opt.mixUpList.List {
			if item == nil || item.UserProfile == nil {
				continue
			}
			if item.UserProfile.Uid != 0 {
				midm[item.UserProfile.Uid] = struct{}{}
			}
		}
	}
	// 推荐用户
	if opt.rcmdUps != nil {
		for _, item := range opt.rcmdUps.Users {
			if item.Uid != 0 {
				midm[item.Uid] = struct{}{}
			}
		}
	}
	// UP主预约
	if len(opt.reserves) > 0 {
		for _, v := range opt.reserves {
			dynID, _ := strconv.ParseInt(v.DynamicId, 10, 64)
			if dynID > 0 {
				reservesDynIDm[dynID] = struct{}{}
			}
			// nolint:exhaustive
			switch v.Type {
			case activitygrpc.UpActReserveRelationType_Archive: // 稿件
				aid, _ := strconv.ParseInt(v.Oid, 10, 64)
				if aid > 0 {
					upAdditionalAids = append(upAdditionalAids, aid)
				}
			case activitygrpc.UpActReserveRelationType_Live: // 直播
				liveID, _ := strconv.ParseInt(v.Oid, 10, 64)
				if liveID > 0 {
					// 直播预约卡
					if uplivemid == nil {
						uplivemid = map[int64][]string{}
					}
					uplivemid[v.Upmid] = append(uplivemid[v.Upmid], v.Oid)
				}
			case activitygrpc.UpActReserveRelationType_Premiere: // 首映
				aid, _ := strconv.ParseInt(v.Oid, 10, 64)
				if aid > 0 {
					premiereAidm[aid] = map[int64]struct{}{}
				}
			}
		}
	}
	// UP主分区动态
	if opt.upRegionRcmds != nil {
		for _, v := range opt.upRegionRcmds.RegionUps {
			for _, upRcmd := range v.UpVideos {
				if upRcmd.Uid != 0 {
					midm[upRcmd.Uid] = struct{}{}
				}
				for _, aid := range upRcmd.AvIds {
					if aid != 0 {
						aidm[aid] = make(map[int64]struct{})
					}
				}
			}
		}
	}
	// 校园动态推荐
	if opt.campusRcmd != nil {
		for _, value := range opt.campusRcmd.List {
			for _, aid := range value.Aids {
				if aid != 0 {
					aidm[int64(aid)] = make(map[int64]struct{})
				}
			}
		}
	}
	// 校园混排推荐流（其他院校/未开放页的推荐feed）
	if len(opt.homeRecommendItem) > 0 {
		for _, v := range opt.homeRecommendItem {
			if v.Aid != 0 {
				aidm[v.Aid] = make(map[int64]struct{})
			}
		}
	}
	// AI校园混排推荐流（其他院校/未开放页的推荐feed）
	if len(opt.homeAiRcmd) > 0 {
		for _, v := range opt.homeAiRcmd {
			switch v.Goto {
			case "av":
				if v.ID != 0 {
					aidm[v.ID] = make(map[int64]struct{})
				}
			case "dynamic":
				if v.ID != 0 {
					drawIDm[v.ID] = struct{}{}
				}
			}
			// 暂时不用
			//if v.UpID != 0 {
			//	midm[v.UpID] = struct{}{}
			//}
		}
	}
	if opt.storyRcmd != nil {
		for _, v := range opt.storyRcmd.StoryUps {
			storyAidm[v.Rid] = make(map[int64]struct{})
			midm[v.Uid] = struct{}{}
		}
	}
	// 并发请求物料
	var midrw = sync.RWMutex{}
	eg := errgroup.WithCancel(c)
	if len(epIDm) != 0 {
		var (
			epIDsInt32 []int32
			epIDs      []int64
		)
		for epid := range epIDm {
			epIDsInt32 = append(epIDsInt32, int32(epid))
			epIDs = append(epIDs, epid)
		}
		// PGC详情
		eg.Go(func(ctx context.Context) error {
			res, err := s.pgcDao.EpList(ctx, epIDsInt32, opt.general, opt.playurlParam)
			if err != nil {
				log.Error("getMaterial mid(%v) EpList(%v), err %v", opt.general.Mid, epIDsInt32, err)
				return nil
			}
			for _, pgc := range res {
				if pgc == nil {
					continue
				}
				if replyID := mdlv2.GetPGCReplyID(pgc); replyID != "" {
					replyIDm[replyID] = struct{}{}
				}
				if likeParam, likeType, isLike := mdlv2.GetPGCLikeID(pgc); isLike {
					likeIDm[likeType] = append(likeIDm[likeType], likeParam)
				}
			}
			ret.ResPGC = res
			return nil
		})
		// 追番附加卡
		eg.Go(func(ctx context.Context) error {
			res, err := s.pgcDao.FollowCard(ctx, opt.general.GetMobiApp(), opt.general.GetDevice(), opt.general.GetPlatform(), opt.general.Mid, int(opt.general.GetBuild()), epIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/pgc.service.dynamic.v1.DynamicService/FollowCard", "request_error")
				log.Error("getMaterial mid(%v) FollowCard(%v), err %v", opt.general.Mid, epIDs, err)
				return nil
			}
			ret.ResAdditionalOGV = res
			return nil
		})
	}
	if len(midm) != 0 {
		var mids []int64
		for mid := range midm {
			mids = append(mids, mid)
		}
		// 装扮信息
		eg.Go(func(ctx context.Context) error {
			res, err := s.accountDao.DecorateCards(ctx, mids)
			if err != nil {
				log.Error("getMaterial mid(%v) DecorateCards(%v), err %v", opt.general.Mid, mids, err)
				return nil
			}
			ret.ResMyDecorate = res
			return nil
		})
		// 直播信息
		eg.Go(func(ctx context.Context) error {
			live, playURl, err := s.liveDao.LiveInfos(ctx, mids, opt.general)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/live.xroom.v1.Room/getMultipleByUids", "request_error")
				log.Error("getMaterial mid(%v) LiveInfos(%v), err %v", opt.general.Mid, mids, err)
				return nil
			}
			ret.ResUserLive = live
			ret.ResUserLivePlayUrl = playURl
			return nil
		})
		// 批量获取nft信息
		eg.Go(func(ctx context.Context) error {
			reply, err := s.accountDao.NFTBatchInfo(ctx, mids)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			for _, v := range reply {
				if v.NftId == "" {
					continue
				}
				nftIdm[v.NftId] = struct{}{}
			}
			ret.ResNFTBatchInfo = reply
			return nil
		})
	}
	// 批量获取IP到属地转换信息
	if len(ipInfoMap) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			ret.ResIP2Loc, err = s.loc.IP2Location(ctx, ipInfoMap)
			if err != nil {
				log.Errorc(ctx, "error doing loc.IP2Location: %v", err)
			}
			return nil
		})
	}
	if len(userIPFrequentM) > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			ret.ResUserFreqLocation, err = s.accountDao.UserFrequentLoc(ctx, userIPFrequentM)
			if err != nil {
				log.Errorc(ctx, "error doing accountDao.UserFrequentLoc: %v", err)
			}
			return nil
		})
	}

	if s.appFeatureGate.UserIPDisplay().Enabled(c) {
		// 获取固定IP的用户信息
		eg.Go(func(ctx context.Context) error {
			ret.ResUserFixedLocation = s.accountDao.FixedUserLocation(ctx)
			return nil
		})
		// 获取管理平台指定的IP显示信息
		eg.Go(func(ctx context.Context) error {
			ret.ResManagerIpDisplay = s.dynDao.ManagerIpDisplay(ctx)
			return nil
		})
	}
	// 付费更新卡信息
	if len(cheeseBatchIDm) != 0 {
		eg.Go(func(ctx context.Context) error {
			var batchIDs []int64
			for batchID := range cheeseBatchIDm {
				batchIDs = append(batchIDs, batchID)
			}
			res, err := s.pgcDao.PGCBatch(ctx, batchIDs, opt.general)
			if err != nil {
				log.Error("getMaterial mid(%v) PGCBatch(%v), err %v", opt.general.Mid, batchIDs, err)
				return nil
			}
			for _, v := range res {
				if v.UpID == 0 {
					continue
				}
				midrw.Lock()
				midm[v.UpID] = struct{}{}
				midrw.Unlock()
			}
			ret.ResCheeseBatch = res
			return nil
		})
	}
	// 付费系列卡信息
	if len(cheeseSeasonIDm) != 0 {
		var seasonIDs []int64
		eg.Go(func(ctx context.Context) error {
			for seasonID := range cheeseSeasonIDm {
				seasonIDs = append(seasonIDs, seasonID)
			}
			res, err := s.pgcDao.PGCSeason(ctx, seasonIDs, opt.general)
			if err != nil {
				log.Error("getMaterial mid(%v) PGCSeason(%v), err %v", opt.general.Mid, seasonIDs, err)
				return nil
			}
			for _, re := range res {
				if re == nil || re.UpID == 0 {
					continue
				}
				midrw.Lock()
				midm[re.UpID] = struct{}{}
				midrw.Unlock()
			}
			ret.ResCheeseSeason = res
			return nil
		})
	}
	// 动态文案(纯文字卡、转发卡等)
	if len(dynIDm) > 0 {
		var dynIDs []int64
		for id := range dynIDm {
			dynIDs = append(dynIDs, id)
		}
		eg.Go(func(ctx context.Context) error {
			res, err := s.dynDao.ListWordText(ctx, opt.general.Mid, dynIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/dynamic.service.feed.svr.v1.Feed/ListWordText", "request_error")
				log.Error("getMaterial mid(%v) ListWordText(%v), error %v", opt.general.Mid, dynIDs, err)
				return nil
			}
			ret.ResWords = res
			return nil
		})
	}
	// 图文动态
	if len(drawIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			drawIDs := make([]int64, 0, len(drawIDm))
			for id := range drawIDm {
				drawIDs = append(drawIDs, id)
			}
			res, err := s.dynDao.DrawDetails(ctx, opt.general, drawIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) DrawDetails(%v), err %v", opt.general.Mid, drawIDs, err)
				return nil
			}
			ret.ResDraw = res
			return nil
		})
	}
	// 专栏
	if len(articlIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var articleIDs []int64
			for id := range articlIDm {
				articleIDs = append(articleIDs, id)
			}
			res, err := s.articleDao.ArticleMetas(ctx, articleIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/article.service.ArticleGRPC/ArticleMetas", "request_error")
				log.Error("getMaterial mid(%v) ArticleMetasMc(%v), err %v", opt.general.Mid, articleIDs, err)
				return nil
			}
			ret.ResArticle = res
			return nil
		})
	}
	// 音频
	if len(musicIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var musicIDs []int64
			for id := range musicIDm {
				musicIDs = append(musicIDs, id)
			}
			res, err := s.musicDao.AudioDetail(ctx, musicIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) AudioDetail(%v), err %v", opt.general.Mid, musicIDs, err)
				return nil
			}
			ret.ResMusic = res
			return nil
		})
	}
	// 通用卡
	if len(commonIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var commonIDs []int64
			for id := range commonIDm {
				commonIDs = append(commonIDs, id)
			}
			res, err := s.dynDao.CommonInfos(ctx, commonIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) CommonInfos(%v), err %v", opt.general.Mid, commonIDs, err)
				return nil
			}
			for _, re := range res {
				if re == nil || re.User == nil {
					continue
				}
				midrw.Lock()
				midm[re.User.UID] = struct{}{}
				midrw.Unlock()
			}
			ret.ResCommon = res
			return nil
		})
	}
	// 播单卡
	if len(medialistIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var medialistIDs []int64
			for id := range medialistIDm {
				medialistIDs = append(medialistIDs, id)
			}
			res, err := s.medialistDao.FavoriteDetail(ctx, medialistIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) FavoriteDetail(%v), err %v", opt.general.Mid, medialistIDs, err)
				return nil
			}
			ret.ResMedialist = res
			return nil
		})
	}
	// 广告卡
	if len(adIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var adIDs []int64
			for id := range adIDm {
				adIDs = append(adIDs, id)
			}
			// 广告 如果是外层卡片且动态类型为广告，那么是原卡
			// 其他情况都视为转发卡
			cardStatus := "other"
			if opt.general.Pattern == "outer" {
				cardStatus = "origin"
			}
			res, err := s.adDao.DynamicAdInfo(ctx, opt.general.Mid, adIDs, opt.general.GetBuildStr(), opt.general.GetBuvid(), opt.general.GetMobiApp(), opt.adExtra, opt.requestID, bcgmdl.SetAdDevice(opt.general.GetMobiApp(), opt.general.GetDevice()), cardStatus, opt.general.AdFrom)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/bcg.sunspot.ad.api.Sunspot/DynamicAdInfo", "request_error")
				log.Error("getMaterial mid(%v) DynamicAdInfo(%v), err %v", opt.general.Mid, adIDs, err)
				return nil
			}
			// 拿用户mid
			for _, v := range res {
				if v.AdverMid > 0 {
					midrw.Lock()
					midm[v.AdverMid] = struct{}{}
					midrw.Unlock()
				}
			}
			ret.ResAD = res
			return nil
		})
	}
	if len(creativeIDm) > 0 && s.c.Resource.CreativeIDmToAvid {
		eg.Go(func(ctx context.Context) error {
			var creativeIDs []int64
			for id := range creativeIDm {
				creativeIDs = append(creativeIDs, id)
			}
			res, err := s.adDao.Creatives(ctx, creativeIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) Creatives(%v), err %v", opt.general.Mid, creativeIDs, err)
				return nil
			}
			for _, aid := range res {
				if aid == 0 {
					continue
				}
				aidm[aid] = make(map[int64]struct{})
			}
			ret.ResCreativeIDs = res
			return nil
		})
	}
	// 小程序卡
	if len(appletIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var appleIDs []int64
			for id := range appletIDm {
				appleIDs = append(appleIDs, id)
			}
			res, err := s.dynDao.DyncApplet(ctx, opt.general.Mid, appleIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/dynamic.service.feed.svr.v1.Feed/ListWidget", "request_error")
				log.Error("getMaterial mid(%v) DyncApplet(%v), err %v", opt.general.Mid, appleIDs, err)
				return nil
			}
			ret.ResApple = res
			return nil
		})
	}
	// 订阅卡
	if len(subIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var subIDs []int64
			for id := range subIDm {
				subIDs = append(subIDs, id)
			}
			res, err := s.subDao.Subscription(ctx, subIDs, opt.general.Mid)
			if err != nil {
				log.Error("getMaterial mid(%v) Subscription(%v), err %v", opt.general.Mid, subIDs, err)
				return nil
			}
			ret.ResSub = res
			return nil
		})
	}
	// 直播推荐卡
	if len(liveRcmdIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var liveRcmdIDs []uint64
			for id := range liveRcmdIDm {
				liveRcmdIDs = append(liveRcmdIDs, uint64(id))
			}
			res, err := s.liveDao.LiveRcmdInfos(ctx, opt.general.Mid, opt.general.GetBuild(), opt.general.GetPlatform(), opt.general.GetDevice(), liveRcmdIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/live.xroomfeed.v1.Dynamic/GetHistoryCardInfo", "request_error")
				log.Error("getMaterial mid(%v) LiveRcmdInfos(%v), err %v", opt.general.Mid, liveRcmdIDs, err)
				return nil
			}
			ret.ResLiveRcmd = res
			return nil
		})
	}
	// 合集卡
	if len(ugcSeasonIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var ugcSeasonIDs []int64
			for id := range ugcSeasonIDm {
				ugcSeasonIDs = append(ugcSeasonIDs, id)
			}
			res, err := s.ugcSeasonDao.Seasons(ctx, ugcSeasonIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/ugcseason.service.v1.UGCSeason/Seasons", "request_error")
				log.Error("getMaterial mid(%v) UGC Seasons(%v), err %v", opt.general.Mid, ugcSeasonIDs, err)
				return nil
			}
			for _, re := range res {
				if re == nil || re.FirstAid == 0 {
					continue
				}
				midrw.Lock()
				midm[re.Mid] = struct{}{}
				midrw.Unlock()
				ugcSeasonAids = append(ugcSeasonAids, re.FirstAid)
			}
			ret.ResUGCSeason = res
			return nil
		})
	}
	// 漫画
	if len(mangaIDm) != 0 {
		eg.Go(func(ctx context.Context) error {
			var comicIDs []int64
			for comicID := range mangaIDm {
				comicIDs = append(comicIDs, comicID)
			}
			comicRes, err := s.comicDao.Comics(ctx, opt.general.Mid, comicIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) Comics(%v), err %+v", opt.general.Mid, comicIDs, err)
				return nil
			}
			ret.ResManga = comicRes
			return nil
		})
	}
	// 付费
	if len(pugvIDm) != 0 {
		eg.Go(func(ctx context.Context) error {
			var cheeseIDs []int64
			for cheeseID := range pugvIDm {
				cheeseIDs = append(cheeseIDs, cheeseID)
			}
			cheeseRes, err := s.cheeseDao.AdditionalCheese(ctx, cheeseIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) AdditionalCheese(%v), err %+v", opt.general.Mid, cheeseIDs, err)
				return nil
			}
			ret.ResPUgv = cheeseRes
			return nil
		})
	}
	// 电竞
	if len(matchIDm) != 0 {
		var matchIDs []int64
		for id := range matchIDm {
			matchIDs = append(matchIDs, id)
		}
		eg.Go(func(ctx context.Context) error {
			resTmp, err := s.esportDao.AdditionalEsport(ctx, opt.general.Mid, matchIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/esports.service.v1.Esports/LiveContests", "request_error")
				log.Error("getMaterial mid(%v) AdditionalEsport(%v), err %+v", opt.general.Mid, matchIDs, err)
				return nil
			}
			ret.ResMatch = resTmp
			return nil
		})
	}
	// 游戏
	if len(gameIDm) != 0 {
		eg.Go(func(ctx context.Context) error {
			var gameIDs []int64
			for gameID := range gameIDm {
				gameIDs = append(gameIDs, gameID)
			}
			gameRes, err := s.gameDao.Games(ctx, opt.general.Mid, opt.general.GetPlatform(), gameIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) Games(%v), err %+v", opt.general.Mid, gameIDs, err)
				return nil
			}
			ret.ResGame = gameRes
			return nil
		})
	}
	// 游戏
	if len(gameActIDm) != 0 {
		eg.Go(func(ctx context.Context) error {
			var gameIDs []int64
			for gameID := range gameActIDm {
				gameIDs = append(gameIDs, gameID)
			}
			gameRes, err := s.gameDao.GameAction(ctx, opt.general.Mid, opt.general.GetPlatform(), gameIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) GameAction(%v), err %+v", opt.general.Mid, gameIDs, err)
				return nil
			}
			ret.ResGameAct = gameRes
			return nil
		})
	}
	// 装扮
	if len(decorationIDm) != 0 {
		eg.Go(func(ctx context.Context) error {
			var decorationIDs []int64
			for decorationID := range decorationIDm {
				decorationIDs = append(decorationIDs, decorationID)
			}
			res, err := s.garbDao.Decorations(ctx, opt.general.Mid, decorationIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/garb.service.v1.Garb/DynamicGarbInfo", "request_error")
				log.Error("getMaterial mid(%v) Decorations(%v), err %+v", opt.general.Mid, decorationIDs, err)
				return nil
			}
			ret.ResDecorate = res
			return nil
		})
	}
	// 帮推
	if len(dynamicActivityArgs) != 0 {
		eg.Go(func(ctx context.Context) error {
			// 动态接口拉取绑定tag
			var dynAttachedPromoInfos []*dynactivitygrpc.DynamicAttachedPromoInfo
			for _, tmpDynamicActivityArgs := range dynamicActivityArgs {
				dynAttachedPromoInfos = append(dynAttachedPromoInfos, tmpDynamicActivityArgs)
			}
			resTmp, err := s.dynDao.DynamicAttachedPromo(ctx, dynAttachedPromoInfos)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/dynamic.service.activity.v1.ActPromoRPC/DynamicAttachedPromo", "request_error")
				log.Error("getMaterial mid(%v) DynamicAttachedPromo(%v), err %+v", opt.general.Mid, dynamicActivityArgs, err)
				return nil
			}
			for _, value := range resTmp {
				if value != nil && value.TagId != 0 {
					attachedPromo[value.DynamicId] = value.TagId
					topicIDm[value.TagId] = struct{}{}
				}
			}
			ret.ResAttachedPromo = attachedPromo
			return nil
		})
	}
	// 必减
	if len(biliCutIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var biliCutIDs []int64
			for id := range biliCutIDm {
				biliCutIDs = append(biliCutIDs, id)
			}
			res, err := s.videoupDao.ExtBiliCut(ctx, biliCutIDs)
			if err != nil {
				log.Error("getMaterial mid(%v) ExtBiliCut(%v), err %v", opt.general.Mid, biliCutIDs, err)
				return nil
			}
			ret.ResBiliCut = res
			return nil
		})
	}
	// 投票
	if len(voteIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			var voteIDs []int64
			for id := range voteIDm {
				voteIDs = append(voteIDs, id)
			}
			res, err := s.dynDao.Votes(ctx, opt.general.Mid, voteIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/dynamic.service.vote.svr.v1.VoteSvr/ListFeedVotes", "request_error")
				log.Error("getMaterial mid(%v) Votes(%v), err %v", opt.general.Mid, voteIDs, err)
				return nil
			}
			ret.ResVote = res
			return nil
		})
	}
	// 商品
	if len(goods) > 0 {
		res := make(map[int64]map[int]map[string]*bcgmdl.GoodsItem) // map[dynamicID]map[sourceType]map[goodsID]*goodsItem
		rw := sync.RWMutex{}
		for _, good := range goods {
			var goodParam = new(bcgmdl.GoodsParams)
			*goodParam = *good
			eg.Go(func(ctx context.Context) error {
				goodsDetail, err := s.adDao.GoodsDetails(ctx, goodParam)
				if err != nil || goodsDetail == nil {
					log.Error("getMaterial mid(%v) GoodsDetials(%v), error(%+v)", opt.general.Mid, goodParam.DynamicID, err)
					return nil
				}
				rw.Lock()
				res[goodParam.DynamicID] = goodsDetail
				rw.Unlock()
				return nil
			})
		}
		ret.ResGood = res
	}
	// 新订阅卡
	if len(subNewIDm) > 0 {
		var subNewIDs []int64
		for id := range subNewIDm {
			subNewIDs = append(subNewIDs, id)
		}
		eg.Go(func(ctx context.Context) error {
			res, err := s.subDao.Tunnel(ctx, subNewIDs, opt.general)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/tunnel.service.v1.Tunnel/DynamicCardMaterial", "request_error")
				log.Error("getMaterial mid(%v) Tunnel(%v), err %v", opt.general.Mid, subNewIDs, err)
				return nil
			}
			ret.ResSubNew = res
			return nil
		})
	}
	// 附加活动
	if len(officActivityIDm) > 0 {
		var officActivityIDs []int64
		for id := range officActivityIDm {
			officActivityIDs = append(officActivityIDs, id)
		}
		eg.Go(func(ctx context.Context) error {
			res, err := s.activityDao.ActivityRelation(ctx, opt.general.Mid, officActivityIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/activity.service.v1.Activity/ActRelationInfo", "request_error")
				log.Error("getMaterial mid(%v) ActivityRelation(%v), err %v", opt.general.Mid, officActivityIDs, err)
				return nil
			}
			ret.ResActivityRelation = res
			for _, act := range res {
				if act == nil {
					continue
				}
				if act.NativeID != 0 {
					additionalActivityIDm[act.NativeID] = struct{}{}
				}
			}
			return nil
		})
	}
	// UP主预约卡信息
	if len(additionalUpIDm) > 0 {
		var additionalUpIDs []int64
		for id := range additionalUpIDm {
			additionalUpIDs = append(additionalUpIDs, id)
		}
		eg.Go(func(ctx context.Context) (err error) {
			relationReply, err = s.activityDao.UpActReserveRelationInfo(ctx, additionalUpIDs, opt.general.Mid)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/activity.service.v1.Activity/UpActReserveRelationInfo", "request_error")
				log.Error("%+v", err)
				return nil
			}
			for k, v := range relationReply.List {
				if ret.ResUpActRelationInfo == nil {
					ret.ResUpActRelationInfo = map[int64]*activitygrpc.UpActReserveRelationInfo{}
				}
				ret.ResUpActRelationInfo[k] = v
				// nolint:exhaustive
				switch v.Type {
				case activitygrpc.UpActReserveRelationType_Archive: // 稿件
					aid, _ := strconv.ParseInt(v.Oid, 10, 64)
					if aid > 0 {
						upAdditionalAids = append(upAdditionalAids, aid)
					}
				case activitygrpc.UpActReserveRelationType_Live: // 直播
					liveID, _ := strconv.ParseInt(v.Oid, 10, 64)
					if liveID > 0 {
						// 直播预约卡
						if uplivemid == nil {
							uplivemid = map[int64][]string{}
						}
						uplivemid[v.Upmid] = append(uplivemid[v.Upmid], v.Oid)
					}
				case activitygrpc.UpActReserveRelationType_Premiere: // 首映
					aid, _ := strconv.ParseInt(v.Oid, 10, 64)
					if aid > 0 {
						premiereAidm[aid] = map[int64]struct{}{}
					}
				}
				dynID, _ := strconv.ParseInt(v.DynamicId, 10, 64)
				midrw.Lock()
				midm[v.Upmid] = struct{}{}
				reservesDynIDm[dynID] = struct{}{}
				midrw.Unlock()
			}
			return nil
		})
	}
	// 分享组件
	if shareReq != nil {
		eg.Go(func(ctx context.Context) error {
			reply, err := s.shareDao.BusinessChannels(ctx, shareReq)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ShareChannel = reply
			return nil
		})
	}
	// 垂搜频道
	if len(opt.channelIDs) > 0 {
		eg.Go(func(ctx context.Context) error {
			reply, err := s.channelDao.SearchChannelsInfo(ctx, opt.general.Mid, opt.channelIDs)
			if err != nil {
				log.Error("%v", err)
				return err
			}
			ret.ResSearchChannels = reply
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			more, err := s.channelDao.RelativeChannel(ctx, opt.general.Mid, opt.channelIDs)
			if err != nil {
				log.Error("%v", err)
				return nil
			}
			ret.ResSearchChannelMore = more
			return nil
		})
		eg.Go(func(ctx context.Context) error {
			hot, err := s.channelDao.ChannelList(ctx, opt.general.Mid, 100, "")
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			ret.ResSearchChannelHot = hot
			return nil
		})
	}
	// 猫儿
	if len(dramaIDm) > 0 {
		var dramaIDs []int64
		for v := range dramaIDm {
			dramaIDs = append(dramaIDs, v)
		}
		eg.Go(func(ctx context.Context) error {
			reply, err := s.dramaseasonDao.FeedCardDrama(ctx, dramaIDs)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResFeedCardDramaInfo = reply
			return nil
		})
	}
	// 追漫卡
	if len(batchIDm) > 0 {
		var batchIDs []int64
		for v := range batchIDm {
			batchIDs = append(batchIDs, v)
		}
		eg.Go(func(ctx context.Context) error {
			reply, err := s.comicDao.BatchInfo(ctx, opt.general.Mid, batchIDs)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResBatch = reply
			return nil
		})
	}
	// 新话题 话题集订阅卡
	if len(newTopicSetIDm) > 0 {
		eg.Go(func(ctx context.Context) error {
			res, err := s.topDao.NewTopicSetDetails(ctx, newTopicSetIDm, opt.general)
			if err != nil {
				log.Errorc(ctx, "error fetching NewTopicSetDetails: %v", err)
			}
			ret.ResNewTopicSet = res
			return nil
		})
	}

	if len(batchIDUid) > 0 {
		var batchIDs []int64
		for v := range batchIDUid {
			batchIDs = append(batchIDs, v)
		}
		eg.Go(func(ctx context.Context) error {
			reply, err := s.comicDao.IsFav(ctx, opt.general.Mid, batchIDs)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResBatchIsFav = reply
			return nil
		})
	}
	if len(mantianxinIds) > 0 {
		var ids []int64
		for v := range mantianxinIds {
			ids = append(ids, v)
		}
		eg.Go(func(ctx context.Context) error {
			reply, err := s.shopDao.ListItemCards(ctx, ids)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResManTianXinm = reply
			return nil
		})
	}
	if len(shoppingIDs) > 0 {
		var tmpIDs []int64
		for v := range shoppingIDs {
			tmpIDs = append(tmpIDs, v)
		}
		eg.Go(func(ctx context.Context) error {
			reply, err := s.shopDao.ItemCard(ctx, tmpIDs)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ShoppingItems = reply
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	// aids聚合到一个map做去重复
	ugcSeasonAids = append(ugcSeasonAids, upAdditionalAids...)
	for _, aid := range ugcSeasonAids {
		if _, ok := aidm[aid]; !ok {
			aidm[aid] = make(map[int64]struct{})
		}
	}
	for k, v := range premiereAidm {
		aidm[k] = v
		playCountIDm[k] = 0
		playCountIDs[k] = 0
		for cid := range v {
			playCountIDm[k] = cid
			playCountIDs[k] = cid
		}
	}
	for _, channel := range ret.ResSearchChannels {
		if channel == nil {
			continue
		}
		for _, video := range channel.GetVideoCards() {
			if video.GetRid() == 0 {
				continue
			}
			aidm[video.GetRid()] = make(map[int64]struct{})
		}
	}
	/*
		第二级物料获取
	*/
	eg2 := errgroup.WithCancel(c)
	// 鸽子蛋
	if relationReply != nil {
		eg2.Go(func(ctx context.Context) error {
			res, err := s.activityDao.CheckReserveDoveAct(ctx, opt.general.Mid, relationReply)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResUpActReserveDove = res
			return nil
		})
	}
	// 预约卡的动态信息
	if len(reservesDynIDm) > 0 {
		var dynIDs []int64
		for id := range reservesDynIDm {
			dynIDs = append(dynIDs, id)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.dynDao.DynSimpleInfos(ctx, dynIDs)
			if err != nil {
				log.Error("DynSimpleInfos mid(%v) DynBriefs(%v) dynIDs, err %v", opt.general.Mid, dynIDs, err)
				return nil
			}
			ret.ResDynSimpleInfos = res
			return nil
		})
	}
	// 分享直播卡
	if len(liveIDm) > 0 {
		var liveIDs []int64
		for id := range liveIDm {
			liveIDs = append(liveIDs, id)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.liveDao.EntryRoomInfo(ctx, liveIDs, []int64{}, opt.general.Mid, opt.general.GetBuild(), opt.general.GetPlatform())
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/live.xroom.v1.Room/entryRoomInfo", "request_error")
				log.Error("EntryRoomInfo mid(%v) EntryRoomInfo(%v) rooms, err %v", opt.general.Mid, liveIDs, err)
				return nil
			}
			ret.ResLive = res
			return nil
		})
	}
	// 直播推荐卡(召回)
	if len(entryLiveUidm) > 0 {
		var entryLiveUids []int64
		for id := range entryLiveUidm {
			entryLiveUids = append(entryLiveUids, id)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.liveDao.EntryRoomInfo(ctx, []int64{}, entryLiveUids, opt.general.Mid, opt.general.GetBuild(), opt.general.GetPlatform())
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/live.xroom.v1.Room/entryRoomInfo", "request_error")
				log.Error("EntryRoomInfo mid(%v) EntryRoomInfoUids(%v) uids, err %v", opt.general.Mid, entryLiveUidm, err)
				return nil
			}
			ret.ResEntryLiveUids = res
			return nil
		})
	}
	// 直播预约卡
	if len(uplivemid) > 0 {
		eg2.Go(func(ctx context.Context) error {
			res, err := s.liveDao.SessionInfo(ctx, uplivemid, opt.general)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResLiveSessionInfo = res
			// 回播aid
			for k, v := range res {
				if info, ok := v.SessionInfoPerLive[k]; ok {
					// 已关播有回放
					if info.Status == 2 && info.Bvid != "" {
						aid, _ := bvid.BvToAv(info.Bvid)
						if aid > 0 {
							liveAidm[aid] = struct{}{}
						}
					}
				}
			}
			return nil
		})
	}
	// 稿件详情
	if len(aidm) != 0 {
		var aids []*archivegrpc.PlayAv
		for aid, cidm := range aidm {
			var ap = &archivegrpc.PlayAv{Aid: aid}
			for cid := range cidm {
				ap.PlayVideos = append(ap.PlayVideos, &archivegrpc.PlayVideo{Cid: cid})
			}
			aids = append(aids, ap)
		}
		if !opt.general.CloseAutoPlay {
			eg2.Go(func(ctx context.Context) error {
				res, err := s.archiveDao.ArcsPlayer(ctx, aids, true, "")
				if err != nil {
					xmetric.DyanmicItemAPI.Inc("/archive.service.v1.Archive/ArcsWithPlayurl", "request_error")
					log.Error("getMaterial mid(%v) ArcsWithPlayurl(%v), err %v", opt.general.Mid, aids, err)
					return nil
				}
				dynArchive = res
				// 动态服务端没有返回合集UP信息 需要回填再获取
				for _, arc := range res {
					if arc != nil && arc.Arc != nil {
						midrw.Lock()
						// 获取预约首映在线人数ID
						if pcid, ok := playCountIDm[arc.Arc.Aid]; ok && pcid == 0 {
							playCountIDs[arc.Arc.Aid] = arc.Arc.FirstCid
						}
						midrw.Unlock()
						if arc.Arc.Author.Mid != 0 {
							midrw.Lock()
							midm[arc.Arc.Author.Mid] = struct{}{}
							midrw.Unlock()
						}
					}
				}
				return nil
			})
		} else {
			eg2.Go(func(ctx context.Context) error {
				var aids []int64
				for aid := range aidm {
					aids = append(aids, aid)
				}
				res, err := s.archiveDao.Archive(ctx, aids, opt.general.GetMobiApp(), opt.general.GetDevice(), opt.general.Mid, opt.general.GetPlatform())
				if err != nil {
					xmetric.DyanmicItemAPI.Inc("/archive.service.v1.Archive/Arcs", "request_error")
					log.Error("getMaterial mid(%v) Arcs(%v), err %v", opt.general.Mid, aids, err)
					return nil
				}
				if dynArchive == nil {
					dynArchive = map[int64]*archivegrpc.ArcPlayer{}
				}
				for _, arc := range res {
					dynArchive[arc.Aid] = &archivegrpc.ArcPlayer{Arc: arc}
					midrw.Lock()
					// 获取预约首映在线人数ID
					if pcid, ok := playCountIDm[arc.Aid]; ok && pcid == 0 {
						playCountIDs[arc.Aid] = arc.FirstCid
					}
					midrw.Unlock()
					// 动态服务端没有返回合集UP信息 需要回填再获取
					if arc.Author.Mid != 0 {
						midrw.Lock()
						midm[arc.Author.Mid] = struct{}{}
						midrw.Unlock()
					}
				}
				return nil
			})
		}
	}
	if len(storyAidm) != 0 {
		var aids []*archivegrpc.PlayAv
		for aid, cidm := range storyAidm {
			var ap = &archivegrpc.PlayAv{Aid: aid}
			for cid := range cidm {
				ap.PlayVideos = append(ap.PlayVideos, &archivegrpc.PlayVideo{Cid: cid})
			}
			aids = append(aids, ap)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.archiveDao.ArcsPlayer(ctx, aids, true, "story")
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/archive.service.v1.Archive/ArcsWithPlayurl", "request_error")
				log.Error("getMaterial mid(%v) ArcsWithPlayurl(%v), err %v", opt.general.Mid, aids, err)
				return nil
			}
			storyArchive = res
			// 动态服务端没有返回合集UP信息 需要回填再获取
			for _, arc := range res {
				if arc != nil && arc.Arc != nil && arc.Arc.Author.Mid != 0 {
					midrw.Lock()
					midm[arc.Arc.Author.Mid] = struct{}{}
					midrw.Unlock()
				}
			}
			return nil
		})
	}
	// 评论外露和计数
	if len(replyIDm) != 0 {
		eg2.Go(func(ctx context.Context) error {
			var replyIDs []string
			for replyID := range replyIDm {
				replyIDs = append(replyIDs, replyID)
			}
			cmtRes, err := s.cmtDao.DynamicFeed(ctx, opt.general.Mid, opt.general.GetBuvid(), replyIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/main.community.reply.v1.ReplyInterface/DynamicFeed", "request_error")
				log.Error("getMaterial mid(%v) DynamicFeed(%v), err %v", opt.general.Mid, replyIDs, err)
				return nil
			}
			ret.ResReply = cmtRes.Feed
			for _, reply := range cmtRes.Feed {
				for _, item := range reply.Replies {
					midrw.Lock()
					midm[item.Mid] = struct{}{}
					midrw.Unlock()
				}
			}
			return nil
		})
	}
	// 用户与动态的点赞关系
	if len(likeIDm) != 0 {
		eg2.Go(func(ctx context.Context) error {
			res, err := s.thumDao.MultiStats(ctx, opt.general.Mid, likeIDm)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/community.service.thumbup.v1.Thumbup/MultiStats", "request_error")
				log.Error("getMaterial mid(%v) MultiStats(%v), err %v", opt.general.Mid, likeIDm, err)
				return nil
			}
			ret.ResLike = res.Business
			return nil
		})
	}
	// 附加卡-帮推-活动信息
	if len(topicIDm) > 0 {
		var tagIDs []int64
		for tagid := range topicIDm {
			tagIDs = append(tagIDs, tagid)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.activityDao.NatInfoFromForeign(ctx, tagIDs, 1)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/natpage.interface.service.v1.NaPage/NatInfoFromForeign", "request_error")
				log.Error("getMaterial mid(%v) NatInfoFromForeign(%v), err %v", opt.general.Mid, tagIDs, err)
				return nil
			}
			ret.ResActivity = res
			return nil
		})
	}
	// 话题大卡-小卡升大卡
	if len(additionalTopicIDm) > 0 {
		var topicIDs []int64
		for topicId := range additionalTopicIDm {
			topicIDs = append(topicIDs, topicId)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.dynDao.ListTopicAdditiveCards(ctx, topicIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/dynamic.service.topic.ext.v1.TopicExt/ListTopicAdditiveCards", "request_error")
				log.Error("getMaterial mid(%v) ListTopicAdditiveCards(%v), err %v", opt.general.Mid, topicIDs, err)
				return nil
			}
			ret.ResTopicAdditiveCard = res
			return nil
		})
	}
	// 附加普通活动
	if len(additionalActivityIDm) > 0 {
		var additionalActivityIDs []int64
		for id := range additionalActivityIDm {
			additionalActivityIDs = append(additionalActivityIDs, id)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.activityDao.NativePageCards(ctx, additionalActivityIDs, opt.general)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/natpage.interface.service.v1.NaPage/NativePageCards", "request_error")
				log.Error("getMaterial mid(%v) NativePageCards(%v), err %v", opt.general.Mid, additionalActivityIDs, err)
				return nil
			}
			ret.ResNativePage = res
			return nil
		})
	}
	// 附加UP发布的活动
	if len(additionalUpActivityIDm) > 0 {
		var additionalActivityIDs []int64
		for id := range additionalUpActivityIDm {
			additionalActivityIDs = append(additionalActivityIDs, id)
		}
		eg2.Go(func(ctx context.Context) error {
			res, err := s.activityDao.NativeAllPageCards(ctx, additionalActivityIDs)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/natpage.interface.service.v1.NaPage/NativeAllPageCards", "request_error")
				log.Error("getMaterial mid(%v) NativeAllPageCards(%v), err %v", opt.general.Mid, additionalActivityIDs, err)
				return nil
			}
			for _, v := range res {
				if v.RelatedUid != 0 {
					midrw.Lock()
					midm[v.RelatedUid] = struct{}{}
					midrw.Unlock()
				}
			}
			ret.NativeAllPageCards = res
			return nil
		})
	}
	// NFT信息
	if len(nftIdm) > 0 {
		var nftIds []string
		for v := range nftIdm {
			nftIds = append(nftIds, v)
		}
		eg2.Go(func(ctx context.Context) error {
			reply, err := s.panguDao.GetNFTRegion(ctx, nftIds)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResNFTRegionInfo = reply
			return nil
		})
	}
	if err := eg2.Wait(); err != nil {
		return nil, err
	}
	// 稿件聚合
	if ret.ResArchive == nil {
		ret.ResArchive = map[int64]*archivegrpc.ArcPlayer{}
	}
	for k, v := range storyArchive {
		ret.ResArchive[k] = v
	}
	for k, v := range dynArchive {
		ret.ResArchive[k] = v
	}
	// 只有动态筛选器继续观看页面才去获取分p详细信息
	if opt.general.DynFrom == _dynFromFilterContinue {
		arcPartCidM = make(map[int64]int64)
		for aid, arc := range ret.ResArchive {
			if arc == nil {
				continue
			}
			// 单p的稿件不用去拉cid
			if arc.Arc.GetVideos() == 1 {
				continue
			}
			if cid := ret.GetArchiveAutoPlayCid(arc); cid > 0 {
				arcPartCidM[cid] = aid
			}
		}
	}
	/*
		第三级调用
	*/
	eg3 := errgroup.WithCancel(c)
	// 获取指定分p信息
	if len(arcPartCidM) > 0 {
		eg3.Go(func(ctx context.Context) error {
			ret.ResArcPart, _ = s.archiveDao.Pages(ctx, arcPartCidM)
			return nil
		})
	}
	// 目前只获取当前登录用户的profile  用于判断school信息
	if opt.general.Mid != 0 {
		eg3.Go(func(ctx context.Context) error {
			res, err := s.accountDao.ProfileWithStat3(ctx, opt.general.Mid)
			if err != nil {
				log.Warnc(ctx, "getMaterial mid(%v) Profile3 error(%v)", opt.general.Mid, err)
				// 只log  不影响主要流程
				return nil
			}
			ret.ResUserProfileStat = map[int64]*accAPI.ProfileStatReply{
				opt.general.Mid: res,
			}
			return nil
		})
	}
	if len(midm) > 0 {
		var mids []int64
		for mid := range midm {
			mids = append(mids, mid)
		}
		eg3.Go(func(ctx context.Context) error {
			res, err := s.accountDao.Cards3New(ctx, mids)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/account.service.Account/Cards3", "request_error")
				log.Warn("getMaterial mid(%v) Cards3New(%v) error(%v)", opt.general.Mid, mids, err)
				return nil
			}
			ret.ResUser = res
			return nil
		})
		eg3.Go(func(ctx context.Context) error {
			ret.ResRelation = s.accountDao.IsAttention(ctx, mids, opt.general.Mid)
			return nil
		})
		eg3.Go(func(ctx context.Context) error {
			res, err := s.relationDao.Stats(ctx, mids)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/account.service.relation.v1.Relation/Stats", "request_error")
				log.Warn("getMaterial mid(%v) Stats(%v), error(%v)", opt.general.Mid, mids, err)
				return nil
			}
			ret.ResStat = res
			return nil
		})
		eg3.Go(func(ctx context.Context) error {
			res, err := s.accountDao.Interrelations(ctx, opt.general.Mid, mids)
			if err != nil {
				xmetric.DyanmicItemAPI.Inc("/account.service.relation.v1.Relation/Interrelations", "request_error")
				log.Error("getMaterial mid(%v) Interrelations(%v), error %v", opt.general.Mid, mids, err)
				return nil
			}
			ret.ResRelationUltima = res
			return nil
		})
	}
	// 直播回放稿件
	if len(liveAidm) > 0 {
		var aids []int64
		for v := range liveAidm {
			aids = append(aids, v)
		}
		eg3.Go(func(ctx context.Context) error {
			res, err := s.archiveDao.Archive(ctx, aids, opt.general.GetMobiApp(), opt.general.GetDevice(), opt.general.Mid, opt.general.GetPlatform())
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResArcs = res
			return nil
		})
	}
	// 获取在线人数
	if len(playCountIDs) > 0 {
		eg3.Go(func(ctx context.Context) error {
			res, err := s.playurlDao.PlayOnline(ctx, playCountIDs)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			ret.ResPlayUrlCount = res
			return nil
		})
	}
	if err := eg3.Wait(); err != nil {
		return nil, err
	}
	/*
		折叠逻辑
	*/
	if opt.fold != nil {
		var (
			fold    = make(map[string]*mdlv2.FoldResItem)
			foldNum = 1
		)
		if len(opt.fold.InplaceFold) != 0 {
			for _, inplaceFold := range opt.fold.InplaceFold {
				if inplaceFold == nil {
					continue
				}
				for _, id := range inplaceFold.DynamicIDs {
					f := &mdlv2.FoldResItem{
						DynID:     id,
						FoldType:  api.FoldType_FoldTypeLimit,
						Group:     foldNum,
						Statement: inplaceFold.Statement,
					}
					fold[strconv.FormatInt(id, 10)] = f
				}
				foldNum++
			}
		}
		if len(opt.fold.FoldMgr) != 0 {
			for _, foldMgr := range opt.fold.FoldMgr {
				if foldMgr == nil {
					continue
				}
				for _, fd := range foldMgr.Folds {
					if fd == nil {
						continue
					}
					for _, id := range fd.DynamicIDs {
						f := &mdlv2.FoldResItem{
							DynID:    id,
							FoldType: mdlv2.TranFoldType(foldMgr.FoldType),
							Group:    foldNum,
						}
						fold[strconv.FormatInt(id, 10)] = f
					}
					foldNum++
				}
			}
		}
		ret.ResFolds = fold
	}
	return ret, nil
}

func (s *Service) procListReply(c context.Context, dynamics []*mdlv2.Dynamic, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam, from string) *mdlv2.FoldList {
	var foldList = &mdlv2.FoldList{}
	var logs []string
	if from == _handleTypeShare || from == _handleTypeForward {
		dynCtx.ForwardFrom = dynCtx.From // 记录转发分享前的from场景
	}
	dynCtx.From = from // 记录来源 分区逻辑
	for _, dyn := range dynamics {
		dynCtx.Dyn = dyn                                             // 原始数据
		dynCtx.DynamicItem = &api.DynamicItem{Extend: &api.Extend{}} // 聚合结果
		dynCtx.Interim = &mdlv2.Interim{}                            // 临时逻辑
		var (
			handlerList []Handler
			ok          bool
		)
		// mid > int32老版本抛弃当前卡片
		if s.checkMidMaxInt32(c, dynCtx.Dyn.UID, general) {
			continue
		}
		switch from {
		case _handleTypeVideo, _handleTypeVideoPersonal, _handleTypeDetail, _handleTypeAll,
			_handleTypeAllPersonal, _handleTypeAllFilter, _handleTypeSpace, _handleTypeSearch, _handleTypeUnLogin,
			_handleTypeSchool, _handleTypeSpaceSearchDetail, _handleTypeServerDetail, _handleTypeSchoolTopicFeed,
			_handleTypeLBS, _handleTypeLegacyTopic:
			handlerList, ok = s.getHandlerList(c, dynCtx, general)
		case _handleTypeShare:
			handlerList, ok = s.getHandlerListShare(c, dynCtx, general)
		case _handleTypeForward:
			handlerList, ok = s.getHandlerListForward(c, dynCtx, general)
		case _handleTypeFake:
			handlerList, ok = s.getHandlerListFake(c, dynCtx, general)
		case _handleTypeRepost:
			handlerList, ok = s.getHandlerRepostList(c, dynCtx, general)
		case _handleTypeLight:
			handlerList, ok = s.getHandlerListLight(c, dynCtx, general)
		}
		if !ok {
			xmetric.DynamicCardError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "handle_error")
			log.Warn("dynamic mid(%v) from(%v) getHandlerList !ok", general.Mid, from)
			continue
		}
		// 执行拼接func
		if err := s.conveyer(c, dynCtx, general, handlerList...); err != nil {
			xmetric.DynamicCardError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "conveyer_error")
			log.Warn("dynamic mid(%v) from(%v) conveyer, err %v", general.Mid, from, err)
			continue
		}
		if dynCtx.Interim.IsPassCard {
			xmetric.DynamicCardError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "pass_card")
			log.Warn("dynamic mid(%v) from(%v) IsPassCard dynid %v", general.Mid, from, dyn.DynamicID)
			continue
		}
		// 日志记录返回值顺序
		logs = append(logs, fmt.Sprintf("dynid(%v) type(%v) rid(%v)", dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid))
		// 监控上报
		if from != _handleTypeShare && from != _handleTypeForward {
			xmetric.DynamicCard.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type))
		}
		// 收割上下文中组装完成的items
		foldList.List = append(foldList.List, &mdlv2.FoldItem{
			Item: dynCtx.DynamicItem,
		})
	}
	log.Warn("dynamic mid(%d) from(%v) reply list(%v)", general.Mid, from, strings.Join(logs, "; "))
	return foldList
}

func (s *Service) getHandlerList(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) ([]Handler, bool) {
	var ret []Handler
	switch {
	case dynCtx.Dyn.IsForward():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardForward, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsAv():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardAv, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsPGC():
		ret = append(ret, s.base, s.authorPGC, s.dispute, s.dynCardPGC, s.dynCardPremiere, s.additional, s.ext, s.stat)
	case dynCtx.Dyn.IsCheeseBatch():
		ret = append(ret, s.base, s.authorCheeseBatch, s.dispute, s.dynCardCourBatch, s.stat)
	case dynCtx.Dyn.IsCourUp():
		ret = append(ret, s.base, s.authorCourUp, s.dispute, s.description, s.dynCardCourUp, s.stat)
	case dynCtx.Dyn.IsWord():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsDraw():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardDraw, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsArticle():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardArticle, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsMusic():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardMusic, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsCommon():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardCommon, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsAD():
		ret = append(ret, s.base, s.dynCardAD, s.interactionAD, s.stat)
	case dynCtx.Dyn.IsApplet():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardApplet, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsSubscription():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardSubscription, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsLiveRcmd(): // 透传
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardLiveRcmd, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsUGCSeason():
		ret = append(ret, s.base, s.authorUGCSeason, s.dispute, s.extNewTopic, s.description, s.dynCardUGCSeason, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsSubscriptionNew():
		ret = append(ret, s.base, s.author, s.dispute, s.extNewTopic, s.description, s.dynCardSubNew, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsBatch():
		ret = append(ret, s.base, s.authorBatch, s.dispute, s.extNewTopic, s.description, s.dynCardBatch, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsNewTopicSet():
		ret = append(ret, s.base, s.authorNewTopicSet, s.description, s.dynCardNewTopicSet)
	default:
		return nil, false
	}
	return ret, true
}

func (s *Service) getHandlerListForward(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) ([]Handler, bool) {
	var ret []Handler
	switch {
	case dynCtx.Dyn.IsForward():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardForward, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsAv():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardAv, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsPGC():
		ret = append(ret, s.base, s.authorShellPGC, s.dispute, s.description, s.dynCardPGC, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsCheeseBatch():
		ret = append(ret, s.base, s.authorShellCheeseBatch, s.dispute, s.description, s.dynCardCourBatch, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsCheeseSeason(): // 仅转发
		ret = append(ret, s.base, s.authorShellCheeseSeason, s.dispute, s.description, s.dynCardCourSeason, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsCourUp():
		ret = append(ret, s.base, s.authorShellCourUp, s.dispute, s.description, s.dynCardCourUp, s.statShell)
	case dynCtx.Dyn.IsLive(): // 仅转发
		ret = append(ret, s.base, s.authorShell, s.dispute, s.description, s.dynCardLive, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsMedialist(): // 仅转发
		ret = append(ret, s.base, s.authorShell, s.dispute, s.description, s.dynCardMedialist, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsWord():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsDraw():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardDraw, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsArticle():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardArticle, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsMusic():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardMusic, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsCommon():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardCommon, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsAD():
		ret = append(ret, s.base, s.dynCardADShell, s.dispute, s.ext, s.statShell)
	case dynCtx.Dyn.IsApplet():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardApplet, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsSubscription():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardSubscription, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsLiveRcmd():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardLiveRcmd, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsUGCSeason():
		ret = append(ret, s.base, s.authorShellUGCSeason, s.dispute, s.extNewTopic, s.description, s.dynCardUGCSeason, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsUGCSeasonShare(): // 仅转发
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardUGCSeasonShare, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsSubscriptionNew():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.extNewTopic, s.description, s.dynCardSubNew, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsBatch():
		ret = append(ret, s.base, s.authorShellBatch, s.dispute, s.extNewTopic, s.description, s.dynCardBatch, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	default:
		return nil, false
	}
	return ret, true
}

func (s *Service) getHandlerListShare(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) ([]Handler, bool) {
	var ret []Handler
	switch {
	case dynCtx.Dyn.IsForward():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardForward, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsAv():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardAv, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsPGC():
		ret = append(ret, s.base, s.authorShellPGC, s.dispute, s.dynCardPGC, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsCheeseBatch():
		ret = append(ret, s.base, s.authorShellCheeseBatch, s.dispute, s.dynCardCourBatch, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsCheeseSeason(): // 仅转发
		ret = append(ret, s.base, s.authorShellCheeseSeason, s.dispute, s.dynCardCourSeason, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsCourUp():
		ret = append(ret, s.base, s.authorShellCourUp, s.dispute, s.description, s.dynCardCourUp, s.statShell)
	case dynCtx.Dyn.IsLive(): // 仅转发
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardLive, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsMedialist(): // 仅转发
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardMedialist, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsWord():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.description, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsDraw():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardDraw, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsArticle():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardArticle, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsMusic():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardMusic, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsCommon():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardCommon, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsAD():
		ret = append(ret, s.base, s.dynCardADShell, s.dispute, s.ext, s.statShell)
	case dynCtx.Dyn.IsApplet():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardApplet, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsSubscription():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardSubscription, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsLiveRcmd():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.description, s.dynCardLiveRcmd, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsUGCSeason():
		ret = append(ret, s.base, s.authorShellUGCSeason, s.dispute, s.dynCardUGCSeason, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsUGCSeasonShare(): // 仅转发
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardUGCSeasonShare, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsSubscriptionNew():
		ret = append(ret, s.base, s.authorShell, s.dispute, s.dynCardSubNew, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	case dynCtx.Dyn.IsBatch():
		ret = append(ret, s.base, s.authorShellBatch, s.dispute, s.description, s.dynCardPremiere, s.additional, s.ext, s.statShell)
	default:
		return nil, false
	}
	return ret, true
}

func (s *Service) getHandlerListFake(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) ([]Handler, bool) {
	var ret []Handler
	switch {
	case dynCtx.Dyn.IsAv():
		ret = append(ret, s.baseFake, s.author, s.dispute, s.description, s.dynCardFake, s.additionalFake, s.extFake, s.interaction, s.statFake)
	case dynCtx.Dyn.IsWord():
		ret = append(ret, s.baseFake, s.author, s.dispute, s.description, s.additionalFake, s.extFake, s.interaction, s.statFake)
	case dynCtx.Dyn.IsDraw():
		ret = append(ret, s.baseFake, s.author, s.dispute, s.description, s.dynCardFake, s.additionalFake, s.extFake, s.interaction, s.statFake)
	default:
		return nil, false
	}
	return ret, true
}

func (s *Service) getHandlerRepostList(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) ([]Handler, bool) {
	var ret []Handler
	switch {
	case dynCtx.Dyn.IsForward():
		ret = append(ret, s.base, s.authorInfo, s.description, s.dynCardForward, s.statRepost)
	default:
		return nil, false
	}
	return ret, true
}

func (s *Service) getHandlerListLight(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) ([]Handler, bool) {
	var ret []Handler
	switch {
	case dynCtx.Dyn.IsAv():
		ret = append(ret, s.base, s.author, s.description, s.dynCardAv, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	case dynCtx.Dyn.IsDraw():
		ret = append(ret, s.base, s.author, s.description, s.dynCardDraw, s.dynCardPremiere, s.additional, s.ext, s.interaction, s.stat)
	default:
		return nil, false
	}
	return ret, true
}

// 详情页
func (s *Service) getHandlerView(_ context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) ([]Handler, bool) {
	var ret []Handler
	switch {
	case dynCtx.Dyn.IsForward():
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardForward, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsAv():
		switch {
		case general.IsPadHD(), general.IsPad(), general.IsAndroidHD():
			ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardAv, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
		default:
			ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardAv, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
		}
	case dynCtx.Dyn.IsPGC():
		switch {
		case general.IsPadHD(), general.IsPad(), general.IsAndroidHD():
			ret = append(ret, s.base, s.authorPGC, s.dispute, s.dynCardPGC, s.detailShareChannel, s.detailRecommend, s.buttom)
		default:
			ret = append(ret, s.base, s.authorPGC, s.dispute, s.dynCardPGC, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
		}
	case dynCtx.Dyn.IsCheeseBatch():
		ret = append(ret, s.base, s.authorCheeseBatch, s.dispute, s.dynCardCourBatch, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsCourUp():
		ret = append(ret, s.base, s.authorCourUp, s.dispute, s.description, s.dynCardCourUp, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsWord():
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsDraw():
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardDraw, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsArticle():
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardArticle, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsMusic():
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardMusic, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsCommon():
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardCommon, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsAD():
		ret = append(ret, s.base, s.top, s.dynCardADShell, s.buttom)
	case dynCtx.Dyn.IsApplet():
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardApplet, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsSubscription():
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardSubscription, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsLiveRcmd(): // 透传
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardLiveRcmd, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsUGCSeason():
		ret = append(ret, s.base, s.top, s.authorUGCSeason, s.dispute, s.extNewTopic, s.description, s.dynCardUGCSeason, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsSubscriptionNew():
		ret = append(ret, s.base, s.top, s.authorView, s.dispute, s.extNewTopic, s.description, s.dynCardSubNew, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	case dynCtx.Dyn.IsBatch():
		ret = append(ret, s.base, s.top, s.authorBatch, s.dispute, s.extNewTopic, s.description, s.dynCardBatch, s.dynCardPremiere, s.additional, s.ext, s.detailShareChannel, s.detailRecommend, s.buttom)
	default:
		return nil, false
	}
	return ret, true
}

func (s *Service) conveyer(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam, f ...Handler) error {
	for _, v := range f {
		err := v(c, dynCtx, general)
		if err != nil {
			log.Errorc(c, "Conveyer failed. dynamic: %v, error: %+v", dynCtx.Dyn.DynamicID, err)
			return err
		}
	}
	return nil
}

func (s *Service) AdditionFollow(c context.Context, cardType string, dynamicID int64, status string) error {
	return s.dynDao.AdditionFollow(c, cardType, dynamicID, status)
}

func (s *Service) procBackfill(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam, foldList *mdlv2.FoldList) {
	// 聚合回填物料
	s.BackfillGetMaterial(c, dynCtx, general)
	// 遍历回填
	for _, foldDynItem := range foldList.List {
		if foldDynItem == nil || foldDynItem.Item == nil {
			continue
		}
		s.backfill(c, dynCtx, foldDynItem.Item, general)
	}
}

// nolint:gocognit
func (s *Service) BackfillGetMaterial(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) {
	// 正文高亮回填
	eg := errgroup.WithCancel(c)
	if len(dynCtx.Emoji) > 0 {
		eg.Go(func(ctx context.Context) error {
			var emoji []string
			for item := range dynCtx.Emoji {
				emoji = append(emoji, item)
			}
			resEmoji, err := s.dynDao.GetEmoji(c, emoji)
			if err != nil {
				log.Error("BackfillGetMaterial mid(%v) GetEmoji(%v), error %v", general.Mid, emoji, err)
				return err
			}
			dynCtx.ResEmoji = resEmoji
			return nil
		})
	}
	// 6.17 版本之后支持av/bv/cv和包含视频/专栏的短链转标题
	if (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DescIdToTitleHightLightIOS) || (general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DescIdToTitleHightLightAndroid) || (general.IsPadHD() && general.GetBuild() >= s.c.BuildLimit.DescIdToTitleHightLightPad) || general.IsPad() || general.IsAndroidHD() {
		var (
			aidm        = make(map[int64]map[int64]struct{})
			ssidm       = make(map[int32]struct{})
			epidm       = make(map[int32]struct{})
			cvidm       = make(map[int64]struct{})
			shortURLm   = make(map[string]struct{})
			shortToLong map[string]string
		)
		// 抽离短链
		for descURL := range dynCtx.BackfillDescURL {
			fIndex := _shortURLRgx.FindStringIndex(descURL)
			if len(fIndex) == 0 {
				continue
			}
			shortURLm[descURL] = struct{}{}
		}
		// 短链转长链
		if len(shortURLm) > 0 {
			var shortURLs []string
			for shortURL := range shortURLm {
				shortURLs = append(shortURLs, shortURL)
			}
			// 短链转长链
			var err error
			shortToLong, err = s.platformDao.ShortUrls(c, shortURLs)
			if err != nil {
				xmetric.DynamicBackfillAPI.Inc("/shorturl.service.v1.ShortUrl/ShortUrls", "request_error")
				log.Error("BackfillGetMaterial mid(%v) ShortUrls(%v), error %v", general.Mid, shortURLs, err)
			}
		}
		// 聚合网页链接数据
		for descURL := range dynCtx.BackfillDescURL {
			var descURLTmp = descURL
			if stl, ok := shortToLong[descURL]; ok && stl != "" {
				descURLTmp = stl
			}
			// archive
			if ugcIndex := _ugcURLRgx.FindStringIndex(descURLTmp); len(ugcIndex) > 0 {
				ugcURL := descURLTmp[ugcIndex[0]:ugcIndex[1]]
				// 拆bvid
				if bvIndex := _bvRgx.FindStringIndex(ugcURL); len(bvIndex) > 0 {
					bv := ugcURL[bvIndex[0]:bvIndex[1]]
					if aid, _ := bvid.BvToAv(bv); aid != 0 {
						if _, ok := aidm[aid]; !ok {
							aidm[aid] = make(map[int64]struct{})
						}
						dynCtx.BackfillDescURL[descURL] = &mdlv2.BackfillDescURLItem{
							Type:  api.DescType_desc_type_bv,
							Title: "",
							Rid:   bv,
						}
					}
					continue
				}
				// 拆avid
				if avIndex := _avRgx.FindStringIndex(ugcURL); len(avIndex) > 0 {
					avid := ugcURL[avIndex[0]:avIndex[1]]
					// 拆id
					if idIndex := _idRgx.FindStringIndex(avid); len(idIndex) > 0 {
						id := avid[idIndex[0]:idIndex[1]]
						if idInt64, _ := strconv.ParseInt(id, 10, 64); idInt64 != 0 {
							if _, ok := aidm[idInt64]; !ok {
								aidm[idInt64] = make(map[int64]struct{})
							}
							dynCtx.BackfillDescURL[descURL] = &mdlv2.BackfillDescURLItem{
								Type:  api.DescType_desc_type_av,
								Title: "",
								Rid:   id,
							}
						}
					}
				}
				continue
			}
			// ogv
			if ogvIndex := _ogvURLRgx.FindStringIndex(descURLTmp); len(ogvIndex) > 0 {
				ogvURL := descURLTmp[ogvIndex[0]:ogvIndex[1]]
				// 拆ssid
				if ssidIndex := _ogvssRgx.FindStringIndex(ogvURL); len(ssidIndex) > 0 {
					ssid := ogvURL[ssidIndex[0]:ssidIndex[1]]
					if idIndex := _idRgx.FindStringIndex(ssid); len(idIndex) > 0 {
						id := ssid[idIndex[0]:idIndex[1]]
						if idInt, _ := strconv.ParseInt(id, 10, 32); idInt != 0 {
							ssidm[int32(idInt)] = struct{}{}
							dynCtx.BackfillDescURL[descURL] = &mdlv2.BackfillDescURLItem{
								Type:  api.DescType_desc_type_ogv_season,
								Title: "",
								Rid:   id,
							}
						}
					}
					continue
				}
				if epidIndex := _ogvepRgx.FindStringIndex(ogvURL); len(epidIndex) > 0 {
					epid := ogvURL[epidIndex[0]:epidIndex[1]]
					if idIndex := _idRgx.FindStringIndex(epid); len(idIndex) > 0 {
						id := epid[idIndex[0]:idIndex[1]]
						if idInt, _ := strconv.ParseInt(id, 10, 32); idInt != 0 {
							epidm[int32(idInt)] = struct{}{}
							dynCtx.BackfillDescURL[descURL] = &mdlv2.BackfillDescURLItem{
								Type:  api.DescType_desc_type_ogv_ep,
								Title: "",
								Rid:   id,
							}
						}
					}
				}
				continue
			}
			// article
			if cvIndex := _articleURLRgx.FindStringIndex(descURLTmp); len(cvIndex) > 0 {
				cvURL := descURLTmp[cvIndex[0]:cvIndex[1]]
				// 拆cvid
				if cvidIndx := _cvRgx.FindStringIndex(cvURL); len(cvidIndx) > 0 {
					cvid := cvURL[cvidIndx[0]:cvidIndx[1]]
					// 拆id
					if idIndex := _idRgx.FindStringIndex(cvid); len(idIndex) > 0 {
						cv := cvid[idIndex[0]:idIndex[1]]
						if cvidInt64, _ := strconv.ParseInt(cv, 10, 64); cvidInt64 != 0 {
							cvidm[cvidInt64] = struct{}{}
							dynCtx.BackfillDescURL[descURL] = &mdlv2.BackfillDescURLItem{
								Type:  api.DescType_desc_type_cv,
								Title: "",
								Rid:   cv,
							}
						}
					}
				}
				continue
			}
		}
		// 聚合id
		for aid := range dynCtx.BackfillAvID {
			if aidInt64, _ := strconv.ParseInt(aid, 10, 64); aidInt64 != 0 {
				if _, ok := aidm[aidInt64]; !ok {
					aidm[aidInt64] = make(map[int64]struct{})
				}
			}
		}
		for bv := range dynCtx.BackfillBvID {
			if aid, _ := bvid.BvToAv(bv); aid != 0 {
				if _, ok := aidm[aid]; !ok {
					aidm[aid] = make(map[int64]struct{})
				}
			}
		}
		for cvid := range dynCtx.BackfillCvID {
			if cvidInt64, _ := strconv.ParseInt(cvid, 10, 64); cvidInt64 != 0 {
				cvidm[cvidInt64] = struct{}{}
			}
		}
		if len(aidm) > 0 {
			eg.Go(func(ctx context.Context) error {
				var aids []*archivegrpc.PlayAv
				for aid, cidm := range aidm {
					var ap = &archivegrpc.PlayAv{Aid: aid, NoPlayer: true}
					for cid := range cidm {
						ap.PlayVideos = append(ap.PlayVideos, &archivegrpc.PlayVideo{Cid: cid})
					}
					aids = append(aids, ap)
				}
				// 假卡不要秒开
				res, err := s.archiveDao.ArcsPlayer(ctx, aids, true, "")
				if err != nil {
					xmetric.DynamicBackfillAPI.Inc("/archive.service.v1.Archive/ArcsWithPlayurl", "request_error")
					log.Error("BackfillGetMaterial mid(%v) ArcsWithPlayurl(%v), err %v", general.Mid, aids, err)
					return err
				}
				dynCtx.ResBackfillArchive = res
				return nil
			})
		}
		if len(cvidm) > 0 {
			eg.Go(func(ctx context.Context) error {
				var cvids []int64
				for id := range cvidm {
					cvids = append(cvids, id)
				}
				res, err := s.articleDao.ArticleMetas(ctx, cvids)
				if err != nil {
					xmetric.DynamicBackfillAPI.Inc("/article.service.ArticleGRPC/ArticleMetas", "request_error")
					log.Error("BackfillGetMaterial mid(%v) ArticleMetasMc(%v), err %v", general.Mid, cvids, err)
					return err
				}
				dynCtx.ResBackfillArticle = res
				return nil
			})
		}
		if len(ssidm) > 0 {
			eg.Go(func(ctx context.Context) error {
				var ssids []int32
				for id := range ssidm {
					ssids = append(ssids, id)
				}
				res, err := s.pgcDao.Seasons(ctx, ssids)
				if err != nil {
					xmetric.DynamicBackfillAPI.Inc("/pgc.service.season.season.v1.Season/Cards", "request_error")
					log.Error("BackfillGetMaterial mid(%v) Seasons(%v), err %v", general.Mid, ssids, err)
					return err
				}
				dynCtx.ResBackfillSeason = res
				return nil
			})
		}
		if len(epidm) > 0 {
			eg.Go(func(ctx context.Context) error {
				var epids []int32
				for id := range epidm {
					epids = append(epids, id)
				}
				res, err := s.pgcDao.Episodes(ctx, epids)
				if err != nil {
					xmetric.DynamicBackfillAPI.Inc("/pgc.service.season.episode.v1.Episode/Cards", "request_error")
					log.Error("BackfillGetMaterial mid(%v) Episodes(%v), err %v", general.Mid, epids, err)
					return err
				}
				dynCtx.ResBackfillEpisode = res
				return nil
			})
		}
	}
	_ = eg.Wait()
}

func (s *Service) shareReqParam(general *mdlv2.GeneralParam, dyn *mdlv2.Dynamic) *shareApi.BusinessChannelsReq {
	if general.ShareID == "" || dyn.Rid == 0 {
		return nil
	}
	if general.Restriction.IsTeenagers {
		return nil
	}
	shareReq := &shareApi.BusinessChannelsReq{
		Mid:       general.Mid,
		ShareId:   general.ShareID,
		Platform:  general.GetPlatform(),
		Buvid:     general.GetBuvid(),
		ShareMode: general.ShareMode,
		MobiApp:   general.GetMobiApp(),
		Device:    general.GetDevice(),
	}
	switch dyn.Type {
	case mdlv2.DynTypeWord, mdlv2.DynTypeForward, mdlv2.DynTypeDraw, mdlv2.DynTypeCommonSquare, mdlv2.DynTypeCommonVertical, mdlv2.DynTypeApplet, mdlv2.DynTypeLive:
		shareReq.ShareOrigin = "dynamic"
		shareReq.Oid = strconv.FormatInt(int64(dyn.DynamicID), 10)
	case mdlv2.DynTypeVideo, mdlv2.DynTypeUGCSeason:
		shareReq.ShareOrigin = "ugc"
		shareReq.Oid = strconv.FormatInt(int64(dyn.Rid), 10)
	case mdlv2.DynTypeArticle:
		shareReq.ShareOrigin = "article"
		shareReq.Oid = strconv.FormatInt(int64(dyn.Rid), 10)
	case mdlv2.DynTypeSubscriptionNew, mdlv2.DynTypeSubscription:
		shareReq.ShareOrigin = "dynamic_subscribe"
		shareReq.Oid = strconv.FormatInt(int64(dyn.DynamicID), 10)
	default:
		return nil
	}
	return shareReq
}

func (s *Service) checkMidMaxInt32(c context.Context, mid int64, general *mdlv2.GeneralParam) bool {
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynMidInt32, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynMidInt32IOS) ||
			(general.IsPad() && general.GetBuild() < s.c.BuildLimit.DynMidInt32IOS) ||
			(general.IsPadHD() && general.GetBuild() < s.c.BuildLimit.DynMidInt32IOSHD) ||
			(general.IsAndroidHD() && general.GetBuild() < s.c.BuildLimit.DynMidInt32AndroidHD) ||
			(general.IsAndroidPick() && general.GetBuild() < s.c.BuildLimit.DynMidInt32Android)}) {
		if mid > math.MaxInt32 {
			return true
		}
	}
	return false
}
