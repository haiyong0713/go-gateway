package channel_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	channelmdl "git.bilibili.co/bapis/bapis-go/community/model/channel"
	dynfeedgrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/feed"
	dynamicTopic "git.bilibili.co/bapis/bapis-go/dynamic/service/topic"
	natgrpc "git.bilibili.co/bapis/bapis-go/natpage/interface/service"
	appCardgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	topiccommon "git.bilibili.co/bapis/bapis-go/topic/common"
	topicsvc "git.bilibili.co/bapis/bapis-go/topic/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
	"go-common/component/metadata/device"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	egv2 "go-common/library/sync/errgroup.v2"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	"go-gateway/app/app-svr/app-channel/interface/model"
	"go-gateway/app/app-svr/app-channel/interface/model/channel"
	chmdl "go-gateway/app/app-svr/app-channel/interface/model/channel_v2"
	dynmdl "go-gateway/app/app-svr/app-channel/interface/model/dynamic"
	topicmdl "go-gateway/app/app-svr/app-channel/interface/model/topic"
	channelSvr "go-gateway/app/app-svr/app-channel/interface/service/channel"
	archivegrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
)

const (
	_channelListTopTitle = "全部频道"
	// 服务端动态类型
	_dynTypeDraw  = 2
	_dynTypeVideo = 8
)

type infocStatistics struct {
	AppID    int32  `json:"appId"`
	Platform int32  `json:"platform"`
	Version  string `json:"version"`
	ABTest   string `json:"abtest"`
}

func (s *Service) Tab(c context.Context) (res []*chmdl.ChannelListTab, err error) {
	var tabs []*channelgrpc.ChannelCategory
	if tabs, err = s.chDao.Tabs(c); err != nil {
		log.Error("%v", err)
		return
	}
	for _, tab := range tabs {
		if tab == nil {
			log.Error("tab nil")
			continue
		}
		i := &chmdl.ChannelListTab{}
		i.FormChannelListTab(tab)
		res = append(res, i)
	}
	return
}

func (s *Service) Tab3(c context.Context, mid int64) (res []*chmdl.ChannelListTab, err error) {
	var (
		mineChannels *channelgrpc.SubscribeReply
		tabs         []*channelgrpc.ChannelCategory
	)
	g, ctx := errgroup.WithContext(c)
	if mid > 0 {
		g.Go(func() (err error) {
			if mineChannels, err = s.chDao.SubscribedChannel(ctx, mid, 1); err != nil {
				log.Error("%v", err)
				return nil
			}
			return
		})
	}
	g.Go(func() (err error) {
		if tabs, err = s.chDao.Tabs(ctx); err != nil {
			log.Error("%v", err)
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	dev, _ := device.FromContext(c)
	title := subscribeReplaceText("我的订阅", "我的收藏", card.FavTextReplace(dev.RawMobiApp, dev.Build))
	// 我订阅的tab
	i := &chmdl.ChannelListTab{
		ID:       chmdl.TabMineID,
		TabType:  chmdl.TabTypeMine,
		Title:    title,
		SubTitle: chmdl.TabSubTitle,
	}
	if mineChannels != nil {
		i.Count = mineChannels.Count
	}
	res = append(res, i)
	// 频道服务端返回的其他tab
	for _, tab := range tabs {
		if tab == nil {
			log.Error("tab nil")
			continue
		}
		i := &chmdl.ChannelListTab{}
		i.FormChannelListTab(tab)
		res = append(res, i)
	}
	return
}

func (s *Service) ChannelList(c context.Context, mid int64, ctype int32, offset string, plat int8, build int, mobiApp, device, spmid string) (res *chmdl.ChannelResult, err error) {
	var list *channelgrpc.ChannelListReply
	if list, err = s.chDao.ChannelList(c, mid, ctype, offset); err != nil {
		log.Error("%v", err)
		return
	}
	res = &chmdl.ChannelResult{Offset: list.GetNextOffset(), Title: _channelListTopTitle}
	if list.GetHasMore() {
		res.HasMore = 1
	}
	//版本判断
	var isHighBuild bool
	if pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsMobiAppIPhone().And().Build(">=", s.c.BuildLimit.OGVChanIOSBuild)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", s.c.BuildLimit.OGVChanAndroidBuild)
	}).FinishOr(false) {
		isHighBuild = true
	}
	for k, cardTmp := range list.GetCards() {
		if cardTmp == nil {
			log.Error("list card nil")
			continue
		}
		i := &chmdl.Channel{}
		i.FormChannel(cardTmp, nil, mobiApp, spmid, int64(build), isHighBuild)
		res.Items = append(res.Items, i)
		// 第一刷的第一个数据
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.ChannelListFirstItem, &feature.OriginResutl{
			BuildLimit: model.IsAndroid(plat) && build >= 5500000 && build < 5510000,
		}) && offset == "" && k == 0 {
			res.Items = append(res.Items, i)
		}
	}
	return
}

