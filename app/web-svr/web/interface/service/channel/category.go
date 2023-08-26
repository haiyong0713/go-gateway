package channel

import (
	"context"

	"go-common/library/log"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	chanmdl "go-gateway/app/web-svr/web/interface/model/channel"

	changrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	cardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	"go-common/library/sync/errgroup.v2"
)

func (s *Service) CategoryList(c context.Context) (*chanmdl.CategoryListReply, error) {
	categoryReply, err := s.chDao.Category(c)
	if err != nil {
		log.Error("[CategoryList] s.chDao.Category (%+v)", err)
		return nil, err
	}
	var categories []*chanmdl.Category
	for _, item := range categoryReply.GetCategorys() {
		if item == nil {
			log.Warn("[CategoryList] Category.ChannelCategory is nil")
			continue
		}
		category := &chanmdl.Category{}
		category.FormChannelCategory(item)
		categories = append(categories, category)
	}
	return &chanmdl.CategoryListReply{Categories: categories}, nil
}

func (s *Service) ChannelArcList(c context.Context, mid int64, req *chanmdl.ChannelArcListReq) (*chanmdl.ChannelArcListReply, error) {
	// 获取分类下的频道列表
	chanResListReq := &changrpc.ChannelResourceListReq{
		Mid:          mid,
		Offset:       req.Offset,
		CategoryType: req.ID,
		Ps:           chanmdl.CategoryChanListPS,    //一次请求需要多少个频道
		Count:        chanmdl.CategoryChanListCount, //一个频道卡里面需要几个视频卡
	}
	chanResListReply, err := s.chDao.ChannelResourceList(c, chanResListReq)
	if err != nil {
		log.Error("[ChannelArcList] s.chDao.ChannelList(%+v) (%+v)", chanResListReq, err)
		return nil, err
	}
	// 格式化 viewChannelCard
	arcChannels := s.formatViewChanCards(c, chanResListReply.GetCard(), true)
	return &chanmdl.ChannelArcListReply{
		HasMore:     chanResListReply.GetHasMore(),
		Offset:      chanResListReply.GetOffset(),
		Total:       chanResListReply.Count,
		ArcChannels: arcChannels,
	}, nil
}

func (s *Service) formatViewChanCards(c context.Context, viewChanCards []*changrpc.ViewChannelCard, needArc bool) []*chanmdl.ArcChannel {
	var (
		err           error
		g             = errgroup.WithContext(c)
		seasonCardMap = make(map[int64]*cardgrpc.SeasonCards)
		arcs          = make(map[int64]*arcgrpc.Arc)
	)
	if needArc {
		// 查询频道剧集数 season_count
		var channelIDs []int64
		for _, card := range viewChanCards {
			if card == nil || card.GetCid() == 0 {
				continue
			}
			channelIDs = append(channelIDs, card.GetCid())
		}
		if len(channelIDs) > 0 {
			g.Go(func(ctx context.Context) error {
				seasonCardMap, err = s.dao.TagOGV(ctx, channelIDs)
				if err != nil {
					log.Error("[formatViewChanCards] s.dao.TagOGV(%+v): %+v", channelIDs, err)
				}
				return nil
			})
		}
		// 查询取稿件数据
		g.Go(func(ctx context.Context) error {
			var aids []int64
			for _, card := range viewChanCards {
				for _, videoCard := range card.GetVideoCards() {
					aids = append(aids, videoCard.GetRid())
				}
			}
			if len(aids) == 0 {
				return nil
			}
			arcs, err = s.dao.Arcs(c, aids)
			if err != nil {
				log.Error("[formatViewChanCards] s.dao.Arcs(%+v) (%+v)", aids, err)
				return err
			}
			return nil
		})
		if err := g.Wait(); err != nil {
			log.Error("[formatViewChanCards] g.Wait (%+v)", err)
		}
	}
	// 组装ArcChannel
	arcChannels := make([]*chanmdl.ArcChannel, 0, len(viewChanCards))
	for _, card := range viewChanCards {
		// 不展示无视频的频道
		if len(card.GetVideoCards()) == 0 {
			continue
		}
		arcChannel := &chanmdl.ArcChannel{Archives: make([]*chanmdl.Archive, 0)}
		// 组装Channel部分
		arcChannel.WebChannel.FormViewChannelCard(card)
		arcChannel.WebChannel.FormSeason(card.GetPGC(), seasonCardMap[card.GetCid()])
		// 组装Archives部分
		if needArc {
			for _, videoCard := range card.GetVideoCards() {
				if _, ok := arcs[videoCard.GetRid()]; !ok {
					continue
				}
				archive := &chanmdl.Archive{}
				archive.FormVideoCard(videoCard)
				archive.FormArc(arcs[videoCard.GetRid()])
				arcChannel.Archives = append(arcChannel.Archives, archive)
			}
		}
		arcChannels = append(arcChannels, arcChannel)
	}
	return arcChannels
}

func (s *Service) ChannelList(c context.Context, mid int64, req *chanmdl.ChannelListReq) (*chanmdl.ChannelListReply, error) {
	listReq := &changrpc.ChannelListReq{
		Mid:          mid,
		CategoryType: req.ID,
		Offset:       req.Offset,
		Ps:           req.PageSize,
	}
	listReply, err := s.chDao.ChannelList(c, listReq)
	if err != nil {
		log.Error("[ChannelList] s.chDao.ChannelList(%+v) (%+v)", listReq, err)
		return nil, err
	}
	channels := make([]*chanmdl.WebChannel, 0)
	for _, card := range listReply.GetCards() {
		ch := &chanmdl.WebChannel{}
		ch.FormChannelCard(card)
		channels = append(channels, ch)
	}
	reply := &chanmdl.ChannelListReply{
		HasMore:  listReply.GetHasMore(),
		Offset:   listReply.GetNextOffset(),
		Total:    listReply.GetCount(),
		Channels: channels,
	}
	return reply, nil
}
