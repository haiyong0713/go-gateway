package resolver

import (
	"context"
	"strconv"

	liveplaygrpc "git.bilibili.co/bapis/bapis-go/live/live-play/v1"
	"go-common/library/log"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

type ResourceOrigin struct{}

func (r ResourceOrigin) Resolve(c context.Context, ss *kernel.Session, natPage *natpagegrpc.NativePage, module *natpagegrpc.Module) config.BaseCfgManager {
	natModule := module.NativeModule
	confSort := natModule.ConfUnmarshal()
	cfg := &config.ResourceOrigin{
		BaseCfgManager: config.NewBaseCfg(natModule),
		ResourceCommon: buildResourceCommon(natModule, ss),
		OriginType:     confSort.RdbType,
		ShowNum:        natModule.Num,
		TabID:          actSortType(module, ss),
		TabList:        actSortList(module),
	}
	r.setBaseCfg(cfg, natModule, ss)
	return cfg
}

func (r ResourceOrigin) setBaseCfg(cfg *config.ResourceOrigin, module *natpagegrpc.NativeModule, ss *kernel.Session) {
	if module.TName == "" {
		return
	}
	confSort := module.ConfUnmarshal()
	switch confSort.RdbType {
	case model.RDBOgv:
		wid, err := strconv.ParseInt(module.TName, 10, 32)
		if err != nil {
			log.Error("Fail to ParseInt wid, wid=%s error=%+v", module.TName, err)
			return
		}
		cfg.Wid = int32(wid)
		_, _ = cfg.AddMaterialParam(model.MaterialQueryWidRly, []int32{cfg.Wid})
	case model.RDBLive:
		actID, err := strconv.ParseInt(module.TName, 10, 64)
		if err != nil {
			log.Error("Fail to ParseInt actID, actID=%s error=%+v", module.TName, actID)
			return
		}
		var isLive int64
		if cfg.DisplayOnlyLive {
			isLive = 1
		}
		cfg.RoomsByActIdReqID, _ = cfg.AddMaterialParam(model.MaterialRoomsByActIdRly, &liveplaygrpc.GetListByActIdReq{
			ActId:    actID,
			TabId:    cfg.TabID,
			Filter:   &liveplaygrpc.Filter{IsLive: isLive},
			PageSize: cfg.Ps,
			Offset:   ss.Offset,
		})
	case model.RDBBizCommodity:
		cfg.ProductDetailReqID, _ = cfg.AddMaterialParam(model.MaterialProductDetail, &model.ProductDetailReq{
			SourceId: module.TName,
			Offset:   ss.Offset,
			Size:     cfg.Ps,
		})
	case model.RDBBizIds:
		cfg.SourceDetailReqID, _ = cfg.AddMaterialParam(model.MaterialSourceDetail, &kernel.SourceDetailReq{
			Req: &model.SourceDetailReq{
				SourceId: module.TName,
				Offset:   ss.Offset,
				Size:     cfg.Ps,
			},
			NeedMultiML: true,
		})
	default:
		log.Warn("unknown RdbType of ResourceOrigin, RdbType=%d", confSort.RdbType)
	}
}
