package bws

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"go-gateway/app/web-svr/activity/ecode"

	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

const (
	_bindingSQL     = "UPDATE act_bws_users SET mid = ? WHERE `key`= ?"
	_userVipKeySQL  = "UPDATE act_bws_vip_token SET mid = ?,bws_date=? WHERE `vip_key`= ? and bid=?"
	_newUserSQL     = "INSERT act_bws_users (bid,mid,`key`) VALUES (?,?,?)"
	_usersMidSQL    = "SELECT id,mid,`key`,ctime,mtime,bid FROM act_bws_users WHERE bid = ? AND mid = ?"
	_usersKeySQL    = "SELECT id,mid,`key`,ctime,mtime,bid FROM act_bws_users WHERE bid = ? AND `key` = ?"
	_bidUsersMidSQL = "SELECT id,mid,`key`,ctime,mtime,bid FROM act_bws_users WHERE bid IN (%s) AND mid = ?"
	_usersMidsSQL   = "SELECT id,mid,`key`,ctime,mtime,bid FROM act_bws_users WHERE bid = ? AND `mid` in (%s)"
	_usersBidsSQL   = "SELECT id,mid,`key`,ctime,mtime,bid FROM act_bws_users WHERE bid in (%s) AND `mid` = ?"
	_usersIDSQL     = "SELECT id,mid,`key`,ctime,mtime,bid FROM act_bws_users WHERE id = ?"
	_userMoreSQL    = "SELECT id,mid,`key`,ctime,mtime,bid FROM act_bws_users WHERE bid in (%s) AND id > ? AND mid > 0 ORDER BY id ASC LIMIT 500"
	_usersKeysSQL   = "SELECT id,mid,`key`,ctime,mtime,bid FROM act_bws_users WHERE bid = ? AND `key` IN (%s)"

	_usersDetailByMidSQL          = "SELECT id,bid,mid,heart,star,lottery_used,star_detail,bws_date,star_last_time,ups,star_in_rank,play_times,play_success_times,ctime,mtime FROM act_bws_user_detail WHERE mid = ? and bid = ? and bws_date = ? and state=1"
	_vipKeyByKeySQL               = "SELECT id,mid,vip_key,ctime,mtime,bid,bws_date FROM act_bws_vip_token WHERE  bid = ? and vip_key = ? and del=0"
	_vipKeyByMidDateSQL           = "SELECT id,mid,vip_key,ctime,mtime,bid,bws_date FROM act_bws_vip_token WHERE  bid = ? and mid = ? and bws_date = ? and del=0"
	_usersDetailByMidForUpdateSQL = "SELECT id,bid,mid,heart,star,lottery_used,star_detail,bws_date,star_last_time,ups,star_in_rank,play_times,play_success_times,ctime,mtime FROM act_bws_user_detail WHERE mid = ? and bid = ? and bws_date = ? and state=1 for update"
	_usersDetailByMidsSQL         = "SELECT id,bid,mid,heart,star,lottery_used,star_detail,bws_date,star_last_time,ups,star_in_rank,play_times,play_success_times,ctime,mtime FROM act_bws_user_detail WHERE mid in (%s) AND `bid` = ? and bws_date = ?  and state=1"
	_newDetailByMidSQL            = "INSERT act_bws_user_detail (bid,mid,bws_date,heart) VALUES (?,?,?,?)"
	_userDetailUpdateSQL          = "UPDATE act_bws_user_detail SET star = ?,heart=?,lottery_used=?,star_detail=?,star_last_time=?,ups=?,star_in_rank=?,play_times=play_times+?,play_success_times=play_success_times+? WHERE id= ?"
	_userTopSQL                   = "SELECT id,bid,mid,heart,star,lottery_used,star_detail,bws_date,star_last_time,star_in_rank,play_times,play_success_times,ctime,mtime FROM act_bws_user_detail WHERE bid = ? and bws_date = ? and state = 1 order by star desc,star_last_time asc LIMIT ?"

	_newHeartLogSQL = "INSERT act_bws_user_heart_log (bid,mid,heart,reason,order_no,token) VALUES (?,?,?,?,?,?)"
	_newStarLogSQL  = "INSERT act_bws_user_star_log (bid,mid,star,reason,order_no,token) VALUES (?,?,?,?,?,?)"

	// bws catch up
	_inCatchUserSQL  = "INSERT IGNORE INTO act_bws_users_catch (bid,user_key,up_mid) VALUES (?,?,?) ON DUPLICATE KEY UPDATE bid=?,user_key=?,up_mid=?"
	_catchUserSQL    = "SELECT up_mid,user_key,ctime,mtime FROM act_bws_users_catch WHERE user_key = ? AND bid = ? AND del = 0 ORDER BY id ASC LIMIT 500"
	_bluetoothUpsSQL = "SELECT `id`,`mid`,`blue_key`,`bid`,`desc`,`ctime`,`mtime` FROM act_bws_bluetooth_ups WHERE del = 0 and bid=?"
)

