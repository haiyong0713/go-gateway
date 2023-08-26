package dao

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"
)

const (
	_redisExpire = 24 * 60 * 60 //24h
	_filePathKey = "dt_file"
)

func NewRedis() (r *redis.Redis, cf func(), err error) {
	var (
		cfg redis.Config
		ct  paladin.Map
	)
	if err = paladin.Get("redis.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	r = redis.NewRedis(&cfg)
	cf = func() { r.Close() }
	return
}

func (d *dao) PingRedis(ctx context.Context) (err error) {
	if _, err = d.redis.Do(ctx, "SET", "ping", "pong"); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}

func (d *dao) genTopTenKey(aid int64) string {
	return fmt.Sprintf("dt_top_%d_%d", aid, time.Now().Day())
}

func (d *dao) genGameSetKey(gameId int64) string {
	return fmt.Sprintf("dt_game_%d", gameId)
}

func (d *dao) genPointKey(gameId int64) string {
	return fmt.Sprintf("dt_point_%d", gameId)
}

func (d *dao) genCommentKey(gameId int64) string {
	return fmt.Sprintf("dt_comment_%d", gameId)
}

func (d *dao) genCommentField(mid int64) string {
	return fmt.Sprintf("%d_%d", mid, time.Now().Unix())
}

func (d *dao) genSTimeKey(gameId int64) string {
	return fmt.Sprintf("dt_stime_%d", gameId)
}

func (d *dao) gameRestartKey(gameId int) string {
	return fmt.Sprintf("game_restart_%d", gameId)
}

func (d *dao) gameExperimentKey(gameId int64) string {
	return fmt.Sprintf("game_experiment_%d", gameId)
}

func (d *dao) ParseCommentField(field string) (mid int64, ts int64, err error) {
	s := strings.Split(field, "_")
	if len(s) != 2 {
		log.Error("unexpected field name: %s", field)
		err = ecode.ServerErr
		return
	}
	if mid, err = strconv.ParseInt(s[0], 10, 64); err != nil {
		return
	}
	if ts, err = strconv.ParseInt(s[1], 10, 64); err != nil {
		return
	}
	return
}

func (d *dao) AddRedisExperiment(c context.Context, gameId int64) error {
	var (
		conn = d.redis.Conn(c)
		key  = d.gameExperimentKey(gameId)
	)
	defer conn.Close()
	_, err := conn.Do("SET", key, gameId, "EX", _redisExpire)
	if err != nil {
		log.Error("AddRedisExperiment key %s err %v", key, err)
	}
	return err
}

func (d *dao) RedisExperiment(c context.Context, gameId int64) (bool, error) {
	var (
		conn = d.redis.Conn(c)
		key  = d.gameExperimentKey(gameId)
	)
	defer conn.Close()
	reply, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		log.Error("RedisExperiment key %s err %v", key, err)
		return false, nil
	}
	if reply == gameId {
		return true, nil
	}
	return false, nil
}

func (d *dao) RedisJoin(c context.Context, gameId int64, position int, player string) (err error) {
	key := d.genGameSetKey(gameId)
	conn := d.redis.Conn(c)
	defer conn.Close()

	if _, err = conn.Do("HSET", key, strconv.Itoa(position), player); err != nil {
		log.Error("RedisJoin key:%s player:%s position:%d err:%v", key, player, position, err)
	}
	if _, err = conn.Do("EXPIRE", key, _redisExpire); err != nil {
		log.Error("conn.Do(EXPIRE, %s, %d) error(%+v)", key, _redisExpire, err)
		return
	}
	return
}

func (d *dao) RedisGetJoinedPlayers(c context.Context, gameId int64) (players map[int]string, err error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := d.genGameSetKey(gameId)
	players = make(map[int]string)
	var values map[string]string
	if values, err = redis.StringMap(conn.Do("HGETALL", key)); err != nil {
		if err == redis.ErrNil {
			return players, nil
		}
		log.Error("conn.Do(HGETALL %s) error(%v)", key, err)
		return
	}

	for k, v := range values {
		intKey, err := strconv.Atoi(k)
		if err != nil {
			log.Error("parse mid to int64 fail mid:%s err:%v", k, err)
			err = nil
			continue
		}
		players[intKey] = v
	}
	return
}

