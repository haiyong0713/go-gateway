//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	v12 "go-gateway/app/app-svr/app-search/internal/dao/v1"
	"go-gateway/app/app-svr/app-search/internal/server/grpc"
	"go-gateway/app/app-svr/app-search/internal/server/http"
	"go-gateway/app/app-svr/app-search/internal/service/v1"

	"github.com/google/wire"
)

//go:generate kratos t wire
func InitApp() (*App, func(), error) {
	panic(wire.Build(v12.Provider, v1.Provider, http.New, grpc.New, NewApp))
}
