package dao

import (
	"context"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/dynamic/service/model"
)

const (
	_rgnArcsSQL = "SELECT aid,attribute,copyright,pubtime,state,typeid FROM archive WHERE id>=? AND id<?"
)

// ArchiveAll get archive info from db .
func (d *Dao) ArchiveAll(ctx context.Context, rid int32, start, end int) (res map[int32][]*api.RegionArc, err error) {
	var (
		rows *sql.Rows
	)
	res = make(map[int32][]*api.RegionArc)
	if rows, err = d.dbArc.Query(ctx, _rgnArcsSQL, start, end); err != nil {
		log.Error("[ArchiveAll] d.db.Query() sql(%s) error(%v)", _rgnArcsSQL, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := &model.RegionArc{}
		if err = rows.Scan(&r.Aid, &r.Attribute, &r.Copyright, &r.PubDate, &r.State, &r.TypeID); err != nil {
			log.Error("[ArchiveAll] rows.Scan error(%v)", err)
			return
		}
		if r.State == -6 || r.State >= 0 {
			tmp := &api.RegionArc{
				Aid:       r.Aid,
				Attribute: r.Attribute,
				Copyright: r.Copyright,
				PubDate:   r.PubDate,
			}
			res[r.TypeID] = append(res[r.TypeID], tmp)
		}
	}
	if err = rows.Err(); err != nil {
		log.Error("[ArchiveAll] rows.Err error(%v)", err)
	}
	return
}
