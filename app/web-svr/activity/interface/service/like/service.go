package like

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go-common/library/conf/env"

	"github.com/bluele/gcache"

	"go-gateway/app/web-svr/activity/tools/lib/initialize"

	"go-common/library/database/bfs"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"
	xtime "go-common/library/time"
	act "go-gateway/app/web-svr/activity/interface/dao/actplat"

	filterdao "go-gateway/app/web-svr/activity/interface/dao/filter"
	silverdao "go-gateway/app/web-svr/activity/interface/dao/silverbullet"
	"go-gateway/app/web-svr/activity/interface/tool"

	tagrpcBapis "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	arccli "go-gateway/app/app-svr/archive/service/api"
	suitapi "go-main/app/account/usersuit/service/api"

	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	upapi "git.bilibili.co/bapis/bapis-go/archive/service/up"
	coinapi "git.bilibili.co/bapis/bapis-go/community/service/coin"
	thumbupapi "git.bilibili.co/bapis/bapis-go/community/service/thumbup"

	"go-common/library/sync/errgroup.v2"

	steinapi "go-gateway/app/app-svr/steins-gate/service/api"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/bnj"
	"go-gateway/app/web-svr/activity/interface/dao/bws"
	"go-gateway/app/web-svr/activity/interface/dao/bwsonline"
	"go-gateway/app/web-svr/activity/interface/dao/cards"
	"go-gateway/app/web-svr/activity/interface/dao/currency"
	"go-gateway/app/web-svr/activity/interface/dao/dynamic"
	"go-gateway/app/web-svr/activity/interface/dao/favorite"
	"go-gateway/app/web-svr/activity/interface/dao/guess"
	"go-gateway/app/web-svr/activity/interface/dao/like"
	"go-gateway/app/web-svr/activity/interface/dao/lottery"
	lotteryV2 "go-gateway/app/web-svr/activity/interface/dao/lottery_v2"
	pre "go-gateway/app/web-svr/activity/interface/dao/prediction"
	"go-gateway/app/web-svr/activity/interface/dao/question"
	rankV3 "go-gateway/app/web-svr/activity/interface/dao/rank_v3"
	"go-gateway/app/web-svr/activity/interface/dao/springfestival2021"
	tagmdl "go-gateway/app/web-svr/activity/interface/dao/tag"
	"go-gateway/app/web-svr/activity/interface/dao/task"
	bnjmdl "go-gateway/app/web-svr/activity/interface/model/bnj"
	curmdl "go-gateway/app/web-svr/activity/interface/model/currency"
	l "go-gateway/app/web-svr/activity/interface/model/like"
	lottmdl "go-gateway/app/web-svr/activity/interface/model/lottery"
	quesmdl "go-gateway/app/web-svr/activity/interface/model/question"
	"go-gateway/app/web-svr/activity/interface/service/archive"
	lotteryApi "go-gateway/app/web-svr/activity/interface/service/lottery"

	figapi "git.bilibili.co/bapis/bapis-go/account/service/figure"
	accRelation "git.bilibili.co/bapis/bapis-go/account/service/relation"
	spyapi "git.bilibili.co/bapis/bapis-go/account/service/spy"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	bbqtaskapi "git.bilibili.co/bapis/bapis-go/bbq/task"
	cheeseapi "git.bilibili.co/bapis/bapis-go/cheese/service/coupon"
	locationapi "git.bilibili.co/bapis/bapis-go/community/service/location"
	passapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	pgcact "git.bilibili.co/bapis/bapis-go/pgc/service/activity"
	silapi "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
	vipactapi "git.bilibili.co/bapis/bapis-go/vip/service/activity"
	vipapi "git.bilibili.co/bapis/bapis-go/vip/service/vipinfo"

	"github.com/robfig/cron"
)

