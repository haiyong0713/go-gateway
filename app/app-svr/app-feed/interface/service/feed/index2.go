package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/exp/ab"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/bplus"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/ai"
	"go-gateway/app/app-svr/app-card/interface/model/card/audio"
	"go-gateway/app/app-svr/app-card/interface/model/card/bangumi"
	"go-gateway/app/app-svr/app-card/interface/model/card/banner"
	"go-gateway/app/app-svr/app-card/interface/model/card/cm"
	"go-gateway/app/app-svr/app-card/interface/model/card/game"
	"go-gateway/app/app-svr/app-card/interface/model/card/live"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	"go-gateway/app/app-svr/app-card/interface/model/stat"
	cardschema "go-gateway/app/app-svr/app-feed/interface-ng/card-schema"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	"go-gateway/app/app-svr/app-feed/interface/common"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	"go-gateway/app/app-svr/app-feed/interface/model"
	"go-gateway/app/app-svr/app-feed/interface/model/feed"
	stat2 "go-gateway/app/app-svr/app-feed/interface/model/stat"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/api/session"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	"go-gateway/pkg/adresource"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	feedMgr "git.bilibili.co/bapis/bapis-go/ai/feed/mgr/service"
	article "git.bilibili.co/bapis/bapis-go/article/model"
	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	taggrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumbupgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	deliverygrpc "git.bilibili.co/bapis/bapis-go/pgc/servant/delivery"
	pgccard "git.bilibili.co/bapis/bapis-go/pgc/service/card"
	pgcAppGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"
	pgcinline "git.bilibili.co/bapis/bapis-go/pgc/service/card/inline"
	episodegrpc "git.bilibili.co/bapis/bapis-go/pgc/service/season/episode"
	tunnelgrpc "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
	resourceV2grpc "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	viprpc "git.bilibili.co/bapis/bapis-go/vip/service"

	"github.com/pkg/errors"
)

const (
	_iosBuild537     = 8330
	_iosNewBlue      = 8090
	_androidBuild537 = 5375000
	_convergeAi      = 100000
	_iosBuild540     = 8470
	_androidBuild540 = 5405000
	// abtest
	_newBannerResource = 2
	_home_transfer_new = 1
	_newThreePoint     = 1
	_newRcmdReason     = 1
	_newRcmdReasonV2   = 2 // 天马新推荐理由样式
	_newAd             = 1 // 是否请求广告新接口
	_newAdBigCard      = 1 // 新广告大卡踢出逻辑
	_showAdGif         = 0 // 优先广告gif 如果(GIF或Inline播放)卡片冲突时, 是否会舍弃广告卡片 1-会, 0-不会
	// abtest
	// gif
	_aiGif   = "ai_gif"
	_adGif   = "ad_gif"
	_rcmdGif = "rcmd_gif"
	// gif
	_bannerCard               = "banner_card"
	_adCard                   = "ad_card"
	_adCardResistReasonGif    = 1 // 避让此广告的原因: 1-GIF/Inline视频避让, 2-热启动Banner下大卡避让
	_adCardResistReasonBanner = 2 // 避让此广告的原因: 1-GIF/Inline视频避让, 2-热启动Banner下大卡避让
	// new user
	_userHideBanner = 1 // 仅不展示banner
	_userHideAd     = 2 // 仅不展示广告
	// dynamic_cover
	_dynamicCoverRcmdGif    = 1 // 运营GIF
	_dynamicCoverAiGif      = 2 // AI GIF
	_dynamicCoverAdGif      = 3 // 广告 GIF
	_dynamicCoverAdInline   = 4 // 广告inline
	_dynamicCoverInlineAv   = 5 // inlineAv
	_dynamicCoverRcmdInline = 6 // 运营Inline
	// ad pk code
	_adPkGifCard = "gif"
	_adPkBigCard = "banner"
	// ai banner
	_aiBannerExp = 1
	// infoc code
	_recsysMode       = 78050
	_recsysModeMsg    = "关注模式"
	_teenagersMode    = 78051
	_teenagersModeMsg = "青少年模式"
	_lessonsMode      = 78052
	_lessonsModeMsg   = "课堂模式"
	_ads              = "A"
	_stock            = "S"
	// abtest 隐藏新手引导 会根据item里面一刷是否有inline卡和gif卡来判断、一刷内有gif或者inline标记当前不能展示新手用户引导
	_hideGuidance             = 1
	_hideGuidanceByAdGif      = 2
	_hideGuidanceByOperateGif = 3
	_hideGuidanceByAIGif      = 4
	_hideGuidanceByInline     = 5
	_hideGuidanceByBugBuild   = 6
	// ai广告
	_aiAdExp = 1

	_newUserInterestPeriod = "0-24"
)

var (
	_cardAdInlineAvm = map[int32]string{
		74: "big",
	}
	_cardAdDynamicm = map[int32]string{ // 起飞新卡，仅双列
		64: "small",
	}
	_cardAdInlineChooseTeam = map[int32]string{ // b站极致说，单双列
		57: "big",
	}
	_cardAdLivem = map[int32]string{
		63: "small",
	}
	_cardAdAvm = map[int32]string{
		1: "small",
	}
	_cardAdWebm = map[int32]string{
		2:  "big",
		7:  "big",
		20: "big",
	}
	_cardAdWebSm = map[int32]string{
		3:  "small",
		26: "small",
	}
	_cardAdPlayerm = map[int32]string{
		27: "big",
	}
	_cardAdInlineGesture = map[int32]string{
		43: "big",
	}
	_cardAdInline360 = map[int32]string{
		42: "big",
	}
	_cardAdInlineLive = map[int32]string{
		44: "big",
	}
	_cardAdWebGif = map[int32]string{
		41: "big",
	}
	_cardAdChoose = map[int32]string{
		54: "big",
	}
	_cardAdPlayerReservation = map[int32]string{
		88: "big",
	}
	_cardAdWebGifReservation = map[int32]string{
		87: "big",
	}
	_cardAdInline3D = map[int32]string{
		100: "big",
	}
	_cardAdInline3DV2 = map[int32]string{
		103: "big",
	}
	_cardAdPgc = map[int32]string{
		97: "small",
	}
	_cardAdInlinePgc = map[int32]string{ //单双列
		98: "big",
	}
	_cardAdColorEgg = map[int32]string{ //单双列
		101: "big",
	}
	_delAdCard = map[int32]struct{}{
		2:   {},
		7:   {},
		20:  {},
		26:  {},
		27:  {},
		41:  {},
		42:  {},
		43:  {},
		44:  {},
		74:  {},
		87:  {},
		88:  {},
		100: {},
		97:  {},
		98:  {},
		101: {},
		103: {},
	}
	_adCardMap = map[int32]string{
		// ad av
		1: "small",
		// ad web
		2:  "big",
		7:  "big",
		20: "big",
		// ad webs
		3:  "small",
		26: "small",
		// ad player
		27: "big",
		41: "big",
		42: "big",
		43: "big",
		44: "big",
		// 	ad choose
		54: "big",
		// AdInlineChooseTeam
		57: "big",
		63: "small",
		// ad dynamic
		64: "small",
		74: "big",
		87: "big",
		88: "big",
		// ad inline 3D
		100: "big",
		// ad pgc
		97: "small",
		// ad inline pgc
		98:  "big",
		101: "big",
		103: "big",
	}
	_followMode = &feed.FollowMode{
		Title: "当前为首页推荐 - 关注模式（内测版）",
		Option: []*feed.Option{
			{Title: "通用模式", Desc: "开启后，推荐你可能感兴趣的内容", Value: 0},
			{Title: "关注模式（内测版）", Desc: "开启后，仅显示关注UP主更新的视频", Value: 1},
		},
		ToastMessage: "关注UP主的内容已经看完啦，请稍后再试",
	}
	inlineGotoSet        = sets.NewString(model.GotoInlineAv, model.GotoInlinePGC, model.GotoInlineLive, model.GotoInlineAvV2, model.GotoInlineBangumi)
	needLikeGoto         = sets.NewString(model.GotoInlineAv, model.GotoInlineAvV2)
	needFavGoto          = sets.NewString(model.GotoInlineAvV2)
	needHideGuidanceGoto = sets.NewString(model.GotoInlineAv, model.GotoInlinePGC, model.GotoInlineLive,
		model.GotoInlineAvV2, model.GotoPlayer, model.GotoPlayerLive, model.GotoPlayerBangumi, model.GotoInlineBangumi)
	needAdReservationGoto     = sets.NewString(model.GotoAdPlayerReservation, model.GotoAdWebGifReservation)
	withoutTunnelMaterialGoto = sets.NewString(model.GotoPGC, model.GotoInlinePGC, model.GotoBangumi, model.GotoGame)
)

//nolint: gocognit
func (s *Service) Index2(c context.Context, buvid string, mid int64, plat int8, param *feed.IndexParam, style int, applist, deviceInfo string, now time.Time) (is []card.Handler, config *feed.Config, infoc *feed.Infoc, info *locgrpc.InfoReply, err error) {
	var (
		rs                         *feed.AIResponse
		adm                        map[int32][]*cm.AdInfo
		advert                     *cm.NewAd
		adAidm, adRoomidm, adEpidm map[int64]struct{}
		banners                    []*banner.Banner
		version                    string
		adInfom, aiAdInfom         map[int32][]*cm.AdInfo
		abtest                     *feed.Abtest
		resourceID                 int
	)
	ip := metadata.String(c, metadata.RemoteIP)
	config = s.indexConfig(c, plat, buvid, mid, param)
	if config.FollowMode == nil {
		param.RecsysMode = 0
	}
	noCache := param.RecsysMode == 1 || param.TeenagersMode == 1 || param.LessonsMode == 1
	followMode := config.FollowMode != nil
	infoc = &feed.Infoc{
		DiscardReason: make(map[int64]*feed.Discard),
	}
	// 后续有变动，单列autoplay会随着ai下发的结果变换为新autoplay逻辑
	infoc.AutoPlayInfoc = fmt.Sprintf("%d|%d", config.AutoplayCard, param.AutoPlayCard)
	if info, err = s.loc.InfoGRPC(c, ip); err != nil {
		log.Warn("s.loc.Info(%v) error(%v)", ip, err)
		err = nil
	}
	group := s.group(mid, buvid)
	// abtest
	abtest = &feed.Abtest{}
	s.initAbtest(abtest, config, param)
	// ai banner
	resourceID, abtest.BannerExp = s.aiBanner(plat, mid, buvid, param, abtest)
	abtest.ResourceID = int64(resourceID)
	// ai广告abtest
	s.aiAd(group, mid, param, abtest)
	if param.PrivacyDisagreeMode == 1 {
		rs = s.fakeRcmdItemsByPrivacyWindow()
		is, infoc.IsRcmd = s.dealItem2(c, mid, buvid, plat, rs.Items, param, true, noCache, followMode, now, abtest, infoc)
		return
	}
	// abtest
	if !s.c.Feed.Index.Abnormal || followMode || param.TeenagersMode == 1 || param.LessonsMode == 1 {
		g, ctx := errgroup.WithContext(c)
		g.Go(func() error {
			rs = s.indexRcmd2(ctx, plat, buvid, mid, param, group, info, style, adAvResource(ctx, plat), infoc.AutoPlayInfoc, noCache, applist, deviceInfo, resourceID, abtest.BannerExp, abtest.AdExp, now, abtest)
			abtest.DislikeExp = rs.DislikeExp
			abtest.ManualInline = rs.ManualInline
			abtest.SingleGuide = rs.SingleGuide
			abtest.RsNewUser = rs.NewUser
			abtest.DislikeText = rs.DislikeText
			abtest.SingleRcmdReason = rs.SingleRcmdReason
			infoc.UserFeature = rs.UserFeature
			infoc.IsRcmd = rs.IsRcmd
			infoc.NewUser = rs.NewUser
			infoc.Code = rs.RespCode
			for index, item := range rs.Items {
				if item == nil {
					FillDiscard(int64(index), "", feed.DiscardReasonNilItem, "", infoc)
					continue
				}
				if item.OgvCreativeId > 0 && item.CreativeId == 0 {
					item.CreativeId = item.OgvCreativeId
				}
				stat.MetricAICardTotal.Inc(stat.BuildRowType(param.Column, plat), item.Goto, item.JumpGoto)
			}
			if i18n.PreferTraditionalChinese(ctx, param.SLocale, param.CLocale) {
				for _, r := range rs.Items {
					if r.RcmdReason != nil {
						i18n.TranslateAsTCV2(&r.RcmdReason.Content)
					}
				}
			}
			return nil
		})
		if param.TeenagersMode == 0 && abtest.NewUser != _userHideBanner && abtest.BannerExp != _aiBannerExp && !s.c.Custom.ResourceDegradeSwitch {
			g.Go(func() (err error) {
				if banners, version, err = s.indexBanner2(ctx, plat, buvid, mid, param, abtest); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				if i18n.PreferTraditionalChinese(ctx, param.SLocale, param.CLocale) {
					for _, b := range banners {
						i18n.TranslateAsTCV2(&b.Title)
					}
				}
				return
			})
		}
		if param.RecsysMode == 0 && param.TeenagersMode == 0 && param.LessonsMode == 0 {
			if abtest.NewUser != _userHideAd && abtest.AdExp != _aiAdExp {
				g.Go(func() (err error) {
					var adStyle = style
					//  兼容老的style逻辑，3为新单列，上报给商业产品的参数定义为：1 单列 2双列
					//nolint:gomnd
					if adStyle == 3 {
						adStyle = 1
					}
					if abtest.IsNewAd == _newAd {
						//nolint:gomnd
						show := atomic.AddUint64(&s.requestCnt, 1) % 2
						atomic.CompareAndSwapUint64(&s.requestCnt, math.MaxUint64, 0)
						abtest.GifType = int(show)
						if adm, adAidm, adRoomidm, advert, infoc.AdCode, err = s.indexAd3(ctx, int(show), plat, buvid, mid, param, info, adStyle, now); err != nil {
							log.Error("%+v", err)
							infoc.AdError = err
							err = nil
						}
					} else {
						if adm, adAidm, adRoomidm, err = s.indexAd2(ctx, plat, buvid, mid, param, info, adStyle, now); err != nil {
							log.Error("%+v", err)
							err = nil
						}
					}
					// 记录广告的card_index
					for _, ads := range adm {
						if len(ads) > 0 {
							isAd := _ads
							if ads[0].AdCb == "" {
								isAd = _stock
							}
							infoc.AdPos = append(infoc.AdPos, strconv.Itoa(int(ads[0].CardIndex))+isAd)
						}
					}
					return
				})
			}
		}
		if param.RecsysMode == 1 {
			infoc.AdCode = _recsysMode
			infoc.AdError = errors.New(_recsysModeMsg)
		}
		if param.TeenagersMode == 1 {
			infoc.AdCode = _teenagersMode
			infoc.AdError = errors.New(_teenagersModeMsg)
		}
		if param.LessonsMode == 1 {
			infoc.AdCode = _lessonsMode
			infoc.AdError = errors.New(_lessonsModeMsg)
		}
		if err = g.Wait(); err != nil {
			log.Error("%+v", err)
			return
		}
		setSessionRecordAIResponse(c, rs)
		if param.RecsysMode == 1 || param.LessonsMode == 1 {
			var tmp []*ai.Item
			for _, r := range rs.Items {
				if r.Goto == model.GotoBanner {
					continue
				}
				if (param.TeenagersMode == 1 || param.LessonsMode == 1) && r.Goto != model.GotoAv {
					continue
				}
				tmp = append(tmp, r)
			}
			if len(tmp) == 0 {
				is = []card.Handler{}
				return
			}
		} else if rs.Ad != nil {
			tmpadm, tmpadAidm, tmpadRoomidm := rs.Ad.AdChange(adAvResource(c, plat), _cardAdAv)
			if len(tmpadm) > 0 && adm == nil {
				adm = map[int32][]*cm.AdInfo{}
			}
			for key, value := range tmpadm {
				adm[key] = value
			}
			if len(tmpadAidm) > 0 && adAidm == nil {
				adAidm = map[int64]struct{}{}
			}
			for key := range tmpadAidm {
				adAidm[key] = struct{}{}
			}
			if len(tmpadRoomidm) > 0 && adRoomidm == nil {
				adRoomidm = map[int64]struct{}{}
			}
			for key := range tmpadRoomidm {
				adRoomidm[key] = struct{}{}
			}
		}
		// ai 广告
		if abtest.AdExp == _aiAdExp && rs != nil {
			adm, adAidm, adRoomidm, adEpidm, aiAdInfom = s.indexAIAd(c, rs, abtest, infoc, adm, adAidm, adRoomidm, adEpidm, buvid, mid, param.IsMelloi)
			infoc.AdCode = rs.AdCode
		}
		rs.Items, adInfom = s.mergeItem2(c, plat, mid, rs.Items, adm, adAidm, adRoomidm, adEpidm, banners, version, followMode, abtest, param, infoc)
		// 如果AI的广告库存大于0直接用AI的广告库存
		if len(aiAdInfom) > 0 {
			adInfom = aiAdInfom
		}
		// interests
		config.Interest = s.interestsList(rs.InterestList)
		if rs.AutoRefreshTime >= 1 && rs.AutoRefreshTime <= 21600 {
			config.AutoRefreshTime = rs.AutoRefreshTime
		}
		config.SceneURI = rs.SceneURI
		config.FeedTopClean = rs.FeedTopClean
		config.NoPreload = rs.NoPreload
		config.TriggerLoadmoreLeftLineNum = rs.TriggerLoadmoreLeftLineNum
		config.AutoRefreshTimeByActive = degradeAutoRefreshTime(rs.AutoRefreshTimeByActive)
		config.AutoRefreshTimeByAppear = degradeAutoRefreshTime(rs.AutoRefreshTimeByAppear)
		config.IsNaviExp = degradeIsNaviExp(param, rs.IsNaviExp)
		config.RefreshTopFirstToast = rs.RefreshTopFirstToast
		config.RefreshTopSecondToast = rs.RefreshTopSecondToast
		config.HistoryCacheSize = rs.HistoryCacheSize
		config.RefreshBarType = rs.RefreshBarType
		config.RefreshOnBack = rs.RefreshOnBack
		config.SmallCoverWhRatio = rs.SmallCoverWhRatio
		config.VideoMode = rs.VideoMode
		config.TopRefreshLatestExp = rs.TopRefreshLatestExp
		config.VisibleArea = visibleArea(rs.ValidShowThres)
		config.PegasusRefreshGuidanceExp = rs.PegasusRefreshGuidanceExp
		config.SpaceEnlargeExp = rs.SpaceEnlargeExp
		config.IconGuidanceExp = rs.IconGuidanceExp
		if rs.RefreshToast != "" {
			config.Toast.HasToast = true
			config.Toast.ToastMessage = rs.RefreshToast
		}
		//nolint: gomnd
		config.InlineSound = func() int8 {
			if s.matchFeatureControl(mid, buvid, "inline_sound") {
				return 2
			}
			if rs.OpenSound == 1 {
				return 1
			}
			return 2
		}()
		s.resolveSingleInlineConfig(ctx, config, param)
		s.resetAutoPlay(ctx, mid, buvid, config, param)
		s.resolveSingleGuide(config, abtest, param)
	} else {
		s.infoProm.Incr("tianma_backup_with_recommend_cache")
		count := s.indexCount(plat, abtest)
		rs.Items = s.recommendCache(count)
		log.Warn("feed index show disaster recovery data len(%d)", len(rs.Items))
	}
	is, infoc.IsRcmd = s.dealItem2(c, mid, buvid, plat, rs.Items, param, infoc.IsRcmd, noCache, followMode, now, abtest, infoc)
	infoc.AutoPlayInfoc = fmt.Sprintf("%d|%d", config.AutoplayCard, param.AutoPlayCard)
	s.dealAdLoc(c, is, param, adInfom, now) // adGif test
	s.cmLog(buvid, mid, plat, param, now, is, abtest.GifType, advert)
	s.configGuidence(config, abtest, param)
	if param.LoginEvent != 0 && rs.RefreshToast == "" && rs.PegasusRefreshGuidanceExp != 1 {
		config.Toast.HasToast = true
		config.Toast.ToastMessage = fmt.Sprintf("发现%d条新内容", len(is))
	}
	if param.DeviceType == 1 && param.InterestId != 0 && param.InterestResult != "" && param.InterestResult != "0" {
		config.Toast.HasToast = true
		config.Toast.ToastMessage = "根据你的兴趣为你推荐"
	}
	return
}

