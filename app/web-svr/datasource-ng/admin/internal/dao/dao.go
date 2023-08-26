package dao

import (
	"context"

	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/sync/pipeline/fanout"
)

//go:generate kratos tool btsgen
// Dao dao interface
//type Dao interface {
//	Close()
//	Ping(ctx context.Context) (err error)
//}

// dao dao.
type Dao struct {
	db    *sql.DB
	cache *fanout.Fanout
}

// New new a dao and return.
func New(db *sql.DB) (*Dao, error) {
	var cfg struct {
	}
	if err := paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	d := &Dao{
		db:    db,
		cache: fanout.New("cache"),
	}
	return d, nil
}

// Close close the resource.
func (d *Dao) Close() {
	d.db.Close()
	d.cache.Close()
}

// Ping ping the resource.
func (d *Dao) Ping(ctx context.Context) error {
	return nil
}

func (d *Dao) TxBegin(ctx context.Context) (*sql.Tx, error) {
	return d.db.Begin(ctx)
}
