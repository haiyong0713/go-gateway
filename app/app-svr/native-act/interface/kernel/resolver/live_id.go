package resolver

import (
	"context"
	"time"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type LiveID struct{}

func (r LiveID) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	nowTime := time.Now().Unix()
	// 不在设置的时间之内，不下发直播卡
	if natModule.Stime > nowTime || natModule.Etime < nowTime || natModule.Fid == 0 {
		return nil
	}
	ryColors := natModule.ColorsUnmarshal()
	cfg := &config.LiveID{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ImageTitle:     natModule.Meta,
		TextTitle:      natModule.Caption,
		ID:             natModule.Fid,
		Stime:          natModule.Stime,
		Cover:          natModule.TName,
		DisplayTitle:   !(natModule.IsAttrHideTitle() == natpagegrpc.AttrModuleYes),
		BgColor:        natModule.BgColor,
		FontColor:      natModule.FontColor,
		DisplayColor:   ryColors.DisplayColor,
		LiveType:       natModule.LiveType,
	}
	_, _ = cfg.BaseCfgManager.AddMaterialParam(model.MaterialLive, []int64{natModule.Fid})
	return cfg
}
