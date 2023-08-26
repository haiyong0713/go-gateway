// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//go:build !wireinject
// +build !wireinject

package dao

// Injectors from wire.go:

func newRecordDao() (*dao, func(), error) {
	db, cleanup, err := NewDB()
	if err != nil {
		return nil, nil, err
	}
	daoDao, cleanup2, err := newDao(db)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	return daoDao, func() {
		cleanup2()
		cleanup()
	}, nil
}
