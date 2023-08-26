package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type CarouselImg struct{}

func (r CarouselImg) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.CarouselImg{
		BaseCfgManager:    config.NewBaseCfg(natModule),
		ContentStyle:      natModule.AvSort,
		BgColor:           natModule.BgColor,
		IndicatorColor:    natModule.MoreColor,
		IsAutoCarousel:    natModule.IsAttrAutoPlay() == natpagegrpc.AttrModuleYes,
		IsTopTabFollowImg: natModule.IsAttrDisplayNum() == natpagegrpc.AttrModuleYes,
		IsTopTabFadeAway:  natModule.IsAttrDisplayDesc() == natpagegrpc.AttrModuleYes,
		ImageTitle:        natModule.Meta,
		Images:            r.buildImages(module.Carousel),
	}
	return cfg
}

func (r CarouselImg) buildImages(carousel *natpagegrpc.Carousel) []*config.CarouselImgItem {
	if carousel == nil {
		return nil
	}
	items := make([]*config.CarouselImgItem, 0, len(carousel.List))
	for _, ext := range carousel.List {
		if ext == nil || ext.Reason == "" {
			continue
		}
		if img, err := model.UnmarshalMixExtCarouselImg(ext.Reason); err == nil {
			items = append(items, img)
		}
	}
	return items
}
