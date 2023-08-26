package dynamicV2

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go-common/library/log"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"go-common/library/net/trace"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	livemdl "go-gateway/app/app-svr/app-dynamic/interface/model/live"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	submdl "go-gateway/app/app-svr/app-dynamic/interface/model/subscription"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
)

const (
	_ShareType  = "3"
	_shareScene = "dynamic"
)

// nolint:gocognit
func (s *Service) base(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	// 初始化公共值
	dynCtx.DynamicItem.Extend.DynIdStr = strconv.FormatInt(dynCtx.Dyn.DynamicID, 10)
	dynCtx.DynamicItem.Extend.BusinessId = strconv.FormatInt(dynCtx.Dyn.Rid, 10)
	dynCtx.DynamicItem.Extend.RType = dynCtx.Dyn.RType
	dynCtx.DynamicItem.Extend.DynType = dynCtx.Dyn.Type
	dynCtx.DynamicItem.Extend.ShareType = _ShareType
	dynCtx.DynamicItem.Extend.ShareScene = _shareScene
	dynCtx.DynamicItem.Extend.IsFastShare = true
	dynCtx.DynamicItem.Extend.TrackId = dynCtx.Dyn.TrackID
	if !dynCtx.Dyn.IsAD() && dynCtx.Dyn.PassThrough != nil {
		dynCtx.DynamicItem.Extend.SourceContent = dynCtx.Dyn.PassThrough.AdSourceContent
	}
	if dynCtx.From != _handleTypeView {
		dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoDyn, strconv.FormatInt(dynCtx.Dyn.DynamicID, 10), model.SuffixHandler(fmt.Sprintf("cardType=%v&rid=%d", dynCtx.Dyn.Type, dynCtx.Dyn.Rid)))
		// 如果当前是广告卡url里面增加is_ad=1
		if dynCtx.Dyn.IsAD() {
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoURL, dynCtx.DynamicItem.Extend.CardUrl, model.SuffixHandler("is_ad=1"))
		}
	}
	// 预处理帮推
	if dynCtx.Dyn.IsWord() || dynCtx.Dyn.IsDraw() {
		if dynCtx.ResAttachedPromo != nil {
			if topicID, ok := dynCtx.ResAttachedPromo[dynCtx.Dyn.DynamicID]; ok {
				if act, ok := dynCtx.ResActivity[topicID]; ok {
					dynCtx.Interim.PromoURI = fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%v?activity_from=dt_dynamic&dynamic_id=%v", act.ID, dynCtx.Dyn.DynamicID)
					dynCtx.Interim.IsPassExtendGameTopic = true
				}
			}
		}
	}
	dynCtx.DynamicItem.Extend.Reply = &api.ExtendReply{
		Uri: model.FillURI(model.GotoDyn, strconv.FormatInt(dynCtx.Dyn.DynamicID, 10), model.SuffixHandler(fmt.Sprintf("cardType=%v&rid=%d", dynCtx.Dyn.Type, dynCtx.Dyn.Rid))),
		Params: []*api.ExtendReplyParam{
			{
				Key:   "comment_on",
				Value: "1",
			},
		},
	}
	/*
		分类型兼容：卡片类型、原卡类型、默认跳转、评论跳转、评论参数
	*/
	// 转发卡
	switch {
	case dynCtx.Dyn.IsForward(): // 转发卡
		dynCtx.DynamicItem.CardType = api.DynamicType_forward
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_forward
		dynCtx.Interim.DynTypeShell = dynCtx.Dyn.Type
		dynCtx.Interim.ShellRID = dynCtx.Dyn.Rid
		if dynCtx.Dyn.Origin != nil {
			dynCtx.Interim.DynTypeKernel = dynCtx.Dyn.Origin.Type
			dynCtx.Interim.KernelRID = dynCtx.Dyn.Origin.Rid
			if !dynCtx.Dyn.Origin.Visible {
				dynCtx.Interim.ForwardOrigFaild = true
				xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "visible_false")
				log.Warn("module error mid(%v) dynid(%v) base vidible false", general.Mid, dynCtx.Dyn.DynamicID)
			}
		} else {
			dynCtx.Interim.ForwardOrigFaild = true // 源卡失效
			xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "dynamic", "origin_nil")
			log.Warn("module error mid(%v) dynid(%v) base origin nil", general.Mid, dynCtx.Dyn.DynamicID)
		}
		// 转发数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		// 转发文案extend.Desc、转发图标extend.OrigImgUrl、转发源动态文案extend.OrigDesc等在dynamic模块中单独处理
	case dynCtx.Dyn.IsAv(): // 视频卡
		dynCtx.DynamicItem.CardType = api.DynamicType_av
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_av
		ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid)
		if !ok || !ap.Arc.IsNormal() {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base av rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		var archive = ap.Arc
		// 付费合集 + Pad不展示付费卡、校园也不展示
		if mdlv2.PayAttrVal(archive) && ((general.IsPad() && general.GetBuild() < s.c.BuildLimit.DynFuFeiIOS || general.IsPadHD() && general.GetBuild() < s.c.BuildLimit.DynFuFeiIOSHD || general.IsAndroidHD() && general.GetBuild() < s.c.BuildLimit.DynFuFeiAndroidHD) || (dynCtx.From == _handleTypeSchool || dynCtx.From == _handleTypeSchoolTopicFeed)) {
			dynCtx.Interim.IsPassCard = true
			return nil
		}
		dynCtx.Interim.CID = archive.FirstCid
		// 新版走非首P逻辑
		if (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.NewPlayerIOS) || (general.IsAndroidPick() && general.GetBuild() >= s.c.BuildLimit.NewPlayerAndroid) {
			dynCtx.Interim.CID = ap.DefaultPlayerCid
			if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.VideoShare != nil {
				dynCtx.Interim.CID = dynCtx.Dyn.Extend.VideoShare.CID
			}
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转和评论
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))
			dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, false))
		}
		replyParam := new(api.ExtendReplyParam)
		replyParam.Key = "comment_on"
		replyParam.Value = "1"
		if archive.AttrVal(arcgrpc.AttrBitIsPGC) == arcgrpc.AttrYes && archive.RedirectURL != "" {
			dynCtx.DynamicItem.Extend.CardUrl = archive.GetRedirectURL()
			dynCtx.DynamicItem.Extend.Reply.Uri = archive.GetRedirectURL()
			replyParam.Key = "reply_id"
			replyParam.Value = "-1"
			if general.IsAndroidPick() {
				replyParam.Key = "comment_state"
				replyParam.Value = "1"
			}
		}
		dynCtx.DynamicItem.Extend.Reply.Params = []*api.ExtendReplyParam{replyParam}
		// 是否是横竖屏
		var (
			vertical bool
		)
		height := archive.Dimension.Height
		width := archive.Dimension.Width
		rotate := archive.Dimension.Rotate
		cid := dynCtx.GetArchiveAutoPlayCid(ap)
		playurl := ap.PlayerInfo[cid]
		if playurl != nil && playurl.PlayerExtra != nil && playurl.PlayerExtra.Dimension != nil {
			height = playurl.PlayerExtra.Dimension.Height
			width = playurl.PlayerExtra.Dimension.Width
			rotate = playurl.PlayerExtra.Dimension.Rotate
		}
		if rotate == 1 || width < height {
			vertical = true
		}
		// 动态视频跳联播页（快速消费页不跳转联播页）
		// Note: 只要是竖屏视频，新版本也会试验性直接跳story
		if (mdlv2.GetArchiveSType(dynCtx.Dyn.SType) == api.VideoType_video_type_dynamic || mdlv2.GetArchiveSType(dynCtx.Dyn.SType) == api.VideoType_video_type_story) ||
			// 分辨率为竖屏的视频，在动态关注流的场景下，也跳转story
			(vertical && (_verticalAvToStory[dynCtx.From] || _verticalAvToStory[dynCtx.ForwardFrom]) && general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, s.c.BuildLimit.DynStoryAndroidV2, s.c.BuildLimit.DynStoryIOSV2)) {
			// 只在粉双端进story，pad设备不进
			if !general.IsPadHD() && !general.IsAndroidHD() && !general.IsPad() {
				if dynCtx.From != _handleTypeAllPersonal && dynCtx.From != _handleTypeVideoPersonal {
					dynCtx.DynamicItem.Extend.CardUrl = fmt.Sprintf("bilibili://following/play_list?oid=%d&type=1&cid=%d", dynCtx.Dyn.DynamicID, archive.FirstCid)
				}
				// 新版本跳转到story页面里面
				if general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual, s.c.BuildLimit.DynStoryAndroid, s.c.BuildLimit.DynStoryIOS) {
					dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoStory, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))
					dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoStory, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, false))
					// story特殊逻辑
					// 无论是否是转卡，都用最上层的vmid
					vmid := dynCtx.Dyn.UID
					if dynCtx.Dyn.Forward != nil {
						vmid = dynCtx.Dyn.Forward.UID
					}
					dynamicScene := "dynamic"
					if general.DynFrom == _dynFromLive || general.DynFrom == _dynFromSpace ||
						// 来自于类似空间/转发在类似空间场景的，均按照空间场景处理
						isDynSpaceLike[dynCtx.From] || isDynSpaceLike[dynCtx.ForwardFrom] {
						dynamicScene = "dynamic_space"
					}
					dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoURL, dynCtx.DynamicItem.Extend.CardUrl, model.StoryHandler(ap, cid, general.IsIOSPlatform(), dynamicScene, vmid, dynCtx.Dyn.DynamicID, ""))
					dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoURL, dynCtx.DynamicItem.Extend.Reply.Uri, model.StoryHandler(ap, cid, general.IsIOSPlatform(), dynamicScene, vmid, dynCtx.Dyn.DynamicID, ""))
				}
			}
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		dynCtx.DynamicItem.Extend.OrigImgUrl = archive.Pic
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, archive.Title, dynCtx, general)
	case dynCtx.Dyn.IsPGC(): // PGC卡
		dynCtx.DynamicItem.CardType = api.DynamicType_pgc
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_pgc
		pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base PGC rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转和评论
			dynCtx.DynamicItem.Extend.CardUrl = pgc.Url
			dynCtx.DynamicItem.Extend.Reply.Uri = pgc.Url
		}
		replyParam := new(api.ExtendReplyParam)
		replyParam.Key = "reply_id"
		replyParam.Value = "-1"
		if general.IsAndroidPick() {
			replyParam.Key = "comment_state"
			replyParam.Value = "1"
		}
		dynCtx.DynamicItem.Extend.Reply.Params = []*api.ExtendReplyParam{replyParam}
		// 转发页面数据
		if pgc.Season != nil {
			dynCtx.Interim.UName = pgc.Season.Title
			dynCtx.DynamicItem.Extend.OrigName = pgc.Season.Title
		}
		dynCtx.DynamicItem.Extend.Uid = dynCtx.Dyn.UID
		dynCtx.DynamicItem.Extend.OrigImgUrl = pgc.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, pgc.CardShowTitle, dynCtx, general)
	case dynCtx.Dyn.IsCheeseBatch(): // 付费批次卡
		dynCtx.DynamicItem.CardType = api.DynamicType_courses
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_courses
		batch, ok := dynCtx.GetResCheeseBatch(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base cheese_batch rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转和评论
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoURL, batch.URL, model.BatchPlayHandler(batch))
			dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoURL, batch.URL, model.BatchPlayHandler(batch))
		}
		replyParam := new(api.ExtendReplyParam)
		replyParam.Key = "comment_on"
		replyParam.Value = "1"
		if general.IsAndroidPick() {
			replyParam.Key = "comment_state"
			replyParam.Value = "1"
		}
		dynCtx.DynamicItem.Extend.Reply.Params = []*api.ExtendReplyParam{replyParam}
		// 转发页面数据
		dynCtx.Interim.UName = batch.UpInfo.Name
		dynCtx.DynamicItem.Extend.OrigName = batch.UpInfo.Name
		dynCtx.DynamicItem.Extend.Uid = batch.UpID
		dynCtx.DynamicItem.Extend.OrigImgUrl = batch.NewEp.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, batch.Title, dynCtx, general)
	case dynCtx.Dyn.IsCheeseSeason(): // 付费系列卡
		dynCtx.DynamicItem.CardType = api.DynamicType_courses_season
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_courses_season
		season, ok := dynCtx.GetResCheeseSeason(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base cheese_season rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转和评论
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoURL, season.URL, model.SeasonPlayHandler(season))
			dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoURL, season.URL, model.SeasonPlayHandler(season))
		}
		replyParam := new(api.ExtendReplyParam)
		replyParam.Key = "comment_on"
		replyParam.Value = "1"
		if general.IsAndroidPick() {
			replyParam.Key = "comment_state"
			replyParam.Value = "1"
		}
		dynCtx.DynamicItem.Extend.Reply.Params = []*api.ExtendReplyParam{replyParam}
		// 转发页面数据
		dynCtx.Interim.UName = season.UpInfo.Name
		dynCtx.DynamicItem.Extend.Uid = season.UpID
		dynCtx.DynamicItem.Extend.OrigName = season.UpInfo.Name
		dynCtx.DynamicItem.Extend.OrigImgUrl = season.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, season.Title, dynCtx, general)
	case dynCtx.Dyn.IsCourUp(): // 课堂UP主推荐
		dynCtx.DynamicItem.CardType = api.DynamicType_cour_up
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_cour_up
		season, ok := dynCtx.GetResCheeseSeason(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base cheese_season rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转和评论
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoURL, season.URL, model.SeasonPlayHandler(season))
			dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoURL, season.URL, model.SeasonPlayHandler(season))
		}
		replyParam := new(api.ExtendReplyParam)
		replyParam.Key = "comment_on"
		replyParam.Value = "1"
		if general.IsAndroidPick() {
			replyParam.Key = "comment_state"
			replyParam.Value = "1"
		}
		dynCtx.DynamicItem.Extend.Reply.Params = []*api.ExtendReplyParam{replyParam}
		// 转发页面数据
		dynCtx.Interim.UName = season.UpInfo.Name
		dynCtx.DynamicItem.Extend.Uid = season.UpID
		dynCtx.DynamicItem.Extend.OrigName = season.UpInfo.Name
		dynCtx.DynamicItem.Extend.OrigImgUrl = season.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, season.Title, dynCtx, general)
	case dynCtx.Dyn.IsLive(): // 直播分享卡(转发)
		dynCtx.DynamicItem.CardType = api.DynamicType_live
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_live
		live, ok := dynCtx.GetResLive(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base live rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoLive, strconv.FormatInt(live.RoomId, 10), nil)
			// 如果当前直播接口返回的跳转地址，则直接用直播返回的链接
			if jumpurl, ok := live.JumpUrl["NONE"]; ok {
				dynCtx.DynamicItem.Extend.CardUrl = jumpurl
			}
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		dynCtx.DynamicItem.Extend.OrigImgUrl = live.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, live.Title, dynCtx, general)
	case dynCtx.Dyn.IsMedialist(): // 播单卡(转发)
		dynCtx.DynamicItem.CardType = api.DynamicType_medialist
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_medialist
		medialist, ok := dynCtx.GetResMedialist(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base medialist rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GOtoMedialist, strconv.FormatInt(medialist.ID, 10), nil)
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		dynCtx.DynamicItem.Extend.OrigImgUrl = medialist.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, medialist.Title, dynCtx, general)
	case dynCtx.Dyn.IsWord(): // 纯文字卡
		dynCtx.DynamicItem.CardType = api.DynamicType_word
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_word
		// 没有正文且没有附加卡则不展示
		if content := s.descriptionWord(dynCtx, general); content == "" && len(dynCtx.Dyn.AttachCardInfos) == 0 {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base word rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
			dynCtx.DynamicItem.Extend.OrigImgUrl = userInfo.Face
		}
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, s.descriptionWord(dynCtx, general), dynCtx, general)
	case dynCtx.Dyn.IsDraw(): // 图文卡
		dynCtx.DynamicItem.CardType = api.DynamicType_draw
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_draw
		draw, ok := dynCtx.GetResDraw(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base draw rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		for _, pic := range draw.Item.Pictures {
			if pic != nil && pic.ImgSrc != "" {
				dynCtx.DynamicItem.Extend.OrigImgUrl = pic.ImgSrc
				break
			}
		}
		var content = draw.Item.Description
		if draw.Item.Title != "" {
			content = draw.Item.Title
		}
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, content, dynCtx, general)
	case dynCtx.Dyn.IsArticle(): // 专栏卡
		dynCtx.DynamicItem.CardType = api.DynamicType_article
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_article
		article, ok := dynCtx.GetResArticle(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base article rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转和评论
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoArticle, strconv.FormatInt(article.ID, 10), nil)
			dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoArticle, strconv.FormatInt(article.ID, 10), nil)
		}
		replyParam := new(api.ExtendReplyParam)
		replyParam.Key = "reply_id"
		replyParam.Value = "-1"
		if general.IsAndroidPick() {
			replyParam.Key = "reply_id"
			replyParam.Value = "-2"
		}
		dynCtx.DynamicItem.Extend.Reply.Params = []*api.ExtendReplyParam{replyParam}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		for _, img := range article.ImageURLs {
			if img != "" {
				dynCtx.DynamicItem.Extend.OrigImgUrl = img
				break
			}
		}
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, article.Title, dynCtx, general)
	case dynCtx.Dyn.IsMusic(): // 音频卡
		dynCtx.DynamicItem.CardType = api.DynamicType_music
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_music
		music, ok := dynCtx.GetResMusic(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base music rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转和评论
			dynCtx.DynamicItem.Extend.CardUrl = music.Schema
			dynCtx.DynamicItem.Extend.Reply.Uri = music.Schema
		}
		replyParam := new(api.ExtendReplyParam)
		replyParam.Key = "from"
		replyParam.Value = "twitter"
		replyParam2 := new(api.ExtendReplyParam)
		replyParam2.Key = "tab_index"
		replyParam2.Value = "1"
		dynCtx.DynamicItem.Extend.Reply.Params = []*api.ExtendReplyParam{replyParam, replyParam2}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		dynCtx.DynamicItem.Extend.OrigImgUrl = music.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, music.Title, dynCtx, general)
	case dynCtx.Dyn.IsCommonSquare(): // 通用卡 方
		dynCtx.DynamicItem.CardType = api.DynamicType_common_square
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_common_square
		common, ok := dynCtx.GetResCommon(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base common square rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		dynCtx.DynamicItem.Extend.OrigImgUrl = common.Sketch.CoverURL
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, common.Sketch.Title, dynCtx, general)
	case dynCtx.Dyn.IsCommonVertical(): // 通用卡 竖
		dynCtx.DynamicItem.CardType = api.DynamicType_common_vertical
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_common_vertical
		common, ok := dynCtx.GetResCommon(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base common vertical rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		dynCtx.DynamicItem.Extend.OrigImgUrl = common.Sketch.CoverURL
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, common.Sketch.Title, dynCtx, general)
	case dynCtx.Dyn.IsAD(): // 广告卡
		dynCtx.DynamicItem.CardType = api.DynamicType_ad
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_ad
		// 广告物料特殊 dynamic内部单独判断
		// 广告起飞卡
		if ok := feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynAdFly, &feature.OriginResutl{
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynAdFlyIOS) ||
				(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynAdFlyAndroid)}); ok &&
			dynCtx.Dyn.PassThrough != nil && dynCtx.Dyn.PassThrough.AdContentType == _adContentFly && dynCtx.Dyn.PassThrough.AdAvid > 0 {
			ap, ok := dynCtx.GetArchive(dynCtx.Dyn.PassThrough.AdAvid)
			if ok && ap.Arc.IsNormal() {
				var archive = ap.Arc
				dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, false))
				if traceInfo, ok := trace.FromContext(c); ok {
					dynCtx.DynamicItem.Extend.Reply.Uri = model.FillReplyURL(dynCtx.DynamicItem.Extend.Reply.Uri, fmt.Sprintf("trackid=%s", traceInfo.TraceID()))
				}
				if dynCtx.Dyn.PassThrough.AdUrlExtra != "" {
					dynCtx.DynamicItem.Extend.Reply.Uri = model.FillReplyURL(dynCtx.DynamicItem.Extend.Reply.Uri, dynCtx.Dyn.PassThrough.AdUrlExtra)
				}
			}
			dynCtx.DynamicItem.Extend.Uid = dynCtx.Dyn.PassThrough.AdverMid
		}
	case dynCtx.Dyn.IsApplet(): // 小程序卡
		dynCtx.DynamicItem.CardType = api.DynamicType_applet
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_applet
		applet, ok := dynCtx.GetResApple(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base applet rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		dynCtx.DynamicItem.Extend.OrigImgUrl = applet.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, applet.Title, dynCtx, general)
	case dynCtx.Dyn.IsSubscription(): // 订阅卡
		dynCtx.DynamicItem.CardType = api.DynamicType_subscription
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_subscription
		sub, ok := dynCtx.GetResSub(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base subscription rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		dynCtx.DynamicItem.Extend.OrigImgUrl = sub.Icon
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, sub.Title, dynCtx, general)
	case dynCtx.Dyn.IsLiveRcmd(): // 直播推荐卡
		dynCtx.DynamicItem.CardType = api.DynamicType_live_rcmd
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_live_rcmd
		dynCtx.Interim.HiddenAuthorLive = true // 隐藏直播标记
		livercmd, ok := dynCtx.GetResLiveRcmd(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base live_rcmd rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		switch livercmd.Type {
		case livemdl.CardTypeLiving:
			if livercmd.LivePlayInfo != nil {
				dynCtx.DynamicItem.Extend.OrigImgUrl = livercmd.LivePlayInfo.Cover
				dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, livercmd.LivePlayInfo.Title, dynCtx, general)
			}
			if dynCtx.From != _handleTypeView {
				// 默认跳转和评论
				dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoLive, strconv.FormatInt(livercmd.LivePlayInfo.RoomId, 10), nil)
				if livercmd.LivePlayInfo.Link != "" {
					dynCtx.DynamicItem.Extend.CardUrl = livercmd.LivePlayInfo.Link
				}
			}
		case livemdl.CardTypePlayBack: // 废弃
			if livercmd.LiveRecordInfo != nil {
				dynCtx.DynamicItem.Extend.OrigImgUrl = livercmd.LiveRecordInfo.Cover
				dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, livercmd.LiveRecordInfo.Title, dynCtx, general)
			}
		}
		// 直播卡是安卓HD跳转直播间
		if general.IsAndroidHD() {
			dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoLive, strconv.FormatInt(livercmd.LivePlayInfo.RoomId, 10), nil)
			if livercmd.LivePlayInfo.Link != "" {
				dynCtx.DynamicItem.Extend.Reply.Uri = livercmd.LivePlayInfo.Link
			}
		}
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynReply, &feature.OriginResutl{
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.DynReplyIOS) ||
				(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.DynReplyAndroid)}) {
			dynCtx.DynamicItem.Extend.Reply.Uri = ""
		}
	case dynCtx.Dyn.IsUGCSeason(): // 合集卡
		dynCtx.DynamicItem.CardType = api.DynamicType_ugc_season
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_ugc_season
		_, ok := dynCtx.GetResUGCSeason(dynCtx.Dyn.UID)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base ugc_season rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.UID)
			return nil
		}
		ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base ugc_season archive rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		var archive = ap.Arc
		// 付费合集 + Pad不展示付费卡、校园也不展示
		if mdlv2.PayAttrVal(archive) && ((general.IsPad() && general.GetBuild() < s.c.BuildLimit.DynFuFeiIOS || general.IsPadHD() && general.GetBuild() < s.c.BuildLimit.DynFuFeiIOSHD || general.IsAndroidHD() && general.GetBuild() < s.c.BuildLimit.DynFuFeiAndroidHD) || (dynCtx.From == _handleTypeSchool || dynCtx.From == _handleTypeSchoolTopicFeed)) {
			dynCtx.Interim.IsPassCard = true
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转和评论
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, archive.FirstCid, true))
			dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(archive.Aid, 10), model.AvPlayHandlerGRPCV2(ap, archive.FirstCid, false))
		}
		replyParam := new(api.ExtendReplyParam)
		replyParam.Key = "comment_on"
		replyParam.Value = "1"
		dynCtx.DynamicItem.Extend.Reply.Params = []*api.ExtendReplyParam{replyParam}
		// 转发页面数据
		dynCtx.DynamicItem.Extend.Uid = archive.Author.Mid
		dynCtx.Interim.UName = archive.Author.Name
		dynCtx.DynamicItem.Extend.OrigName = "" // 兼容 合集订阅卡暂时不展示用户名
		dynCtx.DynamicItem.Extend.OrigImgUrl = archive.Pic
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, archive.Title, dynCtx, general)
	case dynCtx.Dyn.IsUGCSeasonShare(): // UGC合集分享卡（仅转发）
		dynCtx.DynamicItem.CardType = api.DynamicType_medialist
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_medialist
		ss, ok := dynCtx.GetResUGCSeason(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base ugc_season_share rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		user, ok := dynCtx.GetUser(dynCtx.Dyn.UID)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss due to mid(%v) not found. dynid(%v) base ugc_season_share rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeView {
			// 默认跳转和评论
			dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoAv, strconv.FormatInt(ss.FirstAid, 10), nil)
			dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoAv, strconv.FormatInt(ss.FirstAid, 10), nil)
		}
		// 转发页面数据
		dynCtx.Interim.UName = user.Name
		dynCtx.DynamicItem.Extend.OrigName = user.Name
		dynCtx.DynamicItem.Extend.Uid = ss.Mid
		dynCtx.DynamicItem.Extend.OrigImgUrl = ss.Cover
		dynCtx.DynamicItem.Extend.OrigFace = user.Face
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, ss.Title, dynCtx, general)
	case dynCtx.Dyn.IsSubscriptionNew():
		dynCtx.DynamicItem.CardType = api.DynamicType_subscription_new
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_subscription_new
		subNew, ok := dynCtx.GetResSubNew(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base subscription_new rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		// 默认跳转和评论
		switch subNew.Type {
		case submdl.TunnelTypeLive:
			dynCtx.Interim.HiddenAuthorLive = true // 隐藏直播标记
			var subNewLive *submdl.Live
			if err := json.Unmarshal([]byte(subNew.LiveInfo), &subNewLive); err != nil {
				log.Warn("card miss mid(%v) dynid(%v) base subscription_new live rid(%v), error %v", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid, err)
				dynCtx.Interim.IsPassCard = true
				return nil
			}
			if subNewLive != nil && subNewLive.LivePlayInfo != nil {
				if dynCtx.From != _handleTypeView {
					dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoLive, strconv.FormatInt(subNewLive.LivePlayInfo.RoomId, 10), nil)
					if subNewLive.LivePlayInfo.Link != "" {
						dynCtx.DynamicItem.Extend.CardUrl = subNewLive.LivePlayInfo.Link
					}
					if dynCtx.Dyn.Property != nil &&
						dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_ESPORTS_RESERVE {
						dynCtx.DynamicItem.Extend.Reply.Uri = model.FillURI(model.GotoLive, strconv.FormatInt(subNewLive.LivePlayInfo.RoomId, 10), nil)
						if subNewLive.LivePlayInfo.Link != "" {
							dynCtx.DynamicItem.Extend.Reply.Uri = subNewLive.LivePlayInfo.Link
						}
					}
				}
				// 转发页面数据
				dynCtx.DynamicItem.Extend.OrigImgUrl = subNewLive.LivePlayInfo.Cover
				dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, subNewLive.LivePlayInfo.Title, dynCtx, general)
			}
		case submdl.TunnelTypeDraw:
			var sub *submdl.Subscription
			if err := json.Unmarshal([]byte(subNew.ImageInfo), &sub); err != nil {
				log.Warn("card miss mid(%v) dynid(%v) base subscription_new draw rid(%v), error %v", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid, err)
				dynCtx.Interim.IsPassCard = true
				return nil
			}
			// 转发页面数据
			dynCtx.DynamicItem.Extend.OrigImgUrl = sub.Icon
			dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, sub.Title, dynCtx, general)
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
	case dynCtx.Dyn.IsBatch(): // 追漫卡
		dynCtx.DynamicItem.CardType = api.DynamicType_common_square
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_common_square
		batch, ok := dynCtx.ResBatch[dynCtx.Dyn.Rid]
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base batch square rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		if dynCtx.From != _handleTypeForward {
			if isFav, ok := dynCtx.ResBatchIsFav[dynCtx.Dyn.UID]; !ok || !isFav {
				dynCtx.Interim.IsPassCard = true
				return nil
			}
		}
		// 转发页面数据
		dynCtx.Interim.UName = batch.Name
		dynCtx.DynamicItem.Extend.OrigName = batch.Name
		dynCtx.DynamicItem.Extend.OrigImgUrl = batch.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, batch.Title, dynCtx, general)
		dynCtx.DynamicItem.Extend.CardUrl = batch.JumpURL
		dynCtx.DynamicItem.Extend.Reply.Uri = batch.JumpURL
	case dynCtx.Dyn.IsNewTopicSet(): // 新话题 话题集订阅卡
		dynCtx.DynamicItem.CardType = api.DynamicType_topic_set
		tps := dynCtx.GetResNewTopicSet()
		if tps == nil {
			dynCtx.Interim.IsPassCard = true
			log.Warnc(c, "card miss mid(%d) dynid(%d) base new topic-set subscribe rid(%d)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		dynCtx.Interim.UName = tps.SetInfo.GetBasicInfo().GetSetName()
		// 利用帮推的URI设置跳转链接
		dynCtx.Interim.PromoURI = tps.SetInfo.GetBasicInfo().GetJumpUrl()
	default:
		dynCtx.Interim.IsPassCard = true
		log.Warn("card miss mid(%v) dynid(%v) base unknow type(%v) rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid)
		return nil
	}
	// trackid
	if dynCtx.DynamicItem.Extend.CardUrl != "" && dynCtx.Dyn.TrackID != "" {
		dynCtx.DynamicItem.Extend.CardUrl = model.FillReplyURL(dynCtx.DynamicItem.Extend.CardUrl, fmt.Sprintf("trackid=%s", dynCtx.Dyn.TrackID))
		dynCtx.DynamicItem.Extend.Reply.Uri = model.FillReplyURL(dynCtx.DynamicItem.Extend.Reply.Uri, fmt.Sprintf("trackid=%s", dynCtx.Dyn.TrackID))
	}
	// 抽离文案数据 很多地方会用到
	dynCtx.Interim.Desc = s.getDesc(dynCtx, general)
	return nil
}

func (s *Service) baseFake(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	dynCtx.DynamicItem.Extend.RType = dynCtx.Dyn.RType
	dynCtx.DynamicItem.Extend.DynType = dynCtx.Dyn.Type
	dynCtx.DynamicItem.Extend.ShareType = _ShareType
	dynCtx.DynamicItem.Extend.ShareScene = _shareScene
	dynCtx.DynamicItem.Extend.IsFastShare = true
	dynCtx.DynamicItem.Extend.CardUrl = model.FillURI(model.GotoDyn, strconv.FormatInt(dynCtx.Dyn.DynamicID, 10), model.SuffixHandler(fmt.Sprintf("cardType=%v&rid=%d", dynCtx.Dyn.Type, dynCtx.Dyn.Rid)))
	dynCtx.DynamicItem.Extend.DynIdStr = strconv.FormatInt(dynCtx.Dyn.DynamicID, 10)
	dynCtx.DynamicItem.Extend.BusinessId = strconv.FormatInt(dynCtx.Dyn.Rid, 10)
	dynCtx.DynamicItem.Extend.Reply = &api.ExtendReply{
		Uri: model.FillURI(model.GotoDyn, strconv.FormatInt(dynCtx.Dyn.DynamicID, 10), model.SuffixHandler(fmt.Sprintf("cardType=%v&rid=%d", dynCtx.Dyn.Type, dynCtx.Dyn.Rid))),
		Params: []*api.ExtendReplyParam{
			{
				Key:   "comment_on",
				Value: "1",
			},
		},
	}
	switch {
	case dynCtx.Dyn.IsAv():
		dynCtx.DynamicItem.CardType = api.DynamicType_av
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_av
		if dynCtx.Dyn.FakeContent == "" {
			dynCtx.Dyn.FakeContent = "分享视频"
		}
		// 假卡视频不能转发
	case dynCtx.Dyn.IsWord():
		dynCtx.DynamicItem.CardType = api.DynamicType_word
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_word
		// 转发字段
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
			dynCtx.DynamicItem.Extend.OrigImgUrl = userInfo.Face
		}
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, dynCtx.Dyn.FakeContent, dynCtx, general)
	case dynCtx.Dyn.IsDraw():
		dynCtx.DynamicItem.CardType = api.DynamicType_draw
		dynCtx.DynamicItem.Extend.OrigDynType = api.DynamicType_draw
		if dynCtx.Dyn.FakeContent == "" {
			dynCtx.Dyn.FakeContent = "分享图片"
		}
		// 转发字段
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		for _, img := range dynCtx.Dyn.FakeImages {
			if img != nil && img.ImgSrc != "" {
				dynCtx.DynamicItem.Extend.OrigImgUrl = img.ImgSrc
				break
			}
		}
		dynCtx.DynamicItem.Extend.OrigDesc = s.descProc(c, dynCtx.Dyn.FakeContent, dynCtx, general)
	default:
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	dynCtx.Interim.Desc = dynCtx.Dyn.FakeContent
	// 预约卡
	for _, v := range dynCtx.Dyn.AttachCardInfos {
		if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.LotteryTypeCron, &feature.OriginResutl{
			MobiApp: general.GetMobiApp(),
			Device:  general.GetDevice(),
			Build:   general.GetBuild(),
			BuildLimit: (general.IsIPhonePick() && general.GetBuild() >= s.c.BuildLimit.LotteryTypeCronIOS) ||
				(general.IsAndroidPick() && general.GetBuild() > s.c.BuildLimit.LotteryTypeCronAndroid)}) {
			// 当前是预约抽奖卡，且当前是审核中增加小黄条
			if v.CardType != dyncommongrpc.AttachCardType_ATTACH_CARD_RESERVE {
				continue
			}
			res, ok := dynCtx.ResUpActRelationInfo[v.Rid]
			if !ok {
				continue
			}
			if res.LotteryType != activitygrpc.UpActReserveRelationLotteryType_UpActReserveRelationLotteryTypeCron {
				continue
			}
			if res.UpActVisible != activitygrpc.UpActVisible_OnlyUpVisible {
				continue
			}
			if upActState(res.State) != _upAudit && upActState(res.State) != _upStart {
				log.Warn("upActState (%d) %d", upActState(res.State), res.State)
				continue
			}
			dynCtx.Dyn.Extend.Dispute = &mdlv2.Dispute{
				Content: "动态审核中，仅自己可见",
			}
		}
	}
	return nil
}
