package service

import (
	"context"
	"strconv"

	"go-common/library/log"
	errgroupv2 "go-common/library/sync/errgroup.v2"
	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	api "go-gateway/app/app-svr/topic/interface/api"
	"go-gateway/app/app-svr/topic/interface/internal/model"

	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
)

func (s *Service) topicDetailsStoryModeProcess(ctx context.Context, general *topiccardmodel.GeneralParam, req *api.TopicDetailsAllReq) (*api.TopicDetailsAllReply, error) {
	args := &topicsvc.TopicInfoReq{
		TopicId:   req.TopicId,
		Uid:       general.Mid,
		NeedShare: true,
		NeedEntry: true,
		MetaData:  constructTopicCommonMetaDataCtrl(general, ""),
	}
	topicReply, err := s.topicGRPC.TopicInfo(ctx, args)
	if err != nil {
		log.Error("s.topicGRPC.TopicInfo args=%+v, error=%+v", args, err)
		return nil, err
	}
	res := &api.TopicDetailsAllReply{}
	detail := &model.TopicDetailsAll{}
	eg := errgroupv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		res.DetailsTopInfo, err = s.makeTopicTopInfo(ctx, general, req, topicReply)
		if err != nil {
			log.Error("s.makeTopicTopInfo req=%+v, err=%+v", req, err)
			return err
		}
		if res.DetailsTopInfo.TopicInfo != nil {
			res.PubLayer = makeStoryModeTopicPubLayer(topicReply.EntryOption, topicReply.ClappedUrl, topicReply.EntranceCopywriting)
		}
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
		// 构建视频卡
		args := &topicsvc.VideoStoryReq{
			TopicId:    req.TopicId,
			FromSortBy: req.SortBy,
			Offset:     req.Offset,
			PageSize:   int64(req.PageSize),
			Uid:        general.Mid,
			MetaData:   constructTopicCommonMetaDataCtrl(general, _sourceFromStoryModeDetails),
		}
		reply, err := s.topicGRPC.VideoStory(ctx, args)
		if err != nil {
			log.Error("s.topicGRPC.VideoStory args=%+v, error=%+v", args, err)
			return nil
		}
		res.TopicCardList = s.makeStoryModeTopicCardList(ctx, general, reply, req.TopicId, req.Offset, topicReply.ClappedUrl)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		reply, err := s.topicGRPC.TopConfig(ctx, &topicsvc.TopicInlineResReq{TopicId: req.TopicId})
		if err != nil {
			log.Error("s.topicGRPC.TopConfig req=%+v, error=%+v", req, err)
			return nil
		}
		// 获取赛事卡
		detail.EsportInfo = s.makeEsportCard(ctx, reply, general)
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.Wait() error=%+v", err)
		return nil, err
	}
	detail.ResAll = res
	detail.ResAll.TopicTopCards = makeTopicTopCards(detail)
	return detail.ResAll, nil
}

func (s *Service) makeStoryModeTopicCardList(ctx context.Context, general *topiccardmodel.GeneralParam, reply *topicsvc.VideoStoryRsp, topicId int64, offset, clappedUrl string) *api.TopicCardList {
	cardItems := s.makeVideoSmallCardItems(ctx, general, reply, topicId, offset)
	if len(cardItems) == 0 {
		return &api.TopicCardList{NoCardResultReply: &api.NoCardResultReply{DefaultGuideText: _noCardReplyGuideText, ShowButton: &api.ShowButton{ShowText: "参与话题", JumpUrl: clappedUrl}}}
	}
	return &api.TopicCardList{
		TopicCardItems: cardItems,
		Offset:         reply.Offset,
		HasMore:        reply.HasMore,
		TopicSortByConf: &api.TopicSortByConf{
			DefaultSortBy: reply.SortByConf.DefaultSortBy,
			AllSortBy:     convertAllSortBy(reply.SortByConf.AllSortBy),
			ShowSortBy:    reply.SortByConf.ShowSortBy,
		},
	}
}

