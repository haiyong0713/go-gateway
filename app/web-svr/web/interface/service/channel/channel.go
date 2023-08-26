package channel

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	chanDao "go-gateway/app/web-svr/web/interface/dao/channel"
	"go-gateway/app/web-svr/web/interface/model/channel"
	chanmdl "go-gateway/app/web-svr/web/interface/model/channel"

	"go-common/library/sync/errgroup.v2"

	changrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	cardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	"github.com/pkg/errors"
)

func (s *Service) Red(c context.Context, mid int64) (*chanmdl.RedReply, error) {
	notifyReply, err := s.chDao.NewNotify(c, mid)
	if err != nil {
		log.Error("[Red] s.chDao.NewNotify(%+v): %+v", mid, err)
		return nil, err
	}
	return &chanmdl.RedReply{
		Cover:           notifyReply.GetIcon(),
		ChannelID:       notifyReply.GetCid(),
		ChannelName:     notifyReply.GetName(),
		Notify:          notifyReply.GetNotify(),
		Ctype:           notifyReply.GetCtype(),
		SubscribedCount: notifyReply.GetSubscribedCnt(),
	}, nil
}

func (s *Service) SubscribedList(c context.Context, mid int64) (*chanmdl.SubscribedListReply, error) {
	subReply, err := s.chDao.SubscribedChannel(c, mid)
	if err != nil {
		log.Error("[SubscribedList] s.chDao.SubscribedChannel(%+v): %+v", mid, err)
		return nil, err
	}
	stickChannels := make([]*chanmdl.WebChannel, 0)
	normalChannels := make([]*chanmdl.WebChannel, 0)
	if subReply.GetCount() > 0 {
		for _, card := range subReply.GetTops() {
			channel := &chanmdl.WebChannel{}
			channel.FormChannelCard(card)
			stickChannels = append(stickChannels, channel)
		}
		for _, card := range subReply.GetCards() {
			channel := &chanmdl.WebChannel{}
			channel.FormChannelCard(card)
			normalChannels = append(normalChannels, channel)
		}
	}
	total := len(stickChannels) + len(normalChannels)
	return &chanmdl.SubscribedListReply{
		Total:          int64(total),
		StickChannels:  stickChannels,
		NormalChannels: normalChannels,
	}, nil
}

func (s *Service) ViewList(c context.Context, mid int64) (*chanmdl.ViewListReply, error) {
	viewReply, err := s.chDao.ViewChannel(c, mid)
	if err != nil {
		log.Error("[ViewList] s.chDao.ViewChannel(%+v): %+v", mid, err)
		return nil, err
	}
	viewChannels := make([]*chanmdl.ViewChannel, 0)
	for _, card := range viewReply.GetCard() {
		viewChannel := &chanmdl.ViewChannel{}
		viewChannel.FormViewChannelCard(card)
		viewChannels = append(viewChannels, viewChannel)
	}
	return &chanmdl.ViewListReply{
		Total:    int64(len(viewChannels)),
		Channels: viewChannels,
	}, nil
}

func (s *Service) Stick(c context.Context, mid int64, req *chanmdl.StickReq) error {
	if err := s.chDao.UpdateSubscribe(c, mid, req.StickList, req.NormalList); err != nil {
		log.Error("[Stick] s.chDao.UpdateSubscribe(%+v): %+v", req, err)
		return err
	}
	return nil
}

func (s *Service) Subscribe(c context.Context, mid int64, req *chanmdl.SubscribeReq) error {
	if err := s.dao.AddSub(c, mid, []int64{req.ID}); err != nil {
		log.Error("[Subscribe] s.dao.AddSub(%+v, %+v): %+v", mid, []int64{req.ID}, err)
		return err
	}
	return nil
}

func (s *Service) Unsubscribe(c context.Context, mid int64, req *chanmdl.UnsubscribeReq) error {
	if err := s.dao.CancelSub(c, mid, req.ID); err != nil {
		log.Error("[Unsubscribe] s.dao.CancelSub(%+v, %+v): %+v", mid, req.ID, err)
		return err
	}
	return nil
}

