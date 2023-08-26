package dynamicV2

import (
	"context"
	"fmt"
	"strconv"

	"go-common/library/log"
	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/app-dynamic/interface/model"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	feature "go-gateway/app/app-svr/feature/service/sdk"

	cmtGrpc "git.bilibili.co/bapis/bapis-go/community/interface/reply"
)

func (s *Service) interaction(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
		return nil
	}
	var interactionItem []*api.InteractionItem
	// 点赞外显
	if ok, likeUser := dynCtx.Dyn.GetLikeUser(); ok && dynCtx.ResUser != nil {
		likeItem := &api.InteractionItem{
			IconType: api.LocalIconType_local_icon_like,
		}
		for k, uid := range likeUser {
			user, ok := dynCtx.GetUser(uid)
			if !ok {
				xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "interaction", "date_faild")
				log.Warn("module error mid(%v) dynid(%v) interaction_like uid %v", general.Mid, dynCtx.Dyn.DynamicID, uid)
				continue
			}
			var punctuation = "、"
			if k == (len(likeUser) - 1) {
				punctuation = " "
			}
			likeItem.Desc = append(likeItem.Desc, &api.Description{
				Type: api.DescType_desc_type_user,
				Text: fmt.Sprintf("%v%v", user.Name, punctuation),
				Uri:  model.FillURI(model.GotoSpaceDyn, strconv.FormatInt(user.Mid, 10), nil),
			})
		}
		if len(likeItem.Desc) > 0 {
			likeItem.Desc = append(likeItem.Desc, &api.Description{
				Type: api.DescType_desc_type_text,
				Text: "赞了",
			})
			interactionItem = append(interactionItem, likeItem)
		}
	} else if campusLike := dynCtx.Dyn.GetCampusLike(); campusLike != nil && feature.GetBuildLimit(c,
		s.c.Feature.FeatureBuildLimit.CampusDynInteraction, &feature.OriginResutl{
			BuildLimit: general.IsMobileBuildLimitMet(mdlv2.GreaterOrEqual,
				s.c.BuildLimit.CampusDynInteractionAndroid,
				s.c.BuildLimit.CampusDynInteractionIOS),
		}) {
		// 优先展示关注好友的点赞外露
		// 如果没有关注好友的点赞再显示同学点赞
		campusLikeItem := &api.InteractionItem{
			IconType: api.LocalIconType_local_icon_avatar,
			Uri:      fmt.Sprintf("bilibili://campus/like_list/%d", dynCtx.Dyn.DynamicID),
			Stat:     &api.InteractionStat{Like: dynCtx.Dyn.Extend.CampusLike.Total},
		}
		index, maxFaces := 1, 3
		for _, uid := range campusLike {
			user, ok := dynCtx.GetUser(uid)
			if !ok {
				xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "interaction", "date_faild")
				log.Warn("module error mid(%v) dynid(%v) campus_interaction uid %v info not found", general.Mid, dynCtx.Dyn.DynamicID, uid)
				continue
			}
			if index > maxFaces {
				break
			}
			campusLikeItem.Faces = append(campusLikeItem.Faces, &api.InteractionFace{Mid: user.Mid, Face: user.Face})
			index++
		}
		campusLikeItem.Desc = append(campusLikeItem.Desc, &api.Description{
			Type: api.DescType_desc_type_text,
			Text: fmt.Sprintf("%d个同学点赞了", dynCtx.Dyn.Extend.CampusLike.Total),
		})
		interactionItem = append(interactionItem, campusLikeItem)
	}
	/**
	 *	评论外露模块，目前支持动态类型：纯文字、图文、小视频、视频、转发、专栏、直播推荐
	 */
	if community, ok := dynCtx.GetReply(); ok {
		// 外露逻辑
		for _, item := range community.Replies {
			user, ok := dynCtx.GetUser(item.Mid)
			if !ok {
				xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "interaction", "date_faild")
				log.Warn("module error mid(%v) dynid(%v) interaction_reply item.Mid %v", general.Mid, dynCtx.Dyn.DynamicID, item.Mid)
				continue
			}
			communityItem := &api.InteractionItem{
				IconType:  api.LocalIconType_local_icon_comment,
				DynamicId: strconv.FormatInt(dynCtx.Dyn.DynamicID, 10),
			}
			// 神级评论
			if item.Label == cmtGrpc.DynamicFeedReplyMetaReply_Godlike {
				communityItem.IconType = api.LocalIconType_local_icon_cover
				communityItem.Icon = s.c.Resource.Icon.GodReply
			}
			communityItem.CommentMid = user.Mid
			communityItem.Desc = append(communityItem.Desc, &api.Description{
				Type: api.DescType_desc_type_user,
				Text: fmt.Sprintf("%v：", user.Name),
			})
			communityItem.Desc = append(communityItem.Desc, s.descProcCommunity(c, item.Content, dynCtx)...)
			interactionItem = append(interactionItem, communityItem)
		}
	}
	// 聚合所有外露模块
	if len(interactionItem) > 0 {
		module := &api.Module{
			ModuleType: api.DynModuleType_module_interaction,
			ModuleItem: &api.Module_ModuleInteraction{
				ModuleInteraction: &api.ModuleInteraction{
					InteractionItem: interactionItem,
				},
			},
		}
		dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	}
	return nil
}

