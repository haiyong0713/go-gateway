package like

import (
	"context"
	"fmt"
	"github.com/bluele/gcache"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	xsql "go-common/library/database/sql"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
)

const (
	_lotteryIndex    = "/matsuri/api/mission"
	_lotteryAddTimes = "/matsuri/api/add/times"
	_likeItemURI     = "/activity/likes/list/%d"
	_sourceItemURI   = "/activity/web/view/data/%d"
	_tagsURI         = "/x/internal/tag/archive/multi/tags"

	cacheKey4DeployIDPod = "activity:pod:%v"
)

// Dao struct
type Dao struct {
	c                                   *conf.Config
	db                                  *xsql.DB
	subjectStmt                         *xsql.Stmt
	voteLogStmt                         *xsql.Stmt
	mc                                  *memcache.Memcache
	mcLikeExpire                        int32
	mcLikeIPExpire                      int32
	mcPerpetualExpire                   int32
	mcItemExpire                        int32
	mcSubStatExpire                     int32
	mcViewRankExpire                    int32
	mcSourceItemExpire                  int32
	mcProtocolExpire                    int32
	mcRegularExpire                     int32
	mcActMissionExpire                  int32
	mcLikeActExpire                     int32
	mcUserCheckExpire                   int32
	mcReserveOnlyExpire                 int32
	redis                               *redis.Pool
	redisExpire                         int32
	EsLikesExpire                       int32
	matchExpire                         int32
	followExpire                        int32
	hotDotExpire                        int32
	randomExpire                        int32
	stochasticExpire                    int32
	likeTotalExpire                     int32
	likeMidTotalExpire                  int32
	onGoingActExpire                    int32
	likeTokenExpire                     int32
	awardSubjectExpire                  int32
	subRuleExpire                       int32
	actArcsExpire                       int32
	lightCountExpire                    int32
	answerExpire                        int32
	voteCategoryExpire                  int32
	lotteryIndexURL                     string
	addLotteryTimesURL                  string
	likeItemURL                         string
	sourceItemURL                       string
	tagURL                              string
	articleGiantURL                     string
	articleListsURL                     string
	upArtListURL                        string
	artGiantV4URL                       string
	artResetURL                         string
	mallCouponURL                       string
	resAuditURL                         string
	checkTelURL                         string
	epPlayURL                           string
	ticketAddWishURL                    string
	ticketFavCountURL                   string
	ticketAddFavInnerURL                string
	dynamicCreateURL                    string
	client                              *httpx.Client
	singleClient                        *httpx.Client
	addFavInnerClient                   *httpx.Client
	cacheCh                             chan func()
	cache                               *fanout.Fanout
	es                                  *elastic.Elastic
	gaiaRiskPub                         *databus.Databus
	DynamicArc                          map[int64]bool
	DynamicLive                         map[int64]bool
	UpActReserveRelationInfo4LiveGCache gcache.Cache
	dynamicLotteryAuth                  string
	dynamicLotteryBind                  string
	dynamicLotteryPrizeInfo             string
}

var (
	CommonDao *Dao
)

