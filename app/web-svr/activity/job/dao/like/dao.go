package like

import (
	"context"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	"go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/job/conf"
)

const (
	_activity                = "activity"
	_sourceItemURI           = "/activity/web/view/data/%d"
	_selResetURI             = "/x/internal/activity/productrole/reset"
	_createDynamic           = "/dynamic_svr/v0/dynamic_svr/icreate_draw"
	_deleteDynamic           = "/dynamic_svr/v0/dynamic_svr/irm_dynamic"
	_syncActDomain           = "/x/admin/activity/domain/sync"
	_dynamicLotteryPrizeInfo = "/lottery_svr/v1/lottery_svr/lottery_notice"
)

// Dao  dao
type Dao struct {
	c                       *conf.Config
	db                      *sql.DB
	subjectStmt             *sql.Stmt
	inOnlineLog             *sql.Stmt
	mcLike                  *memcache.Memcache
	mcLikeExpire            int32
	redis                   *redis.Pool
	redisNew                *redis.Pool
	redisExpire             int32
	lotteryExpire           int32
	imageUpExpire           int32
	stupidExpire            int32
	taskStateExpire         int32
	httpClient              *blademaster.Client
	httpFate                *blademaster.Client
	httpClientBFS           *blademaster.Client
	es                      *elastic.Elastic
	setObjStatURL           string
	setViewRankURL          string
	setLikeContentURL       string
	addLotteryTimesURL      string
	delLikeURL              string
	preUpURL                string
	preItemUpURL            string
	preSetUpURL             string
	preItemSetUpURL         string
	doTaskURL               string
	lotteryAddTimesURL      string
	likeActStateURL         string
	likeOidsInfoURL         string
	subjectUpURL            string
	likeUpURL               string
	likeCtimeURL            string
	delLikeCtimeURL         string
	actLikeReloadURL        string
	fateShowInfoURL         string
	cache                   *fanout.Fanout
	likeHisAddURL           string
	awardSubjectURL         string
	upArcEventURL           string
	sourceItemURL           string
	selectionResetURL       string
	syncActDomainURL        string
	tunnelPub               *databus.Databus
	tunnelGroupPub          *databus.Databus
	dataExpire              int32
	createDynamicURL        string
	deleteDynamicURL        string
	dynamicLotteryPrizeInfo string
	upReservePushPub        *databus.Databus
}

// New init
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                       c,
		db:                      sql.NewMySQL(c.MySQL.Like),
		mcLike:                  memcache.New(c.Memcache.Like),
		mcLikeExpire:            int32(time.Duration(c.Memcache.LikeExpire) / time.Second),
		redis:                   redis.NewPool(c.Redis.Config),
		redisNew:                redis.NewPool(c.Redis.Cache),
		redisExpire:             int32(time.Duration(c.Redis.Expire) / time.Second),
		lotteryExpire:           int32(time.Duration(c.Lottery.LotteryExpire) / time.Second),
		imageUpExpire:           int32(time.Duration(c.Redis.ImgUpExpire) / time.Second),
		stupidExpire:            int32(time.Duration(c.Redis.StupidExpire) / time.Second),
		taskStateExpire:         int32(time.Duration(c.Redis.TaskStateExpire) / time.Second),
		dataExpire:              int32(time.Duration(c.Redis.TaskDataExpire) / time.Second),
		httpClient:              blademaster.NewClient(c.HTTPClient),
		httpFate:                blademaster.NewClient(c.HTTPFate),
		httpClientBFS:           blademaster.NewClient(c.HTTPClientBFS),
		es:                      elastic.NewElastic(c.Elastic),
		setObjStatURL:           c.Host.APICo + _setObjStatURI,
		setViewRankURL:          c.Host.APICo + _setViewRankURI,
		setLikeContentURL:       c.Host.APICo + _setLikeContentURI,
		addLotteryTimesURL:      c.Host.Activity + _addLotteryTimesURI,
		lotteryAddTimesURL:      c.Host.APICo + _goAddLotteryTimesURI,
		delLikeURL:              c.Host.ActCo + _delLikeURI,
		preUpURL:                c.Host.APICo + _preUpURI,
		preItemUpURL:            c.Host.APICo + _preItemUpURI,
		preSetUpURL:             c.Host.APICo + _preSetUpURI,
		preItemSetUpURL:         c.Host.APICo + _preItemSetUpURI,
		doTaskURL:               c.Host.APICo + _doTaskURI,
		likeActStateURL:         c.Host.APICo + _likeActStateURI,
		likeOidsInfoURL:         c.Host.APICo + _likeOidsInfoURI,
		subjectUpURL:            c.Host.APICo + _subjectUpURI,
		likeUpURL:               c.Host.APICo + _likeUpURI,
		likeCtimeURL:            c.Host.APICo + _likeCtimeURI,
		delLikeCtimeURL:         c.Host.APICo + _delLikeCtimeURI,
		actLikeReloadURL:        c.Host.APICo + _actLikeReloadURI,
		fateShowInfoURL:         c.Host.TbApi + _showInfoURL,
		cache:                   fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
		likeHisAddURL:           c.Host.APICo + _upListHisAddURI,
		awardSubjectURL:         c.Host.APICo + _awardSubURI,
		upArcEventURL:           c.Host.APICo + _upArcEventURI,
		sourceItemURL:           c.Host.Activity + _sourceItemURI,
		selectionResetURL:       c.Host.APICo + _selResetURI,
		syncActDomainURL:        c.Host.Manager + _syncActDomain,
		tunnelPub:               initialize.NewDatabusV1(c.TunnelDatabusPub),
		tunnelGroupPub:          initialize.NewDatabusV1(c.TunnelGroupDatabusPub),
		createDynamicURL:        c.Host.Dynamic + _createDynamic,
		deleteDynamicURL:        c.Host.Dynamic + _deleteDynamic,
		dynamicLotteryPrizeInfo: c.Host.Dynamic + _dynamicLotteryPrizeInfo,
		upReservePushPub:        initialize.NewDatabusV1(c.UpReservePushPub),
	}
	d.subjectStmt = d.db.Prepared(_selSubjectSQL)
	d.inOnlineLog = d.db.Prepared(_inOnlineLogSQL)
	return
}

// Close close
func (d *Dao) Close() {
	d.db.Close()
}

// Ping ping
func (d *Dao) Ping(c context.Context) error {
	return d.db.Ping(c)
}
