package archive

import (
	"fmt"
	"runtime"

	"go-common/library/cache/redis"
	"go-common/library/conf/env"
	"go-common/library/database/sql"
	"go-common/library/database/taishan"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/rpc/warden"
	"go-common/library/stat/prom"
	"go-common/library/sync/pipeline/fanout"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	suApi "git.bilibili.co/bapis/bapis-go/platform/interface/shorturl"
	vipinfoAPI "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"
	"go-gateway/app/app-svr/archive/service/conf"
	locDao "go-gateway/app/app-svr/archive/service/dao/location"
	mngDao "go-gateway/app/app-svr/archive/service/dao/manager"

	"github.com/robfig/cron"
)

// Dao is archive dao.
type Dao struct {
	c        *conf.Config
	resultDB *sql.DB
	statDB   *sql.DB
	// acc rpc
	acc       accapi.AccountClient
	vipClient vipinfoAPI.VipInfoClient
	suClient  suApi.ShortUrlClient
	// redis
	upRds    *redis.Pool
	arcRds   *redis.Pool
	sArcRds  *redis.Pool
	upExpire int32
	// type cache
	tNamem map[int16]string
	// cache chan
	cacheCh  chan func()
	hitProm  *prom.Prom
	missProm *prom.Prom
	errProm  *prom.Prom
	infoProm *prom.Prom
	// player http client
	playerClient *bm.Client
	Taishan      *Taishan
	shareHost    string
	cron         *cron.Cron
	cache        *fanout.Fanout

	mngdao *mngDao.Dao
	locDao *locDao.Dao
}

type Taishan struct {
	client   taishan.TaishanProxyClient
	tableCfg tableConfig
}

type tableConfig struct {
	Table string
	Token string
}

// New new a Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// db
		resultDB: sql.NewMySQL(c.DB.ArcResult),
		statDB:   sql.NewMySQL(c.DB.Stat),
		// redis
		arcRds:   redis.NewPool(c.ArcRedis),
		upRds:    redis.NewPool(c.Redis.Archive.Config),
		sArcRds:  redis.NewPool(c.Redis.SimpleArc),
		upExpire: c.Redis.Archive.UpRdsExpire,
		// cache chan
		cacheCh:      make(chan func(), 1024),
		hitProm:      prom.CacheHit,
		missProm:     prom.CacheMiss,
		errProm:      prom.BusinessErrCount,
		infoProm:     prom.BusinessInfoCount,
		playerClient: bm.NewClient(c.PlayerClient, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		cron:         cron.New(),
		cache:        fanout.New("cache"),
		mngdao:       mngDao.New(c),
		locDao:       locDao.New(c),
	}
	var err error
	if d.acc, err = accapi.NewClient(c.AccClient); err != nil {
		panic(fmt.Sprintf("account GRPC error(%+v)!!!!!!!!!!!!!!!!!!!!!!", err))
	}
	if d.vipClient, err = vipinfoAPI.NewClient(c.VipClient); err != nil {
		panic(fmt.Sprintf("vipinfoAPI GRPC error(%+v)", err))
	}
	if d.suClient, err = suApi.NewClient(c.StClient); err != nil {
		panic(fmt.Sprintf("suApi GRPC error(%+v)", err))
	}
	zone := env.Zone
	if zone == "" {
		panic("env.Zone is nil")
	}
	t, err := taishan.NewClient(&warden.ClientConfig{Zone: zone})
	if err != nil {
		panic(fmt.Sprintf("taishan.NewClient error(%+v)", err))
	}
	d.Taishan = &Taishan{
		client: t,
		tableCfg: tableConfig{
			Table: c.Taishan.Table,
			Token: c.Taishan.Token,
		},
	}
	for i := 0; i < runtime.NumCPU(); i++ {
		// nolint:biligowordcheck
		go d.cacheproc()
	}
	d.initCronTab()
	d.cron.Start()
	return
}

func (d *Dao) initCronTab() {
	// type cache
	d.loadTypes()
	d.loadShortLink()

	if err := d.cron.AddFunc(d.c.Cron.LoadShortHost, d.loadShortLink); err != nil {
		panic(err)
	}
	if err := d.cron.AddFunc(d.c.Cron.LoadTypes, d.loadTypes); err != nil {
		panic(err)
	}
}

// Close close resource.
func (d *Dao) Close() {
	d.resultDB.Close()
	d.statDB.Close()
	d.arcRds.Close()
	d.upRds.Close()
	d.sArcRds.Close()
	d.cache.Close()
}
