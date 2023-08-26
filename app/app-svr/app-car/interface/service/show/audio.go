package show

import (
	"context"
	"strconv"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	xecode "go-gateway/app/app-svr/app-car/ecode"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/audio"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/show"

	listenerChannelgrpc "git.bilibili.co/bapis/bapis-go/car-channel/interface"
)

const (
	_audioHis       = "auido_history"
	_audioFeed      = "audio_feed"
	_audioVertical  = "audio_vertical"
	_defaultAudioPn = 0
	_defaultAudioPs = 10
)

func (s *Service) AudioShow(c context.Context, mid int64, plat int8, buvid string, param *audio.ShowAudioParam) (*show.AudioShow, error) {
	var (
		aids, feedAids []int64
		mutex          sync.Mutex
		channels       []*listenerChannelgrpc.ChannelRecommendInfo
		channelsm      map[int64][]*listenerChannelgrpc.ChannelRecommendInfo
		// 列表顺序
		cardsTypes = []string{_audioFeed}
		cardlist   []*ai.Item
		entrance   string
	)
	channelm := map[int64]*listenerChannelgrpc.ChannelRecommendInfo{}
	// 第一次批量
	group := errgroup.WithContext(c)
	// 推荐
	if mid > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if feedAids, err = s.dyn.RecommendArchives(ctx, mid, buvid, param.Build, param.MobiApp, param.Platform, param.Device, param.Channel); err != nil {
				log.Error("%+v", err)
				return err
			}
			items := []*ai.Item{}
			for _, v := range feedAids {
				item := &ai.Item{Goto: model.GotoAv, ID: v}
				items = append(items, item)
			}
			entrance = model.EntranceAudioFeed
			mutex.Lock()
			cardlist = items
			mutex.Unlock()
			return nil
		})
	}
	// 垂类推荐
	group.Go(func(ctx context.Context) (err error) {
		if channels, err = s.dyn.ChannelRecommend(ctx, _defaultAudioPn, _defaultPs, param.Build, buvid, mid, 0); err != nil {
			log.Error("%+v", err)
			return err
		}
		for _, v := range channels {
			channelm[v.Id] = v
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	// 垂类模块
	for _, v := range channels {
		cardsTypes = append(cardsTypes, strconv.FormatInt(int64(v.Id), 10))
	}
	// 列表里面的历史记录放到末尾
	cardsTypes = append(cardsTypes, _audioHis)
	// 第二次批量
	group = errgroup.WithContext(c)
	if len(channels) > 0 {
		var channelIds []int64
		for _, v := range channels {
			channelIds = append(channelIds, v.Id)
		}
		// 只有未登陆才去获取
		if mid == 0 {
			group.Go(func(ctx context.Context) error {
				// 只获取第一个频道稿件数据
				channelId := channelIds[0]
				channelRcmd, err := s.dyn.ChannelFeedRecommend(ctx, _defaultAudioPn, _defaultAudioPs, param.Build, buvid, mid, channelId)
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				items := []*ai.Item{}
				for _, v := range channelRcmd {
					item := &ai.Item{Goto: model.GotoAv, ID: v}
					items = append(items, item)
				}
				entrance = model.EntranceAudioChannel
				mutex.Lock()
				cardlist = items
				mutex.Unlock()
				return nil
			})
		}
		// 垂类推荐
		group.Go(func(ctx context.Context) (err error) {
			if channelsm, err = s.dyn.ChannelRecommends(ctx, _defaultAudioPn, _defaultPs, param.Build, buvid, mid, channelIds); err != nil {
				log.Error("%+v", err)
				return err
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	for _, v := range cardlist {
		aids = append(aids, v.ID)
	}
	arcs, err := s.arc.ArcsPlayerAll(c, aids)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	materials := &card.Materials{
		ArcPlayers: arcs,
	}
	tabItems := []*show.TabItem{}
	var isDefault bool
	for _, ct := range cardsTypes {
		item := &show.TabItem{Type: ct}
		switch item.Type {
		case _audioFeed, _audioHis:
			// 未登陆的时候为你推荐模块不下发
			if item.Type == _audioFeed {
				if mid == 0 {
					continue
				}
			}
			isDefault = item.FromAudioItem()
		default:
			channelID, _ := strconv.ParseInt(ct, 10, 64)
			channel, ok := channelm[channelID]
			if !ok {
				continue
			}
			// 二级标签
			if chls, ok := channelsm[channel.Id]; ok {
				item.FromAudioTabs(chls, channel, isDefault)
			}
			isDefault = item.FromChannel(isDefault)
			item.Title = channel.Name
			item.Type = _audioVertical
			item.ChannelID = channelID
		}
		tabItems = append(tabItems, item)
	}
	op := &operate.Card{}
	cardParam := &card.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromList,
		MobiApp:  param.MobiApp,
		Build:    param.Build,
	}
	list := s.cardDealItem(cardParam, cardlist, entrance, model.SmallCoverV4, materials, op)
	res := &show.AudioShow{
		TabItems: tabItems,
		Items:    list,
	}
	return res, nil
}

func (s *Service) AudioFeed(c context.Context, plat int8, mid int64, buvid string, param *audio.ShowAudioParam) ([]cardm.Handler, error) {
	feedAids, err := s.dyn.RecommendArchives(c, mid, buvid, param.Build, param.MobiApp, param.Platform, param.Device, param.Channel)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	cards := []*ai.Item{}
	for _, v := range feedAids {
		item := &ai.Item{
			Goto: model.GotoAv,
			ID:   v,
		}
		cards = append(cards, item)
	}
	arcs, err := s.arc.ArcsPlayerAll(c, feedAids)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	materials := &card.Materials{
		ArcPlayers: arcs,
	}
	cardParam := &card.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromList,
		MobiApp:  param.MobiApp,
		Build:    param.Build,
	}
	op := &operate.Card{}
	list := s.cardDealItem(cardParam, cards, model.EntranceCommonSearch, model.SmallCoverV4, materials, op)
	if len(list) == 0 {
		return []cardm.Handler{}, nil
	}
	return list, nil
}

func (s *Service) ReportPlayAction(c context.Context, mid int64, buvid string, param *audio.ReportPlayParam) error {
	reply, err := s.dyn.ReportPlayAction(c, mid, buvid, param.Aid, param.Cid, param.Detail)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	if !reply {
		return xecode.AppReportPlayError
	}
	return nil
}

func (s *Service) AudioChannel(c context.Context, plat int8, mid int64, buvid string, param *audio.ChannelAudioParam) ([]cardm.Handler, error) {
	channelAids, err := s.dyn.ChannelFeedRecommend(c, param.Pn, _defaultAudioPs, param.Build, buvid, mid, param.ChannelID)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	cards := []*ai.Item{}
	for _, v := range channelAids {
		item := &ai.Item{
			Goto: model.GotoAv,
			ID:   v,
		}
		cards = append(cards, item)
	}
	arcs, err := s.arc.ArcsPlayerAll(c, channelAids)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	materials := &card.Materials{
		ArcPlayers: arcs,
	}
	cardParam := &card.CardParam{
		Plat:     plat,
		Mid:      mid,
		FromType: model.FromList,
		MobiApp:  param.MobiApp,
		Build:    param.Build,
	}
	op := &operate.Card{}
	list := s.cardDealItem(cardParam, cards, model.EntranceCommonSearch, model.SmallCoverV4, materials, op)
	if len(list) == 0 {
		return []cardm.Handler{}, nil
	}
	return list, nil
}
