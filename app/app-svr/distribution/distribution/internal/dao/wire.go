//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package dao

import (
	"go-gateway/app/app-svr/distribution/distribution/internal/dao/kv"

	"github.com/google/wire"
)

//go:generate kratos tool wire
func newTestDao() (*dao, func(), error) {
	panic(wire.Build(newDao, NewRedis, kv.NewKV))
}
