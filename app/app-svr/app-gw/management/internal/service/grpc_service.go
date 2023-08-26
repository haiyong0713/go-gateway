package service

import (
	"go-gateway/app/app-svr/app-gw/management/internal/dao"
)

// GrpcService service.
type GrpcService struct {
	common      *CommonService
	dao         dao.Dao
	resourceDao dao.ResourceDao
}

func newGRPC(d dao.Dao) *GrpcService {
	s := &GrpcService{
		common:      newCommon(d),
		dao:         d,
		resourceDao: d.CreateGRPCResourceDao(),
	}
	return s
}
