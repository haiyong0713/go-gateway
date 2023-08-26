package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"go-common/component/metadata/device"
	"go-common/library/log"
	xmetadata "go-common/library/net/metadata"
	errgroupv2 "go-common/library/sync/errgroup.v2"
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	api "go-gateway/app/app-svr/topic/interface/api"
	"go-gateway/app/app-svr/topic/interface/internal/model"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	thumbgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) TopicDetailsAll(ctx context.Context, req *api.TopicDetailsAllReq) (resp *api.TopicDetailsAllReply, err error) {
	general := constructGeneralParamFromCtx(ctx)
	general.SetLocalTime(req.LocalTime)
	general.SetFromSpmid(req.FromSpmid)
	c := constructPlayerArgs(ctx, general, req.PlayerArgs)
	// 话题详情页在网关控制不同模式
	switch req.TopicDetailsExtMode {
	case api.TopicDetailsExtMode_STORY:
		return s.topicDetailsStoryModeProcess(c, general, req)
	default:
		return s.topicDetailsProcess(c, general, req)
	}
}

func (s *Service) TopicDetailsFold(ctx context.Context, req *api.TopicDetailsFoldReq) (resp *api.TopicDetailsFoldReply, err error) {
	general := constructGeneralParamFromCtx(ctx)
	general.SetLocalTime(req.LocalTime)
	c := constructPlayerArgs(ctx, general, req.PlayerArgs)
	return s.topicFoldDetailsProcess(c, general, req)
}

