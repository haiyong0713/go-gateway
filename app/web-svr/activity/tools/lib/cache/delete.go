package cache

import (
	"context"
	"fmt"
	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/log"
)

type Metric4DeleteCache struct {
	Driver interface{}
	Key    string
	Retry  int
	Desc   string
}

const DelCacheMonitor = "[delete_cache_retry_monitor]"

func DeleteCacheRetry(ctx context.Context, metric Metric4DeleteCache) (err error) {
	if metric.Retry == 0 {
		metric.Retry = 3
	}

	if driver, ok := metric.Driver.(*redis.Redis); ok {
		for i := 0; i < metric.Retry; i++ {
			if _, err = driver.Do(ctx, "DEL", metric.Key); err == nil {
				break
			}
		}
		if err != nil {
			err = fmt.Errorf(DelCacheMonitor+"redis command(DEL) key(%s) desc(%s) failed err(%s)", metric.Key, metric.Desc, err.Error())
			log.Errorc(ctx, err.Error())
			return
		}
	}

	if driver, ok := metric.Driver.(*memcache.Memcache); ok {
		for i := 0; i < metric.Retry; i++ {
			if err = driver.Delete(ctx, metric.Key); err == nil || err == memcache.ErrNotFound {
				break
			}
		}
		if err != nil {
			err = fmt.Errorf(DelCacheMonitor+"mc command(delete) key(%s) desc(%s) failed err(%s)", metric.Key, metric.Desc, err.Error())
			log.Errorc(ctx, err.Error())
			return
		}
	}

	err = fmt.Errorf(DelCacheMonitor+"can`t assign driver %+v", metric.Driver)
	log.Errorc(ctx, err.Error())

	return
}
