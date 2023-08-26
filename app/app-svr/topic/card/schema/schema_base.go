package topiccardschema

import (
	"fmt"
	"strconv"

	"go-common/library/log"

	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	midint64 "go-gateway/app/app-svr/app-interface/interface-legacy/middleware/midInt64"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"

	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

const (
	_ShareType  = "3"
	_shareScene = "dynamic"
)

func HandleDynamicCardBase(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	// 初始化公共值
	dynCtx.DynamicItem.Extend = &dynamicapi.Extend{
		DynIdStr:    strconv.FormatInt(dynCtx.Dyn.DynamicID, 10),
		BusinessId:  strconv.FormatInt(dynCtx.Dyn.Rid, 10),
		ShareType:   _ShareType,
		ShareScene:  _shareScene,
		IsFastShare: true,
		RType:       dynCtx.Dyn.RType,
		DynType:     dynCtx.Dyn.Type,
		CardUrl:     topiccardmodel.FillURI(topiccardmodel.GotoDyn, strconv.FormatInt(dynCtx.Dyn.DynamicID, 10), topiccardmodel.SuffixHandler(fmt.Sprintf("cardType=%v&rid=%d", dynCtx.Dyn.Type, dynCtx.Dyn.Rid))),
	}
	// 透传字段处理
	if v, ok := dynSchemaCtx.ServerInfo[dynCtx.Dyn.DynamicID]; ok {
		dynCtx.DynamicItem.ServerInfo = v
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
	dynCtx.DynamicItem.Extend.Reply = &dynamicapi.ExtendReply{
		Uri: topiccardmodel.FillURI(topiccardmodel.GotoDyn, strconv.FormatInt(dynCtx.Dyn.DynamicID, 10), topiccardmodel.SuffixHandler(fmt.Sprintf("cardType=%v&rid=%d", dynCtx.Dyn.Type, dynCtx.Dyn.Rid))),
		Params: []*dynamicapi.ExtendReplyParam{
			{
				Key:   "comment_on",
				Value: "1",
			},
		},
	}
	switch {
	case dynCtx.Dyn.IsForward(): // 转发卡
		dynCtx.DynamicItem.CardType = dynamicapi.DynamicType_forward
		dynCtx.DynamicItem.Extend.OrigDynType = dynamicapi.DynamicType_forward
		dynCtx.Interim.DynTypeShell = dynCtx.Dyn.Type
		dynCtx.Interim.ShellRID = dynCtx.Dyn.Rid
		if dynCtx.Dyn.Origin != nil {
			dynCtx.Interim.DynTypeKernel = dynCtx.Dyn.Origin.Type
			dynCtx.Interim.KernelRID = dynCtx.Dyn.Origin.Rid
			if !dynCtx.Dyn.Origin.Visible {
				dynCtx.Interim.ForwardOrigFaild = true
				log.Warn("isForward warn mid(%v) dynid(%v) base Visible false", general.Mid, dynCtx.Dyn.DynamicID)
			}
		} else {
			dynCtx.Interim.ForwardOrigFaild = true // 源卡失效
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
		dynCtx.DynamicItem.CardType = dynamicapi.DynamicType_av
		dynCtx.DynamicItem.Extend.OrigDynType = dynamicapi.DynamicType_av
		ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid)
		if !ok || !ap.Arc.IsNormal() {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base av rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		var archive = ap.Arc
		dynCtx.Interim.CID = archive.FirstCid
		// 新版走非首P逻辑
		dynCtx.Interim.CID = ap.DefaultPlayerCid
		if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.VideoShare != nil {
			dynCtx.Interim.CID = dynCtx.Dyn.Extend.VideoShare.CID
		}
		// 默认跳转和评论
		dynCtx.DynamicItem.Extend.CardUrl = topiccardmodel.FillURI(topiccardmodel.GotoAv, strconv.FormatInt(archive.Aid, 10), topiccardmodel.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))
		dynCtx.DynamicItem.Extend.Reply.Uri = topiccardmodel.FillURI(topiccardmodel.GotoAv, strconv.FormatInt(archive.Aid, 10), topiccardmodel.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, false))
		replyParam := new(dynamicapi.ExtendReplyParam)
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
		dynCtx.DynamicItem.Extend.Reply.Params = []*dynamicapi.ExtendReplyParam{replyParam, {Key: "auto_float_layer", Value: "99"}}
		// 话题详情页点击视频卡片之后进入story页面，pad端默认进视频详情页
		if general.Source == "" && !general.IsPadHD() && !general.IsAndroidHD() && !general.IsPad() {
			dynCtx.DynamicItem.Extend.CardUrl = topiccardmodel.FillURI(topiccardmodel.GotoStory, strconv.FormatInt(archive.Aid, 10), topiccardmodel.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, true))
			dynCtx.DynamicItem.Extend.Reply.Uri = topiccardmodel.FillURI(topiccardmodel.GotoStory, strconv.FormatInt(archive.Aid, 10), topiccardmodel.AvPlayHandlerGRPCV2(ap, dynCtx.Interim.CID, false))
			// story特殊逻辑
			// 无论是否是转卡，都用最上层的vmid
			vmid := dynCtx.Dyn.UID
			if dynCtx.Dyn.Forward != nil {
				vmid = dynCtx.Dyn.Forward.UID
			}
			dynCtx.DynamicItem.Extend.CardUrl = topiccardmodel.FillURI(topiccardmodel.GotoURL, dynCtx.DynamicItem.Extend.CardUrl, topiccardmodel.SuffixHandler(topiccardmodel.MakeStorySuffixUrl(vmid, dynCtx.Dyn.DynamicID, dynSchemaCtx.TopicId, dynSchemaCtx.SortBy, dynSchemaCtx.Offset, "")))
			dynCtx.DynamicItem.Extend.Reply.Uri = topiccardmodel.FillURI(topiccardmodel.GotoURL, dynCtx.DynamicItem.Extend.Reply.Uri, topiccardmodel.SuffixHandler(topiccardmodel.MakeStorySuffixUrl(vmid, dynCtx.Dyn.DynamicID, dynSchemaCtx.TopicId, dynSchemaCtx.SortBy, dynSchemaCtx.Offset, "")))
		}
		// 转发页面数据
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
		}
		dynCtx.DynamicItem.Extend.OrigImgUrl = archive.Pic
		dynCtx.DynamicItem.Extend.OrigDesc = DescProc(dynCtx, archive.Title, general)
	case dynCtx.Dyn.IsDraw(): // 图文卡
		dynCtx.DynamicItem.CardType = dynamicapi.DynamicType_draw
		dynCtx.DynamicItem.Extend.OrigDynType = dynamicapi.DynamicType_draw
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
		dynCtx.DynamicItem.Extend.OrigDesc = DescProc(dynCtx, content, general)
	case dynCtx.Dyn.IsWord():
		dynCtx.DynamicItem.CardType = dynamicapi.DynamicType_word
		dynCtx.DynamicItem.Extend.OrigDynType = dynamicapi.DynamicType_word
		// 转发字段
		if userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID); ok {
			dynCtx.Interim.UName = userInfo.Name
			dynCtx.DynamicItem.Extend.OrigName = userInfo.Name
			dynCtx.DynamicItem.Extend.Uid = userInfo.Mid
			dynCtx.DynamicItem.Extend.OrigImgUrl = userInfo.Face
		}
		dynCtx.DynamicItem.Extend.OrigDesc = DescProc(dynCtx, descriptionWord(dynCtx), general)
	case dynCtx.Dyn.IsPGC(): // PGC卡
		dynCtx.DynamicItem.CardType = dynamicapi.DynamicType_pgc
		dynCtx.DynamicItem.Extend.OrigDynType = dynamicapi.DynamicType_pgc
		pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base PGC rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		dynCtx.DynamicItem.Extend.CardUrl = pgc.Url
		dynCtx.DynamicItem.Extend.Reply.Uri = pgc.Url
		replyParam := new(dynamicapi.ExtendReplyParam)
		replyParam.Key = "reply_id"
		replyParam.Value = "-1"
		if general.IsAndroidPick() {
			replyParam.Key = "comment_state"
			replyParam.Value = "1"
		}
		dynCtx.DynamicItem.Extend.Reply.Params = []*dynamicapi.ExtendReplyParam{replyParam}
		// 转发页面数据
		if pgc.Season != nil {
			dynCtx.Interim.UName = pgc.Season.Title
			dynCtx.DynamicItem.Extend.OrigName = pgc.Season.Title
		}
		dynCtx.DynamicItem.Extend.Uid = dynCtx.Dyn.UID
		dynCtx.DynamicItem.Extend.OrigImgUrl = pgc.Cover
		dynCtx.DynamicItem.Extend.OrigDesc = DescProc(dynCtx, pgc.CardShowTitle, general)
	case dynCtx.Dyn.IsArticle(): // 专栏卡
		dynCtx.DynamicItem.CardType = dynamicapi.DynamicType_article
		dynCtx.DynamicItem.Extend.OrigDynType = dynamicapi.DynamicType_article
		article, ok := dynCtx.GetResArticle(dynCtx.Dyn.Rid)
		if !ok {
			dynCtx.Interim.IsPassCard = true
			log.Warn("card miss mid(%v) dynid(%v) base article rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Rid)
			return nil
		}
		dynCtx.DynamicItem.Extend.CardUrl = topiccardmodel.FillURI(topiccardmodel.GotoArticle, strconv.FormatInt(article.ID, 10), nil)
		dynCtx.DynamicItem.Extend.Reply.Uri = topiccardmodel.FillURI(topiccardmodel.GotoArticle, strconv.FormatInt(article.ID, 10), nil)
		replyParam := new(dynamicapi.ExtendReplyParam)
		replyParam.Key = "reply_id"
		replyParam.Value = "-1"
		if general.IsAndroidPick() {
			replyParam.Key = "reply_id"
			replyParam.Value = "-2"
		}
		dynCtx.DynamicItem.Extend.Reply.Params = []*dynamicapi.ExtendReplyParam{replyParam}
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
		dynCtx.DynamicItem.Extend.OrigDesc = DescProc(dynCtx, article.Title, general)
	case dynCtx.Dyn.IsCommonSquare(): // 通用卡 方
		dynCtx.DynamicItem.CardType = dynamicapi.DynamicType_common_square
		dynCtx.DynamicItem.Extend.OrigDynType = dynamicapi.DynamicType_common_square
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
		dynCtx.DynamicItem.Extend.OrigDesc = DescProc(dynCtx, common.Sketch.Title, general)
	case dynCtx.Dyn.IsCommonVertical(): // 通用卡 竖
		dynCtx.DynamicItem.CardType = dynamicapi.DynamicType_common_vertical
		dynCtx.DynamicItem.Extend.OrigDynType = dynamicapi.DynamicType_common_vertical
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
		dynCtx.DynamicItem.Extend.OrigDesc = DescProc(dynCtx, common.Sketch.Title, general)
	default:
		dynCtx.Interim.IsPassCard = true
		log.Warn("card miss mid(%v) dynid(%v) base unknown type(%v) rid(%v)", general.Mid, dynCtx.Dyn.DynamicID, dynCtx.Dyn.Type, dynCtx.Dyn.Rid)
		return nil
	}
	if dynSchemaCtx.IsDisableInt64MidVersion && midint64.CheckHasInt64InMids(dynCtx.Dyn.UID) {
		// mid > int32老版本抛弃当前卡片
		dynCtx.Interim.IsPassCard = true
		return nil
	}
	dynCtx.Dyn.Extend = constructSafeDynExtend(dynCtx.Dyn.Extend)
	if dynSchemaCtx.OwnerAppear == 1 && dynSchemaCtx.TopicCreatorMid > 0 && isTopicCreatorLike(dynSchemaCtx) {
		// 话题点赞外露用户
		dynCtx.Dyn.Extend.Display.LikeUsers = append(dynCtx.Dyn.Extend.Display.LikeUsers, dynSchemaCtx.TopicCreatorMid)
	}
	// 抽离文案数据 很多地方会用到
	dynCtx.Interim.Desc = getDesc(dynCtx)
	return nil
}

