package show

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"time"

	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	xecode "go-gateway/app/app-svr/app-card/ecode"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	cardm "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/card/rank"
	"go-gateway/app/app-svr/app-show/interface/model"
	agg "go-gateway/app/app-svr/app-show/interface/model/aggregation"
	"go-gateway/app/app-svr/app-show/interface/model/card"
	"go-gateway/app/app-svr/app-show/interface/model/feed"
	rankmdl "go-gateway/app/app-svr/app-show/interface/model/rank"
	"go-gateway/app/app-svr/app-show/interface/model/show"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
)

var (
	_emptyList2     = make([]cardm.Handler, 0)
	_auditingPass   = 1 // 审核通过
	_allHotEntrance = "全部热门"
	_hotChannelName = "hot-channel"
	_hotTopicName   = "hot-topic"
	_hotH5          = "hot-h5"
	_hotChannelUrl  = "/h5/popular/%d?navhide=1"
)

const _ipadAdaptBuild = 8930

func (s *Service) pickEntrances(c context.Context, mid int64, buvid string, build int, mobiApp, device string, isIpad bool) (entrances []*show.EntranceShow) {
	outputEntrances := make(map[string]struct{}) // 按照module_ID去重
	for _, v := range s.topEntrance {
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.PickEntrances, &feature.OriginResutl{
			BuildLimit: (mobiApp == "iphone" && device == "phone" && build <= 8590) || (mobiApp == "android" && build < 5440000) || (mobiApp == "iphone_b" && build <= 7370),
		}) && len(entrances) >= 3 { // 老版本只出最多3个入口
			break
		}
		if v.ModuleID == _hotChannelName { // 如果是分品类热门，需要判断是不是有视频
			if isIpad { // 分品类热门针对ipad平台 屏蔽
				continue
			}
			if res, resErr := s.dao.CacheAIChannelRes(c, v.ID); resErr != nil || len(res) == 0 {
				continue
			}
			v.Show.URI = s.c.Host.WWW + fmt.Sprintf(_hotChannelUrl, v.Show.EntranceID)
		}
		if _, ok := outputEntrances[v.ModuleID]; (!ok || v.ModuleID == _hotChannelName) && v.CanShow(show.TopEntranceFilterMeta{Mid: mid, Buvid: buvid, Device: device, Build: build, MobiApp: mobiApp}) {
			entrances = append(entrances, v.Show)
			outputEntrances[v.Show.ModuleID] = struct{}{}
		}
	}
	return
}

