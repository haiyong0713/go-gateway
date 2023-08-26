package guess

import (
	"context"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/net/http/blademaster"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/dao"
)

// Dao  dao
type Dao struct {
	db                                              *sql.DB
	redis, guRedis                                  *redis.Pool
	client                                          *blademaster.Client
	mcCourse                                        *memcache.Memcache
	eSportsKey, contestListKey, guessMainDetailsKey string
	imKeyURL, imSendURL, contestsURL                string
	userExpire, listExpire                          int32
}

// New init
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		db:                  sql.NewMySQL(c.MySQL.Like),
		redis:               redis.NewPool(c.Redis.Config),
		guRedis:             redis.NewPool(c.GuRedis.Config),
		mcCourse:            memcache.New(c.S10MC),
		eSportsKey:          c.S10CacheKey.ESportsKey,
		contestListKey:      c.S10CacheKey.ContestListKey,
		guessMainDetailsKey: c.S10CacheKey.GuessMainDetailsKey,
		client:              blademaster.NewClient(c.HTTPClient),
		imKeyURL:            c.Host.ApiVcCo + _getMsgKeyPath,
		imSendURL:           c.Host.ApiVcCo + _sendMsgPath,
		contestsURL:         c.Host.APICo + _contestList,
		userExpire:          int32(time.Duration(c.GuRedis.UserExpire) / time.Second),
		listExpire:          int32(time.Duration(c.GuRedis.ListExpire) / time.Second),
	}
	return
}

// Close close
func (d *Dao) Close() {
	d.redis.Close()
	d.guRedis.Close()
	d.mcCourse.Close()
	d.db.Close()
}

// Ping ping
func (d *Dao) Ping(c context.Context) error {
	eg := errgroup.Group{}
	eg.Go(func(ctx context.Context) error {
		return d.db.Ping(c)
	})
	eg.Go(func(ctx context.Context) error {
		return dao.GlobalReadDB.Ping(ctx)
	})
	return eg.Wait()
}
