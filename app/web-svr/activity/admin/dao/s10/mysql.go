package s10

import (
	"context"
	"fmt"
	"strings"

	"go-gateway/app/web-svr/activity/admin/component"
	"go-gateway/app/web-svr/activity/admin/model/s10"

	"go-common/library/log"
	"go-common/library/xstr"
)

const _lotteryUsersByRobinSQL = "select mid from act_s10_lottery_user where robin=? and mid>? order by mid asc limit 1000;"

func (d *Dao) LotteryUsersByRobin(ctx context.Context, robin int32, offset int64) ([]int64, int64, error) {
	rows, err := component.GlobalDB.Query(ctx, _lotteryUsersByRobinSQL, robin, offset)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.LotteryUsersByRobin(offset:%d) error:%v)", offset, err)
		return nil, 0, err
	}
	defer rows.Close()
	res := make([]int64, 0, 100)
	next := offset
	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, 0, err
		}
		res = append(res, mid)
		if mid > next {
			next = mid
		}
	}
	return res, next, nil
}

const _lotteryUsersSQL = "select mid,id from act_s10_lottery_user where id>? limit 200;"

func (d *Dao) LotteryUsers(ctx context.Context, offset int64) ([]int64, int64, error) {
	rows, err := component.GlobalDB.Query(ctx, _lotteryUsersSQL, offset)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.LotteryUsers(offset:%d) error:%v)", offset, err)
		return nil, 0, err
	}
	defer rows.Close()
	res := make([]int64, 0, 200)
	next := offset
	id := int64(0)
	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid, &id); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, 0, err
		}
		res = append(res, mid)
		if id > next {
			next = id
		}
	}
	return res, next, nil
}

const _giftUsersSQL = "select mid,id from act_s10_user_gift where id>? limit 200;"

func (d *Dao) GiftUsers(ctx context.Context, offset int64) ([]int64, int64, error) {
	rows, err := component.GlobalDB.Query(ctx, _giftUsersSQL, offset)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.GiftUsers(offset:%d) error:%v)", offset, err)
		return nil, 0, err
	}
	defer rows.Close()
	res := make([]int64, 0, 200)
	next := offset
	id := int64(0)
	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid, &id); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, 0, err
		}
		res = append(res, mid)
		if id > next {
			next = id
		}
	}
	return res, next, nil
}

const _goodsByRobinSQL = "select id,stock,state,gname,category,score from act_s10_goods where robin=?;"

