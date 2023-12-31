// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//go:build !wireinject
// +build !wireinject

package di

import (
	"go-gateway/app/web-svr/datasource-ng/service/api"
	"go-gateway/app/web-svr/datasource-ng/service/internal/dao"
	"go-gateway/app/web-svr/datasource-ng/service/internal/server/grpc"
	"go-gateway/app/web-svr/datasource-ng/service/internal/server/http"
	"go-gateway/app/web-svr/datasource-ng/service/internal/service"

	"github.com/google/wire"
)

// Injectors from wire.go:

func InitApp() (*App, func(), error) {
	db, err := dao.NewDB()
	if err != nil {
		return nil, nil, err
	}
	daoDao, err := dao.New(db)
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
	server, err := grpc.New(serviceService)
	if err != nil {
		return nil, nil, err
	}
	app, cleanup, err := NewApp(serviceService, engine, server)
	if err != nil {
		return nil, nil, err
	}
	return app, func() {
		cleanup()
	}, nil
}

// wire.go:

var daoProvider = wire.NewSet(dao.New, dao.NewDB)

var serviceProvider = wire.NewSet(service.New, wire.Bind(new(api.DataSourceNGServer), new(*service.Service)))
