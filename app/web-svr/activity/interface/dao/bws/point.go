package bws

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

const (
	_userLockPointKeyFmt    = "bws_u_lk_p_%d_%d_%s"
	_userLockPointDayKeyFmt = "bws_u_lk_p_%d_%d_%s_%s"
	_userPointKeyFmt        = "bws_u_p_%d_%s"
	_unlockKeyFmt           = "ulk_p_d_%d"
	_userPointsKey          = "us_p_%d_%s"
	_userIncrPointsKey      = "us_i_p_%d_%s"
	_pointsSQL              = "SELECT `id` FROM act_bws_points WHERE del = 0 AND bid = ? ORDER BY ID"
	_pointsLockSQL          = "SELECT `id` FROM act_bws_points WHERE del = 0 AND bid = ? AND lock_type = ? ORDER BY ID"
	_bwsPointsSQL           = "SELECT id,`name`,icon,fid,image,unlocked,lock_type,dic,rule,bid,lose_unlocked,other_ip,ower,ctime,mtime FROM act_bws_points WHERE id in (%s) AND  del = 0"
	_userPointSQL           = "SELECT id,pid,points,ctime FROM act_bws_user_points WHERE bid = ? AND `key` = ? AND del = 0"
	_userSumPointsSQL       = "SELECT SUM(points) FROM act_bws_user_points WHERE bid = ? AND `key` = ? AND del = 0"
	_userLockPointSQL       = "SELECT id,pid,points,ctime FROM act_bws_user_points WHERE bid = ? AND `key` = ? AND del = 0 AND lock_type = ?"
	_userLockPointDaySQL    = "SELECT id,pid,points,ctime FROM act_bws_user_points WHERE bid = ? AND `key` = ? AND del = 0 AND lock_type = ? AND mtime >= ? AND mtime <= ?"
	_batchUserLockPointSQL  = "SELECT id,pid,points,ctime,lock_type FROM act_bws_user_points WHERE bid = ? AND `key` = ? AND del = 0 AND lock_type in (%s)"
	_userPointAddSQL        = "INSERT INTO act_bws_user_points(`bid`,`pid`,`points`,`key`,`lock_type`) VALUES(?,?,?,?,?)"
	_countUnlockSQL         = "SELECT count(1) FROM `act_bws_user_points` WHERE `pid` = ? and `del` = 0"
	_rechargePointSQL       = "UPDATE `act_bws_points` SET `unlocked` = `unlocked` - ? WHERE `id` = ? AND `lock_type` = ?"
)

func keyUserLockPoint(bid, lockType int64, key string) string {
	return fmt.Sprintf(_userLockPointKeyFmt, bid, lockType, key)
}

func keyUserLockPointDay(bid, lockType int64, key, day string) string {
	return fmt.Sprintf(_userLockPointDayKeyFmt, bid, lockType, key, day)
}

func userPointKey(bid int64, key string) string {
	return fmt.Sprintf(_userPointsKey, bid, key)
}

func userIncrPointKey(bid int64, key string) string {
	return fmt.Sprintf(_userIncrPointsKey, bid, key)
}

func keyUserPoint(bid int64, key string) string {
	return fmt.Sprintf(_userPointKeyFmt, bid, key)
}

func keyUnlock(pid int64) string {
	return fmt.Sprintf(_unlockKeyFmt, pid)
}

// RawBwsPoints .
func (d *Dao) RawBwsPoints(c context.Context, ids []int64) (rs map[int64]*bwsmdl.Point, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_bwsPointsSQL, xstr.JoinInts(ids))); err != nil {
		log.Error("RawBwsPoints: db.Exec(%v) error(%v)", ids, err)
		return
	}
	defer rows.Close()
	rs = make(map[int64]*bwsmdl.Point, len(ids))
	for rows.Next() {
		r := new(bwsmdl.Point)
		if err = rows.Scan(&r.ID, &r.Name, &r.Icon, &r.Fid, &r.Image, &r.Unlocked, &r.LockType, &r.Dic, &r.Rule, &r.Bid, &r.LoseUnlocked, &r.OtherIp, &r.Ower, &r.Ctime, &r.Mtime); err != nil {
			log.Error("RawBwsPoints:row.Scan() error(%v)", err)
			return
		}
		rs[r.ID] = r
	}
	err = rows.Err()
	return
}

