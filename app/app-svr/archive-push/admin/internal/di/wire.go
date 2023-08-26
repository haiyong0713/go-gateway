//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	"github.com/google/wire"
	"go-gateway/app/app-svr/archive-push/admin/internal/dao"
	"go-gateway/app/app-svr/archive-push/admin/internal/databus"
	"go-gateway/app/app-svr/archive-push/admin/internal/server/grpc"
	"go-gateway/app/app-svr/archive-push/admin/internal/server/http"
	"go-gateway/app/app-svr/archive-push/admin/internal/service"
)

//go:generate kratos t wire
func InitApp() (*App, func(), error) {
	panic(wire.Build(dao.Provider, service.Provider, databus.New, http.New, grpc.New, NewApp))
}
