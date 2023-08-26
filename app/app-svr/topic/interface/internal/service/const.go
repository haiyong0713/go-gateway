package service

import (
	"context"
	"fmt"
	"time"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/component/metadata/restriction"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	xmetadata "go-common/library/net/metadata"
	xtime "go-common/library/time"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	cardapi "go-gateway/app/app-svr/app-card/interface/model/card/proto"
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	midint64 "go-gateway/app/app-svr/app-interface/interface-legacy/middleware/midInt64"
	"go-gateway/app/app-svr/archive/middleware"
	middlewarev1 "go-gateway/app/app-svr/archive/middleware/v1"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/app/app-svr/topic/card/proto/dyn_handler"
	api "go-gateway/app/app-svr/topic/interface/api"
	"go-gateway/app/app-svr/topic/interface/internal/model"
	"strconv"
	"strings"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	livexroomgate "git.bilibili.co/bapis/bapis-go/live/xroom-gate"
	pgrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"
	esportsgrpc "git.bilibili.co/bapis/bapis-go/operational/esportsservice"
	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

const (
	// 来源参数
	_sourceFromWebDetails       = "Web"
	_sourceFromH5Details        = "H5"
	_sourceFromAPPHalf          = "APP_HALF"
	_sourceFromWebDynamicFeed   = "/x/topic/web/dynamic/rcmd"
	_sourceFromStoryModeDetails = "topic_from_story_mode"

	// 话题卡类型（话题后端定值）
	_dynamicCardType = 0

	// 话题卡类型（网关定值）
	_topicCardFold    = "FOLD"
	_topicCardDynamic = "DYNAMIC"

	// 	live entry
	_newTopicLiveEntry = "new_topic_live_inline"

	// 话题详情页inline资源类型
	_inlineResourceTypeLive = 1 // 直播
	_inlineResourceTypeAv   = 2 // 视频
	_inlineResourceTypeOgv  = 3 // ogv

	// 回填结构
	_ogvssRex      = `(SS|ss|Ss|sS)[0-9]+`
	_ogvepRex      = `(EP|ep|Eo|eP)[0-9]+`
	_shortURLRex   = `(?i)https://(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn)/[1-9A-NP-Za-km-z]{6,10}($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_ogvURLReg     = `(?i)((http(s)?://)?((uat-)?www.bilibili.com/bangumi/(play/|media/)|(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn)/)(ss|ep)[0-9]+)($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_ugcURLReg     = `(?i)(http(s)?://)?(((uat-)?www.bilibili.com)|(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn))(/video)?/((av[0-9]+)|((BV)1[1-9A-NP-Za-km-z]{9}))($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_articleURLReg = `(?i)(http(s)?://)?(uat-)?www.bilibili.com/read/((cv[0-9]+)|(native\?id=[0-9]+)|(app/[0-9]+)|(native/[0-9]+)|(mobile/[0-9]+))($|/|)([/.$*?~=#!%@&-A-Za-z0-9_]*)`
	_idReg         = `[\d]+`
	_avRex         = `(AV|av|Av|aV)[0-9]+`
	_bvRex         = `(BV|bv|Bv|bV)1[1-9A-NP-Za-km-z]{9}`
	_cvRex         = `((CV|cv|Cv|cV)[0-9]+|(mobile/[0-9]+))`

	//三点结构
	_threePointAutoPlayOpenV1Text  = "开启WiFi/免流环境下自动播放"
	_threePointAutoPlayCloseV1Text = "关闭WiFi/免流环境下自动播放"

	// 收藏定义参数
	FavNewTopicFromParam = "new_topic"
	FavTagFromParam      = "tag"

	// 新话题排序枚举
	_topicSortByHot = 2
	_topicSortByNew = 3

	// 创建话题场景区分
	_topicCreateSceneDynamic = "dynamic"
	_topicCreateSceneView    = "view"
	_topicCreateSceneTopic   = "topic"

	// 话题顶部资源类型枚举
	_timeLineResourceType = 4
	// 话题顶部资源类型枚举
	_esportCardType = 5

	// 话题卡片无结果默认引导文案
	_noCardReplyGuideText = "哇！你是第一耶！留下你的看法再走呗！"
)

// construct param
func constructGeneralParamFromCtx(ctx context.Context) *topiccardmodel.GeneralParam {
	const (
		defaultLocalTimeZone = 8
	)
	// 获取鉴权 mid
	au, _ := auth.FromContext(ctx)
	// 获取设备信息
	dev, _ := device.FromContext(ctx)
	// 获取限制条件
	limit, _ := restriction.FromContext(ctx)
	return &topiccardmodel.GeneralParam{
		Restriction: &limit,
		Device:      &dev,
		Mid:         au.Mid,
		IP:          xmetadata.String(ctx, xmetadata.RemoteIP),
		LocalTime:   defaultLocalTimeZone,
	}
}

// 构建时间轴事件
func constructTimeLineEvents(events []*pgrpc.TimeEvent, accuracy int32) []*api.TimeLineEvents {
	var res []*api.TimeLineEvents
	for _, v := range events {
		if v.Title == "" {
			continue
		}
		res = append(res, &api.TimeLineEvents{
			EventId:  v.EventId,
			Title:    v.Title,
			TimeDesc: resolveTimeEventsDesc(v.Stime, accuracy),
			JumpLink: v.JumpLink,
		})
	}
	return res
}

// 构建赛事卡
func constructEsportCard(reply *esportsgrpc.ContestsResponse, liveEntry map[int64]*livexroomgate.EntryRoomInfoResp_EntryList, upStreamRsp *topicsvc.TopicInlineResRsp, localTime int32) *api.EsportInfo {
	info := &api.EsportInfo{
		Id:        upStreamRsp.ResId,
		StartTime: upStreamRsp.StartTime,
		EndTime:   upStreamRsp.EndTime,
	}
	for _, match := range reply.Contests {
		ii := &api.MatchInfo{}
		ii.Id = match.ID
		ii.MatchStage = match.GameStage
		if match.HomeTeam == nil {
			log.Error("constructEsportCard failed, match.HomeTeam empty, reply:%+v", reply)
			return nil
		}
		// 主队信息
		ii.Home = &api.MatchTeamInfo{
			Id:    match.HomeTeam.ID,
			Title: match.HomeTeam.Title,
			Cover: match.HomeTeam.LogoFull,
			Score: match.HomeScore,
		}
		// 客队信息
		ii.Away = &api.MatchTeamInfo{
			Id:    match.AwayTeam.ID,
			Title: match.AwayTeam.Title,
			Cover: match.AwayTeam.LogoFull,
			Score: match.AwayScore,
		}
		var (
			labelText      string
			textColor      string
			textColorNight string
			stime          = match.Stime
			matchState     int32
		)
		if match.ContestStatus == esportsgrpc.ContestStatusEnum_Ing {
			matchState = 2
			labelText = "进行中"
			textColor = "#FB7299"
			textColorNight = "#BB5B76"
		} else if match.ContestStatus == esportsgrpc.ContestStatusEnum_Over {
			matchState = 3
			labelText = "已结束"
			textColor = "#999999"
			textColorNight = "#686868"
		} else {
			matchState = 1
		}
		ii.Status = matchState
		// 比赛状态文案(未开始不下发)
		if labelText != "" && matchState != 1 {
			ii.MatchLabel = &api.MatchCardDisplay{
				Text:           labelText,
				TextColor:      textColor,
				TextColorNight: textColorNight,
			}
		}
		// 比赛开始时间文案
		if timeText := model.FormMatchTime(stime, int64(localTime)); timeText != "" {
			ii.MatchTime = &api.MatchCardDisplay{
				Text: timeText,
			}
		}
		// 比赛引导按钮
		var (
			buttonText, buttonURI, liveLink string
			buttonState                     int32
		)
		buttonText, buttonURI, liveLink, buttonState = model.FormMatchState(matchState, match, liveEntry[match.LiveRoom])
		ii.MatchButton = &api.MatchCardDisplay{
			Text:     buttonText,
			Uri:      buttonURI,
			State:    buttonState,
			LiveLink: liveLink,
		}
		if ii.MatchButton.State == 1 || ii.MatchButton.State == 2 {
			ii.MatchButton.Texts = &api.Texts{
				BookingText:   "订阅",
				UnbookingText: "已订阅",
			}
			if ii.MatchButton.Uri == "" {
				ii.MatchButton.Uri = fmt.Sprintf("https://www.bilibili.com/h5/match/data/detail/%d", match.ID)
			}
		}
		info.Items = append(info.Items, ii)
	}
	return info
}

func constructTimeLineEventsJson(events []*pgrpc.TimeEvent, accuracy int32) []*model.TimeLineEvents {
	if len(events) == 0 {
		return nil
	}
	var res []*model.TimeLineEvents
	for _, v := range events {
		if v.Title == "" {
			continue
		}
		res = append(res, &model.TimeLineEvents{
			EventId:  v.EventId,
			Title:    v.Title,
			TimeDesc: resolveTimeEventsDesc(v.Stime, accuracy),
			JumpLink: v.JumpLink,
		})
	}
	return res
}

// grpc获取秒开参数
func constructPlayerArgs(ctx context.Context, general *topiccardmodel.GeneralParam, playArg *middlewarev1.PlayerArgs) context.Context {
	if playArg == nil {
		return ctx
	}
	// 获取网络信息
	net, _ := network.FromContext(ctx)
	// 秒开参数
	batchArg := middleware.MossBatchPlayArgs(playArg, *general.Device, net, general.Mid)
	return middleware.NewContext(ctx, batchArg)
}

func constructTopicSet(raw *topiccommon.TopicSet) *api.TopicSet {
	if raw == nil {
		return nil
	}
	return &api.TopicSet{
		SetId:   raw.SetId,
		SetName: raw.SetName,
		JumpUrl: raw.JumpUrl,
		Desc:    raw.Desc,
	}
}

func constructSortListReq(req *api.TopicDetailsAllReq, general *topiccardmodel.GeneralParam) *topicsvc.SortListReq {
	return &topicsvc.SortListReq{
		TopicId:         req.TopicId,
		SortBy:          req.SortBy,
		Offset:          req.Offset,
		PageSize:        req.PageSize,
		Uid:             general.Mid,
		NeedRefresh:     req.NeedRefresh,
		MetaData:        constructTopicCommonMetaDataCtrl(general, ""),
		Source:          convertToSourceReq(req.Source),
		NoIndividuation: int32(general.GetDisableRcmdInt()),
	}
}

// make item
func makeTopicItemDesc(view, discuss int64) string {
	if view == 0 || discuss == 0 {
		return ""
	}
	return topiccardmodel.StatString(view, "浏览", "") + "·" + topiccardmodel.StatString(discuss, "讨论", "")
}

func makeTopicServerConfig(ac *paladin.Map) *api.TopicServerConfig {
	if ac == nil {
		log.Warn("paladin map is nil")
		return nil
	}
	config := CustomConfig{}
	if err := ac.Get("CustomConfig").UnmarshalTOML(&config); err != nil {
		log.Warn("CustomConfig is nil, ac.Keys()=%+v", ac.Keys())
		return nil
	}
	if config.TopicServiceConfig == nil {
		return nil
	}
	return &api.TopicServerConfig{
		PubEventsIncreaseThreshold:      config.TopicServiceConfig.PubEventsIncreaseThreshold,
		PubEventsHiddenTimeoutThreshold: config.TopicServiceConfig.PubEventsHiddenTimeoutThreshold,
		VertOnlineRefreshTime:           config.TopicServiceConfig.VertOnlineRefreshTime,
	}
}

func makeTopicFoldCard(foldCount int64, foldDesc string) *api.TopicCardItem {
	return &api.TopicCardItem{
		Type: api.TopicCardType_FOLD,
		FordCardItem: &api.FoldCardItem{
			IsShowFold:   0,
			FoldCount:    foldCount,
			CardShowDesc: fmt.Sprintf("有%d条内容被折叠", foldCount),
			FoldDesc:     foldDesc,
		},
	}
}

func (s *Service) makeFunctionalCardPb(ctx context.Context, reply *topicsvc.TopicActivitiesRsp) *api.FunctionalCard {
	res := &api.FunctionalCard{
		StartTime: reply.StartTime,
		EndTime:   reply.EndTime,
	}
	// 活动引流卡
	if reply.TrafficCard != nil {
		return &api.FunctionalCard{
			TrafficCard: &api.TrafficCard{
				Name:         reply.TrafficCard.Name,
				JumpUrl:      reply.TrafficCard.JumpUrl,
				IconUrl:      reply.TrafficCard.IconUrl,
				BenefitPoint: reply.TrafficCard.BenefitPoint,
				CardDesc:     model.MakeCardTimeDesc(reply.TrafficCard),
				JumpTitle:    reply.TrafficCard.JumpTitle,
			},
		}
	}
	// 跳转胶囊
	if len(reply.Capsules) > 0 {
		for _, v := range reply.Capsules {
			res.Capsules = append(res.Capsules, &api.TopicCapsule{
				Name:    v.Name,
				JumpUrl: v.JumpUrl,
				IconUrl: v.IconUrl,
			})
		}
		return res
	}
	// 游戏下载卡
	if reply.GameCard != nil {
		gameReply, err := s.MultiGameInfos(ctx, []int64{reply.GameCard.GameId})
		if err != nil {
			log.Error("s.MultiGameInfos err=%+v", err)
			return res
		}
		if v, ok := gameReply[reply.GameCard.GameId]; ok {
			res.GameCard = &api.GameCard{
				GameId:   v.GameBaseID,
				GameIcon: v.GameIcon,
				GameName: v.GameName,
				Score:    makeFunctionalCardGameScore(v.Grade),
				GameTags: v.GameTags,
				Notice:   makeFunctionalCardGameNotice(v.Notice),
				GameLink: fmt.Sprintf("%s&sourcefrom=1000220012", v.GameLink),
			}
		}
		return res
	}
	return res
}

// 预约卡
func (s *Service) makeReserveRelationInfo(ctx context.Context, reply *topicsvc.TopicActivitiesRsp, general *topiccardmodel.GeneralParam) *api.ReserveRelationInfo {
	// 预约卡
	if reply.ReserveCard != nil {
		relationReply, err := s.actClient.UpActReserveRelationInfo(ctx, &activitygrpc.UpActReserveRelationInfoReq{
			Mid: general.Mid, Sids: []int64{reply.ReserveCard.ReserveId}, From: activitygrpc.UpCreateActReserveFrom_FROMTOPIC,
		})
		if err != nil {
			log.Error("s.makeFunctionalCardPb err=%+v", err)
			return nil
		}
		res, ok := relationReply.List[reply.ReserveCard.ReserveId]
		if !ok {
			return nil
		}
		return fromUpActReserveRelationInfo(res, reply)
	}
	return nil
}

func fromUpActReserveRelationInfo(s *activitygrpc.UpActReserveRelationInfo, reply *topicsvc.TopicActivitiesRsp) *api.ReserveRelationInfo {
	if s.Type == activitygrpc.UpActReserveRelationType_Live && time.Now().Unix() > int64(s.LivePlanStartTime) {
		// 判断直播预约的预计开播时间过去了就不下发预约卡
		return nil
	}
	i := &api.ReserveRelationInfo{}
	i.Sid = s.Sid
	i.Title = s.Title
	i.Total = s.Total
	i.Stime = int64(s.Stime)
	i.Etime = int64(s.Etime)
	i.IsFollow = s.IsFollow
	i.State = int32(s.State)
	i.Oid = s.Oid
	i.Type = int32(s.Type)
	i.Upmid = s.Upmid
	i.ReserveRecordCtime = int64(s.ReserveRecordCtime)
	i.LivePlanStartTime = int64(s.LivePlanStartTime)
	i.TimeDescText = model.ConstructReserveDescText1(s.Type, s.LivePlanStartTime, s.Desc)
	i.NumberDescText = model.ConstructReserveDescText2(s.Total)
	i.StartTime = reply.StartTime
	i.EndTime = reply.EndTime
	return i
}

func (s *Service) makeFunctionalCardJson(ctx context.Context, reply *topicsvc.TopicActivitiesRsp) *model.FunctionalCard {
	res := &model.FunctionalCard{}
	// 活动引流卡
	if reply.TrafficCard != nil {
		return &model.FunctionalCard{
			TrafficCard: &model.TrafficCard{
				Name:         reply.TrafficCard.Name,
				JumpUrl:      reply.TrafficCard.JumpUrl,
				IconUrl:      reply.TrafficCard.IconUrl,
				BenefitPoint: reply.TrafficCard.BenefitPoint,
				CardDesc:     model.MakeCardTimeDesc(reply.TrafficCard),
				JumpTitle:    reply.TrafficCard.JumpTitle,
			},
		}
	}
	// 跳转胶囊
	if len(reply.Capsules) > 0 {
		for _, v := range reply.Capsules {
			res.Capsules = append(res.Capsules, &model.TopicCapsule{
				Name:    v.Name,
				JumpUrl: v.JumpUrl,
				IconUrl: v.IconUrl,
			})
		}
		return res
	}
	// 游戏下载卡
	if reply.GameCard != nil {
		gameReply, err := s.MultiGameInfos(ctx, []int64{reply.GameCard.GameId})
		if err != nil {
			log.Error("s.MultiGameInfos err=%+v", err)
			return res
		}
		if v, ok := gameReply[reply.GameCard.GameId]; ok {
			res.GameCard = &model.GameCard{
				GameId:   v.GameBaseID,
				GameIcon: v.GameIcon,
				GameName: v.GameName,
				Score:    makeFunctionalCardGameScore(v.Grade),
				GameTags: v.GameTags,
				Notice:   makeFunctionalCardGameNotice(v.Notice),
				GameLink: fmt.Sprintf("https://www.biligame.com/detail/?id=%d&sourcefrom=1000230012", v.GameBaseID),
			}
		}
		return res
	}
	return res
}

func (s *Service) makeTopicTimeLineResource(ctx context.Context, upStreamRsp *topicsvc.TopicInlineResRsp) *api.TimeLineResource {
	if upStreamRsp == nil || upStreamRsp.ResType != _timeLineResourceType {
		return nil
	}
	reply, err := s.managerPopClient.TimeLine(ctx, &pgrpc.TimeLineRequest{LineId: upStreamRsp.ResId, Ps: 2})
	if err != nil {
		log.Error("s.managerPopClient.TimeLine upStreamRsp=%+v, err=%+v", upStreamRsp, err)
		return nil
	}
	events := constructTimeLineEvents(reply.Events, upStreamRsp.TimingAccuracy)
	if len(events) == 0 {
		return nil
	}
	return &api.TimeLineResource{
		TimeLineId:     upStreamRsp.ResId,
		TimeLineTitle:  upStreamRsp.Title,
		TimeLineEvents: events,
		HasMore:        reply.HasMore,
		StartTime:      upStreamRsp.StartTime,
		EndTime:        upStreamRsp.EndTime,
	}
}

// 赛事卡
func (s *Service) makeEsportCard(ctx context.Context, upStreamRsp *topicsvc.TopicInlineResRsp, general *topiccardmodel.GeneralParam) *api.EsportInfo {
	if upStreamRsp == nil || upStreamRsp.ResType != _esportCardType {
		return nil
	}
	reply, err := s.esportGRPC.GetContests(ctx, &esportsgrpc.GetContestsRequest{Mid: general.Mid, Cids: upStreamRsp.MatchList})
	if err != nil {
		log.Error("s.makeEsportCard upStreamRsp=%+v, err=%+v", upStreamRsp, err)
		return nil
	}
	var matchLiveRooms []int64
	for _, v := range reply.Contests {
		matchLiveRooms = append(matchLiveRooms, v.LiveRoom)
	}
	entryReq := &livexroomgate.EntryRoomInfoReq{
		EntryFrom: []string{"NONE"},
		RoomIds:   matchLiveRooms,
		Uid:       general.Mid,
		Uipstr:    general.IP,
		Platform:  general.Device.RawPlatform,
		Build:     general.Device.Build,
		Network:   "other",
	}
	imatchLiveEntryRoom, err := s.roomGateClient.EntryRoomInfo(ctx, entryReq)
	if err != nil {
		log.Error("Failed to get entry room info: %+v: %+v", entryReq, err)
		return nil
	}
	return constructEsportCard(reply, imatchLiveEntryRoom.List, upStreamRsp, general.LocalTime)
}

func makeFunctionalCardGameScore(grade float32) string {
	if grade <= 0 {
		return ""
	}
	return fmt.Sprintf("%.1f分", grade)
}

func makeFunctionalCardGameNotice(notice string) string {
	if notice == "" {
		return ""
	}
	return fmt.Sprintf("公告：%s", notice)
}

func reconstructDynSchemaContext(dynSchemaCtx *topiccardmodel.DynSchemaCtx, params []*topiccardmodel.DynMetaCardListParam) *topiccardmodel.DynSchemaCtx {
	var (
		dynCmtMeta     = make(map[int64]*topiccardmodel.DynCmtMeta, len(params))
		itemFrom       = make(map[int64]string, len(params))
		hiddenAttached = make(map[int64]bool, len(params))
		serverInfo     = make(map[int64]string, len(params))
		mergedResource = make(map[int64]topiccardmodel.MergedResource, len(params))
	)
	for _, v := range params {
		if v.DynCmtMeta != nil {
			dynCmtMeta[v.DynId] = v.DynCmtMeta
		}
		itemFrom[v.DynId] = v.ItemFrom
		hiddenAttached[v.DynId] = v.HiddenAttached
		serverInfo[v.DynId] = v.ServerInfo
		mergedResource[v.DynId] = v.MergedResource
	}
	dynSchemaCtx.DynCmtMode = dynCmtMeta
	dynSchemaCtx.ItemFrom = itemFrom
	dynSchemaCtx.HiddenAttached = hiddenAttached
	dynSchemaCtx.ServerInfo = serverInfo
	dynSchemaCtx.MergedResource = mergedResource
	return dynSchemaCtx
}

func initDynSchemaContext(ctx context.Context, topicId, sortBy int64, offset string) *topiccardmodel.DynSchemaCtx {
	return &topiccardmodel.DynSchemaCtx{
		Ctx:                      ctx,
		TopicId:                  topicId,
		SortBy:                   sortBy,
		Offset:                   offset,
		IsDisableInt64MidVersion: midint64.IsDisableInt64MidVersion(ctx),
	}
}

// convert struct
func convertToTopicInfoPb(params *topicsvc.TopicInfoRsp) *api.TopicInfo {
	raw := params.Info
	return &api.TopicInfo{
		Id:          raw.Id,
		Name:        raw.Name,
		Uid:         raw.Uid,
		View:        raw.View,
		Discuss:     raw.Discuss,
		Fav:         raw.Fav,
		Dynamics:    raw.Dynamics,
		JumpUrl:     raw.JumpUrl,
		Backcolor:   raw.Backcolor,
		IsFav:       raw.IsFav,
		Description: raw.Description,
		SharePic:    params.SharePic,
		Share:       params.Share,
		Like:        params.Like,
		ShareUrl:    params.ShareUrl,
		IsLike:      params.IsLike,
		Type:        raw.Type,
	}
}

func convertToTopicInfoJson(raw *topiccommon.TopicInfo) *model.TopicItem {
	return &model.TopicItem{
		Id:           raw.Id,
		Name:         raw.Name,
		View:         raw.View,
		Discuss:      raw.Discuss,
		Fav:          raw.Fav,
		Dynamics:     raw.Dynamics,
		JumpUrl:      raw.JumpUrl,
		BackColor:    raw.Backcolor,
		Description:  raw.Description,
		CreateSource: raw.CreateSource,
		IsFav:        raw.IsFav,
	}
}

func convertCommonTopicInfoToItems(topicList []*topiccommon.TopicInfo) []*model.TopicItem {
	topicItems := make([]*model.TopicItem, 0, len(topicList))
	for _, v := range topicList {
		topicItems = append(topicItems, &model.TopicItem{
			Id:          v.Id,
			Name:        v.Name,
			View:        v.View,
			Discuss:     v.Discuss,
			Fav:         v.Fav,
			Dynamics:    v.Dynamics,
			JumpUrl:     v.JumpUrl,
			Description: v.Description,
		})
	}
	return topicItems
}

func convertTopicListToItems(topicList []*topicsvc.TopicDetail) []*model.TopicItem {
	topicItems := make([]*model.TopicItem, 0, len(topicList))
	for _, v := range topicList {
		topicItems = append(topicItems, &model.TopicItem{
			Id:          v.TopicId,
			Name:        v.TopicName,
			View:        v.View,
			Discuss:     v.Discuss,
			Fav:         v.Fav,
			Dynamics:    v.Dynamics,
			JumpUrl:     v.JumpUrl,
			Description: v.Desc,
			RcmdIconUrl: v.IconUrl,
			RcmdText:    v.RcmdReason.GetText(),
			LancerInfo:  v.LancerInfo,
			ServerInfo:  v.ServerInfo,
			Rid:         v.Rid,
			UpId:        v.UpId,
		})
	}
	return topicItems
}

func convertToSourceReq(source string) topicsvc.ReqSource {
	switch source {
	case _sourceFromH5Details:
		return topicsvc.ReqSource_OuterH5
	case _sourceFromWebDetails:
		return topicsvc.ReqSource_Web
	case _sourceFromAPPHalf:
		// 目前APP半屏只出现在直播间
		return topicsvc.ReqSource_Live
	default:
		return topicsvc.ReqSource_APP
	}
}

func convertToLargeCoverInlineCardPb(card *jsoncard.LargeCoverInline, resType int64) *api.LargeCoverInline {
	if card == nil {
		return nil
	}
	// pb卡片公共转化部分
	res := &api.LargeCoverInline{
		Base: &cardapi.Base{
			CardType: card.CardType,
			CardGoto: card.CardGoto,
			Goto:     card.Goto,
			Param:    card.Param,
			Cover:    card.Cover,
			Title:    card.Title,
			Uri:      card.URI,
			Args: &cardapi.Args{
				Type:         int32(card.Args.Type),
				UpId:         card.Args.UpID,
				UpName:       card.Args.UpName,
				Rid:          card.Args.Rid,
				Rname:        card.Args.Rname,
				Tid:          card.Args.Tid,
				Tname:        card.Args.Tname,
				TrackId:      card.Args.TrackID,
				State:        card.Args.State,
				ConvergeType: card.Args.ConvergeType,
				Aid:          card.Args.Aid,
			},
			PlayerArgs: &cardapi.PlayerArgs{
				IsLive:    int32(card.PlayerArgs.IsLive),
				Aid:       card.PlayerArgs.Aid,
				Cid:       card.PlayerArgs.Cid,
				SubType:   card.PlayerArgs.SubType,
				RoomId:    card.PlayerArgs.RoomID,
				EpId:      card.PlayerArgs.EpID,
				IsPreview: card.PlayerArgs.IsPreview,
				Type:      card.PlayerArgs.Type,
				Duration:  card.PlayerArgs.Duration,
				SeasonId:  card.PlayerArgs.SeasonID,
			},
			UpArgs: &cardapi.UpArgs{},
			Idx:    card.Idx,
		},
		ExtraUri: card.ExtraURI,
		InlineProgressBar: &api.InlineProgressBar{
			IconDrag:     "https://i0.hdslb.com/bfs/archive/c1461e2c6ca97783ac0298b6ebb2d85d94b8f37c.json",
			IconDragHash: "31df8ce99de871afaa66a7a78f44deec",
			IconStop:     "https://i0.hdslb.com/bfs/archive/6ee2f9b016f20714705cb5b8f15da1446587d172.json",
			IconStopHash: "5648c2926c1c93eb2d30748994ba7b96",
		},
		DisableDanmu:    false,
		HideDanmuSwitch: false,
		CanPlay:         card.CanPlay,
		TopicThreePoint: &api.TopicThreePoint{
			DynThreePointItems: []*dynamicapi.ThreePointItem{dynHandler.TpAutoPlay(nil, _threePointAutoPlayOpenV1Text, _threePointAutoPlayCloseV1Text)},
		},
		RelationData: &api.RelationData{},
	}

	// 按resType各自转化结构
	switch resType {
	case _inlineResourceTypeLive:
		// 直播inline卡角标
		if card.RightTopLiveBadge != nil {
			res.RightTopLiveBadge = &api.RightTopLiveBadge{
				LiveStatus: int64(card.RightTopLiveBadge.LiveStatus),
				InLive: &api.LiveBadgeResource{
					Text:                 card.RightTopLiveBadge.InLive.Text,
					AnimationUrl:         card.RightTopLiveBadge.InLive.AnimationURL,
					AnimationUrlHash:     card.RightTopLiveBadge.InLive.AnimationURLHash,
					BackgroundColorLight: card.RightTopLiveBadge.InLive.BackgroundColorLight,
					BackgroundColorNight: card.RightTopLiveBadge.InLive.BackgroundColorNight,
					AlphaLight:           int64(card.RightTopLiveBadge.InLive.AlphaLight),
					AlphaNight:           int64(card.RightTopLiveBadge.InLive.AlphaNight),
					FontColor:            card.RightTopLiveBadge.InLive.FontColor,
				},
				LiveStatsDesc: card.CoverRightContentDescription, // 直播人气/看过
			}
		}
		// 直播inline卡左下角显示字段
		res.CoverLeftDesc = card.Args.Tname
	case _inlineResourceTypeAv, _inlineResourceTypeOgv:
		// 视频inline卡增加up主参数
		if card.UpArgs != nil {
			res.Base.UpArgs.UpId = card.UpArgs.UpID
			res.Base.UpArgs.UpName = card.UpArgs.UpName
			res.Base.UpArgs.UpFace = card.UpArgs.UpFace
			res.Base.UpArgs.Selected = int64(card.UpArgs.Selected)
		}
		// 视频inline卡左下角显示
		res.CoverLeftText_1 = fmt.Sprintf("%s观看", card.CoverLeftText1)
		res.CoverLeftIcon_1 = int32(card.CoverLeftIcon1)
		res.CoverLeftText_2 = fmt.Sprintf("%s弹幕", card.CoverLeftText2)
		res.CoverLeftIcon_2 = int32(card.CoverLeftIcon2)
		// 视频inline卡增加播放时长字段保证双端一致
		res.DurationText = appcardmodel.DurationString(card.PlayerArgs.Duration)
		// 视频inline卡转化关联资源
		res.RelationData.LikeCount = int64(card.LikeButton.Count)
		res.RelationData.IsLike = card.LikeButton.Selected == 1
		res.RelationData.IsFav = card.IsFav
		res.RelationData.IsCoin = card.IsCoin
		res.RelationData.IsFollow = card.IsAtten
	}

	return res
}

func convertToLargeCoverInlineCardJson(content *jsoncard.LargeCoverInline, resType int64) *model.LargeCoverInline {
	if content == nil {
		return nil
	}
	res := &model.LargeCoverInline{
		LargeCoverInline: content,
	}
	switch resType {
	case _inlineResourceTypeLive:
		res.LiveExtra.LiveStatsDesc = content.CoverRightContentDescription // 直播人气/看过
		res.LiveExtra.LiveStatus = content.RightTopLiveBadge.LiveStatus
	}
	return res
}

func splitInt32s(s string) ([]int32, error) {
	if s == "" {
		return nil, nil
	}
	sArr := strings.Split(s, ",")
	res := make([]int32, 0, len(sArr))
	for _, sc := range sArr {
		i, err := strconv.ParseInt(sc, 10, 64)
		if err != nil {
			return nil, err
		}
		res = append(res, int32(i))
	}
	return res, nil
}

func resolveTimeEventsDesc(stime xtime.Time, accuracy int32) string {
	const (
		accuracyMinute = 1
		accuracyHour   = 2
		accuracyMonth  = 3
		accuracyYear   = 4
		accuracyDay    = 5
	)
	timeObj := stime.Time()
	switch accuracy {
	case accuracyMinute:
		return fmt.Sprintf("%d年%02d月%02d日 %02d:%02d", timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), timeObj.Minute())
	case accuracyHour:
		return fmt.Sprintf("%d年%02d月%02d日 %02d时", timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour())
	case accuracyDay:
		return fmt.Sprintf("%d年%02d月%02d日", timeObj.Year(), timeObj.Month(), timeObj.Day())
	case accuracyMonth:
		return fmt.Sprintf("%d年%02d月", timeObj.Year(), timeObj.Month())
	case accuracyYear:
		return fmt.Sprintf("%d年", timeObj.Year())
	default:
		return fmt.Sprintf("%d年%02d月%02d日 %02d:%02d:%02d", timeObj.Year(), timeObj.Month(), timeObj.Day(), timeObj.Hour(), timeObj.Minute(), timeObj.Second())
	}
}