const (
	_yes           = 1
	_no            = 0
	_typeAll       = "all"
	_typeRegion    = "region"
	_like          = "like"
	_grade         = "grade"
	_vote          = "vote"
	_silenceForbid = 1
	_ios           = 1
	_android       = 2
	_video         = 1
	_pic           = 2
	_drawyoo       = 3
	_article       = 4
	_music         = 5
	_videoAll      = 6
	_videoUp       = 7
	_typeDomain    = "domain_list"
)

// Service struct
type Service struct {
	c                     *conf.Config
	dao                   *like.Dao
	bnjDao                *bnj.Dao
	preDao                *pre.Dao
	taskDao               *task.Dao
	currDao               *currency.Dao
	bwsDao                *bws.Dao
	quesDao               *question.Dao
	dynamicDao            *dynamic.Dao
	guessDao              *guess.Dao
	lottDao               *lottery.Dao
	lottV2Dao             lotteryV2.Dao
	favDao                *favorite.Dao
	tagDao                *tagmdl.Dao
	bwsOnlineDao          *bwsonline.Dao
	springfestival2021Dao *springfestival2021.Dao
	rankv3Dao             *rankV3.Dao
	cardsDao              *cards.Dao
	silverDao             *silverdao.Dao
	filterDao             *filterdao.Dao
	tagGRPC               tagrpcBapis.TagRPCClient
	relClient             relationapi.RelationClient
	bfs                   *bfs.BFS
	accClient             accapi.AccountClient
	coinClient            coinapi.CoinClient
	thumbupClient         thumbupapi.ThumbupClient
	suitClient            suitapi.UsersuitClient
	upClient              upapi.UpClient
	artClient             artapi.ArticleGRPCClient
	spyClient             spyapi.SpyClient
	passportClient        passapi.PassportUserClient
	steinClient           steinapi.SteinsGateClient
	cheeseClient          cheeseapi.CouponClient
	figureClient          figapi.FigureClient
	locationClient        locationapi.LocationClient
	vipActClient          vipactapi.VipActivityClient
	silverClient          silapi.SilverbulletProxyClient
	vipInfoClient         vipapi.VipInfoClient
	bbqtaskClient         bbqtaskapi.TaskClient
	cache                 *fanout.Fanout
	cron                  *cron.Cron
	arcType               map[int32]*arccli.Tp
	dialectTags           map[int64]struct{}
	dialectRegions        map[int32]struct{}
	reward                map[int]*bnjmdl.Reward
	r                     *rand.Rand
	newestSubTs           int64
	questionBase          map[string]*quesmdl.Base
	awardConf             map[int64]*conf.AwardRule
	starConf              map[int64]*conf.Star
	//arcListData           *l.ArcListData
	//scholarshipArcData    *l.ArcListData
	//springCardArcData     *l.ArcListData
	certificateData  *curmdl.CertificateMsg
	taafData         map[string]*l.TaafData
	taafLikes        *l.ListInfo
	steinData        map[string]*l.SteinMemData
	timemachineLikes *l.ListInfo
	entData          map[int64]*l.EntData
	//entDataV2             map[int64]*l.EntDataV2
	resAuditData map[string][]int64
	bdfData      map[int64][]int64
	//shaDArcData           *l.ArcListData
	specialArcData map[int64]*l.SpecialArcList
	//restartArcData        *l.ArcListData
	//yellowGreenArcData    *l.ArcListData
	//mobileGameArcData     *l.ArcListData
	clockInSubIDs []int64
	giantLikes    map[int64]*l.Item
	//stupidArcData         *l.ArcListData
	gameHolidayArcData               *l.ArcListData
	totalRankArcData                 *l.ArcListData
	daySelectArcData                 *l.ArcListData
	doubl11VidelArcData              *l.ArcListData
	doubl11ChannelArcData            *l.ArcListData
	timeMachineArcData               *l.ArcListData
	internalLottSids                 map[string]struct{}
	wxLotteryGift                    *lottmdl.WxLotteryGiftRes
	internalQuestionSids             map[int64]struct{}
	lotterySvr                       *lotteryApi.Service
	articleDayAwardMap               *atomic.Value
	contributionAwards               []*l.ContriAwards
	funnyVideoListArcData            *l.ArcListData
	answerQuestionDetails            map[int64]map[int64]*quesmdl.Detail
	hourPeopleInfo                   *l.HourPeople
	answerRank                       []*l.UserRank
	archive                          *archive.Service
	pgcActClient                     pgcact.ActivityClient
	HotActRelationInfoStore          map[int64]*l.ActRelationInfo
	HotActSubjectInfoStore           map[int64]*l.SubjectItem
	HotActSubjectReserveIDsInfoStore map[int64]int64
	knowledge                        map[int64][]*l.LIDWithVote
	accRelation                      accRelation.RelationClient
	reserveVideoSourceTags           map[int64]*l.ActVideoSourceRelationReserve
	webViewData                      map[int64][]*l.WebData
	gCache                           gcache.Cache
	actDao                           *act.Dao
}

