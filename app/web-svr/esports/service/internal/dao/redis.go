package dao

import (
	"bytes"
	"context"
	"math/rand"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/log"
)

const (
	// 赛程编辑锁
	ContestEditRedisLock    = "esports:contest:edit:lock:sid:%d:cid:%d:externalId:%d"
	ContestEditRedisLockTTL = 60
	// 赛季
	seasonInfoCache    = "esports:season:redis:id:%d"
	seasonInfoCacheTTL = 300

	// 游戏列表
	gameInfoMapCache = "esports:games:redis:cache:map"
	// 赛事列表
	matchInfoMapCache = "esports:matches:redis:cache:map"

	// 战队列表
	teamInfoMapCache = "esports:team:redis:cache:id:%d"

	// 赛程缓存
	contestInfoCache    = "esports:contestModel:redis:cache:id:%d"
	contestInfoCacheTTL = 300

	contestSeriesInfoCache    = "esports:contest_series:redis:cache:id:%d"
	contestSeriesInfoCacheTTL = 600

	seasonContestSeriesListCache    = "esports:contest_series:redis:cache:season_id:%d"
	seasonContestSeriesListCacheTTL = 120

	esContestCache    = "esports:es:contestIdsTotal:md5:%s"
	esContestCacheTTL = 5
)

func (d *dao) RedisUniqueValue() string {
	baseValue := time.Now().UnixNano()
	random := rand.New(rand.NewSource(baseValue)).Intn(1000000)
	var buffer bytes.Buffer
	buffer.WriteString(strconv.FormatInt(baseValue, 10))
	buffer.WriteString(strconv.Itoa(random))
	return buffer.String()
}

func (d *dao) RedisLock(ctx context.Context, key string, value string, ttl int64, retry int, internalMillSeconds int64) (err error) {
	for {
		_, err = d.redis.Do(ctx, "set", key, value, "EX", ttl, "NX")
		if err != nil {
			if err == redis.ErrNil {
				log.Warnc(ctx, "[Dao][RedisLock][AlreadyExist], key:%s, err: %+v", key, err)
			} else {
				log.Errorc(ctx, "[Dao][RedisLock][Error], key:%s, err: %+v", key, err)
			}
		} else {
			break
		}
		if retry <= 0 {
			break
		}
		retry--
	}
	return
}

func (d *dao) RedisUnLock(ctx context.Context, key string, value string) (err error) {
	// 此处考虑到场景不太复杂，实现简单版本的redLock
	bytes, err := redis.Bytes(d.redis.Do(ctx, "get", key))
	if err != nil {
		log.Errorc(ctx, "[Dao][RedisUnLock][Get][Error], err:%+v", err)
		return
	}
	if string(bytes) == value {
		_, err = d.redis.Do(ctx, "del", key)
		if err != nil {
			log.Errorc(ctx, "[Dao][Redis][RedisUnLock][Del][Error], err:%+v", err)
			return
		}
	}
	return
}

func (d *dao) connClose(ctx context.Context, conn redis.Conn) {
	err := conn.Close()
	if err != nil {
		log.Errorc(ctx, "[Dao][Redis][Conn][Close][Error], err:%+v", err)
	}
}
