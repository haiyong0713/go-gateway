package ott

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/pkg/errors"
)

const (
	_gameKey       = "ott_game_%d"
	_gameTimeKey   = "time_gap_%d"
	_gamePkgKey    = "game_pkg"
	_gameQRCodeKey = "game_qrcode_%d"
)

func gameKey(id int64) string {
	return fmt.Sprintf(_gameKey, id)
}

func gameTimeKey(id int64) string {
	return fmt.Sprintf(_gameTimeKey, id)
}

func needDelKey(gameId int64, mids []int64) []string {
	var keys []string
	keys = append(keys, fmt.Sprintf(_gameKey, gameId))          // game
	keys = append(keys, fmt.Sprintf(_playerCommentKey, gameId)) // comment
	for _, mid := range mids {
		keys = append(keys, fmt.Sprintf(_playerComboKey, gameId, mid)) // combo
		keys = append(keys, playerStatKey(gameId, mid))                // stat
	}
	return keys
}

func (d *dao) cacheGame(c context.Context, id int64) (*model.OttGame, error) {
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

func (d *dao) addCacheGame(c context.Context, id int64, game *model.OttGame) error {
	var (
		conn = d.redis.Get(c)
		key  = gameKey(id)
	)
	defer conn.Close()
	data, err := json.Marshal(game)
	if err != nil {
		return errors.Wrapf(err, "addCacheGame gameId(%d) data(%v)", id, game)
	}
	if _, err = conn.Do("SET", key, data, "EX", d.gameExpire); err != nil {
		return errors.Wrapf(err, "addCacheGame gameId(%d) data(%v)", id, game)
	}
	return nil
}

func (d *dao) DelCacheGame(c context.Context, gameId int64) error {
	var (
		conn = d.redis.Get(c)
		key  = gameKey(gameId)
	)
	defer conn.Close()
	if _, err := conn.Do("DEL", key); err != nil {
		return errors.Wrapf(err, "DelCacheGame gameId(%d)", gameId)
	}
	return nil
}

func (d *dao) AddCacheGameGap(c context.Context, gameId, gap int64) error {
	var (
		conn = d.redis.Get(c)
		key  = gameTimeKey(gameId)
	)
	defer conn.Close()
	if _, err := conn.Do("SET", key, gap); err != nil {
		return errors.Wrapf(err, "AddCacheGameGap gameId(%d) gap(%d)", gameId, gap)
	}
	return nil
}

func (d *dao) AddCacheGamePkg(c context.Context, url string) error {
	var (
		conn = d.redis.Get(c)
		key  = _gamePkgKey
	)
	defer conn.Close()
	if _, err := conn.Do("SET", key, url); err != nil {
		return errors.Wrapf(err, "AddCacheGamePkg url(%s)", url)
	}
	return nil
}

func (d *dao) CacheGamePkg(c context.Context) (string, error) {
	var (
		url string
		key = _gamePkgKey
	)
	if retryErr := retry.WithAttempts(c, "CacheGamePkg", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		var (
			conn = d.redis.Get(ctx)
			err  error
		)
		defer conn.Close()
		url, err = redis.String(conn.Do("GET", key))
		return err
	}); retryErr != nil {
		log.Error("CacheGamePkg cache failed. err(%v)", retryErr)
		return url, retryErr
	}
	return url, nil
}

func (d *dao) CacheGameQRCode(c context.Context, id int64) (string, error) {
	var (
		conn = d.redis.Get(c)
		key  = fmt.Sprintf(_gameQRCodeKey, id)
	)
	defer conn.Close()
	value, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return value, errors.Wrapf(err, "CacheGameQRCode gameId(%d)", id)
	}
	return value, nil
}

func (d *dao) AddCacheQRCode(c context.Context, id int64, value string) error {
	var (
		conn = d.redis.Get(c)
		key  = fmt.Sprintf(_gameQRCodeKey, id)
	)
	defer conn.Close()
	if _, err := conn.Do("SET", key, value, "EX", d.gameExpire); err != nil {
		return errors.Wrapf(err, "addCacheQRCode value(%s)", value)
	}
	return nil
}

func (d *dao) DelCaches(c context.Context, gameId int64, mids []int64) error {
	var conn = d.redis.Get(c)
	defer conn.Close()
	keys := needDelKey(gameId, mids)
	for _, key := range keys {
		if err := conn.Send("DEL", key); err != nil {
			return errors.Wrapf(err, "DelPlayersCombo key(%s)", key)
		}
	}
	if err := conn.Flush(); err != nil {
		return errors.Wrapf(err, "DelPlayersCombo keys(%v)", keys)
	}
	for _, key := range keys {
		if _, err := conn.Receive(); err != nil {
			return errors.Wrapf(err, "DelPlayersCombo key(%s)", key)
		}
	}
	return nil
}
