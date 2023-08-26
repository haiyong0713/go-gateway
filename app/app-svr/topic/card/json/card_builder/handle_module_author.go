package cardbuilder

import (
	"fmt"
	"strconv"

	"go-common/library/log"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	relationgrpc "git.bilibili.co/bapis/bapis-go/account/service/relation"
	dyncommongrpc "git.bilibili.co/bapis/bapis-go/dynamic/common"
)

func handleModuleAuthorPgc(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.ModuleAuthor {
	res := &jsonwebcard.ModuleAuthor{
		AuthorType: jsonwebcard.AuthorTypePgc,
		UserInfo: jsonwebcard.UserInfo{
			Mid: dynCtx.Dyn.UID,
		},
		PubTime:   topiccardmodel.ConstructPubTime(metaCtx.LocalTime, dynCtx.Dyn.Timestamp),
		PubAction: constructPubAction(metaCtx, dynCtx),
		PubTs:     dynCtx.Dyn.Timestamp,
	}
	pgc, ok := dynCtx.GetResPGC(int32(dynCtx.Dyn.Rid))
	if !ok || pgc.Season == nil {
		res.UserInfo.Face = "https://i0.hdslb.com/bfs/feed-admin/c4cf44cc63cbe7e9642482a600db915500fd4d2f.png"
		return res
	}
	res.UserInfo.Name = pgc.Season.Title
	res.UserInfo.Face = pgc.Season.Cover
	return res
}

func handleModuleAuthorUgc(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.ModuleAuthor {
	if dynCtx.Dyn.UID == 0 {
		return nil
	}
	userInfo, ok := dynCtx.GetUser(dynCtx.Dyn.UID)
	if !ok {
		log.Warn("handleModuleAuthorUgc module miss mid(%d) dynId(%d)", dynCtx.Dyn.UID, dynCtx.Dyn.DynamicID)
		return nil
	}
	res := &jsonwebcard.ModuleAuthor{
		AuthorType: jsonwebcard.AuthorTypeNormal,
		UserInfo: jsonwebcard.UserInfo{
			Face:    userInfo.Face,
			Name:    userInfo.Name,
			Mid:     userInfo.Mid,
			FaceNft: userInfo.FaceNftNew,
		},
		Following: metaCtx.Mid == dynCtx.Dyn.UID || constructRelationFollowing(dynCtx.Dyn.UID, dynCtx.ResRelationUltima),
		PubTime:   topiccardmodel.ConstructPubTime(metaCtx.LocalTime, dynCtx.Dyn.Timestamp),
		PubAction: constructPubAction(metaCtx, dynCtx),
		PubTs:     dynCtx.Dyn.Timestamp,
	}
	// 装扮
	res.Pendant.Pid = int64(userInfo.Pendant.Pid)
	res.Pendant.Name = userInfo.Pendant.Name
	res.Pendant.Image = userInfo.Pendant.Image
	res.Pendant.Expire = userInfo.Pendant.Expire
	res.Pendant.ImageEnhance = userInfo.Pendant.ImageEnhance
	res.Pendant.ImageEnhanceFrame = userInfo.Pendant.ImageEnhanceFrame
	// 会员信息
	res.Vip.Label = userInfo.Vip.Label
	res.Vip.VipStatus = userInfo.Vip.Status
	res.Vip.Type = userInfo.Vip.Type
	res.Vip.DueDate = userInfo.Vip.DueDate
	res.Vip.ThemeType = userInfo.Vip.ThemeType
	res.Vip.VipPayType = userInfo.Vip.VipPayType
	res.Vip.Role = userInfo.Vip.Role
	res.Vip.AvatarSubscript = userInfo.Vip.AvatarSubscript
	res.Vip.AvatarSubscriptUrl = userInfo.Vip.AvatarSubscriptUrl
	res.Vip.NicknameColor = userInfo.Vip.NicknameColor
	// 认证信息
	res.OfficialVerify.Type = userInfo.Official.Type
	res.OfficialVerify.Desc = userInfo.Official.Desc
	// 装扮
	res.Decorate = resolveAuthorDecorate(dynCtx, userInfo)
	return res
}

func resolveAuthorDecorate(dynCtx *dynmdlV2.DynamicContext, userInfo *accountgrpc.Card) *jsonwebcard.Decorate {
	decoInfo, ok := dynCtx.ResMyDecorate[userInfo.Mid]
	if !ok {
		return nil
	}
	res := &jsonwebcard.Decorate{
		Id:      decoInfo.ID,
		Type:    int64(decoInfo.ItemType),
		Name:    decoInfo.Name,
		CardUrl: decoInfo.CardURL,
		JumpUrl: decoInfo.JumpURL,
	}
	res.DecorateFan.IsFan = decoInfo.Fan.IsFan == 1
	res.DecorateFan.Number = int32(decoInfo.Fan.Number)
	res.DecorateFan.NumStr = constructDecorateNumStr(decoInfo.Fan.Number)
	res.DecorateFan.Color = decoInfo.Fan.Color
	if decoInfo.ImageEnhance != "" {
		res.CardUrl = decoInfo.ImageEnhance
	}
	return res
}

func constructDecorateNumStr(number int) string {
	// nolint:gomnd
	if number < 100000 {
		return fmt.Sprintf("%06d", number)
	}
	return strconv.Itoa(number)
}

func constructPubAction(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) string {
	if v, ok := metaCtx.Config.HiddenAttached[dynCtx.Dyn.DynamicID]; ok && v {
		return topiccardmodel.TopicHiddenAttatchedText
	}
	if !dynCtx.Dyn.IsAv() {
		return ""
	}
	if dynCtx.Dyn.Property != nil && dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_ARCHIVE {
		return "预约的视频"
	}
	if dynCtx.Dyn.Property != nil && dynCtx.Dyn.Property.RcmdType == dyncommongrpc.FeedRcmdType_FEED_RCMD_TYPE_RESERVE_LIVE_PLAY_BACK {
		return "预约的直播"
	}
	if ap, ok := dynCtx.GetArchive(dynCtx.Dyn.Rid); ok {
		var archive = ap.Arc
		if archive.Rights.IsCooperation == 1 {
			return "与他人联合创作"
		}
	}
	switch dynCtx.Dyn.SType {
	case dynmdlV2.VideoStypeDynamic, dynmdlV2.VideoStypeDynamicStory:
		return "发布了动态"
	case dynmdlV2.VideoStypePlayback:
		return "投稿了直播回放"
	}
	return "投稿了视频"
}

// nolint:gomnd
func constructRelationFollowing(mid int64, relations map[int64]*relationgrpc.InterrelationReply) bool {
	rel, ok := relations[mid]
	if !ok {
		return false
	}
	switch rel.Attribute {
	case 2, 6: // 用户关注UP主
		return true
	default:
		return false
	}
}
