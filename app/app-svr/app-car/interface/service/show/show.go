package show

import (
	"context"
	"fmt"
	"sync"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/interface/model"
	"go-gateway/app/app-svr/app-car/interface/model/bangumi"
	"go-gateway/app/app-svr/app-car/interface/model/card"
	cardm "go-gateway/app/app-svr/app-car/interface/model/card"
	"go-gateway/app/app-svr/app-car/interface/model/card/ai"
	"go-gateway/app/app-svr/app-car/interface/model/card/operate"
	"go-gateway/app/app-svr/app-car/interface/model/dynamic"
	"go-gateway/app/app-svr/app-car/interface/model/show"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	"go-common/library/sync/errgroup.v2"

	cardappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	seasongrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/season"
)

const (
	_popular         = "popular"
	_myBangumi       = "my_bangumi"
	_bangumi         = "bangumi"
	_domestic        = "domestic"
	_myCinema        = "my_cinema"
	_cinema          = "cinema"
	_cinemaDoc       = "cinema_doc"
	_feed            = "feed"
	_banner          = "banner"
	_dynamicVideoNew = "dynamic_video_new"
	_max             = 20
	_popularmax      = 25
	_defaultPn       = 1
	_defaultPs       = 20
	// 动态最小值
	_dynListMin = 3
	// channel type
	_sound = "sound"
)

var (
	pgcType = map[string]string{
		_myBangumi: _followTypeBangumi,
		_myCinema:  _followTypeCinema,
		_bangumi:   _followTypeBangumi,
		_domestic:  _followTypeDomestic,
		_cinema:    _followTypeCinema,
		_cinemaDoc: _followTypeCinemaDoc,
	}
)

func regionKey(rid int64) string {
	return fmt.Sprintf("region_%d", rid)
}

