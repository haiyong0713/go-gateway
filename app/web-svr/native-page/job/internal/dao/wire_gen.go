// Code generated by Wire. DO NOT EDIT.

//go:build !wireinject
// +build !wireinject

package dao

// Injectors from wire.go:

//go:generate kratos tool wire
func newTestDao() (*dao, func(), error) {
	redis, cleanup, err := NewRedis()
	if err != nil {
		return nil, nil, err
	}
	db, cleanup2, err := NewDB()
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	daoDao, cleanup3, err := newDao(redis, db)
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
