// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"go-gateway/app/app-svr/kvo/job/internal/dao"
	"go-gateway/app/app-svr/kvo/job/internal/server/http"
	"go-gateway/app/app-svr/kvo/job/internal/service"

	"github.com/google/wire"
)

// Injectors from wire.go:

func InitApp() (*App, func(), error) {
	redis, err := dao.NewRedis()
	if err != nil {
		return nil, nil, err
	}
	db, err := dao.NewDB()
	if err != nil {
		return nil, nil, err
	}
	daoDao, err := dao.New(redis, db)
	if err != nil {
		return nil, nil, err
	}
	serviceService, err := service.New(daoDao)
	if err != nil {
		return nil, nil, err
	}
	engine, err := http.New(serviceService)
	if err != nil {
		return nil, nil, err
	}
	app, cleanup, err := NewApp(serviceService, engine)
	if err != nil {
		return nil, nil, err
	}
	return app, func() {
		cleanup()
	}, nil
}

// wire.go:

type defaultHttp interface {
}

var daoProvider = wire.NewSet(dao.New, dao.NewDB, dao.NewRedis)

var serviceProvider = wire.NewSet(service.New, wire.Bind(new(defaultHttp), new(*service.Service)))
