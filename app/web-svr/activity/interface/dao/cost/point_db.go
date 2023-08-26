package cost

import (
	"context"
	"database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
	cmdl "go-gateway/app/web-svr/activity/interface/model/cost"
	"time"
)

// GetUserAllCost 获取用户全部消耗
func (d *dao) GetUserAllCost(ctx context.Context, activityId string, mid int64, isSplitTable bool) (total int, list []*cmdl.UserCostInfoDB, dbErr error) {
	total = 0
	var lastID int64 = 0
	list = make([]*cmdl.UserCostInfoDB, 0)
	for {
		// 分批1000/每次的取
		tmpList, dbErr := d.fetchCostFromDB(ctx, activityId, mid, lastID)
		if dbErr != nil {
			log.Error("fetchCostFromDB err:,err is (%v)!", dbErr)
			break
		}

		if len(tmpList) > 0 {
			lastID = tmpList[len(tmpList)-1].ID
			list = append(list, tmpList...)
		}
		if len(tmpList) < 1000 {
			break
		}
	}
	// 累加
	if len(list) <= 0 {
		return
	}
	for _, v := range list {
		total += v.CostValue
	}
	return
}

const (
	sql4FetchCostValue = `
SELECT id,mid,award_id,activity_id,order_id,cost_type,cost_value,ctime,mtime
FROM user_cost_info
WHERE mid = ?
    AND activity_id = ?
    AND status = 1
    AND id > ?
order by id ASC
LIMIT 1000
`
)

func (d *dao) fetchCostFromDB(ctx context.Context, activityId string, mid int64, lastID int64) ([]*cmdl.UserCostInfoDB, error) {
	list := make([]*cmdl.UserCostInfoDB, 0)
	rows, err := d.db.Query(ctx, sql4FetchCostValue, mid, activityId, lastID)
	if err != nil {
		return list, err
	}

	defer rows.Close()

	for rows.Next() {
		tmp := new(cmdl.UserCostInfoDB)
		tmpErr := rows.Scan(&tmp.ID, &tmp.Mid, &tmp.AwardId, &tmp.ActivityId,
			&tmp.OrderId, &tmp.CostType, &tmp.CostValue, &tmp.Ctime, &tmp.Mtime)
		if tmpErr != nil {
			return list, tmpErr
		}
		list = append(list, tmp)
	}
	err = rows.Err()
	return list, err
}

const _selUsrCostByOrderId = "select id,mid,order_id,award_id,activity_id,cost_type,cost_value,status,ctime,mtime " +
	"from user_cost_info where order_id = ? and status =1"

func (d *dao) getUserCostByOrderId(ctx context.Context, orderId string) (res *cmdl.UserCostInfoDB, err error) {
	res = new(cmdl.UserCostInfoDB)
	row := d.db.QueryRow(ctx, _selUsrCostByOrderId, orderId)
	err = row.Scan(&res.ID, &res.Mid, &res.OrderId, &res.AwardId, &res.ActivityId,
		&res.CostType, &res.CostValue, &res.Status, &res.Ctime, &res.Mtime)
	if err != nil {
		log.Errorc(ctx, "getUserCostByOrderId:row.Scan err, error(%v)", err)
		if err == sql.ErrNoRows {
			err = nil
			return
		}
	}

	return
}

const _addOneUserCostSQL = "INSERT INTO `user_cost_info` (mid,order_id,award_id,activity_id,cost_type,cost_value,status,ctime,mtime) VALUES(?,?,?,?,?,?,?,?,?)"

// InsertOneUserCost db新增一条消耗记录 ignore幂等
func (d *dao) InsertOneUserCost(ctx context.Context, record *cmdl.UserCostInfoDB, isSplitTable bool) (int64, error) {
	now := time.Now()
	res, err := d.db.Exec(ctx, _addOneUserCostSQL,
		record.Mid, record.OrderId,
		record.AwardId, record.ActivityId,
		record.CostType, record.CostValue,
		record.Status, now, now)

	if err != nil {
		log.Errorc(ctx, "InsertOneUserCost:d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.LastInsertId()

}

const (
	sql4FetchCostValueByDate = `
SELECT id,mid,award_id,activity_id,order_id,cost_type,cost_value,ctime,mtime
FROM user_cost_info
WHERE mid = ?
    AND activity_id = ?
    AND status = 1
    AND cost_type = ?
    AND ctime >= ?
LIMIT 1000
`
)

// GetUserCostListByDate
func (d *dao) GetUserCostListByDate(ctx context.Context, mid int64, activityId string, costTyp int, timeS xtime.Time) (list []*cmdl.UserCostInfoDB, err error) {
	list = make([]*cmdl.UserCostInfoDB, 0)
	rows, err := d.db.Query(ctx, sql4FetchCostValueByDate, mid, activityId, costTyp, timeS)
	if err != nil {
		return list, err
	}
	defer rows.Close()

	for rows.Next() {
		tmp := new(cmdl.UserCostInfoDB)
		if tmpErr := rows.Scan(&tmp.ID, &tmp.Mid, &tmp.AwardId, &tmp.ActivityId,
			&tmp.OrderId, &tmp.CostType, &tmp.CostValue, &tmp.Ctime, &tmp.Mtime); tmpErr == nil {
			list = append(list, tmp)
		}
	}
	err = rows.Err()
	return list, err

}
