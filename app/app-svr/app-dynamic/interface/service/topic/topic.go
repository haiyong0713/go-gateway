package topic

import (
	"context"
	"strconv"
	"strings"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-dynamic/interface/model"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	topicmdl "go-gateway/app/app-svr/app-dynamic/interface/model/topic"

	archivegrpc "go-gateway/app/app-svr/archive/service/api"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	topicgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
)

func (s *Service) Square(c context.Context, mid int64, buvid string, req *topicmdl.SquareReq) (*topicmdl.SquareReply, error) {
	g := errgroup.WithContext(c)
	// 发起活动
	var isLaunchedActivity bool
	g.Go(func(ctx context.Context) error {
		var err error
		if isLaunchedActivity, err = s.nativePageDao.IsUpActUID(ctx, mid); err != nil {
			return err
		}
		return nil
	})
	// 我的关注
	var mySubTopic *channelgrpc.SubscribeReply
	g.Go(func(ctx context.Context) error {
		var err error
		if mySubTopic, err = s.channelDao.SubscribedChannel(ctx, mid); err != nil {
			return err
		}
		return nil
	})
	// 推荐话题
	var topicRcmds []*topicgrpc.HotListDetail
	g.Go(func(ctx context.Context) error {
		var err error
		if topicRcmds, err = s.topicDao.OldRcmdActList(ctx, mid, buvid, req); err != nil {
			return err
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	// 物料获取
	topicContext, err := s.TopicContext(c, mid, req, mySubTopic, topicRcmds)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	// 组装【发起活动】【我的关注】【推荐话题】
	var res = &topicmdl.SquareReply{
		LaunchedActivity: s.FormTopicLaunchedActivity(isLaunchedActivity),
		Subscription:     s.FormTopicSubscription(req, mySubTopic, topicContext),
		Recommend:        s.FormTopicRecommend(req, topicRcmds, mySubTopic, topicContext),
	}
	return res, nil
}

// 具体话题所需物料
// nolint:gocognit
func (s *Service) TopicContext(c context.Context, mid int64, _ *topicmdl.SquareReq, mySubTopic *channelgrpc.SubscribeReply, topicRcmds []*topicgrpc.HotListDetail) (*topicmdl.Context, error) {
	var (
		actIDm  = make(map[int64]struct{})
		midm    = make(map[int64]struct{})
		aidm    = make(map[int64]map[int64]struct{})
		drawIDm = make(map[int64]struct{})
	)
	if mySubTopic != nil {
		// 置顶频道
		for _, top := range mySubTopic.GetTops() {
			if top.GetActAttr() == 1 {
				if top.GetChannelId() != 0 {
					actIDm[top.GetChannelId()] = struct{}{}
				}
			}
		}
		// 非置顶
		for _, card := range mySubTopic.GetCards() {
			if card.GetActAttr() == 1 {
				if card.GetChannelId() != 0 {
					actIDm[card.GetChannelId()] = struct{}{}
				}
			}
		}
	}
	for _, topicRcmd := range topicRcmds {
		if topicRcmd.Uid != 0 {
			midm[topicRcmd.Uid] = struct{}{}
		}
		switch topicRcmd.Type {
		case dynmdlV2.DynTypeVideo:
			if topicRcmd.Rid != 0 {
				aidm[topicRcmd.Rid] = nil
			}
		case dynmdlV2.DynTypeDraw:
			if topicRcmd.Rid != 0 {
				drawIDm[topicRcmd.Rid] = struct{}{}
			}
		}
	}
	var res = new(topicmdl.Context)
	g := errgroup.WithContext(c)
	// 活动接口
	if len(actIDm) > 0 {
		var actIDs []int64
		for actID := range actIDm {
			actIDs = append(actIDs, actID)
		}
		g.Go(func(ctx context.Context) error {
			resTmp, err := s.activityDao.NatInfoFromForeign(ctx, actIDs, 1)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			res.Activitys = resTmp
			return nil
		})
	}
	// 账号信息
	if len(midm) > 0 {
		var mids []int64
		for midTmp := range midm {
			mids = append(mids, midTmp)
		}
		g.Go(func(ctx context.Context) error {
			resTmp, err := s.accountDao.Cards3New(ctx, mids)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			res.Accounts = resTmp
			return nil
		})
	}
	// 稿件
	if len(aidm) > 0 {
		var aids []*archivegrpc.PlayAv
		for aid := range aidm {
			aids = append(aids, &archivegrpc.PlayAv{Aid: aid})
		}
		g.Go(func(ctx context.Context) error {
			resTmp, err := s.archiveDao.ArcsPlayer(ctx, aids, false, "")
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			res.Archives = resTmp
			return nil
		})
	}
	// 图文动态详情
	if (len(drawIDm)) > 0 {
		var drawIDs []int64
		for drawID := range drawIDm {
			drawIDs = append(drawIDs, drawID)
		}
		g.Go(func(ctx context.Context) error {
			generalParam := dynmdlV2.NewGeneralParamFromCtx(ctx)
			generalParam.Mid = mid
			resTmp, err := s.dynDao.DrawDetails(ctx, generalParam, drawIDs)
			if err != nil {
				log.Error("%+v", err)
				return err
			}
			res.Draws = resTmp
			return nil
		})
	}
	_ = g.Wait()
	return res, nil
}

func (s *Service) FormTopicLaunchedActivity(isLaunchedActivity bool) *topicmdl.LaunchedActivity {
	if !isLaunchedActivity {
		return nil
	}
	return &topicmdl.LaunchedActivity{
		Title: "发起活动",
		URL:   "https://www.bilibili.com/blackboard/up-sponsor.html?act_from=square_topic",
	}
}

func (s *Service) FormTopicSubscription(req *topicmdl.SquareReq, mySubTopic *channelgrpc.SubscribeReply, topicContext *topicmdl.Context) *topicmdl.Subscription {
	if mySubTopic == nil {
		return nil
	}
	var res = &topicmdl.Subscription{Title: "我的订阅"}
	// 置顶频道
	for _, mcs := range mySubTopic.GetTops() {
		if mcs == nil {
			log.Error("mine sub top nil req %+v", req)
			continue
		}
		i := &topicmdl.SubscriptionItem{}
		i.FormSubscriptionItem(mcs, topicContext.Activitys)
		i.SubType = "top"
		res.Top = append(res.Top, i)
	}
	// 非置顶频道
	for _, mcs := range mySubTopic.GetCards() {
		if mcs == nil {
			log.Error("mine sub card nil req %+v", req)
			continue
		}
		i := &topicmdl.SubscriptionItem{}
		i.FormSubscriptionItem(mcs, topicContext.Activitys)
		i.SubType = "card"
		res.Card = append(res.Card, i)
	}
	return res
}

// 聚合推荐话题
// nolint:gocognit
func (s *Service) FormTopicRecommend(req *topicmdl.SquareReq, topicRcmds []*topicgrpc.HotListDetail, mySubTopic *channelgrpc.SubscribeReply, topicContext *topicmdl.Context) *topicmdl.Recommend {
	var res *topicmdl.Recommend
	if len(topicRcmds) < 1 {
		log.Warn("FormTopicRecommend get rcmd len 0")
	}
	for _, topicRcmd := range topicRcmds {
		if topicRcmd == nil {
			continue
		}
		i := new(topicmdl.RecommendItem)
		i.DefauleURL = topicRcmd.TopicLink
		i.Type = topicRcmd.Type
		i.Rid = topicRcmd.Rid
		i.Mid = topicRcmd.Uid
		// 用户部分
		if topicContext.Accounts != nil {
			if user, ok := topicContext.Accounts[topicRcmd.Uid]; ok {
				i.Author = &topicmdl.RecommendItemAuthor{
					Name:   user.Name,
					Mid:    user.Mid,
					Face:   user.Face,
					URL:    topicRcmd.TopicLink,
					Suffix: "发起",
				}
			}
		}
		// 话题部分
		i.Topic = &topicmdl.RecommendItemTopic{
			Icon: "",
			ID:   topicRcmd.TopicId,
			Name: topicRcmd.TopicName,
			URL:  topicRcmd.TopicLink,
		}
		var labels []string
		if topicRcmd.HeatInfo != nil {
			if topicRcmd.HeatInfo.View != 0 {
				labels = append(labels, model.StatString(topicRcmd.HeatInfo.View, "浏览"))
			}
			if topicRcmd.HeatInfo.Discuss != 0 {
				labels = append(labels, model.StatString(topicRcmd.HeatInfo.Discuss, "讨论"))
			}
		}
		i.Topic.Label = strings.Join(labels, "·")
		// 订阅关系
		if mySubTopic != nil {
			for _, top := range mySubTopic.GetTops() {
				if top.ChannelId == topicRcmd.TopicId {
					i.Topic.IsSub = true
					break
				}
			}
			for _, card := range mySubTopic.GetCards() {
				if card.ChannelId == topicRcmd.TopicId {
					i.Topic.IsSub = true
					break
				}
			}
		}
		// 卡面部分
		switch topicRcmd.Type {
		case dynmdlV2.DynTypeVideo:
			if topicContext.Archives == nil {
				// TODO 销卡日志+告警
				continue
			}
			archive, ok := topicContext.Archives[topicRcmd.Rid]
			if !ok {
				// TODO 销卡日志+告警
				continue
			}
			i.Topic.Desc = archive.Arc.Dynamic
			i.Cover = &topicmdl.RecommendItemCover{
				Cover: archive.Arc.Pic,
				URL:   model.FillURI(model.GotoAv, strconv.FormatInt(archive.Arc.Aid, 10), nil),
				Labels: map[string]string{
					"duration": model.DurationString(archive.Arc.Duration),
					"view":     model.StatString(int64(archive.Arc.Stat.View), "观看"),
					"danmaku":  model.StatString(int64(archive.Arc.Stat.Danmaku), "弹幕"),
				},
			}
		case dynmdlV2.DynTypeDraw:
			if topicContext.Draws == nil {
				// TODO 销卡日志+告警
				continue
			}
			draw, ok := topicContext.Draws[topicRcmd.Rid]
			if !ok || draw.Item == nil {
				// TODO 销卡日志+告警
				continue
			}
			i.Topic.Desc = draw.Item.Description
			var cover string
			for _, pic := range draw.Item.Pictures {
				if pic != nil && pic.ImgSrc != "" {
					cover = pic.ImgSrc
					break
				}
			}
			i.Cover = &topicmdl.RecommendItemCover{
				Cover: cover,
				URL:   topicRcmd.TopicLink,
			}
		}
		if res == nil {
			res = new(topicmdl.Recommend)
			res.Title = "推荐活动"
		}
		res.List = append(res.List, i)
	}
	return res
}

// HotList 聚合热门话题
// nolint:gocognit
func (s *Service) HotList(c context.Context, mid int64, buvid string, req *topicmdl.HotListReq) (*topicmdl.HotListReply, error) {
	// 原始list
	hotlist, err := s.topicDao.OldHotList(c, mid, buvid, req)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	// 物料获取
	var (
		aidm    = make(map[int64]map[int64]struct{})
		drawIDm = make(map[int64]struct{})
	)
	for _, l := range hotlist.HotList {
		switch l.Type {
		case dynmdlV2.DynTypeVideo:
			if l.Rid != 0 {
				aidm[l.Rid] = nil
			}
		case dynmdlV2.DynTypeDraw:
			if l.Rid != 0 {
				drawIDm[l.Rid] = struct{}{}
			}
		}
	}
	var (
		archives map[int64]*archivegrpc.ArcPlayer
		draws    map[int64]*dynmdlV2.DrawDetailRes
	)
	g := errgroup.WithContext(c)
	// 稿件
	if len(aidm) > 0 {
		var aids []*archivegrpc.PlayAv
		for aid := range aidm {
			aids = append(aids, &archivegrpc.PlayAv{Aid: aid})
		}
		g.Go(func(ctx context.Context) error {
			if archives, err = s.archiveDao.ArcsPlayer(ctx, aids, false, ""); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	// 图文动态详情
	if (len(drawIDm)) > 0 {
		var drawIDs []int64
		for drawID := range drawIDm {
			drawIDs = append(drawIDs, drawID)
		}
		g.Go(func(ctx context.Context) error {
			generalParam := dynmdlV2.NewGeneralParamFromCtx(ctx)
			generalParam.Mid = mid
			if draws, err = s.dynDao.DrawDetails(ctx, generalParam, drawIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	err = g.Wait()
	// 数据聚合
	var res = new(topicmdl.HotListReply)
	for _, tab := range hotlist.HotTabList {
		resTab := &topicmdl.HotListTab{Name: tab.HotListDesc, TypeID: tab.HotListType}
		res.Tab = append(res.Tab, resTab)
	}
	for _, l := range hotlist.HotList {
		i := new(topicmdl.HotListItem)
		switch l.Type {
		case dynmdlV2.DynTypeVideo:
			archive, ok := archives[l.Rid]
			if !ok {
				// TODO 销卡日志+告警
				continue
			}
			i.ID = l.TopicId
			i.Name = l.TopicName
			i.Desc = archive.Arc.Dynamic
			i.URL = l.TopicLink
			i.Cover = archive.Arc.Pic
		case dynmdlV2.DynTypeDraw:
			draw, ok := draws[l.Rid]
			if !ok || draw.Item == nil {
				// TODO 销卡日志+告警
				continue
			}
			i.ID = l.TopicId
			i.Name = l.TopicName
			i.Desc = draw.Item.Description
			i.URL = l.TopicLink
			var cover string
			for _, pic := range draw.Item.Pictures {
				if pic != nil && pic.ImgSrc != "" {
					cover = pic.ImgSrc
					break
				}
			}
			i.Cover = cover
		}
		var labels []string
		if l.HeatInfo != nil {
			if l.HeatInfo.View != 0 {
				labels = append(labels, model.StatString(l.HeatInfo.View, "浏览"))
			}
			if l.HeatInfo.Discuss != 0 {
				labels = append(labels, model.StatString(l.HeatInfo.Discuss, "讨论"))
			}
		}
		i.Label = strings.Join(labels, "·")
		res.List = append(res.List, i)
	}
	res.Offset = hotlist.Offset
	if hotlist.HasMore == 1 {
		res.HasMore = true
	}
	return res, nil
}

func (s *Service) SubscribeSave(c context.Context, mid int64, req *topicmdl.SubscribeSaveReq) error {
	if err := s.channelDao.ChannelSort(c, mid, req.Action, req.Top, req.Card); err != nil {
		log.Error("%v", err)
	}
	return nil
}
