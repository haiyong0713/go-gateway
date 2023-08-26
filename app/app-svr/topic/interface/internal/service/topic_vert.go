package service

import (
	"context"

	"go-common/library/log"
	errgroupv2 "go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/app/app-svr/topic/interface/internal/model"

	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"

	"github.com/pkg/errors"
)

// nolint:gomnd
func (s *Service) VertTopicOnline(ctx context.Context, params *model.VertTopicOnlineReq) (*model.VertTopicOnlineRsp, error) {
	general := constructGeneralParamFromCtx(ctx)
	reply, err := s.topicGRPC.DetailOnline(ctx, &topicsvc.DetailOnlineReq{
		TopicId:  params.TopicId,
		Mid:      general.Mid,
		MetaData: constructTopicCommonMetaDataCtrl(general, "/vert/online"),
	})
	if err != nil {
		log.Error("s.topicGRPC.DetailOnline error=%+v", err)
		return nil, nil
	}
	if reply.OnlineNum < 10 {
		// 人数小于10不出文案
		return &model.VertTopicOnlineRsp{OnlineNum: reply.OnlineNum}, nil
	}
	return &model.VertTopicOnlineRsp{
		OnlineNum:  reply.OnlineNum,
		OnlineText: topiccardmodel.StatString(reply.OnlineNum, "人正在看", ""),
	}, nil
}

func (s *Service) VertTopicCenter(ctx context.Context, params *model.VertTopicCenterReq) (*model.VertTopicCenterRsp, error) {
	res, general := &model.VertTopicCenterRsp{}, constructGeneralParamFromCtx(ctx)
	eg := errgroupv2.WithContext(ctx)
	if params.Offset == "" {
		// 首页调用
		if general.Mid > 0 {
			eg.Go(func(ctx context.Context) (err error) {
				res.FavTopics, err = s.resolveFavTopicsInTopicCenter(ctx, general)
				if err != nil {
					log.Error("s.resolveFavTopicsInTopicCenter, error=%+v", err)
					return nil
				}
				return nil
			})
		}
		eg.Go(func(ctx context.Context) (err error) {
			args := &topicsvc.HotNewTopicsReq{
				Uid:             general.Mid,
				MetaData:        constructTopicCommonMetaDataCtrl(general, params.Source),
				NoIndividuation: int32(general.GetDisableRcmdInt()),
			}
			reply, err := s.topicGRPC.HotNewTopics(ctx, args)
			if err != nil {
				log.Error("s.topicGRPC.HotNewTopics args=%+v, error=%+v", args, err)
				return nil
			}
			res.HotTopics = &model.HotTopics{HotItems: convertTopicListToItems(reply.TopicList)}
			// 获取话题展示信息
			res.HotTopics.HotItems = s.addVertTopicDescriptionDisplay(ctx, res.HotTopics.HotItems, general)
			return nil
		})
	}
	if general.Mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			reply, err := s.topicGRPC.HasCreatedTopic(ctx, &topicsvc.HasCreatedTopicReq{Uid: general.Mid})
			if err != nil {
				log.Error("s.topicGRPC.HasCreatedTopic mid=%d, error=%+v", general.Mid, err)
				return nil
			}
			if reply.HasCreated {
				res.EntranceButton = model.ConstructTopicEntranceButton()
			}
			return nil
		})
		eg.Go(func(ctx context.Context) (err error) {
			reply, err := s.topicGRPC.HasCreateJurisdiction(ctx, &topicsvc.HasCreateJurisdictionReq{Uid: general.Mid})
			if err != nil {
				log.Error("s.topicGRPC.HasCreateJurisdiction mid=%d, error=%+v", general.Mid, err)
				return nil
			}
			res.HasCreateJurisdiction = reply.HasJurisdiction
			return nil
		})
	}
	eg.Go(func(ctx context.Context) (err error) {
		args := &topicsvc.AllNewTopicsReq{
			Uid:             general.Mid,
			MetaData:        constructTopicCommonMetaDataCtrl(general, params.Source),
			NoIndividuation: int32(general.GetDisableRcmdInt()),
			PageInfo: &topicsvc.PageInfo{
				PageSize: params.PageSize,
				Offset:   params.Offset,
			},
		}
		reply, err := s.topicGRPC.AllNewTopics(ctx, args)
		if err != nil {
			log.Error("s.topicGRPC.AllNewTopics args=%+v, error=%+v", args, err)
			return nil
		}
		res.TopicItems = convertTopicListToItems(reply.TopicList)
		// 获取话题展示信息
		res.TopicItems = s.addVertTopicDescriptionDisplay(ctx, res.TopicItems, general)
		res.PageInfo = makeTopicReplyPageInfo(params.Offset, reply.PageInfo)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("VertTopicCenter eg.Wait() error=%+v", err)
		return nil, err
	}
	return res, nil
}

func makeTopicReplyPageInfo(reqOffset string, pageInfo *topicsvc.PageInfo) *topicsvc.PageInfo {
	if pageInfo == nil {
		return &topicsvc.PageInfo{Offset: reqOffset, HasMore: false}
	}
	return pageInfo
}

