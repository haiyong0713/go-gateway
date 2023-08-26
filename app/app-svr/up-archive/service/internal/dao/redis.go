package dao

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/up-archive/service/api"
	"go-gateway/app/app-svr/up-archive/service/internal/model"

	"github.com/pkg/errors"
)

type Redis struct {
	r  *redis.Redis
	dr *redis.Redis
}

func NewRedis() (r *Redis, cf func(), err error) {
	var (
		ct        paladin.Map
		cfg, dCfg redis.Config
	)
	if err = paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		err = errors.WithStack(err)
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	if err = ct.Get("Degrade").UnmarshalTOML(&dCfg); err != nil {
		err = errors.WithStack(err)
		return
	}
	r = &Redis{
		r:  redis.NewRedis(&cfg),
		dr: redis.NewRedis(&dCfg),
	}
	cf = func() {
		r.r.Close()
		r.dr.Close()
	}
	return
}

func (d *dao) PingRedis(ctx context.Context) (err error) {
	if _, err = d.redis.Do(ctx, "SET", "ping", "pong"); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}

func arcPassedKey(mid int64, without api.Without) string {
	switch without {
	case api.Without_none:
		return fmt.Sprintf("%d_arc_passed", mid)
	case api.Without_staff:
		return fmt.Sprintf("%d_arc_simple", mid)
	default:
		return fmt.Sprintf("%d_arc_%s", mid, without.String())
	}
}

func (d *dao) emptyExpire() int32 {
	rand.Seed(time.Now().UnixNano())
	return d.emptyCacheExpire + rand.Int31n(d.emptyCacheRand)
}

func (d *dao) CacheArcPassed(ctx context.Context, mid, start, end int64, isAsc bool, without api.Without) ([]int64, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcPassedKey(mid, without)
	if key == "" {
		return nil, nil
	}
	cmd := "ZREVRANGE"
	if isAsc {
		cmd = "ZRANGE"
	}
	aids, err := redis.Int64s(conn.Do(cmd, key, start, end))
	if err != nil {
		return nil, errors.Wrapf(err, "CacheArcPassed key:%s start:%d end:%d", key, start, end)
	}
	return aids, nil
}

func (d *dao) CacheArcPassedTotal(ctx context.Context, mid int64, without api.Without) (int64, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcPassedKey(mid, without)
	if key == "" {
		return 0, nil
	}
	total, err := redis.Int64(conn.Do("ZCARD", key))
	if err != nil {
		return 0, errors.Wrapf(err, "CacheArcPassedTotal key:%s", key)
	}
	return total, nil
}

func (d *dao) CacheArcsPassed(ctx context.Context, mids []int64, start, end int64, isAsc bool, without api.Without) (map[int64][]int64, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	for _, mid := range mids {
		key := arcPassedKey(mid, without)
		if key == "" {
			return nil, nil
		}
		cmd := "ZREVRANGE"
		if isAsc {
			cmd = "ZRANGE"
		}
		if err := conn.Send(cmd, key, start, end); err != nil {
			return nil, errors.Wrapf(err, "CacheArcsPassed key:%s start:%d end:%d", key, start, end)
		}
	}
	if err := conn.Flush(); err != nil {
		return nil, err
	}
	res := map[int64][]int64{}
	for _, mid := range mids {
		aids, err := redis.Int64s(conn.Receive())
		if err != nil {
			log.Error("%+v", err)
			continue
		}
		res[mid] = aids
	}
	return res, nil
}

func (d *dao) CacheArcsPassedTotal(ctx context.Context, mids []int64, without api.Without) (map[int64]int64, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	for _, mid := range mids {
		key := arcPassedKey(mid, without)
		if key == "" {
			return nil, nil
		}
		if err := conn.Send("ZCARD", key); err != nil {
			return nil, errors.Wrapf(err, "CacheArcsPassedTotal key:%s", key)
		}
	}
	if err := conn.Flush(); err != nil {
		return nil, err
	}
	res := map[int64]int64{}
	for _, mid := range mids {
		total, err := redis.Int64(conn.Receive())
		if err != nil {
			return nil, err
		}
		res[mid] = total
	}
	return res, nil
}

func (d *dao) ExpireEmptyArcPassed(ctx context.Context, mid int64, without api.Without) error {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcPassedKey(mid, without)
	if key == "" {
		return nil
	}
	ttl, err := redis.Int64(conn.Do("TTL", key))
	if err != nil {
		return errors.Wrapf(err, "ExpireEmptyArcPassed TTL key:%s", key)
	}
	// only expire if not has expire
	if ttl != -1 {
		return nil
	}
	if _, err = conn.Do("EXPIRE", key, d.emptyExpire()); err != nil {
		return errors.Wrapf(err, "ExpireEmptyArcPassed EXPIRE key:%s", key)
	}
	return nil
}

func (d *dao) CacheArcPassedExists(ctx context.Context, mid int64, without api.Without) (bool, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcPassedKey(mid, without)
	if key == "" {
		return false, nil
	}
	exist, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return false, errors.Wrapf(err, "CacheArcPassedExists key:%s", key)
	}
	return exist, nil
}