// RawUsersByBid .
func (d *Dao) RawUsersByBid(c context.Context, bids []int64, id int64) (list []*bwsmdl.Users, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_userMoreSQL, xstr.JoinInts(bids)), id); err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		res := &bwsmdl.Users{}
		if err = rows.Scan(&res.ID, &res.Mid, &res.Key, &res.Ctime, &res.Mtime, &res.Bid); err != nil {
			log.Errorc(c, "RawUsersByBid:row.Scan error(%v)", err)
			return
		}
		list = append(list, res)
	}
	err = rows.Err()
	return
}

// RawUsersMid get users by mid
func (d *Dao) RawUsersMid(c context.Context, bid, mid int64) (res *bwsmdl.Users, err error) {
	res = &bwsmdl.Users{}
	row := d.db.QueryRow(c, _usersMidSQL, bid, mid)
	if err = row.Scan(&res.ID, &res.Mid, &res.Key, &res.Ctime, &res.Mtime, &res.Bid); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "RawUsersMid:row.Scan error(%v)", err)
		}
	}
	return
}

// RawUsersKey get users by key
func (d *Dao) RawUsersKey(c context.Context, bid int64, key string) (res *bwsmdl.Users, err error) {
	res = &bwsmdl.Users{}
	row := d.db.QueryRow(c, _usersKeySQL, bid, key)
	if err = row.Scan(&res.ID, &res.Mid, &res.Key, &res.Ctime, &res.Mtime, &res.Bid); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "RawUsersKey:row.Scan error(%v)", err)
		}
	}
	return
}

// RawUserDetail get users detail by mid
func (d *Dao) RawUserDetail(c context.Context, bid, mid int64, date string) (res *bwsmdl.UserDetail, err error) {
	res = &bwsmdl.UserDetail{}
	row := d.db.QueryRow(c, _usersDetailByMidSQL, mid, bid, date)
	if err = row.Scan(&res.Id, &res.Bid, &res.Mid, &res.Heart, &res.Star, &res.LotteryUsed, &res.StarDetail, &res.BwsDate, &res.StarLastTime, &res.Ups, &res.StarInRank, &res.PlayTimes, &res.PlaySuccessTimes, &res.Ctime, &res.Mtime); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "RawUserDetail:row.Scan error(%v)", err)
		}
	}
	return
}

// RawUserDetails ...
func (d *Dao) RawUserDetails(c context.Context, mids []int64, bid int64, date string) (list map[int64]*bwsmdl.UserDetail, err error) {
	var (
		strLen = len(mids)
		rows   *xsql.Rows
	)
	if strLen == 0 {
		return
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_usersDetailByMidsSQL, xstr.JoinInts(mids)), bid, date); err != nil {
		return
	}
	defer rows.Close()
	list = make(map[int64]*bwsmdl.UserDetail, strLen)
	for rows.Next() {
		res := &bwsmdl.UserDetail{}
		if err = rows.Scan(&res.Id, &res.Bid, &res.Mid, &res.Heart, &res.Star, &res.LotteryUsed, &res.StarDetail, &res.BwsDate, &res.StarLastTime, &res.Ups, &res.StarInRank, &res.PlayTimes, &res.PlaySuccessTimes, &res.Ctime, &res.Mtime); err != nil {
			log.Errorc(c, "RawUserDetails:row.Scan error(%v)", err)
			return nil, nil
		}
		list[res.Mid] = res
	}
	err = rows.Err()
	return
}

