package dao

import (
	"sync"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/queue/databus"
	xtime "go-common/library/time"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/steins-gate/job/internal/model"
)

// Dao dao.
type Dao struct {
	daoClosed     bool
	db            *sql.DB // db
	redis         *redis.Pool
	arcClient     arcgrpc.ArchiveClient // archive service grpc client
	msgClient     *bm.Client            // http client
	retryCh       chan *model.RetryOp   // retry and waiter
	waiter        *sync.WaitGroup
	steinsGatePub *databus.Databus // databus
	messageHost   string
	lockExpire    int32
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// New new a dao and return.
func New() (dao *Dao) {
	var (
		dc struct {
			SteinsGate *sql.Config
		}
		rds struct {
			SteinsGate struct {
				*redis.Config
				LockExpire xtime.Duration
			}
		}
		grpc struct {
			archive *warden.ClientConfig
		}
		http struct {
			MessageClient *bm.ClientConfig
			Host          struct {
				Message string
			}
		}
		databusCfg struct {
			SteinsGate *databus.Config
		}
		err error
	)
	checkErr(paladin.Get("http.toml").UnmarshalTOML(&http))
	checkErr(paladin.Get("mysql.toml").UnmarshalTOML(&dc))
	checkErr(paladin.Get("redis.toml").UnmarshalTOML(&rds))
	checkErr(paladin.Get("databus.toml").UnmarshalTOML(&databusCfg))
	dao = &Dao{
		// mysql
		db: sql.NewMySQL(dc.SteinsGate),
		// redis
		redis:      redis.NewPool(rds.SteinsGate.Config),
		lockExpire: int32(time.Duration(rds.SteinsGate.LockExpire) / time.Second),
		// http
		msgClient: bm.NewClient(http.MessageClient),
		// retry channel
		retryCh: make(chan *model.RetryOp, 1024),
		// waiter
		waiter: new(sync.WaitGroup),
		// databus
		steinsGatePub: databus.New(databusCfg.SteinsGate),
		//message host
		messageHost: http.Host.Message,
	}
	dao.arcClient, err = arcgrpc.NewClient(grpc.archive)
	checkErr(err)
	dao.waiter.Add(1)
	//nolint:biligowordcheck
	go dao.retryproc()
	return
}

// Close close the resource.
func (d *Dao) Close() {
	d.daoClosed = true
	log.Warn("Dao Closed!")
	time.Sleep(2 * time.Second)
	close(d.retryCh)
	d.steinsGatePub.Close() // close databus
	d.db.Close()
	d.waiter.Wait()

}
