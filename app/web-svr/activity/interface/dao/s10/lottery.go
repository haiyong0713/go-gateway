package s10

import (
	"context"
	"database/sql"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/model/s10"
	"go-gateway/app/web-svr/activity/interface/tool"
)

const _userLotterySQL = "select robin,gid,extra,state from act_s10_user_gift where mid=? and state != 1;"

func (d *Dao) UserLottery(ctx context.Context, mid int64) (map[int32]*s10.Lucky, error) {
	rows, err := component.S10GlobalDB.Query(ctx, _userLotterySQL, mid)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UserLottery(mid:%d) error(%v)", mid, err)
		return nil, err
	}
	defer rows.Close()
	resMap := make(map[int32]*s10.Lucky, 5)
	for rows.Next() {
		tmp := new(s10.Lucky)
		robin := int32(0)
		if err = rows.Scan(&robin, &tmp.Gid, &tmp.Extra, &tmp.State); err != nil {
			log.Errorc(ctx, "s10 rows.Scan(mid:%d) error(%v)", mid, err)
			return nil, err
		}
		resMap[robin] = tmp

	}
	return resMap, rows.Err()
}

const _userLotteryByRobinSQL = "select gid,extra,state from act_s10_user_gift where mid=? and robin=?;"

func (d *Dao) UserLotteryByRobin(ctx context.Context, mid int64, robin int32) (*s10.Lucky, error) {
	if !tool.IsLimiterAllowedByUniqBizKey(s10.S10LimitTypeGoodsExhcange, s10.S10LimitBusinessUserReceiveGoods) {
		return nil, xecode.LimitExceed
	}
	row := component.S10GlobalDB.Master().QueryRow(ctx, _userLotteryByRobinSQL, mid, robin)
	tmp := new(s10.Lucky)
	err := row.Scan(&tmp.Gid, &tmp.Extra, &tmp.State)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		log.Errorc(ctx, "s10 row.Scan() error(%v)", err)
		return nil, err
	}
	return tmp, nil
}

const _correctUserLotteryByRobinSQL = "update act_s10_user_gift set state=0 where mid=? and robin=?;"

func (d *Dao) CorrectUserLotteryByRobin(ctx context.Context, mid int64, robin int32) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _correctUserLotteryByRobinSQL, mid, robin)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.CorrectUserLotteryByRobin(mid:%d,robin:%d) error(%v)", mid, robin, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _updateUserLotteryStateSQL = "update act_s10_user_gift set state=2, user_number=?, user_name=?, addr=? where mid=? and robin=?;"
const _updateUserLotteryState2SQL = "update act_s10_user_gift set state=2 where mid=? and robin=?;"

func (d *Dao) UpdateUserLotteryState(ctx context.Context, robin, typ int32, mid int64, number, name, addr string) (int64, error) {
	SQL := _updateUserLotteryState2SQL
	params := []interface{}{mid, robin}
	if typ == 0 {
		SQL = _updateUserLotteryStateSQL
		params = []interface{}{number, name, addr, mid, robin}
	}
	row, err := component.S10GlobalDB.Exec(ctx, SQL, params...)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UpdateUserLotteryState(mid:%d,robin:%d) error(%v)", mid, robin, err)
		return 0, nil
	}
	return row.RowsAffected()
}

const _usersLotteryInfoByRobinSQL = "select user_name,bonus from act_s10_lucky_user where robin=? limit 100;"

func (d *Dao) UsersLotteryInfoByRobin(ctx context.Context, robin int32) ([]*s10.UserLotteryInfo, error) {
	rows, err := component.S10GlobalDB.Query(ctx, _usersLotteryInfoByRobinSQL, robin)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UserLotteryInfoByRobin(robin:%d) error(%v)", robin, err)
		return nil, err
	}
	defer rows.Close()
	res := make([]*s10.UserLotteryInfo, 0, 100)
	for rows.Next() {
		tmp := new(s10.UserLotteryInfo)
		if err = rows.Scan(&tmp.Name, &tmp.Bonus); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error(%v)", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, rows.Err()
}

const _ackUserCostGiftSQL = "update act_s10_user_gift set ack=1 where mid=? and robin=?"

func (d *Dao) AckUserCostGift(ctx context.Context, mid int64, robin int32) (int64, error) {
	row, err := component.S10GlobalDB.Exec(ctx, _ackUserCostGiftSQL, mid, robin)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.AckUserCostGift(mid:%d,robin:%d) error:%v", mid, robin, err)
		return 0, err
	}
	return row.RowsAffected()
}