var (
	AsyncProducer    *databus.Databus
	AsyncConsumerCfg *databus.Config

	consumerExitChan chan int
	once             sync.Once

	CommonLikeService *Service
)

func init() {
	consumerExitChan = make(chan int, 1)
}

// New Service
func New(c *conf.Config) (s *Service) {
	if CommonLikeService != nil {
		return CommonLikeService
	}

	s = &Service{
		c:                     c,
		cache:                 fanout.New("like_service_cache", fanout.Worker(1), fanout.Buffer(1024)),
		dao:                   like.New(c),
		bnjDao:                bnj.New(c),
		preDao:                pre.New(c),
		taskDao:               task.New(c),
		currDao:               currency.New(c),
		dynamicDao:            dynamic.New(c),
		bwsDao:                bws.New(c),
		lottDao:               lottery.New(c),
		lottV2Dao:             lotteryV2.New(c),
		quesDao:               question.New(c),
		guessDao:              guess.New(c),
		tagDao:                tagmdl.New(c),
		silverDao:             silverdao.New(c),
		filterDao:             filterdao.New(c),
		bwsOnlineDao:          bwsonline.New(c),
		r:                     rand.New(rand.NewSource(time.Now().UnixNano())),
		bfs:                   bfs.New(c.BFS),
		cron:                  cron.New(),
		favDao:                favorite.New(c),
		lotterySvr:            lotteryApi.New(c),
		articleDayAwardMap:    new(atomic.Value),
		archive:               archive.New(c),
		springfestival2021Dao: springfestival2021.New(c),
		rankv3Dao:             rankV3.New(c),
		cardsDao:              cards.New(c),
		actDao:                act.New(c),

		gCache: gcache.New(5).LRU().LoaderExpireFunc(func(key interface{}) (value interface{}, duration *time.Duration, err error) {
			if key == _typeDomain {
				ctx := context.Background()
				log.Infoc(ctx, "local cache miss , get data from redis , expire_time:%v", c.ActDomainConf.ExpireSecond)
				value, err = s.dao.HGetAllDomain(ctx)
				expire := time.Duration(c.ActDomainConf.ExpireSecond) * time.Second
				duration = &expire
			}
			return
		}).Build(),
	}

	var err error
	m := make(map[string][]*l.ArticleDayAward, 0)
	s.articleDayAwardMap.Store(m)
	if s.accClient, err = accapi.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.relClient, err = relationapi.NewClient(c.RelClient); err != nil {
		panic(err)
	}
	if s.coinClient, err = coinapi.NewClient(c.CoinClient); err != nil {
		panic(err)
	}
	if s.thumbupClient, err = thumbupapi.NewClient(c.ThumbupClient); err != nil {
		panic(err)
	}
	if s.suitClient, err = suitapi.NewClient(c.SuitClient); err != nil {
		panic(err)
	}
	if s.upClient, err = upapi.NewClient(c.UpClient); err != nil {
		panic(err)
	}
	if s.artClient, err = artapi.NewClient(c.ArtClient); err != nil {
		panic(err)
	}
	if s.spyClient, err = spyapi.NewClient(c.SpyClient); err != nil {
		panic(err)
	}
	if s.passportClient, err = passapi.NewClient(c.PassClient); err != nil {
		panic(err)
	}
	if s.steinClient, err = steinapi.NewClient(c.SteinClient); err != nil {
		panic(err)
	}
	if s.cheeseClient, err = cheeseapi.NewClient(c.CheeseClient); err != nil {
		panic(err)
	}
	if s.figureClient, err = figapi.NewClient(c.Figure); err != nil {
		panic(err)
	}
	if s.locationClient, err = locationapi.NewClient(c.LocationRPC); err != nil {
		panic(err)
	}
	if s.vipActClient, err = vipactapi.NewClient(c.VipActClient); err != nil {
		panic(err)
	}
	if s.silverClient, err = silapi.NewClient(c.SilverClient); err != nil {
		panic(err)
	}
	if s.vipInfoClient, err = vipapi.NewClient(c.VipClient); err != nil {
		panic(err)
	}
	if s.bbqtaskClient, err = bbqtaskapi.NewClient(c.BbqTaskClient); err != nil {
		panic(err)
	}
	if s.accRelation, err = accRelation.NewClient(c.RelationClient); err != nil {
		panic(err)
	}
	if s.pgcActClient, err = pgcact.NewClient(c.PgcActClient); err != nil {
		panic(err)
	}
	if s.tagGRPC, err = tagrpcBapis.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	s.initDialect()
	s.initReward()
	s.initAward()
	s.initStar()
	s.initInternalLottSids()
	s.initInternalQuestionSids()
	s.loadTaafWebData()
	s.loadTaafLikes()
	s.loadSteinWebData()
	s.loadTmLikes()
	s.loadEntWebData()
	s.loadResAuditData()
	//s.loadEntV2WebData()
	//s.springCardArcDataproc()
	//s.loadShaDArcData()
	s.loadSpecialArcData()
	//s.loadRestartArcData()
	//s.loadYellowGreenArcData()
	//s.loadBdfAids()
	s.bdfData = make(map[int64][]int64)
	//s.loadMobileGameArcData()
	s.loadClockInSubIDs()
	s.loadGiantArticles()
	//s.loadStupidArcData()
	s.loadArticleDayAward()
	s.loadGameHolidayArcData()
	//s.loadWxLotteryGift()
	s.loadContributionAwards()
	s.loadTotalRankArcData()
	s.loadDaySelectArcData()
	s.loadDouble11ChannelData()
	s.loadDouble11VideoData()
	s.loadTimeMachineData()
	//s.loadFunnyVideoArcData()
	initialize.Call(s.loadActKnowledgeData)
	s.loadQuestionBaseData()
	// s10答题，依赖loadQuestionBaseData()方法
	s.loadQuestionDetail()
	s.loadAnswerHourPeople()
	s.initSelCategory()
	initialize.Call(s.initInternalGetActRelationInfoFromCacheSetInfoMemory)
	initialize.Call(s.initInternalGetActSubjectInfoFromCacheSetInfoMemory)
	initialize.Call(s.initInternalGetActSubjectReserveIDsInfoFromCacheSetInfoMemory)
	initialize.Call(s.initGetActReservedMapVideoSourceTags)
	s.createCron()

	//go s.loadOperationData()
	go s.ASyncHotMainDetailList(context.Background())
	go s.InternalGetActRelationInfoFromCacheSetInfoMemory()
	go s.InternalGetActSubjectInfoFromCacheSetInfoMemory()
	go s.InternalGetActSubjectReserveIDsInfoFromCacheSetInfoMemory()
	go s.InternalSyncActSubjectRuleIntoMemory()

	go s.GetActReservedMapVideoSourceTags(context.Background())

	// 获取up主发起预约的四种类型数据
	//	initialize.CallC(s.GetUpActReserveWhiteList)

	initialize.Call(s.asyncReserveConsumeProc)

	CommonLikeService = s

	return
}

