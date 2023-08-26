package service

import (
	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
	managementjob "go-gateway/app/app-svr/app-gw/management-job/api"
	"go-gateway/app/app-svr/app-gw/management/internal/dao"
)

// CommonService service.
type CommonService struct {
	ac            *paladin.Map
	dao           dao.Dao
	role          *RoleManager
	managementjob managementjob.ManagementJobClient
}

// newCommon new a common service and return.
func newCommon(d dao.Dao) *CommonService {
	GRPCClient := struct {
		ManagementJob *warden.ClientConfig
	}{}
	s := &CommonService{
		ac:  &paladin.TOML{},
		dao: d,
	}
	if err := paladin.Watch("application.toml", s.ac); err != nil {
		panic(err)
	}
	s.role = newRoleManager(s.ac, s.authZByRole)
	if err := paladin.Get("grpc.toml").UnmarshalTOML(&GRPCClient); err != nil {
		panic(err)
	}
	managementjob, err := managementjob.NewClient(GRPCClient.ManagementJob)
	if err != nil {
		panic(err)
	}
	s.managementjob = managementjob
	return s
}
