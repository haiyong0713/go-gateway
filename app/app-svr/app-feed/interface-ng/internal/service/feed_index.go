package service

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sync"
	"sync/atomic"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	feedcard "go-gateway/app/app-svr/app-feed/interface-ng/feed-card"
	"go-gateway/app/app-svr/app-feed/interface-ng/internal/model"
	feedmdl "go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/api/session"

	"go-common/library/sync/errgroup.v2"

	"github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
)

var (
	// ai banner
	_aiAdResource = map[int8]int64{
		feedmdl.PlatIPhone:   1685,
		feedmdl.PlatIPhoneB:  1685,
		feedmdl.PlatAndroid:  1690,
		feedmdl.PlatAndroidB: 1690,
		feedmdl.PlatIPad:     1974,
		feedmdl.PlatIPadHD:   1974,
	}
)

func (s *Service) fanoutLoad(ctx context.Context, loader *feedFanoutLoader) (*feedcard.FanoutResult, error) {
	// account subset should be load later
	accountSubset := loader.Account
	loader.Account = loaderAccountSubset{}

	if feedcard.UsingChannelAsTag(loader.Device) {
		loader.WithChannel(loader.TagID...)
	}
	loader.WithBannerResourceID()
	fanoutResult, err := loader.Load(ctx, s.dao)
	if err != nil {
		return nil, err
	}
	secondLoader := loader.DeriveEmpty()
	secondLoader.Account = accountSubset
	for _, a := range fanoutResult.Archive.Archive {
		secondLoader.WithAccountProfile(a.Arc.Author.Mid)
	}
	for _, a := range fanoutResult.Archive.StoryArchive {
		secondLoader.WithAccountProfile(a.Arc.Author.Mid)
	}
	for _, r := range fanoutResult.Live.Room {
		secondLoader.WithAccountProfile(r.UID)
	}
	for _, a := range fanoutResult.Article {
		secondLoader.WithAccountProfile(a.Author.Mid)
	}
	for _, r := range fanoutResult.Live.InlineRoom {
		secondLoader.WithAccountProfile(r.UID)
	}
	for _, episodeCard := range fanoutResult.Bangumi.InlinePGC {
		secondLoader.WithThumbUpArchive(episodeCard.Aid)
	}
	secondLoaderResult, err := secondLoader.Load(ctx, s.dao)
	if err != nil {
		return nil, err
	}
	// concat fanout result
	fanoutResult.Account = secondLoaderResult.Account
	fanoutResult.ThumbUp.HasLikeArchive = constructThumbup(fanoutResult.ThumbUp.HasLikeArchive, secondLoaderResult.ThumbUp.HasLikeArchive)
	fanoutResult.HotAidSet = s.HotAidSetFunc()
	return fanoutResult, nil
}

func constructThumbup(archive map[int64]int8, pgc map[int64]int8) map[int64]int8 {
	out := make(map[int64]int8, len(archive)+len(pgc))
	for aid, status := range archive {
		out[aid] = status
	}
	for aid, status := range pgc {
		out[aid] = status
	}
	return out
}

// ValidateSession is
func (s *Service) ValidateSession(ctx context.Context, session *session.IndexSession) ([]cardschema.FeedCard, error) {
	param, device, err := parseFeedParam(session)
	if err != nil {
		return nil, err
	}

	aiResponse := &feed.AIResponse{}
	if err := json.Unmarshal([]byte(session.AIRecommendResponse), aiResponse); err != nil {
		return nil, errors.Wrap(err, "ai recommend response")
	}
	setAIInternalAd(aiResponse)
	//nolint:gomnd
	gifType := atomic.AddUint64(&RequestCount, 1) % 2
	loader := defaultFanoutPreProcesser.processRcmd(aiResponse.Items...)
	loader.fanoutCommon.CurrentMid = session.Mid
	loader.fanoutCommon.Device = device
	loader.fanoutCommon.FeedParam = *param
	loader.fanoutCommon.PreferGIFType = convertPreferGIFType(gifType)
	fanoutResult, err := s.fanoutLoad(ctx, loader)
	if err != nil {
		return nil, err
	}
	allocateGIFPermission(aiResponse.Items, fanoutResult, loader)
	output := make([]cardschema.FeedCard, 0, len(session.AIRecommendResponse))
	userSession := feedcard.NewUserSession(session.Mid, fanoutResult.Account.IsAttention, param)
	fCtx := feedcard.NewFeedContext(userSession, device, time.Now())
	if aiResponse.DislikeExp == 1 {
		fCtx.FeatureGates().EnableFeature(cardschema.FeatureNewDislike)
	}

	now := time.Now()
	for i, item := range aiResponse.Items {
		item.SetRequestAt(now)
		item.SetGotoStoryDislikeReason(s.customConfig.GotoStoryDislikeReason)
		builder, ok := GlobalCardBuilderResolver.getBuilder(fCtx, item.Goto)
		if !ok {
			log.Error("Unsupported ai goto: %q", item.Goto)
			continue
		}
		cardOutput, err := builder.Build(fCtx, int64(i), item, fanoutResult)
		if err != nil {
			log.Error("Failed to build card output: %+v", err)
			continue
		}
		output = append(output, cardOutput)
	}
	setFinallyIdx(fCtx, output)
	return output, nil
}

