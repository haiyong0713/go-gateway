package show

import (
	"context"
	"fmt"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"

	agg "go-gateway/app/app-svr/app-show/interface/model/aggregation"
)

const (
	_aggregationSQL  = "SELECT id,hot_title,title,subtitle,image,state FROM hotword_aggregation WHERE id=? AND state!=4"
	_aggregationsSQL = "SELECT id,hot_title,title,subtitle,image,state FROM hotword_aggregation WHERE id IN (%s) AND state!=4"
)

// RawAggregation .
func (d *Dao) RawAggregation(ctx context.Context, hotID int64) (res *agg.Aggregation, err error) {
	res = &agg.Aggregation{}
	if err = d.db.QueryRow(ctx, _aggregationSQL, hotID).Scan(&res.ID, &res.HotTitle, &res.Title, &res.Subtitle, &res.Image, &res.State); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Error("[RawAggregation] d.db.QueryRow() hotID(%d) error(%v)", hotID, err)
	}
	return
}

// RawAggregations .
func (d *Dao) RawAggregations(ctx context.Context, hotIDs []int64) (res map[int64]*agg.Aggregation, err error) {
	var rows *sql.Rows
	res = make(map[int64]*agg.Aggregation)
	if rows, err = d.db.Query(ctx, fmt.Sprintf(_aggregationsSQL, xstr.JoinInts(hotIDs))); err != nil {
		log.Error("[RawAggregations]  d.db.Query() error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &agg.Aggregation{}
		if err = rows.Scan(&a.ID, &a.HotTitle, &a.Title, &a.Subtitle, &a.Image, &a.State); err != nil {
			log.Error("[RawAggregations] rows.Scan error(%v)", err)
			return
		}
		res[a.ID] = a
	}
	if err = rows.Err(); err != nil {
		log.Error("rows.Err() error(%v)", err)
	}
	return
}
