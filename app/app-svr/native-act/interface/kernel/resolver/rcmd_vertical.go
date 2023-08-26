package resolver

import (
	"context"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type RcmdVertical struct{}

func (r RcmdVertical) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.RcmdVertical{
		BaseCfgManager: config.NewBaseCfg(natModule),
		RcmdCommon:     buildRcmdCommon(natModule),
		RcmdUsers:      r.rcmdUsers(module.Recommend),
	}
	r.setBaseCfg(cfg)
	return cfg
}

func (r RcmdVertical) rcmdUsers(rcmd *natpagegrpc.Recommend) []*config.RcmdUser {
	if rcmd == nil || len(rcmd.List) == 0 {
		return nil
	}
	users := make([]*config.RcmdUser, 0, len(rcmd.List))
	for _, ext := range rcmd.List {
		if ext == nil || ext.ForeignID == 0 {
			continue
		}
		rcmdUser := &config.RcmdUser{Mid: ext.ForeignID}
		if ext.Reason != "" {
			if userExt, err := model.UnmarshalRcmdVerticalExt(ext.Reason); err == nil {
				rcmdUser.Reason = userExt.Reason
				rcmdUser.Uri = userExt.Uri
			}
		}
		users = append(users, rcmdUser)
	}
	return users
}

func (r RcmdVertical) setBaseCfg(cfg *config.RcmdVertical) {
	var mids []int64
	for _, user := range cfg.RcmdUsers {
		mids = append(mids, user.Mid)
	}
	_, _ = cfg.AddMaterialParam(model.MaterialRelation, mids)
	_, _ = cfg.AddMaterialParam(model.MaterialAccountCard, mids)
}
