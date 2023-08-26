package resolver

import (
	"context"

	populargrpc "git.bilibili.co/bapis/bapis-go/manager/service/popular"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type TimelineOrigin struct{}

func (r TimelineOrigin) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	confSort := natModule.ConfUnmarshal()
	colors := natModule.ColorsUnmarshal()
	cfg := &config.TimelineOrigin{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ImageTitle:     natModule.Meta,
		TextTitle:      natModule.Caption,
		TimePrecision:  confSort.TimeSort,
		BgColor:        natModule.BgColor,
		CardBgColor:    colors.TitleBgColor,
		TimelineColor:  colors.TimelineColor,
		ShowNum:        natModule.Num,
		ViewMoreText:   natModule.Remark,
		Ps:             50, //固定为下拉展示方式
	}
	r.setBaseCfg(cfg, natModule, ss)
	return cfg
}

func (r TimelineOrigin) setBaseCfg(cfg *config.TimelineOrigin, module *natpagegrpc.NativeModule, ss *kernel.Session) {
	if module.Fid <= 0 {
		return
	}
	cfg.TimelineReqID, _ = cfg.AddMaterialParam(model.MaterialTimelineRly, &populargrpc.TimeLineRequest{
		LineId: module.Fid,
		Ps:     int32(cfg.Ps),
		Offset: int32(ss.Offset),
	})
}