// _pointsLockSQL
func (d *Dao) RawPointsByLock(c context.Context, bid int64, lockType int) (res []int64, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _pointsLockSQL, bid, lockType); err != nil {
		log.Error("RawPointsByLock: db.Exec(%d) error(%v)", bid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.Point)
		if err = rows.Scan(&r.ID); err != nil {
			log.Error("RawPointsByLock:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r.ID)
	}
	err = rows.Err()
	return
}

// RawPoints points list
func (d *Dao) RawPoints(c context.Context, bid int64) (res []int64, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _pointsSQL, bid); err != nil {
		log.Error("RawPoints: db.Exec(%d) error(%v)", bid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.Point)
		if err = rows.Scan(&r.ID); err != nil {
			log.Error("RawPoints:row.Scan() error(%v)", err)
			return
		}
		res = append(res, r.ID)
	}
	err = rows.Err()
	return
}

// RawBatchUserLockPoints .
func (d *Dao) RawBatchUserLockPoints(c context.Context, bid int64, lockType []int64, key string) (rs map[int64][]*bwsmdl.UserPoint, err error) {
	var (
		rows *xsql.Rows
		sqls []string
		args []interface{}
	)
	if len(lockType) == 0 {
		return
	}
	args = append(args, bid, key)
	for _, lt := range lockType {
		sqls = append(sqls, "?")
		args = append(args, lt)
	}
	if rows, err = d.db.Query(c, fmt.Sprintf(_batchUserLockPointSQL, strings.Join(sqls, ",")), args...); err != nil {
		log.Error("RawBatchUserLockPoints:db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	rs = make(map[int64][]*bwsmdl.UserPoint)
	for rows.Next() {
		r := &bwsmdl.UserPointDetail{UserPoint: &bwsmdl.UserPoint{}}
		if err = rows.Scan(&r.ID, &r.Pid, &r.Points, &r.Ctime, &r.LockType); err != nil {
			log.Error("RawBatchUserLockPoints:row.Scan() error(%v)", err)
			return
		}
		rs[r.LockType] = append(rs[r.LockType], &bwsmdl.UserPoint{ID: r.ID, Pid: r.Pid, Points: r.Points, Ctime: r.Ctime})
	}
	err = rows.Err()
	return
}

// RawUserLockPoints .
func (d *Dao) RawUserLockPoints(c context.Context, bid, lockType int64, key string) (rs []*bwsmdl.UserPoint, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _userLockPointSQL, bid, key, lockType); err != nil {
		log.Error("RawUserLockPoints:db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.UserPoint)
		if err = rows.Scan(&r.ID, &r.Pid, &r.Points, &r.Ctime); err != nil {
			log.Error("RawUserLockPoints:row.Scan() error(%v)", err)
			return
		}
		rs = append(rs, r)
	}
	err = rows.Err()
	return
}

func (d *Dao) RawUserLockPointsDay(ctx context.Context, bid, lockType int64, key, day string) ([]*bwsmdl.UserPoint, error) {
	layout := "20060102 15:04:05"
	mtimeFrom, err := time.ParseInLocation(layout, day+" 00:00:00", time.Local)
	if err != nil {
		return nil, err
	}
	mtimeTo, err := time.ParseInLocation(layout, day+" 23:59:59", time.Local)
	if err != nil {
		return nil, err
	}
	rows, err := d.db.Query(ctx, _userLockPointDaySQL, bid, key, lockType, mtimeFrom, mtimeTo)
	if err != nil {
		log.Error("RawUserLockPointsDay:db.Query error(%v)", err)
		return nil, err
	}
	defer rows.Close()
	var rs []*bwsmdl.UserPoint
	for rows.Next() {
		r := new(bwsmdl.UserPoint)
		if err = rows.Scan(&r.ID, &r.Pid, &r.Points, &r.Ctime); err != nil {
			log.Error("RawUserLockPointsDay:row.Scan() error(%v)", err)
			return nil, err
		}
		rs = append(rs, r)
	}
	if err = rows.Err(); err != nil {
		log.Error("RawUserLockPointsDay:row.Err() error(%v)", err)
		return nil, err
	}
	return rs, nil
}

// RawUserHp .
func (d *Dao) RawUserHp(c context.Context, bid int64, key string) (rs int64, err error) {
	row := d.db.QueryRow(c, _userSumPointsSQL, bid, key)
	var countNull sql.NullInt64
	if err = row.Scan(&countNull); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
	}
	rs = countNull.Int64
	return
}

// RawUserPoints .
func (d *Dao) RawUserPoints(c context.Context, bid int64, key string) (rs []*bwsmdl.UserPoint, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _userPointSQL, bid, key); err != nil {
		log.Error("RawUserPoints:db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.UserPoint)
		if err = rows.Scan(&r.ID, &r.Pid, &r.Points, &r.Ctime); err != nil {
			log.Error("RawUserPoints:row.Scan() error(%v)", err)
			return
		}
		rs = append(rs, r)
	}
	err = rows.Err()
	return
}

// RawCountUnlock .
func (d *Dao) RawCountUnlock(c context.Context, pid int64) (total int64, err error) {
	row := d.db.QueryRow(c, _countUnlockSQL, pid)
	if err = row.Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			log.Error("RawCountUnlock row.Scan() error(%v)", err)
		}
	}
	return
}

