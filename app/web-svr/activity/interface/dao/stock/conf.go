package stock

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/stock"
	"time"
)

const (
	_addStockServerConf        = "INSERT INTO act_stock_server_conf (resource_id,resource_ver,foreign_act_id,describe_info,rules_info,stock_start_time , stock_end_time) VALUES (?,?,?,?,?,?,?)"
	_updateStockServerConf     = "UPDATE act_stock_server_conf SET resource_id = ? , resource_ver = ? ,foreign_act_id = ? , describe_info = ? , rules_info = ? , stock_start_time= ? , stock_end_time = ?  WHERE id = ?"
	_stockServerConfListByRid  = "SELECT id , resource_id,resource_ver,foreign_act_id,describe_info  ,ctime ,mtime , rules_info , stock_start_time , stock_end_time FROM act_stock_server_conf WHERE resource_id=? and foreign_act_id=? and state = 0 "
	_stockServerConfByID       = "SELECT id , resource_id,resource_ver,foreign_act_id,describe_info  ,ctime ,mtime , rules_info , stock_start_time , stock_end_time FROM act_stock_server_conf WHERE id=? and state = 0 "
	_stockServerConfListByIDs  = "SELECT id , resource_id,resource_ver,foreign_act_id,describe_info  ,ctime ,mtime , rules_info , stock_start_time , stock_end_time FROM act_stock_server_conf WHERE id IN (%s) and state = 0 "
	_stockServerConfListByTime = "SELECT id , resource_id,resource_ver,foreign_act_id,describe_info  ,ctime ,mtime , rules_info , stock_start_time , stock_end_time FROM act_stock_server_conf WHERE stock_start_time < ? and stock_end_time > ? and state = 0 order by id asc limit ? , ?"
)

func (d *Dao) AddStockServerConf(ctx context.Context, value *stock.ConfItemDB) (int64, error) {
	row, err := d.db.Exec(ctx, _addStockServerConf, value.ResourceId, value.ResourceVer, value.ForeignActId, value.DescribeInfo, value.RulesInfo, value.StockStartTime, value.StockEndTime)
	if err != nil {
		return 0, errors.Wrap(err, "AddUserAwardPackage")
	}
	return row.LastInsertId()
}

func (d *Dao) UpdateStockServerConf(ctx context.Context, value *stock.ConfItemDB) (int64, error) {
	res, err := d.db.Exec(ctx, _updateStockServerConf, value.ResourceId, value.ResourceVer, value.ForeignActId, value.DescribeInfo, value.RulesInfo, value.StockStartTime, value.StockEndTime, value.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (d *Dao) RawGetConfListByRid(ctx context.Context, rid string, foreignId string) (res []*stock.ConfItemDB, err error) {
	rows, err := d.db.Query(ctx, _stockServerConfListByRid, rid, foreignId)
	if err != nil {
		return nil, errors.Wrap(err, "RawUserPackage Query")
	}
	defer rows.Close()
	for rows.Next() {
		conf := stock.ConfItemDB{}
		if err = rows.Scan(&conf.ID, &conf.ResourceId, &conf.ResourceVer, &conf.ForeignActId,
			&conf.DescribeInfo, &conf.Ctime, &conf.Mtime, &conf.RulesInfo, &conf.StockStartTime, &conf.StockEndTime); err != nil {
			return nil, errors.Wrap(err, "RawGetConfListByRid Scan")
		}
		res = append(res, &conf)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawGetConfListByRid rows")
	}
	return res, nil
}

func (d *Dao) RawGetConfByID(ctx context.Context, id int64) (conf *stock.ConfItemDB, err error) {
	row := d.db.QueryRow(ctx, _stockServerConfByID, id)
	conf = new(stock.ConfItemDB)
	if err := row.Scan(&conf.ID, &conf.ResourceId, &conf.ResourceVer, &conf.ForeignActId,
		&conf.DescribeInfo, &conf.Ctime, &conf.Mtime, &conf.RulesInfo, &conf.StockStartTime, &conf.StockEndTime); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "RawGetConfListByID Scan")
	}
	return conf, nil
}

func (d *Dao) RawGetConfListByIDs(ctx context.Context, ids []int64) (confList []*stock.ConfItemDB, err error) {
	var rows *sql.Rows
	rows, err = d.db.Query(ctx, fmt.Sprintf(_stockServerConfListByIDs, xstr.JoinInts(ids)))
	if err != nil {
		return nil, errors.Wrap(err, "RawGetConfListByIDs Query")
	}
	defer rows.Close()
	for rows.Next() {
		var conf = new(stock.ConfItemDB)
		if err := rows.Scan(&conf.ID, &conf.ResourceId, &conf.ResourceVer, &conf.ForeignActId,
			&conf.DescribeInfo, &conf.Ctime, &conf.Mtime, &conf.RulesInfo, &conf.StockStartTime, &conf.StockEndTime); err != nil {
			return nil, errors.Wrap(err, "RawGetConfListByIDs Scan")
		}
		confList = append(confList, conf)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "RawGetConfListByIDs rows")
	}
	return confList, nil
}

func (d *Dao) RawGetConfListByTime(ctx context.Context, beginTime, endTime int64, offset, limit int32) (res []*stock.ConfItemDB, err error) {
	layout := "2006-01-02 15:04:05"
	beginTimeStr := time.Unix(beginTime, 0).Format(layout)
	endTimeStr := time.Unix(endTime, 0).Format(layout)

	rows, err := d.db.Query(ctx, _stockServerConfListByTime, beginTimeStr, endTimeStr, offset, limit)
	log.Infoc(ctx, "RawGetConfListByTime beginTimeStr:%v, endTimeStr:%v , offset:%v ,limit:%v", beginTimeStr, endTimeStr, offset, limit)
	if err != nil {
		return nil, errors.Wrap(err, "RawGetConfListByTime Query")
	}
	defer rows.Close()
	return convert2ConfItemDB(rows)
}

func convert2ConfItemDB(rows *sql.Rows) (res []*stock.ConfItemDB, err error) {
	for rows.Next() {
		conf := stock.ConfItemDB{}
		if err = rows.Scan(&conf.ID, &conf.ResourceId, &conf.ResourceVer, &conf.ForeignActId,
			&conf.DescribeInfo, &conf.Ctime, &conf.Mtime, &conf.RulesInfo, &conf.StockStartTime, &conf.StockEndTime); err != nil {
			return nil, errors.Wrap(err, "convert2ConfItemDB Scan")
		}
		res = append(res, &conf)
	}
	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "convert2ConfItemDB rows")
	}
	return
}
