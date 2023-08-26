package builder

import (
	"context"
	"fmt"
	"strconv"

	articlemdl "git.bilibili.co/bapis/bapis-go/article/model"
	xtime "go-common/library/time"

	appcardmdl "go-gateway/app/app-svr/app-card/interface/model"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	"go-gateway/app/app-svr/native-act/interface/kernel/passthrough"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Timeline struct{}

func (bu Timeline) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	tlCfg, ok := cfg.(*config.Timeline)
	if !ok {
		logCfgAssertionError(config.Timeline{})
		return nil
	}
	hasMore, offset, items := bu.buildModuleItems(tlCfg, material, ss)
	if len(items) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeTimeline.String(),
		ModuleId:    tlCfg.ModuleBase().ModuleID,
		ModuleColor: &api.Color{BgColor: tlCfg.BgColor, CardBgColor: tlCfg.CardBgColor, TimelineColor: tlCfg.TimelineColor},
		ModuleUkey:  tlCfg.ModuleBase().Ukey,
		HasMore:     hasMore,
	}
	if tlCfg.ViewMoreType == model.TimelineMoreByExpand {
		items = bu.withUnfoldItems(items, tlCfg.ShowNum, tlCfg.ViewMoreText)
	} else if hasMore {
		if model.IsFromIndex(ss.ReqFrom) {
			items = append(items, bu.buildTimelineMore(tlCfg, offset))
		} else {
			module.SubpageParams = bu.supernatantParams(offset, -1, tlCfg.ModuleBase().ModuleID)
		}
	}
	module.ModuleItems = unshiftTitleCard(items, tlCfg.ImageTitle, tlCfg.TextTitle, ss.ReqFrom)
	return module
}

func (bu Timeline) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu Timeline) buildModuleItems(cfg *config.Timeline, material *kernel.Material, ss *kernel.Session) (bool, int64, []*api.ModuleItem) {
	mixRly, ok := material.MixExtsRlys[cfg.MixExtsReqID]
	if !ok || len(mixRly.List) == 0 {
		return false, 0, nil
	}
	items := make([]*api.ModuleItem, 0, len(mixRly.List)*2)
	var cnt int64
	offset := ss.Offset
	for _, ext := range mixRly.List {
		offset++
		if ext == nil {
			continue
		}
		remark := ext.RemarkUnmarshal()
		var item *api.ModuleItem
		switch ext.MType {
		case natpagegrpc.MixAvidType:
			arc, ok := material.Arcs[ext.ForeignID]
			if !ok || !arc.IsNormal() {
				continue
			}
			item = bu.buildArchive(arc, ss)
		case natpagegrpc.MixCvidType:
			art, ok := material.Articles[ext.ForeignID]
			if !ok || !art.IsNormal() {
				continue
			}
			item = bu.buildArticle(art)
		case natpagegrpc.MixTimelineText:
			item = bu.buildTimelineText(remark)
		case natpagegrpc.MixTimelinePic:
			item = bu.buildTimelineImage(remark)
		case natpagegrpc.MixTimeline:
			item = bu.buildTimelineImagetext(remark)
		default:
			continue
		}
		head := bu.buildTimelineHead(cfg, remark)
		items = append(items, head, item)
		cnt++
		if cnt >= cfg.Ps {
			break
		}
	}
	if len(items) == 0 {
		return false, 0, nil
	}
	hasMore := mixRly.HasMore == 1
	if !hasMore && offset < mixRly.Offset {
		hasMore = true
	}
	return hasMore, offset, items
}

func (bu Timeline) buildArchive(arc *arcgrpc.Arc, ss *kernel.Session) *api.ModuleItem {
	cd := &api.TimelineEventResourceCard{
		Title:         arc.GetTitle(),
		CoverImageUri: arc.GetPic(),
		Position1:     arc.GetAuthor().Name,
		Position2:     appcardmdl.StatString(arc.GetStat().View, "观看"),
		ReportDic:     &api.ReportDic{BizType: model.ReportBizTypeUGC},
		ResourceType:  model.ResourceTypeUGC,
	}
	cd.Uri = appcardmdl.FillURI(appcardmdl.GotoAv, ss.RawDevice().Plat(), int(ss.RawDevice().Build), strconv.FormatInt(arc.GetAid(), 10),
		appcardmdl.ArcPlayHandler(arc, nil, ss.TraceId(), nil, int(ss.RawDevice().Build), ss.RawDevice().RawMobiApp, false))
	return &api.ModuleItem{
		CardType:   model.CardTypeTimelineEventResource.String(),
		CardId:     strconv.FormatInt(arc.GetAid(), 10),
		CardDetail: &api.ModuleItem_TimelineEventResourceCard{TimelineEventResourceCard: cd},
	}
}