// AddUserPoint .
func (d *Dao) AddUserPoint(c context.Context, bid, pid, lockType, points int64, key string) (lastID int64, err error) {
	var res sql.Result
	if res, err = d.db.Exec(c, _userPointAddSQL, bid, pid, points, key, lockType); err != nil {
		log.Error("AddUserPoint error d.db.Exec(%d,%d,%d,%s) error(%v)", bid, pid, points, key, err)
		return
	}
	return res.LastInsertId()
}

// RechargePoint lockType only for rechargeType.
func (d *Dao) RechargePoint(c context.Context, pid, lockType, points int64) (err error) {
	if _, err = d.db.Exec(c, _rechargePointSQL, points, pid, lockType); err != nil {
		log.Error("RechargePoint error d.db.Exec(%d,%d) error(%v)", pid, points, err)
	}
	return
}

// CacheBatchUserLockPoints .
func (d *Dao) CacheBatchUserLockPoints(c context.Context, bid int64, lockType []int64, key string) (res map[int64][]*bwsmdl.UserPoint, err error) {
	if len(lockType) == 0 {
		return
	}
	var (
		conn = d.redis.Get(c)
		max  = len(lockType)
	)
	defer conn.Close()
	for _, v := range lockType {
		cacheKey := keyUserLockPoint(bid, v, key)
		if err = conn.Send("ZRANGE", cacheKey, 0, -1); err != nil {
			log.Error("CacheBatchUserLockPoints conn.Send(ZRANGE, %s) error(%v)", cacheKey, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("CacheBatchUserLockPoints conn.Flush() error(%v)", err)
		return
	}
	res = make(map[int64][]*bwsmdl.UserPoint, max)
	for i := 0; i < max; i++ {
		var (
			values []interface{}
			tep    []*bwsmdl.UserPoint
		)
		if values, err = redis.Values(conn.Receive()); err != nil {
			log.Error("CacheUserLockPoints conn.Receive(ZRANGE, %d) error(%v)", i, err)
			return
		}
		if len(values) == 0 {
			continue
		}
		for len(values) > 0 {
			var bs []byte
			if values, err = redis.Scan(values, &bs); err != nil {
				log.Error("CacheUserLockPoints redis.Scan(%v) error(%v)", values, err)
				err = nil
				continue
			}
			item := new(bwsmdl.UserPoint)
			if err = json.Unmarshal(bs, &item); err != nil {
				log.Error("CacheUserLockPoints json.Unmarshal error(%v)", err)
				err = nil
				continue
			}
			tep = append(tep, item)
		}
		if len(tep) > 0 {
			res[lockType[i]] = tep
		}
	}
	return
}

// CacheUserLockPoints .
func (d *Dao) CacheUserLockPoints(c context.Context, bid, lockType int64, key string) (res []*bwsmdl.UserPoint, err error) {
	var (
		values   []interface{}
		cacheKey = keyUserLockPoint(bid, lockType, key)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if values, err = redis.Values(conn.Do("ZRANGE", cacheKey, 0, -1)); err != nil {
		log.Error("CacheUserLockPoints conn.Do(ZRANGE, %s) error(%v)", cacheKey, err)
		return
	} else if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		var bs []byte
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("CacheUserLockPoints redis.Scan(%v) error(%v)", values, err)
			return
		}
		item := new(bwsmdl.UserPoint)
		if err = json.Unmarshal(bs, &item); err != nil {
			log.Error("CacheUserLockPoints conn.Do(ZRANGE, %s) error(%v)", cacheKey, err)
			continue
		}
		res = append(res, item)
	}
	return
}

// CacheUserLockPoints .
func (d *Dao) CacheUserLockPointsDay(ctx context.Context, bid, lockType int64, key, day string) (res []*bwsmdl.UserPoint, err error) {
	var (
		values   []interface{}
		cacheKey = keyUserLockPointDay(bid, lockType, key, day)
		conn     = d.redis.Get(ctx)
	)
	defer conn.Close()
	if values, err = redis.Values(conn.Do("ZRANGE", cacheKey, 0, -1)); err != nil {
		log.Error("CacheUserLockPoints conn.Do(ZRANGE, %s) error(%v)", cacheKey, err)
		return
	} else if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		var bs []byte
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("CacheUserLockPoints redis.Scan(%v) error(%v)", values, err)
			return
		}
		item := new(bwsmdl.UserPoint)
		if err = json.Unmarshal(bs, &item); err != nil {
			log.Error("CacheUserLockPoints conn.Do(ZRANGE, %s) error(%v)", cacheKey, err)
			continue
		}
		res = append(res, item)
	}
	return
}