//nolint:gomnd
func visibleArea(thres int64) int64 {
	if thres != 0 {
		return thres
	}
	return 80
}

//nolint:gocognit
func (s *Service) initAbtest(abtest *feed.Abtest, cfg *feed.Config, param *feed.IndexParam) {
	abtest.IpadHDThreeColumn = cfg.IpadHDAbtest
	// gif abtest
	abtest.IsNewAd = _newAd
	// new ad big card
	abtest.IsNewAdBigCard = _newAdBigCard
	// banner
	// feature Index2AbtestNewBanner
	if (param.MobiApp == "iphone" && param.Build > 8510) ||
		(param.MobiApp == "android" && param.Build > 5415000) ||
		(param.MobiApp == "ipad" && param.Build > 12110) ||
		(param.MobiApp == "iphone_b" && param.Build > 8110) ||
		(param.MobiApp == "android_b" && param.Build > 591240) ||
		(param.MobiApp == "android_i") || (param.MobiApp == "iphone_i" && param.Build >= 64400200) ||
		(param.MobiApp == "win") {
		abtest.Banner = _newBannerResource
	}
	// feature Index2AbtestRcmdReason
	if (param.MobiApp == "iphone" && param.Device == "phone" && param.Build > 9150) || (param.MobiApp == "android" &&
		param.Build > 5525000) {
		abtest.RcmdReason = _newRcmdReason
	}
	// feature Index2AbtestRcmdReasonV2
	if (param.MobiApp == "android" && param.Build > 6025000) || (param.MobiApp == "iphone" && param.Build > 10130) ||
		(param.MobiApp == "ipad" && param.Build >= 32100000) || (param.MobiApp == "android_hd" && param.Build >= 1030000) ||
		(param.MobiApp == "win") {
		abtest.RcmdReason = _newRcmdReasonV2
	}
	// feature Index2AbtestStoryTP
	if ((param.MobiApp == "iphone" && param.Build > 10150) || (param.MobiApp == "android" && param.Build > 6055000)) &&
		s.c.Custom.StoryThreePoint { // 老story卡片三点控制
		abtest.StoryThreePoint = true
	}
	abtest.LiveContentMode = 1
}

var _osVerRgx = regexp.MustCompile(`osVer[^ ]*`)

//nolint:gomnd
func degradeIsNaviExp(param *feed.IndexParam, exp int8) int8 {
	if param.MobiApp != "iphone" {
		return exp
	}
	osVerList := _osVerRgx.FindStringSubmatch(param.Ua)
	if len(osVerList) == 0 {
		return exp
	}
	sub := strings.SplitAfter(osVerList[0], "osVer/")
	if len(sub) < 2 {
		return exp
	}
	osVerSubs := strings.Split(sub[1], ".")
	if len(osVerSubs) == 0 {
		return exp
	}
	osVer, err := strconv.ParseFloat(osVerSubs[0], 64)
	if err != nil {
		log.Error("Failed to parse float: %s, %+v", osVerSubs[0], errors.WithStack(err))
		return exp
	}
	if param.Build < 67400000 && osVer < 13.0 {
		return 0
	}
	return exp
}

func degradeAutoRefreshTime(in int64) int64 {
	const _defaultTime = 1200
	if in == 0 {
		return _defaultTime
	}
	return in
}

func (s *Service) fakeRcmdItemsByPrivacyWindow() *feed.AIResponse {
	items := make([]*ai.Item, 0, len(s.c.Custom.PrivacyModeAid))
	for _, aid := range s.c.Custom.PrivacyModeAid {
		items = append(items, &ai.Item{ID: aid, Goto: model.GotoAv, TrackID: "gateway_privacy_mode", IconType: cdm.AIUpIconType})
	}
	out := &feed.AIResponse{
		Items: items,
	}
	return out
}

func adAvResource(ctx context.Context, plat int8) int64 {
	scene := adresource.EmptyScene
	switch plat {
	case model.PlatIPhone, model.PlatIPhoneB:
		scene = adresource.PegasusAdAvIOS
	case model.PlatIPadHD, model.PlatIPad:
		scene = adresource.PegasusAdAvIPad
	case model.PlatAndroid, model.PlatAndroidB:
		scene = adresource.PegasusAdAvAndroid
	default:
		log.Info("Failed to match scene by plat: %d", plat)
	}
	resourceId, ok := adresource.CalcResourceID(ctx, scene)
	if !ok {
		return 0
	}
	return int64(resourceId)
}

func (s *Service) resolveSingleGuide(config *feed.Config, abtest *feed.Abtest, param *feed.IndexParam) {
	const (
		_guideSourceIcon     = "https://i0.hdslb.com/bfs/activity-plat/static/20210526/0977767b2e79d8ad0a36a731068a83d7/J7xHLt28y5.png"
		_guideAnimation      = "https://i0.hdslb.com/bfs/activity-plat/static/20210601/0977767b2e79d8ad0a36a731068a83d7/GDRCzjjjG6.gif"
		_guideAnimationNight = "https://i0.hdslb.com/bfs/activity-plat/static/20210601/0977767b2e79d8ad0a36a731068a83d7/1rOF4YW5Bh.gif"
	)
	if abtest.CanResetColumn() {
		config.NeedResetColumn = true
		config.Column = cdm.ColumnSvrSingle
		param.Column = cdm.ColumnSvrSingle
		if !abtest.RsNewUser {
			config.RecoverColumnGuidance = &feed.PopupGuidance{
				Title:          "已切换至单列模式",
				SubTitle:       "可以在[右下角三点]切换单/双列模式",
				SourceURL:      _guideSourceIcon,
				SourceNightURL: _guideSourceIcon,
				Option: []*feed.GuideOption{
					{
						Desc:  "切换至双列",
						Value: cdm.FlagConfirm,
						Type:  cdm.ColumnSvrDouble,
						Toast: "已成功切换至双列模式 可以在[右下角三点]切换单/双列模式",
					},
				},
			}
		}
	}
	if abtest.CanSupportGuide() {
		config.SwitchColumnGuidance = &feed.PopupGuidance{
			Title:          "邀你体验「推荐」单列模式",
			SubTitle:       "推荐内容直接看，浏览更方便",
			SourceURL:      _guideAnimation,
			SourceNightURL: _guideAnimationNight,
			Option: []*feed.GuideOption{
				{
					Desc:  "开启单列模式",
					Value: cdm.FlagConfirm,
					Type:  cdm.ColumnSvrSingle,
					Toast: "已成功切换至单列模式\n 可以在[右下角三点]切换单/双列模式",
				},
				{
					Desc:  "不了",
					Value: cdm.FlagCancel,
				},
			},
		}
	}
}

func (s *Service) resetColumn(ctx context.Context, mid int64, buvid string, config *feed.Config, param *feed.IndexParam) {
	if !s.c.Custom.Resetting.ColumnOpen {
		return
	}
	if !s.matchFeatureControl(mid, buvid, "reset") {
		return
	}
	if s.c.Custom.Resetting.ColumnTimestamp > param.ColumnTimestamp && param.OpenEvent == "cold" &&
		canEnableReset(ctx, param.MobiApp, int64(param.Build)) {
		config.NeedResetColumn = true
		config.Column = cdm.ColumnStatus(s.c.Custom.Resetting.Column)
		param.Column = cdm.ColumnStatus(s.c.Custom.Resetting.Column)
	}
}

func (s *Service) resetAutoPlay(ctx context.Context, mid int64, buvid string, config *feed.Config, param *feed.IndexParam) {
	if !s.c.Custom.Resetting.AutoplayOpen {
		return
	}
	if !s.matchFeatureControl(mid, buvid, "reset") {
		return
	}
	if s.c.Custom.Resetting.AutoplayTimestamp > param.AutoplayTimestamp && param.OpenEvent == "cold" &&
		canEnableReset(ctx, param.MobiApp, int64(param.Build)) {
		config.NeedResetAutoplay = true
		config.AutoplayCard = int8(s.c.Custom.Resetting.Autoplay)
	}
}

func (s *Service) resolveSingleInlineConfig(ctx context.Context, config *feed.Config, param *feed.IndexParam) {
	if feature.GetBuildLimit(ctx, "service.SingleInline", nil) {
		config.SingleAutoplayFlag = 1
	}
	if config.SingleAutoplayFlag == 1 { //命中实验
		//nolint:gomnd
		switch param.AutoPlayCard {
		case 0, 11, 1, 2:
			config.AutoplayCard = s.c.Custom.SingleInlineAutoPlay // 配置值第一阶段为1，第二阶段为11
		default:
			config.AutoplayCard = 0
		}
	}
}

func isSingleInline(item *ai.Item) bool {
	return item.SingleInline > 0
}

func canEnableReset(ctx context.Context, mobiApp string, build int64) bool {
	return feature.GetBuildLimit(ctx, "service.resetting", &feature.OriginResutl{
		BuildLimit: (mobiApp == "iphone" && build >= 62700000) ||
			(mobiApp == "android" && build >= 6270000)})
}

func canEnable4GWiFiAutoPlay(mobiApp string, build int64) bool {
	return (mobiApp == "android" && build >= 6140000) ||
		(mobiApp == "iphone" && build > 10350)
}

func (s *Service) indexConfig(ctx context.Context, plat int8, buvid string, mid int64, param *feed.IndexParam) (config *feed.Config) {
	config = &feed.Config{
		AutoRefreshTime: int64(time.Duration(s.c.Custom.AutoRefreshTime) / time.Second),
	}
	if s.c.Feed.Inline != nil {
		config.ShowInlineDanmaku = s.c.Feed.Inline.ShowInlineDanmaku
	}
	// feature IndexConfigColumn
	if param.MobiApp == "iphone_b" && param.Build > _iosNewBlue { // ios蓝 2.5之后 默认0、1（实验组：单）、2（实验组：双）直接返回单列、用户的3（单）、4（双）还是正常控制
		switch param.Column {
		case cdm.ColumnDefault, cdm.ColumnSvrSingle, cdm.ColumnSvrDouble:
			config.Column = cdm.ColumnSvrSingle
		default:
			config.Column = cdm.Columnm[param.Column]
		}
	} else if model.IsPad(plat) {
		config.Column = cdm.ColumnSvrSingle
	} else {
		config.Column = cdm.Columnm[param.Column]
	}
	s.resetColumn(ctx, mid, buvid, config, param)
	// if mid > 0 && mid%20 == 19 {
	// 	config.FeedCleanAbtest = 1
	// } else {
	// 	config.FeedCleanAbtest = 0
	// }
	config.FeedCleanAbtest = 0
	// 转场动画abtest
	// feature HomeTransferTest
	if s.c.Custom.TransferSwitch && param.MobiApp == "iphone_b" && param.Build == 8030 && crc32.ChecksumIEEE([]byte(buvid+"_blueversion"))%20 < 10 {
		config.HomeTransferTest = _home_transfer_new
	}
	config.IpadHDAbtest = s.ipadHDThreeColumnAbtest(ctx, param)
	config.IsBackToHomepage = true
	config.EnableRcmdGuide = true
	// ipad 不允许自动播放、不在实验里面也不允许自动播放
	config.AutoplayCard = 2
	config.CardDensityExp = 1
	if !model.IsPad(plat) {
		switch cdm.Columnm[param.Column] {
		case cdm.ColumnSvrDouble:
			// 6.14 按配置文件内容下发值，默认不下发
			// feature CanEnable4GWiFiAutoPlay
			if s.c.Custom.Prefer4GAutoPlay && canEnable4GWiFiAutoPlay(param.MobiApp, int64(param.Build)) {
				config.AutoplayCard = 0
			}
			// ios 6.14 遇到 3 返回 2
			if model.IsIOS(plat) && param.Build == 10370 && param.AutoPlayCard == 3 {
				config.AutoplayCard = 2
			}
		case cdm.ColumnSvrSingle:
			//nolint:gomnd
			switch param.AutoPlayCard {
			case 3, 10:
				config.AutoplayCard = 1
			case 0, 1, 2, 4, 11:
				config.AutoplayCard = 2
			default:
				config.AutoplayCard = 2
			}
			if s.c.Custom.SingleAutoPlayForce > 0 {
				config.AutoplayCard = s.c.Custom.SingleAutoPlayForce
			}
		default:
		}
	}
	if s.enableAndroidBAutoplay(ctx, param) {
		config.AutoplayCard = 11
		config.NeedResetAutoplay = true
	}
	if mid < 1 {
		return
	}
	if _, ok := s.followModeList[mid]; ok {
		tmpConfig := &feed.FollowMode{}
		if s.c.Feed.Index.FollowMode == nil {
			*tmpConfig = *_followMode
		} else {
			*tmpConfig = *s.c.Feed.Index.FollowMode
		}
		if param.RecsysMode != 1 {
			tmpConfig.ToastMessage = ""
		}
		config.FollowMode = tmpConfig
	}
	return
}

func (s *Service) enableAndroidBAutoplay(ctx context.Context, param *feed.IndexParam) bool {
	return s.c.Custom.AndroidBAutoplaySwitch &&
		(s.c.Custom.AndroidBAutoplayTimestamp >= param.AutoplayTimestamp) &&
		feature.GetBuildLimit(ctx, "service.androidbAutoplay", nil)
}

func (s *Service) ipadHDThreeColumnAbtest(ctx context.Context, param *feed.IndexParam) int8 {
	if !feature.GetBuildLimit(ctx, "service.ipadHDThreeColumn", &feature.OriginResutl{
		BuildLimit: (param.MobiApp == "ipad" && param.Build >= 31700000) ||
			(param.MobiApp == "iphone" && param.Device == "pad" && param.Build >= 63800000),
	}) {
		return 0
	}
	return 1
}

func (s *Service) indexRcmd2(c context.Context, plat int8, buvid string, mid int64, param *feed.IndexParam, group int, zone *locgrpc.InfoReply, style int, avAdResource int64, autoPlay string,
	noCache bool, applist, deviceInfo string, resourceID, bannerExp, adExp int, now time.Time, abtest *feed.Abtest) (res *feed.AIResponse) {
	count := s.indexCount(plat, abtest)
	resource := s.adResource(c, plat, param.Build)
	if buvid != "" || mid > 0 {
		var (
			err    error
			zoneID int64
		)
		if zone != nil {
			zoneID = zone.ZoneId
		}
		stat2.MetricCMResource.Inc(strconv.FormatInt(resource, 10), strconv.FormatInt(int64(plat), 10))
		if res, err = s.rcmd.Recommend(c, plat, buvid, mid, param.Build, param.LoginEvent, param.ParentMode,
			param.RecsysMode, param.TeenagersMode, param.LessonsMode, zoneID, group, param.Interest, param.Network,
			style, param.Column, param.Flush, count, param.DeviceType, avAdResource, resource, autoPlay,
			param.DeviceName, param.OpenEvent, param.BannerHash, applist, deviceInfo, param.InterestV2, resourceID,
			bannerExp, adExp, param.MobiApp, param.AdExtra, param.Pull, param.RedPoint, param.InlineSound,
			param.InlineDanmu, now, param.ScreenWindowType, param.DisableRcmd, param.LocalBuvid, param.OpenAppURL,
			param.DituiLanding, param.InterestId, param.InterestResult, param.VideoMode); err != nil {
			log.Error("%+v", err)
		}
		if noCache {
			res.IsRcmd = true
			return
		}
		if len(res.Items) != 0 {
			res.IsRcmd = true
		}
		var fromCache bool
		if len(res.Items) == 0 && mid > 0 && !ecode.ServiceUnavailable.Equal(err) {
			s.infoProm.Incr("tianma_backup_with_index_cache")
			res.Items = s.recommendCache(count)
			if len(res.Items) != 0 {
				s.pHit.Incr("index_cache")
			} else {
				s.pMiss.Incr("index_cache")
			}
			fromCache = true
		}
		if len(res.Items) == 0 || (fromCache && len(res.Items) < count) {
			s.infoProm.Incr("tianma_backup_with_recommend_cache")
			res.Items = s.recommendCache(count)
		}
	} else {
		s.errProm.Incr("Buvid_empty")
		paramStr, _ := json.Marshal(param)
		log.Warn("[BuvidEmpty] Plat %d, Build %d, Param %s", plat, param.Build, paramStr)
		res = &feed.AIResponse{Items: s.recommendCache(count)}
	}
	return
}

func (s *Service) indexAd2(c context.Context, plat int8, buvid string, mid int64, param *feed.IndexParam,
	zone *locgrpc.InfoReply, style int, now time.Time) (adm map[int32][]*cm.AdInfo, adAidm, adRoomidm map[int64]struct{}, err error) {
	var advert *cm.Ad
	resource := s.adResource(c, plat, param.Build)
	if resource == 0 {
		return
	}
	//  兼容老的style逻辑，3为新单列，上报给商业产品的参数定义为：1 单列 2双列
	// if style == 3 {
	// 	style = 1
	// }
	stat2.MetricCMResource.Inc(strconv.FormatInt(resource, 10), strconv.FormatInt(int64(plat), 10))
	var country, province, city string
	if zone != nil {
		country = zone.Country
		province = zone.Province
		city = zone.City
	}
	if advert, err = s.ad.Ad(c, mid, param.Build, buvid, []int64{resource}, country, province, city, param.Network,
		param.MobiApp, param.Device, param.OpenEvent, param.AdExtra, style, now); err != nil {
		return
	}
	adm, adAidm, adRoomidm = advert.AdChange(resource, _cardAdAv)
	return
}

func (s *Service) indexAd3(c context.Context, isShow int, plat int8, buvid string, mid int64, param *feed.IndexParam, zone *locgrpc.InfoReply, style int, now time.Time) (adm map[int32][]*cm.AdInfo, adAidm, adRoomidm map[int64]struct{}, advert *cm.NewAd, respCode int, err error) {
	resource := s.adResource(c, plat, param.Build)
	if resource == 0 {
		return
	}
	stat2.MetricCMResource.Inc(strconv.FormatInt(resource, 10), strconv.FormatInt(int64(plat), 10))
	var country, province, city string
	if zone != nil {
		country = zone.Country
		province = zone.Province
		city = zone.City
	}
	if advert, respCode, err = s.ad.NewAd(c, mid, param.Build, buvid, []int64{resource}, country, province, city, param.Network,
		param.MobiApp, param.Device, param.OpenEvent, param.AdExtra, style, isShow, now); err != nil {
		return
	}
	adm, adAidm, adRoomidm = advert.NewAdChange(resource, _cardAdAv)
	for _, ads := range adm {
		for _, ad := range ads {
			if isShow == 1 && (ad.CreativeStyle == 2 || ad.CreativeType == 4) {
				s.infoProm.Incr("impossibility_gif")
			}
		}
	}
	return
}

