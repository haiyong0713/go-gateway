package dynHandler

import (
	"fmt"
	"strconv"

	"go-common/library/log"

	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	topiccardschema "go-gateway/app/app-svr/topic/card/schema"
)

func (schema *CardSchema) interaction(dynSchemaCtx *topiccardmodel.DynSchemaCtx, general *topiccardmodel.GeneralParam) error {
	dynCtx := dynSchemaCtx.DynCtx
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
		return nil
	}
	var interactionItem []*dynamicapi.InteractionItem
	// 点赞外显
	if ok, likeUser := dynCtx.Dyn.GetLikeUser(); ok && dynCtx.ResUser != nil {
		likeItem := &dynamicapi.InteractionItem{
			IconType: dynamicapi.LocalIconType_local_icon_like,
		}
		for k, uid := range likeUser {
			user, ok := dynCtx.GetUser(uid)
			if !ok {
				log.Warn("module error mid(%v) dynid(%v) interaction_like uid %v", general.Mid, dynCtx.Dyn.DynamicID, uid)
				continue
			}
			var punctuation = "、"
			if k == (len(likeUser) - 1) {
				punctuation = " "
			}
			likeItem.Desc = append(likeItem.Desc, &dynamicapi.Description{
				Type: dynamicapi.DescType_desc_type_user,
				Text: fmt.Sprintf("%v%v", user.Name, punctuation),
				Uri:  topiccardmodel.FillURI(topiccardmodel.GotoSpaceDyn, strconv.FormatInt(user.Mid, 10), nil),
			})
		}
		if len(likeItem.Desc) > 0 {
			likeItem.Desc = append(likeItem.Desc, &dynamicapi.Description{
				Type: dynamicapi.DescType_desc_type_text,
				Text: "赞了",
			})
			interactionItem = append(interactionItem, likeItem)
		}
	}
	/**
	 *	评论外露模块
	 */
	// 话题控制评论外露
	cmtShowStat, cmtShowNum := topiccardmodel.MakeDynCmtMode(dynSchemaCtx.DynCmtMode, dynCtx.Dyn.DynamicID)
	if community, ok := dynCtx.GetReply(); ok {
		// 外露逻辑
		for num, item := range community.Replies {
			if !cmtShowStat || num > cmtShowNum {
				break
			}
			user, ok := dynCtx.GetUser(item.Mid)
			if !ok {
				log.Warn("module error mid(%v) dynid(%v) interaction_reply item.Mid %v", general.Mid, dynCtx.Dyn.DynamicID, item.Mid)
				continue
			}
			communityItem := &dynamicapi.InteractionItem{
				IconType:   dynamicapi.LocalIconType_local_icon_comment,
				DynamicId:  strconv.FormatInt(dynCtx.Dyn.DynamicID, 10),
				CommentMid: user.Mid,
			}
			communityItem.Desc = append(communityItem.Desc, &dynamicapi.Description{
				Type: dynamicapi.DescType_desc_type_user,
				Text: fmt.Sprintf("%v：", user.Name),
			})
			communityItem.Desc = append(communityItem.Desc, topiccardschema.DescProcCommunity(item.Content, dynCtx)...)
			interactionItem = append(interactionItem, communityItem)
		}
	}
	// 聚合所有外露模块
	if len(interactionItem) > 0 {
		module := &dynamicapi.Module{
			ModuleType: dynamicapi.DynModuleType_module_interaction,
			ModuleItem: &dynamicapi.Module_ModuleInteraction{
				ModuleInteraction: &dynamicapi.ModuleInteraction{
					InteractionItem: interactionItem,
				},
			},
		}
		dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	}
	return nil
}
