package builder

import (
	"context"

	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type TimelineOrigin struct{}

func (bu TimelineOrigin) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	toCfg, ok := cfg.(*config.TimelineOrigin)
	if !ok {
		logCfgAssertionError(config.TimelineOrigin{})
		return nil
	}
	tlRly, ok := material.TimelineRlys[toCfg.TimelineReqID]
	if !ok || len(tlRly.GetEvents()) == 0 {
		return nil
	}
	items := bu.buildModuleItems(toCfg, material)
	if len(items) == 0 {
		return nil
	}
	items = unshiftTitleCard(Timeline{}.withUnfoldItems(items, toCfg.ShowNum, toCfg.ViewMoreText), toCfg.ImageTitle, toCfg.TextTitle, ss.ReqFrom)
	module := &api.Module{
		ModuleType:  model.ModuleTypeTimeline.String(),
		ModuleId:    toCfg.ModuleBase().ModuleID,
		ModuleColor: &api.Color{BgColor: toCfg.BgColor, CardBgColor: toCfg.CardBgColor, TimelineColor: toCfg.TimelineColor},
		ModuleItems: items,
		ModuleUkey:  toCfg.ModuleBase().Ukey,
	}
	return module
}

func (bu TimelineOrigin) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu TimelineOrigin) buildModuleItems(cfg *config.TimelineOrigin, material *kernel.Material) []*api.ModuleItem {
	tlRly, ok := material.TimelineRlys[cfg.TimelineReqID]
	if !ok || len(tlRly.Events) == 0 {
		return nil
	}
	items := make([]*api.ModuleItem, 0, len(tlRly.Events)*2)
	tlBuilder := Timeline{}
	for _, event := range tlRly.Events {
		if event == nil {
			continue
		}
		item := &api.ModuleItem{
			CardType: model.CardTypeTimelineEventImagetext.String(),
			CardDetail: &api.ModuleItem_TimelineEventImagetextCard{
				TimelineEventImagetextCard: bu.buildTimelineImagetext(event),
			},
		}
		headItem := &api.ModuleItem{
			CardType: model.CardTypeTimelineHead.String(),
			CardDetail: &api.ModuleItem_TimelineHeadCard{
				TimelineHeadCard: tlBuilder.buildHeadTime(event.GetStime(), cfg.TimePrecision),
			},
		}
		items = append(items, headItem, item)
	}
	return items
}

func (bu TimelineOrigin) buildTimelineImagetext(event *populargrpc.TimeEvent) *api.TimelineEventImagetextCard {
	return &api.TimelineEventImagetextCard{
		Content: event.GetTitle(),
		Image:   event.GetPic(),
		Uri:     event.GetJumpLink(),
	}
}
