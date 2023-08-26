package dynamicV2

import (
	"context"
	"fmt"
	"strconv"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	backfillVideo   = "common_video_icon"
	backfillArticle = "common_article_icon"
)

// nolint:gocognit
func (s *Service) backfill(_ context.Context, dynCtx *mdlv2.DynamicContext, dynItem *api.DynamicItem, _ *mdlv2.GeneralParam) {
	for _, item := range dynItem.Modules {
		//nolint:exhaustive
		switch item.ModuleType {
		case api.DynModuleType_module_desc:
			module := item.ModuleItem.(*api.Module_ModuleDesc)
			for _, descItem := range module.ModuleDesc.Desc {
				s.formDescTtem(descItem, dynCtx)
			}
		case api.DynModuleType_module_interaction:
			module := item.ModuleItem.(*api.Module_ModuleInteraction)
			for _, interactionItem := range module.ModuleInteraction.InteractionItem {
				if interactionItem == nil {
					continue
				}
				if interactionItem.IconType == api.LocalIconType_local_icon_comment {
					for _, descItem := range interactionItem.Desc {
						if descItem.Type == api.DescType_desc_type_emoji {
							emoji, ok := dynCtx.ResEmoji[descItem.Text]
							if !ok {
								descItem.Type = api.DescType_desc_type_text
								continue
							}
							descItem.Uri = emoji.URL
							if dynCtx.From == _handleTypeRepost {
								descItem.EmojiSize = int32(emoji.Meta.Size)
							}
						}
					}
				}
			}
		case api.DynModuleType_module_dynamic:
			if dynItem.CardType == api.DynamicType_forward {
				module := item.ModuleItem.(*api.Module_ModuleDynamic)
				forward := module.ModuleDynamic.ModuleItem.(*api.ModuleDynamic_DynForward)
				for _, kernalItem := range forward.DynForward.Item.Modules {
					//nolint:exhaustive
					switch kernalItem.ModuleType {
					case api.DynModuleType_module_desc:
						kernalModule := kernalItem.ModuleItem.(*api.Module_ModuleDesc)
						for _, descItem := range kernalModule.ModuleDesc.Desc {
							s.formDescTtem(descItem, dynCtx)
						}
					}
				}
			}
		}
	}
	// 快速转发原卡文案
	for _, descItem := range dynItem.Extend.OrigDesc {
		s.formDescTtem(descItem, dynCtx)
	}
}

// nolint:gocognit
func (s *Service) formDescTtem(descItem *api.Description, dynCtx *mdlv2.DynamicContext) {
	descItem.OrigText = descItem.Text
	if descItem.Type == api.DescType_desc_type_emoji {
		emoji, ok := dynCtx.ResEmoji[descItem.Text]
		if !ok {
			descItem.Type = api.DescType_desc_type_text
		} else {
			descItem.Uri = emoji.URL
			if dynCtx.From == _handleTypeRepost {
				descItem.EmojiSize = int32(emoji.Meta.Size)
			}
		}
	}
	if descItem.Type == api.DescType_desc_type_av {
		if aid, _ := strconv.ParseInt(descItem.Rid, 10, 64); aid != 0 {
			if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
				var archive = ap.Arc
				descItem.OrigText = descItem.Text
				descItem.Text = s.TitleLimit(archive.Title)
				descItem.IconName = backfillVideo
			}
		}
	}
	if descItem.Type == api.DescType_desc_type_bv {
		if aid, _ := bvid.BvToAv(descItem.Rid); aid != 0 {
			if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
				var archive = ap.Arc
				descItem.OrigText = descItem.Text
				descItem.Text = s.TitleLimit(archive.Title)
				descItem.IconName = backfillVideo
			}
		}
	}
	if descItem.Type == api.DescType_desc_type_cv {
		if cvid, _ := strconv.ParseInt(descItem.Rid, 10, 64); cvid != 0 {
			if article, ok := dynCtx.ResBackfillArticle[cvid]; ok {
				descItem.OrigText = descItem.Text
				descItem.Text = s.TitleLimit(article.Title)
				descItem.IconName = backfillArticle
			}
		}
	}
	if descItem.Type == api.DescType_desc_type_web {
		if descURL, ok := dynCtx.BackfillDescURL[descItem.Uri]; ok && descURL != nil {
			if descURL.Type == api.DescType_desc_type_av {
				if aid, _ := strconv.ParseInt(descURL.Rid, 10, 64); aid != 0 {
					if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
						var archive = ap.Arc
						descItem.OrigText = descItem.Text
						descItem.Text = s.TitleLimit(archive.Title)
						descItem.IconName = backfillVideo
					}
				}
			}
			if descURL.Type == api.DescType_desc_type_bv {
				if aid, _ := bvid.BvToAv(descURL.Rid); aid != 0 {
					if ap, ok := dynCtx.ResBackfillArchive[aid]; ok {
						var archive = ap.Arc
						descItem.OrigText = descItem.Text
						descItem.Text = s.TitleLimit(archive.Title)
						descItem.IconName = backfillVideo
					}
				}
			}
			if descURL.Type == api.DescType_desc_type_ogv_season {
				if ssid, _ := strconv.ParseInt(descURL.Rid, 10, 64); ssid != 0 {
					if season, ok := dynCtx.ResBackfillSeason[int32(ssid)]; ok {
						descItem.OrigText = descItem.Text
						descItem.Text = s.TitleLimit(season.Title)
						descItem.IconName = backfillVideo
					}
				}
			}
			if descURL.Type == api.DescType_desc_type_ogv_ep {
				if epid, _ := strconv.ParseInt(descURL.Rid, 10, 64); epid != 0 {
					if episode, ok := dynCtx.ResBackfillEpisode[int32(epid)]; ok && episode.Season != nil {
						descItem.OrigText = descItem.Text
						descItem.Text = s.TitleLimit(episode.Season.Title)
						descItem.IconName = backfillVideo
					}
				}
			}
			if descURL.Type == api.DescType_desc_type_cv {
				if cvid, _ := strconv.ParseInt(descURL.Rid, 10, 64); cvid != 0 {
					if article, ok := dynCtx.ResBackfillArticle[cvid]; ok {
						descItem.OrigText = descItem.Text
						descItem.Text = s.TitleLimit(article.Title)
						descItem.IconName = backfillArticle
					}
				}
			}
		}
	}
	// 兼容逻辑 客户端暂不支持mail类型
	if descItem.Type == api.DescType_desc_type_mail {
		descItem.Type = api.DescType_desc_type_text
	}
}

func (s *Service) TitleLimit(title string) string {
	if tmp := []rune(title); s.c.Ctrl.DescTitleLimit != 0 && len(tmp) > s.c.Ctrl.DescTitleLimit {
		return fmt.Sprintf("%v...", string(tmp[:s.c.Ctrl.DescTitleLimit]))
	}
	return title
}