func (d *dao) RedisIncrPoint(c context.Context, gameId int64, mid int64, delta int64) (err error) {
	key := d.genPointKey(gameId)
	conn := d.redis.Conn(c)
	defer conn.Close()
	if _, err = conn.Do("HINCRBY", key, strconv.FormatInt(mid, 10), delta); err != nil {
		log.Error("RedisIncrPoint key:%s mid:%d delta:%d err:%v", key, mid, delta, err)
	}
	if _, err = conn.Do("EXPIRE", key, _redisExpire); err != nil {
		log.Error("conn.Do(EXPIRE, %s, %d) error(%+v)", key, mid, err)
		return
	}
	return
}

func (d *dao) RedisGetPoints(c context.Context, gameId int64) (res map[int64]int64, err error) {
	key := d.genPointKey(gameId)
	var (
		values map[string]int
		conn   = d.redis.Conn(c)
	)
	defer conn.Close()
	res = make(map[int64]int64)
	if values, err = redis.IntMap(conn.Do("HGETALL", key)); err != nil {
		if err == redis.ErrNil {
			return res, nil
		}
		log.Error("conn.Do(HGETALL %s) error(%v)", key, err)
		return
	}

	for k, v := range values {
		intKey, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			log.Error("parse mid to int64 fail mid:%s err:%v", k, err)
			err = nil
			continue
		}
		res[intKey] = int64(v)
	}
	return
}

func (d *dao) RedisSetComment(c context.Context, gameId int64, mid int64, comment string) (err error) {
	key := d.genCommentKey(gameId)
	conn := d.redis.Conn(c)
	defer conn.Close()
	if _, err = conn.Do("HSET", key, d.genCommentField(mid), comment); err != nil {
		log.Error("RedisSetComment key:%s mid:%d comment:%s err:%v", key, mid, comment, err)
	}
	if _, err = conn.Do("EXPIRE", key, _redisExpire); err != nil {
		log.Error("conn.Do(EXPIRE, %s, %d) error(%+v)", key, mid, err)
		return
	}
	return
}

func (d *dao) redisDelComment(c context.Context, gameId int64, fields []string) (err error) {
	if len(fields) == 0 {
		return
	}
	conn := d.redis.Conn(c)
	defer conn.Close()
	args := []interface{}{d.genCommentKey(gameId)}
	for _, f := range fields {
		args = append(args, f)
	}
	if _, err = conn.Do("HDEL", args...); err != nil {
		log.Error("conn.Do(HDEL %+v) error(%v)", args, err)
	}
	return
}

func (d *dao) RedisGetComments(c context.Context, gameId int64) (res map[int64]string, err error) {

	key := d.genCommentKey(gameId)
	var (
		values map[string]string
		conn   = d.redis.Conn(c)
	)
	defer conn.Close()
	res = make(map[int64]string, 0)
	timestamps := make(map[int64]int64, 0)
	if values, err = redis.StringMap(conn.Do("HGETALL", key)); err != nil {
		if err == redis.ErrNil {
			return res, nil
		}
		log.Error("conn.Do(HGETALL %s) error(%v)", key, err)
		return
	}

	var fields []string
	for f, v := range values {
		fields = append(fields, f)
		mid, ts, err := d.ParseCommentField(f)
		if err != nil {
			continue
		}

		_, ok := res[mid]
		if !ok {
			res[mid] = v
			timestamps[mid] = ts
		} else {
			if ts > timestamps[mid] {
				res[mid] = v
				timestamps[mid] = ts
			}
		}
	}

	_ = d.redisDelComment(c, gameId, fields)
	return
}

func (d *dao) RedisSetFilePath(c context.Context, filePath string) (err error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	if _, err = conn.Do("SET", _filePathKey, filePath); err != nil {
		log.Error("RedisSetFilePath key:%s filePath:%s err:%v", _filePathKey, filePath, err)
	}
	return
}

