package ott

import (
	"context"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/dance-taiko/interface/internal/model"

	"github.com/pkg/errors"
)

const (
	_cacheRankLey = "rank_%d_%d"
)

func cacheRankKey(cid int64) string {
	return fmt.Sprintf(_cacheRankLey, cid, getFirstDateOfWeek())
}

func getFirstDateOfWeek() int64 {
	now := time.Now()
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	return weekStartDate.UnixNano() / int64(time.Millisecond)
}

func (d *dao) CacheRank(c context.Context, cid int64, pn, ps int) ([]*model.PlayerHonor, error) {
	var (
		conn = d.redis.Get(c)
		key  = cacheRankKey(cid)
		res  = make([]*model.PlayerHonor, 0)
	)
	defer conn.Close()
	reply, err := redis.Values(conn.Do("ZREVRANGE", key, pn*ps, (pn+1)*ps-1, "WITHSCORES"))
	if err != nil {
		if err == redis.ErrNil {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "CacheRank cid(%d)", cid)
	}
	for len(reply) > 0 {
		player := new(model.PlayerHonor)
		if reply, err = redis.Scan(reply, &player.Mid, &player.Score); err != nil {
			log.Warn("redis.Scan err key:%s", key)
			continue
		}
		res = append(res, player)
	}
	return res, nil
}

func (d *dao) AddCacheRanks(c context.Context, cid int64, players []*model.PlayerHonor) error {
	var (
		conn = d.redis.Get(c)
		args = redis.Args{}.Add(cacheRankKey(cid))
	)
	defer conn.Close()
	for _, player := range players {
		args = args.Add(player.Score).Add(player.Mid)
	}
	if err := conn.Send("ZADD", args...); err != nil {
		return errors.Wrapf(err, "ZADD cid(%d) players(%v)", cid, players)
	}
	if err := conn.Send("EXPIRE", d.rankExpire); err != nil {
		return errors.Wrapf(err, "EX cid(%d) players(%v)", cid, players)
	}
	if err := conn.Flush(); err != nil {
		return err
	}
	return nil
}

func (d *dao) AddCacheRank(c context.Context, cid int64, players []*model.PlayerHonor) error {
	var (
		conn = d.redis.Get(c)
		key  = cacheRankKey(cid)
	)
	defer conn.Close()
	// 先查分
	for _, play := range players {
		if err := conn.Send("ZSCORE", key, play.Mid); err != nil {
			return err
		}
	}
	if err := conn.Flush(); err != nil {
		log.Error("RedisSetUserPoints conn.Send(ZADD, %s) error(%v)", key, err)
		return err
	}
	args := redis.Args{}.Add(key)
	for _, player := range players {
		score, _ := redis.Int64(conn.Receive())
		if score < player.Score {
			score = player.Score // 如果本次得分更高，替换
			args = args.Add(score).Add(player.Mid)
		}
	}
	if err := conn.Send("ZADD", args...); err != nil {
		return errors.Wrapf(err, "ZADD cid(%d) players(%v)", cid, players)
	}
	if err := conn.Send("EXPIRE", d.rankExpire); err != nil {
		return errors.Wrapf(err, "EX cid(%d) players(%v)", cid, players)
	}
	return conn.Flush()
}

func (d *dao) CachePlayersRank(c context.Context, cid int64, mids []int64) (map[int64]int, error) {
	var (
		conn = d.redis.Get(c)
		key  = cacheRankKey(cid)
	)
	defer conn.Close()
	for _, mid := range mids {
		if err := conn.Send("ZREVRANK", key, mid); err != nil {
			return nil, err
		}
	}
	if err := conn.Flush(); err != nil {
		log.Error("CachePlayersRank conn.Send(ZRANK, %s) error(%v)", key, err)
		return nil, err
	}
	var res = make(map[int64]int)
	for _, mid := range mids {
		reply, err := conn.Receive()
		if err != nil || reply == nil {
			res[mid] = -1
			continue
		}
		rank, _ := redis.Int(reply, nil)
		res[mid] = rank
	}
	return res, nil
}

func (d *dao) CachePlayerScore(c context.Context, cid, mid int64) (int, error) {
	var (
		conn = d.redis.Get(c)
		key  = cacheRankKey(cid)
	)
	defer conn.Close()
	reply, err := redis.Int(conn.Do("ZSCORE", key, mid))
	if err != nil {
		return 0, errors.Wrapf(err, "CachePlayerScore cid(%d) mid(%d)", cid, mid)
	}
	return reply, nil
}
