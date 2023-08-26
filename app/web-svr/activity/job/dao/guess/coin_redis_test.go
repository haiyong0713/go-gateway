package guess

import (
	"context"
	"testing"
	xtime "time"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/time"
)

var (
	tmpDao *Dao
)

func init() {
	tmpDao = new(Dao)
}

func TestCoinBiz(t *testing.T) {
	newCfg := &redis.Config{
		Name:  "local",
		Proto: "tcp",
		Addr:  "127.0.0.1:6379",
		Config: &pool.Config{
			IdleTimeout: time.Duration(10 * xtime.Second),
			Idle:        2,
			Active:      8,
		},
		WriteTimeout: time.Duration(10 * xtime.Second),
		ReadTimeout:  time.Duration(10 * xtime.Second),
		DialTimeout:  time.Duration(10 * xtime.Second),
	}

	tmpDao.redis = redis.NewPool(newCfg)
	tmpDao.guRedis = redis.NewPool(newCfg)
	t.Run("batch add user guess list", batchAddList)
	t.Run("reset userCoin", resetUserCoin)
	t.Run("delete user guess log cache", deleteUserGuessLogCache)
}

func deleteUserGuessLogCache(t *testing.T) {
	if err := tmpDao.DeleteUserGuessLogCache(context.Background(), 88888888, 1); err != nil {
		t.Error(err)
	}
}

func resetUserCoin(t *testing.T) {
	if err := tmpDao.ResetUserCoinCache(context.Background(), 88888888, 88888888); err != nil {
		t.Error(err)
	}
}

func batchAddList(t *testing.T) {
	m := make(map[int64]float64, 0)
	{
		m[6] = 6
		m[66] = 66
	}
	if err := tmpDao.BatchAddUserListCache(context.Background(), 88888888, m); err != nil {
		t.Error(err)
	}
}
