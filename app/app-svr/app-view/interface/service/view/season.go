package view

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go-common/component/metadata/device"
	dev "go-common/component/metadata/device"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"

	appecode "go-gateway/app/app-svr/app-card/ecode"
	"go-gateway/app/app-svr/app-card/interface/model/i18n"
	viewApi "go-gateway/app/app-svr/app-view/interface/api/view"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/view"
	"go-gateway/pkg/idsafe/bvid"

	"go-gateway/app/app-svr/archive/service/api"

	accApi "git.bilibili.co/bapis/bapis-go/account/service"
	dmApi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	location "git.bilibili.co/bapis/bapis-go/community/service/location"
	vuApi "git.bilibili.co/bapis/bapis-go/videoup/open/service"
	"git.bilibili.co/go-tool/libbdevice/pkg/pd"
)

// nolint:gocognit
func (s *Service) ActivitySeason(c context.Context, mid, aid int64, plat int8, build, autoplay int, mobiApp, device, buvid, ip, cdnIP, network, adExtra, from, spmid, fromSpmid, platform, filtered, isMelloi, brand, slocale, clocale, trackid, pageVersion string, now time.Time, vp *api.ViewReply, disableRcmdMode int, extra map[string]string) (*viewApi.ViewReply, error) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	v, err := s.SeasonView(c, mid, plat, build, mobiApp, device, cdnIP, buvid, vp, pageVersion, spmid, platform, extra)
	if err != nil {
		log.Error("ActivitySeason sid(%d) aid(%d) SeasonView err(%+v)", vp.SeasonID, aid, err)
		return nil, err
	}
	// config
	if v == nil {
		return nil, ecode.NothingFound
	}
	var (
		tagIDs                                              []int64
		isFavSeason, isReserve, isPlayStory, landscapeStory bool
		reserveId                                           int64
		storyIcon, landscapeIcon                            string
	)
	for _, tag := range v.Tag {
		tagIDs = append(tagIDs, tag.TagID)
	}
	eg := errgroup.WithContext(c)
	//获取竖屏视频切全屏是否进story
	eg.Go(func(ctx context.Context) error {
		isPlayStory, storyIcon, landscapeStory, landscapeIcon = s.playStoryABTest(ctx, mid, buvid)
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		s.initReqUser(ctx, v, mid, plat, build, buvid, platform, brand, network, mobiApp)
		//关注动效等级大型活动页不同步
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		s.initHonor(ctx, v, plat, build, mobiApp, device)
		// 保持线上逻辑，无荣誉榜单 && 3 < 稿件排行 <= 10 返回文字排行
		if v.Honor == nil && v.Stat.HisRank > s.c.Custom.HonorRank && v.Stat.HisRank <= s.c.Custom.HonorRankMax {
			v.Rank = &viewApi.Rank{
				Icon:      model.ActSeasonRankIcon,
				IconNight: model.ActSeasonRankIcon,
				Text:      fmt.Sprintf("全站排行榜最高第%d名", v.Stat.HisRank),
			}
		}
		return nil
	})
	if !cfg.skipRelate {
		eg.Go(func(ctx context.Context) (err error) {
			s.initActivityCM(ctx, v, plat, build, mid, buvid, mobiApp, device, network, adExtra, from, spmid, fromSpmid, platform, filtered, tagIDs, isMelloi)
			return nil
		})
		if v.MngAct.ActivePlay.IsContainedRecom { //如果包含AI的相关推荐
			eg.Go(func(ctx context.Context) (err error) {
				s.initAIRelate(ctx, v, plat, build, autoplay, mid, buvid, mobiApp, device, from, spmid, fromSpmid, trackid, filtered, slocale, clocale, isMelloi, ip, now, pageVersion)
				return nil
			})
		}
	}
	eg.Go(func(ctx context.Context) (err error) {
		s.initElecRank(ctx, v, mobiApp, platform, device, build)
		return nil
	})
	asDesc := ""
	arcDescV2 := []*api.DescV2{
		{
			RawText: v.Desc,
			Type:    api.DescType_DescTypeText,
		},
	}
	accountInfos := &accApi.InfosReply{}
	eg.Go(func(ctx context.Context) (err error) {
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
	if v.AttrVal(api.AttrBitHasArgument) == api.AttrYes {
		eg.Go(func(ctx context.Context) (err error) {
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
	eg.Go(func(ctx context.Context) (err error) {
		if v.PlayerIcon, err = cfg.dep.Resource.PlayerIcon(ctx, v.Aid, mid, tagIDs, v.TypeID, true, build, mobiApp, device); err != nil {
			log.Error("ActivitySeason sid(%d) aid(%d) PlayerIcon err(%+v) tagids(%+v) typeid(%d)", v.SeasonID, v.Aid, err, tagIDs, v.TypeID)
		}
		return nil
	})
	if s.matchNGBuilder(mid, buvid, "tf_panel") {
		eg.Go(func(ctx context.Context) (err error) {
			customizedPanel, err := cfg.dep.Resource.GetPlayerCustomizedPanel(ctx, tagIDs)
			if err != nil {
				log.Error("ActivitySeason sid(%d) aid(%d) GetPlayerCustomizedPanel err(%+v) tids(%+v)", v.SeasonID, v.Aid, err, tagIDs)
				return nil
			}
			v.TfPanelCustomized = view.FromPlayerCustomizedPanel(customizedPanel)
			return nil
		})
	}
	if mid > 0 || buvid != "" {
		eg.Go(func(ctx context.Context) (err error) {
			v.History, _ = cfg.dep.History.Progress(ctx, v.Aid, mid, buvid)
			return nil
		})
	}
	if mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if v.MngAct.ActivePlay.IsContainedLive && !model.IsIPad(plat) && v.MngAct.ActivePlay.Stime > now.Unix() {
				if lb, ok := v.MngAct.GetAppPlay().GetMenus()[model.LiveBefore]; ok {
					isReserve = s.actDao.IsReserveAct(ctx, lb.MenuId, mid)
				}
			}
			return nil
		})
	}
	//获取预约id
	eg.Go(func(ctx context.Context) error {
		d, _ := dev.FromContext(ctx)
		res, err := s.dmDao.Commands(ctx, aid, v.Arc.FirstCid, mid, d)
		if err != nil {
			log.Error("Commands fail aid:%+v cid:%+v mid:%+v err:%+v", aid, v.Arc.FirstCid, mid, err)
			return nil
		}
		reserveId = ActivityReserveId(res)
		return nil
	})
	//评论样式
	eg.Go(func(ctx context.Context) error {
		res, err := cfg.dep.Reply.GetReplyListPreface(ctx, mid, aid, buvid)
		if err != nil {
			log.Error("GetReplyListPreface fail mid:%d, aid:%d err:%+v", mid, aid, err)
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
	if mobiApp == "android" {
		eg.Go(func(ctx context.Context) error {
			if s.popupConfig(ctx, mid, buvid) {
				v.Config = &view.Config{
					PopupInfo: true,
				}
			}
			return nil
		})
	}
	eg.Go(func(ctx context.Context) error {
		s.initLabel(ctx, v, s.displaySteinsLabel(c, v.ViewStatic, mobiApp, device, build))
		return nil
	})
	//获取在看人数信息
	eg.Go(func(ctx context.Context) error {
		s.initOnline(v, buvid, mid, v.Aid)
		return nil
	})

	if err = eg.Wait(); err != nil {
		log.Error("ActivitySeason sid(%d) aid(%d) eg.wait() err(%+v)", v.SeasonID, v.Aid, err)
		return nil, err
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
	v.ShortLink = fmt.Sprintf(_shortLinkHost+"/av%d", v.Aid)
	if v.BvID != "" {
		v.ShortLink = fmt.Sprintf(_shortLinkHost+"/%s", v.BvID)
	}
	v.SubTitleChange()
	v.DislikeReasons(c, s.c.Feature, mobiApp, device, build, disableRcmdMode)
	if asDesc != "" {
		v.Desc = asDesc
	}
	//竖屏进story实验
	if v.Config == nil {
		v.Config = &view.Config{}
	}
	v.Config.PlayStory = isPlayStory
	v.Config.StoryIcon = storyIcon
	v.Config.LandscapeStory = landscapeStory
	v.Config.LandscapeIcon = landscapeIcon
	//竖屏进story实验
	res := &viewApi.ViewReply{
		ActivitySeason: &viewApi.ActivitySeason{
			Arc:               v.Arc,
			Pages:             view.FromPages(v.Pages),
			OwnerExt:          view.FromOwnerExt(v.OwnerExt),
			ReqUser:           v.ReqUser,
			ElecRank:          v.ElecRank,
			History:           v.History,
			Dislike:           v.DislikeV2,
			PlayerIcon:        view.FromPlayerIcon(v.PlayerIcon),
			Bvid:              v.BvID,
			Honor:             v.Honor,
			Staff:             view.FromStaff(v.Staff),
			ArgueMsg:          v.ArgueMsg,
			ShortLink:         v.ShortLink,
			Label:             v.Label,
			UgcSeason:         view.FromUgcSeason(v.UgcSeason),
			ShareSubtitle:     v.ShareSubtitle,
			CmConfig:          v.CMConfigNew,
			Rank:              v.Rank,
			TfPanelCustomized: v.TfPanelCustomized,
			BadgeUrl:          v.BadgeUrl,
			DescV2:            s.DescV2ParamsMerge(c, arcDescV2, accountInfos),
			Config:            view.FromConfig(v.Config),
			Online:            v.Online,
			ReplyPreface:      v.ReplyStyle,
		},
	}
	s.HandleArcPubLocation(mid, mobiApp, device, fromSpmid, v.Arc, res, true)
	if v.ReqUser != nil {
		isFavSeason = v.ReqUser.FavSeason == 1
	}
	if model.IsIPad(plat) {
		s.initIpadMngAct(res, v)
	} else {
		s.initAppMngAct(c, res, v, isFavSeason, isReserve, now, reserveId)
	}
	return res, nil
}

func ActivityReserveId(dm []*dmApi.CommandDm) int64 {
	for _, r := range dm {
		extraStr := r.GetExtra()
		if extraStr != "" {
			extra, err := dmExtra(extraStr)
			if err != nil {
				log.Error("ActivityReserveId dmExtra is err %+v", err)
				continue
			}
			if extra.ReserveType == 2 || extra.ReserveType == 3 {
				return extra.ReserveId
			}
		}
	}
	return 0
}

func (s *Service) SeasonView(c context.Context, mid int64, plat int8, build int, mobiApp, device, cdnIP string, buvid string, vp *api.ViewReply, pageVersion, spmid, platform string, extra map[string]string) (v *view.View, err error) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	vs := &view.ViewStatic{Arc: vp.Arc}
	s.initPages(c, vs, vp.Pages, mobiApp, build)
	vs.Stat.DisLike = 0
	bvID, err := bvid.AvToBv(vs.Aid)
	if err != nil {
		log.Error("ActivitySeason sid(%d) aid(%d) avtobv err(%v)", vp.SeasonID, vp.Aid, err)
		return nil, ecode.NothingFound
	}
	v = &view.View{ViewStatic: vs, DMSeg: 1, BvID: bvID}
	if err = s.checkAccess(c, mid, v.Aid, int(v.State), int(v.Access), vs.Arc); err != nil {
		log.Error("ActivitySeason sid(%d) aid(%d) checkAccess err(%+v)", vp.SeasonID, vp.Aid, err)
		// archive is ForbitFixed and Transcoding and StateForbitDistributing need analysis history body .
		return nil, err
	}
	if v.Access > 0 {
		v.Stat.View = 0
	}
	var arcAddit *vuApi.ArcViewAdditReply
	eg := errgroup.WithContext(c)
	// 地区版权校验
	eg.Go(func(ctx context.Context) (err error) {
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
				log.Error("ActivitySeason sid(%d) aid(%d) ipLimit mid(%d) ip(%s) cdn_ip(%s) error(%+v)", vp.SeasonID, vp.Aid, mid, metadata.String(ctx, metadata.RemoteIP), cdnIP, err)
				return err
			} else if v.AttrVal(api.AttrBitSteinsGate) == api.AttrYes {
				download = int64(location.StatusDown_ForbiddenDown)
			}
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
	eg.Go(func(ctx context.Context) (err error) {
		if arcAddit, err = s.vuDao.ArcViewAddit(ctx, v.Aid); err != nil || arcAddit == nil {
			log.Error("ActivitySeason sid(%d) aid(%d) ArcViewAddit err(%+v) or arcAddit=nil", vp.SeasonID, vp.Aid, err)
			return nil
		}
		if arcAddit.ForbidReco != nil {
			v.ForbidRec = arcAddit.ForbidReco.State
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		ugcSn, ok := s.bnjSeasons[v.SeasonID]
		if !ok {
			if ugcSn, err = s.seasonDao.Season(ctx, v.SeasonID); err != nil || ugcSn == nil {
				log.Error("ActivitySeason sid(%d) aid(%d) Season err(%+v)", vp.SeasonID, vp.Aid, err)
				return err
			}
		}
		v.UgcSeason = new(view.UgcSeason)
		v.UgcSeason.FromSeason(ugcSn)
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		initRly := s.initTag(ctx, v.Arc, v.Arc.FirstCid, mid, plat, build, pageVersion, buvid, mobiApp, spmid, platform, extra)
		s.initResult(v, initRly)
		return nil
	})
	// 获取后台活动合集配置
	eg.Go(func(ctx context.Context) (err error) {
		return s.initMngAct(ctx, v, mid, plat)
	})
	if err = eg.Wait(); err != nil {
		log.Error("ActivitySeason sid(%d) aid(%d) eg.wait() err(%+v)", vp.SeasonID, vp.Aid, err)
		return nil, err
	}
	return v, nil
}

func (s *Service) initMngAct(c context.Context, v *view.View, mid int64, plat int8) error {
	asPlat := model.PlatActSeasonApp
	if model.IsIPad(plat) {
		asPlat = model.PlatActSeasonHD
	}
	if as, ok := s.bnjActSeason[s.ActSeasonKey(asPlat, v.SeasonID)]; ok {
		if !s.checkActSeasonAccess(mid, as.ActivePlay.Whitelist) {
			log.Error("ActivitySeason sid(%d) aid(%d) fallback checkActSeasonAccess not allowed mid(%d) white(%+v)", v.SeasonID, v.Aid, mid, as.ActivePlay.Whitelist)
			return appecode.AppActivitySeasonFallback
		}
		v.MngAct = as
		return nil
	}
	var err error
	if v.MngAct, err = s.mngDao.CommonActivity(c, v.SeasonID, mid, int32(asPlat)); err != nil || v.MngAct == nil || v.MngAct.ActivePlay == nil {
		log.Error("ActivitySeason sid(%d) aid(%d) fallback CommonActivity err(%+v) mid(%d) plat(%d) or nil", v.SeasonID, v.Aid, err, mid, plat)
		return appecode.AppActivitySeasonFallback
	}
	return nil
}

func (s *Service) checkActSeasonAccess(mid int64, whiteList []int64) bool {
	if len(whiteList) == 0 {
		return true
	}
	if mid <= 0 {
		return false
	}
	for _, v := range whiteList {
		if mid == v {
			return true
		}
	}
	return false
}

func (s *Service) initActivityCM(c context.Context, v *view.View, plat int8, build int, mid int64, buvid, mobiApp, device, network, adExtra, from, spmid, fromSpmid, platform, filtered string, tids []int64, isMelloi string) {
	cfg := FromContextOrCreate(c, s.defaultViewConfigCreater())
	// 审核版本，和有屏蔽推荐池属性的稿件下 不出相关推荐任何信息
	if filtered == "1" || v.ForbidRec == 1 || v.AttrValV2(model.AttrBitV2CleanMode) == api.AttrYes || model.IsIPad(plat) {
		log.Warn("ActivitySeason sid(%d) aid(%d) no cm filtered(%s) ForbidRec(%d) plat(%d)", v.SeasonID, v.Aid, filtered, v.ForbidRec, plat)
		return
	}
	resourceID := []int32{_androidPlayerCM}
	if model.IsIPhone(plat) {
		resourceID = []int32{_iphonePlayerCM}
	}
	s.infocAd(c, mobiApp, build, network, mid, buvid, v.Aid, resourceID, platform, isMelloi, "view_request_before")
	advertNew, err := cfg.dep.AD.AdGRPC(c, mobiApp, buvid, device, build, mid, v.Author.Mid, v.Aid, v.TypeID, tids, resourceID, network, adExtra, spmid, fromSpmid, from, false)
	s.infocAd(c, mobiApp, build, network, mid, buvid, v.Aid, resourceID, platform, isMelloi, "view_request")
	if err != nil {
		log.Error("ActivitySeason sid(%d) aid(%d) AdGRPC err(%+v)", v.SeasonID, v.Aid, err)
		return
	}
	if advertNew != nil {
		v.CMConfigNew = &viewApi.CMConfig{
			AdsControl: advertNew.AdsControl,
		}
	}
}

// ipadHD有活动tab(仅有背景图)、无相关推荐模块
func (s *Service) initIpadMngAct(vr *viewApi.ViewReply, v *view.View) {
	hdMng := v.MngAct.GetHdPlay()
	if hdMng == nil {
		return
	}
	vr.ActivitySeason.SupportDislike = hdMng.SupportDislike
	vr.ActivitySeason.ActivityResource = &viewApi.ActivityResource{
		ModPoolName:     hdMng.ModPoolName,
		ModResourceName: hdMng.ModResourceName,
		BgColor:         hdMng.GetHdColor().GetBgColor(),
		SelectedBgColor: hdMng.GetHdColor().GetSelectedBgColor(),
		TextColor:       hdMng.GetHdColor().GetTextColor(),
		LightTextColor:  hdMng.GetHdColor().GetLightTextColor(),
		DarkTextColor:   hdMng.GetHdColor().GetDarkTextColor(),
		DividerColor:    hdMng.GetHdColor().GetDividerColor(),
	}
	// tab模块，HD仅支持背景图
	if hdMng.Background != "" {
		vr.ActivitySeason.Tab = &viewApi.Tab{
			Background: hdMng.Background,
		}
	}
}

//nolint:gomnd
func (s *Service) initAppMngAct(c context.Context, vr *viewApi.ViewReply, v *view.View, isFav bool, isReserve bool, now time.Time, reserveId int64) {
	appMng := v.MngAct.GetAppPlay()
	if appMng == nil {
		return
	}
	// 是否支持点踩
	vr.ActivitySeason.SupportDislike = appMng.SupportDislike
	// 活动mod资源
	vr.ActivitySeason.ActivityResource = &viewApi.ActivityResource{
		ModPoolName:     appMng.ModPoolName,
		ModResourceName: appMng.ModResourceName,
		BgColor:         appMng.GetAppColor().GetBgColor(),
		SelectedBgColor: appMng.GetAppColor().GetSelectedBgColor(),
		TextColor:       appMng.GetAppColor().GetTextColor(),
		LightTextColor:  appMng.GetAppColor().GetLightTextColor(),
		DarkTextColor:   appMng.GetAppColor().GetDarkTextColor(),
		DividerColor:    appMng.GetAppColor().GetDividerColor(),
	}
	var relateItems []*viewApi.RelateItem
	// 相关推荐模块
	if r, ok := appMng.CommRecommend[v.Aid]; ok {
		for _, v := range r.CommRecommends {
			relateItems = append(relateItems, &viewApi.RelateItem{
				Url:   v.CommJumpUrl,
				Cover: v.CommPic,
			})
		}
	}
	// 运营推荐和AI推荐任意有数据即可
	if len(relateItems) > 0 || len(v.Relates) > 0 {
		//版本判断, 繁体版不下发
		buildBool := pd.WithContext(c).IsPlatAndroidI().Or().IsPlatIPhoneI().Or().IsPlatIPadI().FinishOr(false)
		if !buildBool {
			vr.ActivitySeason.OperationRelate = &viewApi.OperationRelate{
				Title:        s.c.ActivitySeason.RelateTitle,
				RelateItem:   relateItems,
				AiRelateItem: view.FromRelates(v.Relates),
			}
		}
	}
	// tab模块
	vr.ActivitySeason.Tab = func() *viewApi.Tab {
		// 背景图和其他内容可独立配置，如果没有tab内容
		if appMng.TabType == 0 {
			if appMng.Background == "" {
				return nil
			}
			return &viewApi.Tab{Background: appMng.Background}
		}
		tmpTab := &viewApi.Tab{
			Otype:      viewApi.TabOtype(appMng.TabLinkType),
			Style:      viewApi.TabStyle(appMng.TabType),
			Background: appMng.Background,
		}
		// 数据完整性校验，
		invalid := false
		switch appMng.TabLinkType {
		case 1: // h5
			if appMng.TabUrl == "" {
				log.Error("活动页告警 后台配置 ActivitySeason sid(%d) aid(%d) Tab H5 invalid (%+v)", v.SeasonID, v.Aid, appMng)
				invalid = true
			}
			tmpTab.Uri = appMng.TabUrl
		case 2: // native
			oid, _ := strconv.ParseInt(appMng.TabUrl, 10, 64)
			if oid <= 0 {
				log.Error("活动页告警 后台配置 ActivitySeason sid(%d) aid(%d) Tab native invalid (%+v)", v.SeasonID, v.Aid, appMng)
				invalid = true
			}
			tmpTab.Oid = oid
		default:
			log.Error("活动页告警 后台配置 ActivitySeason sid(%d) aid(%d) Tab unknown link_type (%+v)", v.SeasonID, v.Aid, appMng)
			invalid = true
		}
		if appMng.TabContent == "" {
			log.Error("活动页告警 后台配置 ActivitySeason sid(%d) aid(%d)  Tab invalid tab_content (%+v)", v.SeasonID, v.Aid, appMng)
			invalid = true
		}
		switch appMng.TabType {
		case 1: // 文字
			tmpTab.Text = appMng.TabContent
		case 2: // 图片
			tmpTab.Pic = appMng.TabContent
		default:
			log.Error("活动页告警 后台配置 ActivitySeason sid(%d) aid(%d) Tab unknown tab_type (%+v)", v.SeasonID, v.Aid, appMng)
			invalid = true
		}
		if invalid {
			if tmpTab.Background == "" {
				return nil
			}
			return &viewApi.Tab{Background: tmpTab.Background}
		}
		return tmpTab
	}()
	// 预约模块
	order := &viewApi.Order{
		Title:             v.UgcSeason.Title,
		SeasonStatView:    int64(v.UgcSeason.Stat.View),
		SeasonStatDanmaku: int64(v.UgcSeason.Stat.Danmaku),
		Intro:             v.UgcSeason.Intro,
	}
	var button, buttonSelected string
	// 未开播 走预约活动逻辑
	if v.MngAct.ActivePlay.IsContainedLive && v.MngAct.ActivePlay.Stime > now.Unix() {
		if lb, ok := appMng.Menus[model.LiveBefore]; ok {
			order.Status = isReserve
			order.OrderParam = &viewApi.Order_Reserve{
				Reserve: &viewApi.BizReserveActivityParam{
					ActivityId: lb.MenuId,
					From:       "video_page",
					Type:       "video_page",
					Oid:        v.SeasonID,
					ReserveId:  reserveId,
				},
			}
			order.OrderType = viewApi.BizType_BizTypeReserveActivity
			button = lb.UnclickedText
			buttonSelected = lb.ClickedText
		} else {
			log.Error("活动页告警 后台配置 ActivitySeason sid(%d) aid(%d) live before config miss (%+v)", v.SeasonID, v.Aid, appMng)
		}
		order.ButtonTitle = view.FormatOrderButton(button, false)
		order.ButtonSelectedTitle = view.FormatOrderButton(buttonSelected, true)
	} else {
		//已开播 走收藏合集逻辑
		order.Status = isFav
		order.OrderParam = &viewApi.Order_FavSeason{
			FavSeason: &viewApi.BizFavSeasonParam{
				SeasonId: v.SeasonID,
			},
		}
		order.OrderType = viewApi.BizType_BizTypeFavSeason
		if la, ok := appMng.Menus[model.LiveAfter]; ok {
			button = la.UnclickedText
			buttonSelected = la.ClickedText
		} else {
			log.Error("活动页告警 后台配置 ActivitySeason sid(%d) aid(%d) live after config miss (%+v)", v.SeasonID, v.Aid, appMng)
		}
		order.ButtonTitle = view.FormatFavButton(button, false)
		order.ButtonSelectedTitle = view.FormatFavButton(buttonSelected, true)
	}
	vr.ActivitySeason.Order = order
}

func (s *Service) ClickActivitySeason(c context.Context, arg *viewApi.ClickActivitySeasonReq, mid int64, dev device.Device) error {
	err := func() error {
		switch arg.OrderType {
		case viewApi.BizType_BizTypeReserveActivity:
			param := arg.GetReserve()
			if param == nil {
				return ecode.RequestErr
			}
			return s.actDao.Reserve(c, param.ActivityId, mid, param.Oid, arg.Action, arg.Spmid, param.From, param.Type, dev)
		case viewApi.BizType_BizTypeFavSeason:
			param := arg.GetFavSeason()
			if param == nil {
				return ecode.RequestErr
			}
			return s.favDao.AddFav(c, mid, param.SeasonId, arg.Action, model.FavTypeSeason, dev.RawMobiApp, dev.RawPlatform, dev.Device)
		default:
			return ecode.RequestErr
		}
	}()
	if err != nil {
		log.Error("ClickActivitySeason error(%+v) arg(%+v) mid(%d) dev(%+v)", err, arg, mid, dev)
		return err
	}
	return nil
}

func (s *Service) initAIRelate(c context.Context, v *view.View, plat int8, build, autoplay int, mid int64, buvid, mobiApp, device, from, spmid, fromSpmid, trackid, filtered, slocale, clocale, isMelloi, ip string, now time.Time, pageVersion string) {
	// 审核版本，和有屏蔽推荐池属性的稿件下 不出相关推荐任何信息
	if filtered == "1" || v.ForbidRec == 1 {
		log.Warn("ActivitySeason sid(%d) aid(%d) no relate filtered(%s) ForbidRec(%d)", v.SeasonID, v.Aid, filtered, v.ForbidRec)
		return
	}
	var (
		rls        []*view.Relate
		err        error
		relateConf *view.RelateConf
	)
	if mid > 0 || buvid != "" {
		if rls, v.TabInfo, v.PlayParam, v.UserFeature, v.ReturnCode, v.PvFeature, relateConf, err = s.newRcmdRelate(c, plat, v.Aid, mid, v.ZoneID, buvid, mobiApp, from, trackid, model.RelateCmd, "", build, 0, autoplay, 1, true, pageVersion, fromSpmid); err != nil {
			log.Error("ActivitySeason s.newRcmdRelate(%d) error(%+v)", v.Aid, err)
		}
		if relateConf != nil && v.Config != nil {
			v.Config.AutoplayCountdown = s.c.ViewConfig.AutoplayCountdown
			if relateConf.AutoplayCountdown > 0 {
				v.Config.AutoplayCountdown = relateConf.AutoplayCountdown
			}
			v.Config.PageRefresh = relateConf.ReturnPage
			v.Config.AutoplayDesc = relateConf.AutoplayToast
			v.Config.RelatesStyle = relateConf.RelatesStyle
			v.Config.RelateGifExp = relateConf.GifExp
			v.Config.RecThreePointStyle = relateConf.RecThreePointStyle
		}
	}
	// ai：code=-3表示无有效结果稿;code=5表示屏蔽用户黑名单;code=-2表示内部拉用户信息缺失
	if len(rls) == 0 && v.ReturnCode != "-3" && v.ReturnCode != "-5" { // -3和-5不要取灾备数据
		rls, _ = s.dealRcmdRelate(c, plat, v.Aid, mid, build, mobiApp, device)
		log.Warn("ActivitySeason s.dealRcmdRelate aid(%d) mid(%d) buvid(%s) build(%d) mobiApp(%s) device(%s)", v.Aid, mid, buvid, build, mobiApp, device)
	} else {
		v.IsRec = 1
		log.Info("ActivitySeason s.newRcmdRelate returncode(%s) aid(%d) mid(%d) buvid(%s)", v.ReturnCode, v.Aid, mid, buvid)
	}
	v.RelatesInfoc = &view.RelatesInfoc{}
	v.RelatesInfoc.SetAdCode("NULL")
	if len(rls) == 0 {
		s.prom.Incr("没有任何相关推荐")
		v.RelatesInfoc.SetPKCode(view.AdFirstForRelate0)
		return
	}
	v.Relates = s.sortRelates(v, false, rls, nil, nil, nil, plat, build)
	if i18n.PreferTraditionalChinese(c, slocale, clocale) {
		for _, rl := range v.Relates {
			i18n.TranslateAsTCV2(&rl.Title)
		}
	}
	// 相关推荐曝光上报
	s.RelateInfoc(mid, v.Aid, int(plat), strconv.Itoa(build), buvid, ip, model.PathView, v.ReturnCode, v.UserFeature,
		from, "", v.Relates, now, v.IsRec, int(autoplay), v.PlayParam, trackid, model.PageTypeRelate, fromSpmid,
		spmid, v.PvFeature, v.TabInfo, isMelloi, v.RelatesInfoc, 0)
}
