package like

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	xsql "go-common/library/database/sql"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/component"
	l "go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

// like_action sql and like state
const (
	_likeActInfosSQL       = "SELECT id,lid FROM like_action WHERE lid in (%s) and mid = ?"
	_likeActBatchAddSQL    = "INSERT INTO like_action(lid,mid,action,sid,ipv6,ctime,mtime) VALUES %s"
	_storyActSumSQL        = "SELECT sum(action) likes FROM like_action WHERE sid = ? AND mid = ? AND ctime >= ? AND ctime <= ?"
	_upWholeActionSQL      = "SELECT id,lid,mid  likes FROM like_action WHERE sid = ? AND mid = ?"
	_upWholeUserSQL        = "SELECT id,lid,mid  likes FROM like_action WHERE sid = ? AND lid = ?"
	_storyEachActSumSQL    = "SELECT sum(action) likes FROM like_action WHERE sid = ? AND mid = ? AND lid = ? AND ctime >= ? AND ctime <= ?"
	_likeActLikesSQL       = "SELECT lid,action,ctime FROM like_action WHERE sid = ? AND mid = ? ORDER BY id ASC"
	_likeActAddSQL         = "INSERT INTO like_action(lid,mid,action,sid,ipv6,ctime,mtime,extra_action) VALUES(?,?,?,?,?,?,?,?)"
	_likeExtraTimesSQL     = "SELECT id,sid,mid,num,ctime FROM like_extra_times WHERE sid = ? AND mid = ?"
	_likeExtendTokenSQL    = "SELECT id,sid,mid,token,max_limit,ctime FROM like_extend_token WHERE sid = ? AND mid = ? ORDER BY id DESC LIMIT 1"
	_likeExtendInfoSQL     = "SELECT id,sid,mid,token,max_limit,ctime FROM like_extend_token WHERE sid = ? AND token = ? ORDER BY id DESC LIMIT 1"
	_likeExtraTimesAddSQL  = "INSERT INTO like_extra_times(sid,mid,num) VALUES(?,?,?)"
	_likeExtendTokenAddSQL = "INSERT INTO like_extend_token(sid,mid,token,max_limit,ctime,mtime) VALUES(?,?,?,?,?,?)"
	HasLike                = 1
	NoLike                 = -1
	//Total number of activities set the old is bilibili-activity:like:%d
	_likeActScoreKeyFmt = "go:bl-a:l:%d"
	//Total number of comments for different types of manuscripts
	_likeActScoreTyoeKeyFmt = "go:bl:a:l:%d:%d"
	//liked key the old is bilibili-activity:like:%d:%d:%d
	_likeActKeyFmt = "go:bl-act:l:%d:%d:%d"
	//Total number of like the old is likes:oid:%d
	_likeLidKeyFmt = "go:ls:oid:%d"
	//Total number of activities like the old is sb:likes:count:%d
	_likeCountKeyFmt = "go:sb:ls:count:%d"
	_likeLimitKeyFmt = "go:li:lmt:%d:%d"
)

func likeLimitKey(sid, mid int64) string {
	return fmt.Sprintf(_likeLimitKeyFmt, sid, mid)
}

// likeActScoreKey .
func likeActScoreKey(sid int64) string {
	return fmt.Sprintf(_likeActScoreKeyFmt, sid)
}

// likeActScoreTypeKey .
func likeActScoreTypeKey(sid int64, ltype int64) string {
	return fmt.Sprintf(_likeActScoreTyoeKeyFmt, ltype, sid)
}

func likeActKey(sid, lid, mid int64) string {
	return fmt.Sprintf(_likeActKeyFmt, sid, lid, mid)
}

// likeLidKey .
func likeLidKey(oid int64) string {
	return fmt.Sprintf(_likeLidKeyFmt, oid)
}

// likeCountKey .
func likeCountKey(sid int64) string {
	return fmt.Sprintf(_likeCountKeyFmt, sid)
}

