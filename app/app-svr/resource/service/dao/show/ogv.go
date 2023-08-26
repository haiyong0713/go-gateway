package show

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/resource/service/model"
)

const (
	_searchOgv = "SELECT id,pgc_ids FROM search_ogv WHERE `check` = 2 AND stime <= ?"
)

// SearchOgv .
func (d *Dao) SearchOgv(c context.Context) (res map[int64][]int64, err error) {
	rows, err := d.db.Query(c, _searchOgv, time.Now())
	if err != nil {
		log.Error("SearchOgv Query error (%v)", err)
		return
	}
	defer rows.Close()
	res = make(map[int64][]int64)
	for rows.Next() {
		tmp := &model.SearchOgv{}
		if err = rows.Scan(&tmp.ID, &tmp.PgcIDs); err != nil {
			log.Error("SearchOgv rows.Scan err (%v)", err)
			return
		}
		if tmp.PgcIDs != "" {
			tmpPgcIDs, e := xstr.SplitInts(tmp.PgcIDs)
			if e != nil {
				log.Error("SearchOgv xstr.SplitInts ID(%d) pgcIDs(%v) err (%v)", tmp.ID, tmp.PgcIDs, e)
				continue
			}
			if len(tmpPgcIDs) == 0 {
				continue
			}
			res[tmp.ID] = tmpPgcIDs
		}
	}
	err = rows.Err()
	return
}
