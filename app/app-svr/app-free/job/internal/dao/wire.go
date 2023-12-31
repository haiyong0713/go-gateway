//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package dao

import (
	"github.com/google/wire"
)

//go:generate kratos tool wire
func newRecordDao() (*dao, func(), error) {
	panic(wire.Build(newDao, NewDB))
}
