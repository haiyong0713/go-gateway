package dao

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"

	"go-gateway/app/web-svr/dance-taiko/interface/api"
	"go-gateway/app/web-svr/dance-taiko/job/model"
)

const (
	_playerScoreKey   = "players_score_%d"
	_playerCommentKey = "players_comment_%d"
	_playerComboKey   = "combo_%d_%d"
)

func playerStatKey(gameId, mid int64) string {
	return fmt.Sprintf("ott_stat_%d_%d", gameId, mid)
}

func playerScoreKey(id int64) string {
	return fmt.Sprintf(_playerScoreKey, id)
}

func playerCommentKey(id int64) string {
	return fmt.Sprintf(_playerCommentKey, id)
}

func playerComboKey(gameId, mid int64) string {
	return fmt.Sprintf(_playerComboKey, gameId, mid)
}

// PickPlayerStats 获取区间内的动作数据，用于对关键帧进行评分；start/end是自然时间
func (d *Dao) PickPlayerStats(c context.Context, gameId, mid, start, end int64) ([]*api.StatAcc, error) {
	conn := d.redis.Get(c)
	defer conn.Close()

	key := playerStatKey(gameId, mid)
	res := make([]*api.StatAcc, 0)
	values, err := redis.Values(conn.Do("ZRANGEBYSCORE", key, start, end, "WITHSCORES"))
	log.Warn("ZRANGEBYSCORE %s %d %d WITHSCORES, Res %v", key, start, end, values)
	if err != nil {
		log.Error("conn.Do(ZRANGEBYSCORE, %s) error(%v)", key, err)
		return res, err
	}
	if len(values) == 0 {
		return res, nil
	}
	for len(values) > 0 {
		var acc float64
		var ts int64
		if values, err = redis.Scan(values, &acc, &ts); err != nil {
			log.Error("redis.Scan(%v) error(%v)", values, err)
			return res, err
		}
		object := &api.StatAcc{
			Ts:  ts,
			Acc: acc,
		}
		res = append(res, object)
	}
	return res, nil
}

// DelUnusedStats 让key过期
func (d *Dao) DelUnusedStats(c context.Context, gameId, mid int64) error {
	conn := d.redis.Get(c)
	defer conn.Close()

	key := playerStatKey(gameId, mid)
	if _, err := conn.Do("EXPIRE", key, d.gameExpire); err != nil {
		log.Error("conn.Send(EXPIRE) key(%s) error(%v)", key, err)
		return err
	}

	return nil
}

// AddCacheComments 添加玩家的评分信息
func (d *Dao) AddCacheComments(c context.Context, gameId int64, players []model.PlayerComment) error {
	if len(players) == 0 {
		return nil
	}
	var (
		key  = playerCommentKey(gameId)
		args = redis.Args{}.Add(key)
	)
	for _, v := range players {
		if v.Mid == 0 {
			continue
		}
		args = args.Add(v.Mid).Add(v.Comment)
	}
	if retryErr := retry.WithAttempts(c, "AddCacheComments", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		conn := d.redis.Get(ctx)
		defer conn.Close()
		_, err := conn.Do("HSET", args...)
		return err
	}); retryErr != nil {
		log.Error("AddCacheComments key %s cache failed. Err(%v)", key, retryErr)
		return retryErr
	}
	return nil
}

func (d *Dao) CachePlayersCombo(c context.Context, gameId int64, mids []int64) (map[int64]int64, error) {
	var (
		conn = d.redis.Get(c)
		args = redis.Args{}
	)
	defer conn.Close()
	for _, mid := range mids {
		key := playerComboKey(gameId, mid)
		args = args.Add(key)
	}
	reply, err := redis.Int64s(conn.Do("MGET", args...))
	if err != nil {
		return nil, errors.Wrapf(err, "CachePlayersCombo gameId(%d) mids(%d)", gameId, mids)
	}
	var res = make(map[int64]int64)
	for index, value := range reply {
		if value == 0 {
			log.Warn("CachePlayersCombo value(%d) err(%v)", value, err)
			continue
		}
		res[mids[index]] = value
	}
	return res, nil
}

func (d *Dao) AddCachePlayersCombo(c context.Context, gameId int64, combos []model.PlayerCombo) error {
	var (
		conn = d.redis.Get(c)
		args = redis.Args{}
		keys = make([]string, 0)
	)
	defer conn.Close()
	for _, combo := range combos {
		key := playerComboKey(gameId, combo.Mid)
		keys = append(keys, key)
		args = args.Add(key).Add(combo.Combo)
	}
	if _, err := conn.Do("MSET", args...); err != nil {
		return errors.Wrapf(err, "conn.Do(MSET) keys:%+v error", keys)
	}
	return nil
}
