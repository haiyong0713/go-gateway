package s10

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-gateway/app/web-svr/activity/interface/model/s10"
	"go-gateway/app/web-svr/activity/interface/tool"

	"go-common/library/database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-common/library/xstr"
)

const _userCostRecordSQL = "select gid,cost,gname,ctime from act_s10_user_cost where mid=? and state=0;"

func UserCostRecord(ctx context.Context, mid int64) ([]*s10.CostRecord, error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeBackToData, s10.S10LimitBusinessPointsDetail) {
		return nil, xecode.LimitExceed
	}
	rows, err := component.S10GlobalDB.Query(ctx, _userCostRecordSQL, mid)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UserCostRecord(mid:%d) error(%v)", mid, err)
		return nil, err
	}
	defer rows.Close()
	res := make([]*s10.CostRecord, 0, 10)
	for rows.Next() {
		tmp := &s10.CostRecord{}
		if err = rows.Scan(&tmp.Gid, &tmp.Cost, &tmp.Name, &tmp.Ctime); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error(%v)", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	err = rows.Err()
	return res, nil
}

const _userCostRecordSubSQL = "select gid,cost,gname,ctime from act_s10_user_cost_%s where mid=? and state=0;"

func UserCostRecordSub(ctx context.Context, mid int64) ([]*s10.CostRecord, error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeBackToData, s10.S10LimitBusinessPointsDetail) {
		return nil, xecode.LimitExceed
	}
	rows, err := component.S10GlobalDB.Query(ctx, fmt.Sprintf(_userCostRecordSubSQL, subTabCostRecord(mid)), mid)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UserCostRecord(mid:%d) error(%v)", mid, err)
		return nil, err
	}
	defer rows.Close()
	res := make([]*s10.CostRecord, 0, 10)
	for rows.Next() {
		tmp := &s10.CostRecord{}
		if err = rows.Scan(&tmp.Gid, &tmp.Cost, &tmp.Name, &tmp.Ctime); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error(%v)", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, rows.Err()
}

const _userCountCostSQL = "select COALESCE(sum(cost),0) from act_s10_user_cost where mid=? and state=0;"

func (d *Dao) UserCountCost(ctx context.Context, mid int64) (total int32, err error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeBackToData, s10.S10LimitBusinessPoints) {
		return 0, xecode.LimitExceed
	}
	row := component.S10GlobalDB.QueryRow(ctx, _userCountCostSQL, mid)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Errorc(ctx, "s10 row.Scan() error(%v)", err)
	}
	return
}

func (d *Dao) UserCountCostMaster(ctx context.Context, mid int64) (total int32, err error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeGoodsExhcange, s10.S10LimitBusinessUserExchangeGoods) {
		return 0, xecode.LimitExceed
	}
	row := component.S10GlobalDB.Master().QueryRow(ctx, _userCountCostSQL, mid)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Errorc(ctx, "s10 row.Scan() error(%v)", err)
	}
	return
}

const _userCountCostSubSQL = "select COALESCE(sum(cost),0) from act_s10_user_cost_%s where mid=? and state=0;"

func (d *Dao) UserCountCostSub(ctx context.Context, mid int64) (total int32, err error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeBackToData, s10.S10LimitBusinessPoints) {
		return 0, xecode.LimitExceed
	}
	row := component.S10GlobalDB.QueryRow(ctx, fmt.Sprintf(_userCountCostSubSQL, subTabCostRecord(mid)), mid)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Errorc(ctx, "s10 row.Scan() error(%v)", err)
	}
	return
}

func (d *Dao) UserCountCostMasterSub(ctx context.Context, mid int64) (total int32, err error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeGoodsExhcange, s10.S10LimitBusinessUserExchangeGoods) {
		return 0, xecode.LimitExceed
	}
	row := component.S10GlobalDB.Master().QueryRow(ctx, fmt.Sprintf(_userCountCostSubSQL, subTabCostRecord(mid)), mid)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Errorc(ctx, "s10 row.Scan() error(%v)", err)
	}
	return
}

const _addUserCostRecordSQL = "insert into act_s10_user_cost(mid,gid,cost,gname,ctime) values(?,?,?,?,?);"

func (d *Dao) AddUserCostRecord(ctx context.Context, mid int64, gid, cost int32, name string, addTime xtime.Time) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _addUserCostRecordSQL, mid, gid, cost, name, addTime)
	if err != nil {
		log.Errorc(ctx, "s10 db.Exec() error(%v)", err)
		return 0, err
	}
	return row.LastInsertId()
}

func subTabCostRecord(mid int64) string {
	return fmt.Sprintf("%03d", mid%200)
}

const _addUserCostRecordToSubTabSQL = "insert into act_s10_user_cost_%s(mid,gid,cost,gname,ctime) values(?,?,?,?,?);"

func (d *Dao) AddUserCostRecordToSubTab(ctx context.Context, mid int64, gid, cost int32, name string, addTime xtime.Time) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, fmt.Sprintf(_addUserCostRecordToSubTabSQL, subTabCostRecord(mid)), mid, gid, cost, name, addTime)
	if err != nil {
		log.Errorc(ctx, "s10 db.Exec() error(%v)", err)
		return 0, err
	}
	return row.LastInsertId()
}