func setAIInternalAd(rs *feed.AIResponse) {
	if len(rs.Items) == 0 {
		return
	}
	adIndexMap := constructAdIndexMap(rs)
	setAIInternalStatusAndID(rs, adIndexMap)
}

func (s *Service) constructAdIndexMapFromCm(ctx context.Context, gifType uint64, param *model.ConstructCardParam) map[int32][]*cm.AdInfo {
	resource := adResource(param.IndexNgReq.Device)
	if resource == 0 {
		return nil
	}
	var country, province, city string
	if param.Zone != nil {
		country = param.Zone.Country
		province = param.Zone.Province
		city = param.Zone.City
	}
	advert, err := s.dao.Ad().Ad(ctx, &model.AdReq{
		Mid:          param.IndexNgReq.Mid,
		Buvid:        param.IndexNgReq.Device.Buvid(),
		Build:        param.IndexNgReq.FeedParam.Build,
		Resource:     []int64{resource},
		Country:      country,
		Province:     province,
		City:         city,
		Network:      param.IndexNgReq.FeedParam.Network,
		MobiApp:      param.IndexNgReq.FeedParam.MobiApp,
		Device:       param.IndexNgReq.FeedParam.Device,
		OpenEvent:    param.IndexNgReq.FeedParam.OpenEvent,
		AdExtra:      param.IndexNgReq.FeedParam.AdExtra,
		Style:        param.IndexNgReq.Style,
		MayResistGif: int(gifType),
	})
	if err != nil {
		log.Error("Failed to request ad: %+v", err)
		return nil
	}
	return cm.ConstructAdIndexMapFrom(resource, advert)
}

func setAIInternalStatusAndID(rs *feed.AIResponse, adIndexMap map[int32][]*cm.AdInfo) {
	for _, r := range rs.Items {
		r.SetManualInline(rs.ManualInline)
		if !AiAdSet.Has(r.Goto) {
			continue
		}
		ads, ok := adIndexMap[r.BizIdx]
		if !ok || len(ads) == 0 {
			continue
		}
		ad := ads[0]
		if ad.CreativeID == 0 {
			log.Warn("ad creative id is nil: %+v", ad)
			continue
		}
		r.SetCardStatusAd(ad)
		if r.Goto == feedmdl.GotoAdAv && ad.CreativeContent != nil {
			r.ID = ad.CreativeContent.VideoID
		}
	}
}

func allocateGIFPermission(items []*ai.Item, _ *feedcard.FanoutResult, loader *feedFanoutLoader) {
	gifAlreadyStateSet := sets.String{}
	for _, r := range items {
		if archiveGotoTypeSet.Has(r.Goto) && r.CoverGif != "" {
			gifAlreadyStateSet.Insert("ai_gif_already")
		}
		if potentialAdGifSet.Has(r.Goto) && isAdGIFStyle(r) {
			gifAlreadyStateSet.Insert("ad_gif_already")
			if loader.fanoutCommon.PreferGIFType == PreferGIFTypeAdvertisement {
				r.SetAllowGIF()
				r.SetDynamicCoverInfoc(constructAdDynamicCover(r.CardStatusAd().CreativeStyle))
				continue
			}
			if directAdGotoTypeSet.Has(r.Goto) {
				if r.CardStatusAd().CardIndex <= _delAdIndex && hasBanner(items) && DelAdCardSet.Has(int(r.CardStatusAd().CardType)) {
					r.SetAdPKCode(_adPkBigCard)
				}
				continue
			}
			r.SetAdPKCode(_adPkGifCard)
		}
	}
	for _, r := range items {
		if archiveGotoTypeSet.Has(r.Goto) && r.CoverGif != "" &&
			!gifAlreadyStateSet.HasAny("ad_gif_already", "rcmd_gif_already") {
			r.SetAllowGIF()
			r.SetDynamicCoverInfoc(_dynamicCoverAiGif)
		}
	}
}

func constructAdDynamicCover(style int32) int32 {
	//nolint:gomnd
	switch style {
	case 2:
		return _dynamicCoverAdGif
	case 4:
		return _dynamicCoverAdInline
	default:
		return 0
	}
}

