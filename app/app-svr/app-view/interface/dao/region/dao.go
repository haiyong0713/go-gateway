package region

import (
	"context"

	xsql "go-common/library/database/sql"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/app-view/interface/model/region"
)

const (
	// region
	_secSQL = "SELECT rid,name,logo,rank,goto,param,plat,area FROM region_copy WHERE state=1 AND reid!=0"
)

type Dao struct {
	db  *xsql.DB
	get *xsql.Stmt
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db: xsql.NewMySQL(c.MySQL.Show),
	}
	// prepare
	d.get = d.db.Prepared(_secSQL)
	return
}

// Seconds get all second region.
func (d *Dao) Seconds(ctx context.Context) (rm map[int8]map[int]*region.Region, err error) {
	rows, err := d.get.Query(ctx)
	if err != nil {
		return
	}
	defer rows.Close()
	rm = map[int8]map[int]*region.Region{}
	for rows.Next() {
		a := &region.Region{}
		if err = rows.Scan(&a.Rid, &a.Name, &a.Logo, &a.Rank, &a.Goto, &a.Param, &a.Plat, &a.Area); err != nil {
			return
		}
		if rs, ok := rm[a.Plat]; ok {
			rs[a.Rid] = a
		} else {
			rm[a.Plat] = map[int]*region.Region{a.Rid: a}
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return
}