// FeedIndex2 feed index
// nolint:gomnd
func (s *Service) FeedIndex2(c context.Context, mid, idx, entranceId int64, plat int8, build, loginEvent int, mobiApp, device, buvid, spmid string, now time.Time) (res []cardm.Handler, ver string, config *show.HotConfig, err error) {
	var (
		ps                   = 10
		isIpad               = plat == model.PlatIPad
		cards                []*card.PopularCard
		infocs               []*feed.Item
		style                int8
		channelOrder         int // 下面三个字段用于上报赋值
		channelID            = entranceId
		channelName          string
		userfeature, trackID string
		resCode              int
		isrcmd               bool
	)
	config = &show.HotConfig{
		ItemTitle:       s.c.ShowHotConfig.ItemTitle,
		BottomText:      s.c.ShowHotConfig.BottomText,
		BottomTextCover: s.c.ShowHotConfig.BottomTextCover,
		BottomTextURL:   s.c.ShowHotConfig.BottomTextURL,
		HeadImage:       s.middleTopPhoto,
		TopItems:        s.pickEntrances(c, mid, buvid, build, mobiApp, device, isIpad),
	}
	config.ShareInfo = s.handleShareInfo(config.ShareInfo, entranceId)
	if isIpad {
		ps = 20
	}
	var key int
	if mid > 0 {
		key = int((mid / 1000) % 10)
	} else {
		key = int((crc32.ChecksumIEEE([]byte(buvid)) / 1000) % 10)
	}
	// HotDynamic====================
	// cards = append(cards[:0], append([]*card.PopularCard{&card.PopularCard{Type: model.GotoHotDynamic, ReasonType: 0, FromType: "recommend"}}, cards[0:]...)...)
	// HotDynamic====================
	//build
	if (mobiApp == "iphone" && device == "phone" && build > 8230) || (mobiApp == "android" && build > 5345000) ||
		(mobiApp == "android_i" && build >= 2025000) || (mobiApp == "iphone_b" && build > 7370) || (mobiApp == "h5") ||
		(mobiApp == "iphone" && device == "pad" && build > _ipadAdaptBuild) {
		style = cdm.HotCardStyleHideUp
	} else {
		style = cdm.HotCardStyleOld
	}
	if entranceId == 0 {
		// 实验组内走ai下发控制，非实验组走老的redis中获取
		if s.AIGroup(c, mid, buvid) {
			cards, userfeature, _, resCode, isrcmd = s.hotRcmd2(c, idx, mid, 0, plat, build, buvid, mobiApp, nil, entranceId, 0, ps, key, nil)
		} else {
			cards = s.PopularCardTenList(c, key, int(idx), ps)
		}
		if len(cards) == 0 {
			err = xecode.AppNotData
			res = _emptyList2
			return
		}
		channelOrder = 0
		channelName = _allHotEntrance
	} else {
		cardCache := s.PopularEntranceAI(c, entranceId)
		if len(cardCache) > int(idx) {
			cards = cardCache[idx:]
		} else {
			err = xecode.AppNotData
			res = _emptyList2
			return
		}
		for i := 0; i < len(config.TopItems); i++ {
			if config.TopItems[i].EntranceID == entranceId {
				channelOrder = i
				channelName = config.TopItems[i].Title
				break
			}
		}
	}
	//build
	res, infocs = s.dealItem2(c, plat, build, ps, cards, mid, idx, style, mobiApp, device, isrcmd)
	ver = strconv.FormatInt(now.Unix(), 10)
	if len(res) == 0 {
		err = xecode.AppNotData
		res = _emptyList2
		return
	}
	//infoc
	for i := 0; i < len(infocs); i++ { // 对应的卡片信息的这部分上报数据是一样的
		infocs[i].ChannelID = strconv.Itoa(int(channelID))
		infocs[i].ChannelOrder = strconv.Itoa(channelOrder)
		infocs[i].ChannelName = channelName
	}
	if len(infocs) > 0 {
		trackID = infocs[0].TrackID
	}
	infoc := &feedInfoc{
		mobiApp:     mobiApp,
		device:      device,
		build:       strconv.Itoa(build),
		now:         now.Format("2006-01-02 15:04:05"),
		loginEvent:  strconv.Itoa(loginEvent),
		mid:         strconv.FormatInt(mid, 10),
		buvid:       buvid,
		page:        strconv.Itoa((int(idx) / ps) + 1),
		spmid:       spmid,
		feed:        infocs,
		url:         "/x/v2/show/popular/index",
		env:         env.DeployEnv,
		trackid:     trackID,
		userfeature: userfeature,
		returnCode:  strconv.Itoa(resCode),
		flush:       "0",
	}
	// 	code=0	正常返回结果
	// code=-2	ai内部超时，服务端出灾备
	// code=-3	所有内容已经刷完
	// 其余取值	错误，服务端需要出灾备
	if isrcmd {
		infoc.isrec = "1"
	}
	s.infocfeed(infoc)
	return
}

