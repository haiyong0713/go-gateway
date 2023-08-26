package dao

import (
	"fmt"
	"runtime"

	"go-common/library/cache/redis"
	"go-common/library/database/sql"
	"go-common/library/log"

	vasGrpc "git.bilibili.co/bapis/bapis-go/vas/trans/service"

	arcapi "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/ugc-season/service/conf"
)

// Dao is ugc-season dao
type Dao struct {
	c *conf.Config
	// db
	season *sql.DB
	stat   *sql.DB
	redis  *redis.Pool
	// cache chan
	cacheCh   chan func()
	ArcClient arcapi.ArchiveClient
	vasGRPC   vasGrpc.VasTransServiceClient
}

// New new a Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:       c,
		season:  sql.NewMySQL(c.SeasonDB),
		stat:    sql.NewMySQL(c.StatDB),
		redis:   redis.NewPool(c.Redis),
		cacheCh: make(chan func(), 1024),
	}
	var err error
	if d.ArcClient, err = arcapi.NewClient(c.ArcClient); err != nil {
		panic(fmt.Sprintf("archive GRPC error(%+v)!!!!!!!!!!!!!!!!!!!!!!", err))
	}
	if d.vasGRPC, err = vasGrpc.NewClientVasTransService(c.VasGRPC); err != nil {
		panic(fmt.Sprintf("vasGrpc.NewClientVasTransService error (%+v)", err))
	}
	// nolint:biligowordcheck
	for i := 0; i < runtime.NumCPU(); i++ {
		go d.cacheproc()
	}
	return
}

func (d *Dao) addCache(f func()) {
	select {
	case d.cacheCh <- f:
	default:
		log.Warn("d.cacheCh is full")
	}
}

func (d *Dao) cacheproc() {
	for {
		f, ok := <-d.cacheCh
		if !ok {
			return
		}
		f()
	}
}

// Close close resource.
func (d *Dao) Close() {
	d.season.Close()
	d.stat.Close()
	d.redis.Close()
}
