package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type Rcmd struct{}

func (r Rcmd) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.Rcmd{
		BaseCfgManager: config.NewBaseCfg(natModule),
		RcmdCommon:     buildRcmdCommon(natModule),
		RcmdUsers:      r.rcmdUsers(module.Recommend),
	}
	r.setBaseCfg(cfg)
	return cfg
}

func (r Rcmd) rcmdUsers(rcmd *natpagegrpc.Recommend) []*config.RcmdUser {
	if rcmd == nil || len(rcmd.List) == 0 {
		return nil
	}
	users := make([]*config.RcmdUser, 0, len(rcmd.List))
	for _, ext := range rcmd.List {
		if ext == nil || ext.ForeignID == 0 {
			continue
		}
		users = append(users, &config.RcmdUser{Mid: ext.ForeignID, Reason: ext.Reason})
	}
	return users
}

func (r Rcmd) setBaseCfg(cfg *config.Rcmd) {
	var mids []int64
	for _, user := range cfg.RcmdUsers {
		mids = append(mids, user.Mid)
	}
	_, _ = cfg.AddMaterialParam(model.MaterialRelation, mids)
	_, _ = cfg.AddMaterialParam(model.MaterialAccountCard, mids)
}

func buildRcmdCommon(module *natpagegrpc.NativeModule) config.RcmdCommon {
	return config.RcmdCommon{
		ImageTitle:         module.Meta,
		BgColor:            module.BgColor,
		CardTitleFontColor: module.TitleColor,
	}
}
