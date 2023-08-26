package cardbuilder

import (
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"

	thumgrpc "git.bilibili.co/bapis/bapis-go/community/service/thumbup"
)

func handleModuleStat(cardType jsonwebcard.CardType, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.ModuleStat {
	return &jsonwebcard.ModuleStat{
		Forward: makeModuleStatForward(dynCtx),
		Comment: makeModuleStatComment(cardType, dynCtx),
		Like:    makeModuleStatLike(cardType, dynCtx),
	}
}

func makeModuleStatLike(cardType jsonwebcard.CardType, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.MdlStatItem {
	var (
		likeState  string
		likeNumber int64
	)
	busParam, busType, ok := resolveStatLikeParams(dynCtx)
	if ok && dynCtx.ResLike != nil && dynCtx.ResLike[busType] != nil {
		if r, ok := dynCtx.ResLike[busType].Records[busParam.MsgID]; ok {
			likeNumber = r.LikeNumber
			if r.LikeState == thumgrpc.State_STATE_LIKE {
				likeState = "STATE_LIKE"
			}
		}
	}
	switch cardType {
	case jsonwebcard.CardDynamicTypeAv:
		ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid)
		if !ok {
			return nil
		}
		return &jsonwebcard.MdlStatItem{
			Count:  int64(ap.Arc.Stat.Like),
			Text:   topiccardmodel.StatString(int64(ap.Arc.Stat.Like), "", ""),
			Status: likeState,
		}
	default:
		return &jsonwebcard.MdlStatItem{
			Count:  likeNumber,
			Text:   topiccardmodel.StatString(likeNumber, "", ""),
			Status: likeState,
		}
	}
}

func resolveStatLikeParams(dynCtx *dynmdlV2.DynamicContext) (*dynmdlV2.ThumbsRecord, string, bool) {
	if dynCtx.Dyn.IsPGC() {
		pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
		if !ok {
			return nil, "", false
		}
		busParam, busType, isThum := dynmdlV2.GetPGCLikeID(pgc)
		if !isThum {
			return nil, "", false
		}
		return busParam, busType, true
	}
	busParam, busType, isThum := dynCtx.Dyn.GetLikeID()
	if !isThum {
		return nil, "", false
	}
	return busParam, busType, true
}

func makeModuleStatComment(cardType jsonwebcard.CardType, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.MdlStatItem {
	// 评论服务端-禁止评论
	if replyInfo, ok := dynCtx.GetReply(); ok && replyInfo.NoComment {
		return &jsonwebcard.MdlStatItem{Forbidden: true}
	}
	// 动态服务端-禁止评论
	if dynCtx.Dyn.ACL != nil && dynCtx.Dyn.ACL.CommentBan == 1 {
		return &jsonwebcard.MdlStatItem{Forbidden: true}
	}
	switch cardType {
	case jsonwebcard.CardDynamicTypeAv:
		ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid)
		if !ok {
			return nil
		}
		return &jsonwebcard.MdlStatItem{
			Count:     int64(ap.Arc.Stat.Reply),
			Forbidden: false,
			Text:      topiccardmodel.StatString(int64(ap.Arc.Stat.Reply), "", ""),
		}
	case jsonwebcard.CardDynamicTypeDraw:
		draw, ok := dynCtx.GetResDraw(dynCtx.Dyn.Rid)
		if !ok {
			return nil
		}
		return &jsonwebcard.MdlStatItem{
			Count: int64(draw.Item.Reply),
			Text:  topiccardmodel.StatString(int64(draw.Item.Reply), "", ""),
		}
	case jsonwebcard.CardDynamicTypeForward, jsonwebcard.CardDynamicTypeWord:
		replyTmp, ok := dynCtx.GetReply()
		if !ok {
			return nil
		}
		return &jsonwebcard.MdlStatItem{
			Count: replyTmp.Count,
			Text:  topiccardmodel.StatString(replyTmp.Count, "", ""),
		}
	default:
		return nil
	}
}

func makeModuleStatForward(dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.MdlStatItem {
	// 评论服务端-禁止评论
	if dynCtx.Dyn.ACL != nil && dynCtx.Dyn.ACL.RepostBan == 1 {
		return &jsonwebcard.MdlStatItem{Forbidden: true}
	}
	// 动态服务端-禁止评论
	if dynCtx.Dyn.IsForward() && dynCtx.Interim.ForwardOrigFaild {
		return &jsonwebcard.MdlStatItem{Forbidden: true}
	}
	res := &jsonwebcard.MdlStatItem{
		Count: dynCtx.Dyn.Repost,
		Text:  topiccardmodel.StatString(dynCtx.Dyn.Repost, "", ""),
	}
	return res
}
