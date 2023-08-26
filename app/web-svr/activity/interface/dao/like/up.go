package like

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/cache"
	"go-common/library/cache/redis"
	"go-common/library/log"
	xtime "go-common/library/time"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_queryActUpSQL          = "SELECT id,mid,title,statement,aid,state,offline,suffix,attribute,finish_count,stime,etime,ctime,mtime FROM act_up WHERE mid = ? ORDER BY id DESC LIMIT 1"
	_queryActUpByIDSQL      = "SELECT id,mid,title,statement,aid,state,offline,suffix,attribute,finish_count,stime,etime,ctime,mtime FROM act_up WHERE id = ? ORDER BY id DESC LIMIT 1"
	_queryActUpByAidSQL     = "SELECT id,mid,title,statement,aid,state,offline,suffix,attribute,finish_count,stime,etime,ctime,mtime FROM act_up WHERE aid = ? ORDER BY id DESC LIMIT 1"
	_queryActUpUserStateSQL = "SELECT id,sid,mid,bid,round,times,finish,result,ctime,mtime FROM act_up_user_state_%d WHERE sid = ? AND mid = ?"
	_addActUpSQL            = "INSERT INTO `act_up` (`mid`,`title`,`statement`,`stime`,`etime`,`aid`)VALUES(?,?,?,?,?,?)"
	_addUserLogSQL          = "INSERT INTO `act_up_user_log` (`sid`,`mid`,`bid`,`round`)VALUES(?,?,?,?)"
	_upUserStateSQL         = "UPDATE act_up_user_state_%d SET times=?,finish=? WHERE sid=? AND mid=? AND bid=? AND round=?"
	_addUserStateSQL        = "INSERT INTO act_up_user_state_%d(`sid`,`mid`,`bid`,`round`,`times`,`finish`,`result`)VALUES(?,?,?,?,?,?,?)"
	_actUpMidKey            = "act_up_mid_%d"
	_actUpSidKey            = "act_up_sid_%d"
	_actUpAidKey            = "act_up_aid_%d"
	_rankMax                = 100
)

func actUpKey(mid int64) string {
	return fmt.Sprintf(_actUpMidKey, mid)
}

func actUpBySidKey(sid int64) string {
	return fmt.Sprintf(_actUpSidKey, sid)
}

func actUpByAidKey(aid int64) string {
	return fmt.Sprintf(_actUpAidKey, aid)
}

func userStateKey(id, mid int64) string {
	return fmt.Sprintf("act_up_%d_%d", id, mid)
}

func RoundMapKey(round int64) string {
	return fmt.Sprintf("round_%d", round)
}

func actUpRankKey(sid int64) string {
	return fmt.Sprintf("act_up_rank_%d", sid)
}

func buildUserDays(days, ctime int64) float64 {
	return float64(days) + float64(ctime)*0.0000000001
}

func (dao *Dao) RawActUp(c context.Context, mid int64) (res *like.ActUp, err error) {
	row := dao.db.QueryRow(c, _queryActUpSQL, mid)
	res = new(like.ActUp)
	if err = row.Scan(&res.ID, &res.Mid, &res.Title, &res.Statement, &res.Aid, &res.State, &res.Offline, &res.Suffix, &res.Attribute, &res.FinishCount, &res.Stime, &res.Etime, &res.Ctime, &res.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			err = errors.Wrap(err, "RawActUp()")
		}
	}
	return
}

func (dao *Dao) RawActUpBySid(c context.Context, sid int64) (res *like.ActUp, err error) {
	row := dao.db.QueryRow(c, _queryActUpByIDSQL, sid)
	res = new(like.ActUp)
	if err = row.Scan(&res.ID, &res.Mid, &res.Title, &res.Statement, &res.Aid, &res.State, &res.Offline, &res.Suffix, &res.Attribute, &res.FinishCount, &res.Stime, &res.Etime, &res.Ctime, &res.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			err = errors.Wrap(err, "RawActUpBySid()")
		}
	}
	return
}