func (s *Service) makeVideoSmallCardItems(ctx context.Context, general *topiccardmodel.GeneralParam, upStreamRsp *topicsvc.VideoStoryRsp, topicId int64, offset string) []*api.TopicCardItem {
	aids, storyItems, ok := resolveStoryItemPlayAvAids(general, upStreamRsp)
	if !ok {
		return nil
	}
	reply, err := s.arcsPlayer(ctx, aids, true, _sourceFromStoryModeDetails)
	if err != nil {
		log.Error("makeVideoSmallCardItems s.arcsPlayer mid=%d, aids=%+v, error=%+v", general.Mid, aids, err)
		return nil
	}
	var res []*api.TopicCardItem
	for _, item := range upStreamRsp.Items {
		if reply == nil || reply[item.Rid] == nil {
			continue
		}
		res = append(res, &api.TopicCardItem{
			Type:               api.TopicCardType_VIDEO_SMALL_CARD,
			VideoSmallCardItem: constructVideoSmallCardItem(reply[item.Rid], storyItems[item.Rid], topicId, offset, item),
		})
	}
	return res
}

func constructVideoSmallCardItem(arcPlayer *archivegrpc.ArcPlayer, storyItem *model.StoryItemFromTopic, topicId int64, offset string, item *topicsvc.StoryItem) *api.VideoSmallCardItem {
	if arcPlayer.Arc == nil || storyItem == nil {
		return nil
	}
	arc := arcPlayer.Arc
	uri := topiccardmodel.FillURI(topiccardmodel.GotoStory, strconv.FormatInt(arc.Aid, 10), topiccardmodel.AvPlayHandlerGRPCV2(arcPlayer, 0, true))
	cardUri := topiccardmodel.FillURI(topiccardmodel.GotoURL, uri, topiccardmodel.SuffixHandler(topiccardmodel.MakeStorySuffixUrl(storyItem.Vmid, arc.Aid, topicId, storyItem.SortBy, offset, topiccardmodel.GotoStory)))
	return &api.VideoSmallCardItem{
		VideoCardBase: &api.VideoCardBase{
			Cover:    arc.Pic,
			Title:    arc.Title,
			UpName:   arc.Author.GetName(),
			Play:     int64(arc.Stat.View),
			JumpLink: cardUri,
			Aid:      arc.Aid,
		},
		CoverLeftBadgeText: storyItem.CornerMark,
		CardStatIcon_1:     int64(appcardmodel.IconUp),
		CardStatText_1:     arc.Author.GetName(),
		CardStatIcon_2:     int64(appcardmodel.IconPlay),
		CardStatText_2:     topiccardmodel.StatString(int64(arc.Stat.View), "", "-"),
		ServerInfo:         item.ServerInfo,
	}
}

func resolveStoryItemPlayAvAids(general *topiccardmodel.GeneralParam, upStreamRsp *topicsvc.VideoStoryRsp) ([]*archivegrpc.PlayAv, map[int64]*model.StoryItemFromTopic, bool) {
	items := upStreamRsp.Items
	if len(items) == 0 {
		return nil, nil, false
	}
	var (
		playAvids  []*archivegrpc.PlayAv
		storyItems = make(map[int64]*model.StoryItemFromTopic, len(items))
	)
	for _, v := range items {
		playAvids = append(playAvids, &archivegrpc.PlayAv{Aid: v.Rid})
		storyItems[v.Rid] = &model.StoryItemFromTopic{
			Vmid:       general.Mid,
			SortBy:     upStreamRsp.SortByConf.GetShowSortBy(),
			CornerMark: v.CornerMark,
		}
	}
	return playAvids, storyItems, true
}

func makeStoryModeTopicPubLayer(entryOption int32, clappedUrl, entranceCopywriting string) *api.PubLayer {
	const (
		alwaysBigButtonType = 12
	)
	if entryOption == 0 {
		return &api.PubLayer{ClosePubLayerEntry: true}
	}
	if entranceCopywriting == "" {
		// 配置文案暂时兜底
		entranceCopywriting = "立即参与"
	}
	res := &api.PubLayer{
		ShowType: alwaysBigButtonType,
		JumpLink: clappedUrl,
		ButtonMeta: &api.ButtonMeta{
			Text: entranceCopywriting,
			Icon: "https://i0.hdslb.com/bfs/activity-plat/static/ce06d65bc0a8d8aa2a463747ce2a4752/FBeLdCuYC1.png",
		},
	}
	return res
}
