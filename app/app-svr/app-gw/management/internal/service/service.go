package service

import (
	"context"

	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/dao"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.ManagementServer), new(*Service)))

// Service service.
type Service struct {
	pb.ManagementServer

	Snapshot *SnapshotService
	Deploy   *DeploymentService
	Common   *CommonService
	HTTP     *HttpService
	GRPC     *GrpcService

	dao dao.Dao
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		Snapshot: newSnapshot(d.CreateSnapshotDao()),
		Deploy:   newDeployment(d),
		Common:   newCommon(d),
		HTTP:     newHTTP(d),
		GRPC:     newGRPC(d),
		dao:      d,
	}
	cf = s.Close
	s.Common.initialTokenSecret(context.Background())
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}
