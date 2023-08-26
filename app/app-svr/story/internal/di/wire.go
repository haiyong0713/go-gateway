//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	infocv2 "go-common/library/log/infoc.v2"
	"go-gateway/app/app-svr/story/internal/dao"
	"go-gateway/app/app-svr/story/internal/server/grpc"
	"go-gateway/app/app-svr/story/internal/server/http"
	"go-gateway/app/app-svr/story/internal/service"

	"github.com/google/wire"
)

//go:generate kratos t wire
func InitApp(ic infocv2.Infoc) (*App, func(), error) {
	panic(wire.Build(dao.Provider, service.Provider, http.New, grpc.New, NewApp))
}
