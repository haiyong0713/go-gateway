package show

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	"go-common/component/metadata/locale"
	"go-common/library/conf/env"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/text/translate/chinese"

	roomgategrpc "git.bilibili.co/bapis/bapis-go/live/xroom-gate"

	xecode "go-gateway/app/app-svr/app-card/ecode"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	cdmc "go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	cardapi "go-gateway/app/app-svr/app-card/interface/model/card/proto"
	cardmv2 "go-gateway/app/app-svr/app-card/interface/model/card/v2"
	swEcode "go-gateway/app/app-svr/app-show/ecode"
	svApi "go-gateway/app/app-svr/app-show/interface/api"
	api "go-gateway/app/app-svr/app-show/interface/api/popular"
	"go-gateway/app/app-svr/app-show/interface/model"
	agg "go-gateway/app/app-svr/app-show/interface/model/aggregation"
	"go-gateway/app/app-svr/app-show/interface/model/card"
	"go-gateway/app/app-svr/app-show/interface/model/feed"
	popularmodel "go-gateway/app/app-svr/app-show/interface/model/popular"
	rankmdl "go-gateway/app/app-svr/app-show/interface/model/rank"
	rcmod "go-gateway/app/app-svr/app-show/interface/model/recommend"
	"go-gateway/app/app-svr/app-show/interface/model/show"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	bgroupApi "git.bilibili.co/bapis/bapis-go/platform/service/bgroup/v2"
)

var (
	_emptyList3  = make([]*cardapi.Card, 0)
	_hotTagName  = "hot-tab"
	_hotPageName = "hot-page"
	_allHotName  = "全站热门"
	_svOff       = 0
)

const (
	_specialKey   = 10
	_allHotID     = "0"
	_allHotOrder  = "0"
	_articlesFrom = 3 //热门用户的from是3
	_svideoIndex  = 2
	_svideoAggr   = 3
	_uriSVideo    = "bilibili://inline/play_list/%d/%d"
	_uriH5Index   = "/h5/popular/%d?navhide=1"
)

func buildSVideoURI(index, entranceID, focusAID int64) string {
	u := fmt.Sprintf(_uriSVideo, index, entranceID)
	p := url.Values{}
	if focusAID > 0 {
		p.Set("focus_aid", strconv.FormatInt(focusAID, 10))
	}
	query := p.Encode()
	if query == "" {
		return u
	}
	return fmt.Sprintf("%s?%s", u, query)
}

func (s *Service) pickEntrancesV2(c context.Context, mid int64, buvid string, build int, mobiApp, device string, isPad bool, plat int8) (entrances []*api.EntranceShow) {
	outputEntrances := make(map[string]struct{}) // 按照module_ID去重
	bgroupResult := s.batchFetchBgroupResult(c, s.topEntrance, mid)
	for _, v := range s.topEntrance {
		topEntraceFilterMeta := show.TopEntranceFilterMeta{Mid: mid, Buvid: buvid, MobiApp: mobiApp, Build: build, Device: device, BGroupResult: bgroupResult}
		if _, ok := outputEntrances[v.ModuleID]; (!ok || v.ModuleID == _hotChannelName || v.ModuleID == _hotTopicName || v.ModuleID == _hotH5) && v.CanShow(topEntraceFilterMeta) {
			if v.Show != nil {
				if v.ModuleID == _hotChannelName || v.ModuleID == _hotTopicName { // 如果是分品类热门，需要判断是不是有视频
					if isPad { // 分品类热门针对ipad平台 屏蔽
						continue
					}
					if res, resErr := s.dao.CacheAIChannelRes(c, v.ID); resErr != nil || len(res) == 0 {
						continue
					}
					if v.ModuleID == _hotTopicName && s.displaySVideo(buvid, build, plat) {
						// v.Show.URI = fmt.Sprintf(_uriSVideo, _svideoIndex, v.Show.EntranceID)
						v.Show.URI = buildSVideoURI(_svideoIndex, v.Show.EntranceID, 0)
					} else {
						v.Show.URI = s.c.Host.WWW + fmt.Sprintf(_uriH5Index, v.Show.EntranceID)
					}
				}
				vproto := &api.EntranceShow{
					Icon:       v.Show.Icon,
					Title:      v.Show.Title,
					ModuleId:   v.Show.ModuleID,
					Uri:        v.Show.URI,
					EntranceId: v.Show.EntranceID,
					TopPhoto:   v.Show.TopPhoto,
				}
				if v.Show.Bubble != nil {
					bubble := &api.Bubble{
						BubbleContent: v.Show.Bubble.BubbleContent,
						Version:       v.Show.Bubble.Version,
						Stime:         v.Show.Bubble.Stime,
					}
					vproto.Bubble = bubble
				}
				entrances = append(entrances, vproto)
			}
			outputEntrances[v.Show.ModuleID] = struct{}{}
		}
	}
	return
}

