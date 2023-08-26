//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	"go-gateway/app/web-svr/web-goblin/admin/internal/dao"
	"go-gateway/app/web-svr/web-goblin/admin/internal/server/http"
	"go-gateway/app/web-svr/web-goblin/admin/internal/service"

	"github.com/google/wire"
)

var daoProvider = wire.NewSet(dao.New, dao.NewDB, dao.NewRedis, dao.NewMC)
var serviceProvider = wire.NewSet(service.New)

func InitApp() (*App, func(), error) {
	panic(wire.Build(daoProvider, serviceProvider, http.New, NewApp))
}