func (s *Service) interactionAD(c context.Context, dynCtx *mdlv2.DynamicContext, general *mdlv2.GeneralParam) error {
	if dynCtx.Interim.IsPassCard {
		return nil
	}
	if general.IsPadHD() || general.IsAndroidHD() || general.IsPad() {
		return nil
	}
	if feature.GetBuildLimit(c, s.c.Feature.FeatureBuildLimit.DynAdFlyReply, &feature.OriginResutl{
		BuildLimit: (general.IsIPhonePick() && general.GetBuild() < s.c.BuildLimit.DynAdFlyReplyIOS) ||
			(general.IsAndroidPick() && general.GetBuild() <= s.c.BuildLimit.DynAdFlyReplyAndroid)}) {
		return nil
	}
	var interactionItem []*api.InteractionItem
	if community, ok := dynCtx.GetReply(); ok {
		// 外露逻辑
		for _, item := range community.Replies {
			user, ok := dynCtx.GetUser(item.Mid)
			if !ok {
				xmetric.DynamicModuleError.Inc(s.fromName(dynCtx.From), mdlv2.DynamicName(dynCtx.Dyn.Type), "interaction", "date_faild")
				log.Warn("module error mid(%v) dynid(%v) interaction_reply item.Mid %v", general.Mid, dynCtx.Dyn.DynamicID, item.Mid)
				continue
			}
			communityItem := &api.InteractionItem{
				IconType:  api.LocalIconType_local_icon_comment,
				DynamicId: strconv.FormatInt(dynCtx.Dyn.DynamicID, 10),
			}
			communityItem.CommentMid = user.Mid
			communityItem.Desc = append(communityItem.Desc, &api.Description{
				Type: api.DescType_desc_type_user,
				Text: fmt.Sprintf("%v：", user.Name),
			})
			communityItem.Desc = append(communityItem.Desc, s.descProcCommunity(c, item.Content, dynCtx)...)
			interactionItem = append(interactionItem, communityItem)
		}
	}
	// 聚合所有外露模块
	if len(interactionItem) > 0 {
		module := &api.Module{
			ModuleType: api.DynModuleType_module_interaction,
			ModuleItem: &api.Module_ModuleInteraction{
				ModuleInteraction: &api.ModuleInteraction{
					InteractionItem: interactionItem,
				},
			},
		}
		dynCtx.DynamicItem.Modules = append(dynCtx.DynamicItem.Modules, module)
	}
	return nil
}
