package resolver

import (
	"context"
	"strings"

	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

const (
	_colorParts = 2
)

type Navigation struct{}

func (r Navigation) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Navigation{
		BaseCfgManager: config.NewBaseCfg(natModule),
	}
	if natModule.BgColor != "" {
		cfg.UnselectedBgColor, cfg.NtUnselectedBgColor = r.extractColors(natModule.BgColor)
	}
	if natModule.FontColor != "" {
		cfg.UnselectedFontColor, cfg.NtUnselectedFontColor = r.extractColors(natModule.FontColor)
	}
	if natModule.TitleColor != "" {
		cfg.SelectedBgColor, cfg.NtSelectedBgColor = r.extractColors(natModule.TitleColor)
	}
	if natModule.MoreColor != "" {
		cfg.SelectedFontColor, cfg.NtSelectedFontColor = r.extractColors(natModule.MoreColor)
	}
	return cfg
}

func (r Navigation) extractColors(colors string) (day string, night string) {
	parts := strings.Split(colors, ",")
	if len(parts) >= 1 {
		day = parts[0]
	}
	if len(parts) >= _colorParts {
		night = parts[1]
	}
	return day, night
}
