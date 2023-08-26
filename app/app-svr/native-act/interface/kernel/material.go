package kernel

import (
	"context"
	"fmt"
	"time"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	appshowgrpc "git.bilibili.co/bapis/bapis-go/app/show/v1"
	articlemdl "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	commscoregrpc "git.bilibili.co/bapis/bapis-go/community/service/score"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dynfeedgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dyntopicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	hmtgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	liveplaygrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	xroomfeedgrpc "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	pgcappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcfollowgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	chargrpc "git.bilibili.co/bapis/bapis-go/pgc/service/media"
	actplatv2grpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"github.com/google/uuid"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	appdyngrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

const (
	// 批量请求数
	_arcReqMax         = 100
	_liveReqMax        = 50
	_gameReqMax        = 20
	_weekReqMax        = 50
	_roomReqMax        = 100
	_articleReqMax     = 50
	_epReqMax          = 50
	_actSubProtoReqMax = 50
	_folderReqMax      = 30
	_accountReqMax     = 50
	_dynDetailReqMax   = 50
	_tagReqMax         = 30
	_playAvsMax        = 50
	_relFidsMax        = 30
	_naCardMax         = 100
	_naAllMax          = 100
	_naPagesMax        = 150
	_channelMax        = 30
	_upRsvInfoMax      = 30
	_ssidReqMax        = 50
	_dynVoteInfoReqMax = 30
	_actSubReqMax      = 50
	_actProgGroupMax   = 20
	_pgcFollowReqMax   = 20
	_comicInfoReqMax   = 20
	_actRsvFollowMax   = 20
	_ticketFavStateMax = 20
	_scoreIdsMax       = 20
	// addItem 入参数
	_roomParams         = 2
	_folderParams       = 2
	_actProgGroupParams = 2
)

type RequestID string

type Material struct {
	WeekCard           map[int64]*appshowgrpc.SerieConfig                             //每周必看
	GameCard           map[int64]*model.GaItem                                        //游戏
	LiveCard           map[uint64]*xroomfeedgrpc.LiveCardInfo                         //直播房间ids,长号card
	Arcs               map[int64]*arcgrpc.Arc                                         //稿件
	LiveRooms          map[int64]map[int64]*liveplaygrpc.RoomList                     //直播间：isLive=>map
	Articles           map[int64]*articlemdl.Meta                                     //专栏
	Episodes           map[int64]*model.EpPlayer                                      //剧集
	Folders            map[int32]map[int64]*favmdl.Folder                             //播单
	ActSubProtos       map[int64]*activitygrpc.ActSubProtocolReply                    //活动和扩展信息
	Accounts           map[int64]*accountgrpc.Info                                    //用户信息
	HasDynsRlys        map[RequestID]*dyntopicgrpc.HasDynsRsp                         //话题下是否有更多动态
	ListDynsRlys       map[RequestID]*dyntopicgrpc.ListDynsRsp                        //话题下动态feed流
	DynDetails         map[RequestID]map[int64]*appdyngrpc.DynamicItem                //动态卡片详情
	ActLikesRlys       map[RequestID]*activitygrpc.LikesReply                         //根据sid查询活动列表信息
	DynRevsRlys        map[RequestID]*dynfeedgrpc.FetchDynIdByRevsRsp                 //根据(type, rid)数组拉取对应DynId数组
	Tags               map[int64]*taggrpc.Tag                                         //标签
	MixExtsRlys        map[RequestID]*natpagegrpc.ModuleMixExtsReply                  //根据module_id获取所有配置的id信息
	GetHisRlys         map[RequestID]*actplatv2grpc.GetHistoryResp                    //获取单个用户指定counter下的积分历史
	PageArcsRlys       map[RequestID]*populargrpc.PageArcsResp                        //
	MixExtRlys         map[RequestID]*natpagegrpc.ModuleMixExtReply                   //根据module_id获取配置的id信息
	RankRstRlys        map[RequestID]*activitygrpc.RankResultResp                     //榜单排行结果
	SelSerieRlys       map[RequestID]*appshowgrpc.SelectedSerieRly                    //热门精选获取指定期
	UpListRlys         map[RequestID]*activitygrpc.UpListReply                        //up主活动数据源列表
	RelInfosRlys       map[RequestID]*chargrpc.CharacterRelInfosReply                 //角色关联信息
	BriefDynsRlys      map[RequestID]*model.BriefDynsRly                              //
	QueryWidRlys       map[int32]*pgcappgrpc.QueryWidReply                            //给主站活动平台查询wid对应的item的接口
	RoomsByActIdRlys   map[RequestID]*liveplaygrpc.GetListByActIdResp                 //通过 活动id 获取直播间信息
	ArcsPlayer         map[int64]*arcgrpc.ArcPlayer                                   //带秒开信息的稿件
	ChannelFeedRlys    map[RequestID]*hmtgrpc.ChannelFeedReply                        //港澳台垂类feed流数据
	Relations          map[int64]*relationgrpc.FollowingReply                         //用户关注关系
	AccountCards       map[int64]*accountgrpc.Card                                    //用户卡片信息
	NativePageCards    map[int64]*natpagegrpc.NativePageCard                          //批量获取话题活动卡-处理跳转地址
	NativeAllPages     map[int64]*natpagegrpc.NativePage                              //批量获取话题活动信息接口-返回所有状态
	NativePages        map[int64]*natpagegrpc.NativePage                              //批量获取话题活动信息接口-仅仅返回有效状态
	Channels           map[int64]*channelgrpc.Channel                                 //频道信息列表
	VoteRankRlys       map[RequestID]*activitygrpc.GetVoteActivityRankResp            //投票组件-查看活动下的投票排行
	UpRsvInfos         map[RequestID]map[int64]*activitygrpc.UpActReserveRelationInfo //up主预约关联活动基本信息
	RoomSessionInfos   map[int64]*roomgategrpc.SessionInfos                           //获取场次信息-直播预约（批量：多个主播，每个主播多个场次）
	TimelineRlys       map[RequestID]*populargrpc.TimeLineReply                       //咨询后台-时间轴
	SeasonCards        map[int32]*pgcappgrpc.SeasonCardInfoProto                      //season卡片信息
	SeasonByPlayIdRlys map[RequestID]*pgcappgrpc.SeasonByPlayIdReply                  //根据片单id返回卡片信息
	ActiveUsersRlys    map[RequestID]*model.ActiveUsersRly                            //话题浏览量
	DynVoteInfos       map[int64]*dyncommongrpc.VoteInfo                              //
	ActSubjects        map[RequestID]map[int64]*activitygrpc.Subject                  //活动信息
	ActProgressGroups  map[int64]map[int64]*activitygrpc.ActivityProgressGroup        //进度组件数值
	SourceDetailRlys   map[RequestID]*model.SourceDetailRly                           //数据源详情查询
	ProductDetailRlys  map[RequestID]*model.ProductDetailRly                          //商品卡数据源详情查询
	PgcFollowStatuses  map[int32]*pgcfollowgrpc.FollowStatusProto                     //追番状态
	ComicInfos         map[int64]*model.ComicInfo                                     //漫画信息
	ActRsvFollows      map[int64]*activitygrpc.ReserveFollowingReply                  //批量查询预约状态
	AwardStates        map[int64]*activitygrpc.AwardSubjectStateReply                 //领奖组件获取奖励状态接口
	TicketFavStates    map[int64]bool                                                 //会员购票务“想去”状态
	ActRelationInfos   map[int64]*activitygrpc.ActRelationInfoReply                   //获取活动关联平台信息
	PlatCounterResRlys map[RequestID]*actplatv2grpc.GetCounterResResp                 //获取counter数值列表，通常用于获取单个用户指定counter当日值
	PlatTotalResRlys   map[RequestID]*actplatv2grpc.GetTotalResResp                   //获取单个用户指定counter下分数
	LotUnusedRlys      map[string]*activitygrpc.LotteryUnusedTimesReply               //获取剩余抽奖次数
	ScoreTargets       map[int64]*commscoregrpc.ScoreTarget                           //批量查询评分对象的评分 e.g:话题页电影列表使用
}

type MaterialLoader struct {
	c   context.Context
	dep dao.Dependency
	ss  *Session

	aids               []int64
	liveIDs            []int64 //直播房间ids,长号
	gameIDs            []int64 //游戏ids
	weekIDs            []int64 //每周必看ids
	roomIDs            map[int64][]int64
	cvids              []int64
	epids              []int64
	folderIDs          map[int32][]int64
	actSubProtoIDs     []int64
	mids               []int64
	hasDynsReqs        map[RequestID]*dyntopicgrpc.HasDynsReq
	listDynsReqs       map[RequestID]*dyntopicgrpc.ListDynsReq
	dynDetailReqs      map[RequestID]*appdyngrpc.DynServerDetailsReq
	actLikesReqs       map[RequestID]*ActLikesReq
	dynRevsReqs        map[RequestID]*dynfeedgrpc.FetchDynIdByRevsReq
	tagIDs             []int64
	mixExtsReqs        map[RequestID]*ModuleMixExtsReq
	getHisReqs         map[RequestID]*actplatv2grpc.GetHistoryReq
	pageArcsReqs       map[RequestID]*populargrpc.PageArcsReq
	mixExtReqs         map[RequestID]*natpagegrpc.ModuleMixExtReq
	rankRstReqs        map[RequestID]*RankResultReq
	selSerieReqs       map[RequestID]*appshowgrpc.SelectedSerieReq
	upListReqs         map[RequestID]*UpListReq
	relInfosReqs       map[RequestID]*RelInfosReq
	briefDynsReqs      map[RequestID]*BriefDynsReq
	wids               []int32
	roomsByActIdReqs   map[RequestID]*liveplaygrpc.GetListByActIdReq
	playAvs            []*arcgrpc.PlayAv
	channelFeedReqs    map[RequestID]*ChannelFeedReq
	relFids            []int64
	cardMids           []int64
	pidsOfNaCard       []int64
	pidsOfNaAll        []int64
	pidsOfNaPages      []int64
	channelIDs         []int64
	VoteRankReqs       map[RequestID]*activitygrpc.GetVoteActivityRankReq
	upRsvIDsReqs       map[RequestID]*UpRsvIDsReq
	uidLiveIDs         map[int64][]string
	timelineReqs       map[RequestID]*populargrpc.TimeLineRequest
	ssids              []int32
	seasonByPlayIdReqs map[RequestID]*pgcappgrpc.SeasonByPlayIdReq
	activeUsersReqs    map[RequestID]*model.ActiveUsersReq
	dynVoteIDs         []int64
	actSidsReqs        map[RequestID]*ActSidsReq
	actSidGroupIDs     map[int64][]int64
	sourceDetailReqs   map[RequestID]*SourceDetailReq
	productDetailReqs  map[RequestID]*model.ProductDetailReq
	pgcFollowSeasonIds []int32
	comicIds           []int64
	actRsvIds          []int64
	awardIds           []int64
	ticketFavIds       []int64
	actRelationIds     []int64
	platCounterReqs    map[RequestID]*actplatv2grpc.GetCounterResReq
	platTotalReqs      map[RequestID]*actplatv2grpc.GetTotalResReq
	lotteryIds         []string
	scoreIds           []int64
}

type ModuleMixExtsReq struct {
	Req         *natpagegrpc.ModuleMixExtsReq
	NeedMultiML bool
	IsLive      int64
	ArcType     model.MaterialType
}

type ActLikesReq struct {
	Req         *activitygrpc.ActLikesReq
	NeedMultiML bool
	ArcType     model.MaterialType
}

type RelInfosReq struct {
	Req         *chargrpc.CharacterIdsOidsReq
	NeedMultiML bool
	ShowNum     int64 //默认无限制
}

type BriefDynsReq struct {
	Req         *model.BriefDynsReq
	NeedMultiML bool
	ArcType     model.MaterialType
}

type ChannelFeedReq struct {
	Req         *hmtgrpc.ChannelFeedReq
	NeedMultiML bool
}

type UpListReq struct {
	Req         *activitygrpc.UpListReq
	NeedMultiML bool
}

type RankResultReq struct {
	Req         *activitygrpc.RankResultReq
	NeedMultiML bool
}

type UpRsvIDsReq struct {
	IDs         []int64
	NeedMultiML bool
	NeedAccount bool
}

type ActSidsReq struct {
	IDs         []int64
	NeedMultiML bool
	NeedAccount bool
}

type SourceDetailReq struct {
	Req         *model.SourceDetailReq
	NeedMultiML bool
}

type MatLoaderFactory struct {
	c   context.Context
	dep dao.Dependency
	ss  *Session
}

func NewMatLoaderFactory(c context.Context, dep dao.Dependency, ss *Session) *MatLoaderFactory {
	return &MatLoaderFactory{c: c, dep: dep, ss: ss}
}

func (mlf *MatLoaderFactory) NewMaterialLoader() *MaterialLoader {
	return NewMaterialLoader(mlf.c, mlf.dep, mlf.ss)
}

func NewMaterialLoader(c context.Context, dep dao.Dependency, ss *Session) *MaterialLoader {
	return &MaterialLoader{c: c, dep: dep, ss: ss}
}

func (ml *MaterialLoader) AddItem(matType model.MaterialType, data ...interface{}) (RequestID, error) {
	switch matType {
	case model.MaterialWeeks:
		return "", ml.addWeekIDs(data...)
	case model.MaterialGame:
		return "", ml.addGameIDs(data...)
	case model.MaterialLive:
		return "", ml.addLiveIDs(data...)
	case model.MaterialArchive:
		return "", ml.addAids(data...)
	case model.MaterialLiveRoom:
		return "", ml.addRoomIDs(data...)
	case model.MaterialArticle:
		return "", ml.addCvids(data...)
	case model.MaterialEpisode:
		return "", ml.addEpids(data...)
	case model.MaterialFolder:
		return "", ml.addFolderIDs(data...)
	case model.MaterialActSubProto:
		return "", ml.addActSubProtoIDs(data...)
	case model.MaterialAccount:
		return "", ml.addMids(data...)
	case model.MaterialHasDynsRly:
		return ml.addHasDynsReq(data...)
	case model.MaterialListDynsRly:
		return ml.addListDynsReq(data...)
	case model.MaterialDynDetail:
		return ml.addDynDetailReq(data...)
	case model.MaterialActLikesRly:
		return ml.addActLikesReq(data...)
	case model.MaterialDynRevsRly:
		return ml.addDynRevsReq(data...)
	case model.MaterialTag:
		return "", ml.addTagIDs(data...)
	case model.MaterialMixExtsRly:
		return ml.addMixExtsReq(data...)
	case model.MaterialGetHisRly:
		return ml.addGetHisReq(data...)
	case model.MaterialPageArcsRly:
		return ml.addPageArcsReq(data...)
	case model.MaterialMixExtRly:
		return ml.addMixExtReq(data...)
	case model.MaterialRankRstRly:
		return ml.addRankRstReq(data...)
	case model.MaterialSelSerieRly:
		return ml.addSelSerieReq(data...)
	case model.MaterialUpListRly:
		return ml.addUpListReq(data...)
	case model.MaterialRelInfosRly:
		return ml.addRelInfosReq(data...)
	case model.MaterialBriefDynsRly:
		return ml.addBriefDynsReq(data...)
	case model.MaterialQueryWidRly:
		return "", ml.addWids(data...)
	case model.MaterialRoomsByActIdRly:
		return ml.addRoomsByActIdReq(data...)
	case model.MaterialArcPlayer:
		return "", ml.addPlayAvs(data...)
	case model.MaterialChannelFeedRly:
		return ml.addChannelFeedReq(data...)
	case model.MaterialRelation:
		return "", ml.addRelFids(data...)
	case model.MaterialAccountCard:
		return "", ml.addCardMids(data...)
	case model.MaterialNativeCard:
		return "", ml.addPidsOfNaCard(data...)
	case model.MaterialNativeAllPage:
		return "", ml.addPidsOfNaAll(data...)
	case model.MaterialNativePages:
		return "", ml.addPidsOfNaPages(data...)
	case model.MaterialChannel:
		return "", ml.addChannelIDs(data...)
	case model.MaterialVoteRankRly:
		return ml.addVoteRankReq(data...)
	case model.MaterialUpRsvInfo:
		return ml.addUpRsvIDsReq(data...)
	case model.MaterialRoomSessionInfo:
		return "", ml.addUidLiveIDs(data...)
	case model.MaterialTimelineRly:
		return ml.addTimelineReq(data...)
	case model.MaterialSeasonCard:
		return "", ml.addSsids(data...)
	case model.MaterialSeasonByPlayIdRly:
		return ml.addSeasonByPlayIdReq(data...)
	case model.MaterialActiveUsersRly:
		return ml.addActiveUsersReq(data...)
	case model.MaterialDynVoteInfo:
		return "", ml.addDynVoteIDs(data...)
	case model.MaterialActSubject:
		return ml.addActSidsReq(data...)
	case model.MaterialActProgressGroup:
		return "", ml.addActSidGroupIDs(data...)
	case model.MaterialSourceDetail:
		return ml.addSourceDetailReq(data...)
	case model.MaterialProductDetail:
		return ml.addProductDetailReq(data...)
	case model.MaterialPgcFollowStatus:
		return "", ml.addPgcFollowSeasonIds(data...)
	case model.MaterialComicInfo:
		return "", ml.addComicIds(data...)
	case model.MaterialActReserveFollow:
		return "", ml.addActRsvIds(data...)
	case model.MaterialActAwardState:
		return "", ml.addAwardIds(data...)
	case model.MaterialTicketFavState:
		return "", ml.addTicketFavIds(data...)
	case model.MaterialActRelationInfo:
		return "", ml.addActRelationIds(data...)
	case model.MaterialPlatCounterRes:
		return ml.addPlatCounterResReqs(data...)
	case model.MaterialPlatTotalRes:
		return ml.addPlatTotalResReqs(data...)
	case model.MaterialLotteryUnused:
		return "", ml.addLotteryIds(data...)
	case model.MaterialScoreTarget:
		return "", ml.addScoreIds(data...)
	default:
		log.Warn("unknown material_type=%+v", matType)
	}
	return "", nil
}

// nolint:gocognit
func (ml *MaterialLoader) JoinLoader(sourceLoader *MaterialLoader) {
	if sourceLoader == nil {
		return
	}
	if len(sourceLoader.weekIDs) > 0 {
		ml.weekIDs = append(ml.weekIDs, sourceLoader.weekIDs...)
	}
	if len(sourceLoader.gameIDs) > 0 {
		ml.gameIDs = append(ml.gameIDs, sourceLoader.gameIDs...)
	}
	if len(sourceLoader.liveIDs) > 0 {
		ml.liveIDs = append(ml.liveIDs, sourceLoader.liveIDs...)
	}
	if len(sourceLoader.aids) > 0 {
		ml.aids = append(ml.aids, sourceLoader.aids...)
	}
	if sourceLoader.roomIDs != nil {
		if ml.roomIDs == nil {
			ml.roomIDs = make(map[int64][]int64)
		}
		for isLive, roomIDs := range sourceLoader.roomIDs {
			ml.roomIDs[isLive] = append(ml.roomIDs[isLive], roomIDs...)
		}
	}
	if len(sourceLoader.cvids) > 0 {
		ml.cvids = append(ml.cvids, sourceLoader.cvids...)
	}
	if len(sourceLoader.epids) > 0 {
		ml.epids = append(ml.epids, sourceLoader.epids...)
	}
	if sourceLoader.folderIDs != nil {
		if ml.folderIDs == nil {
			ml.folderIDs = make(map[int32][]int64)
		}
		for typ, folderIDs := range sourceLoader.folderIDs {
			ml.folderIDs[typ] = append(ml.folderIDs[typ], folderIDs...)
		}
	}
	if len(sourceLoader.actSubProtoIDs) > 0 {
		ml.actSubProtoIDs = append(ml.actSubProtoIDs, sourceLoader.actSubProtoIDs...)
	}
	if len(sourceLoader.mids) > 0 {
		ml.mids = append(ml.mids, sourceLoader.mids...)
	}
	if sourceLoader.hasDynsReqs != nil {
		if ml.hasDynsReqs == nil {
			ml.hasDynsReqs = make(map[RequestID]*dyntopicgrpc.HasDynsReq)
		}
		for reqID, req := range sourceLoader.hasDynsReqs {
			ml.hasDynsReqs[reqID] = req
		}
	}
	if sourceLoader.listDynsReqs != nil {
		if ml.listDynsReqs == nil {
			ml.listDynsReqs = make(map[RequestID]*dyntopicgrpc.ListDynsReq)
		}
		for reqID, req := range sourceLoader.listDynsReqs {
			ml.listDynsReqs[reqID] = req
		}
	}
	if sourceLoader.dynDetailReqs != nil {
		if ml.dynDetailReqs == nil {
			ml.dynDetailReqs = make(map[RequestID]*appdyngrpc.DynServerDetailsReq)
		}
		for reqID, req := range sourceLoader.dynDetailReqs {
			ml.dynDetailReqs[reqID] = req
		}
	}
	if sourceLoader.actLikesReqs != nil {
		if ml.actLikesReqs == nil {
			ml.actLikesReqs = make(map[RequestID]*ActLikesReq)
		}
		for reqID, req := range sourceLoader.actLikesReqs {
			ml.actLikesReqs[reqID] = req
		}
	}
	if sourceLoader.dynRevsReqs != nil {
		if ml.dynRevsReqs == nil {
			ml.dynRevsReqs = make(map[RequestID]*dynfeedgrpc.FetchDynIdByRevsReq)
		}
		for reqID, req := range sourceLoader.dynRevsReqs {
			ml.dynRevsReqs[reqID] = req
		}
	}
	if len(sourceLoader.tagIDs) > 0 {
		ml.tagIDs = append(ml.tagIDs, sourceLoader.tagIDs...)
	}
	if sourceLoader.mixExtsReqs != nil {
		if ml.mixExtsReqs == nil {
			ml.mixExtsReqs = make(map[RequestID]*ModuleMixExtsReq)
		}
		for reqID, req := range sourceLoader.mixExtsReqs {
			ml.mixExtsReqs[reqID] = req
		}
	}
	if sourceLoader.getHisReqs != nil {
		if ml.getHisReqs == nil {
			ml.getHisReqs = make(map[RequestID]*actplatv2grpc.GetHistoryReq)
		}
		for reqID, req := range sourceLoader.getHisReqs {
			ml.getHisReqs[reqID] = req
		}
	}
	if sourceLoader.pageArcsReqs != nil {
		if ml.pageArcsReqs == nil {
			ml.pageArcsReqs = make(map[RequestID]*populargrpc.PageArcsReq)
		}
		for reqID, req := range sourceLoader.pageArcsReqs {
			ml.pageArcsReqs[reqID] = req
		}
	}
	if sourceLoader.mixExtReqs != nil {
		if ml.mixExtReqs == nil {
			ml.mixExtReqs = make(map[RequestID]*natpagegrpc.ModuleMixExtReq)
		}
		for reqID, req := range sourceLoader.mixExtReqs {
			ml.mixExtReqs[reqID] = req
		}
	}
	if sourceLoader.rankRstReqs != nil {
		if ml.rankRstReqs == nil {
			ml.rankRstReqs = make(map[RequestID]*RankResultReq)
		}
		for reqID, req := range sourceLoader.rankRstReqs {
			ml.rankRstReqs[reqID] = req
		}
	}
	if sourceLoader.selSerieReqs != nil {
		if ml.selSerieReqs == nil {
			ml.selSerieReqs = make(map[RequestID]*appshowgrpc.SelectedSerieReq)
		}
		for reqID, req := range sourceLoader.selSerieReqs {
			ml.selSerieReqs[reqID] = req
		}
	}
	if sourceLoader.upListReqs != nil {
		if ml.upListReqs == nil {
			ml.upListReqs = make(map[RequestID]*UpListReq)
		}
		for reqID, req := range sourceLoader.upListReqs {
			ml.upListReqs[reqID] = req
		}
	}
	if sourceLoader.relInfosReqs != nil {
		if ml.relInfosReqs == nil {
			ml.relInfosReqs = make(map[RequestID]*RelInfosReq)
		}
		for reqID, req := range sourceLoader.relInfosReqs {
			ml.relInfosReqs[reqID] = req
		}
	}
	if sourceLoader.briefDynsReqs != nil {
		if ml.briefDynsReqs == nil {
			ml.briefDynsReqs = make(map[RequestID]*BriefDynsReq)
		}
		for reqID, req := range sourceLoader.briefDynsReqs {
			ml.briefDynsReqs[reqID] = req
		}
	}
	if sourceLoader.wids != nil {
		ml.wids = append(ml.wids, sourceLoader.wids...)
	}
	if sourceLoader.roomsByActIdReqs != nil {
		if ml.roomsByActIdReqs == nil {
			ml.roomsByActIdReqs = make(map[RequestID]*liveplaygrpc.GetListByActIdReq)
		}
		for reqID, req := range sourceLoader.roomsByActIdReqs {
			ml.roomsByActIdReqs[reqID] = req
		}
	}
	if sourceLoader.playAvs != nil {
		ml.playAvs = append(ml.playAvs, sourceLoader.playAvs...)
	}
	if sourceLoader.channelFeedReqs != nil {
		if ml.channelFeedReqs == nil {
			ml.channelFeedReqs = make(map[RequestID]*ChannelFeedReq)
		}
		for reqID, req := range sourceLoader.channelFeedReqs {
			ml.channelFeedReqs[reqID] = req
		}
	}
	if sourceLoader.relFids != nil {
		ml.relFids = append(ml.relFids, sourceLoader.relFids...)
	}
	if sourceLoader.cardMids != nil {
		ml.cardMids = append(ml.cardMids, sourceLoader.cardMids...)
	}
	if sourceLoader.pidsOfNaCard != nil {
		ml.pidsOfNaCard = append(ml.pidsOfNaCard, sourceLoader.pidsOfNaCard...)
	}
	if sourceLoader.pidsOfNaAll != nil {
		ml.pidsOfNaAll = append(ml.pidsOfNaAll, sourceLoader.pidsOfNaAll...)
	}
	if sourceLoader.pidsOfNaPages != nil {
		ml.pidsOfNaPages = append(ml.pidsOfNaPages, sourceLoader.pidsOfNaPages...)
	}
	if sourceLoader.channelIDs != nil {
		ml.channelIDs = append(ml.channelIDs, sourceLoader.channelIDs...)
	}
	if sourceLoader.VoteRankReqs != nil {
		if ml.VoteRankReqs == nil {
			ml.VoteRankReqs = make(map[RequestID]*activitygrpc.GetVoteActivityRankReq)
		}
		for reqID, req := range sourceLoader.VoteRankReqs {
			ml.VoteRankReqs[reqID] = req
		}
	}
	if sourceLoader.upRsvIDsReqs != nil {
		if ml.upRsvIDsReqs == nil {
			ml.upRsvIDsReqs = make(map[RequestID]*UpRsvIDsReq)
		}
		for reqID, req := range sourceLoader.upRsvIDsReqs {
			ml.upRsvIDsReqs[reqID] = req
		}
	}
	if sourceLoader.uidLiveIDs != nil {
		if ml.uidLiveIDs == nil {
			ml.uidLiveIDs = make(map[int64][]string)
		}
		for mid, liveIDs := range sourceLoader.uidLiveIDs {
			ml.uidLiveIDs[mid] = append(ml.uidLiveIDs[mid], liveIDs...)
		}
	}
	if sourceLoader.timelineReqs != nil {
		if ml.timelineReqs == nil {
			ml.timelineReqs = make(map[RequestID]*populargrpc.TimeLineRequest)
		}
		for reqID, req := range sourceLoader.timelineReqs {
			ml.timelineReqs[reqID] = req
		}
	}
	if len(sourceLoader.ssids) > 0 {
		ml.ssids = append(ml.ssids, sourceLoader.ssids...)
	}
	if sourceLoader.seasonByPlayIdReqs != nil {
		if ml.seasonByPlayIdReqs == nil {
			ml.seasonByPlayIdReqs = make(map[RequestID]*pgcappgrpc.SeasonByPlayIdReq)
		}
		for reqID, req := range sourceLoader.seasonByPlayIdReqs {
			ml.seasonByPlayIdReqs[reqID] = req
		}
	}
	if sourceLoader.activeUsersReqs != nil {
		if ml.activeUsersReqs == nil {
			ml.activeUsersReqs = make(map[RequestID]*model.ActiveUsersReq)
		}
		for reqID, req := range sourceLoader.activeUsersReqs {
			ml.activeUsersReqs[reqID] = req
		}
	}
	if len(sourceLoader.dynVoteIDs) > 0 {
		ml.dynVoteIDs = append(ml.dynVoteIDs, sourceLoader.dynVoteIDs...)
	}
	if sourceLoader.actSidsReqs != nil {
		if ml.actSidsReqs == nil {
			ml.actSidsReqs = make(map[RequestID]*ActSidsReq)
		}
		for reqID, req := range sourceLoader.actSidsReqs {
			ml.actSidsReqs[reqID] = req
		}
	}
	if sourceLoader.actSidGroupIDs != nil {
		if ml.actSidGroupIDs == nil {
			ml.actSidGroupIDs = make(map[int64][]int64)
		}
		for sid, gids := range sourceLoader.actSidGroupIDs {
			ml.actSidGroupIDs[sid] = append(ml.actSidGroupIDs[sid], gids...)
		}
	}
	if sourceLoader.sourceDetailReqs != nil {
		if ml.sourceDetailReqs == nil {
			ml.sourceDetailReqs = make(map[RequestID]*SourceDetailReq)
		}
		for reqID, req := range sourceLoader.sourceDetailReqs {
			ml.sourceDetailReqs[reqID] = req
		}
	}
	if sourceLoader.productDetailReqs != nil {
		if ml.productDetailReqs == nil {
			ml.productDetailReqs = make(map[RequestID]*model.ProductDetailReq)
		}
		for reqID, req := range sourceLoader.productDetailReqs {
			ml.productDetailReqs[reqID] = req
		}
	}
	if len(sourceLoader.pgcFollowSeasonIds) > 0 {
		ml.pgcFollowSeasonIds = append(ml.pgcFollowSeasonIds, sourceLoader.pgcFollowSeasonIds...)
	}
	if len(sourceLoader.comicIds) > 0 {
		ml.comicIds = append(ml.comicIds, sourceLoader.comicIds...)
	}
	if len(sourceLoader.actRsvIds) > 0 {
		ml.actRsvIds = append(ml.actRsvIds, sourceLoader.actRsvIds...)
	}
	if len(sourceLoader.awardIds) > 0 {
		ml.awardIds = append(ml.awardIds, sourceLoader.awardIds...)
	}
	if len(sourceLoader.ticketFavIds) > 0 {
		ml.ticketFavIds = append(ml.ticketFavIds, sourceLoader.ticketFavIds...)
	}
	if len(sourceLoader.actRelationIds) > 0 {
		ml.actRelationIds = append(ml.actRelationIds, sourceLoader.actRelationIds...)
	}
	if sourceLoader.platCounterReqs != nil {
		if ml.platCounterReqs == nil {
			ml.platCounterReqs = make(map[RequestID]*actplatv2grpc.GetCounterResReq)
		}
		for reqID, req := range sourceLoader.platCounterReqs {
			ml.platCounterReqs[reqID] = req
		}
	}
	if sourceLoader.platTotalReqs != nil {
		if ml.platTotalReqs == nil {
			ml.platTotalReqs = make(map[RequestID]*actplatv2grpc.GetTotalResReq)
		}
		for reqID, req := range sourceLoader.platTotalReqs {
			ml.platTotalReqs[reqID] = req
		}
	}
	if len(sourceLoader.lotteryIds) > 0 {
		ml.lotteryIds = append(ml.lotteryIds, sourceLoader.lotteryIds...)
	}
	if len(sourceLoader.scoreIds) > 0 {
		ml.scoreIds = append(ml.scoreIds, sourceLoader.scoreIds...)
	}
}

func (ml *MaterialLoader) Load(material *Material) *Material {
	if material == nil {
		material = &Material{}
	}
	eg := errgroup.WithContext(ml.c)
	ml.doWeekCards(eg, material)
	ml.doGameCard(eg, material)
	ml.doLiveCard(eg, material)
	ml.doArcs(eg, material)
	ml.doLiveRooms(eg, material)
	ml.doArticles(eg, material)
	ml.doEpisodes(eg, material)
	ml.doFolders(eg, material)
	ml.doActSubProtos(eg, material)
	ml.doAccounts(eg, material)
	ml.doHasDynsRlys(eg, material)
	ml.doListDynsRlys(eg, material)
	ml.doDynDetails(eg, material)
	ml.doActLikesRlys(eg, material)
	ml.doDynRevsRlys(eg, material)
	ml.doTags(eg, material)
	ml.doMixExtsRlys(eg, material)
	ml.doGetHisRlys(eg, material)
	ml.doPageArcsRlys(eg, material)
	ml.doMixExtRlys(eg, material)
	ml.doRankRstRlys(eg, material)
	ml.doSelSerieRlys(eg, material)
	ml.doUpListRlys(eg, material)
	ml.doRelInfosRlys(eg, material)
	ml.doBriefDynsRlys(eg, material)
	ml.doQueryWidRlys(eg, material)
	ml.doRoomsByActIdRlys(eg, material)
	ml.doArcsPlayer(eg, material)
	ml.doChannelFeedRlys(eg, material)
	ml.doRelations(eg, material)
	ml.doAccountCards(eg, material)
	ml.doNativePageCards(eg, material)
	ml.doNativeAllPages(eg, material)
	ml.doNativePages(eg, material)
	ml.doChannels(eg, material)
	ml.doVoteRankRlys(eg, material)
	ml.doUpRsvInfos(eg, material)
	ml.doRoomSessionInfos(eg, material)
	ml.doTimelineRlys(eg, material)
	ml.doSeasonCards(eg, material)
	ml.doSeasonByPlayIdRly(eg, material)
	ml.doActiveUsersRly(eg, material)
	ml.doDynVoteInfo(eg, material)
	ml.doActSubject(eg, material)
	ml.doActProgressGroups(eg, material)
	ml.doSourceDetailRly(eg, material)
	ml.doProductDetailRly(eg, material)
	ml.doPgcFollowStatuses(eg, material)
	ml.doComicInfos(eg, material)
	ml.doActRsvFollows(eg, material)
	ml.doActAwardStates(eg, material)
	ml.doTicketStates(eg, material)
	ml.doActRelationInfos(eg, material)
	ml.doPlatCounterResRlys(eg, material)
	ml.doPlatTotalResRlys(eg, material)
	ml.doLotUnusedRlys(eg, material)
	ml.doScoreTargets(eg, material)
	_ = eg.Wait()
	return material
}

func requestID() RequestID {
	if u, err := uuid.NewRandom(); err == nil {
		return RequestID(u.String())
	}
	return RequestID(fmt.Sprintf("req-%d", time.Now().UnixNano()))
}