// RawUsersVipKey ...
func (d *Dao) RawUsersVipKey(c context.Context, bid int64, vipKey string) (res *bwsmdl.VipUsersToken, err error) {
	res = &bwsmdl.VipUsersToken{}
	row := d.db.QueryRow(c, _vipKeyByKeySQL, bid, vipKey)
	if err = row.Scan(&res.ID, &res.Mid, &res.VipKey, &res.Ctime, &res.Mtime, &res.Bid, &res.BwsDate); err != nil {
		if err == xsql.ErrNoRows {
			err = ecode.ActivityBwsVipKeyErr
		} else {
			log.Errorc(c, "RawUserDetail:row.Scan error(%v)", err)
		}
	}
	return
}

// RawUsersVipMidDate ...
func (d *Dao) RawUsersVipMidDate(c context.Context, bid int64, mid int64, date string) (res *bwsmdl.VipUsersToken, err error) {
	res = &bwsmdl.VipUsersToken{}
	row := d.db.QueryRow(c, _vipKeyByMidDateSQL, bid, mid, date)
	if err = row.Scan(&res.ID, &res.Mid, &res.VipKey, &res.Ctime, &res.Mtime, &res.Bid, &res.BwsDate); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "RawUserDetail:row.Scan error(%v)", err)
		}
	}
	return
}

// RawUserDetailForUpdate get users detail by mid
func (d *Dao) RawUserDetailForUpdate(c context.Context, tx *xsql.Tx, bid, mid int64, date string) (res *bwsmdl.UserDetail, err error) {
	res = &bwsmdl.UserDetail{}
	row := tx.QueryRow(_usersDetailByMidForUpdateSQL, mid, bid, date)
	if err = row.Scan(&res.Id, &res.Bid, &res.Mid, &res.Heart, &res.Star, &res.LotteryUsed, &res.StarDetail, &res.BwsDate, &res.StarLastTime, &res.Ups, &res.StarInRank, &res.PlayTimes, &res.PlaySuccessTimes, &res.Ctime, &res.Mtime); err != nil {
		log.Errorc(c, "RawUserDetailForUpdate:row.Scan error(%v)", err)
		return
	}
	return
}

// RawUsersBids .
func (d *Dao) RawUsersBids(c context.Context, bids []int64, mid int64) (list map[int64]*bwsmdl.Users, err error) {
	var (
		strLen = len(bids)
		rows   *xsql.Rows
	)
	if strLen == 0 {
		return
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_usersBidsSQL, xstr.JoinInts(bids)), mid); err != nil {
		return
	}
	defer rows.Close()
	list = make(map[int64]*bwsmdl.Users, strLen)
	for rows.Next() {
		res := &bwsmdl.Users{}
		if err = rows.Scan(&res.ID, &res.Mid, &res.Key, &res.Ctime, &res.Mtime, &res.Bid); err != nil {
			log.Errorc(c, "RawUsersKey:row.Scan error(%v)", err)
			return
		}
		list[res.Bid] = res
	}
	err = rows.Err()
	return
}

// RawUsersMids .
func (d *Dao) RawUsersMids(c context.Context, bid int64, mids []int64) (list map[int64]*bwsmdl.Users, err error) {
	var (
		strLen = len(mids)
		rows   *xsql.Rows
	)
	if strLen == 0 {
		return
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_usersMidsSQL, xstr.JoinInts(mids)), bid); err != nil {
		return
	}
	defer rows.Close()
	list = make(map[int64]*bwsmdl.Users, strLen)
	for rows.Next() {
		res := &bwsmdl.Users{}
		if err = rows.Scan(&res.ID, &res.Mid, &res.Key, &res.Ctime, &res.Mtime, &res.Bid); err != nil {
			log.Errorc(c, "RawUsersKey:row.Scan error(%v)", err)
			return
		}
		list[res.Mid] = res
	}
	err = rows.Err()
	return
}

// Binding binding mid
func (d *Dao) Binding(c context.Context, loginMid int64, p *bwsmdl.ParamBinding) (err error) {
	if _, err = d.db.Exec(c, _bindingSQL, loginMid, p.Key); err != nil {
		log.Error("Binding: db.Exec(%d,%s) error(%v)", loginMid, p.Key, err)
	}
	return
}

