package bws

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"
)

const (
	_fmtPointsKey      = "bws_p_lel_%d"
	_fmtPointsAwardKey = "bws_p_aw_%d"
	_bwsLevelsSQL      = "SELECT `id`,`pid`,`is_delete`,`level`,`points`,`unlock`,`bid`,`ctime`,`mtime` FROM `act_bws_points_level` WHERE `id` in(%s) AND is_delete = 0"
	_bwsAwardsSQL      = "SELECT `id`,`bid`,`is_delete`,`pl_id`,`name`,`icon`,`amount`,`ctime`,`mtime` FROM `act_bws_points_award` WHERE `id` in(%s) AND is_delete = 0"
	_bwsPointLevelSQL  = "SELECT `id` FROM `act_bws_points_level` WHERE `bid` = ? AND pid IN (%s) AND is_delete = 0"
	_bwsPointAwardSQL  = "SELECT `id` FROM `act_bws_points_award` WHERE `pl_id` = ? AND is_delete = 0"
)

func pointLevelKey(id int64) string {
	return fmt.Sprintf(_fmtPointsKey, id)
}

func pointAwardKey(id int64) string {
	return fmt.Sprintf(_fmtPointsAwardKey, id)
}

// CachePointLevels .
func (d *Dao) CachePointLevels(c context.Context, bid int64) ([]int64, error) {
	var (
		key = pointLevelKey(bid)
	)
	return d.CacheSetMembers(c, key)
}

// DelCachePointLevels .
func (d *Dao) DelCachePointLevels(c context.Context, bid int64) error {
	cacheKey := pointLevelKey(bid)
	return d.DelCacheSetMembers(c, cacheKey)
}

// AddCachePointLevels .
func (d *Dao) AddCachePointLevels(c context.Context, bid int64, ids []int64) error {
	var (
		key = pointLevelKey(bid)
	)
	return d.AddCacheSetMembers(c, key, ids)
}

// CachePointsAward .
func (d *Dao) CachePointsAward(c context.Context, plID int64) ([]int64, error) {
	var (
		key = pointAwardKey(plID)
	)
	return d.CacheSetMembers(c, key)
}

// CachePointsAward .
func (d *Dao) AddCachePointsAward(c context.Context, plID int64, ids []int64) error {
	var (
		key = pointAwardKey(plID)
	)
	return d.AddCacheSetMembers(c, key, ids)
}

// CachePointsAward .
func (d *Dao) DelCachePointsAward(c context.Context, plID int64) error {
	var (
		key = pointAwardKey(plID)
	)
	return d.DelCacheSetMembers(c, key)
}

// CacheSetMembers .
func (d *Dao) CacheSetMembers(c context.Context, key string) (ids []int64, err error) {
	var (
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if ids, err = redis.Int64s(conn.Do("SMEMBERS", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CachePointLevels %s error(%v)", key, err)
		}
	}
	return
}

// DelCacheSetMembers .
func (d *Dao) DelCacheSetMembers(c context.Context, cacheKey string) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		log.Error("DelCacheSetMembers conn.Do(DEL) key(%s) error(%v)", cacheKey, err)
	}
	return
}

func (d *Dao) AddCacheSetMembers(c context.Context, key string, ids []int64) (err error) {
	if len(ids) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	args = args.Add(key)
	for _, apID := range ids {
		args = args.Add(apID)
	}
	if _, err = conn.Do("SADD", args...); err != nil {
		log.Error("AddCachePointLevels conn.Send(SADD %s %v) error(%v)", key, args, err)
	}
	return
}

// RawPointsAward .
func (d *Dao) RawPointsAward(c context.Context, bid int64) (ids []int64, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, _bwsPointAwardSQL, bid); err != nil {
		log.Error("RawPointAwards: db.Exec(%d) error(%v)", bid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.PointsAward)
		if err = rows.Scan(&r.ID); err != nil {
			log.Error("RawPointAwards:row.Scan() error(%v)", err)
			return
		}
		ids = append(ids, r.ID)
	}
	err = rows.Err()
	return
}

// RawPointLevels .
func (d *Dao) RawPointLevels(c context.Context, bid int64) (ids []int64, err error) {
	var plids []int64
	if plids, err = d.RawPointsByLock(c, bid, bwsmdl.ChargeType); err != nil {
		log.Error("d.RawPointsByLock(%d) error(%v)", bid, err)
		return
	}
	if len(plids) == 0 {
		return
	}
	return d.PointLevelsByID(c, bid, plids)
}

// PointLevelsByID .
func (d *Dao) PointLevelsByID(c context.Context, bid int64, plids []int64) (ids []int64, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_bwsPointLevelSQL, xstr.JoinInts(plids)), bid); err != nil {
		log.Error("RawPointLevels: db.Exec(%d) error(%v)", bid, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.PointsLevel)
		if err = rows.Scan(&r.ID); err != nil {
			log.Error("RawPointLevels:row.Scan() error(%v)", err)
			return
		}
		ids = append(ids, r.ID)
	}
	err = rows.Err()
	return
}

// RawRechargeLevels .
func (d *Dao) RawRechargeAwards(c context.Context, ids []int64) (rs map[int64]*bwsmdl.PointsAward, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_bwsAwardsSQL, xstr.JoinInts(ids))); err != nil {
		log.Error("RawRechargeAwards: db.Exec(%v) error(%v)", ids, err)
		return
	}
	defer rows.Close()
	rs = make(map[int64]*bwsmdl.PointsAward, len(ids))
	for rows.Next() {
		r := new(bwsmdl.PointsAward)
		if err = rows.Scan(&r.ID, &r.Bid, &r.IsDelete, &r.PlID, &r.Name, &r.Icon, &r.Amount, &r.Ctime, &r.Mtime); err != nil {
			log.Error("RawRechargeAwards:row.Scan() error(%v)", err)
			return
		}
		rs[r.ID] = r
	}
	err = rows.Err()
	return
}

// RawRechargeLevels .
func (d *Dao) RawRechargeLevels(c context.Context, ids []int64) (rs map[int64]*bwsmdl.PointsLevel, err error) {
	var (
		rows *xsql.Rows
	)
	if rows, err = d.db.Query(c, fmt.Sprintf(_bwsLevelsSQL, xstr.JoinInts(ids))); err != nil {
		log.Error("RawRechargeLevels: db.Exec(%v) error(%v)", ids, err)
		return
	}
	defer rows.Close()
	rs = make(map[int64]*bwsmdl.PointsLevel, len(ids))
	for rows.Next() {
		r := new(bwsmdl.PointsLevel)
		if err = rows.Scan(&r.ID, &r.Pid, &r.IsDelete, &r.Level, &r.Points, &r.Unlock, &r.Bid, &r.Ctime, &r.Mtime); err != nil {
			log.Error("RawRechargeLevels:row.Scan() error(%v)", err)
			return
		}
		rs[r.ID] = r
	}
	err = rows.Err()
	return
}
