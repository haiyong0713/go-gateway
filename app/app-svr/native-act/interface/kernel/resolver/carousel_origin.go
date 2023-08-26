package resolver

import (
	"context"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type CarouselOrigin struct{}

func (r CarouselOrigin) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.CarouselOrigin{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ContentStyle:   natModule.AvSort,
		BgColor:        natModule.BgColor,
		IndicatorColor: natModule.MoreColor,
		IsAutoCarousel: natModule.IsAttrAutoPlay() == natpagegrpc.AttrModuleYes,
		ImageTitle:     natModule.Meta,
		ImgHeight:      natModule.Length,
		ImgWidth:       natModule.Width,
	}
	r.setBaseCfg(cfg, natModule, ss)
	return cfg
}

func (r CarouselOrigin) setBaseCfg(cfg *config.CarouselOrigin, module *natpagegrpc.NativeModule, ss *kernel.Session) {
	if module.Fid <= 0 {
		return
	}
	confSort := module.ConfUnmarshal()
	if confSort.SourceType != model.SourceTypeActUp {
		return
	}
	sortType := model.SortTypeCtime
	if confSort.SortType != "" {
		sortType = confSort.SortType
	}
	cfg.UpListReqID, _ = cfg.AddMaterialParam(model.MaterialUpListRly, &kernel.UpListReq{
		Req: &activitygrpc.UpListReq{Sid: module.Fid, Type: sortType, Pn: 1, Ps: 8, Mid: ss.Mid()},
	})
}