func (s *Service) asyncReserveConsumeProc() error {
	if d := os.Getenv("DEPLOYMENT_ID"); d != "" && env.Color == conf.Conf.AsyncReserveConfig.Color {
		if !tool.IsBnj2021LiveApplication() && conf.Conf.AsyncReserveConfig.Topic != "" {
			if index, err := s.dao.IncrActivityPodIndex(d); err == nil && index > 0 {
				if index == 1 {
					if conf.Conf.AsyncReserveConfig.Concurrency <= 0 {
						conf.Conf.AsyncReserveConfig.Concurrency = 1
					}
					go s.asyncReserveConsume()
				}
			}
		}
	}
	return nil
}

// Close service
func (s *Service) Close() {
	once.Do(func() {
		s.cron.Stop()
		s.cache.Close()
		s.dao.Close()
		s.preDao.Close()
		s.taskDao.Close()
		s.lottDao.Close()

		// set timeout as 2 seconds, make sure that kafka consumers exit
		//     >>> in order to increase the missing of reserve data
		ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
		close(consumerExitChan)
		select {
		case <-ctx.Done():
			// Do nothing now
		}
	})
}

// Ping service
func (s *Service) Ping(c context.Context) (err error) {
	err = s.dao.Ping(c)
	return
}

func (s *Service) initDialect() {
	tmpTag := make(map[int64]struct{}, len(s.c.Rule.DialectTags))
	for _, v := range s.c.Rule.DialectTags {
		tmpTag[v] = struct{}{}
	}
	tmpRegion := make(map[int32]struct{}, len(s.c.Rule.DialectRegions))
	for _, v := range s.c.Rule.DialectRegions {
		tmpRegion[v] = struct{}{}
	}
	s.dialectTags = tmpTag
	s.dialectRegions = tmpRegion
}

