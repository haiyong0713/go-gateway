//go:build wireinject
// +build wireinject

package dao

import (
	"github.com/google/wire"
)

//go:generate wire

var daoProvider = wire.NewSet(New, NewDB, NewRedis, NewMC)

func NewDao() (Dao, error) {
	panic(wire.Build(daoProvider))
}
