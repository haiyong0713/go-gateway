package archive

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/xstr"

	achmdl "go-gateway/app/app-svr/archive/service/api"
)

var (
	_internalsSQL = "SELECT `id`,`aid`,`attribute` FROM `archive_internal` WHERE `aid` IN (%s)"
)

func (d *Dao) RawInternals(c context.Context, aids []int64) (map[int64]*achmdl.ArcInternal, error) {
	rows, err := d.resultDB.Query(c, fmt.Sprintf(_internalsSQL, xstr.JoinInts(aids)))
	if err != nil {
		log.Errorc(c, "d.d.resultDB.Query error(%v)", err)
		return nil, err
	}
	defer rows.Close()
	eps := make(map[int64]*achmdl.ArcInternal)
	for rows.Next() {
		ep := &achmdl.ArcInternal{}
		if err = rows.Scan(&ep.ID, &ep.Aid, &ep.Attribute); err != nil {
			log.Error("rows.Scan err (%v)", err)
			continue
		}
		eps[ep.Aid] = ep
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error=%+v", err)
		return nil, err
	}
	return eps, nil
}
