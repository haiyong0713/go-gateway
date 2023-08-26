package dao

import (
	"context"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/job/model"
)

const (
	sql4MaxContestID                    = `select max(id) from es_contests`
	sql4GetAllContestSeriesInfo4Refresh = `
select id,type from contest_series where is_deleted=0
`
)

func (d *Dao) FetchMaxContestID(ctx context.Context) (maxID int64, err error) {
	err = d.db.QueryRow(ctx, sql4MaxContestID).Scan(&maxID)

	return
}

func (d *Dao) GetAllSeriesInfo4Refresh(ctx context.Context) (series []*model.ContestSeries4InfoRefresh, err error) {
	series = make([]*model.ContestSeries4InfoRefresh, 0)
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, sql4GetAllContestSeriesInfo4Refresh)
	if err != nil {
		return
	}
	defer func() {
		_ = rows.Close()
		if err = rows.Err(); err != nil {
			log.Errorc(ctx, "GetAllSeriesInfo4Refresh rows.Err() error(%v)", err)
		}
	}()
	for rows.Next() {
		t := &model.ContestSeries4InfoRefresh{}
		err = rows.Scan(&t.ID, &t.Type)
		if err != nil {
			return
		}
		series = append(series, t)
	}
	return

}
