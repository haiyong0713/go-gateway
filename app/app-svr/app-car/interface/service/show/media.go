package show

import (
	"context"
	"fmt"
	"regexp"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-car/ecode"
	"go-gateway/app/app-svr/app-car/interface/model"
	bgmmodel "go-gateway/app/app-svr/app-car/interface/model/bangumi"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/popular"
	"go-gateway/app/app-svr/app-car/interface/model/search"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

const (
	// 美妆护肤 分区
	_regionMakeups = 157

	//近期热门
	_InterveneTypeHot = 1

	//搜索tab
	_InterveneTypeTab = 2
)

func (s *Service) MediaPopularWeb(c context.Context, param *popular.MediaPopularParam) ([]*cardm.MediaItemWeb, error) {
	start := (param.Pn - 1) * param.Ps
	cards := s.PopularCardTenList(c, 0, start, param.Ps)
	var (
		aids  []int64
		arcs  map[int64]*arcgrpc.Arc
		seams map[int32]*episodegrpc.EpisodeCardsProto
	)
	for _, v := range cards {
		if v.Value == 0 {
			continue
		}
		switch v.Type {
		case model.GotoAv:
			aids = append(aids, v.Value)
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) != 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsByAids(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	is := []*cardm.MediaItemWeb{}
	for _, v := range cards {
		var (
			main interface{}
			gt   string
		)
		if v.Value == 0 {
			continue
		}
		switch v.Type {
		case model.GotoAv:
			// 如果当前稿件是pgc视频，转成pgc卡片处理，否则还是稿件卡片
			main = arcs
			gt = model.GotoAv
			if _, ok := seams[int32(v.Value)]; ok {
				main = seams
				gt = model.GotoPGC
			}
		}
		i := &cardm.MediaItemWeb{}
		ok := i.FromMediaItemWeb(v.Value, gt, main)
		if !ok {
			continue
		}
		is = append(is, i)
	}
	if len(is) == 0 {
		return []*cardm.MediaItemWeb{}, nil
	}
	return is, nil
}

func (s *Service) MediaSearchWeb(c context.Context, param *search.MediaSearchParam) ([]*cardm.MediaItemWeb, error) {
	var (
		cardItem []*ai.Item
		aids     []int64
		ssids    []int32
		arcs     map[int64]*arcgrpc.Arc
		seams    map[int32]*seasongrpc.CardInfoProto
		seamAids map[int32]*episodegrpc.EpisodeCardsProto
	)
	all, err := s.srch.Search(c, 0, 0, param.Pn, param.Ps, param.Keyword, "")
	if err != nil {
		log.Error("%+v", err)
		return []*cardm.MediaItemWeb{}, nil
	}
	if all == nil || all.Result == nil {
		return []*cardm.MediaItemWeb{}, nil
	}
	// 转换成统一结构体
	// pgc
	for _, v := range all.Result.MediaBangumi {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	for _, v := range all.Result.MediaFt {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	// archive
	for _, v := range all.Result.Video {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoAv, ID: int64(v.ID)})
	}
	for _, v := range cardItem {
		if v.ID == 0 {
			continue
		}
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		case model.GotoPGC:
			ssids = append(ssids, int32(v.ID))
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seamAids, err = s.bgm.CardsByAids(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(ssids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsAll(ctx, ssids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return []*cardm.MediaItemWeb{}, nil
	}
	is := []*cardm.MediaItemWeb{}
	for _, v := range cardItem {
		var (
			main interface{}
			gt   string
		)
		switch v.Goto {
		case model.GotoAv:
			// 如果当前稿件是pgc视频，转成pgc卡片处理，否则还是稿件卡片
			main = arcs
			gt = model.GotoAv
			if _, ok := seamAids[int32(v.ID)]; ok {
				main = seamAids
				gt = model.GotoPGC
			}
		case model.GotoPGC:
			main = seams
			gt = model.GotoPGC
		}
		i := &cardm.MediaItemWeb{}
		ok := i.FromMediaItemWeb(v.ID, gt, main)
		if !ok {
			continue
		}
		is = append(is, i)
	}
	if len(is) == 0 {
		return []*cardm.MediaItemWeb{}, nil
	}
	return is, nil
}

func (s *Service) MediaPGCWeb(c context.Context, param *bgmmodel.MediaPGCParam) ([]*cardm.MediaItemWeb, error) {
	// 18 番剧推荐
	// 19 国创推荐
	// 88 电影热播
	// 87 纪录片热播
	var (
		followType int
		cardItem   []*ai.Item
	)
	switch param.FollowType {
	case _followTypeBangumi:
		followType = 18
	case _followTypeCinema:
		followType = 88
	default:
		return []*cardm.MediaItemWeb{}, nil
	}
	// 实际PGC这接口一期没有分页，第一页就返回了所有数据
	if param.Pn > 1 {
		return []*cardm.MediaItemWeb{}, nil
	}
	list, err := s.bgm.Module(c, followType, model.AndroidBilithings, "")
	if err != nil {
		log.Warn("%+v", err)
		return []*cardm.MediaItemWeb{}, nil
	}
	var (
		ssids   []int32
		seasonm map[int32]*seasongrpc.CardInfoProto
	)
	for _, l := range list {
		if l.SeasonID == 0 {
			continue
		}
		ssids = append(ssids, l.SeasonID)
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(l.SeasonID)})
	}
	if len(ssids) > 0 {
		var err error
		if seasonm, err = s.bgm.CardsAll(c, ssids); err != nil {
			log.Error("%+v", err)
			return nil, err
		}
	}
	is := []*cardm.MediaItemWeb{}
	for _, v := range cardItem {
		i := &cardm.MediaItemWeb{}
		ok := i.FromMediaItemWeb(v.ID, v.Goto, seasonm)
		if !ok {
			continue
		}
		is = append(is, i)
	}
	if len(is) == 0 {
		return []*cardm.MediaItemWeb{}, nil
	}
	return is, nil
}

func (s *Service) MediaRegion(c context.Context, pn, ps int64) ([]*cardm.MediaItem, string, error) {
	moreURL := fmt.Sprintf("bilithings://player?sourceType=%s&rid=%d", model.EntranceRegion, _regionMakeups)
	reply, err := s.reg.RanksArcs(c, _regionMakeups, pn, ps)
	if err != nil {
		return nil, moreURL, err
	}
	var (
		aids  []int64
		arcs  map[int64]*arcgrpc.Arc
		seams map[int32]*episodegrpc.EpisodeCardsProto
	)
	for _, v := range reply {
		if v.Aid == 0 {
			continue
		}
		aids = append(aids, v.Aid)
	}
	group := errgroup.WithContext(c)
	if len(aids) != 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsByAids(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	is := []*cardm.MediaItem{}
	for _, v := range reply {
		var (
			main interface{}
			gt   string
		)
		if v.Aid == 0 {
			continue
		}
		// 如果当前稿件是pgc视频，转成pgc卡片处理，否则还是稿件卡片
		main = arcs
		gt = model.GotoAv
		if _, ok := seams[int32(v.Aid)]; ok {
			main = seams
			gt = model.GotoPGC
		}
		materials := &cardm.Materials{Prune: cardm.GtPrune(model.GotoAv, v.Aid)}
		i := &cardm.MediaItem{}
		ok := i.FromItem(v.Aid, gt, main, materials)
		if !ok {
			continue
		}
		is = append(is, i)
	}
	if len(is) == 0 {
		return nil, moreURL, xecode.AppMediaNotData
	}
	return is, moreURL, nil
}

// MediaRegionV2 后续取代MediaRegion方法 ,更多按钮跳转scheme,使用搜索落地页承接
// 根据来源是小鹏时，定制处理：小鹏美妆空间近期热门 ：从B站App - 时尚-美妆护肤二级- 化妆教程频道-综合-近期热门-排行榜 类目下拉取指定页数的数据
func (s *Service) MediaRegionV2(c context.Context, pn, ps int64) ([]*cardm.MediaItem, string, error) {
	keyword := s.c.CustomModule.XiaoPengKeywordRegion
	moreURL := fmt.Sprintf("bilithings://search?from=beauty_space&resource=more&keyword=%s&ps=%d", keyword, ps)
	var (
		arg = &channelgrpc.ResourceListReq{
			ChannelId: s.c.CustomModule.ChannelMakeups,
			TabType:   channelgrpc.TabType_TAB_TYPE_TOTAL,
			SortType:  channelgrpc.TotalSortType_SORT_BY_HOT,
			PageSize:  int32(ps)}
	)

	var (
		aids []int64
		arcs map[int64]*arcgrpc.Arc
	)
	//插卡的数据获取点,是否打开开关，和aids做去重，取前N个数据，且是ugc的
	if !s.c.CustomModule.CircuitIntervene {
		interv, ok := s.xiaoPengRecs[fmt.Sprintf("%d_%s", _InterveneTypeHot, keyword)]
		if ok && interv != nil {
			//合并
			aids = append(aids, interv.Items...)
		}
	}
	//检索频道数据
	reply, err := s.channelDao.ResourceList(c, arg)
	if err != nil {
		//对于有err的情况，则自动降级
		log.Warnc(c, "ResourceList err: arg: %+v, er: %+v", arg, err)
	} else {
		for _, v := range reply.Cards {
			//频道下面，只有ugc内容
			if v.CardType != channelgrpc.ChannelCardType_CARD_TYPE_VIDEO_ARCHIVE {
				//只选取普通视频卡片
				continue
			}
			if v.VideoCard.Rid == 0 {
				continue
			}
			aids = append(aids, v.VideoCard.Rid)
		}
	}
	aids = s.mergeAndUniq(aids, nil)

	group := errgroup.WithContext(c)
	if len(aids) != 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	is := []*cardm.MediaItem{}
	for _, aid := range aids {
		var (
			main interface{}
			gt   string
		)
		if aid == 0 {
			continue
		}
		// 如果当前稿件是pgc视频，转成pgc卡片处理，否则还是稿件卡片
		main = arcs
		gt = model.GotoAv
		materials := &cardm.Materials{Prune: cardm.GtPrune(model.GotoAv, aid)}
		i := &cardm.MediaItem{}
		ok := i.FromItem(aid, gt, main, materials)
		if !ok {
			continue
		}
		is = append(is, i)
	}
	if len(is) == 0 {
		return nil, moreURL, xecode.AppMediaNotData
	}
	//取头部的N个
	if len(is) > int(ps) {
		is = is[0:ps]
	}
	return is, moreURL, nil
}

// mergeAndUniq 合并两个数组，l1 排在l2 前面，并且去重
func (s *Service) mergeAndUniq(l1 []int64, l2 []int64) []int64 {
	result := make([]int64, 0)
	m := make(map[int64]bool) //map的值不重要
	for _, v := range l1 {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	for _, v := range l2 {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}

func (s *Service) MediaRegionSearch(c context.Context, pn, ps int, keyword string) ([]*cardm.MediaItem, string, error) {
	var (
		cardItem []*ai.Item
		aids     []int64
		ssids    []int32
		arcs     map[int64]*arcgrpc.Arc
		seams    map[int32]*seasongrpc.CardInfoProto
		seamAids map[int32]*episodegrpc.EpisodeCardsProto
	)
	keywordURL := fmt.Sprintf("bilithings://search?keyword=%s", keyword)
	all, err := s.srch.Search(c, 0, _regionMakeups, pn, ps, keyword, "")
	if err != nil {
		return nil, keywordURL, err
	}
	if all == nil || all.Result == nil {
		return nil, keywordURL, xecode.AppMediaNotData
	}
	// 转换成统一结构体
	// pgc
	for _, v := range all.Result.MediaBangumi {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	for _, v := range all.Result.MediaFt {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
	}
	// archive
	for _, v := range all.Result.Video {
		cardItem = append(cardItem, &ai.Item{Goto: model.GotoAv, ID: int64(v.ID)})
	}
	for _, v := range cardItem {
		if v.ID == 0 {
			continue
		}
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		case model.GotoPGC:
			ssids = append(ssids, int32(v.ID))
		}
	}
	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seamAids, err = s.bgm.CardsByAidsAll(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(ssids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsAll(ctx, ssids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return []*cardm.MediaItem{}, keywordURL, nil
	}
	is := []*cardm.MediaItem{}
	for _, v := range cardItem {
		var (
			main interface{}
			gt   string
		)
		switch v.Goto {
		case model.GotoAv:
			// 如果当前稿件是pgc视频，转成pgc卡片处理，否则还是稿件卡片
			main = arcs
			gt = model.GotoAv
			if _, ok := seamAids[int32(v.ID)]; ok {
				main = seamAids
				gt = model.GotoPGC
			}
		case model.GotoPGC:
			main = seams
			gt = model.GotoPGC
		}
		materials := &cardm.Materials{Prune: cardm.GtPrune(model.GotoAv, v.ID)}
		i := &cardm.MediaItem{}
		ok := i.FromItem(v.ID, gt, main, materials)
		if !ok {
			continue
		}
		is = append(is, i)
	}
	if len(is) == 0 {
		return nil, keywordURL, xecode.AppMediaNotData
	}
	return is, keywordURL, nil
}

// MediaRegionSearchV2 小鹏资源 后续取代MediaRegionSearch方法
func (s *Service) MediaRegionSearchV2(c context.Context, pn, ps int, keyword string) ([]*cardm.MediaItem, string, error) {
	var (
		cardItem []*ai.Item
		aids     []int64
		ssids    []int32
		arcs     map[int64]*arcgrpc.Arc
		seams    map[int32]*seasongrpc.CardInfoProto
		seamAids map[int32]*episodegrpc.EpisodeCardsProto
	)
	keywordURL := fmt.Sprintf("bilithings://search?from=beauty_space&resource=more&keyword=%s&ps=%d", keyword, ps)
	//如果是 小鹏美妆空间XXX教程 从时尚分区 - 美妆护肤二级分区 类目下进行XXX检索
	reg := regexp.MustCompile(s.c.CustomModule.XiaoPengKeywordTab)
	orgKeyword := keyword
	matchArr := reg.FindStringSubmatch(keyword)
	if len(matchArr) > 0 {
		keyword = matchArr[len(matchArr)-1]
	}
	//插卡的数据获取点,是否打开开关，和aids做去重，取前N个数据，且是ugc的
	_m := make(map[int64]bool) //map的值不重要,用于去重
	cardItem, _m = s.makeCardItem(orgKeyword, cardItem, _m)
	//查询大搜
	all, _ := s.srch.Search(c, 0, _regionMakeups, pn, ps, keyword, "")
	//if err != nil || all == nil || all.Result == nil {
	//	log.Warnc(c, "Search err %+v", err)
	// 检索结果转换成统一结构体
	if all != nil && all.Result != nil {
		// pgc
		for _, v := range all.Result.MediaBangumi {
			cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
		}
		for _, v := range all.Result.MediaFt {
			cardItem = append(cardItem, &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID)})
		}
		// archive
		for _, v := range all.Result.Video {
			if _, ok := _m[v.ID]; !ok { //去重
				cardItem = append(cardItem, &ai.Item{Goto: model.GotoAv, ID: int64(v.ID)})
			}
		}
	}
	for _, v := range cardItem {
		if v.ID == 0 {
			continue
		}
		switch v.Goto {
		case model.GotoAv:
			aids = append(aids, v.ID)
		case model.GotoPGC:
			ssids = append(ssids, int32(v.ID))
		}
	}

	group := errgroup.WithContext(c)
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.Archives(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seamAids, err = s.bgm.CardsByAidsAll(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(ssids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsAll(ctx, ssids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
		return []*cardm.MediaItem{}, keywordURL, nil
	}
	is := []*cardm.MediaItem{}
	for _, v := range cardItem {
		var (
			main interface{}
			gt   string
		)
		switch v.Goto {
		case model.GotoAv:
			// 如果当前稿件是pgc视频，转成pgc卡片处理，否则还是稿件卡片
			main = arcs
			gt = model.GotoAv
			if _, ok := seamAids[int32(v.ID)]; ok {
				main = seamAids
				gt = model.GotoPGC
			}
		case model.GotoPGC:
			main = seams
			gt = model.GotoPGC
		}
		materials := &cardm.Materials{Prune: cardm.GtPrune(model.GotoAv, v.ID)}
		i := &cardm.MediaItem{}
		ok := i.FromItem(v.ID, gt, main, materials)
		if !ok {
			continue
		}
		is = append(is, i)
	}
	if len(is) == 0 {
		return nil, keywordURL, xecode.AppMediaNotData
	}
	//取头部的N个
	if len(is) > ps {
		is = is[0:ps]
	}
	return is, keywordURL, nil
}

func (s *Service) makeCardItem(orgKeyword string, cardItem []*ai.Item, _m map[int64]bool) ([]*ai.Item, map[int64]bool) {
	//插卡的数据获取点,是否打开开关，和aids做去重，取前N个数据，且是ugc的
	if !s.c.CustomModule.CircuitIntervene {
		interv, ok := s.xiaoPengRecs[fmt.Sprintf("%d_%s", _InterveneTypeTab, orgKeyword)]
		if ok && interv != nil {
			for _, _aid := range interv.Items {
				if _, ok := _m[_aid]; !ok { //去重
					cardItem = append(cardItem, &ai.Item{Goto: model.GotoAv, ID: _aid})
					_m[_aid] = true
				}
			}
		}
	}
	return cardItem, _m
}
