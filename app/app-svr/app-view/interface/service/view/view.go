package view

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	pagy "git.bilibili.co/bapis/bapis-go/bilibili/pagination"
	"github.com/thoas/go-funk"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	fksmeta "go-common/component/metadata/fawkes"
	"go-common/component/metadata/network"
	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/exp/ab"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/stat/prom"
	egV2 "go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	appecode "go-gateway/app/app-svr/app-card/ecode"
	xecode "go-gateway/app/app-svr/app-card/ecode"
	cdm "go-gateway/app/app-svr/app-card/interface/model"
	"go-gateway/app/app-svr/app-card/interface/model/card"
	"go-gateway/app/app-svr/app-card/interface/model/card/operate"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	appresapi "go-gateway/app/app-svr/app-resource/interface/api/v1"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/bangumi"
	"go-gateway/app/app-svr/app-view/interface/model/creative"
	musicmdl "go-gateway/app/app-svr/app-view/interface/model/music"
	"go-gateway/app/app-svr/app-view/interface/model/tag"
	"go-gateway/app/app-svr/app-view/interface/model/view"
	"go-gateway/app/app-svr/app-view/interface/tools"
	avecode "go-gateway/app/app-svr/archive/ecode"
	"go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	resource "go-gateway/app/app-svr/resource/service/model"
	steinApi "go-gateway/app/app-svr/steins-gate/service/api"
	seasonApi "go-gateway/app/app-svr/ugc-season/service/api"
	mainEcode "go-gateway/ecode"
	"go-gateway/pkg/idsafe/bvid"
	"go-gateway/pkg/riskcontrol"

	listenerDao "go-gateway/app/app-svr/app-view/interface/dao/listener"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	upApi "git.bilibili.co/bapis/bapis-go/archive/service/up"
	checkin "git.bilibili.co/bapis/bapis-go/community/interface/checkin"
	dmgrpc "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	buzzword "git.bilibili.co/bapis/bapis-go/community/interface/dm-buzzword"
	sharerpc "git.bilibili.co/bapis/bapis-go/community/interface/share"
	appConf "git.bilibili.co/bapis/bapis-go/community/service/appconfig"
	favecode "git.bilibili.co/bapis/bapis-go/community/service/favorite/ecode"
	location "git.bilibili.co/bapis/bapis-go/community/service/location"
	thumbup "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	tecode "git.bilibili.co/bapis/bapis-go/community/service/thumbup/ecode"
	flowcontrolapi "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	votegrpc "git.bilibili.co/bapis/bapis-go/dynamic/service/vote"
	vcloud "git.bilibili.co/bapis/bapis-go/video/vod/playurlstory"
	vuApi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"

	"github.com/pkg/errors"
)

const (
	_promptCoin               = 1
	_promptFav                = 2
	_avTypeAv                 = 1
	_businessLike             = "archive"
	_coinAv                   = 1
	_shortLinkHost            = "https://b23.tv"
	_defaultBgColor           = "#000000"
	_steinsLabelIos           = 8790
	_steinsLabelAndroid       = 5460500
	_steinsLabelAndroidBlue   = 5330000
	_steinsLabelIosBlue       = 8030
	_steinsLabelIpad          = 12200
	_steinsLabelAndroidI      = 3000000
	_steinsLabelIphoneI       = 64400200
	_musicHonorControlAndroid = 6680000
	_musicHonorControlIos     = 66800000
	_superReplyControlAndroid = 6480300
	_superReplyControlIos     = 64800000
	_typeAv                   = "av"
	_shareType                = 3
	_biJianType               = 51
	_biJianBiz                = 1
	_contractTitle            = "成为UP主的\"老粉\""
	_contractSubtitle         = "助力UP主成长，让更多人发现TA"
	_contractInlineTitle      = "投资当前UP主"
	_signPyText               = "付费"
	_likeGray                 = 10000
	_contractToast            = "三连推荐成功，感谢原始粉丝~"
	_oldFansToast             = "三连推荐成功，感谢老粉~"
	_hardCoreToast            = "三连成功，硬核推荐力Upup！"
	_harCoreCoinToast         = "已收到来自硬核指挥部的硬币！"
	_superReplyKey            = "aid_superb_reply"
	_hotMusicKey              = "aid_music_top_list"
	_arcPubTimeForm           = "2006-01-02 15:04:05"
	_seasonAbilityCheck       = "打卡"
)

var (
	pipVal                = ab.String("is_auto_pip_allusers_iOS", "view", _missABValue)
	smallWindowABtest     = ab.String("auto_pip_iPhone", "view", _missABValue)
	newSwindowABTestFlage = ab.String("new_device_new_miniplayer", "新版小窗实验", _missABValue)
	relatesBiserialABtest = ab.String("is_double_row", "相关推荐HD双列实验", _missABValue)
)

func (s *Service) HasCustomConfig(c context.Context, aid int64) bool {
	return s.rscDao.HasCustomConfig(c, resource.CustomConfigTPArchive, aid)
}

func (s *Service) NothingFoundUrl(aid int64) string {
	return fmt.Sprintf("http://www.bilibili.com/h5/special-404/%d?navhide=1", aid)
}

// View  all view data.
func (s *Service) ViewHttp(c context.Context, mid, aid int64, plat int8, build, parentMode, autoplay, teenagersMode, lessonsMode int, disableRcmdMode int,
	mobiApp, device, buvid, cdnIP, network, adExtra, from, spmid, fromSpmid, trackid, platform, filtered, withoutCharge string, isMelloi, brand, slocale, clocale string) (v *view.View, err error) {
	vp, extra, err := s.ArcView(c, aid, 0, "", "", "", plat)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	//针对android_tv屏蔽付费稿件
	if mobiApp == "android_tv" && vp.Arc.AttrValV2(api.AttrBitV2Pay) == api.AttrYes {
		return nil, ecode.NothingFound
	}
	return s.ViewInfo(c, mid, aid, plat, build, parentMode, autoplay, teenagersMode, lessonsMode,
		mobiApp, device, buvid, cdnIP, network, adExtra, from, spmid, fromSpmid, trackid, platform, filtered,
		withoutCharge, false, isMelloi, brand, slocale, clocale, "", vp, disableRcmdMode, 0, 0, "", "", 0, 0, 0, extra)
}