func (bu Timeline) buildArticle(art *articlemdl.Meta) *api.ModuleItem {
	cd := &api.TimelineEventResourceCard{
		Title:        art.Title,
		Uri:          fmt.Sprintf("https://www.bilibili.com/read/cv%d", art.ID),
		Position1:    art.Author.Name,
		Position2:    appcardmdl.Stat64String(art.Stats.View, "观看"),
		Badge:        &api.Badge{Text: "文章", BgColor: "#FB7299"},
		ReportDic:    &api.ReportDic{BizType: model.ReportBizTypeArticle},
		ResourceType: model.ResourceTypeArticle,
	}
	if len(art.ImageURLs) > 0 {
		cd.CoverImageUri = art.ImageURLs[0]
	}
	return &api.ModuleItem{
		CardType:   model.CardTypeTimelineEventResource.String(),
		CardId:     strconv.FormatInt(art.ID, 10),
		CardDetail: &api.ModuleItem_TimelineEventResourceCard{TimelineEventResourceCard: cd},
	}
}

func (bu Timeline) buildTimelineText(ext *natpagegrpc.MixReason) *api.ModuleItem {
	cd := &api.TimelineEventTextCard{
		Title:    ext.Title,
		Subtitle: ext.SubTitle,
		Content:  ext.Desc,
		Uri:      ext.Url,
	}
	return &api.ModuleItem{
		CardType:   model.CardTypeTimelineEventText.String(),
		CardDetail: &api.ModuleItem_TimelineEventTextCard{TimelineEventTextCard: cd},
	}
}

func (bu Timeline) buildTimelineImage(ext *natpagegrpc.MixReason) *api.ModuleItem {
	cd := &api.TimelineEventImageCard{
		Image: &api.SizeImage{
			Image:  ext.Image,
			Height: int64(ext.Length),
			Width:  int64(ext.Width),
		},
		Title: ext.Title,
		Uri:   ext.Url,
	}
	return &api.ModuleItem{
		CardType:   model.CardTypeTimelineEventImage.String(),
		CardDetail: &api.ModuleItem_TimelineEventImageCard{TimelineEventImageCard: cd},
	}
}

func (bu Timeline) buildTimelineImagetext(ext *natpagegrpc.MixReason) *api.ModuleItem {
	cd := &api.TimelineEventImagetextCard{
		Title:    ext.Title,
		Subtitle: ext.SubTitle,
		Content:  ext.Desc,
		Image:    ext.Image,
		Uri:      ext.Url,
	}
	return &api.ModuleItem{
		CardType:   model.CardTypeTimelineEventImagetext.String(),
		CardDetail: &api.ModuleItem_TimelineEventImagetextCard{TimelineEventImagetextCard: cd},
	}
}

func (bu Timeline) buildTimelineHead(cfg *config.Timeline, ext *natpagegrpc.MixReason) *api.ModuleItem {
	var cd *api.TimelineHeadCard
	if cfg.NodeType == model.TimelineNodeTime {
		cd = bu.buildHeadTime(xtime.Time(ext.Stime), cfg.TimePrecision)
	} else {
		cd = bu.buildHeadText(ext.Name)
	}
	return &api.ModuleItem{
		CardType:   model.CardTypeTimelineHead.String(),
		CardDetail: &api.ModuleItem_TimelineHeadCard{TimelineHeadCard: cd},
	}
}

func (bu Timeline) buildHeadText(stage string) *api.TimelineHeadCard {
	return &api.TimelineHeadCard{Stage: stage}
}

