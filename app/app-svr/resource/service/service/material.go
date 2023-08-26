package service

import (
	"context"
	"go-common/library/ecode"
	"go-common/library/log"
	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
)

const (
	_materialCacheKey    = "tianma_material_cache"
	_materialCacheExpire = 86400 //redis缓存超时时间
)

func (s *Service) loadMaterialCache() {
	var (
		materialMapCacheTmp map[int64]*pb2.Material
		err                 error
	)

	materialMapCacheTmp, err = s.show.GetMaterialMap(context.Background())
	if err != nil {
		log.Error("service.loadMaterialCache GetMaterialMap err(%+v)", err)
		return
	}
	if len(materialMapCacheTmp) == 0 {
		log.Error("service.loadMaterialCache empty")
		return
	}
	s.materialMapCache = materialMapCacheTmp
	if err = s.show.SetMaterial2Cache(context.Background(), _materialCacheKey, _materialCacheExpire, materialMapCacheTmp); err != nil {
		log.Error("service.loadMaterialCache SetMaterial2Cache err(%+v)", err)
		return
	}
	log.Warn("service.loadMaterialCache success")
}

func (s *Service) GetMaterial(c context.Context, req *pb2.MaterialReq) (res *pb2.MaterialResp, err error) {
	res = &pb2.MaterialResp{}
	if req.Id == nil || len(req.Id) == 0 {
		err = ecode.RequestErr
		return
	}

	if len(s.materialMapCache) == 0 {
		s.materialMapCache, err = s.show.GetMaterialFromCache(context.Background(), _materialCacheKey)
		if err != nil {
			log.Error("service.GetMaterial GetMaterialFromCache err(%+v) key(%+v)", err, _materialCacheKey)
		}
		if len(s.materialMapCache) == 0 {
			log.Error("service.GetMaterial materialMapCache is nil")
			err = ecode.ServerErr
			return
		}
	}

	for _, v := range req.Id {
		if material, ok := s.materialMapCache[v]; ok {
			res.Material = append(res.Material, material)
		}
	}

	if len(res.Material) == 0 {
		err = ecode.NothingFound
		return
	}

	return
}