func (s *Service) initReward() {
	tmp := make(map[int]*bnjmdl.Reward, len(s.c.Bnj2019.Reward))
	for _, v := range s.c.Bnj2019.Reward {
		tmp[v.Step] = v
	}
	s.reward = tmp
}

func (s *Service) initAward() {
	tmp := make(map[int64]*conf.AwardRule, len(s.c.Rule.AwardRule))
	for _, v := range s.c.Rule.AwardRule {
		tmp[v.Sid] = v
	}
	s.awardConf = tmp
}

func (s *Service) initStar() {
	tmp := make(map[int64]*conf.Star, len(s.c.Stars))
	for _, v := range s.c.Stars {
		tmp[v.Sid] = v
	}
	s.starConf = tmp
}

func (s *Service) initInternalLottSids() {
	tmp := make(map[string]struct{})
	tmp[s.c.WxLottery.CommonSid] = struct{}{}
	tmp[s.c.WxLottery.NewCashSid] = struct{}{}
	tmp[s.c.WxLottery.VipSid] = struct{}{}
	tmp[s.c.WxLottery.NormalSid] = struct{}{}
	for _, v := range s.c.Rule.InterLottSids {
		tmp[v] = struct{}{}
	}
	s.internalLottSids = tmp
}

func (s *Service) initInternalQuestionSids() {
	tmp := make(map[int64]struct{})
	for _, v := range s.c.Rule.InterQuestionIds {
		tmp[v] = struct{}{}
	}
	s.internalQuestionSids = tmp
}

