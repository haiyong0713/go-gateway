package service

import (
	"context"

	appdyngrpc "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	"go-gateway/app/app-svr/native-act/interface/api"
	"go-gateway/app/app-svr/native-act/interface/internal/dao"
	"go-gateway/app/app-svr/native-act/interface/internal/model"
	"go-gateway/app/app-svr/native-act/interface/kernel"
	"go-gateway/app/app-svr/native-act/interface/kernel/config"
)

type IndexReqID struct {
	DynDetail kernel.RequestID
}

func loadMaterials(loader *kernel.MaterialLoader, cfgs []config.BaseCfgManager) *kernel.Material {
	if loader == nil {
		return nil
	}
	for _, cfg := range cfgs {
		loader.JoinLoader(cfg.MaterialParams())
	}
	return loader.MultiLoad()
}

func newIndexML(c context.Context, dep dao.Dependency, ss *kernel.Session, req *api.IndexReq) (*IndexReqID, *kernel.MaterialLoader) {
	ml := kernel.NewMaterialLoader(c, dep, ss)
	reqID := &IndexReqID{}
	if req.DynamicId > 0 && model.NeedLayerDynamic(req.ActivityFrom) {
		reqID.DynDetail, _ = ml.AddItem(model.MaterialDynDetail, &appdyngrpc.DynServerDetailsReq{DynamicIds: []int64{req.DynamicId}})
	}
	return reqID, ml
}
