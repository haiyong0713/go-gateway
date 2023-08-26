//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	"go-gateway/app/app-svr/distribution/distribution/internal/dao"
	"go-gateway/app/app-svr/distribution/distribution/internal/server/grpc"
	"go-gateway/app/app-svr/distribution/distribution/internal/server/http"
	"go-gateway/app/app-svr/distribution/distribution/internal/service"

	"github.com/google/wire"
)

//go:generate kratos t wire
func InitApp() (*App, func(), error) {
	panic(wire.Build(dao.Provider, service.Provider, http.New, grpc.New, NewApp))
}