func (s *Service) indexBanner2(c context.Context, plat int8, buvid string, mid int64, param *feed.IndexParam, abtest *feed.Abtest) (banners []*banner.Banner, version string, err error) {
	hash := param.BannerHash
	if param.LoginEvent != 0 {
		hash = ""
	}
	banners, version, err = s.banners(c, plat, param.Build, mid, buvid, param.Network, param.MobiApp, param.Device, param.OpenEvent, param.AdExtra, hash, param.SplashID, abtest, param.LessonsMode, nil, param.TeenagersMode)
	return
}

//nolint: gocognit
func (s *Service) mergeItem2(_ context.Context, plat int8, mid int64, rs []*ai.Item, adms map[int32][]*cm.AdInfo, adAidm, adRoomidm, adEpidm map[int64]struct{}, banners []*banner.Banner, version string, followMode bool, abtest *feed.Abtest, param *feed.IndexParam, ic *feed.Infoc) (is []*ai.Item, adInfoms map[int32][]*cm.AdInfo) {
	if len(rs) == 0 {
		return
	}
	const (
		cardIndex     = 7
		cardIndexIPad = 17
		cardOffset    = 2
	)
	var showBanner bool // 是否有展示banner
	// ai接口返回的banner
	for _, r := range rs {
		if r.Goto == model.GotoBanner {
			showBanner = true
			break
		}
	}
	// 老逻辑的banner
	if len(banners) != 0 {
		rs = append([]*ai.Item{{Goto: model.GotoBanner, Banners: banners, Version: version}}, rs...)
		showBanner = true
	}
	if showBanner && (abtest == nil || abtest.AdExp != _aiAdExp) {
		for index, adm := range adms {
			for _, ad := range adm {
				// 广告大卡
				_, adWebOk := _cardAdWebm[ad.CardType]
				// 广告inline卡
				_, adPlayerOk := _cardAdPlayerm[ad.CardType]
				if (adWebOk || adPlayerOk) && ((model.IsPad(plat) && index <= cardIndexIPad) || index <= cardIndex) {
					ad.CardIndex = ad.CardIndex + cardOffset
				}
			}
		}
	}
	is = make([]*ai.Item, 0, len(rs)+len(adms))
	adInfoms = map[int32][]*cm.AdInfo{}
	var (
		existsAdWeb bool
		// card_index和实际的位置不匹配
		insert = map[int32]*ai.Item{}
	)
	for _, r := range rs {
		if abtest != nil && abtest.AdExp == _aiAdExp {
			// ai广告
			switch r.Goto {
			case model.GotoAdAv, model.GotoAdWeb, model.GotoAdWebS, model.GotoAdPlayer, model.GotoAdInlineGesture,
				model.GotoAdInline360, model.GotoAdInlineLive, model.GotoAdWebGif, model.GotoAdInlineChoose, model.GotoAdLive,
				model.GotoAdDynamic, model.GotoAdInlineChooseTeam, model.GotoAdInlineAv, model.GotoAdWebGifReservation,
				model.GotoAdPlayerReservation, model.GotoAdInline3D, model.GotoAdPgc, model.GotoAdInlinePgc,
				model.GotoAdInlineEggs, model.GotoAdInline3DV2:
				var item *ai.Item
				ads, ok := adms[r.BizIdx]
				if !ok || len(ads) == 0 {
					FillDiscard(r.ID, r.Goto, feed.DiscardReasonAd, "商业卡ad_info为空", ic)
					continue
				}
				// 强行获取第一个
				ad := ads[0]
				if ad.CreativeID == 0 {
					FillDiscard(r.ID, r.Goto, feed.DiscardReasonAd, "商业卡creative_id为0", ic)
					continue
				}
				var adID int64
				if r.Goto == model.GotoAdAv || r.Goto == model.GotoAdInlineAv {
					if ad.CreativeContent != nil {
						adID = ad.CreativeContent.VideoID
					}
				}
				if r.Goto == model.GotoAdLive || r.Goto == model.GotoAdInlineLive {
					adID = ad.RoomID
				}
				if r.Goto == model.GotoAdWebGifReservation || r.Goto == model.GotoAdPlayerReservation {
					adID = ad.LiveBookingID
				}
				if r.Goto == model.GotoAdInlinePgc || r.Goto == model.GotoAdPgc {
					adID = ad.EpId
				}
				item = &ai.Item{ID: adID, Goto: r.Goto, Ads: ads, SingleAdNew: r.SingleAdNew, TrackID: r.TrackID, AvFeature: r.AvFeature}
				// moni数据上报
				str := fmt.Sprintf("ad:%s_index:%d_banner:%t_openEvent:%s", _adCardMap[ad.CardType], ad.CardIndex, len(banners) == 0, param.OpenEvent)
				s.infoProm.Incr(str)
				is = append(is, item)
				continue
			default:
			}
		} else {
			// for 循环确保广告优先
			for {
				ads, ok := adms[int32(len(is))]
				if !ok || len(ads) == 0 {
					break
				}
				// 强行获取第一个
				ad := ads[0]
				if ad.CreativeID == 0 {
					adInfoms[ad.CardIndex-1] = ads
					break
				}
				var (
					adCardType string
					item       *ai.Item
				)
				if adCardType, ok = _cardAdAvm[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdAv, Ads: ads}
				} else if adCardType, ok = _cardAdLivem[ad.CardType]; ok {
					item = &ai.Item{ID: ad.RoomID, Goto: model.GotoAdLive, Ads: ads}
				} else if adCardType, ok = _cardAdWebm[ad.CardType]; ok {
					item = &ai.Item{Goto: model.GotoAdWeb, Ads: ads}
					existsAdWeb = true
				} else if adCardType, ok = _cardAdWebSm[ad.CardType]; ok {
					item = &ai.Item{Goto: model.GotoAdWebS, Ads: ads}
				} else if adCardType, ok = _cardAdPlayerm[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdPlayer, Ads: ads}
				} else if adCardType, ok = _cardAdInline3D[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdInline3D, Ads: ads}
				} else if adCardType, ok = _cardAdInline3DV2[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdInline3DV2, Ads: ads}
				} else if adCardType, ok = _cardAdInlineGesture[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdInlineGesture, Ads: ads}
				} else if adCardType, ok = _cardAdInline360[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdInline360, Ads: ads}
				} else if adCardType, ok = _cardAdInlineLive[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdInlineLive, Ads: ads}
				} else if adCardType, ok = _cardAdWebGif[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdWebGif, Ads: ads}
				} else if adCardType, ok = _cardAdColorEgg[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdInlineEggs, Ads: ads}
				} else if adCardType, ok = _cardAdChoose[ad.CardType]; ok {
					item = &ai.Item{Goto: model.GotoAdInlineChoose, Ads: ads}
				} else if adCardType, ok = _cardAdInlineChooseTeam[ad.CardType]; ok {
					item = &ai.Item{Goto: model.GotoAdInlineChooseTeam, Ads: ads}
				} else if adCardType, ok = _cardAdDynamicm[ad.CardType]; ok {
					item = &ai.Item{Goto: model.GotoAdDynamic, Ads: ads}
				} else if adCardType, ok = _cardAdInlineAvm[ad.CardType]; ok {
					item = &ai.Item{ID: ad.CreativeContent.VideoID, Goto: model.GotoAdInlineAv, Ads: ads}
				} else if adCardType, ok = _cardAdPlayerReservation[ad.CardType]; ok {
					item = &ai.Item{ID: ad.LiveBookingID, Goto: model.GotoAdPlayerReservation, Ads: ads}
				} else if adCardType, ok = _cardAdWebGifReservation[ad.CardType]; ok {
					item = &ai.Item{ID: ad.LiveBookingID, Goto: model.GotoAdWebGifReservation, Ads: ads}
				} else if adCardType, ok = _cardAdPgc[ad.CardType]; ok {
					item = &ai.Item{ID: ad.EpId, Goto: model.GotoAdPgc, Ads: ads}
				} else if adCardType, ok = _cardAdInlinePgc[ad.CardType]; ok {
					item = &ai.Item{ID: ad.EpId, Goto: model.GotoAdInlinePgc, Ads: ads}
				} else {
					b, _ := json.Marshal(ad)
					log.Error("ad---%s", b)
					// 不能识别的卡片类型抛弃，往下聚合ai卡片
					break
				}
				// 目前广告说用到3、5、6
				str := fmt.Sprintf("ad:%s_index:%d_banner:%t_openEvent:%s", adCardType, ad.CardIndex, len(banners) == 0, param.OpenEvent)
				s.infoProm.Incr(str)
				if int32(len(is)) != ad.CardIndex-1 {
					insert[ad.CardIndex-1] = item
					break
				}
				is = append(is, item)
			}
		}
		if r.Goto == model.GotoAv {
			if _, ok := adAidm[r.ID]; ok {
				FillDiscard(r.ID, r.Goto, feed.DiscardReasonRepeatedId, "avid与广告卡avid重复", ic)
				continue
			}
		} else if r.Goto == model.GotoLive {
			if _, ok := adRoomidm[r.ID]; ok {
				FillDiscard(r.ID, r.Goto, feed.DiscardReasonRepeatedId, "直播卡与广告卡重复", ic)
				continue
			}
		} else if r.Goto == model.GotoPGC {
			if _, ok := adEpidm[r.ID]; ok {
				FillDiscard(r.ID, r.Goto, feed.DiscardReasonRepeatedId, "pgc卡与广告卡重复", ic)
				continue
			}
		} else if r.Goto == model.GotoBanner && (len(is) != 0 || abtest != nil && abtest.NewUser == _userHideBanner) {
			FillDiscard(r.ID, r.Goto, feed.DiscardReasonOther, "banner卡不在第一位", ic)
			// banner 必须在第一位
			continue
		} else if r.Goto == model.GotoRank && existsAdWeb {
			continue
		} else if r.Goto == model.GotoLogin && mid > 0 {
			continue
		} else if r.Goto == model.GotoFollowMode && !followMode {
			continue
		}
		is = append(is, r)
		if v, ok := insert[int32(len(is))]; ok {
			is = append(is, v)
		}
	}
	return
}

func (s *Service) dealAdLoc(_ context.Context, is []card.Handler, param *feed.IndexParam, adInfom map[int32][]*cm.AdInfo, now time.Time) {
	il := len(is)
	if il == 0 {
		return
	}
	if param.Idx < 1 {
		param.Idx = now.Unix()
	}
	for i, h := range is {
		if param.Pull {
			h.Get().Idx = param.Idx + int64(il-i)
		} else {
			h.Get().Idx = param.Idx - int64(i+1)
		}
		var (
			ads []*cm.AdInfo
			ok  bool
		)
		if ads, ok = adInfom[int32(i)]; ok {
			for _, ad := range ads {
				h.Get().AdInfo = ad
				if h.Get().Rcmd != nil {
					h.Get().Rcmd.Ad = ad
				}
				break
			}
		}
		if !ok {
			if h.Get().AdInfo != nil {
				h.Get().AdInfo.CardIndex = int32(i + 1)
			}
		}
	}
}