// UseVipKey 使用vipkey
func (d *Dao) UseVipKey(c context.Context, loginMid int64, vipkey string, date string, bid int64) (err error) {
	if _, err = d.db.Exec(c, _userVipKeySQL, loginMid, date, vipkey, bid); err != nil {
		log.Error("Binding: db.Exec(%d,%s,%s) error(%v)", loginMid, vipkey, date, err)
	}
	return

}

// UserByID .
func (d *Dao) UserByID(c context.Context, keyID int64) (res *bwsmdl.Users, err error) {
	res = &bwsmdl.Users{}
	row := d.db.QueryRow(c, _usersIDSQL, keyID)
	if err = row.Scan(&res.ID, &res.Mid, &res.Key, &res.Ctime, &res.Mtime, &res.Bid); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			log.Errorc(c, "UserByID:row.Scan error(%v)", err)
		}
	}
	return
}

// CreateUser insert user data.
func (d *Dao) CreateUser(c context.Context, bid, mid int64, key string) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _newUserSQL, bid, mid, key); err != nil {
		log.Errorc(c, "CreateUser error d.db.Exec(%d,%d,%s) error(%v)", bid, mid, key, err)
		return
	}
	return res.LastInsertId()
}

// CreateUserDetail insert user data.
func (d *Dao) CreateUserDetail(c context.Context, bid, mid int64, date string, heart int64) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _newDetailByMidSQL, bid, mid, date, heart); err != nil {
		log.Errorc(c, "CreateUserDetail error d.db.Exec(%d,%d,%s) error(%v)", bid, mid, date, err)
		return
	}
	return res.LastInsertId()
}

// RawUserDetailRank .
func (d *Dao) RawUserDetailRank(c context.Context, bid int64, date string, limit int64) ([]*bwsmdl.UserDetail, error) {
	rows, err := d.db.Query(c, _userTopSQL, bid, date, limit)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*bwsmdl.UserDetail
	for rows.Next() {
		b := &bwsmdl.UserDetail{}
		if err := rows.Scan(&b.Id, &b.Bid, &b.Mid, &b.Heart, &b.Star, &b.LotteryUsed, &b.StarDetail, &b.BwsDate, &b.StarLastTime, &b.StarInRank, &b.PlayTimes, &b.PlaySuccessTimes, &b.Ctime, &b.Mtime); err != nil {
			log.Errorc(c, "RawUserDetailRank %+v", err)
			return nil, err
		}
		res = append(res, b)
	}
	return res, rows.Err()
}

// CreateHeartLog insert user data.
func (d *Dao) CreateHeartLog(c context.Context, tx *xsql.Tx, bid, mid, heart int64, reason, orderNo, key string) (lastID int64, err error) {
	var res sql.Result
	if res, err = tx.Exec(_newHeartLogSQL, bid, mid, heart, reason, orderNo, key); err != nil {
		log.Errorc(c, "CreateHeartLog error d.db.Exec(%d,%d,%d,%s,%s,%s) error(%v)", bid, mid, heart, reason, orderNo, key, err)
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = ecode.ActivityBwsDuplicateErr
		}
		return
	}
	return res.LastInsertId()
}

// CreateStarLog insert user data.
func (d *Dao) CreateStarLog(c context.Context, tx *xsql.Tx, bid, mid, star int64, reason, orderNo, key string) (lastID int64, err error) {
	var res sql.Result
	if res, err = tx.Exec(_newStarLogSQL, bid, mid, star, reason, orderNo, key); err != nil {

		log.Errorc(c, "CreateStarLog error d.db.Exec(%d,%d,%d,%s,%s,%s) error(%v)", bid, mid, star, reason, orderNo, key, err)
		if strings.Contains(err.Error(), "Duplicate entry") {
			err = ecode.ActivityBwsDuplicateErr
		}
		return
	}
	return res.LastInsertId()
}

