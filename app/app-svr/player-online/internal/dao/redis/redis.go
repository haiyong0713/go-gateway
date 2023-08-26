package redis

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-common/library/stat/prom"

	"go-gateway/app/app-svr/player-online/internal/conf"
)

const (
	_onlineCountKey            = "player_online_count_%d_%d"
	_sdmShowKey                = "special_dm_%d_%d_%s"
	_premiereWatchCountKey     = "premiere_count_%d"
	_premiereUserWatchKey      = "premiere_%d_%s"
	_premiereUserWatchLock     = "premiere_%d_%s_lock"
	_premiereRoomStatisticsKey = "premiere_room_%d"
)

type Dao struct {
	c        *conf.Config
	redis    *redis.Redis
	hitProm  *prom.Prom
	missProm *prom.Prom
	errProm  *prom.Prom
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:        c,
		redis:    redis.NewRedis(c.Redis.Online),
		hitProm:  prom.CacheHit,
		missProm: prom.CacheMiss,
		errProm:  prom.BusinessErrCount,
	}
	return
}

func premiereRoomStatisticsKey(aid int64) string {
	return fmt.Sprintf(_premiereRoomStatisticsKey, aid)
}

func premiereCountKey(aid int64) string {
	return fmt.Sprintf(_premiereWatchCountKey, aid)
}

func premiereUserWatchKey(aid int64, buvid string) string {
	return fmt.Sprintf(_premiereUserWatchKey, aid, buvid)
}

func premiereUserWatchLock(aid int64, buvid string) string {
	return fmt.Sprintf(_premiereUserWatchLock, aid, buvid)
}

func (d *Dao) ExistPremiereUserWatch(c context.Context, aid int64, buvid string) (bool, error) {
	_, err := redis.Int64(d.redis.Do(c, "GET", premiereUserWatchKey(aid, buvid)))
	if err == redis.ErrNil {
		return false, err
	}
	if err != nil {
		log.Error("d.ExistPremiereUserWatch error key(%+v) err(%+v)", premiereUserWatchKey(aid, buvid), err)
	}
	return true, nil
}

func (d *Dao) SetPremiereUserWatch(c context.Context, aid int64, buvid string, exp int64) error {
	if _, err := d.redis.Do(c, "SETEX", premiereUserWatchKey(aid, buvid), exp, 1); err != nil {
		log.Error("d.SetPremiereUserWatch error key(%+v) exp(%+v) err(%+v)",
			premiereUserWatchKey(aid, buvid), exp, err)
		return err
	}
	return nil
}

func (d *Dao) IncreasePremiereCountCache(c context.Context, aid int64, buvid string) error {
	uid, _ := uuid.NewUUID()
	if ok, err := d.TryLock(c, premiereUserWatchLock(aid, buvid), uid.String(), 1); !ok {
		log.Error("d.IncreasePremiereCountCache try lock failed key(%+v) uid(%s) err(%+v)",
			premiereUserWatchLock(aid, buvid), uid.String(), err)
		return err
	}
	if _, err := d.redis.Do(c, "INCR", premiereCountKey(aid)); err != nil {
		log.Error("d.IncreasePremiereCountCache INCR error key(%+v) err(%+v)", premiereCountKey(aid), err)
		return err
	}
	if ok := d.UnLock(c, premiereUserWatchLock(aid, buvid), uid.String()); !ok {
		log.Error("d.IncreasePremiereCountCache unlock failed key(%+v) uid(%s)",
			premiereUserWatchLock(aid, buvid), uid.String())
	}
	log.Error("d.IncreasePremiereCountCache aid(%d), buvid(%s), key(%s)", aid, buvid, premiereCountKey(aid))
	return nil
}

func (d *Dao) GetPremiereCountCache(c context.Context, aid int64) (int64, error) {
	cnt, err := redis.Int64(d.redis.Do(c, "GET", premiereCountKey(aid)))
	if err == redis.ErrNil {
		return 0, nil
	}
	if err != nil {
		log.Error("d.GetPremiereCountCache error key(%+v) err(%+v)", premiereCountKey(aid), err)
		return 0, err
	}
	return cnt, nil
}