func hasBanner(items []*ai.Item) bool {
	for _, item := range items {
		if item.Goto == feedmdl.GotoBanner {
			return true
		}
	}
	return false
}

var (
	AiAdSet = sets.NewString(feedmdl.GotoAdAv, feedmdl.GotoAdWeb, feedmdl.GotoAdWebS, feedmdl.GotoAdPlayer,
		feedmdl.GotoAdInlineGesture, feedmdl.GotoAdInline360, feedmdl.GotoAdInlineLive, feedmdl.GotoAdWebGif)
	DelAdCardSet = sets.NewInt(2, 7, 20, 26, 27, 41, 42, 43, 44)
	_delAdIndex  = int32(6)
)

func constructAdIndexMap(rs *feed.AIResponse) map[int32][]*cm.AdInfo {
	adIndexMap := map[int32][]*cm.AdInfo{}
	for _, item := range rs.Items {
		if !AiAdSet.Has(item.Goto) {
			continue
		}
		if int(item.BizIdx) >= len(rs.BizData.AdSelected) {
			log.Warn("BizIdx is bigger than the lenth of adselected, item: %+v, adselected: %+v", item, rs.BizData.AdSelected)
			continue
		}
		adInfos := cm.ConstructAdInfos(rs.BizData.AdSelected[item.BizIdx])
		// todo AdAv aid
		adIndexMap[item.BizIdx] = adInfos
	}
	return adIndexMap
}

func isAdGIFStyle(r *ai.Item) bool {
	if r.CardStatusAd() == nil {
		return false
	}
	if r.CardStatusAd().CreativeStyle == 2 || r.CardStatusAd().CreativeStyle == 4 { // creative_style: 1 静态图文  2 gif动态图文  3 静态视频  4 inline 广告位播放视频
		return true
	}
	return false
}

func convertPreferGIFType(gifType uint64) string {
	if gifType == 0 {
		return PreferGIFTypeAdvertisement
	}
	return PreferGIFTypeOperator
}

