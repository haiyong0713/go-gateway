package family

import (
	"time"

	"go-common/library/cache/redis"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
)

type Dao struct {
	redis              *redis.Redis
	qrcodeExpire       int64
	qrcodeStatusExpire int64
	lockExpire         int64
	timelockPwdExpire  int64
}

func NewDao(cfg *conf.Config) *Dao {
	return &Dao{
		redis:              redis.NewRedis(cfg.Redis.Family.Config),
		qrcodeExpire:       int64(time.Duration(cfg.Redis.Family.QrcodeExpire) / time.Second),
		qrcodeStatusExpire: int64(time.Duration(cfg.Redis.Family.QrcodeStatusExpire) / time.Second),
		lockExpire:         int64(time.Duration(cfg.Redis.Family.LockExpire) / time.Second),
		timelockPwdExpire:  int64(time.Duration(cfg.Redis.Family.TimelockPwdExpire) / time.Second),
	}
}
