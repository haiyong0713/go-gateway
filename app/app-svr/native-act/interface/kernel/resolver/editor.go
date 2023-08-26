package resolver

import (
	"context"
	"encoding/json"
	"strconv"

	actplatv2grpc "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Editor struct{}

func (r Editor) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Editor{
		BaseCfgManager:    config.NewBaseCfg(natModule),
		Position:          buildEditorPosition(natModule),
		DisplayMoreButton: natModule.IsAttrDisplayOp() == natpagegrpc.AttrModuleYes,
		BgColor:           natModule.BgColor,
	}
	r.setBaseCfg(cfg, natModule, ss)
	return cfg
}

func buildEditorPosition(module *natpagegrpc.NativeModule) config.Position {
	pos := config.Position{}
	if module.TName == "" {
		return pos
	}
	if err := json.Unmarshal([]byte(module.TName), &pos); err != nil {
		log.Error("Fail to unmarshal editor position, t_name=%+v error=%+v", module.TName, err)
		return config.Position{}
	}
	return pos
}

func (r Editor) setBaseCfg(cfg *config.Editor, module *natpagegrpc.NativeModule, ss *kernel.Session) {
	cfg.MixExtsReqID, _ = cfg.AddMaterialParam(model.MaterialMixExtsRly, &kernel.ModuleMixExtsReq{
		Req:         &natpagegrpc.ModuleMixExtsReq{ModuleID: cfg.ModuleBase().ModuleID, Ps: module.Num},
		NeedMultiML: true,
		ArcType:     model.MaterialArchive,
	})
	func() {
		confSort := module.ConfUnmarshal()
		if confSort.Sid == 0 || confSort.Counter == "" || ss.Mid() == 0 {
			return
		}
		cfg.GetHisReqID, _ = cfg.AddMaterialParam(model.MaterialGetHisRly, &actplatv2grpc.GetHistoryReq{
			Activity: strconv.FormatInt(confSort.Sid, 10),
			Counter:  confSort.Counter,
			Mid:      ss.Mid(),
		})
	}()
}