func (d *Dao) GetRoomStatisticsCache(c context.Context, aid int64) (int64, error) {
	cnt, err := redis.Int64(d.redis.Do(c, "GET", premiereRoomStatisticsKey(aid)))
	if err == redis.ErrNil {
		return 0, err
	}
	if err != nil {
		log.Error("d.GetRoomStatisticsCache error key(%+v) err(%+v)", premiereRoomStatisticsKey(aid), err)
		return 0, err
	}
	return cnt, nil
}

func (d *Dao) SetRoomStatisticsCache(c context.Context, aid int64, exp int64, value int64) error {
	if _, err := d.redis.Do(c, "SETEX", premiereRoomStatisticsKey(aid), exp, value); err != nil {
		log.Error("d.SetRoomStatisticsCache error key(%+v) exp(%+v) value(%+v) err(%+v)",
			premiereRoomStatisticsKey(aid), exp, value, err)
		return err
	}
	return nil
}

func onlineCountKey(aid int64, cid int64) string {
	return fmt.Sprintf(_onlineCountKey, aid, cid)
}

func sdmShowKey(aid int64, cid int64, buvid string) string {
	return fmt.Sprintf(_sdmShowKey, aid, cid, buvid)
}

func (d *Dao) SetOnlineCountCache(c context.Context, aid int64, cid int64, exp int64, value int64) error {
	if _, err := d.redis.Do(c, "SETEX", onlineCountKey(aid, cid), exp, value); err != nil {
		log.Error("d.SetOnlineCountCache error key(%+v) exp(%+v) value(%+v) err(%+v)", onlineCountKey(aid, cid),
			exp, value, err)
		return err
	}
	return nil
}

func (d *Dao) GetOnlineCountCache(c context.Context, aid int64, cid int64) (int64, error) {
	cnt, err := redis.Int64(d.redis.Do(c, "GET", onlineCountKey(aid, cid)))
	if err == redis.ErrNil {
		d.missProm.Incr("online_count")
		return 0, err
	}

	if err != nil {
		d.errProm.Incr("online_count")
		log.Error("d.GetOnlineCountCache error key(%+v) err(%+v)", onlineCountKey(aid, cid), err)
		return 0, err
	}

	d.hitProm.Incr("online_count")
	return cnt, nil
}

func (d *Dao) SetSdmCache(c context.Context, aid int64, cid int64, buvid string, exp int64, value int64) error {
	if _, err := d.redis.Do(c, "SETEX", sdmShowKey(aid, cid, buvid), exp, value); err != nil {
		log.Error("d.SetSdmCache error key(%+v) exp(%+v) value(%+v) err(%+v)",
			sdmShowKey(aid, cid, buvid), exp, value, err)
		return err
	}
	return nil
}

func (d *Dao) GetSdmCache(c context.Context, aid int64, cid int64, buvid string) (int64, error) {
	cnt, err := redis.Int64(d.redis.Do(c, "GET", sdmShowKey(aid, cid, buvid)))
	if err == redis.ErrNil {
		d.missProm.Incr("sdm_show")
		return 0, err
	}

	if err != nil {
		d.errProm.Incr("sdm_show")
		log.Error("d.GetSdmCache error key(%+v) err(%+v)", sdmShowKey(aid, cid, buvid), err)
		return 0, err
	}

	d.hitProm.Incr("sdm_show")
	return cnt, nil
}

func (d *Dao) DelSdmCache(c context.Context, aid int64, cid int64, buvid string) error {
	if _, err := d.redis.Do(c, "DEL", sdmShowKey(aid, cid, buvid)); err != nil {
		log.Error("d.DelSdmCache error key(%+v) err(%+v)", sdmShowKey(aid, cid, buvid), err)
		return err
	}
	return nil
}