func (s *Service) createCron() {
	var err error
	if err = s.cron.AddFunc(s.c.Cron.ResAudit, s.loadResAuditData); err != nil {
		panic(err)
	}
	//if err = s.cron.AddFunc(s.c.Cron.SelectArc, s.springCardArcDataproc); err != nil {
	//	panic(err)
	//}
	//if err = s.cron.AddFunc(s.c.Cron.EntV2, s.loadEntV2WebData); err != nil {
	//	panic(err)
	//}
	//if err = s.cron.AddFunc(s.c.Cron.Bdf, s.loadBdfAids); err != nil {
	//	panic(err)
	//}
	//if err = s.cron.AddFunc(s.c.Cron.SelectArc, s.loadShaDArcData); err != nil {
	//	panic(err)
	//}
	if err = s.cron.AddFunc("@every 10m", s.loadNewestSubTs); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 10m", s.loadQuestionBaseData); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 50m", s.loadCertificateData); err != nil {
		panic(err)
	}
	//if err = s.cron.AddFunc("@every 51m", s.loadScholarshipArcData); err != nil {
	//	panic(err)
	//}
	//if err = s.cron.AddFunc("@every 52m", s.loadArchiveList); err != nil {
	//	panic(err)
	//}
	if err = s.cron.AddFunc("@every 53m", s.loadTaafWebData); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 54m", s.loadTaafLikes); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 55m", s.loadSteinWebData); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 56m", s.loadTmLikes); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 57m", s.loadEntWebData); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 10m", s.loadArcType); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 1h", s.loadActSource); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 5m", s.loadArticleDayAward); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.SpecialArc, s.loadSpecialArcData); err != nil {
		panic(err)
	}
	//if err = s.cron.AddFunc(s.c.Cron.SelectArc, s.loadRestartArcData); err != nil {
	//	panic(err)
	//}
	//if err = s.cron.AddFunc(s.c.Cron.SpecialArc, s.loadYellowGreenArcData); err != nil {
	//	panic(err)
	//}
	//if err = s.cron.AddFunc(s.c.Cron.SpecialArc, s.loadMobileGameArcData); err != nil {
	//	panic(err)
	//}
	if err = s.cron.AddFunc(s.c.Cron.SubRule, s.loadClockInSubIDs); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.GiantV4, s.cronGiantArticles); err != nil {
		panic(err)
	}
	//if err = s.cron.AddFunc("@every 11m", s.loadStupidArcData); err != nil {
	//	panic(err)
	//}
	if err = s.cron.AddFunc("@every 9m", s.loadGameHolidayArcData); err != nil {
		panic(err)
	}
	//if err = s.cron.AddFunc(s.c.WxLottery.GiftCron, s.loadWxLotteryGift); err != nil {
	//	panic(err)
	//}
	if err = s.cron.AddFunc(s.c.S10Answer.DetailCron, s.loadQuestionDetail); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 1m", s.loadContributionAwards); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 12m", s.loadTotalRankArcData); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 13m", s.loadDaySelectArcData); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 14m", s.loadDouble11ChannelData); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 15m", s.loadDouble11VideoData); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc("@every 10m", s.loadTimeMachineData); err != nil {
		panic(err)
	}
	//if err = s.cron.AddFunc("@every 1m", s.loadFunnyVideoArcData); err != nil {
	//	panic(err)
	//}
	//if err = s.cron.AddFunc("@every 1s", s.loadActKnowledgeData); err != nil {
	//	panic(err)
	//}
	if err = s.cron.AddFunc("@every 1m", s.loadAnswerHourPeople); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Selection.VoteCategoryCron, s.initProductRole); err != nil {
		panic(err)
	}
	s.cron.Start()
}