func isTopicCreatorLike(dynSchemaCtx *topiccardmodel.DynSchemaCtx) bool {
	var (
		busParam *dynmdlV2.ThumbsRecord
		busType  string
		isThum   bool
	)
	dynCtx := dynSchemaCtx.DynCtx
	switch {
	case dynCtx.Dyn.IsPGC():
		if pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid)); ok {
			if busParam, busType, isThum = dynmdlV2.GetPGCLikeID(pgc); !isThum {
				return false
			}
		}
	default:
		if busParam, busType, isThum = dynCtx.Dyn.GetLikeID(); !isThum {
			return false
		}
	}
	if busParam == nil {
		return false
	}
	if dynSchemaCtx.TopicCreatorLike != nil && dynSchemaCtx.TopicCreatorLike[busType] != nil {
		if r, ok := dynSchemaCtx.TopicCreatorLike[busType].Records[busParam.MsgID]; ok {
			return r.LikeState == thumgrpc.State_STATE_LIKE
		}
	}
	return false
}

func constructSafeDynExtend(extend *dynmdlV2.Extend) *dynmdlV2.Extend {
	if extend == nil {
		return &dynmdlV2.Extend{Display: &dynmdlV2.Display{}}
	}
	if extend.Display == nil {
		extend.Display = &dynmdlV2.Display{}
	}
	return extend
}