// LikeActInfos get likesaction logs.
func (dao *Dao) LikeActInfos(c context.Context, lids []int64, mid int64) (likeActs map[int64]*l.Action, err error) {
	var rows *xsql.Rows
	if rows, err = dao.db.Query(c, fmt.Sprintf(_likeActInfosSQL, xstr.JoinInts(lids)), mid); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "LikeActInfos:Query(%s)", _likeActInfosSQL)
			return
		}
	}
	defer rows.Close()
	likeActs = make(map[int64]*l.Action, len(lids))
	for rows.Next() {
		a := &l.Action{}
		if err = rows.Scan(&a.ID, &a.Lid); err != nil {
			err = errors.Wrap(err, "LikeActInfos:scan()")
			return
		}
		likeActs[a.Lid] = a
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "LikeActInfos:rows.Err()")
	}
	return
}

// UpWholeAction get like action sid.
func (dao *Dao) UpWholeAction(c context.Context, sid int64, mid int64) (likeActs []*l.Action, err error) {
	var rows *xsql.Rows
	if rows, err = dao.db.Query(c, _upWholeActionSQL, sid, mid); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "UpWholeAction:Query(%s)", _upWholeActionSQL)
			return
		}
	}
	defer rows.Close()
	for rows.Next() {
		a := &l.Action{}
		if err = rows.Scan(&a.ID, &a.Lid, &a.Mid); err != nil {
			err = errors.Wrap(err, "UpWholeAction:scan()")
			return
		}
		likeActs = append(likeActs, a)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "UpWholeAction:rows.Err()")
	}
	return
}

// UpWholeUsers get like action sid lid.
func (dao *Dao) UpWholeUsers(c context.Context, sid int64, lid int64) (likeActs []*l.Action, err error) {
	var rows *xsql.Rows
	if rows, err = dao.db.Query(c, _upWholeUserSQL, sid, lid); err != nil {
		if err == xsql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrapf(err, "UpWholeUsers:Query(%s)", _upWholeUserSQL)
			return
		}
	}
	defer rows.Close()
	for rows.Next() {
		a := &l.Action{}
		if err = rows.Scan(&a.ID, &a.Lid, &a.Mid); err != nil {
			err = errors.Wrap(err, "UpWholeUsers:scan()")
			return
		}
		likeActs = append(likeActs, a)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "UpWholeUsers:rows.Err()")
	}
	return
}

// StoryLikeActSum .
func (dao *Dao) StoryLikeActSum(c context.Context, sid, mid int64, stime, etime string) (res int64, err error) {
	var tt sql.NullInt64
	row := dao.db.QueryRow(c, _storyActSumSQL, sid, mid, stime, etime)
	if err = row.Scan(&tt); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "row.Scan()")
		}
	}
	res = tt.Int64
	return
}

// StoryEachLikeAct .
func (dao *Dao) StoryEachLikeAct(c context.Context, sid, mid, lid int64, stime, etime string) (res int64, err error) {
	var tt sql.NullInt64
	row := dao.db.QueryRow(c, _storyEachActSumSQL, sid, mid, lid, stime, etime)
	if err = row.Scan(&tt); err != nil {
		if err == sql.ErrNoRows {
			err = nil
		} else {
			err = errors.Wrap(err, "row.Scan()")
		}
	}
	res = tt.Int64
	return
}

// DelLikeListLikes .
func (dao *Dao) DelLikeListLikes(c context.Context, sid int64, items []*l.Item) (err error) {
	if len(items) == 0 {
		return
	}
	var (
		conn = component.GlobalRedisStore.Conn(c)
		key  = likeActScoreKey(sid)
		max  int
	)
	defer conn.Close()
	args := redis.Args{}.Add(key)
	for _, v := range items {
		args = args.Add(v.ID)
		if v.Type != 0 {
			typeKey := likeActScoreTypeKey(sid, v.Type)
			if err = conn.Send("ZREM", typeKey, v.ID); err != nil {
				log.Error("DelLikeListLikes:conn.Send(%v) error(%v)", v, err)
				return
			}
			max++
		}
	}
	if err = conn.Send("ZREM", args...); err != nil {
		log.Error("DelLikeListLikes:conn.Send(%v) error(%v)", args, err)
		return
	}
	max++
	if err = conn.Flush(); err != nil {
		log.Error("DelLikeListLikes:conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < max; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("DelLikeListLikes:conn.Receive(%d) error(%v)", i, err)
			return
		}
	}
	return
}

