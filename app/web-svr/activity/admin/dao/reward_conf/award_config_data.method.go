// Code generated by bcurd. DO NOT EDIT.
package reward_conf

import (
	"context"
	xsql "database/sql"
	"strings"

	"github.com/pkg/errors"

	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/time"
)

const (
	awardConfigDataID         = "id"
	awardConfigDataAwardID    = "award_id"
	awardConfigDataStockID    = "stock_id"
	awardConfigDataCostType   = "cost_type"
	awardConfigDataCostValue  = "cost_value"
	awardConfigDataShowTime   = "show_time"
	awardConfigDataOrder      = "order"
	awardConfigDataCreator    = "creator"
	awardConfigDataStatus     = "status"
	awardConfigDataCtime      = "ctime"
	awardConfigDataMtime      = "mtime"
	awardConfigDataActivityID = "activity_id"
	awardConfigDataEndTime    = "end_time"
	sqlFullFields             = "`id`,`award_id`,`stock_id`,`cost_type`,`cost_value`,`show_time`,`order`,`creator`,`status`,`ctime`,`mtime`,`activity_id`,`end_time`"
)

// queryAwardConfigDataRow   Select a AwardConfigData record
func (dao *Dao) queryAwardConfigDataRow(ctx context.Context, sqlStr string, args ...interface{}) (*AwardConfigData, error) {
	var err error
	awardConfigData := AwardConfigData{}
	err = dao.db.QueryRow(ctx, sqlStr, args...).Scan(&awardConfigData.ID, &awardConfigData.AwardID, &awardConfigData.StockID, &awardConfigData.CostType, &awardConfigData.CostValue, &awardConfigData.ShowTime, &awardConfigData.Order, &awardConfigData.Creator, &awardConfigData.Status, &awardConfigData.Ctime, &awardConfigData.Mtime, &awardConfigData.ActivityID, &awardConfigData.EndTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorc(ctx, "queryAwardConfigDataRow sql:%v args:%v err:%v", sqlStr, args, err)
		return nil, errors.Wrap(err, "queryAwardConfigDataRow err")
	}
	return &awardConfigData, nil
}

// queryAwardConfigDataCount Select count AwardConfigData record
func (dao *Dao) queryAwardConfigDataCount(ctx context.Context, sqlStr string, args ...interface{}) (int64, error) {
	var count int64
	err := dao.db.QueryRow(ctx, sqlStr, args...).Scan(&count)
	if err != nil {
		log.Errorc(ctx, "queryAwardConfigDataCount sql:%v args:%v err:%v", sqlStr, args, err)
		return 0, errors.Wrap(err, "queryAwardConfigDataCount err")
	}
	return count, nil
}

// queryAwardConfigDataRows   Select AwardConfigData records
func (dao *Dao) queryAwardConfigDataRows(ctx context.Context, sqlStr string, args ...interface{}) ([]*AwardConfigData, error) {
	q, err := dao.db.Query(ctx, sqlStr, args...)
	if err != nil {
		log.Errorc(ctx, "queryAwardConfigDataRows Query sql:%v args:%v err:%v", sqlStr, args, err)
		return nil, errors.Wrap(err, "queryAwardConfigDataRows err")
	}
	defer q.Close()
	res := make([]*AwardConfigData, 0)
	for q.Next() {
		awardConfigData := AwardConfigData{}
		err = q.Scan(&awardConfigData.ID, &awardConfigData.AwardID, &awardConfigData.StockID, &awardConfigData.CostType, &awardConfigData.CostValue, &awardConfigData.ShowTime, &awardConfigData.Order, &awardConfigData.Creator, &awardConfigData.Status, &awardConfigData.Ctime, &awardConfigData.Mtime, &awardConfigData.ActivityID, &awardConfigData.EndTime)
		if err != nil {
			log.Errorc(ctx, "queryAwardConfigDataRows Scan sql:%v args:%v err:%v", sqlStr, args, err)
			return nil, errors.Wrap(err, "queryAwardConfigDataRows err")
		}
		res = append(res, &awardConfigData)
	}
	if q.Err() != nil {
		log.Errorc(ctx, "queryAwardConfigDataRows Err() sql:%v args:%v err:%v", sqlStr, args, err)
		return nil, errors.Wrap(err, "queryAwardConfigDataRows err")
	}
	return res, nil
}