func (s *Service) topicDetailsProcess(ctx context.Context, general *topiccardmodel.GeneralParam, req *api.TopicDetailsAllReq) (*api.TopicDetailsAllReply, error) {
	var (
		res         = &api.TopicDetailsAllReply{}
		timeLineRes *api.TimeLineResource
		detail      = &model.TopicDetailsAll{}
	)
	arg := &topicsvc.TopicInfoReq{
		TopicId:   req.TopicId,
		Uid:       general.Mid,
		NeedShare: true,
		NeedEntry: true,
		MetaData:  constructTopicCommonMetaDataCtrl(general, ""),
	}
	topicReply, err := s.topicGRPC.TopicInfo(ctx, arg)
	if err != nil {
		log.Error("s.topicGRPC.TopicInfo arg=%+v, error=%+v", arg, err)
		return nil, err
	}
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		res.DetailsTopInfo, err = s.makeTopicTopInfo(ctx, general, req, topicReply)
		if err != nil {
			log.Error("s.makeTopicTopInfo req=%+v, reply=%+v, err=%+v", req, topicReply, err)
			return nil
		}
		res.PubLayer = s.makeTopicPubLayer(ctx, general.Mid, res.DetailsTopInfo.TopicInfo.Id, topicReply.EntryOption)
		res.TopicServerConfig = makeTopicServerConfig(s.ac)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		args := constructTopicActivitiesParams(req.TopicId, general, req.Source)
		reply, err := s.topicGRPC.TopicActivities(ctx, args)
		if err != nil {
			log.Error("s.topicGRPC.TopicActivities args=%+v, error=%+v", args, err)
			return nil
		}
		res.FunctionalCard = s.makeFunctionalCardPb(ctx, reply)
		// 获取预约卡
		detail.ReserveInfo = s.makeReserveRelationInfo(ctx, reply, general)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		args := constructSortListReq(req, general)
		reply, err := s.topicGRPC.TopicSortList(ctx, args)
		if err != nil {
			log.Error("s.topicGRPC.TopicSortList args=%+v, error=%+v", args, err)
			return nil
		}
		dynSchemaCtx := initDynSchemaContext(ctx, req.TopicId, reply.SortByConf.GetShowSortBy(), req.Offset)
		dynSchemaCtx.TopicCreatorMid = topicReply.Info.Uid
		dynSchemaCtx.OwnerAppear = topicReply.OwnerAppear
		res.TopicCardList = s.makeTopicCardList(dynSchemaCtx, general, reply)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if pd.WithContext(ctx).IsPlatIPad().And().Build("<", int64(67100000)).MustFinish() {
			// 粉pad版本控制不出时间轴
			return nil
		}
		reply, err := s.topicGRPC.TopConfig(ctx, &topicsvc.TopicInlineResReq{TopicId: req.TopicId})
		if err != nil {
			log.Error("s.topicGRPC.TopConfig req=%+v, error=%+v", req, err)
			return nil
		}
		timeLineRes = s.makeTopicTimeLineResource(ctx, reply)
		// 获取赛事卡
		detail.EsportInfo = s.makeEsportCard(ctx, reply, general)
		return nil
	})
	if req.Source == _sourceFromAPPHalf {
		// app 半屏获取在线人数
		eg.Go(func(ctx context.Context) error {
			var (
				onlineRes *model.VertTopicOnlineRsp
				err       error
			)
			params := &model.VertTopicOnlineReq{
				TopicId: req.TopicId,
			}
			if onlineRes, err = s.VertTopicOnline(ctx, params); err != nil {
				return nil
			}
			res.TopicOnline = &api.TopicOnline{
				OnlineNum:  onlineRes.OnlineNum,
				OnlineText: onlineRes.OnlineText,
			}
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		log.Error("eg.Wait() error=%+v", err)
		return nil, err
	}
	res.TimeLineResource = timeLineRes
	detail.ResAll = res
	detail.ResAll.TopicTopCards = makeTopicTopCards(detail)
	adjustTopicDetailsAllReplySent(detail.ResAll)
	return detail.ResAll, nil
}

func adjustTopicDetailsAllReplySent(res *api.TopicDetailsAllReply) {
	if res.GetTopicCardList() != nil && res.GetTopicCardList().GetNoCardResultReply() != nil && res.GetTopicCardList().GetNoCardResultReply().GetShowButton() != nil {
		res.TopicCardList.NoCardResultReply.ShowButton.JumpUrl = res.GetPubLayer().GetJumpLink()
	}
	if res.GetFunctionalCard() != nil && res.GetDetailsTopInfo().GetOperationContent().GetOperationCard().GetLargeCoverInline() != nil && res.GetTimeLineResource() != nil {
		// 当同时配置了功能卡（功能卡除掉胶囊外），顶部视频和时间轴时，隐藏功能卡
		if len(res.GetFunctionalCard().GetCapsules()) == 0 {
			res.FunctionalCard = nil
		}
	}
}

// 封装话题顶部卡
func makeTopicTopCards(details *model.TopicDetailsAll) []*api.TopicTopCard {
	var topicTopCards []*api.TopicTopCard
	//排序规则：
	//1.功能卡：预约、活动引流、游戏转化卡、跳转胶囊
	//2.播放器：视频、直播间
	//3.内容卡：赛程、时间轴
	//三类卡内部彼此互斥（即同一个话题不可以配2个同类卡）
	//三类卡之间可以随意排列组合，最多2个（取最新生效的两个）
	// 功能卡
	if card := getFunctionalCard(details); card != nil {
		topicTopCards = append(topicTopCards, card)
	}
	// 播放器卡
	if card := getInlineCard(details); card != nil {
		topicTopCards = append(topicTopCards, card)
	}
	// 内容卡
	if card := getContentCard(details); card != nil {
		topicTopCards = append(topicTopCards, card)
	}
	// 配置多于2个,取最新生效的两个
	if topicTopCards != nil && len(topicTopCards) > 2 {
		sort.SliceStable(topicTopCards, func(i, j int) bool {
			// 最新生效的在前
			return topicTopCards[i].GetStartTime() > topicTopCards[j].GetStartTime()
		})
		// 排序之后取前面的两个
		topicTopCards = topicTopCards[:2]
	}
	return topicTopCards
}

// 获取功能卡
func getFunctionalCard(details *model.TopicDetailsAll) *api.TopicTopCard {
	var topicTopCard *api.TopicTopCard
	// 跳转胶囊卡
	if card := details.ResAll.GetFunctionalCard().GetCapsules(); card != nil {
		topicTopCard = &api.TopicTopCard{
			Type:      api.TopicTopCardType_Capsules_Type,
			StartTime: details.ResAll.GetFunctionalCard().StartTime,
			EndTime:   details.ResAll.GetFunctionalCard().EndTime,
			CardItem: &api.TopicTopCard_Capsules{
				Capsules: &api.TopicCapsuleInfo{
					Capsules: card,
				},
			},
		}
		return topicTopCard
	}
	// 引流卡
	if card := details.ResAll.GetFunctionalCard().GetTrafficCard(); card != nil {
		topicTopCard = &api.TopicTopCard{
			Type:      api.TopicTopCardType_Traffic_Card_Type,
			StartTime: details.ResAll.GetFunctionalCard().StartTime,
			EndTime:   details.ResAll.GetFunctionalCard().EndTime,
			CardItem: &api.TopicTopCard_TrafficCard{
				TrafficCard: card,
			},
		}
		return topicTopCard
	}
	// 游戏下载卡
	if card := details.ResAll.GetFunctionalCard().GetGameCard(); card != nil {
		topicTopCard = &api.TopicTopCard{
			Type:      api.TopicTopCardType_Game_Card_Type,
			StartTime: details.ResAll.GetFunctionalCard().StartTime,
			EndTime:   details.ResAll.GetFunctionalCard().EndTime,
			CardItem: &api.TopicTopCard_GameCard{
				GameCard: card,
			},
		}
		return topicTopCard
	}
	// 预约卡
	if details.ReserveInfo != nil {
		topicTopCard = &api.TopicTopCard{
			Type:      api.TopicTopCardType_Reservation_Card_Type,
			StartTime: details.ReserveInfo.StartTime,
			EndTime:   details.ReserveInfo.EndTime,
			CardItem: &api.TopicTopCard_ReservationCard{
				ReservationCard: details.ReserveInfo,
			},
		}
		return topicTopCard
	}
	return nil
}

// 获取播放器inline卡
func getInlineCard(details *model.TopicDetailsAll) *api.TopicTopCard {
	var topicTopCard *api.TopicTopCard
	// inline
	if card := details.ResAll.GetDetailsTopInfo().GetOperationContent().GetOperationCard().GetLargeCoverInline(); card != nil {
		topicTopCard = &api.TopicTopCard{
			Type:      api.TopicTopCardType_Large_Cover_Inline_Type,
			StartTime: details.ResAll.GetDetailsTopInfo().GetOperationContent().StartTime,
			EndTime:   details.ResAll.GetDetailsTopInfo().GetOperationContent().EndTime,
			CardItem: &api.TopicTopCard_LargeCoverInline{
				LargeCoverInline: card,
			},
		}
		return topicTopCard
	}
	return nil
}

// 获取内容卡
func getContentCard(details *model.TopicDetailsAll) *api.TopicTopCard {
	var topicTopCard *api.TopicTopCard
	// 时间轴
	if card := details.ResAll.GetTimeLineResource(); card != nil {
		topicTopCard = &api.TopicTopCard{
			Type:      api.TopicTopCardType_Time_Line_Type,
			StartTime: details.ResAll.GetTimeLineResource().StartTime,
			EndTime:   details.ResAll.GetTimeLineResource().EndTime,
			CardItem: &api.TopicTopCard_TimeLineResource{
				TimeLineResource: card,
			},
		}
		return topicTopCard
	}
	// 赛程卡
	if details.EsportInfo != nil {
		topicTopCard = &api.TopicTopCard{
			Type:      api.TopicTopCardType_Esport_Card_Type,
			StartTime: details.EsportInfo.StartTime,
			EndTime:   details.EsportInfo.EndTime,
			CardItem: &api.TopicTopCard_EsportCard{
				EsportCard: details.EsportInfo,
			},
		}
		return topicTopCard
	}
	return nil
}

func (s *Service) makeTopicPubLayer(ctx context.Context, userMid, topicId int64, entryOption int32) *api.PubLayer {
	if pd.WithContext(ctx).IsOverseas().MustFinish() {
		// 繁体版只进动态发布
		return &api.PubLayer{ShowType: 1, JumpLink: fmt.Sprintf("bilibili://following/publish?topicV2ID=%d", topicId)}
	}
	if entryOption == 0 {
		return &api.PubLayer{ClosePubLayerEntry: true}
	}
	return s.makeTopicPubLayerWithAvatar(ctx, userMid, topicId)
}

func (s *Service) makeTopicPubLayerWithAvatar(ctx context.Context, userMid, topicId int64) *api.PubLayer {
	const (
		_unfoldAvatar      = 20
		_unfoldDefaultIcon = 21
	)
	res := &api.PubLayer{ShowType: _unfoldDefaultIcon, JumpLink: fmt.Sprintf("bilibili://uper/center_plus?topic_id=%d&tab_index=4&relation_from=topic&jump_from=topic", topicId)}
	if userMid == 0 {
		return res
	}
	info, err := s.accGRPC.Info3(ctx, &accgrpc.MidReq{Mid: userMid, RealIp: xmetadata.String(ctx, xmetadata.RemoteIP)})
	if err != nil {
		log.Error("s.accGRPC.Info3 mid=%d, err=%+v", userMid, err)
		return res
	}
	res.ShowType = _unfoldAvatar
	res.UserAvatar = info.GetInfo().GetFace()
	return res
}

func (s *Service) topicFoldDetailsProcess(ctx context.Context, general *topiccardmodel.GeneralParam, req *api.TopicDetailsFoldReq) (*api.TopicDetailsFoldReply, error) {
	args := &topicsvc.FoldListReq{
		TopicId:    req.TopicId,
		FromSortBy: req.FromSortBy,
		Offset:     req.Offset,
		PageSize:   req.PageSize,
	}
	reply, err := s.topicGRPC.TopicFoldList(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.TopicFoldList args=%+v, error=%+v", args, err)
		return nil, err
	}
	dynSchemaCtx := initDynSchemaContext(ctx, req.TopicId, req.FromSortBy, req.Offset)
	cardItems, err := s.dynCardProcess(dynSchemaCtx, general, convertFoldListToDynMetaCardListItem(reply.Items))
	if err != nil {
		return nil, err
	}
	return &api.TopicDetailsFoldReply{
		TopicCardList: &api.TopicCardList{
			TopicCardItems: cardItems,
			Offset:         reply.Offset,
			HasMore:        reply.HasMore,
		},
		FoldCount: int64(len(cardItems)),
	}, nil
}

func constructTopicCommonMetaDataCtrl(general *topiccardmodel.GeneralParam, from string) *topiccommon.MetaDataCtrl {
	switch from {
	case _sourceFromWebDetails, _sourceFromH5Details, _sourceFromWebDynamicFeed:
		// 特制一些来源传MetaData
		return &topiccommon.MetaDataCtrl{
			Platform: "Web",
			From:     from,
		}
	default:
		var networkType int32
		if general.Device != nil {
			networkType = general.Device.NetworkType
		}
		return &topiccommon.MetaDataCtrl{
			Platform:     general.GetPlatform(),
			Build:        general.GetBuildStr(),
			MobiApp:      general.GetMobiApp(),
			Buvid:        general.GetBuvid(),
			Device:       general.GetDevice(),
			From:         from,
			Network:      networkType,
			Ip:           general.IP,
			TeenagerMode: int32(general.GetTeenagerInt()),
			FromSpmid:    general.FromSpmid,
		}
	}
}

func constructTopicActivitiesParams(topicId int64, general *topiccardmodel.GeneralParam, from string) *topicsvc.TopicActivitiesReq {
	metaDataCtrl := constructTopicCommonMetaDataCtrl(general, from)
	return &topicsvc.TopicActivitiesReq{
		TopicId:  topicId,
		Metadata: metaDataCtrl,
	}
}

func (s *Service) makeTopicTopInfo(ctx context.Context, general *topiccardmodel.GeneralParam, req *api.TopicDetailsAllReq, reply *topicsvc.TopicInfoRsp) (*api.DetailsTopInfo, error) {
	topInfo := &api.DetailsTopInfo{
		TopicInfo:        convertToTopicInfoPb(reply),
		StatsDesc:        makeTopicItemDesc(reply.Info.View, reply.Info.Discuss),
		HeadImgUrl:       reply.HeadImgUrl,
		HeadImgBackcolor: reply.HeadImgBackcolor,
		WordColor:        reply.WordColor,
		TopicSet:         constructTopicSet(reply.TopicSet),
		Symbol:           reply.Symbol,
	}
	if req.TopicDetailsExtMode == api.TopicDetailsExtMode_STORY && reply.MissionUrl != "" {
		topInfo.MissionText = "活动详情"
		topInfo.MissionUrl = reply.MissionUrl
		topInfo.MissionPageShowType = convertToMissionPageShowType(reply.MissionPageType)
	}
	eg := errgroupv2.WithContext(ctx)
	if reply.Info.Uid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			midReply, err := s.accGRPC.Card3(ctx, &accgrpc.MidReq{
				Mid:    reply.Info.Uid,
				RealIp: xmetadata.String(ctx, xmetadata.RemoteIP),
			})
			if err != nil {
				log.Error("s.accGRPC.Card3 reply.Info=%+v, error=%+v", reply.Info, err)
				return nil
			}
			topInfo.User = &api.User{
				Uid:      reply.Info.Uid,
				Face:     midReply.Card.Face,
				Name:     midReply.Card.Name,
				NameDesc: "发起",
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) (err error) {
		jurisdictionReply, err := s.topicGRPC.HasCreateJurisdiction(ctx, &topicsvc.HasCreateJurisdictionReq{
			Uid: general.Mid,
		})
		if err != nil {
			log.Error("s.topicGRPC.HasCreateJurisdiction mid=%d, err=%+v", general.Mid, err)
			return nil
		}
		topInfo.HasCreateJurisdiction = jurisdictionReply.HasJurisdiction
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		inlineRes, err := s.topicGRPC.TopicInlineRes(ctx, &topicsvc.TopicInlineResReq{TopicId: reply.Info.Id})
		if err != nil {
			log.Error("s.topicGRPC.TopicInlineRes mid=%d, err=%+v", general.Mid, err)
			return nil
		}
		topInfo.OperationContent = &api.OperationContent{
			StartTime: inlineRes.StartTime,
			EndTime:   inlineRes.EndTime,
			OperationCard: &api.OperationCard{Card: &api.OperationCard_LargeCoverInline{
				LargeCoverInline: convertToLargeCoverInlineCardPb(s.makeOperationContent(ctx, general, inlineRes, &model.CommonDetailsParams{
					TopicId: req.TopicId,
					SortBy:  req.SortBy,
					Offset:  req.Offset,
				}), inlineRes.ResType)},
			},
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("makeTopicTopInfo eg.Wait() error=%+v", err)
		return nil, err
	}
	return topInfo, nil
}

// nolint:gomnd
func convertToMissionPageShowType(pageType int32) int32 {
	// MissionPageShowType(1.半屏na活动页 2.NA活动页 3.半屏h5 4.全屏h5)
	// MissionPageType(1.简版活动页 2.na活动页)
	const (
		halfNaPageShowType = 1
		naPageShowType     = 2
		halfH5ShowType     = 3
		allH5ShowType      = 4
	)
	switch pageType {
	case 1:
		return halfNaPageShowType
	case 2:
		return naPageShowType
	case 3:
		return allH5ShowType
	default:
		return halfH5ShowType
	}
}

func (s *Service) makeOperationContent(ctx context.Context, general *topiccardmodel.GeneralParam, rawResource *topicsvc.TopicInlineResRsp, req *model.CommonDetailsParams) *jsoncard.LargeCoverInline {
	if rawResource.ResId == 0 {
		return nil
	}
	loader := &TopicFanoutLoader{General: constructGeneralParamFromCtx(ctx), Service: s}
	switch rawResource.ResType {
	case _inlineResourceTypeOgv:
		loader.Bangumi.EPID = []int32{int32(rawResource.ResId)}
		fanout, err := loader.doTopicCardFanoutLoad(ctx)
		if err != nil {
			log.Error("makeOperationContent doTopicCardFanoutLoad OGV resId=%+v, error=%+v", rawResource.ResId, err)
			return nil
		}
		card, err := buildTopicOgvInlineCard(ctx, rawResource, fanout)
		if err != nil {
			log.Error("makeOperationContent buildTopicLiveInlineCard resId=%+v, error=%+v", rawResource.ResId, err)
			return nil
		}
		return card
	case _inlineResourceTypeLive:
		loader.Live.InlineRoomIDs = []int64{rawResource.ResId}
		fanout, err := loader.doTopicCardFanoutLoad(ctx)
		if err != nil {
			log.Error("makeOperationContent doTopicCardFanoutLoad Live resId=%+v, error=%+v", rawResource.ResId, err)
			return nil
		}
		card, err := buildTopicLiveInlineCard(ctx, rawResource, fanout)
		if err != nil {
			log.Error("makeOperationContent buildTopicLiveInlineCard resId=%+v, error=%+v", rawResource.ResId, err)
			return nil
		}
		return card
	case _inlineResourceTypeAv:
		loader.Archive.Aids = []int64{rawResource.ResId}
		fanout, err := loader.doTopicCardFanoutLoad(ctx)
		if err != nil {
			log.Error("makeOperationContent doTopicCardFanoutLoad Av resId=%+v, error=%+v", rawResource.ResId, err)
			return nil
		}
		card, err := buildTopicAvInlineCard(ctx, general, rawResource, fanout, req)
		if err != nil {
			log.Error("makeOperationContent buildTopicAvInlineCard resId=%+v, error=%+v", rawResource.ResId, err)
			return nil
		}
		return card
	default:
		log.Warn("makeOperationContent unrecognized inlineResource(%+v) type", rawResource)
		return nil
	}
}

func (s *Service) makeTopicCardList(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam, sortListMeta *topicsvc.SortListRsp) *api.TopicCardList {
	cardItems, err := s.dynCardProcess(dynSchemaCtx, general, convertSortListToDynMetaCardListItem(sortListMeta.Items))
	if err != nil || len(cardItems) == 0 {
		return &api.TopicCardList{NoCardResultReply: &api.NoCardResultReply{DefaultGuideText: _noCardReplyGuideText, ShowButton: &api.ShowButton{ShowText: "参与话题"}}}
	}
	// 折叠条逻辑
	if !sortListMeta.HasMore && sortListMeta.IsShowFold == 1 && sortListMeta.FoldCount > 0 {
		cardItem := makeTopicFoldCard(sortListMeta.FoldCount, sortListMeta.FoldDesc)
		cardItems = append(cardItems, cardItem)
	}
	// 相关话题插入
	if sortListMeta.RelatedTopics != nil {
		cardItems = insertRelatedTopicsDynamicItem(cardItems, sortListMeta.RelatedTopics)
	}
	return &api.TopicCardList{
		TopicCardItems: cardItems,
		Offset:         sortListMeta.Offset,
		HasMore:        sortListMeta.HasMore,
		TopicSortByConf: &api.TopicSortByConf{
			DefaultSortBy: sortListMeta.SortByConf.DefaultSortBy,
			AllSortBy:     convertAllSortBy(sortListMeta.SortByConf.AllSortBy),
			ShowSortBy:    sortListMeta.SortByConf.ShowSortBy,
		},
	}
}

func insertRelatedTopicsDynamicItem(cardItems []*api.TopicCardItem, relatedTopic *topicsvc.RelatedTopic) []*api.TopicCardItem {
	if len(relatedTopic.Topics) == 0 || len(relatedTopic.Topics) > 3 || relatedTopic.ShowIndex < 0 || int(relatedTopic.ShowIndex) > len(cardItems) {
		return cardItems
	}
	modules := []*dynamicapi.Module{
		{
			ModuleType: dynamicapi.DynModuleType_module_title,
			ModuleItem: &dynamicapi.Module_ModuleTitle{
				ModuleTitle: &dynamicapi.ModuleTitle{
					Title: "相关话题",
				},
			},
		},
	}
	for _, v := range relatedTopic.Topics {
		desc := makeTopicItemDesc(v.View, v.Discuss)
		if desc == "" {
			desc = v.Description
		}
		module := &dynamicapi.Module{
			ModuleType: dynamicapi.DynModuleType_module_topic_brief,
			ModuleItem: &dynamicapi.Module_ModuleTopicBrief{
				ModuleTopicBrief: &dynamicapi.ModuleTopicBrief{
					Topic: &dynamicapi.TopicItem{
						TopicId:   v.Id,
						TopicName: v.Name,
						Url:       v.JumpUrl,
						Desc_2:    desc,
					},
				},
			},
		}
		modules = append(modules, module)
	}
	// 插入相关话题
	cardItems = append(cardItems[:relatedTopic.ShowIndex], append([]*api.TopicCardItem{
		{
			Type: api.TopicCardType_DYNAMIC,
			DynamicItem: &dynamicapi.DynamicItem{
				CardType: dynamicapi.DynamicType_topic_rcmd,
				Modules:  modules,
			},
		},
	}, cardItems[relatedTopic.ShowIndex:]...)...)
	return cardItems
}

func convertAllSortBy(raws []*topicsvc.SortContent) []*api.SortContent {
	var res []*api.SortContent
	for _, raw := range raws {
		res = append(res, &api.SortContent{
			SortBy:   raw.SortBy,
			SortName: raw.SortName,
		})
	}
	return res
}

func convertSortListToDynMetaCardListItem(sortList []*topicsvc.SortListItem) []*topiccardmodel.DynMetaCardListParam {
	var res []*topiccardmodel.DynMetaCardListParam
	for _, v := range sortList {
		if v.Type != _dynamicCardType {
			continue
		}
		res = append(res, &topiccardmodel.DynMetaCardListParam{
			DynId: v.Rid,
			DynCmtMeta: &topiccardmodel.DynCmtMeta{
				CmtShowStat: v.CmtShowStat,
				CmtMode:     v.CmtMode,
			},
			ItemFrom:       v.ItemFrom,
			HiddenAttached: v.HiddenAttached,
			ServerInfo:     v.ServerInfo,
			MergedResource: topiccardmodel.MergedResource{MergeType: v.MergeType, MergedResCnt: v.MergedResCnt},
		})
	}
	return res
}

func convertFoldListToDynMetaCardListItem(foldList []*topicsvc.FoldListItem) []*topiccardmodel.DynMetaCardListParam {
	var res []*topiccardmodel.DynMetaCardListParam
	for _, v := range foldList {
		if v.Type != _dynamicCardType {
			continue
		}
		res = append(res, &topiccardmodel.DynMetaCardListParam{
			DynId: v.Rid,
			DynCmtMeta: &topiccardmodel.DynCmtMeta{
				CmtShowStat: v.CmtShowStat,
				CmtMode:     v.CmtMode,
			},
		})
	}
	return res
}

func (s *Service) TopicReport(ctx context.Context, mid int64, params *model.TopicReportReq) (*empty.Empty, error) {
	args := &topicsvc.TopicReportReq{
		Uid:     mid,
		TopicId: params.TopicId,
		Reason:  params.Reason,
	}
	if _, err := s.topicGRPC.TopicReport(ctx, args); err != nil {
		log.Error("s.topicGRPC.TopicReport args=%+v, error=%+v", args, err)
		return nil, nil
	}
	return nil, nil
}

func (s *Service) TopicResReport(ctx context.Context, mid int64, params *model.TopicResReportReq) (*empty.Empty, error) {
	if params.ResId == 0 {
		resId, err := strconv.ParseInt(params.ResIdStr, 10, 64)
		if err != nil {
			log.Error("TopicResReport ResIdStr parse params=%+v, error=%+v", params, err)
			return nil, nil
		}
		params.ResId = resId
	}
	args := &topicsvc.TopicResReportReq{
		Uid:     mid,
		TopicId: params.TopicId,
		ResId:   params.ResId,
		ResType: params.ResType,
		Reason:  params.Reason,
	}
	if _, err := s.topicGRPC.TopicResReport(ctx, args); err != nil {
		log.Error("s.topicGRPC.TopicResReport args=%+v, error=%+v", args, err)
		return nil, nil
	}
	return nil, nil
}

func (s *Service) TopicLike(ctx context.Context, mid int64, device device.Device, params *model.TopicLikeReq) (*empty.Empty, error) {
	args := &thumbgrpc.LikeReq{
		Business:  params.Business,
		Mid:       mid,
		UpMid:     params.UpMid,
		MessageID: params.TopicId,
		IP:        xmetadata.RemoteIP,
		MobiApp:   device.RawMobiApp,
		Platform:  device.RawPlatform,
		Device:    device.Device,
	}
	switch params.Action {
	case "like":
		args.Action = thumbgrpc.Action_ACTION_LIKE
	case "cancel_like":
		args.Action = thumbgrpc.Action_ACTION_CANCEL_LIKE
	default:
		log.Error("unrecognized topic like action, mid=%d, params=%+v", mid, params)
		return nil, nil
	}
	if _, err := s.thumbGRPC.Like(ctx, args); err != nil {
		log.Error("s.thumbGRPC.Like args=%+v, error=%+v", args, err)
		return nil, err
	}
	return nil, nil
}

func (s *Service) TopicDisLike(ctx context.Context, mid int64, params *model.TopicDislikeReq) (*empty.Empty, error) {
	args := &topicsvc.DislikeReq{
		Uid:     mid,
		TopicId: params.TopicId,
	}
	if _, err := s.topicGRPC.Dislike(ctx, args); err != nil {
		log.Error("s.topicGRPC.Dislike args=%+v, error=%+v", args, err)
		return nil, nil
	}
	return nil, nil
}
