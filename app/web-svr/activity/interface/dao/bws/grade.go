package bws

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/cache/redis"
	bwsmdl "go-gateway/app/web-svr/activity/interface/model/bws"

	"github.com/pkg/errors"
)

const (
	_userGradeAddSQL      = "INSERT INTO act_bws_grade(pid,`key`,amount) VALUES(?,?,?) ON DUPLICATE KEY UPDATE amount = ?"
	_usersAchieveGradeSQL = "SELECT `key`,amount,mtime FROM act_bws_grade WHERE pid = ? AND `key` IN (%s) AND state = 0"
	_userGradeInfoSQL     = "SELECT pid,`key`,amount,mtime FROM act_bws_grade where `key` = ? AND state = 0"
	_userGradeSQL         = "SELECT `key`,amount,mtime FROM act_bws_grade WHERE pid = ? AND state = 0"
)

func bwsUserGrade(pid int64) string {
	return fmt.Sprintf("bws_user_grade_%d", pid)
}

func bwsAchieveGrade(pid int64, ukey string) string {
	return fmt.Sprintf("bws_achieve_grade_%d_%s", pid, ukey)
}

func buildUserGrade(num, ctime int64) float64 {
	return float64(num) + float64(_rankMaxTime-ctime)*0.0000000001
}

// CacheUserRank 获取用户成绩.
func (d *Dao) CacheUserGrade(c context.Context, pid, mid int64) (rank int, score float64, err error) {
	cacheKey := bwsUserGrade(pid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("ZREVRANK", cacheKey, mid); err != nil {
		err = errors.Wrap(err, "CacheUserGrade ZREVRANK()")
		return
	}
	if err = conn.Send("ZSCORE", cacheKey, mid); err != nil {
		err = errors.Wrap(err, "CacheUserGrade ZSCORE()")
		return
	}
	if err = conn.Flush(); err != nil {
		return
	}
	if rank, err = redis.Int(conn.Receive()); err != nil {
		rank = bwsmdl.DefaultRank
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrap(err, "CacheUserGrade redis.Int()")
		}
		return
	}
	if score, err = redis.Float64(conn.Receive()); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrap(err, "CacheUserGrade redis.Float64()")
		}
	}
	return
}