// nolint: gocognit
func (s *Service) Show(c context.Context, mid int64, plat int8, buvid string, param *show.ShowParam) ([]*show.Item, *show.Config, error) {
	var (
		regionType   = []int64{3, 129, 4, 36, 202, 223, 160}
		cardsTypes   = []string{_popular, _myBangumi, _bangumi, _domestic, _myCinema, _cinema, _cinemaDoc}
		cardsTypesV2 = []string{_banner, _dynamicVideoNew, _feed, regionKey(model.CustomModuleRid51), regionKey(model.CustomModuleRid61Childhood), regionKey(model.CustomModuleRid61Eden), regionKey(model.CustomModuleRidDW),
			_popular, regionKey(3), regionKey(129), regionKey(4), regionKey(36), regionKey(160), regionKey(202), regionKey(223)}
		cardsTypesV3 = []string{_banner, _dynamicVideoNew, _feed, regionKey(model.CustomModuleRid51), regionKey(model.CustomModuleRid61Childhood), regionKey(model.CustomModuleRid61Eden), regionKey(model.CustomModuleRidDW),
			_popular, regionKey(223), regionKey(160), regionKey(3), regionKey(129), regionKey(4), regionKey(36), regionKey(202)}
		mutex             sync.Mutex
		aids              []int64
		ssids, epinlinIds []int32
		seams             map[int32]*episodegrpc.EpisodeCardsProto
		seasonm           map[int32]*seasongrpc.CardInfoProto
		epm               map[int32]*pgcinline.EpisodeCard
		dynList           *dynamic.DynVideoListRes
		dynTypeList       = []string{"8"}
		dynUpdateNum      int64
	)
	// config配置
	config := &show.Config{DynUpdateNumber: param.DynUpdateNumber}
	listm := map[string][]*ai.Item{}
	newlistm := map[string][]*ai.Item{}
	// 物料
	arcs := map[int64]*arcgrpc.ArcPlayer{}
	animem := map[int32]*cardappgrpc.CardSeasonProto{}
	regionArcm := map[int64]*arcgrpc.Arc{}
	bangumim := map[int32]*bangumi.Module{}
	// 分区key映射关系
	regionm := map[string]int64{}
	// 第一次批量
	group := errgroup.WithContext(c)
	// 动态最新视频
	if param.Build >= 1100000 && mid > 0 {
		group.Go(func(ctx context.Context) error {
			// 获取用户关注链信息(关注的up、追番、购买的课程）
			following, pgcFollowing, err := s.followings(c, mid)
			if err != nil {
				log.Error("%+v", err)
			}
			attentions := dynamic.GetAttentionsParams(mid, following, pgcFollowing)
			dynList, err = s.dyn.DynVideoList(c, mid, "", "20", dynTypeList, attentions, param.Build, param.Platform, param.MobiApp, buvid, param.Device)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			items := []*ai.Item{}
			for _, v := range dynList.Dynamics {
				item := v.DynamicCardChange()
				items = append(items, item)
			}
			dynUpdateNum = dynList.UpdateNum
			mutex.Lock()
			listm[_dynamicVideoNew] = items
			mutex.Unlock()
			return nil
		})
	}
	// 热门
	group.Go(func(ctx context.Context) error {
		key := s.popularGroup(mid, buvid)
		list := s.PopularCardTenList(ctx, key, 0, _popularmax)
		items := []*ai.Item{}
		for _, v := range list {
			item := v.PopularCardToAiChange()
			item.Entrance = model.EntrancePopular
			items = append(items, item)
		}
		mutex.Lock()
		listm[_popular] = items
		// v1.4热门一部署数据用于banner
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.HotBanner, &feature.OriginResutl{
			MobiApp:    param.MobiApp,
			Device:     param.Device,
			Build:      int64(param.Build),
			BuildLimit: param.Build >= 1040000,
		}) && len(items) >= 10 {
			listm[_banner] = items[:5]
			listm[_popular] = items[5:]
		} else {
			if len(items) > _max {
				listm[_popular] = items[:20]
			}
		}
		mutex.Unlock()
		return nil
	})
	// 天马
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.Feed, &feature.OriginResutl{
		MobiApp:    param.MobiApp,
		Device:     param.Device,
		Build:      int64(param.Build),
		BuildLimit: param.Build >= 1030000,
	}) && mid > 0 {
		group.Go(func(ctx context.Context) error {
			group := s.group(mid, buvid)
			list, err := s.rcmd.FeedRecommend(ctx, plat, param.MobiApp, buvid, mid, param.Build, param.LoginEvent, group, _max, 0)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{}
				*item = *v
				item.Entrance = model.EntranceCommonSearch
				items = append(items, item)
			}
			mutex.Lock()
			listm[_feed] = items
			mutex.Unlock()
			return nil
		})
	}
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.Region, &feature.OriginResutl{
		MobiApp:    param.MobiApp,
		Device:     param.Device,
		Build:      int64(param.Build),
		BuildLimit: param.Build >= 1010000,
	}) {
		// 分区
		for _, rid := range regionType {
			tmp := rid
			group.Go(func(ctx context.Context) error {
				list, err := s.reg.RegionDynamic(ctx, tmp, _defaultPn, _defaultPs)
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				items := []*ai.Item{}
				for _, v := range list {
					item := &ai.Item{Goto: model.GotoAv, ID: v.Aid, Entrance: model.EntranceRegion}
					items = append(items, item)
					mutex.Lock()
					regionArcm[v.Aid] = v
					mutex.Unlock()
				}
				mutex.Lock()
				listm[regionKey(tmp)] = items
				regionm[regionKey(tmp)] = tmp
				mutex.Unlock()
				return nil
			})
		}

		// 五一特辑
		func() {
			if s.c.CustomModule51 == nil || !s.c.CustomModule51.EnableCustomModule {
				return
			}
			if s.c.CustomModule51.ChannelAids[param.Channel] == nil {
				return
			}
			if len(s.c.CustomModule51.ChannelAids[param.Channel]) < s.c.CustomModule51.MinNumbers {
				log.Warn("app car custom 51 aid numbers less than 20.")
				return
			}
			items := make([]*ai.Item, 0)
			for _, aid := range s.c.CustomModule51.ChannelAids[param.Channel] {
				item := &ai.Item{Goto: model.GotoAv, ID: aid, Entrance: model.EntranceRegion}
				items = append(items, item)
			}
			mutex.Lock()
			listm[regionKey(model.CustomModuleRid51)] = items
			regionm[regionKey(model.CustomModuleRid51)] = model.CustomModuleRid51
			mutex.Unlock()
		}()

		// 61 童年回来了
		func() {
			if s.c.CustomModule61Childhood == nil || !s.c.CustomModule61Childhood.EnableCustomModule {
				return
			}
			if s.c.CustomModule61Childhood.ChannelAids[param.Channel] == nil {
				return
			}
			if len(s.c.CustomModule61Childhood.ChannelAids[param.Channel]) < s.c.CustomModule61Childhood.MinNumbers {
				log.Warn("app car custom 61 aid numbers less than 20.")
				return
			}
			items := make([]*ai.Item, 0)
			for _, aid := range s.c.CustomModule61Childhood.ChannelAids[param.Channel] {
				item := &ai.Item{Goto: model.GotoAv, ID: aid, Entrance: model.EntranceRegion}
				items = append(items, item)
			}
			mutex.Lock()
			listm[regionKey(model.CustomModuleRid61Childhood)] = items
			regionm[regionKey(model.CustomModuleRid61Childhood)] = model.CustomModuleRid61Childhood
			mutex.Unlock()
		}()

		// 61 小朋友乐园
		func() {
			if s.c.CustomModule61Eden == nil || !s.c.CustomModule61Eden.EnableCustomModule {
				return
			}
			if s.c.CustomModule61Eden.ChannelAids[param.Channel] == nil {
				return
			}
			if len(s.c.CustomModule61Eden.ChannelAids[param.Channel]) < s.c.CustomModule61Eden.MinNumbers {
				log.Warn("app car custom 61 aid numbers less than 20.")
				return
			}
			items := make([]*ai.Item, 0)
			for _, aid := range s.c.CustomModule61Eden.ChannelAids[param.Channel] {
				item := &ai.Item{Goto: model.GotoAv, ID: aid, Entrance: model.EntranceRegion}
				items = append(items, item)
			}
			mutex.Lock()
			listm[regionKey(model.CustomModuleRid61Eden)] = items
			regionm[regionKey(model.CustomModuleRid61Eden)] = model.CustomModuleRid61Eden
			mutex.Unlock()
		}()

		// 端午 "粽”有陪伴
		func() {
			if s.c.CustomModuleDW == nil || !s.c.CustomModuleDW.EnableCustomModule {
				return
			}
			if s.c.CustomModuleDW.ChannelAids[param.Channel] == nil {
				return
			}
			if len(s.c.CustomModuleDW.ChannelAids[param.Channel]) < s.c.CustomModuleDW.MinNumbers {
				log.Warn("app car custom dw aid numbers less than 20.")
				return
			}
			items := make([]*ai.Item, 0)
			for _, aid := range s.c.CustomModuleDW.ChannelAids[param.Channel] {
				item := &ai.Item{Goto: model.GotoAv, ID: aid, Entrance: model.EntranceRegion}
				items = append(items, item)
			}
			mutex.Lock()
			listm[regionKey(model.CustomModuleRidDW)] = items
			regionm[regionKey(model.CustomModuleRidDW)] = model.CustomModuleRidDW
			mutex.Unlock()
		}()

	} else {
		if mid > 0 {
			// 我的追番
			group.Go(func(ctx context.Context) error {
				const followType = 1
				list, err := s.bgm.MyFollows(ctx, mid, followType, _defaultPn, _defaultPs)
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				items := []*ai.Item{}
				for _, v := range list {
					item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonId), Entrance: model.EntranceMyAnmie}
					items = append(items, item)
					mutex.Lock()
					animem[v.SeasonId] = v
					mutex.Unlock()
				}
				mutex.Lock()
				listm[_myBangumi] = items
				mutex.Unlock()
				return nil
			})
			// 我的追剧
			group.Go(func(ctx context.Context) error {
				const followType = 2
				list, err := s.bgm.MyFollows(ctx, mid, followType, _defaultPn, _defaultPs)
				if err != nil {
					log.Error("%+v", err)
					return nil
				}
				items := []*ai.Item{}
				for _, v := range list {
					item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonId), Entrance: model.EntranceMyAnmie}
					items = append(items, item)
					mutex.Lock()
					animem[v.SeasonId] = v
					mutex.Unlock()
				}
				mutex.Lock()
				listm[_myCinema] = items
				mutex.Unlock()
				return nil
			})
		}
		// 18 番剧推荐
		group.Go(func(ctx context.Context) error {
			const followType = 18
			list, err := s.bgm.Module(ctx, followType, param.MobiApp, buvid)
			if err != nil {
				log.Error("%+v", err)
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID), Entrance: model.EntrancePgcList}
				items = append(items, item)
				mutex.Lock()
				bangumim[v.SeasonID] = v
				mutex.Unlock()
			}
			mutex.Lock()
			listm[_bangumi] = items
			mutex.Unlock()
			return nil
		})
		// 19 国创推荐
		group.Go(func(ctx context.Context) error {
			const followType = 19
			list, err := s.bgm.Module(ctx, followType, param.MobiApp, buvid)
			if err != nil {
				log.Error("%+v", err)
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID), Entrance: model.EntrancePgcList}
				items = append(items, item)
				mutex.Lock()
				bangumim[v.SeasonID] = v
				mutex.Unlock()
			}
			mutex.Lock()
			listm[_domestic] = items
			mutex.Unlock()
			return nil
		})
		// 88 电影热播
		group.Go(func(ctx context.Context) error {
			const followType = 88
			list, err := s.bgm.Module(ctx, followType, param.MobiApp, buvid)
			if err != nil {
				log.Error("%+v", err)
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID), Entrance: model.EntrancePgcList}
				items = append(items, item)
				mutex.Lock()
				bangumim[v.SeasonID] = v
				mutex.Unlock()
			}
			mutex.Lock()
			listm[_cinema] = items
			mutex.Unlock()
			return nil
		})
		// 87 纪录片热播
		group.Go(func(ctx context.Context) error {
			const followType = 87
			list, err := s.bgm.Module(ctx, followType, param.MobiApp, buvid)
			if err != nil {
				log.Error("%+v", err)
			}
			items := []*ai.Item{}
			for _, v := range list {
				item := &ai.Item{Goto: model.GotoPGC, ID: int64(v.SeasonID), Entrance: model.EntrancePgcList}
				items = append(items, item)
				mutex.Lock()
				bangumim[v.SeasonID] = v
				mutex.Unlock()
			}
			mutex.Lock()
			listm[_cinemaDoc] = items
			mutex.Unlock()
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	for _, list := range listm {
		for _, v := range list {
			switch v.Goto {
			case model.GotoAv:
				aids = append(aids, v.ID)
			case model.GotoPGC:
				ssids = append(ssids, int32(v.ID))
			}
		}
	}
	// 第二次批量
	group = errgroup.WithContext(c)
	// 获取稿件物料
	if len(aids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if arcs, err = s.arc.ArcsPlayerAll(ctx, aids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		group.Go(func(ctx context.Context) (err error) {
			if seams, err = s.bgm.CardsByAidsAll(ctx, aids); err != nil {
				log.Error("%+v", err)
				return nil
			}
			for _, ep := range seams {
				epinlinIds = append(epinlinIds, ep.EpisodeId)
			}
			return nil
		})
	}
	if len(ssids) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if seasonm, err = s.bgm.CardsAll(ctx, ssids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	// 第三次批量
	group = errgroup.WithContext(c)
	if len(epinlinIds) > 0 {
		group.Go(func(ctx context.Context) (err error) {
			if epm, err = s.bgm.InlineCardsAll(ctx, epinlinIds, param.MobiApp, param.Platform, param.Device, param.Build); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("%+v", err)
	}
	// 除重，如果动态最新视频<=3和天马feed流里面匹配去重复
	for k, v := range listm {
		newlistm[k] = v
	}
	// 上一次的动态更新数+当前动态更新数
	config.DynUpdateNumber += dynUpdateNum
	// nolint:gomnd
	if param.Build >= 1110000 {
		if dynlist, ok := listm[_dynamicVideoNew]; ok && (config.DynUpdateNumber <= _dynListMin || len(dynlist) <= _dynListMin) {
			if len(listm[_feed]) > 0 {
				var newfeedlist []*ai.Item
				for _, v := range dynlist {
					d := &ai.Item{}
					*d = *v
					d.Entrance = model.EntranceCommonSearch
					newfeedlist = append(newfeedlist, d)
				}
				newfeedlistm := map[string]struct{}{}
				if len(newfeedlist) > _dynListMin {
					newfeedlist = newfeedlist[:_dynListMin]
				}
				for _, v := range newfeedlist {
					key := fmt.Sprintf("%d_%s", v.ID, v.Goto)
					newfeedlistm[key] = struct{}{}
				}
				if feedlist, ok := listm[_feed]; ok {
					for _, v := range feedlist {
						tmp := v
						key := fmt.Sprintf("%d_%s", tmp.ID, tmp.Goto)
						if _, ok := newfeedlistm[key]; ok {
							continue
						}
						newfeedlist = append(newfeedlist, tmp)
					}
				}
				newlistm[_feed] = newfeedlist
			}
			// 音响和车载去除最新视频
			cardsTypesV2 = []string{_banner, _feed, regionKey(model.CustomModuleRid51), regionKey(model.CustomModuleRid61Childhood), regionKey(model.CustomModuleRid61Eden), regionKey(model.CustomModuleRidDW),
				_popular, regionKey(3), regionKey(129), regionKey(4), regionKey(36), regionKey(160), regionKey(202), regionKey(223)}
			cardsTypesV3 = []string{_banner, _feed, regionKey(model.CustomModuleRid51), regionKey(model.CustomModuleRid61Childhood), regionKey(model.CustomModuleRid61Eden), regionKey(model.CustomModuleRidDW),
				_popular, regionKey(223), regionKey(160), regionKey(3), regionKey(129), regionKey(4), regionKey(36), regionKey(202)}
		}
	}
	list := cardsTypes
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.ShowListTab, &feature.OriginResutl{
		MobiApp:    param.MobiApp,
		Device:     param.Device,
		Build:      int64(param.Build),
		BuildLimit: param.Build >= 1010000,
	}) {
		// 默认分区车载第一位
		list = cardsTypesV3
		// 当前是音响则车载分区放最后
		if _, ok := s.c.Custom.ChannelType.Sound[param.Channel]; param.ChannelType == _sound || ok {
			list = cardsTypesV2
		}
	}
	// 列表处理
	if arcs == nil {
		arcs = map[int64]*arcgrpc.ArcPlayer{}
	}
	for _, v := range regionArcm {
		arcs[v.Aid] = &arcgrpc.ArcPlayer{
			Arc: v,
		}
	}
	materials := &card.Materials{
		Animem:             animem,
		ArcPlayers:         arcs,
		Bangumim:           bangumim,
		Seams:              seasonm,
		EpisodeCardsProtom: seams,
		EpInlinem:          epm,
	}
	items := []*show.Item{}
	for _, ct := range list {
		var (
			entrance string
			cardType model.CardType
		)
		op := &operate.Card{FollowType: pgcType[ct]}
		cards, ok := newlistm[ct]
		if !ok {
			continue
		}
		if regionID, ok := regionm[ct]; ok {
			op.Rid = regionID
		}
		cardParam := &card.CardParam{
			Plat:     plat,
			Mid:      mid,
			FromType: model.FromList,
			MobiApp:  param.MobiApp,
			Build:    param.Build,
		}
		cardType = model.SmallCoverV1
		switch ct {
		case _feed:
			entrance = model.EntranceCommonSearch
		case _banner:
			// 小度音响不出banner
			if param.Channel == "xiaodu" {
				continue
			}
			cardType = model.BannerV1
			cardParam.IsPlayer = true
			entrance = model.EntrancePopular
		case _popular:
			entrance = model.EntrancePopular
		case _myBangumi, _myCinema:
			entrance = model.EntranceMyAnmie
			cardType = model.SmallCoverV2
		case _domestic, _cinema, _cinemaDoc, _bangumi:
			entrance = model.EntrancePgcList
			cardType = model.VerticalCoverV1
		case _dynamicVideoNew:
			entrance = model.EntranceDynamicVideoNew
		case regionKey(model.CustomModuleRid51), regionKey(model.CustomModuleRid61Childhood), regionKey(model.CustomModuleRid61Eden), regionKey(model.CustomModuleRidDW):
			entrance = model.EntranceRegion
		default:
			for _, rid := range regionType {
				if regionKey(rid) == ct {
					entrance = model.EntranceRegion
					break
				}
			}
		}
		list := s.cardDealItem(cardParam, cards, entrance, cardType, materials, op)
		if len(list) == 0 {
			continue
		}
		item := &show.Item{Type: ct, Items: list}
		// 动态更新数
		if ct == _dynamicVideoNew {
			item.UpdateNumber = dynUpdateNum
		}
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.ShowFromTag, &feature.OriginResutl{
			MobiApp:    param.MobiApp,
			Device:     param.Device,
			Build:      int64(param.Build),
			BuildLimit: param.Build >= 1010000,
		}) {
			item.FromItem2(param.Channel)
		} else {
			item.FromItem()
		}
		items = append(items, item)
	}
	return items, config, nil
}

func (s *Service) cardDealItem(param *card.CardParam, feedCard []*ai.Item, entrance string, cardType model.CardType, materials *card.Materials, op *operate.Card) []cardm.Handler {
	is := []cardm.Handler{}
	for _, v := range feedCard {
		var (
			main interface{}
		)
		op.TrackID = v.TrackID
		switch param.FromType {
		case model.FromView:
			cardType = model.SmallCoverV4
		}
		op.From(model.CardGt(v.Goto), entrance, v.ID, param.Plat, param.Build, param.MobiApp)
		op.DynCtime = v.DynCtime
		switch v.Goto {
		case model.GotoAv:
			materials.Prune = cardm.GtPrune(v.Goto, v.ID)
			arc, ok := materials.ArcPlayers[v.ID]
			if !ok {
				continue
			}
			main = map[int64]*arcgrpc.Arc{arc.Arc.Aid: arc.Arc}
			// 只有banner需要秒开逻辑，其他卡片都不要
			if param.IsPlayer {
				main = materials.ArcPlayers
			}
			if ep, ok := materials.EpisodeCardsProtom[int32(v.ID)]; ok {
				main = materials.EpisodeCardsProtom
				v.Goto = model.GotoPGC
				op.Epid = ep.EpisodeId
			}
		case model.GotoAvHis, model.GotoAvView:
			materials.Prune = cardm.GtPrune(v.Goto, v.ID)
			_, ok := materials.ViewReplym[v.ID]
			if !ok {
				continue
			}
			op.Goto = model.GotoAv
			main = materials.ViewReplym
			if ep, ok := materials.EpisodeCardsProtom[int32(v.ID)]; ok {
				main = materials.EpisodeCardsProtom
				v.Goto = model.GotoPGC
				op.Epid = ep.EpisodeId
			}
		case model.GotoPGC:
			main = materials.Seams
			switch entrance {
			case model.EntranceMyAnmie:
				anim, ok := materials.Animem[int32(v.ID)]
				if !ok {
					continue
				}
				main = anim
				materials.Prune = cardm.PGCFollowPrune(anim)
			case model.EntrancePgcList:
				anim, ok := materials.Bangumim[int32(v.ID)]
				if !ok {
					continue
				}
				materials.Prune = cardm.PGCModulePrune(anim)
			}
		case model.GotoPGCEp:
			switch entrance {
			case model.EntrancePgcRcmdList:
				main = materials.Epms
			}
		}
		h := cardm.Handle(param.Plat, model.CardGt(v.Goto), cardType, v, materials)
		if h == nil || !h.From(main, op) {
			continue
		}
		// 过滤互动视频
		if h.Get().Filter == model.FilterAttrBitSteinsGate {
			continue
		}
		is = append(is, h)
	}
	if len(is) == 0 {
		return []cardm.Handler{}
	}
	return is
}
