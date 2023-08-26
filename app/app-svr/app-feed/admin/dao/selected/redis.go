package selected

import (
	"context"
	"encoding/json"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"

	"github.com/pkg/errors"
)

const (
	_prefixRankKey = "erank_%d"
	_cronJobLock   = "%s_cron_job_lock"
	_selectedSerie = "%d_%s_serie"
)

// 获取分布式锁 true 标识获取到分布式锁
func (d *Dao) GetCronJobLock(ctx context.Context, cronJobName string) (lock bool, err error) {
	key := fmt.Sprintf(_cronJobLock, cronJobName)
	reply, err := d.selRedis.Get(ctx).Do("SET", key, 1, "EX", 60, "NX")
	if err != nil {
		log.Error("bulkLockCache key(%s) err(%v)", key, err)
		return false, err
	}
	if reply == nil {
		return false, err
	}
	log.Warn("key: %s, reply: %+v", key, reply)
	return true, err
}

func (d *Dao) AddRankCache(ctx context.Context, id, rank int) (err error) {
	var (
		key  = fmt.Sprintf(_prefixRankKey, id)
		conn = d.selRedis.Get(ctx)
	)
	defer conn.Close()
	if _, err = conn.Do("SET", key, rank); err != nil {
		log.Error("conn.Do(SET, %s, %d) error(%v)", key, rank, err)
	}
	return
}

func (d *Dao) CacheRankCache(c context.Context, id int) (res int, err error) {
	var (
		key  = fmt.Sprintf(_prefixRankKey, id)
		conn = d.selRedis.Get(c)
	)
	defer conn.Close()
	if res, err = redis.Int(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("conn.Do(GET, %s) error(%v)", key, err)
	}
	return
}

func oneSerieKey(number int64, stype string) string {
	return fmt.Sprintf(_selectedSerie, number, stype)
}

func (d *Dao) PickSerieCache(c context.Context, sType string, number int64) (serie *selected.SerieFull, err error) {
	var (
		key  = oneSerieKey(number, sType)
		conn = d.selRedis.Get(c)
	)
	defer conn.Close()
	values, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil, nil
		}
		return nil, errors.WithMessagef(err, "PickSerieCache redis Get")
	}
	err = json.Unmarshal(values, &serie)
	if err != nil {
		return nil, errors.WithMessagef(err, "PickSerieCache json.Unmarshal error")
	}
	return serie, nil
}
