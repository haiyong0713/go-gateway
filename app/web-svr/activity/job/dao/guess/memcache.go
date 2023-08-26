package guess

import (
	"context"
	"fmt"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/stat/prom"

	"go-gateway/app/web-svr/activity/job/model/guess"
)

const (
	_userGuessOid = "user_guess_oid_%d"
)

func keyUserGuessOid(mid int64) string {
	return fmt.Sprintf(_userGuessOid, mid)
}

// ContestList 获取赛程列表.
func (d *Dao) ContestList(c context.Context) (res []int64, err error) {
	if err = d.mcCourse.Get(c, d.eSportsKey).Scan(&res); err != nil {
		if err == memcache.ErrNotFound {
			return nil, nil
		}
		prom.BusinessErrCount.Incr("mc:ContestList")
		log.Error("Memcache ContestList key:%v error:%+v", d.eSportsKey, err)
		return nil, err
	}
	return
}

func (d *Dao) DelCacheUserGuessOid(c context.Context, mid int64) (err error) {
	key := keyUserGuessOid(mid)
	if err = d.mcCourse.Delete(c, key); err != nil {
		if err == memcache.ErrNotFound {
			err = nil
		} else {
			log.Error("DelCacheUserGuessOid(%d) value(%v) error(%v)", key, err)
		}
	}
	return
}

func (d *Dao) SetCacheDetailOption(ctx context.Context, detailOptions map[int64][]*guess.DetailOption) (err error) {
	item := &memcache.Item{Key: d.guessMainDetailsKey, Object: detailOptions, Expiration: 86400, Flags: memcache.FlagJSON}
	if err := d.mcCourse.Set(ctx, item); err != nil {
		log.Errorc(ctx, "SetCacheDetailOption(%s) error(%v)", d.guessMainDetailsKey, err)
	}
	return
}
