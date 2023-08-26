package dynamicV2

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/exp/ab"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	submdl "go-gateway/app/app-svr/app-dynamic/interface/model/subscription"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	feature "go-gateway/app/app-svr/feature/service/sdk"
	"go-gateway/pkg/idsafe/bvid"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
	livexroomfeed "git.bilibili.co/bapis/bapis-go/live/xroom-feed"
)

const (
	_play_online_total = "total"
)

// nolint:gocognit
func (s *Service) dynCardForward(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	// 获取文案
	var tips = s.c.Resource.Text.ModuleDynamicForwardDefaultTips
	if dynCtx.Dyn.Origin != nil {
		if !dynCtx.Dyn.Origin.Visible {
			// 源卡不可展示
			if dynCtx.Dyn.Origin.Tips != "" {
				tips = dynCtx.Dyn.Origin.Tips
			}
		}
	}
	// 默认模块
	var module = &api.Module{
		ModuleType: api.DynModuleType_module_item_null,
		ModuleItem: &api.Module_ModuleItemNull{
			ModuleItemNull: &api.ModuleItemNull{
				Icon: s.c.Resource.Icon.ModuleDynamicItemNull,
				Text: tips,
			},
		},
	}
	// 处理正常数据
	if !dynCtx.Interim.ForwardOrigFaild {
		// 获取拼接卡片的func handler
		var (
			dyn       = new(mdlv2.Dynamic)
			dynCtxTmp = new(mdlv2.DynamicContext)
			foldList  *mdlv2.FoldList
		)
		*dyn = *dynCtx.Dyn.Origin
		*dynCtxTmp = *dynCtx
		// 感知转卡信息
		dyn.Forward = dynCtx.Dyn
		if dynCtx.Dyn.RType == mdlv2.DynShare {
			foldList = s.procListReply(c, []*mdlv2.Dynamic{dyn}, dynCtxTmp, general, _handleTypeShare)
		} else {
			foldList = s.procListReply(c, []*mdlv2.Dynamic{dyn}, dynCtxTmp, general, _handleTypeForward)
		}
		if len(foldList.List) == 0 || foldList.List[0].Item == nil {
			xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "date_faild")
			goto END
		}
		xmetric.DynamicForward.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtxTmp.Dyn.Type))
		var (
			card        = new(api.MdlDynForward)
			dynamicItem = foldList.List[0].Item
		)
		card.Item = dynamicItem
		card.Rtype = dynCtx.Dyn.RType
		// 转发卡物料
		// 表情
		for k := range dynCtxTmp.Emoji {
			if dynCtx.Emoji == nil {
				dynCtx.Emoji = make(map[string]struct{})
			}
			dynCtx.Emoji[k] = struct{}{}
		}
		// cv
		for k := range dynCtxTmp.BackfillCvID {
			if dynCtx.BackfillCvID == nil {
				dynCtx.BackfillCvID = make(map[string]struct{})
			}
			dynCtx.BackfillCvID[k] = struct{}{}
		}
		// bv
		for k := range dynCtxTmp.BackfillBvID {
			if dynCtx.BackfillBvID == nil {
				dynCtx.BackfillBvID = make(map[string]struct{})
			}
			dynCtx.BackfillBvID[k] = struct{}{}
		}
		// av
		for k := range dynCtxTmp.BackfillAvID {
			if dynCtx.BackfillAvID == nil {
				dynCtx.BackfillAvID = make(map[string]struct{})
			}
			dynCtx.BackfillAvID[k] = struct{}{}
		}
		// url
		for k := range dynCtxTmp.BackfillDescURL {
			if dynCtx.BackfillDescURL == nil {
				dynCtx.BackfillDescURL = make(map[string]*mdlv2.BackfillDescURLItem)
			}
			dynCtx.BackfillDescURL[k] = nil
		}
		// 扩展字段
		dynCtx.Interim.VoteID = dynCtxTmp.Interim.VoteID
		dynCtx.DynamicItem.Extend.OrigDynIdStr = dynamicItem.Extend.DynIdStr
		dynCtx.DynamicItem.Extend.OrigName = dynamicItem.Extend.OrigName     // 转发卡使用内层物料数据
		dynCtx.DynamicItem.Extend.OrigImgUrl = dynamicItem.Extend.OrigImgUrl // 转发卡使用内层物料数据
		dynCtx.DynamicItem.Extend.OrigDesc = dynamicItem.Extend.OrigDesc     // 转发卡使用内层物料数据
		dynCtx.DynamicItem.Extend.OrigDynType = dynamicItem.CardType
		dynCtx.DynamicItem.ItemType = dynamicItem.CardType // 兼容逻辑, android 6.15版本使用
		for _, origItem := range dynamicItem.Modules {
			//nolint:exhaustive
			switch origItem.ModuleType {
			case api.DynModuleType_module_extend:
				origModule := origItem.ModuleItem.(*api.Module_ModuleExtend)
				if origModule.ModuleExtend != nil && len(origModule.ModuleExtend.Extend) > 0 {
					dynCtx.Interim.IsPassExtend = true
				}
			case api.DynModuleType_module_additional:
				origModule := origItem.ModuleItem.(*api.Module_ModuleAdditional)
				if origModule.ModuleAdditional != nil {
					dynCtx.Interim.IsPassAddition = true
				}
			}
		}
		module = &api.Module{
			ModuleType: api.DynModuleType_module_dynamic,
			ModuleItem: &api.Module_ModuleDynamic{
				ModuleDynamic: &api.ModuleDynamic{
					Type: api.ModuleDynamicType_mdl_dyn_forward,
					ModuleItem: &api.ModuleDynamic_DynForward{
						DynForward: card,
					},
				},
			},
		}
	}
