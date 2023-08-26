package resolver

import (
	"context"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type DynamicAct struct{}

func (r DynamicAct) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.DynamicAct{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ImageTitle:     natModule.Meta,
		TextTitle:      natModule.Caption,
		IsFeed:         natModule.IsAttrLast() == natpagegrpc.AttrModuleYes,
		Sid:            natModule.Fid,
		SortType:       actSortType(module, ss),
		SortList:       actSortList(module),
		SubpageTitle:   natModule.Title,
	}
	r.setColor(cfg, natModule)
	r.setPageSize(cfg, natModule, ss)
	r.setBaseCfg(cfg, ss)
	r.setOldVersion(cfg, natPage)
	return cfg
}

func (r DynamicAct) setColor(cfg *config.DynamicAct, natModule *natpagegrpc.NativeModule) {
	colors := natModule.ColorsUnmarshal()
	cfg.BgColor = natModule.BgColor
	cfg.FontColor = colors.DisplayColor
}

func (r DynamicAct) setPageSize(cfg *config.DynamicAct, natModule *natpagegrpc.NativeModule, ss *kernel.Session) {
	if ss.ReqFrom == model.ReqFromSubPage {
		cfg.PageSize = 10
		return
	}
	pageSize := int32(natModule.Num)
	if cfg.IsFeed {
		pageSize = 5
	}
	cfg.PageSize = pageSize
}

func (r DynamicAct) setBaseCfg(cfg *config.DynamicAct, ss *kernel.Session) {
	req := &kernel.ActLikesReq{
		Req: &activitygrpc.ActLikesReq{
			Sid:      cfg.Sid,
			Mid:      ss.Mid(),
			SortType: int32(cfg.SortType),
			Ps:       cfg.PageSize,
			Offset:   ss.Offset,
		},
		ArcType: model.MaterialArchive,
	}
	cfg.ActLikesReqID, _ = cfg.BaseCfgManager.AddMaterialParam(model.MaterialActLikesRly, req)
}

func (r DynamicAct) setOldVersion(cfg *config.DynamicAct, natPage *natpagegrpc.NativePage) {
	cfg.PageID = natPage.ID
}