func (d *Dao) GoodsByRobin(ctx context.Context, robin int32) ([]*s10.Goods, error) {
	rows, err := component.GlobalDB.Query(ctx, _goodsByRobinSQL, robin)
	if err != nil {
		log.Errorc(ctx, "s10 dao.GoodsByRobin(robin:%d) error:%v", robin, err)
		return nil, err
	}
	defer rows.Close()
	res := make([]*s10.Goods, 0, 20)
	for rows.Next() {
		tmp := new(s10.Goods)
		if err = rows.Scan(&tmp.Gid, &tmp.Stock, &tmp.State, &tmp.Gname, &tmp.Type, &tmp.Score); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, nil
}

const _giftCertsByGidSQL = "select cert from act_gift_certificates where gid=?;"

func (d *Dao) GiftCertsByGid(ctx context.Context, gid int32) ([]string, error) {
	rows, err := component.GlobalDB.Query(ctx, _giftCertsByGidSQL, gid)
	if err != nil {
		log.Errorc(ctx, "s10 dao.GiftCertsByGid(gid:%d) error:%v", err)
		return nil, err
	}
	defer rows.Close()
	res := make([]string, 0, 50)
	for rows.Next() {
		var tmp string
		if err = rows.Scan(&tmp); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, nil
}

const _batchAddLuckyCertUser = "insert into act_s10_user_gift (mid,robin,gid,extra) values (?,?,?,?) on duplicate key update gid=?,state=0,extra=?;"

func (d *Dao) BatchAddLuckyCertUser(ctx context.Context, gid, robin int32, mids []int64, certs []string) (int64, error) {
	for i, mid := range mids {
		_, err := component.GlobalDB.Exec(ctx, _batchAddLuckyCertUser, mid, robin, gid, certs[i], gid, certs[i])
		if err != nil {
			log.Errorc(ctx, "s10 d.dao.BatchAddLuckyUser(user:%v) error:%v", mid, err)
			return 0, err
		}
	}
	return 0, nil
}

const _batchAddLuckyUser = "insert into act_s10_user_gift (mid,robin,gid) values (?,?,?) on duplicate key update gid=?,extra='',state=0;"

func (d *Dao) BatchAddLuckyUser(ctx context.Context, gid, robin int32, mids []int64) (int64, error) {
	params := make([]string, 0, len(mids))
	for _, mid := range mids {
		_, err := component.GlobalDB.Exec(ctx, _batchAddLuckyUser, mid, robin, gid, gid)
		if err != nil {
			log.Errorc(ctx, "s10 d.dao.BatchAddLuckyUser(params:%v) error:%v", params, err)
			return 0, err
		}
	}
	return 0, nil
}

const _lotteryUserByGidSQL = "select mid from act_s10_user_gift where gid=?;"

func (d *Dao) LotteryUserByGid(ctx context.Context, gid int32) ([]int64, error) {
	rows, err := component.GlobalDB.Query(ctx, _lotteryUserByGidSQL, gid)
	if err != nil {
		log.Errorc(ctx, "s10 dao.LotteryUserByGid(gid:%d) error:%v", err)
		return nil, err
	}
	defer rows.Close()
	res := make([]int64, 0, 50)
	for rows.Next() {
		var tmp int64
		if err = rows.Scan(&tmp); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, nil
}

const _lotteryUserByGidsSQL = "select mid from act_s10_user_gift where gid in (%s);"

func (d *Dao) LotteryUserByGids(ctx context.Context, gids []int64) ([]int64, error) {
	rows, err := component.GlobalDB.Query(ctx, fmt.Sprintf(_lotteryUserByGidsSQL, xstr.JoinInts(gids)))
	if err != nil {
		log.Errorc(ctx, "s10 dao.LotteryUserByGids(gid:%d) error:%v", err)
		return nil, err
	}
	defer rows.Close()
	res := make([]int64, 0, 50)
	for rows.Next() {
		var tmp int64
		if err = rows.Scan(&tmp); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, nil
}

const _realGoodsUserSQL = "select id,mid,user_name,gid,user_number,addr,state from act_s10_user_gift where gid in (%s) and state=2;"

func (d *Dao) RealGoodsUser(ctx context.Context, gids []int64) ([]*s10.RealGiftRecord, error) {
	rows, err := component.GlobalDB.Query(ctx, fmt.Sprintf(_realGoodsUserSQL, xstr.JoinInts(gids)))
	if err != nil {
		log.Errorc(ctx, "s10 dao.LotteryUserByGid(gid:%d) error:%v", err)
		return nil, err
	}
	defer rows.Close()
	res := make([]*s10.RealGiftRecord, 0, 50)
	for rows.Next() {
		tmp := new(s10.RealGiftRecord)
		if err = rows.Scan(&tmp.ID, &tmp.Mid, &tmp.UserName, &tmp.Gid, &tmp.Number, &tmp.Addr, &tmp.State); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, nil
}

func subTableUserCost(mid int64) string {
	return fmt.Sprintf("%03d", mid%200)
}

const _sentOutGoodsSQL = "update act_s10_user_gift set state=3 where id in (%s)"

func (d *Dao) SentOutGoods(ctx context.Context, ids []int64) (int64, error) {
	rows, err := component.GlobalDB.Exec(ctx, fmt.Sprintf(_sentOutGoodsSQL, xstr.JoinInts(ids)))
	if err != nil {
		log.Errorc(ctx, "s10 dao.SentOutGoods(gid:%d) error:%v", err)
		return 0, err
	}
	return rows.RowsAffected()
}

const _updateLotteryUserStateByGid = "update act_s10_user_gift set state=1 where gid=?"

func (d *Dao) UpdateLotteryUserStateByGid(ctx context.Context, gid int32) (int64, error) {
	rows, err := component.GlobalDB.Exec(ctx, _updateLotteryUserStateByGid, gid)
	if err != nil {
		log.Errorc(ctx, "s10 dao.UpdateLotteryUserStateByGid(gid:%d) error:%v", err)
		return 0, err
	}
	return rows.RowsAffected()
}

const _updateLotteryUserStateByGids = "update act_s10_user_gift set state=1 where gid in (%s)"

func (d *Dao) UpdateLotteryUserStateByGids(ctx context.Context, gids []int64) (int64, error) {
	rows, err := component.GlobalDB.Exec(ctx, fmt.Sprintf(_updateLotteryUserStateByGids, xstr.JoinInts(gids)))
	if err != nil {
		log.Errorc(ctx, "s10 dao.UpdateLotteryUserStateByGids(gid:%d) error:%v", err)
		return 0, err
	}
	return rows.RowsAffected()
}

const _userCostExceptSQL = "select gid,gname,ctime,cost,ack,id from act_s10_user_cost where mid=? and state=0"

func (d *Dao) UserCostExcept(ctx context.Context, mid int64) ([]*s10.UserCostRecord, error) {
	rows, err := component.GlobalDB.Query(ctx, _userCostExceptSQL, mid)
	if err != nil {
		log.Errorc(ctx, "s10 dao.UserCostExcept(gid:%d) error:%v", err)
		return nil, err
	}
	defer rows.Close()
	res := make([]*s10.UserCostRecord, 0, 50)
	for rows.Next() {
		tmp := new(s10.UserCostRecord)
		if err = rows.Scan(&tmp.Gid, &tmp.Name, &tmp.Ctime, &tmp.Cost, &tmp.Ack, &tmp.ID); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, nil
}

const _userCostExceptSubSQL = "select gid,gname,ctime,cost,ack,id from act_s10_user_cost_%s where mid=? and state=0"

func (d *Dao) UserCostExceptSub(ctx context.Context, mid int64) ([]*s10.UserCostRecord, error) {
	rows, err := component.GlobalDB.Query(ctx, fmt.Sprintf(_userCostExceptSubSQL, subTableUserCost(mid)), mid)
	if err != nil {
		log.Errorc(ctx, "s10 dao.UserCostExcept(gid:%d) error:%v", err)
		return nil, err
	}
	defer rows.Close()
	res := make([]*s10.UserCostRecord, 0, 50)
	for rows.Next() {
		tmp := new(s10.UserCostRecord)
		if err = rows.Scan(&tmp.Gid, &tmp.Name, &tmp.Ctime, &tmp.Cost, &tmp.Ack, &tmp.ID); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, err
		}
		res = append(res, tmp)
	}
	return res, nil
}

const _userCostExceptRow = "select gid,gname,ctime,cost,ack,state  from act_s10_user_cost where id=?;"

func (d *Dao) UserCostRecordByID(ctx context.Context, id, mid int64) (*s10.UserCostRecord, error) {
	row := component.GlobalDB.QueryRow(ctx, _userCostExceptRow, id)
	tmp := new(s10.UserCostRecord)
	err := row.Scan(&tmp.Gid, &tmp.Name, &tmp.Ctime, &tmp.Cost, &tmp.Ack, &tmp.State)
	if err != nil {
		log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
		return nil, err
	}
	return tmp, err
}

const _userCostExceptRowSubSQL = "select gid,gname,ctime,cost,ack,state  from act_s10_user_cost_%s where id=?;"

func (d *Dao) UserCostRecordByIDSub(ctx context.Context, id, mid int64) (*s10.UserCostRecord, error) {
	row := component.GlobalDB.QueryRow(ctx, fmt.Sprintf(_userCostExceptRowSubSQL, subTableUserCost(mid)), id)
	tmp := new(s10.UserCostRecord)
	err := row.Scan(&tmp.Gid, &tmp.Name, &tmp.Ctime, &tmp.Cost, &tmp.Ack, &tmp.State)
	if err != nil {
		log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
		return nil, err
	}
	return tmp, err
}

const _userGiftExceptRow = "select gid,robin,state,ack from act_s10_user_gift where id=? and state!=1;"

func (d *Dao) UserGiftByID(ctx context.Context, id int64) (*s10.UserGiftRecord, error) {
	row := component.GlobalDB.QueryRow(ctx, _userGiftExceptRow, id)
	tmp := new(s10.UserGiftRecord)
	err := row.Scan(&tmp.Gid, &tmp.Robin, &tmp.State, &tmp.Ack)
	if err != nil {
		log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
		return nil, err
	}
	return tmp, err
}

const _ackUserCostActSubSQL = "update act_s10_user_cost_%s set ack=1 where id=?;"

func (d *Dao) AckUserCostActSub(ctx context.Context, mid, id int64) (int64, error) {
	row, err := component.GlobalDB.Exec(ctx, fmt.Sprintf(_ackUserCostActSubSQL, subTableUserCost(mid)), id)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.AckUserCostAct(id:%d,mid:%d) error:%v", id, mid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _ackUserCostActSQL = "update act_s10_user_cost set ack=1 where id=?;"

func (d *Dao) AckUserCostAct(ctx context.Context, mid, id int64) (int64, error) {
	row, err := component.GlobalDB.Exec(ctx, _ackUserCostActSQL, id)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.AckUserCostAct(id:%d,mid:%d) error:%v", id, mid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _updateUserCostStateSubSQL = "update act_s10_user_cost_%s set state=1 where id=?;"

func (d *Dao) UpdateUserCostStateSub(ctx context.Context, mid, id int64) (int64, error) {
	row, err := component.GlobalDB.Exec(ctx, fmt.Sprintf(_updateUserCostStateSubSQL, subTableUserCost(mid)), id)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UpdateUserCostState(id:%d,mid:%d) error:%v", id, mid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _updateUserCostStateSQL = "update act_s10_user_cost set state=1 where id=?;"

func (d *Dao) UpdateUserCostState(ctx context.Context, mid, id int64) (int64, error) {
	row, err := component.GlobalDB.Exec(ctx, _updateUserCostStateSQL, id)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.UpdateUserCostState(id:%d,mid:%d) error:%v", id, mid, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _goodsSQL = "select category from act_s10_goods where id=?;"

func (d *Dao) Goods(ctx context.Context, gid int32) (typ int32, err error) {
	row := component.GlobalDB.QueryRow(ctx, _goodsSQL, gid)
	if err = row.Scan(&typ); err != nil {
		log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
	}
	return
}

const _ackUserGiftActSQL = "update act_s10_user_gift set state=2,ack=1 where id=? and state!=1;"

func (d *Dao) AckUserGiftAct(ctx context.Context, id int64) (int64, error) {
	row, err := component.GlobalDB.Exec(ctx, _ackUserGiftActSQL, id)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.AckUserCostAct(id:%d) error:%v", id, err)
		return 0, err
	}
	return row.RowsAffected()
}

const _batchAddSuperLuckyUserSQL = "insert into act_s10_lucky_user(robin,user_name,bonus) values %s;"

func (d *Dao) BatchAddSuperLuckyUser(ctx context.Context, users []*s10.SuperLotteryUserInfo) (int64, error) {
	params := make([]string, 0, len(users))
	for _, v := range users {
		params = append(params, fmt.Sprintf("(%d,'%s','%s')", v.Robin, v.Name, v.Gname))
	}
	rows, err := component.GlobalDB.Exec(ctx, fmt.Sprintf(_batchAddSuperLuckyUserSQL, strings.Join(params, ",")))
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.BatchAddSuperLuckyUser(params:%v) error:%v", params, err)
		return 0, err
	}
	return rows.LastInsertId()
}

const _delSuperLuckyUserSQL = "delete from act_s10_lucky_user where robin=?;"

func (d *Dao) DelSuperLuckyUser(ctx context.Context, robin int32) (int64, error) {
	rows, err := component.GlobalDB.Exec(ctx, _delSuperLuckyUserSQL, robin)
	if err != nil {
		log.Errorc(ctx, "s10 d.dao.DelSuperLuckyUser(robin:%d) error:%v", robin, err)
		return 0, err
	}
	return rows.RowsAffected()
}

const _batchAddSuperGiftUserSQL = "insert into act_s10_user_gift(mid,robin,gid,extra) values (?,?,?,?) on duplicate key update gid=?,extra='',state=0;"

func (d *Dao) BatchAddSuperGiftUser(ctx context.Context, users []*s10.SuperLotteryUserInfo) (int64, error) {
	for _, v := range users {
		_, err := component.GlobalDB.Exec(ctx, _batchAddSuperGiftUserSQL, v.Mid, v.Robin, v.Gid, "", v.Gid)
		if err != nil {
			log.Errorc(ctx, "s10 d.dao.BatchAddSuperLuckyUser(params:%v) error:%v", v, err)
			return 0, err
		}
	}
	return 0, nil
}

const _existUsersByRobinSQL = "select mid from act_s10_user_gift where mid in (%s) and robin=? and state!=1;"

func (d *Dao) ExistUsersByRobin(ctx context.Context, mids []int64, robin int32) ([]int64, error) {
	rows, err := component.GlobalDB.Query(ctx, fmt.Sprintf(_existUsersByRobinSQL, xstr.JoinInts(mids)), robin)
	if err != nil {
		log.Errorc(ctx, "s10 dao.ExistUsersByRobin(robin:%d) error:%v", err)
		return nil, err
	}
	defer rows.Close()
	res := make([]int64, 0, len(mids))
	for rows.Next() {
		var mid int64
		if err = rows.Scan(&mid); err != nil {
			log.Errorc(ctx, "s10 rows.Scan() error:%v", err)
			return nil, err
		}
		res = append(res, mid)
	}
	return res, nil
}

const _userCostRecordSQL = "select gid,cost,gname,ctime from act_s10_user_cost where mid=? and state=0;"

func (d *Dao) UserCostRecord(ctx context.Context, mid int64) ([]*s10.CostRecord, error) {
	rows, err := component.GlobalDB.Query(ctx, _userCostRecordSQL, mid)
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
	return res, nil
}

const _userCostRecordSubSQL = "select gid,cost,gname,ctime from act_s10_user_cost_%s where mid=? and state=0;"

func (d *Dao) UserCostRecordSub(ctx context.Context, mid int64) ([]*s10.CostRecord, error) {
	rows, err := component.GlobalDB.Query(ctx, fmt.Sprintf(_userCostRecordSubSQL, subTableUserCost(mid)), mid)
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
	return res, nil
}

const _userLotteryInfoSQL = "select gid from act_s10_user_cost where mid=? and state=0 and gid in (%s) "

func (d *Dao) UserLotteryInfo(ctx context.Context, mid int64, gids []int64) ([]int32, error) {
	rows, err := component.GlobalDB.Query(ctx, fmt.Sprintf(_userLotteryInfoSQL, xstr.JoinInts(gids)), mid)
	if err != nil {
		log.Errorc(ctx, "d.dao.UserLotteryInfo(mid:%d,gids:%v) error(%v)", mid, gids, err)
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
	return res, nil
}

const _userLotteryInfoSubSQL = "select gid from act_s10_user_cost_%s where mid=? and state=0 and gid in (%s) "

func (d *Dao) UserLotterySubInfo(ctx context.Context, mid int64, gids []int64) ([]int32, error) {
	rows, err := component.GlobalDB.Query(ctx, fmt.Sprintf(_userLotteryInfoSubSQL, subTableUserCost(mid), xstr.JoinInts(gids)), mid)
	if err != nil {
		log.Errorc(ctx, "d.dao.UserLotteryInfo(mid:%d,gids:%v) error(%v)", mid, gids, err)
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
	return res, nil
}

const _userLotterySQL = "select robin,gid,extra,state from act_s10_user_gift where mid=? and state!=1;"

func (d *Dao) UserLottery(ctx context.Context, mid int64) (map[int32]*s10.Lucky, error) {
	rows, err := component.GlobalDB.Query(ctx, _userLotterySQL, mid)
	if err != nil {
		log.Errorc(ctx, "d.dao.UserLottery(mid:%d) error(%v)", mid, err)
		return nil, err
	}
	defer rows.Close()
	resMap := make(map[int32]*s10.Lucky, 5)
	for rows.Next() {
		tmp := new(s10.Lucky)
		robin := int32(0)
		if err = rows.Scan(&robin, &tmp.Gid, &tmp.Extra, &tmp.State); err != nil {
			log.Errorc(ctx, "rows.Scan(mid:%d) error(%v)", mid, err)
			return nil, err
		}
		resMap[robin] = tmp

	}
	return resMap, nil
}

const _userGiftSQL = "select id,robin,gid,state,ack from act_s10_user_gift where mid=? and state!=1;"

func (d *Dao) UserGift(ctx context.Context, mid int64) ([]*s10.UserGiftRecord, []int64, error) {
	rows, err := component.GlobalDB.Query(ctx, _userGiftSQL, mid)
	if err != nil {
		log.Errorc(ctx, "d.dao.UserLottery(mid:%d) error(%v)", mid, err)
		return nil, nil, err
	}
	defer rows.Close()
	res := make([]*s10.UserGiftRecord, 0, 5)
	gids := make([]int64, 0, 5)
	for rows.Next() {
		tmp := new(s10.UserGiftRecord)
		if err = rows.Scan(&tmp.ID, &tmp.Robin, &tmp.Gid, &tmp.State, &tmp.Ack); err != nil {
			log.Errorc(ctx, "rows.Scan(mid:%d) error(%v)", mid, err)
			return nil, nil, err
		}
		res = append(res, tmp)
		gids = append(gids, int64(tmp.Gid))

	}
	return res, gids, nil
}

const _goodsesSQL = "select id,gname,category from act_s10_goods where id in (%s)"

func (d *Dao) Goodses(ctx context.Context, gids []int64) (map[int32]*s10.Goods, error) {
	rows, err := component.GlobalDB.Query(ctx, _goodsesSQL, xstr.JoinInts(gids))
	if err != nil {
		log.Errorc(ctx, "d.dao.Goodses(gids:%v) error(%v)", gids, err)
		return nil, err
	}
	defer rows.Close()
	res := make(map[int32]*s10.Goods, 5)
	for rows.Next() {
		tmp := new(s10.Goods)
		var id int32
		if err = rows.Scan(&id, &tmp.Gname, &tmp.Type); err != nil {
			log.Errorc(ctx, "rows.Scan() error:%v", err)
			return nil, err
		}
		res[id] = tmp
	}
	return res, nil
}