func (d *dao) RedisGetFilePath(c context.Context) (filePath string, err error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	if filePath, err = redis.String(conn.Do("GET", _filePathKey)); err != nil {
		log.Error("redis conn.Do(GET %s) error(%v)", _filePathKey, err)
	}
	return
}

func (d *dao) RedisDelPoints(c context.Context, gameId int64) (err error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := d.genPointKey(gameId)
	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("redis conn.Do(DEL %s) error(%v)", key, err)
	}
	return
}

func (d *dao) RedisSetSTime(c context.Context, gameId int64, sTime int64) (err error) {
	key := d.genSTimeKey(gameId)
	conn := d.redis.Conn(c)
	defer conn.Close()
	if _, err = conn.Do("SET", key, sTime, "EX", _redisExpire); err != nil {
		log.Error("RedisSetSTime key:%s gameId:%d filePath:%d err:%v", key, gameId, sTime, err)
	}
	return
}

func (d *dao) RedisGetSTime(c context.Context, gameId int64) (sTime int64, err error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := d.genSTimeKey(gameId)
	if sTime, err = redis.Int64(conn.Do("GET", key)); err != nil {
		log.Error("redis conn.Do(GET %s) error(%v)", key, err)
	}
	return
}

func (d *dao) RedisSetUserPoints(c context.Context, aid int64, points map[int64]int64) error {
	if len(points) == 0 { // 空跑没有用户，不提示
		return nil
	}

	conn := d.redis.Conn(c)
	defer conn.Close()
	key := d.genTopTenKey(aid)

	mids := []int64{}
	for mid := range points {
		// 查分，来确保保留用户在当前aid的最高分
		if err := conn.Send("ZSCORE", key, mid); err != nil {
			return err
		}
		mids = append(mids, mid)
	}
	if err := conn.Flush(); err != nil {
		log.Error("RedisSetUserPoints conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	args := redis.Args{}.Add(key)
	for i := 0; i < len(mids); i++ {
		score, _ := redis.Int64(conn.Receive())
		if score < points[mids[i]] {
			score = points[mids[i]] // 如果本次得分更高，替换
		}
		args = args.Add(score).Add(mids[i])
	}
	if _, err := conn.Do("ZADD", args...); err != nil {
		log.Error("RedisSetUserPoints conn.Send(ZADD, %s, %v) error(%v)", key, args, err)
		return err
	}
	return nil
}

func (d *dao) RedisGetUserPoints(c context.Context, aid int64, number int64) ([]*model.PlayerHonor, error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := d.genTopTenKey(aid)

	values, err := redis.Values(conn.Do("ZREVRANGE", key, 0, number-1, "WITHSCORES"))
	if err != nil {
		log.Error("日志报警 conn.Do(ZREVRANGE, %s) error(%v)", key, err)
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	points := make([]*model.PlayerHonor, 0)
	for len(values) > 0 {
		var mid, score int64
		if values, err = redis.Scan(values, &mid, &score); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return nil, err
		}
		points = append(points, &model.PlayerHonor{
			Mid:   mid,
			Score: score,
		})
	}
	return points, nil
}

func (d *dao) RedisSetGame(c context.Context, gameId int) error {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := d.gameRestartKey(gameId)

	if _, err := conn.Do("SET", key, gameId, "EX", d.conf.DemoExpire); err != nil {
		log.Error("RedisSetGame gameId(%d) err(%v)", gameId, err)
	}
	return nil
}

func (d *dao) RedisGetGame(c context.Context, gameId int) (int, error) {
	conn := d.redis.Conn(c)
	defer conn.Close()
	key := d.gameRestartKey(gameId)

	reply, err := redis.Int(conn.Do("GET", key))
	if err != nil {
		log.Error("RedisGetGame gameId(%d) err(%v)", gameId, err)
		return 0, err
	}
	return reply, nil
}