func (s *Service) HotList(c context.Context, mid int64, req *chanmdl.HotListReq) (*chanmdl.HotListReply, error) {
	hotReq := &changrpc.HotChannelReq{
		Mid:    mid,
		Offset: req.Offset,
		Ps:     req.PageSize,
		Count:  chanmdl.HotListCount,
		Typ:    chanDao.TypWeb,
	}
	hotListReply, err := s.chDao.HotChannel(c, mid, hotReq)
	if err != nil {
		log.Error("[HotList] s.chDao.HotChannel(%+v) (%+v)", hotReq, err)
		return nil, err
	}
	arcChannels := s.formatViewChanCards(c, hotListReply.GetCard(), req.NeedArc)
	return &chanmdl.HotListReply{
		Offset:      hotListReply.Offset,
		ArcChannels: arcChannels,
	}, nil
}

func (s *Service) Detail(c context.Context, mid int64, req *chanmdl.WebDetailReq) (*chanmdl.DetailReply, error) {
	var (
		channelDetail *changrpc.ChannelDetailReply
		seasonCards   *cardgrpc.SeasonCards
	)
	g := errgroup.WithContext(c)
	g.Go(func(ctx context.Context) error {
		var err error
		channelDetail, err = s.chDao.ChannelDetail(ctx, mid, req.ID)
		if err != nil {
			log.Error("[Detail] s.chDao.ChannelDetail(%+v, %+v) (%+v)", mid, req.ID, err)
			return err
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		tagCards, err := s.dao.TagOGV(ctx, []int64{req.ID})
		if err != nil {
			log.Error("[Detail] s.dao.TagOGV(%+v) (%+v)", []int64{req.ID}, err)
			return nil
		}
		if info, ok := tagCards[req.ID]; ok {
			seasonCards = info
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		log.Error("[Detail] g.Wait (%+v)", err)
		return nil, err
	}
	if channelDetail == nil {
		return nil, ecode.NothingFound
	}
	// 组装detailReply
	detailReply := &chanmdl.DetailReply{TagChannels: make([]*chanmdl.WebChannel, 0)}
	detailReply.WebChannel.FormChannelCard(channelDetail.GetChannel())
	detailReply.WebChannel.FormSeason(channelDetail.GetPGC(), seasonCards)
	// pr管控判断
	var isPR bool
	for _, channelID := range s.c.PRLimit.ChannelList {
		if channelID == req.ID {
			isPR = true
			break
		}
	}
	// 精选-添加全部选项
	if !isPR {
		featuredTab := &chanmdl.DetailTab{
			Type:    chanmdl.DetailTabF,
			Options: make([]*chanmdl.DetailOption, 0),
		}
		if len(channelDetail.GetFeaturedOptions()) > 0 {
			featuredTab.Options = append(featuredTab.Options, &chanmdl.DetailOption{Title: "全部", Value: "0"})
			for _, op := range channelDetail.GetFeaturedOptions() {
				option := &chanmdl.DetailOption{}
				option.FormFeaturedOption(op)
				featuredTab.Options = append(featuredTab.Options, option)
			}
			detailReply.Tabs = []*chanmdl.DetailTab{featuredTab}
		}
	}
	// 综合tab
	multiTab := &chanmdl.DetailTab{
		Type: chanmdl.DetailTabM,
	}
	multiTab.Options = append(multiTab.Options, chanmdl.AllSortHotWeb)
	if !isPR {
		multiTab.Options = append(multiTab.Options, chanmdl.AllSortVieWeb, chanmdl.AllSortNewWeb)
	}
	detailReply.Tabs = append(detailReply.Tabs, multiTab)
	for _, assocChannel := range channelDetail.GetChannel().GetAssocChannels() {
		tagChannel := &chanmdl.WebChannel{}
		tagChannel.FormAssocChannel(assocChannel)
		detailReply.TagChannels = append(detailReply.TagChannels, tagChannel)
	}
	return detailReply, nil
}

func (s *Service) FeaturedList(c context.Context, mid int64, req *chanmdl.FeaturedListReq) (*chanmdl.FeaturedListReply, error) {
	// 获取精选列表
	resListReq := &changrpc.ResourceListReq{
		ChannelId:  req.ChannelID,
		TabType:    changrpc.TabType_TAB_TYPE_FEATURED,
		FilterType: req.FilterType,
		Offset:     req.Offset,
		PageSize:   req.PageSize,
		Mid:        mid,
	}
	resListReply, err := s.chDao.ResourceList(c, resListReq)
	if err != nil {
		log.Error("[FeaturedList] s.chDao.ResourceList(%+v) (%+v)", resListReq, err)
		return nil, err
	}
	// 只有第一页需要展示剧集
	needSeason := false
	if req.Offset == "" {
		needSeason = true
	}
	tabListItem := make([]interface{}, 0)
	if formatRes, err := s.formatResourceList(c, req.ChannelID, resListReply, needSeason, false, int64(req.FilterType), ""); err == nil {
		tabListItem = formatRes
	}
	return &chanmdl.FeaturedListReply{
		Offset:  resListReply.GetNextOffset(),
		HasMore: resListReply.GetHasMore(),
		List:    tabListItem,
	}, nil
}

func (s *Service) MultipleList(c context.Context, mid int64, req *chanmdl.MultipleListReq) (*chanmdl.MultipleListReply, error) {
	// 获取综合列表
	var sortType changrpc.TotalSortType
	switch req.SortType {
	case "hot":
		sortType = changrpc.TotalSortType_SORT_BY_HOT
	case "view":
		sortType = changrpc.TotalSortType_SORT_BY_VIEW_CNT
	case "new":
		sortType = changrpc.TotalSortType_SORT_BY_PUB_TIME
	default:
		err := ecode.RequestErr
		return nil, err
	}
	resListReq := &changrpc.ResourceListReq{
		ChannelId: req.ChannelID,
		TabType:   changrpc.TabType_TAB_TYPE_TOTAL,
		SortType:  sortType,
		Offset:    req.Offset,
		PageSize:  req.PageSize,
		Mid:       mid,
	}
	resListReply, err := s.chDao.ResourceList(c, resListReq)
	if err != nil {
		log.Error("[MultipleList] s.chDao.ResourceList(%+v): %+v", resListReq, err)
		return nil, err
	}
	tabListItem := make([]interface{}, 0)
	needRank := false
	if sortType == changrpc.TotalSortType_SORT_BY_HOT {
		needRank = true
	}
	if formatRes, err := s.formatResourceList(c, req.ChannelID, resListReply, false, needRank, 0, req.SortType); err == nil {
		tabListItem = formatRes
	}
	return &chanmdl.MultipleListReply{
		Offset:  resListReply.GetNextOffset(),
		HasMore: resListReply.GetHasMore(),
		List:    tabListItem,
	}, nil
}

// nolint:gocognit
func (s *Service) formatResourceList(c context.Context, chanId int64, resListReply *changrpc.ResourceListReply, needSeason, needRank bool, ftype int64, stype string) ([]interface{}, error) {
	var (
		err         error
		tabListItem = make([]interface{}, 0)
	)
	// 组装数据
	var aids []int64
	for _, card := range resListReply.GetCards() {
		switch card.GetCardType() {
		case changrpc.ChannelCardType_CARD_TYPE_VIDEO_ARCHIVE:
			if card.GetVideoCard() != nil {
				aids = append(aids, card.GetVideoCard().Rid)
			}
		case changrpc.ChannelCardType_CARD_TYPE_CUSTOM_CARD:
			// 过滤掉自定义卡
		case changrpc.ChannelCardType_CARD_TYPE_RANK_CARD:
			// 只有频道综合 && 近期热门 才展示排行榜
			if needRank && card.GetRankCard() != nil {
				for _, rankCard := range card.GetRankCard().Cards {
					aids = append(aids, rankCard.GetRid())
				}
			}
		default:
			log.Warn("[formatResourceList] unknown card_type %+v", card)
		}
	}
	var (
		arcsMap     = make(map[int64]*arcgrpc.Arc)
		seasonCards = &cardgrpc.SeasonCards{}
	)
	g := errgroup.WithContext(c)
	// 获取稿件
	if len(aids) > 0 {
		g.Go(func(ctx context.Context) error {
			arcsMap, err = s.dao.Arcs(ctx, aids)
			if err != nil {
				log.Error("[formatResourceList] s.dao.Arcs(%+v) (%+v)", aids, err)
				return err
			}
			return nil
		})
	}
	// 获取剧集
	if needSeason && resListReply.GetPGC() {
		g.Go(func(ctx context.Context) error {
			tagCards, err := s.dao.TagOGV(ctx, []int64{chanId})
			if err != nil {
				log.Error("[formatResourceList] s.dao.TagOGV(%+v): %+v", []int64{chanId}, err)
				return err
			}
			if item, ok := tagCards[chanId]; ok {
				seasonCards = item
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("[formatResourceList] g.Wait (%+v)", err)
		return nil, err
	}
	// 区分剧集、稿件和排行榜
	type SeasonType struct {
		chanmdl.Season
		Sort string `json:"sort"` //用于web端透传上报
		Filt int64  `json:"filt"` //用于web端透传上报
	}
	type ArchiveType struct {
		chanmdl.Archive
		CardType string `json:"card_type,omitempty"`
		Sort     string `json:"sort"`
		Filt     int64  `json:"filt"`
	}
	var (
		doneArcIDs = make(map[int64]struct{})
		seasonList = &chanmdl.TabListItem{CardType: chanmdl.TabTypeS, Items: make([]interface{}, 0)}
		rankList   = &chanmdl.TabListItem{CardType: chanmdl.TabTypeR, Items: make([]interface{}, 0)}
	)
	// 剧集
	if needSeason && len(seasonCards.GetCards()) > 0 {
		for _, seasonCard := range seasonCards.GetCards() {
			if seasonCard == nil {
				continue
			}
			season := &SeasonType{Sort: stype, Filt: ftype}
			season.FormSeasonCard(seasonCard)
			seasonList.Items = append(seasonList.Items, season)
		}
		tabListItem = append(tabListItem, seasonList)
	}
	// 排行榜
	if needRank {
		for _, card := range resListReply.GetCards() {
			if card.GetCardType() != changrpc.ChannelCardType_CARD_TYPE_RANK_CARD || card.GetRankCard() == nil {
				continue
			}
			if card.GetRankCard().GetDetail() != nil {
				rankList.PublishRange = card.GetRankCard().GetDetail().GetPubRange()
				rankList.UpdateTime = card.GetRankCard().GetDetail().GetUpdateTime()
				rankList.Title = card.GetRankCard().GetDetail().GetTitle()
			}
			for _, rankCard := range card.GetRankCard().GetCards() {
				if _, ok := arcsMap[rankCard.GetRid()]; !ok {
					continue
				}
				if _, ok := doneArcIDs[rankCard.GetRid()]; ok {
					continue
				}
				archive := &ArchiveType{Sort: stype, Filt: ftype}
				archive.FormVideoCard(rankCard)
				archive.FormArc(arcsMap[rankCard.GetRid()])
				rankList.Items = append(rankList.Items, archive)
				doneArcIDs[rankCard.GetRid()] = struct{}{}
			}
		}
		if len(rankList.Items) != 0 {
			tabListItem = append(tabListItem, rankList)
		}
	}
	// 稿件
	for _, card := range resListReply.GetCards() {
		if card.GetCardType() != changrpc.ChannelCardType_CARD_TYPE_VIDEO_ARCHIVE || card.GetVideoCard() == nil {
			continue
		}
		videoCard := card.GetVideoCard()
		if _, ok := arcsMap[videoCard.GetRid()]; !ok {
			continue
		}
		// RankCard 和 VideoCard 需去重
		if _, ok := doneArcIDs[videoCard.GetRid()]; ok {
			continue
		}
		archive := &ArchiveType{CardType: chanmdl.TabTypeA, Sort: stype, Filt: ftype}
		archive.FormVideoCard(videoCard)
		archive.FormArc(arcsMap[videoCard.GetRid()])
		tabListItem = append(tabListItem, archive)
	}
	return tabListItem, nil
}

// nolint:gocognit
func (s *Service) Search(c context.Context, mid int64, req *chanmdl.SearchReq) (*chanmdl.SearchReply, error) {
	var (
		err            error
		cids, hideCids []int64
		g              = errgroup.WithContext(c)
		esResult       = &chanmdl.EsRes{}
	)
	// 搜索匹配频道
	g.Go(func(ctx context.Context) (err error) {
		if esResult, cids, err = s.chDao.SearchEs(ctx, req.Keyword, req.Page, req.PageSize, chanmdl.EsStateOK); err != nil {
			log.Error("[Search] s.chDao.SearchEs(%+v, %+v) (%+v)", req, chanmdl.EsStateOK, err)
			return err
		}
		return nil
	})
	// 搜索隐藏频道
	g.Go(func(ctx context.Context) (err error) {
		if _, hideCids, err = s.chDao.SearchEs(ctx, req.Keyword, 1, 50, chanmdl.EsStateHide); err != nil {
			log.Error("[Search] s.chDao.SearchEs(%+v, %+v) (%+v)", req, chanmdl.EsStateHide, err)
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("[Search] g.Wait() (%+v)", err)
		return nil, err
	}
	var (
		isFirstPage   bool
		moreChannels  []*changrpc.RelativeChannel
		hotChannels   []*changrpc.ChannelCard
		chSeasonCards map[int64]*cardgrpc.SeasonCards
		channels      = make(map[int64]*changrpc.SearchChannelCard)
		moreCids      = hideCids
		g2            = errgroup.WithContext(c)
	)
	if req.Page == 1 {
		isFirstPage = true
		moreCids = append(moreCids, cids...)
	}
	// 获取频道详情
	if len(cids) > 0 {
		g2.Go(func(ctx context.Context) (err error) {
			searchReq := &changrpc.SearchChannelsInfoReq{Mid: mid, Cids: cids, Count: chanmdl.SearchVideoCount}
			searchReply, err := s.chDao.SearchChannelsInfo(c, searchReq)
			if err != nil {
				log.Error("[Search] s.chDao.SearchChannelsInfo(%+v) (%+v)", searchReq, err)
				return err
			}
			for _, card := range searchReply.GetCards() {
				if card == nil {
					continue
				}
				channels[card.GetCid()] = card
			}
			return nil
		})
	}
	// 获取频道剧集数
	if len(cids) > 0 {
		g2.Go(func(ctx context.Context) (err error) {
			chSeasonCards, err = s.dao.TagOGV(ctx, cids)
			if err != nil {
				log.Error("[Search] s.dao.TagOGV(%+v) (%+v)", cids, err)
				return err
			}
			return nil
		})
	}
	// 获取更多频道
	if isFirstPage && len(moreCids) > 0 {
		g2.Go(func(ctx context.Context) (err error) {
			relateReq := &changrpc.RelativeChannelReq{Mid: mid, Cids: moreCids}
			relateReply, err := s.chDao.RelativeChannel(ctx, relateReq)
			if err != nil {
				log.Error("[Search] s.chDao.RelativeChannel(%+v) (%+v)", relateReq, err)
				return nil
			}
			for _, card := range relateReply.GetCards() {
				if card == nil {
					continue
				}
				moreChannels = append(moreChannels, card)
			}
			return nil
		})
	}
	// 获取热门频道
	if isFirstPage {
		g2.Go(func(ctx context.Context) (err error) {
			hotReq := &changrpc.ChannelListReq{Mid: mid, CategoryType: chanmdl.CategoryHot}
			hotReply, err := s.chDao.ChannelList(ctx, hotReq)
			if err != nil {
				log.Error("[Search] s.chDao.ChannelList(%+v) (%+v)", hotReq, err)
				return nil
			}
			for _, card := range hotReply.GetCards() {
				if card == nil {
					continue
				}
				hotChannels = append(hotChannels, card)
			}
			return nil
		})
	}
	if err = g2.Wait(); err != nil {
		log.Error("[Search] g.Wait() (%+v)", err)
		return nil, err
	}
	// 获取稿件
	var (
		aids []int64
		arcs map[int64]*arcgrpc.Arc
	)
	for _, card := range channels {
		for _, video := range card.GetVideoCards() {
			if video.GetRid() == 0 {
				continue
			}
			aids = append(aids, video.GetRid())
		}
	}
	if len(aids) > 0 {
		arcs, err = s.dao.Arcs(c, aids)
		if err != nil {
			log.Error("[Search] s.dao.Arcs(%+v) (%+v)", aids, err)
			return nil, err
		}
	}
	// 组装数据
	if esResult == nil || esResult.Data == nil {
		return nil, errors.New("esResult || esResult.Data is nil")
	}
	arcChannels := make([]*chanmdl.ArcChannel, 0, len(esResult.Data.Result))
	for _, cid := range cids {
		if _, ok := channels[cid]; !ok {
			continue
		}
		chCard := channels[cid]
		// 组装Channel部分
		arcChannel := &chanmdl.ArcChannel{Archives: make([]*chanmdl.Archive, 0, len(chCard.GetVideoCards()))}
		arcChannel.FormSearchChannelCard(chCard)
		arcChannel.FormSeason(chCard.GetPGC(), chSeasonCards[cid])
		// 组装Archives部分
		for _, card := range chCard.GetVideoCards() {
			if _, ok := arcs[card.GetRid()]; !ok || card == nil {
				continue
			}
			archive := &chanmdl.Archive{}
			archive.FormVideoCard(card)
			archive.FormArc(arcs[card.GetRid()])
			arcChannel.Archives = append(arcChannel.Archives, archive)
		}
		arcChannels = append(arcChannels, arcChannel)
	}
	searchReply := &chanmdl.SearchReply{
		Pages:       esResult.Data.Page.Num,
		Total:       esResult.Data.Page.Total,
		ArcChannels: arcChannels,
	}
	// 非隐藏结果>1不展示更多、热门
	if !isFirstPage || len(searchReply.ArcChannels) > 1 {
		return searchReply, nil
	}
	var extChannels []*chanmdl.WebChannel
	// 优先更多
	for _, card := range moreChannels {
		ch := &chanmdl.WebChannel{}
		ch.FormRelativeChannel(card)
		extChannels = append(extChannels, ch)
	}
	if len(extChannels) > 0 {
		searchReply.ExtType = chanmdl.ExtMore
		searchReply.ExtChannels = extChannels
		return searchReply, nil
	}
	if len(searchReply.ArcChannels) > 0 {
		return searchReply, nil
	}
	for _, card := range hotChannels {
		ch := &chanmdl.WebChannel{}
		ch.FormChannelCard(card)
		extChannels = append(extChannels, ch)
	}
	if len(extChannels) > 0 {
		searchReply.ExtType = chanmdl.ExtHot
		searchReply.ExtChannels = extChannels
	}
	return searchReply, nil
}

// 按照播放量，精选->综合
func (s *Service) TopList(c context.Context, mid int64, req *chanmdl.TopListReq) (*chanmdl.TopListReply, error) {
	var (
		featuredReply *chanmdl.FeaturedListReply
		mulReply      *chanmdl.MultipleListReply
		err           error
	)
	if featuredReply, err = func() (*chanmdl.FeaturedListReply, error) {
		freq := &channel.FeaturedListReq{
			ChannelID:  req.ChannelID,
			Offset:     req.Offset,
			FilterType: 0,
			PageSize:   req.PageSize,
		}
		return s.FeaturedList(c, mid, freq)
	}(); err != nil || featuredReply == nil || (featuredReply.List != nil && len(featuredReply.List) <= 0) {
		// 获取综合的
		mreq := &channel.MultipleListReq{
			ChannelID: req.ChannelID,
			Offset:    req.Offset,
			PageSize:  req.PageSize,
			SortType:  "view",
		}
		if mulReply, err = s.MultipleList(c, mid, mreq); err != nil {
			return nil, err
		}
		return &chanmdl.TopListReply{
			List:    mulReply.List,
			Offset:  mulReply.Offset,
			HasMore: mulReply.HasMore,
		}, nil
	}
	return &chanmdl.TopListReply{
		List:    featuredReply.List,
		Offset:  featuredReply.Offset,
		HasMore: featuredReply.HasMore,
	}, nil
}