func (d *Dao) AddCacheUserGrade(c context.Context, pid int64, midGrade map[int64]*bwsmdl.UserGrade) (err error) {
	if len(midGrade) == 0 {
		return
	}
	cacheKey := bwsUserGrade(pid)
	conn := d.redis.Get(c)
	defer conn.Close()
	for mid, val := range midGrade {
		if err = conn.Send("ZADD", cacheKey, buildUserGrade(val.Amount, int64(val.Mtime)), mid); err != nil {
			err = errors.Wrapf(err, "conn.Send(ZADD,%s,%f,%d)", cacheKey, buildUserGrade(val.Amount, int64(val.Mtime)), mid)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		err = errors.Wrap(err, "AddCacheUserGrade conn.Flush()")
		return
	}
	for i := 0; i < len(midGrade); i++ {
		if _, err = conn.Receive(); err != nil {
			err = errors.Wrap(err, "conn.Receive()")
		}
	}
	return
}

// AddUserGrade .
func (d *Dao) AddUserGrade(c context.Context, pid, amount int64, key string) (lastID int64, err error) {
	res, err := d.db.Exec(c, _userGradeAddSQL, pid, key, amount, amount)
	if err != nil {
		err = errors.Wrapf(err, "AddUserGrade d.db.Exec(%d,%s,%d)", pid, key, amount)
		return
	}
	return res.LastInsertId()
}

// DelUserGrade 排行版数据删除.
func (d *Dao) DelUserGrade(c context.Context, pid int64) (err error) {
	cacheKey := bwsUserGrade(pid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", cacheKey); err != nil {
		err = errors.Wrap(err, "DelUserGrade conn.Do(DEL)")
	}
	return
}

// CacheUsersRank score list .
func (d *Dao) CacheUsersRank(c context.Context, pid int64, start, end int) (data []*bwsmdl.RankUserGrade, err error) {
	key := bwsUserGrade(pid)
	conn := d.redis.Get(c)
	defer conn.Close()
	vs, err := redis.Values(conn.Do("ZREVRANGE", key, start, end, "WITHSCORES"))
	if err != nil {
		err = errors.Wrap(err, "CacheUsersRank conn.Do(ZREVRANG)")
		return
	}
	if len(vs) == 0 {
		return
	}
	data = make([]*bwsmdl.RankUserGrade, 0, len(vs))
	for len(vs) > 0 {
		rug := new(bwsmdl.RankUserGrade)
		if vs, err = redis.Scan(vs, &rug.Mid, &rug.Amount); err != nil {
			err = errors.Wrap(err, "CacheUsersRank redis.Scan()")
			return
		}
		data = append(data, rug)
	}
	return
}

// CacheAchievesGrade 获取用户成就点数.
func (d *Dao) CacheAchievesGrade(c context.Context, pid int64, ukey []string) (list map[string]int64, err error) {
	if len(ukey) == 0 {
		return
	}
	var ss []int64
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for _, v := range ukey {
		args = args.Add(bwsAchieveGrade(pid, v))
	}
	if ss, err = redis.Int64s(conn.Do("MGET", args...)); err != nil {
		err = errors.Wrap(err, "CacheAchievesGrade conn.Do(MGET)")
		return
	}
	list = make(map[string]int64, len(ukey))
	for key, val := range ss {
		if val == 0 {
			continue
		}
		list[ukey[key]] = val
	}
	return
}

// AddCacheAchievesGrade 设置用户考分(回源时使用).
func (d *Dao) AddCacheAchievesGrade(c context.Context, pid int64, miss map[string]int64) (err error) {
	if len(miss) == 0 {
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	args := redis.Args{}
	for k, v := range miss {
		args = args.Add(bwsAchieveGrade(pid, k)).Add(v)
	}
	if _, err = redis.String(conn.Do("MSET", args...)); err != nil {
		err = errors.Wrap(err, "AddCacheAchievesGrade conn.Do(MSET)")
	}
	return
}

// AchievesGrade get data from cache if miss will call source method, then add to cache.
func (d *Dao) AchievesGrade(c context.Context, pid int64, ukey []string) (res map[string]int64, err error) {
	if len(ukey) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheAchievesGrade(c, pid, ukey); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []string
	for _, key := range ukey {
		if _, ok := res[key]; !ok {
			miss = append(miss, key)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[string]int64
	missData, err = d.RawUsersAchievesGrade(c, pid, miss)
	if res == nil {
		res = make(map[string]int64, len(ukey))
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
	d.AddCacheAchievesGrade(c, pid, missData)
	return
}

// RawUsersAchievesGrade get user achievements from db.
func (d *Dao) RawUsersAchievesGrade(c context.Context, pid int64, ukeys []string) (rs map[string]int64, err error) {
	ukesLen := len(ukeys)
	if ukesLen == 0 {
		return
	}
	args := make([]interface{}, 0)
	args = append(args, pid)
	var str []string
	for _, v := range ukeys {
		str = append(str, "?")
		args = append(args, v)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_usersAchieveGradeSQL, strings.Join(str, ",")), args...)
	if err != nil {
		err = errors.Wrap(err, "RawUsersAchievesGrade db.Query()")
		return
	}
	defer rows.Close()
	rs = make(map[string]int64)
	for rows.Next() {
		r := new(bwsmdl.UserGrade)
		if err = rows.Scan(&r.Key, &r.Amount, &r.Mtime); err != nil {
			err = errors.Wrap(err, "RawUsersAchievesGrade rows.Scan()")
			return
		}
		rs[r.Key] = r.Amount
	}
	err = rows.Err()
	return
}

func (d *Dao) RawGradeInfo(c context.Context, key string) (res []*bwsmdl.UserGrade, err error) {
	rows, err := d.db.Query(c, _userGradeInfoSQL, key)
	if err != nil {
		err = errors.Wrap(err, "RawGradeInfo db.Query()")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.UserGrade)
		if err = rows.Scan(&r.Pid, &r.Key, &r.Amount, &r.Mtime); err != nil {
			err = errors.Wrap(err, "RawGradeInfo rows.Scan()")
			return
		}
		res = append(res, r)
	}
	err = rows.Err()
	return
}

// RawUserGrade
func (d *Dao) RawUsersGrade(c context.Context, pid int64) (res []*bwsmdl.UserGrade, err error) {
	rows, err := d.db.Query(c, _userGradeSQL, pid)
	if err != nil {
		err = errors.Wrap(err, "RawUserGrade db.Query")
		return
	}
	defer rows.Close()
	for rows.Next() {
		r := new(bwsmdl.UserGrade)
		if err = rows.Scan(&r.Key, &r.Amount, &r.Mtime); err != nil {
			err = errors.Wrap(err, "RawUserGrade rows.Scan")
			return
		}
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawUserGrade rows.Err")
	}
	return
}
