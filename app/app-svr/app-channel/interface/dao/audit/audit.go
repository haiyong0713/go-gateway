package audit

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-channel/interface/conf"
)

const (
	_getSQL = "SELECT mobi_app,build FROM audit"
)

// Dao is audit dao.
type Dao struct {
	db *sql.DB
}

// New new a audit dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db: sql.NewMySQL(c.MySQL.Show),
	}
	return
}

// Audits get all audit build.
func (d *Dao) Audits(ctx context.Context) (res map[string]map[int]struct{}, err error) {
	rows, err := d.db.Query(ctx, _getSQL)
	if err != nil {
		log.Error("query error(%v)", err)
		return
	}
	defer rows.Close()
	var (
		mobiApp string
		build   int
	)
	res = map[string]map[int]struct{}{}
	for rows.Next() {
		if err = rows.Scan(&mobiApp, &build); err != nil {
			log.Error("rows.Scan error(%v)", err)
			res = nil
			return
		}
		if plat, ok := res[mobiApp]; ok {
			plat[build] = struct{}{}
		} else {
			res[mobiApp] = map[int]struct{}{
				build: {},
			}
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}

// Close close memcache resource.
func (dao *Dao) Close() {
	if dao.db != nil {
		dao.db.Close()
	}
}
