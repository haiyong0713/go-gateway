package guess

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/model/guess"
)

func mdKey(mainID, business int64) string {
	return fmt.Sprintf("gmd_%d_%d", mainID, business)
}

func oKey(oid, business int64) string {
	return fmt.Sprintf("go_%d_%d", oid, business)
}

func muKey(mainID, mid int64) string {
	return fmt.Sprintf("mu_%d_%d", mainID, mid)
}

func mainKey(mainID int64) string {
	return fmt.Sprintf("gm_%d", mainID)
}

func statKey(mid, stakeType, business int64) string {
	return fmt.Sprintf("st_%d_%d_%d", mid, stakeType, business)
}

func userKey(mid, business int64) string {
	return fmt.Sprintf("gu_%d_%d", mid, business)
}

// CacheUserGuess .
func (d *Dao) CacheUserGuess(c context.Context, mainIDs []int64, mid int64) (res map[int64]*guess.UserGuessLog, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, mainID := range mainIDs {
		key = muKey(mainID, mid)
		args = args.Add(key)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheUserGuess conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*guess.UserGuessLog, len(mainIDs))
	for _, bs := range bss {
		userLog := new(guess.UserGuessLog)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, &userLog); err != nil {
			log.Error("CacheUserGuess json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[userLog.MainID] = userLog
	}
	return
}

// AddCacheUserGuess .
func (d *Dao) AddCacheUserGuess(c context.Context, data map[int64]*guess.UserGuessLog, mid int64) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsMDs = redis.Args{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	for k, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheUserGuess json.Marshal err(%v)", err)
			continue
		}
		keyID = muKey(k, mid)
		keyIDs = append(keyIDs, keyID)
		argsMDs = argsMDs.Add(keyID).Add(string(bs))
	}
	if err = conn.Send("MSET", argsMDs...); err != nil {
		log.Error("AddCacheUserGuess conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	for _, v := range keyIDs {
		count++
		if err = conn.Send("EXPIRE", v, d.guessExpire); err != nil {
			log.Error("AddCacheUserGuess conn.Send(Expire, %s, %d) error(%v)", v, d.guessExpire, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// CacheOidMids .
func (d *Dao) CacheOidMIDs(c context.Context, oid, business int64) (res []*guess.MainID, err error) {
	var (
		bs   []byte
		key  = oKey(oid, business)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheOidMids .
func (d *Dao) AddCacheOidMIDs(c context.Context, oid int64, data []*guess.MainID, business int64) (err error) {
	var (
		bs   []byte
		key  = oKey(oid, business)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s,%d) error(%v)", key, oid, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.guessExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// CacheOidsMIDs .
func (d *Dao) CacheOidsMIDs(c context.Context, oids []int64, business int64) (res map[int64][]*guess.MainID, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, oid := range oids {
		key = oKey(oid, business)
		args = args.Add(key)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheOidsMIDs conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64][]*guess.MainID, len(oids))
	for _, bs := range bss {
		var (
			mains []*guess.MainID
			oid   int64
		)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, &mains); err != nil {
			log.Error("CacheOidsMIDs json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		for _, main := range mains {
			oid = main.OID
			break
		}
		res[oid] = mains
	}
	return
}

// AddCacheOidsMIDs .
func (d *Dao) AddCacheOidsMIDs(c context.Context, data map[int64][]*guess.MainID, business int64) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsMDs = redis.Args{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	for k, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheOidsMIDs json.Marshal err(%v)", err)
			continue
		}
		keyID = oKey(k, business)
		keyIDs = append(keyIDs, keyID)
		argsMDs = argsMDs.Add(keyID).Add(string(bs))
	}
	if err = conn.Send("MSET", argsMDs...); err != nil {
		log.Error("AddCacheOidsMIDs conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	for _, v := range keyIDs {
		count++
		if err = conn.Send("EXPIRE", v, d.guessExpire); err != nil {
			log.Error("AddCacheOidsMIDs conn.Send(Expire, %s, %d) error(%v)", v, d.guessExpire, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// CacheMDResult .
func (d *Dao) CacheMDResult(c context.Context, mainID, business int64) (res *guess.MainRes, err error) {
	var (
		bs   []byte
		key  = mdKey(mainID, business)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheMDResult add main detail list to cache.
func (d *Dao) AddCacheMDResult(c context.Context, id int64, data *guess.MainRes, business int64) (err error) {
	var (
		bs   []byte
		key  = mdKey(id, business)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s,%d) error(%v)", key, id, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.guessExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// CacheMDsResult .
func (d *Dao) CacheMDsResult(c context.Context, mainIDs []int64, business int64) (res map[int64]*guess.MainRes, err error) {
	var (
		key  string
		args = redis.Args{}
		bss  [][]byte
	)
	for _, mainID := range mainIDs {
		key = mdKey(mainID, business)
		args = args.Add(key)
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	if bss, err = redis.ByteSlices(conn.Do("MGET", args...)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("CacheMDsResult conn.Do(MGET,%s) error(%v)", key, err)
		}
		return
	}
	res = make(map[int64]*guess.MainRes, len(mainIDs))
	for _, bs := range bss {
		mainRes := new(guess.MainRes)
		if bs == nil {
			continue
		}
		if err = json.Unmarshal(bs, &mainRes); err != nil {
			log.Error("CacheMDsResult json.Unmarshal(%s) error(%v)", string(bs), err)
			err = nil
			continue
		}
		res[mainRes.ID] = mainRes
	}
	return
}

// AddCacheMDsResult .
func (d *Dao) AddCacheMDsResult(c context.Context, data map[int64]*guess.MainRes, business int64) (err error) {
	if len(data) == 0 {
		return
	}
	var (
		bs      []byte
		keyID   string
		keyIDs  []string
		argsMDs = redis.Args{}
	)
	conn := d.redis.Get(c)
	defer conn.Close()
	for k, v := range data {
		if bs, err = json.Marshal(v); err != nil {
			log.Error("AddCacheMDsResult json.Marshal err(%v)", err)
			continue
		}
		keyID = mdKey(k, business)
		keyIDs = append(keyIDs, keyID)
		argsMDs = argsMDs.Add(keyID).Add(string(bs))
	}
	if err = conn.Send("MSET", argsMDs...); err != nil {
		log.Error("AddCacheMDsResult conn.Send(MSET) error(%v)", err)
		return
	}
	count := 1
	for _, v := range keyIDs {
		count++
		if err = conn.Send("EXPIRE", v, d.guessExpire); err != nil {
			log.Error("AddCacheMDsResult conn.Send(Expire, %s, %d) error(%v)", v, d.guessExpire, err)
			return
		}
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// CacheGuessMain .
func (d *Dao) CacheGuessMain(c context.Context, mainID int64) (res *guess.MainGuess, err error) {
	var (
		bs   []byte
		key  = mainKey(mainID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheGuessMain .
func (d *Dao) AddCacheGuessMain(c context.Context, mainID int64, mains *guess.MainGuess) (err error) {
	var (
		bs   []byte
		key  = mainKey(mainID)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(mains); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s,%d) error(%v)", key, mainID, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.guessExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// CacheUserStat get user business guess data.
func (d *Dao) CacheUserStat(c context.Context, mid, stakeType, business int64) (res *api.UserGuessDataReply, err error) {
	var (
		bs   []byte
		key  = statKey(mid, stakeType, business)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET,%s) error(%v)", key, err)
		}
		return
	}
	if err = json.Unmarshal(bs, &res); err != nil {
		log.Error("json.Unmarshal(%s) error(%v)", string(bs), err)
	}
	return
}

// AddCacheUserStat add user business guess data to cache.
func (d *Dao) AddCacheUserStat(c context.Context, mid int64, miss *api.UserGuessDataReply, stakeType, business int64) (err error) {
	var (
		bs   []byte
		key  = statKey(mid, stakeType, business)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if bs, err = json.Marshal(miss); err != nil {
		log.Error("json.Marshal() error(%v)", err)
		return
	}
	if err = conn.Send("SET", key, bs); err != nil {
		log.Error("conn.Send(SET,%s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.guessExpire); err != nil {
		log.Error("conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("add conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("add conn.Receive()%d error(%v)", i+1, err)
			return
		}
	}
	return
}

// CacheUserGuessList .
func (d *Dao) CacheUserGuessList(c context.Context, mid, business int64) (res []*guess.UserGuessLog, err error) {
	key := userKey(mid, business)
	conn := d.redis.Get(c)
	defer conn.Close()
	values, err := redis.Values(conn.Do("ZREVRANGE", key, 0, -1, "WITHSCORES"))
	if err != nil {
		log.Error("conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return
	}
	if len(values) == 0 {
		return
	}
	var num int64
	for len(values) > 0 {
		bs := []byte{}
		if values, err = redis.Scan(values, &bs, &num); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return
		}
		match := &guess.UserGuessLog{}
		if err = json.Unmarshal(bs, match); err != nil {
			log.Error("json.Unmarshal(%v) error(%v)", bs, err)
			return
		}
		res = append(res, match)
	}
	return
}

// AppendCacheUserGuessList .
func (d *Dao) AppendCacheUserGuessList(c context.Context, mid int64, miss []*guess.UserGuessLog, business int64) (err error) {
	var ok bool
	key := userKey(mid, business)
	conn := d.redis.Get(c)
	defer conn.Close()
	if ok, err = redis.Bool(conn.Do("EXPIRE", key, d.guessExpire)); err != nil {
		log.Error("conn.Do(EXPIRE %s) error(%v)", key, err)
		return
	}
	if ok {
		err = d.AddCacheUserGuessList(c, mid, miss, business)
	}
	return
}

// AddCacheUserGuessList .
func (d *Dao) AddCacheUserGuessList(c context.Context, mid int64, miss []*guess.UserGuessLog, business int64) (err error) {
	var (
		count int
		total int64
	)
	key := userKey(mid, business)
	conn := d.redis.Get(c)
	defer conn.Close()
	for _, userGuess := range miss {
		bs, _ := json.Marshal(userGuess)
		if err = conn.Send("ZADD", key, userGuess.ID, bs); err != nil {
			log.Error("conn.Send(ZADD, %s, %s) error(%v)", key, string(bs), err)
			return
		}
		count++
	}
	if err = conn.Send("EXPIRE", key, d.guessExpire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, d.guessExpire, err)
		return
	}
	count++
	if err = conn.Send("ZCARD", key); err != nil {
		log.Error("conn.Send(ZCARD %s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	if total, err = redis.Int64(conn.Receive()); err != nil {
		log.Error("total conn.Receive() error(%v)", err)
	}
	delCount := total - _userViewMax
	if delCount > 0 {
		if _, err = conn.Do("ZREMRANGEBYRANK", key, 0, delCount-1); err != nil {
			log.Error("conn.Send(ZREMRANGEBYRANK) key(%s) error(%v)", key, err)
		}
	}
	return
}

// DelGuessCache delete  guess cache.
func (d *Dao) DelGuessCache(c context.Context, oid, business, mainID, mid, stakeType int64) (err error) {
	oidK := oKey(oid, business)
	mdK := mdKey(mainID, business)
	mK := mainKey(mainID)
	statK := statKey(mid, stakeType, business)
	conn := d.redis.Get(c)
	defer conn.Close()
	if err = conn.Send("DEL", oidK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", oidK, err)
		return
	}
	if err = conn.Send("DEL", mdK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", mdK, err)
		return
	}
	if err = conn.Send("DEL", mK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", mK, err)
		return
	}
	if err = conn.Send("DEL", statK); err != nil {
		log.Error("conn.Send(DEL, %s) error(%v)", statK, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 4; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func redisKey(key string) string {
	return "gnx_" + key
}

// RsSetNX Dao
func (d *Dao) RsSetNX(c context.Context, key string, expire int32) (res bool, err error) {
	var (
		rkey = redisKey(key)
		conn = component.GlobalRedisStore.Conn(c)
		rly  interface{}
	)
	defer conn.Close()
	if rly, err = conn.Do("SET", rkey, "1", "EX", expire, "NX"); err != nil {
		log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
		return
	}
	if rly != nil {
		res = true
	}
	return
}