// dealItem feed item
// nolint:gocognit
func (s *Service) dealItem2(c context.Context, plat int8, build, ps int, cards []*card.PopularCard, mid, idx int64, style int8, mobiApp, device string, isrcmd bool) (is []cardm.Handler, infocs []*feed.Item) {
	var (
		max                           = int64(100)
		_fTypeOperation               = "operation"
		aids, avUpIDs, upIDs, rnUpIDs []int64
		amplayer                      map[int64]*arcgrpc.ArcPlayer
		feedcards                     []*card.PopularCard
		err                           error
		rank                          *operate.Card
		accountm                      map[int64]*accountgrpc.Card
		isAtten                       map[int64]int8
		statm                         map[int64]*relationgrpc.StatReply
		aggRes                        map[int64]*agg.Aggregation
		hotIDs                        []int64
		rcmdArc                       map[int64]struct{}
	)
	cardSet := map[int64]*operate.Card{}
	eventTopic := map[int64]*operate.Card{}
LOOP:
	for pos, ca := range cards {
		var cardIdx = idx + int64(pos+1)
		if !isrcmd {
			if cardIdx > max && ca.FromType != _fTypeOperation {
				continue
			}
			if plat != model.PlatH5 { // h5 gives all operation cards
				cardPlat := plat
				if mobiApp == "iphone" && device == "pad" { // 如果是新版ipad粉，需要在PopularCardPlat里面用iphone的plat
					cardPlat = model.PlatIPhone
				}
				if config, ok := ca.PopularCardPlat[cardPlat]; ok {
					for _, l := range config {
						if model.InvalidBuild(build, l.Build, l.Condition) {
							continue LOOP
						}
					}
				} else if ca.FromType == _fTypeOperation {
					continue LOOP
				}
			}
		}
		tmp := &card.PopularCard{}
		*tmp = *ca
		tmp.Idx = cardIdx
		feedcards = append(feedcards, tmp)
		switch ca.Type {
		case model.GotoAv:
			aids = append(aids, ca.Value)
			if ca.HotwordID != 0 {
				hotIDs = append(hotIDs, ca.HotwordID)
			}
		case model.GotoRank:
			if plat == model.PlatH5 || plat == model.PlatIPad { // h5和ipad 不展示 排行榜卡片
				continue
			}
			rank = &operate.Card{}
			rank.FromRank(s.rankCache2)
		case model.GotoUpRcmdNew:
			cardm, as, upid := s.cardSetChange(c, ca.Value)
			aids = append(aids, as...)
			for id, card := range cardm {
				cardSet[id] = card
			}
			rnUpIDs = append(rnUpIDs, upid)
		case model.GotoEventTopic:
			if plat == model.PlatIPad { // ipad不出事件专题卡
				continue
			}
			eventTopic = s.eventTopicChange(c, plat, ca.Value)
		}
		if len(feedcards) == ps && !isrcmd {
			break
		}
	}
	if len(aids) != 0 {
		var aidsReq []*arcgrpc.PlayAv
		//没有cid，只需处理aid
		for _, v := range aids {
			aidsReq = append(aidsReq, &arcgrpc.PlayAv{Aid: v})
		}
		if amplayer, err = s.arc.ArcsPlayer(c, aidsReq); err != nil {
			log.Error("%+v", err)
			return
		}
		for _, a := range amplayer {
			if a == nil || a.Arc == nil {
				continue
			}
			avUpIDs = append(avUpIDs, a.Arc.Author.Mid)
		}
	}
	switch style {
	case cdm.HotCardStyleShowUp, cdm.HotCardStyleHideUp:
		upIDs = append(upIDs, avUpIDs...)
	}
	upIDs = append(upIDs, rnUpIDs...)
	avUpIDs = append(avUpIDs, rnUpIDs...)
	g, ctx := errgroup.WithContext(c)
	var innerArc map[int64]*rankmdl.InnerAttr
	if len(aids) != 0 {
		g.Go(func() error {
			innerArc = s.controld.CircleReqInternalAttr(ctx, aids)
			return nil
		})
	}
	if len(avUpIDs) > 0 {
		g.Go(func() (err error) {
			if accountm, err = s.acc.Cards3GRPC(ctx, avUpIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(upIDs) > 0 {
		g.Go(func() (err error) {
			if statm, err = s.reldao.StatsGRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		if mid != 0 {
			g.Go(func() error {
				isAtten = s.acc.IsAttentionGRPC(ctx, upIDs, mid)
				return nil
			})
		}
	}
	if len(hotIDs) != 0 {
		g.Go(func() (err error) {
			if aggRes, err = s.dao.Aggregations(ctx, hotIDs); err != nil {
				log.Error("[dealItem2] s.agg.Aggregation() error(%v)", err)
			}
			return nil
		})
	}
	g.Go(func() (err error) {
		if rcmdArc, err = s.rcmmnd.Recommend(ctx); err != nil {
			log.Error("%v", err)
			return nil
		}
		return
	})
	_ = g.Wait()
	for _, ca := range feedcards {
		var (
			r        = ca.PopularCardToAiChange()
			main     interface{}
			cardType cdm.CardType
			op       = &operate.Card{TrackID: ca.TrackID}
		)
		r.Style = style
		op.From(cdm.CardGt(r.Goto), r.ID, 0, plat, build, mobiApp)
		switch r.Style {
		case cdm.HotCardStyleShowUp, cdm.HotCardStyleHideUp:
			switch r.Goto {
			case model.GotoAv:
				cardType = cdm.SmallCoverV5
			}
		}
		switch r.Goto {
		case model.GotoAv:
			var isOverseas bool
			if innerArc != nil && innerArc[r.ID] != nil {
				isOverseas = innerArc[r.ID].OverSeaBlock
			}
			if a, ok := amplayer[r.ID]; ok && a != nil && a.Arc != nil && (!isOverseas || !model.IsOverseas(plat)) {
				main = map[int64]*arcgrpc.ArcPlayer{a.Arc.Aid: a}
				r.HideButton = true
				if cardType == cdm.SmallCoverV5 {
					op.Switch = cdm.SwitchCooperationShow
				} else {
					op.Switch = cdm.SwitchCooperationHide
				}
			}
			if res, ok := aggRes[ca.HotwordID]; ok && res.State == _auditingPass {
				op.ShowHotword = true
				op.Tid = res.ID
				op.Cover = s.c.Aggregation.Icon
				op.RedirectURL = agg.ToResH5URl(res.ID)
				op.Subtitle = res.Title
			}
			if rcmdArc != nil {
				if _, ok := rcmdArc[r.ID]; ok {
					op.IsPopular = true
				}
			}
		case model.GotoRank:
			ams := map[int64]*arcgrpc.ArcPlayer{}
			for aid, a := range s.rankArchivesCache {
				ams[aid] = &arcgrpc.ArcPlayer{Arc: a}
			}
			main = map[cdm.Gt]interface{}{cdm.GotoAv: ams}
			op = rank
		case model.GotoUpRcmdNew:
			main = amplayer
			op = cardSet[r.ID]
			if op != nil {
				op.Plat = plat // 通过op将plat传入，便于对ipad进行新人卡视频数量过滤
			}
		case model.GotoEventTopic:
			op = eventTopic[r.ID]
		}
		h := cardm.Handle(plat, cdm.CardGt(r.Goto), cardType, cdm.ColumnSvrSingle, r, nil, isAtten, nil, statm, accountm, nil)
		if h == nil {
			log.Warn("dealItem2NotMatchHandle ca(%v)", ca)
			continue
		}
		op.FromDev(mobiApp, plat, build)
		_ = h.From(main, op)
		h.Get().FromType = ca.FromType
		h.Get().Idx = ca.Idx
		if h.Get().Right {
			h.Get().ThreePointWatchLater()
			is = append(is, h)
		}
		// infoc
		tinfo := &feed.Item{
			Goto:       ca.Type,
			Param:      strconv.FormatInt(ca.Value, 10),
			URI:        h.Get().URI,
			FromType:   ca.FromType,
			Idx:        h.Get().Idx,
			CornerMark: ca.CornerMark,
			CardStyle:  r.Style,
			HotAggreID: ca.HotwordID,
			TrackID:    ca.TrackID,
			Source:     ca.Source,
			AvFeature:  []byte(`""`),
		}
		if ca.AvFeature != nil {
			tinfo.AvFeature = ca.AvFeature
		}
		if r.RcmdReason != nil {
			tinfo.RcmdContent = r.RcmdReason.Content
		}
		if r.CoverGif != "" && r.Goto == model.GotoAv {
			tinfo.CoverType = "gif"
		} else {
			tinfo.CoverType = "pic"
		}
		if op != nil {
			switch r.Goto {
			case model.GotoEventTopic:
				tinfo.Item = append(tinfo.Item, &feed.Item{Param: op.URI, Goto: string(op.Goto)})
			default:
				for _, tmp := range op.Items {
					tinfo.Item = append(tinfo.Item, &feed.Item{Param: strconv.FormatInt(tmp.ID, 10), Goto: string(tmp.Goto)})
				}
			}
		}
		infocs = append(infocs, tinfo)
		// infoc
		if len(is) == ps {
			break
		}
	}
	rl := len(is)
	if rl == 0 {
		is = _emptyList2
		return
	}
	return
}

func (s *Service) RankCard() (ranks []*rank.Rank, aids []int64) {
	const _limit = 3
	ranks = make([]*rank.Rank, 0, _limit)
	aids = make([]int64, 0, _limit)
	for _, rank := range s.rankCache2 {
		ranks = append(ranks, rank)
		aids = append(aids, rank.Aid)
		if len(ranks) == _limit {
			break
		}
	}
	return
}

// 兜底操作
func (s *Service) handleShareInfo(in show.ShareInfo, entranceId int64) (out show.ShareInfo) {
	if entranceId == 0 {
		in.CurrentTopPhoto = s.middleTopPhoto
		in.CurrentTitle = _allHotEntrance
	}
	for _, item := range s.topEntrance {
		if item.ID == entranceId {
			in.CurrentTopPhoto = item.TopPhoto
			in.CurrentTitle = item.Title
			in.ShareDesc = item.ShareDesc
			in.ShareTitle = item.ShareTitle
			in.ShareSubTitle = item.ShareSubTitle
			in.ShareIcon = item.ShareIcon
		}
	}
	if in.ShareDesc == "" {
		in.ShareDesc = s.c.ShowHotConfig.ShareDesc
	}
	if in.ShareTitle == "" {
		in.ShareTitle = s.c.ShowHotConfig.ShareTitle
	}
	if in.ShareSubTitle == "" {
		in.ShareSubTitle = s.c.ShowHotConfig.ShareSubTitle
	}
	if in.ShareIcon == "" {
		in.ShareIcon = s.c.ShowHotConfig.ShareIcon
	}
	return in
}