func (s *Service) ActListInfo(c context.Context, mid int64, types, platform int, region int, appType int64) (res []*l.SubAndProtocol, err error) {
	var (
		typeIds []int64
		sids    []int64
		nowT    = time.Now().Unix()
	)
	switch types {
	case _video:
		typeIds = l.VIDEOS
	case _pic:
		typeIds = l.PICS
	case _drawyoo:
		typeIds = l.DRAWYOOS
	case _article:
		typeIds = l.ARTICLES
	case _music:
		typeIds = l.MUSICS
	case _videoAll:
		typeIds = l.VIDEOALL
	case _videoUp:
		typeIds = l.VIDEOUP
	default:
		err = ecode.ActivityNotExist
		return
	}
	if types == _video && (platform == _ios || platform == _android) {
		typeIds = append(typeIds, l.PHONEVIDEO)
	}
	if (types == _videoAll || types == _videoUp) && (platform == _ios || platform == _android) {
		if mid > 0 {
			s.cache.Do(c, func(ctx context.Context) {
				s.ClearRetDot(ctx, mid)
			})
		}
	}
	if sids, err = s.dao.ActSubjectsOnGoing(c, typeIds); err != nil {
		log.Error("ActListInfo s.dao.ActSubjectsOnGoing(%v) error(%v)", typeIds, err)
		return
	}
	if len(sids) == 0 {
		log.Warn("ActListInfo no suitable sid")
		return
	}
	subjects := make(map[int64]*l.SubjectItem, len(sids))
	protocols := make(map[int64]*l.ActSubjectProtocol, len(sids))
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if subjects, e = s.dao.ActSubjects(ctx, sids); e != nil {
			log.Error("ActListInfo s.dao.ActSubjects(%v) error(%v)", sids, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if protocols, e = s.dao.ActSubjectProtocols(ctx, sids); e != nil {
			log.Error("ActListInfo s.dao.ActSubjects(%v) error(%v)", sids, e)
		}
		return
	})
	eg.Wait()
	for _, v := range sids {
		if subjects[v] == nil || protocols[v] == nil {
			log.Warn("ActListInfo subject(%v) or protocol(%v) nil", subjects[v], protocols[v])
			continue
		}
		if region != 0 && protocols[v].Types != "" {
			var ok bool
			strSlice := strings.Split(protocols[v].Types, ",")
			for _, val := range strSlice {
				if i, _ := strconv.Atoi(val); i != region {
					continue
				}
				ok = true
				break
			}
			if !ok {
				continue
			}
		}
		if subjects[v].Type == l.CLOCKIN && subjects[v].IsForbidHotList() {
			continue
		}
		if appType > 0 && appType&protocols[v].TagShowPlatform == 0 {
			continue
		}
		tmp := &l.SubAndProtocol{SubjectItem: subjects[v], Protocol: protocols[v]}
		if xtime.Time(nowT) < protocols[v].WeightStime || xtime.Time(nowT) > protocols[v].WeightEtime {
			tmp.Protocol.GlobalWeight = 0
			tmp.Protocol.RegionWeight = 0
		}
		res = append(res, tmp)
	}

	// 获取不包含分区的tag
	if region == -1 {
		noRegionActTags := make([]*l.SubAndProtocol, 0)
		for _, v := range res {
			if v.Protocol.Types == "" && v.Protocol.Tags != "" {
				noRegionActTags = append(noRegionActTags, v)
			}
		}
		res = noRegionActTags
	}

	return
}

// ActsProtocol .
func (s *Service) ActsProtocol(c context.Context, sids []int64) (res map[int64]*l.SubProtocol, err error) {
	var (
		subs  map[int64]*l.SubjectItem
		pros  map[int64]*l.ActSubjectProtocol
		rules map[int64][]*l.SubjectRule
	)
	eg := errgroup.WithContext(c)
	eg.Go(func(ctx context.Context) (e error) {
		if subs, e = s.dao.ActSubjects(c, sids); e != nil {
			log.Error("s.dao.ActSubjects(%v) error(%+v)", sids, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if pros, e = s.dao.ActSubjectProtocols(ctx, sids); e != nil {
			log.Error("s.dao.ActSubjectProtocols(%v) error(%+v)", sids, e)
		}
		return
	})
	eg.Go(func(ctx context.Context) (e error) {
		if rules, e = s.dao.RawSubjectRulesBySids(c, sids); e != nil {
			log.Error("s.dao.RawSubjectRulesBySids(%v) error(%+v)", sids, err)
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	res = make(map[int64]*l.SubProtocol)
	for _, v := range sids {
		if _, ok := subs[v]; !ok {
			continue
		}
		tmp := &l.SubProtocol{SubjectItem: subs[v]}
		if tVal, tok := pros[v]; tok {
			tmp.ActSubjectProtocol = tVal
		}
		if tmp.SubjectItem.Type == l.CLOCKIN {
			if ruleVal, ok := rules[v]; ok {
				tmp.Rules = ruleVal
			}
		}
		res[v] = tmp
	}
	return
}
