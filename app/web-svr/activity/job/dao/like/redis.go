package like

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/model/like"

	"github.com/pkg/errors"
)

const (
	_prefixAttention = "lg_"
	_likeLidKeyFmt   = "go:ls:oid:%d"
	_lotteryFmt      = "go:lt:s:%d:%s"
	// storyKing LikeAct cache
	_keyStoryDilyLikeFmt = "go:s:d:m:%s:%d:%d"
	prefix               = "act"
	separator            = ":"
	_articleRightKey     = "a_day2_right"
	_articleMidKey       = "a_day2_%d"
	synvFavLock          = "fav_lock"
	lotteryPrefix        = "lottery_new"
)

func redisKey(key string) string {
	return _prefixAttention + key
}

// likeLidKey .
func likeLidKey(oid int64) string {
	return fmt.Sprintf(_likeLidKeyFmt, oid)
}

func lotteryKey(mid int64, dt string) string {
	return fmt.Sprintf(_lotteryFmt, mid, dt)
}

func keyStoryLikeKey(sid, mid int64, daily string) string {
	return fmt.Sprintf(_keyStoryDilyLikeFmt, daily, sid, mid)
}

func entUpRankKey(sid int64) string {
	return fmt.Sprintf("ent_up_%d", sid)
}

func subjectRulesKey(sid int64) string {
	return fmt.Sprintf("sub_rule_%d", sid)
}

func likeTypeCountKey(sid int64) string {
	return fmt.Sprintf("like_type_cnt_%d", sid)
}

// articleMidKey .
func articleMidKey(mid int64) string {
	return fmt.Sprintf(_articleMidKey, mid)
}

func keyYellowGreenPeriod(yellowSid, GreenSid int64) string {
	return fmt.Sprintf("yingyuan_vote_%d_%d", yellowSid, GreenSid)
}

// buildKey build key
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return prefix + separator + strings.Join(strArgs, separator)
}

// buildKey build key
func buildKeyLottery(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return lotteryPrefix + separator + strings.Join(strArgs, separator)
}

// LotteryGet .
func (d *Dao) LotteryGet(c context.Context, mid int64) (count int, err error) {
	var (
		dt   = time.Now().Format("2006-01-02")
		key  = lotteryKey(mid, dt)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if count, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("LotterySet conn.do(%s) error(%v)", key, err)
		}
	}
	return
}

