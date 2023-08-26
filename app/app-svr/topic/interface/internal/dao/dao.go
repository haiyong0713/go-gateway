package dao

import (
	"context"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
}

type dao struct {
}

// New new a dao and return.
func New() (d *dao, cf func(), err error) {
	return newDao()
}

func newDao() (d *dao, cf func(), err error) {
	d = &dao{}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
	// Do nothing because of no content in dao struct.
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}
