// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"go-gateway/app/api-gateway/api-manager/internal/dao"
	"go-gateway/app/api-gateway/api-manager/internal/server/grpc"
	"go-gateway/app/api-gateway/api-manager/internal/server/http"
	"go-gateway/app/api-gateway/api-manager/internal/service"
)

// Injectors from wire.go:

func InitApp() (*App, func(), error) {
	db, cleanup, err := dao.NewDB()
	if err != nil {
		return nil, nil, err
	}
	daoDao, cleanup2, err := dao.New(db)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	serviceService, cleanup3, err := service.New(daoDao)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	engine, err := http.New(serviceService)
	if err != nil {
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	server, err := grpc.New(serviceService)
	if err != nil {
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	app, cleanup4, err := NewApp(serviceService, engine, server)
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
