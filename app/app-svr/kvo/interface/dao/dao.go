package dao

import (
	"context"
	"time"

	"go-gateway/app/app-svr/kvo/interface/conf"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/queue/databus"
)

// Dao kvo data access obj with bfs
type Dao struct {
	rds            *redis.Redis
	rdsExpire      int32
	rdsIncrExpire  int32
	rdsUcDocExpire int32
	// http client for bfs req
	db *sql.DB

	taskPub      *databus.Databus
	buvidTaskPub *databus.Databus

	getUserConf *sql.Stmt
}

// New new data access
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		rds:            redis.NewRedis(c.Redis.Redis),
		rdsExpire:      int32(time.Duration(c.Redis.Expire) / time.Second),
		rdsUcDocExpire: int32(time.Duration(c.Redis.UcDocExpire) / time.Second),
		rdsIncrExpire:  int32(time.Duration(c.Redis.IncrExpire) / time.Second),

		db:           sql.NewMySQL(c.Mysql),
		taskPub:      databus.New(c.TaskPub),
		buvidTaskPub: databus.New(c.BuvidTaskPub),
	}
	d.getUserConf = d.db.Prepared(_getUserConf)
	return
}

// Ping check if health
func (d *Dao) Ping(ctx context.Context) (err error) {
	if err = d.pingRedis(ctx); err != nil {
		return
	}
	if err = d.db.Ping(ctx); err != nil {
		return
	}
	return
}

// BeginTx begin trans
func (d *Dao) BeginTx(c context.Context) (*sql.Tx, error) {
	return d.db.Begin(c)
}
