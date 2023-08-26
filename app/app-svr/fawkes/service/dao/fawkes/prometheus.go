package fawkes

import (
	"context"

	"go-common/library/database/sql"

	prometheusmdl "go-gateway/app/app-svr/fawkes/service/model/prometheus"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

const (
	// _ciInWaiting 目前由于企业包pkg_type=3 和 企业包debug pkg_type=6都是重签名得来，不占用打包机资源， 所以排除掉
	_ciInWaiting = `SELECT p.app_key,count(CASE WHEN p.status=1 THEN 1 ELSE NULL END) FROM (SELECT app_key,status,state,pkg_type,ctime FROM build_pack ORDER BY id desc LIMIT 10000) AS p WHERE p.state=0 AND p.pkg_type NOT IN (3,6) AND p.ctime>DATE_SUB(NOW(),INTERVAL 2 HOUR) GROUP BY p.app_key`
)

// CIInWaiting monitor android ci in waiting status number
func (d *Dao) CIInWaiting(c context.Context) (res *prometheusmdl.CIInWaiting, err error) {
	rows, err := d.db.Query(c, _ciInWaiting)
	if err != nil {
		log.Error("%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		res = &prometheusmdl.CIInWaiting{}
		if err = rows.Scan(&res.AppKey, &res.Count); err != nil {
			if err == sql.ErrNoRows {
				err = nil
			} else {
				log.Error("%v", err)
			}
			return
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return
}
