package knowledge

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
)

const (
	_userKnowledgeTask = "knowledge_task_%d_%s"
)

func keyCacheKnowTask(mid int64, table string) string {
	return fmt.Sprintf(_userKnowledgeTask, mid, table)
}

func (d *Dao) CacheUserKnowledgeTask(ctx context.Context, mid int64, table string) (res map[string]int64, err error) {
	key := keyCacheKnowTask(mid, table)
	if err = d.mc.Get(ctx, key).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
			res = nil
		} else {
			log.Errorc(ctx, "CacheUserKnowledgeTask conn.Get error(%v)", err)
		}
	}
	return
}

func (d *Dao) AddCacheUserKnowledgeTask(ctx context.Context, mid int64, table string, value map[string]int64) (err error) {
	key := keyCacheKnowTask(mid, table)
	if err = d.mc.Set(ctx, &memcache.Item{
		Key:        key,
		Object:     value,
		Expiration: 300,
		Flags:      memcache.FlagJSON,
	}); err != nil {
		log.Error("AddCacheUserKnowledgeTask(%d) value(%v) error(%v)", key, value, err)
		return
	}
	return
}

func (d *Dao) DelCacheUserKnowledgeTask(ctx context.Context, mid int64, table string) (err error) {
	key := keyCacheKnowTask(mid, table)
	if err = retry.WithAttempts(ctx, "user_knowledge_task_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		if err = d.mc.Delete(ctx, key); err == memcache.ErrNotFound {
			return nil
		}
		return err
	}); err != nil {
		log.Error("DelCacheUserKnowledgeTask(%d) error(%v)", key, err)
		return err
	}
	return
}

func (d *Dao) UserKnowledgeTask(ctx context.Context, mid int64, table string) (res map[string]int64, err error) {
	if res, err = d.CacheUserKnowledgeTask(ctx, mid, table); err != nil {
		if err != nil {
			log.Errorc(ctx, "UserKnowledgeTask d.CacheUserKnowledgeTask() mid(%d) error(%+v)", mid, err)
			err = nil
		}
	}
	if len(res) > 0 {
		return
	}
	if res, err = d.RawUserKnowledgeTask(ctx, mid, table); err != nil {
		log.Errorc(ctx, "UserKnowledgeTask d.RawUserKnowledgeTask() mid(%d) error(%+v)", mid, err)
		return
	}
	d.cache.Do(ctx, func(c context.Context) {
		d.AddCacheUserKnowledgeTask(c, mid, table, res)
	})
	return
}