func (d *dao) CacheArcPassedCursor(ctx context.Context, mid, score, ps int64, isAsc, containScore bool, without api.Without) ([]*api.ArcPassed, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcPassedKey(mid, without)
	if key == "" {
		return nil, nil
	}
	start := fmt.Sprintf("(%d", score)
	if containScore {
		start = fmt.Sprintf("%d", score)
	}
	values, err := func() ([]interface{}, error) {
		if isAsc {
			return redis.Values(conn.Do("ZRANGEBYSCORE", key, start, "+inf", "WITHSCORES", "LIMIT", 0, ps))
		}
		return redis.Values(conn.Do("ZREVRANGEBYSCORE", key, start, "-inf", "WITHSCORES", "LIMIT", 0, ps))
	}()
	if err != nil {
		return nil, errors.Wrapf(err, "CacheArcPassedCursor conn.Do key:%s error:%+v", key, err)
	}
	var list []*api.ArcPassed
	for len(values) > 0 {
		aidScore := new(api.ArcPassed)
		if values, err = redis.Scan(values, &aidScore.Aid, &aidScore.Score); err != nil {
			return nil, errors.Wrapf(err, "CacheArcPassedCursor redis.Scan key:%s error:%+v", key, err)
		}
		list = append(list, aidScore)
	}
	return list, nil
}

func (d *dao) CacheArcPassedScoreRank(ctx context.Context, mid, aid int64, isAsc bool, without api.Without) (score, rank int64, err error) {
	p := d.redis.Pipeline()
	key := arcPassedKey(mid, without)
	if key == "" {
		return 0, 0, nil
	}
	p.Send("ZSCORE", key, aid)
	cmd := "ZREVRANK"
	if isAsc {
		cmd = "ZRANK"
	}
	p.Send(cmd, key, aid)
	rs, err := p.Exec(ctx)
	if err != nil {
		return 0, 0, err
	}
	if score, err = redis.Int64(rs.Scan()); err != nil {
		return 0, 0, err
	}
	if rank, err = redis.Int64(rs.Scan()); err != nil {
		return 0, 0, err
	}
	return score, rank, nil
}

// nolint:gomnd
func (d *dao) CacheUpsPassed(ctx context.Context, mids []int64, start, end int64, isAsc bool, without api.Without) (map[int64][]*api.AidPubTime, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	for _, mid := range mids {
		key := arcPassedKey(mid, without)
		if key == "" {
			return nil, nil
		}
		cmd := "ZREVRANGE"
		if isAsc {
			cmd = "ZRANGE"
		}
		if err := conn.Send(cmd, key, start, end, "WITHSCORES"); err != nil {
			return nil, errors.Wrapf(err, "CacheUpsPassed key:%s start:%d end:%d", key, start, end)
		}
	}
	if err := conn.Flush(); err != nil {
		return nil, err
	}
	res := map[int64][]*api.AidPubTime{}
	for _, mid := range mids {
		aidScores, err := redis.Int64s(conn.Receive())
		if err != nil {
			log.Error("%+v", err)
			continue
		}
		for i := 0; i < len(aidScores); i += 2 {
			aid := aidScores[i]
			score := aidScores[i+1]
			ptime := score >> 21
			copyright := int32(score & 3)
			res[mid] = append(res[mid], &api.AidPubTime{Aid: aid, Pubdate: xtime.Time(ptime), Copyright: copyright})
		}
	}
	return res, nil
}

func (d *dao) CacheArcPassedExist(ctx context.Context, mid, aid int64, without api.Without) (bool, error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	key := arcPassedKey(mid, without)
	if key == "" {
		return false, nil
	}
	if _, err := redis.Int64(conn.Do("ZRANK", key, aid)); err != nil {
		if err != redis.ErrNil {
			return false, errors.Wrapf(err, "CacheArcPassedExist key:%s aid%d", key, aid)
		}
		return false, nil
	}
	return true, nil
}

func (d *dao) degradeExpire() int32 {
	rand.Seed(time.Now().UnixNano())
	return d.degradeCacheExpire + rand.Int31n(d.degradeCacheRand)
}

func (d *dao) degradeEmptyExpire() int32 {
	rand.Seed(time.Now().UnixNano())
	return d.degradeEmptyCacheExpire + rand.Int31n(d.emptyCacheRand)
}

func cacheSFSearch(prefix string, val ...interface{}) string {
	vs := make([]string, 0, len(val))
	for _, arg := range val {
		vs = append(vs, fmt.Sprintf("%v", arg))
	}
	s := fmt.Sprintf("%s_%x", prefix, md5.Sum([]byte(strings.Join(vs, "-"))))
	log.Info("cacheSFSearch %+v,%s", val, s)
	return s
}

func (d *dao) CacheArcSearch(ctx context.Context, key string) (*model.ResultCache, error) {
	bs, err := redis.Bytes(d.dRedis.Do(ctx, "GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, err
	}
	var data *model.ResultCache
	if err := json.Unmarshal(bs, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *dao) AddCacheArcSearch(ctx context.Context, key string, miss interface{}) error {
	data := struct {
		Reply interface{} `json:"reply"`
		Ctime time.Time   `json:"ctime"`
	}{
		Reply: miss,
		Ctime: time.Now(),
	}
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	expire := d.degradeExpire()
	if miss == nil {
		expire = d.degradeEmptyExpire()
	}
	_, err = d.dRedis.Do(ctx, "SETEX", key, expire, bs)
	return err
}
