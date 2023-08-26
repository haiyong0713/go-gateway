package service

import (
	"context"

	"go-common/library/log"
	xmetadata "go-common/library/net/metadata"
	errgroupv2 "go-common/library/sync/errgroup.v2"

	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	"go-gateway/app/app-svr/topic/interface/internal/model"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

func (s *Service) WebTopicInfo(ctx context.Context, mid int64, params *model.WebTopicInfoReq) (*model.WebTopicInfoRsp, error) {
	// 获取并预处理话题信息
	args := &topicsvc.TopicInfoReq{
		TopicId:   params.TopicId,
		Uid:       mid,
		NeedShare: true,
		NeedEntry: true,
	}
	reply, err := s.topicGRPC.TopicInfo(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.TopicInfo args=%+v, error=%+v", args, err)
		return nil, err
	}
	topInfo, general := webTopDetailsPreProcess(reply), constructGeneralParamFromCtx(ctx)
	// 填充接口信息
	res := &model.WebTopicInfoRsp{}
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		actArgs := constructTopicActivitiesParams(params.TopicId, general, params.Source)
		actReply, err := s.topicGRPC.TopicActivities(ctx, actArgs)
		if err != nil {
			log.Error("s.topicGRPC.TopicActivities args=%+v, error=%+v", args, err)
			return nil
		}
		res.FunctionalCard = s.makeFunctionalCardJson(ctx, actReply)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		inlineRes, err := s.topicGRPC.TopicInlineRes(ctx, &topicsvc.TopicInlineResReq{TopicId: reply.Info.Id})
		if err != nil {
			log.Error("s.topicGRPC.TopicInlineRes mid=%d, err=%+v", mid, err)
			return nil
		}
		topInfo.OperationContent = &model.OperationContent{LargeCoverInline: convertToLargeCoverInlineCardJson(s.makeOperationContent(ctx, general, inlineRes, &model.CommonDetailsParams{
			TopicId: params.TopicId,
			Source:  params.Source,
		}), inlineRes.ResType)}
		return nil
	})
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
			topInfo.TopicCreator = &model.TopicCreator{
				Uid:  reply.Info.Uid,
				Face: midReply.Card.Face,
				Name: midReply.Card.Name,
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) (err error) {
		JurisdictionReply, err := s.topicGRPC.HasCreateJurisdiction(ctx, &topicsvc.HasCreateJurisdictionReq{
			Uid: mid,
		})
		if err != nil {
			log.Error("s.topicGRPC.HasCreateJurisdiction mid=%d, error=%+v", mid, err)
			return nil
		}
		topInfo.HasCreateJurisdiction = JurisdictionReply.HasJurisdiction
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	res.TopDetails = topInfo
	return res, nil
}

func webTopDetailsPreProcess(reply *topicsvc.TopicInfoRsp) *model.TopDetails {
	topicItem := convertToTopicInfoJson(reply.Info)
	topicItem.SharePic = reply.SharePic
	topicItem.Share = reply.Share
	topicItem.ShareUrl = reply.ShareUrl
	topicItem.Like = reply.Like
	topicItem.IsLike = reply.IsLike
	return &model.TopDetails{
		TopicItem: topicItem, HeadImgUrl: reply.HeadImgUrl, HeadImgBackcolor: reply.HeadImgBackcolor, WordColor: reply.WordColor,
		ClosePubLayerEntry: reply.EntryOption == 0,
	}
}

