package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Icon struct{}

func (r Icon) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Icon{
		BaseCfgManager: config.NewBaseCfg(natModule),
		BgColor:        natModule.BgColor,
		FontColor:      natModule.FontColor,
		Items:          r.iconItems(module.Icon),
	}
	return cfg
}

func (r Icon) iconItems(iconRly *natpagegrpc.Icon) []*config.IconItem {
	if iconRly == nil {
		return nil
	}
	items := make([]*config.IconItem, 0, len(iconRly.List))
	for _, ext := range iconRly.List {
		if ext == nil {
			continue
		}
		if iconExt, err := model.UnmarshalIconExt(ext.Reason); err == nil {
			items = append(items, &config.IconItem{
				Title: iconExt.Content,
				Image: iconExt.ImgUrl,
				Uri:   iconExt.RedirectUrl,
			})
		}
	}
	return items
}
