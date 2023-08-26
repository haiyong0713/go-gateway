package cardbuilder

import (
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
)

func BackfillCard(card jsonwebcard.TopicCard, dynCtx *dynmdlV2.DynamicContext) *jsonwebcard.TopicCard {
	modules := card.GetModules()
	if modules.ModuleDynamic == nil || modules.ModuleDynamic.Desc == nil {
		return &card
	}
	for _, descItem := range modules.ModuleDynamic.Desc.RichTextNode {
		formDescItem(descItem, dynCtx)
	}
	if modules.ModuleInteraction == nil || modules.ModuleInteraction.Comment == nil || modules.ModuleInteraction.Comment.Desc == nil {
		return &card
	}
	for _, descItem := range modules.ModuleInteraction.Comment.Desc.RichTextNode {
		if descItem == nil {
			continue
		}
		if descItem.DescItemType == jsonwebcard.RichTextNodeTypeEmoji {
			emoji, ok := dynCtx.ResEmoji[descItem.Text]
			if !ok {
				descItem.DescItemType = jsonwebcard.RichTextNodeTypeText
				continue
			}
			descItem.RichTextNodeEmoji.IconUrl = emoji.URL
			descItem.RichTextNodeEmoji.Text = emoji.Text
			descItem.RichTextNodeEmoji.Size = int64(emoji.Meta.Size)
		}
	}

	return &card
}