func (s *Service) batchFetchBgroupResult(ctx context.Context, in []*show.EntranceMem, mid int64) map[string]bool {
	var groups []*bgroupApi.MemberInReq_MemberInReqSingle
	for _, v := range in {
		if v.BGroup.Business == "" || v.BGroup.Name == "" {
			continue
		}
		groups = append(groups, &bgroupApi.MemberInReq_MemberInReqSingle{Business: v.BGroup.Business, Name: v.BGroup.Name})
	}
	if len(groups) == 0 {
		return nil
	}
	req := &bgroupApi.MemberInReq{
		Member:    strconv.FormatInt(mid, 10),
		Groups:    groups,
		Dimension: bgroupApi.Mid,
	}
	reply, err := s.bGroupClient.MemberIn(ctx, req)
	if err != nil {
		log.Error("s.bGroupClient.MemberIn error(%+v), mid(%d)", err, mid)
		//出错返回空，都不会展示
		return nil
	}
	result := make(map[string]bool, len(reply.Results))
	for _, v := range reply.Results {
		if v == nil {
			continue
		}
		result[show.BGroupKey(v.Business, v.Name)] = v.In
	}
	return result
}

// FeedIndex3 feed index
// nolint:gocognit,gomnd
func (s *Service) FeedIndex3(c context.Context, mid, entranceId, idx int64, plat int8, build int, loginEvent int32, mobiApp, device, buvid, spmid string, topAids []int64, source int32, now time.Time, loc locale.Locale, flush int32, ad *api.PopularAd) (res []*cardapi.Card, ver string, config *api.Config, err error) {
	var (
		ps           = 10
		isPad        = (plat == model.PlatIPad) || (plat == model.PlatAndroidHD)
		cards        []*card.PopularCard
		infocs       []*feed.Item
		style        = cdm.HotCardStyleHideUp
		hasLargeCard bool
		channelOrder int // 下面三个字段用于上报赋值
		channelID    = entranceId
		channelName  string
		topAidMap    = make(map[int64]struct{})
		userfeature  string
		resCode      int
		isrcmd       bool
		hit          int64 //0表示没有命中，1表示命中
		adBizData    *rcmod.BizData
	)
	if buvidUint := crc32.ChecksumIEEE([]byte(buvid)); buvidUint%10 < uint32(s.c.Custom.Hit) {
		hit = 1
	}
	config = &api.Config{
		ItemTitle:       s.c.ShowHotConfig.ItemTitle,
		BottomText:      s.c.ShowHotConfig.BottomText,
		BottomTextCover: s.c.ShowHotConfig.BottomTextCover,
		BottomTextUrl:   s.c.ShowHotConfig.BottomTextURL,
		HeadImage:       s.middleTopPhoto,
		TopItems:        s.pickEntrancesV2(c, mid, buvid, build, mobiApp, device, isPad, plat),
		Hit:             hit,
	}
	config.PageItems = s.dealTopItems(config.TopItems) // 给中间页入口项赋值和过滤
	if isPad {
		ps = 20
	}
	var key int
	if mid > 0 {
		key = int((mid / 1000) % 10)
	} else {
		key = int((crc32.ChecksumIEEE([]byte(buvid)) / 1000) % 10)
	}
	if entranceId == 0 {
		if _, ok := s.largeCardsMids[mid]; ok {
			key = _specialKey
		}
		// 实验组内走ai下发控制，非实验组走老的redis中获取
		if s.AIGroup(c, mid, buvid) {
			cards, userfeature, adBizData, resCode, isrcmd = s.hotRcmd2(c, idx, mid, source, plat, build, buvid, mobiApp, topAids, entranceId, 0, ps, key, ad)
		} else {
			cards = s.PopularCardTenList(c, key, int(idx), ps)
		}
		if len(cards) == 0 {
			err = xecode.AppNotData
			res = _emptyList3
			return
		}
		if idx == 0 && (plat == model.PlatIPhone || plat == model.PlatAndroid) { // build号过滤
			cards, _ = s.dealLargeCardAndEventTopic(cards, key, mid)
		}
		channelOrder = 0
		channelName = _allHotEntrance
	} else {
		cardCache := s.PopularEntranceAI(c, entranceId)
		if len(cardCache) > int(idx) {
			cards = cardCache[idx:]
		} else {
			err = xecode.AppNotData
			res = _emptyList3
			return
		}
		for i := 0; i < len(config.PageItems); i++ {
			if config.PageItems[i].EntranceId == entranceId {
				channelOrder = i
				channelName = config.PageItems[i].Title
				break
			}
		}
		if channelOrder == 0 {
			log.Info("EntranceIdNotMatchPageItems entranceId(%d), mid(%d), buvid(%s), build(%d), mobiApp(%s), device(%s)", entranceId, mid, buvid, build, mobiApp, device)
		}
	}
	for _, item := range topAids { // 把置顶aid转成map，方便操作
		topAidMap[item] = struct{}{}
	}
	trackID := s.getTrackID(config.PageItems, entranceId, source) // 根据来源页获取trackID, 下发到视频小卡的uri里面
	//build
	res, infocs, hasLargeCard = s.dealItem3(c, plat, build, ps, cards, mid, idx, style, mobiApp, buvid, device, topAidMap, trackID, loc, isrcmd, adBizData)
	if len(config.TopItems) > 0 && idx == 0 && len(cards) > 0 && ((mobiApp == "iphone" && device == "phone") || (mobiApp == "android") || (mobiApp == "android_i") || (mobiApp == "iphone_i")) && source == 0 && entranceId == 0 {
		res = s.dealEntrance(config.TopItems, plat, hasLargeCard, res)
	}
	ver = strconv.FormatInt(now.Unix(), 10)
	if len(res) == 0 {
		err = xecode.AppNotData
		res = _emptyList3
		return
	}
	for _, item := range infocs { // 对应的卡片信息的这部分上报数据是一样的
		item.ChannelID = strconv.Itoa(int(channelID))
		item.ChannelOrder = strconv.Itoa(channelOrder)
		item.ChannelName = channelName
		if item.TrackID != "" {
			trackID = item.TrackID
		}
	}
	//infoc
	adBizDataByte, _ := json.Marshal(adBizData)
	infoc := &feedInfoc{
		mobiApp:     mobiApp,
		device:      device,
		build:       strconv.Itoa(build),
		now:         now.Format("2006-01-02 15:04:05"),
		loginEvent:  strconv.Itoa(int(loginEvent)),
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
		flush:       strconv.Itoa(int(flush)),
		adBizData:   string(adBizDataByte),
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

// FeedIndex3 feed index
func (s *Service) FeedIndexSvideo(c context.Context, entranceId int64, idx int64) (res *svApi.IndexSVideoReply, err error) {
	cardCache := s.PopularEntranceAI(c, entranceId)
	if len(cardCache) <= int(idx) {
		return nil, swEcode.ActivityNothingMore
	}
	cards := cardCache[idx:]
	list := make([]*svApi.SVideoItem, 0, len(cards))
	var (
		nextIdx int
		ps      = 20
		hasMore int32
	)
	for k, card := range cards {
		if card.Type != model.GotoAv {
			continue
		}
		tmp := &svApi.SVideoItem{
			Rid:   card.Value,
			Uid:   0,
			Index: idx + int64(k),
		}
		list = append(list, tmp)
		if len(list) >= ps {
			nextIdx = int(idx) + k + 1
			break
		}
	}
	if nextIdx > 0 && nextIdx < len(cardCache) {
		hasMore = _hasMore
	}
	res = &svApi.IndexSVideoReply{
		List:    list,
		Offset:  strconv.FormatInt(int64(nextIdx), 10),
		HasMore: hasMore,
		Top:     s.entranceTop(entranceId),
	}
	return
}

func (s *Service) entranceTop(entranceID int64) (res *svApi.SVideoTop) {
	for _, v := range s.topEntrance {
		if v.ID == entranceID {
			res = &svApi.SVideoTop{
				Title: v.Title,
				Desc:  v.ShareDesc,
			}
		}
	}
	return
}

// dealItem feed item
// nolint:gocognit
func (s *Service) dealItem3(c context.Context, plat int8, build, ps int, cards []*card.PopularCard, mid, idx int64, style int8,
	mobiApp, buvid, dev string, topAids map[int64]struct{}, trackID string, loc locale.Locale, isrcmd bool, adBizData *rcmod.BizData) (is []*cardapi.Card, infocs []*feed.Item, hasLargeCard bool) {
	var (
		max                                          = int64(100)
		_fTypeOperation                              = "operation"
		aids, avUpIDs, upIDs, rnUpIDs, lives, artIDs []int64
		amplayer                                     map[int64]*arcgrpc.ArcPlayer
		innerArc                                     map[int64]*rankmdl.InnerAttr
		liveplayer                                   map[int64]*roomgategrpc.EntryRoomInfoResp_EntryList
		feedcards                                    []*card.PopularCard
		rank                                         *operate.Card
		accountm                                     map[int64]*accountgrpc.Card
		isAtten                                      map[int64]int8
		authorRelations                              map[int64]*relationgrpc.InterrelationReply
		statm                                        map[int64]*relationgrpc.StatReply
		aggRes                                       map[int64]*agg.Aggregation
		hotIDs                                       []int64
		metam                                        map[int64]*article.Meta
		clocale, slocale                             string
		isHant                                       bool
	)
	clocale = loc.CLocale.Region
	if loc.CLocale.Language != "" || loc.CLocale.Script != "" {
		var clocaleTmp = strings.Join([]string{loc.CLocale.Language, loc.CLocale.Script}, "-")
		clocale = strings.Join([]string{clocaleTmp, clocale}, "_")
	}
	slocale = loc.SLocale.Region
	if loc.SLocale.Language != "" || loc.SLocale.Script != "" {
		var slocaleTmp = strings.Join([]string{loc.SLocale.Language, loc.SLocale.Script}, "-")
		slocale = strings.Join([]string{slocaleTmp, slocale}, "_")
	}
	if model.IsHant(clocale, slocale) {
		isHant = true
	}
	cardSet := map[int64]*operate.Card{}
	eventTopic := map[int64]*operate.Card{}
	cardLarge := map[int64]*operate.Card{}
	cardLive := map[int64]*operate.Card{}
	cardArticle := map[int64]*operate.Card{}
LOOP:
	for pos, ca := range cards {
		var cardIdx = idx + int64(pos+1)
		if !isrcmd {
			if cardIdx > max && ca.FromType != _fTypeOperation {
				continue
			}
			if plat != model.PlatH5 { // h5 gives all operation cards
				tmpPlat := plat
				if mobiApp == "iphone" && dev == "pad" {
					tmpPlat = model.PlatIPhone
				}
				if config, ok := ca.PopularCardPlat[tmpPlat]; ok {
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
		switch ca.Type {
		case model.GotoAv, model.GotoAd:
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
		case model.GotoUpRcmdNew, model.GotoUpRcmdNewSingle:
			cardm, as, upid := s.cardSetChange(c, ca.Value)
			// 国际版过滤
			if model.IsOverseas(plat) {
				if s.MidControl(upid) {
					continue
				}
			}
			aids = append(aids, as...)
			for id, card := range cardm {
				if card.CardGoto == model.GotoUpRcmdNewSingle {
					tmp.Type = model.GotoUpRcmdNewSingle
					// 561之前的版本不展示单视频模式新星卡
					if (model.IsAndroid(plat) && build <= s.c.BuildLimit.StarsSingleAndroid) || (model.IsIPhone(plat) && build <= s.c.BuildLimit.StarsSingleIOS) {
						continue LOOP
					}
					if model.IsIPad(plat) { // ipad 不下发
						continue LOOP
					}
				}
				cardSet[id] = card
			}
			rnUpIDs = append(rnUpIDs, upid)
		case model.GotoEventTopic:
			if plat == model.PlatIPad { // ipad不出事件专题卡
				continue
			}
			eventTopic = s.eventTopicChange(c, plat, ca.Value)
		case model.LargeCardType:
			if plat != model.PlatAndroid && plat != model.PlatIPhone { // 非粉版全部不出大卡
				continue
			}
			cardm, as := s.handleLargeCard(c, ca.Value)
			if as > 0 {
				aids = append(aids, as)
			}
			for id, card := range cardm {
				cardLarge[id] = card
			}
		case model.LiveCardType:
			if plat == model.PlatIPad { // ipad不出直播小卡
				continue
			}
			cardm, roomId := s.handleLiveCard(c, ca.Value)
			if roomId > 0 {
				lives = append(lives, roomId)
			}
			for id, card := range cardm {
				cardLive[id] = card
			}
		case model.GotoReadCard:
			if plat == model.PlatIPad { // ipad不出专栏
				continue
			}
			cardm, articleID := s.handleArticleCard(c, ca.Value)
			if articleID == 0 || cardm == nil {
				continue
			}
			artIDs = append(artIDs, articleID)
			cardArticle[ca.Value] = cardm
		}
		feedcards = append(feedcards, tmp)
		if len(feedcards) == ps && !isrcmd {
			break
		}
	}
	if len(topAids) > 0 && len(feedcards) > 0 && !isrcmd { //处理热门定位，把对应aid放到置顶卡下面，且当前不是AI返回的才处理此逻辑
		feedcards = s.dealTopCard(feedcards, idx, topAids)
		for k := range topAids {
			aids = append(aids, k)
		}
	}
	g := errgroup.WithContext(c)
	if len(aids) != 0 {
		g.Go(func(ctx context.Context) error {
			innerArc = s.controld.CircleReqInternalAttr(ctx, aids)
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			var aidsReq []*arcgrpc.PlayAv
			//没有cid，只需处理aid
			for _, v := range aids {
				aidsReq = append(aidsReq, &arcgrpc.PlayAv{Aid: v})
			}
			if amplayer, err = s.arc.ArcsPlayer(ctx, aidsReq); err != nil {
				log.Error("%+v", err)
				return
			}
			// 国际版mid控制
			if model.IsOverseas(plat) {
				for aid, aVal := range amplayer {
					a := aVal.GetArc()
					if a == nil {
						continue
					}
					// 命中 剔除对应稿件
					if s.MidControl(a.Author.Mid) {
						delete(amplayer, aid)
					}
					// 简转繁
					if isHant {
						out := chinese.Converts(ctx, a.Title, a.Desc)
						a.Title = out[a.Title]
						a.Desc = out[a.Desc]
					}
				}
			}
			for _, a := range amplayer {
				if a == nil || a.Arc == nil {
					continue
				}
				avUpIDs = append(avUpIDs, a.Arc.Author.Mid)
			}
			return
		})
	}
	if len(lives) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if liveplayer, err = s.lv.EntryRoomInfo(ctx, lives, mid); err != nil {
				log.Error("[dealItem3] s.lv.GetMultiple() error(%v)", err)
			}
			for key, live := range liveplayer {
				// 国际版mid控制
				if model.IsOverseas(plat) {
					if s.MidControl(live.Uid) {
						delete(amplayer, key)
					}
				}
				avUpIDs = append(avUpIDs, live.Uid)
			}
			return nil
		})
	}
	if len(artIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if metam, err = s.art.ArticleMetas(ctx, artIDs, _articlesFrom); err != nil {
				log.Error("%+v", err)
			}
			for key, meta := range metam {
				if meta.Author != nil {
					// 国际版mid控制
					if model.IsOverseas(plat) {
						if s.MidControl(meta.Author.Mid) {
							delete(amplayer, key)
						}
					}
					avUpIDs = append(avUpIDs, meta.Author.Mid)
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("dealItem3 errgroup error(%+v)", err)
	}
	switch style {
	case cdm.HotCardStyleShowUp, cdm.HotCardStyleHideUp:
		upIDs = append(upIDs, avUpIDs...)
	}
	upIDs = append(upIDs, rnUpIDs...)
	avUpIDs = append(avUpIDs, rnUpIDs...)
	g = errgroup.WithContext(c)
	if len(avUpIDs) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if accountm, err = s.acc.Cards3GRPC(ctx, avUpIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(upIDs) > 0 {
		g.Go(func(ctx context.Context) (err error) {
			if statm, err = s.reldao.StatsGRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		if mid != 0 {
			g.Go(func(ctx context.Context) (err error) {
				if authorRelations, err = s.reldao.RelationsInterrelations(ctx, mid, upIDs); err != nil {
					log.Error("%+v", err)
				}
				return nil
			})
			g.Go(func(ctx context.Context) error {
				isAtten = s.acc.IsAttentionGRPC(ctx, upIDs, mid)
				return nil
			})
		}
	}
	if len(hotIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if aggRes, err = s.dao.Aggregations(ctx, hotIDs); err != nil {
				log.Error("[dealItem2] s.agg.Aggregation() error(%v)", err)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("dealItem3 g.Wait() error(%v)", err)
	}
	for _, ca := range feedcards {
		var (
			r        = ca.PopularCardToAiChange()
			main     interface{}
			cardType cdm.CardType
			op       = &operate.Card{TrackID: ca.TrackID}
		)
		if (model.IsAndroid(plat) && build >= s.c.BuildLimit.HotCardOptimizeAndroid) ||
			(model.IsIPhone(plat) && build >= s.c.BuildLimit.HotCardOptimizeIPhone) ||
			(model.IsIPad(plat) && build >= s.c.BuildLimit.HotCardOptimizeIPad) {
			switch r.CornerMark {
			case 0:
				r.CornerMark = 6
			case 1:
				r.CornerMark = 7
			}
		}
		// 国际版 推荐理由简转繁
		if model.IsOverseas(plat) {
			if isHant && r.RcmdReason != nil {
				r.RcmdReason.Content = chinese.Convert(c, r.RcmdReason.Content)
			}
		}
		op.From(cdm.CardGt(r.Goto), r.ID, 0, plat, build, mobiApp)
		r.Style = style
		switch r.Style {
		case cdm.HotCardStyleShowUp, cdm.HotCardStyleHideUp:
			switch r.Goto {
			case model.GotoAv, model.LiveCardType:
				cardType = cdm.SmallCoverV5
			case model.LargeCardType:
				cardType = cdm.LargeCoverV4
			case model.GotoUpRcmdNewSingle:
				cardType = cdm.RcmdOneItem
			case model.GotoAd:
				cardType = cdm.SmallCoverV5Ad
			}
		}
		switch r.Goto {
		case model.GotoAv, model.GotoAd:
			var isOverseas bool
			if innerArc != nil && innerArc[r.ID] != nil {
				isOverseas = innerArc[r.ID].OverSeaBlock
			}
			if a, ok := amplayer[r.ID]; ok && a != nil && a.Arc != nil && (!isOverseas || !model.IsOverseas(plat)) {
				main = map[int64]*arcgrpc.ArcPlayer{a.Arc.Aid: a}
				r.HideButton = true
				if op.TrackID == "" {
					op.TrackID = trackID
				}
				if cardType == cdm.SmallCoverV5 || cardType == cdm.SmallCoverV5Ad {
					op.Switch = cdm.SwitchCooperationShow
				} else {
					op.Switch = cdm.SwitchCooperationHide
				}
			}
			if res, ok := aggRes[ca.HotwordID]; ok && res.State == _auditingPass {
				op.ShowHotword = true
				op.Tid = res.ID
				if (plat == model.PlatAndroid && build >= s.c.BuildLimit.SVideoAndroid) || (plat == model.PlatIPhone && build > s.c.BuildLimit.SVideoIOS) {
					op.Cover = s.c.Aggregation.IconV2
					op.Subtitle = "热点"
					op.SvideoShow = true
				} else {
					op.Cover = s.c.Aggregation.Icon
					op.Subtitle = res.Title
				}
				if s.displaySVideo(buvid, build, plat) {
					// op.RedirectURL = fmt.Sprintf(_uriSVideo, _svideoAggr, res.ID)
					op.RedirectURL = buildSVideoURI(_svideoAggr, res.ID, r.ID)
				} else {
					op.RedirectURL = agg.ToResH5URl(res.ID)
				}
			}
			op.Share = s.c.Share
		case model.GotoRank:
			ams := map[int64]*arcgrpc.ArcPlayer{}
			for aid, a := range s.rankArchivesCache {
				ams[aid] = &arcgrpc.ArcPlayer{Arc: a}
			}
			main = map[cdm.Gt]interface{}{cdm.GotoAv: ams}
			op = rank
		case model.GotoUpRcmdNew, model.GotoUpRcmdNewSingle:
			main = amplayer
			op = cardSet[r.ID]
			if op != nil {
				op.Plat = plat // 通过op将plat传入，便于对ipad进行新人卡视频数量过滤
				op.Share = s.c.Share
				op.BuildLimit = &operate.BuildLimit{
					IsAndroid:              model.IsAndroid(plat),
					IsIphone:               model.IsIPhone(plat),
					IsIPad:                 model.IsIPad(plat),
					HotCardOptimizeAndroid: s.c.BuildLimit.HotCardOptimizeAndroid,
					HotCardOptimizeIPhone:  s.c.BuildLimit.HotCardOptimizeIPhone,
					HotCardOptimizeIPad:    s.c.BuildLimit.HotCardOptimizeIPad,
				}
			}
		case model.GotoEventTopic:
			op = eventTopic[r.ID]
		case model.LargeCardType:
			main = amplayer
			op = cardLarge[r.ID]
		case model.LiveCardType:
			main = liveplayer
			op = cardLive[r.ID]
			if op != nil {
				op.Plat = plat
				op.Share = s.c.LiveShare
			}
		case model.GotoReadCard:
			main = metam
			op = cardArticle[r.ID]
			if op != nil {
				op.Share = s.c.Share
			}
		}
		h := cardmv2.Handle(plat, cdm.CardGt(r.Goto), cardType, cdm.ColumnSvrSingle, r, nil, isAtten, nil, statm, accountm, authorRelations, adBizData.ToCardAdInfo())
		if h == nil {
			continue
		}
		op.FromDev(mobiApp, plat, build)
		h.From(main, op)
		h.Get().FromType = ca.FromType
		h.Get().Idx = ca.Idx
		if h.Get().Right && !shouldDisableMidMaxIn32Card(h.Get(), device.Device{RawMobiApp: mobiApp, Build: int64(build)}) {
			if h.Get().ThreePointV4 == nil {
				h.Get().ThreePointWatchLater(op)
			}
			is = append(is, cardmv2.AddCard(h))
			if r.Goto == model.LargeCardType || r.Goto == model.GotoEventTopic {
				hasLargeCard = true
			}
		}
		// infoc
		tinfo := &feed.Item{
			Goto:       ca.Type,
			Param:      strconv.FormatInt(ca.Value, 10),
			URI:        h.Get().Uri,
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
		switch tinfo.Goto {
		case model.GotoUpRcmdNewSingle:
			tinfo.Goto = model.GotoUpRcmdNewV2
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
			case model.GotoReadCard:
				tinfo.Param = op.Param
			case model.LiveCardType, model.LargeCardType: // 直播小卡、视频大卡 上报的Param为房间号
				tinfo.Param = op.SubParam
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
		is = _emptyList3
		return
	}
	return
}

func shouldDisableMidMaxIn32Card(card *cardmv2.Card, dev device.Device) bool {
	if !cdmc.CheckMidMaxInt32Version(dev) {
		return false
	}
	mid := card.GetThreePointV4().GetSharePlane().GetAuthorId()
	if card.CardType == cdm.LargeCoverV4 {
		mid = card.GetArgs().GetUpId()
	}
	return cdmc.CheckMidMaxInt32(mid)
}

func (s *Service) dealEntrance(items []*api.EntranceShow, plat int8, hasLargeCard bool, originList []*cardapi.Card) (res []*cardapi.Card) {
	if len(originList) == 0 {
		return
	}
	var (
		main     interface{}
		cardType = cdm.CardType(model.GotoPopularTopEntrance)
		idx      int
	)
	op := &operate.Card{Desc: s.c.ShowHotConfig.ItemTitle}
	for _, item := range items {
		if item.Bubble == nil {
			item.Bubble = new(api.Bubble)
		}
		op.EntranceItems = append(op.EntranceItems, &cdm.EntranceItem{
			Goto:         model.GotoEntrances,
			Icon:         item.Icon,
			Title:        item.Title,
			ModuleId:     item.ModuleId,
			Uri:          item.Uri,
			EntranceId:   item.EntranceId,
			EntranceType: item.EntranceType,
			Bubble: &cdm.Bubble{
				BubbleContent: item.Bubble.BubbleContent,
				Version:       item.Bubble.Version,
				Stime:         item.Bubble.Stime,
			},
		})
	}
	h := cardmv2.Handle(plat, model.GotoEntrances, cardType, cdm.ColumnSvrSingle, nil, nil, nil, nil, nil, nil, nil, nil)
	if h != nil {
		h.From(main, op)
		h.Get().FromType = model.GotoPopularTopEntrance
		if hasLargeCard {
			idx = 1
		} else {
			idx = 0
		}
		h.Get().CardType = model.GotoPopularTopEntrance
		h.Get().Goto = model.GotoWeb
		for i, list := range originList {
			if i == idx {
				res = append(res, cardmv2.AddCard(h))
			}
			res = append(res, list)
		}
	}
	if len(res) == 0 {
		return originList
	}
	return
}

// nolint:gocognit
func (s *Service) dealLargeCardAndEventTopic(originCards []*card.PopularCard, key int, mid int64) (res []*card.PopularCard, hasLargeCard bool) {
	log.Info("dealLargeCard key(%d) mid(%d)", key, mid)
	if len(originCards) == 0 {
		return
	}
	func() {
		for i := 0; i < len(originCards); i++ {
			switch originCards[i].Type {
			case model.LargeCardType:
				for _, item := range s.largeCards {
					if item.ID == originCards[i].Value && item.Sticky == 1 {
						if item.WhiteList != "" {
							item.WhiteList = fmt.Sprintf(",%s,", item.WhiteList)
							if key == _specialKey && !strings.Contains(item.WhiteList, fmt.Sprintf(",%d,", mid)) {
								log.Info("dealLargeCard item(%+v)", item)
								continue
							}
						}
						if i != 0 {
							res = append([]*card.PopularCard{originCards[i]}, originCards[0:i-1]...)
							res = append(res, originCards[i:]...)
						}
						hasLargeCard = true
						return
					}
				}
			case model.GotoEventTopic:
				if item, ok := s.eventTopicCache[originCards[i].Value]; ok {
					if item.Sticky == 1 {
						if i != 0 {
							res = append([]*card.PopularCard{originCards[i]}, originCards[0:i-1]...)
							res = append(res, originCards[i:]...)
						}
						hasLargeCard = true
						return
					}
				}
			}
		}
	}()
	if len(res) == 0 {
		return originCards, hasLargeCard
	}
	return
}

func (s *Service) dealTopCard(originCards []*card.PopularCard, idx int64, aids map[int64]struct{}) (res []*card.PopularCard) {
	if len(originCards) <= 1 {
		return originCards
	}
	if originCards[0].Type == model.LargeCardType || originCards[0].Type == model.GotoEventTopic { // 视频大卡和事件卡保持第一位，注意后面有置顶卡(如事件卡)也都需要这里修改
		res = append(res, originCards[0])
	}
	if idx == 0 { // 判断是不是首页，首页需要增加，非首页需要去掉
		for k := range aids {
			res = append(res, &card.PopularCard{
				Type:  model.GotoAv,
				Value: k,
			})
		}
	}
	for index, item := range originCards {
		if (item.Type == model.LargeCardType || item.Type == model.GotoEventTopic) && index == 0 { // 视频大卡和事件卡跳过
			continue
		}
		if item.Type == model.GotoAv {
			if _, ok := aids[item.Value]; ok { // 因为需要置顶的视频小卡已经加到前面，后面只要发现有，直接跳过
				continue
			}
		}
		res = append(res, item)
	}
	for pos, cardTemp := range res {
		cardTemp.Idx = idx + int64(pos+1)
	}
	if len(res) > 0 {
		return res
	}
	return originCards
}

func (s *Service) dealTopItems(items []*api.EntranceShow) (res []*api.EntranceShow) {
	// EntranceType为1代表分品类
	res = append(res, &api.EntranceShow{ // 默认下发全站热门在第一位
		Icon:         s.c.ShowHotConfig.BottomTextCover,
		Title:        _allHotName,
		Uri:          s.c.ShowHotConfig.BottomTextURL,
		EntranceId:   0,
		TopPhoto:     s.middleTopPhoto,
		EntranceType: 1,
	})
	for _, item := range items {
		if item.ModuleId == _hotChannelName || item.ModuleId == _hotTopicName {
			item.EntranceType = 1
			res = append(res, item)
		}
	}
	return res
}

func (s *Service) getTrackID(items []*api.EntranceShow, id int64, source int32) string {
	var idx int
	for i, item := range items {
		if item.EntranceId == id {
			idx = i
		}
	}
	if source == 0 { // 根据来源进行不同的下发
		return fmt.Sprintf("%s.%d", _hotTagName, time.Now().Unix())
	}
	return fmt.Sprintf("%s.%d.%d.%d", _hotPageName, id, idx, time.Now().Unix())
}

func (s *Service) displaySVideo(buvid string, build int, plat int8) bool {
	if s.c.Custom.HotContinuousPlay == _svOff {
		return false
	}
	if buvid == "" {
		return false
	}
	if plat != model.PlatAndroid && plat != model.PlatIPhone {
		return false
	}
	if (plat == model.PlatAndroid && build < s.c.BuildLimit.SVideoAndroid) || (plat == model.PlatIPhone && build <= s.c.BuildLimit.SVideoIOS) {
		return false
	}
	return int64(crc32.ChecksumIEEE([]byte(buvid))%100) < s.c.Custom.HotContinuousGray
}

func (s *Service) MidControl(hmid int64) bool {
	for _, mid := range s.c.Intl.MidControl {
		if mid == hmid {
			return true
		}
	}
	return false
}

func (s *Service) hotRcmd2(c context.Context, idx, mid int64, sourceID int32, plat int8, build int, buvid, mobiApp string, locationIDs []int64, entranceID int64, hotWordID, ps, key int, ad *api.PopularAd) ([]*card.PopularCard, string, *rcmod.BizData, int, bool) {
	var (
		res  []*card.PopularCard
		page int
	)
	if int(idx)%ps != 0 {
		page = int(idx)/ps + 1
	} else {
		page = int(idx) / ps
	}
	var zoneID int64
	locReply, err := s.loc.Info(c, metadata.String(c, metadata.RemoteIP))
	if err == nil {
		zoneID = locReply.ZoneId
	} else {
		log.Error("s.loc.Info error(%+v), mid(%d)", err, mid)
	}
	hotrcmd, userfeature, bizData, resCode, err := s.rcmmnd.HotAiRcmd(c, mid, sourceID, plat, build, buvid, mobiApp, page, ps, entranceID, hotWordID, locationIDs, ad, zoneID)
	if err != nil {
		log.Error("日志报警 ai hot rcmd error(%v)", err)
		// 服务error用老数据
		return s.PopularCardTenList(c, key, int(idx), ps), userfeature, nil, resCode, false
	}
	if len(hotrcmd) == 0 {
		return nil, userfeature, nil, resCode, true
	}
	for _, v := range hotrcmd {
		if v == nil {
			continue
		}
		res = append(res, v.HotItemChange())
	}
	return res, userfeature, bizData, resCode, true
}

// nolint:gomnd
func (s *Service) AIGroup(c context.Context, mid int64, buvid string) bool {
	// 强制回滚到老得redis逻辑上
	if s.c.Custom.AIHotAbnormal {
		return false
	}
	// 如果mid为0且buvid为空，这类请求直接降级到redis逻辑上
	if mid == 0 && buvid == "" {
		return false
	}
	if _, ok := s.c.Custom.AIHotMid[strconv.FormatInt(mid, 10)]; !ok {
		if mid > 0 {
			group := int((mid / 100) % 100)
			if _, ok := s.c.Custom.AIGroupMid[strconv.Itoa(group)]; !ok {
				return false
			}
		} else {
			group := int((crc32.ChecksumIEEE([]byte(buvid)) / 100) % 100)
			if _, ok := s.c.Custom.AIGroupBuvid[strconv.Itoa(group)]; !ok {
				return false
			}
		}
	}
	return true
}

//nolint:gomnd
func (s *Service) PurePopularArchive(ctx context.Context, req *popularmodel.PopularArchiveRequest) (*popularmodel.PopularArchiveReply, error) {
	dev, _ := device.FromContext(ctx)
	authN, _ := auth.FromContext(ctx)
	key := int((crc32.ChecksumIEEE([]byte(dev.Buvid)) / 1000) % 10)
	if authN.Mid > 0 {
		key = int((authN.Mid / 1000) % 10)
	}
	cards, _, _, _, _ := s.hotRcmd2(ctx, 0, authN.Mid, 0, dev.Plat(), int(dev.Build), dev.Buvid, dev.RawMobiApp, nil, 0, 0, 100, key, nil)
	if len(cards) <= 0 {
		return &popularmodel.PopularArchiveReply{}, nil
	}

	aids := make([]int64, 0, len(cards))
	for _, c := range cards {
		switch c.Type {
		case model.GotoAv:
			aids = append(aids, c.Value)
		}
	}
	archives, err := s.arc.ArchivesPB(ctx, aids, authN.Mid, dev.RawMobiApp, dev.Device)
	if err != nil {
		return nil, err
	}

	reply := &popularmodel.PopularArchiveReply{}
	for _, c := range cards {
		switch c.Type {
		case model.GotoAv:
			arc, ok := archives[c.Value]
			if !ok {
				continue
			}
			reply.List = append(reply.List, &popularmodel.PopularArchiveItem{
				Aid:      arc.Aid,
				Title:    arc.Title,
				Cover:    arc.Pic,
				URI:      cdm.FillURI(cdm.GotoAv, 0, 0, strconv.FormatInt(arc.Aid, 10), nil),
				Play:     int64(arc.Stat.View),
				Danmaku:  int64(arc.Stat.Danmaku),
				Duration: cdm.DurationString(arc.Duration),
			})
		}
	}
	return reply, nil
}
