package resolver

import (
	"context"

	pgcappgrpc "git.bilibili.co/bapis/bapis-go/pgc/service/card/app"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type OgvOrigin struct{}

func (r OgvOrigin) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	cfg := &config.OgvOrigin{
		BaseCfgManager: config.NewBaseCfg(natModule),
		OgvCommon:      buildOgvCommon(natModule, ss),
		PlaylistID:     int32(natModule.Fid),
	}
	r.setBaseCfg(cfg, ss)
	return cfg
}

func (r OgvOrigin) setBaseCfg(cfg *config.OgvOrigin, ss *kernel.Session) {
	if cfg.PlaylistID <= 0 {
		return
	}
	cfg.SeasonByPlayIdReq, _ = cfg.AddMaterialParam(model.MaterialSeasonByPlayIdRly, &pgcappgrpc.SeasonByPlayIdReq{
		PlaylistId: cfg.PlaylistID,
		Offset:     int32(ss.Offset),
		PageSize:   int32(cfg.Ps),
		User:       &pgcappgrpc.UserReq{Mid: ss.Mid()},
	})
}
