//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package di

import (
	packSvc "go-gateway/app/app-svr/fawkes/job/internal/service/pack"

	"go-gateway/app/app-svr/fawkes/job/internal/dao"
	"go-gateway/app/app-svr/fawkes/job/internal/server/http"
	modSvc "go-gateway/app/app-svr/fawkes/job/internal/service/mod"

	"github.com/google/wire"
)

//go:generate kratos t wire
func InitApp() (*App, func(), error) {
	panic(wire.Build(dao.Provider, modSvc.Provider, packSvc.Provider, http.Provider, NewApp))
}
