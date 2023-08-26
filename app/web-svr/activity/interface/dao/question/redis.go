package question

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/component"
)

func poolKey(baseID, poolID int64) string {
	return fmt.Sprintf("pool_%d_%d", baseID, poolID)
}

func latestPoolID(baseID int64) string {
	return fmt.Sprintf("pool_latest_%d", baseID)
}

func quesLimitKey(mid, baseID int64, day string) string {
	return fmt.Sprintf("ques_lmt_%d_%d_%s", mid, baseID, day)
}

func (d *Dao) PoolQuestionIDsWithDefault(c context.Context, baseID, nowTs int64, count int) (data map[int64]int64, poolID int64, err error) {
	poolID = nowTs - 10
	data, err = d.PoolQuestionIDs(c, baseID, poolID, count)
	if len(data) == 0 && err == nil {
		// 兜底逻辑
		poolID, err = redis.Int64(component.GlobalRedis.Do(c, "GET", latestPoolID(baseID)))
		if err != nil {
			log.Errorc(c, "PoolQuestionIDsWithDefault conn.Do(GET %v) error(%v)", latestPoolID(baseID), err)
			return nil, poolID, nil
		}
		if poolID <= 0 {
			return nil, poolID, nil
		}
		data, err = d.PoolQuestionIDs(c, baseID, poolID, count)
		return
	}
	return
}

// PoolQuestionIDs pool question ids.
func (d *Dao) PoolQuestionIDs(c context.Context, baseID, poolID int64, count int) (data map[int64]int64, err error) {
	key := poolKey(baseID, poolID)
	args := redis.Args{}.Add(key)
	for i := 1; i <= count; i++ {
		args = args.Add(i)
	}
	values, err := redis.Int64s(component.GlobalRedis.Do(c, "HMGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "PoolQuestionIDs conn.Do(HMGET %v) error(%v)", args, err)
		return
	}
	data = make(map[int64]int64, count)
	for i := 1; i <= count; i++ {
		if id := values[i-1]; id > 0 {
			data[int64(i)] = values[i-1]
		}
	}
	return
}

// PoolIndexQuestionID get pool index question id.
func (d *Dao) PoolIndexQuestionID(c context.Context, baseID, poolID, index int64) (id int64, err error) {
	key := poolKey(baseID, poolID)
	args := redis.Args{}.Add(key).Add(index)
	id, err = redis.Int64(component.GlobalRedis.Do(c, "HGET", args...))
	if err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Errorc(c, "conn.Do(HMGET %v) error(%v)", args, err)
	}
	return
}

// IncrQuesLimit .
func (d *Dao) IncrQuesLimit(c context.Context, mid, baseID int64, day string) (err error) {
	cacheKey := quesLimitKey(mid, baseID, day)
	if _, err = component.GlobalRedis.Do(c, "INCRBY", cacheKey, 1); err != nil {
		log.Errorc(c, "IncrQuesLimit conn.Do(INCRBY) key(%s) error(%v)", cacheKey, err)
	}
	return
}

// QuesLimit .
func (d *Dao) QuesLimit(c context.Context, mid, baseID int64, day string) (count int64, err error) {
	cacheKey := quesLimitKey(mid, baseID, day)
	if count, err = redis.Int64(component.GlobalRedis.Do(c, "GET", cacheKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(c, "QuesLimit redis.Int64(%s) error(%v)", cacheKey, err)
		}
	}
	return
}

func (d *Dao) FilterRepeatedReport(ctx context.Context, mid int64, year int, usedTime int, score int, province, course string) (res bool, err error) {
	redisKey := fmt.Sprintf("activity_gaokao_2021_repeared_report_%d_%d_%d_%d_%s_%s", mid, year, usedTime, score, province, course)
	if res, err = redis.Bool(component.GlobalRedis.Do(ctx, "SETNX", redisKey, 1)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(ctx, "FilterRepeatedReport conn.Do(SETNX(%s)) error(%v)", redisKey, err)
			return
		}
	}
	if res {
		var times int32
		times = 60
		if _, err = redis.Bool(component.GlobalRedis.Do(ctx, "EXPIRE", redisKey, times)); err != nil {
			log.Errorc(ctx, "FilterRepeatedReport conn.Do(EXPIRE, %s, %d) error(%v)", redisKey, times, err)
			return
		}
	}
	log.Infoc(ctx, "FilterRepeatedReport redisKey:%v , result:%v", redisKey, res)
	return
}

const gaoKaoCacheExpire = 5
const rankRedisKeyPre = "activity_gaokao_2021_rank_key"

func (d *Dao) CacheTotalGaokaoCount(ctx context.Context, total int64) (err error) {
	redisKey := fmt.Sprintf("%s_%s", rankRedisKeyPre, "total_count")
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", redisKey, gaoKaoCacheExpire, total); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(ctx, "CacheTotalGaokaoCount conn.Do(SETNX(%s)) error(%v)", redisKey, err)
			return
		}
	}
	log.Infoc(ctx, "CacheTotalGaokaoCount redisKey :%v , total:%v", redisKey, total)
	return
}

func (d *Dao) GetTotalGaokaoCount(ctx context.Context) (total int64, err error) {
	redisKey := fmt.Sprintf("%s_%s", rankRedisKeyPre, "total_count")
	if total, err = redis.Int64(component.GlobalRedis.Do(ctx, "GET", redisKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Error("GetTotalGaokaoCount redis.Int64(%s) error(%v)", redisKey, err)
		}
	}
	log.Infoc(ctx, "GetTotalGaokaoCount redisKey :%v , total:%v", redisKey, total)
	return
}

func (d *Dao) CacheRankByScore(ctx context.Context, score int, province, course string, count int64) (err error) {
	redisKey := fmt.Sprintf("%s_%s_%s_%s_%d", rankRedisKeyPre, "socre_rank", province, course, score)
	if _, err = component.GlobalRedis.Do(ctx, "SETEX", redisKey, gaoKaoCacheExpire, count); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(ctx, "CacheRankByScore conn.Do(SETNX(%s)) error(%v)", redisKey, err)
			return
		}
	}
	log.Infoc(ctx, "CacheRankByScore redisKey :%v , count:%v", redisKey, count)
	return
}

func (d *Dao) GetRankByScore(ctx context.Context, score int, province, course string) (count int64, err error) {
	redisKey := fmt.Sprintf("%s_%s_%s_%s_%d", rankRedisKeyPre, "socre_rank", province, course, score)
	if count, err = redis.Int64(component.GlobalRedis.Do(ctx, "GET", redisKey)); err != nil {
		if err == redis.ErrNil {
			err = nil
		} else {
			log.Errorc(ctx, "GetRankByScore conn.Do(SETNX(%s)) error(%v)", redisKey, err)
			return
		}
	}
	log.Infoc(ctx, "GetRankByScore redisKey :%v , count:%v", redisKey, count)
	return
}