// LotterySet .
func (d *Dao) LotterySet(c context.Context, mid int64) (ok bool, err error) {
	var (
		dt   = time.Now().Format("2006-01-02")
		key  = lotteryKey(mid, dt)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("SETNX", key, 1); err != nil {
		log.Error("LotterySet conn.Send(%s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, 86400); err != nil {
		log.Error("LotterySet conn.Send(EXPIRE,%s) error(%v)", key, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("LotterySet conn.Flush() error(%v)", err)
		return
	}
	if ok, err = redis.Bool(conn.Receive()); err != nil {
		log.Error(" LotterySetconn.Receive() error(%v)", err)
		return
	}
	if _, err = conn.Receive(); err != nil {
		log.Error(" LotterySetconn.Receive() error(%v)", err)
		return
	}
	return
}

// RsSet set res
func (d *Dao) RsSet(c context.Context, key string, value string) (err error) {
	var (
		rkey = redisKey(key)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("SET", rkey, value); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", rkey, value, err)
		return
	}
	return
}

// RsSetWithNX set res
func (d *Dao) RsSetEx(c context.Context, key string, value string, expire int32) (err error) {
	var (
		rkey = redisKey(key)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("SETEX", rkey, expire, value); err != nil {
		log.Error("conn.Send(SETEX, %s, %s) error(%v)", rkey, value, err)
		return
	}
	return
}

// RbSet setRb
func (d *Dao) RbSet(c context.Context, key string, value []byte) (err error) {
	var (
		rkey = redisKey(key)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if _, err = conn.Do("SET", rkey, value); err != nil {
		log.Error("conn.Send(SET, %s, %d) error(%v)", rkey, value, err)
		return
	}
	return
}

// RsGet getRs
func (d *Dao) RsGet(c context.Context, key string) (res string, err error) {
	var (
		rkey = redisKey(key)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.String(conn.Do("GET", rkey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
		}
		return
	}
	return
}

// RiGet get int value.
func (d *Dao) RiGet(c context.Context, key string) (res int, err error) {
	var (
		rkey = redisKey(key)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int(conn.Do("GET", rkey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
		}
		return
	}
	return
}

// RsSetNX NXset get
func (d *Dao) RsSetNX(c context.Context, key string, expire int32) (res bool, err error) {
	var (
		rkey = redisKey(key)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Bool(conn.Do("SETNX", rkey, 1)); err != nil {
		log.Error("conn.Do(SETNX key(%s)) error(%v)", rkey, err)
		return
	}
	if res {
		if _, err = redis.Bool(conn.Do("EXPIRE", key, expire)); err != nil {
			log.Error("conn.Do(EXPIRE, %s, %d) error(%v)", key, expire, err)
			return
		}
	}
	return
}

// RsDelNx .
func (d *Dao) RsDelNx(c context.Context, key string) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	_, err = conn.Do("DEL", key)
	return
}

// Incr incr
func (d *Dao) Incr(c context.Context, key string, expire int32) (res bool, err error) {
	var (
		rkey = redisKey(key)
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Bool(conn.Do("incr", rkey)); err != nil {
		log.Error("conn.Do(incr key(%s)) error(%v)", rkey, err)
		return
	}
	if expire > 0 {
		if _, err = conn.Do("EXPIRE", key, expire); err != nil {
			log.Error("conn.Do(EXPIRE,%s) err(%+v)", key, err)
		}
	}
	return
}

// CreateSelection Create selection
func (d *Dao) CreateSelection(c context.Context, aid int64, stage int64) (err error) {
	key := strconv.FormatInt(aid, 10) + ":" + strconv.FormatInt(stage, 10)
	var (
		rkeyYes = redisKey(key + ":yes")
		rkeyNo  = redisKey(key + ":no")
		conn    = d.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("SET", rkeyYes, 0); err != nil {
		log.Error("conn.Send(SET %s) error(%v)", rkeyYes, err)
		return
	}
	if err = conn.Send("SET", rkeyNo, 0); err != nil {
		log.Error("conn.Send(SET %s) error(%v)", rkeyNo, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// Selection selection
func (d *Dao) Selection(c context.Context, aid int64, stage int64) (yes int64, no int64, err error) {
	key := strconv.FormatInt(aid, 10) + ":" + strconv.FormatInt(stage, 10)
	var (
		rkeyYes = redisKey(key + ":yes")
		rkeyNo  = redisKey(key + ":no")
		conn    = d.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("GET", rkeyYes); err != nil {
		log.Error("conn.Send(SET %s) error(%v)", rkeyYes, err)
		return
	}
	if err = conn.Send("GET", rkeyNo); err != nil {
		log.Error("conn.Send(SET %s) error(%v)", rkeyNo, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush() error(%v)", err)
		return
	}
	if yes, err = redis.Int64(conn.Receive()); err != nil {
		log.Error("conn.Receive(yes) error(%v)", err)
		return
	}
	if no, err = redis.Int64(conn.Receive()); err != nil {
		log.Error("conn.Receive(no) error(%v)", err)
		return
	}
	return
}

// LikeActLidCounts get lid score.
func (d *Dao) LikeActLidCounts(c context.Context, lids []int64) (res map[int64]int64, err error) {
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

// StoryLikeSum .
func (d *Dao) StoryLikeSum(c context.Context, sid, mid int64) (res int64, err error) {
	var (
		now = time.Now().Format("2006-01-02")
		key = keyStoryLikeKey(sid, mid, now)
	)
	if res, err = redis.Int64(component.GlobalRedisStore.Do(c, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			res = -1
		} else {
			err = errors.Wrap(err, "redis.Do(get)")
		}
	}
	return
}

func (d *Dao) AddCacheLikeTotal(c context.Context, key string, mid int64, expire int32) (err error) {
	var (
		conn = d.redis.Get(c)
	)
	defer conn.Close()
	if err = conn.Send("INCR", key); err != nil {
		log.Error("conn.Send(INCR, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, expire); err != nil {
		log.Error("conn.Send(Expire, %s, %d) error(%v)", key, expire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) SetEntUpRank(c context.Context, sid int64, list map[int64]*likemdl.LidLikeRes) (err error) {
	conn := component.GlobalRedisStore.Conn(c)
	defer conn.Close()
	key := entUpRankKey(sid)
	count := 0
	if err = conn.Send("DEL", key); err != nil {
		log.Error("conn.Send(DEL %s) error(%v)", key, err)
		return
	}
	count++
	for _, v := range list {
		var bs []byte
		bs, err = json.Marshal(v)
		if err != nil {
			return err
		}
		if err = conn.Send("ZADD", key, v.Score, bs); err != nil {
			log.Error("conn.Send(ZADD) score(%d) bs(%s) error(%v)", v.Score, string(bs), err)
			return
		}
		count++
	}
	if err = conn.Flush(); err != nil {
		log.Error("conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < count; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("conn.Receive error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) DelCacheSubjectRulesBySid(c context.Context, sid int64) (err error) {
	key := subjectRulesKey(sid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheSubjectRulesBySid conn.Do(DEL key:%s) error(%v)", key, err)
	}
	return nil
}

func (d *Dao) SetFactionCache(c context.Context, data []*like.Faction, hour string) (err error) {
	var value []byte
	if value, err = json.Marshal(data); err != nil {
		log.Error("SetFactionCache json marshal error(%v)", err)
		return
	}
	conn := d.redis.Get(c)
	defer conn.Close()
	key := "faction_rank"
	if err = conn.Send("SET", key, value); err != nil {
		log.Error("SetFactionCache conn.Send(SET, %s, %d) error(%v)", key, value, err)
		return
	}
	hourKey := key + "_" + hour
	if err = conn.Send("SET", hourKey, value); err != nil {
		log.Error("SetFactionCache conn.Send(SET, %s, %d) error(%v)", hourKey, value, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("SetFactionCache conn.Flush key:%s hourKey:%s error(%v)", key, hourKey, err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("SetFactionCache conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) UpCacheLikeTypeCount(ctx context.Context, sid, typeID int64) error {
	key := likeTypeCountKey(sid)
	values, err := redis.Int64Map(component.GlobalRedisStore.Do(ctx, "HGETALL", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil
		}
		log.Error("UpCacheLikeTypeCount HGETALL key:%s error(%v)", key, err)
		return err
	}
	data := make(map[int64]int64, len(values))
	for k, v := range values {
		field, e := strconv.ParseInt(k, 10, 64)
		if e != nil {
			log.Warn("CacheLikeTypeCount field(%s) strconv.ParseInt error(%v)", k, e)
			continue
		}
		data[field] = v
	}
	// reset cache
	if len(data) == 0 {
		return d.ResetCacheLikeTypeCount(ctx, sid)
	}
	if _, err = component.GlobalRedisStore.Do(ctx, "HINCRBY", key, typeID, 1); err != nil {
		log.Error("UpCacheLikeTypeCount HINCRBY key:%s typeID:%d error(%v)", key, typeID, err)
		return err
	}
	return nil
}

func (d *Dao) ResetCacheLikeTypeCount(ctx context.Context, sid int64) error {
	data, err := d.RawLikeTypeCount(ctx, sid)
	if err != nil {
		log.Error("ResetCacheLikeTypeCount RawLikeTypeCount sid:%d error(%v)", sid, err)
		return err
	}
	if len(data) == 0 {
		return nil
	}
	key := likeTypeCountKey(sid)
	// delete key
	if _, err = component.GlobalRedisStore.Do(ctx, "DEL", key); err != nil {
		log.Error("ResetCacheLikeTypeCount DEL key:%s error(%v)", key, err)
		return err
	}
	args := redis.Args{}.Add(key)
	for typeID, count := range data {
		args = append(args, typeID, count)
	}
	if _, err = component.GlobalRedisStore.Do(ctx, "HMSET", args...); err != nil {
		log.Error("ResetCacheLikeTypeCount HMSET key:%s data:%+v error(%v)", key, data, err)
		return err
	}
	//if err = conn.Flush(); err != nil {
	//	log.Error("ResetCacheLikeTypeCount Flush key:%s error(%v)", key, err)
	//	return err
	//}
	//for i := 0; i < 2; i++ {
	//	if _, err = conn.Receive(); err != nil {
	//		log.Error("ResetCacheLikeTypeCount Receive key:%s error(%v)", key, err)
	//		return err
	//	}
	//}
	return nil
}

func (d *Dao) creationKey(mid int64, activityUID string) string {
	return fmt.Sprintf("creation_%d_%s", mid, activityUID)
}

func (d *Dao) invitesKey(inviterMid int64, activityUID string) string {
	return fmt.Sprintf("invites_%d_%s", inviterMid, activityUID)
}

func (d *Dao) DelNewstar(ctx context.Context, mid, inviterMid int64, activityUID string) (err error) {
	creationKey := d.creationKey(mid, activityUID)
	inviteKey := d.invitesKey(inviterMid, activityUID)
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if err = conn.Send("DEL", creationKey); err != nil {
		log.Error("DelNewstar DEL creationKey:%s error(%v)", creationKey, err)
		return err
	}
	if err = conn.Send("DEL", inviteKey); err != nil {
		log.Error("DelNewstar DEL inviteKey:%s error(%v)", inviteKey, err)
		return err
	}
	if err = conn.Flush(); err != nil {
		log.Error("DelNewstar Flush error(%v)", err)
		return err
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("DelNewstar Receive error(%v)", err)
			return err
		}
	}
	return
}

// SetArticleDay .
func (d *Dao) SetArticleDay(c context.Context, data *like.ArticleDay) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	var bs []byte
	if bs, err = json.Marshal(data); err != nil {
		log.Errorc(c, "SetArticleDay json.Marshal(%v) error(%v)", data, err)
		return
	}
	if _, err = conn.Do("SET", _articleRightKey, bs); err != nil {
		log.Errorc(c, "SetArticleDay conn.Do(SETEX, %s, %d, %d)", _articleRightKey, -1, bs)
	}
	return
}

func (d *Dao) DelCacheArticleDayByMid(c context.Context, mid int64) (err error) {
	key := articleMidKey(mid)
	conn := d.redis.Get(c)
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DayClockIn DelCacheArticleDayByMid conn.Do(DEL key:%s) error(%v)", key, err)
	}
	return nil
}

// CacheIntValue .
func (d *Dao) CacheIntValue(ctx context.Context, key string) (res int64, err error) {
	if res, err = redis.Int64(component.GlobalRedis.Do(ctx, "GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			log.Warn("CacheIntValue(%s) return nil", key)
		} else {
			log.Error("CacheIntValue conn.Do(GET key(%v)) error(%v)", key, err)
		}
	}
	return
}

// AddCacheIntValue .
func (d *Dao) AddCacheIntValue(ctx context.Context, key string, val int64) (err error) {
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 86400, val); err != nil {
		log.Error("AddCacheIntValue conn.Send(SETEX, %s, %d, %d) error(%v)", key, 86400, val, err)
	}
	return
}

// FavSyncLock 收藏夹同步锁
func (d *Dao) FavSyncLock(c context.Context, fid int64, activityID, counter string) (err error) {
	var reply interface{}
	conn := d.redis.Get(c)
	var key = buildKey(synvFavLock, fid, activityID, counter)
	defer conn.Close()
	if reply, err = conn.Do("SET", key, "LOCK", "EX", 300, "NX"); err != nil {
		log.Errorc(c, "SETEX(%v) error(%v)", key, err)
		return
	}
	if reply == nil {
		err = errors.Errorf("FavSyncLock redis.setexnx(%d) already has", fid)
		fmt.Println(err, "err")
	}
	return
}

// DelFavSyncLock 收藏夹同步锁删除
func (d *Dao) DelFavSyncLock(c context.Context, fid int64, activityID, counter string) (err error) {
	conn := d.redis.Get(c)
	var key = buildKey(synvFavLock, fid, activityID, counter)
	defer conn.Close()
	_, err = conn.Do("DEL", key)
	return
}

// AddCacheYellowGreenVote .
func (d *Dao) AddCacheYellowGreenVote(ctx context.Context, vote *like.YgVote, period *like.YellowGreenPeriod) (err error) {
	var (
		bs []byte
	)
	key := keyYellowGreenPeriod(period.YellowYingYuanSid, period.GreenYingYuanSid)
	if bs, err = json.Marshal(vote); err != nil {
		log.Error("AddCacheYellowGreenVote json.Marshal(%v) error (%v)", vote, err)
		return
	}
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", key, 864000, bs); err != nil {
		log.Error("AddCacheYellowGreenVote conn.Send(SETEX, %s, %d, %d) error(%v)", key, 864000, bs, err)
	}
	return
}
