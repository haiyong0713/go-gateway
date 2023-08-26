package service

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"
	errgroupv2 "go-common/library/sync/errgroup.v2"

	topicecode "go-gateway/app/app-svr/topic/ecode"
	"go-gateway/app/app-svr/topic/interface/internal/model"

	dyncommonapi "git.bilibili.co/bapis/bapis-go/dynamic/common"
	dyntopicapi "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	topicextapi "git.bilibili.co/bapis/bapis-go/dynamic/service/topic-ext"
	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"

	"github.com/pkg/errors"
)

func (s *Service) SubFavTopics(ctx context.Context, mid int64, params *model.FavSubListReq) (*model.FavSubListRsp, error) {
	var err error
	res := &model.FavSubListRsp{}
	switch params.From {
	case FavNewTopicFromParam:
		// 只请求topic接口
		res.TopicList, err = s.subNewTopicsList(ctx, mid, params)
		if err != nil {
			log.Error("s.subNewTopicsList error=%+v", err)
			return nil, err
		}
		return res, nil
	case FavTagFromParam:
		// 只请求tag接口
		res.TagList, err = s.subTagList(ctx, mid, params)
		if err != nil {
			log.Error("s.subTagList error=%+v", err)
			return nil, err
		}
		return res, nil
	default:
		// 并发请求
		eg := errgroupv2.WithContext(ctx)
		eg.Go(func(ctx context.Context) (err error) {
			res.TopicList, err = s.subNewTopicsList(ctx, mid, params)
			if err != nil {
				log.Error("s.subNewTopicsList error=%+v", err)
				return nil
			}
			return nil
		})
		eg.Go(func(ctx context.Context) (err error) {
			res.TagList, err = s.subTagList(ctx, mid, params)
			if err != nil {
				log.Error("s.subTagList error=%+v", err)
				return nil
			}
			return nil
		})
		if err := eg.Wait(); err != nil {
			return nil, err
		}
		res.FavTab = &model.FavTab{
			Topic: res.TopicList != nil && len(res.TopicList.TopicItems) != 0,
			Tag:   res.TagList != nil && len(res.TagList.TagItems) != 0,
		}
		return res, nil
	}
}

func convertToFavTopicItems(raws []*topiccommon.TopicInfo) []*model.TopicItem {
	var res []*model.TopicItem
	for _, raw := range raws {
		item := &model.TopicItem{
			Id:      raw.Id,
			Name:    raw.Name,
			View:    raw.View,
			Discuss: raw.Discuss,
		}
		// 收藏跳链构造
		item.JumpUrl = fmt.Sprintf("https://m.bilibili.com/topic-detail?topic_id=%d", raw.Id)
		item.StatDesc = makeTopicItemDesc(raw.View, raw.Discuss)
		res = append(res, item)
	}
	return res
}

func convertToFavTagItems(raws []*dyntopicapi.SubTopicInfo, details map[int64]*topicextapi.TopicInfoDetail) []*model.TagItem {
	var res []*model.TagItem
	for _, raw := range raws {
		var jumpUrl string
		if detail, ok := details[raw.Id]; ok {
			jumpUrl = detail.TopicLink
		}
		res = append(res, &model.TagItem{
			Id:       raw.Id,
			Name:     raw.Name,
			View:     raw.View,
			Discuss:  raw.Discuss,
			JumpUrl:  jumpUrl,
			StatDesc: makeTopicItemDesc(raw.View, raw.Discuss),
		})
	}
	return res
}

func (s *Service) subNewTopicsList(ctx context.Context, mid int64, params *model.FavSubListReq) (*model.TopicFavList, error) {
	args := &topicsvc.FavTopicsReq{
		Uid: mid,
		PageInfo: &topiccommon.PaginationReq{
			PageSize: params.PageSize,
			PageNum:  params.PageNum,
			Offset:   params.Offset,
		},
	}
	topicList, err := s.topicGRPC.FavTopics(ctx, args)
	if err != nil {
		return nil, errors.Wrapf(err, "s.topicGRPC.FavTopics args=%+v, error=%+v", args, err)
	}
	res := &model.TopicFavList{
		TopicItems: convertToFavTopicItems(topicList.Topics),
		PageInfo: &model.PaginationRsp{
			CurPageNum: topicList.PageInfo.CurPageNum,
			Offset:     topicList.PageInfo.Offset,
			HasMore:    topicList.PageInfo.HasMore,
			Total:      topicList.PageInfo.Total,
		},
	}
	return res, nil
}