//nolint:gocognit, bilirailguncheck, unparam
func (s *Service) dealItem2(c context.Context, mid int64, buvid string, plat int8, rs []*ai.Item, param *feed.IndexParam, isRcmd, noCache, followMode bool,
	now time.Time, abtest *feed.Abtest, infoc *feed.Infoc) (is []card.Handler, isAI bool) {
	if len(rs) == 0 {
		is = []card.Handler{}
		return
	}
	var (
		aids, tids, roomIDs, inlineRoomIDs, metaIDs, audioIDs, picIDs, channeIDs, storyAids, inlineAids, favAids, feedCreativeIds, gameIDs, reservationIds []int64
		seasonIDs, sids, epPlayerIDs, specialSeasonIds, pgcEpids                                                                                           []int32
		upIDs, avUpIDs, rmUpIDs, inlineRmUpIDs, mtUpIDs, storyUpIDs, bangumiAids                                                                           []int64
		gameParams                                                                                                                                         []*game.GameParam
		amplayer, storyamplayer                                                                                                                            map[int64]*arcgrpc.ArcPlayer
		tagm                                                                                                                                               map[int64]*taggrpc.Tag
		rm, inlinerm                                                                                                                                       map[int64]*live.Room
		hasUpdate, getBanner, vipRenew, hideInlineCard                                                                                                     bool
		update                                                                                                                                             *bangumi.Update
		pgcRemind                                                                                                                                          *bangumi.Remind
		metam                                                                                                                                              map[int64]*article.Meta
		audiom                                                                                                                                             map[int64]*audio.Audio
		cardm                                                                                                                                              map[int64]*accountgrpc.Card
		statm                                                                                                                                              map[int64]*relationgrpc.StatReply
		moe                                                                                                                                                *bangumi.Moe
		isAtten                                                                                                                                            map[int64]int8
		arcOK                                                                                                                                              bool
		seasonm, sm                                                                                                                                        map[int32]*episodegrpc.EpisodeCardsProto
		banners                                                                                                                                            []*banner.Banner
		version                                                                                                                                            string
		picm                                                                                                                                               map[int64]*bplus.Picture
		vipRenewReply                                                                                                                                      *viprpc.TipsRenewReply
		tunnels                                                                                                                                            map[int64]*tunnelgrpc.FeedCard
		eppm                                                                                                                                               map[int32]*pgcinline.EpisodeCard
		haslike                                                                                                                                            = make(map[int64]int8)
		epHasLike                                                                                                                                          = make(map[int64]int8)
		pgcSeasonm                                                                                                                                         map[int32]*pgcAppGrpc.SeasonCardInfoProto
		multiMaterials                                                                                                                                     map[int64]*feedMgr.Material
		gamem                                                                                                                                              map[int64]*game.Game
		reservationm                                                                                                                                       map[int64]*activitygrpc.UpActReserveRelationInfo
		// gif count
		cardGifCount, aiCardGifCount, adCardGifCount, adCardCount int
		// gif count
		channelm, channelDetailm map[int64]*channelgrpc.ChannelCard
		bannerInfoItem           []*ai.BannerInfoItem
		hasFav                   map[int64]int8
		gatherOids               [][]int64
		hasCoin                  map[int64]int64
		episodeSeasonCardm       map[int64]*pgccard.EpisodeCard // 新OGV物料来源
		pgcCardm                 map[int32]*pgccard.EpisodeCard
		specialIds               []int64
		specialCardm             map[int64]*resourceV2grpc.AppSpecialCard
		openCoursePegasusMark    map[int64]bool
		likeState                map[int64]*thumbupgrpc.StatState
		epMaterialReq            []*deliverygrpc.EpMaterialReq
		epMaterialm              map[int64]*deliverygrpc.EpMaterial
	)
	convergem := map[int64]*operate.Card{}
	followm := map[int64]*operate.Card{}
	specialm := map[int64]*operate.Card{}
	liveUpm := map[int64][]*live.Card{}
	avconvergem := map[int64]*operate.Card{}
	specialCardIndex := map[string]int32{}
	isAI = isRcmd
	rowType := stat.BuildRowType(param.Column, plat)
	for _, r := range rs {
		if r == nil {
			continue
		}
		r.SingleInline = s.rcmdSingleInline(c, param)
		if r.CreativeId != 0 && !withoutTunnelMaterialGoto.Has(r.Goto) {
			feedCreativeIds = append(feedCreativeIds, r.CreativeId)
		}
		if r.CreativeId != 0 && r.Goto == model.GotoGame && r.PosRecUniqueID != "" {
			feedCreativeIds = append(feedCreativeIds, r.CreativeId)
		}
		if needHideGuidanceGoto.Has(r.Goto) || isSingleInline(r) {
			abtestHideGuidance(abtest, _hideGuidanceByInline)
		}
		switch r.Goto {
		case model.GotoBanner:
			if len(r.Banners) != 0 {
				banners = r.Banners
				version = r.Version
			} else {
				getBanner = true
			}
			specialCardIndex[_bannerCard] = 0
			hideInlineCard = true
			if r.BannerInfo != nil {
				fixIOSInlineBannerBug(c, r)
				bannerInfoItem = r.BannerInfo.Items
				aids, epPlayerIDs, inlineRoomIDs = bannerAddTo(r, aids, epPlayerIDs, inlineRoomIDs, abtest)
			}
		case model.GotoAv, model.GotoPlayer, model.GotoUpRcmdAv, model.GotoInlineAv, model.GotoInlineAvV2:
			if r.ID != 0 {
				// feature StoryAids
				if r.JumpGoto == model.GotoVerticalAv && (param.MobiApp == "iphone" && param.Build > 10030 ||
					param.MobiApp == "android" && param.Build > 6025500 || param.MobiApp == "android_i" &&
					param.Build >= 6790300 || param.MobiApp == "iphone_i" && param.Build >= 67900200) || r.StNewCover == 1 {
					storyAids = append(storyAids, r.ID)
				} else {
					aids = append(aids, r.ID)
				}
				if needLikeGoto.Has(r.Goto) || isSingleInline(r) {
					// 用于查询稿件是否点赞
					inlineAids = append(inlineAids, r.ID)
				}
				if needFavGoto.Has(r.Goto) || isSingleInline(r) {
					// 用于查询稿件是否收藏
					favAids = append(favAids, r.ID)
				}
			}
			if r.Tid != 0 {
				tids = append(tids, r.Tid)
			}
			if r.CoverGif != "" {
				s.allGifState(_aiGif, abtest)
				abtestHideGuidance(abtest, _hideGuidanceByAIGif)
			}
		case model.GotoAdAv:
			if r.ID != 0 {
				aids = append(aids, r.ID)
			}
			if r.Tid != 0 {
				tids = append(tids, r.Tid)
			}
			if _, ok := s.adCreativeStyle(r.Ads); ok {
				s.allGifState(_adGif, abtest)
				abtestHideGuidance(abtest, _hideGuidanceByAdGif)
			}
			adCardCount++
		case model.GotoAdInlineAv:
			if r.ID != 0 {
				aids = append(aids, r.ID)
				inlineAids = append(inlineAids, r.ID)
				favAids = append(favAids, r.ID)
			}
			if r.Tid != 0 {
				tids = append(tids, r.Tid)
			}
			if _, ok := s.adCreativeStyle(r.Ads); ok {
				s.allGifState(_adGif, abtest)
				abtestHideGuidance(abtest, _hideGuidanceByInline)
			}
			adCardCount++
		case model.GotoAdLive:
			if r.ID != 0 {
				roomIDs = append(roomIDs, r.ID)
			}
			if _, ok := s.adCreativeStyle(r.Ads); ok {
				s.allGifState(_adGif, abtest)
				abtestHideGuidance(abtest, _hideGuidanceByAdGif)
			}
			adCardCount++
		case model.GotoAdInlineLive:
			if r.ID != 0 {
				inlineRoomIDs = append(inlineRoomIDs, r.ID)
			}
			if _, ok := s.adCreativeStyle(r.Ads); ok {
				s.allGifState(_adGif, abtest)
				abtestHideGuidance(abtest, _hideGuidanceByInline)
			}
			adCardCount++
		case model.GotoAdPgc:
			if r.ID != 0 {
				pgcEpids = append(pgcEpids, int32(r.ID))
			}
			if _, ok := s.adCreativeStyle(r.Ads); ok {
				s.allGifState(_adGif, abtest)
				abtestHideGuidance(abtest, _hideGuidanceByAdGif)
			}
			adCardCount++
		case model.GotoAdInlinePgc:
			if r.ID != 0 {
				epPlayerIDs = append(epPlayerIDs, int32(r.ID))
			}
			if _, ok := s.adCreativeStyle(r.Ads); ok {
				s.allGifState(_adGif, abtest)
				abtestHideGuidance(abtest, _hideGuidanceByInline)
			}
			adCardCount++
		case model.GotoAdWebS, model.GotoAdWeb, model.GotoAdPlayer, model.GotoAdInlineGesture, model.GotoAdInline360,
			model.GotoAdWebGif, model.GotoAdInlineChoose, model.GotoAdDynamic, model.GotoAdInlineChooseTeam,
			model.GotoAdPlayerReservation, model.GotoAdWebGifReservation, model.GotoAdInline3D, model.GotoAdInlineEggs,
			model.GotoAdInline3DV2:
			var (
				ad *cm.AdInfo
				ok bool
			)
			if ad, ok = s.adCreativeStyle(r.Ads); ok {
				s.allGifState(_adGif, abtest)
				abtestHideGuidance(abtest, _hideGuidanceByAdGif)
			}
			if ad != nil {
				if cindex := specialCardIndex[_adCard]; cindex == 0 || cindex > ad.CardIndex {
					specialCardIndex[_adCard] = ad.CardIndex
				}
				if _, ok := _delAdCard[ad.CardType]; ok {
					hideInlineCard = true
				}
				adCardCount++
			}
			if needAdReservationGoto.Has(r.Goto) {
				reservationIds = append(reservationIds, r.ID)
			}
		case model.GotoLive, model.GotoPlayerLive:
			func() {
				if r.ID != 0 {
					if isSingleV1Inline(r, param) {
						inlineRoomIDs = append(inlineRoomIDs, r.ID)
						return
					}
					roomIDs = append(roomIDs, r.ID)
				}
			}()
		case model.GotoInlineLive:
			if r.ID != 0 {
				inlineRoomIDs = append(inlineRoomIDs, r.ID)
			}
		case model.GotoBangumi:
			if r.ID != 0 {
				sids = append(sids, int32(r.ID))
				aids = append(aids, r.ID)
				bangumiAids = append(bangumiAids, r.ID)
			}
			if r.Tid != 0 {
				tids = append(tids, r.Tid)
			}
			if r.CreativeId > 0 {
				epMaterialReq = append(epMaterialReq, &deliverygrpc.EpMaterialReq{
					MaterialNo: r.CreativeId,
					Epid:       r.Epid,
				})
			}
			if isSingleV1Inline(r, param) && r.Epid > 0 {
				epPlayerIDs = append(epPlayerIDs, r.Epid)
			}
		case model.GotoPGC:
			if r.ID != 0 {
				pgcEpids = append(pgcEpids, int32(r.ID))
			}
			if isSingleV1Inline(r, param) {
				epPlayerIDs = append(epPlayerIDs, int32(r.ID))
			}
			if r.CreativeId > 0 {
				epMaterialReq = append(epMaterialReq, &deliverygrpc.EpMaterialReq{
					MaterialNo: r.CreativeId,
					Epid:       int32(r.ID),
				})
			}
		case model.GotoBangumiRcmd:
			hasUpdate = true
		case model.GotoConvergeAi:
			if r.ConvergeInfo == nil {
				continue
			}
			id := r.ID + _convergeAi
			card, aiAids := s.convergeCardAi(c, r.ConvergeInfo, id)
			if err := s.cardPub.Send(c, card.Param, card); err != nil {
				continue
			}
			convergem[id] = card
			aids = append(aids, aiAids...)
		case model.GotoArticleS:
			if r.ID != 0 {
				metaIDs = append(metaIDs, r.ID)
			}
		case model.GotoAudio:
			if r.ID != 0 {
				audioIDs = append(audioIDs, r.ID)
			}
		case model.GotoLiveUpRcmd:
			cardm, upID := s.liveUpRcmdCard(c, r.ID)
			for id, card := range cardm {
				liveUpm[id] = card
			}
			upIDs = append(upIDs, upID...)
		case model.GotoChannelRcmd:
			cardm, aid, tid := s.channelRcmdCard(c, r.ID)
			for id, card := range cardm {
				followm[id] = card
			}
			aids = append(aids, aid...)
			tids = append(tids, tid...)
		case model.GotoSpecial, model.GotoSpecialS:
			if plat == model.PlatIPad || plat == model.PlatIPadHD || plat == model.PlatWPhone {
				FillDiscard(r.ID, r.Goto, feed.DiscardReasonCannotBuildCard, "ipad不支持特殊卡", infoc)
				continue
			}
			specialIds = append(specialIds, r.ID)
			if r.SingleSpecialInfo != nil {
				id := r.SingleSpecialInfo.SpID
				switch r.SingleSpecialInfo.SpType {
				case "av":
					aids = append(aids, id)
					// 用于查询稿件是否点赞
					inlineAids = append(inlineAids, id)
					// 用于查询稿件是否收藏
					favAids = append(favAids, id)
				case "pgc":
					epPlayerIDs = append(epPlayerIDs, int32(id))
					pgcEpids = append(pgcEpids, int32(id))
				case "live":
					roomIDs = append(roomIDs, id)
					inlineRoomIDs = append(inlineRoomIDs, id)
				case "article":
					metaIDs = append(metaIDs, id)
				case "season":
					specialSeasonIds = append(specialSeasonIds, int32(id))
				default:
				}
			}
			if s.teenagerSpecialCondition(param, r) {
				if cdm.Columnm[param.Column] == cdm.ColumnSvrSingle {
					r.Style = 2
				}
			}
		case model.GotoPicture:
			if r.ID != 0 {
				picIDs = append(picIDs, r.ID)
			}
			if r.RcmdReason != nil && r.RcmdReason.Style == 4 {
				upIDs = append(upIDs, r.RcmdReason.FollowedMid)
			}
		case model.GotoPlayerBangumi, model.GotoInlinePGC, model.GotoInlineBangumi:
			if r.ID != 0 {
				epPlayerIDs = append(epPlayerIDs, int32(r.ID))
			}
			if r.CreativeId > 0 {
				epMaterialReq = append(epMaterialReq, &deliverygrpc.EpMaterialReq{
					MaterialNo: r.CreativeId,
					Epid:       int32(r.ID),
				})
			}
		case model.GotoVipRenew:
			vipRenew = true
		case model.GotoAvConverge, model.GotoMultilayerConverge:
			avconverge, aiAids := s.avConvergeCard(c, r)
			aids = append(aids, aiAids...)
			avconvergem[r.ID] = avconverge
			if r.Tid != 0 {
				tids = append(tids, r.Tid)
			}
		case model.GotoSpecialChannel:
			specialcard, channeID := s.specialCardChannel(c, r.ID)
			channeIDs = append(channeIDs, channeID)
			specialm[r.ID] = specialcard
		case model.GotoTunnel:
			gatherOids = append(gatherOids, []int64{r.ID})
		case model.GotoNewTunnel:
			gatherOids = newTunnelAddTo(r.MsgIDs, gatherOids)
		case model.GotoBigTunnel:
			gatherOids = append(gatherOids, []int64{r.ID})
			aids, epPlayerIDs, inlineRoomIDs = bigTunnelAddTo(r, aids, epPlayerIDs, inlineRoomIDs, abtest)
		case model.GotoAiStory:
			sAids, sTids := s.storyCard(c, r.StoryInfo)
			storyAids = append(storyAids, sAids...)
			tids = append(tids, sTids...)
		case model.GotoGame:
			gameIDs = append(gameIDs, r.ID)
			if r.PosRecUniqueID == "" && r.CreativeId > 0 {
				gameParams = append(gameParams, &game.GameParam{
					GameId:     r.ID,
					CreativeId: r.CreativeId,
				})
			}
		default:
		}
	}
	g, ctx := errgroup.WithContext(c)
	if getBanner && !s.c.Custom.ResourceDegradeSwitch {
		g.Go(func() (err error) {
			if banners, version, err = s.banners(ctx, plat, param.Build, mid, buvid, param.Network, param.MobiApp, param.Device, param.OpenEvent, param.AdExtra, "", param.SplashID, abtest, param.LessonsMode, bannerInfoItem, param.TeenagersMode); err != nil {
				log.Error("%+v", err)
				err = nil
			} else {
				specialCardIndex[_bannerCard] = 0
			}
			return
		})
	}
	if len(aids) != 0 {
		g.Go(func() (err error) {
			if amplayer, err = s.ArcsPlayer(ctx, aids); err != nil {
				return
			}
			for _, a := range amplayer {
				avUpIDs = append(avUpIDs, a.Arc.Author.Mid)
			}
			if i18n.PreferTraditionalChinese(ctx, param.SLocale, param.CLocale) {
				for _, a := range amplayer {
					i18n.TranslateAsTCV2(&a.Arc.Title, &a.Arc.Desc, &a.Arc.TypeName)
				}
			}
			arcOK = true
			return
		})
		if s.canEnableClassBadge(mid, buvid) {
			g.Go(func() (err error) {
				if openCoursePegasusMark, err = s.creative.OpenCoursePegasusMark(ctx, aids); err != nil {
					log.Error("Failed to request OpenCoursePegasusMark: %+v", err)
					err = nil
					return
				}
				return nil
			})
		}
	}
	if len(inlineAids) != 0 {
		g.Go(func() (err error) {
			if haslike, err = s.thumbupDao.HasLike(ctx, buvid, mid, inlineAids); err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			return nil
		})
		g.Go(func() (err error) {
			if mid <= 0 {
				return nil
			}
			if hasCoin, err = s.coin.ArchiveUserCoins(ctx, inlineAids, mid); err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			return nil
		})
		if !s.c.Custom.DisableLikeStat {
			g.Go(func() (err error) {
				if likeState, err = s.thumbupDao.GetLikeStates(ctx, inlineAids); err != nil {
					log.Error("Failed to request GetLikeStates: %+v", err)
					return nil
				}
				return nil
			})
		}
	}
	if len(favAids) != 0 {
		g.Go(func() (err error) {
			favResult, err := s.fav.IsFavVideos(ctx, mid, favAids)
			if err != nil {
				log.Error("%+v", err)
				return nil
			}
			hasFav = favResult
			return nil
		})
	}
	if len(storyAids) != 0 {
		g.Go(func() (err error) {
			if storyamplayer, err = s.storyArcsPlayer(ctx, storyAids); err != nil {
				log.Error("%+v", err)
				err = nil
				return
			}
			for _, a := range storyamplayer {
				storyUpIDs = append(storyUpIDs, a.Arc.Author.Mid)
			}
			arcOK = true
			return nil
		})
	}
	if len(tids) != 0 {
		g.Go(func() (err error) {
			if tagm, err = s.tg.TagsInfoByIDs(ctx, mid, tids); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			return
		})
		if plat == model.PlatIPhone && param.Build > 8820 || plat == model.PlatAndroid && param.Build > 5479999 {
			g.Go(func() (err error) {
				if channelDetailm, err = s.channelDao.Details(ctx, tids); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				return
			})
		}
	}
	if len(roomIDs) != 0 && cdm.ShowLive(param.MobiApp, param.Device, param.Build) {
		g.Go(func() (err error) {
			if rm, err = s.lv.AppMRoom(ctx, roomIDs, mid, param.Platform, param.DeviceName, param.AccessKey, param.ActionKey, param.AppKey, param.Device, param.MobiApp, param.Statistics, buvid, param.Network, param.Build, param.TeenagersMode, param.Appver, param.Filtered, param.HttpsUrlReq, 0); err != nil {
				log.Error("%+v", err)
				err = nil
			}
			for _, r := range rm {
				if r == nil {
					continue
				}
				rmUpIDs = append(rmUpIDs, r.UID)
			}
			return
		})
	}
	// 直播inline卡需要有房间过滤限制
	if len(inlineRoomIDs) != 0 {
		g.Go(func() (err error) {
			if inlinerm, err = s.lv.AppMRoom(ctx, inlineRoomIDs, mid, param.Platform, param.DeviceName, param.AccessKey, param.ActionKey, param.AppKey, param.Device, param.MobiApp, param.Statistics, buvid, param.Network, param.Build, param.TeenagersMode, param.Appver, param.Filtered, param.HttpsUrlReq, 1); err != nil {
				log.Error("%+v", err)
				return nil
			}
			for _, r := range inlinerm {
				if r == nil {
					continue
				}
				inlineRmUpIDs = append(inlineRmUpIDs, r.UID)
			}
			return
		})
	}
	if len(sids) != 0 {
		g.Go(func() (err error) {
			if sm, err = s.bgm.CardsByAids(ctx, sids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(bangumiAids) != 0 {
		g.Go(func() (err error) {
			if episodeSeasonCardm, err = s.bgm.EpCardsFromPgcByAids(ctx, bangumiAids); err != nil {
				log.Error("s.bgm.EpCardsFromPgcByAids err(%+v)", err)
			}
			return nil
		})
	}
	if len(seasonIDs) != 0 {
		g.Go(func() (err error) {
			if seasonm, err = s.bgm.CardsInfoReply(ctx, seasonIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(specialSeasonIds) != 0 {
		g.Go(func() (err error) {
			if pgcSeasonm, err = s.bgm.SeasonBySeasonId(ctx, specialSeasonIds, mid); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(pgcEpids) > 0 {
		g.Go(func() (err error) {
			if pgcCardm, err = s.bgm.EpCardsFromPgcByEpids(ctx, pgcEpids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(epPlayerIDs) != 0 {
		g.Go(func() (err error) {
			if eppm, err = s.bgm.InlineCards(ctx, epPlayerIDs, param.MobiApp, param.Platform, param.Device, param.Build, mid, false, false, false, buvid, nil); err != nil {
				log.Error("%+v", err)
				return nil
			}
			epAids := make([]int64, 0, len(eppm))
			for _, v := range eppm {
				epAids = append(epAids, v.Aid)
			}
			if epHasLike, err = s.thumbupDao.HasLike(ctx, buvid, mid, epAids); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(gameIDs) > 0 {
		g.Go(func() (err error) {
			if gamem, err = s.game.MultiGameInfos(ctx, gameIDs, param.MobiApp, param.Build, gameParams); err != nil {
				log.Error("Failed to request MultiGameInfos: %+v", err)
			}
			return nil
		})
	}
	if len(reservationIds) > 0 {
		g.Go(func() (err error) {
			if reservationm, err = s.actDao.ActReserveCard(ctx, mid, reservationIds); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if hasUpdate && mid > 0 {
		g.Go(func() (err error) {
			if (model.IsIOS(plat) && param.Build > _iosBuild537) || (model.IsAndroid(plat) && param.Build > _androidBuild537) {
				if pgcRemind, err = s.bgm.Remind(ctx, mid); err != nil {
					log.Error("%+v", err)
				}
			} else {
				if update, err = s.bgm.Updates(ctx, mid, now); err != nil {
					log.Error("%+v", err)
				}
			}
			return nil
		})
	}
	if len(metaIDs) != 0 {
		g.Go(func() (err error) {
			if metam, err = s.art.Articles(ctx, metaIDs); err != nil {
				log.Error("%+v", err)
			}
			for _, meta := range metam {
				if meta.Author != nil {
					mtUpIDs = append(mtUpIDs, meta.Author.Mid)
				}
			}
			return nil
		})
	}
	if len(audioIDs) != 0 {
		g.Go(func() (err error) {
			if audiom, err = s.audio.Audios(ctx, audioIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(picIDs) != 0 {
		g.Go(func() (err error) {
			if picm, err = s.bplus.DynamicDetail(ctx, param.Platform, param.MobiApp, param.Device, param.Build, picIDs...); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if vipRenew && mid > 0 {
		g.Go(func() (err error) {
			var (
				platformInt int64
				_ios        = int64(1)
				_ipad       = int64(2)
				_android    = int64(4)
			)
			if model.IsPad(plat) {
				platformInt = _ipad
			} else if model.IsIOS(plat) {
				platformInt = _ios
			} else if model.IsAndroid(plat) {
				platformInt = _android
			}
			if vipRenewReply, err = s.vip.TipsRenew(ctx, param.Build, platformInt, mid); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	// 新频道
	if len(channeIDs) > 0 {
		g.Go(func() (err error) {
			if channelm, err = s.Channels(ctx, channeIDs, mid); err != nil {
				log.Error("%v", err)
			}
			return nil
		})
	}
	if len(gatherOids) > 0 {
		g.Go(func() (err error) {
			if tunnels, err = s.tunnelDao.FeedCards(ctx, param.MobiApp, mid, int64(param.Build), gatherOids); err != nil {
				log.Error("%v", err)
			}
			return nil
		})
	}
	if len(feedCreativeIds) > 0 {
		g.Go(func() (err error) {
			if multiMaterials, err = s.rsc.MultiMaterials(ctx, feedCreativeIds); err != nil {
				log.Error("Failed to MultiMaterials: %+v, %+v", feedCreativeIds, err)
			}
			return nil
		})
	}
	if len(specialIds) > 0 && !s.c.Custom.ResourceDegradeSwitch {
		g.Go(func() (err error) {
			if specialCardm, err = s.specialCardV2(ctx, specialIds); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
	}
	if len(epMaterialReq) > 0 && !s.c.Custom.PGCMaterialDegradeSwitch {
		g.Go(func() (err error) {
			if epMaterialm, err = s.bgm.BatchEpMaterial(ctx, epMaterialReq); err != nil {
				log.Error("Failed to request BatchEpMaterial: %+v", err)
				return nil
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		if noCache {
			is = []card.Handler{}
			return
		}
		if isRcmd {
			count := s.indexCount(plat, abtest)
			s.infoProm.Incr("tianma_backup_with_recommend_cache")
			rs = s.recommendCache(count)
		}
	} else {
		upIDs = append(upIDs, avUpIDs...)
		upIDs = append(upIDs, rmUpIDs...)
		upIDs = append(upIDs, inlineRmUpIDs...)
		upIDs = append(upIDs, mtUpIDs...)
		upIDs = append(upIDs, storyUpIDs...)
		g, ctx = errgroup.WithContext(c)
		if len(upIDs) != 0 {
			g.Go(func() (err error) {
				if cardm, err = s.acc.Cards3GRPC(ctx, upIDs); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				return
			})
			g.Go(func() (err error) {
				if statm, err = s.rel.StatsGRPC(ctx, upIDs); err != nil {
					log.Error("%+v", err)
					err = nil
				}
				return
			})
			if mid > 0 && param.RecsysMode == 0 {
				g.Go(func() error {
					isAtten = s.acc.IsAttentionGRPC(ctx, upIDs, mid)
					return nil
				})
			}
		}
		if err := g.Wait(); err != nil {
			log.Error("Failed to wait: %+v", err)
		}
	}
	isAI = isAI && arcOK
	if moe != nil {
		moePos := s.c.Feed.Index.MoePosition
		if moePos-1 >= 0 && moePos-1 <= len(rs) {
			rs = append(rs[:moePos-1], append([]*ai.Item{{ID: moe.ID, Goto: model.GotoMoe}}, rs[moePos-1:]...)...)
		}
	}
	var cardTotal int
	is = make([]card.Handler, 0, len(rs))
	insert := map[int32]card.Handler{}
	hotAidSet := convertHotAid(s.hotAids)
	canEnableDoubleUGCClickLike := s.canEnableDoubleUGCClickLike(mid, buvid)
	canEnableSingleUGCClickLike := s.canEnableSingleUGCClickLike(mid, buvid)
LOOP:
	for index, r := range rs {
		if r == nil {
			continue
		}
		var (
			main     interface{}
			cardType = cdm.CardType(r.CardType)
		)
		r.SetRequestAt(now)
		r.SetGotoStoryDislikeReason(s.c.Custom.GotoStoryDislikeReason)
		r.SetSingleInlineDbClickLike(canEnableSingleUGCClickLike)
		r.SetDoubleInlineDbClickLike(canEnableDoubleUGCClickLike)
		r.SetOgvHasScore(true)
		r.SetAllowGameBadge(s.c.Custom.AllowGameBadge)
		op := &operate.Card{}
		op.From(cdm.CardGt(r.Goto), r.ID, r.Tid, plat, param.Build, param.MobiApp)
		op.SwitchLargeCoverShow = cdm.SwitchLargeCoverShowAll
		if abtest != nil {
			r.SetManualInline(abtest.ManualInline)
			op.NeedSwitchColumnThreePoint = card.CanEnableSwitchColumnThreePoint(param.MobiApp, param.Build, abtest)
			op.Column = param.Column
			op.ReplaceDislikeTitle = abtest.DislikeText == 1
		}
		// 卡片展示点赞数实验
		// if mid%20 == 11 && ((plat == model.PlatIPhone && param.Build >= 8290) || (plat == model.PlatAndroid && param.Build >= 5360000)) {
		// 	op.FromSwitch(cdm.SwitchFeedIndexLike)
		// }
		// 变化卡片类型

		switch r.Goto {
		case model.GotoAv:
			if isSingleInline(r) && cdm.Columnm[param.Column] == cdm.ColumnSvrSingle {
				cardType = cdm.LargeCoverSingleV9
				if s.c.Feed.Inline != nil {
					setOperateFromInlineConf(op, s.c.Feed.Inline)
				}
				r.DynamicCover = _dynamicCoverInlineAv
				op.HasFav = hasFav
				op.HotAidSet = hotAidSet
				op.HasCoin = hasCoin
			}
			if cdm.Columnm[param.Column] == cdm.ColumnSvrDouble && r.StNewCover == 1 {
				cardType = cdm.SmallCoverV11
			}
		case model.GotoSpecialS, model.GotoGameDownloadS, model.GotoShoppingS:
			//nolint:gomnd
			if r.Style == 2 {
				cardType = cdm.LargeCoverV1
			}
			if isSingleSpecialS(r, param) {
				switch r.SingleSpecialInfo.SpType {
				case "av":
					cardType = cdm.LargeCoverSingleV9
				case "pgc":
					cardType = cdm.LargeCoverSingleV7
				case "live":
					cardType = cdm.LargeCoverSingleV8
				}
			}
		case model.GotoPicture:
			if p, ok := picm[r.ID]; ok {
				// feature PicSwitchStyle
				if (param.MobiApp == "iphone" && param.Device == "phone" && param.Build >= 8290) || (param.MobiApp == "android" && param.Build >= 5360000) {
					if op.SwitchStyle == nil {
						op.SwitchStyle = map[cdm.Switch]struct{}{cdm.SwitchPictureLike: {}}
					} else {
						op.SwitchStyle[cdm.SwitchPictureLike] = struct{}{}
					}
				}
				// 图文卡tag新旧频道开关
				// feature PicIsNewChannel
				if (param.MobiApp == "iphone" && param.Device == "phone" && param.Build > s.c.BuildLimit.NewChannelIOS) || (param.MobiApp == "android" && param.Build > s.c.BuildLimit.NewChannelAndroid) {
					p.IsNewChannel = true
				}
				switch cdm.Columnm[param.Column] {
				case cdm.ColumnSvrSingle:
					//nolint:gomnd
					if len(p.Imgs) < 3 {
						cardType = cdm.OnePicV1
					} else {
						cardType = cdm.ThreePicV1
					}
				case cdm.ColumnSvrDouble:
					//nolint:gomnd
					if len(p.Imgs) < 3 {
						// http层强转了plat，所以这里用mobi_app判断
						// feature OnePicV3
						if (param.MobiApp == "iphone" && param.Build > _iosBuild540) || (plat == model.PlatAndroid && param.Build > _androidBuild540) {
							cardType = cdm.OnePicV3
						} else if (param.MobiApp == "iphone" && param.Build > 8300) || (plat == model.PlatAndroid && param.Build > 5365000) {
							// fature OnePicV2
							cardType = cdm.OnePicV2
						} else {
							cardType = cdm.SmallCoverV2
						}
					} else {
						// feature ThreePicV3
						if (param.MobiApp == "iphone" && param.Build > _iosBuild540) || (plat == model.PlatAndroid && param.Build > _androidBuild540) {
							cardType = cdm.ThreePicV3
						} else {
							cardType = cdm.ThreePicV2
						}
					}
				default:
					stat.MetricDiscardCardTotal.Inc(rowType, r.Goto, r.JumpGoto, string(cardType), "unexpeted_column")
					continue
				}
			} else {
				stat.MetricDiscardCardTotal.Inc(rowType, r.Goto, r.JumpGoto, string(cardType), "picture_resource_not_exist")
				FillDiscard(r.ID, r.Goto, feed.DiscardReasonOther, "图文卡资源获取失败", infoc)
				continue
			}
		case model.GotoGame:
			cardType = cdm.SmallCoverV10
		case model.GotoFollowMode:
			cardType = cdm.Select
		case model.GotoBanner:
			if abtest != nil && abtest.Banner == _newBannerResource {
				cardType = s.bannerCardType(abtest, plat, param.Column)
			}
		case model.GotoLive, model.GotoPlayerLive:
			if s.c.Feed.Inline != nil {
				setOperateFromInlineConf(op, s.c.Feed.Inline)
			}
			func() {
				if isSingleV1Inline(r, param) {
					cardType = cdm.LargeCoverSingleV8
					return
				}
				// feature LiveV9Custom
				if (plat == model.PlatIPhone && param.Build >= 62100000) ||
					(param.MobiApp == "android" && param.Build >= 6210000) ||
					(param.MobiApp == "ipad" && param.Build >= 32000000) ||
					(plat == model.PlatIPad && param.Build >= 63300000) ||
					(param.MobiApp == "iphone_i" && param.Build >= 64400200) ||
					(param.MobiApp == "win") {
					if s.c.V9Custom != nil {
						if val, ok := s.c.V9Custom.LeftBottomBadgeStyle[s.c.V9Custom.LeftBottomBadgeKey]; ok {
							op.LiveLeftBottomBadgeStyle = val
						}
						op.LiveLeftCoverBadgeStyle = s.c.V9Custom.LeftCoverBadgeStyle
					}
					switch cdm.Columnm[param.Column] {
					case cdm.ColumnSvrDouble:
						cardType = cdm.SmallCoverV9
					default:
					}
					if param.MobiApp == "ipad" || plat == model.PlatIPad || param.MobiApp == "win" {
						cardType = cdm.SmallCoverV9
					}
				}
			}()
		case model.GotoBangumi, model.GotoPGC:
			switch cdm.Columnm[param.Column] {
			case cdm.ColumnSvrDouble:
				if op.SwitchStyle == nil {
					op.SwitchStyle = map[cdm.Switch]struct{}{cdm.SwitchPGCHideSubtitle: {}}
				} else {
					op.SwitchStyle[cdm.SwitchPGCHideSubtitle] = struct{}{}
				}
			default:
			}
			if isOgvSmallCover(r, param) {
				cardType = cdm.OgvSmallCover
			}
			if (r.Goto == model.GotoBangumi || r.Goto == model.GotoPGC) && isSingleV1Inline(r, param) {
				if s.c.Feed.Inline != nil {
					setOperateFromInlineConf(op, s.c.Feed.Inline)
				}
				cardType = cdm.LargeCoverSingleV7
			}
		case model.GotoInlineAv:
			if hideInlineCard {
				switch cdm.Columnm[param.Column] {
				case cdm.ColumnSvrDouble:
					cardType = cdm.SmallCoverV2
					r.Goto = model.GotoAv
					// ai ad监控
					if abtest != nil && abtest.AdExp == _aiAdExp {
						s.infoProm.Incr("miss_ai_inline_av")
					}
				default:
				}
			} else {
				// feature InlineAV2
				if (param.MobiApp == "iphone" && param.Build > 10130) || (param.MobiApp == "android" && param.Build > 6045000) {
					// inline2.0样式
					cardType = cdm.LargeCoverV6
					if s.c.Feed.Inline != nil {
						setOperateFromInlineConf(op, s.c.Feed.Inline)
					}
				}
				r.DynamicCover = _dynamicCoverInlineAv
			}
			if isSingleV1Inline(r, param) {
				cardType = cdm.LargeCoverSingleV9
				if s.c.Feed.Inline != nil {
					setOperateFromInlineConf(op, s.c.Feed.Inline)
				}
				r.DynamicCover = _dynamicCoverInlineAv
				op.HasFav = hasFav
				op.HotAidSet = hotAidSet
				op.HasCoin = hasCoin
			}
		case model.GotoInlineAvV2:
			if hideInlineCard {
				switch cdm.Columnm[param.Column] {
				case cdm.ColumnSvrDouble:
					cardType = cdm.SmallCoverV2
					r.Goto = model.GotoAv
					// ai ad监控
					if abtest != nil && abtest.AdExp == _aiAdExp {
						s.infoProm.Incr("miss_ai_inline_av_v2")
					}
				default:
				}
			} else {
				cardType = cdm.LargeCoverV9
				if s.c.Feed.Inline != nil {
					setOperateFromInlineConf(op, s.c.Feed.Inline)
				}
				r.DynamicCover = _dynamicCoverInlineAv
				op.HasFav = hasFav
				op.HotAidSet = hotAidSet
				op.HasCoin = hasCoin
			}
		case model.GotoInlinePGC:
			if hideInlineCard {
				switch cdm.Columnm[param.Column] {
				case cdm.ColumnSvrDouble:
					cardType = cdm.SmallCoverV2
					r.Goto = model.GotoPGC
					s.infoProm.Incr("miss_inline_pgc")
				default:
				}
			} else {
				if s.c.Feed.Inline != nil {
					setOperateFromInlineConf(op, s.c.Feed.Inline)
				}
			}
			if isSingleV1Inline(r, param) {
				cardType = cdm.LargeCoverSingleV7
			}
		case model.GotoInlineBangumi:
			//if s.c.Feed.Inline != nil {
			//	setOperateFromInlineConf(op, s.c.Feed.Inline)
			//}
			//cardType = cdm.LargeCoverSingleV7
		case model.GotoInlineLive:
			if hideInlineCard {
				switch cdm.Columnm[param.Column] {
				case cdm.ColumnSvrDouble:
					cardType = cdm.SmallCoverV2
					r.Goto = model.GotoLive
					s.infoProm.Incr("miss_inline_live")
				default:
				}
			} else {
				if s.c.Feed.Inline != nil {
					setOperateFromInlineConf(op, s.c.Feed.Inline)
				}
			}
			if isSingleV1Inline(r, param) {
				cardType = cdm.LargeCoverSingleV8
			}
		case model.GotoTunnel:
			// feature FeedTunnel
			if param.MobiApp == "android" && (param.Build >= 6090600 && param.Build <= 6091000) {
				t, ok := tunnels[r.ID]
				if !ok {
					continue
				}
				cardType = cdm.SmallCoverV4
				if t.ResourceType == "game" {
					cardType = cdm.SmallCoverV7
				}
			}
		case model.GotoAdPlayer, model.GotoAdInline3D, model.GotoAdInlineEggs, model.GotoAdInline3DV2:
			if r.SingleAdNew == 1 {
				cardType = cdm.CmSingleV1
			}
		}
		// new double
		columnStatus := param.Column
		if len(haslike) == 0 {
			haslike = make(map[int64]int8)
		}
		for aid, state := range epHasLike {
			haslike[aid] = state
		}
		// new double
		h := card.Handle(plat, cdm.CardGt(r.Goto), cardType, columnStatus, r, tagm, isAtten, haslike, statm, cardm, nil)
		if h == nil {
			stat.MetricDiscardCardTotal.Inc(rowType, r.Goto, r.JumpGoto, string(cardType), "unexpected_card_template")
			FillDiscard(r.ID, r.Goto, feed.DiscardReasonUnexpectedCardTemplate, "找不到对应的卡片模板", infoc)
			continue
		}
		switch r.Goto {
		case model.GotoAv, model.GotoUpRcmdAv, model.GotoPlayer, model.GotoInlineAv, model.GotoInlineAvV2:
			if !arcOK {
				if r.Archive != nil {
					i18n.TranslateAsTCV2(&r.Archive.Title, &r.Archive.Desc, &r.Archive.TypeName)
					amplayer = map[int64]*arcgrpc.ArcPlayer{r.Archive.Aid: {Arc: r.Archive}}
				}
				if r.Tag != nil {
					tagm = map[int64]*taggrpc.Tag{r.Tag.Id: r.Tag}
					op.Tid = r.Tag.Id
				}
			}
			if r.IconType != 0 {
				op.GotoIcon = ConstructGotoIcon(s.c.Feed.StoryIcon)
			}
			if r.JumpGoto == model.GotoVerticalAv {
				op.GotoIcon = ConstructGotoIcon(s.c.Feed.StoryIcon)
				if (param.MobiApp == "iphone" && param.Build > 10030) ||
					(param.MobiApp == "android" && param.Build > 6025500) ||
					(param.MobiApp == "android_i" && param.Build >= 6790300) ||
					(param.MobiApp == "iphone_i" && param.Build >= 67900200) ||
					r.StNewCover == 1 {
					if a, ok := storyamplayer[r.ID]; ok {
						main = storyamplayer
						op.TrackID = r.TrackID
						if isPGCArchive(r, a.Arc) {
							op.RedirectURL = a.Arc.RedirectURL
							r.JumpGoto = ""
						}
					}
				} else {
					r.JumpGoto = ""
				}
			} else if a, ok := amplayer[r.ID]; ok {
				main = amplayer
				op.TrackID = r.TrackID
				if isPGCArchive(r, a.Arc) {
					op.RedirectURL = a.Arc.RedirectURL
				}
			}
			if channelDetailm != nil {
				if cd, ok := channelDetailm[op.Tid]; ok && cd != nil {
					var (
						channelID   = cd.GetChannelId()
						channelName = cd.GetChannelName()
					)
					if channelID == op.Tid {
						op.Channel = &operate.Channel{
							ChannelID:   channelID,
							ChannelName: channelName,
						}
					}
				}
			}
			if plat == model.PlatIPhone && param.Build > 8290 || plat == model.PlatAndroid && param.Build > 5365000 ||
				param.MobiApp == "ipad" && param.Build > 12520 || plat == model.PlatIPad && param.Build >= 63100000 {
				op.Switch = cdm.SwitchCooperationShow
			} else {
				op.Switch = cdm.SwitchCooperationHide
			}
			// infoc
			if r.CoverGif != "" {
				isShowGif := s.showCardGif(_aiGif, abtest)
				if isShowGif {
					op.GifCover = r.CoverGif
					aiCardGifCount++
					r.DynamicCover = _dynamicCoverAiGif
				}
				if infoc != nil {
					if infoc.IsGifCover == nil {
						infoc.IsGifCover = map[int64]int{}
					}
					infoc.IsGifCover[r.ID] = 0
					if isShowGif {
						infoc.IsGifCover[r.ID] = 1
					}
				}
			}
		case model.GotoLive, model.GotoPlayerLive:
			main = rm
			op.Network = param.Network
			if isSingleV1Inline(r, param) {
				main = inlinerm
			}
		case model.GotoInlineLive:
			main = inlinerm
			op.Network = param.Network
		case model.GotoBangumi:
			switch cardType {
			case cdm.OgvSmallCover:
				main = episodeSeasonCardm
			case cdm.LargeCoverSingleV7:
				main = eppm
			default:
				main = sm
				if r.Tag != nil {
					tagm = map[int64]*taggrpc.Tag{r.Tag.Id: r.Tag}
					op.Tid = r.Tag.Id
				}
				if a, ok := amplayer[r.ID]; ok {
					op.Desc = a.Arc.TypeName
				}
			}
		case model.GotoPlayerBangumi, model.GotoInlinePGC, model.GotoInlineBangumi:
			op.SwitchLargeCoverShow = cdm.SwitchLargeCoverShowBottom
			main = eppm
		case model.GotoPGC:
			main = seasonm
			if isSingleV1Inline(r, param) {
				main = eppm
			}
		case model.GotoSpecial, model.GotoSpecialS, model.GotoSpecialB:
			cardm := convertSpecialCardmToCardm(specialCardm)
			if s.teenagerSpecialCondition(param, r) {
				cardm = s.teenagersSpecialCard(c)
			}
			for id, card := range cardm {
				// gif count
				if card.GifCover != "" && r.StaticCover == 0 {
					s.allGifState(_rcmdGif, abtest)
					abtestHideGuidance(abtest, _hideGuidanceByOperateGif)
				}
				// gif count
				specialm[id] = card
			}
			op = specialm[r.ID]
			if op != nil {
				op.Network = param.Network
				// feature SpecialSwitchStyle
				if (param.MobiApp == "android" && param.Build > 5455000) || (param.MobiApp == "iphone" && param.Build > 8790) ||
					(param.MobiApp == "iphone_i" && param.Build >= 64400200) {
					switch op.Goto {
					case cdm.GotoAv, cdm.GotoLive, cdm.GotoArticle, cdm.GotoBangumi, cdm.GotoPGC:
						main = map[cdm.Gt]interface{}{cdm.GotoAv: amplayer, cdm.GotoLive: rm, cdm.GotoArticle: metam, cdm.GotoBangumi: pgcCardm, cdm.GotoPGC: pgcSeasonm}
					default:
					}
					if op.SwitchStyle == nil {
						op.SwitchStyle = map[cdm.Switch]struct{}{cdm.SwitchSpecialInfo: {}}
					} else {
						op.SwitchStyle[cdm.SwitchSpecialInfo] = struct{}{}
					}
				}
				if r.StaticCover == 1 {
					op.GifCover = ""
				}
				if !s.showCardGif(_rcmdGif, abtest) || op.GifCover == "" {
					op.GifCover = ""
					// ai ad监控
					if abtest != nil && abtest.AdExp == _aiAdExp {
						s.infoProm.Incr("miss_ai_rcmd_gif")
					}
				}
			}
		case model.GotoBangumiRcmd:
			if (model.IsIOS(plat) && param.Build > _iosBuild537) || (model.IsAndroid(plat) && param.Build > _androidBuild537) {
				main = pgcRemind
			} else {
				main = update
			}
		case model.GotoBanner:
			if inlineBannersWithoutIPadSet.Has(abtest.ResourceID) && common.IsInlineBanner(
				param.MobiApp, int64(param.Build)) {
				main = &card.BannerInline{
					Archive: amplayer,
					PGC:     eppm,
					Live:    inlinerm,
				}
			}
			r.Banners = banners
			op.FromBanner(banners, version)
			op.HasFav = hasFav
			op.HasCoin = hasCoin
			if s.c.Feed.Inline != nil {
				setOperateFromInlineConf(op, s.c.Feed.Inline)
			}
			if infoc != nil {
				infoc.BannerHash = version
			}
		case model.GotoConvergeAi:
			main = map[cdm.Gt]interface{}{cdm.GotoAv: amplayer, cdm.GotoLive: rm, cdm.GotoArticle: metam}
			switch r.Goto {
			case model.GotoConvergeAi:
				op = convergem[r.ID+_convergeAi]
			default:
				op = convergem[r.ID]
			}
		case model.GotoArticleS:
			main = metam
		case model.GotoAudio:
			main = audiom
		case model.GotoChannelRcmd:
			main = amplayer
			op = followm[r.ID]
		case model.GotoLiveUpRcmd:
			main = liveUpm
		case model.GotoMoe:
			main = moe
		case model.GotoPicture:
			main = picm
		case model.GotoAdAv, model.GotoAdInlineAv:
			var missAds []*cm.AdInfo
			op.TrackID = r.TrackID
			main = amplayer
			for _, ad := range r.Ads {
				if ad.CreativeStyle == 2 || ad.CreativeStyle == 4 { // "creative_style": 1,  // 1 静态图文  2 gif动态图文  3 静态视频  4 inline 广告位播放视频
					if !s.showCardGif(_adGif, abtest) {
						missAds = append(missAds, ad)
						s.infoProm.Incr("missAd")
						if infoc != nil {
							infoc.AdPkCode = append(infoc.AdPkCode, _adPkGifCard)
						}
						// ai ad监控
						if abtest != nil && abtest.AdExp == _aiAdExp {
							s.infoProm.Incr("miss_ai_biz_gif")
						}
						stat.MetricDiscardCardTotal.Inc(rowType, r.Goto, r.JumpGoto, getCardType(h), "skip_gif_card")
						continue
					} else {
						adCardGifCount++
						s.infoProm.Incr("hitAd")
						//nolint:gomnd
						switch ad.CreativeStyle {
						case 2:
							r.DynamicCover = _dynamicCoverAdGif
						case 4:
							r.DynamicCover = _dynamicCoverAdInline
						default:
						}
					}
				}
				r.Ad = ad
				op.FromAdAv(ad)
				break
			}
			s.AdCardDataBus(c, buvid, mid, _adCardResistReasonGif, missAds, r.Ad, param.IsMelloi)
		case model.GotoAdLive:
			main = rm
			if s.c.V9Custom != nil {
				if val, ok := s.c.V9Custom.LeftBottomBadgeStyle[s.c.V9Custom.LeftBottomBadgeKey]; ok {
					op.LiveLeftBottomBadgeStyle = val
				}
				op.LiveLeftCoverBadgeStyle = s.c.V9Custom.LeftCoverBadgeStyle
			}
			for _, ad := range r.Ads {
				// 广告直播小卡暂时不支持gif
				r.Ad = ad
				op.FromAdLive(ad)
				break
			}
		case model.GotoAdInlineLive:
			main = inlinerm
			for _, ad := range r.Ads {
				// 广告直播inline卡暂时不支持gif
				r.Ad = ad
				op.FromAdLiveInLine(ad)
				break
			}
		case model.GotoAdPgc:
			main = pgcCardm
			for _, ad := range r.Ads {
				r.Ad = ad
				break
			}
		case model.GotoAdInlinePgc:
			main = eppm
			for _, ad := range r.Ads {
				r.Ad = ad
				break
			}
		case model.GotoAdWebS, model.GotoAdWeb, model.GotoAdPlayer, model.GotoAdInlineGesture, model.GotoAdInline360,
			model.GotoAdWebGif, model.GotoAdInlineChoose, model.GotoAdDynamic, model.GotoAdInlineChooseTeam,
			model.GotoAdWebGifReservation, model.GotoAdPlayerReservation, model.GotoAdInline3D, model.GotoAdInlineEggs,
			model.GotoAdInline3DV2:
			const (
				_delAdIndex = 6
			)
			var missAds []*cm.AdInfo
			for _, ad := range r.Ads {
				if abtest != nil && abtest.IsNewAdBigCard == _newAdBigCard {
					// ad big card
					_, bannerok := specialCardIndex[_bannerCard]
					_, bigcard := _delAdCard[ad.CardType]
					if ad.CardIndex <= _delAdIndex && bannerok && bigcard {
						s.AdCardDataBus(c, buvid, mid, _adCardResistReasonBanner, []*cm.AdInfo{ad}, nil, param.IsMelloi)
						if infoc != nil {
							infoc.AdPkCode = append(infoc.AdPkCode, _adPkBigCard)
						}
						// ai ad监控
						if abtest != nil && abtest.AdExp == _aiAdExp {
							s.infoProm.Incr("miss_ai_biz_banner")
						}
						continue LOOP
					}
				} else {
					// old del ad card
					//nolint:gomnd
					if adCardCount >= 2 {
						_, bannerok := specialCardIndex[_bannerCard]
						_, bigcard := _delAdCard[ad.CardType]
						if cindex, adok := specialCardIndex[_adCard]; adok && bannerok && cindex == ad.CardIndex && bigcard {
							s.AdCardDataBus(c, buvid, mid, _adCardResistReasonBanner, []*cm.AdInfo{ad}, nil, param.IsMelloi)
							if infoc != nil {
								infoc.AdPkCode = append(infoc.AdPkCode, _adPkBigCard)
							}
							continue LOOP
						}
					}
				}
				if ad.CreativeStyle == 2 || ad.CreativeStyle == 4 { // "creative_style": 1,  // 1 静态图文  2 gif动态图文  3 静态视频  4 inline 广告位播放视频
					if !s.showCardGif(_adGif, abtest) {
						missAds = append(missAds, ad)
						s.infoProm.Incr("missAd")
						// ai ad监控
						if abtest != nil && abtest.AdExp == _aiAdExp {
							s.infoProm.Incr("miss_ai_biz_gif")
						}
						stat.MetricDiscardCardTotal.Inc(rowType, r.Goto, r.JumpGoto, getCardType(h), "skip_gif_card")
						continue
					} else {
						adCardGifCount++
						s.infoProm.Incr("hitAd")
						//nolint:gomnd
						switch ad.CreativeStyle {
						case 2:
							r.DynamicCover = _dynamicCoverAdGif
						case 4:
							r.DynamicCover = _dynamicCoverAdInline
						}
					}
				}
				r.Ad = ad
				main = ad
				break
			}
			s.AdCardDataBus(c, buvid, mid, _adCardResistReasonGif, missAds, r.Ad, param.IsMelloi)
		case model.GotoFollowMode:
			var (
				title  string
				desc   string
				button []string
			)
			if s.c.Feed.Index.FollowMode != nil && s.c.Feed.Index.FollowMode.Card != nil {
				title = s.c.Feed.Index.FollowMode.Card.Title
				desc = s.c.Feed.Index.FollowMode.Card.Desc
				button = s.c.Feed.Index.FollowMode.Card.Button
			}
			op.FromFollowMode(title, desc, button)
		case model.GotoVipRenew:
			main = vipRenewReply
			op.FromVipRenew(vipRenewReply)
		case model.GotoAvConverge, model.GotoMultilayerConverge:
			tmp := avconvergem[r.ID]
			op.FromAvConvergeCard(tmp)
			if _, ok := amplayer[op.ID]; ok {
				main = amplayer
			}
			if r.Tag != nil {
				tagm = map[int64]*taggrpc.Tag{r.Tag.Id: r.Tag}
				op.Tid = r.Tag.Id
			}
			if channelDetailm != nil {
				if cd, ok := channelDetailm[op.Tid]; ok && cd != nil {
					var (
						channelID   = cd.GetChannelId()
						channelName = cd.GetChannelName()
					)
					if channelID == op.Tid {
						op.Channel = &operate.Channel{
							ChannelID:   channelID,
							ChannelName: channelName,
						}
					}
				}
			}
		case model.GotoIntroduction:
			if r.ConvergeInfo != nil {
				op.Title = r.ConvergeInfo.Title
			}
		case model.GotoSpecialChannel:
			op = specialm[r.ID]
			main = channelm
		case model.GotoTunnel:
			main = tunnels[r.ID]
		case model.GotoNewTunnel:
			main = tunnels
		case model.GotoBigTunnel:
			func() {
				if r.BigTunnelObject == "" {
					main = tunnels[r.ID]
					return
				}
				bigTunnelObject := &ai.BigTunnelObject{}
				if err := json.Unmarshal([]byte(r.BigTunnelObject), &bigTunnelObject); err != nil {
					log.Error("Failed to unmarshal big tunnel object: %+v", err)
					return
				}
				id, _ := strconv.ParseInt(bigTunnelObject.Resource, 10, 64)
				if s.c.Feed.Inline != nil {
					op.InlinePlayIcon = operate.InlinePlayIcon{
						IconDrag:     s.c.Feed.Inline.IconDrag,
						IconDragHash: s.c.Feed.Inline.IconDragHash,
						IconStop:     s.c.Feed.Inline.IconStop,
						IconStopHash: s.c.Feed.Inline.IconStopHash,
					}
				}
				op.HasFav = hasFav
				op.HasCoin = hasCoin
				if s.c.Feed.Inline != nil {
					setOperateFromInlineConf(op, s.c.Feed.Inline)
				}
				main = &card.BigTunnelInline{
					Tunnel:  tunnels[r.ID],
					Archive: amplayer[id],
					PGC:     eppm[int32(id)],
					Live:    inlinerm[id],
				}
			}()
		case model.GotoAiStory:
			main = storyamplayer
		case model.GotoGame:
			main = gamem
		default:
			log.Warn("v2 unexpected goto(%s) %+v", r.Goto, r)
			stat.MetricDiscardCardTotal.Inc(rowType, r.Goto, r.JumpGoto, getCardType(h), "unexpected_goto")
			FillDiscard(r.ID, r.Goto, feed.DiscardReasonUnexpectedGoto, "", infoc)
			continue
		}
		if op != nil {
			op.Plat = plat
			op.Build = param.Build
			op.MobiApp = param.MobiApp
			if abtest != nil {
				switch abtest.RcmdReason {
				case _newRcmdReason:
					if op.SwitchStyle == nil {
						op.SwitchStyle = map[cdm.Switch]struct{}{cdm.SwitchNewReason: {}}
					} else {
						op.SwitchStyle[cdm.SwitchNewReason] = struct{}{}
					}
				case _newRcmdReasonV2:
					if op.SwitchStyle == nil {
						op.SwitchStyle = map[cdm.Switch]struct{}{cdm.SwitchNewReasonV2: {}}
					} else {
						op.SwitchStyle[cdm.SwitchNewReasonV2] = struct{}{}
					}
				}
			}
		}
		if material, ok := multiMaterials[r.CreativeId]; ok && material.GifCover != "" &&
			s.showCardGif(_rcmdGif, abtest) && r.StaticCover == 0 {
			r.DynamicCover = _dynamicCoverRcmdGif
			r.SetAllowGIF()
		}
		if epm, ok := epMaterialm[r.CreativeId]; ok && epm.GifCover != "" && s.showCardGif(_rcmdGif, abtest) &&
			r.StaticCover == 0 {
			r.DynamicCover = _dynamicCoverRcmdGif
			r.SetAllowGIF()
		}
		// 运营inline 卡片goto为inline_av、inline_live、inline_pgc、inline_av_v2，通过dalao_uniq_id>0可判断为运营卡片
		if inlineGotoSet.Has(r.Goto) && r.PosRecID > 0 {
			r.DynamicCover = _dynamicCoverRcmdInline
		}
		if r.PosRecID > 0 && isSingleInline(r) {
			r.DynamicCover = _dynamicCoverRcmdInline
		}
		feedCtx := buildFeedCtx(c, param, mid, isAtten)
		addFeatureGates(feedCtx, op, abtest)
		materials := &Materials{
			Archive:                  amplayer,
			StoryArchive:             storyamplayer,
			Picture:                  picm,
			Tag:                      tagm,
			AccountCard:              cardm,
			Channel:                  channelDetailm,
			RelationStatMid:          statm,
			IsAttention:              isAtten,
			Article:                  metam,
			Room:                     rm,
			InlineRoom:               inlinerm,
			Season:                   seasonm,
			SeasonByAid:              sm,
			InlinePGC:                eppm,
			HasLike:                  haslike,
			Banner:                   banners,
			BannerVersion:            version,
			Remind:                   pgcRemind,
			Update:                   update,
			Tunnel:                   tunnels,
			Vip:                      vipRenewReply,
			HasFavourite:             hasFav,
			HotAidSet:                hotAidSet,
			HasCoin:                  hasCoin,
			PgcEpisodeByAids:         episodeSeasonCardm,
			PgcEpisodeByEpids:        pgcCardm,
			LiveLeftBottomBadgeKey:   s.c.V9Custom.LeftBottomBadgeKey,
			LiveLeftBottomBadgeStyle: s.c.V9Custom.LeftBottomBadgeStyle,
			LiveLeftCoverBadgeStyle:  s.c.V9Custom.LeftCoverBadgeStyle,
			MultiMaterials:           multiMaterials,
			Specials:                 specialm,
			Game:                     gamem,
			Reservation:              reservationm,
			SpecialCard:              specialCardm,
			PgcSeason:                pgcSeasonm,
			OpenCourseMark:           openCoursePegasusMark,
			LikeStatState:            likeState,
			EpMaterial:               epMaterialm,
		}
		r.SetDynamicCoverInfoc(r.DynamicCover)
		newHandler, ok, needThreePoint := s.constructHandler(h, feedCtx, r, main, materials, op, rowType, mid, buvid, index, isAI, infoc)
		if !ok {
			continue
		}
		h = newHandler
		switch r.Goto {
		case model.GotoAdAv, model.GotoAdWebS, model.GotoAdWeb, model.GotoAdPlayer, model.GotoAdInlineGesture,
			model.GotoAdInline360, model.GotoAdInlineLive, model.GotoAdWebGif, model.GotoAdInlineChoose,
			model.GotoAdLive, model.GotoAdInlineChooseTeam, model.GotoAdDynamic, model.GotoAdInlineAv,
			model.GotoAdWebGifReservation, model.GotoAdPlayerReservation, model.GotoAdInline3D, model.GotoAdPgc,
			model.GotoAdInlinePgc, model.GotoAdInlineEggs, model.GotoAdInline3DV2:
			// 判断结果列表长度，如果列表的末尾不是广告位，则放到插入队列里
			if r.Ad != nil {
				if int32(len(is)) != r.Ad.CardIndex-1 {
					insert[r.Ad.CardIndex-1] = h
					// 插入队列后一定要continue，否则就直接加到队列末尾了
					continue
				}
			}
		}
		is, cardTotal = s.appendItem(plat, is, h, param.Column, cardTotal, param.Build, param.MobiApp, abtest, param.DisableRcmd, needThreePoint)
		// 从插入队列里获取广告
		if h, ok := insert[int32(len(is))]; ok {
			is, cardTotal = s.appendItem(plat, is, h, param.Column, cardTotal, param.Build, param.MobiApp, abtest, param.DisableRcmd, needThreePoint)
		}
	}
	s.infoProm.Incr(fmt.Sprintf("cover_gif_%d", (cardGifCount + aiCardGifCount + adCardGifCount)))
	// 双列末尾卡片去空窗
	if !model.IsPad(plat) {
		if cdm.Columnm[param.Column] == cdm.ColumnSvrDouble {
			if len(is) > 0 && cardTotal%2 == 1 {
				statDiscardByCardLen(rowType, infoc, is[len(is)-1])
			}
			is = is[:len(is)-cardTotal%2]
		}
	} else if abtest != nil && abtest.IpadHDThreeColumn == 1 {
		//nolint:gomnd
		if cardTotal%3 == 2 {
			statDiscardByCardLen(rowType, infoc, is[len(is)-2])
		}
		if cardTotal%3 == 1 {
			statDiscardByCardLen(rowType, infoc, is[len(is)-1])
		}
		is = is[:len(is)-cardTotal%3]
	} else {
		// 复杂的ipad去空窗逻辑
		//nolint:gomnd
		if cardTotal%4 == 3 {
			//nolint:gomnd
			if is[len(is)-2].Get().CardLen == 2 {
				statDiscardByCardLen(rowType, infoc, is[len(is)-2:]...)
				is = is[:len(is)-2]
			} else {
				statDiscardByCardLen(rowType, infoc, is[len(is)-3:]...)
				is = is[:len(is)-3]
			}
		} else if cardTotal%4 == 2 {
			//nolint:gomnd
			if is[len(is)-1].Get().CardLen == 2 {
				statDiscardByCardLen(rowType, infoc, is[len(is)-1:]...)
				is = is[:len(is)-1]
			} else {
				statDiscardByCardLen(rowType, infoc, is[len(is)-2:]...)
				is = is[:len(is)-2]
			}
		} else if cardTotal%4 == 1 {
			statDiscardByCardLen(rowType, infoc, is[len(is)-1:]...)
			is = is[:len(is)-1]
		}
	}
	if len(is) == 0 {
		is = []card.Handler{}
		return
	}
	return
}

func isPGCArchive(r *ai.Item, a *arcgrpc.Arc) bool {
	return r.Goto == model.GotoAv && a.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && a.RedirectURL != ""
}

func convertSpecialCardmToCardm(asc map[int64]*resourceV2grpc.AppSpecialCard) map[int64]*operate.Card {
	out := make(map[int64]*operate.Card, len(asc))
	for _, v := range asc {
		op := &operate.Card{}
		op.FromAppSpecialCard(v)
		out[v.Id] = op
	}
	return out
}

func (s *Service) rcmdSingleInline(ctx context.Context, param *feed.IndexParam) int8 {
	if cdm.Columnm[param.Column] != cdm.ColumnSvrSingle {
		return 0
	}
	if !feature.GetBuildLimit(ctx, "service.SingleInline", nil) {
		return 0
	}
	return 1
}

func addFeatureGates(ctx cardschema.FeedContext, op *operate.Card, abtest *feed.Abtest) {
	if op == nil || abtest == nil {
		return
	}
	if op.NeedSwitchColumnThreePoint {
		ctx.FeatureGates().EnableFeature(cardschema.FeatureSwitchColumnThreePoint)
	}
	if abtest.DislikeExp == 1 {
		ctx.FeatureGates().EnableFeature(cardschema.FeatureNewDislike)
	}
	if ctx.IndexParam().IsCloseRcmd() == 1 {
		ctx.FeatureGates().EnableFeature(cardschema.FeatureCloseRcmd)
	}
	if abtest.DislikeText == 1 {
		ctx.FeatureGates().EnableFeature(cardschema.FeatureDislikeText)
	}
	if abtest.SingleRcmdReason == 1 {
		ctx.FeatureGates().EnableFeature(cardschema.FeatureSingleRcmdReason)
	}
	ctx.FeatureGates().SetFeatureState(cardschema.FeatureLiveContentMode, abtest.LiveContentMode)
}

func (s *Service) constructHandler(h card.Handler, feedCtx cardschema.FeedContext, r *ai.Item, main interface{}, materials *Materials, op *operate.Card, rowType string, mid int64, buvid string, index int, isAI bool, infoc *feed.Infoc) (card.Handler, bool, bool) {
	if NgMergeCardSet.Has(cardKey(string(h.Get().CardType), r.Goto)) {
		newhandler, ok := s.buildCardNG(feedCtx, string(h.Get().CardType), r.Goto, r, int64(index), materials, infoc)
		if !ok {
			stat.MetricDiscardCardTotal.Inc(rowType, r.Goto, r.JumpGoto, getCardType(h), "ng_build_card_failed")
			return nil, false, false
		}
		return newhandler, true, false
	}
	if err := h.From(main, op); err != nil {
		log.Error("Fail to build card, Context={rowType=%s cardType=%s goto=%s jumpGoto=%s} Error={%+v}",
			rowType, getCardType(h), r.Goto, r.JumpGoto, err)
		card.StatBuildCardErr(err, rowType, r.Goto, r.JumpGoto, getCardType(h))
		FillDiscard(r.ID, r.Goto, feed.DiscardReasonCannotBuildCard, err.Error(), infoc)
		return nil, false, false
	}
	// 卡片不正常要continue
	if !h.Get().Right {
		stat.MetricDiscardCardTotal.Inc(rowType, r.Goto, r.JumpGoto, getCardType(h), "unexpected_build_card_failed")
		FillDiscard(r.ID, r.Goto, feed.DiscardReasonCardIsNotNormal, "", infoc)
		return nil, false, false
	}
	if s.matchNGBuilder(mid, buvid, string(h.Get().CardType), r.Goto, r.PosRecID, isAI) {
		func() {
			newhandler, ok := s.buildCardNG(feedCtx, string(h.Get().CardType), r.Goto, r, int64(index), materials, infoc)
			if !ok {
				log.Error("Failed to buildCardNG, %s, %s, %d, item: %+v", string(h.Get().CardType),
					feedCtx.Device().RawMobiApp(), feedCtx.Device().Build(), r)
				return
			}
			equal, left, right, err := isHandlerEqual(h, newhandler)
			if err != nil {
				log.Error("Failed to compare: %+v", err)
				return
			}
			if !equal {
				log.Error("The result of card matching is abnormal, mid: %d, buvid: %s, cardType: %s, cardGoto: %s, left: %s, right: %s", mid, buvid, string(h.Get().CardType), string(h.Get().CardGoto), left, right)
				return
			}
			newhandler.Get().ThreePoint = nil
			newhandler.Get().ThreePointV2 = nil
			h = newhandler
		}()
	}
	return h, true, true
}

func newTunnelAddTo(msgIDS string, gatherOids [][]int64) [][]int64 {
	slots := strings.Split(msgIDS, ",")
	for _, slot := range slots {
		oids := strings.Split(slot, "|")
		gatherOid := make([]int64, 0, len(oids))
		for _, oidStr := range oids {
			oid, err := strconv.ParseInt(oidStr, 10, 64)
			if err != nil {
				log.Error("Failed to parse msg id: %q, ids: %q: %+v", oidStr, msgIDS, err)
				continue
			}
			gatherOid = append(gatherOid, oid)
		}
		gatherOids = append(gatherOids, gatherOid)
	}
	return gatherOids
}

func (s *Service) bannerCardType(abtest *feed.Abtest, plat int8, column cdm.ColumnStatus) cdm.CardType {
	if inlineBannersSet.Has(abtest.ResourceID) {
		if model.IsPad(plat) {
			return cdm.BannerIPadV8
		}
		switch cdm.Columnm[column] {
		case cdm.ColumnSvrSingle:
			return cdm.BannerSingleV8
		case cdm.ColumnSvrDouble:
			return cdm.BannerV8
		default:
			log.Error("Failed to match column: %d", column)
			return ""
		}
	}
	return ""
}

func bigTunnelAddTo(r *ai.Item, aids []int64, epPlayerIDs []int32, inlineRoomIDs []int64, abtest *feed.Abtest) ([]int64, []int32, []int64) {
	if r.BigTunnelObject == "" {
		return aids, epPlayerIDs, inlineRoomIDs
	}
	var tunnelObject *ai.BigTunnelObject
	if err := json.Unmarshal([]byte(r.BigTunnelObject), &tunnelObject); err != nil {
		log.Error("Failed to unmarshal big tunnel object: %+v", err)
		return aids, epPlayerIDs, inlineRoomIDs
	}
	id, err := strconv.ParseInt(tunnelObject.Resource, 10, 64)
	if err != nil {
		log.Error("Failed to parse tunnel object resource: %+v", errors.WithStack(err))
		return aids, epPlayerIDs, inlineRoomIDs
	}
	switch tunnelObject.Type {
	case "ugc":
		aids = append(aids, id)
		abtestHideGuidance(abtest, _hideGuidanceByInline)
	case "pgc":
		epPlayerIDs = append(epPlayerIDs, int32(id))
		abtestHideGuidance(abtest, _hideGuidanceByInline)
	case "live":
		inlineRoomIDs = append(inlineRoomIDs, id)
		abtestHideGuidance(abtest, _hideGuidanceByInline)
	case "image":
	default:
		log.Warn("Unknown tunnel object type: %s", tunnelObject.Type)
	}
	return aids, epPlayerIDs, inlineRoomIDs
}

func ConstructGotoIcon(icon map[string]*cdm.GotoIcon) map[int64]*cdm.GotoIcon {
	out := make(map[int64]*cdm.GotoIcon, len(icon))
	for key, value := range icon {
		iconType, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			continue
		}
		out[iconType] = value
	}
	return out
}

//nolint:gocognit
func (s *Service) appendItem(plat int8, rs []card.Handler, h card.Handler, column cdm.ColumnStatus, cardTotal, build int, mobiApp string, abtest *feed.Abtest, isCloseRcmd int, needThreePoint bool) (is []card.Handler, total int) {
	var dislikeExp int
	if abtest != nil {
		dislikeExp = abtest.DislikeExp
	}
	if needThreePoint {
		if abtest != nil && abtest.ThreePoint == _newThreePoint {
			h.Get().ThreePointFromV3(mobiApp, build, dislikeExp)
		} else {
			switch h.Get().CardGoto {
			case cdm.CardGotoAiStory:
				// story卡片并且实验开关开启，story卡才能展示三点按钮
				if abtest != nil && abtest.StoryThreePoint {
					h.Get().ThreePointFrom(mobiApp, build, dislikeExp, abtest, column, isCloseRcmd)
				}
			default:
				h.Get().ThreePointFrom(mobiApp, build, dislikeExp, abtest, column, isCloseRcmd)
			}
		}
	}
	if !model.IsPad(plat) {
		// 双列大小卡换位去空窗
		if cdm.Columnm[column] == cdm.ColumnSvrDouble {
			// 通栏卡
			if h.Get().CardLen == 0 {
				if cardTotal%2 == 1 {
					is = card.SwapTwoItem(rs, h)
				} else {
					is = append(rs, h)
				}
			} else {
				is = append(rs, h)
			}
		} else {
			is = append(rs, h)
		}
	} else {
		// ipad卡片不展示标签
		h.Get().DescButton = nil
		// ipad大小卡换位去空窗
		//nolint:gomnd
		if h.Get().CardLen == 0 {
			// 通栏卡
			//nolint:gomnd
			if cardTotal%4 == 3 {
				is = card.SwapFourItem(rs, h)
			} else if cardTotal%4 == 2 {
				//nolint:gomnd
				//nolint:gomnd
				if len(rs) < 2 {
					is = card.SwapTwoItem(rs, h)
				} else {
					is = card.SwapThreeItem(rs, h)
				}
			} else if cardTotal%4 == 1 {
				is = card.SwapTwoItem(rs, h)
			} else {
				is = append(rs, h)
			}
		} else if h.Get().CardLen == 2 {
			// 半栏卡
			//nolint:gomnd
			if cardTotal%4 == 3 {
				is = card.SwapTwoItem(rs, h)
			} else if cardTotal%4 == 2 {
				is = append(rs, h)
			} else if cardTotal%4 == 1 {
				is = card.SwapTwoItem(rs, h)
			} else {
				is = append(rs, h)
			}
		} else {
			is = append(rs, h)
		}
	}
	total = cardTotal + h.Get().CardLen
	return
}

func (s *Service) Converge(c context.Context, mid int64, plat int8, param *feed.ConvergeParam, buvid string, now time.Time) (is []card.Handler,
	converge *operate.Card, isAi bool, list *ai.ConvergeInfoV2, respCode int, err error) {
	var (
		rs []*ai.Item
	)
	cardm, _, _, _ := s.convergeCard(c, 0, param.ID)
	converge, ok := cardm[param.ID]
	if !ok && param.ID > _convergeAi {
		cardID := param.ID - _convergeAi
		if list, respCode, err = s.rcmd.ConvergeList(c, plat, buvid, mid, cardID, param.Build, param.DisplayID, param.ConvergeParam, 0, now); err != nil {
			log.Error("%+v", err)
			err = nil
			is = []card.Handler{}
			return
		} else {
			isAi = true
			for _, item := range list.Items {
				switch item.Goto {
				case model.GotoAv:
					rs = append(rs, item)
				}
			}
			converge = &operate.Card{
				Title:    list.Title,
				Goto:     cdm.Gt(cdm.CardGotoConverge),
				Param:    strconv.FormatInt(param.ID, 10),
				CardGoto: cdm.CardGotoConvergeAi,
			}
		}
	} else if converge != nil {
		rs = make([]*ai.Item, 0, len(converge.Items))
		for _, item := range converge.Items {
			rs = append(rs, &ai.Item{ID: item.ID, Goto: string(item.CardGoto)})
		}
	}
	indexParam := &feed.IndexParam{
		MobiApp: param.MobiApp,
		Device:  param.Device,
		Build:   param.Build,
	}
	is, _ = s.dealItem2(c, mid, "", plat, rs, indexParam, false, true, false, now, nil, nil)
	for _, item := range is {
		// 运营tab页没有不感兴趣
		item.Get().ThreePointWatchLater()
	}
	return
}

func (s *Service) AvConverge(c context.Context, mid int64, plat int8, buvid string, param *feed.ConvergeParam, now time.Time) (is []card.Handler, list *ai.ConvergeInfoV2, respCode int, err error) {
	var (
		rs []*ai.Item
	)
	if list, respCode, err = s.rcmd.ConvergeList(c, plat, buvid, mid, param.ID, param.Build, param.DisplayID, param.ConvergeParam, param.ConvergeType, now); err != nil {
		log.Error("%+v", err)
		is = []card.Handler{}
		err = nil
		return
	}
	if list.Desc != "" && len(list.Items) > 0 {
		rs = append(rs, &ai.Item{CardType: string(cdm.Introduction), Goto: model.GotoIntroduction, ConvergeInfo: &ai.ConvergeInfo{Title: list.Desc}})
	}
	for _, item := range list.Items {
		tmp := &ai.Item{}
		*tmp = *item
		tmp.CardType = string(cdm.SmallCoverV8)
		rs = append(rs, tmp)
	}
	indexParam := &feed.IndexParam{
		MobiApp:  param.MobiApp,
		Device:   param.Device,
		Build:    param.Build,
		Platform: param.Platform,
		Column:   cdm.ColumnSvrSingle,
	}
	is, _ = s.dealItem2(c, mid, "", plat, rs, indexParam, false, true, false, now, nil, nil)
	for _, item := range is {
		item.Get().ThreePointV2 = nil
		item.Get().ThreePoint = nil
		item.Get().ThreePointWatchLater()
	}
	return
}

//func (s *Service) cardGifCount(param *feed.IndexParam, r *ai.Item, oldCount int) (count int) {
//	switch r.Goto {
//	case model.GotoSpecialB:
//		switch cdm.Columnm[param.Column] {
//		case cdm.ColumnSvrDouble:
//			count = oldCount + 1
//		default:
//		}
//	default:
//		count = oldCount + 1
//	}
//	return
//}

// func (s *Service) hideAdGif(ad *cm.AdInfo, abtest *feed.Abtest) bool {
// 	if ad.CreativeStyle == 2 && abtest.GifSwitch != _showAdGif { // 如果是gif广告且不能展示广告gif 则跳过本次循环
// 		return true
// 	}
// 	return false
// }

func (s *Service) adCreativeStyle(ads []*cm.AdInfo) (adInfo *cm.AdInfo, gif bool) {
	for _, ad := range ads {
		adInfo = ad
		if ad.CreativeStyle == 2 || ad.CreativeStyle == 4 {
			gif = true
			return
		}
	}
	return
}

func (s *Service) allGifState(key string, abtest *feed.Abtest) {
	if abtest == nil {
		return
	}
	if abtest.AllGifState == nil {
		abtest.AllGifState = map[string]struct{}{
			key: {},
		}
	}
	abtest.AllGifState[key] = struct{}{}
}

func (s *Service) showCardGif(key string, abtest *feed.Abtest) bool {
	if abtest == nil {
		return false
	}
	switch abtest.IsNewAd {
	case _newAd:
		switch abtest.GifType { // 0 运营gif优先、1 广告gif优先
		case _showAdGif:
			switch key {
			case _adGif:
				return true
			case _rcmdGif:
				if _, ok := abtest.AllGifState[_adGif]; ok {
					return false
				}
				return true
			case _aiGif:
				for abkey := range abtest.AllGifState {
					if abkey != key {
						return false
					}
				}
				return true
			default:
			}
		default:
			switch key {
			case _rcmdGif:
				return true
			case _aiGif:
				if _, ok := abtest.AllGifState[_rcmdGif]; ok {
					return false
				}
				return true
			case _adGif:
				for abkey := range abtest.AllGifState {
					if abkey != key {
						return false
					}
				}
				return true
			default:
			}
		}
	default:
		switch key {
		case _rcmdGif:
			return true
		case _aiGif:
			if _, ok := abtest.AllGifState[_rcmdGif]; ok {
				return false
			}
			return true
		default:
			return true
		}
	}
	return false
}

// AdCardDataBus is
func (s *Service) AdCardDataBus(c context.Context, buvid string, mid int64, resistReason int, missAds []*cm.AdInfo, hitAd *cm.AdInfo, isMelloi string) {
	if isMelloi != "" {
		return
	}

	if len(missAds) == 0 {
		return
	}

	var data = map[string]interface{}{
		"request_id":  missAds[0].RequestID,
		"mid":         mid,
		"buvid":       buvid,
		"source_id":   missAds[0].Source,
		"resource_id": missAds[0].Resource,
		"card_index":  missAds[0].CardIndex,
		"index":       missAds[0].Index,
	}
	var missData []interface{}
	for _, miss := range missAds {
		if resistReason == 0 {
			switch miss.DiscardReason {
			case "gif":
				resistReason = _adCardResistReasonGif
			case "banner":
				resistReason = _adCardResistReasonBanner
			}
		}
		missData = append(missData, map[string]interface{}{
			"resist_reason":  resistReason,
			"ad_cb":          miss.AdCb,
			"card_type":      miss.CardType,
			"creative_style": miss.CreativeStyle,
		})
	}
	data["ad_resisted"] = missData
	if hitAd != nil {
		data["ad_selected"] = []interface{}{
			map[string]interface{}{
				"ad_cb":          hitAd.AdCb,
				"card_type":      hitAd.CardType,
				"creative_style": hitAd.CreativeStyle,
			},
		}
	}
	switch resistReason {
	case _adCardResistReasonGif:
		s.infoProm.Incr("ad_card_gif")
	case _adCardResistReasonBanner:
		s.infoProm.Incr("ad_card_banner")
	default:
	}
	//nolint: bilirailguncheck
	_ = s.fanout.Do(c, func(ctx context.Context) {
		for i := 0; i < 3; i++ {
			err := s.adFeedPub.Send(ctx, missAds[0].RequestID, data)
			if err == nil {
				return
			}
			log.Error("s.adFeedPub.Send error(%v)", err)
			time.Sleep(100 * time.Millisecond)
		}
	})
}

func (s *Service) cmLog(buvid string, mid int64, plat int8, param *feed.IndexParam, now time.Time, is []card.Handler, show int, advert *cm.NewAd) {
	// 有banner的情况
	// 最前的广告卡是大卡 _delAdCard
	// 日志打印
	var hasBanner, hasSmall bool
	for _, i := range is {
		if i.Get().CardGoto == cdm.CardGotoBanner || i.Get().Goto == model.GotoBanner {
			hasBanner = true
		}
		adInfo := i.Get().AdInfo
		if adInfo == nil {
			continue
		}
		if _, ok := _delAdCard[adInfo.CardType]; ok {
			if hasBanner && !hasSmall {
				var rs []*feed.Item
				// 自己的返回 goto + id
				for _, i := range is {
					rs = append(rs, &feed.Item{
						Goto:  string(i.Get().CardGoto),
						Param: i.Get().Param,
					})
				}
				// 广告
				a, _ := json.Marshal(&rs)
				b, _ := json.Marshal(&advert)
				log.Warn("cmlog %v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,item-(%v),cm-(%v)", param.MobiApp, param.Device, plat, param.Build, buvid, mid, param.LoginEvent, param.BannerHash, now.Unix(), param.OpenEvent, show, string(a), string(b))
				break
			}
			continue
		}
		hasSmall = true
	}
}

func (s *Service) interestsList(interestList []*ai.Interest) (res *feed.Interest) {
	if len(interestList) > 0 {
		interests := &feed.Interest{
			TitleHide: s.c.Feed.Index.NewInterest.TitleHide,
			DescHide:  s.c.Feed.Index.NewInterest.DescHide,
			TitleShow: s.c.Feed.Index.NewInterest.TitleShow,
			DescShow:  s.c.Feed.Index.NewInterest.DescShow,
			Message:   s.c.Feed.Index.NewInterest.Message,
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
			res = interests
		}
	}
	return
}

func (s *Service) aiBanner(plat int8, _ int64, _ string, param *feed.IndexParam, abtest *feed.Abtest) (resourceID, bannerExp int) {
	ctx := context.Background()
	if param.LessonsMode == 1 {
		resourceID = int(common.BannerLessonResource(ctx, plat))
		return
	}
	if abtest != nil && abtest.Banner == _newBannerResource {
		if s.canEnable169BannerResourceID(param.MobiApp, param.Build) {
			resourceID = int(common.InlineBannerResource(ctx, plat))
		}
	} else {
		resourceID = int(common.OldBannerResource(ctx, plat))
	}
	if param.TeenagersMode == 1 {
		resourceID = int(common.BannerTeenagerResource(ctx, plat))
	}
	bannerExp = _aiBannerExp
	return
}

func (s *Service) canEnable169BannerResourceID(mobiApp string, build int) bool {
	return common.IsInlineBanner(mobiApp, int64(build))
}

func (s *Service) configGuidence(config *feed.Config, abtest *feed.Abtest, param *feed.IndexParam) {
	if config == nil || abtest == nil || param == nil {
		return
	}
	defer func() {
		// 当前一刷存在inline或者gif卡不能展示新用户应到或者当前不是新设备
		if param.Guidance == 0 {
			config.InterGuidance = 0
			return
		}
		if abtest.HideGuidance >= _hideGuidance {
			config.InterGuidance = 0
			stat.MetricFeedGuidanceTotal.Inc(matchStatMobi(param.MobiApp, param.Device), matchHideGuidanceReason(abtest.HideGuidance))
			return
		}
		if param.MobiApp == "android" && param.Build == 6010600 {
			stat.MetricFeedGuidanceTotal.Inc(matchStatMobi(param.MobiApp, param.Device), matchHideGuidanceReason(_hideGuidanceByBugBuild))
			config.InterGuidance = 0
			return
		}
	}()
	// 展示新用户引导
	config.InterGuidance = 1
}

func matchHideGuidanceReason(guidance int8) string {
	switch guidance {
	case _hideGuidanceByAdGif:
		return "广告gif"
	case _hideGuidanceByOperateGif:
		return "运营gif"
	case _hideGuidanceByAIGif:
		return "AIgif"
	case _hideGuidanceByInline:
		return "Inline"
	case _hideGuidanceByBugBuild:
		return "BugBuild"
	}
	return "_blank"
}

func matchStatMobi(mobiApp, device string) string {
	switch mobiApp {
	case "android":
		return "android"
	case "iphone":
		if device == "phone" {
			return "iphone"
		}
	default:
	}
	return "_blank"
}

//nolint: gocognit
func (s *Service) indexAIAd(c context.Context, rs *feed.AIResponse, abtest *feed.Abtest, infoc *feed.Infoc, adm map[int32][]*cm.AdInfo, adAidm, adRoomidm, adEpidm map[int64]struct{}, buvid string, mid int64, isMelloi string) (resAd map[int32][]*cm.AdInfo, resAdAidm, resAdRoomidm, resAdEpidm map[int64]struct{}, adInfoms map[int32][]*cm.AdInfo) {
	if adm == nil {
		adm = map[int32][]*cm.AdInfo{}
	}
	if adAidm == nil {
		adAidm = map[int64]struct{}{}
	}
	if adRoomidm == nil {
		adRoomidm = map[int64]struct{}{}
	}
	if adEpidm == nil {
		adEpidm = map[int64]struct{}{}
	}
	adInfoms = map[int32][]*cm.AdInfo{}
	resAd = adm
	resAdAidm = adAidm
	resAdRoomidm = adRoomidm
	resAdEpidm = adEpidm
	if rs == nil || rs.BizData == nil || abtest == nil {
		return
	}
	// 记录广告的card_index
	if rs.BizData.BizResult != "" {
		infoc.AdPos = strings.Split(rs.BizData.BizResult, ",")
	}
	var (
		hitAds = map[int32]*cm.AdInfo{}
	)
	// AI接口挂了
	if rs.RespCode != 0 {
		for _, v := range rs.BizData.AdSelected {
			if v == nil {
				continue
			}
			aiAd, aiAdAid, aiAdRoomid, aiAdEpid := cm.AdChangeV2(v, _cardAdAv)
			resAd[v.CardIndex-1] = aiAd
			resAdAidm[aiAdAid] = struct{}{}
			resAdRoomidm[aiAdRoomid] = struct{}{}
			resAdEpidm[aiAdEpid] = struct{}{}
			abtest.AdExp = 0
			if len(aiAd) > 0 {
				hitAds[aiAd[0].Source] = aiAd[0]
			}
		}
	} else {
		// 广告
		for _, v := range rs.Items {
			switch v.Goto {
			case model.GotoAdAv, model.GotoAdWeb, model.GotoAdWebS, model.GotoAdPlayer, model.GotoAdInlineGesture,
				model.GotoAdInline360, model.GotoAdInlineLive, model.GotoAdWebGif, model.GotoAdInlineChoose, model.GotoAdLive,
				model.GotoAdInlineChooseTeam, model.GotoAdDynamic, model.GotoAdInlineAv, model.GotoAdPlayerReservation,
				model.GotoAdWebGifReservation, model.GotoAdInline3D, model.GotoAdInlinePgc, model.GotoAdPgc,
				model.GotoAdInlineEggs, model.GotoAdInline3DV2:
				if int(v.BizIdx) >= len(rs.BizData.AdSelected) {
					FillDiscard(v.ID, v.Goto, feed.DiscardReasonAd, "ai的biz_idx大于等于广告的ad_selected", infoc)
					continue
				}
				aiAd, aiAdAid, aiAdRoomid, aiAdEpid := cm.AdChangeV2(rs.BizData.AdSelected[v.BizIdx], _cardAdAv)
				resAd[v.BizIdx] = aiAd
				resAdAidm[aiAdAid] = struct{}{}
				resAdRoomidm[aiAdRoomid] = struct{}{}
				resAdEpidm[aiAdEpid] = struct{}{}
				if len(aiAd) > 0 {
					hitAds[aiAd[0].Source] = aiAd[0]
				}
			}
		}
	}
	// 广告库存
	for _, v := range rs.BizData.Stocks {
		if v == nil {
			continue
		}
		aiAd, _, _, _ := cm.AdChangeV2(v, _cardAdAv)
		if len(aiAd) == 0 {
			continue
		}
		adInfoms[v.CardIndex-1] = aiAd
	}
	// 被抛弃的广告上报databus
	for _, v := range rs.BizData.AdDiscarded {
		if v == nil {
			continue
		}
		missAiAd, _, _, _ := cm.AdChangeV2(v, _cardAdAv)
		for _, miss := range missAiAd {
			switch miss.DiscardReason {
			case "gif":
				if infoc != nil {
					infoc.AdPkCode = append(infoc.AdPkCode, _adPkGifCard)
				}
				s.infoProm.Incr("miss_ai_ad_gif")
				if _, ok := hitAds[v.Source]; ok {
					s.infoProm.Incr("hit_ai_ad_gif")
				}
			case "banner":
				if infoc != nil {
					infoc.AdPkCode = append(infoc.AdPkCode, _adPkBigCard)
				}
				s.infoProm.Incr("miss_ai_ad_banner")
				if _, ok := hitAds[v.Source]; ok {
					s.infoProm.Incr("hit_ai_ad_banner")
				}
			}
		}
		s.AdCardDataBus(c, buvid, mid, 0, missAiAd, hitAds[v.Source], isMelloi)
	}
	return
}

func (s *Service) aiAd(group int, mid int64, param *feed.IndexParam, abtest *feed.Abtest) {
	if abtest == nil || param == nil {
		return
	}
	if param.RecsysMode == 0 && param.TeenagersMode == 0 && param.LessonsMode == 0 {
		if env.DeployEnv == "pre" {
			if mid > 0 {
				if _, ok := s.c.Custom.AIAdMid[strconv.FormatInt(mid, 10)]; ok {
					abtest.AdExp = _aiAdExp
					return
				}
				if _, ok := s.c.Custom.AIAdGroupMid[strconv.Itoa(group)]; ok { // 实验组 mid
					abtest.AdExp = _aiAdExp
					return
				}
				return
			}
			if _, ok := s.c.Custom.AIAdGroupBuvid[strconv.Itoa(group)]; ok { // 实验组 buvid
				abtest.AdExp = _aiAdExp
				return
			}
			return
		}
		abtest.AdExp = _aiAdExp
	}
}

func getCardType(h card.Handler) string {
	base := h.Get()
	if base == nil {
		return ""
	}
	return string(base.CardType)
}

func statDiscardByCardLen(rowType string, infoc *feed.Infoc, discard ...card.Handler) {
	for _, c := range discard {
		base := c.Get()
		if base == nil {
			continue
		}
		rcmd := base.Rcmd
		if rcmd == nil {
			continue
		}
		if infoc == nil {
			continue
		}
		FillDiscard(rcmd.ID, rcmd.Goto, feed.DiscardReasonEmptyWindow, "", infoc)
		stat.MetricDiscardCardTotal.Inc(rowType, rcmd.Goto, rcmd.JumpGoto, string(base.CardType), "discard_by_card_len")
	}
}

func setSessionRecordAIResponse(ctx context.Context, ai *feed.AIResponse) {
	si, ok := session.FromContext(ctx)
	if !ok {
		return
	}
	raw, err := json.Marshal(ai)
	if err != nil {
		return
	}
	si.AIRecommendResponse = string(raw)
}

func convertHotAid(in map[int64]struct{}) sets.Int64 {
	out := sets.Int64{}
	for aid := range in {
		out.Insert(aid)
	}
	return out
}

const (
	_inlineTypeAv   = "av"
	_inlineTypePGC  = "pgc"
	_inlineTypeLive = "live"
)

func bannerAddTo(r *ai.Item, aids []int64, epPlayerIDs []int32, roomIDs []int64, abtest *feed.Abtest) ([]int64, []int32, []int64) {
	for _, i := range r.BannerInfo.Items {
		if i.InlineID == "" {
			continue
		}
		id, err := strconv.ParseInt(i.InlineID, 10, 64)
		if err != nil {
			log.Error("Failed to parse inline id: %s, %+v", i.InlineID, errors.WithStack(err))
			continue
		}
		switch i.InlineType {
		case _inlineTypeAv:
			aids = append(aids, id)
			abtestHideGuidance(abtest, _hideGuidanceByInline)
		case _inlineTypePGC:
			epPlayerIDs = append(epPlayerIDs, int32(id))
			abtestHideGuidance(abtest, _hideGuidanceByInline)
		case _inlineTypeLive:
			roomIDs = append(roomIDs, id)
			abtestHideGuidance(abtest, _hideGuidanceByInline)
		default:
			log.Error("Unrecognized inline type: %+v", i.InlineType)
		}
	}
	return aids, epPlayerIDs, roomIDs
}

func abtestHideGuidance(abtest *feed.Abtest, target int) {
	if abtest != nil {
		abtest.HideGuidance = int8(target)
	}
}

func (s *Service) matchFeatureControl(mid int64, buvid, feature string) bool {
	if s.c.FeatureControl.DisableAll {
		return false
	}
	if feature == "" {
		return false
	}
	policy, ok := s.c.FeatureControl.Feature[feature]
	if !ok {
		return false
	}
	if len(policy) == 0 {
		return true
	}
	for _, v := range policy {
		fn, err := parsePolicy(v)
		if err != nil {
			log.Error("Failed to parse policy: %+v", err)
			continue
		}
		if fn(mid, buvid) {
			return true
		}
	}
	return false
}

func setOperateFromInlineConf(op *operate.Card, inline *conf.Inline) {
	op.LikeButtonShowCount = inline.LikeButtonShowCount
	op.LikeResource = &operate.LikeButtonResource{
		URL:  inline.LikeResource,
		Hash: inline.LikeResourceHash,
	}
	op.LikeNightResource = &operate.LikeButtonResource{
		URL:  inline.LikeNightResource,
		Hash: inline.LikeNightResourceHash,
	}
	op.DisLikeResource = &operate.LikeButtonResource{
		URL:  inline.DisLikeResource,
		Hash: inline.DisLikeResourceHash,
	}
	op.DisLikeNightResource = &operate.LikeButtonResource{
		URL:  inline.DisLikeNightResource,
		Hash: inline.DisLikeNightResourceHash,
	}
	op.InlinePlayIcon = operate.InlinePlayIcon{
		IconDrag:     inline.IconDrag,
		IconDragHash: inline.IconDragHash,
		IconStop:     inline.IconStop,
		IconStopHash: inline.IconStopHash,
	}
	op.InlineThreePoint = operate.InlineThreePoint{
		PanelType: inline.ThreePointPanelType,
	}
}

func isSingleV1Inline(r *ai.Item, param *feed.IndexParam) bool {
	return r.SingleInline == cdm.SingleInlineV1 && (param.Column == cdm.ColumnSvrSingle || param.Column == cdm.ColumnUserSingle)
}

func isOgvSmallCover(r *ai.Item, param *feed.IndexParam) bool {
	return r.Goto == model.GotoBangumi && cdm.Columnm[param.Column] == cdm.ColumnSvrDouble &&
		((param.MobiApp == "iphone" && param.Build > 62500000) || (param.MobiApp == "android" && param.Build > 6250000))
}

func (s *Service) canEnableDoubleUGCClickLike(mid int64, buvid string) bool {
	if mid == 0 {
		return false
	}
	if s.matchFeatureControl(mid, buvid, "double_like") {
		return true
	}
	return crc32.ChecksumIEEE([]byte(strconv.FormatInt(mid, 10)+"_double_inline_double_click"))%10 < uint32(s.c.Custom.DoubleInlineLike)
}

func (s *Service) canEnableSingleUGCClickLike(mid int64, buvid string) bool {
	if mid == 0 {
		return false
	}
	if s.matchFeatureControl(mid, buvid, "single_like") {
		return true
	}
	return crc32.ChecksumIEEE([]byte(strconv.FormatInt(mid, 10)+"_single_inline_double_click"))%10 < uint32(s.c.Custom.SingleInlineLike)
}

func fixIOSInlineBannerBug(ctx context.Context, r *ai.Item) {
	if feature.GetBuildLimit(ctx, "service.iosBugInlineBannerBuild", nil) {
		for _, i := range r.BannerInfo.Items {
			i.InlineID = ""
			i.InlineType = ""
		}
	}
}

func isSingleSpecialS(r *ai.Item, param *feed.IndexParam) bool {
	return isSingleInline(r) && cdm.Columnm[param.Column] == cdm.ColumnSvrSingle && r.SingleSpecialInfo != nil
}

func FillDiscard(id int64, goto_ string, reasonID int8, err string, infoc *feed.Infoc) {
	if infoc == nil {
		return
	}
	if _, ok := infoc.DiscardReason[id]; ok {
		return
	}
	infoc.DiscardReason[id] = &feed.Discard{
		ID:            id,
		Goto:          goto_,
		DiscardReason: reasonID,
		Error:         err,
	}
}

func (s *Service) teenagerSpecialCondition(param *feed.IndexParam, r *ai.Item) bool {
	return s.c.Feed.Index.TeenagersSpecialCard != nil && r.ID == s.c.Feed.Index.TeenagersSpecialCard.ID && (param.TeenagersMode == 1 || param.LessonsMode == 1)
}

func (s *Service) specialCardV2(ctx context.Context, ids []int64) (map[int64]*resourceV2grpc.AppSpecialCard, error) {
	out, err := s.rsc.SpecialV2(ctx, ids)
	if err != nil {
		return nil, err
	}
	out[s.c.Feed.Index.TeenagersSpecialCard.ID] = &resourceV2grpc.AppSpecialCard{
		Id:      s.c.Feed.Index.TeenagersSpecialCard.ID,
		Title:   s.c.Feed.Index.TeenagersSpecialCard.Title,
		Cover:   s.c.Feed.Index.TeenagersSpecialCard.Cover,
		ReType:  0,
		ReValue: s.c.Feed.Index.TeenagersSpecialCard.URL,
	}
	return out, nil
}

//nolint:bilirailguncheck
func (s *Service) FeedAppListProduce(c *bm.Context, param *feed.IndexParam, mid int64, buvid, applist string) {
	if param.LoginEvent == 0 || applist == "" {
		return
	}
	applistPubParam := &feed.FeedAppListParam{
		Mid:      mid,
		Buvid:    buvid,
		MobiApp:  param.MobiApp,
		Device:   param.Device,
		Platform: param.Platform,
		Build:    param.Build,
		IP:       metadata.String(c, metadata.RemoteIP),
		Ua:       c.Request.UserAgent(),
		Referer:  c.Request.Referer(),
		Origin:   c.Request.Header.Get("Origin"),
		CdnIp:    c.Request.Header.Get("X-Cache-Server-Addr"),
		Channel:  c.Request.URL.Query().Get("channel"),
		Brand:    c.Request.URL.Query().Get("brand"),
		Model:    c.Request.URL.Query().Get("model"),
		Osver:    c.Request.URL.Query().Get("osver"),
		Applist:  applist,
	}
	bt, err := applistPubParam.MarshalJSON()
	if err != nil {
		return
	}
	if err = FeedAppListPub.Send(c, buvid, bt); err != nil {
		log.Error("Failed to pub applist: %s, %+v", bt, err)
	}
}

var (
	interestFlag = ab.Int("newuser_interest_select", "new user interest select", 0)
)

func (s *Service) interestAbTest(ctx context.Context, buvid string, mid int64) int64 {
	if mid > 0 {
		return 0
	}
	t, ok := ab.FromContext(ctx)
	if !ok {
		return 0
	}
	t.Add(ab.KVString("buvid", buvid))
	exp := interestFlag.Value(t)
	return exp
}

var _interestChooseMap = map[int64]*feed.InterestChoose{
	8: {
		Style: 12,
		Items: []*feed.InterestChooseItem{
			{
				Name: "动漫游戏",
				Id:   1,
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220613/0977767b2e79d8ad0a36a731068a83d7/ZNo8NC5a66.png",
				Desc: "海量番剧",
			},
			{
				Name: "知识学习",
				Id:   3,
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220613/0977767b2e79d8ad0a36a731068a83d7/KeZho1w4ND.png",
				Desc: "丰富专业",
			},
			{
				Name: "影视剧集",
				Id:   7,
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220613/0977767b2e79d8ad0a36a731068a83d7/upl6y9j7A0.png",
				Desc: "独家剧集",
			},
			{
				Name: "明星娱乐",
				Id:   8,
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220613/0977767b2e79d8ad0a36a731068a83d7/CfQGbDbpgo.png",
				Desc: "精彩剪辑",
			},
			{
				Name: "搞笑萌宠",
				Id:   5,
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220613/0977767b2e79d8ad0a36a731068a83d7/W2CBkLlAli.png",
				Desc: "沙雕视频",
			},
			{
				Name: "科技数码",
				Id:   9,
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220613/0977767b2e79d8ad0a36a731068a83d7/T6l9Cvc3tl.png",
				Desc: "专业评测",
			},
			{
				Name: "寻味美食",
				Id:   10,
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220613/0977767b2e79d8ad0a36a731068a83d7/Gb29Y0NiB4.png",
				Desc: "探店吃播",
			},
			{
				Name: "运动健身",
				Id:   11,
				Icon: "https://i0.hdslb.com/bfs/activity-plat/static/20220613/0977767b2e79d8ad0a36a731068a83d7/3ieobodqyj.png",
				Desc: "瘦身塑形",
			},
		},
		Title:          "选择你感兴趣的方向",
		SubTitle:       "为你推荐丰富多样的内容",
		ConfirmText:    "一键开启推荐",
		ConfirmOutText: "选好了，开启内容推荐",
		CancelText:     "跳过",
		UniqueId:       12,
	},
}

func (s *Service) IndexInterest(ctx *bm.Context, mid int64, buvid string) (*feed.InterestChoose, error) {
	isNewBuvid := s.acc.CheckRegTime(ctx, &accountgrpc.CheckRegTimeReq{Buvid: buvid, Periods: _newUserInterestPeriod})
	if !isNewBuvid {
		return nil, ecode.Errorf(ecode.RequestErr, "buvid非新：%s", buvid)
	}
	result := s.interestAbTest(ctx, buvid, mid)
	if result == 0 {
		return nil, nil
	}
	return &feed.InterestChoose{
		Style:          _interestChooseMap[result].Style,
		Items:          _interestChooseMap[result].Items,
		Title:          _interestChooseMap[result].Title,
		SubTitle:       _interestChooseMap[result].SubTitle,
		ConfirmText:    _interestChooseMap[result].ConfirmText,
		ConfirmOutText: _interestChooseMap[result].ConfirmOutText,
		CancelText:     _interestChooseMap[result].CancelText,
		UniqueId:       _interestChooseMap[result].UniqueId,
	}, nil
}

func (s *Service) canEnableClassBadge(mid int64, buvid string) bool {
	if s.matchFeatureControl(mid, buvid, "class_badge") {
		return true
	}
	return int(crc32.ChecksumIEEE([]byte(buvid+"_open_class_badge"))%10) < s.c.Custom.ClassBadgeGroup
}
