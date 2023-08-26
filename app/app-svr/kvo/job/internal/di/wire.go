//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	"go-gateway/app/app-svr/kvo/job/internal/dao"
	"go-gateway/app/app-svr/kvo/job/internal/server/http"
	"go-gateway/app/app-svr/kvo/job/internal/service"

	"github.com/google/wire"
)

type defaultHttp interface {
}

var daoProvider = wire.NewSet(dao.New, dao.NewDB, dao.NewRedis)
var serviceProvider = wire.NewSet(service.New, wire.Bind(new(defaultHttp), new(*service.Service)))

func InitApp() (*App, func(), error) {
	panic(wire.Build(daoProvider, serviceProvider, http.New, NewApp))
}
