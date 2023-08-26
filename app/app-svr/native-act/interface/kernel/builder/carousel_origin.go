package builder

import (
	"context"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type CarouselOrigin struct{}

func (bu CarouselOrigin) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	coCfg, ok := cfg.(*config.CarouselOrigin)
	if !ok {
		logCfgAssertionError(config.CarouselOrigin{})
		return nil
	}
	upListRly, ok := material.UpListRlys[coCfg.UpListReqID]
	if !ok || len(upListRly.List) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeCarouselImg.String(),
		ModuleId:      cfg.ModuleBase().ModuleID,
		ModuleColor:   &api.Color{BgColor: coCfg.BgColor, IndicatorColor: coCfg.IndicatorColor},
		ModuleItems:   bu.buildModuleItems(coCfg, upListRly.List),
		ModuleUkey:    cfg.ModuleBase().Ukey,
		ModuleSetting: &api.Setting{AutoCarousel: coCfg.IsAutoCarousel},
	}
	return module
}

func (bu CarouselOrigin) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu CarouselOrigin) buildModuleItems(cfg *config.CarouselOrigin, upList []*activitygrpc.UpListItem) []*api.ModuleItem {
	items := make([]*api.ModuleItem, 0, 2)
	if cfg.ContentStyle == model.CarouselCSSlide && cfg.ImageTitle != "" {
		items = append(items, card.NewImageTitle(cfg.ImageTitle).Build())
	}
	cd := &api.CarouselImgCard{
		ContentStyle: cfg.ContentStyle,
		Images:       make([]*api.CarouselImgItem, 0, len(upList)),
	}
	for _, upItem := range upList {
		if upItem.Content == nil {
			continue
		}
		cd.Images = append(cd.Images, &api.CarouselImgItem{
			Image:  upItem.Content.Image,
			Uri:    upItem.Content.Link,
			Height: cfg.ImgHeight,
			Width:  cfg.ImgWidth,
		})
	}
	items = append(items, &api.ModuleItem{
		CardType:   model.CardTypeCarouselImg.String(),
		CardDetail: &api.ModuleItem_CarouselImgCard{CarouselImgCard: cd},
	})
	return items
}