// GetLikeLimitNum .
func (dao *Dao) GetLikeLimitNum(c context.Context, sid, mid int64) (score int64, err error) {
	var (
		key = likeLimitKey(sid, mid)
	)
	if score, err = redis.Int64(component.GlobalRedisStore.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("SetLikesReload:conn.Do(GET) error(%v)", err)
		}
	}
	return
}

// SetLikeLimitNum .
func (dao *Dao) SetLikeLimitNum(c context.Context, sid, mid int64, score int64) (err error) {
	var (
		key = likeLimitKey(sid, mid)
	)
	if _, err = component.GlobalRedisStore.Do(c, "INCRBY", key, score); err != nil {
		log.Error("SetLikeLimit INCRBY(%s) error(%v)", key, err)
	}
	return
}

// SetLikesReload reload likes set cache .
func (dao *Dao) SetLikesReload(c context.Context, lid, sid, ltype int64) (err error) {
	var (
		conn   = component.GlobalRedisStore.Conn(c)
		key    = likeActScoreKey(sid)
		lidKey = likeLidKey(lid)
		score  int64
		max    int
	)
	defer conn.Close()
	if score, err = redis.Int64(conn.Do("GET", lidKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("SetLikesReload:conn.Do(GET) error(%v)", err)
		}
		return
	}
	if score <= 0 {
		return
	}
	if err = conn.Send("ZADD", key, score, lid); err != nil {
		log.Error("SetLikesReload:conn.Do(ZADD) %s error(%v)", key, err)
		return
	}
	max++
	if ltype > 0 {
		tempKey := likeActScoreTypeKey(sid, ltype)
		if err = conn.Send("ZADD", tempKey, score, lid); err != nil {
			log.Error("SetLikesReload:conn.Do(ZADD) %s error(%v)", tempKey, err)
			return
		}
		max++
	}
	if err = conn.Flush(); err != nil {
		log.Error("SetLikesReload conn.Flush(%s) error(%v)", key, err)
		return
	}
	for i := 0; i < max; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("SetLikesReload conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// SetRedisCache .
func (dao *Dao) SetRedisCache(c context.Context, sid, lid, score, likeType int64) (err error) {
	var (
		conn        = component.GlobalRedisStore.Conn(c)
		key         = likeActScoreKey(sid)
		lidKey      = likeLidKey(lid)
		lidCountKey = likeCountKey(sid)
		max         = 3
	)
	defer conn.Close()
	if err = conn.Send("ZINCRBY", key, score, lid); err != nil {
		err = errors.Wrap(err, "conn.Send(ZINCRBY) likeActScoreKey")
		return
	}
	if err = conn.Send("INCRBY", lidKey, score); err != nil {
		err = errors.Wrap(err, "conn.Send(INCR) likeLidKey")
		return
	}
	if likeType != 0 {
		max++
		typeKey := likeActScoreTypeKey(sid, likeType)
		if err = conn.Send("ZINCRBY", typeKey, score, lid); err != nil {
			err = errors.Wrap(err, "conn.Send(ZINCRBY) likeActScoreTypeKey")
			return
		}
	}
	if err = conn.Send("INCRBY", lidCountKey, score); err != nil {
		err = errors.Wrap(err, "conn.Send(INCR) likeLidKey")
		return
	}
	if err = conn.Flush(); err != nil {
		err = errors.Wrap(err, " conn.Set()")
		return
	}
	for i := 0; i < max; i++ {
		if _, err = conn.Receive(); err != nil {
			err = errors.Wrap(err, fmt.Sprintf("conn.Receive()%d", i+1))
			return
		}
	}
	return
}

// BatchSetLikeScoreCache .
func (dao *Dao) BatchSetLikeScoreCache(c context.Context, sid, score, likeType int64, lids []int64) (err error) {
	var (
		conn        = component.GlobalRedisStore.Conn(c)
		key         = likeActScoreKey(sid)
		lidCountKey = likeCountKey(sid)
		max         int
	)
	defer conn.Close()
	for _, lid := range lids {
		if err = conn.Send("ZINCRBY", key, score, lid); err != nil {
			err = errors.Wrap(err, "conn.Send(ZINCRBY) likeActScoreKey")
			return
		}
		max++
		if likeType != 0 {
			typeKey := likeActScoreTypeKey(sid, likeType)
			if err = conn.Send("ZINCRBY", typeKey, score, lid); err != nil {
				err = errors.Wrap(err, "conn.Send(ZINCRBY) likeActScoreTypeKey")
				return
			}
			max++
		}
		lidKey := likeLidKey(lid)
		if err = conn.Send("INCRBY", lidKey, score); err != nil {
			err = errors.Wrap(err, "conn.Send(INCR) likeLidKey")
			return
		}
		max++
	}
	incrScore := score * int64(len(lids))
	if err = conn.Send("INCRBY", lidCountKey, incrScore); err != nil {
		err = errors.Wrap(err, "conn.Send(INCR) likeLidKey")
		return
	}
	max++
	if err = conn.Flush(); err != nil {
		err = errors.Wrap(err, "conn.Set()")
		return
	}
	for i := 0; i < max; i++ {
		if _, err = conn.Receive(); err != nil {
			err = errors.Wrap(err, fmt.Sprintf("conn.Receive()%d", i+1))
			return
		}
	}
	return
}

// RedisCache get cache order by like .
func (dao *Dao) RedisCache(c context.Context, sid, ltype, start, end int64) (res []*l.LidLikeRes, err error) {
	var (
		key string
	)
	if ltype > 0 {
		key = likeActScoreTypeKey(sid, ltype)
	} else {
		key = likeActScoreKey(sid)
	}
	values, err := redis.Values(component.GlobalRedisStore.Do(c, "ZREVRANGE", key, start, end, "WITHSCORES"))
	if err != nil {
		err = errors.Wrap(err, "conn.Do(ZREVRANGE)")
		return
	}
	if len(values) == 0 {
		return
	}
	res = make([]*l.LidLikeRes, 0, len(values))
	for len(values) > 0 {
		t := &l.LidLikeRes{}
		if values, err = redis.Scan(values, &t.Lid, &t.Score); err != nil {
			err = errors.Wrap(err, "redis.Scan")
			return
		}
		res = append(res, t)
	}
	return
}

// LikeActZscore .
func (dao *Dao) LikeActZscore(c context.Context, sid, lid int64) (res int64, err error) {
	var (
		key = likeActScoreKey(sid)
	)
	if res, err = redis.Int64(component.GlobalRedisStore.Do(c, "ZSCORE", key, lid)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrap(err, "conn.Do(ZSCORE)")
		}
	}
	return
}

// SetInitializeLikeCache initialize like_action like data .
func (dao *Dao) SetInitializeLikeCache(c context.Context, sid int64, lidLikeAct map[int64]int64, typeLike map[int64]int64) (err error) {
	var (
		conn = component.GlobalRedisStore.Conn(c)
		max  = 0
		key  = likeActScoreKey(sid)
		args = redis.Args{}.Add(key)
	)
	defer conn.Close()
	for k, val := range lidLikeAct {
		args = args.Add(val).Add(k)
		if typeLike[k] != 0 {
			keyType := likeActScoreTypeKey(sid, typeLike[k])
			argsType := redis.Args{}.Add(keyType).Add(val).Add(k)
			if err = conn.Send("ZADD", argsType...); err != nil {
				log.Error("SetInitializeLikeCache:conn.Send(zadd) args(%v) error(%v)", argsType, err)
				return
			}
			max++
		}
	}
	if err = conn.Send("ZADD", args...); err != nil {
		log.Error("SetInitializeLikeCache:conn.Send(zadd) args(%v) error(%v)", args, err)
		return
	}
	max++
	if err = conn.Flush(); err != nil {
		err = errors.Wrap(err, "SetInitializeLikeCache:conn.Set()")
		return
	}
	for i := 0; i < max; i++ {
		if _, err = conn.Receive(); err != nil {
			err = errors.Wrapf(err, "SetInitializeLikeCache:conn.Receive()%d", i+1)
		}
	}
	return
}

// LikeActAdd add like_action .
func (dao *Dao) LikeActAdd(c context.Context, likeAct *l.Action) (id int64, err error) {
	var res sql.Result
	if res, err = dao.db.Exec(c, _likeActAddSQL, likeAct.Lid, likeAct.Mid, likeAct.Action, likeAct.Sid, likeAct.IPv6, time.Now(), time.Now(), likeAct.ExtraAction); err != nil {
		err = errors.Wrapf(err, "d.db.Exec(%s)", _likeActAddSQL)
		return
	}
	return res.LastInsertId()
}

func (dao *Dao) BatchLikeActAdd(c context.Context, adds map[int64]*l.Action) (err error) {
	var args []interface{}
	sqls := make([]string, 0, len(adds))
	if len(adds) == 0 {
		return
	}
	now := time.Now()
	for _, v := range adds {
		sqls = append(sqls, "(?,?,?,?,?,?,?)")
		args = append(args, v.Lid, v.Mid, v.Action, v.Sid, v.IPv6, now, now)
	}
	if _, err := dao.db.Exec(c, fmt.Sprintf(_likeActBatchAddSQL, strings.Join(sqls, ",")), args...); err != nil {
		log.Error("BatchAddExtend:dao.db.Exec(%v)", sqls)
	}
	return
}

// LikeActLidCounts get lid score.
func (dao *Dao) LikeActLidCounts(c context.Context, lids []int64) (res map[int64]int64, err error) {
	if len(lids) == 0 {
		return
	}
	var (
		args = redis.Args{}
		ss   []int64
	)
	for _, lid := range lids {
		args = args.Add(likeLidKey(lid))
	}
	if ss, err = redis.Int64s(component.GlobalRedisStore.Do(c, "MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			err = errors.Wrapf(err, "redis.Ints(conn.Do(HMGET,%v)", args)
		}
		return
	}
	res = make(map[int64]int64, len(lids))
	for key, val := range ss {
		res[lids[key]] = val
	}
	return
}

// LikeActs get data from cache if miss will call source method, then add to cache.
func (dao *Dao) LikeActs(c context.Context, sid, mid int64, lids []int64) (res map[int64]int, err error) {
	var (
		miss         []int64
		likeActInfos map[int64]*l.Action
		missVal      map[int64]int
	)
	if len(lids) == 0 {
		return
	}
	addCache := true
	res, err = dao.CacheLikeActs(c, sid, mid, lids)
	if err != nil {
		addCache = false
		res = nil
		err = nil
	}
	for _, key := range lids {
		if (res == nil) || (res[key] == 0) {
			miss = append(miss, key)
		}
	}
	if len(miss) == 0 {
		return
	}
	if likeActInfos, err = dao.LikeActInfos(c, miss, mid); err != nil {
		err = errors.Wrapf(err, "dao.LikeActInfos(%v) error(%v)", miss, err)
		return
	}
	if res == nil {
		res = make(map[int64]int)
	}
	missVal = make(map[int64]int, len(miss))
	for _, mcLid := range miss {
		if _, ok := likeActInfos[mcLid]; ok {
			res[mcLid] = HasLike
		} else {
			res[mcLid] = NoLike
		}
		missVal[mcLid] = res[mcLid]
	}
	if !addCache {
		return
	}
	//异步回源
	dao.cache.Do(c, func(c context.Context) {
		dao.AddCacheLikeActs(c, sid, mid, missVal)
	})
	return
}

// CacheLikeActs res value val -1:no like 1:has like 0:no value.
func (dao *Dao) CacheLikeActs(c context.Context, sid, mid int64, lids []int64) (res map[int64]int, err error) {
	l := len(lids)
	if l == 0 {
		return
	}
	keysMap := make(map[string]int64, l)
	keys := make([]string, 0, l)
	for _, id := range lids {
		key := likeActKey(sid, id, mid)
		keysMap[key] = id
		keys = append(keys, key)
	}
	replies, err := dao.mc.GetMulti(c, keys)
	if err != nil {
		log.Errorv(c, log.KV("CacheLikeActs", fmt.Sprintf("%+v", err)), log.KV("keys", keys))
		return
	}
	for _, key := range replies.Keys() {
		var v string
		err = replies.Scan(key, &v)
		if err != nil {
			log.Errorv(c, log.KV("CacheLikeActs", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
		r, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Errorv(c, log.KV("CacheLikeActs", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return res, err
		}
		if res == nil {
			res = make(map[int64]int, len(keys))
		}
		res[keysMap[key]] = int(r)
	}
	return
}

// AddCacheLikeActs Set data to mc
func (dao *Dao) AddCacheLikeActs(c context.Context, sid, mid int64, values map[int64]int) (err error) {
	if len(values) == 0 {
		return
	}
	for id, val := range values {
		key := likeActKey(sid, id, mid)
		bs := []byte(strconv.FormatInt(int64(val), 10))
		item := &memcache.Item{Key: key, Value: bs, Expiration: dao.mcLikeActExpire, Flags: memcache.FlagRAW}
		if err = dao.mc.Set(c, item); err != nil {
			log.Errorv(c, log.KV("AddCacheLikeActs", fmt.Sprintf("%+v", err)), log.KV("key", key))
			return
		}
	}
	return
}

// RawLikeActLids raw like act lids.
func (dao *Dao) RawLikeActLids(c context.Context, sid, mid int64) (lids []*l.LidItem, err error) {
	var rows *xsql.Rows
	if rows, err = dao.db.Query(c, _likeActLikesSQL, sid, mid); err != nil {
		err = errors.Wrapf(err, "RawLikeActLids:Query(%s)", _likeActInfosSQL)
		return
	}
	defer rows.Close()
	for rows.Next() {
		item := new(l.LidItem)
		if err = rows.Scan(&item.Lid, &item.Action, &item.ActTime); err != nil {
			err = errors.Wrap(err, "RawLikeActLids:scan()")
			return
		}
		lids = append(lids, item)
	}
	if err = rows.Err(); err != nil {
		err = errors.Wrap(err, "RawLikeActLids:rows.Err()")
	}
	return
}

// AddCacheLikeActLids set cache like act lids.
func (dao *Dao) AddCacheLikeActLids(c context.Context, sid int64, lids []*l.LidItem, mid int64) (err error) {
	if len(lids) == 0 {
		return
	}
	var (
		count int
		cache []*l.LidItem
	)
	key := likeActLikesKey(sid, mid)
	conn := component.GlobalRedis.Conn(c)
	defer conn.Close()
	if cache, err = dao.CacheLikeActLids(c, sid, mid); err != nil {
		log.Error("dao.CacheLikeActLids sid(%d) mid(%d) error(%v)", sid, mid, err)
		return
	}
	// empty cache
	if len(cache) == 1 && cache[0] != nil && cache[0].Lid == -1 {
		if err = conn.Send("DEL", key); err != nil {
			log.Error("dao.CacheLikeActLids sid(%d) mid(%d) error(%v)", sid, mid, err)
			return
		}
		count++
	}
	for _, v := range lids {
		bs, e := json.Marshal(v)
		if e != nil {
			log.Error("AddCacheLikeActLids json.Marshal error(%v)", e)
			continue
		}
		if err = conn.Send("ZADD", key, v.ActTime, bs); err != nil {
			log.Error("AddCacheLikeActLids conn.Send(ZADD %s %+v) error(%v)", key, v, err)
			return
		}
		count++
	}
	if err = conn.Send("EXPIRE", key, dao.likeTotalExpire); err != nil {
		log.Error("AddCacheLikeActLids conn.Do(SETEX) error(%v)", err)
		return
	}
	count++
	if err = conn.Flush(); err != nil {
		log.Error("AddCacheLikeActLids conn.Flush(%s) error(%v)", key, err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("AddCacheLikeActLids conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// AppendCacheLikeActLids append cache.
func (dao *Dao) AppendCacheLikeActLids(c context.Context, sid int64, item *l.LidItem, mid int64) (err error) {
	var ok bool
	key := likeActLikesKey(sid, mid)
	if ok, err = redis.Bool(component.GlobalRedis.Do(c, "EXPIRE", key, dao.likeTotalExpire)); err != nil {
		log.Error("AppendCacheLikeActLids conn.Do(EXPIRE %s) error(%v)", key, err)
		return
	}
	if ok {
		err = dao.AddCacheLikeActLids(c, sid, []*l.LidItem{item}, mid)
	}
	return
}

// CacheLikeActLids .
func (dao *Dao) CacheLikeActLids(c context.Context, sid, mid int64) (lids []*l.LidItem, err error) {
	key := likeActLikesKey(sid, mid)
	values, err := redis.Values(component.GlobalRedis.Do(c, "ZRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZREVRANGE,%s,%d,%d) error(%v)", key, 0, -1, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var (
		memStr string
		bs     []byte
	)
	for len(values) > 0 {
		if values, err = redis.Scan(values, &memStr, &bs); err != nil {
			log.Error("redis.Scan() error(%v)", err)
			return
		}
		lidItem := new(l.LidItem)
		if e := json.Unmarshal(bs, &lidItem); e != nil {
			log.Error("json.Unmarshal error(%v)", e)
			continue
		}
		lids = append(lids, lidItem)
	}
	return
}

// AddCacheHisLikeScore .
func (dao *Dao) AddCacheHisLikeScore(c context.Context, sid int64, value string) (err error) {
	key := likeActHisKey(sid)
	if err = dao.RsSet(c, key, value); err != nil {
		err = errors.Wrap(err, "AddCacheHisLikeScore:RsSet")
	}
	return
}

// CacheHisLikeScore .
func (dao *Dao) CacheHisLikeScore(c context.Context, sid int64) (list []*l.LidLikeRes, err error) {
	var tmpStr string
	key := likeActHisKey(sid)
	if tmpStr, err = dao.RsGet(c, key); err != nil {
		err = errors.Wrap(err, "CacheHisLikeScore:RsGet")
		return
	}
	if len(tmpStr) == 0 {
		log.Warn("CacheHisLikeScore len(tmpStr) == 0")
		return
	}
	if err = json.Unmarshal([]byte(tmpStr), &list); err != nil {
		err = errors.Wrap(err, "CacheHisLikeScore:json.Unmarshal")
	}
	return
}

func (dao *Dao) IncrLikeExtraTimes(c context.Context, sid, mid, num int64) (id int64, err error) {
	var res sql.Result
	if res, err = dao.db.Exec(c, _likeExtraTimesAddSQL, sid, mid, num); err != nil {
		err = errors.Wrapf(err, "d.db.Exec(%s)", _likeExtraTimesAddSQL)
		return
	}
	return res.LastInsertId()
}

func (dao *Dao) IncrLikeExtendToken(c context.Context, sid, mid, max int64, token string) (id int64, err error) {
	var res sql.Result
	if res, err = dao.db.Exec(c, _likeExtendTokenAddSQL, sid, mid, token, max, time.Now(), time.Now()); err != nil {
		err = errors.Wrapf(err, "d.db.Exec(%s)", _likeExtendTokenAddSQL)
		return
	}
	return res.LastInsertId()
}
