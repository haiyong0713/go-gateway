package dao

import (
	"context"

	"go-common/library/conf/paladin.v2"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) error
}

// dao dao.
type dao struct{}

// New new a dao and return.
func New() (d Dao, cf func(), err error) {
	return newDao()
}

func newDao() (d *dao, cf func(), err error) {
	var cfg struct {
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