// DelCacheUserLockPoints .
func (d *Dao) DelCacheUserLockPoints(c context.Context, bid, lockType int64, key string) (err error) {
	var (
		cacheKey = keyUserLockPoint(bid, lockType, key)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		log.Error("DelCacheUserLockPoints conn.Do(DEL) key(%s) error(%v)", cacheKey, err)
	}
	return
}

func (d *Dao) DelCacheUserLockPointsDay(c context.Context, bid, lockType int64, key, day string) (err error) {
	var (
		cacheKey = keyUserLockPointDay(bid, lockType, key, day)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		log.Error("DelCacheUserLockPoints conn.Do(DEL) key(%s) error(%v)", cacheKey, err)
	}
	return
}

// AddCacheBatchUserLockPoints .
func (d *Dao) AddCacheBatchUserLockPoints(c context.Context, bid int64, userPoints map[int64][]*bwsmdl.UserPoint, key string) (err error) {
	var (
		bs  []byte
		max int
	)
	if len(userPoints) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	for lockType, data := range userPoints {
		if len(data) == 0 {
			continue
		}
		cacheKey := keyUserLockPoint(bid, lockType, key)
		args := redis.Args{}.Add(cacheKey)
		for _, v := range data {
			if bs, err = json.Marshal(v); err != nil {
				log.Error("AddCacheBatchUserLockPoints json.Marshal() error(%v)", err)
				return
			}
			args = args.Add(v.ID).Add(bs)
		}
		if err = conn.Send("ZADD", args...); err != nil {
			log.Error("AddCacheBatchUserLockPoints conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
			return
		}
		max++
		if err = conn.Send("EXPIRE", cacheKey, d.userPointExpire); err != nil {
			log.Error("AddCacheBatchUserLockPoints conn.Send(Expire, %s) error(%v)", cacheKey, err)
			return
		}
		max++
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheBatchUserLockPoints conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < max; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheBatchUserLockPoints conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// AddCacheUserLockPoints .
func (d *Dao) AddCacheUserLockPoints(c context.Context, bid int64, data []*bwsmdl.UserPoint, lockType int64, key string) (err error) {
	var bs []byte
	if len(data) == 0 {
		return
	}
	cacheKey := keyUserLockPoint(bid, lockType, key)
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(cacheKey)
	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheUserLockPoints json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.ID).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddCacheUserLockPoints conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
		return
	}
	if err = conn.Send("EXPIRE", cacheKey, d.userPointExpire); err != nil {
		log.Error("AddCacheUserLockPoints conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheUserLockPoints conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheUserLockPoints conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) AddCacheUserLockPointsDay(ctx context.Context, bid int64, data []*bwsmdl.UserPoint, lockType int64, key, day string) (err error) {
	var bs []byte
	if len(data) == 0 {
		return
	}
	cacheKey := keyUserLockPointDay(bid, lockType, key, day)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	args := redis.Args{}.Add(cacheKey)
	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheUserLockPoints json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.ID).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddCacheUserLockPoints conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
		return
	}
	if err = conn.Send("EXPIRE", cacheKey, d.userPointExpire); err != nil {
		log.Error("AddCacheUserLockPoints conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheUserLockPoints conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheUserLockPoints conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// CacheUserPoints .
func (d *Dao) CacheUserPoints(c context.Context, bid int64, key string) (res []*bwsmdl.UserPoint, err error) {
	var (
		values   []interface{}
		cacheKey = keyUserPoint(bid, key)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if values, err = redis.Values(conn.Do("ZRANGE", cacheKey, 0, -1)); err != nil {
		log.Error("CacheUserAchieves conn.Do(ZRANGE, %s) error(%v)", cacheKey, err)
		return
	} else if len(values) == 0 {
		return
	}
	for len(values) > 0 {
		var bs []byte
		if values, err = redis.Scan(values, &bs); err != nil {
			log.Error("CacheUserAchieves redis.Scan(%v) error(%v)", values, err)
			return
		}
		item := new(bwsmdl.UserPoint)
		if err = json.Unmarshal(bs, &item); err != nil {
			log.Error("CacheUserAchieves conn.Do(ZRANGE, %s) error(%v)", cacheKey, err)
			continue
		}
		res = append(res, item)
	}
	return
}

// AddCacheUserPoints .
func (d *Dao) AddCacheUserPoints(c context.Context, bid int64, data []*bwsmdl.UserPoint, key string) (err error) {
	var bs []byte
	if len(data) == 0 {
		return
	}
	cacheKey := keyUserPoint(bid, key)
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}.Add(cacheKey)
	for _, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheUserPoints json.Marshal() error(%v)", err)
			return
		}
		args = args.Add(v.ID).Add(bs)
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AddCacheUserPoints conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
		return
	}
	if err = conn.Send("EXPIRE", cacheKey, d.userPointExpire); err != nil {
		log.Error("AddCacheUserPoints conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheUserPoints conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheUserPoints conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// AppendUserLockPointsCache .
func (d *Dao) AppendUserLockPointsCache(c context.Context, bid, lockType int64, key string, point *bwsmdl.UserPoint) (err error) {
	var (
		bs       []byte
		ok       bool
		cacheKey = keyUserLockPoint(bid, lockType, key)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if ok, err = redis.Bool(conn.Do("EXPIRE", cacheKey, d.userPointExpire)); err != nil || !ok {
		log.Error("AppendUserLockPointsCache conn.Do(EXPIRE %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(cacheKey)
	if bs, err = json.Marshal(point); err != nil {
		log.Error("AppendUserLockPointsCache json.Marshal() error(%v)", err)
		return
	}
	args = args.Add(point.ID).Add(bs)
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AppendUserLockPointsCache conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
		return
	}
	if err = conn.Send("EXPIRE", cacheKey, d.userPointExpire); err != nil {
		log.Error("AppendUserLockPointsCache conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AppendUserLockPointsCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AppendUserLockPointsCache conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// AppendUserLockPointsCache .
func (d *Dao) AppendUserLockPointsDayCache(c context.Context, bid, lockType int64, key, day string, point *bwsmdl.UserPoint) (err error) {
	var (
		bs       []byte
		ok       bool
		cacheKey = keyUserLockPointDay(bid, lockType, key, day)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if ok, err = redis.Bool(conn.Do("EXPIRE", cacheKey, d.userPointExpire)); err != nil || !ok {
		log.Error("AppendUserLockPointsCache conn.Do(EXPIRE %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(cacheKey)
	if bs, err = json.Marshal(point); err != nil {
		log.Error("AppendUserLockPointsCache json.Marshal() error(%v)", err)
		return
	}
	args = args.Add(point.ID).Add(bs)
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AppendUserLockPointsCache conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
		return
	}
	if err = conn.Send("EXPIRE", cacheKey, d.userPointExpire); err != nil {
		log.Error("AppendUserLockPointsCache conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AppendUserLockPointsCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AppendUserLockPointsCache conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// AppendUserPointsCache .
func (d *Dao) AppendUserPointsCache(c context.Context, bid int64, key string, point *bwsmdl.UserPoint) (err error) {
	var (
		bs       []byte
		ok       bool
		cacheKey = keyUserPoint(bid, key)
		conn     = d.redis.Get(c)
	)
	defer conn.Close()
	if ok, err = redis.Bool(conn.Do("EXPIRE", cacheKey, d.userPointExpire)); err != nil || !ok {
		log.Error("AppendUserPointsCache conn.Do(EXPIRE %s) error(%v)", key, err)
		return
	}
	args := redis.Args{}.Add(cacheKey)
	if bs, err = json.Marshal(point); err != nil {
		log.Error("AppendUserPointsCache json.Marshal() error(%v)", err)
		return
	}
	args = args.Add(point.ID).Add(bs)
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("AppendUserPointsCache conn.Send(ZADD, %s, %v) error(%v)", cacheKey, args, err)
		return
	}
	if err = conn.Send("EXPIRE", cacheKey, d.userPointExpire); err != nil {
		log.Error("AppendUserPointsCache conn.Send(Expire, %s) error(%v)", cacheKey, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AppendUserPointsCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AppendUserPointsCache conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// DelCacheUserPoints .
func (d *Dao) DelCacheUserPoints(c context.Context, bid int64, key string) (err error) {
	cacheKey := keyUserPoint(bid, key)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		log.Error("DelCacheUserPoints conn.Do(DEL) key(%s) error(%v)", cacheKey, err)
	}
	return
}

// IncrUnlock .
func (d *Dao) IncrUnlock(c context.Context, pid int64, num int) (err error) {
	cacheKey := keyUnlock(pid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("INCRBY", cacheKey, num); err != nil {
		log.Error("IncrUnlock conn.Do(INCRBY) key(%s) error(%v)", cacheKey, err)
	}
	return
}

// AddUnlock .
func (d *Dao) AddUnlock(c context.Context, pid int64, num int64) (err error) {
	cacheKey := keyUnlock(pid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", cacheKey, num); err != nil {
		log.Error("AddUnlock conn.Do(SET) key(%s) error(%v)", cacheKey, err)
	}
	return
}

// CacheUnlock .
func (d *Dao) CacheUnlock(c context.Context, pid int64) (total int64, err error) {
	cacheKey := keyUnlock(pid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if total, err = redis.Int64(conn.Do("GET", cacheKey)); err != nil {
		if err == redis.ErrNil {
			total = -1
			err = nil
		} else {
			log.Error("CacheUnlock conn.Do(GET) key(%s) error(%v)", cacheKey, err)
		}
	}
	return
}

// CacheUserHp get data from mc
func (d *Dao) CacheUserHp(c context.Context, id int64, ukey string) (res int64, err error) {
	key := userPointKey(id, ukey)
	conn := d.redis.Get(c)
	defer conn.Close()
	if res, err = redis.Int64(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheUserHp conn.Do(GET) key(%s) error(%v)", key, err)
		}
	}
	return
}

// UserHp get data from cache if miss will call source method, then add to cache.
func (d *Dao) UserHp(c context.Context, bid int64, ukey string) (res int64, err error) {
	addCache := true
	res, err = d.CacheUserHp(c, bid, ukey)
	if err != nil {
		addCache = false
		err = nil
	}
	if res != 0 {
		return
	}
	res, err = d.RawUserHp(c, bid, ukey)
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	d.AddCacheUserHp(c, bid, res, ukey)
	return
}

// DelCacheUserHp get data from redis .
func (d *Dao) DelCacheUserHp(c context.Context, id int64, ukey string) (err error) {
	key := userPointKey(id, ukey)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheUserHp conn.Do(DEL) key(%s) error(%v)", key, err)
	}
	return
}

// AddCacheUserHp Set data to mc
func (d *Dao) AddCacheUserHp(c context.Context, id int64, val int64, ukey string) (err error) {
	key := userPointKey(id, ukey)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, val); err != nil {
		log.Error("AddCacheUserHp conn.Do(SET) key(%s) error(%v)", key, err)
	}
	return
}

// IncrUserHp Set data to mc
func (d *Dao) IncrUserHp(c context.Context, id int64, val int64, ukey string) (err error) {
	cacheKey := userPointKey(id, ukey)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("INCRBY", cacheKey, val); err != nil {
		log.Error("IncrUserHp conn.Do(INCRBY) key(%s) error(%v)", cacheKey, err)
	}
	return
}

// IncrUserPoints incr user points
func (d *Dao) IncrUserPoints(c context.Context, id int64, val int64, ukey string) (point int64, err error) {
	cacheKey := userIncrPointKey(id, ukey)
	conn := d.redis.Get(c)
	defer conn.Close()
	if point, err = redis.Int64(conn.Do("INCRBY", cacheKey, val)); err != nil {
		log.Error("IncrUserPoints conn.Do(INCRBY) key(%s) error(%v)", cacheKey, err)
	}
	return
}

// PointsByBid .
func (d *Dao) PointsByBid(c context.Context, bid int64) (list *bwsmdl.Points, err error) {
	var (
		ids []int64
		pt  map[int64]*bwsmdl.Point
	)
	if ids, err = d.Points(c, bid); err != nil {
		log.Error(" d.Points(%d) error(%v)", bid, err)
		return
	}
	list = &bwsmdl.Points{}
	if len(ids) == 0 {
		return
	}
	if pt, err = d.BwsPoints(c, ids); err != nil {
		log.Error(" d.BwsPoints(%v) error(%v)", ids, err)
		return
	}
	for _, v := range ids {
		if _, ok := pt[v]; ok {
			list.Points = append(list.Points, pt[v])
		}
	}
	return
}

// BatchUserLockPoints get data from cache if miss will call source method, then add to cache.
func (d *Dao) BatchUserLockPoints(c context.Context, bid int64, lockType []int64, key string) (res map[int64][]*bwsmdl.UserPoint, err error) {
	if len(lockType) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheBatchUserLockPoints(c, bid, lockType, key); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range lockType {
		if (res == nil) || (len(res[key]) == 0) {
			miss = append(miss, key)
		}
	}
	cache.MetricHits.Add(float64(len(lockType)-len(miss)), "bts:BatchUserLockPoints")
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64][]*bwsmdl.UserPoint
	cache.MetricMisses.Add(float64(len(miss)), "bts:BatchUserLockPoints")
	missData, err = d.RawBatchUserLockPoints(c, bid, miss, key)
	if res == nil {
		res = make(map[int64][]*bwsmdl.UserPoint, len(lockType))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	d.AddCacheBatchUserLockPoints(c, bid, missData, key)
	return
}
