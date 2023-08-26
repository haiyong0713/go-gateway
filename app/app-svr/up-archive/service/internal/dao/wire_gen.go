// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//go:build !wireinject
// +build !wireinject

package dao

// Injectors from wire.go:

func newTestDao() (*dao, func(), error) {
	redis, cleanup, err := NewRedis()
	if err != nil {
		return nil, nil, err
	}
	elastic, cleanup2, err := NewElastic()
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	daoDao, cleanup3, err := newDao(redis, elastic)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	return daoDao, func() {
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}