const (
	_userCostRecordCountByGid1SQL = "select count(*) from act_s10_user_cost where mid=? and state=0 and gid=?;"
	_userCostRecordCountByGid2SQL = "select count(*) from act_s10_user_cost where mid=? and state=0 and gid=? and ctime>?;"
)

func (d *Dao) UserCostRecordCountByGid(ctx context.Context, mid int64, gid int32, currdate xtime.Time) (res int32, err error) {
	SQL := _userCostRecordCountByGid1SQL
	params := []interface{}{mid, gid}
	if currdate > 0 {
		SQL = _userCostRecordCountByGid2SQL
		params = append(params, currdate)
	}
	row := component.S10GlobalDB.Master().QueryRow(ctx, SQL, params...)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Errorc(ctx, "s10 d.dao.UserMatchStageCostRecord(mid:%d,gid:%d,) error(%v)", mid, gid, err)
	}
	return
}

const (
	_userCostRecordCountByGidSub1SQL = "select count(*) from act_s10_user_cost_%s where mid=? and state=0 and gid=?;"
	_userCostRecordCountByGidSub2SQL = "select count(*) from act_s10_user_cost_%s where mid=? and state=0 and gid=? and ctime>?;"
)

func (d *Dao) UserCostRecordCountByGidSub(ctx context.Context, mid int64, gid int32, currdate xtime.Time) (res int32, err error) {
	SQL := _userCostRecordCountByGidSub1SQL
	params := []interface{}{mid, gid}
	if currdate > 0 {
		SQL = _userCostRecordCountByGidSub2SQL
		params = append(params, currdate)
	}
	row := component.S10GlobalDB.Master().QueryRow(ctx, fmt.Sprintf(SQL, subTabCostRecord(mid)), params...)
	if err = row.Scan(&res); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Errorc(ctx, "s10 d.dao.UserCostRecordCountByGidSub(mid:%d,gid:%d,) error(%v)", mid, gid, err)
	}
	return
}

const _updateUserCostRecordStateSQL = "update act_s10_user_cost set state=1 where id=?'"

func (d *Dao) UpdateUserCostRecordState(ctx context.Context, id int64) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _updateUserCostRecordStateSQL, id)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UpdateUserCostRecordState(id:%d) error(%v)", id, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _updateUserCostRecordStateSubSQL = "update act_s10_user_cost_%s set state=1 where id=?'"

func (d *Dao) UpdateUserCostRecordStateSub(ctx context.Context, id, mid int64) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, fmt.Sprintf(_updateUserCostRecordStateSubSQL, subTabCostRecord(mid)), id)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UpdateUserCostRecordStateSub(id:%d) error(%v)", id, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _userLotteryInfoSQL = "select gid from act_s10_user_cost where mid=? and state=0 and gid in (%s) "

func (d *Dao) UserLotteryInfo(ctx context.Context, mid int64, gids []int64) ([]int32, error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeBackToData, s10.S10LimitBusinessUserLotteryGoods) {
		return nil, xecode.LimitExceed
	}
	rows, err := component.S10GlobalDB.Query(ctx, fmt.Sprintf(_userLotteryInfoSQL, xstr.JoinInts(gids)), mid)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UserLotteryInfo(mid:%d,gids:%v) error(%v)", mid, gids, err)
		return nil, err
	}
	defer rows.Close()
	res := make([]int32, 0, len(gids))
	for rows.Next() {
		var gid int32
		if err = rows.Scan(&gid); err != nil {
			log.Errorc(ctx, "rows.Scan() error(%v)", err)
			return nil, err
		}
		res = append(res, gid)
	}
	return res, rows.Err()
}

const _userLotteryInfoSubSQL = "select gid from act_s10_user_cost_%s where mid=? and state=0 and gid in (%s) "

func (d *Dao) UserLotteryInfoSub(ctx context.Context, mid int64, gids []int64) ([]int32, error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeBackToData, s10.S10LimitBusinessUserLotteryGoods) {
		return nil, xecode.LimitExceed
	}
	rows, err := component.S10GlobalDB.Query(ctx, fmt.Sprintf(_userLotteryInfoSubSQL, subTabCostRecord(mid), xstr.JoinInts(gids)), mid)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UserLotteryInfoSub(mid:%d,gids:%v) error(%v)", mid, gids, err)
		return nil, err
	}
	defer rows.Close()
	res := make([]int32, 0, len(gids))
	for rows.Next() {
		var gid int32
		if err = rows.Scan(&gid); err != nil {
			log.Errorc(ctx, "rows.Scan() error(%v)", err)
			return nil, err
		}
		res = append(res, gid)
	}
	return res, rows.Err()
}

const _ackUserCostRecordSQL = "update act_s10_user_cost set ack=1 where id=?"

func (d *Dao) AckUserCostRecord(ctx context.Context, id int64) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _ackUserCostRecordSQL, id)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.AckUserCostRecord(id:%d) error:%v", id, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _ackUserCostRecordSubSQL = "update act_s10_user_cost_%s set ack=1 where id=?"

func (d *Dao) AckUserCostRecordSub(ctx context.Context, id, mid int64) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, fmt.Sprintf(_ackUserCostRecordSubSQL, subTabCostRecord(mid)), id)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.AckUserCostRecordSub(id:%d) error:%v", id, err)
		return 0, err
	}
	return row.RowsAffected()
}
