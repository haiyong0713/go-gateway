package timemachine

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/component"
	"sync/atomic"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/env"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/conf"

	"go-common/library/sync/errgroup.v2"

	"go-common/library/database/hbase.v2"
)

type Dao struct {
	c                        *conf.Config
	redis                    *redis.Redis
	hbase                    *hbase.Client
	tmProcStart              int64
	tmProcStop               int64
	UserYearReport2020Expire int32
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c:                        c,
		redis:                    component.TimeMachineRedis,
		hbase:                    hbase.NewClient(c.Hbase),
		UserYearReport2020Expire: int32(time.Duration(c.Redis.UserYearReport2020Expire) / time.Second),
	}
	go d.startTmproc(context.Background())
	return d
}

func (d *Dao) startTmproc(c context.Context) {
	if env.DeployEnv != env.DeployEnvPre {
		return
	}
	for {
		time.Sleep(time.Second)
		if atomic.LoadInt64(&d.tmProcStart) != 0 {
			go func() {
				// scan key
				max := 10000000000
				step := max / 10000
				prefix := step - 1
				group := errgroup.WithContext(c)
				group.GOMAXPROCS(15)
				for i := 0; i < max; i += step {
					startRow := fmt.Sprintf("%0*d", 10, i)
					endRow := fmt.Sprintf("%0*d", 10, i+prefix)
					group.Go(func(ctx context.Context) error {
						if err := d.timemachineScan(ctx, startRow, endRow); err != nil {
							log.Error("startTmproc timemachineScan startRow(%s) endRow(%s) error(%v)", startRow, endRow, err)
							return nil
						}
						log.Info("startTmproc finish startRow(%s) endRow(%s)", startRow, endRow)
						return nil
					})
				}
				group.Wait()
			}()
			break
		}
	}
}

// StartTmProc start time machine proc.
func (d *Dao) StartTmProc() {
	atomic.StoreInt64(&d.tmProcStart, 1)
}

func (d *Dao) StopTmproc() {
	atomic.StoreInt64(&d.tmProcStop, 1)
}

// Close .
func (d *Dao) Close() {
	if d.hbase != nil {
		d.hbase.Close()
	}
}
