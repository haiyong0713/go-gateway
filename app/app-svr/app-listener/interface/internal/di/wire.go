//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	dao2 "go-gateway/app/app-svr/app-listener/interface/internal/dao"
	grpc2 "go-gateway/app/app-svr/app-listener/interface/internal/server/grpc"
	http2 "go-gateway/app/app-svr/app-listener/interface/internal/server/http"
	service2 "go-gateway/app/app-svr/app-listener/interface/internal/service"

	"github.com/google/wire"
)

//go:generate kratos t wire
func InitApp() (*App, func(), error) {
	panic(wire.Build(dao2.Provider, service2.Provider, http2.New, grpc2.New, NewApp))
}