// nolint:gocognit
func (s *Service) ViewInfo(c context.Context, mid, aid int64, plat int8, build, parentMode, autoplay, teenagersMode,
	lessonsMode int, mobiApp, device, buvid, cdnIP, network, adExtra, from, spmid, fromSpmid, trackid, platform,
	filtered, withoutCharge string, viewGRPC bool, isMelloi, brand, slocale, clocale, pageVersion string,
	vp *api.ViewReply, disableRcmdMode int, deviceType int64, pageIndex int64, sessionId, playMode string, inFeedPlay, refreshNum, refreshType int32, extra map[string]string) (v *view.View, err error) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	if v, err = s.ViewPage(c, mid, plat, build, mobiApp, device, cdnIP, true, buvid, slocale, clocale, vp, pageVersion, spmid, platform, teenagersMode, extra); err != nil {
		log.Error("%+v", err)
		return
	}
	// config
	if v == nil {
		return nil, ecode.NothingFound
	}
	defer HideArcAttribute(v.Arc)
	if v.Config == nil {
		v.Config = &view.Config{}
	}
	v.Config.RelatesTitle = s.c.ViewConfig.RelatesTitle
	v.SubTitleChange()
	v.Config.ShareStyle = 1
	// end page abtest
	v.Config.EndPageHalf, v.Config.EndPageFull = s.endPageTest(buvid, mid)
	//相关推荐是否双列展示
	if _, ok := s.c.RelatesBiserialWhiteList[strconv.FormatInt(mid, 10)]; ok || cfg.relatesBiserialExp {
		v.Config.RelatesBiserial = true
		prom.BusinessInfoCount.Incr("相关推荐-双列展示")
	}
	var tagIDs []int64
	for _, tagData := range v.Tag {
		tagIDs = append(tagIDs, tagData.TagID)
	}
	//新版本去掉活动tag
	v.Tag = s.NewTopicDelActTag(c, v.Tag, buvid)
	//点赞场景化定制
	if s.c.Custom.LikeCustomSwitch {
		v.LikeCustom = &viewApi.LikeCustom{
			FullToHalfProgress: s.c.Custom.FullToHalfProgress,
			NonFullProgress:    s.c.Custom.NonFullProgress,
			UpdateCount:        s.c.Custom.LikeCustomUpdateCount,
		}
		if v.Stat.View >= int32(s.c.Custom.LikeCustomVideoView) && v.SeasonID == 0 {
			v.LikeCustom.LikeSwitch = true
		}
	}
	// config
	g := egV2.WithContext(c)
	g.Go(func(ctx context.Context) (err error) {
		s.initReqUser(ctx, v, mid, plat, build, buvid, platform, brand, network, mobiApp)
		return
	})
	//获取竖屏视频切全屏是否进story
	g.Go(func(ctx context.Context) error {
		v.Config.PlayStory, v.Config.StoryIcon, v.Config.LandscapeStory, v.Config.LandscapeIcon = s.playStoryABTest(ctx, mid, buvid)
		//单p &&非互动视频 && 非付费视频(不论是否支持免费试看) && autoplay等于1
		var isSingle bool
		if v.GetVideos() == 1 && !v.IsSteinsGate() && !v.IsBitV2Pay() && v.GetRights().Autoplay == 1 {
			isSingle = true
		}
		if v.Config.PlayStory && isSingle { //命中实验
			v.Config.ArcPlayStory = true
		}
		if v.Config.LandscapeStory && isSingle {
			v.Config.ArcLandscapeStory = true
		}
		return nil
	})
	//是否展示听视频按钮
	if mid > 0 && mid%100 < s.c.Custom.ListenButtonGrey {
		//一级分区
		pid := s.ArchiveTypesMap[v.TypeID]
		if funk.Contains(s.c.Custom.ListenButtonType, int(pid)) {
			g.Go(func(ctx context.Context) (err error) {
				copyrightReply, err := s.copyright.GetArcBanPlay(ctx, v.Aid)
				if err != nil {
					log.Error("s.copyright.GetArcBanPlay err(%+v)", err)
					return nil
				}
				if !copyrightReply {
					v.Config.ShowListenButton = true
				}
				return nil
			})
		}
	}
	// 获取点赞动画
	if mid > 0 {
		g.Go(func(ctx context.Context) error {
			equipRly, e := cfg.dep.Garb.ThumbupUserEquip(ctx, mid)
			if e != nil {
				log.Error("s.garbDao.ThumbupUserEquip(%d) error(%v)", mid, e)
				return nil
			}
			if equipRly != nil {
				v.UserGarb = &viewApi.UserGarb{UrlImageAniCut: equipRly.URLImageAniCut}
				if equipRly.URLImageAniCut != "" {
					s.prom.Incr("ThumbupUserEquip-HasValue")
				}
			}
			return nil
		})
	}
	//首映资源(预约+文案+首映状态等)
	if v.Premiere != nil {
		g.Go(func(ctx context.Context) error {
			s.initPremiere(c, v, mid)
			return nil
		})
	}
	if teenagersMode == 0 && lessonsMode == 0 {
		g.Go(func(ctx context.Context) (err error) {
			s.initHonor(ctx, v, plat, build, mobiApp, device)
			// 保持线上逻辑，无荣誉榜单 && 3 < 稿件排行 <= 10 返回文字排行
			if v.Honor == nil && v.Stat.HisRank > s.c.Custom.HonorRank && v.Stat.HisRank <= s.c.Custom.HonorRankMax {
				v.Rank = &viewApi.Rank{
					Icon:      model.RankIcon,
					IconNight: model.RankIconNight,
					Text:      fmt.Sprintf("全站排行榜最高第%d名", v.Stat.HisRank),
				}
			}
			return nil
		})
		if !cfg.skipRelate {
			g.Go(func(ctx context.Context) (err error) {
				if viewGRPC {
					//全量切AI推荐，旧逻辑已删除
					s.initRelateCMTagNewV2(ctx, v, plat, build, parentMode, autoplay, mid, buvid, mobiApp, device, network, adExtra, from, spmid, fromSpmid, trackid, filtered, tagIDs, slocale, clocale, pageVersion, cfg, disableRcmdMode, deviceType, pageIndex, sessionId, playMode, inFeedPlay, refreshNum, refreshType)
				} else {
					s.initRelateCMTag(ctx, v, plat, build, parentMode, autoplay, mid, buvid, mobiApp, device, network, adExtra, from, spmid, fromSpmid, trackid, platform, filtered, tagIDs, isMelloi, slocale, clocale, pageVersion)
				}
				return nil
			})
		}
	}
	asDesc := ""
	arcDescV2 := []*api.DescV2{
		{
			RawText: v.Desc,
			Type:    api.DescType_DescTypeText,
		},
	}
	accountInfos := &accApi.InfosReply{}
	if v.AttrVal(api.AttrBitIsPGC) != api.AttrYes {
		g.Go(func(ctx context.Context) (err error) {
			// 从6.10版本开始去除对dm.SubjectInfos调用
			if (plat == model.PlatIPhone && build >= s.c.BuildLimit.DmInfoIOSBuild) || (plat == model.PlatAndroid && build >= s.c.BuildLimit.DmInfoAndBuild) {
				return
			}
			s.initDM(ctx, v)
			return
		})
		if teenagersMode == 0 {
			g.Go(func(ctx context.Context) (err error) {
				s.initAudios(ctx, v)
				return
			})
		}
		if teenagersMode == 0 && lessonsMode == 0 {
			g.Go(func(ctx context.Context) (err error) {
				if model.IsIPhoneB(plat) || (model.IsIPhone(plat) && (build >= 7000 && build <= 8000)) {
					return
				}
				s.initElec(ctx, v, mobiApp, platform, device, build, mid)
				return
			})
		}
		g.Go(func(ctx context.Context) (err error) {
			desc, descV2, mids, err := cfg.dep.Archive.DescriptionV2(ctx, v.Aid)
			if err != nil {
				log.Error("s.arcDao.DescriptionV2 aid(%d),err(%+v)", v.Aid, err)
				return nil
			}
			arcDescV2 = descV2
			asDesc = desc
			//拉取用户最新数据
			if len(mids) > 0 {
				accountInfos, err = cfg.dep.Account.GetInfos(ctx, mids)
				if err != nil {
					log.Error("s.accDao.GetInfos aid(%d),err(%+v)", v.Aid, err)
				}
			}
			return nil
		})
	}
	g.Go(func(ctx context.Context) (err error) {
		if v.Bgm, v.Sticker, v.VideoSource, err = cfg.dep.VideoUP.GetMaterialList(ctx, v.Aid, v.FirstCid); err != nil {
			log.Error("s.vuDao.GetMaterialList aid(%d),err(%+v)", v.Aid, err)
		}
		return nil
	})
	g.Go(func(ctx context.Context) (err error) {
		uplikeImg, err := cfg.dep.Archive.UpLikeImgCreative(ctx, v.Author.Mid, v.Aid)
		if err != nil {
			log.Error("cfg.dep.Archive.UpLikeImgCreative is error %+v", err)
		}
		if uplikeImg != nil {
			v.UpLikeImg = uplikeImg
		}
		return nil
	})
	if v.AttrVal(api.AttrBitHasArgument) == api.AttrYes {
		g.Go(func(ctx context.Context) (err error) {
			req := &vuApi.MultiArchiveArgumentReq{
				Aids: []int64{v.Aid},
			}
			reply, err := s.vuDao.MultiArchiveArgument(ctx, req)
			if err != nil {
				log.Error("Failed to get archive argument: %+v: %+v", req, err)
				return nil
			}
			if r, ok := reply.Arguments[v.Aid]; ok {
				v.ArgueMsg = r.ArgueMsg
			}
			return nil
		})
	}
	g.Go(func(ctx context.Context) error {
		// 版本控制
		var showPlayicon bool
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.ViewPlayIcon, &feature.OriginResutl{
			MobiApp: mobiApp,
			Device:  device,
			Build:   int64(build),
			BuildLimit: (mobiApp == "iphone" && build >= conf.Conf.BuildLimit.PlayIconIOSBuildLimit) ||
				(mobiApp == "android" && build >= conf.Conf.BuildLimit.PlayIconAndroidBuildLimit) ||
				(mobiApp == "ipad" && build >= conf.Conf.BuildLimit.PlayIconIpadHDBuildLimit),
		}) {
			showPlayicon = true
		}
		playerIconRly, err := cfg.dep.Resource.PlayerIconNew(ctx, v.Aid, mid, tagIDs, v.TypeID, showPlayicon, build, mobiApp, device)
		if err != nil {
			log.Error("PlayerIconNew err(%+v) aid(%d) tagids(%+v) typeid(%d)", err, v.Aid, tagIDs, v.TypeID)
			return nil
		}
		if playerIconRly != nil {
			v.PlayerIcon = playerIconRly.Item
		}
		return nil
	})
	//评论样式
	g.Go(func(ctx context.Context) error {
		res, err := cfg.dep.Reply.GetReplyListPreface(ctx, mid, aid, buvid)
		if err != nil {
			log.Error("GetReplyListPreface fail mid:%d,aid:%d,err%+v", mid, aid, err)
			return nil
		}
		v.BadgeUrl = res.BadgeUrl
		v.ReplyStyle = &viewApi.ReplyStyle{
			BadgeUrl:  res.BadgeUrl,
			BadgeText: res.BadgeText,
			BadgeType: res.BadgeType,
		}
		return nil
	})
	g.Go(func(ctx context.Context) error {
		res, err := s.thumbupDao.GetMultiLikeAnimation(ctx, aid)
		if err != nil {
			log.Error("s.thumbupDao.GetMultiLikeAnimation aid:%d,err%+v", aid, err)
			return nil
		}
		if like, ok := res[aid]; ok {
			v.LikeAnimation = &viewApi.LikeAnimation{
				LikeIcon:      like.LikeIcon,
				LikedIcon:     like.LikedIcon,
				LikeAnimation: like.LikeAnimation,
			}
			if like.LikeCartoon != "" {
				v.IsLikeAnimation = true
				v.OperationLikeAnimation = like.LikeCartoon
			}
		}
		return nil
	})
	if s.matchNGBuilder(mid, buvid, "tf_panel") {
		g.Go(func(ctx context.Context) (err error) {
			customizedPanel, err := cfg.dep.Resource.GetPlayerCustomizedPanel(ctx, tagIDs)
			if err != nil && !ecode.EqualError(ecode.NothingFound, err) {
				log.Error("Failed to get player customized panel with tids: %+v: %+v", tagIDs, err)
				return nil
			}
			v.TfPanelCustomized = view.FromPlayerCustomizedPanel(customizedPanel)
			return nil
		})
	}
	// (在白名单 || 命中分组) && 非up主 && 非合集
	if s.liveBookingControl(buvid, mid, vp.Author.Mid, vp.SeasonID, teenagersMode, lessonsMode) {
		g.Go(func(ctx context.Context) error {
			reply, err := cfg.dep.AppResource.CheckEntranceInfoc(ctx, &appresapi.CheckEntranceInfocRequest{Mid: mid, UpMid: vp.Author.Mid, Business: "live_reserve"})
			if err != nil {
				log.Error("s.appResourceClient.CheckEntranceInfoc error(%+v)", err)
				return nil
			}
			if reply == nil || reply.GetIsExisted() { //用户主动关闭提示条
				return nil
			}
			actReply, err := cfg.dep.Activity.LiveBooking(ctx, mid, vp.Author.Mid)
			if err != nil {
				return nil
			}
			// 状态不对 || 已开播 || 已预约 || 仅up主可见状态 不展示
			if actReply.State != actgrpc.UpActReserveRelationState_UpReserveRelated || actReply.LivePlanStartTime.Time().Unix() <= time.Now().Unix() ||
				actReply.IsFollow == 1 || actReply.UpActVisible == actgrpc.UpActVisible_OnlyUpVisible {
				return nil
			}
			v.LiveOrderInfo = &viewApi.LiveOrderInfo{
				IsFollow:          false,
				LivePlanStartTime: actReply.LivePlanStartTime.Time().Unix(),
				Sid:               actReply.Sid,
				Text:              s.c.Custom.LiveOrderText,
			}
			return nil
		})
	}
	g.Go(func(ctx context.Context) error {
		//首映稿件 + 首映前 + 首映中不返回小窗
		if v.Premiere != nil &&
			(v.Premiere.State == api.PremiereState_premiere_before || v.Premiere.State == api.PremiereState_premiere_in) {
			v.Config.AutoSwindow = false
			return nil
		}
		newDevice := s.accDao.IsNewDevice(ctx, buvid, "0-24")
		func() {
			if mobiApp != "iphone" || !s.c.Custom.PipSwitchOn {
				return
			}
			if newDevice {
				v.Config.AutoSwindow = true
				return
			}
			v.Config.AutoSwindow = cfg.autoSwindowExp
		}()

		if _, ok := s.c.NewSwindowWhiteList[strconv.FormatInt(mid, 10)]; ok || (newDevice && cfg.newSwindowExp) {
			v.Config.NewSwindow = true
		}

		return nil
	})

	g.Go(func(ctx context.Context) error {
		opt := listenerDao.ListenerSwitchOpt{
			Mid:   mid,
			Aid:   aid,
			Spmid: spmid,
		}
		ret, err := s.listenerDao.ListenerConfig(ctx, opt)
		if err != nil {
			log.Warn("s.listenerDao.ListenerConfig Error (%v)", err)
		} else {
			v.Config.ListenerConfig = ret
		}

		return nil
	})

	if mobiApp == "iphone" {
		g.Go(func(ctx context.Context) error {
			v.Config.AbTestSmallWindow = _smallWindowKeep
			if cfg.smallWindowExp {
				v.Config.AbTestSmallWindow = _smallWindowOpen
			}
			return nil
		})
	}
	if mobiApp == "android" {
		g.Go(func(ctx context.Context) error {
			if s.popupConfig(ctx, mid, buvid) {
				v.Config.PopupInfo = true
			}
			return nil
		})
	}
	g.Go(func(ctx context.Context) error {
		s.initLabel(ctx, v, s.displaySteinsLabel(c, v.ViewStatic, mobiApp, device, build))
		return nil
	})
	//是否进行跳转
	g.Go(func(ctx context.Context) error {
		//获取archive_redirect数据
		redirect, err := cfg.dep.Archive.ArcRedirectUrl(ctx, aid)
		if err != nil {
			return nil
		}
		if redirect.RedirectTarget == "" || redirect.PolicyId == 0 {
			return nil
		}
		//location策略获取返回数据
		if redirect.GetPolicyType() == api.RedirectPolicyType_PolicyTypeLocation {
			locs, err := cfg.dep.Location.GetGroups(ctx, []int64{redirect.PolicyId})
			if err != nil {
				log.Error("GetGroups is err (%+v)", err)
				return nil
			}
			loc, ok := locs[redirect.PolicyId]
			if !ok {
				return nil
			}
			//是否需要跳转
			if loc.Play != int64(location.Status_Forbidden) {
				v.Season = &bangumi.Season{
					IsJump:     1,
					OGVPlayURL: redirect.RedirectTarget,
					SeasonID:   "1", //为了兼容android的逻辑，写死一个不存在的season_id+title
					Title:      "forcejump",
				}
			}
		}
		return nil
	})
	// 直接获取点赞数据
	if s.LikeGrayControl(v.Aid) {
		g.Go(func(ctx context.Context) error {
			reply, err := s.thumbupDao.GetStates(ctx, _businessLike, []int64{aid})
			if err != nil {
				log.Error("s.thumbupDao.GetStates err:%+v", err)
				return nil
			}
			if likeState, ok := reply.Stats[aid]; ok && likeState != nil {
				v.Arc.Stat.Like = int32(likeState.LikeNumber)
			}
			return nil
		})
	}
	// 判断是否为硬核会员
	if mid > 0 {
		g.Go(func(ctx context.Context) error {
			info, err := s.accDao.GetInfo(ctx, mid)
			if err != nil {
				log.Error("s.accDao.GetInfo err:%+v", err)
				return nil
			}
			if info != nil && info.IsSeniorMember == 1 {
				v.IsHardCoreFans = true
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	//返回
	if asDesc != "" {
		v.Desc = asDesc
	}
	if len(arcDescV2) > 0 {
		v.DescV2 = s.DescV2ParamsMerge(c, arcDescV2, accountInfos)
	}
	if v.AttrValV2(api.AttrBitV2OnlyFavView) == api.AttrYes {
		if ok := func() bool {
			if mid == 0 {
				return false
			}
			if mid == v.Author.Mid {
				return true
			}
			for _, sf := range v.StaffInfo {
				if sf.Mid == mid {
					return true
				}
			}
			if v.ReqUser != nil && v.ReqUser.Favorite == 1 {
				return true
			}
			return false
		}(); !ok {
			return nil, ecode.NothingFound
		}
	}
	//获取short_link
	s.setShortLink(v)
	if v.AttrValV2(model.AttrBitV2CleanMode) == api.AttrYes {
		v.Tag = nil
	}
	//过滤up mid64位的相关推荐
	if tools.CheckNeedFilterMid64(c) {
		relates := make([]*view.Relate, 0, len(v.Relates))
		for _, re := range v.Relates {
			if re.Author == nil || tools.IsInt32Mid(re.Author.Mid) {
				relates = append(relates, re)
				continue
			}
			prom.BusinessInfoCount.Incr("相关推荐mid64过滤卡片")
			jsonValue, _ := json.Marshal(re)
			log.Warn("相关推荐mid64过滤: (%+v)", string(jsonValue))
		}
		v.Relates = relates
	}
	//首映被风控，则首映标签资源都为空
	if v.PremiereRiskStatus && v.Label != nil && v.Label.Type == 4 {
		v.Label = nil
	}
	// 点赞三连动画设置
	s.LikeAnimate(v)
	s.TripleAnimate(v)
	// 投币toast设置
	s.CoinConfig(v)
	//新充电按钮
	s.setElecCharging(mobiApp, device, v)
	return
}

func (s *Service) CoinConfig(v *view.View) {
	// 硬核会员定制投币标语
	if v.IsHardCoreFans {
		v.CoinCustom = &viewApi.CoinCustom{
			Toast: _harCoreCoinToast,
		}
	}
}

// LikeAnimate avid指定>装扮皮肤＞契约者/老粉>硬核会员＞普通
func (s *Service) LikeAnimate(v *view.View) {
	if v.IsLikeAnimation {
		v.UserGarb = &viewApi.UserGarb{
			UrlImageAniCut: v.OperationLikeAnimation,
		}
		return
	}
	// 如果有装扮皮肤
	if v.UserGarb != nil && v.UserGarb.UrlImageAniCut != "" {
		return
	}
	var vType int64
	var toast string
	// 如果是契约者
	if v.IsContractor {
		vType = 3
		// 如果是老粉
		if v.IsOldFans {
			vType = 4
		}
		toast = s.likeToast(vType, int64(v.Arc.Stat.Like))
		v.UserGarb = &viewApi.UserGarb{
			UrlImageAniCut: "https://i0.hdslb.com/bfs/app/b2e96bda9f13d17dd75cf26dad29b7f1500f5991.bin",
			LikeToast:      toast,
		}
		return
	}
	// 如果是硬核会员
	if v.IsHardCoreFans {
		vType = 5
		toast = s.likeToast(vType, int64(v.Arc.Stat.Like))
		// 硬核会员点赞动画
		v.UserGarb = &viewApi.UserGarb{
			UrlImageAniCut: "https://i0.hdslb.com/bfs/app/c43ec1b96be0c75f5bb0f0f875824b0cac54d3a8.bin",
			LikeToast:      toast,
		}
		return
	}
}

// TripleAnimate avid维度下发＞UP主维度下发＞契约者/老粉>硬核会员＞普通
func (s *Service) TripleAnimate(v *view.View) {
	// 如果有avid或up主维度下发
	if v.UpLikeImg != nil && v.UpLikeImg.SucImg != "" {
		return
	}
	// 如果是契约者
	if v.IsContractor {
		v.UpLikeImg = &viewApi.UpLikeImg{
			PreImg:  "https://i0.hdslb.com/bfs/dm/b043246c382f2ebb8614a648cfe3ecf83433c252.gif",
			SucImg:  "https://i0.hdslb.com/bfs/app/73d0966623c2bf6141f04e2bd4ef229aedb27b70.gif",
			Content: _contractToast,
			Type:    1,
		}
		// 如果是老粉
		if v.IsOldFans {
			v.UpLikeImg.Content = _oldFansToast
		}
		return
	}
	// 如果是硬核会员
	if v.IsHardCoreFans {
		// 硬核会员三连动画
		v.UpLikeImg = &viewApi.UpLikeImg{
			PreImg:  "https://i0.hdslb.com/bfs/app/85b1b7e5fcaecca79eb414a5a9ad7266c1901eca.gif",
			SucImg:  "https://i0.hdslb.com/bfs/app/a79dfb1f8e8d19e33d0930286797b4178d160271.gif",
			Content: _hardCoreToast,
			Type:    2,
		}
		return
	}
}

func (s *Service) NewTopicDelActTag(c context.Context, channelTag []*tag.Tag, buvid string) []*tag.Tag {
	res := []*tag.Tag{}
	//版本判断
	buildBool := pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", int64(s.c.BuildLimit.NewTopicAndroidBuild))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build(">=", int64(s.c.BuildLimit.NewTopicIOSBuild))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPadHD().And().Build(">=", int64(s.c.BuildLimit.NewTopicIPadHDBuild))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatIPad().And().Build(">=", int64(s.c.BuildLimit.NewTopicIPadBuild))
	}).MustFinish()
	//满足版本号 + 灰度逻辑 去掉活动tag
	if buildBool && !s.NewTopicActTagGrey(buvid) {
		for _, v := range channelTag {
			if v.TagType == "act" {
				continue
			}
			res = append(res, v)
		}
		return res
	}
	return channelTag
}

func (s *Service) DescV2ParamsMerge(c context.Context, arcDescV2 []*api.DescV2, accountInfos *accApi.InfosReply) []*viewApi.DescV2 {
	viewDescV2 := []*viewApi.DescV2{}
	if len(arcDescV2) == 0 {
		return nil
	}
	for _, val := range arcDescV2 {
		if val == nil {
			continue
		}
		uri := ""
		if val.BizId > 0 && viewApi.DescType(val.Type) == viewApi.DescType_DescTypeAt {
			midStr := strconv.FormatInt(val.BizId, 10)
			uri = model.FillURI(model.GotoSpace, midStr, nil)
		}
		if accountInfos != nil {
			if rawText, ok := accountInfos.Infos[val.BizId]; ok {
				val.RawText = rawText.Name
			}
		}
		viewDescV2 = append(viewDescV2, &viewApi.DescV2{
			Text: val.RawText,
			Type: viewApi.DescType(val.Type),
			Rid:  val.BizId,
			Uri:  uri,
		})
	}
	return viewDescV2
}

//nolint:gomnd
func (s *Service) liveBookingControl(buvid string, mid, upMid, seasonId int64, teenagersMode, lessonsMode int) bool {
	if teenagersMode != 0 || lessonsMode != 0 {
		return false
	}
	if mid == upMid || seasonId > 0 { //up主本人 || 是合集
		return false
	}
	//白名单
	if _, ok := s.c.LiveOrderMid[strconv.FormatInt(mid, 10)]; ok {
		return true
	}
	//灰度逻辑
	group := crc32.ChecksumIEEE([]byte(buvid+"_ugc_booking_live")) % 10
	if _, ok := s.c.LiveOrderGray[strconv.Itoa(int(group))]; ok {
		return true
	}
	return false
}

func (s *Service) ArcView(c context.Context, aid int64, mid int64, mobiApp, device, platform string, plat int8) (*api.ViewReply, map[string]string, error) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	var (
		vp    *api.ViewReply
		extra map[string]string
	)
	v, ok := s.bnjArcs[aid]
	if ok {
		arc := *v.Arc
		vp = &api.ViewReply{
			Arc:   &arc,
			Pages: v.Pages,
		}
		// bnj活动页extra 配置到配置文件内处理，暂不增加
		extra = make(map[string]string)
	} else {
		var err error
		eg := egV2.WithContext(c)
		eg.Go(func(c context.Context) (err error) { //错误直接抛出
			vp, err = cfg.dep.Archive.View3(c, aid, mid, mobiApp, device, platform)
			return
		})
		eg.Go(func(c context.Context) error {
			extra, _ = cfg.dep.ArchiveExtra.GetArchiveExtraValue(c, aid)
			return nil
		})
		if err = eg.Wait(); err != nil {
			return nil, nil, err
		}
	}
	if vp == nil || vp.Arc == nil || len(vp.Pages) == 0 || !vp.Arc.IsNormalV2() || vp.Arc.AttrVal(api.AttrBitIsPUGVPay) == api.AttrYes {
		return nil, nil, ecode.NothingFound
	}
	//旧版本首映前不展示 + 首映稿件首映前是竖屏不展示 +只有iphone和android返回首映前的稿件
	if IsPremierePortrait(c, vp.Arc, plat) {
		return nil, nil, ecode.NothingFound
	}
	if _, ok := s.specialMids[vp.Author.Mid]; ok && env.DeployEnv == env.DeployEnvProd {
		log.Error("aid(%d) mid(%d) can not view on prod", vp.Aid, vp.Author.Mid)
		return nil, nil, ecode.NothingFound
	}
	if tools.CheckNeedFilterMid64(c) && !tools.IsInt32Mid(vp.Author.Mid) {
		s.prom.Incr("mid64:arcview不支持")
		log.Error("aid(%d) mid(%d) mid64:arcview不支持", vp.Aid, vp.Author.Mid)
		return nil, nil, ecode.NothingFound
	}
	return vp, extra, nil
}

func IsPremierePortrait(c context.Context, a *api.Arc, plat int8) bool {
	//无首映信息跳过
	if a.Premiere == nil {
		return false
	}
	//首映稿件首映前是竖屏不展示
	width := a.Dimension.Width
	height := a.Dimension.Height
	//交换位置
	if a.Dimension.Rotate > 0 {
		width, height = height, width
	}
	//是竖屏
	if height > width {
		//首映稿件首映前是竖屏不展示
		if a.Premiere.State == api.PremiereState_premiere_before {
			return true
		} else if a.Premiere.State == api.PremiereState_premiere_in || a.Premiere.State == api.PremiereState_premiere_after {
			//首映中和首映后去掉首映信息，按照普通稿件返回
			a.Premiere = nil
			return false
		}
	}
	//旧版本 首映前不展示(只展示iphone + android 新版本)
	if pd.WithContext(c).Where(func(pd *pd.PDContext) {
		pd.IsPlatIPhone().And().Build(">=", int64(66700000))
	}).OrWhere(func(pd *pd.PDContext) {
		pd.IsPlatAndroid().And().Build(">=", int64(6670000))
	}).FinishOr(true) {
		return false
	}
	if a.Premiere.State == api.PremiereState_premiere_before {
		return true
	}
	return false
}