func (dao *Dao) RawActUpByAid(c context.Context, aid int64) (res *like.ActUp, err error) {
	row := dao.db.QueryRow(c, _queryActUpByAidSQL, aid)
	res = new(like.ActUp)
	if err = row.Scan(&res.ID, &res.Mid, &res.Title, &res.Statement, &res.Aid, &res.State, &res.Offline, &res.Suffix, &res.Attribute, &res.FinishCount, &res.Stime, &res.Etime, &res.Ctime, &res.Mtime); err != nil {
		if err == sql.ErrNoRows {
			err = nil
			res = nil
		} else {
			err = errors.Wrap(err, "RawActUpByAid()")
		}
	}
	return
}

func (dao *Dao) InsertUpAct(c context.Context, title, statement string, stime, etime xtime.Time, aid, mid int64) (err error) {
	if _, err = dao.db.Exec(c, _addActUpSQL, mid, title, statement, stime, etime, aid); err != nil {
		err = errors.Wrap(err, "dao.db.Exec()")
	}
	return
}

func (dao *Dao) CacheActUp(c context.Context, mid int64) (res *like.ActUp, err error) {
	var (
		key  = actUpKey(mid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheActUp(%s) return nil", key)
		} else {
			err = errors.Wrap(err, "conn.Do()")
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		err = errors.Wrap(err, "json.Unmarshal")
	}
	return
}

func (dao *Dao) AddCacheActUp(c context.Context, mid int64, val *like.ActUp) (err error) {
	var (
		key  = actUpKey(mid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(val); err != nil {
		err = errors.Wrap(err, "json.Unmarshal")
		return
	}
	if _, err = conn.Do("SETEX", key, dao.matchExpire, bs); err != nil {
		err = errors.Wrap(err, "conn.Send(SETEX)")
	}
	return
}

func (dao *Dao) CacheActUpBySid(c context.Context, sid int64) (res *like.ActUp, err error) {
	var (
		key  = actUpBySidKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheActUpBySid(%s) return nil", key)
		} else {
			err = errors.Wrap(err, "conn.Do()")
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		err = errors.Wrap(err, "json.Unmarshal")
	}
	return
}

func (dao *Dao) AddCacheActUpBySid(c context.Context, sid int64, val *like.ActUp) (err error) {
	if val == nil {
		return
	}
	var (
		key  = actUpBySidKey(sid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(val); err != nil {
		err = errors.Wrap(err, "json.Unmarshal")
		return
	}
	if _, err = conn.Do("SETEX", key, dao.matchExpire, bs); err != nil {
		err = errors.Wrap(err, "conn.Send(SETEX)")
	}
	return
}

func (dao *Dao) CacheActUpByAid(c context.Context, aid int64) (res *like.ActUp, err error) {
	var (
		key  = actUpByAidKey(aid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheActUpByAid(%s) return nil", key)
		} else {
			err = errors.Wrap(err, "conn.Do()")
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		err = errors.Wrap(err, "json.Unmarshal")
	}
	return
}

func (dao *Dao) AddCacheActUpByAid(c context.Context, aid int64, val *like.ActUp) (err error) {
	if val == nil {
		return
	}
	var (
		key  = actUpByAidKey(aid)
		conn = dao.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(val); err != nil {
		err = errors.Wrap(err, "json.Unmarshal")
		return
	}
	if _, err = conn.Do("SETEX", key, dao.matchExpire, bs); err != nil {
		err = errors.Wrap(err, "conn.Send(SETEX)")
	}
	return
}

func (dao *Dao) RawUpActUserState(c context.Context, act *like.ActUp, mid int64) (res map[string]*like.UpActUserState, err error) {
	rows, err := dao.db.Query(c, fmt.Sprintf(_queryActUpUserStateSQL, act.Suffix), act.ID, mid)
	if err != nil {
		log.Error("d.dao.AllBadge() error(%v)", err)
		return nil, err
	}
	defer rows.Close()
	res = make(map[string]*like.UpActUserState)
	for rows.Next() {
		state := &like.UpActUserState{}
		if err = rows.Scan(&state.ID, &state.Sid, &state.Mid, &state.Bid, &state.Round, &state.Times, &state.Finish, &state.Result, &state.Ctime, &state.Mtime); err != nil {
			err = errors.Wrap(err, "rows.Scan()")
			return
		}
		res[RoundMapKey(state.Round)] = state
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawUpActUserState: rows.Err()")
	}
	return
}

// CacheUpActUserState get up act user cache.
func (dao *Dao) CacheUpActUserState(c context.Context, act *like.ActUp, mid int64) (list map[string]*like.UpActUserState, err error) {
	if act == nil || act.Mid == 0 {
		return
	}
	key := userStateKey(act.ID, mid)
	args := redis.Args{}.Add(key)
	nowTs := time.Now().Unix()
	var upList []string
	if !act.UpIsNoCycle() {
		for i := act.UpRound(nowTs); i >= 0; i-- {
			args = args.Add(RoundMapKey(i))
			upList = append(upList, RoundMapKey(i))
		}
	} else {
		args = args.Add(RoundMapKey(0))
		upList = append(upList, RoundMapKey(0))
	}
	conn := dao.redis.Get(c)
	defer conn.Close()
	var values [][]byte
	if values, err = redis.ByteSlices(conn.Do("HMGET", args...)); err != nil {
		if err != redis.ErrNil {
			log.Error("CacheUpActUserState redis.ByteSlices(%s,%d) error(%v)", key, act.ID, err)
			return
		}
	}
	list = make(map[string]*like.UpActUserState)
	for index, v := range upList {
		if values[index] == nil {
			continue
		}
		state := &like.UpActUserState{}
		if e := json.Unmarshal(values[index], state); e != nil {
			log.Warn("CacheUserTaskState json.Unmarshal(%x) error(%v)", values[index], err)
			continue
		}
		list[v] = state
	}
	return
}

// AddCacheUpActUserState add user state cache.
func (dao *Dao) AddCacheUpActUserState(c context.Context, act *like.ActUp, missData map[string]*like.UpActUserState, mid int64) (err error) {
	var (
		bs []byte
	)
	if len(missData) == 0 {
		return
	}
	conn := dao.redis.Get(c)
	defer conn.Close()
	key := userStateKey(act.ID, mid)
	args := redis.Args{}.Add(key)
	for k, v := range missData {
		if bs, err = json.Marshal(v); err != nil {
			log.Warn("AddCacheUpActUserState json.Marshal(%v) error(%v)", v, err)
			continue
		}
		args = args.Add(k).Add(bs)
	}
	if err = conn.Send("HMSET", args...); err != nil {
		log.Error("AddCacheUpActUserState conn.Send(HMSET, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, dao.matchExpire); err != nil {
		log.Error("AddCacheUpActUserState conn.Send(Expire, %s, %d) error(%v)", key, dao.matchExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheUpActUserState conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheUpActUserState conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// UpActUserState get user state by upact.
func (dao *Dao) UpActUserState(c context.Context, act *like.ActUp, mid int64) (res map[string]*like.UpActUserState, err error) {
	if act == nil || act.ID == 0 || act.Mid == 0 {
		return
	}
	addCache := true
	if res, err = dao.CacheUpActUserState(c, act, mid); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	if res != nil {
		cache.MetricHits.Inc("UpActUserState")
		return
	}
	cache.MetricMisses.Inc("UpActUserState")
	res, err = dao.RawUpActUserState(c, act, mid)
	if err != nil {
		return
	}
	missData := res
	if !addCache {
		return
	}
	dao.cache.Do(c, func(c context.Context) {
		dao.AddCacheUpActUserState(c, act, missData, mid)
	})
	return
}

// SetCacheUpActUserState set one cache user state
func (d *Dao) SetCacheUpActUserState(c context.Context, act *like.ActUp, val *like.UpActUserState, mid, round int64) (err error) {
	var (
		bs []byte
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	key := userStateKey(act.ID, mid)
	args := redis.Args{}.Add(key)
	if bs, err = json.Marshal(val); err != nil {
		log.Warn("SetCacheUpActUserState json.Marshal(%v) error(%v)", val, err)
		return
	}
	args = args.Add(RoundMapKey(round)).Add(bs)
	if err = conn.Send("HSET", args...); err != nil {
		log.Error("SetCacheUpActUserState conn.Send(HSET, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.matchExpire); err != nil {
		log.Error("SetCacheUpActUserState conn.Send(Expire, %s, %d) error(%v)", key, d.matchExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("SetCacheUpActUserState conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("SetCacheUpActUserState conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (dao *Dao) AddUserLog(c context.Context, sid, mid, bid, round int64) (err error) {
	if _, err = dao.db.Exec(c, _addUserLogSQL, sid, mid, bid, round); err != nil {
		log.Error("AddUserLog error dao.db.Exec(%d,%d,%d,%d) error(%v)", sid, mid, bid, round, err)
	}
	return
}

func (dao *Dao) AddUserState(c context.Context, sid, mid, bid, round, finish, times, suffix int64, result string) (err error) {
	if _, err = dao.db.Exec(c, fmt.Sprintf(_addUserStateSQL, suffix), sid, mid, bid, round, times, finish, result); err != nil {
		log.Error("AddUserState error dao.db.Exec(%d,%d,%d,%d) error(%v)", sid, mid, bid, round, err)
	}
	return
}

func (dao *Dao) UpUserState(c context.Context, sid, mid, bid, round, finish, times, suffix int64) (err error) {
	if _, err = dao.db.Exec(c, fmt.Sprintf(_upUserStateSQL, suffix), times, finish, sid, mid, bid, round); err != nil {
		log.Error("AddUserState error dao.db.Exec(%d,%d,%d,%d) error(%v)", sid, mid, bid, round, err)
	}
	return
}

// CacheUpUsersRank score list .
func (dao *Dao) CacheUpUsersRank(c context.Context, sid int64, start, end int) (data []*like.RankUserDays, err error) {
	key := actUpRankKey(sid)
	conn := dao.redis.Get(c)
	defer conn.Close()
	vs, err := redis.Values(conn.Do("ZREVRANGE", key, start, end, "WITHSCORES"))
	if err != nil {
		err = errors.Wrap(err, "CacheUpUsersRank conn.Do(ZREVRANG)")
		return
	}
	if len(vs) == 0 {
		return
	}
	data = make([]*like.RankUserDays, 0, len(vs))
	for len(vs) > 0 {
		rud := new(like.RankUserDays)
		if vs, err = redis.Scan(vs, &rud.Mid, &rud.Days); err != nil {
			err = errors.Wrap(err, "CacheUpUsersRank redis.Scan()")
			return
		}
		data = append(data, rud)
	}
	return
}

// CacheUpUsersRank score list .
func (dao *Dao) CacheUpUserDays(c context.Context, sid, mid int64) (days float64, err error) {
	key := actUpRankKey(sid)
	conn := dao.redis.Get(c)
	defer conn.Close()
	if days, err = redis.Float64(conn.Do("ZSCORE", key, mid)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrap(err, "conn.Do(ZSCORE)")
		}
	}
	return
}

func (dao *Dao) AddUpUsersRank(c context.Context, sid, mid, nowT, days int64) (err error) {
	key := actUpRankKey(sid)
	conn := dao.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("ZADD", key, buildUserDays(days, nowT), mid); err != nil {
		err = errors.Wrapf(err, "conn.Send(ZADD,%s,%f,%d)", key, buildUserDays(int64(days), nowT), mid)
		return
	}
	if err = conn.Send("ZCARD", key); err != nil {
		log.Error("conn.Send(ZCARD %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		err = errors.Wrap(err, "AddCacheUserGrade conn.Flush()")
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			err = errors.Wrap(err, "conn.Receive()")
			return
		}
	}
	total := int64(0)
	if total, err = redis.Int64(conn.Receive()); err != nil {
		log.Error("AddUpUsersRank conn.Receive() error(%v)", err)
	}
	delCount := total - _rankMax
	if delCount > 0 {
		if _, err = conn.Do("ZREMRANGEBYRANK", key, 0, delCount-1); err != nil {
			log.Error("conn.Send(ZREMRANGEBYRANK) key(%s) error(%v)", key, err)
		}
	}
	return
}

// DeleteCacheActUpBySid ...
func (dao *Dao) DeleteCacheActUpBySid(c context.Context, sid int64) (err error) {
	var (
		key  = actUpBySidKey(sid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DeleteCacheActUpBySid conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// DeleteCacheActUpBySid ...
func (dao *Dao) DeleteCacheActUp(c context.Context, mid int64) (err error) {
	var (
		key  = actUpKey(mid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DeleteCacheActUp conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}

// DeleteCacheActUpByAid ...
func (dao *Dao) DeleteCacheActUpByAid(c context.Context, aid int64) (err error) {
	var (
		key  = actUpByAidKey(aid)
		conn = dao.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DeleteCacheActUp conn.Do(DEL, %s) error(%v)", key, err)
	}
	return
}
