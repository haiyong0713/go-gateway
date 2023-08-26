package ott

import (
	"context"
	"fmt"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"strconv"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/pkg/errors"
)

const (
	_playerScoreKey   = "players_score_%d"
	_playerCommentKey = "players_comment_%d"
	_playerComboKey   = "combo_%d_%d"
)

func playerScoreKey(id int64) string {
	return fmt.Sprintf(_playerScoreKey, id)
}

func playerCommentKey(id int64) string {
	return fmt.Sprintf(_playerCommentKey, id)
}

func playerComboKey(gameId, mid int64) string {
	return fmt.Sprintf(_playerComboKey, gameId, mid)
}

func (d *dao) CachePlayer(c context.Context, gameId int64) ([]*model.PlayerHonor, error) {
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
		if err == redis.ErrNil {
			err = nil
		}
		return err
	}); retryErr != nil {
		log.Error("CachePlayer cache failed. GameId(%d) Err(%v)", gameId, retryErr)
		return nil, retryErr
	}
	res := make([]*model.PlayerHonor, 0)
	for i := 0; i < len(reply); i += 2 {
		mid := reply[i]
		score := reply[i+1]
		res = append(res, &model.PlayerHonor{Mid: mid, Score: score})
	}
	return res, nil
}

func (d *dao) AddCachePLayer(c context.Context, gameId int64, players []*model.PlayerHonor) error {
	if len(players) == 0 {
		return nil
	}
	var (
		conn = d.redis.Get(c)
		key  = playerScoreKey(gameId)
		args = redis.Args{}.Add(key)
	)
	defer conn.Close()
	for _, player := range players {
		args = args.Add(player.Mid).Add(player.Score)
	}
	if _, err := conn.Do("HSET", args...); err != nil {
		return errors.Wrapf(err, "addCachePLayer gameIds(%d) args(%v)", gameId, args)
	}
	return nil
}

func (d *dao) CachePlayerComment(c context.Context, gameId int64) (map[int64]string, error) {
	var (
		conn = d.redis.Get(c)
		key  = playerCommentKey(gameId)
	)
	defer conn.Close()
	reply, err := redis.Strings(conn.Do("HGETALL", key))
	if err != nil {
		return nil, err
	}
	res := make(map[int64]string)
	for i := 0; i < len(reply); i += 2 {
		mid, _ := strconv.ParseInt(reply[i], 10, 64)
		comment := reply[i+1]
		res[mid] = comment
	}
	return res, nil
}

func (d *dao) DelPlayerComment(c context.Context, gameId int64) error {
	var (
		conn = d.redis.Get(c)
		key  = playerCommentKey(gameId)
	)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		return errors.Wrapf(err, "DelPlayerComment gameId(%d)", gameId)
	}
	return nil
}

func (d *dao) CachePlayersCombo(c context.Context, gameId int64, mids []int64) (map[int64]int, error) {
	var (
		conn = d.redis.Get(c)
		args = redis.Args{}
	)
	defer conn.Close()
	for _, mid := range mids {
		key := playerComboKey(gameId, mid)
		args = args.Add(key)
	}
	reply, err := redis.Ints(conn.Do("MGET", args...))
	if err != nil {
		return nil, errors.Wrapf(err, "CachePlayersCombo gameId(%d) mids(%d)", gameId, mids)
	}
	var res = make(map[int64]int)
	for index, value := range reply {
		res[mids[index]] = value
	}
	return res, nil
}