// execAwardConfigDataQuery   exec AwardConfigData query
func (dao *Dao) execAwardConfigDataQuery(ctx context.Context, sqlStr string, args ...interface{}) (int64, error) {
	result, err := dao.db.Exec(ctx, sqlStr, args...)
	if err != nil {
		log.Errorc(ctx, "execAwardConfigDataQuery sql:%v args:%v err:%v", sqlStr, args, err)
		return 0, errors.Wrap(err, "execAwardConfigDataQuery err")
	}
	return result.RowsAffected()
}

// InsertAwardConfigData  Insert a record
func (dao *Dao) InsertAwardConfigData(ctx context.Context, awardConfigData *AwardConfigData) error {
	var err error
	const sqlStr = "INSERT INTO  `award_config_data` (" +
		"`award_id`,`stock_id`,`cost_type`,`cost_value`,`show_time`,`order`,`creator`,`status`,`activity_id`,`end_time`" +
		`) VALUES (` +
		` ?,?,?,?,?,?,?,?,?,?` +
		`)`

	var result xsql.Result
	result, err = dao.db.Exec(ctx, sqlStr, &awardConfigData.AwardID, &awardConfigData.StockID, &awardConfigData.CostType, &awardConfigData.CostValue, &awardConfigData.ShowTime, &awardConfigData.Order, &awardConfigData.Creator, &awardConfigData.Status, &awardConfigData.ActivityID, &awardConfigData.EndTime)

	if err != nil {
		log.Errorc(ctx, "AwardConfigData Insert (%v) Exec err: %v", *awardConfigData, err)
		return errors.Wrap(err, "InsertAwardConfigData err")
	}

	var id int64
	id, err = result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "InsertAwardConfigData err")
	}
	awardConfigData.ID = uint64(id)

	return nil
}

// MultiInsertAwardConfigData  Insert multi record
func (dao *Dao) MultiInsertAwardConfigData(ctx context.Context, awardConfigDatas []*AwardConfigData) (int64, error) {
	var err error
	var sqlStr = "INSERT INTO  `award_config_data` (" +
		"`award_id`,`stock_id`,`cost_type`,`cost_value`,`show_time`,`order`,`creator`,`status`,`activity_id`,`end_time`" +
		`) VALUES ` +
		strings.TrimRight(strings.Repeat(`(?,?,?,?,?,?,?,?,?,?),`, len(awardConfigDatas)), ",")
	var params []interface{}
	for _, awardConfigData := range awardConfigDatas {
		params = append(params, awardConfigData.AwardID, awardConfigData.StockID, awardConfigData.CostType, awardConfigData.CostValue, awardConfigData.ShowTime, awardConfigData.Order, awardConfigData.Creator, awardConfigData.Status, awardConfigData.ActivityID, awardConfigData.EndTime)
	}
	var result xsql.Result
	result, err = dao.db.Exec(ctx, sqlStr, params...)
	if err != nil {
		log.Errorc(ctx, "AwardConfigData Insert Exec err: %v", err)
		return 0, errors.Wrap(err, "MultiInsertAwardConfigData err")
	}
	return result.RowsAffected()
}

// DeleteAwardConfigDataByID Delete by primary key:`id`
func (dao *Dao) DeleteAwardConfigDataByID(ctx context.Context, id uint64) (int64, error) {
	const sqlStr = "DELETE FROM `award_config_data` WHERE `id` = ?"
	return dao.execAwardConfigDataQuery(ctx, sqlStr, id)
}

// UpdateAwardConfigDataByID Update a record
func (dao *Dao) UpdateAwardConfigDataByID(ctx context.Context, awardConfigData *AwardConfigData) (int64, error) {
	const sqlStr = "UPDATE `award_config_data` SET " +
		"`award_id` = ? ,`stock_id` = ? ,`cost_type` = ? ,`cost_value` = ? ,`show_time` = ? ,`order` = ? ,`creator` = ? ,`status` = ? ,`activity_id` = ? ,`end_time` = ? " +
		" WHERE `id` = ?"
	return dao.execAwardConfigDataQuery(ctx, sqlStr, awardConfigData.AwardID, awardConfigData.StockID, awardConfigData.CostType, awardConfigData.CostValue, awardConfigData.ShowTime, awardConfigData.Order, awardConfigData.Creator, awardConfigData.Status, awardConfigData.ActivityID, awardConfigData.EndTime, awardConfigData.ID)
}

// mapUpdateAwardConfigDataByID Update a record
func (dao *Dao) mapUpdateAwardConfigDataByID(ctx context.Context, id uint64, awardConfigData map[string]interface{}) (int64, error) {
	cols := make([]string, 0, len(awardConfigData))
	vars := make([]interface{}, 0, len(awardConfigData))
	for k, v := range awardConfigData {
		cols = append(cols, "`"+k+"` = ?")
		vars = append(vars, v)
	}
	var sqlStr = "UPDATE `award_config_data` SET " +
		strings.Join(cols, ",") +
		" WHERE `id` = ?"
	vars = append(vars, id)
	return dao.execAwardConfigDataQuery(ctx, sqlStr, vars...)
}

