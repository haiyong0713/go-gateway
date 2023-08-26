//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	"go-gateway/app/app-svr/app-gw/management-job/internal/dao"
	"go-gateway/app/app-svr/app-gw/management-job/internal/server/grpc"
	"go-gateway/app/app-svr/app-gw/management-job/internal/server/http"
	"go-gateway/app/app-svr/app-gw/management-job/internal/service"

	"github.com/google/wire"
)

//go:generate kratos t wire
func InitApp() (*App, func(), error) {
	panic(wire.Build(dao.Provider, service.Provider, http.New, grpc.New, NewApp))
}
