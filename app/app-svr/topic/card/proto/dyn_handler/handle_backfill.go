package dynHandler

import (
	"fmt"
	"strconv"

	dynamicapi "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	dynmdlV2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	topiccardmodel "go-gateway/app/app-svr/topic/card/model"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	backfillVideo     = "common_video_icon"
	backfillArticle   = "common_article_icon"
	_handleTypeRepost = "repost"
	_descTitleLimit   = 20
)

func (schema *CardSchema) Backfill(dynCtx *dynmdlV2.DynamicContext, dynItem *dynamicapi.DynamicItem, _ *topiccardmodel.GeneralParam) {
	for _, item := range dynItem.Modules {
		switch item.ModuleType {
		case dynamicapi.DynModuleType_module_desc:
			module := item.ModuleItem.(*dynamicapi.Module_ModuleDesc)
			for _, descItem := range module.ModuleDesc.Desc {
				schema.formDescTtem(descItem, dynCtx)
			}
		case dynamicapi.DynModuleType_module_interaction:
			module := item.ModuleItem.(*dynamicapi.Module_ModuleInteraction)
			for _, interactionItem := range module.ModuleInteraction.InteractionItem {
				schema.backFillModuleIteraction(interactionItem, dynCtx)
			}
		case dynamicapi.DynModuleType_module_dynamic:
			if dynItem.CardType == dynamicapi.DynamicType_forward {
				module := item.ModuleItem.(*dynamicapi.Module_ModuleDynamic)
				forward := module.ModuleDynamic.ModuleItem.(*dynamicapi.ModuleDynamic_DynForward)
				for _, kernalItem := range forward.DynForward.Item.Modules {
					switch kernalItem.ModuleType {
					case dynamicapi.DynModuleType_module_desc:
						kernalModule := kernalItem.ModuleItem.(*dynamicapi.Module_ModuleDesc)
						for _, descItem := range kernalModule.ModuleDesc.Desc {
							schema.formDescTtem(descItem, dynCtx)
						}
					}
				}
			}
		}
	}
	// 快速转发原卡文案
	for _, descItem := range dynItem.Extend.OrigDesc {
		schema.formDescTtem(descItem, dynCtx)
	}
}

func (schema *CardSchema) backFillModuleIteraction(interactionItem *dynamicapi.InteractionItem, dynCtx *dynmdlV2.DynamicContext) {
	if interactionItem == nil {
		return
	}
	if interactionItem.IconType != dynamicapi.LocalIconType_local_icon_comment {
		return
	}
	for _, descItem := range interactionItem.Desc {
		if descItem.Type == dynamicapi.DescType_desc_type_emoji {
			emoji, ok := dynCtx.ResEmoji[descItem.Text]
			if !ok {
				descItem.Type = dynamicapi.DescType_desc_type_text
				continue
			}
			descItem.Uri = emoji.URL
			if dynCtx.From == _handleTypeRepost {
				descItem.EmojiSize = int32(emoji.Meta.Size)
			}
		}
	}
}

// nolint:gocognit
func (schema *CardSchema) formDescTtem(descItem *dynamicapi.Description, dynCtx *dynmdlV2.DynamicContext) {
	descItem.OrigText = descItem.Text
	switch descItem.Type {
	case dynamicapi.DescType_desc_type_emoji:
		emoji, ok := dynCtx.ResEmoji[descItem.Text]
		if !ok {
			descItem.Type = dynamicapi.DescType_desc_type_text
		} else {
			descItem.Uri = emoji.URL
			if dynCtx.From == _handleTypeRepost {
				descItem.EmojiSize = int32(emoji.Meta.Size)
			}
		}
	case dynamicapi.DescType_desc_type_av:
		if aid, _ := strconv.ParseInt(descItem.Rid, 10, 64); aid != 0 {
			if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
				var archive = ap.Arc
				descItem.OrigText = descItem.Text
				descItem.Text = schema.TitleLimit(archive.Title)
				descItem.IconName = backfillVideo
			}
		}
	case dynamicapi.DescType_desc_type_bv:
		if aid, _ := bvid.BvToAv(descItem.Rid); aid != 0 {
			if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
				var archive = ap.Arc
				descItem.OrigText = descItem.Text
				descItem.Text = schema.TitleLimit(archive.Title)
				descItem.IconName = backfillVideo
			}
		}
	case dynamicapi.DescType_desc_type_cv:
		if cvid, _ := strconv.ParseInt(descItem.Rid, 10, 64); cvid != 0 {
			if article, ok := dynCtx.ResBackfillArticle[cvid]; ok {
				descItem.OrigText = descItem.Text
				descItem.Text = schema.TitleLimit(article.Title)
				descItem.IconName = backfillArticle
			}
		}
	case dynamicapi.DescType_desc_type_web:
		if descURL, ok := dynCtx.BackfillDescURL[descItem.Uri]; ok && descURL != nil {
			if descURL.Type == dynamicapi.DescType_desc_type_av {
				if aid, _ := strconv.ParseInt(descURL.Rid, 10, 64); aid != 0 {
					if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
						var archive = ap.Arc
						descItem.OrigText = descItem.Text
						descItem.Text = schema.TitleLimit(archive.Title)
						descItem.IconName = backfillVideo
					}
				}
			}
			if descURL.Type == dynamicapi.DescType_desc_type_bv {
				if aid, _ := bvid.BvToAv(descURL.Rid); aid != 0 {
					if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
						var archive = ap.Arc
						descItem.OrigText = descItem.Text
						descItem.Text = schema.TitleLimit(archive.Title)
						descItem.IconName = backfillVideo
					}
				}
			}
			if descURL.Type == dynamicapi.DescType_desc_type_ogv_season {
				if ssid, _ := strconv.ParseInt(descURL.Rid, 10, 64); ssid != 0 {
					if season, ok := dynCtx.ResBackfillSeason[int32(ssid)]; ok {
						descItem.OrigText = descItem.Text
						descItem.Text = schema.TitleLimit(season.Title)
						descItem.IconName = backfillVideo
					}
				}
			}
			if descURL.Type == dynamicapi.DescType_desc_type_ogv_ep {
				if epid, _ := strconv.ParseInt(descURL.Rid, 10, 64); epid != 0 {
					if episode, ok := dynCtx.ResBackfillEpisode[int32(epid)]; ok && episode.Season != nil {
						descItem.OrigText = descItem.Text
						descItem.Text = schema.TitleLimit(episode.Season.Title)
						descItem.IconName = backfillVideo
					}
				}
			}
			if descURL.Type == dynamicapi.DescType_desc_type_cv {
				if cvid, _ := strconv.ParseInt(descURL.Rid, 10, 64); cvid != 0 {
					if article, ok := dynCtx.ResBackfillArticle[cvid]; ok {
						descItem.OrigText = descItem.Text
						descItem.Text = schema.TitleLimit(article.Title)
						descItem.IconName = backfillArticle
					}
				}
			}
		}
	case dynamicapi.DescType_desc_type_mail:
		// 兼容逻辑 客户端暂不支持mail类型
		descItem.Type = dynamicapi.DescType_desc_type_text
	default:
	}
}

func (schema *CardSchema) TitleLimit(title string) string {
	if tmp := []rune(title); len(tmp) > _descTitleLimit {
		return fmt.Sprintf("%v...", string(tmp[:_descTitleLimit]))
	}
	return title
}
