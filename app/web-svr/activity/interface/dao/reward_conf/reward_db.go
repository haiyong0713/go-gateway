package reward_conf

import (
	"context"
	"database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
	cmdl "go-gateway/app/web-svr/activity/interface/model/cost"
)

const _selAwardConfByIdAndDate = "select id,activity_id,award_id,stock_id,cost_type,cost_value,show_time,end_time,`order`,creator,`status`,ctime,mtime " +
	"from award_config_data where activity_id = ? and award_id = ? and cost_type = ? and show_time <= ? and end_time >= ? and status =1"

func (d *dao) GetAwardConfByIdAndDate(ctx context.Context, sid string, awardId string, costType int, timeS xtime.Time) (res *cmdl.AwardConfigDataDB, err error) {
	res = new(cmdl.AwardConfigDataDB)
	row := d.db.QueryRow(ctx, _selAwardConfByIdAndDate, sid, awardId, costType, timeS, timeS)
	err = row.Scan(&res.ID, &res.ActivityId, &res.AwardId, &res.StockId, &res.CostType, &res.CostValue,
		&res.ShowTime, &res.EndTime, &res.Order, &res.Creator, &res.Status, &res.Ctime, &res.Mtime)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "getAwardConfByIdAndDate:row.Scan err, error(%v)", err)
	}

	return
}

const (
	sql4FetchAward = "SELECT id,activity_id,award_id,stock_id,cost_type,cost_value,show_time,end_time,`order`,creator,`status`,ctime,mtime " +
		"FROM award_config_data " +
		"WHERE activity_id = ? and cost_type = ? " +
		"and show_time <= ? and end_time >= ? and status = 1 " +
		"order by `order` ASC " +
		"LIMIT 1000"
)

// FetchAwardFromDB 获取奖品列表
func (d *dao) FetchAwardFromDB(ctx context.Context, activityId string, costType int, timeS xtime.Time) ([]*cmdl.AwardConfigDataDB, error) {
	list := make([]*cmdl.AwardConfigDataDB, 0)
	rows, err := d.db.Query(ctx, sql4FetchAward, activityId, costType, timeS, timeS)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		tmp := new(cmdl.AwardConfigDataDB)
		if tmpErr := rows.Scan(&tmp.ID, &tmp.ActivityId, &tmp.AwardId,
			&tmp.StockId, &tmp.CostType, &tmp.CostValue, &tmp.ShowTime, &tmp.EndTime, &tmp.Order,
			&tmp.Creator, &tmp.Status, &tmp.Ctime, &tmp.Mtime); tmpErr == nil {
			list = append(list, tmp)
		}
	}
	err = rows.Err()
	return list, err
}
