package cardbuilder

import (
	"fmt"
	"strconv"

	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
)

func handleModuleInteraction(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.ModuleInteraction {
	return &jsonwebcard.ModuleInteraction{
		Like:    makeInteractionLike(dynCtx),
		Comment: makeInteractionComment(metaCtx, dynCtx),
	}
}

func makeInteractionComment(metaCtx jsonwebcard.MetaContext, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.InteractiveItem {
	community, ok := dynCtx.GetReply()
	if !ok {
		return nil
	}
	res := &jsonwebcard.InteractiveItem{
		JumpUrl: fmt.Sprintf("https://m.bilibili.com/dynamic/%d", dynCtx.Dyn.DynamicID),
	}
	cmtShowStat, cmtShowNum := topiccardmodel.MakeDynCmtMode(metaCtx.Config.DynCmtTopicControl, dynCtx.Dyn.DynamicID)
	if cmtShowStat {
		return nil
	}
	var nodes []*jsonwebcard.RichTextNode
	for num, item := range community.Replies {
		if num > cmtShowNum {
			break
		}
		user, ok := dynCtx.GetUser(item.Mid)
		if !ok {
			continue
		}
		nodes = append(nodes, &jsonwebcard.RichTextNode{
			DescItemType: jsonwebcard.RichTextNodeTypeUser,
			Text:         fmt.Sprintf("%v：", user.Name),
		})
	}
	res.Desc = &jsonwebcard.DynDesc{RichTextNode: nodes}
	return res
}

func makeInteractionLike(dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.InteractiveItem {
	ok, likeUser := dynCtx.Dyn.GetLikeUser()
	if !ok || dynCtx.ResUser == nil {
		return nil
	}
	res := &jsonwebcard.InteractiveItem{
		JumpUrl: fmt.Sprintf("https://m.bilibili.com/dynamic/%d", dynCtx.Dyn.DynamicID),
	}
	var nodes []*jsonwebcard.RichTextNode
	for k, uid := range likeUser {
		user, ok := dynCtx.GetUser(uid)
		if !ok {
			continue
		}
		var punctuation = "、"
		if k == (len(likeUser) - 1) {
			punctuation = " "
		}
		nodes = append(nodes, &jsonwebcard.RichTextNode{
			Text:         fmt.Sprintf("%v%v", user.Name, punctuation),
			DescItemType: jsonwebcard.RichTextNodeTypeUser,
			JumpUrl:      topiccardmodel.FillURI(topiccardmodel.GotoWebSpace, strconv.FormatInt(user.Mid, 10), nil),
		})
	}
	if len(nodes) > 0 {
		res.Desc = &jsonwebcard.DynDesc{
			Text:         "赞了",
			RichTextNode: nodes,
		}
	}
	return res
}
