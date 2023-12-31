// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package dao

// Injectors from wire.go:

//go:generate kratos tool wire
func newTestDao() (*dao, func(), error) {
	daoDao, cleanup, err := newDao()
	if err != nil {
		return nil, nil, err
	}
	return daoDao, func() {
		cleanup()
	}, nil
}
