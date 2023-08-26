package dynHandler

import (
	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

const (
	_moduleStatNoComment              = "这条动态已被封印，当前不可评论╮(๑•́ ₃•̀๑)╭"
	_moduleStatNoForward              = "这条动态已被封印，当前不可转发╮(๑•́ ₃•̀๑)╭"
	_moduleStatNoForwardForwardFailed = "源动态不见惹，不可以转发噢~"
)

func (schema *CardSchema) statInfo(dynCtx *dynmdlV2.DynamicContext) *dynamicapi.ModuleStat {
	stat := &dynamicapi.ModuleStat{
		Repost: dynCtx.Dyn.Repost, // 禁止转发会置为0
	}
	var (
		reply, like, share int64
		replyuri           string
		islike             bool
	)
	switch {
	case dynCtx.Dyn.IsForward():
		reply, like, replyuri, islike = schema.statCommon(dynCtx)
	case dynCtx.Dyn.IsAv():
		reply, like, replyuri, islike = schema.statAV(dynCtx)
	case dynCtx.Dyn.IsDraw():
		reply, like, replyuri, islike = schema.statDraw(dynCtx)
	case dynCtx.Dyn.IsWord():
		reply, like, replyuri, islike = schema.statCommon(dynCtx)
	}
	stat.Reply = reply // 禁止评论会置为0
	stat.Like = like
	stat.ReplyUrl = replyuri
	if share > 0 {
		stat.Repost = share
	}
	if islike {
		if stat.LikeInfo == nil {
			stat.LikeInfo = new(dynamicapi.LikeInfo)
		}
		stat.LikeInfo.IsLike = true
	}
	// 点赞动画
	if dynCtx.Dyn.Extend != nil && dynCtx.Dyn.Extend.LikeIcon != nil {
		if stat.LikeInfo == nil {
			stat.LikeInfo = new(dynamicapi.LikeInfo)
		}
		stat.LikeInfo.Animation = &dynamicapi.LikeAnimation{
			Begin:      dynCtx.Dyn.Extend.LikeIcon.Begin,
			Proc:       dynCtx.Dyn.Extend.LikeIcon.Proc,
			End:        dynCtx.Dyn.Extend.LikeIcon.End,
			LikeIconId: dynCtx.Dyn.Extend.LikeIcon.NewIconID,
		}
	}
	// 优先，评论服务端-禁止评论
	if replyInfo, ok := dynCtx.GetReply(); ok && replyInfo.NoComment {
		stat.NoComment = true
		stat.Reply = 0
		stat.NoCommentText = _moduleStatNoComment
	} else if dynCtx.Dyn.ACL != nil && dynCtx.Dyn.ACL.CommentBan == 1 { // 否则，动态服务端-禁止评论
		stat.NoComment = true
		stat.Reply = 0
		stat.NoCommentText = _moduleStatNoComment
	}
	// 禁止转发
	if dynCtx.Dyn.ACL != nil && dynCtx.Dyn.ACL.RepostBan == 1 {
		stat.NoForward = true
		stat.Repost = 0
		stat.NoForwardText = _moduleStatNoForward
	} else if dynCtx.Dyn.IsForward() && dynCtx.Interim.ForwardOrigFaild {
		stat.NoForward = true
		stat.Repost = 0
		stat.NoForwardText = _moduleStatNoForwardForwardFailed
	}
	return stat
}

func (schema *CardSchema) stat(dynSchemaCtx *topiccardmodel.DynSchemaCtx, _ *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_stat,
		ModuleItem: &dynamicapi.Module_ModuleStat{
			ModuleStat: schema.statInfo(dynCtx),
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (schema *CardSchema) statShell(dynSchemaCtx *topiccardmodel.DynSchemaCtx, _ *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	module := &dynamicapi.Module{
		ModuleType: dynamicapi.DynModuleType_module_stat_forward,
		ModuleItem: &dynamicapi.Module_ModuleStatForward{
			ModuleStatForward: schema.statInfo(dynCtx),
		},
	}
	dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	return nil
}

func (schema *CardSchema) statCommon(dynCtx *dynmdlV2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	reply = schema.getReply(dynCtx)
	islike, like = schema.getLike(dynCtx)
	return
}

func (schema *CardSchema) statAV(dynCtx *dynmdlV2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	if ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid); ok {
		var archive = ap.Arc
		like = int64(archive.Stat.Like)
		reply = int64(archive.Stat.Reply)
	}
	islike, _ = schema.getLike(dynCtx)
	return
}

func (schema *CardSchema) statDraw(dynCtx *dynmdlV2.DynamicContext) (reply, like int64, replyuri string, islike bool) {
	if draw, ok := dynCtx.GetResDraw(dynCtx.Dyn.Rid); ok {
		reply = int64(draw.Item.Reply)
	}
	islike, like = schema.getLike(dynCtx)
	return
}

func (schema *CardSchema) getLike(dynCtx *dynmdlV2.DynamicContext) (bool, int64) {
	var (
		busParam *dynmdlV2.ThumbsRecord
		busType  string
		isThum   bool
	)
	if dynCtx.Dyn.IsPGC() {
		if pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid)); ok {
			if busParam, busType, isThum = dynmdlV2.GetPGCLikeID(pgc); !isThum {
				return false, 0
			}
		}
	} else {
		if busParam, busType, isThum = dynCtx.Dyn.GetLikeID(); !isThum {
			return false, 0
		}
	}
	if busParam == nil {
		return false, 0
	}
	if dynCtx.ResLike != nil && dynCtx.ResLike[busType] != nil {
		if r, ok := dynCtx.ResLike[busType].Records[busParam.MsgID]; ok {
			if r.LikeState == thumgrpc.State_STATE_LIKE {
				return true, r.LikeNumber
			}
			return false, r.LikeNumber
		}
	}
	return false, 0
}

func (schema *CardSchema) getReply(dynCtx *dynmdlV2.DynamicContext) int64 {
	if replyTmp, ok := dynCtx.GetReply(); ok {
		return replyTmp.Count
	}
	return 0
}