// RawAwardConfigDataByID   Select a record by primary key:`id`
func (dao *Dao) RawAwardConfigDataByID(ctx context.Context, id uint64) (*AwardConfigData, error) {
	const sqlStr = `SELECT ` +
		sqlFullFields +
		" FROM  `award_config_data` " +
		"WHERE `id` = ?"
	return dao.queryAwardConfigDataRow(ctx, sqlStr, id)
}

// RawdaoByIDList   Select a record by primary key:`id`
func (dao *Dao) RawAwardConfigDataByIDList(ctx context.Context, id []uint64) ([]*AwardConfigData, error) {
	var sqlStr = `SELECT ` +
		sqlFullFields +
		" FROM  `award_config_data` " +
		"WHERE `id` IN (" + strings.TrimRight(strings.Repeat("?,", len(id)), ",") + `)`
	args := make([]interface{}, 0, len(id))
	for _, v := range id {
		args = append(args, v)
	}
	return dao.queryAwardConfigDataRows(ctx, sqlStr, args...)
}

// RawAwardConfigDataByAwardIDShowTime Select by index name:`k_awardid_showtime`
func (dao *Dao) RawAwardConfigDataByAwardIDShowTime(ctx context.Context, awardID string, showTime time.Time, offset, limit int) ([]*AwardConfigData, error) {
	const sqlStr = `SELECT ` +
		sqlFullFields +
		" FROM  `award_config_data` " +
		"WHERE `award_id` = ? AND `show_time` = ?  LIMIT ?,?"
	return dao.queryAwardConfigDataRows(ctx, sqlStr, awardID, showTime, offset, limit)
}

func (dao *Dao) RawCountAwardConfigDataByAwardIDShowTime(ctx context.Context, awardID string, showTime time.Time) (int64, error) {
	const sqlStr = "SELECT count(*) FROM  `award_config_data` " +
		"WHERE `award_id` = ? AND `show_time` = ? "
	return dao.queryAwardConfigDataCount(ctx, sqlStr, awardID, showTime)
}

// RawAwardConfigDataByMtime Select by index name:`ix_mtime`
func (dao *Dao) RawAwardConfigDataByMtime(ctx context.Context, mtime time.Time, offset, limit int) ([]*AwardConfigData, error) {
	const sqlStr = `SELECT ` +
		sqlFullFields +
		" FROM  `award_config_data` " +
		"WHERE `mtime` = ?  LIMIT ?,?"
	return dao.queryAwardConfigDataRows(ctx, sqlStr, mtime, offset, limit)
}

func (dao *Dao) RawCountAwardConfigDataByMtime(ctx context.Context, mtime time.Time) (int64, error) {
	const sqlStr = "SELECT count(*) FROM  `award_config_data` " +
		"WHERE `mtime` = ? "
	return dao.queryAwardConfigDataCount(ctx, sqlStr, mtime)
}

// RawAwardConfigDataByActivityIDAwardIDCostTypeShowTimeStatus Select by index name:`k_sid_awardid_showtime`
func (dao *Dao) RawAwardConfigDataByActivityIDAwardIDCostTypeShowTimeStatus(ctx context.Context, activityID string, awardID string, costType int8, showTime time.Time, status int8, offset, limit int) ([]*AwardConfigData, error) {
	const sqlStr = `SELECT ` +
		sqlFullFields +
		" FROM  `award_config_data` " +
		"WHERE `activity_id` = ? AND `award_id` = ? AND `cost_type` = ? AND `show_time` = ? AND `status` = ?  LIMIT ?,?"
	return dao.queryAwardConfigDataRows(ctx, sqlStr, activityID, awardID, costType, showTime, status, offset, limit)
}

func (dao *Dao) RawCountAwardConfigDataByActivityIDAwardIDCostTypeShowTimeStatus(ctx context.Context, activityID string, awardID string, costType int8, showTime time.Time, status int8) (int64, error) {
	const sqlStr = "SELECT count(*) FROM  `award_config_data` " +
		"WHERE `activity_id` = ? AND `award_id` = ? AND `cost_type` = ? AND `show_time` = ? AND `status` = ? "
	return dao.queryAwardConfigDataCount(ctx, sqlStr, activityID, awardID, costType, showTime, status)
}