//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	pb "go-gateway/app/web-svr/datasource-ng/admin/api"
	"go-gateway/app/web-svr/datasource-ng/admin/internal/dao"
	"go-gateway/app/web-svr/datasource-ng/admin/internal/server/grpc"
	"go-gateway/app/web-svr/datasource-ng/admin/internal/server/http"
	"go-gateway/app/web-svr/datasource-ng/admin/internal/service"

	"github.com/google/wire"
)

var daoProvider = wire.NewSet(dao.New, dao.NewDB, dao.NewRedis, dao.NewMC)
var serviceProvider = wire.NewSet(service.New, wire.Bind(new(pb.DemoServer), new(*service.Service)))

func InitApp() (*App, func(), error) {
	panic(wire.Build(daoProvider, serviceProvider, http.New, grpc.New, NewApp))
}
