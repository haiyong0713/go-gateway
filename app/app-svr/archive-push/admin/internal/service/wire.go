//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package service

import (
	"github.com/google/wire"

	"go-gateway/app/app-svr/archive-push/admin/internal/dao"
	blizzardDao "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/dao"
	qqDao "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/dao"
)

//go:generate kratos t wire
func InitService() (*Service, func(), error) {
	panic(wire.Build(New, dao.Provider, qqDao.Init, blizzardDao.Init))
}
