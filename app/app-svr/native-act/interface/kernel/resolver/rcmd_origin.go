package resolver

import (
	"context"

	activitygrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

const (
	_rcmdRankMax = 10
)

type RcmdOrigin struct{}

func (r RcmdOrigin) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	confSort := natModule.ConfUnmarshal()
	cfg := &config.RcmdOrigin{
		BaseCfgManager:   config.NewBaseCfg(natModule),
		RcmdCommon:       buildRcmdCommon(natModule),
		SourceType:       confSort.SourceType,
		DisplayRankScore: natModule.IsAttrDisplayRecommend() == natpagegrpc.AttrModuleYes,
	}
	r.setBaseCfg(cfg, natModule, ss)
	return cfg
}

func (r RcmdOrigin) setBaseCfg(cfg *config.RcmdOrigin, module *natpagegrpc.NativeModule, ss *kernel.Session) {
	if module.Fid <= 0 {
		return
	}
	confSort := module.ConfUnmarshal()
	switch confSort.SourceType {
	case model.SourceTypeActUp:
		sortType := module.ConfUnmarshal().SortType
		if sortType == "" {
			sortType = model.SortTypeCtime
		}
		cfg.UpListReqID, _ = cfg.AddMaterialParam(model.MaterialUpListRly, &kernel.UpListReq{
			Req:         &activitygrpc.UpListReq{Sid: module.Fid, Type: sortType, Pn: 1, Ps: 40, Mid: ss.Mid()},
			NeedMultiML: true,
		})
	case model.SourceTypeRank:
		ps := module.Num
		if ps > _rcmdRankMax {
			ps = _rcmdRankMax
		}
		cfg.MixExtReqID, _ = cfg.AddMaterialParam(model.MaterialMixExtRly, &natpagegrpc.ModuleMixExtReq{
			ModuleID: cfg.ModuleBase().ModuleID,
			Ps:       ps,
			MType:    natpagegrpc.MixRankIcon,
		})
		cfg.RankRstReqID, _ = cfg.AddMaterialParam(model.MaterialRankRstRly, &kernel.RankResultReq{
			Req: &activitygrpc.RankResultReq{
				RankID: module.Fid,
				Pn:     1,
				Ps:     ps,
			},
			NeedMultiML: true,
		})
	default:
		log.Warn("unknown sourceType of RcmdOrigin, sourceType=%s module=%+v", confSort.SourceType, module)
		return
	}
}
