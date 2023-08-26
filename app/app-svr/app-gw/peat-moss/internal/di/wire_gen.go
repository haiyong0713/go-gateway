// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"go-gateway/app/app-svr/app-gw/peat-moss/internal/dao"
	"go-gateway/app/app-svr/app-gw/peat-moss/internal/server/http"
	"go-gateway/app/app-svr/app-gw/peat-moss/internal/service"
)

// Injectors from wire.go:

func InitApp() (*App, func(), error) {
	daoDao, cleanup, err := dao.New()
	if err != nil {
		return nil, nil, err
	}
	serviceService, cleanup2, err := service.New(daoDao)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	engine, cleanup3, err := http.New(serviceService)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	app, cleanup4, err := NewApp(serviceService, engine)
	if err != nil {
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	return app, func() {
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}