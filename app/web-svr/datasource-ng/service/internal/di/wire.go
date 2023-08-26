//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	pb "go-gateway/app/web-svr/datasource-ng/service/api"
	"go-gateway/app/web-svr/datasource-ng/service/internal/dao"
	"go-gateway/app/web-svr/datasource-ng/service/internal/server/grpc"
	"go-gateway/app/web-svr/datasource-ng/service/internal/server/http"
	"go-gateway/app/web-svr/datasource-ng/service/internal/service"

	"github.com/google/wire"
)

var daoProvider = wire.NewSet(dao.New, dao.NewDB)
var serviceProvider = wire.NewSet(service.New, wire.Bind(new(pb.DataSourceNGServer), new(*service.Service)))

func InitApp() (*App, func(), error) {
	panic(wire.Build(daoProvider, serviceProvider, http.New, grpc.New, NewApp))
}
