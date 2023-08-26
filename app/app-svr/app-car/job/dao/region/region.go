package region

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-car/job/conf"
	"go-gateway/app/app-svr/app-car/job/model/region"
)

const (
	_regionAndroid = `SELECT r.rid,r.reid,r.name FROM region_copy AS r, language AS l WHERE r.plat=0 AND r.state=1 AND l.id=r.lang_id AND l.name="hans" ORDER BY r.rank DESC`
)

type Dao struct {
	db *sql.DB
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		db: sql.NewMySQL(c.MySQL.Show),
	}
	return d
}

func (d *Dao) AndroidAll(ctx context.Context) (map[int32]*region.Region, error) {
	rows, err := d.db.Query(ctx, _regionAndroid)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	rs := map[int32]*region.Region{}
	for rows.Next() {
		r := &region.Region{}
		if err := rows.Scan(&r.Rid, &r.Reid, &r.Name); err != nil {
			log.Error("row.Scan error(%v)", err)
			return nil, err
		}
		rs[r.Rid] = r
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return rs, nil
}
