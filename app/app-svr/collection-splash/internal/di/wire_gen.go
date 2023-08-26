// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"go-gateway/app/app-svr/collection-splash/internal/dao"
	"go-gateway/app/app-svr/collection-splash/internal/server/grpc"
	"go-gateway/app/app-svr/collection-splash/internal/server/http"
	"go-gateway/app/app-svr/collection-splash/internal/service"
)

// Injectors from wire.go:

//go:generate kratos t wire
func InitApp() (*App, func(), error) {
	db, cleanup, err := dao.NewDB()
	if err != nil {
		return nil, nil, err
	}
	redis, cleanup2, err := dao.NewRedis()
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	daoDao, cleanup3, err := dao.New(db, redis)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	serviceService, cleanup4, err := service.New(daoDao)
	if err != nil {
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	engine, err := http.New(serviceService)
	if err != nil {
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	server, err := grpc.New(serviceService)
	if err != nil {
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	app, cleanup5, err := NewApp(serviceService, engine, server)
	if err != nil {
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	return app, func() {
		cleanup5()
		cleanup4()
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}
