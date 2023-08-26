package builder

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type CarouselWord struct{}

func (bu CarouselWord) Build(c context.Context, ss *kernel.Session, dep dao.Dependency, cfg config.BaseCfgManager, material *kernel.Material) *api.Module {
	cwCfg, ok := cfg.(*config.CarouselWord)
	if !ok {
		logCfgAssertionError(config.CarouselWord{})
		return nil
	}
	if len(cwCfg.Words) == 0 {
		return nil
	}
	module := &api.Module{
		ModuleType:  model.ModuleTypeCarouselWord.String(),
		ModuleId:    cfg.ModuleBase().ModuleID,
		ModuleColor: bu.buildModuleColor(cwCfg),
		ModuleItems: bu.buildModuleItems(cwCfg),
		ModuleUkey:  cfg.ModuleBase().Ukey,
	}
	return module
}

func (bu CarouselWord) After(data *AfterContextData, current *api.Module) bool {
	return true
}

func (bu CarouselWord) buildModuleColor(cfg *config.CarouselWord) *api.Color {
	color := &api.Color{BgColor: cfg.BgColor, FontColor: cfg.FontColor}
	if cfg.ContentStyle == model.CarouselCSMultiLine {
		color.CardBgColor = cfg.CardBgColor
	}
	return color
}

func (bu CarouselWord) buildModuleItems(cfg *config.CarouselWord) []*api.ModuleItem {
	if cfg.ContentStyle == model.CarouselCSMultiLine {
		cfg.Words = cfg.Words[:1]
	}
	cd := &api.CarouselWordCard{
		ContentStyle: cfg.ContentStyle,
		Words:        make([]*api.CarouselWordItem, 0, len(cfg.Words)),
	}
	if cfg.ContentStyle == model.CarouselCSSingleLine {
		cd.ScrollType = cfg.ScrollType
	}
	for _, word := range cfg.Words {
		cd.Words = append(cd.Words, &api.CarouselWordItem{
			Content: word.Content,
		})
	}
	return []*api.ModuleItem{{
		CardType:   model.CardTypeCarouselWord.String(),
		CardDetail: &api.ModuleItem_CarouselWordCard{CarouselWordCard: cd},
	}}
}
