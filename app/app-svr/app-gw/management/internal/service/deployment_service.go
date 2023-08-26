package service

import "go-gateway/app/app-svr/app-gw/management/internal/dao"

// DeploymentService service.
type DeploymentService struct {
	*SnapshotService
	http   *HttpService
	grpc   *GrpcService
	common *CommonService
	dao    dao.Dao
}

func newDeployment(d dao.Dao) *DeploymentService {
	s := &DeploymentService{
		SnapshotService: newSnapshot(d.CreateSnapshotDao()),
		http:            newHTTP(d),
		grpc:            newGRPC(d),
		common:          newCommon(d),
		dao:             d,
	}
	return s
}