func (bu Timeline) buildHeadTime(stime xtime.Time, precision int64) *api.TimelineHeadCard {
	y, m, d := stime.Time().Date()
	h, min, sec := stime.Time().Clock()
	var stage string
	switch precision {
	case model.TimelineTimeMonth:
		stage = fmt.Sprintf("%d年%d月", y, m)
	case model.TimelineTimeDay:
		stage = fmt.Sprintf("%d年%d月%d日", y, m, d)
	case model.TimelineTimeHour:
		stage = fmt.Sprintf("%d年%d月%d日 %02d时", y, m, d, h)
	case model.TimelineTimeMin:
		stage = fmt.Sprintf("%d年%d月%d日 %02d:%02d", y, m, d, h, min)
	case model.TimelineTimeSec:
		stage = fmt.Sprintf("%d年%d月%d日 %02d:%02d:%02d", y, m, d, h, min, sec)
	default:
		stage = fmt.Sprintf("%d年", y)
	}
	return &api.TimelineHeadCard{Stage: stage}
}

func (bu Timeline) buildTimelineMore(cfg *config.Timeline, lastIndex int64) *api.ModuleItem {
	moreText := cfg.ViewMoreText
	if moreText == "" {
		moreText = "查看更多"
	}
	params := bu.supernatantParams(0, lastIndex, cfg.ModuleBase().ModuleID)
	return card.NewTimelineMore(moreText, cfg.SupernatantTitle, params).Build()
}

func (bu Timeline) supernatantParams(offset, lastIndex, moduleID int64) string {
	return passthrough.Marshal(&api.TimelineSupernatantParams{LastIndex: lastIndex, Offset: offset, ModuleId: moduleID})
}

func (bu Timeline) withUnfoldItems(items []*api.ModuleItem, showNum int64, viewMoreText string) []*api.ModuleItem {
	var showCardsNum = int(showNum * 2)
	if len(items) <= showCardsNum {
		return items
	}
	before := items[:showCardsNum]
	after := items[showCardsNum:]
	return append(before, bu.buildTimelineUnfold(viewMoreText, after))
}

func (bu Timeline) buildTimelineUnfold(unfoldText string, modules []*api.ModuleItem) *api.ModuleItem {
	if unfoldText == "" {
		unfoldText = "展开"
	}
	cd := &api.TimelineUnfoldCard{
		UnfoldText: unfoldText,
		FoldText:   "收起",
		Cards:      make([]*api.TimelineUnfoldCard_Card, 0, len(modules)),
	}
	for _, module := range modules {
		var unfoldCd *api.TimelineUnfoldCard_Card

		switch v := module.CardDetail.(type) {
		case *api.ModuleItem_TimelineHeadCard:
			unfoldCd = &api.TimelineUnfoldCard_Card{
				CardDetail: &api.TimelineUnfoldCard_Card_TimelineHeadCard{TimelineHeadCard: v.TimelineHeadCard},
			}
		case *api.ModuleItem_TimelineEventTextCard:
			unfoldCd = &api.TimelineUnfoldCard_Card{
				CardDetail: &api.TimelineUnfoldCard_Card_TimelineEventTextCard{TimelineEventTextCard: v.TimelineEventTextCard},
			}
		case *api.ModuleItem_TimelineEventImageCard:
			unfoldCd = &api.TimelineUnfoldCard_Card{
				CardDetail: &api.TimelineUnfoldCard_Card_TimelineEventImageCard{TimelineEventImageCard: v.TimelineEventImageCard},
			}
		case *api.ModuleItem_TimelineEventImagetextCard:
			unfoldCd = &api.TimelineUnfoldCard_Card{
				CardDetail: &api.TimelineUnfoldCard_Card_TimelineEventImagetextCard{TimelineEventImagetextCard: v.TimelineEventImagetextCard},
			}
		case *api.ModuleItem_TimelineEventResourceCard:
			unfoldCd = &api.TimelineUnfoldCard_Card{
				CardDetail: &api.TimelineUnfoldCard_Card_TimelineEventResourceCard{TimelineEventResourceCard: v.TimelineEventResourceCard},
			}
		default:
			continue
		}
		cd.Cards = append(cd.Cards, unfoldCd)
	}
	return &api.ModuleItem{
		CardType:   model.CardTypeTimelineUnfold.String(),
		CardDetail: &api.ModuleItem_TimelineUnfoldCard{TimelineUnfoldCard: cd},
	}
}