func (s *Service) addVertTopicDescriptionDisplay(ctx context.Context, items []*model.TopicItem, general *topiccardmodel.GeneralParam) []*model.TopicItem {
	var (
		rids, upIds []int64
	)
	for _, v := range items {
		if v.UpId != 0 {
			upIds = append(upIds, v.UpId)
		}
		if v.Rid != 0 {
			rids = append(rids, v.Rid)
		}
	}
	loader := &TopicFanoutLoader{General: general, Service: s, Dynamic: loaderDynamicSubset{DynamicIds: rids, TopicUpIds: upIds}}
	fanout, err := loader.doDynamicCardFanoutLoad(ctx)
	if err != nil {
		log.Error("addVertTopicDescriptionDisplay doDynamicCardFanoutLoad error=%+v", err)
	}
	for _, topicItem := range items {
		if topicItem.Description != "" {
			topicItem.DescriptionSubject = getDescriptionSubject(fanout, topicItem.UpId)
			topicItem.DescriptionContent = topicItem.Description
		}
		if subject, content, ok := coverDescriptionSubjectAndContent(fanout, topicItem.Rid); ok {
			topicItem.DescriptionSubject, topicItem.DescriptionContent = subject, content
		}
	}
	return items
}

func getDescriptionSubject(fanout *FanoutResult, upId int64) string {
	const (
		_defaultDescriptionSubject = "话题简介："
	)
	if fanout == nil || fanout.Dynamic.ResTopicUser == nil {
		return _defaultDescriptionSubject
	}
	user, ok := fanout.Dynamic.ResTopicUser[upId]
	if !ok || user == nil {
		return _defaultDescriptionSubject
	}
	return user.Name + "："
}

func (s *Service) resolveFavTopicsInTopicCenter(ctx context.Context, general *topiccardmodel.GeneralParam) (*model.FavTopics, error) {
	const (
		_maxTopicsFavShowNum = 9
	)
	topicsFavReply, err := s.topicGRPC.FavTopics(ctx, &topicsvc.FavTopicsReq{
		Uid:      general.Mid,
		PageInfo: &topiccommon.PaginationReq{PageSize: _maxTopicsFavShowNum},
	})
	if err != nil {
		return nil, errors.Wrapf(err, "no FavTopics resolved in topic center, general=%+v", general)
	}
	// 构造话题中心收藏物料
	if topicsFavReply != nil && len(topicsFavReply.Topics) > 0 {
		return &model.FavTopics{
			FavItems: convertToFavTopicItems(topicsFavReply.Topics),
			MoreLink: "bilibili://main/favorite?tab=topic_list",
		}, nil
	}
	return &model.FavTopics{MoreLink: "bilibili://main/favorite"}, nil
}

func (s *Service) VertSearchTopics(ctx context.Context, mid int64, params *model.VertSearchTopicsReq) (*model.VertSearchTopicsRsp, error) {
	args := &topicsvc.VertSearchTopicsReq{
		Uid:     mid,
		KeyWord: params.Keywords,
		PageInfo: &topiccommon.PaginationReq{
			PageSize: params.PageSize,
			PageNum:  params.PageNum,
			Offset:   params.Offset,
		},
	}
	reply, err := s.topicGRPC.VertSearchTopics(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.VertSearchTopics args=%+v, error=%+v", args, err)
		return nil, err
	}
	return convertToVertSearchTopicsRsp(reply), nil
}

func convertToVertSearchTopicsRsp(params *topicsvc.VertSearchTopicsRsp) *model.VertSearchTopicsRsp {
	topicItems := make([]*model.TopicItem, 0, len(params.Topics))
	for _, topic := range params.Topics {
		item := &model.TopicItem{
			Id:      topic.Id,
			Name:    topic.Name,
			JumpUrl: topic.JumpUrl,
		}
		topicItems = append(topicItems, item)
	}
	return &model.VertSearchTopicsRsp{
		TopicItems: topicItems,
		PageInfo:   params.PageInfo,
	}
}

func coverDescriptionSubjectAndContent(fanout *FanoutResult, rid int64) (string, string, bool) {
	if fanout == nil || fanout.Dynamic.ResUser == nil || fanout.Dynamic.ResDynSimpleInfo == nil {
		return "", "", false
	}
	dynSimpleInfo, ok := fanout.Dynamic.ResDynSimpleInfo[rid]
	if !ok {
		return "", "", false
	}
	user, ok := fanout.Dynamic.ResUser[dynSimpleInfo.Uid]
	if !ok {
		return "", "", false
	}
	subject := user.Name + "："
	switch dynSimpleInfo.Type {
	case dynamicV2.DynTypeVideo: // 视频卡
		if fanout.Dynamic.ResArchive[dynSimpleInfo.Rid] == nil || fanout.Dynamic.ResArchive[dynSimpleInfo.Rid].Arc == nil {
			return "", "", false
		}
		return subject, fanout.Dynamic.ResArchive[dynSimpleInfo.Rid].Arc.Title, true
	case dynamicV2.DynTypeForward, dynamicV2.DynTypeWord: //转发卡，纯文字卡
		if fanout.Dynamic.ResWords[dynSimpleInfo.Rid] == "" {
			return "", "", false
		}
		return subject, fanout.Dynamic.ResWords[dynSimpleInfo.Rid], true
	case dynamicV2.DynTypeDraw: // 图文卡
		if fanout.Dynamic.ResDraw[dynSimpleInfo.Rid] == nil || fanout.Dynamic.ResDraw[dynSimpleInfo.Rid].Item == nil {
			return "", "", false
		}
		return subject, fanout.Dynamic.ResDraw[dynSimpleInfo.Rid].Item.Description, true
	case dynamicV2.DynTypeArticle: // 专栏卡
		if fanout.Dynamic.ResArticle[dynSimpleInfo.Rid] == nil {
			return "", "", false
		}
		return subject, fanout.Dynamic.ResArticle[dynSimpleInfo.Rid].Title, true
	default:
		return "", "", false
	}
}