func (s *Service) subTagList(ctx context.Context, mid int64, params *model.FavSubListReq) (*model.TagFavList, error) {
	args := &dyntopicapi.SubTopicsReq{
		Uid:      mid,
		PageSize: params.PageSize,
		PageNum:  params.PageNum,
	}
	tagList, err := s.dynTopicGRPC.SubTopics(ctx, args)
	if err != nil {
		log.Error("s.dynTopicGRPC.SubTopics args=%+v, error=%+v", args, err)
		return nil, err
	}
	// 需要ListTopicDetails接口获取跳链
	details, err := s.listTopicDetails(ctx, mid, tagList.Topics)
	if err != nil {
		log.Error("s.listTopicDetails error=%+v", err)
		return nil, err
	}
	var tagItems []*model.TagItem
	if details != nil && details.Topics != nil {
		tagItems = convertToFavTagItems(tagList.Topics, details.Topics)
	}
	res := &model.TagFavList{
		TagItems: tagItems,
		PageInfo: &model.PaginationRsp{
			CurPageNum: tagList.CurPageNum,
			Offset:     tagList.Offset,
			HasMore:    tagList.HasMore,
		},
	}
	return res, nil
}

func (s *Service) listTopicDetails(ctx context.Context, mid int64, raws []*dyntopicapi.SubTopicInfo) (*topicextapi.RspListTopicDetails, error) {
	general := constructGeneralParamFromCtx(ctx)
	topicIds := makeTagListIds(raws)
	args := &topicextapi.ReqListTopicDetails{
		Uid: mid,
		MetaData: &dyncommonapi.CmnMetaData{
			Build:    general.GetBuildStr(),
			Platform: general.GetPlatform(),
			MobiApp:  general.GetMobiApp(),
			Device:   general.GetDevice(),
			Buvid:    general.GetBuvid(),
		},
		TopicIds: topicIds,
	}
	tmp, err := s.topicExtGRPC.ListTopicDetails(ctx, args)
	if err != nil {
		return nil, errors.Wrapf(err, "s.topicExtGRPC.ListTopicDetails args=%+v", args)
	}
	return tmp, nil
}

func makeTagListIds(topics []*dyntopicapi.SubTopicInfo) []int64 {
	var res []int64
	for _, v := range topics {
		res = append(res, v.Id)
	}
	return res
}

func (s *Service) AddFav(ctx context.Context, mid int64, params *model.AddFavReq) error {
	args := &topicsvc.AddFavReq{
		Uid:   mid,
		Tid:   params.TopicId,
		SetId: params.TopicSetId,
	}
	reply, err := s.topicGRPC.AddFav(ctx, args)
	if err != nil {
		if ecode.EqualError(topicecode.TopicSubFavFailed, err) {
			return ecode.Error(topicecode.TopicSubFavFailed, reply.GetMsg())
		}
		return err
	}
	return ecode.Error(ecode.OK, reply.Msg)
}

func (s *Service) CancelFav(ctx context.Context, mid int64, params *model.CancelFavReq) error {
	args := &topicsvc.CancelFavReq{
		Uid:   mid,
		Tid:   params.TopicId,
		SetId: params.TopicSetId,
	}
	reply, err := s.topicGRPC.CancelFav(ctx, args)
	if err != nil {
		if ecode.EqualError(topicecode.TopicSubFavFailed, err) {
			return ecode.Error(topicecode.TopicSubFavFailed, reply.GetMsg())
		}
		return err
	}
	return ecode.Error(ecode.OK, reply.Msg)
}
