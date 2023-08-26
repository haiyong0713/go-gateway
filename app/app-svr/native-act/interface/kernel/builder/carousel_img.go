package builder

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/builder/card"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type CarouselImg struct{}

func (bu CarouselImg) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	ciCfg, ok := cfg.(*config.CarouselImg)
	if !ok {
		logCfgAssertionError(config.CarouselImg{})
		return nil
	}
	if len(ciCfg.Images) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:    model.ModuleTypeCarouselImg.String(),
		ModuleId:      cfg.ModuleBase().ModuleID,
		ModuleColor:   &api.Color{BgColor: ciCfg.BgColor, IndicatorColor: ciCfg.IndicatorColor},
		ModuleItems:   bu.buildModuleItems(ciCfg),
		ModuleUkey:    cfg.ModuleBase().Ukey,
		ModuleSetting: &api.Setting{AutoCarousel: ciCfg.IsAutoCarousel, TopTabFollowImg: ciCfg.IsTopTabFollowImg, TopTabFadeAway: ciCfg.IsTopTabFadeAway},
	}
	return module
}

func (bu CarouselImg) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu CarouselImg) buildModuleItems(cfg *config.CarouselImg) []*api.ModuleItem {
	items := make([]*api.ModuleItem, 0, 2)
	if cfg.ContentStyle == model.CarouselCSSlide && cfg.ImageTitle != "" {
		items = append(items, card.NewImageTitle(cfg.ImageTitle).Build())
	}
	cd := &api.CarouselImgCard{
		ContentStyle: cfg.ContentStyle,
		Images:       make([]*api.CarouselImgItem, 0, len(cfg.Images)),
	}
	for _, image := range cfg.Images {
		imgItem := &api.CarouselImgItem{
			Image:  image.ImgUrl,
			Uri:    image.RedirectUrl,
			Height: image.Length,
			Width:  image.Width,
		}
		switch image.BgType {
		case model.TopTabBgImg:
			imgItem.TopTab = &api.TopTab{
				BgImage1:  image.BgImage1,
				BgImage2:  image.BgImage2,
				FontColor: image.FontColor,
				BarType:   image.BarType,
			}
		case model.TopTabBgColor:
			imgItem.TopTab = &api.TopTab{
				TabTopColor:    image.TabTopColor,
				TabMiddleColor: image.TabMiddleColor,
				TabBottomColor: image.TabBottomColor,
				FontColor:      image.FontColor,
				BarType:        image.BarType,
			}
		}
		cd.Images = append(cd.Images, imgItem)
	}
	items = append(items, &api.ModuleItem{
		CardType:   model.CardTypeCarouselImg.String(),
		CardDetail: &api.ModuleItem_CarouselImgCard{CarouselImgCard: cd},
	})
	return items
}
