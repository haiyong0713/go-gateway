package s10

import (
	"context"
	"database/sql"
	"fmt"
	"go-gateway/app/web-svr/activity/job/dao"
	"go-gateway/app/web-svr/activity/job/model/s10"

	"go-common/library/log"
)

const _addLotteryUser = "insert into act_s10_lottery_user(mid,robin) values(?,?);"

func (d *Dao) AddLotteryUser(ctx context.Context, mid int64, robin int32) (int64, error) {
	row, err := dao.GlobalDB.Exec(ctx, _addLotteryUser, mid, robin)
	if err != nil {
		log.Errorc(ctx, "d.dao.AddLotteryUser(mid:%d,robin:%d) error(%v)", mid, robin, err)
		return 0, err
	}
	return row.LastInsertId()
}

const _userCostRecordSQL = "select gid,cost,gname,ctime from act_s10_user_cost where mid=? and state=0;"

func (d *Dao) UserCostRecord(ctx context.Context, mid int64) ([]*s10.CostRecord, error) {
	rows, err := dao.GlobalDB.Query(ctx, _userCostRecordSQL, mid)
	if err != nil {
		log.Errorc(ctx, "d.dao.UserCostRecord(mid:%d) error(%v)", mid, err)
		return nil, err
	}
	defer rows.Close()
	res := make([]*s10.CostRecord, 0, 10)
	for rows.Next() {
		tmp := &s10.CostRecord{}
		if err = rows.Scan(&tmp.Gid, &tmp.Cost, &tmp.Name, &tmp.Ctime); err != nil {
			log.Errorc(ctx, "rows.Scan() error(%v)", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, rows.Err()
}

func subTabCostRecord(mid int64) string {
	return fmt.Sprintf("%03d", mid%200)
}

const _userCostRecordSubTabSQL = "select gid,cost,gname,ctime from act_s10_user_cost_%s where mid=? and state=0;"

func (d *Dao) UserCostRecordSubTab(ctx context.Context, mid int64) ([]*s10.CostRecord, error) {
	rows, err := dao.GlobalDB.Query(ctx, fmt.Sprintf(_userCostRecordSubTabSQL, subTabCostRecord(mid)), mid)
	if err != nil {
		log.Errorc(ctx, "d.dao.UserCostRecord(mid:%d) error(%v)", mid, err)
		return nil, err
	}
	defer rows.Close()
	res := make([]*s10.CostRecord, 0, 10)
	for rows.Next() {
		tmp := &s10.CostRecord{}
		if err = rows.Scan(&tmp.Gid, &tmp.Cost, &tmp.Name, &tmp.Ctime); err != nil {
			log.Errorc(ctx, "rows.Scan() error(%v)", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, rows.Err()
}

const _addRawFreeFlowSQL = "insert ignore into act_s10_tel(tel) values (?);"

func (d *Dao) AddRawFreeFlow(ctx context.Context, tel string) (int64, error) {
	rows, err := dao.GlobalDB.Exec(ctx, _addRawFreeFlowSQL, tel)
	if err != nil {
		log.Errorc(ctx, "s10 AddRawFreeFlow(tel:%s) error:%v", tel, err)
		return 0, err
	}
	return rows.LastInsertId()
}

const _rawFreeFlowSQL = "select id from act_s10_tel where tel=?;"

func (d *Dao) RawFreeFlow(ctx context.Context, tel string) (id int64, err error) {
	row := dao.GlobalDB.QueryRow(ctx, _rawFreeFlowSQL, tel)
	if err = row.Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Errorc(ctx, "s10 RawFreeFlow(tel:%s) error:%v", tel, err)
	}
	return
}

const _addFreeFlowUserSQL = "insert ignore into act_s10_user_flow(mid,source) values (?,?);"

func (d *Dao) AddFreeFlowUser(ctx context.Context, mid int64, source int32) (int64, error) {
	rows, err := dao.GlobalDB.Exec(ctx, _addFreeFlowUserSQL, mid, source)
	if err != nil {
		log.Errorc(ctx, "s10 AddRawFreeFlow(mid:%d,source:%d) error:%v", mid, source, err)
		return 0, err
	}
	return rows.LastInsertId()
}
