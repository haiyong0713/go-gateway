package dao

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/dance-taiko/job/model"
)

const (
	_gameKey     = "ott_game_%d"
	_gameTimeKey = "time_gap_%d"
)

func gameTimeKey(id int64) string {
	return fmt.Sprintf(_gameTimeKey, id)
}

func gameKey(id int64) string {
	return fmt.Sprintf(_gameKey, id)
}

func (d *Dao) AddCacheGame(c context.Context, game *model.OttGame) error {
	var (
		data []byte
		key  = gameKey(game.GameId)
		err  error
	)
	if data, err = json.Marshal(game); err != nil {
		return errors.Wrapf(err, "AddCacheGame gameId(%d) data(%v)", game.GameId, game)
	}
	if retryErr := retry.WithAttempts(c, "AddCacheGame", 5, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		conn := d.redis.Get(ctx)
		defer conn.Close()
		_, err := conn.Do("SET", key, data, "EX", d.gameExpire)
		return err
	}); retryErr != nil {
		log.Error("AddCacheGame cache failed. Value(%v) Err(%v)", game, retryErr)
		return retryErr
	}
	return nil
}

func (d *Dao) AddCachePlayers(c context.Context, gameId int64, players []model.PlayerHonor) error {
	if len(players) == 0 {
		return nil
	}
	key := playerScoreKey(gameId)
	conn := d.redis.Get(c)
	defer conn.Close()

	mids := []int64{}
	for _, v := range players {
		mids = append(mids, v.Mid)
	}
	args := redis.Args{}.Add(key).AddFlat(mids)

	scores, err := redis.Ints(conn.Do("HMGET", args...))
	if err != nil {
		log.Error("redis.Ints(HMGET) key(%s) args(%v) error(%v)", key, args, err)
		return err
	}

	args = redis.Args{}.Add(key)
	for k, v := range players {
		if myScore := int64(scores[k]) + v.Score; myScore > int64(d.c.Cfg.MaxScore-1) {
			args = args.Add(v.Mid).Add(int64(d.c.Cfg.MaxScore) - 1)
			log.Error("日志报警 GameID %d, Mid %d, Score %d, 超过5000分啦", gameId, v.Mid, myScore)
		} else {
			args = args.Add(v.Mid).Add(myScore)
		}
	}
	if _, err = conn.Do("HSET", args...); err != nil {
		log.Error("AddCachePlayers GameID %d Err %v", gameId, err)
		return err
	}
	return nil
}

func (d *Dao) CachePlayerMap(c context.Context, gameId int64) (map[int64]*model.PlayerHonor, error) {
	var (
		key   = playerScoreKey(gameId)
		reply []int64
	)
	if retryErr := retry.WithAttempts(c, "CachePlayer", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		var (
			conn = d.redis.Get(ctx)
			err  error
		)
		defer conn.Close()
		reply, err = redis.Int64s(conn.Do("HGETALL", key))
		return err
	}); retryErr != nil {
		log.Error("CachePlayer cache failed. GameId(%d) Err(%v)", gameId, retryErr)
		return nil, retryErr
	}
	res := make(map[int64]*model.PlayerHonor)
	for i := 0; i < len(reply); i += 2 {
		mid := reply[i]
		score := reply[i+1]
		res[mid] = &model.PlayerHonor{Mid: mid, Score: score}
	}
	return res, nil
}

// CacheGameGap picks the gap time
func (d *Dao) CacheGameGap(c context.Context, gameId int64) (int64, error) {
	conn := d.redis.Get(c)
	defer conn.Close()
	gap, err := redis.Int64(conn.Do("GET", gameTimeKey(gameId)))
	if err != nil {
		if err == redis.ErrNil {
			return 0, nil
		}
		return 0, errors.Wrapf(err, "AddCacheGameGap gameId(%d) gap(%d)", gameId, gap)
	}
	return gap, nil
}

func (d *Dao) CacheGame(c context.Context, id int64) (*model.OttGame, error) {
	var (
		conn = d.redis.Get(c)
		key  = gameKey(id)
	)
	defer conn.Close()
	reply, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "cacheGame gameId(%d)", id)
	}
	res := new(model.OttGame)
	if err = json.Unmarshal(reply, &res); err != nil {
		return nil, errors.Wrapf(err, "cacheGame gameId(%d)", id)
	}
	return res, nil
}