func (s *Service) WebTopicCards(ctx context.Context, params *model.WebTopicCardsReq) (*model.WebTopicCardsRsp, error) {
	var (
		reply      *topicsvc.SortListRsp
		topicReply *topicsvc.TopicInfoRsp
	)
	general := constructGeneralParamFromCtx(ctx)
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		args := &topicsvc.SortListReq{
			TopicId:     params.TopicId,
			SortBy:      params.SortBy,
			Offset:      params.Offset,
			PageSize:    params.PageSize,
			Uid:         general.Mid,
			NeedRefresh: params.NeedRefresh,
			MetaData:    constructTopicCommonMetaDataCtrl(general, _sourceFromWebDetails),
			Source:      convertToSourceReq(params.Source),
		}
		reply, err = s.topicGRPC.TopicSortList(ctx, args)
		if err != nil {
			log.Error("s.topicGRPC.TopicSortList args=%+v, error=%+v", args, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		// 获取话题信息
		topicReply, err = s.topicGRPC.TopicInfo(ctx, &topicsvc.TopicInfoReq{
			TopicId:   params.TopicId,
			Uid:       general.Mid,
			NeedShare: false,
			NeedEntry: false,
			MetaData:  constructTopicCommonMetaDataCtrl(general, _sourceFromWebDetails),
		})
		if err != nil {
			log.Error("s.topicGRPC.TopicInfo params=%+v, error=%+v", params, err)
			return nil
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	dynSchemaCtx := initDynSchemaContext(ctx, params.TopicId, reply.SortByConf.GetShowSortBy(), params.Offset)
	if topicReply != nil && topicReply.Info != nil {
		dynSchemaCtx.TopicCreatorMid = topicReply.Info.Uid
		dynSchemaCtx.OwnerAppear = topicReply.OwnerAppear
	}
	cardItems, err := s.dynWebCardProcess(dynSchemaCtx, general, convertSortListToDynMetaCardListItem(reply.Items))
	if err != nil {
		log.Error("s.dynWebCardProcess params=%+v, error=%+v", params, err)
		return nil, err
	}
	res := &model.WebTopicCardsRsp{TopicCardList: &model.TopicCardList{
		Items:   mergeTopicCards(cardItems, reply),
		Offset:  reply.Offset,
		HasMore: reply.HasMore,
		TopicSortByConf: &model.TopicSortByConf{
			DefaultSortBy: reply.SortByConf.DefaultSortBy,
			AllSortBy:     reply.SortByConf.AllSortBy,
			ShowSortBy:    reply.SortByConf.ShowSortBy,
		},
	}}
	res.RelatedTopics = &model.RelatedTopics{TopicItems: convertCommonTopicInfoToItems(reply.RelatedTopics.GetTopics())}
	return res, nil
}

func (s *Service) WebTopicFoldCards(ctx context.Context, params *model.WebTopicFoldCardsReq) (*model.WebTopicFoldCardsRsp, error) {
	args := &topicsvc.FoldListReq{
		TopicId:    params.TopicId,
		FromSortBy: params.FromSortBy,
		Offset:     params.Offset,
		PageSize:   params.PageSize,
	}
	reply, err := s.topicGRPC.TopicFoldList(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.TopicFoldList args=%+v, error=%+v", args, err)
		return nil, err
	}
	dynSchemaCtx := initDynSchemaContext(ctx, params.TopicId, params.FromSortBy, params.Offset)
	cardItems, err := s.dynWebCardProcess(dynSchemaCtx, constructGeneralParamFromCtx(ctx), convertFoldListToDynMetaCardListItem(reply.Items))
	if err != nil {
		log.Error("s.dynWebCardProcess args=%+v, error=%+v", args, err)
		return nil, err
	}
	return &model.WebTopicFoldCardsRsp{TopicCardList: &model.TopicCardList{
		Items:   makeDynamicCardList(cardItems),
		Offset:  reply.Offset,
		HasMore: reply.HasMore,
	}}, nil
}

func mergeTopicCards(items []*jsonwebcard.TopicCard, sortListMeta *topicsvc.SortListRsp) []*model.TopicCardItem {
	// 先插入动态卡
	cardItems := makeDynamicCardList(items)
	// 判断折叠条逻辑
	if !sortListMeta.HasMore && sortListMeta.IsShowFold == 1 && sortListMeta.FoldCount > 0 {
		cardItem := makeTopicWebFoldCard(sortListMeta.FoldCount, sortListMeta.FoldDesc)
		cardItems = append(cardItems, &model.TopicCardItem{TopicType: _topicCardFold, FoldCardItem: &cardItem})
	}
	return cardItems
}

func makeDynamicCardList(items []*jsonwebcard.TopicCard) []*model.TopicCardItem {
	var cardItems []*model.TopicCardItem
	for _, v := range items {
		cardItems = append(cardItems, &model.TopicCardItem{TopicType: _topicCardDynamic, DynamicCardItem: v})
	}
	return cardItems
}

func makeTopicWebFoldCard(foldCount int64, foldDesc string) jsonwebcard.TopicCard {
	return jsonwebcard.ConstructTopicFoldCard(foldCount, foldDesc)
}

func (s *Service) WebSubFavTopics(ctx context.Context, mid int64, params *model.WebFavSubListReq) (*model.WebFavSubListRsp, error) {
	res, err := s.subNewTopicsList(ctx, mid, &model.FavSubListReq{PageNum: params.PageNum, PageSize: params.PageSize})
	if err != nil {
		log.Error("WebSubFavTopics mid=%d, error=%+v", mid, err)
		return nil, err
	}
	return &model.WebFavSubListRsp{TopicList: res}, nil
}

func (s *Service) WebDynamicRcmdTopics(ctx context.Context, params *model.WebDynamicRcmdReq) (*model.WebDynamicRcmdRsp, error) {
	general := constructGeneralParamFromCtx(ctx)
	args := &topicsvc.RcmdNewTopicsReq{
		Uid:             general.Mid,
		MetaData:        constructTopicCommonMetaDataCtrl(general, _sourceFromWebDynamicFeed),
		PageSize:        params.PageSize,
		NoIndividuation: int32(general.GetDisableRcmdInt()),
	}
	res, err := s.topicGRPC.RcmdNewTopics(ctx, args)
	if err != nil {
		log.Error("WebDynamicRcmdTopics args=%+v, error=%+v", args, err)
		return nil, err
	}
	return &model.WebDynamicRcmdRsp{TopicItems: convertTopicListToItems(res.TopicList)}, nil
}