func (s *Service) Mine(c context.Context, plat int8, build int, mid int64, mobiApp, spmid string) (res *chmdl.ChannelMineResult, err error) {
	subscribeText := subscribeReplaceText("订阅", "收藏", card.FavTextReplace(mobiApp, int64(build)))
	if mid <= 0 {
		res = &chmdl.ChannelMineResult{
			Config: &chmdl.SubscribeConfig{
				Title: _channelListTopTitle,
				Label: fmt.Sprintf("首页会展示你%s的前10个频道", subscribeText),
				LoginButton: &chmdl.Button{
					Label: "你还没有登录哦~\n赶紧登录打开新世界的大门",
					Text:  "立即登录",
				},
			},
		}
		return
	}
	var (
		mineChannels *channelgrpc.SubscribeReply
		scaneds      []*channelgrpc.ViewChannelCard
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		var isNewSub int32
		if pd.WithContext(c).Where(func(pd *pd.PDContext) {
			pd.IsPlatAndroid().Or().IsPlatAndroidG().And().Build(">", int64(s.c.BuildLimit.MineNewSubAndroid))
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatIPhone().And().Build(">", int64(s.c.BuildLimit.MineNewSubIOS))
		}).Or().IsPlatIPhoneI().OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatAndroidI().And().Build(">", int64(2042030))
		}).MustFinish() {
			isNewSub = 1
		}
		if mineChannels, err = s.chDao.SubscribedChannel(ctx, mid, isNewSub); err != nil {
			log.Error("%v", err)
		}
		return
	})
	g.Go(func() (err error) {
		if scaneds, err = s.chDao.ViewChannel(ctx, mid); err != nil {
			log.Error("%v", err)
			return nil
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	// 固定文案、背景图
	res = &chmdl.ChannelMineResult{
		Config: &chmdl.SubscribeConfig{
			Label:      fmt.Sprintf("首页会展示你%s的前10个频道", subscribeText),
			NoSubLabel: fmt.Sprintf("你还没%s过频道\n来看看最近浏览的频道吧~", subscribeText),
			SubLabel:   fmt.Sprintf("%s完成后\n刷新试试～", subscribeText),
			NoSubButton: &chmdl.Button{
				Param: strconv.Itoa(chmdl.TabSelect),
				Label: fmt.Sprintf("你还没有%s过频道\n去看看近期热门的频道吧~", subscribeText),
				Text:  "去看看",
				URI:   model.FillURI(model.GotoChannelTab, strconv.Itoa(chmdl.TabSelect), 0, 0, 0, nil),
			},
			NoMoreButton: &chmdl.Button{
				Param: strconv.Itoa(chmdl.TabSelect),
				Text:  fmt.Sprintf("没有更多%s啦～\n去看看还有哪些", subscribeText) + chmdl.MarkRed("热门频道"),
				URI:   model.FillURI(model.GotoChannelTab, strconv.Itoa(chmdl.TabSelect), 0, 0, 0, nil),
			},
		},
	}
	var actInfos map[int64]*natgrpc.NativePage
	if s.c.Switch.MineActive {
		var actTids []int64
		for _, mcs := range mineChannels.GetTops() {
			if mcs.GetActAttr() == chmdl.ActiveTag {
				actTids = append(actTids, mcs.GetChannelId())
			}
		}
		for _, mcs := range mineChannels.GetCards() {
			if mcs.GetActAttr() == chmdl.ActiveTag {
				actTids = append(actTids, mcs.GetChannelId())
			}
		}
		if len(actTids) > 0 {
			if actInfos, err = s.natDao.NatInfoFromForeigns(c, actTids, 1); err != nil {
				log.Error("%v", err)
				err = nil
			}
		}
	}
	// 我订阅的频道不需要显示默认运营频道
	if mineChannels.GetCount() > 0 {
		//版本判断
		var isHighBuild bool
		if pd.WithContext(ctx).Where(func(pd *pd.PDContext) {
			pd.IsMobiAppIPhone().And().Build(">=", s.c.BuildLimit.OGVChanIOSBuild)
		}).OrWhere(func(pd *pd.PDContext) {
			pd.IsPlatAndroid().And().Build(">=", s.c.BuildLimit.OGVChanAndroidBuild)
		}).FinishOr(false) {
			isHighBuild = true
		}
		// 置顶频道
		for _, mcs := range mineChannels.GetTops() {
			if mcs == nil {
				log.Error("mine stick nil")
				continue
			}
			i := &chmdl.Channel{}
			i.FormChannelMine(mcs, actInfos, isHighBuild, spmid)
			i.SubType = "stick"
			res.Stick = append(res.Stick, i)
		}
		// 非置顶频道
		for _, mcs := range mineChannels.GetCards() {
			if mcs == nil {
				log.Error("mine normal nil")
				continue
			}
			i := &chmdl.Channel{}
			i.FormChannelMine(mcs, actInfos, isHighBuild, spmid)
			i.SubType = "normal"
			res.Normal = append(res.Normal, i)
		}
	} else {
		for _, scaned := range scaneds {
			if scaned == nil {
				continue
			}
			i := &chmdl.Channel{}
			i.FormChannelMineScaned(scaned, subscribeText)
			res.Scaned = append(res.Scaned, i)
		}
	}
	return
}

func (s *Service) ChannelSort(c context.Context, mid int64, action int32, stick, normal string) (err error) {
	if err = s.chDao.ChannelSort(c, mid, action, stick, normal); err != nil {
		log.Error("%v", err)
	}
	return
}

// nolint:gocognit
func (s *Service) Square(c context.Context, svr *channelSvr.Service, mid, timeIso, ts int64, build, teenagersMode, autoRefresh int, plat int8,
	_, mobiApp, device, lang, offset, buvid, fromSpmid, reqURL, statistics, paramChannel string, pn int) (res *chmdl.SquareResult, err error) {
	var (
		regions         []*channel.Region
		recents         []*channelgrpc.ChannelCard
		myChannels      *channelgrpc.MyChannelsReply
		myChannelBadges = make(map[int64]map[int64]*operate.ChannelBadge)
		ip              = metadata.String(c, metadata.RemoteIP)
	)
	res = &chmdl.SquareResult{}
	g, ctx := errgroup.WithContext(c)
	// 获取分区
	g.Go(func() (err error) {
		_, _, regions, err = svr.RegionList(ctx, build, teenagersMode, mobiApp, device, lang, paramChannel)
		if err != nil {
			log.Error("%v", err)
			err = nil
			return
		}
		res.Region = regions
		return
	})
	// 我订阅的频道+我订阅的频道更新
	g.Go(func() (err error) {
		var isNoFeed bool
		if (model.IsAndroid(plat) && build > s.c.BuildLimit.NoSquareFeedAndroid) || (model.IsIPhone(plat) && build > s.c.BuildLimit.NoSquareFeedIOS) {
			isNoFeed = true
		}
		if myChannels, err = s.chDao.MyChannels(ctx, mid, offset, isNoFeed, chmdl.OldSubVersion); err != nil {
			log.Error("%v", err)
			err = nil
			res.News = &chmdl.New{
				Offset:  offset,
				HasMore: 1,
			}
			return
		}
		if myChannels == nil {
			log.Error("square guanzhu mychannels is nil")
			return
		}
		// 顶部我订阅的频道
		var mcsAll []*channelgrpc.ChannelCard
		res.Subscribe = &chmdl.Subscribe{
			Count: myChannels.GetCount(),
		}
		// fc cp
		for _, tc := range myChannels.GetTops() {
			if tc == nil {
				log.Error("square guanzhu stick nil")
				continue
			}
			mcsAll = append(mcsAll, tc)
		}
		for _, cc := range myChannels.GetNormals() {
			if cc == nil {
				log.Error("square guanzhu normal nil")
				continue
			}
			mcsAll = append(mcsAll, cc)
		}
		for _, mcs := range mcsAll {
			if mcs == nil {
				log.Error("square guanzhu mcs nil")
				continue
			}
			i := &chmdl.Channel{}
			i.FormChannel(mcs, nil, mobiApp, "", int64(build), false)
			res.Subscribe.Items = append(res.Subscribe.Items, i)
			// nolint:gomnd
			if len(res.Subscribe.Items) == 9 {
				break
			}
		}
		// 我订阅的频道的更新
		res.News = &chmdl.New{
			Offset: myChannels.GetNextOffset(),
		}
		if myChannels.GetHasMore() {
			res.News.HasMore = 1
		}
		var (
			totalUpChannelNum = myChannels.GetUpdatedChannelNum()
			aids              []int64
			cardItems         = make(map[int64][]*operate.Card)
		)
		for _, nc := range myChannels.GetCards() {
			if nc == nil {
				log.Error("square guanzhu gengxin card nil")
				continue
			}
			var tmpAids []int64
			for _, resource := range nc.GetResourceCards() {
				if resource != nil && resource.GetVideoCard() != nil && resource.GetVideoCard().GetRid() != 0 {
					tmpAids = append(tmpAids, resource.GetVideoCard().GetRid())
					ci := &operate.Card{
						ID: resource.GetVideoCard().GetRid(),
					}
					cardItems[nc.GetChannelId()] = append(cardItems[nc.GetChannelId()], ci)
					if resource.GetVideoCard().GetBadgeTitle() != "" && resource.GetVideoCard().GetBadgeBackground() != "" {
						var (
							myChannelBadge map[int64]*operate.ChannelBadge
							ok             bool
						)
						if myChannelBadge, ok = myChannelBadges[nc.GetChannelId()]; !ok {
							myChannelBadge = make(map[int64]*operate.ChannelBadge)
							myChannelBadges[nc.GetChannelId()] = myChannelBadge
						}
						myChannelBadge[resource.GetVideoCard().GetRid()] = &operate.ChannelBadge{
							Text:  resource.GetVideoCard().GetBadgeTitle(),
							Cover: resource.GetVideoCard().GetBadgeBackground(),
						}
					}

				}
			}
			aids = append(aids, tmpAids...)
		}
		if len(aids) == 0 {
			log.Error("square guanzhu gengxin update aids 0")
			return
		}
		var (
			amplayer map[int64]*archivegrpc.ArcPlayer
			isFav    map[int64]bool
			coins    map[int64]int64
		)
		// 广场页维度:我订阅的更新err不影响整体,err需要置为nil
		// 我订阅的更新维度:amplayer强依赖,err必须触发return
		g2, ctx2 := errgroup.WithContext(ctx)
		g2.Go(func() (err error) {
			var (
				aidsV2    []*archivegrpc.PlayAv
				isPlayurl bool
			)
			for _, aid := range aids {
				aidsV2 = append(aidsV2, &archivegrpc.PlayAv{Aid: aid})
				if aid != 0 {
					if (model.IsAndroid(plat) && build > s.c.BuildLimit.MiaokaiAndroid) || (model.IsIOS(plat) && build > s.c.BuildLimit.MiaokaiIOS) {
						isPlayurl = true
					}
				}
			}
			amplayer, err = s.Archives(ctx2, aidsV2, isPlayurl)
			if err != nil {
				log.Error("square guanzhu gengxin ArcsWithPlayurl error%v", err)
			}
			return
		})
		g2.Go(func() (err error) {
			if isFav, err = s.favDao.IsFavoreds(ctx2, mid, aids); err != nil {
				log.Error("square guanzhu gengxin IsFavoreds error%v", err)
				err = nil
			}
			return
		})
		if len(aids) > 0 && mid > 0 {
			g2.Go(func() (err error) {
				if coins, err = s.coinDao.IsCoins(ctx2, aids, mid); err != nil {
					log.Error("square guanzhu gengxin IsFavoreds error%v", err)
					err = nil
				}
				return
			})
		}
		if err = g2.Wait(); err != nil {
			err = nil
			return
		}
		var (
			cardNum           int
			channelInfocItems []*chmdl.InfocItem
			position          int64
		)
		for _, nc := range myChannels.GetCards() {
			if nc == nil {
				continue
			}
			cis, ok := cardItems[nc.ChannelId]
			if !ok || len(cis) == 0 {
				continue
			}
			h := cardm.Handle(plat, cdm.CardGt("channel_new"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
			if h == nil {
				continue
			}
			badges := myChannelBadges[nc.ChannelId]
			op := &operate.Card{
				ID:    nc.ChannelId,
				Plat:  plat,
				Title: nc.ChannelName,
				Cover: nc.Icon,
				Param: strconv.FormatInt(nc.ChannelId, 10),
				Channel: &operate.Channel{
					LastUpTime:     nc.GetLastUpdateTs(),
					UpCnt:          nc.GetUpdatedRsNum(),
					TodayCnt:       nc.GetTodayRsNum(),
					FeatureCnt:     nc.GetFeaturedCnt(),
					OfficiaLVerify: nc.GetVerify(),
					CType:          nc.GetCtype(),
					Badges:         badges,
					IsFav:          isFav,
					Coins:          coins,
					Position:       position,
				},
				Items:   cis,
				Build:   build,
				MobiApp: mobiApp,
			}
			_ = h.From(amplayer, op)
			// 频道卡不完整，totalUpChannelNum减1
			if h.Get() != nil && !h.Get().Right && op.Channel.UpCnt > 0 {
				totalUpChannelNum--
			}
			if h.Get() != nil && h.Get().Right {
				position = h.Get().Idx
				res.News.Items = append(res.News.Items, h)
				// infoc: archive card in channel card.
				var cii []*chmdl.InfocItem
				for _, ci := range op.Items {
					cardNum++
					ciii := &chmdl.InfocItem{
						OID:      ci.ID,
						CardType: "av_2r",
						Pos:      cardNum,
					}
					if b, ok := badges[ci.ID]; ok {
						ciii.Corner = b.Text
					}
					cii = append(cii, ciii)
				}
				// infoc: channel card.
				channelInfoc := &chmdl.InfocItem{
					ChannelID: nc.ChannelId,
					Items:     cii,
				}
				channelInfocItems = append(channelInfocItems, channelInfoc)
			}
		}
		if myChannels.GetDefaultFeed() {
			res.News.Label = "推荐频道"
		} else if totalUpChannelNum > 0 {
			res.News.Label = fmt.Sprintf("我订阅的%s个频道有更新", chmdl.MarkRed(strconv.Itoa(int(totalUpChannelNum))))
		} else {
			res.News.Label = "我订阅的频道"
		}
		// infoc: the whole square page.
		var ss = &infocStatistics{}
		if statistics != "" {
			if errTmp := json.Unmarshal([]byte(statistics), &ss); errTmp != nil {
				log.Warn("Failed to parse request statistics: %+v, err %+v", ss, errTmp)
			}
		}
		infoc := &chmdl.ChannelInfoc{
			EventId:     "cardshow",
			Page:        fromSpmid,
			Items:       channelInfocItems,
			CardNum:     cardNum,
			RequestUrl:  reqURL,
			TimeIso:     timeIso,
			Ip:          ip,
			AppId:       ss.AppID,
			Platform:    ss.Platform,
			Buvid:       buvid,
			Version:     ss.Version,
			VersionCode: strconv.Itoa(build),
			Mid:         strconv.FormatInt(mid, 10),
			Ctime:       strconv.FormatInt(ts*1000, 10),
			Abtest:      ss.ABTest,
			AutoRefresh: autoRefresh,
			Pos:         "我订阅的频道更新",
			CurRefresh:  pn,
		}
		s.infoc(infoc)
		return
	})
	if mid != 0 {
		// 频道历史,只有第一刷调用
		if offset == "" {
			g.Go(func() (err error) {
				if recents, err = s.chDao.Recent(ctx, mid); err != nil {
					log.Error("square history error %v", err)
					err = nil
					return
				}
				for _, recent := range recents {
					if recent == nil {
						continue
					}
					i := &chmdl.Channel{}
					i.FormChannelRecent(recent)
					res.Recent = append(res.Recent, i)
				}
				return
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
	}
	return
}

// nolint:gocognit
func (s Service) SquareAlpha(c context.Context, _, mobiApp, _, buvid string, plat int8, autoRefresh, build int,
	mid, ts, timeIso int64, fromSpmid, reqURL, statistics string) (res *chmdl.SquareResult, err error) {
	var (
		ip              = metadata.String(c, metadata.RemoteIP)
		scaneds         []*channelgrpc.ChannelFeedCard
		scanedCardItems = make(map[int64][]*operate.Card)
		scanedBadges    = make(map[int64]map[int64]*operate.ChannelBadge)
		rcmds           []*channelgrpc.ChannelFeedCard
		rcmdCardItems   = make(map[int64][]*operate.Card)
		rcmdBadges      = make(map[int64]map[int64]*operate.ChannelBadge)
		aids            []int64
	)
	g, ctx := errgroup.WithContext(c)
	if mid > 0 {
		// 最近浏览过的频道
		g.Go(func() (err error) {
			if scaneds, err = s.chDao.ScanedChannels(ctx, mid); err != nil {
				log.Error("%v", err)
				return nil
			}
			for _, scaned := range scaneds {
				if scaned == nil {
					log.Error("square alpha scaned card nil")
					continue
				}
				var tmpAids []int64
				for _, video := range scaned.GetVideoCards() {
					if video == nil || video.GetRid() == 0 {
						continue
					}
					tmpAids = append(tmpAids, video.GetRid())
					scanedCardItems[scaned.GetCid()] = append(scanedCardItems[scaned.GetCid()], &operate.Card{ID: video.GetRid()})
					if video.GetBadgeTitle() != "" && video.GetBadgeBackground() != "" {
						var (
							scanedBadge map[int64]*operate.ChannelBadge
							ok          bool
						)
						if scanedBadge, ok = scanedBadges[scaned.GetCid()]; !ok {
							scanedBadge = make(map[int64]*operate.ChannelBadge)
							scanedBadges[scaned.GetCid()] = scanedBadge
						}
						scanedBadge[video.GetRid()] = &operate.ChannelBadge{Text: video.GetBadgeTitle(), Cover: video.GetBadgeBackground()}
					}
				}
				aids = append(aids, tmpAids...)
			}
			return
		})
	}
	// 更多精彩频道
	g.Go(func() (err error) {
		if rcmds, err = s.chDao.Rcmd(c, mid); err != nil {
			log.Error("%v", err)
			return nil
		}
		for _, rcmd := range rcmds {
			if rcmd == nil {
				continue
			}
			for _, video := range rcmd.GetVideoCards() {
				if video == nil {
					continue
				}
				if video.GetRid() == 0 {
					continue
				}
				aids = append(aids, video.GetRid())
				rcmdCardItems[rcmd.GetCid()] = append(rcmdCardItems[rcmd.GetCid()], &operate.Card{ID: video.GetRid()})
				if video.GetBadgeTitle() != "" && video.GetBadgeBackground() != "" {
					var (
						rcmdBadge map[int64]*operate.ChannelBadge
						ok        bool
					)
					if rcmdBadge, ok = rcmdBadges[rcmd.GetCid()]; !ok {
						rcmdBadge = make(map[int64]*operate.ChannelBadge)
						rcmdBadges[rcmd.GetCid()] = rcmdBadge
					}
					rcmdBadge[video.GetRid()] = &operate.ChannelBadge{Text: video.GetBadgeTitle(), Cover: video.GetBadgeBackground()}
				}
			}
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	if len(aids) == 0 {
		log.Error("square alpha aids 0")
		return
	}
	var (
		amplayer map[int64]*archivegrpc.ArcPlayer
		isFav    map[int64]bool
		coins    map[int64]int64
	)
	// 获取所有物料信息
	// 整个页面维度:amplayer强依赖,err必须触发return
	g2, ctx2 := errgroup.WithContext(c)
	if len(aids) > 0 {
		g2.Go(func() (err error) {
			var aidsV2 []*archivegrpc.PlayAv
			for _, aid := range aids {
				if aid != 0 {
					aidsV2 = append(aidsV2, &archivegrpc.PlayAv{Aid: aid})
				}
			}
			if amplayer, err = s.Archives(ctx2, aidsV2, true); err != nil {
				log.Error("square alpha Archives error%v", err)
			}
			return err
		})
		g2.Go(func() (err error) {
			if isFav, err = s.favDao.IsFavoreds(ctx2, mid, aids); err != nil {
				log.Error("square alpha IsFavoreds error%v", err)
			}
			return nil
		})
		if mid > 0 {
			g2.Go(func() (err error) {
				if coins, err = s.coinDao.IsCoins(ctx2, aids, mid); err != nil {
					log.Error("square alpha IsCoins error%v", err)
					err = nil
				}
				return
			})
		}
	}
	if err = g2.Wait(); err != nil {
		err = nil
		return
	}
	res = &chmdl.SquareResult{}
	if len(scaneds) > 0 {
		res.Scaned = &chmdl.Scaned{
			Label: "最近浏览过的频道",
		}
		var (
			position          int64
			channelInfocItems []*chmdl.InfocItem
			cardNum           int
		)
		for _, scaned := range scaneds {
			if scaned == nil {
				continue
			}
			scis, ok := scanedCardItems[scaned.GetCid()]
			if !ok || len(scis) == 0 {
				continue
			}
			h := cardm.Handle(plat, cdm.CardGt("channel_scaned"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
			if h == nil {
				continue
			}
			badges := scanedBadges[scaned.GetCid()]
			op := &operate.Card{
				ID:    scaned.GetCid(),
				Plat:  plat,
				Title: scaned.GetCname(),
				Param: strconv.FormatInt(scaned.GetCid(), 10),
				Channel: &operate.Channel{
					LastUpTime: scaned.GetLastUpdateTs(),
					FeatureCnt: scaned.GetFeaturedCnt(),
					IsAtten:    scaned.GetSubscribed(),
					AttenCnt:   scaned.GetSubCnt(),
					TodayCnt:   scaned.GetTodayRsNum(),
					Badges:     badges,
					IsFav:      isFav,
					Coins:      coins,
					Position:   position,
				},
				Items:   scis,
				Build:   build,
				MobiApp: mobiApp,
			}
			_ = h.From(amplayer, op)
			if h.Get() != nil && h.Get().Right {
				position = h.Get().Idx
				res.Scaned.Items = append(res.Scaned.Items, h)
				// infoc: archive card in channel card.
				var cii []*chmdl.InfocItem
				for _, ci := range op.Items {
					cardNum++
					ciii := &chmdl.InfocItem{
						OID:      ci.ID,
						CardType: "av_2r",
						Pos:      cardNum,
					}
					if b, ok := badges[ci.ID]; ok {
						ciii.Corner = b.Text
					}
					cii = append(cii, ciii)
				}
				// infoc: channel card.
				channelInfoc := &chmdl.InfocItem{
					ChannelID: scaned.GetCid(),
					Items:     cii,
				}
				channelInfocItems = append(channelInfocItems, channelInfoc)
			}
		}
		var ss = &infocStatistics{}
		if statistics != "" {
			if errTmp := json.Unmarshal([]byte(statistics), &ss); errTmp != nil {
				log.Warn("Failed to parse request statistics: %+v, err %+v", ss, errTmp)
			}
		}
		infoc := &chmdl.ChannelInfoc{
			EventId:     "cardshow",
			Page:        fromSpmid,
			Items:       channelInfocItems,
			CardNum:     cardNum,
			RequestUrl:  reqURL,
			TimeIso:     timeIso,
			Ip:          ip,
			AppId:       ss.AppID,
			Platform:    ss.Platform,
			Buvid:       buvid,
			Version:     ss.Version,
			VersionCode: strconv.Itoa(build),
			Mid:         strconv.FormatInt(mid, 10),
			Ctime:       strconv.FormatInt(ts*1000, 10),
			Abtest:      ss.ABTest,
			AutoRefresh: autoRefresh,
			Pos:         "最近浏览过的频道",
		}
		s.infoc(infoc)
	}
	if len(rcmds) > 0 {
		res.Rcmd = &chmdl.Rcmd{
			Label: "更多精彩频道",
		}
		var (
			position          int64
			channelInfocItems []*chmdl.InfocItem
			cardNum           int
		)
		for _, rcmd := range rcmds {
			if rcmd == nil {
				continue
			}
			scis, ok := rcmdCardItems[rcmd.GetCid()]
			if !ok || len(scis) == 0 {
				continue
			}
			h := cardm.Handle(plat, cdm.CardGt("channel_rcmd_v2"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
			if h == nil {
				continue
			}
			badges := rcmdBadges[rcmd.GetCid()]
			op := &operate.Card{
				ID:    rcmd.GetCid(),
				Plat:  plat,
				Title: rcmd.GetCname(),
				Param: strconv.FormatInt(rcmd.GetCid(), 10),
				Channel: &operate.Channel{
					LastUpTime: rcmd.GetLastUpdateTs(),
					FeatureCnt: rcmd.GetFeaturedCnt(),
					IsAtten:    rcmd.GetSubscribed(),
					AttenCnt:   rcmd.GetSubCnt(),
					TodayCnt:   rcmd.GetTodayRsNum(),
					Badges:     badges,
					IsFav:      isFav,
					Coins:      coins,
					Position:   position,
				},
				Items:   scis,
				Build:   build,
				MobiApp: mobiApp,
			}
			_ = h.From(amplayer, op)
			if h.Get() != nil && h.Get().Right {
				position = h.Get().Idx
				res.Rcmd.Items = append(res.Rcmd.Items, h)
				// infoc: archive card in channel card.
				var cii []*chmdl.InfocItem
				for _, ci := range op.Items {
					cardNum++
					ciii := &chmdl.InfocItem{
						OID:      ci.ID,
						CardType: "av_2r",
						Pos:      cardNum,
					}
					if b, ok := badges[ci.ID]; ok {
						ciii.Corner = b.Text
					}
					cii = append(cii, ciii)
				}
				// infoc: channel card.
				channelInfoc := &chmdl.InfocItem{
					ChannelID: rcmd.GetCid(),
					Items:     cii,
				}
				channelInfocItems = append(channelInfocItems, channelInfoc)
			}
		}
		var ss = &infocStatistics{}
		if statistics != "" {
			if errTmp := json.Unmarshal([]byte(statistics), &ss); errTmp != nil {
				log.Warn("Failed to parse request statistics: %+v, err %+v", ss, errTmp)
			}
		}
		infoc := &chmdl.ChannelInfoc{
			EventId:     "cardshow",
			Page:        fromSpmid,
			Items:       channelInfocItems,
			CardNum:     cardNum,
			RequestUrl:  reqURL,
			TimeIso:     timeIso,
			Ip:          ip,
			AppId:       ss.AppID,
			Platform:    ss.Platform,
			Buvid:       buvid,
			Version:     ss.Version,
			VersionCode: strconv.Itoa(build),
			Mid:         strconv.FormatInt(mid, 10),
			Ctime:       strconv.FormatInt(ts*1000, 10),
			Abtest:      ss.ABTest,
			AutoRefresh: autoRefresh,
			Pos:         "更多精彩频道",
		}
		s.infoc(infoc)
	}
	return
}

func (s *Service) Detail(c context.Context, mid, channelID int64, plat int8, build int64, mobiApp, spmid, platform string, externalArg *chmdl.ChanelDetailExternalArgs) (res *chmdl.Detail, err error) {
	var (
		detail        *channelgrpc.ChannelCard
		selectSort    []*channelgrpc.FeaturedOption
		tabs          []*channelgrpc.ShowTab
		seasonm       map[int64]*appCardgrpc.SeasonCards
		defaultTabIdx int64
		labels        []*channelgrpc.ShowLabel
		ogvSwitch     bool
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		reply, err := s.chDao.Detail(ctx, mid, channelID, &channelmdl.MetaDataCtrl{Platform: platform, MobiApp: mobiApp, Build: build, Args: externalArg.Args})
		if err != nil {
			log.Error("%+v", err)
			return err
		}
		detail, selectSort, ogvSwitch, tabs, defaultTabIdx, labels = reply.GetChannel(), reply.GetFeaturedOptions(), reply.GetPGC(), reply.GetTabs(), reply.GetDefaultTabIdx(), reply.GetLabels()
		return
	})
	if s.c.Switch != nil && s.c.Switch.DetailVerify {
		g.Go(func() (err error) {
			if seasonm, err = s.pgcDao.TagOGV(ctx, []int64{channelID}); err != nil {
				log.Error("%v", err)
				err = nil
			}
			return
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	if detail == nil {
		log.Error("detail nil")
		return
	}
	res = &chmdl.Detail{}
	var menus = s.menuCache[detail.ChannelId]
	//版本判断
	var isHighBuild bool
	if pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsMobiAppIPhone().And().Build(">=", s.c.BuildLimit.OGVChanIOSBuild)
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", s.c.BuildLimit.OGVChanAndroidBuild)
	}).FinishOr(false) {
		isHighBuild = true
	}
	res.FormDetail(c, detail, menus, selectSort, seasonm, ogvSwitch, model.IsOverseas(plat), isHighBuild, s.c.PRLimit, build, defaultTabIdx, mobiApp, spmid, tabs, labels)
	return
}

func (s *Service) Multiple(c context.Context, channelID, mid, timeIso, ts int64, build int, plat int8,
	_, sort, offset, _, fromSpmid, buvid, reqURL, statistics, theme, from, mobiApp string, pn int) (res *chmdl.ChannelListResult, err error) {
	var (
		ip      = metadata.String(c, metadata.RemoteIP)
		cSort   channelgrpc.TotalSortType
		isBadge bool
	)
	switch sort {
	case "hot":
		cSort = channelgrpc.TotalSortType_SORT_BY_HOT
		isBadge = true
	case "view":
		cSort = channelgrpc.TotalSortType_SORT_BY_VIEW_CNT
	case "new":
		cSort = channelgrpc.TotalSortType_SORT_BY_PUB_TIME
	}
	var (
		cards *channelgrpc.ResourceListReply
		arg   = &channelgrpc.ResourceListReq{ChannelId: channelID, TabType: channelgrpc.TabType_TAB_TYPE_TOTAL, SortType: cSort, Offset: offset, PageSize: 20, Mid: mid}
	)
	if cards, err = s.chDao.ResourceList(c, arg); err != nil {
		log.Error("%v", err)
		return
	}
	res = &chmdl.ChannelListResult{Offset: cards.GetNextOffset()}
	if cards.GetHasMore() {
		res.HasMore = 1
	}
	if cards.UpdateHotCnt > 0 {
		res.Label = fmt.Sprintf("有%s", model.StatString(cards.UpdateHotCnt, "个视频更新"))
	}
	var hs []card.Handler
	if hs, err = s.dealItems(c, cards, channelID, mid, timeIso, ts, build, plat,
		ip, fromSpmid, buvid, sort, reqURL, statistics, 0, isBadge, theme, from, mobiApp, pn, false); err != nil {
		log.Error("%v", err)
		return
	}
	res.Items = hs
	return
}

func (s *Service) Selected(c context.Context, channelID, mid, timeIso, ts int64, cFilter int32, build int, plat int8,
	_, offset, _, fromSpmid, buvid, reqURL, statistics, theme, from, mobiApp string, pn int) (res *chmdl.ChannelListResult, err error) {
	var (
		ip    = metadata.String(c, metadata.RemoteIP)
		cards *channelgrpc.ResourceListReply
		arg   = &channelgrpc.ResourceListReq{ChannelId: channelID, TabType: channelgrpc.TabType_TAB_TYPE_FEATURED, FilterType: cFilter, Offset: offset, PageSize: 20, Mid: mid}
	)
	if cards, err = s.chDao.ResourceList(c, arg); err != nil {
		log.Error("%v", err)
		return
	}
	res = &chmdl.ChannelListResult{Offset: cards.GetNextOffset()}
	if cards.GetHasMore() {
		res.HasMore = 1
	}
	if cards.UpdateHotCnt > 0 {
		res.Label = fmt.Sprintf("有%s", model.StatString(cards.UpdateHotCnt, "个视频更新"))
	}
	var (
		hs    []card.Handler
		isOGV bool
	)
	if (offset == "" || offset == "0") && cFilter == 0 {
		isOGV = true
	}
	if hs, err = s.dealItems(c, cards, channelID, mid, timeIso, ts, build, plat,
		ip, fromSpmid, buvid, "", reqURL, statistics, cFilter, true, theme, from, mobiApp, pn, isOGV); err != nil {
		log.Error("%v", err)
		return
	}
	res.Items = hs
	return
}

// nolint:gocognit
func (s *Service) dealItems(c context.Context, cards *channelgrpc.ResourceListReply, channelID, mid, timeIso, ts int64, build int,
	plat int8, ip, fromSpmid, buvid, cSort, reqURL, statistics string, cFilter int32,
	isBadge bool, theme, from, mobiApp string, pn int, isOGV bool) (is []card.Handler, err error) {
	var (
		cs = cards.GetCards()
		// aids为全量视频卡；oids是包含双列视频卡
		aids, oids   []int64
		videoBadges  = make(map[int64]*operate.ChannelBadge)
		customBadges = make(map[int64]*operate.ChannelBadge)
	)
	for _, cardTmp := range cs {
		switch cardTmp.GetCardType() {
		case chmdl.CardTypeVideo:
			if cardTmp.GetVideoCard() != nil && cardTmp.GetVideoCard().Rid != 0 {
				aids = append(aids, cardTmp.GetVideoCard().Rid)
				oids = append(oids, cardTmp.GetVideoCard().Rid)
				if isBadge && cardTmp.GetVideoCard().GetBadgeTitle() != "" && cardTmp.GetVideoCard().GetBadgeBackground() != "" {
					videoBadges[cardTmp.GetVideoCard().Rid] = &operate.ChannelBadge{
						Text:  cardTmp.GetVideoCard().GetBadgeTitle(),
						Cover: cardTmp.GetVideoCard().GetBadgeBackground(),
					}
				}
			}
		case chmdl.CardTypeCustom:
			if customs := cardTmp.GetCustomCard(); customs != nil {
				for _, cardTmp := range customs.GetCards() {
					if cardTmp != nil && cardTmp.GetRid() != 0 {
						aids = append(aids, cardTmp.GetRid())
						if isBadge && cardTmp.GetBadgeTitle() != "" && cardTmp.GetBadgeBackground() != "" {
							customBadges[cardTmp.GetRid()] = &operate.ChannelBadge{
								Text:  cardTmp.GetBadgeTitle(),
								Cover: cardTmp.GetBadgeBackground(),
							}
						}
					}
				}
			}
		case chmdl.CardTypeRank:
			if ranks := cardTmp.GetRankCard(); ranks != nil {
				for _, rank := range ranks.GetCards() {
					if rank != nil && rank.Rid != 0 {
						aids = append(aids, rank.GetRid())
					}
				}
			}
		default:
			log.Warn("channel dealItem unknown type %+v", cardTmp)
		}
	}
	var (
		arcs    map[int64]*archivegrpc.ArcPlayer
		isFav   map[int64]bool
		coins   map[int64]int64
		seasonm map[int64]*appCardgrpc.SeasonCards
	)
	g, ctx := errgroup.WithContext(c)
	if len(aids) > 0 {
		var (
			aidsV2    []*archivegrpc.PlayAv
			isPlayurl bool
		)
		for _, aid := range aids {
			if aid != 0 {
				aidsV2 = append(aidsV2, &archivegrpc.PlayAv{Aid: aid})
				if (model.IsAndroid(plat) && build > s.c.BuildLimit.MiaokaiAndroid) || (model.IsIOS(plat) && build > s.c.BuildLimit.MiaokaiIOS) ||
					(plat == model.PlatAndroidI && build > s.c.BuildLimit.ArcWithPlayerAndroid) || (plat == model.PlatIPhoneI && build > s.c.BuildLimit.ArcWithPlayerIOS) {
					isPlayurl = true
				}
			}
		}
		g.Go(func() (err error) {
			arcs, err = s.Archives(ctx, aidsV2, isPlayurl)
			if err != nil {
				log.Error("%v", err)
				err = nil
			}
			return
		})
		if len(oids) > 0 {
			if mid > 0 {
				g.Go(func() (err error) {
					if coins, err = s.coinDao.IsCoins(ctx, oids, mid); err != nil {
						log.Error("%v", err)
						err = nil
					}
					return
				})
			}
			g.Go(func() (err error) {
				if isFav, err = s.favDao.IsFavoreds(ctx, mid, aids); err != nil {
					log.Error("%v", err)
					err = nil
				}
				return
			})
		}
		if s.c.Switch != nil && s.c.Switch.DetailVerify && isOGV && cards.GetPGC() {
			g.Go(func() (err error) {
				if seasonm, err = s.pgcDao.TagOGV(ctx, []int64{channelID}); err != nil {
					log.Error("%v", err)
				}
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		cardCount, cardTotal int
		infocItems           []*chmdl.InfocItem
		position             int64
	)
	is = make([]card.Handler, 0, len(cs))
	for _, season := range seasonm {
		if season == nil || len(season.GetCards()) == 0 {
			continue
		}
		var (
			gt    cdm.CardGt
			ctype cdm.CardType
		)
		if len(season.GetCards()) == 1 {
			gt = cdm.CardGt("channel_ogv")
			ctype = cdm.CardType("channel_ogv")
		} else {
			gt = cdm.CardGt("channel_ogv_large")
			ctype = cdm.CardType("channel_ogv_large")
		}
		op := &operate.Card{
			ID:      channelID,
			Plat:    plat,
			Build:   build,
			Param:   strconv.FormatInt(channelID, 10),
			Channel: &operate.Channel{Position: position, Sort: cSort, Filt: cFilter},
			MobiApp: mobiApp,
		}
		if s.c.Switch != nil {
			if s.c.Switch.ListOGVFold && len(season.GetCards()) > 3 {
				op.Channel.HasFold = true
			}
			op.Channel.HasMore = s.c.Switch.ListOGVMore
		}
		op.From(gt, channelID, 0, plat, build, mobiApp)
		h := cardm.Handle(plat, gt, ctype, cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
		if h == nil {
			continue
		}
		_ = h.From(season, op)
		if h.Get() == nil || !h.Get().Right {
			continue
		}
		position = h.Get().Idx
		// infoc: archive card in channel card.
		var i []*chmdl.InfocItem
		for _, season := range season.GetCards() {
			ii := &chmdl.InfocItem{
				OID:      int64(season.GetSeasonId()),
				CardType: "pgc",
			}
			i = append(i, ii)
		}
		is, infocItems, cardTotal = s.dealAppend(is, h, infocItems, i, cardTotal)
	}
	for _, card := range cs {
		// nolint:exhaustive
		switch card.GetCardType() {
		case chmdl.CardTypeVideo:
			if card.GetVideoCard() == nil || card.GetVideoCard().Rid == 0 {
				continue
			}
			h := cardm.Handle(plat, cdm.CardGt("channel_new_detail"), "", cdm.ColumnSvrDouble, nil, nil, nil, nil, nil, nil, nil)
			if h == nil {
				continue
			}
			op := &operate.Card{
				Channel: &operate.Channel{Badges: videoBadges, IsFav: isFav, Coins: coins, Position: position, Sort: cSort, Filt: cFilter},
			}
			op.From(cdm.CardGt("channel_new_detail"), card.GetVideoCard().Rid, 0, plat, build, mobiApp)
			_ = h.From(arcs, op)
			if h.Get() == nil || !h.Get().Right {
				continue
			}
			position = h.Get().Idx
			// infoc: archive card in detail.
			var i []*chmdl.InfocItem
			ii := &chmdl.InfocItem{
				OID:      op.ID,
				CardType: "av_2r",
			}
			if videoBadges != nil {
				if vb, ok := videoBadges[op.ID]; ok {
					ii.Corner = vb.Text
				}
			}
			i = append(i, ii)
			is, infocItems, cardTotal = s.dealAppend(is, h, infocItems, i, cardTotal)
		case chmdl.CardTypeCustom:
			var (
				customs = card.GetCustomCard()
				cis     []*operate.Card
			)
			if customs == nil {
				continue
			}
			for _, card := range customs.GetCards() {
				if card == nil || card.GetRid() == 0 {
					continue
				}
				var (
					arc *archivegrpc.ArcPlayer
					ok  bool
				)
				if arc, ok = arcs[card.GetRid()]; !ok || arc == nil {
					continue
				}
				cis = append(cis, &operate.Card{ID: card.GetRid()})
			}
			if len(cis) == 0 {
				continue
			}
			op := &operate.Card{
				ID:      channelID,
				Plat:    plat,
				Build:   build,
				Param:   strconv.FormatInt(channelID, 10),
				Items:   cis,
				Channel: &operate.Channel{Position: position, Badges: customBadges, Sort: cSort, Filt: cFilter},
				MobiApp: mobiApp,
			}
			if customs.GetDetail() != nil {
				op.Title = customs.GetDetail().GetName()
				op.Channel.CustomDesc = customs.GetDetail().GetJumpDesc()
				op.Channel.CustomURI = customs.GetDetail().GetJumpUrl()
			}
			op.From(cdm.CardGt("channel_new_detail_custom"), channelID, 0, plat, build, mobiApp)
			h := cardm.Handle(plat, cdm.CardGt("channel_new_detail_custom"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
			if h == nil {
				continue
			}
			_ = h.From(arcs, op)
			if h.Get() == nil || !h.Get().Right {
				continue
			}
			position = h.Get().Idx
			// infoc: archive card in channel card.
			var i []*chmdl.InfocItem
			for _, ci := range op.Items {
				ii := &chmdl.InfocItem{
					OID:      ci.ID,
					CardType: "av_3r",
				}
				if customBadges != nil {
					if cb, ok := customBadges[ci.ID]; ok {
						ii.Corner = cb.Text
					}
				}
				i = append(i, ii)
			}
			is, infocItems, cardTotal = s.dealAppend(is, h, infocItems, i, cardTotal)
		case chmdl.CardTypeRank:
			var (
				ranks = card.GetRankCard()
				cis   []*operate.Card
			)
			if ranks == nil {
				continue
			}
			for _, rank := range ranks.GetCards() {
				if rank == nil || rank.GetRid() == 0 {
					continue
				}
				var (
					arc *archivegrpc.ArcPlayer
					ok  bool
				)
				if arc, ok = arcs[rank.GetRid()]; !ok || arc == nil {
					continue
				}
				cis = append(cis, &operate.Card{ID: rank.GetRid()})
			}
			if len(cis) == 0 {
				continue
			}
			op := &operate.Card{
				ID:      channelID,
				Plat:    plat,
				Build:   build,
				Param:   strconv.FormatInt(channelID, 10),
				Items:   cis,
				Channel: &operate.Channel{Position: position, Sort: cSort, Filt: cFilter},
				MobiApp: mobiApp,
			}
			if ranks.GetDetail() != nil {
				op.Title = ranks.GetDetail().GetTitle()
				op.Channel.CustomDesc = ranks.GetDetail().GetJumpDesc()
				op.Channel.CustomURI = fmt.Sprintf(chmdl.RankURL, channelID, url.QueryEscape(theme))
				op.Channel.RankType = ranks.Detail.SortType
			}
			op.From(cdm.CardGt("channel_new_detail_rank"), channelID, 0, plat, build, mobiApp)
			h := cardm.Handle(plat, cdm.CardGt("channel_new_detail_rank"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
			if h == nil {
				continue
			}
			_ = h.From(arcs, op)
			if h.Get() == nil || !h.Get().Right {
				continue
			}
			position = h.Get().Idx
			// infoc: archive card in channel card.
			var i []*chmdl.InfocItem
			for _, ci := range op.Items {
				ii := &chmdl.InfocItem{
					OID:      ci.ID,
					CardType: "rank_3r",
				}
				i = append(i, ii)
			}
			is, infocItems, cardTotal = s.dealAppend(is, h, infocItems, i, cardTotal)
		}
	}
	// 双列末尾卡片去空窗(暂时不去末尾空窗,因为有些频道视频少,去空窗后与页面上的"x个视频"文案有出入
	// is = is[:len(is)-cardTotal%2]
	// infocItems = infocItems[:len(infocItems)-cardTotal%2]
	// infoc: channel detail page.
	for _, infocItem := range infocItems {
		cardCount++
		infocItem.Pos = cardCount
	}
	var channelInfocItems []*chmdl.InfocItem
	channelInfocItems = append(channelInfocItems, &chmdl.InfocItem{
		ChannelID: channelID,
		Items:     infocItems,
	})
	// infoc: the whole detail page.
	var ss = &infocStatistics{}
	if statistics != "" {
		if errTmp := json.Unmarshal([]byte(statistics), &ss); errTmp != nil {
			log.Warn("Failed to parse request statistics: %+v, err %+v", ss, errTmp)
		}
	}
	infoc := &chmdl.ChannelInfoc{
		EventId:     "cardshow",
		Page:        fromSpmid,
		Sort:        cSort,
		Filt:        strconv.Itoa(int(cFilter)),
		Items:       channelInfocItems,
		CardNum:     cardCount,
		RequestUrl:  reqURL,
		TimeIso:     timeIso,
		Ip:          ip,
		AppId:       ss.AppID,
		Platform:    ss.Platform,
		Buvid:       buvid,
		Version:     ss.Version,
		VersionCode: strconv.Itoa(build),
		Mid:         strconv.FormatInt(mid, 10),
		Ctime:       strconv.FormatInt(ts*1000, 10),
		Abtest:      ss.ABTest,
		From:        from,
		CurRefresh:  pn,
	}
	s.infoc(infoc)
	return
}

func (s *Service) dealAppend(rs []card.Handler, h card.Handler, infocs []*chmdl.InfocItem, infoc []*chmdl.InfocItem, cardTotal int) (is []card.Handler, infocs2 []*chmdl.InfocItem, total int) {
	if h.Get().CardLen == 0 {
		if cardTotal%2 == 1 {
			is = card.SwapTwoItem(rs, h)
			infocs2 = append(infocs[:len(infocs)-1], append(infoc, infocs[len(infocs)-1])...)
		} else {
			is = append(rs, h)
			infocs2 = append(infocs, infoc...)
		}
	} else {
		is = append(rs, h)
		infocs2 = append(infocs, infoc...)
	}
	total = cardTotal + h.Get().CardLen
	return
}

func (s *Service) RankList(c context.Context, id int64, offset string, ps int) (res *chmdl.RankResult, err error) {
	var rankCards *channelgrpc.RankCardReply
	if rankCards, err = s.chDao.RankCards(c, id, offset, int32(ps)); err != nil {
		log.Error("%v", err)
		return
	}
	if rankCards == nil {
		return
	}
	var (
		aids []int64
		arcm map[int64]*archivegrpc.ArcPlayer
	)
	for _, v := range rankCards.GetCards() {
		if v == nil || v.Rid == 0 {
			continue
		}
		aids = append(aids, v.GetRid())
	}
	if len(aids) == 0 {
		log.Warn("RankList id(%d) offset(%s) get empty aids", id, offset)
		return
	}
	var aidsV2 []*archivegrpc.PlayAv
	for _, aid := range aids {
		aidsV2 = append(aidsV2, &archivegrpc.PlayAv{Aid: aid})
	}
	if arcm, err = s.Archives(c, aidsV2, false); err != nil {
		log.Error("%v", err)
		return
	}
	res = &chmdl.RankResult{
		Offset: rankCards.GetNextOffset(),
	}
	if rankCards.GetHasMore() {
		res.HasMore = 1
	}
	if len(rankCards.GetCards()) == 0 {
		log.Warn("RankList id(%d) offset(%s) get empty cards", id, offset)
		return
	}
	var rankType int32
	if rankCards.GetDetail() != nil {
		res.Title = rankCards.GetDetail().GetTitle()
		res.Label = fmt.Sprintf("近%d日内投稿，每%d分钟更新一次", rankCards.GetDetail().GetPubRange(), rankCards.GetDetail().GetUpdateTime())
		rankType = rankCards.GetDetail().SortType
	}
	dev, _ := device.FromContext(c)
	for _, aid := range aids {
		if ap, ok := arcm[aid]; ok && ap != nil && ap.Arc != nil {
			if cardm.CheckMidMaxInt32(ap.GetArc().GetAuthor().Mid) && cardm.CheckMidMaxInt32Version(dev) {
				log.Warn("Filtered by mid int64: %d", ap.GetArc().GetAuthor().Mid)
				continue
			}
			var (
				labelText string
				arc       = ap.Arc
			)
			// nolint:gomnd
			switch rankType {
			case 1:
				labelText = model.StatString(arc.Stat.View, " 播放")
			case 4:
				labelText = model.StatString(arc.Stat.Fav, " 收藏")
			case 5:
				labelText = model.StatString(arc.Stat.Coin, " 投币")
			}
			i := &chmdl.Rank{
				ID:       aid,
				Title:    arc.Title,
				Cover:    arc.Pic,
				Param:    strconv.FormatInt(aid, 10),
				Goto:     model.GotoAv,
				Label:    labelText,
				Duration: arc.Duration,
				Author:   arc.Author,
			}
			res.Items = append(res.Items, i)
		}
	}
	return
}

func (s *Service) Share(c context.Context, id, mid int64) (res *chmdl.Share, err error) {
	if s.c.Share == nil || s.c.Share.Items == nil {
		return
	}
	reply, err := s.chDao.Detail(c, mid, id, nil)
	if err != nil {
		log.Error("%v", err)
		return
	}
	detail := reply.Channel
	if detail == nil {
		log.Error("detail nil")
		return
	}
	res = &chmdl.Share{
		Share: &chmdl.ShareItem{},
	}
	if s.c.Share.Items.Weibo {
		res.Share.Weibo = true
	}
	if s.c.Share.Items.Wechat {
		res.Share.Wechart = true
	}
	if s.c.Share.Items.WechatMonment {
		res.Share.WechartMonment = true
	}
	if s.c.Share.Items.QQ {
		res.Share.QQ = true
	}
	if s.c.Share.Items.QZone {
		res.Share.QZone = true
	}
	if s.c.Share.Items.Copy {
		res.Share.Copy = true
	}
	if s.c.Share.Items.More {
		res.Share.More = true
	}
	res.ID = id
	res.ShareURI = s.c.Share.JumpURI + strconv.FormatInt(detail.GetChannelId(), 10)
	var (
		title string
		descs []string
	)
	title = fmt.Sprintf("共%d个视频", detail.GetRCnt())
	if detail.GetFeaturedCnt() > 0 {
		title = fmt.Sprintf("为你精选了%d个视频", detail.GetFeaturedCnt())
		descs = append(descs, fmt.Sprintf("共%v", model.StatString(detail.GetRCnt(), "视频")))
	}
	descs = append(descs, model.StatString(detail.GetSubscribedCnt(), "人订阅"))
	res.Title = fmt.Sprintf("频道 | %s：%s", detail.GetChannelName(), title)
	res.Desc = fmt.Sprintf("bilibili『%s频道』%s", detail.GetChannelName(), strings.Join(descs, "，"))
	res.Icon = detail.Icon
	res.Param = strconv.FormatInt(id, 10)
	res.ChannelURI = constructChannelURI(res.Param, reply)
	return
}

func constructChannelURI(param string, reply *channelgrpc.ChannelDetailReply) string {
	if len(reply.Tabs) > int(reply.DefaultTabIdx) && reply.Tabs[reply.DefaultTabIdx] != nil {
		return reply.Tabs[reply.DefaultTabIdx].Url
	}
	return model.FillURI(model.GotoChannelNew, param, 0, 0, 0, model.ChannelHandler("tab=all&sort=hot"))
}

func (s *Service) Red(c context.Context, mid int64) (res *chmdl.Red, err error) {
	var resTmp bool
	if resTmp, err = s.chDao.RedPoint2(c, mid); err != nil {
		log.Error("%v", err)
	}
	if resTmp {
		res = &chmdl.Red{Type: chmdl.RedTypePoint}
	}
	return
}

// nolint:gocognit
func (s *Service) Rcmd(c context.Context, _, mobiApp, _, buvid string, plat int8, autoRefresh, build int,
	mid, ts, timeIso int64, fromSpmid, reqURL, statistics string) (res *chmdl.Rcmd, err error) {
	var (
		ip            = metadata.String(c, metadata.RemoteIP)
		rcmds         []*channelgrpc.ChannelFeedCard
		rcmdCardItems = make(map[int64][]*operate.Card)
		rcmdBadges    = make(map[int64]map[int64]*operate.ChannelBadge)
		aids          []int64
	)
	if rcmds, err = s.chDao.Rcmd(c, mid); err != nil {
		log.Error("%v", err)
		return
	}
	for _, rcmd := range rcmds {
		if rcmd == nil {
			continue
		}
		for _, video := range rcmd.GetVideoCards() {
			if video == nil {
				continue
			}
			if video.GetRid() == 0 {
				continue
			}
			aids = append(aids, video.GetRid())
			rcmdCardItems[rcmd.GetCid()] = append(rcmdCardItems[rcmd.GetCid()], &operate.Card{ID: video.GetRid()})
			if video.GetBadgeTitle() != "" && video.GetBadgeBackground() != "" {
				var (
					rcmdBadge map[int64]*operate.ChannelBadge
					ok        bool
				)
				if rcmdBadge, ok = rcmdBadges[rcmd.GetCid()]; !ok {
					rcmdBadge = make(map[int64]*operate.ChannelBadge)
					rcmdBadges[rcmd.GetCid()] = rcmdBadge
				}
				rcmdBadge[video.GetRid()] = &operate.ChannelBadge{Text: video.GetBadgeTitle(), Cover: video.GetBadgeBackground()}
			}
		}
	}
	if len(aids) == 0 {
		log.Error("rcmd get aids 0")
		return
	}
	var (
		amplayer map[int64]*archivegrpc.ArcPlayer
		isFav    map[int64]bool
		coins    map[int64]int64
	)
	// 获取所有物料信息
	// 整个页面维度:amplayer强依赖,err必须触发return
	g, ctx := errgroup.WithContext(c)
	if len(aids) > 0 {
		g.Go(func() (err error) {
			var aidsV2 []*archivegrpc.PlayAv
			for _, aid := range aids {
				aidsV2 = append(aidsV2, &archivegrpc.PlayAv{Aid: aid})
			}
			if amplayer, err = s.Archives(ctx, aidsV2, true); err != nil {
				log.Error("rcmd ArcsPlayer error%v", err)
			}
			return
		})
		g.Go(func() (err error) {
			if isFav, err = s.favDao.IsFavoreds(ctx, mid, aids); err != nil {
				log.Error("rcmd IsFavoreds error%v", err)
			}
			return nil
		})
		if mid > 0 {
			g.Go(func() (err error) {
				if coins, err = s.coinDao.IsCoins(ctx, aids, mid); err != nil {
					log.Error("rcmd IsCoins error%v", err)
				}
				return nil
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	res = &chmdl.Rcmd{
		Label: "更多精彩频道",
	}
	var (
		position          int64
		channelInfocItems []*chmdl.InfocItem
		cardNum           int
	)
	for _, rcmd := range rcmds {
		if rcmd == nil {
			continue
		}
		scis, ok := rcmdCardItems[rcmd.GetCid()]
		if !ok || len(scis) == 0 {
			continue
		}
		h := cardm.Handle(plat, cdm.CardGt("channel_rcmd_v2"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
		if h == nil {
			continue
		}
		badges := rcmdBadges[rcmd.GetCid()]
		op := &operate.Card{
			ID:    rcmd.GetCid(),
			Plat:  plat,
			Title: rcmd.GetCname(),
			Param: strconv.FormatInt(rcmd.GetCid(), 10),
			Channel: &operate.Channel{
				LastUpTime: rcmd.GetLastUpdateTs(),
				FeatureCnt: rcmd.GetFeaturedCnt(),
				IsAtten:    rcmd.GetSubscribed(),
				AttenCnt:   rcmd.GetSubCnt(),
				TodayCnt:   rcmd.GetTodayRsNum(),
				Badges:     badges,
				IsFav:      isFav,
				Coins:      coins,
				Position:   position,
			},
			Items:   scis,
			Build:   build,
			MobiApp: mobiApp,
		}
		_ = h.From(amplayer, op)
		if h.Get() != nil && h.Get().Right {
			position = h.Get().Idx
			res.Items = append(res.Items, h)
			// infoc: archive card in channel card.
			var cii []*chmdl.InfocItem
			for _, ci := range op.Items {
				cardNum++
				ciii := &chmdl.InfocItem{
					OID:      ci.ID,
					CardType: "av_2r",
					Pos:      cardNum,
				}
				if b, ok := badges[ci.ID]; ok {
					ciii.Corner = b.Text
				}
				cii = append(cii, ciii)
			}
			// infoc: channel card.
			channelInfoc := &chmdl.InfocItem{
				ChannelID: rcmd.GetCid(),
				Items:     cii,
			}
			channelInfocItems = append(channelInfocItems, channelInfoc)
		}
	}
	var ss = &infocStatistics{}
	if statistics != "" {
		if errTmp := json.Unmarshal([]byte(statistics), &ss); errTmp != nil {
			log.Warn("Failed to parse request statistics: %+v, err %v", ss, errTmp)
		}
	}
	infoc := &chmdl.ChannelInfoc{
		EventId:     "cardshow",
		Page:        fromSpmid,
		Items:       channelInfocItems,
		CardNum:     cardNum,
		RequestUrl:  reqURL,
		TimeIso:     timeIso,
		Ip:          ip,
		AppId:       ss.AppID,
		Platform:    ss.Platform,
		Buvid:       buvid,
		Version:     ss.Version,
		VersionCode: strconv.Itoa(build),
		Mid:         strconv.FormatInt(mid, 10),
		Ctime:       strconv.FormatInt(ts*1000, 10),
		Abtest:      ss.ABTest,
		AutoRefresh: autoRefresh,
		Pos:         "更多精彩频道",
	}
	s.infoc(infoc)
	return
}

func (s *Service) RegionList(c context.Context, svr *channelSvr.Service, params *chmdl.Param) (res []*channel.Region, err error) {
	if _, _, res, err = svr.RegionList(c, params.Build, params.TeenagersMode, params.MobiApp, params.Device, params.Lang, params.Channel); err != nil {
		log.Error("%v", err)
		return
	}
	if i18n.PreferTraditionalChinese(c, params.SLocal, params.CLocal) {
		for _, r := range res {
			i18n.TranslateAsTCV2(&r.Name)
		}
	}
	return
}

// nolint:gocognit
func (s *Service) Square3(c context.Context, params *chmdl.Param) (res []*chmdl.SquareItem, err error) {
	var (
		myChannels   *channelgrpc.MyChannelsReply
		channelList  *channelgrpc.ChannelListReply
		viewChannels []*channelgrpc.ViewChannelCard
		hotChannels  *channelgrpc.HotChannelReply
		plat         = model.Plat(params.MobiApp, params.Device)
		ip           = metadata.String(c, metadata.RemoteIP)
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if channelList, err = s.chDao.ChannelList(ctx, params.MID, 0, ""); err != nil {
			log.Error("%v", err)
			return nil
		}
		return
	})
	g.Go(func() (err error) {
		if myChannels, err = s.chDao.MyChannels2(ctx, params.MID, params.OffsetNew, true, chmdl.NewSubVersion); err != nil {
			log.Error("%v", err)
		}
		return
	})
	if params.MID > 0 {
		g.Go(func() (err error) {
			if viewChannels, err = s.chDao.ViewChannel(ctx, params.MID); err != nil {
				log.Error("%v", err)
				return nil
			}
			return
		})
	}
	g.Go(func() (err error) {
		if hotChannels, err = s.chDao.HotChannel(ctx, params.MID, params.OffsetRcmd); err != nil {
			log.Error("%v", err)
			return nil
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		aids                []int64
		myChannelCardItems  = make(map[int64][]*operate.Card)
		myChannelBadges     = make(map[int64]map[int64]*operate.ChannelBadge)
		hotChannelBadges    = make(map[int64]map[int64]*operate.ChannelBadge)
		hotChannelCardItems = make(map[int64][]*operate.Card)
		actTids             []int64
	)
	if myChannels != nil {
		for _, nc := range myChannels.GetCards() {
			if nc == nil {
				continue
			}
			var tmpAids []int64
			for _, resource := range nc.GetResourceCards() {
				if resource.GetVideoCard() != nil && resource.GetVideoCard().GetRid() != 0 {
					tmpAids = append(tmpAids, resource.GetVideoCard().GetRid())
					ci := &operate.Card{
						ID: resource.GetVideoCard().GetRid(),
					}
					myChannelCardItems[nc.GetChannelId()] = append(myChannelCardItems[nc.GetChannelId()], ci)
					if resource.GetVideoCard().GetBadgeTitle() != "" && resource.GetVideoCard().GetBadgeBackground() != "" {
						var (
							myChannelbadge map[int64]*operate.ChannelBadge
							ok             bool
						)
						if myChannelbadge, ok = myChannelBadges[nc.GetChannelId()]; !ok {
							myChannelbadge = make(map[int64]*operate.ChannelBadge)
							myChannelBadges[nc.GetChannelId()] = myChannelbadge
						}
						myChannelbadge[resource.GetVideoCard().GetRid()] = &operate.ChannelBadge{
							Text:  resource.GetVideoCard().GetBadgeTitle(),
							Cover: resource.GetVideoCard().GetBadgeBackground(),
						}
					}
				}
			}
			aids = append(aids, tmpAids...)
		}
		for _, dynamic := range myChannels.GetDynamicList() {
			if dynamic.GetRid() == 0 {
				continue
			}
			aids = append(aids, dynamic.GetRid())
		}
		for _, top := range myChannels.GetTops() {
			if top.GetActAttr() == chmdl.ActiveTag {
				actTids = append(actTids, top.GetChannelId())
			}
		}
		for _, normals := range myChannels.GetNormals() {
			if normals.GetActAttr() == chmdl.ActiveTag {
				actTids = append(actTids, normals.GetChannelId())
			}
		}
	}
	if hotChannels != nil {
		for _, hc := range hotChannels.GetCard() {
			if hc == nil {
				continue
			}
			var tmpAids []int64
			for _, video := range hc.GetVideoCards() {
				if video.GetRid() == 0 {
					continue
				}
				tmpAids = append(tmpAids, video.GetRid())
				ci := &operate.Card{
					ID: video.GetRid(),
				}
				hotChannelCardItems[hc.Cid] = append(hotChannelCardItems[hc.Cid], ci)
				if video.GetBadgeTitle() != "" && video.GetBadgeBackground() != "" {
					var (
						hotChannelbadge map[int64]*operate.ChannelBadge
						ok              bool
					)
					if hotChannelbadge, ok = hotChannelBadges[hc.Cid]; !ok {
						hotChannelbadge = make(map[int64]*operate.ChannelBadge)
						hotChannelBadges[hc.Cid] = hotChannelbadge
					}
					hotChannelbadge[video.GetRid()] = &operate.ChannelBadge{
						Text:  video.BadgeTitle,
						Cover: video.BadgeBackground,
					}
				}
			}
			aids = append(aids, tmpAids...)
		}
	}
	var (
		amplayer map[int64]*archivegrpc.ArcPlayer
		isFav    map[int64]bool
		coins    map[int64]int64
		actInfos map[int64]*natgrpc.NativePage
	)
	g2, ctx2 := errgroup.WithContext(c)
	if len(aids) > 0 {
		var (
			aidsV2    []*archivegrpc.PlayAv
			isPlayurl bool
		)
		for _, aid := range aids {
			aidsV2 = append(aidsV2, &archivegrpc.PlayAv{Aid: aid})
			if (model.IsAndroid(plat) && params.Build > s.c.BuildLimit.MiaokaiAndroid) || (model.IsIOS(plat) && params.Build > s.c.BuildLimit.MiaokaiIOS) || (plat == model.PlatAndroidI && params.Build > s.c.BuildLimit.ArcWithPlayerAndroid) || (plat == model.PlatIPhoneI && params.Build > s.c.BuildLimit.ArcWithPlayerIOS) {
				isPlayurl = true
			}
		}
		g2.Go(func() (err error) {
			amplayer, err = s.Archives(ctx2, aidsV2, isPlayurl)
			if err != nil {
				log.Error("%v", err)
			}
			return
		})
		if params.MID > 0 {
			g2.Go(func() (err error) {
				if isFav, err = s.favDao.IsFavoreds(ctx2, params.MID, aids); err != nil {
					log.Error("%v", err)
					return nil
				}
				return
			})
			g2.Go(func() (err error) {
				if coins, err = s.coinDao.IsCoins(ctx2, aids, params.MID); err != nil {
					log.Error("%v", err)
					return nil
				}
				return
			})
		}
	}
	if s.c.Switch.SquareActive && len(actTids) > 0 {
		g2.Go(func() (err error) {
			if actInfos, err = s.natDao.NatInfoFromForeigns(c, actTids, 1); err != nil {
				log.Error("%v", err)
				return nil
			}
			return
		})
	}
	if err = g2.Wait(); err != nil {
		return
	}
	var isBreak bool
	if params.OffsetNew != "" {
		isBreak = true
	}
	for _, mt := range s.c.Square.Models {
		var re *chmdl.SquareItem
		switch mt {
		case chmdl.ModelTypeSearch: // 搜索模块
			if !isBreak {
				re = &chmdl.SquareItem{ModelType: chmdl.ModelTypeSearch, ModelTitle: chmdl.ModelNameSearch}
			}
		case chmdl.ModelTypeSubscribe: // 订阅频道
			if !isBreak {
				re = &chmdl.SquareItem{ModelType: chmdl.ModelTypeSubscribe, ModelTitle: chmdl.ModelNameSubscribe}
				re.Items = s.SquareSubscribe(myChannels, channelList, actInfos, params.MobiApp, params.Device, int64(params.Build))
			}
		case chmdl.ModelTypeNew: // 订阅更新
			if len(myChannels.GetCards()) == 0 {
				continue
			}
			var (
				label string
				items []card.Handler
			)
			if label, items = s.SquareNew(myChannels, amplayer, myChannelCardItems, myChannelBadges, isFav, coins, params, plat, ip); len(items) == 0 {
				continue
			}
			re = &chmdl.SquareItem{ModelType: chmdl.ModelTypeNew, ModelTitle: chmdl.ModelNameNew, Offset: myChannels.NextOffset}
			if myChannels.HasMore {
				re.HasMore = 1
				isBreak = true
			} else {
				isBreak = false
			}
			re.DescButton = &cardm.Button{
				Text: fmt.Sprintf("管理订阅 %d", myChannels.Count),
				URI:  model.FillURI(model.GotoChannelTab, strconv.Itoa(chmdl.TabMineID), 0, 0, 0, nil),
			}
			re.Label = label
			re.Items = items
		case chmdl.ModelTypeScaned: // 最近看过的频道
			if len(viewChannels) == 0 {
				continue
			}
			if !isBreak {
				re = &chmdl.SquareItem{ModelType: chmdl.ModelTypeScaned, ModelTitle: chmdl.ModelNameScaned, Label: "最近看过的频道"}
				re.Items = s.SquareScaned(viewChannels)
			}
		case chmdl.ModelTypeRcmd: // 热门频道：频道 + 频道动态 + 频道卡
			if hotChannels == nil || len(hotChannels.Card) == 0 {
				continue
			}
			if !isBreak {
				var (
					items   *chmdl.SquareHot
					dynamic []*channelgrpc.DynamicCard
				)
				if myChannels != nil {
					dynamic = myChannels.GetDynamicList()
				}
				if items = s.SquareHot(hotChannels.Card, amplayer, hotChannelCardItems, hotChannelBadges, dynamic, isFav, coins, params, plat, ip); len(items.List) == 0 || len(items.Rcmd) == 0 {
					continue
				}
				re = &chmdl.SquareItem{ModelType: chmdl.ModelTypeRcmd, ModelTitle: chmdl.ModelNameRcmd, Label: "热门频道"}
				re.Offset = hotChannels.Offset
				re.Items = items
			}
		}
		if re != nil {
			res = append(res, re)
		}
	}
	return
}

// nolint:gocognit
func (s *Service) Rcmd2(c context.Context, params *chmdl.Param) (res *chmdl.SquareItem, err error) {
	var (
		hotChannels *channelgrpc.HotChannelReply
		myChannels  *channelgrpc.MyChannelsReply
		plat        = model.Plat(params.MobiApp, params.Device)
		ip          = metadata.String(c, metadata.RemoteIP)
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if hotChannels, err = s.chDao.HotChannel(ctx, params.MID, params.Offset); err != nil {
			log.Error("%v", err)
		}
		return
	})
	g.Go(func() (err error) {
		if myChannels, err = s.chDao.MyChannels(ctx, params.MID, params.OffsetNew, true, chmdl.NewSubVersion); err != nil {
			log.Error("%v", err)
			return nil
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	if hotChannels == nil || len(hotChannels.Card) == 0 {
		return
	}
	var (
		aids                []int64
		hotChannelBadges    = make(map[int64]map[int64]*operate.ChannelBadge)
		hotChannelCardItems = make(map[int64][]*operate.Card)
	)
	if myChannels != nil {
		for _, dynamic := range myChannels.GetDynamicList() {
			if dynamic.GetRid() == 0 {
				continue
			}
			aids = append(aids, dynamic.GetRid())
		}
	}
	for _, hc := range hotChannels.GetCard() {
		if hc == nil {
			continue
		}
		var tmpAids []int64
		for _, video := range hc.GetVideoCards() {
			if video.GetRid() == 0 {
				continue
			}
			tmpAids = append(tmpAids, video.GetRid())
			ci := &operate.Card{
				ID: video.GetRid(),
			}
			hotChannelCardItems[hc.Cid] = append(hotChannelCardItems[hc.Cid], ci)
			if video.GetBadgeTitle() != "" && video.GetBadgeBackground() != "" {
				var (
					hotChannelbadge map[int64]*operate.ChannelBadge
					ok              bool
				)
				if hotChannelbadge, ok = hotChannelBadges[hc.Cid]; !ok {
					hotChannelbadge = make(map[int64]*operate.ChannelBadge)
					hotChannelBadges[hc.Cid] = hotChannelbadge
				}
				hotChannelbadge[video.GetRid()] = &operate.ChannelBadge{
					Text:  video.BadgeTitle,
					Cover: video.BadgeBackground,
				}
			}
		}
		aids = append(aids, tmpAids...)
	}
	if len(aids) == 0 {
		return
	}
	var (
		amplayer  map[int64]*archivegrpc.ArcPlayer
		isFav     map[int64]bool
		coins     map[int64]int64
		aidsV2    []*archivegrpc.PlayAv
		isPlayurl bool
	)
	for _, aid := range aids {
		if aid != 0 {
			aidsV2 = append(aidsV2, &archivegrpc.PlayAv{Aid: aid})
			if (model.IsAndroid(plat) && params.Build > s.c.BuildLimit.MiaokaiAndroid) || (model.IsIOS(plat) && params.Build > s.c.BuildLimit.MiaokaiIOS) ||
				(plat == model.PlatAndroidI && params.Build > s.c.BuildLimit.ArcWithPlayerAndroid) || (plat == model.PlatIPhoneI && params.Build > s.c.BuildLimit.ArcWithPlayerIOS) {
				isPlayurl = true
			}
		}
	}
	g2, ctx2 := errgroup.WithContext(c)
	g2.Go(func() (err error) {
		amplayer, err = s.Archives(ctx2, aidsV2, isPlayurl)
		if err != nil {
			log.Error("%v", err)
		}
		return
	})
	if params.MID > 0 {
		g2.Go(func() (err error) {
			if isFav, err = s.favDao.IsFavoreds(ctx2, params.MID, aids); err != nil {
				log.Error("%v", err)
				return nil
			}
			return
		})
		g2.Go(func() (err error) {
			if coins, err = s.coinDao.IsCoins(ctx2, aids, params.MID); err != nil {
				log.Error("%v", err)
				return nil
			}
			return
		})
	}
	if err = g2.Wait(); err != nil {
		return
	}
	var (
		items   *chmdl.SquareHot
		dynamic []*channelgrpc.DynamicCard
	)
	dynamic = myChannels.GetDynamicList()
	if items = s.SquareHot(hotChannels.Card, amplayer, hotChannelCardItems, hotChannelBadges, dynamic, isFav, coins, params, plat, ip); len(items.List) == 0 || len(items.Rcmd) == 0 {
		return
	}
	res = &chmdl.SquareItem{ModelType: chmdl.ModelTypeRcmd, ModelTitle: chmdl.ModelTypeRcmd, Label: "热门频道"}
	res.Offset = hotChannels.Offset
	res.Items = items
	return
}

func (s *Service) SquareSubscribe(myChannels *channelgrpc.MyChannelsReply, channelList *channelgrpc.ChannelListReply, actInfos map[int64]*natgrpc.NativePage, mobiApp, device string, build int64) (res []*chmdl.Channel) {
	var pos = int64(1)
	if myChannels != nil && myChannels.Count > 0 {
		// 顶部我订阅的频道
		var mcsAll []*channelgrpc.ChannelCard
		// fc cp
		for _, tc := range myChannels.GetTops() {
			if tc == nil {
				log.Error("stick nil")
				continue
			}
			mcsAll = append(mcsAll, tc)
		}
		for _, cc := range myChannels.GetNormals() {
			if cc == nil {
				log.Error("normal nil")
				continue
			}
			mcsAll = append(mcsAll, cc)
		}
		for _, mcs := range mcsAll {
			if mcs == nil {
				log.Error("mcs nil")
				continue
			}
			i := &chmdl.Channel{}
			i.Position = pos
			i.FormChannel(mcs, actInfos, mobiApp, "", build, false)
			res = append(res, i)
			pos++
			// nolint:gomnd
			if len(res) == 9 {
				break
			}
		}
	}
	var (
		mvpItem = &chmdl.Channel{Cover: chmdl.SquareAllChannelIcon, Position: pos}
		addItem *chmdl.Channel
	)
	if len(res) == 0 {
		mvpItem.Title = _channelListTopTitle
		pos++
		addItem = &chmdl.Channel{
			Cover:    chmdl.SquareAddChannelIcon,
			Position: pos,
			Title:    "添加订阅",
			URI:      model.FillURI(model.GotoChannelTab, "100", 0, 0, 0, nil),
		}
	} else {
		mvpItem.Title = "更多频道"
	}
	if channelList != nil {
		var count int64
		// nolint:gomnd
		if count = channelList.Count; count > 9999 {
			count = 9999
		}
		mvpItem.CoverLabel = strconv.FormatInt(count, 10)
		mvpItem.CoverLabel2 = "个频道"
	}
	mvpItem.URI = model.FillURI(model.GotoChannelTab, "100", 0, 0, 0, nil)
	res = append(res, mvpItem)
	if addItem != nil {
		res = append(res, addItem)
	}
	return
}

func (s *Service) SquareNew(myChannels *channelgrpc.MyChannelsReply, amplayer map[int64]*archivegrpc.ArcPlayer, cardItems map[int64][]*operate.Card,
	myChannelBadges map[int64]map[int64]*operate.ChannelBadge, isFav map[int64]bool, coins map[int64]int64, params *chmdl.Param, plat int8, ip string) (label string, res []card.Handler) {
	var (
		cardNum           int
		channelInfocItems []*chmdl.InfocItem
		totalUpChannelNum = myChannels.GetUpdatedChannelNum()
		position          int64
	)
	build := params.Build
	mobiApp := params.MobiApp
	for _, nc := range myChannels.GetCards() {
		if nc == nil {
			continue
		}
		cis, ok := cardItems[nc.ChannelId]
		if !ok || len(cis) == 0 {
			continue
		}
		h := cardm.Handle(plat, cdm.CardGt("channel_new"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
		if h == nil {
			continue
		}
		badges := myChannelBadges[nc.ChannelId]
		op := &operate.Card{
			ID:    nc.ChannelId,
			Plat:  plat,
			Title: nc.ChannelName,
			Cover: nc.Icon,
			Param: strconv.FormatInt(nc.ChannelId, 10),
			Channel: &operate.Channel{
				LastUpTime:     nc.GetLastUpdateTs(),
				UpCnt:          nc.GetUpdatedRsNum(),
				TodayCnt:       nc.GetTodayRsNum(),
				FeatureCnt:     nc.GetFeaturedCnt(),
				OfficiaLVerify: nc.GetVerify(),
				CType:          nc.GetCtype(),
				Badges:         badges,
				IsFav:          isFav,
				Coins:          coins,
				Position:       position,
			},
			Items:   cis,
			Build:   build,
			MobiApp: mobiApp,
		}
		_ = h.From(amplayer, op)
		// 频道卡不完整，totalUpChannelNum减1
		if h.Get() != nil && !h.Get().Right && op.Channel.UpCnt > 0 {
			totalUpChannelNum--
		}
		if h.Get() != nil && h.Get().Right {
			position = h.Get().Idx
			res = append(res, h)
			// infoc: archive card in channel card.
			var cii []*chmdl.InfocItem
			for _, ci := range op.Items {
				cardNum++
				ciii := &chmdl.InfocItem{
					OID:      ci.ID,
					CardType: "av_2r",
					Pos:      cardNum,
				}
				if b, ok := badges[ci.ID]; ok {
					ciii.Corner = b.Text
				}
				cii = append(cii, ciii)
			}
			// infoc: channel card.
			channelInfoc := &chmdl.InfocItem{
				ChannelID: nc.ChannelId,
				Items:     cii,
			}
			channelInfocItems = append(channelInfocItems, channelInfoc)
		}
	}
	text := subscribeReplaceText("订阅", "收藏", card.FavTextReplace(params.MobiApp, int64(params.Build)))
	if myChannels.GetDefaultFeed() {
		label = "推荐频道"
	} else if totalUpChannelNum > 0 {
		label = fmt.Sprintf("我%s的%s个频道有更新", text, chmdl.MarkRed(strconv.Itoa(int(totalUpChannelNum))))
	} else {
		label = fmt.Sprintf("我%s的频道", text)
	}
	// infoc: the whole square page.
	var ss = &infocStatistics{}
	if params.Statistics != "" {
		if errTmp := json.Unmarshal([]byte(params.Statistics), &ss); errTmp != nil {
			log.Warn("Failed to parse request statistics: %+v, err %v", ss, errTmp)
		}
	}
	infoc := &chmdl.ChannelInfoc{
		EventId:     "cardshow",
		Page:        params.Spmid,
		Items:       channelInfocItems,
		CardNum:     cardNum,
		RequestUrl:  params.ReqURL,
		TimeIso:     params.TimeIso,
		Ip:          ip,
		AppId:       ss.AppID,
		Platform:    ss.Platform,
		Buvid:       params.Buvid,
		Version:     ss.Version,
		VersionCode: strconv.Itoa(params.Build),
		Mid:         strconv.FormatInt(params.MID, 10),
		Ctime:       strconv.FormatInt(params.TS*1000, 10),
		Abtest:      ss.ABTest,
		AutoRefresh: params.AutoRefresh,
		Pos:         "我订阅的频道更新",
		CurRefresh:  params.PN,
	}
	s.infoc(infoc)
	return
}

func (s *Service) SquareScaned(scaneds []*channelgrpc.ViewChannelCard) (res []*chmdl.SquareScaned) {
	var pos = int64(1)
	for _, scaned := range scaneds {
		if scaned == nil {
			continue
		}
		re := &chmdl.SquareScaned{
			CardType: "channel_scaned_v2",
			CardGoto: "channel_scaned_v2",
			Position: pos,
		}
		re.FormSquareScaned(scaned)
		res = append(res, re)
		pos++
	}
	return
}

func (s *Service) SquareHot(hots []*channelgrpc.ViewChannelCard, amplayer map[int64]*archivegrpc.ArcPlayer, cardItems map[int64][]*operate.Card,
	hotChannelBadges map[int64]map[int64]*operate.ChannelBadge, dynamics []*channelgrpc.DynamicCard, isFav map[int64]bool, coins map[int64]int64,
	params *chmdl.Param, plat int8, ip string) (res *chmdl.SquareHot) {
	build := params.Build
	mobiApp := params.MobiApp
	res = &chmdl.SquareHot{}
	for _, dynamic := range dynamics {
		if dynamic == nil {
			continue
		}
		d := &chmdl.HotDynamic{}
		d.FormHotDynamic(dynamic, amplayer)
		if d.Desc != "" {
			res.Dynamic = append(res.Dynamic, d)
		}
	}
	var (
		cardNum           int
		channelInfocItems []*chmdl.InfocItem
		position          int64
		listPosition      = int64(1)
	)
	for _, hot := range hots {
		if hot == nil {
			continue
		}
		scis, ok := cardItems[hot.GetCid()]
		if !ok || len(scis) == 0 {
			continue
		}
		h := cardm.Handle(plat, cdm.CardGt("channel_rcmd_v2"), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil)
		if h == nil {
			continue
		}
		badges := hotChannelBadges[hot.GetCid()]
		op := &operate.Card{
			ID:    hot.GetCid(),
			Plat:  plat,
			Title: hot.GetCname(),
			Param: strconv.FormatInt(hot.GetCid(), 10),
			Channel: &operate.Channel{
				LastUpTime: int64(hot.GetScanTs()),
				FeatureCnt: hot.GetFeaturedCnt(),
				IsAtten:    hot.GetSubscribed(),
				AttenCnt:   hot.GetSubscribedCnt(),
				Badges:     badges,
				IsFav:      isFav,
				Coins:      coins,
				Position:   position,
			},
			Items:   scis,
			Build:   build,
			MobiApp: mobiApp,
		}
		_ = h.From(amplayer, op)
		if h.Get() != nil && h.Get().Right {
			// 顶部频道list
			l := &chmdl.HotList{}
			l.FormHotList(hot, listPosition)
			res.List = append(res.List, l)
			listPosition++
			// 底部推荐详情
			position = h.Get().Idx
			res.Rcmd = append(res.Rcmd, h)
			// infoc: archive card in channel card.
			var cii []*chmdl.InfocItem
			for _, ci := range op.Items {
				cardNum++
				ciii := &chmdl.InfocItem{
					OID:      ci.ID,
					CardType: "av_2r",
					Pos:      cardNum,
				}
				if b, ok := badges[ci.ID]; ok {
					ciii.Corner = b.Text
				}
				cii = append(cii, ciii)
			}
			// infoc: channel card.
			channelInfoc := &chmdl.InfocItem{
				ChannelID: hot.GetCid(),
				Items:     cii,
			}
			channelInfocItems = append(channelInfocItems, channelInfoc)
		}
	}
	if len(channelInfocItems) == 0 {
		return
	}
	var ss = &infocStatistics{}
	if params.Statistics != "" {
		if errTmp := json.Unmarshal([]byte(params.Statistics), &ss); errTmp != nil {
			log.Warn("Failed to parse request statistics: %+v, err %v", ss, errTmp)
		}
	}
	infoc := &chmdl.ChannelInfoc{
		EventId:     "cardshow",
		Page:        params.Spmid,
		Items:       channelInfocItems,
		CardNum:     cardNum,
		RequestUrl:  params.ReqURL,
		TimeIso:     params.TimeIso,
		Ip:          ip,
		AppId:       ss.AppID,
		Platform:    ss.Platform,
		Buvid:       params.Buvid,
		Version:     ss.Version,
		VersionCode: strconv.Itoa(params.Build),
		Mid:         strconv.FormatInt(params.MID, 10),
		Ctime:       strconv.FormatInt(params.TS*1000, 10),
		Abtest:      ss.ABTest,
		AutoRefresh: params.AutoRefresh,
		Pos:         "热门频道",
	}
	s.infoc(infoc)
	return
}

func (s *Service) Home2(ctx context.Context, params *chmdl.Param) (*chmdl.Home2, error) {
	res := &chmdl.Home2{}
	eg := egv2.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		reply, err := s.topicDao.HasCreatedTopic(ctx, &topicsvc.HasCreatedTopicReq{Uid: params.MID})
		if err != nil {
			log.Error("s.topicDao.HasCreatedTopic mid=%d, error=%+v", params.MID, err)
			return nil
		}
		if reply.HasCreated {
			res.EntranceButton = constructEntranceButton()
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		squareItems, err := s.Home(ctx, params)
		if err != nil {
			log.Error("s.Home params=%+v, error=%+v", params, err)
			return nil
		}
		res.SquareItems = squareItems
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Home2 eg.Wait() err=%+v", err)
		return nil, err
	}
	return res, nil
}

func constructEntranceButton() *chmdl.EntranceButton {
	// EntranceButton结构目前只会出我的话题
	return &chmdl.EntranceButton{
		Type: "myTopic",
		Text: "我的话题",
		Link: "https://www.bilibili.com/h5/topic-active/my-topic?navhide=1",
	}
}

// nolint:gocognit
func (s *Service) Home(c context.Context, params *chmdl.Param) (res []*chmdl.SquareItem, err error) {
	var (
		myChannels   *channelgrpc.MyChannelsReply
		hotTopics    []*dynamicTopic.HotListDetail
		hotChannels  *channelgrpc.HotChannelReply
		plat         = model.Plat(params.MobiApp, params.Device)
		ip           = metadata.String(c, metadata.RemoteIP)
		newHotTopics []*topicmdl.NewHotTopicDetail
	)
	g, ctx := errgroup.WithContext(c)
	g.Go(func() (err error) {
		if myChannels, err = s.chDao.MyChannels2(ctx, params.MID, params.OffsetNew, true, chmdl.NewSubVersion); err != nil {
			log.Error("%v", err)
		}
		return
	})
	if params.MID > 0 {
		g.Go(func() (err error) {
			if hotTopics, err = s.topicDao.RcmdTopicsBigCard(ctx, params.MID, params.Build, params.Platform, params.MobiApp, params.Device, "", "", params.Device); err != nil {
				log.Error("%v", err)
				return nil
			}
			return
		})
		g.Go(func() (err error) {
			if !chmdl.CanNewTopicOnline(ctx) {
				return nil
			}
			args := constructHotNewTopicsArgs(params)
			newHotTopicsRsp, err := s.topicDao.HotNewTopics(ctx, args)
			if err != nil {
				log.Error("s.topicDao.HotNewTopics args=%+v, err=%+v", args, err)
				return nil
			}
			// 取封面用的动态资源id
			var dynIds []int64
			for _, v := range newHotTopicsRsp.TopicList {
				dynIds = append(dynIds, v.Rid)
			}
			simpleInfo, err := s.dynDao.DynSimpleInfos(ctx, &dynfeedgrpc.DynSimpleInfosReq{DynIds: dynIds})
			if err != nil {
				log.Error("s.dynDao.DynSimpleInfos args=%+v, err=%+v", args, err)
				return nil
			}
			for _, v := range newHotTopicsRsp.TopicList {
				if v == nil {
					continue
				}
				tmp := &topicmdl.NewHotTopicDetail{TopicDetail: v}
				if info, ok := simpleInfo.DynSimpleInfos[v.Rid]; ok {
					tmp.DynamicResourceId = info.Rid
				}
				newHotTopics = append(newHotTopics, tmp)
			}
			return nil
		})
	}
	g.Go(func() (err error) {
		if hotChannels, err = s.chDao.HotChannel(ctx, params.MID, params.OffsetRcmd); err != nil {
			log.Error("%v", err)
			return nil
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%v", err)
		return
	}
	var (
		aids                []int64
		myChannelCardItems  = make(map[int64][]*operate.Card)
		myChannelBadges     = make(map[int64]map[int64]*operate.ChannelBadge)
		hotChannelBadges    = make(map[int64]map[int64]*operate.ChannelBadge)
		hotChannelCardItems = make(map[int64][]*operate.Card)
		actTids             []int64
		drawIDs             []int64
	)
	if myChannels != nil {
		for _, nc := range myChannels.GetCards() {
			if nc == nil {
				continue
			}
			var tmpAids []int64
			for _, resource := range nc.GetResourceCards() {
				if resource.GetVideoCard() != nil && resource.GetVideoCard().GetRid() != 0 {
					tmpAids = append(tmpAids, resource.GetVideoCard().GetRid())
					ci := &operate.Card{
						ID: resource.GetVideoCard().GetRid(),
					}
					myChannelCardItems[nc.GetChannelId()] = append(myChannelCardItems[nc.GetChannelId()], ci)
					if resource.GetVideoCard().GetBadgeTitle() != "" && resource.GetVideoCard().GetBadgeBackground() != "" {
						var (
							myChannelbadge map[int64]*operate.ChannelBadge
							ok             bool
						)
						if myChannelbadge, ok = myChannelBadges[nc.GetChannelId()]; !ok {
							myChannelbadge = make(map[int64]*operate.ChannelBadge)
							myChannelBadges[nc.GetChannelId()] = myChannelbadge
						}
						myChannelbadge[resource.GetVideoCard().GetRid()] = &operate.ChannelBadge{
							Text:  resource.GetVideoCard().GetBadgeTitle(),
							Cover: resource.GetVideoCard().GetBadgeBackground(),
						}
					}
				}
			}
			aids = append(aids, tmpAids...)
		}
		for _, dynamic := range myChannels.GetDynamicList() {
			if dynamic.GetRid() == 0 {
				continue
			}
			aids = append(aids, dynamic.GetRid())
		}
		for _, top := range myChannels.GetTops() {
			if top.GetActAttr() == chmdl.ActiveTag {
				actTids = append(actTids, top.GetChannelId())
			}
		}
		for _, normals := range myChannels.GetNormals() {
			if normals.GetActAttr() == chmdl.ActiveTag {
				actTids = append(actTids, normals.GetChannelId())
			}
		}
	}
	if hotChannels != nil {
		for _, hc := range hotChannels.GetCard() {
			if hc == nil {
				continue
			}
			var tmpAids []int64
			for _, video := range hc.GetVideoCards() {
				if video.GetRid() == 0 {
					continue
				}
				tmpAids = append(tmpAids, video.GetRid())
				ci := &operate.Card{
					ID: video.GetRid(),
				}
				hotChannelCardItems[hc.Cid] = append(hotChannelCardItems[hc.Cid], ci)
				if video.GetBadgeTitle() != "" && video.GetBadgeBackground() != "" {
					var (
						hotChannelbadge map[int64]*operate.ChannelBadge
						ok              bool
					)
					if hotChannelbadge, ok = hotChannelBadges[hc.Cid]; !ok {
						hotChannelbadge = make(map[int64]*operate.ChannelBadge)
						hotChannelBadges[hc.Cid] = hotChannelbadge
					}
					hotChannelbadge[video.GetRid()] = &operate.ChannelBadge{
						Text:  video.BadgeTitle,
						Cover: video.BadgeBackground,
					}
				}
			}
			aids = append(aids, tmpAids...)
		}
	}
	for _, hotTopic := range hotTopics {
		// nolint:gomnd
		switch hotTopic.Type {
		case 2:
			if hotTopic.Rid != 0 {
				drawIDs = append(drawIDs, hotTopic.Rid)
			}
		case 8:
			if hotTopic.Rid != 0 {
				aids = append(aids, hotTopic.Rid)

			}
		}
	}
	for _, newTopic := range newHotTopics {
		switch newTopic.TopicDetail.DynamicType {
		case _dynTypeDraw:
			if newTopic.DynamicResourceId != 0 {
				drawIDs = append(drawIDs, newTopic.DynamicResourceId)
			}
		case _dynTypeVideo:
			if newTopic.DynamicResourceId != 0 {
				aids = append(aids, newTopic.DynamicResourceId)
			}
		}
	}
	var (
		amplayer map[int64]*archivegrpc.ArcPlayer
		isFav    map[int64]bool
		coins    map[int64]int64
		actInfos map[int64]*natgrpc.NativePage
		draws    map[int64]*dynmdl.DrawDetailRes
	)
	g2, ctx2 := errgroup.WithContext(c)
	if len(aids) > 0 {
		var aidsV2 []*archivegrpc.PlayAv
		for _, aid := range aids {
			aidsV2 = append(aidsV2, &archivegrpc.PlayAv{Aid: aid})
		}
		g2.Go(func() (err error) {
			amplayer, err = s.Archives(ctx2, aidsV2, true)
			if err != nil {
				log.Error("%v", err)
			}
			return
		})
		if params.MID > 0 {
			g2.Go(func() (err error) {
				if isFav, err = s.favDao.IsFavoreds(ctx2, params.MID, aids); err != nil {
					log.Error("%v", err)
				}
				return nil
			})
			g2.Go(func() (err error) {
				if coins, err = s.coinDao.IsCoins(ctx2, aids, params.MID); err != nil {
					log.Error("%v", err)
				}
				return nil
			})
		}
	}
	if s.c.Switch.SquareActive && len(actTids) > 0 {
		g2.Go(func() (err error) {
			if actInfos, err = s.natDao.NatInfoFromForeigns(c, actTids, 1); err != nil {
				log.Error("%v", err)
			}
			return
		})
	}
	if len(drawIDs) > 0 {
		g2.Go(func() error {
			if draws, err = s.dynDao.DrawDetails(ctx2, params.MID, drawIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if err = g2.Wait(); err != nil {
		return
	}
	var isBreak bool
	if params.OffsetNew != "" {
		isBreak = true
	}
	if !isBreak {
		// 我的订阅(我的收藏)
		if items := s.homeSubscribe(myChannels, actInfos, params.MobiApp, int64(params.Build)); len(items) > 0 {
			text := subscribeReplaceText(chmdl.ModelNameSubscribe, chmdl.ModelNameFav,
				card.FavTextReplace(params.MobiApp, int64(params.Build)))
			res = append(res, &chmdl.SquareItem{
				ModelType:  chmdl.ModelTypeSubscribe,
				ModelTitle: text,
				Label:      text,
				Items:      items,
				DescButton: &cardm.Button{
					Text: "查看全部",
					URI:  model.FillURI(model.GotoChannelTab, strconv.Itoa(chmdl.TabAll), 0, 0, 0, nil),
				},
			})
		}
		// 推荐话题
		res = append(res, s.makeHotTopicProcess(newHotTopics, hotTopics, amplayer, draws)...)
	}
	// 订阅更新
	if len(myChannels.GetCards()) > 0 {
		var (
			label string
			items []card.Handler
		)
		if label, items = s.SquareNew(myChannels, amplayer, myChannelCardItems, myChannelBadges, isFav, coins, params, plat, ip); len(items) > 0 {
			text := subscribeReplaceText("订阅", "收藏", card.FavTextReplace(params.MobiApp, int64(params.Build)))
			re := &chmdl.SquareItem{
				ModelType: chmdl.ModelTypeNew,
				ModelTitle: subscribeReplaceText(chmdl.ModelNameNew, chmdl.ModelNameFavNew,
					card.FavTextReplace(params.MobiApp, int64(params.Build))),
				Label:  label,
				Offset: myChannels.NextOffset,
				DescButton: &cardm.Button{
					Text: fmt.Sprintf("管理%s %d", text, myChannels.Count),
					URI:  model.FillURI(model.GotoChannelTab, strconv.Itoa(chmdl.TabMineID), 0, 0, 0, nil),
				},
				Items: items,
			}
			if myChannels.HasMore {
				re.HasMore = 1
				isBreak = true
			} else {
				isBreak = false
			}
			res = append(res, re)
		}
	}
	// 热门频道：频道 + 频道动态 + 频道卡
	if hotChannels != nil && len(hotChannels.Card) > 0 && !isBreak {
		var (
			items   *chmdl.SquareHot
			dynamic []*channelgrpc.DynamicCard
		)
		if myChannels != nil {
			dynamic = myChannels.GetDynamicList()
		}
		if items = s.SquareHot(hotChannels.Card, amplayer, hotChannelCardItems, hotChannelBadges, dynamic, isFav, coins, params, plat, ip); len(items.List) > 0 && len(items.Rcmd) > 0 {
			res = append(res, &chmdl.SquareItem{
				ModelType:  chmdl.ModelTypeRcmd,
				ModelTitle: chmdl.ModelNameRcmd,
				Label:      chmdl.ModelNameRcmd,
				Offset:     hotChannels.Offset,
				Items:      items,
			})
		}
	}
	return
}

func constructHotNewTopicsArgs(params *chmdl.Param) *topicsvc.HotNewTopicsReq {
	return &topicsvc.HotNewTopicsReq{
		Uid: params.MID,
		MetaData: &topiccommon.MetaDataCtrl{
			Platform:  params.Platform,
			Build:     strconv.Itoa(params.Build),
			MobiApp:   params.MobiApp,
			Buvid:     params.Buvid,
			Device:    params.Device,
			FromSpmid: params.Spmid,
			From:      "app-channel",
			Network:   params.NetType,
		},
	}
}

func subscribeReplaceText(srcText, dstText string, replaceCondition bool) string {
	if replaceCondition {
		return dstText
	}
	return srcText
}

func (s *Service) homeSubscribe(myChannels *channelgrpc.MyChannelsReply, actInfos map[int64]*natgrpc.NativePage,
	mobiApp string, build int64) (res []*chmdl.Channel) {
	var pos = int64(1)
	if myChannels != nil && myChannels.Count > 0 {
		// 顶部我订阅的频道
		var mcsAll []*channelgrpc.ChannelCard
		// fc cp
		for _, tc := range myChannels.GetTops() {
			if tc == nil {
				log.Error("stick nil")
				continue
			}
			mcsAll = append(mcsAll, tc)
		}
		for _, cc := range myChannels.GetNormals() {
			if cc == nil {
				log.Error("normal nil")
				continue
			}
			mcsAll = append(mcsAll, cc)
		}
		for _, mcs := range mcsAll {
			if mcs == nil {
				log.Error("mcs nil")
				continue
			}
			i := &chmdl.Channel{}
			i.Position = pos
			i.FormChannel(mcs, actInfos, mobiApp, "", build, false)
			res = append(res, i)
			pos++
			// nolint:gomnd
			if len(res) == 10 {
				break
			}
		}
	}
	return
}

func (s *Service) makeHotTopicProcess(newHotTopics []*topicmdl.NewHotTopicDetail, hotTopics []*dynamicTopic.HotListDetail, amplayer map[int64]*archivegrpc.ArcPlayer, draws map[int64]*dynmdl.DrawDetailRes) []*chmdl.SquareItem {
	var res []*chmdl.SquareItem
	if len(newHotTopics) > 0 {
		if items := s.homeNewHotTopic(newHotTopics, amplayer, draws); len(items) > 0 {
			res = append(res, &chmdl.SquareItem{
				ModelType:  chmdl.ModelTypeHotTopic,
				ModelTitle: chmdl.ModelNameHotTopic,
				Label:      chmdl.ModelNameHotTopic,
				Items:      items,
				DescButton: &cardm.Button{
					Text: "查看更多",
					URI:  "https://www.bilibili.com/h5/topic-active/topic-center?navhide=1",
				},
			})
		}
		return res
	}
	if len(hotTopics) > 0 {
		if items := s.homeHotTopic(hotTopics, amplayer, draws); len(items) > 0 {
			res = append(res, &chmdl.SquareItem{
				ModelType:  chmdl.ModelTypeHotTopic,
				ModelTitle: chmdl.ModelNameHotTopic,
				Label:      chmdl.ModelNameHotTopic,
				Items:      items,
				DescButton: &cardm.Button{
					Text: "查看更多",
					URI:  "https://www.bilibili.com/blackboard/topic-active.html?from_spmid=dt.dt.0.0&from_module=activity-card",
				},
			})
		}
	}
	return res
}

func (s *Service) homeNewHotTopic(hotTopics []*topicmdl.NewHotTopicDetail, amplayer map[int64]*archivegrpc.ArcPlayer, draws map[int64]*dynmdl.DrawDetailRes) []*chmdl.HotTopic {
	var res []*chmdl.HotTopic
	for _, hotTopic := range hotTopics {
		if hotTopic == nil {
			continue
		}
		i := &chmdl.HotTopic{
			ID:    hotTopic.TopicDetail.TopicId,
			Title: hotTopic.TopicDetail.TopicName,
			URI:   hotTopic.TopicDetail.JumpUrl,
		}
		var labels []string
		if hotTopic.TopicDetail.View != 0 {
			labels = append(labels, model.Stat64String(hotTopic.TopicDetail.View, "浏览"))
		}
		if hotTopic.TopicDetail.Discuss != 0 {
			labels = append(labels, model.Stat64String(hotTopic.TopicDetail.Discuss, "讨论"))
		}
		i.Label = strings.Join(labels, " · ")
		switch hotTopic.TopicDetail.DynamicType {
		case _dynTypeDraw:
			if draw, ok := draws[hotTopic.DynamicResourceId]; ok {
				for _, pic := range draw.Item.Pictures {
					if pic != nil && pic.ImgSrc != "" {
						i.Cover = pic.ImgSrc
						break
					}
				}
			}
		case _dynTypeVideo:
			if archive, ok := amplayer[hotTopic.DynamicResourceId]; ok {
				i.Cover = archive.Arc.Pic
			}
		}
		if hotTopic.TopicDetail.RcmdReason != nil {
			i.SedType = hotTopic.TopicDetail.RcmdReason.Text
			i.RcmdReason = &chmdl.ReasonStyle{
				Text:             hotTopic.TopicDetail.RcmdReason.Text,
				TextColor:        "#FFFA8E57",
				TextColorNight:   "#FFBA6B45",
				BgColor:          "#FFFFF1EA",
				BgColorNight:     "#FF3B352E",
				BorderColor:      "#FFFFF1EA",
				BorderColorNight: "#FF3B352E",
				BgStyle:          1,
			}
		}
		res = append(res, i)
	}
	return res
}

func (s *Service) homeHotTopic(hotTopics []*dynamicTopic.HotListDetail, amplayer map[int64]*archivegrpc.ArcPlayer, draws map[int64]*dynmdl.DrawDetailRes) (res []*chmdl.HotTopic) {
	for _, hotTopic := range hotTopics {
		if hotTopic == nil {
			continue
		}
		i := &chmdl.HotTopic{
			ID:      hotTopic.TopicId,
			Title:   hotTopic.TopicName,
			URI:     hotTopic.TopicLink,
			SedType: hotTopic.RcmdDesc,
		}
		var labels []string
		if hotTopic.HeatInfo != nil {
			if hotTopic.HeatInfo.View != 0 {
				labels = append(labels, model.Stat64String(hotTopic.HeatInfo.View, "浏览"))
			}
			if hotTopic.HeatInfo.Discuss != 0 {
				labels = append(labels, model.Stat64String(hotTopic.HeatInfo.Discuss, "讨论"))
			}
		}
		i.Label = strings.Join(labels, " · ")
		// nolint:gomnd
		switch hotTopic.Type {
		case 2: // 图文动态
			if draw, ok := draws[hotTopic.Rid]; ok {
				for _, pic := range draw.Item.Pictures {
					if pic != nil && pic.ImgSrc != "" {
						i.Cover = pic.ImgSrc
						break
					}
				}
			}
		case 8: // 视频动态
			if archive, ok := amplayer[hotTopic.Rid]; ok {
				i.Cover = archive.Arc.Pic
			}
		}
		if hotTopic.RcmdDesc != "" {
			i.RcmdReason = &chmdl.ReasonStyle{
				Text:             hotTopic.RcmdDesc,
				TextColor:        "#FFFA8E57",
				TextColorNight:   "#FFBA6B45",
				BgColor:          "#FFFFF1EA",
				BgColorNight:     "#FF3B352E",
				BorderColor:      "#FFFFF1EA",
				BorderColorNight: "#FF3B352E",
				BgStyle:          1,
			}
		}
		res = append(res, i)
	}
	return res
}
