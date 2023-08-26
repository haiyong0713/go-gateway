package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type CarouselWord struct{}

func (r CarouselWord) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.CarouselWord{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ContentStyle:   natModule.AvSort,
		BgColor:        natModule.BgColor,
		FontColor:      natModule.FontColor,
		CardBgColor:    natModule.TitleColor,
		ScrollType:     int64(natModule.DySort),
		Words:          r.buildWords(module.Carousel),
	}
	return cfg
}

func (r CarouselWord) buildWords(carousel *natpagegrpc.Carousel) []*config.CarouselWordItem {
	if carousel == nil {
		return nil
	}
	items := make([]*config.CarouselWordItem, 0, len(carousel.List))
	for _, ext := range carousel.List {
		if ext == nil || ext.Reason == "" {
			continue
		}
		if word, err := model.UnmarshalMixExtCarouselWord(ext.Reason); err == nil {
			items = append(items, word)
		}
	}
	return items
}