END:
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// nolint:gocognit
func (s *Service) dynCardAv(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	ap, _ := dynCtx.GetArchive(dynCtx.Dyn.Rid)
	var archive = ap.Arc
	card := &api.MdlDynArchive{
		Title:           s.getTitle(archive.Title, dynCtx),
		Cover:           archive.Pic,
		CoverLeftText_1: s.videoDuration(archive.Duration),
		CoverLeftText_2: fmt.Sprintf("%s观看", s.numTransfer(int(archive.Stat.View))),
		CoverLeftText_3: fmt.Sprintf("%s弹幕", s.numTransfer(int(archive.Stat.Danmaku))),
		Avid:            archive.Aid,
		Cid:             archive.FirstCid,
		MediaType:       api.MediaType_MediaTypeUGC,
		Dimension: &api.Dimension{
			Height:          archive.Dimension.Height,
			Width:           archive.Dimension.Width,
			Rotate:          archive.Dimension.Rotate,
			ForceHorizontal: true, // UGC视频卡不区分横竖屏全量用横屏封面
		},
		Duration: archive.Duration,
		View:     archive.Stat.View,
	}
	card.Bvid, _ = bvid.AvToBv(archive.Aid)
	var (
		playurl *arcgrpc.PlayerInfo
		ok      bool
	)
	if playurl, ok = ap.PlayerInfo[dynCtx.Interim.CID]; !ok {
		if playurl, ok = ap.PlayerInfo[ap.DefaultPlayerCid]; !ok {
			playurl = ap.PlayerInfo[ap.Arc.FirstCid]
		}
	}
	if playurl != nil && playurl.PlayerExtra != nil {
		// progress 单位是毫秒
		card.PartProgress = playurl.PlayerExtra.GetProgress() / time.Second.Milliseconds()
		if archive.Videos > 1 {
			card.PartDuration = dynCtx.GetArcPart(dynCtx.GetArchiveAutoPlayCid(ap)).GetDuration()
		} else {
			card.PartDuration = archive.Duration
		}
		// 付费视频不支持inline也不秒开，不下发进度条
		if card.PartDuration > 0 && general.DynFrom == _dynFromFilterContinue && !mdlv2.PayAttrVal(archive) {
			card.ShowProgress = true
		}
		if playurl.PlayerExtra.Dimension != nil {
			card.Cid = playurl.PlayerExtra.Cid
			card.Dimension.Height = playurl.PlayerExtra.Dimension.Height
			card.Dimension.Width = playurl.PlayerExtra.Dimension.Width
			card.Dimension.Rotate = playurl.PlayerExtra.Dimension.Rotate
		}
	}
	if dynCtx.Dyn.Property != nil && (dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_ARCHIVE || dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_PLAY_BACK ||
		dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_PREMIERE_RESERVE) {
		// UP主预约是否召回
		card.ReserveType = api.ReserveType_reserve_recall
	}
	if g, ok := dynCtx.Grayscale[s.c.Grayscale.ShowPlayIcon.Key]; ok {
		switch g {
		case 1:
			card.PlayIcon = s.c.Resource.Icon.ModuleDynamicPlayIcon
		}
	}
	card.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))

	// PGC特殊逻辑
	if archive.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && archive.RedirectURL != "" {
		card.Uri = archive.RedirectURL
		card.IsPGC = true
		if playurl, ok = ap.PlayerInfo[ap.DefaultPlayerCid]; ok && playurl.PlayerExtra != nil && playurl.PlayerExtra.PgcPlayerExtra != nil {
			if playurl.PlayerExtra.PgcPlayerExtra.IsPreview == 1 {
				card.IsPreview = true
			}
			card.EpisodeId = playurl.PlayerExtra.PgcPlayerExtra.EpisodeId
			card.SubType = playurl.PlayerExtra.PgcPlayerExtra.SubType
			card.PgcSeasonId = playurl.PlayerExtra.PgcPlayerExtra.PgcSeasonId
		}
	}
	// 小视频特殊处理
	card.Stype = mdlv2.GetArchiveSType(dynCtx.Dyn.SType)
	if card.Stype == api.VideoType_video_type_story {
		if !feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynStory, &feature.OriginResutl{
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynStoryIOS) ||
				(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynStoryAndroid)}) {
			card.Stype = api.VideoType_video_type_dynamic
		}
	}
	if card.Stype == api.VideoType_video_type_dynamic || card.Stype == api.VideoType_video_type_story {
		card.Title = ""
		card.IsPGC = false
	}
	if archive.Rights.IsCooperation == 1 {
		card.Badge = append(card.Badge, mdlv2.CooperationBadge)
	}
	if archive.Rights.UGCPay == 1 {
		card.Badge = append(card.Badge, mdlv2.PayBadge)
	}
	// 付费合集（但付费合集中的免费稿件在单视频卡片上不展示付费角标）
	if mdlv2.PayAttrVal(archive) && archive.Rights.GetArcPayFreeWatch() == 0 {
		card.Badge = append(card.Badge, mdlv2.PayBadge)
	}
	if card.Stype == api.VideoType_video_type_playback {
		card.Badge = append(card.Badge, mdlv2.PlayBackBadge)
	}
	// 新版本才出story角标
	if card.Stype == api.VideoType_video_type_story || card.Stype == api.VideoType_video_type_dynamic {
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynStory, &feature.OriginResutl{
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynStoryIOS) ||
				(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynStoryAndroid)}) {
			card.Badge = append(card.Badge, mdlv2.StoryBadge)
		}
	}
	if len(card.Badge) == 0 {
		// 新版本才出角标
		if (general.Device.MobiApp() == "iphone" && general.Device.Build >= 62000000 || general.Device.MobiApp() == "android" && general.Device.Build >= 6195000 || general.Device.MobiApp() == "ipad" && general.Device.Build >= 31500100) && archive.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes || general.IsAndroidHD() || general.IsPad() {
			// 新的角标
			if dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.PgcBadge != nil && dynCtx.Dyn.PassThrough.PgcBadge.EpisodeId > 0 {
				if dynCtx.Dyn.PassThrough.PgcBadge.SectionType == 0 {
					// 是否是PGC正片，上报字段
					card.IsFeature = true
					card.IsPGC = true
				}
				if pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.PassThrough.PgcBadge.EpisodeId)); ok {
					if pgc.Season != nil && pgc.SectionType == 0 {
						if dynCtx.Dyn.PassThrough.PgcBadge.Show {
							// 追番人数角标
							if pgc.Stat.FollowDesc != "" {
								card.BadgeCategory = append(card.BadgeCategory, mdlv2.BadgeStyleFrom(mdlv2.BgColorTransparentGray, pgc.Stat.FollowDesc))
							}
							// PGC角标
							if pgc.Season.TypeName != "" {
								card.BadgeCategory = append(card.BadgeCategory, mdlv2.BadgeStyleFrom(mdlv2.BgColorPink, pgc.Season.TypeName))
							}
						}
					}
				}
			}
		}
	}
	// 首映 且 是召回卡
	if dynCtx.Dyn.Property != nil && dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_PREMIERE_RESERVE && archive.Premiere != nil &&
		(general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynPropertyIOS || general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.DynPropertyAndroid) {
		countInfo, ok := dynCtx.ResPlayUrlCount[ap.Arc.Aid]
		// nolint:exhaustive
		switch archive.Premiere.State {
		case arcgrpc.PremiereState_premiere_in: // 首映中
			if ok {
				card.Badge = nil
				card.CoverLeftText_2 = ""
				card.CoverLeftText_3 = ""
				card.BadgeCategory = append(card.BadgeCategory, mdlv2.BadgeStyleFrom(mdlv2.BgColorGray, fmt.Sprintf("%d人在线", countInfo.Count[_play_online_total])))
				card.ShowPremiereBadge = true
				// 首映中稿件原卡添加详情页特殊浮层私参
				card.Uri = s.inArchivePremiereArg()(card.Uri)
			}
		}
	}
	card.CanPlay = mdlv2.CanPlay(archive.Rights.Autoplay)
	dynamic := &api.ModuleDynamic{
		Type: api.ModuleDynamicType_mdl_dyn_archive,
		ModuleItem: &api.ModuleDynamic_DynArchive{
			DynArchive: card,
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: dynamic,
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardPGC(_ context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	pgc, _ := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
	card := &api.MdlDynPGC{
		Title:           s.getTitle(pgc.CardShowTitle, dynCtx),
		Cover:           pgc.Cover,
		Uri:             pgc.Url,
		CoverLeftText_1: s.videoDuration(pgc.Duration),
		CoverLeftText_2: fmt.Sprintf("%s观看", s.numTransfer(int(pgc.Stat.Play))),
		CoverLeftText_3: fmt.Sprintf("%s弹幕", s.numTransfer(int(pgc.Stat.Danmaku))),
		Cid:             pgc.Cid,
		Epid:            int64(pgc.EpisodeId),
		Aid:             pgc.Aid,
		MediaType:       api.MediaType_MediaTypePGC,
		IsPreview:       mdlv2.Int32ToBool(int32(pgc.IsPreview)),
		Dimension: &api.Dimension{
			Height: int64(pgc.Dimension.Height),
			Width:  int64(pgc.Dimension.Width),
			Rotate: int64(pgc.Dimension.Rotate),
		},
		Duration: pgc.Duration,
		SubType:  dynCtx.Dyn.GetPGCSubType(),
	}
	if g, ok := dynCtx.Grayscale[s.c.Grayscale.ShowPlayIcon.Key]; ok {
		switch g {
		case 1:
			card.PlayIcon = s.c.Resource.Icon.ModuleDynamicPlayIcon
		}
	}
	if pgc.Season != nil {
		season := &api.PGCSeason{
			IsFinish: int32(pgc.Season.IsFinish),
			Title:    pgc.Season.Title,
			Type:     int32(pgc.Season.Type),
		}
		card.Season = season
		card.SeasonId = int64(pgc.Season.SeasonId)
		if general.Device.MobiApp() == "iphone" && general.Device.Build >= 62000000 || general.Device.MobiApp() == "android" && general.Device.Build >= 6195000 || general.Device.MobiApp() == "ipad" && general.Device.Build >= 31500100 || general.IsAndroidHD() || general.IsPad() {
			if pgc.SectionType == 0 {
				// 是否是PGC正片，上报字段
				card.IsFeature = true
				// 新版本才出角标
				if dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.PgcBadge != nil && dynCtx.Dyn.PassThrough.PgcBadge.Show {
					// 追番人数角标
					if pgc.Stat.FollowDesc != "" {
						card.BadgeCategory = append(card.BadgeCategory, mdlv2.BadgeStyleFrom(mdlv2.BgColorTransparentGray, pgc.Stat.FollowDesc))
					}
					// PGC角标
					if pgc.Season.TypeName != "" {
						card.BadgeCategory = append(card.BadgeCategory, mdlv2.BadgeStyleFrom(mdlv2.BgColorPink, pgc.Season.TypeName))
					}
				}
			}
		}
	}
	card.CanPlay = pgc.PlayerInfo != nil
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type: api.ModuleDynamicType_mdl_dyn_pgc,
				ModuleItem: &api.ModuleDynamic_DynPgc{
					DynPgc: card,
				},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardCourBatch(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	batch, _ := dynCtx.GetResCheeseBatch(dynCtx.Dyn.Rid)
	card := &api.MdlDynCourBatch{
		Title:           s.getTitle(batch.Title, dynCtx),
		Cover:           batch.NewEp.Cover,
		Uri:             batch.URL,
		Text_1:          batch.NewEp.Title,
		CoverLeftText_1: s.videoDuration(batch.Stat.Duration),
		CoverLeftText_2: fmt.Sprintf("%s观看", s.numTransfer(int(batch.Stat.PlayCount))),
		CoverLeftText_3: fmt.Sprintf("%s弹幕", s.numTransfer(int(batch.Stat.DmCount))),
		Avid:            batch.InlineVideo.Aid,
		Cid:             batch.InlineVideo.Cid,
		Epid:            batch.InlineVideo.Epid,
		Duration:        batch.InlineVideo.Duration,
		SeasonId:        int64(batch.SeasonID),
		IsPreview:       batch.InlineVideo.IsPreview,
	}
	if batch.InlineVideo.Url != "" {
		card.CanPlay = true
	}
	card.Uri = model.FillURI(model.GotoURL, batch.URL, model.BatchPlayHandler(batch))
	if g, ok := dynCtx.Grayscale[s.c.Grayscale.ShowPlayIcon.Key]; ok {
		switch g {
		case 1:
			card.PlayIcon = s.c.Resource.Icon.ModuleDynamicPlayIcon
		}
	}
	badge := &api.VideoBadge{
		Text:           batch.Badge.Text,
		TextColor:      batch.Badge.TextColor,
		TextColorNight: batch.Badge.TextDarkColor,
		BgColor:        batch.Badge.BgColor,
		BgColorNight:   batch.Badge.BgDarkColor,
		BgStyle:        model.BgStyleFill,
	}
	card.Badge = badge
	if batch.UpdateCount > 1 {
		card.Text_2 = fmt.Sprintf("等%d个视频", batch.UpdateCount)
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type: api.ModuleDynamicType_mdl_dyn_cour_batch,
				ModuleItem: &api.ModuleDynamic_DynCourBatch{
					DynCourBatch: card,
				},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardCourUp(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	season, _ := dynCtx.GetResCheeseSeason(dynCtx.Dyn.Rid)
	card := &api.MdlDynCourUp{
		Title:     s.getTitle(season.Title, dynCtx),
		Cover:     season.Cover,
		Uri:       season.URL,
		Text_1:    season.UpdateInfo,
		Desc:      season.Subtitle,
		Avid:      season.InlineVideo.Aid,
		Cid:       season.InlineVideo.Cid,
		Epid:      season.InlineVideo.Epid,
		Duration:  season.InlineVideo.Duration,
		SeasonId:  int64(season.ID),
		IsPreview: season.InlineVideo.IsPreview,
	}
	if season.InlineVideo.Url != "" {
		card.CanPlay = true
	}
	card.Uri = model.FillURI(model.GotoURL, season.URL, model.SeasonPlayHandler(season))
	if g, ok := dynCtx.Grayscale[s.c.Grayscale.ShowPlayIcon.Key]; ok {
		switch g {
		case 1:
			card.PlayIcon = s.c.Resource.Icon.ModuleDynamicPlayIcon
		}
	}
	badge := &api.VideoBadge{
		Text:           season.Badge.Text,
		TextColor:      season.Badge.TextColor,
		TextColorNight: season.Badge.TextDarkColor,
		BgColor:        season.Badge.BgColor,
		BgColorNight:   season.Badge.BgDarkColor,
		BgStyle:        model.BgStyleFill,
	}
	card.Badge = badge
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type: api.ModuleDynamicType_mdl_dyn_cour_up,
				ModuleItem: &api.ModuleDynamic_DynCourBatchUp{
					DynCourBatchUp: card,
				},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardCourSeason(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	season, _ := dynCtx.GetResCheeseSeason(dynCtx.Dyn.Rid)
	card := &api.MdlDynCourSeason{
		Title:     s.getTitle(season.Title, dynCtx),
		Cover:     season.Cover,
		Uri:       season.URL,
		Text_1:    season.UpdateInfo,
		Desc:      season.Subtitle,
		Avid:      season.InlineVideo.Aid,
		Cid:       season.InlineVideo.Cid,
		Epid:      season.InlineVideo.Epid,
		Duration:  season.InlineVideo.Duration,
		SeasonId:  dynCtx.Dyn.Rid,
		IsPreview: season.InlineVideo.IsPreview,
	}
	if season.InlineVideo.Url != "" {
		card.CanPlay = true
	}
	card.Uri = model.FillURI(model.GotoURL, season.URL, model.SeasonPlayHandler(season))
	if g, ok := dynCtx.Grayscale[s.c.Grayscale.ShowPlayIcon.Key]; ok {
		switch g {
		case 1:
			card.PlayIcon = s.c.Resource.Icon.ModuleDynamicPlayIcon
		}
	}
	badge := &api.VideoBadge{
		Text:           season.Badge.Text,
		TextColor:      season.Badge.TextColor,
		TextColorNight: season.Badge.TextDarkColor,
		BgColor:        season.Badge.BgColor,
		BgColorNight:   season.Badge.BgDarkColor,
		BgStyle:        model.BgStyleFill,
	}
	card.Badge = badge
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_cour_season,
				ModuleItem: &api.ModuleDynamic_DynCourSeason{DynCourSeason: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardLive(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	live, _ := dynCtx.GetResLive(dynCtx.Dyn.Rid)
	card := &api.MdlDynLive{
		Id:         live.RoomId,
		Title:      live.Title,
		Cover:      live.Cover,
		CoverLabel: live.AreaName,
	}
	if dynCtx.Dyn.Property != nil && dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE {
		// UP主预约是否召回
		card.ReserveType = api.ReserveType_reserve_recall
	}
	if live.PopularityCount > 0 {
		card.CoverLabel2 = model.StatString(live.PopularityCount, "人气")
		if show := live.WatchedShow; show != nil && show.TextLarge != "" {
			card.CoverLabel2 = show.TextLarge
		}
	}
	card.Badge = &api.VideoBadge{Text: s.c.Resource.Text.ModuleDynamicLiveBadgeFinish}
	if live.LiveStatus == 1 {
		card.LiveState = api.LiveState_live_live
		card.Badge = &api.VideoBadge{Text: s.c.Resource.Text.ModuleDynamicLiveBadgeLiving}
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_live,
				ModuleItem: &api.ModuleDynamic_DynCommonLive{DynCommonLive: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardMedialist(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	medialist, _ := dynCtx.GetResMedialist(dynCtx.Dyn.Rid)
	card := &api.MdlDynMedialist{
		Id:        medialist.ID,
		Title:     s.getTitle(medialist.Title, dynCtx),
		SubTitle:  fmt.Sprintf("%d个内容", medialist.MediaCount),
		Cover:     medialist.Cover,
		CoverType: medialist.CoverType,
		Badge: &api.VideoBadge{
			Text:             s.c.Resource.Others.ModuleDynamicMedialistBadge.Text,
			TextColor:        s.c.Resource.Others.ModuleDynamicMedialistBadge.TextColor,
			TextColorNight:   s.c.Resource.Others.ModuleDynamicMedialistBadge.TextColorNight,
			BgColor:          s.c.Resource.Others.ModuleDynamicMedialistBadge.BgColor,
			BgColorNight:     s.c.Resource.Others.ModuleDynamicMedialistBadge.BgColorNight,
			BorderColor:      s.c.Resource.Others.ModuleDynamicMedialistBadge.BorderColor,
			BorderColorNight: s.c.Resource.Others.ModuleDynamicMedialistBadge.BorderColorNight,
			BgStyle:          s.c.Resource.Others.ModuleDynamicMedialistBadge.BgStyle,
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_medialist,
				ModuleItem: &api.ModuleDynamic_DynMedialist{DynMedialist: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardUGCSeasonShare(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	ss, ok := dynCtx.GetResUGCSeason(dynCtx.Dyn.Rid)
	if !ok || ss == nil {
		dynCtx.Interim.IsPassCard = true
	}
	card := &api.MdlDynMedialist{
		Id:        ss.ID,
		Uri:       model.FillURI(model.GotoAv, strconv.FormatInt(ss.FirstAid, 10), model.SuffixHandler("auto_float_layer=3")), // 跳入时唤起合集浮层
		Title:     s.getTitle(ss.Title, dynCtx),
		SubTitle:  fmt.Sprintf("%d集 %s", ss.EpCount, model.StatString(int64(ss.Stat.View), "观看")),
		Cover:     ss.Cover,
		CoverType: 2, // 视频封面
		Badge: &api.VideoBadge{
			Text:             "合集",
			TextColor:        s.c.Resource.Others.ModuleDynamicMedialistBadge.TextColor,
			TextColorNight:   s.c.Resource.Others.ModuleDynamicMedialistBadge.TextColorNight,
			BgColor:          s.c.Resource.Others.ModuleDynamicMedialistBadge.BgColor,
			BgColorNight:     s.c.Resource.Others.ModuleDynamicMedialistBadge.BgColorNight,
			BorderColor:      s.c.Resource.Others.ModuleDynamicMedialistBadge.BorderColor,
			BorderColorNight: s.c.Resource.Others.ModuleDynamicMedialistBadge.BorderColorNight,
			BgStyle:          s.c.Resource.Others.ModuleDynamicMedialistBadge.BgStyle,
		},
		CoverBottomRightIcon: "https://i0.hdslb.com/bfs/activity-plat/static/20220913/fd43ade10c04329bcc177dcb1cdefce0/PLB7rpAr6O.png",
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_medialist,
				ModuleItem: &api.ModuleDynamic_DynMedialist{DynMedialist: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardDraw(_ context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	draw, _ := dynCtx.GetResDraw(dynCtx.Dyn.Rid)
	card := &api.MdlDynDraw{
		Uri:         dynCtx.Interim.PromoURI,
		Id:          dynCtx.Dyn.Rid,
		IsDrawFirst: general.Config.IsDetailDrawFirst(),
	}
	for _, pic := range draw.Item.Pictures {
		if pic == nil {
			xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "date_faild")
			continue
		}
		i := &api.MdlDynDrawItem{
			Src:    pic.ImgSrc,
			Width:  pic.ImgWidth,
			Height: pic.ImgHeight,
			Size_:  pic.ImgSize,
		}
		for _, picTag := range pic.ImgTags {
			if picTag == nil {
				xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "date_faild")
				continue
			}
			switch picTag.Type {
			case mdlv2.DrawTagTypeCommon:
				i.Tags = append(i.Tags, &api.MdlDynDrawTag{
					Type: api.MdlDynDrawTagType_mdl_draw_tag_common,
					Item: &api.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Url:         picTag.Url,
					},
				})
			case mdlv2.DrawTagTypeGoods:
				i.Tags = append(i.Tags, &api.MdlDynDrawTag{
					Type: api.MdlDynDrawTagType_mdl_draw_tag_goods,
					Item: &api.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Url:         picTag.Url,
						ItemId:      picTag.ItemID,
						Source:      picTag.Source,
						SchemaUrl:   picTag.SchemaURL,
					},
				})
			case mdlv2.DrawTagTypeUser:
				i.Tags = append(i.Tags, &api.MdlDynDrawTag{
					Type: api.MdlDynDrawTagType_mdl_draw_tag_user,
					Item: &api.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Mid:         picTag.Mid,
						Url:         model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(picTag.Mid, 10), nil),
					},
				})
			case mdlv2.DrawTagTypeTopic:
				var topicURL string
				topicInfos, _ := dynCtx.Dyn.GetTopicInfo()
				for _, topic := range topicInfos {
					if topic != nil && topic.TopicName == picTag.Text {
						topicURL = topic.TopicLink
						break
					}
				}
				i.Tags = append(i.Tags, &api.MdlDynDrawTag{
					Type: api.MdlDynDrawTagType_mdl_draw_tag_topic,
					Item: &api.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Tid:         picTag.Tid,
						Url:         topicURL,
					},
				})
			case mdlv2.DrawTagTypeLBS:
				var (
					lbs *mdlv2.DrawTagLBS
					uri string
				)
				if err := json.Unmarshal([]byte(picTag.Poi), &lbs); err != nil {
					xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "date_invalid")
					continue
				}
				if lbs != nil && lbs.PoiInfo != nil && lbs.PoiInfo.Location != nil {
					uri = fmt.Sprintf(model.LBSURI, lbs.PoiInfo.Poi, lbs.PoiInfo.Type, lbs.PoiInfo.Location.Lat, lbs.PoiInfo.Location.Lng, url.QueryEscape(lbs.PoiInfo.Title), url.QueryEscape(lbs.PoiInfo.Address))
				}
				i.Tags = append(i.Tags, &api.MdlDynDrawTag{
					Type: api.MdlDynDrawTagType_mdl_draw_tag_lbs,
					Item: &api.MdlDynDrawTagItem{
						X:           picTag.X,
						Y:           picTag.Y,
						Text:        picTag.Text,
						Orientation: picTag.Orientation,
						Poi:         picTag.Poi,
						Url:         uri,
					},
				})
			default:
				xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "date_invalid")
			}
		}
		card.Items = append(card.Items, i)
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_draw,
				ModuleItem: &api.ModuleDynamic_DynDraw{DynDraw: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	// 详情页上图下文
	if general.Config.IsDetailDrawFirst() {
		// 先保存图文模块的idx
		drawIdx := len(dynCtx.DynamicItem.Modules) - 1
		mdlDraw := dynCtx.DynamicItem.Modules[drawIdx]
		// 查找author模块的idx 然后把图文挪到author之前
		authorIdx := -1
		for i, m := range dynCtx.DynamicItem.Modules {
			if m.ModuleType == api.DynModuleType_module_author {
				authorIdx = i
				break
			}
		}
		// 确认找到的情况下插进去
		if authorIdx != -1 {
			copy(dynCtx.DynamicItem.Modules[authorIdx+1:], dynCtx.DynamicItem.Modules[authorIdx:])
			dynCtx.DynamicItem.Modules[authorIdx] = mdlDraw
		}
	}
	return nil
}

func (s *Service) dynCardArticle(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	article, _ := dynCtx.GetResArticle(dynCtx.Dyn.Rid)
	card := &api.MdlDynArticle{
		Id:         article.ActID,
		Title:      s.getTitle(article.Title, dynCtx),
		Desc:       article.Summary,
		Covers:     article.ImageURLs,
		Label:      model.StatString(article.Stats.View, "阅读"),
		TemplateID: article.TemplateID,
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_article,
				ModuleItem: &api.ModuleDynamic_DynArticle{DynArticle: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardMusic(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	music, _ := dynCtx.GetResMusic(dynCtx.Dyn.Rid)
	card := &api.MdlDynMusic{
		Id:     music.ID,
		UpId:   music.UpId,
		Title:  s.getTitle(music.Title, dynCtx),
		Cover:  music.Cover,
		Upper:  music.Upper,
		Label1: music.TypeInfo,
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_music,
				ModuleItem: &api.ModuleDynamic_DynMusic{DynMusic: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardCommon(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	common, _ := dynCtx.GetResCommon(dynCtx.Dyn.Rid)
	card := &api.MdlDynCommon{
		Oid:      common.Sketch.BizID,
		Uri:      common.Sketch.TagURL,
		Title:    s.getTitle(common.Sketch.Title, dynCtx),
		Desc:     common.Sketch.DescText,
		Cover:    common.Sketch.CoverURL,
		Label:    common.Sketch.Text,
		BizType:  int32(common.Sketch.BizType),
		SketchID: common.Sketch.SketchID,
		Style:    api.MdlDynCommonType_mdl_dyn_common_vertica,
	}
	if dynCtx.Dyn.IsCommonSquare() {
		card.Style = api.MdlDynCommonType_mdl_dyn_common_square
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynCommonLabel, &feature.OriginResutl{
			MobiApp: general.GetMobiApp(),
			Device:  general.GetDevice(),
			Build:   general.GetBuild(),
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynCommonLabelIOS) ||
				(general.IsAndroidPick() && general.GetBuild() <= s.c.BuildLimit.DynCommonLabelAndroid) || general.IsAndroidHD() || general.IsPad() || general.IsPadHD()}) {
			card.Label = "" // 兼容逻辑：iOS多读了字段
		}
	}
	var tags []*mdlv2.DynamicCommonCardTags
	if err := json.Unmarshal(common.Sketch.Tags, &tags); err != nil {
		xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "date_faild")
	}
	for _, tag := range tags {
		if tag == nil || tag.Name == "" {
			continue
		}
		var color = tag.Color
		if !strings.Contains(color, "#") {
			color = fmt.Sprintf("#%s", color)
		}
		card.Badge = append(card.Badge, &api.VideoBadge{
			Text:             tag.Name,
			TextColor:        s.c.Resource.Others.ModuleDynamicCommonBadge.TextColor,
			TextColorNight:   s.c.Resource.Others.ModuleDynamicCommonBadge.TextColorNight,
			BgColor:          color,
			BgColorNight:     color,
			BorderColor:      color,
			BorderColorNight: color,
			BgStyle:          s.c.Resource.Others.ModuleDynamicCommonBadge.BgStyle,
		})
	}
	// 按钮
	if button := common.Sketch.Button; button != nil {
		if button.JumpStyle != nil {
			card.Button = &api.AdditionalButton{
				Type:    api.AddButtonType_bt_jump,
				JumpUrl: button.JumpURL,
				JumpStyle: &api.AdditionalButtonStyle{
					Icon:    button.JumpStyle.Icon,
					Text:    button.JumpStyle.Text,
					BgStyle: api.AddButtonBgStyle_fill,
					Disable: api.DisableState_highlight,
				},
			}
			if button.Status == mdlv2.AttachButtonStatusCheck {
				card.Button = &api.AdditionalButton{
					Type:   api.AddButtonType_bt_button,
					Status: api.AdditionalButtonStatus_check,
					Check: &api.AdditionalButtonStyle{
						Icon:    button.JumpStyle.Icon,
						Text:    button.JumpStyle.Text,
						BgStyle: api.AddButtonBgStyle_gray,
						Disable: api.DisableState_gary,
					},
				}
			}
		}
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_common,
				ModuleItem: &api.ModuleDynamic_DynCommon{DynCommon: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardBatch(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	const (
		_batchBizType = 311
	)
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	batch, ok := dynCtx.ResBatch[dynCtx.Dyn.Rid]
	if !ok {
		return nil
	}
	card := &api.MdlDynCommon{
		Oid:     batch.ID,
		Uri:     batch.JumpURL,
		Title:   batch.Title,
		Desc:    fmt.Sprintf("%s %s", batch.Area, batch.Style),
		Label:   fmt.Sprintf("%s：%s", batch.FromBatchFinish(), batch.UpdateFreq),
		Cover:   batch.Cover,
		BizType: _batchBizType,
		Style:   api.MdlDynCommonType_mdl_dyn_common_vertica,
		Badge:   []*api.VideoBadge{mdlv2.BadgeStyleFrom(mdlv2.BgColorPink, batch.FromBatchPay())},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_common,
				ModuleItem: &api.ModuleDynamic_DynCommon{DynCommon: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardAD(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if dynCtx.Dyn.PassThrough == nil {
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	// 广告AdSourceContent特殊判断
	if dynCtx.Dyn.PassThrough.AdSourceContent == nil || dynCtx.Dyn.PassThrough.AdSourceContent.Size() == 0 {
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	moduleAd := &api.Module_ModuleAd{
		ModuleAd: &api.ModuleAd{
			SourceContent: dynCtx.Dyn.PassThrough.AdSourceContent,
			ModuleAuthor:  s.authorUser(c, dynCtx.Dyn.PassThrough.AdverMid, dynCtx, general),
		},
	}
	if moduleAd.ModuleAd.ModuleAuthor != nil {
		moduleAd.ModuleAd.ModuleAuthor.TpList = s.threePointAd(c, dynCtx, general)
	}
	if dynCtx.Dyn.PassThrough != nil {
		moduleAd.ModuleAd.AdContentType = dynCtx.Dyn.PassThrough.AdContentType
	}
	// 起飞广告
	if dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.AdContentType == _adContentFly && dynCtx.Dyn.PassThrough.AdAvid > 0 {
		ap, ok := dynCtx.GetArchive(dynCtx.Dyn.PassThrough.AdAvid)
		if ok && ap.Arc.IsNormal() {
			moduleAd.ModuleAd.CoverLeftText_1 = s.videoDuration(ap.Arc.Duration)
			moduleAd.ModuleAd.CoverLeftText_2 = fmt.Sprintf("%s观看", s.numTransfer(int(ap.Arc.Stat.View)))
			moduleAd.ModuleAd.CoverLeftText_3 = fmt.Sprintf("%s弹幕", s.numTransfer(int(ap.Arc.Stat.Danmaku)))
		}
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_ad,
		ModuleItem: moduleAd,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardADShell(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	res, ok := dynCtx.ResAD[dynCtx.Dyn.Rid]
	if !ok || res == nil {
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	// 广告SourceContent特殊判断
	if res.SourceContent == nil || res.SourceContent.Size() == 0 {
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	moduleAd := &api.Module_ModuleAd{
		ModuleAd: &api.ModuleAd{
			SourceContent: res.SourceContent,
			ModuleAuthor:  s.authorUser(c, res.AdverMid, dynCtx, general),
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_ad,
		ModuleItem: moduleAd,
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardApplet(_ context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	applet, _ := dynCtx.GetResApple(dynCtx.Dyn.Rid)
	card := &api.MdlDynApplet{
		Id:       applet.GetRid(),
		Uri:      applet.GetTargetUrl(),
		Title:    s.getTitle(applet.GetTitle(), dynCtx),
		SubTitle: applet.GetDesc(),
		Cover:    applet.GetCover(),
	}
	if applet.GetTag() != "" {
		appletLabel := new(mdlv2.AppletLabel)
		if err := json.Unmarshal([]byte(applet.GetTag()), &appletLabel); err != nil {
			xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "date_faild")
		} else {
			card.Icon = appletLabel.Icon
			card.Label = appletLabel.ProgramText
			card.ButtonTitle = appletLabel.JumpText
		}
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_applet,
				ModuleItem: &api.ModuleDynamic_DynApplet{DynApplet: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardSubscription(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	sub, _ := dynCtx.GetResSub(dynCtx.Dyn.Rid)
	card := &api.MdlDynSubscription{
		Id:    sub.OID,
		Uri:   sub.JumpURL,
		Title: s.getTitle(sub.Title, dynCtx),
		Cover: sub.Icon,
		Badge: &api.VideoBadge{
			Text:             sub.TagName,
			TextColor:        s.c.Resource.Others.ModuleDynamicSubscriptionBadge.TextColor,
			TextColorNight:   s.c.Resource.Others.ModuleDynamicSubscriptionBadge.TextColorNight,
			BgColor:          sub.TagColor,
			BgColorNight:     sub.TagColor,
			BorderColor:      sub.TagColor,
			BorderColorNight: sub.TagColor,
			BgStyle:          s.c.Resource.Others.ModuleDynamicSubscriptionBadge.BgStyle,
		},
		Tips: sub.Tips,
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_subscription,
				ModuleItem: &api.ModuleDynamic_DynSubscription{DynSubscription: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

type pendantMeta struct {
	Name     string
	Icon     string
	ID       int64
	Priority int64
}

var (
	_allPendentMetas = []*pendantMeta{
		{
			Name:     "生日会",
			Icon:     "https://i0.hdslb.com/bfs/archive/8f429757e1c8fb7c8dbab669f404711caff8cf74.png",
			ID:       389,
			Priority: 1,
		},
		{
			Name:     "生日会",
			Icon:     "https://i0.hdslb.com/bfs/archive/8f429757e1c8fb7c8dbab669f404711caff8cf74.png",
			ID:       387,
			Priority: 1,
		},
		{
			Name:     "生日会",
			Icon:     "https://i0.hdslb.com/bfs/archive/8f429757e1c8fb7c8dbab669f404711caff8cf74.png",
			ID:       863,
			Priority: 1,
		},
		{
			Name:     "红包直播",
			Icon:     "https://i0.hdslb.com/bfs/archive/bbed1abeec6a65e16a044c10bf2e2e925debc911.png",
			ID:       math.MinInt64,
			Priority: 2,
		},
		{
			Name:     "天选时刻",
			Icon:     "https://i0.hdslb.com/bfs/archive/e97a6b9ba4ab5ba3583128483490ea8ce1d533e1.png",
			ID:       504,
			Priority: 4,
		},
		{
			Name:     "PK对决",
			Icon:     "https://i0.hdslb.com/bfs/archive/b2b64c3191e86cb9f26d4b7f42fbc3e1911bff72.png",
			ID:       math.MinInt64,
			Priority: 5,
		},
		{
			Name:     "付费直播-大航海",
			Icon:     "http://i0.hdslb.com/bfs/feed-admin/b4e39a36de8b3c86d4fd05486c9ddd2f59554192.png",
			ID:       1145,
			Priority: 3,
		},
	}
)
var pendantMetas = map[int64]*pendantMeta{}

func init() {
	for _, m := range _allPendentMetas {
		pendantMetas[m.ID] = m
	}
}

func constructLivePendent(in *livexroomfeed.LivePlayInfo) *api.LivePendant {
	if in.Pendants == nil {
		return nil
	}
	type livePendant struct {
		*api.LivePendant
		priority int64
	}

	allPendants := make([]*livePendant, 0, len(in.Pendants.List))
	for _, pp := range in.Pendants.List {
		for _, p := range pp.List {
			pm, ok := pendantMetas[p.PendantId]
			if !ok {
				continue
			}
			lp := &livePendant{
				LivePendant: &api.LivePendant{
					Text:      pm.Name,
					Icon:      pm.Icon,
					PendantId: p.PendantId,
				},
				priority: pm.Priority,
			}
			allPendants = append(allPendants, lp)
		}
	}
	if len(allPendants) <= 0 {
		return nil
	}
	sort.Slice(allPendants, func(i, j int) bool {
		return allPendants[i].priority < allPendants[j].priority
	})
	return allPendants[0].LivePendant
}

var (
	livePendantFlag = ab.Int("live_pendant_v2", "livePendant", 0)
)

func (s *Service) buvidABTest(ctx context.Context, flag *ab.IntFlag) bool {
	dev, ok := device.FromContext(ctx)
	if !ok {
		return false
	}
	t, ok := ab.FromContext(ctx)
	if !ok {
		return false
	}
	t.Add(ab.KVString("buvid", dev.Buvid))
	exp := flag.Value(t)
	return exp == 1
}

func (s *Service) dynCardLiveRcmd(ctx context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	livercmd, _ := dynCtx.GetResLiveRcmd(dynCtx.Dyn.Rid)
	var (
		reserveType    api.ReserveType
		reserveLiveURL string
	)
	if dynCtx.Dyn.Property != nil && (dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE || dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_HISTORY) {
		// UP主预约是否召回
		reserveType = api.ReserveType_reserve_recall
		// 处理预约跳转地址 TODO 临时方案方案有接口额外调用且涉及对透传数据的修改,严重不合理,推进修改
		if entryLive, ok := dynCtx.GetResEntryLive(dynCtx.Dyn.UID); ok {
			reserveLiveURL = entryLive.JumpUrl["dt_top_live_card"]
		}
	}
	// room_paid_type 值的含义: 1.大航海付费直播间
	if livercmd.LivePlayInfo != nil && reserveLiveURL != "" && livercmd.LivePlayInfo.RoomPaidType != 1 {
		livercmd.LivePlayInfo.Link = reserveLiveURL
	}
	content, err := json.Marshal(livercmd)
	if err != nil {
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	if string(content) == "" {
		xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "date_invalid")
	}
	card := &api.MdlDynLiveRcmd{
		Content:     string(content),
		ReserveType: reserveType,
	}
	if s.buvidABTest(ctx, livePendantFlag) {
		card.Pendant = constructLivePendent(livercmd.LivePlayInfo)
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_live_rcmd,
				ModuleItem: &api.ModuleDynamic_DynLiveRcmd{DynLiveRcmd: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardUGCSeason(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	ugcSeason, _ := dynCtx.GetResUGCSeason(dynCtx.Dyn.UID)
	ap, _ := dynCtx.GetArchive(dynCtx.Dyn.Rid)
	var archive = ap.Arc
	card := &api.MdlDynUGCSeason{
		Id:              dynCtx.Dyn.Rid,
		Avid:            archive.Aid,
		Cid:             archive.FirstCid,
		Title:           s.getTitle(archive.Title, dynCtx),
		Cover:           archive.Pic,
		CoverLeftText_1: s.videoDuration(archive.Duration),
		CoverLeftText_2: fmt.Sprintf("%s观看", s.numTransfer(int(ugcSeason.Stat.View))),
		CoverLeftText_3: fmt.Sprintf("%s弹幕", s.numTransfer(int(ugcSeason.Stat.Danmaku))),
		Dimension: &api.Dimension{
			Height: archive.Dimension.Height,
			Width:  archive.Dimension.Width,
			Rotate: archive.Dimension.Rotate,
		},
		Duration: archive.Duration,
	}
	var (
		playurl *arcgrpc.PlayerInfo
		ok      bool
	)
	if playurl, ok = ap.PlayerInfo[ap.DefaultPlayerCid]; !ok {
		playurl = ap.PlayerInfo[ap.Arc.FirstCid]
	}
	if playurl != nil && playurl.PlayerExtra != nil && playurl.PlayerExtra.Dimension != nil {
		card.Cid = playurl.PlayerExtra.Cid
		card.Dimension.Height = playurl.PlayerExtra.Dimension.Height
		card.Dimension.Width = playurl.PlayerExtra.Dimension.Width
		card.Dimension.Rotate = playurl.PlayerExtra.Dimension.Rotate
	}
	if g, ok := dynCtx.Grayscale[s.c.Grayscale.ShowPlayIcon.Key]; ok {
		switch g {
		case 1:
			card.PlayIcon = s.c.Resource.Icon.ModuleDynamicPlayIcon
		}
	}
	card.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, archive.FirstCid, true))
	card.CanPlay = mdlv2.CanPlay(archive.Rights.Autoplay)
	// 付费合集
	if mdlv2.PayAttrVal(archive) {
		card.Badge = append(card.Badge, mdlv2.PayBadge)
	}
	dynamic := &api.ModuleDynamic{
		Type: api.ModuleDynamicType_mdl_dyn_ugc_season,
		ModuleItem: &api.ModuleDynamic_DynUgcSeason{
			DynUgcSeason: card,
		},
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: dynamic,
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) dynCardSubNew(_ context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	subNew, _ := dynCtx.GetResSubNew(dynCtx.Dyn.Rid)
	card := &api.MdlDynSubscriptionNew{}
	switch subNew.Type {
	case submdl.TunnelTypeLive:
		card.Style = api.MdlDynSubscriptionNewStyle_mdl_dyn_subscription_new_style_live
		card.Item = &api.MdlDynSubscriptionNew_DynLiveRcmd{
			DynLiveRcmd: &api.MdlDynLiveRcmd{
				Content: subNew.LiveInfo,
			},
		}
	case submdl.TunnelTypeDraw:
		var sub *submdl.Subscription
		if err := json.Unmarshal([]byte(subNew.ImageInfo), &sub); err != nil {
			dynCtx.Interim.IsPassCard = true
			return nil
		}
		card.Style = api.MdlDynSubscriptionNewStyle_mdl_dyn_subscription_new_style_draw
		card.Item = &api.MdlDynSubscriptionNew_DynSubscription{
			DynSubscription: &api.MdlDynSubscription{
				Id:    sub.OID,
				Uri:   sub.JumpURL,
				Title: s.getTitle(sub.Title, dynCtx),
				Cover: sub.Icon,
				Badge: &api.VideoBadge{
					Text:             sub.TagName,
					TextColor:        s.c.Resource.Others.ModuleDynamicSubscriptionBadge.TextColor,
					TextColorNight:   s.c.Resource.Others.ModuleDynamicSubscriptionBadge.TextColorNight,
					BgColor:          sub.TagColor,
					BgColorNight:     sub.TagColorNight,
					BorderColor:      sub.TagColor,
					BorderColorNight: sub.TagColorNight,
					BgStyle:          s.c.Resource.Others.ModuleDynamicSubscriptionBadge.BgStyle,
				},
				Tips: sub.Tips,
			},
		}
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type:       api.ModuleDynamicType_mdl_dyn_subscription_new,
				ModuleItem: &api.ModuleDynamic_DynSubscriptionNew{DynSubscriptionNew: card},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (s *Service) videoDuration(du int64) string {
	hour := du / mdlv2.PerHour
	du = du % mdlv2.PerHour
	minute := du / mdlv2.PerMinute
	second := du % mdlv2.PerMinute
	if hour != 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
	}
	return fmt.Sprintf("%02d:%02d", minute, second)
}

// nolint:gomnd
func (s *Service) numTransfer(num int) string {
	if num < 10000 {
		return strconv.Itoa(num)
	}
	integer := num / 10000
	decimals := num % 10000
	decimals = decimals / 1000
	return fmt.Sprintf("%d.%d万", integer, decimals)
}

func (s *Service) dynCardFake(_ context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	var module *api.Module
	if dynCtx.Dyn.IsAv() {
		card := &api.MdlDynArchive{
			Cover:           dynCtx.Dyn.FakeCover,
			CoverLeftText_1: s.videoDuration(dynCtx.Dyn.Duration),
			CoverLeftText_2: "0观看",
			CoverLeftText_3: "0弹幕",
		}
		module = &api.Module{
			ModuleType: api.DynModuleType_module_dynamic,
			ModuleItem: &api.Module_ModuleDynamic{
				ModuleDynamic: &api.ModuleDynamic{
					Type:       api.ModuleDynamicType_mdl_dyn_archive,
					ModuleItem: &api.ModuleDynamic_DynArchive{DynArchive: card},
				},
			},
		}
	} else if dynCtx.Dyn.IsDraw() {
		card := &api.MdlDynDraw{}
		for _, images := range dynCtx.Dyn.FakeImages {
			if images == nil {
				continue
			}
			i := &api.MdlDynDrawItem{
				Src:    images.ImgSrc,
				Width:  images.ImgWidth,
				Height: images.ImgHeight,
				Size_:  images.ImgSize,
			}
			card.Items = append(card.Items, i)
		}
		module = &api.Module{
			ModuleType: api.DynModuleType_module_dynamic,
			ModuleItem: &api.Module_ModuleDynamic{
				ModuleDynamic: &api.ModuleDynamic{
					Type:       api.ModuleDynamicType_mdl_dyn_draw,
					ModuleItem: &api.ModuleDynamic_DynDraw{DynDraw: card},
				},
			},
		}
	} else {
		return nil
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

// nolint:gocognit
func (s *Service) dynCardPremiere(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynPropertyIOS || general.IsAndroidPick() && general.GetBuild() < s.c.BuildLimit.DynPropertyAndroid || general.IsPad() || general.IsPadHD() || general.IsAndroidHD() {
		return nil
	}
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	for _, v := range dynCtx.Dyn.AttachCardInfos {
		// nolint:exhaustive
		switch v.CardType {
		case dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE:
			up, ok := dynCtx.ResUpActRelationInfo[v.Rid]
			if !ok {
				continue
			}
			// 非首映的直接抛弃
			if up.Type != activitygrpc.UpActReserveRelationType_Premiere {
				continue
			}
			aid, _ := strconv.ParseInt(up.Oid, 10, 64)
			ap, ok := dynCtx.GetArchive(aid)
			if !ok {
				continue
			}
			var archive = ap.Arc
			card := &api.MdlDynArchive{
				Cover:           archive.Pic,
				CoverLeftText_1: s.videoDuration(archive.Duration),
				CoverLeftText_2: fmt.Sprintf("%s观看", s.numTransfer(int(archive.Stat.View))),
				CoverLeftText_3: fmt.Sprintf("%s弹幕", s.numTransfer(int(archive.Stat.Danmaku))),
				Avid:            archive.Aid,
				Cid:             archive.FirstCid,
				MediaType:       api.MediaType_MediaTypeUGC,
				Dimension: &api.Dimension{
					Height: archive.Dimension.Height,
					Width:  archive.Dimension.Width,
					Rotate: archive.Dimension.Rotate,
				},
				Duration:     archive.Duration,
				View:         archive.Stat.View,
				PremiereCard: true,
				PlayIcon:     s.c.Resource.Icon.ModuleDynamicPlayIcon,
			}
			card.Bvid, _ = bvid.AvToBv(archive.Aid)
			playurl, ok := ap.PlayerInfo[dynCtx.Interim.CID]
			if !ok {
				if playurl, ok = ap.PlayerInfo[ap.DefaultPlayerCid]; !ok {
					playurl = ap.PlayerInfo[ap.Arc.FirstCid]
				}
			}
			if playurl != nil && playurl.PlayerExtra != nil && playurl.PlayerExtra.Dimension != nil {
				card.Cid = playurl.PlayerExtra.Cid
				card.Dimension.Height = playurl.PlayerExtra.Dimension.Height
				card.Dimension.Width = playurl.PlayerExtra.Dimension.Width
				card.Dimension.Rotate = playurl.PlayerExtra.Dimension.Rotate
			}
			card.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))
			// PGC特殊逻辑
			if archive.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && archive.RedirectURL != "" {
				card.Uri = archive.RedirectURL
				card.IsPGC = true
				if playurl, ok = ap.PlayerInfo[ap.DefaultPlayerCid]; ok && playurl.PlayerExtra != nil && playurl.PlayerExtra.PgcPlayerExtra != nil {
					if playurl.PlayerExtra.PgcPlayerExtra.IsPreview == 1 {
						card.IsPreview = true
					}
					card.EpisodeId = playurl.PlayerExtra.PgcPlayerExtra.EpisodeId
					card.SubType = playurl.PlayerExtra.PgcPlayerExtra.SubType
					card.PgcSeasonId = playurl.PlayerExtra.PgcPlayerExtra.PgcSeasonId
				}
			}
			// 小视频特殊处理
			card.Stype = mdlv2.GetArchiveSType(dynCtx.Dyn.SType)
			if card.Stype == api.VideoType_video_type_story {
				if !feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynStory, &feature.OriginResutl{
					BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynStoryIOS) ||
						(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynStoryAndroid)}) {
					card.Stype = api.VideoType_video_type_dynamic
				}
			}
			if card.Stype == api.VideoType_video_type_dynamic || card.Stype == api.VideoType_video_type_story {
				card.Title = ""
				card.IsPGC = false
			}
			card.CanPlay = mdlv2.CanPlay(archive.Rights.Autoplay)
			// 首映状态
			if archive.Premiere == nil {
				continue
			}
			countInfo, ok := dynCtx.ResPlayUrlCount[ap.Arc.Aid]
			switch archive.Premiere.State {
			case arcgrpc.PremiereState_premiere_before: // 首映前
				card.CoverLeftText_1 = ""
				card.CoverLeftText_3 = ""
				if ok {
					card.CoverLeftText_2 = model.UpStatString(countInfo.Count[_play_online_total], "人在等待")
				}
				card.CanPlay = false
			case arcgrpc.PremiereState_premiere_in: // 首映中
				if dynCtx.From != _handleTypeForward {
					// 首映结中且非转发卡，直接抛弃整个动态卡
					dynCtx.Interim.IsPassCard = true
				}
				card.CoverLeftText_2 = ""
				card.CoverLeftText_3 = ""
				card.Uri = s.inArchivePremiereArg()(card.Uri)
				if ok {
					card.BadgeCategory = append(card.BadgeCategory, mdlv2.BadgeStyleFrom(mdlv2.BgColorGray, fmt.Sprintf("%d人在线", countInfo.Count[_play_online_total])))
					card.ShowPremiereBadge = true
				}
			default: // 首映结束
				if dynCtx.From != _handleTypeForward {
					// 首映结束且非转发卡，直接抛弃整个动态卡
					dynCtx.Interim.IsPassCard = true
				}
				card.CoverLeftText_2 = model.UpStatString(int64(archive.Stat.View), "观看")
				card.CoverLeftText_3 = model.UpStatString(int64(archive.Stat.Danmaku), "弹幕")
				card.PremiereCard = false
				card.Title = s.getTitle(archive.Title, dynCtx)
			}
			dynamic := &api.ModuleDynamic{
				Type: api.ModuleDynamicType_mdl_dyn_archive,
				ModuleItem: &api.ModuleDynamic_DynArchive{
					DynArchive: card,
				},
			}
			module := &api.Module{
				ModuleType: api.DynModuleType_module_dynamic,
				ModuleItem: &api.Module_ModuleDynamic{
					ModuleDynamic: dynamic,
				},
			}
			// card ext
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))
			if ap.GetArc().GetPremiere().GetState() == arcgrpc.PremiereState_premiere_in {
				dynCtx.DynamicItem.Extend.CardUrl = s.inArchivePremiereArg()(dynCtx.DynamicItem.Extend.CardUrl)
			}
			dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, false))
			dynCtx.DynamicItem.Extend.OrigImgUrl = archive.Pic
			dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, archive.Title, dynCtx, general)
			dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
		}
	}
	return nil
}

func (s *Service) dynCardNewTopicSet(c context.Context, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	tps := dynCtx.GetResNewTopicSet()
	if tps == nil {
		// 主动丢弃卡片
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	topicSet := &api.MdlDynTopicSet{
		MoreBtn: &api.IconButton{
			IconTail: s.c.Resource.Icon.DynMixTopicSquareMore,
			Text:     "查看更多话题",
			JumpUri:  tps.SetInfo.BasicInfo.GetJumpUrl(),
		},
		TopicSetId: tps.SetInfo.GetBasicInfo().GetSetId(),
		PushId:     dynCtx.Dyn.GetNewTopicSetPushId(),
	}
	for _, t := range tps.TopicList.Topics {
		topicSet.Topics = append(topicSet.Topics, &api.TopicItem{
			TopicId:   t.GetId(),
			TopicName: t.GetName(),
			Url:       t.GetJumpUrl(),
		})
	}
	module := &api.Module{
		ModuleType: api.DynModuleType_module_dynamic,
		ModuleItem: &api.Module_ModuleDynamic{
			ModuleDynamic: &api.ModuleDynamic{
				Type: api.ModuleDynamicType_mdl_dyn_topic_set,
				ModuleItem: &api.ModuleDynamic_DynTopicSet{
					DynTopicSet: topicSet,
				},
			},
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)

	return nil
}