// CompareSession is
func (s *Service) CompareSession(ctx context.Context, session *session.IndexSession) ([]*model.CompareSessionReply, error) {
	_, device, err := parseFeedParam(session)
	if err != nil {
		return nil, err
	}

	sessionResults, err := s.ValidateSession(ctx, session)
	if err != nil {
		return nil, err
	}
	out := []*model.CompareSessionReply{}
	sessionMap, err := makeSessionMap(session.Response)
	if err != nil {
		return nil, err
	}
	var mutex sync.Mutex
	eg := errgroup.WithContext(ctx)
	eg.GOMAXPROCS(15)
	for _, v := range sessionResults {
		v := v
		eg.Go(func(ctx context.Context) error {
			cardType := v.Get().CardType
			cardGoto := v.Get().CardGoto
			gt := v.Get().Goto
			param := v.Get().Param
			if cardType == "storys_v2" || cardType == "storys_v1" {
				//nolint:gosimple
				switch v.(type) {
				case *card.Storys:
					storys := v.(*card.Storys)
					for _, item := range storys.Items {
						diffStr, err := jsonDiff(sessionMap[sessionKey(string(cardType), "vertical_av", "vertical_av", item.Param)], item)
						if err != nil {
							log.Error("Failed to diff: %+v", errors.WithStack(err))
							continue
						}
						csr := &model.CompareSessionReply{
							CardType: string(cardType),
							CardGoto: "vertical_av",
							Goto:     "vertical_av",
							Param:    item.Param,
							MobiApp:  device.MobiApp(),
							Result:   diffStr,
						}
						mutex.Lock()
						out = append(out, csr)
						mutex.Unlock()
					}
				default:
					log.Warn("unexpected result, goto: %s, param: %s", gt, param)
				}
				return nil
			}
			diffStr, err := jsonDiff(sessionMap[sessionKey(string(cardType), string(cardGoto), string(gt), param)], v)
			if err != nil {
				log.Error("Failed to diff: %+v", errors.WithStack(err))
				return err
			}
			csr := &model.CompareSessionReply{
				CardType: string(cardType),
				CardGoto: string(cardGoto),
				Goto:     string(gt),
				Param:    param,
				MobiApp:  device.MobiApp(),
				Result:   diffStr,
			}
			mutex.Lock()
			out = append(out, csr)
			mutex.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

func makeSessionMap(sessionResponse string) (map[string]interface{}, error) {
	sessionArray := []json.RawMessage{}
	if err := json.Unmarshal([]byte(sessionResponse), &sessionArray); err != nil {
		return nil, errors.WithStack(err)
	}
	sessionMap := map[string]interface{}{}
	for _, v := range sessionArray {
		js, err := simplejson.NewJson(v)
		if err != nil {
			log.Error("Failed to new simplejson: %+v", errors.WithStack(err))
			continue
		}
		cardType, _ := js.Get("card_type").String()
		param, _ := js.Get("param").String()
		cardGoto, _ := js.Get("card_goto").String()
		goto_, _ := js.Get("goto").String()
		if cardType == "storys_v2" || cardType == "storys_v1" {
			storys := &card.Storys{}
			if err := json.Unmarshal(v, &storys); err != nil {
				log.Error("Failed to unmarshal storys: %+v", err)
				continue
			}
			for _, item := range storys.Items {
				sessionMap[sessionKey(cardType, "vertical_av", "vertical_av", item.Param)] = item
			}
			continue
		}
		sessionMap[sessionKey(cardType, cardGoto, goto_, param)] = v
	}
	return sessionMap, nil
}

func sessionKey(cardType, cardGoto, gt, param string) string {
	return fmt.Sprintf("%s:%s:%s:%s", cardType, cardGoto, gt, param)
}

func jsonDiff(a, b interface{}) (string, error) {
	aJSON, err := json.Marshal(a)
	if err != nil {
		return "", errors.WithStack(err)
	}
	bJSON, err := json.Marshal(b)
	if err != nil {
		return "", errors.WithStack(err)
	}

	differ := gojsondiff.New()
	d, err := differ.Compare(aJSON, bJSON)
	if err != nil {
		return "", errors.WithStack(err)
	}
	var leftObject map[string]interface{}
	if err := json.Unmarshal(aJSON, &leftObject); err != nil {
		return "", errors.WithStack(err)
	}
	formatter := formatter.NewAsciiFormatter(leftObject,
		formatter.AsciiFormatterConfig{})
	diffString, err := formatter.Format(d)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return diffString, nil
}

func (s *Service) IndexNg(ctx context.Context, req *model.IndexNgReq) (*model.IndexNgReply, error) {
	config := s.indexConfig(req.Device, req)
	newBanner := feedcard.UsingNewBanner(req.Device)
	ip := metadata.String(ctx, metadata.RemoteIP)
	zone, _ := s.dao.Location().InfoGRPC(ctx, ip)
	if config.FollowMode == nil {
		req.FeedParam.RecsysMode = 0
	}
	aiReq := &model.AiReq{
		IndexNgReq:   *req,
		Group:        s.group(req.Mid, req.Device.Buvid()),
		AvAdResource: _aiAdResource[req.Device.Plat()],
		AutoPlay:     fmt.Sprintf("%d|%d", config.AutoplayCard, req.FeedParam.AutoPlayCard),
		NoCache:      req.FeedParam.RecsysMode == 1 || req.FeedParam.TeenagersMode == 1 || req.FeedParam.LessonsMode == 1,
		ResourceID:   s.bannerResourceID(req, newBanner),
		BannerExp:    1, // banner实验已全量
		AdExp:        1,
		Zone:         zone,
		Now:          time.Now(),
	}
	aiResponse, err := s.recommend(ctx, aiReq)
	if err != nil {
		log.Error("Failed to request recommend: %+v", err)
		return nil, err
	}

	config.Interest = s.interestsList(aiResponse.InterestList)
	if aiResponse.AutoRefreshTime >= 1 && aiResponse.AutoRefreshTime <= 21600 {
		config.AutoRefreshTime = aiResponse.AutoRefreshTime
	}
	config.SceneURI = aiResponse.SceneURI
	config.FeedTopClean = aiResponse.FeedTopClean
	config.NoPreload = aiResponse.NoPreload

	constructCardReq := &model.ConstructCardParam{
		IndexNgReq: req,
		AIResponse: aiResponse,
		Zone:       zone,
	}
	items, fanoutResult, err := s.fetchCard(ctx, constructCardReq, config)
	if err != nil {
		log.Error("Failed to construct items: %+v", err)
		return nil, err
	}
	if req.FeedParam.LoginEvent != 0 {
		config.Toast.HasToast = true
		config.Toast.ToastMessage = fmt.Sprintf("发现%d条新内容", len(items))
	}
	s.IndexInfoc(ctx, fanoutResult, aiReq, constructCardReq)
	reply := &model.IndexNgReply{
		Items:  items,
		Config: config,
	}
	return reply, nil
}

var _followMode = &feed.FollowMode{
	Title: "当前为首页推荐 - 关注模式（内测版）",
	Option: []*feed.Option{
		{Title: "通用模式", Desc: "开启后，推荐你可能感兴趣的内容", Value: 0},
		{Title: "关注模式（内测版）", Desc: "开启后，仅显示关注UP主更新的视频", Value: 1},
	},
	ToastMessage: "关注UP主的内容已经看完啦，请稍后再试",
}

func (s *Service) indexConfig(device *feedcard.CtxDevice, req *model.IndexNgReq) *feed.Config {
	config := &feed.Config{
		AutoRefreshTime: int64(time.Duration(s.customConfig.AutoRefreshTime) / time.Second),
		Column:          cdm.Columnm[req.FeedParam.Column],
		FeedCleanAbtest: 0,
		AutoplayCard:    2, //1：自动播放；2：不自动播放
	}
	if s.customConfig.Inline != nil {
		config.ShowInlineDanmaku = s.customConfig.Inline.ShowInlineDanmaku
	}
	if feedcard.IsIOSNewBlue(device) { // ios蓝 2.5之后 默认0、1（实验组：单）、2（实验组：双）直接返回单列、用户的3（单）、4（双）还是正常控制
		switch req.FeedParam.Column {
		case cdm.ColumnDefault, cdm.ColumnSvrSingle, cdm.ColumnSvrDouble:
			config.Column = cdm.ColumnSvrSingle
		default:
			config.Column = cdm.Columnm[req.FeedParam.Column]
		}
	}
	if feedcard.IsIPad(device) {
		config.Column = cdm.ColumnSvrSingle
	}
	if s.customConfig.TransferSwitch && device.RawMobiApp() == "iphone_b" && device.Build() == 8030 && crc32.ChecksumIEEE([]byte(req.Device.Buvid()+"_blueversion"))%20 < 10 {
		config.HomeTransferTest = 1 //_home_transfer_new
	}
	if !feedcard.IsIPad(device) && crc32.ChecksumIEEE([]byte(req.Device.Buvid()+"tianma2.0_autoplay_card"))%100 < 5 {
		//nolint:gomnd
		switch req.FeedParam.AutoPlayCard {
		case 0, 1, 2, 3:
			config.AutoplayCard = 1
		}
	}
	if cdm.Columnm[req.FeedParam.Column] == cdm.ColumnSvrDouble {
		// 6.14 按配置文件内容下发值，默认不下发
		if s.customConfig.Prefer4GAutoPlay && feedcard.CanEnable4GWiFiAutoPlay(req.Device) {
			config.AutoplayCard = 0
		}
		// ios 6.14 遇到 3 返回 2
		if feedcard.IsIOS(req.Device) && req.FeedParam.Build == 10370 && req.FeedParam.AutoPlayCard == 3 {
			config.AutoplayCard = 2
		}
	}
	if req.Mid < 1 {
		return config
	}
	if s.autoplayMidSetFunc(req.Mid) && req.FeedParam.AutoPlayCard != 4 {
		if !feedcard.CanEnable4GWiFiAutoPlay(req.Device) {
			config.AutoplayCard = 1
		}
		if cdm.Columnm[req.FeedParam.Column] == cdm.ColumnSvrSingle {
			config.AutoplayCard = 1
		}
	}
	if cdm.Columnm[req.FeedParam.Column] == cdm.ColumnSvrDouble {
		// 6.14 按配置文件内容下发值，默认不下发
		if s.customConfig.Prefer4GAutoPlay && feedcard.CanEnable4GWiFiAutoPlay(req.Device) {
			config.AutoplayCard = 0
		}
		// ios 6.14 遇到 3 返回 2
		if feedcard.IsIOS(req.Device) && req.FeedParam.Build == 10370 && req.FeedParam.AutoPlayCard == 3 {
			config.AutoplayCard = 2
		}
	}
	if s.followModeSetFunc(req.Mid) {
		tmpConfig := &feed.FollowMode{}
		*tmpConfig = *_followMode
		if req.FeedParam.RecsysMode != 1 {
			tmpConfig.ToastMessage = ""
		}
		config.FollowMode = tmpConfig
	}
	return config
}

func (s *Service) group(mid int64, buvid string) int {
	group := -1
	if mid == 0 && buvid == "" {
		return group
	}
	if mid != 0 {
		if v, ok := s.dispatchMid2GroupFunc(mid); ok {
			group = v
			return group
		}
		group = int(mid % 20)
		return group
	}
	// group = int(crc32.ChecksumIEEE([]byte(buvid)) % 20) 老的buvid实验组逻辑
	// ai新的buvid实验组规则 https://www.tapd.bilibili.co/20095661/prong/stories/view/1120095661001258044
	group = int(crc32.ChecksumIEEE([]byte(fmt.Sprintf("%s_1CF61D5DE42C7852", buvid))) % 4)
	return group
}

var (
	_banners = map[int8]int{
		feedmdl.PlatIPhoneB:  467,
		feedmdl.PlatIPhone:   467,
		feedmdl.PlatAndroid:  631,
		feedmdl.PlatIPad:     771,
		feedmdl.PlatIPhoneI:  947,
		feedmdl.PlatAndroidG: 1285,
		feedmdl.PlatAndroidI: 1707,
		feedmdl.PlatIPadI:    1117,
	}
	// abtest
	_bannersABtest = map[int8]int{
		feedmdl.PlatIPhone:  3143,
		feedmdl.PlatAndroid: 3150,
		feedmdl.PlatIPad:    3179,
	}
	// lesson
	_bannersLesson = map[int8]int{
		feedmdl.PlatIPhone:  3848,
		feedmdl.PlatAndroid: 3852,
		feedmdl.PlatIPad:    3856,
	}
)

func (s *Service) bannerResourceID(req *model.IndexNgReq, newBanner bool) int {
	resourceID := _banners[req.Device.Plat()]
	if req.FeedParam.LessonsMode == 1 {
		resourceID = _bannersLesson[req.Device.Plat()]
		return resourceID
	}
	if newBanner {
		resourceID = _bannersABtest[req.Device.Plat()]
	}
	return resourceID
}

func (s *Service) recommend(ctx context.Context, req *model.AiReq) (*feed.AIResponse, error) {
	indexCount := indexCount(req.IndexNgReq.Device)
	resource := adResource(req.IndexNgReq.Device)
	if req.IndexNgReq.Device.Buvid() == "" && req.IndexNgReq.Mid <= 0 {
		return nil, errors.New("empty mid and buvid")
	}
	var zoneID int64
	if req.Zone != nil {
		zoneID = req.Zone.ZoneId
	}
	recommendReq := &model.RecommendReq{
		Mid:            req.IndexNgReq.Mid,
		Plat:           req.IndexNgReq.Device.Plat(),
		Buvid:          req.IndexNgReq.Device.Buvid(),
		Build:          req.IndexNgReq.FeedParam.Build,
		LoginEvent:     req.IndexNgReq.FeedParam.LoginEvent,
		ParentMode:     req.IndexNgReq.FeedParam.ParentMode,
		RecsysMode:     req.IndexNgReq.FeedParam.RecsysMode,
		TeenagersMode:  req.IndexNgReq.FeedParam.TeenagersMode,
		LessonsMode:    req.IndexNgReq.FeedParam.LessonsMode,
		ZoneID:         zoneID,
		Group:          req.Group,
		Interest:       req.IndexNgReq.FeedParam.Interest,
		Network:        req.IndexNgReq.FeedParam.Network,
		Style:          req.IndexNgReq.Style,
		Column:         req.IndexNgReq.FeedParam.Column,
		Flush:          req.IndexNgReq.FeedParam.Flush,
		IndexCount:     indexCount,
		DeviceType:     req.IndexNgReq.FeedParam.DeviceType,
		AvAdResource:   req.AvAdResource,
		AdResource:     resource,
		AutoPlay:       req.AutoPlay,
		DeviceName:     req.IndexNgReq.FeedParam.DeviceName,
		OpenEvent:      req.IndexNgReq.FeedParam.OpenEvent,
		BannerHash:     req.IndexNgReq.FeedParam.BannerHash,
		AppList:        req.IndexNgReq.AppList,
		DeviceInfo:     req.IndexNgReq.DeviceInfo,
		InterestSelect: req.IndexNgReq.FeedParam.InterestV2,
		ResourceID:     req.ResourceID,
		BannerExp:      req.BannerExp,
		AdExp:          1,
		MobiApp:        req.IndexNgReq.FeedParam.MobiApp,
		AdExtra:        req.IndexNgReq.FeedParam.AdExtra,
		Pull:           req.IndexNgReq.FeedParam.Pull,
		RedPoint:       req.IndexNgReq.FeedParam.RedPoint,
		InlineSound:    req.IndexNgReq.FeedParam.InlineSound,
		InlineDanmu:    req.IndexNgReq.FeedParam.InlineDanmu,
		Now:            req.Now,
	}
	res, err := s.dao.Recommend().Recommend(ctx, recommendReq)
	if err != nil {
		res = &feed.AIResponse{
			RespCode: ecode.ServerErr.Code(),
		}
		if errors.Cause(err) == context.DeadlineExceeded { // ai timeout{
			res.RespCode = ecode.Deadline.Code()
		}
		log.Error("Failed to request ai: %+v", err)
	}
	if req.NoCache || len(res.Items) != 0 {
		res.IsRcmd = true
		return res, nil
	}
	return res, nil
}

func indexCount(dev cardschema.Device) int {
	if feedcard.IsIPad(dev) {
		//nolint:gomnd
		return 20
	}
	//nolint:gomnd
	return 10
}

func adResource(dev cardschema.Device) int64 {
	var cmResourceMap = map[int8]int64{
		1: 1890,
		0: 1897,
		2: 1975,
	}
	if feedcard.IsCmResource(dev) {
		return cmResourceMap[dev.Plat()]
	}
	return 0
}

func (s *Service) fetchCard(ctx context.Context, req *model.ConstructCardParam, config *feed.Config) ([]card.Handler, *feedcard.FanoutResult, error) {
	if specialModeCheck(req.AIResponse, req.IndexNgReq.FeedParam) {
		return []card.Handler{}, &feedcard.FanoutResult{}, nil
	}
	param := convertParam(req.IndexNgReq.FeedParam, req.IndexNgReq.AppList, req.IndexNgReq.DeviceInfo)
	setAIInternalAd(req.AIResponse)
	//nolint:gomnd
	gifType := atomic.AddUint64(&RequestCount, 1) % 2
	if s.customConfig.CmAdSwitch {
		setAIInternalStatusAndID(req.AIResponse, s.constructAdIndexMapFromCm(ctx, gifType, req))
	}
	loader := defaultFanoutPreProcesser.processRcmd(req.AIResponse.Items...)
	loader.fanoutCommon.CurrentMid = req.IndexNgReq.Mid
	loader.fanoutCommon.Device = req.IndexNgReq.Device
	loader.fanoutCommon.FeedParam = *param
	loader.fanoutCommon.PreferGIFType = convertPreferGIFType(gifType)
	fanoutResult, err := s.fanoutLoad(ctx, loader)
	if err != nil {
		return nil, &feedcard.FanoutResult{}, err
	}
	allocateGIFPermission(req.AIResponse.Items, fanoutResult, loader)
	output := make([]cardschema.FeedCard, 0, len(req.AIResponse.Items))
	resolveConfig(config, req.AIResponse, param)
	userSession := feedcard.NewUserSession(req.IndexNgReq.Mid, fanoutResult.Account.IsAttention, param)
	fCtx := feedcard.NewFeedContext(userSession, req.IndexNgReq.Device, time.Now())
	addFeatureGates(fCtx, req)
	now := time.Now()
	for i, item := range req.AIResponse.Items {
		item.SetRequestAt(now)
		item.SetGotoStoryDislikeReason(s.customConfig.GotoStoryDislikeReason)
		builder, ok := GlobalCardBuilderResolver.getBuilder(fCtx, item.Goto)
		if !ok {
			log.Error("Unsupported ai goto: %q", item.Goto)
			continue
		}
		cardOutput, err := builder.Build(fCtx, int64(i), item, fanoutResult)
		if err != nil {
			log.Error("Failed to build card output: %+v", err)
			continue
		}
		output = append(output, cardOutput)
	}
	setFinallyIdx(fCtx, output)
	markAsAdStock(req.AIResponse, output)
	return output, fanoutResult, nil
}

func addFeatureGates(fCtx cardschema.FeedContext, param *model.ConstructCardParam) {
	if param.AIResponse.DislikeExp == 1 {
		fCtx.FeatureGates().EnableFeature(cardschema.FeatureNewDislike)
	}
	if CanSupportSwitchColumnThreePoint(param.AIResponse.SingleGuide) &&
		feedcard.CanEnableSwitchColumn(param.IndexNgReq.Device) {
		fCtx.FeatureGates().EnableFeature(cardschema.FeatureSwitchColumnThreePoint)
	}
}

func resolveConfig(config *feed.Config, ai *feed.AIResponse, param *feedcard.IndexParam) {
	if CanSupportSwitchColumnGuide(ai.SingleGuide) {
		config.SwitchColumnGuidance = &feed.PopupGuidance{
			Title:          "邀你体验「推荐」单列模式",
			SubTitle:       "推荐内容直接看，浏览更方便",
			SourceURL:      "",
			SourceNightURL: "",
			Option: []*feed.GuideOption{
				{Desc: "开启单列模式", Value: cdm.FlagConfirm, Toast: "已成功切换至单列模式 可以在[右下角三点]进行单/双列切换哟~"},
				{Desc: "不了", Value: cdm.FlagCancel},
			},
		}
	}
	if CanResetColumn(ai.SingleGuide) {
		config.NeedResetColumn = true
		config.Column = cdm.ColumnSvrSingle
		param.Column = cdm.ColumnSvrSingle
		if !ai.NewUser {
			config.RecoverColumnGuidance = &feed.PopupGuidance{
				Title:          "已切换至单列模式",
				SubTitle:       "可以在[右下角三点]进行单/双列切换哟~",
				SourceURL:      "",
				SourceNightURL: "",
				Option: []*feed.GuideOption{
					{Desc: "切换至双列", Value: cdm.FlagConfirm, Toast: "已成功切换至双列模式 可以在[右下角三点]进行单/双列切换哟~"},
				},
			}
		}
	}
}

func markAsAdStock(aiResponse *feed.AIResponse, output []card.Handler) {
	if aiResponse.BizData == nil || len(aiResponse.BizData.Stocks) == 0 {
		return
	}
	adInfoMap := make(map[int32][]*cm.AdInfo)
	for _, v := range aiResponse.BizData.Stocks {
		if v == nil {
			continue
		}
		adInfos := cm.ConstructAdInfos(v)
		if len(adInfos) == 0 {
			continue
		}
		adInfoMap[v.CardIndex-1] = adInfos
	}
	for i, h := range output {
		var (
			ads []*cm.AdInfo
			ok  bool
		)
		if ads, ok = adInfoMap[int32(i)]; ok {
			for _, ad := range ads {
				h.Get().AdInfo = ad
				break
			}
		}
		if !ok {
			if h.Get().AdInfo != nil {
				h.Get().AdInfo.CardIndex = int32(i) + 1
			}
		}
	}
}

func convertParam(param feed.IndexParam, appList, deviceInfo string) *feedcard.IndexParam {
	return &feedcard.IndexParam{
		Idx:               param.Idx,
		Pull:              param.Pull,
		Column:            param.Column,
		LoginEvent:        param.LoginEvent,
		OpenEvent:         param.OpenEvent,
		BannerHash:        param.BannerHash,
		AdExtra:           param.AdExtra,
		Interest:          param.Interest,
		Flush:             param.Flush,
		AutoPlayCard:      param.AutoPlayCard,
		DeviceType:        param.DeviceType,
		ParentMode:        param.ParentMode,
		RecsysMode:        param.RecsysMode,
		TeenagersMode:     param.TeenagersMode,
		LessonsMode:       param.LessonsMode,
		DeviceName:        param.DeviceName,
		AccessKey:         param.AccessKey,
		ActionKey:         param.ActionKey,
		Statistics:        param.Statistics,
		Appver:            param.Appver,
		Filtered:          param.Filtered,
		AppKey:            param.AppKey,
		HttpsUrlReq:       param.HttpsUrlReq,
		InterestV2:        param.InterestV2,
		SplashID:          param.SplashID,
		Guidance:          param.Guidance,
		AppList:           appList,
		DeviceInfo:        deviceInfo,
		ColumnTimestamp:   param.ColumnTimestamp,
		AutoplayTimestamp: param.AutoplayTimestamp,
	}
}

func specialModeCheck(rs *feed.AIResponse, param feed.IndexParam) bool {
	flag := false
	var item []*ai.Item
	if param.RecsysMode == 1 || param.TeenagersMode == 1 || param.LessonsMode == 1 {
		for _, r := range rs.Items {
			if r.Goto == feedmdl.GotoBanner {
				continue
			}
			if (param.TeenagersMode == 1 || param.LessonsMode == 1) && r.Goto != feedmdl.GotoAv {
				continue
			}
			item = append(item, r)
		}
		if len(item) == 0 {
			flag = true
		}
	}
	return flag
}

func (s *Service) interestsList(interestList []*ai.Interest) *feed.Interest {
	if len(interestList) <= 0 {
		return nil
	}
	interests := &feed.Interest{
		TitleHide: "选择感兴趣内容",
		DescHide:  "获得更精准的首页推荐",
		TitleShow: "选择感兴趣的内容",
		DescShow:  "选择后首页推荐会更精彩哦",
		Message:   "已根据你的兴趣，推荐相关内容",
	}
	for _, interest := range interestList {
		if interest == nil || interest.Text == "" || interest.CateID == 0 {
			continue
		}
		tmp := &feed.InterestItem{
			ID:    interest.CateID,
			Title: interest.Text,
		}
		for _, it := range interest.Items {
			if it == nil || it.Text == "" || it.SubCateID == 0 {
				continue
			}
			item := &feed.InterestItem{
				ID:    it.SubCateID,
				Title: it.Text,
			}
			tmp.Option = append(tmp.Option, item)
		}
		if len(tmp.Option) == 0 {
			continue
		}
		interests.Items = append(interests.Items, tmp)
	}
	if len(interests.Items) > 0 {
		return interests
	}
	return nil
}

func CanSupportSwitchColumnThreePoint(flag int64) bool {
	return (flag>>feed.FlagBitCanSupportThreePoint)&1 == feed.FlagYes
}

func CanSupportSwitchColumnGuide(flag int64) bool {
	return (flag>>feed.FlagBitCanSupportGuide)&1 == feed.FlagYes
}

func CanResetColumn(flag int64) bool {
	return (flag>>feed.FlagBitCanResetColumn)&1 == feed.FlagYes
}
