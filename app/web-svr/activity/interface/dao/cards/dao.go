package cards

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

const (
	springFestivalPrefix = "cards"
	midCard              = "mid_card"
	midCardNew           = "mid_card_new"
	cardConfig           = "card_config"
	midArchive           = "mid_archive"
	midCardToken         = "mid_card"
	tokenKey             = "token"
	midKey               = "mid"
	inviterKey           = "inviter"
	midCompose           = "mid_compose"
	separator            = ":"
	qpsLimitKey          = "qps"
)

// Dao ...
type Dao struct {
	c                 *conf.Config
	redisStore        *redis.Redis
	redisCache        *redis.Redis
	MidCardExpire     int32
	SevenDayExpire    int32
	fiveMinutesExpire int32
	ActivityEndExpire int32
	singleClient      *httpx.Client
	qpsLimitExpire    int32

	// newyear2021DataBusPub *databus.Databus
}

// New ...
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:                 c,
		redisCache:        component.GlobalRedis,
		redisStore:        component.GlobalRedisStore,
		MidCardExpire:     int32(time.Duration(c.Redis.MidCardExpire) / time.Second),
		SevenDayExpire:    int32(time.Duration(c.Redis.SevenDayExpire) / time.Second),
		fiveMinutesExpire: int32(time.Duration(c.Redis.FiveMinutesExpire) / time.Second),
		ActivityEndExpire: int32(time.Duration(c.Redis.SpringFestivialActivityEndExpire) / time.Second),
		singleClient:      httpx.NewClient(c.HTTPClientSingle),
		qpsLimitExpire:    int32(time.Duration(c.Redis.QPSLimitExpire) / time.Second),
		// newyear2021DataBusPub: databus.New(c.DataBus.NewYear2021Pub),
	}
	return d
}

// buildKey ...
func buildKey(args ...interface{}) string {
	strArgs := make([]string, len(args), len(args))
	for i, val := range args {
		strArgs[i] = fmt.Sprint(val)
	}
	return springFestivalPrefix + separator + strings.Join(strArgs, separator)
}

// BeginTran begin transcation.
func (d *Dao) BeginTran(c context.Context) (tx *sql.Tx, err error) {
	return component.GlobalDB.Begin(c)
}