// ViewPage view page data.
// nolint:gocognit
func (s *Service) ViewPage(c context.Context, mid int64, plat int8, build int, mobiApp, device, cdnIP string, nMovie bool, buvid, slocale, clocale string, vp *api.ViewReply, pageVersion, spmid, platform string, teenagersMode int, extra map[string]string) (v *view.View, err error) {
	const (
		_androidMovie     = 5220000
		_iPhoneMovie      = 6500
		_iPadMovie        = 6720
		_iPadHDMovie      = 12020
		_androidUgcSeason = 5425000
		_iPhoneUgcSeason  = 8530 // 5.42.1
	)
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	vs := &view.ViewStatic{Arc: vp.Arc}
	if i18n.PreferTraditionalChinese(c, slocale, clocale) {
		i18n.TranslateAsTCV2(&vs.Title, &vs.Desc)
	}
	if s.displaySteins(c, vs, mobiApp, device, build) {
		vp.Pages = []*api.Page{}
	} else {
		s.initPages(c, vs, vp.Pages, mobiApp, build)
	}
	// TODO 产品最帅了！
	vs.Stat.DisLike = 0
	bvID, err := bvid.AvToBv(vs.Aid)
	if err != nil {
		log.Error("avtobv aid:%d err(%v)", vs.Aid, err)
		return nil, ecode.NothingFound
	}
	v = &view.View{ViewStatic: vs, DMSeg: 1, BvID: bvID}
	if v.AttrVal(api.AttrBitIsPGC) != api.AttrYes {
		// check access
		if err = s.checkAccess(c, mid, v.Aid, int(v.State), int(v.Access), vs.Arc); err != nil {
			// archive is ForbitFixed and Transcoding and StateForbitDistributing need analysis history body .
			return nil, err
		}
		if v.Access > 0 {
			v.Stat.View = 0
		}
	}
	var arcAddit *vuApi.ArcViewAdditReply
	g := egV2.WithContext(c)
	// 地区版权校验
	g.Go(func(ctx context.Context) (err error) {
		if s.overseaCheckV2(ctx, vs.Arc, plat) {
			return ecode.AreaLimit
		}
		// check region area limit
		if err = s.areaLimit(ctx, plat, int(vs.TypeID)); err != nil {
			return err
		}
		loc, _ := cfg.dep.Location.Info2(c)
		// 相关推荐AI使用zoneID取zoneID[3]
		if loc != nil && len(loc.ZoneId) >= 4 {
			v.ZoneID = loc.ZoneId[3]
		}
		download := int64(location.StatusDown_AllowDown) // by default it's allowed
		if v.AttrVal(api.AttrBitLimitArea) == api.AttrYes {
			if v.ZoneID == 0 {
				return ecode.NothingFound
			}
			if download, err = s.ipLimit(ctx, mid, v.Aid, cdnIP); err != nil {
				log.Error("aid(%d) mid(%d) ip(%s) cdn_ip(%s) error(%+v)", v.Aid, mid, metadata.String(ctx, metadata.RemoteIP), cdnIP, err)
				return err
			} else if v.AttrVal(api.AttrBitSteinsGate) == api.AttrYes {
				download = int64(location.StatusDown_ForbiddenDown)
			}
		}
		// 付费稿件不能下载
		if v.AttrValV2(api.AttrBitV2Pay) == api.AttrYes && v.Rights.ArcPayFreeWatch == 0 {
			download = int64(location.StatusDown_ForbiddenDown)
		}
		if download == int64(location.StatusDown_ForbiddenDown) {
			v.Rights.Download = int32(download)
			return
		}
		for _, p := range v.Pages {
			if p.From == "qq" {
				download = int64(location.StatusDown_ForbiddenDown)
				break
			}
		}
		v.Rights.Download = int32(download)
		return nil
	})
	// 校验稿件审核屏蔽状态
	g.Go(func(ctx context.Context) (err error) {
		if arcAddit, err = cfg.dep.VideoUP.ArcViewAddit(ctx, v.Aid); err != nil || arcAddit == nil {
			log.Error("s.vuDao.ArcViewAddit aid(%d) err(%+v) or arcAddit=nil", v.Aid, err)
			err = nil
			return
		}
		if arcAddit.ForbidReco != nil {
			v.ForbidRec = arcAddit.ForbidReco.State
		}
		return
	})
	if s.displaySteins(c, vs, mobiApp, device, build) {
		g.Go(func(ctx context.Context) (err error) {
			var steinView *steinApi.ViewReply
			if steinView, err = cfg.dep.Steins.View(ctx, v.Aid, mid, buvid); err != nil {
				log.Error("s.steinDao.View err(%v)", err)
				if ecode.EqualError(mainEcode.NonValidGraph, err) {
					err = ecode.NothingFound
				}
				return
			}
			if steinView.Graph == nil {
				err = ecode.NothingFound
				return
			}
			vp.Pages = []*api.Page{view.ArchivePage(steinView.Page)}
			vp.FirstCid = steinView.Page.Cid
			v.Interaction = &viewApi.Interaction{
				GraphVersion: steinView.Graph.Id,
				Mark:         steinView.Mark,
			}
			if steinView.Evaluation != "" { // 稿件综合评分
				v.Interaction.Evaluation = steinView.Evaluation
			}
			if steinView.ToastMsg != "" {
				v.Interaction.Msg = steinView.ToastMsg
			}
			if steinView.CurrentNode != nil {
				v.Interaction.HistoryNode = &viewApi.Node{
					Cid:    steinView.CurrentNode.Cid,
					Title:  steinView.CurrentNode.Name,
					NodeId: steinView.CurrentNode.Id,
				}
			}
			s.initPages(c, vs, vp.Pages, mobiApp, build)
			return
		})
	}
	if ((plat == model.PlatAndroid && build > _androidUgcSeason) ||
		(plat == model.PlatIPhone && build > _iPhoneUgcSeason) || plat == model.PlatIpadHD ||
		(plat == model.PlatIPad && build >= s.c.BuildLimit.UgcSeasonIPadBuild) ||
		(plat == model.PlatAndroidI && build >= s.c.BuildLimit.UgcSeasonAndroidIBuild) ||
		(plat == model.PlatAndroidHD && build >= s.c.BuildLimit.UgcSeasonAndroidHDBuild) ||
		(plat == model.PlatIPhoneI && build >= s.c.BuildLimit.UgcSeasonIphoneIBuild)) && v.SeasonID != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if ugcSn, err := cfg.dep.UGCSeason.Season(ctx, v.SeasonID); err == nil && ugcSn != nil { // ugc剧集
				v.UgcSeason = new(view.UgcSeason)
				v.UgcSeason.FromSeason(ugcSn)
				if !s.newSeasonTypeBuild(ctx) && v.UgcSeason.SeasonType == viewApi.SeasonType_Base { //老版本不展示基础合集
					v.UgcSeason = nil
					return nil
				}
				//青少年模式下不返回ugc的入口
				if teenagersMode == 1 && ugcSn.Season.AttrVal(seasonApi.AttrSnTeenager) != seasonApi.AttrSnYes {
					v.UgcSeason = nil
					return nil
				}
				// 是否是合集付费类型
				if ugcSn.Season.AttrVal(seasonApi.SeasonAttrSnPay) == seasonApi.AttrSnYes {
					v.UgcSeason.SeasonPay = true
					v.UgcSeason.LabelTextNew = _signPyText
					// 根据当前稿件aid查找付费合集信息
					if v.Arc.Pay == nil {
						log.Error("日志告警 商品信息错误,付费合集商品信息为空 seasonId:%+v", v.SeasonID)
						return nil
					}
					for _, good := range v.Arc.Pay.GoodsInfo {
						// 一个稿件绑定多个商品，找到第一个类型为合集的商品
						if good.Category == api.Category_CategorySeason {
							// 设置ugc goodinfo
							v.UgcSeason.FormGoodInfo(good)
							// 设置购买按钮
							v.UgcSeason.NewPayedButton()
							// 更新Episode付费状态
							v.UgcSeason.UpdateEpisodePayState()
							break
						}
					}
				}
				//基础合集返回"是否展示连续播放":只有pad端基础合集情况下使用
				if ugcSn.Season.AttrVal(seasonApi.AttrSnType) == seasonApi.AttrSnYes {
					v.UgcSeason.ShowContinualButton = s.c.Custom.SeasonContinualButtonSwitch
				}
				//是否有合集打卡活动
				if v.UgcSeason != nil {
					req := &checkin.ActivityReq{
						Type: 1,
						Oid:  v.SeasonID,
						Aid:  v.Aid,
						Cid:  v.FirstCid,
						Mid:  mid,
					}
					v.UgcSeason.Activity, err = s.CheckinSeasonActivity(ctx, req)
					if err != nil && !ecode.EqualError(ecode.NothingFound, err) {
						log.Error("s.CheckinSeasonActivity is err %+v %+v", err, req)
					}
					if v.UgcSeason.Activity != nil {
						v.UgcSeason.SeasonAbility = []string{_seasonAbilityCheck}
					}
				}
				//合集数量为1 && 不属于付费合集 && 不为合集打卡 则隐藏
				if ugcSn.Season.GetEpCount() == 1 && ugcSn.Season.AttrVal(seasonApi.SeasonAttrSnPay) != seasonApi.AttrSnYes && v.UgcSeason.Activity == nil {
					v.UgcSeason = nil
					return nil
				}
				s.prom.Incr("Season_Show")
			}
			return
		})
	}
	if mid > 0 || buvid != "" {
		g.Go(func(ctx context.Context) (err error) {
			v.History, _ = cfg.dep.History.Progress(ctx, v.Aid, mid, buvid)
			return
		})
	}
	if v.AttrVal(api.AttrBitIsPGC) == api.AttrYes {
		if (v.AttrVal(api.AttrBitIsMovie) != api.AttrYes) || (plat == model.PlatAndroid && build >= _androidMovie) || (plat == model.PlatIPhone && build >= _iPhoneMovie) || (plat == model.PlatIPad && build >= _iPadMovie) ||
			(plat == model.PlatIpadHD && build > _iPadHDMovie) || plat == model.PlatAndroidTVYST || plat == model.PlatAndroidTV || plat == model.PlatAndroidI || plat == model.PlatIPhoneB {
			g.Go(func(ctx context.Context) error {
				return s.initPGC(ctx, v, mid, build, mobiApp, device)
			})
		} else {
			g.Go(func(ctx context.Context) error {
				return s.initMovie(ctx, v, mid, build, mobiApp, device, nMovie)
			})
		}
	} else {
		if v.Rights.UGCPay == 1 && mid != v.Author.Mid {
			g.Go(func(ctx context.Context) (err error) {
				if err = s.initUGCPay(ctx, v, plat, mid, build); err != nil {
					log.Error("%+v", err)
					err = nil
					return
				}
				return nil
			})
		}
	}
	//获取在看人数信息
	g.Go(func(ctx context.Context) error {
		s.initOnline(v, buvid, mid, v.Aid)
		return nil
	})
	// 获取播放页小黄条,全站热歌 > 妙评
	var musicH, miaoH *viewApi.Honor
	g.Go(func(ctx context.Context) error {
		// 全站热歌
		if (mobiApp == "android" && build >= _musicHonorControlAndroid) || (mobiApp == "iphone" && build >= _musicHonorControlIos) {
			if musicID, ok := extra[_hotMusicKey]; ok {
				musicHonor, err := cfg.dep.Music.ToplistEntrance(ctx, v.Aid, musicID)
				if err != nil {
					log.Error("d.d.musicClient.ToplistEntrance aid:%d musicId:%s err:%+v", v.Aid, musicID, err)
					return nil
				}
				if musicHonor != nil && musicHonor.ArcHonor != nil {
					musicH = view.FromMusicHonor(musicHonor)
				}
				return nil
			}
		}
		return nil
	})
	// 秒评
	g.Go(func(ctx context.Context) error {
		if (mobiApp == "android" && build >= _superReplyControlAndroid) || (mobiApp == "iphone" && build >= _superReplyControlIos) {
			if _, ok := extra[_superReplyKey]; ok {
				replyHonor, err := cfg.dep.Reply.GetArchiveHonor(ctx, v.Aid)
				if err != nil {
					log.Error("s.replyDap.GetArchiveHonor aid:%d err:%+v", v.Aid, err)
					return nil
				}
				if replyHonor != nil && replyHonor.ArchiveHonor != nil {
					miaoH = view.FromReplyHonor(replyHonor)
				}
				return nil
			}
		}
		return nil
	})
	var initRly *view.InitTag
	g.Go(func(ctx context.Context) error {
		//其他标签
		initRly = s.initTag(ctx, v.Arc, v.FirstCid, mid, plat, build, pageVersion, buvid, mobiApp, spmid, platform, extra)
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	// tag 返回值处理
	s.initResult(v, initRly)
	// honor 返回值处理 全站热歌 > 妙评
	v.Honor = musicH
	if v.Honor == nil {
		v.Honor = miaoH
	}
	return v, nil
}

func (s *Service) CheckinSeasonActivity(ctx context.Context, req *checkin.ActivityReq) (*viewApi.UgcSeasonActivity, error) {
	reply, err := s.checkinDao.CheckinActivity(ctx, req)
	if err != nil {
		return nil, err
	}
	if reply == nil {
		return nil, ecode.NothingFound
	}
	resp := &viewApi.UgcSeasonActivity{
		Type:             reply.Type,
		Oid:              reply.Oid,
		ActivityId:       reply.ActivityId,
		Title:            reply.Title,
		Intro:            reply.Intro,
		DayCount:         reply.DayCount,
		UserCount:        reply.UserCount,
		JoinDeadline:     reply.JoinDeadline,
		ActivityDeadline: reply.ActivityDeadline,
		CheckinViewTime:  reply.CheckinViewTime,
		NewActivity:      reply.NewActivity,
	}
	if reply.UserActivity != nil {
		resp.UserActivity = &viewApi.UserActivity{
			UserState:       reply.UserActivity.UserState,
			LastCheckinDate: reply.UserActivity.LastCheckinDate,
			CheckinToday:    reply.UserActivity.CheckinToday,
			UserDayCount:    reply.UserActivity.UserDayCount,
			UserViewTime:    reply.UserActivity.UserViewTime,
			Portrait:        reply.UserActivity.Portrait,
		}
	}
	if reply.SeasonShow != nil {
		resp.SeasonShow = &viewApi.SeasonShow{
			ButtonText:    reply.SeasonShow.ButtonText,
			JoinText:      reply.SeasonShow.JoinText,
			RuleText:      reply.SeasonShow.RuleText,
			CheckinText:   reply.SeasonShow.CheckinText,
			CheckinPrompt: reply.SeasonShow.CheckinPrompt,
		}
	}
	return resp, nil
}

func HideArcAttribute(arc *api.Arc) {
	if arc == nil {
		return
	}
	arc.Access = 0
	arc.Attribute = 0
	arc.AttributeV2 = 0
}

// ShareIcon .
func (s *Service) ShareIcon(c context.Context, mid, aid, build int64, plat int8, buvid string) (icon *view.ShareIcon) {
	icon = &view.ShareIcon{}
	if model.IsOverseas(plat) { // 国际版对有些渠道不支持，所以始终返回default
		icon.ShareChannel = model.ShareDefaultStr
		return
	}
	rly, err := s.shareDao.LastChannel(c, &sharerpc.LastChannelReq{
		Mid:   mid,
		Aid:   aid,
		Buvid: buvid,
		Ip:    metadata.String(c, metadata.RemoteIP),
		Type:  _shareType,
		Build: build,
	})
	// 不下发错误，默认返回0值
	if err != nil {
		log.Error("s.shareClient.LastChannel(%d) error(%v)", mid, err)
		return
	}
	var ok bool
	if icon.ShareChannel, ok = model.ShareChannelToString[rly.Channel]; !ok {
		icon.ShareChannel = model.ShareDefaultStr
	}
	return
}

// AddShare add a share
func (s *Service) AddShare(c context.Context, aid, mid, build int64, shareChannel, ip, buvid string) (share int64, isReport bool, upID int64, toast string, err error) {
	var a *api.Arc
	if a, err = s.arcDao.Archive(c, aid); err != nil {
		if errors.Cause(err) == ecode.NothingFound {
			err = avecode.ArchiveNotExist
		}
		return
	}
	if !a.IsNormal() {
		err = avecode.ArchiveNotExist
		return
	}
	upID = a.Author.Mid
	shareReply, err := s.shareDao.AddShareClick(c, &view.ShareParam{
		OID:          aid,
		Type:         model.ShareTypeAV,
		ShareChannel: shareChannel,
		Build:        build,
	}, mid, upID, buvid, "old", nil)
	if shareReply != nil && shareReply.IsFirstShare && mid > 0 {
		toast = "每天首次分享成功，等级经验值+5"
	} else {
		toast = "分享成功"
	}
	if err != nil {
		if ecode.EqualError(mainEcode.ShareAlreadyAdd, err) {
			err = nil
			return
		}
		toast = ""
		log.Error("s.shareClient.AddShare(%d, %d, 3) error(%v)", aid, mid, err)
		return
	}
	if shareReply != nil && shareReply.Count > int64(a.Stat.Share) {
		share = shareReply.Count
		isReport = true
	}
	return
}

// Shot shot service
func (s *Service) Shot(c context.Context, aid, cid int64, mobiApp string, build int32) (shot *view.Videoshot, err error) {
	var (
		arcShot        *api.VideoShot
		videoViewReply *vuApi.VideoPointsReply
		eg             = egV2.WithContext(c)
	)
	shot = new(view.Videoshot)
	// 获取设备信息
	dev, _ := device.FromContext(c)
	plat := model.PlatNew(dev.RawMobiApp, dev.Device)
	eg.Go(func(ctx context.Context) (err error) {
		//获取缩略图
		if arcShot, err = s.arcDao.Shot(ctx, aid, cid, plat, dev.Build, dev, dev.Model); err != nil {
			log.Error("s.arcDao.Shot err(%+v)", err)
			return nil
		}
		if arcShot != nil {
			if mobiApp == "android_b" || (mobiApp == "android" && build < s.c.Custom.VideoShotAndBuild && s.c.Custom.VideoShotGray > aid%100) ||
				(mobiApp == "iphone" && build > s.c.Custom.VideoShotIOSBuild && s.c.Custom.VideoShotGrayIOS > aid%100) {
				for k, img := range arcShot.Image {
					arcShot.Image[k] = handleVideoShot(img)
				}
			}
			shot.VideoShot = arcShot
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		//获取"高能看点"
		videoViewReply, err = s.vuDao.GetVideoViewPoints(ctx, aid, cid)
		if err != nil {
			log.Error("GetVideoViewPoints err(%+v)", err)
			return nil
		}
		if videoViewReply != nil && len(videoViewReply.Points) > 0 {
			for _, p := range videoViewReply.Points {
				tmpPoint := creative.Points{
					Type:    int(p.Type),
					From:    int64(p.From),
					To:      int64(p.To),
					Content: p.Content,
					Cover:   p.ImgUrl,
				}
				videoShotVersion := feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.VideoShotBuild, nil)
				//旧版本不返回type=2（分段章节）的数据
				if p.Type == 2 && !videoShotVersion {
					continue
				}
				shot.Points = append(shot.Points, &tmpPoint)
			}
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.wait() err(%+v) aid(%d) cid(%d)", err, aid, cid)
	}
	return
}

// Like add a like.
func (s *Service) Like(c context.Context, aid, mid int64, status int8, ogvType int64, buvid, platform, path, appkey, ua, build, mobiApp, device string) (upperID int64, toast string, err error) {
	var (
		a    *api.SimpleArc
		typ  thumbup.Action
		stat *thumbup.LikeReply
	)
	if status == 0 {
		if a, err = s.arcDao.SimpleArc(c, aid); err != nil {
			if errors.Cause(err) == ecode.NothingFound {
				err = avecode.ArchiveNotExist
			}
			return
		}
		if !a.IsNormal() {
			err = avecode.ArchiveNotExist
			return
		}
		upperID = a.Mid
		typ = thumbup.Action_ACTION_LIKE
		// 点赞前先判断风控
		tec := &view.SilverEventCtx{
			Action:     model.SilverActionLike,
			Aid:        aid,
			UpID:       upperID,
			Mid:        mid,
			PubTime:    time.Unix(a.Pubdate, 0).Format("2006-01-02 15:04:05"),
			LikeSource: model.SilverSourceLike,
			Buvid:      buvid,
			Ip:         metadata.String(c, metadata.RemoteIP),
			Platform:   platform,
			Ctime:      time.Now().Format("2006-01-02 15:04:05"),
			Api:        path,
			Origin:     appkey,
			UserAgent:  ua,
			Build:      build,
			Token:      riskcontrol.ReportedLoginTokenFromCtx(c),
		}
		if s.silverDao.RuleCheck(c, tec, model.SilverSceneLike) {
			err = mainEcode.SilverBulletLikeReject
			return
		}
	} else if status == 1 {
		typ = thumbup.Action_ACTION_CANCEL_LIKE
	}
	if stat, err = s.thumbupDao.Like(c, mid, upperID, _businessLike, aid, typ, true, mobiApp, device, platform); err != nil {
		if ecode.EqualError(tecode.ThumbupDupLikeErr, err) {
			log.Error("%+v", err)
			err = nil
			toast = "点赞收到！视频可能推荐哦"
		}
		return
	}
	if typ == thumbup.Action_ACTION_LIKE {
		toast = s.likeToast(ogvType, stat.LikeNumber)
	}
	return
}

//nolint:gomnd
func (s *Service) likeToast(vType int64, likeNumber int64) string {
	toast := map[int64][]string{
		0: {"点赞收到！视频可能推荐哦", "感谢点赞，推荐已收到啦", "get！视频也许更多人能看见！", "点赞爆棚，感谢推荐！"},                      // ugc
		1: {"眼光独到，感谢推荐！", "收到好评，您的推荐已收到！", "多谢好评！作品有几率被推荐哦！", "好评爆棚！作品被推荐的姿势增加了！"},               // 电影
		2: {"眼光独到，感谢推荐本集作品！", "收到好评，您对本集的推荐已收到！", "多谢好评！本集有几率被推荐哦！", "好评爆棚！本集被推荐的姿势增加了！"},        // 非电影
		3: {"感谢原始粉丝，此视频将可能被更多人看见", "多谢好评，感谢原始粉丝的支持~", "感谢原始粉丝，此视频将更大可能被推荐哦~", "好评爆棚！感谢原始粉丝的陪伴~"}, // 契约者
		4: {"感谢老粉，此视频将可能被更多人看见", "多谢好评，感谢老粉的支持~", "感谢老粉，此视频将更大可能被推荐哦~", "好评爆棚！感谢老粉的陪伴~"},         // 老粉
		5: {"发现宝藏！感谢硬核推荐！", "硬核点赞，感谢你的推荐！", "已收到来自硬核指挥部的推荐！", "硬核推荐能量发动！"},                       // 硬核会员
	}
	tt, ok := toast[vType]
	if !ok {
		tt = toast[0]
	}
	if likeNumber <= 99 {
		return tt[0]
	} else if likeNumber >= 100 && likeNumber <= 999 {
		return tt[1]
	} else if likeNumber >= 1000 && likeNumber <= 9999 {
		return tt[2]
	} else {
		return tt[3]
	}
}

// LikeNoLogin is for no login user like
func (s *Service) LikeNoLogin(c context.Context, aid, ogvType int64, status int32, buvid, platform, path, appkey, ua, build string) (int64, *view.LikeNoLoginRes, error) {
	toast := ""
	upperID := int64(0)
	typ := thumbup.Action_ACTION_LIKE
	if status == 1 {
		typ = thumbup.Action_ACTION_CANCEL_LIKE
	}
	if typ == thumbup.Action_ACTION_LIKE {
		a, err := s.arcDao.Archive(c, aid)
		if err != nil {
			if errors.Cause(err) == ecode.NothingFound {
				err = avecode.ArchiveNotExist
			}
			return 0, nil, err
		}
		if !a.IsNormal() {
			err = avecode.ArchiveNotExist
			return 0, nil, err
		}
		likeNum := a.Stat.Like + 1
		upperID = a.Author.Mid
		// 点赞前先判断风控
		tec := &view.SilverEventCtx{
			Action:     model.SilverActionLike,
			Aid:        aid,
			UpID:       upperID,
			Mid:        0,
			PubTime:    a.PubDate.Time().Format("2006-01-02 15:04:05"),
			LikeSource: model.SilverSourceNologin,
			Buvid:      buvid,
			Ip:         metadata.String(c, metadata.RemoteIP),
			Platform:   platform,
			Ctime:      time.Now().Format("2006-01-02 15:04:05"),
			Api:        path,
			Origin:     appkey,
			UserAgent:  ua,
			Build:      build,
			Token:      riskcontrol.ReportedLoginTokenFromCtx(c),
		}
		if s.silverDao.RuleCheck(c, tec, model.SilverSceneLike) {
			return upperID, nil, mainEcode.SilverBulletLikeReject
		}
		toast = s.likeToast(ogvType, int64(likeNum))
	}
	// 未登录点赞是否需要唤起登录 上线时默认否
	res := &view.LikeNoLoginRes{Toast: toast, NeedLogin: s.c.Custom.LikeNeedLogin}
	if err := s.thumbupDao.LikeNoLogin(c, upperID, _businessLike, buvid, aid, typ, true); err != nil {
		log.Error("s.thumbupDao.LikeNoLogin err:%+v", err)
		if ecode.EqualError(tecode.ThumbupDupLikeErr, err) {
			return upperID, res, nil
		}
		return upperID, nil, err
	}
	return upperID, res, nil
}

// Dislike add a dislike.
func (s *Service) Dislike(c context.Context, aid, mid int64, status int8, mobiApp, device, platform string) (upperID int64, err error) {
	var (
		a   *api.SimpleArc
		typ thumbup.Action
	)
	if status == 0 {
		if a, err = s.arcDao.SimpleArc(c, aid); err != nil {
			if errors.Cause(err) == ecode.NothingFound {
				err = avecode.ArchiveNotExist
			}
			return
		}
		if !a.IsNormal() {
			err = avecode.ArchiveNotExist
			return
		}
		upperID = a.Mid
		typ = thumbup.Action_ACTION_DISLIKE
	} else if status == 1 {
		typ = thumbup.Action_ACTION_CANCEL_DISLIKE
	}
	_, err = s.thumbupDao.Like(c, mid, upperID, _businessLike, aid, typ, false, mobiApp, device, platform)
	return
}

// AddCoin add a coin
func (s *Service) AddCoin(c context.Context, aid, mid, upID, avtype, multiply int64, selectLike int, buvid, platform, path, appkey, ua, build, mobiApp, device string) (bool, bool, error) {
	var maxCoin int64 = 2
	var typeID int16
	var pubTime int64
	if avtype == _avTypeAv {
		a, err := s.arcDao.Archive(c, aid)
		if err != nil {
			if errors.Cause(err) == ecode.NothingFound {
				err = avecode.ArchiveNotExist
			}
			return false, false, err
		}
		if !a.IsNormal() {
			return false, false, avecode.ArchiveNotExist
		}
		upID = a.Author.Mid
		typeID = int16(a.TypeID)
		pubTime = int64(a.PubDate)
		// 投币&点赞前先判断风控
		tec := &view.SilverEventCtx{
			Action:    model.SilverActionCoin,
			Mid:       mid,
			UpID:      upID,
			Aid:       aid,
			ItemType:  _typeAv,
			CoinNum:   multiply,
			Title:     a.Title,
			PlayNum:   int64(a.Stat.View),
			PubTime:   a.PubDate.Time().Format("2006-01-02 15:04:05"),
			Buvid:     buvid,
			Ip:        metadata.String(c, metadata.RemoteIP),
			Platform:  platform,
			Ctime:     time.Now().Format("2006-01-02 15:04:05"),
			Api:       path,
			Origin:    appkey,
			UserAgent: ua,
			Build:     build,
			Token:     riskcontrol.ReportedLoginTokenFromCtx(c),
		}
		scene := model.SilverSceneCoin
		if selectLike == 1 {
			tec.Action = model.SilverActionCointolike
			scene = model.SilverSceneCointolike
		}
		if s.silverDao.RuleCheck(c, tec, scene) {
			return false, true, mainEcode.SilverBulletLikeReject
		}
		// pgc视频不做maxCoin数限制
		if a.AttrVal(api.AttrBitIsPGC) != api.AttrYes && a.Copyright == int32(api.CopyrightCopy) {
			maxCoin = 1
		}
	}
	err := s.coinDao.AddCoins(c, aid, mid, upID, maxCoin, avtype, multiply, typeID, pubTime, mobiApp, device, platform)
	if err != nil {
		return false, false, err
	}
	like := false
	prompt := false
	eg := egV2.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if avtype == _avTypeAv && selectLike == 1 {
			if _, err = s.thumbupDao.Like(ctx, mid, upID, _businessLike, aid, thumbup.Action_ACTION_LIKE, false, mobiApp, device, platform); err != nil {
				log.Error("s.thumbupDao.Like aid(%d) mid(%d) err(%+v)", aid, mid, err)
			} else {
				like = true
			}
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if prompt, err = s.relDao.Prompt(ctx, mid, upID, _promptCoin); err != nil {
			log.Error("s.relDao.Prompt mid(%d) aid(%d) upid(%d) err(%+v)", mid, aid, upID, err)
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.wait() err(%+v)", err)
	}
	return prompt, like, nil
}

// Paster get paster if nologin.
func (s *Service) Paster(c context.Context, plat, adType int8, aid, typeID, buvid string) (p *resource.Paster, err error) {
	if p, err = s.rscDao.Paster(c, plat, adType, aid, typeID, buvid); err != nil {
		log.Error("%+v", err)
	}
	return
}

// VipPlayURL get playurl token.
//
//nolint:gomnd
func (s *Service) VipPlayURL(c context.Context, aid, cid, mid int64) (res *view.VipPlayURL, err error) {
	var (
		a    *api.SimpleArc
		card *accApi.Card
	)
	res = &view.VipPlayURL{
		From: "app",
		Ts:   time.Now().Unix(),
		Aid:  aid,
		Cid:  cid,
		Mid:  mid,
	}
	if card, err = s.accDao.Card3(c, mid); err != nil || card == nil {
		log.Error("s.accDao.Card3 err(%+v) or card=nil", err)
		err = ecode.AccessDenied
		return
	}
	if res.VIP = int(card.Level); res.VIP > 6 {
		res.VIP = 6
	}
	if card.Vip.Type != 0 && card.Vip.Status == 1 {
		res.SVIP = 1
	}
	if a, err = s.arcDao.SimpleArc(c, aid); err != nil {
		log.Error("%+v", err)
		err = ecode.NothingFound
		return
	}
	if mid == a.Mid {
		res.Owner = 1
	}
	params := url.Values{}
	params.Set("from", res.From)
	params.Set("ts", strconv.FormatInt(res.Ts, 10))
	params.Set("aid", strconv.FormatInt(res.Aid, 10))
	params.Set("cid", strconv.FormatInt(res.Cid, 10))
	params.Set("mid", strconv.FormatInt(res.Mid, 10))
	params.Set("vip", strconv.Itoa(res.VIP))
	params.Set("svip", strconv.Itoa(res.SVIP))
	params.Set("owner", strconv.Itoa(res.Owner))
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	mh := md5.Sum([]byte(strings.ToLower(tmp) + s.c.PlayURL.Secret))
	res.Fcs = hex.EncodeToString(mh[:])
	return
}

// Follow get auto follow switch from creative and acc.
func (s *Service) Follow(c context.Context, vmid, mid int64) (res *creative.PlayerFollow, err error) {
	var (
		fl bool
		fs *upApi.UpSwitchReply
	)
	g := egV2.WithContext(c)
	if mid > 0 {
		g.Go(func(ctx context.Context) (err error) {
			fl, err = s.accDao.Following3(ctx, mid, vmid)
			if err != nil {
				log.Error("%+v", err)
			}
			return
		})
	}
	g.Go(func(ctx context.Context) (err error) {
		fs, err = s.creativeDao.FollowSwitch(ctx, vmid)
		if err != nil {
			log.Error("%+v", err)
		}
		return
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	res = &creative.PlayerFollow{}
	if fs != nil && fs.State == 1 && !fl {
		res.Show = true
	}
	return
}

// UpperRecmd is
func (s *Service) UpperRecmd(c context.Context, plat int8, platform, mobiApp, device, buvid string, build int, mid, vimd int64) (res card.Handler, err error) {
	var (
		upIDs          []int64
		follow         *operate.Card
		cardm          map[int64]*accApi.Card
		statm          map[int64]*relationgrpc.StatReply
		interrelations map[int64]*relationgrpc.InterrelationReply
	)

	if follow, err = s.searchFollow(c, platform, mobiApp, device, buvid, build, mid, vimd); err != nil {
		log.Error("%+v", err)
		return
	}
	if follow == nil {
		err = xecode.AppNotData
		log.Error("follow is nil")
		return
	}
	for _, item := range follow.Items {
		upIDs = append(upIDs, item.ID)
	}
	g := egV2.WithContext(c)
	if len(upIDs) != 0 {
		g.Go(func(ctx context.Context) (err error) {
			if cardm, err = s.accDao.Cards3(ctx, upIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		g.Go(func(ctx context.Context) (err error) {
			if statm, err = s.relDao.StatsGRPC(ctx, upIDs); err != nil {
				log.Error("%+v", err)
			}
			return nil
		})
		if mid != 0 {
			g.Go(func(ctx context.Context) error {
				if interrelations, err = s.relDao.Interrelations(ctx, mid, upIDs); err != nil {
					log.Error("%+v", err)
				}
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		log.Error("g.wait() err(%+v)", err)
	}
	op := &operate.Card{}
	op.From(cdm.CardGt(model.GotoSearchUpper), 0, 0, plat, build, mobiApp)
	h := card.Handle(plat, cdm.CardGt(model.GotoSearchUpper), "", cdm.ColumnSvrSingle, nil, nil, nil, nil, statm, cardm, interrelations)
	if h == nil {
		err = xecode.AppNotData
		return
	}
	op = follow
	_ = h.From(nil, op)
	if h.Get().Right {
		res = h
	} else {
		err = xecode.AppNotData
	}
	return
}

// LikeTriple like & coin & fav
//
//nolint:gomnd
func (s *Service) LikeTriple(c context.Context, aid, mid int64, buvid, platform, path, appkey, ua, build, mobiApp, device string) (*view.TripleRes, bool, error) {
	maxCoin := int64(1)
	multiply := int64(1)
	a, err := s.arcDao.Archive(c, aid)
	if err != nil {
		if errors.Cause(err) == ecode.NothingFound {
			err = avecode.ArchiveNotExist
		}
		return nil, false, err
	}
	if !a.IsNormal() {
		err = avecode.ArchiveNotExist
		return nil, false, err
	}
	res := &view.TripleRes{
		UpID: a.Author.Mid,
	}
	// 三连前先判断风控
	tec := &view.SilverEventCtx{
		Action:    model.SilverActionTriple,
		Mid:       mid,
		UpID:      a.Author.Mid,
		Aid:       aid,
		ItemType:  _typeAv,
		Title:     a.Title,
		PlayNum:   int64(a.Stat.View),
		PubTime:   a.PubDate.Time().Format("2006-01-02 15:04:05"),
		Buvid:     buvid,
		Ip:        metadata.String(c, metadata.RemoteIP),
		Platform:  platform,
		Ctime:     time.Now().Format("2006-01-02 15:04:05"),
		Api:       path,
		Origin:    appkey,
		UserAgent: ua,
		Build:     build,
		Token:     riskcontrol.ReportedLoginTokenFromCtx(c),
	}
	if s.silverDao.RuleCheck(c, tec, model.SilverSceneTriple) {
		return res, true, mainEcode.SilverBulletLikeReject
	}
	if a.Copyright == int32(api.CopyrightOriginal) {
		maxCoin = 2
		multiply = 2
	}
	isReport := false
	eg := egV2.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		userCoins, _ := s.coinDao.UserCoins(ctx, mid)
		if userCoins < 1 { // 用户当前没有硬币，如之前投过则返回点亮成功
			arcUserCoins, _ := s.coinDao.ArchiveUserCoins(ctx, aid, mid, _coinAv)
			if arcUserCoins != nil && arcUserCoins.Multiply > 0 {
				res.Coin = true
			}
			return
		} else if userCoins < 2 { // 用户当前不足2个币，则只投1个
			multiply = 1
		}
		if err = s.coinDao.AddCoins(ctx, aid, mid, a.Author.Mid, maxCoin, _coinAv, multiply, int16(a.TypeID), int64(a.PubDate), mobiApp, device, platform); err != nil {
			if ecode.EqualError(xecode.CoinOverMax, err) { // 投币超上限，返回点亮成功
				isReport = true
				res.Coin = true
			} else { // 如其他错误，则判断用户之前是否投过币
				arcUserCoins, _ := s.coinDao.ArchiveUserCoins(ctx, aid, mid, _coinAv)
				if arcUserCoins != nil && arcUserCoins.Multiply > 0 {
					res.Coin = true
				}
			}
			log.Error("s.coinDao.AddCoins err(%+v) aid(%d) mid(%d)", err, aid, mid)
			err = nil
		} else {
			res.Multiply = multiply
			isReport = true
			res.Coin = true
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		res.Fav = s.favDao.IsFav(ctx, mid, aid, model.FavTypeVideo)
		if err = s.favDao.AddFav(ctx, mid, aid, 0, model.FavTypeVideo, mobiApp, platform, device); err != nil {
			log.Error("s.favDao.AddFav err(%+v) aid(%d) mid(%d)", err, aid, mid)
			if ecode.EqualError(favecode.FavVideoExist, err) || ecode.EqualError(favecode.FavResourceExist, err) {
				res.Fav = true
			}
			err = nil
		} else {
			res.Fav = true
			isReport = true
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if _, err = s.thumbupDao.Like(ctx, mid, res.UpID, _businessLike, aid, thumbup.Action_ACTION_LIKE, false, mobiApp, device, platform); err != nil {
			log.Error("s.thumbupDao.Like err(%+v) aid(%d) mid(%d)", err, aid, mid)
			if ecode.EqualError(tecode.ThumbupDupLikeErr, err) {
				res.Like = true
			}
			err = nil
		} else {
			res.Like = true
			isReport = true
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if res.Prompt, err = s.relDao.Prompt(ctx, mid, a.Author.Mid, _promptFav); err != nil {
			log.Error("s.relDao.Prompt err(%+v)", err)
			err = nil
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.wait() err(%+v", err)
	}
	if !res.Coin && !res.Fav {
		res.Prompt = false
	}
	return res, isReport, nil
}

// Material 全屏模式运营活动素材
// nolint:ineffassign,gomnd
func (s *Service) Material(c context.Context, params *view.MaterialParam) (res view.MaterialResArr, err error) {
	var material []*vuApi.MaterialViewRes
	if material, err = s.vuDao.MaterialView(c, params); err != nil {
		log.Error("s.vuDao.MaterialView err(%+v)", err)
		return
	}
	if len(material) == 0 {
		return
	}
	for _, m := range material {
		name, err := view.MaterialName(m.Type, m.Name)
		if err != nil {
			log.Error("MaterialView err type(%v)", err)
			err = nil
			continue
		}
		res = append(res, &view.MaterialRes{
			ID:       m.Id,
			Icon:     m.Icon,
			URL:      m.Url,
			Typ:      m.Type,
			Name:     name,
			BgColor:  _defaultBgColor,   // 后续等后台配置，图片优先
			JumpType: int32(m.JumpType), //跳转类型
		})
	}
	//优先级判断+大于2个只展示一个
	if len(res) >= 2 {
		res = priorityMaterial(res)
		return
	}
	return
}

// 优先级判断：1：活动；2：bgm；3：贴纸；4：Native 话题活动（B剪）；5:拍同款;6：合拍
// 当有2个及以上导流位时，仅露出1个，优先级为：4>6>5>1 >2 >3
func priorityMaterial(data view.MaterialResArr) view.MaterialResArr {
	sortArr := []int32{4, 6, 5, 1, 2, 3} //优先级
	for _, arr := range sortArr {
		for _, material := range data {
			if material.Typ != arr {
				continue
			}
			return view.MaterialResArr{material}
		}
	}
	return nil
}

func (s *Service) displaySteins(c context.Context, a *view.ViewStatic, mobiApp, device string, build int) bool {
	return a.AttrVal(api.AttrBitSteinsGate) == api.AttrYes && feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DisplaySteins, &feature.OriginResutl{
		MobiApp: mobiApp,
		Device:  device,
		Build:   int64(build),
		BuildLimit: (mobiApp == "iphone" && build > s.c.Custom.SteinsBuild.IosPink) ||
			(mobiApp == "android" && build >= s.c.Custom.SteinsBuild.Android) ||
			(mobiApp == "android_b" && build >= s.c.Custom.SteinsBuild.AndroidBlue) ||
			(mobiApp == "iphone_b" && build >= s.c.Custom.SteinsBuild.IosBlue) ||
			(mobiApp == "ipad" && build >= s.c.Custom.SteinsBuild.IpadHD) ||
			(mobiApp == "android_i" && build >= s.c.Custom.SteinsBuild.AndroidI) ||
			(mobiApp == "iphone_i" && build >= s.c.Custom.SteinsBuild.IphoneI) ||
			(mobiApp == "android_hd"),
	})
}

func (s *Service) displaySteinsLabel(c context.Context, a *view.ViewStatic, mobiApp, device string, build int) bool {
	return a.AttrVal(api.AttrBitSteinsGate) == api.AttrYes && feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DisplaySteinsLabel, &feature.OriginResutl{
		MobiApp: mobiApp,
		Device:  device,
		Build:   int64(build),
		BuildLimit: (mobiApp == "iphone" && build > _steinsLabelIos) ||
			(mobiApp == "android" && build > _steinsLabelAndroid) ||
			(mobiApp == "android_b" && build > _steinsLabelAndroidBlue) ||
			(mobiApp == "iphone_b" && build > _steinsLabelIosBlue) ||
			(mobiApp == "ipad" && build > _steinsLabelIpad) ||
			(mobiApp == "android_i" && build > _steinsLabelAndroidI) ||
			(mobiApp == "iphone_i" && build >= _steinsLabelIphoneI) ||
			(mobiApp == "android_hd"),
	})
}

// ShareClick when share click
// nolint:ineffassign
func (s *Service) ShareClick(c context.Context, params *view.ShareParam, mid int64, buvid, ua, path, referer, sid string) (int64, bool, error) {
	upID, arc, err := s.shareCheck(c, params)
	if err != nil {
		log.Error("shareCheck err(%+v) %+v", err, arc)
		return 0, false, err
	}
	if params.IsMelloi != "" {
		params.Type = model.ShareTypeMelloi
	}
	actData := &sharerpc.Metadata{
		Ua:      ua,
		Referer: referer,
		Url:     path,
		Csid:    sid,
		Ip:      metadata.String(c, metadata.RemoteIP),
		AppKey:  params.AppKey,
	}
	if _, err := s.shareDao.AddShareClick(c, params, mid, upID, buvid, "", actData); err != nil {
		if ecode.EqualError(mainEcode.ShareAlreadyAdd, err) {
			err = nil
			return upID, false, nil
		}
		log.Error("s.shareDao.AddShareClick param(%+v) mid(%d) error(%v)", params, mid, err)
		return 0, false, err
	}
	return upID, true, nil
}

// ShareComplete when share complete
func (s *Service) ShareComplete(c context.Context, params *view.ShareParam, mid int64, buvid string) (string, error) {
	upID, _, err := s.shareCheck(c, params)
	if err != nil {
		log.Error("shareCheck err(%+v)", err)
		return "", err
	}
	toast := "分享成功"
	if params.IsMelloi != "" {
		params.Type = model.ShareTypeMelloi
	}
	shareReply, err := s.shareDao.AddShareComplete(c, params, mid, upID, buvid)
	if err != nil {
		if ecode.EqualError(mainEcode.ShareAlreadyAdd, err) {
			return toast, nil
		}
		log.Error("s.shareDao.AddShareComplete param(%+v) mid(%d) error(%v)", params, mid, err)
		return "", err
	}
	if shareReply != nil && shareReply.IsFirstShare && mid > 0 {
		toast = "每天首次分享成功，等级经验值+5"
	}
	return toast, nil
}

func (s *Service) shareCheck(c context.Context, params *view.ShareParam) (int64, *api.Arc, error) {
	switch params.Type {
	case model.ShareTypeAV:
		a, err := s.arcDao.Archive(c, params.OID)
		if err != nil {
			if ecode.EqualError(ecode.NothingFound, err) {
				err = avecode.ArchiveNotExist
			}
			return 0, nil, err
		}
		if !a.IsNormal() {
			return 0, nil, avecode.ArchiveNotExist
		}
		return a.Author.Mid, a, nil
	case model.ShareTypeLive:
		return params.UpID, nil, nil
	default:
		return 0, nil, ecode.RequestErr
	}
}

func (s *Service) endPageTest(buvid string, mid int64) (half, full int) {
	for _, cmid := range s.c.Custom.EndPageMids {
		if mid == cmid {
			return 1, 1
		}
	}
	if buvid == "" {
		return 0, 0
	}
	// half screen end page test
	if crc32.ChecksumIEEE([]byte(buvid+"_endpage"))%10 == s.c.Custom.EndPageHalfGroup {
		half = 1
	}
	// full screen end page test
	if crc32.ChecksumIEEE([]byte(buvid+"_endpage"))%10 == s.c.Custom.EndPageFullGroup {
		full = 1
	}
	return
}

func (s *Service) ViewMaterial(c context.Context, arg *viewApi.ViewMaterialReq, mid int64, build int32, device, mobiApp, platform, buvid string) (*viewApi.ViewMaterialReply, error) {
	rly := &viewApi.ViewMaterialReply{}
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	var (
		matRyl   view.MaterialResArr
		musicRly *musicmdl.Entrance
	)
	g := egV2.WithContext(c)
	g.Go(func(ctx context.Context) error {
		//获取material
		req := &view.MaterialParam{AID: arg.Aid, CID: arg.Cid, MobiApp: mobiApp, Build: int32(build), Platform: platform, Device: device}
		matRyl, _ = s.Material(ctx, req)
		return nil
	})
	//获取音乐
	g.Go(func(ctx context.Context) (e error) {
		if musicRly, e = cfg.dep.Music.BgmEntrance(ctx, arg.Aid, arg.Cid, platform); e != nil {
			log.Error("cfg.dep.Music.BgmEntrance(%d,%d) error(%v)", arg.Aid, arg.Cid, e)
			e = nil
		}
		return
	})
	_ = g.Wait()
	//组装
	rly.MaterialRes = make([]*viewApi.MaterialRes, 0)
	var isCrashVer bool
	if mobiApp == "android" && build == 6680100 {
		isCrashVer = true
	}
	for _, v := range matRyl {
		if v == nil {
			continue
		}
		if isCrashVer && v.Typ != 1 && v.Typ != 2 && v.Typ != 3 {
			continue
		}
		rly.MaterialRes = append(rly.MaterialRes, &viewApi.MaterialRes{
			Id:       v.ID,
			Icon:     v.Icon,
			Url:      v.URL,
			Typ:      v.Typ,
			Name:     v.Name,
			BgColor:  v.BgColor,
			BgPic:    v.BgPic,
			JumpType: v.JumpType,
		})
	}
	if musicRly != nil && musicRly.MusicInfo != nil {
		rly.MaterialLeft = &viewApi.MaterialLeft{StaticIcon: view.PlayMusicStaticIcon, Text: musicRly.MusicInfo.MusicTitle, Url: musicRly.MusicInfo.JumpUrl, LeftType: "bgm", Param: musicRly.MusicInfo.MusicId, OperationalType: "7"}
		if !s.c.Custom.CloseMusicIcon {
			rly.MaterialLeft.Icon = view.PlayMusicIcon
		}
	}
	return rly, nil
}

func (s *Service) ViewTag(c context.Context, arg *viewApi.ViewTagReq, mid int64, plat int8, build int, device, pageVersion, buvid, mobiApp, platform string) (*viewApi.ViewTagReply, error) {
	res := &viewApi.ViewTagReply{}
	// 获取稿件信息
	vp, _, err := s.ArcView(c, arg.Aid, 0, "", "", "", plat)
	if err != nil {
		return nil, err
	}
	//获取tag信息
	initRly := s.initTag(c, vp.Arc, arg.Cid, mid, plat, build, pageVersion, buvid, mobiApp, arg.Spmid, platform, nil)
	if initRly == nil {
		return res, nil
	}

	res.SpecialCellNew = initRly.SpecialCellNew
	res.MaterialLeft = initRly.MaterialLeft
	return res, nil
}

// nolint:ineffassign,staticcheck
func (s *Service) ViewGRPC(c context.Context, arg *viewApi.ViewReq, mid int64, plat int8, teenagersMode, lessonsMode, build int, mobiApp, buvid, device, net, platform, ip, cdnip, filterd, isMelloi, brand, slocale, clocale string, now time.Time, disableRcmdMode int) (*viewApi.ViewReply, error) {
	res := new(viewApi.ViewReply)
	vp, extra, err := s.ArcView(c, arg.Aid, mid, mobiApp, device, platform, plat)
	if err != nil {
		return nil, err
	}
	bizExtra := decodeBizExtra(arg.BizExtra)
	if bizExtra.TeenagerExempt == 0 && teenagersMode == 1 && vp.Arc.AttrVal(api.AttrBitTeenager) == api.AttrNo {
		return nil, xecode.AppTeenagersFilter
	}
	// 总开关开 && 商业流量优先
	adTab := !s.c.Custom.DisableAdTab && decodeBizExtra(arg.BizExtra).AdPlayPage == 1
	cfg := s.defaultViewConfigCreater()()
	opts := []ViewOption{
		WithPopupExp(s.buvidABTest(c, buvid, popupFlag)),
		WithAutoSwindowExp(s.buvidABTest(c, buvid, pipVal)),
		WithAdTab(adTab),
		WithSmallWindowExp(s.SmallWindowConfig(c, buvid, smallWindowABtest)),
		WithNewSwindowExp(s.buvidABTest(c, buvid, newSwindowABTestFlage)),
		WithRelatesBiserialExp(s.RelatesBiserialConfig(c, int64(build), mobiApp, device, buvid)),
	}
	cfg.Apply(opts...)
	c = WithContext(c, cfg)

	// 不在白名单内或获取合集配置失败 降级普通播放页
	if s.displayActSeason(vp, teenagersMode, lessonsMode, build, mobiApp, arg.Spmid, mid, plat) {
		if res, err = s.ActivitySeason(c, mid, arg.Aid, plat, build, int(arg.Autoplay), mobiApp, device, buvid, ip, cdnip, net, arg.AdExtra, arg.From, arg.Spmid, arg.FromSpmid, platform, filterd, isMelloi, brand, slocale, clocale, arg.Trackid, arg.PageVersion, now, vp, disableRcmdMode, extra); err == nil {
			s.prom.Incr("展示活动合集")
			HideArcAttribute(res.GetActivitySeason().GetArc())
			return res, nil
		} else {
			log.Error("ActivitySeason sid(%d) aid(%d) s.ActivitySeason err(%+v)", vp.SeasonID, vp.Aid, err)
			if err != appecode.AppActivitySeasonFallback {
				return nil, err
			}
			s.prom.Incr("降级普通合集")
			err = nil
		}
	}

	//获取页码
	pageIndex, err := s.getRelatePageIndex(arg.RelatesPage, arg.Pagination)
	if err != nil {
		return nil, err
	}

	v, err := s.ViewInfo(c, mid, arg.Aid, plat, build, 0, int(arg.Autoplay), teenagersMode, lessonsMode,
		mobiApp, device, buvid, cdnip, net, arg.AdExtra, arg.From, arg.Spmid, arg.FromSpmid, arg.Trackid, platform,
		filterd, "1", true, isMelloi, brand, slocale, clocale, arg.PageVersion, vp,
		disableRcmdMode, arg.DeviceType, pageIndex, arg.SessionId, arg.PlayMode, arg.InFeedPlay, arg.RefreshNum, arg.Refresh, extra)
	if err != nil {
		return nil, err
	}
	v.DislikeReasons(c, s.c.Feature, mobiApp, device, build, disableRcmdMode)
	// 相关推荐曝光上报 后台连播不上报相关推荐数据
	if arg.PlayMode != "background" {
		s.RelateInfoc(mid, arg.Aid, int(plat), strconv.Itoa(build), buvid, ip, model.PathView, v.ReturnCode, v.UserFeature,
			arg.From, "", v.Relates, now, v.IsRec, int(arg.Autoplay), v.PlayParam, arg.Trackid, model.PageTypeRelate,
			arg.FromSpmid, arg.Spmid, v.PvFeature, v.TabInfo, isMelloi, v.RelatesInfoc, pageIndex)
	}
	res = &viewApi.ViewReply{
		Arc:                v.Arc,
		Pages:              view.FromPages(v.Pages),
		OwnerExt:           view.FromOwnerExt(v.OwnerExt),
		ReqUser:            v.ReqUser,
		Tag:                view.FromTag(v.Tag),
		DescTag:            view.FromTag(v.DescTag),
		TIcon:              v.TIcon,
		Season:             view.FromSeason(v.Season),
		ElecRank:           v.ElecRank,
		History:            v.History,
		Relates:            view.FromRelates(v.Relates),
		Dislike:            v.DislikeV2,
		PlayerIcon:         view.FromPlayerIcon(v.PlayerIcon),
		VipActive:          v.VIPActive,
		Bvid:               v.BvID,
		Honor:              v.Honor,
		RelateTab:          v.RelateTab,
		ActivityUrl:        v.ActivityURL,
		Bgm:                v.Bgm,
		Staff:              view.FromStaff(v.Staff),
		ArgueMsg:           v.ArgueMsg,
		ShortLink:          v.ShortLink,
		PlayParam:          int32(v.PlayParam),
		Label:              v.Label,
		UgcSeason:          view.FromUgcSeason(v.UgcSeason),
		Config:             view.FromConfig(v.Config),
		ShareSubtitle:      v.ShareSubtitle,
		Interaction:        v.Interaction,
		Cms:                v.CMSNew,
		CmConfig:           v.CMConfigNew,
		Rank:               v.Rank,
		TfPanelCustomized:  v.TfPanelCustomized,
		UpAct:              v.UpAct,
		UserGarb:           v.UserGarb,
		BadgeUrl:           v.BadgeUrl,
		ReplyPreface:       v.ReplyStyle,
		LiveOrderInfo:      v.LiveOrderInfo,
		DescV2:             v.DescV2,
		Sticker:            v.Sticker,
		CmIpad:             v.IPadCM,
		LikeCustom:         v.LikeCustom,
		UpLikeImg:          v.UpLikeImg,
		SpecialCell:        v.SpecialCell,
		Online:             v.Online,
		CmUnderPlayer:      v.CmUnderPlayer,
		VideoSource:        v.VideoSource,
		Premiere:           v.PremiereResource,
		SpecialCellNew:     v.SpecialCellNew,
		RefreshSpecialCell: v.RefreshSpecialCell,
		MaterialLeft:       v.MaterialLeft,
		NotesCount:         v.NotesCount,
		PullAction:         v.ClientAction,
		Pagination: &pagy.PaginationReply{
			Next: v.Next,
		},
		LikeAnimation: v.LikeAnimation,
		RefreshPage:   v.RefreshPage,
		CoinCustom:    v.CoinCustom,
	}
	s.HandleArcPubLocation(mid, mobiApp, device, arg.FromSpmid, v.Arc, res, false)
	// 首映上报
	s.ReportPremiereWatch(c, isMelloi, v.Arc, buvid)
	return res, nil
}

func (s *Service) GetArcsPlayerGRPC(c context.Context, arg *viewApi.GetArcsPlayerReq, deviceParams device.Device) (*viewApi.GetArcsPlayerReply, error) {
	res := new(viewApi.GetArcsPlayerReply)
	arcsPlayerReq, err := s.SetArcsPlayReq(arg)
	if err != nil {
		return nil, err
	}
	arcsPlayerReply, err := s.GetArcsPlayerURI(c, arg, arcsPlayerReq, deviceParams)
	if err != nil {
		log.Error("s.GetArcsPlayerURI() err: %+v", err)
		return nil, err
	}
	res.ArcsPlayer = arcsPlayerReply
	return res, nil
}

func (s *Service) ContinuousPlayGRPC(c context.Context, arg *viewApi.ContinuousPlayReq, mid int64, plat int8, deviceParams device.Device, net, isMelloi string) (*viewApi.ContinuousPlayReply, error) {
	res := new(viewApi.ContinuousPlayReply)
	//9501-后台播放灰度
	if arg.From == "9501" {
		if !(crc32.ChecksumIEEE([]byte(deviceParams.Buvid+"continuous_9501"))%100 < uint32(s.c.Custom.PlayBackgroundGrey)) {
			res.Relates = nil
			return res, nil
		}
	}
	//ai参数
	aiParams, playInfoParams := s.AiContinuousParam(c, arg, mid, plat, deviceParams, net)
	//ai连播数据
	relate, err := s.ContinuousPlayRelate(c, aiParams, playInfoParams)
	if err != nil {
		return nil, err
	}
	res.Relates = view.FromRelates(relate)
	//数据上报
	if isMelloi == "" {
		s.infoc(playInfoParams)
	}
	return res, nil
}

func (s *Service) GetPremiereGRPC(c context.Context, arg *viewApi.PremiereArchiveReq) (*viewApi.PremiereArchiveReply, error) {
	res := &viewApi.PremiereArchiveReply{
		RiskReason: s.c.Custom.PremiereRiskReason,
	}
	vp, _, err := s.ArcView(c, arg.Aid, 0, "", "", "", 0)
	if err != nil {
		return nil, err
	}
	if vp.Premiere == nil {
		return nil, ecode.NothingFound
	}
	res.Premiere = &viewApi.Premiere{
		PremiereState: viewApi.PremiereState(vp.Premiere.State),
		StartTime:     vp.Premiere.StartTime,
		ServiceTime:   time.Now().Unix(),
		RoomId:        vp.Premiere.RoomId,
	}
	_, err = s.pgcDao.GetUGCPremiereRoomStatus(c, vp.Premiere.RoomId)
	if err != nil {
		if !ecode.EqualError(xecode.PremiereRoomRisk, err) {
			log.Error("s.pgcDao.GetUGCPremiereRoomStatus is err %+v %+v", vp.Premiere.RoomId, err)
		}
	}
	//房间被风控了
	if ecode.EqualError(xecode.PremiereRoomRisk, err) {
		res.RiskStatus = true
	}
	return res, nil
}

func (s *Service) SeasonWidgetExpose(ctx context.Context, arg *viewApi.SeasonWidgetExposeReq) (*viewApi.SeasonWidgetExposeReply, error) {
	req := &checkin.WidgetExposeReq{
		Mid:        arg.Mid,
		Type:       1,
		Oid:        arg.SeasonId,
		ActivityId: arg.ActivityId,
		Aid:        arg.Aid,
		Cid:        arg.Cid,
		Scene:      arg.Scene,
	}
	err := s.checkinDao.CheckinWidgetExpose(ctx, req)
	if err != nil {
		return nil, err
	}
	return &viewApi.SeasonWidgetExposeReply{
		SeasonId:   arg.SeasonId,
		ActivityId: arg.ActivityId,
	}, nil

}

func (s *Service) SeasonActivityRecordRPC(ctx context.Context, arg *viewApi.SeasonActivityRecordReq, dev device.Device, mid int64, net network.Network) (*viewApi.SeasonActivityRecordReply, error) {
	req := &checkin.AddUserActivityRecordReq{
		Mid:        mid,
		Type:       1,
		Oid:        arg.SeasonId,
		ActivityId: arg.ActivityId,
		ActedAt:    xtime.Time(time.Now().Unix()),
		Action:     arg.Action,
		Aid:        arg.Aid,
		Cid:        arg.Cid,
		Scene:      arg.Scene,
	}
	req.Common = &checkin.CommonReq{
		Platform: dev.RawPlatform,
		Build:    int32(dev.Build),
		Buvid:    dev.Buvid,
		MobiApp:  dev.RawMobiApp,
		Device:   dev.Device,
		Ip:       net.RemoteIP,
		Spmid:    arg.Spmid,
	}
	//合集打卡
	reply, err := s.checkinDao.CheckinAddUserActivityRecord(ctx, req)
	if err != nil {
		log.Error("checkinDao.CheckinAddUserActivityRecord %+v %+v %+v", arg, req, err)
		switch {
		case ecode.EqualError(xecode.ActivityIsExpired, err):
			err = ecode.Error(xecode.ActivityIsExpired, "报名时间已过期，无法报名")
		case ecode.EqualError(xecode.ActivityNotExists, err):
			err = ecode.Error(xecode.ActivityNotExists, "活动不存在或已失效")
		case ecode.EqualError(xecode.UserActivityInProgress, err):
			err = ecode.Error(xecode.UserActivityInProgress, "已成功报名，无需重复操作")
		case ecode.EqualError(xecode.UserActivityNotFound, err):
			err = ecode.Error(xecode.UserActivityNotFound, "当前没有进行中的打卡")
		case ecode.EqualError(xecode.UserActivityError, err):
			err = ecode.Error(xecode.UserActivityError, "打卡出错了，请刷新查看")
		default:
			err = ecode.Error(xecode.ActivityNetError, "网络错误，请刷新查看")
		}
		return nil, err
	}
	activity := &viewApi.UgcSeasonActivity{
		Type:             reply.Type,
		Oid:              reply.Oid,
		ActivityId:       reply.ActivityId,
		Title:            reply.Title,
		Intro:            reply.Intro,
		DayCount:         reply.DayCount,
		UserCount:        reply.UserCount,
		JoinDeadline:     reply.JoinDeadline,
		ActivityDeadline: reply.ActivityDeadline,
		CheckinViewTime:  reply.CheckinViewTime,
	}
	if reply.UserActivity != nil {
		activity.UserActivity = &viewApi.UserActivity{
			UserState:       reply.UserActivity.UserState,
			LastCheckinDate: reply.UserActivity.LastCheckinDate,
			CheckinToday:    reply.UserActivity.CheckinToday,
			UserDayCount:    reply.UserActivity.UserDayCount,
			UserViewTime:    reply.UserActivity.UserViewTime,
			Portrait:        reply.UserActivity.Portrait,
		}
	}
	if reply.SeasonShow != nil {
		activity.SeasonShow = &viewApi.SeasonShow{
			ButtonText:    reply.SeasonShow.ButtonText,
			JoinText:      reply.SeasonShow.JoinText,
			RuleText:      reply.SeasonShow.RuleText,
			CheckinText:   reply.SeasonShow.CheckinText,
			CheckinPrompt: reply.SeasonShow.CheckinPrompt,
		}
	}
	return &viewApi.SeasonActivityRecordReply{
		Activity: activity,
	}, nil
}

func (s *Service) ReserveRPC(c context.Context, arg *viewApi.ReserveReq, mid int64) (*viewApi.ReserveReply, error) {
	//取消预约
	if arg.ReserveAction == 1 {
		err := s.actDao.ReserveCancel(c, arg.ReserveId, mid)
		if err != nil {
			log.Error("ReserveRPC_cancel %+v %+v %+v", arg, mid, err)
			return nil, err
		}
	}
	//预约
	if arg.ReserveAction == 0 {
		err := s.actDao.AddReserve(c, arg.ReserveId, mid, arg.UpId)
		if err != nil {
			log.Error("ReserveRPC_reserve %+v %+v %+v", arg, mid, err)
			return nil, err
		}
	}
	return &viewApi.ReserveReply{
		ReserveId: arg.ReserveId,
	}, nil
}
func (s *Service) SetArcsPlayReq(req *viewApi.GetArcsPlayerReq) (out []*api.PlayAv, err error) {
	playAvsTmp := req.PlayAvs
	cmpAidCidMap := make(map[int64]int64) //比较是否有重复的参数
	//最多20个
	if len(req.PlayAvs) > int(20) {
		return nil, ecode.RequestErr
	}
	for _, tmp := range playAvsTmp {
		if tmp == nil {
			continue
		}
		//验证
		if cmpCid, ok := cmpAidCidMap[tmp.Aid]; ok && cmpCid == tmp.Cid {
			continue
		}
		cmpAidCidMap[tmp.GetAid()] = tmp.Cid
		playVideo := []*api.PlayVideo{}
		//拼接
		playVideo = append(playVideo, &api.PlayVideo{Cid: tmp.Cid})
		arcsPlayerReq := &api.PlayAv{
			Aid:        tmp.Aid,
			PlayVideos: playVideo,
		}
		out = append(out, arcsPlayerReq)
	}
	return
}

func (s *Service) AiContinuousParam(c context.Context, arg *viewApi.ContinuousPlayReq, mid int64, plat int8, deviceParams device.Device, net string) (*view.RecommendReq, *view.ContinuousInfo) {
	ip := metadata.String(c, metadata.RemoteIP)
	aiReq := &view.RecommendReq{
		Cmd:        "continuous",
		SourcePage: arg.From,
		Aid:        arg.Aid,
		SessionId:  arg.SessionId,
		DisplayId:  arg.DisplayId,
		TrackId:    arg.Trackid,
		Mid:        mid,
		Buvid:      deviceParams.Buvid,
		Build:      int(deviceParams.Build),
		MobileApp:  deviceParams.RawMobiApp,
		Plat:       plat,
		Network:    net,
		Ip:         ip,
		Spmid:      arg.Spmid,
		FromSpmid:  arg.FromSpmid,
	}
	//zone_id
	loc, _ := s.locDao.Info2(c)
	// 相关推荐AI使用zoneID取zoneID[3]
	if loc != nil && len(loc.ZoneId) >= 4 {
		aiReq.ZoneId = loc.ZoneId[3]
	}
	//上报参数
	now := time.Now().Unix() //时间
	infoParams := &view.ContinuousInfo{
		Ip:          ip,
		Now:         strconv.FormatInt(now, 10),
		Api:         model.PathContinuousPlay,
		Buvid:       deviceParams.Buvid,
		Mid:         strconv.FormatInt(mid, 10),
		Client:      strconv.Itoa(int(plat)),
		MobiApp:     deviceParams.RawMobiApp,
		From:        arg.From,
		Build:       strconv.FormatInt(deviceParams.Build, 10),
		Network:     net,
		FromTrackId: arg.Trackid,
		Spmid:       arg.Spmid,
		FromSpmid:   arg.FromSpmid,
		DisplayId:   strconv.FormatInt(arg.DisplayId, 10),
		FromAv:      strconv.FormatInt(arg.Aid, 10),
	}
	return aiReq, infoParams
}

func (s *Service) buzzwordShowConfigPeriod(ctx context.Context, req *buzzword.BuzzwordShowConfigPeriodReq) []*viewApi.BuzzwordConfig {
	reply, err := s.dmDao.BuzzwordShowConfigPeriod(ctx, req)
	if err != nil {
		log.Error("Failed to get buzzword show config period: %+v: %+v", req, err)
		return nil
	}
	out := make([]*viewApi.BuzzwordConfig, 0, len(reply.Periods))
	for _, v := range reply.Periods {
		out = append(out, &viewApi.BuzzwordConfig{
			Name:          v.Name,
			Schema:        v.Schema,
			Source:        v.Source,
			Start:         v.Start,
			End:           v.End,
			FollowControl: v.FollowControl,
			Id:            v.Id,
			BuzzwordId:    v.BuzzwordId,
			SchemaType:    v.SchemaType,
			Picture:       v.Picture,
		})
	}
	return out
}

//nolint:gocognit
func (s *Service) ViewProgress(c context.Context, arg *viewApi.ViewProgressReq, mid int64, dev device.Device) (*viewApi.ViewProgressReply, error) {
	var (
		dmCommands      []*viewApi.CommandDm
		playerCards     *viewApi.VideoGuide
		chronos         *viewApi.Chronos
		arcShot         *viewApi.VideoShot    //缩略图
		videoPoints     []*viewApi.VideoPoint //分段章节
		pointMaterial   *viewApi.PointMaterial
		pointPermanent  bool //分段章节是否常驻
		buzzwordPeriods []*viewApi.BuzzwordConfig
		eg              = egV2.WithContext(c)
	)
	eg.Go(func(ctx context.Context) (err error) {
		if playerCards, err = s.videoGuidesAttentions(ctx, arg.Aid, arg.Cid, mid, dev); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if dmCommands, err = s.videoDmCommands(ctx, arg.Aid, arg.Cid, mid, dev); err != nil {
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		if s.canUseChronosV2(dev) {
			if arg.ServiceKey == "" { // default value
				arg.ServiceKey = "service_danmaku"
			}
			reply, err := s.ChronosPkg(ctx, &view.ChronosPkgReq{
				ServiceKey:    arg.ServiceKey,
				EngineVersion: arg.EngineVersion,
				Aid:           arg.Aid,
			})
			if err != nil {
				log.Error("ViewProgress s.ChronosPkg error(%+v) aid(%d), mid(%d), buvid(%s), build(%d), mobi_app(%s)", err, arg.Aid, mid, dev.Buvid, dev.Build, dev.RawMobiApp)
				return nil
			}
			chronos = reply
			return nil
		}
		chronos = s.checkChronos(arg.Aid, mid, dev.Build, dev.RawPlatform, dev.Buvid)
		return nil
	})
	//获取缩略图
	eg.Go(func(ctx context.Context) (err error) {
		plat := model.PlatNew(dev.RawMobiApp, dev.Device)
		shot, err := s.arcDao.Shot(ctx, arg.Aid, arg.Cid, plat, dev.Build, dev, dev.Model)
		if err != nil {
			log.Error("s.arcDao.Shot err(%+v)", err)
			return nil
		}
		if shot != nil {
			if dev.RawMobiApp == "android_b" || dev.RawMobiApp == "android" || dev.RawMobiApp == "iphone" {
				for k, img := range shot.Image {
					shot.Image[k] = s.HandleVideoShotV2(img)
				}
			}
			arcShot = &viewApi.VideoShot{
				PvData:   shot.PvData,
				ImgXLen:  shot.XLen,
				ImgYLen:  shot.YLen,
				ImgXSize: shot.XSize,
				ImgYSize: shot.YSize,
				Image:    shot.Image,
			}
		}
		return nil
	})
	//获取"高能看点"
	eg.Go(func(ctx context.Context) (err error) {
		videoViewReply, err := s.vuDao.GetVideoViewPoints(ctx, arg.Aid, arg.Cid)
		if err != nil {
			log.Error("GetVideoViewPoints err(%+v)", err)
			return nil
		}
		if videoViewReply != nil && len(videoViewReply.Points) > 0 {
			for _, p := range videoViewReply.Points {
				tmpPoint := viewApi.VideoPoint{
					Type:    p.Type,
					From:    int64(p.From),
					To:      int64(p.To),
					Content: p.Content,
					Cover:   p.ImgUrl,
					LogoUrl: p.LogoUrl,
				}
				videoPoints = append(videoPoints, &tmpPoint)
			}
			//是否常驻
			pointPermanent = videoViewReply.Permanent
			//获取必剪素材
			if videoViewReply.Oid > 0 {
				req := view.BiJianMaterialReq{Ids: videoViewReply.Oid, Type: _biJianType, Biz: _biJianBiz}
				material, err := s.arcDao.GetBiJianMaterial(ctx, &req)
				if err != nil {
					log.Error("s.arcDao.GetBiJianMaterial err(%+v)", err)
					return nil
				}
				if len(material.Data.List) == 0 {
					return nil
				}
				pointMaterial = &viewApi.PointMaterial{
					Url:            material.Data.List[0].DownloadUrl,
					MaterialSource: viewApi.MaterialSource_BiJian,
				}
			}
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		buzzwordPeriods = s.buzzwordShowConfigPeriod(ctx, &buzzword.BuzzwordShowConfigPeriodReq{
			Oid:  arg.Cid,
			Type: 1,
			Aid:  arg.Aid,
			Mid:  mid,
			Common: &buzzword.Common{
				Build:    dev.Build,
				Buvid:    dev.Buvid,
				MobiApp:  dev.RawMobiApp,
				Platform: dev.RawPlatform,
				Device:   dev.Device,
				Channel:  dev.Channel,
				Brand:    dev.Brand,
				Model:    dev.Model,
				Osver:    dev.Osver,
			},
		})
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("eg.wait() err(%+v) aid(%d) cid(%d) mid(%d)", err, arg.Aid, arg.Cid, mid)
	}
	//暂时判断下，后续产品确定热梗新展示入口，可以删除逻辑，先判断热梗数据，减少对bgm接口的压力
	var isNewVersion bool
	if (dev.RawMobiApp == "android" && dev.Build >= int64(s.c.BuildLimit.MusicAndroidBuild)) || (dev.RawMobiApp == "iphone" && dev.Build >= int64(s.c.BuildLimit.MusicIOSBuild)) {
		isNewVersion = true
	}
	if isNewVersion && len(buzzwordPeriods) > 0 {
		//判断是否有音乐数据
		cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
		musicRly, _ := cfg.dep.Music.BgmEntrance(c, arg.Aid, arg.Cid, dev.RawPlatform)
		if musicRly != nil && musicRly.MusicInfo != nil { //如果有音乐数据，则不展示热梗
			buzzwordPeriods = nil
		}
	}
	//暂时判断下，后续产品确定热梗新展示入口
	rly := &viewApi.ViewProgressReply{
		VideoGuide: &viewApi.VideoGuide{
			Attention:        playerCards.GetAttention(),
			CommandDms:       dmCommands,
			OperationCard:    playerCards.GetOperationCard(),
			OperationCardNew: playerCards.GetOperationCardNew(),
			ContractCard:     playerCards.GetContractCard(),
			CardsSecond:      playerCards.GetCardsSecond(),
		},
		Chronos:         chronos,
		ArcShot:         arcShot,
		Points:          videoPoints,
		PointMaterial:   pointMaterial,
		PointPermanent:  pointPermanent,
		BuzzwordPeriods: buzzwordPeriods,
	}
	return rly, nil
}

func (s *Service) canUseChronosV2(dev device.Device) bool {
	buildLimit, ok := s.c.Custom.ChronosV2SwitchOnMap[dev.RawMobiApp]
	if !ok {
		return false
	}
	if dev.Build < buildLimit {
		return false
	}
	return true
}

func dmExtra(extraStr string) (*view.DmExtra, error) {
	res := &view.DmExtra{}
	err := json.Unmarshal([]byte(extraStr), &res)
	if err != nil {
		log.Error("extraStr Unmarshal is err %+v", err)
		return nil, err
	}
	return res, nil
}

func (s *Service) videoDmCommands(c context.Context, aid, cid, mid int64, dev device.Device) (commands []*viewApi.CommandDm, err error) {
	res, e := s.dmDao.Commands(c, aid, cid, mid, dev)
	if e != nil {
		log.Error("s.dmDao.Commands(%d,%d,%d) error(%+v)", aid, cid, mid, e)
		return nil, e
	}
	if len(res) <= 0 {
		return
	}
	version := feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DmCommandBuild, nil)
	for _, r := range res {
		//type=3的类型老版本不返回
		extraStr := r.GetExtra()
		if r.GetType() == 10 && extraStr != "" {
			extra, err := dmExtra(extraStr)
			if err != nil {
				log.Error("dmExtra is err %+v", err)
				continue
			}
			if extra.ReserveType == 3 && !version {
				continue
			}
		}
		commands = append(commands, &viewApi.CommandDm{
			Id:       r.GetId(),
			Oid:      r.GetOid(),
			Mid:      r.GetMid(),
			Command:  r.GetCommand(),
			Content:  r.GetContent(),
			Progress: r.GetProgress(),
			Ctime:    r.GetCtime(),
			Mtime:    r.GetMtime(),
			Extra:    r.GetExtra(),
			IdStr:    r.GetIdStr(),
		})
	}
	return
}

func (s *Service) videoGuidesAttentions(c context.Context, aid, cid, mid int64, dev device.Device) (*viewApi.VideoGuide, error) {
	res, err := s.confDao.VideoGuide(c, aid, cid, mid, dev)
	if err != nil {
		log.Error("s.confDao.VideoGuide(%d,%d,%d) error(%v)", aid, cid, mid, err)
		return nil, err
	}
	playerCards := &viewApi.VideoGuide{}
	for _, v := range res.GetAttentions() {
		if v == nil {
			continue
		}
		playerCards.Attention = append(playerCards.Attention, &viewApi.Attention{
			StartTime: v.From,
			EndTime:   v.To,
			PosX:      v.PosX,
			PosY:      v.PosY,
		})
	}
	for _, v := range res.GetSkips() {
		if v == nil {
			continue
		}
		playerCards.OperationCard = append(playerCards.OperationCard, &viewApi.OperationCard{
			StartTime:  v.From,
			EndTime:    v.To,
			Icon:       v.Icon,
			Title:      v.Label,
			ButtonText: v.Button,
			Url:        v.Native,
			Content:    v.Content,
		})
	}
	for _, v := range res.GetOperations() {
		if v == nil {
			continue
		}
		cardType := viewApi.OperationCardType(v.CardType)
		bizType := viewApi.BizType(v.BizType)
		//todo 后续如果增加类型和社区沟通向后顺延一位
		if v.GetBizType() == appConf.BizTypeReserveGame { //游戏预约，社区为4，view为5
			bizType = viewApi.BizType_BizTypeReserveGame
		}
		tmpCard := &viewApi.OperationCardNew{
			Id:       v.Id,
			From:     v.From,
			To:       v.To,
			Status:   v.Status,
			CardType: cardType,
			BizType:  bizType,
		}
		switch v.GetParam().(type) {
		case *appConf.OperationCard_Follow:
			if v.GetFollow() == nil {
				continue
			}
			tmpCard.Param = &viewApi.OperationCardNew_Follow{
				Follow: &viewApi.BizFollowVideoParam{
					SeasonId: v.GetFollow().SeasonID,
				},
			}
		case *appConf.OperationCard_Jump:
			if v.GetJump() == nil {
				continue
			}
			tmpCard.Param = &viewApi.OperationCardNew_Jump{
				Jump: &viewApi.BizJumpLinkParam{
					Url: v.GetJump().Url,
				},
			}
		case *appConf.OperationCard_Reserve:
			if v.GetReserve() == nil {
				continue
			}
			tmpCard.Param = &viewApi.OperationCardNew_Reserve{
				Reserve: &viewApi.BizReserveActivityParam{
					ActivityId: v.GetReserve().ActivityID,
				},
			}
		case *appConf.OperationCard_Game:
			if v.GetGame() == nil {
				continue
			}
			tmpCard.Param = &viewApi.OperationCardNew_Game{
				Game: &viewApi.BizReserveGameParam{
					GameId: v.GetGame().GameID,
				},
			}
		default:
			log.Error("unknown OperationCardNew param(%v)", v.Param)
			continue
		}
		switch cardType {
		case viewApi.OperationCardType_CardTypeStandard:
			sd := v.GetStandard()
			if sd == nil {
				continue
			}
			tmpCard.Render = &viewApi.OperationCardNew_Standard{
				Standard: &viewApi.StandardCard{
					Title:               sd.Title,
					ButtonTitle:         sd.ButtonTitle,
					ButtonSelectedTitle: sd.ButtonSelectedTitle,
					ShowSelected:        sd.ShowSelected,
				},
			}
		case viewApi.OperationCardType_CardTypeSkip:
			sk := v.GetSkip()
			if sk == nil {
				continue
			}
			tmpCard.Render = &viewApi.OperationCardNew_Skip{
				Skip: &viewApi.OperationCard{
					Icon:       sk.Icon,
					Title:      sk.Label,
					ButtonText: sk.Button,
					Url:        sk.Native,
					Content:    sk.Content,
				},
			}
		default:
			log.Warn("unknown operation card type(%d)", v.CardType)
			continue
		}
		playerCards.OperationCardNew = append(playerCards.OperationCardNew, tmpCard)
	}
	for _, v := range res.GetCardsV2() {
		if v == nil {
			continue
		}
		bizType := viewApi.BizType(v.BizType)
		if v.GetBizType() == appConf.BizTypeReserveGame { //游戏预约，社区为4，view为5
			bizType = viewApi.BizType_BizTypeReserveGame
		}
		tmpCardV2 := &viewApi.OperationCardV2{
			Id:      v.Id,
			From:    v.From,
			To:      v.To,
			Status:  v.Status,
			BizType: bizType,
			Content: &viewApi.OperationCardV2Content{
				Title:               v.GetContent().GetTitle(),
				Subtitle:            v.GetContent().GetSubtitle(),
				Icon:                v.GetContent().GetIcon(),
				ButtonTitle:         v.GetContent().GetButtonTitle(),
				ButtonSelectedTitle: v.GetContent().GetButtonSelectedTitle(),
				ShowSelected:        v.GetContent().GetShowSelected(),
			},
		}
		switch v.GetParam().(type) {
		case *appConf.OperationCardV2_Follow:
			if v.GetFollow() == nil {
				continue
			}
			tmpCardV2.Param = &viewApi.OperationCardV2_Follow{
				Follow: &viewApi.BizFollowVideoParam{
					SeasonId: v.GetFollow().GetSeasonID(),
				},
			}
		case *appConf.OperationCardV2_Jump:
			if v.GetJump() == nil {
				continue
			}
			tmpCardV2.Param = &viewApi.OperationCardV2_Jump{
				Jump: &viewApi.BizJumpLinkParam{
					Url: v.GetJump().GetUrl(),
				},
			}
		case *appConf.OperationCardV2_Reserve:
			if v.GetReserve() == nil {
				continue
			}
			tmpCardV2.Param = &viewApi.OperationCardV2_Reserve{
				Reserve: &viewApi.BizReserveActivityParam{
					ActivityId: v.GetReserve().GetActivityID(),
				},
			}
		case *appConf.OperationCardV2_Game:
			if v.GetGame() == nil {
				continue
			}
			tmpCardV2.Param = &viewApi.OperationCardV2_Game{
				Game: &viewApi.BizReserveGameParam{
					GameId: v.GetGame().GetGameID(),
				},
			}
		default:
			log.Error("unknown OperationCardV2 param(%v)", v.Param)
			continue
		}
		playerCards.CardsSecond = append(playerCards.CardsSecond, tmpCardV2)
	}
	//契约卡
	if res.Contract != nil {
		upperInfo := &viewApi.UpperInfos{}
		if res.Contract.Upper != nil {
			upperInfo = &viewApi.UpperInfos{
				FansCount:            res.Contract.Upper.FansCount,
				ArcCountLastHalfYear: res.Contract.Upper.ArcCountLastHalfYear,
				FirstUpDates:         res.Contract.Upper.FirstUpDates,
				TotalPlayCount:       res.Contract.Upper.TotalPlayCount,
			}
		}
		playerCards.ContractCard = &viewApi.ContractCard{
			DisplayProgress:          res.Contract.DisplayProgress,
			DisplayAccuracy:          res.Contract.DisplayAccuracy,
			DisplayDuration:          res.Contract.DisplayDuration,
			ShowMode:                 res.Contract.ShowMode,
			PageType:                 res.Contract.PageType,
			Upper:                    upperInfo,
			IsFollowDisplay:          res.Contract.IsFollowDisplay,
			FollowDisplayEndDuration: res.Contract.FollowDisplayEndDuration,
			IsPlayDisplay:            res.Contract.IsPlayDisplay,
			IsInteractDisplay:        res.Contract.IsInteractDisplay,
			PlayDisplaySwitch:        s.c.Custom.ContractPlayDisplaySwitch,
			Text: &viewApi.ContractText{
				Title:       _contractTitle,
				Subtitle:    _contractSubtitle,
				InlineTitle: _contractInlineTitle,
			},
		}
	}

	// 特殊处理: 兼容6.43版本ios问题, 三个版本后可删除
	extendDefaultOperationCardOnIOS643(model.PlatNew(dev.RawMobiApp, dev.Device), dev.Build, playerCards)

	return playerCards, nil
}

//nolint:gomnd
func (s *Service) FilterEmoji(content string) string {
	newContent := ""
	for _, value := range content {
		_, size := utf8.DecodeRuneInString(string(value))
		if size <= 3 {
			newContent += string(value)
		}
	}
	return newContent
}

func (s *Service) checkChronos(aid, mid, build int64, platform, buvid string) *viewApi.Chronos {
	conf := s.chronosConf
	for _, v := range conf {
		if v == nil {
			continue
		}
		// avid check
		if !v.AllAvids && !view.IsInIDs(v.Avids, aid) {
			continue
		}
		// mid check
		if !v.AllMids && !view.IsInIDs(v.Mids, mid) {
			continue
		}
		// build check
		bis, ok := v.BuildLimit[platform]
		if !ok || len(bis) == 0 {
			continue
		}
		if ok := func() bool {
			for _, b := range bis {
				// 有配置限制才下发皮肤
				if view.InvalidBuild(build, b.Value, b.Condition) {
					// 有一个版本校验不通过时，则认为不满足条件
					return false
				}
			}
			return true
		}(); !ok {
			continue
		}
		// gray check
		if crc32.ChecksumIEEE([]byte(buvid))%view.MaxGray < uint32(v.Gray) {
			return &viewApi.Chronos{File: strings.Replace(v.File, "http://", "https://", 1), Md5: v.MD5}
		}
	}
	return nil
}

func (s *Service) ClickPlayerCard(c context.Context, arg *viewApi.ClickPlayerCardReq, mid int64, dev device.Device) error {
	return s.confDao.ClickPlayerCard(c, arg, mid, dev)
}

func (s *Service) ClickPlayerCardV2(c context.Context, arg *viewApi.ClickPlayerCardReq, mid int64, dev device.Device) (string, error) {
	res, err := s.confDao.ClickPlayerCardV2(c, arg, mid, dev)
	if err != nil {
		return "", err
	}
	return res.GetMessage(), nil
}

// nolint:gomnd
func (s *Service) VideoOnline(c context.Context, arg *view.VideoOnlineParam) (*view.VideoOnlineRes, error) {
	if s.c.Online == nil {
		return nil, nil
	}
	if !s.judgeSwitchState(arg.Scene) {
		return nil, nil
	}
	count, canShow := s.psDao.PlayOnlineTotal(c, arg.Aid, arg.Cid)
	if !canShow {
		return nil, nil
	}
	var res = new(view.VideoOnlineRes)
	//点赞场景化-在线人数大于等于后台配置值
	if count >= s.c.Custom.LikeCustomOnlineCount {
		res.LikeSwitch = true
	}
	//在线人数小于10则不展示
	if count < 10 {
		return res, nil
	}
	res.Online.TotalText = fmt.Sprintf(s.c.Online.Text, s.onlineText(count))
	return res, nil
}

// nolint:gomnd
func (s *Service) judgeSwitchState(scene int64) bool {
	switch scene {
	case 1: //ugc竖版全屏
		return s.c.Online.SwitchOnUS
	case 2: //story
		return s.c.Online.SwitchOnStory
	default:
		return s.c.Online.SwitchOn
	}
}

// nolint:gomnd
func (s *Service) onlineText(number int64) string {
	if number < 100 {
		return strconv.FormatInt(number/10*10, 10) + "+"
	}
	if number < 1000 {
		return strconv.FormatInt(number/100*100, 10) + "+"
	}
	if number < 10000 {
		return strconv.FormatInt(number/1000*1000, 10) + "+"
	}
	if number < 100000 {
		return strings.TrimSuffix(strconv.FormatFloat(float64(number)/10000, 'f', 1, 64), ".0") + "万+"
	}
	return "10万+"

}

func (s *Service) VideoDownload(ctx context.Context, arg *view.VideoDownloadReq) (*view.VideoDownloadReply, error) {
	out := &view.VideoDownloadReply{
		ShortFormVideoDownloadReply: &viewApi.ShortFormVideoDownloadReply{
			HasDownloadUrl: false,
		},
	}
	//ios审核态不返回下载分享
	if arg.MobiApp == model.MobileAppIphone && arg.Restriction.GetIsReview() {
		return out, nil
	}
	//版本控制
	if (arg.MobiApp == model.MobileAppIphone && arg.Build < s.c.Custom.VideoDownloadBuildIphone) || (arg.MobiApp == model.MobileAppAndroid && arg.Build <= s.c.Custom.VideoDownloadBuildAndroid) {
		return out, nil
	}
	if !s.matchNGBuilder(arg.Mid, arg.Buvid, "video_download") {
		return out, nil
	}
	archive, err := s.arcDao.Archive(ctx, arg.Aid)
	if err != nil {
		return nil, err
	}
	if !s.isValidArchive(archive) {
		return out, nil
	}
	eg := egV2.WithContext(ctx)
	var shortFormVideoInfo *vcloud.ResponseItem
	eg.Go(func(ctx context.Context) error {
		param := &vcloud.RequestMsg{
			Cids:      []uint64{uint64(arg.Cid)},
			Uip:       metadata.String(ctx, metadata.RemoteIP),
			Platform:  arg.Platform,
			Mid:       uint64(arg.Mid),
			TfType:    vcloud.TFType(arg.TfType),
			BackupNum: 1,
		}
		reply, err := s.vcloudDao.ShortFormVideoInfo(ctx, param, arg.Cid)
		if err != nil {
			log.Error("s.vcloudDao.ShortFormVideoInfo: %+v", err)
			return err
		}
		shortFormVideoInfo = reply
		return nil
	})
	var upSwitch *upApi.UpSwitchReply
	eg.Go(func(ctx context.Context) error {
		reply, err := s.creativeDao.UpDownloadSwitch(ctx, archive.Author.Mid)
		if err != nil {
			return err
		}
		upSwitch = reply
		return nil
	})
	var flowInfo *flowcontrolapi.FlowCtlInfoReply
	eg.Go(func(ctx context.Context) error {
		res, err := s.flowDao.GetCtlInfo(ctx, arg.Aid)
		if err != nil {
			log.Error("s.flowDao.GetCtlInfo failed: %+v", err)
		}
		flowInfo = res
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Failed to request: %+v", err)
		return nil, err
	}
	if !hasDownloadURL(shortFormVideoInfo, upSwitch) {
		return out, nil
	}
	//单一稿件是否支持下载分享 -- 默认支持
	if !IsArcDownloadShare(flowInfo) {
		return out, nil
	}
	setShortFormVideoInfo(out, shortFormVideoInfo)
	return out, nil
}

// 单一稿件是否支持下载分享，默认支持
func IsArcDownloadShare(flowCtl *flowcontrolapi.FlowCtlInfoReply) bool {
	if flowCtl == nil || flowCtl.ForbiddenItems == nil || len(flowCtl.ForbiddenItems) == 0 {
		return true
	}
	//数组转map
	r := funk.Map(flowCtl.ForbiddenItems, func(item *flowcontrolapi.ForbiddenItem) (string, int32) {
		return item.Key, item.Value
	})
	//forbiddenItemsMap
	forbiddenItemsMap := r.(map[string]int32)
	forbidden, ok := forbiddenItemsMap["no_share_download"]
	if !ok {
		return true
	}
	if forbidden == 1 { //1-禁止分享
		return false
	}
	return true
}

func (s *Service) isValidArchive(archive *api.Arc) bool {
	if !archive.IsNormal() {
		return false
	}
	if archive.Stat.View < s.c.Custom.VideoViewNumber { //观看总数
		return false
	}
	//是否是互动视频、是否是PGC、是否是付费、是否仅收藏可见、是否是番剧、是否有地区限制
	if archive.IsSteinsGate() || IsPGCArchive(archive) || IsPayArchive(archive) ||
		IsFavView(archive) || IsBangumi(archive) || IsLimitArea(archive) {
		return false
	}
	return true
}

func hasDownloadURL(info *vcloud.ResponseItem, upSwitch *upApi.UpSwitchReply) bool {
	// 短视频下载开关，0开启，1关闭
	if upSwitch.State != 0 {
		return false
	}
	if info == nil || info.Url == "" {
		return false
	}
	return true
}

func setShortFormVideoInfo(out *view.VideoDownloadReply, info *vcloud.ResponseItem) {
	out.DownloadUrl = info.Url
	out.Md5 = info.FileInfo.Md5
	out.Size_ = info.FileInfo.Filesize
	out.HasDownloadUrl = true
	if len(info.BackupUrl) > 0 {
		out.BackupDownloadUrl = info.BackupUrl[0]
	}
}

// 是否是PGC
func IsPGCArchive(archive *api.Arc) bool {
	return archive.AttrVal(api.AttrBitIsPGC) == api.AttrYes
}

// 是否付费
func IsPayArchive(archive *api.Arc) bool {
	return archive.AttrVal(api.AttrBitIsPUGVPay) == api.AttrYes ||
		archive.AttrVal(api.AttrBitUGCPay) == api.AttrYes ||
		archive.AttrVal(api.AttrBitUGCPayPreview) == api.AttrYes ||
		(archive.AttrValV2(api.AttrBitV2Pay) == api.AttrYes && archive.Rights.ArcPayFreeWatch == api.AttrNo)
}

// 是否仅收藏可见
func IsFavView(archive *api.Arc) bool {
	return archive.AttrValV2(api.AttrBitV2OnlyFavView) == api.AttrYes
}

// 是否是番剧
func IsBangumi(archive *api.Arc) bool {
	return archive.AttrVal(api.AttrBitIsBangumi) == api.AttrYes
}

// 是否地区限制
func IsLimitArea(archive *api.Arc) bool {
	return archive.AttrVal(api.AttrBitLimitArea) == api.AttrYes
}

func IsRecentArchive(archive *api.Arc) bool {
	nowTime := time.Now()
	return nowTime.Sub(time.Unix(int64(archive.Ctime), 0)).Hours() <= 28*24
}

func (s *Service) displayActSeason(vp *api.ViewReply, teenagersMode, lessonsMode, build int, mobiApp, spmid string, mid int64, plat int8) bool {
	// 活动合集
	// 青少年及课堂模式 播单页请求view接口 不支持
	// 粉蓝app 及 ipadHD 支持 + 港澳台版本
	validBuild := false
	if (plat == model.PlatIPhone && build > s.c.ActivitySeason.IphoneBuild) || (plat == model.PlatIPhoneB && build > s.c.ActivitySeason.IphoneBlueBuild) ||
		(plat == model.PlatAndroid && build > s.c.ActivitySeason.AndroidBuild) || (plat == model.PlatAndroidB && build > s.c.ActivitySeason.AndroidBlueBuild) ||
		(plat == model.PlatIpadHD && build > s.c.ActivitySeason.IpadHDBuild || (plat == model.PlatIPad && build >= s.c.ActivitySeason.IpadBuild) ||
			(plat == model.PlatAndroidI && build > s.c.ActivitySeason.AndroidIBuild || plat == model.PlatIPhoneI && build > s.c.ActivitySeason.IPhoneIBuild)) {
		validBuild = true
	}
	if vp.AttrValV2(api.AttrBitV2ActSeason) == api.AttrYes {
		if vp.SeasonID > 0 && teenagersMode == 0 && lessonsMode == 0 && spmid != _playlistSpmid && validBuild {
			return true
		}
		log.Warn("ActivitySeason sid(%d) aid(%d) displayActSeason invalid mid(%d) AttrValV2(%d) teenagersMode(%d) lessonsMode(%d) spmid(%s) validBuild(%t) mobiApp(%s) build(%d) plat(%d)", vp.SeasonID, vp.Aid, mid, vp.AttributeV2, teenagersMode, lessonsMode, spmid, validBuild, mobiApp, build, plat)
	}
	return false
}

func (s *Service) ExposePlayerCard(c context.Context, arg *viewApi.ExposePlayerCardReq, mid int64, dev device.Device) error {
	return s.confDao.ExposePlayerCard(c, arg, mid, dev)
}

func (s *Service) AddContract(c context.Context, arg *viewApi.AddContractReq, mid int64, dev device.Device) error {
	return s.contractDao.AddContract(c, arg, mid, dev)
}

func (s *Service) DmVote(ctx context.Context, req *view.DmVoteReq) (*view.DmVoteReply, error) {
	fanoutResult := s.doVoteFanout(ctx, req)
	if fanoutResult.vote == nil {
		return nil, ecode.Error(ecode.RequestErr, "投票失败")
	}
	reply := &view.DmVoteReply{
		Vote: &view.VoteReply{
			Uid:  fanoutResult.vote.Uid,
			Type: fanoutResult.vote.Type,
		},
	}
	if fanoutResult.userCard.GetLevel() <= 0 {
		return reply, nil
	}
	dmReply, err := s.dmClient.PostByVote(ctx, &dmgrpc.PostByVoteReq{
		Progress:      req.Progress,
		Build:         req.Build,
		TeenagersMode: req.TeenagersMode,
		LessonsMode:   req.LessonsMode,
		Aid:           req.AID,
		Cid:           req.CID,
		Mid:           req.Mid,
		Msg:           strconv.FormatInt(int64(req.Vote), 10),
		Platform:      req.Platform,
		Buvid:         req.Buvid,
		MobiApp:       req.MobiApp,
		Device:        req.Device,
	})
	if err != nil {
		log.Error("Failed to request PostByVote: %+v", err)
		return reply, nil
	}
	reply.Dm = &view.DmReply{
		DmID:    dmReply.GetDmid(),
		DmIDStr: dmReply.GetDmidStr(),
		Visible: dmReply.GetVisible(),
		Action:  dmReply.GetAction(),
	}
	return reply, nil
}

type voteFanoutResult struct {
	userCard *accApi.Card
	vote     *votegrpc.DoVoteRsp
}

func (s *Service) doVoteFanout(ctx context.Context, args *view.DmVoteReq) *voteFanoutResult {
	result := &voteFanoutResult{}
	eg := egV2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if result.userCard, err = s.accDao.Card3(ctx, args.Mid); err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if result.vote, err = s.dynamicDao.Vote(ctx, &votegrpc.DoVoteReq{
			VoteId:   args.VoteID,
			Votes:    []int32{args.Vote},
			VoterUid: args.Mid,
		}); err != nil {
			log.Error("%+v", err)
			err = nil
		}
		return
	})
	if err := eg.Wait(); err != nil {
		log.Error("Failed to execute errgroup: %+v", err)
	}
	return result
}

func (s *Service) Season(c context.Context, req *viewApi.SeasonReq) (*viewApi.SeasonReply, error) {
	ugcSn, err := s.seasonDao.Season(c, req.SeasonId)
	if err != nil {
		log.Error("s.seasonDao.Season req(%+v) err(%+v)", req, err)
		return nil, err
	}
	season := new(view.UgcSeason)
	season.FromSeason(ugcSn)
	return &viewApi.SeasonReply{Season: view.FromUgcSeason(season)}, nil
}

func (s *Service) GetStatSrv(c context.Context, req *view.StatReq) (*view.StatReply, error) {
	archive, err := s.arcDao.Archive(c, req.Aid)
	if err != nil {
		return nil, err
	}
	res := &view.StatReply{}
	res.Stat.Like = int64(archive.Stat.Like)
	return res, nil
}

func buggyIOS643(plat int8, build int64) bool {
	platMatch := plat == device.PlatIPad || plat == device.PlatIPhone
	buildMatch := build == int64(64300100) || build == int64(64301100)
	return platMatch && buildMatch
}

func extendDefaultOperationCardOnIOS643(plat int8, build int64, playerCards *viewApi.VideoGuide) {
	if buggyIOS643(plat, build) && len(playerCards.OperationCardNew) == 0 {
		tmpCard := &viewApi.OperationCardNew{
			Id:       0,
			From:     0,
			To:       0,
			Status:   false,
			CardType: 99,
			BizType:  0,
		}
		playerCards.OperationCardNew = append(playerCards.OperationCardNew, tmpCard)
	}
}

func decodeBizExtra(in string) view.BizExtra {
	if in == "" {
		return view.BizExtra{}
	}
	decodedValue, err := url.QueryUnescape(in)
	if err != nil {
		log.Error("Failed to decode bizExtra: %+v", err)
		return view.BizExtra{}
	}
	res := view.BizExtra{}
	err = json.Unmarshal([]byte(decodedValue), &res)
	if err != nil {
		log.Error("Failed to Unmarshal bizExtra: %+v", err)
		return view.BizExtra{}
	}
	return res
}

func handleVideoShot(img string) string {
	u, err := url.Parse(img)
	if err != nil {
		return img
	}
	if strings.Contains(u.Host, "boss") {
		return img
	}
	return img + "@50q.webp"
}

func (s *Service) HandleVideoShotV2(img string) string {
	u, err := url.Parse(img)
	if err != nil {
		return img
	}
	if strings.Contains(u.Host, "boss") {
		u.Host = s.c.Custom.VideoShotHost
		u.Path = u.Path + "@.webp"
		return u.Scheme + "://" + u.Host + u.Path
	}
	return img + "@50q.webp"
}

func (s *Service) CacheViewGRPC(c context.Context, arg *viewApi.CacheViewReq, mid int64, plat int8, teenagersMode,
	lessonsMode, build int, mobiApp, buvid, device, net, platform, cdnip, filterd, brand, slocale,
	clocale string, disableRcmdMode int) (*viewApi.CacheViewReply, error) {

	cfg := s.defaultViewConfigCreater()()
	opts := []ViewOption{
		SkipRelate(true),
		SkipSpecialCell(true),
		WithPopupExp(false),
		WithAutoSwindowExp(false),
		WithSmallWindowExp(false),
		WithAdTab(false),
	}
	cfg.Apply(opts...)
	c = WithContext(c, cfg)

	//获取稿件信息
	vp, extra, err := s.ArcView(c, arg.Aid, 0, "", "", "", plat)
	if err != nil {
		log.Error("s.CacheViewGRPC ArcView error(%+v)", err)
		return nil, err
	}

	v, err := s.CacheViewInfo(c, mid, arg.Aid, plat, build, teenagersMode, lessonsMode, mobiApp, device, buvid,
		cdnip, net, platform, filterd, brand, slocale, clocale, "v1", arg.Spmid, vp, extra)
	if err != nil {
		log.Error("s.CacheViewGRPC CacheViewInfo error(%+v)", err)
		return nil, err
	}

	//获取"我不想看"信息
	v.DislikeReasons(c, s.c.Feature, mobiApp, device, build, disableRcmdMode)

	res := &viewApi.CacheViewReply{
		Arc:               v.Arc,
		Pages:             view.FromPages(v.Pages),
		OwnerExt:          view.FromOwnerExt(v.OwnerExt),
		ReqUser:           v.ReqUser,
		Season:            view.FromSeason(v.Season),
		ElecRank:          v.ElecRank,
		History:           v.History,
		Dislike:           v.DislikeV2,
		PlayerIcon:        view.FromPlayerIcon(v.PlayerIcon),
		Bvid:              v.BvID,
		ShortLink:         v.ShortLink,
		ShareSubtitle:     v.ShareSubtitle,
		TfPanelCustomized: v.TfPanelCustomized,
		Online:            v.Online,
	}
	return res, nil
}

// nolint:gocognit
func (s *Service) CacheViewInfo(c context.Context, mid, aid int64, plat int8, build, teenagersMode, lessonsMode int,
	mobiApp, device, buvid, cdnIP, network, platform, withoutCharge string, brand, slocale, clocale, pageVersion, spmid string,
	vp *api.ViewReply, extra map[string]string) (*view.View, error) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())

	v, err := s.ViewPage(c, mid, plat, build, mobiApp, device, cdnIP, true, buvid, slocale, clocale, vp, pageVersion, spmid, platform, teenagersMode, extra)
	if err != nil {
		log.Error("s.CacheViewInfo ViewPage error(%+v)", err)
		return nil, err
	}

	if v == nil {
		return nil, ecode.NothingFound
	}
	defer HideArcAttribute(v.Arc)

	var tagIDs []int64
	for _, tagData := range v.Tag {
		tagIDs = append(tagIDs, tagData.TagID)
	}
	//新版本去掉活动tag
	v.Tag = s.NewTopicDelActTag(c, v.Tag, buvid)

	// config
	g := egV2.WithContext(c)

	//获取用户和up关系
	g.Go(func(ctx context.Context) (err error) {
		s.initReqUser(ctx, v, mid, plat, build, buvid, platform, brand, network, mobiApp)
		return
	})
	if v.AttrVal(api.AttrBitIsPGC) != api.AttrYes {
		g.Go(func(ctx context.Context) (err error) {
			// 从6.10版本开始去除对dm.SubjectInfos调用
			if (plat == model.PlatIPhone && build >= s.c.BuildLimit.DmInfoIOSBuild) || (plat == model.PlatAndroid && build >= s.c.BuildLimit.DmInfoAndBuild) {
				return
			}
			s.initDM(ctx, v)
			return
		})
		g.Go(func(ctx context.Context) (err error) {
			s.initAudios(ctx, v)
			return
		})
	}
	//获取充电排行
	if v.AttrVal(api.AttrBitIsPGC) != api.AttrYes && teenagersMode == 0 && lessonsMode == 0 {
		g.Go(func(ctx context.Context) (err error) {
			if model.IsIPhoneB(plat) || (model.IsIPhone(plat) && (build >= 7000 && build <= 8000)) {
				return
			}
			if withoutCharge == "1" {
				s.initElecRank(ctx, v, mobiApp, platform, device, build)
				return
			}
			s.initElec(ctx, v, mobiApp, platform, device, build, mid)
			return
		})
	}
	//获取播放进度条装扮
	g.Go(func(ctx context.Context) error {
		// 版本控制
		var showPlayicon bool
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.ViewPlayIcon, &feature.OriginResutl{
			MobiApp:    mobiApp,
			Device:     device,
			Build:      int64(build),
			BuildLimit: (mobiApp == "iphone" && build >= conf.Conf.BuildLimit.PlayIconIOSBuildLimit) || (mobiApp == "android" && build >= conf.Conf.BuildLimit.PlayIconAndroidBuildLimit),
		}) {
			showPlayicon = true
		}
		if v.PlayerIcon, err = cfg.dep.Resource.PlayerIcon(ctx, v.Aid, mid, tagIDs, v.TypeID, showPlayicon, build, mobiApp, device); err != nil {
			log.Error("PlayerIcon err(%+v) aid(%d) tagids(%+v) typeid(%d)", err, v.Aid, tagIDs, v.TypeID)
		}
		return nil
	})
	//获取免流信息
	if s.matchNGBuilder(mid, buvid, "tf_panel") {
		g.Go(func(ctx context.Context) (err error) {
			customizedPanel, err := cfg.dep.Resource.GetPlayerCustomizedPanel(ctx, tagIDs)
			if err != nil && !ecode.EqualError(ecode.NothingFound, err) {
				log.Error("Failed to get player customized panel with tids: %+v: %+v", tagIDs, err)
				return nil
			}
			v.TfPanelCustomized = view.FromPlayerCustomizedPanel(customizedPanel)
			return nil
		})
	}
	//是否进行跳转
	g.Go(func(ctx context.Context) error {
		//获取archive_redirect数据
		redirect, err := cfg.dep.Archive.ArcRedirectUrl(ctx, aid)
		if err != nil {
			return nil
		}
		if redirect.RedirectTarget == "" || redirect.PolicyId == 0 {
			return nil
		}
		//location策略获取返回数据
		if redirect.GetPolicyType() == api.RedirectPolicyType_PolicyTypeLocation {
			locs, err := cfg.dep.Location.GetGroups(ctx, []int64{redirect.PolicyId})
			if err != nil {
				log.Error("GetGroups is err (%+v)", err)
				return nil
			}
			loc, ok := locs[redirect.PolicyId]
			if !ok {
				return nil
			}
			//是否需要跳转
			if loc.Play != int64(location.Status_Forbidden) {
				v.Season = &bangumi.Season{
					IsJump:     1,
					OGVPlayURL: redirect.RedirectTarget,
					SeasonID:   "1", //为了兼容android的逻辑，写死一个不存在的season_id+title
					Title:      "forcejump",
				}
			}
		}
		return nil
	})
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}

	//获取分享副标题
	v.SubTitleChange()
	//获取short_link
	s.setShortLink(v)

	return v, nil
}

func (s *Service) setShortLink(v *view.View) {
	v.ShortLink = fmt.Sprintf(_shortLinkHost+"/av%d", v.Aid)
	bvID, err := bvid.AvToBv(v.Aid)
	if err != nil {
		log.Error("bvid.AvToBv aid(%d) err(%+v)", v.Aid, err)
	} else {
		v.ShortLink = fmt.Sprintf(_shortLinkHost+"/%s", bvID)
	}
}

func (s *Service) ChronosPkg(ctx context.Context, req *view.ChronosPkgReq) (*viewApi.Chronos, error) {
	authN, _ := auth.FromContext(ctx)
	device, ok := device.FromContext(ctx)
	if !ok {
		return nil, errors.New("failed to find device info")
	}
	network, _ := network.FromContext(ctx)
	fks, ok := fksmeta.FromContext(ctx)
	if !ok {
		return nil, errors.New("failed to find fawkes info")
	}
	ruleMeta := &view.RuleMeta{
		AppKey:        fks.AppKey,
		ServiceKey:    req.ServiceKey,
		Mid:           authN.Mid,
		RomVersion:    device.Osver,
		NetType:       int64(network.Type),
		MobiApp:       device.RawMobiApp,
		EngineVersion: req.EngineVersion,
		Buvid:         device.Buvid,
		Build:         device.Build,
		DeviceType:    device.Brand,
		Aid:           req.Aid,
		Device:        device.Device,
	}
	if pd.WithContext(ctx).IsIOSAll().FinishOr(false) { //ios的品牌使用model
		ruleMeta.DeviceType = device.Model
	}
	chronosPkgList, ok := s.chronosPkgInfo[genChronosPkgKey(ruleMeta.AppKey, ruleMeta.ServiceKey)]
	if !ok {
		return nil, errors.WithMessagef(ecode.NothingFound, "Failed to find package with appkey(%s) and servicekey(%s)", ruleMeta.AppKey, ruleMeta.ServiceKey)
	}
	chronosPkgListInOrder := reOrderChronosPkgListByRank(chronosPkgList)
	for _, v := range chronosPkgListInOrder {
		if !judgePackageRules(v, ruleMeta) {
			continue
		}
		return &viewApi.Chronos{
			File: v.ResourceUrl,
			Md5:  v.Md5,
			Sign: v.Sign,
		}, nil
	}
	return nil, errors.WithMessagef(ecode.NothingFound, "Failed to match rules appkey(%s) servicekey(%s)", ruleMeta.AppKey, ruleMeta.ServiceKey)
}

func genChronosPkgKey(appKey, serviceKey string) string {
	return fmt.Sprintf("%s:%s", appKey, serviceKey)
}

func splitMessageByDotAndConvertToInt(in string) map[int64]struct{} {
	messages := strings.Split(in, ",")
	out := make(map[int64]struct{}, len(messages))
	for _, v := range messages {
		vi, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		out[vi] = struct{}{}
	}
	return out
}

func splitMessageByDot(in string) map[string]struct{} {
	messages := strings.Split(in, ",")
	out := make(map[string]struct{}, len(messages))
	for _, v := range messages {
		out[v] = struct{}{}
	}
	return out
}

func reOrderChronosPkgListByRank(in []*view.PackageInfo) []*view.PackageInfo {
	sort.Slice(in, func(i, j int) bool {
		return in[i].Rank > in[j].Rank
	})
	return in
}

func judgePackageRules(info *view.PackageInfo, rule *view.RuleMeta) bool {
	if info.WhiteList != "" {
		if _, ok := splitMessageByDotAndConvertToInt(info.WhiteList)[rule.Mid]; ok {
			return true
		}
	}
	if info.BlackList != "" {
		if _, ok := splitMessageByDotAndConvertToInt(info.BlackList)[rule.Mid]; ok {
			return false
		}
	}
	if info.VideoList != "" {
		if _, ok := splitMessageByDotAndConvertToInt(info.VideoList)[rule.Aid]; !ok {
			return false
		}
	}
	if info.BuildLimitExp != "" {
		if !pd.WithDevice(pd.NewCommonDevice(rule.MobiApp, rule.Device, "", rule.Build)).ParseCondition(info.BuildLimitExp).FinishOr(false) {
			return false
		}
	}
	if info.RomVersion != "" {
		if _, ok := splitMessageByDot(info.RomVersion)[rule.RomVersion]; !ok {
			return false
		}
	}
	if info.NetType != "" {
		if _, ok := splitMessageByDotAndConvertToInt(info.NetType)[rule.NetType]; !ok {
			return false
		}
	}
	if info.EngineVersion != "" {
		if _, ok := splitMessageByDot(info.EngineVersion)[rule.EngineVersion]; !ok {
			return false
		}
	}
	if info.DeviceType != "" {
		deviceType := strings.ToLower(info.DeviceType)
		if _, ok := splitMessageByDot(deviceType)[strings.ToLower(rule.DeviceType)]; !ok {
			return false
		}
	}
	return crc32.ChecksumIEEE([]byte(rule.Buvid))%10000 < uint32(info.Gray)
}

// LikeGrayControl 白名单和灰度控制
func (s *Service) LikeGrayControl(aid int64) bool {
	// 白名单
	_, ok := s.c.LikeNumGrayControl.Aid[strconv.FormatInt(aid, 10)]
	// 灰度控制
	group := aid % _likeGray
	return ok || group < s.c.LikeNumGrayControl.Gray
}

func (s *Service) HandleArcPubLocation(mid int64, mobiApp, device string, fromSpmid string, arc *api.Arc, res *viewApi.ViewReply, isActivitySeason bool) {
	var pubLocation string
	if isActivitySeason {
		pubLocation = res.ActivitySeason.Arc.PubLocation
		//隐藏arc中的IP属地
		res.ActivitySeason.Arc.PubLocation = ""
	} else {
		pubLocation = res.Arc.PubLocation
		//隐藏arc中的IP属地
		res.Arc.PubLocation = ""
	}

	versionMatch := mobiApp == "iphone" || mobiApp == "android" || (mobiApp == "ipad" && device == "pad") ||
		mobiApp == "android_hd"
	if !versionMatch {
		return
	}

	if !s.AppFeatureGate.UserIPDisplay().Enabled() {
		return
	}

	if s.AppFeatureGate.UserIPDisplay().SelfVisibleOnly() && mid != arc.Author.Mid {
		return
	}

	if _, ok := s.c.Custom.ShowArcPubIpFromSpmid[fromSpmid]; !ok {
		return
	}

	parse, _ := time.Parse(_arcPubTimeForm, s.c.Custom.ShowArcPubIpAfterTime)
	if arc.PubDate.Time().Before(parse) {
		return
	}

	if pubLocation == "" {
		return
	}

	if isActivitySeason {
		res.ActivitySeason.ArcExtra = &viewApi.ArcExtra{
			ArcPubLocation: "IP属地：" + pubLocation,
		}
	} else {
		res.ArcExtra = &viewApi.ArcExtra{
			ArcPubLocation: "IP属地：" + pubLocation,
		}
	}
}

func (s *Service) setElecCharging(mobiApp, device string, v *view.View) {
	versionMatch := (mobiApp == "iphone" && device != "pad") || mobiApp == "android"
	if !versionMatch {
		return
	}
	if v.ReqUser != nil && v.Elec != nil && len(v.Elec.UpowerTitle) > 0 && len(v.Elec.UpowerJumpUrl) > 0 {
		v.ReqUser.ElecPlusBtn = &viewApi.Button{
			Icon:  v.Elec.UpowerIconUrl,
			Title: v.Elec.UpowerTitle,
			Uri:   v.Elec.UpowerJumpUrl,
		}
	}
}