// UpdateUserDetail ...
func (d *Dao) UpdateUserDetail(c context.Context, tx *xsql.Tx, id, star, heart, lastStarTime, lotteryUsed, ups int64, starDetail string, starInRank int64, playTimes, playSuccessTimes int64) (err error) {
	if _, err = tx.Exec(_userDetailUpdateSQL, star, heart, lotteryUsed, starDetail, lastStarTime, ups, starInRank, playTimes, playSuccessTimes, id); err != nil {
		log.Errorc(c, "UpdateUserDetail: db.Exec( star(%d), heart(%d), lotteryUsed(%d), starDetail(%s), lastStarTime(%d),playTimes(%d), playSuccessTimes(%d) id(%d) error(%v)", star, heart, lotteryUsed, starDetail, lastStarTime, playTimes, playSuccessTimes, id, err)
		return
	}
	return
}

// RawBidUsersMid get users by mid
func (d *Dao) RawBidUsersMid(c context.Context, bid []int64, mid int64) (list map[int64]*bwsmdl.Users, err error) {
	var (
		rows *xsql.Rows
		sqls []string
		args []interface{}
	)
	if len(bid) == 0 {
		return
	}
	for _, id := range bid {
		sqls = append(sqls, "?")
		args = append(args, id)
	}
	args = append(args, mid)
	if rows, err = d.db.Query(c, fmt.Sprintf(_bidUsersMidSQL, strings.Join(sqls, ",")), args...); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
		return
	}
	defer rows.Close()
	list = make(map[int64]*bwsmdl.Users)
	for rows.Next() {
		res := &bwsmdl.Users{}
		if err = rows.Scan(&res.ID, &res.Mid, &res.Key, &res.Ctime, &res.Mtime, &res.Bid); err != nil {
			log.Error("RawUsersMid:row.Scan error(%v)", err)
			return
		}
		list[res.Bid] = res
	}
	err = rows.Err()
	return
}

// RawUsersKeys .
func (d *Dao) RawUsersKeys(c context.Context, bid int64, keys []string) (list map[string]*bwsmdl.Users, err error) {
	var (
		strLen = len(keys)
		rows   *xsql.Rows
		args   = make([]interface{}, 0)
		str    []string
	)
	if strLen == 0 {
		return
	}
	args = append(args, bid)
	for _, v := range keys {
		str = append(str, "?")
		args = append(args, v)
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_usersKeysSQL, strings.Join(str, ",")), args...); err != nil {
		return
	}
	defer rows.Close()
	list = make(map[string]*bwsmdl.Users, strLen)
	for rows.Next() {
		res := &bwsmdl.Users{}
		if err = rows.Scan(&res.ID, &res.Mid, &res.Key, &res.Ctime, &res.Mtime, &res.Bid); err != nil {
			log.Errorc(c, "RawUsersKey:row.Scan error(%v)", err)
			return
		}
		list[res.Key] = res
	}
	err = rows.Err()
	return
}

// InCatchUser .
func (d *Dao) InCatchUser(c context.Context, bid, mid int64, key string) error {
	if _, err := d.db.Exec(c, _inCatchUserSQL, bid, key, mid, bid, key, mid); err != nil {
		log.Errorc(c, "%+v", err)
		return err
	}
	return nil
}

// CatchUser .
func (d *Dao) CatchUser(c context.Context, bid int64, key string) ([]*bwsmdl.CatchUser, error) {
	rows, err := d.db.Query(c, _catchUserSQL, key, bid)
	if err != nil {
		log.Errorc(c, "%+v", err)
		return nil, err
	}
	defer rows.Close()
	var res []*bwsmdl.CatchUser
	for rows.Next() {
		b := &bwsmdl.CatchUser{}
		if err := rows.Scan(&b.Mid, &b.Key, &b.Ctime, &b.Mtime); err != nil {
			log.Errorc(c, "%+v", err)
			return nil, err
		}
		res = append(res, b)
	}
	return res, rows.Err()
}

func (d *Dao) BluetoothUps(c context.Context, bid int64) ([]*bwsmdl.BluetoothUp, error) {
	rows, err := d.db.Query(c, _bluetoothUpsSQL, bid)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	defer rows.Close()
	res := []*bwsmdl.BluetoothUp{}
	for rows.Next() {
		u := &bwsmdl.BluetoothUp{}
		if err := rows.Scan(&u.Id, &u.Mid, &u.Key, &u.Bid, &u.Desc, &u.Ctime, &u.Mtime); err != nil {
			log.Errorc(c, "%+v", err)
			return nil, err
		}
		res = append(res, u)
	}
	return res, rows.Err()
}
