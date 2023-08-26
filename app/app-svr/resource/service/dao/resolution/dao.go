package resolution

import (
	"context"
	"time"

	"go-gateway/app/app-svr/resource/service/conf"
	rm "go-gateway/app/app-svr/resource/service/model"

	xsql "go-common/library/database/sql"
)

const (
	_limitFree = "select aid, limit_free, subtitle from resolution_limit_free where is_deleted=0 and state=1 and stime<? and etime>?"
)

type Dao struct {
	gwdb *xsql.DB
	get  *xsql.Stmt
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		gwdb: xsql.NewMySQL(c.DB.GWDB),
	}
	d.get = d.gwdb.Prepared(_limitFree)
	return
}

func (d *Dao) FetchAllLimitFreeOnline(ctx context.Context) ([]*rm.LimitFreeInfo, error) {
	now := time.Now()
	rows, err := d.get.Query(ctx, now.Unix(), now.Unix())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reply []*rm.LimitFreeInfo
	for rows.Next() {
		limitFreeInfo := &rm.LimitFreeInfo{}
		if err := rows.Scan(&limitFreeInfo.Aid, &limitFreeInfo.LimitFree, &limitFreeInfo.Subtitle); err != nil {
			return nil, err
		}
		reply = append(reply, limitFreeInfo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *Dao) Close() {
	if d.gwdb != nil {
		_ = d.gwdb.Close()
	}
}