// New init
func New(c *conf.Config) *Dao {
	if CommonDao != nil {
		return CommonDao
	}

	CommonDao = &Dao{
		c:                                   c,
		db:                                  component.GlobalDB,
		mc:                                  memcache.New(c.Memcache.Like),
		mcLikeExpire:                        int32(time.Duration(c.Memcache.LikeExpire) / time.Second),
		mcLikeIPExpire:                      int32(time.Duration(c.Memcache.LikeIPExpire) / time.Second),
		mcPerpetualExpire:                   int32(time.Duration(c.Memcache.PerpetualExpire) / time.Second),
		mcItemExpire:                        int32(time.Duration(c.Memcache.ItemExpire) / time.Second),
		mcSubStatExpire:                     int32(time.Duration(c.Memcache.SubStatExpire) / time.Second),
		mcViewRankExpire:                    int32(time.Duration(c.Memcache.ViewRankExpire) / time.Second),
		mcSourceItemExpire:                  int32(time.Duration(c.Memcache.SourceItemExpire) / time.Second),
		mcProtocolExpire:                    int32(time.Duration(c.Memcache.ProtocolExpire) / time.Second),
		mcRegularExpire:                     int32(time.Duration(c.Memcache.RegularExpire) / time.Second),
		mcActMissionExpire:                  int32(time.Duration(c.Memcache.ActMissionExpire) / time.Second),
		mcLikeActExpire:                     int32(time.Duration(c.Memcache.LikeActExpire) / time.Second),
		mcUserCheckExpire:                   int32(time.Duration(c.Memcache.UserCheckExpire) / time.Second),
		redis:                               redis.NewPool(c.Redis.Config),
		cacheCh:                             make(chan func(), 1024),
		cache:                               fanout.New("cache", fanout.Worker(1), fanout.Buffer(1024)),
		redisExpire:                         int32(time.Duration(c.Redis.Expire) / time.Second),
		matchExpire:                         int32(time.Duration(c.Redis.MatchExpire) / time.Second),
		followExpire:                        int32(time.Duration(c.Redis.FollowExpire) / time.Second),
		hotDotExpire:                        int32(time.Duration(c.Redis.HotDotExpire) / time.Second),
		randomExpire:                        int32(time.Duration(c.Redis.RandomExpire) / time.Second),
		stochasticExpire:                    int32(time.Duration(c.Redis.StochasticExpire) / time.Second),
		EsLikesExpire:                       int32(time.Duration(c.Redis.EsLikesExpire) / time.Second),
		likeTotalExpire:                     int32(time.Duration(c.Redis.LikeTotalExpire) / time.Second),
		likeMidTotalExpire:                  int32(time.Duration(c.Redis.LikeMidTotalExpire) / time.Second),
		onGoingActExpire:                    int32(time.Duration(c.Redis.OnGoingActivityExpire) / time.Second),
		likeTokenExpire:                     int32(time.Duration(c.Redis.LikeTokenExpire) / time.Second),
		awardSubjectExpire:                  int32(time.Duration(c.Redis.AwardSubjectExpire) / time.Second),
		subRuleExpire:                       int32(time.Duration(c.Redis.SubRuleExpire) / time.Second),
		actArcsExpire:                       int32(time.Duration(c.Redis.ActArcsExpire) / time.Second),
		lightCountExpire:                    int32(time.Duration(c.Redis.LightVideoCountExpire) / time.Second),
		answerExpire:                        int32(time.Duration(c.S10Answer.AnswerActExpire) / time.Second),
		voteCategoryExpire:                  int32(time.Duration(c.Selection.VoteCategoryExpire) / time.Second),
		lotteryIndexURL:                     c.Host.Activity + _lotteryIndex,
		addLotteryTimesURL:                  c.Host.Activity + _lotteryAddTimes,
		likeItemURL:                         c.Host.Activity + _likeItemURI,
		sourceItemURL:                       c.Host.Activity + _sourceItemURI,
		tagURL:                              c.Host.APICo + _tagsURI,
		articleGiantURL:                     c.Host.APICo + _articleGiantURI,
		articleListsURL:                     c.Host.APICo + _articleListURI,
		upArtListURL:                        c.Host.APICo + _upArtListURI,
		artGiantV4URL:                       c.Host.APICo + _articleGiantV4URI,
		artResetURL:                         c.Host.APICo + _articleResetURI,
		mallCouponURL:                       c.Host.Mall + _mallCouponURI,
		resAuditURL:                         c.Host.APICo + _resAuditURI,
		checkTelURL:                         c.Host.APICo + _checkTelURI,
		epPlayURL:                           c.Host.APICo + _epPlayURI,
		ticketAddWishURL:                    c.Host.ShowCo + _ticketAddWishURI,
		ticketFavCountURL:                   c.Host.ShowCo + _ticketFavCountURI,
		ticketAddFavInnerURL:                c.Host.ShowCo + _ticketAddFavInner,
		dynamicCreateURL:                    c.Host.Dynamic + _dynamicCreateURI,
		client:                              httpx.NewClient(c.HTTPClient),
		singleClient:                        httpx.NewClient(c.HTTPClientSingle),
		addFavInnerClient:                   httpx.NewClient(c.HttpClientAddFavInner),
		es:                                  elastic.NewElastic(c.Elastic),
		mcReserveOnlyExpire:                 int32(time.Duration(c.Memcache.McReserveOnlyExpire) / time.Second),
		UpActReserveRelationInfo4LiveGCache: gcache.New(c.UpActReserveRelationInfo4Live.GCacheLen).LRU().Build(),
		dynamicLotteryAuth:                  c.Host.Dynamic + _dynamicAuth,
		dynamicLotteryBind:                  c.Host.Dynamic + _dynamicBind,
		dynamicLotteryPrizeInfo:             c.Host.Dynamic + _dynamicPrizeInfo,
	}
	CommonDao.subjectStmt = CommonDao.db.Prepared(_selSubjectSQL)
	CommonDao.voteLogStmt = CommonDao.db.Prepared(_votLogSQL)
	go CommonDao.cacheproc()

	return CommonDao
}

// make sure that only one pod can lock
func (dao *Dao) IncrActivityPodIndex(deployID string) (index int64, err error) {
	conn := dao.redis.Get(context.Background())
	defer func() {
		_ = conn.Close()
	}()

	cacheKey := fmt.Sprintf(cacheKey4DeployIDPod, deployID)
	reply, incrErr := conn.Do("INCR", cacheKey)
	if incrErr != nil {
		err = incrErr

		return
	}

	if d, ok := reply.(int64); ok {
		index = d
	}

	// set expire as 1 day
	if index == 1 {
		_, _ = conn.Do("EXPIRE", cacheKey, 86400)
	}

	return
}

// CVoteLog chan Vote Log
func (dao *Dao) CVoteLog(c context.Context, sid int64, aid int64, mid int64, stage int64, vote int64) {
	dao.cacheCh <- func() {
		dao.VoteLog(c, sid, aid, mid, stage, vote)
	}
}

// Close Dao
func (dao *Dao) Close() {
	if dao.redis != nil {
		dao.redis.Close()
	}
	if dao.mc != nil {
		dao.mc.Close()
	}
	close(dao.cacheCh)
}

// Ping Dao
func (dao *Dao) Ping(c context.Context) error {
	return dao.db.Ping(c)
}

func (dao *Dao) cacheproc() {
	for {
		f, ok := <-dao.cacheCh
		if !ok {
			return
		}
		f()
	}
}
