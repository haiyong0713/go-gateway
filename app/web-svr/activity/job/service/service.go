package service

import (
	"context"
	"encoding/json"
	"go-common/library/net/trace"
	"go-gateway/app/web-svr/activity/job/dao/stock"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/library/conf/env"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	arcapi "go-gateway/app/app-svr/archive/service/api"
	actapi "go-gateway/app/web-svr/activity/interface/api"
	actrpc "go-gateway/app/web-svr/activity/interface/rpc/client"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/dao/bnj"
	"go-gateway/app/web-svr/activity/job/dao/bws"
	"go-gateway/app/web-svr/activity/job/dao/college"
	"go-gateway/app/web-svr/activity/job/dao/dm"
	"go-gateway/app/web-svr/activity/job/dao/dubbing"
	"go-gateway/app/web-svr/activity/job/dao/funny"
	"go-gateway/app/web-svr/activity/job/dao/guess"
	"go-gateway/app/web-svr/activity/job/dao/handwrite"
	"go-gateway/app/web-svr/activity/job/dao/knowledgetask"
	"go-gateway/app/web-svr/activity/job/dao/like"
	"go-gateway/app/web-svr/activity/job/dao/pay"
	"go-gateway/app/web-svr/activity/job/dao/question"
	"go-gateway/app/web-svr/activity/job/dao/rank"
	rankv2 "go-gateway/app/web-svr/activity/job/dao/rank_v2"
	s10dao "go-gateway/app/web-svr/activity/job/dao/s10"
	"go-gateway/app/web-svr/activity/job/dao/share"
	silverdao "go-gateway/app/web-svr/activity/job/dao/silverbullet"
	bnjmdl "go-gateway/app/web-svr/activity/job/model/bnj"
	l "go-gateway/app/web-svr/activity/job/model/like"
	"go-gateway/app/web-svr/activity/job/model/match"
	quesmdl "go-gateway/app/web-svr/activity/job/model/question"
	"go-gateway/app/web-svr/activity/job/model/s10"
	acg "go-gateway/app/web-svr/activity/job/service/2020acg"
	"go-gateway/app/web-svr/activity/job/service/bnj2021"
	fitSvr "go-gateway/app/web-svr/activity/job/service/fit"
	rankSvr "go-gateway/app/web-svr/activity/job/service/rank"
	rankv3Svr "go-gateway/app/web-svr/activity/job/service/rank_v3"
	sourceSvr "go-gateway/app/web-svr/activity/job/service/source"
	"go-gateway/app/web-svr/activity/job/service/wish_2021_spring"
	"go-gateway/app/web-svr/activity/job/tool"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	suitapi "go-main/app/account/usersuit/service/api"
	favoriteapi "go-main/app/community/favorite/service/api"
	tagapi "go-main/app/community/tag/service/api"

	api "git.bilibili.co/bapis/bapis-go/account/service"
	memberAPI "git.bilibili.co/bapis/bapis-go/account/service/member"
	relationapi "git.bilibili.co/bapis/bapis-go/account/service/relation"
	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	bbqtaskapi "git.bilibili.co/bapis/bapis-go/bbq/task"
	cheeseapi "git.bilibili.co/bapis/bapis-go/cheese/service/coupon"
	tagnewapi "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	coinapi "git.bilibili.co/bapis/bapis-go/community/service/coin"
	actplatapi "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	gaialibapi "git.bilibili.co/bapis/bapis-go/silverbullet/gaia/service/lib"
	videoupapi "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"go-common/library/sync/pipeline/fanout"

	"github.com/robfig/cron"
)

const (
	_matchObjTable         = "act_matchs_object"
	_subjectTable          = "act_subject"
	_subjectStatTable      = "subject_stat"
	_subjectRuleTable      = "act_subject_rule"
	_likesTable            = "likes"
	_likeContentTable      = "like_content"
	_likeActionTable       = "like_action"
	_predictionTable       = "prediction"
	_predictionItemTable   = "prediction_item"
	_archiveTable          = "archive"
	_guessMainTable        = "act_guess_main"
	_guessUserTable        = "act_guess_user_"
	_actReserve            = "act_reserve_"
	_actReserveNew         = "act_reserve_new_"
	_wxLotteryPrefix       = "wx_lottery_log_"
	_lotteryActionPrefix   = "act_lottery_action_"
	_lotteryAddtimesPrefix = "act_lottery_addtimes_"
	_taskUserLogPrefix     = "task_user_log_"
	_taskUserStatePrefix   = "task_user_state_"
	_taskStatePrefix       = "task_state_"
	_manuScriptAuditPrefix = "user_commit_manuscript_audit_material_"
	_userCommitTablePrefix = "user_commit_manuscript_tmp_"
	_actBwsReservePrefix   = "act_bws_online_inter_reserve_"
	_objectPieceSize       = 100
	_retryTimes            = 3
	_typeArc               = "archive"
	_typeArt               = "article"
	_sharding              = 10
	_s10UserCostRecord     = "act_s10_user_cost"
	_s10UserGiftRecord     = "act_s10_user_gift"

	_s10UserCostRecordPrefix = "act_s10_user_cost_"
	_upActReserveRelation    = "up_act_reserve_relation"
)

// Service service
type Service struct {
	c           *conf.Config
	dao         *like.Dao
	handWrite   handwrite.Dao
	college     college.Dao
	share       share.Dao
	rank        rank.Dao
	rankv2      rankv2.Dao
	bnj         *bnj.Dao
	dm          *dm.Dao
	dubbing     dubbing.Dao
	question    *question.Dao
	guessDao    *guess.Dao
	bws         *bws.Dao
	pay         *pay.Dao
	funny       funny.Dao
	silverDao   *silverdao.Dao
	rankSvr     *rankSvr.Service
	rankV3Svr   *rankv3Svr.Service
	sourceSvr   *sourceSvr.Service
	FitSvr      *fitSvr.Service
	knowTaskDao *knowledgetask.Dao
	// waiter
	waiter sync.WaitGroup
	// lotteryConsumewaiter
	lotteryConsumewaiter sync.WaitGroup
	closed               bool
	// bts: type, upper
	// arc rpc
	actRPC  *actrpc.Service
	actGRPC actapi.ActivityClient
	//grpc
	accClient      api.AccountClient
	arcClient      arcapi.ArchiveClient
	artClient      artapi.ArticleGRPCClient
	coinClient     coinapi.CoinClient
	suitClient     suitapi.UsersuitClient
	tagClient      tagapi.TagRPCClient
	tagNewClient   tagnewapi.TagRPCClient
	cheeseClient   cheeseapi.CouponClient
	relationClient relationapi.RelationClient
	favoriteClient favoriteapi.FavoriteClient
	actplatClient  actplatapi.ActPlatClient
	memberClient   memberAPI.MemberClient
	bbqtaskClient  bbqtaskapi.TaskClient
	videoupClient  videoupapi.VideoUpOpenClient
	gaialibClient  gaialibapi.LibClient
	// databus
	actSub                 *databus.Databus
	bnjBinlogSub           *databus.Databus
	coinSub                *databus.Databus
	thumbupSub             *databus.Databus
	articlePassSub         *databus.Databus
	archiveSub             *databus.Databus
	liveFollowSub          *databus.Databus
	replySub               *databus.Databus
	lotteryAwardSub        *databus.Databus
	vipLotterySub          *databus.Databus
	vipCardSub             *databus.Databus
	ottVipLotterySub       *databus.Databus
	customizeLotterySub    *databus.Databus
	bnjSub                 *databus.Databus
	bnjAwardSub            *databus.Databus
	subRuleStatSub         *databus.Databus
	reserveNotifyPub       *databus.Databus
	reserveNotifySub       *databus.Databus
	newYearReservePub      *databus.Databus
	rewardsAwardSendingSub *databus.Databus
	archiveBinLogSub       *databus.Databus

	actPlatHistorySub   *databus.Databus // 任务平台notify流
	actPlatHistorySpSub *databus.Databus // 任务平台notify流
	lotteryAddtimesSub  *databus.Databus // 抽奖增加次数流
	upReservePushPub    *databus.Databus // 稿件预约催核销流

	bnjValue         int64
	bnjReservedCount int64
	// channel
	subActionCh     []chan *l.Action
	subActI         int64
	arcActionCh     chan *l.Archive
	lotteryActionch chan *l.LotteryMsg
	bnjAwardch      map[int]chan *bnjmdl.AwardAction

	// actPlatHistoryCh 活动任务平台notify
	actPlatHistoryCh     chan *l.ActPlatHistoryMsg
	actPlatHistoryWaitCh chan struct {
	}
	lotteryAddtimesWaitCh chan struct {
	}
	lotteryAddtimesCh chan *l.LotteryAddTimesMsg

	// cron
	cron *cron.Cron
	// question
	questionBase map[string]*quesmdl.NewBaseItem
	questionRand *rand.Rand
	// live seid
	liveSeids map[int64]struct {
	}
	// lottery passes
	lotteryAdds  map[int64]*conf.LotteryAddRule
	suitAwardIDs map[int64]struct {
	}
	// vip lottery
	vipLotteryIDs map[int64]int64
	// bws
	bwsLotteryRand *rand.Rand
	// guess
	guessSesaon map[int64]struct {
	}
	// ott lottery platform ids
	ottVipPids map[int64]int64
	// staff arcs
	staffArc map[int64]struct {
	}
	// restart2020 arcs
	restartArc map[int64]struct {
	}
	awardSubjectTask map[int64]*l.AwardSubject
	// yellow and green arcs
	yelAndGreenArc map[int64]struct {
	}
	// mobile game arcs
	mobileGameArc map[int64]struct {
	}
	// stupid arcs
	stupidArc map[int64]struct {
	}
	// collegeTabList 校园tab列表
	collegeTabList        map[int64][]int64
	collegeArchiveTopList map[int64][]int64

	lotteryTypeAddTimes map[int][]*l.Lottery

	wxLottery *l.TransInfo
	// running
	wxLotteryLogPageRunning bool
	// handWriteRankRunning
	handWriteRankRunning sync.Mutex
	// handWriteFavRunning
	handWriteFavRunning sync.Mutex
	// handWriteDataRunning
	handWriteDataRunning sync.Mutex
	// remixRunning
	remixRunning sync.Mutex
	// remixDataRunning
	remixDataRunning sync.Mutex
	// gameHolidaySyncRunning
	gameHolidaySyncRunning sync.Mutex
	// contributionSyncRunning
	contributionSyncRunning sync.Mutex
	// funnySyncRunning
	funnySyncRunning    sync.Mutex
	actKnowledgeRunning sync.Mutex

	// s10 course
	contestID map[int64]struct {
	}
	mainID map[int64]struct {
	}
	mainContestID map[int64]int64
	contestMutex  sync.Mutex
	s10Dao        *s10dao.Dao
	s10General    *s10.S10General
	freeFlowSub   *databus.Databus
	gaiaRiskSub   *databus.Databus

	// fanout
	natWorker *fanout.Fanout
	// native_page 子父页面关系
	natRelations map[int64]int64

	// collegeRankRunning
	collegeRankRunning sync.Mutex
	// collegeVersionUpdateRunning
	collegeVersionUpdateRunning sync.Mutex
	// shareURLUpdateRunning
	shareURLUpdateRunning sync.Mutex
	// dubbingRunning
	dubbingRunning sync.Mutex
	// collegeBonusRunning
	collegeBonusRunning sync.Mutex
	// collegeScoreRunning
	collegeScoreRunning sync.Mutex
	// dubbingDataRunning
	dubbingDataRunning sync.Mutex
	AcgSrv             *acg.Service
	// collegeScoreAutoRunning
	collegeScoreAutoRunning sync.Mutex
	// handwrite2021Running
	handwrite2021Running sync.Mutex
	// handwrite2021DataRunning
	handwrite2021DataRunning sync.Mutex
	// rankRunning
	rankRunning           sync.Mutex
	lotteryRailgun        *railgun.Railgun
	isUpActReserveLottery sync.Map
	missionTaskRailgun    *railgun.Railgun
	stockDao              *stock.Dao
}

var (
	Ctx4Worker        context.Context
	CancelFunc4Worker context.CancelFunc
)

func init() {
	Ctx4Worker, CancelFunc4Worker = context.WithCancel(context.Background())
}

// New is archive service implementation.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                   c,
		dao:                 like.New(c),
		handWrite:           handwrite.New(c),
		college:             college.New(c),
		dubbing:             dubbing.New(c),
		rank:                rank.New(c),
		rankv2:              rankv2.New(c),
		dm:                  dm.New(c),
		bnj:                 bnj.New(c),
		question:            question.New(c),
		guessDao:            guess.New(c),
		bws:                 bws.New(c),
		pay:                 pay.New(c),
		share:               share.New(c),
		silverDao:           silverdao.New(c),
		knowTaskDao:         knowledgetask.New(c),
		actRPC:              actrpc.New(c.ActRPC),
		actSub:              initialize.NewDatabusV1(c.ActSub),
		bnjBinlogSub:        initialize.NewDatabusV1(c.BnjMainWebSvrBinlogSub),
		coinSub:             initialize.NewDatabusV1(c.ArticleCoinSub),
		thumbupSub:          initialize.NewDatabusV1(c.ArticleLikeSub),
		articlePassSub:      initialize.NewDatabusV1(c.ArticlePassSub),
		archiveSub:          initialize.NewDatabusV1(c.ArchiveSub),
		liveFollowSub:       initialize.NewDatabusV1(c.LiveFollowSub),
		replySub:            initialize.NewDatabusV1(c.ReplySub),
		lotteryAwardSub:     initialize.NewDatabusV1(c.LotteryAwardSub),
		vipLotterySub:       initialize.NewDatabusV1(c.VipLotterySub),
		vipCardSub:          initialize.NewDatabusV1(c.VipCardSub),
		ottVipLotterySub:    initialize.NewDatabusV1(c.OttVipLotterySub),
		customizeLotterySub: initialize.NewDatabusV1(c.CustomizeLotterySub),
		bnjSub:              initialize.NewDatabusV1(c.BnjSub),
		bnjAwardSub:         initialize.NewDatabusV1(c.BnjAwardSub),
		subRuleStatSub:      initialize.NewDatabusV1(c.SubRuleStatSub),
		reserveNotifyPub:    initialize.NewDatabusV1(c.ReserveNotifyPub),
		reserveNotifySub:    initialize.NewDatabusV1(c.ReserveNotifySub),
		newYearReservePub:   initialize.NewDatabusV1(c.NewYearReservePub),
		actPlatHistorySub:   initialize.NewDatabusV1(c.ActPlatHistorySub),
		actPlatHistorySpSub: initialize.NewDatabusV1(c.ActPlatHistorySpSub),
		lotteryAddtimesSub:  initialize.NewDatabusV1(c.LotteryAddtimesSub),
		archiveBinLogSub:    initialize.NewDatabusV1(c.ArchiveBinLogSub),
		upReservePushPub:    initialize.NewDatabusV1(c.UpReservePushPub),
		//rewardsAwardSendingSub: initialize.NewDatabusV1(c.RewardsAwardSendingSub),
		rewardsAwardSendingSub: initialize.NewDatabusV1(c.RewardsAwardSendingSub),
		cron:                   cron.New(),
		questionRand:           rand.New(rand.NewSource(time.Now().UnixNano())),
		bwsLotteryRand:         rand.New(rand.NewSource(time.Now().UnixNano())),
		bnjAwardch:             make(map[int]chan *bnjmdl.AwardAction, len(awardTypes)),
		wxLottery:              &l.TransInfo{TransDesc: c.WxLottery.TransDesc, WithdrawEndTime: c.WxLottery.WithdrawEndTime, WithdrawStartHour: c.WxLottery.WithdrawStartHour},

		s10Dao:      s10dao.New(c),
		s10General:  c.S10General,
		freeFlowSub: initialize.NewDatabusV1(c.FreeFlowSub),
		gaiaRiskSub: initialize.NewDatabusV1(c.GaiaRiskSub),

		natWorker: fanout.New("nat-worker", fanout.Worker(1), fanout.Buffer(1024)),
		funny:     funny.New(c),

		sourceSvr:     sourceSvr.New(c),
		rankSvr:       rankSvr.New(c),
		rankV3Svr:     rankv3Svr.New(c),
		FitSvr:        fitSvr.New(c),
		actplatClient: client.ActplatClient,
		arcClient:     client.ArcClient,
	}
	var err error
	if s.actGRPC, err = actapi.NewClient(nil); err != nil {
		panic(err)
	}
	if s.accClient, err = api.NewClient(c.AccClient); err != nil {
		panic(err)
	}
	if s.coinClient, err = coinapi.NewClient(c.CoinClient); err != nil {
		panic(err)
	}
	if s.suitClient, err = suitapi.NewClient(c.SuitClient); err != nil {
		panic(err)
	}
	if s.artClient, err = artapi.NewClient(c.ArtClient); err != nil {
		panic(err)
	}
	if s.tagClient, err = tagapi.NewClient(c.TagClient); err != nil {
		panic(err)
	}
	if s.tagNewClient, err = tagnewapi.NewClient(c.TagNewClient); err != nil {
		panic(err)
	}
	if s.cheeseClient, err = cheeseapi.NewClient(c.CheeseClient); err != nil {
		panic(err)
	}
	if s.relationClient, err = relationapi.NewClient(c.RelationClient); err != nil {
		panic(err)
	}
	if s.favoriteClient, err = favoriteapi.New(c.FavoriteClient); err != nil {
		panic(err)
	}
	if s.memberClient, err = memberAPI.NewClient(c.MemberClient); err != nil {
		panic(err)
	}
	if s.videoupClient, err = videoupapi.NewClient(c.VideoUpClient); err != nil {
		panic(err)
	}
	if s.bbqtaskClient, err = bbqtaskapi.NewClient(c.BbqTaskClient); err != nil {
		panic(err)
	}
	if s.gaialibClient, err = gaialibapi.NewClient(c.GaiaLibClient); err != nil {
		panic(err)
	}
	initialize.New(acg.New, func() {
		s.AcgSrv = acg.New(c, s.dao, s.arcClient)
	})
	s.subActI = 0
	for i := 0; i < _sharding; i++ {
		s.subActionCh = append(s.subActionCh, make(chan *l.Action, 10240))
		s.waiter.Add(1)
		go s.actionDealProc(i)
	}
	s.initRules()
	s.initBnj()
	s.arcActionCh = make(chan *l.Archive, 10240)
	// 增加抽奖机会
	s.lotteryActionch = make(chan *l.LotteryMsg, 10240)
	s.actPlatHistoryCh = make(chan *l.ActPlatHistoryMsg, 1)
	s.lotteryAddtimesCh = make(chan *l.LotteryAddTimesMsg, 1)
	s.lotteryTypeAddTimes = make(map[int][]*l.Lottery)
	s.actPlatHistoryWaitCh = make(chan struct{})
	s.lotteryAddtimesWaitCh = make(chan struct{})

	// 拜年祭奖励channel
	for _, typ := range awardTypes {
		s.bnjAwardch[typ] = make(chan *bnjmdl.AwardAction, 102400)
	}
	s.getLotteryTimes()
	s.waiter.Add(1)
	// 免流databus
	go s.freeFlowProc()
	s.waiter.Add(1)
	go s.lotteryProc()
	// 消费评论databus
	s.waiter.Add(1)
	go s.replyproc()
	// 稿件下架处理
	s.waiter.Add(1)
	go s.actLikeproc()
	// archive pass databus
	s.waiter.Add(1)
	go s.archiveCanal()
	// archive action databus
	s.waiter.Add(1)
	go s.archiveActionCanal()
	// activity db canal
	s.waiter.Add(1)
	go s.consumeCanal()
	s.waiter.Add(1)
	go s.consumeBnjCanal()
	// coin databus
	s.waiter.Add(1)
	go s.statCoinproc()
	// thumbup databus
	s.waiter.Add(1)
	go s.statThumbupproc()
	// article pass databus
	s.waiter.Add(1)
	go s.articlePassproc()
	// live follow databus
	s.waiter.Add(1)
	go s.liveFollowproc()
	s.waiter.Add(1)
	go s.awardproc()
	s.waiter.Add(1)
	go s.vipLotteryproc()
	s.waiter.Add(1)
	go s.bwsVipCardproc()
	s.waiter.Add(1)
	go s.ottVipLotteryproc()
	s.waiter.Add(1)
	go s.bnjproc()
	s.waiter.Add(1)
	go s.bnjAwardproc()
	s.waiter.Add(1)
	go s.customizeLotteryproc()
	s.waiter.Add(1)
	go s.StartRankJob()
	go s.updateLotteryTimesLoop()
	for i := 0; i < s.c.Bnj2020.PugvCount; i++ {
		s.waiter.Add(1)
		go s.bnjPugvAwardproc()
	}
	for i := 0; i < s.c.Bnj2020.ComicCount; i++ {
		s.waiter.Add(1)
		go s.bnjComicAwardproc()
	}
	for i := 0; i < s.c.Bnj2020.PendantCount; i++ {
		s.waiter.Add(1)
		go s.bnjPendantAwardproc()
	}
	for i := 0; i < s.c.Bnj2020.MallCount; i++ {
		s.waiter.Add(1)
		go s.bnjMallAwardproc()
	}
	for i := 0; i < s.c.Bnj2020.LiveCount; i++ {
		s.waiter.Add(1)
		go s.bnjLiveAwardproc()
	}
	s.waiter.Add(1)
	go s.subjectRuleStatproc()
	s.waiter.Add(1)
	go s.syncUserActionActivityProc()
	s.waiter.Add(1)
	go s.userActionStatSyncProc()
	s.waiter.Add(1)
	go s.syncFullData2CounterProc()
	s.waiter.Add(1)
	go s.LoadNotifySubjectInfoProc()
	s.waiter.Add(1)
	go s.ReserveNotifySubProc()
	s.waiter.Add(1)
	go s.SyncFavAvid2Counter()
	s.waiter.Add(1)
	go s.gaiaRiskProc()
	s.waiter.Add(1)
	s.lotteryConsumewaiter.Add(1)
	go s.actPlatHistoryMsgToCh()
	s.waiter.Add(1)
	s.lotteryConsumewaiter.Add(1)
	go s.actPlatHistoryMsgSpToCh()
	s.waiter.Add(1)
	go s.ActTolotteryAddTimes()
	s.waiter.Add(1)
	go s.lotteryAddtimesMsgToCh()
	s.waiter.Add(1)
	go s.lotteryAddTimesConsume()
	s.bnjValue = s.c.Bnj2020.MaxValue
	s.subsRankproc()
	s.questionBaseproc()
	s.steinListproc()
	s.loadGuessOidsproc()
	s.loadStaffArc()
	s.loadRestartArc()
	s.loadAwardSubjects()
	//s.loadYelAndGreenArc()
	s.loadMobileGameArc()
	s.loadStupidArc()
	s.startDatabus(c)

	s.createCron()

	go s.RetryFailedFinishGuessLoop()

	s.contestID = make(map[int64]struct{})
	s.mainID = make(map[int64]struct{})
	s.mainContestID = make(map[int64]int64)
	s.loadContest()
	go s.loadCourseProc()
	//go s.loadOperationData()
	go s.upActReserveLiveStateExpire()

	// 拜年纪2021：
	// 初始化拜年纪预约直播间礼签
	_ = bnj2021.UpdateBnjReserveLiveAwardCfg()
	// 初始化拜年纪2021所有奖池信息
	_ = bnj2021.UpdateBnjRewardCfg()
	// 初始化拜年纪2021兜底奖励信息
	_ = bnj2021.UpdateBnjDefaultReward()
	// 初始化拜年纪预约规则
	_ = bnj2021.UpdateReserveRewardRule()
	// 初始化拜年纪直播观看时长达标奖励规则
	_ = bnj2021.UpdateLiveDurationRule()
	// 初始化拜年纪答题统计配置
	_ = bnj2021.UpdateExamStatRule()
	// 初始化拜年纪业务限流规则
	_ = bnj2021.UpdateBizLimitRule()
	// watch拜年纪配置变更
	bnj2021.RegisterFileWatcher()
	// 初始化AR游戏兑换奖券kafka配置
	bnj2021.InitARRewardConsumerCfg(c.Bnj2021ARSub)
	// 初始化直播间领奖kafka配置
	bnj2021.InitBnjLiveDrawRecConfig(c.Bnj2021LiveDrawRec)
	// 初始化直播间抽奖券获取配置
	bnj2021.InitBnjLiveDrawCouponConfig(c.Bnj2021LiveDrawARCoupon)
	// 初始化拜年祭直播间观看时长达标用户记录落库kafka配置
	bnj2021.InitBnjUserInLiveRoomConfig(c.Bnj2021LiveUser)
	// AR扭蛋券兑换
	go initialize.CallC(bnj2021.ASyncARRewardConsumer)
	// AR扭蛋券兑换(灾备队列)
	go initialize.CallC(bnj2021.ASyncARExchangeFromBackupMQ)
	// 实时消费拜年祭直播间观看时长达标用户记录落库
	go initialize.CallC(bnj2021.ASyncUserLogInLiveRoom)
	// 基于拜年纪直播间观看时长达标抽奖规则轮训触发抽奖(实际不发放仅记录)
	go initialize.CallC(bnj2021.StartBnjLotteryBizByRules)
	// 异步完成用户在拜年纪直播领取提前抽中的奖品
	go initialize.CallC(bnj2021.ASyncLiveRewardReceiveBiz)
	// 异步完成用户在拜年纪直播领取提前抽中的奖品(灾备队列)
	go initialize.CallC(bnj2021.ASyncLiveAwardRecFromBackupMQ)
	// 异步完成用户在拜年纪扭蛋券获取到直播间抽奖券抽奖逻辑(实际不发放仅记录)
	go initialize.CallC(bnj2021.ASyncLiveARCouponBiz)
	// 异步完成用户在拜年纪扭蛋券获取到直播间抽奖券抽奖逻辑(实际不发放仅记录 灾备队列)
	go initialize.CallC(bnj2021.ASyncLiveARCouponFromBackupMQ)
	// 实时进行预约阶段奖励业务
	go initialize.CallC(bnj2021.ASyncReserveRewardRuleAndPay)
	// 实时更新拜年纪预约人数
	go initialize.CallC(bnj2021.ASyncBnjReservedCount)
	// 更新资源位配置
	go initialize.CallC(bnj2021.ASyncResetPCConfiguration)
	// 实时更新答题统计
	go initialize.CallC(bnj2021.ASyncUpdateExamStats)

	// 异步奖励发放
	go s.ASyncRewardsAwardSending(Ctx4Worker)
	// 常规许愿/稿件活动类用户数据消费
	go wish_2021_spring.ASynCommonActivityUserCommitConsumeFromBackupMQ(Ctx4Worker)

	// 投票数据刷新
	go initialize.CallC(s.VoteRefreshDSItemNotEnd)
	go initialize.CallC(s.VoteRefreshDSItemEndWithin90)
	go initialize.CallC(s.VoteRefreshRankNotEnd)
	go initialize.CallC(s.VoteRefreshRankEndWithin90)
	go initialize.CallC(s.RefreshBindConfig)
	s.startMissionTaskConsumer()
	go initialize.CallC(s.refreshValidActivity)
	go initialize.CallC(s.makeUpReceiveRecords)

	//cpc100 pv
	if env.DeployEnv == env.DeployEnvProd {
		go initialize.CallC(s.RefreshCpc100PV)
	}
	go initialize.CallC(s.RefreshCpc100Topic)
	go initialize.CallC(s.refreshOlympicContest)
	return s
}

// startDatabus ...
func (s *Service) startDatabus(cfg *conf.Config) {
	s.lotteryRailgun = railgun.NewRailGun("抽奖次数任务消费", nil,
		railgun.NewDatabusV1Inputer(&railgun.DatabusV1Config{Config: s.c.LotteryDatabus.ActHistoryDatabus}),
		railgun.NewSingleProcessor(s.c.LotteryDatabus.LotteryRailgun, s.actPointMsgConsume, s.actPointAddLotteryTimes))
	s.lotteryRailgun.Start()
}

func (s *Service) startMissionTaskConsumer() {
	s.missionTaskRailgun = railgun.NewRailGun(
		"MissionTaskConsumer",
		nil,
		railgun.NewDatabusV1Inputer(&railgun.DatabusV1Config{Config: s.c.MissionTaskSub.Databus}),
		railgun.NewSingleProcessor(s.c.MissionTaskSub.Railgun, s.actPointMsgConsume, s.missionGroupsConsumer),
	)
	s.missionTaskRailgun.Start()
}

func (s *Service) initRules() {
	s.liveSeids = make(map[int64]struct{}, len(s.c.Live.Seids))
	for _, id := range s.c.Live.Seids {
		s.liveSeids[id] = struct{}{}
	}
	s.lotteryAdds = make(map[int64]*conf.LotteryAddRule, len(s.c.Rule.LotteryAddRule))
	for _, v := range s.c.Rule.LotteryAddRule {
		s.lotteryAdds[v.Sid] = v
	}
	s.suitAwardIDs = make(map[int64]struct{}, len(s.c.YeGr.SuitAwardIDs))
	for _, id := range s.c.YeGr.SuitAwardIDs {
		s.suitAwardIDs[id] = struct{}{}
	}
	s.vipLotteryIDs = make(map[int64]int64, len(s.c.Rule.VipLotteryIDs))
	for _, v := range s.c.Rule.VipLotteryIDs {
		s.vipLotteryIDs[v.AppID] = v.LotteryID
	}
	s.ottVipPids = make(map[int64]int64, len(s.c.Rule.OttVipLottery))
	for _, v := range s.c.Rule.OttVipLottery {
		s.vipLotteryIDs[v.Pid] = v.LotteryID
	}
}

func (s *Service) likeArc(c context.Context, sub *l.Subject) (res *l.Subject, err error) {
	if sub != nil {
		if sub.ID == 0 {
			res = nil
		} else {
			res = sub
			var (
				ok   bool
				arcs *arcapi.ArcsReply
				aids []int64
			)
			for _, l := range res.List {
				aids = append(aids, l.Wid)
			}
			argAids := &arcapi.ArcsRequest{
				Aids: aids,
			}
			if arcs, err = s.arcClient.Arcs(c, argAids); err != nil {
				log.Error("s.arcRPC.Archives(arcAids:(%v), arcs), err(%v)", aids, err)
				return
			}
			for _, l := range res.List {
				if l.Archive, ok = arcs.Arcs[l.Wid]; !ok {
					log.Info("s.arcs.wid:(%d), (%v)", l.Wid, ok)
					continue
				}
			}
		}
	}
	return
}

func (s *Service) consumeBnjCanal() {
	defer s.waiter.Done()
	if s.bnjBinlogSub == nil {
		return
	}
	var c = context.Background()
	for {
		msg, ok := <-s.bnjBinlogSub.Messages()
		if !ok {
			log.Info("consumeBnjCanal databus bnj binlog consumer exit!")
			return
		}
		if msg != nil && msg.Value != nil {
			log.Info("consumeBnjCanal receive msg: %s", string(msg.Value))
		}
		msg.Commit()
		m := &match.Message{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("consumeBnjCanal json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		log.Info("consumeBnjCanal key:%s partition:%d offset:%d table:%s", msg.Key, msg.Partition, msg.Offset, m.Table)
		if strings.HasPrefix(m.Table, _userCommitTablePrefix) {
			//一键投稿的送机审能力
			_, pubErr := s.actGRPC.CommonActivityAuditPub(c, &actapi.CommonActivityAuditPubReq{
				ActionType: m.Action,
				TableName:  m.Table,
				RawMessage: m.New,
			})
			log.Infoc(c, "_userCommitTablePrefix Action:%v , table:%v", m.Action, m.Table)
			if pubErr != nil {
				log.Warnc(c, "_userCommitTablePrefix error , message:%s , err:%+v", m.New, pubErr)
			}
		}
	}
}

func (s *Service) consumeCanal() {
	defer s.waiter.Done()
	if s.actSub == nil {
		return
	}
	var c = context.Background()
	for {
		msg, ok := <-s.actSub.Messages()
		if !ok {
			log.Info("databus: activity-job binlog consumer exit!")
			return
		}
		if msg != nil && msg.Value != nil {
			log.Info("consumeCanal receive msg: %s", string(msg.Value))
		}
		msg.Commit()
		m := &match.Message{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		if strings.HasPrefix(m.Table, _guessUserTable) {
			if m.Action == match.ActInsert {
				s.addGuessLotteryTimes(c, m.New)
				s.upUserGuess(c, m.New)
				s.pubPredictMsg(c, m.New)
			} else if m.Action == match.ActUpdate {
				s.upUserGuess(c, m.New)
			}
		} else if strings.HasPrefix(m.Table, _actReserve) {
			if m.Action == match.ActInsert {
				newReserve := &l.Reserve{}
				if err := json.Unmarshal(m.New, newReserve); err != nil {
					log.Error("consumeCanal json.Unmarshal(%s) error(%+v)", m.New, err)
					continue
				}
				// 同步countermid列表
				s.SyncMid2Counter(c, newReserve)
				if awardTask, ok := s.awardSubjectTask[newReserve.Sid]; ok && awardTask != nil {
					if time.Now().Unix() <= int64(awardTask.Etime) {
						s.singleDoTask(c, newReserve.Mid, awardTask.TaskID)
					}
				}
				sid := strconv.FormatInt(newReserve.Sid, 10)
				orderNo := strconv.FormatInt(newReserve.Mid, 10) + sid + strconv.FormatInt(newReserve.ID, 10)
				s.goAddLotteryTimesByType(c, sid, newReserve.Mid, l.LotteryArcType, orderNo)
				s.dao.AsyncSendGroupDatabus(c, newReserve, time.Now().Unix()) // 预约活动 人群包
				// tunnel push.
				s.pushTunnel(c, newReserve)
				// up发起抽奖预约私信通知
				upActReserveRelationLotteryNotify := trace.SimpleServerTrace(context.Background(), "up_act_reserve_relation_lottery_notify")
				if err := s.UpActReserveRelationLotteryNotify(upActReserveRelationLotteryNotify, newReserve); err != nil {
					log.Errorc(upActReserveRelationLotteryNotify, l.UpActReserveRelationLotteryNotify+err.Error())
				}
			} else if m.Action == match.ActUpdate {
				s.pushEditTunnelGroup(c, m.New, m.Old) // 预约活动 人群包
				s.pushEditTunnel(c, m.New, m.Old)
			}
			if m.Action == match.ActInsert || m.Action == match.ActUpdate {
				newObj := new(l.ActReserveField)
				if err := json.Unmarshal(m.New, newObj); err != nil {
					log.Errorc(c, "[act_reserve_*] json.Unmarshal data[%s] err[%v]", string(m.New), err)
					return
				}
				s.PubReserveUserUpdate(c, newObj)
				s.PubReserveActivityData(c, newObj)
				reserveRelationLotteryUserReserveState := trace.SimpleServerTrace(context.Background(), "up_act_reserve_relation_lottery_user_reserve_state")
				if err := s.PubUpActReserveRelationLotteryReserve(reserveRelationLotteryUserReserveState, newObj); err != nil {
					log.Errorc(reserveRelationLotteryUserReserveState, l.UpActReserveRelationLotteryUserReserveState+err.Error())
				}
				// s.CopyReserveItem2NewTable(c, m.Action, newObj)
			}
		} else if strings.HasPrefix(m.Table, _wxLotteryPrefix) {
			s.wxLotteryLogHandle(m)

		} else if strings.HasPrefix(m.Table, _lotteryActionPrefix) {
			s.lotteryActionRedisClean(c, m)
		} else if strings.HasPrefix(m.Table, _lotteryActionPrefix) {
			s.lotteryActionRedisClean(c, m)
		} else if strings.HasPrefix(m.Table, _lotteryAddtimesPrefix) {
			s.lotteryAddRedisClean(c, m)
		} else if strings.HasPrefix(m.Table, _taskUserLogPrefix) {
			s.upTaskStat(c, m)
		} else if strings.HasPrefix(m.Table, _s10UserCostRecordPrefix) {
			s.userCostRecord(m)
		} else if strings.HasPrefix(m.Table, _taskUserStatePrefix) {
			s.PubClockInUserUpdate(c, m.New, m.Old)
		} else if strings.HasPrefix(m.Table, _taskStatePrefix) {
			s.PubClockInRuleStatUpdate(c, m.New, m.Old)

		} else if strings.HasPrefix(m.Table, _actBwsReservePrefix) {
			// bws 互动预约
			if m.Action == match.ActInsert {
				s.PubBwsReserve(c, m.New, m.Old)
			}
		} else {
			switch m.Table {
			case _matchObjTable:
				if m.Action == match.ActUpdate {
					s.upMatchUser(c, m.New, m.Old)
				}
			case _subjectTable:
				if m.Action == match.ActInsert || m.Action == match.ActUpdate {
					s.upSubject(c, m.New, m.Old)

					newObj := new(l.ActSubject)
					if err := json.Unmarshal(m.New, newObj); err != nil {
						log.Error("[JobDataMonitorError] act_subject insert json.Unmarshal err(%+v)", err)
						return
					}
					if m.Action == match.ActInsert {
						// up主发起预约 需要去渠道中台创建通知包 预约类型为 24 => up主发起稿件预约 25 => up主发起直播预约
						groupCtx := trace.SimpleServerTrace(context.Background(), "up_act_reserve_group")
						if tool.InInt64Slice(int64(newObj.Type), []int64{l.UPRESERVATIONARC, l.UPRESERVATIONLIVE}) {
							if err := s.CreateUpActNotify2Platform(groupCtx, newObj); err != nil {
								log.Errorc(groupCtx, l.UpActReserveLogPrefix+"CreateUpActNotify2Platform err(%+v)", err)
							}
						}
					}
				} else if m.Action == match.ActDelete {
					s.upSubject(c, m.Old, []byte{})
				}
			case _likesTable:
				var err error
				if m.Action == match.ActInsert {
					s.AddLike(c, m.New)
					newArc := &l.Item{}
					if err = json.Unmarshal(m.New, newArc); err != nil {
						log.Error("consumeCanal json.Unmarshal(%s) error(%+v)", m.New, err)
						continue
					}
					// 审核通过增加抽奖机会
					if newArc.Sid > 0 && newArc.Mid > 0 && newArc.State > 0 {
						// 更新稿件分区计数
						s.dao.UpCacheLikeTypeCount(c, newArc.Sid, int64(newArc.Type))
						sid := strconv.FormatInt(newArc.Sid, 10)
						orderNo := strconv.FormatInt(newArc.Mid, 10) + sid + strconv.FormatInt(newArc.Wid, 10)
						s.goAddLotteryTimesByType(c, sid, newArc.Mid, l.LotteryArcType, orderNo)
						if awardTask, ok := s.awardSubjectTask[newArc.Sid]; ok && awardTask != nil {
							if time.Now().Unix() <= int64(awardTask.Etime) {
								s.singleDoTask(c, newArc.Mid, awardTask.TaskID)
							}
						} else {
							s.dao.AwardSubject(c, newArc.Sid, newArc.Mid)
						}
						// 专栏活动
						if taskID, yok := s.c.Image.ArticleTaskIDs[strconv.FormatInt(newArc.Sid, 10)]; yok && taskID > 0 {
							s.singleDoTask(c, newArc.Mid, taskID)
						}
						// 同步counter wid
						s.SyncAvid2Counter(c, newArc)
						// 专栏挑战更新当日打卡
						s.DayClockIn(c, newArc, 1)
					}
				} else if m.Action == match.ActUpdate {
					s.UpLike(c, m.New, m.Old)
					newArc := &l.Item{}
					oldArc := &l.Item{}
					if err = json.Unmarshal(m.New, newArc); err != nil {
						log.Error("consumeCanal json.Unmarshal(%s) error(%+v)", m.New, err)
						continue
					}
					if newArc.Sid > 0 && newArc.Wid > 0 {
						if err = json.Unmarshal(m.Old, oldArc); err != nil {
							log.Error("consumeCanal json.Unmarshal(%s) error(%+v)", m.Old, err)
							continue
						}
						if oldArc.State <= 0 && newArc.Sid > 0 && newArc.Mid > 0 && newArc.State > 0 {
							// 更新稿件分区计数
							s.dao.UpCacheLikeTypeCount(c, newArc.Sid, int64(newArc.Type))
							sid := strconv.FormatInt(newArc.Sid, 10)
							orderNo := strconv.FormatInt(newArc.Mid, 10) + sid + strconv.FormatInt(newArc.Wid, 10)
							s.goAddLotteryTimesByType(c, sid, newArc.Mid, l.LotteryArcType, orderNo)
							if awardTask, ok := s.awardSubjectTask[newArc.Sid]; ok && awardTask != nil {
								if time.Now().Unix() <= int64(awardTask.Etime) {
									s.singleDoTask(c, newArc.Mid, awardTask.TaskID)
								}
							} else {
								s.dao.AwardSubject(c, newArc.Sid, newArc.Mid)
							}
							// 专栏活动
							if taskID, yok := s.c.Image.ArticleTaskIDs[strconv.FormatInt(newArc.Sid, 10)]; yok && taskID > 0 {
								s.singleDoTask(c, newArc.Mid, taskID)
							}
							// 专栏挑战更新当日打卡
							s.DayClockIn(c, newArc, 2)
						}
						// 同步counter wid
						s.SyncAvid2Counter(c, newArc)
					}
				} else if m.Action == match.ActDelete {
					s.DelLike(c, m.Old)
				}
			case _likeContentTable:
				if m.Action == match.ActDelete {
					s.upLikeContent(c, m.Old)
				} else {
					s.upLikeContent(c, m.New)
				}
			case _likeActionTable:
				if m.Action == match.ActInsert {
					s.actionProc(m.New)
					s.actLikeLotteryProc(m.New)
				}
			case _predictionTable:
				if m.Action == match.ActInsert || m.Action == match.ActUpdate {
					s.UpPreMc(c, m.New)
				} else if m.Action == match.ActDelete {
					s.DelPreMc(c, m.Old)
				}
			case _predictionItemTable:
				if m.Action == match.ActInsert || m.Action == match.ActUpdate {
					s.UpItemPreMc(c, m.New)
				} else if m.Action == match.ActDelete {
					s.DelItemPreMc(c, m.Old)
				}
			case _guessMainTable:
				if m.Action == match.ActInsert {
					s.DelGuessCache(c, m.New)
				} else if m.Action == match.ActUpdate {
					s.BalanceGuess(c, m.New, m.Old) // 代码里有预测完成的判断 在相关位置推送预测完成的消息到任务系统
				}
			case _subjectRuleTable:
				if m.Action == match.ActInsert || m.Action == match.ActUpdate {
					s.upSubjectRule(c, m.New)
				}
			case _subjectStatTable:
				newObj := new(l.SubjectStat)
				if err := json.Unmarshal(m.New, newObj); err != nil {
					log.Errorc(c, "Subject State json.Unmarshal data[%+v] err[%v]", m.New, err)
					continue
				}
				s.UpActReserveChanged(c, newObj)
				if m.Action == match.ActInsert || m.Action == match.ActUpdate {
					s.PubReserveActivityStatUpdate(c, m.New, m.Old)
				}
			case _s10UserCostRecord:
				s.userCostRecord(m)
			case _s10UserGiftRecord:
				s.userLotteryRecord(m)
			case _upActReserveRelation:
				newObj := &l.UpActReserveRelation{}
				oldObj := &l.UpActReserveRelation{}
				if m.Action == match.ActUpdate {
					if err := json.Unmarshal(m.Old, oldObj); err != nil {
						log.Errorc(c, "UpActReserveRelation Old json.Unmarshal(%s) error(%+v)", m.Old, err)
						continue
					}
				}
				if err := json.Unmarshal(m.New, newObj); err != nil {
					log.Errorc(c, "UpActReserveRelation New json.Unmarshal(%s) error(%+v)", m.New, err)
					continue
				}
				// 鸽子蛋活动推流
				//s.UpActReserveRelationChanged(ctx, oldObj, newObj)
				// 原始数据推流
				updateBinLogCtx := trace.SimpleServerTrace(context.Background(), "up_act_reserve_relation_table_monitor")
				s.UpActReserveRelationTableMonitor(updateBinLogCtx, m.Action, oldObj, newObj)
				// 新版卡片绑定和解绑
				platformCard := trace.SimpleServerTrace(context.Background(), "up_act_reserve_relation_platform_card")
				if err := s.BindCard2PlatformRepeatableNew(platformCard, oldObj, newObj); err != nil {
					log.Errorc(platformCard, l.UpActReserveLogPrefix+"BindCard2PlatformRepeatableNew err(%+v)", err)
				}
				// 删除关联动态
				deleteDynamic := trace.SimpleServerTrace(context.Background(), "up_act_reserve_relation_delete_dynamic")
				if err := s.DeleteDynamicRelatedByLiveReserve(deleteDynamic, oldObj, newObj); err != nil {
					log.Errorc(deleteDynamic, "DeleteDynamicRelatedByLiveReserve err:(%+v)", err)
				}
				// 动态+抽奖+预约 审核联动
				reserveRelationChannelAuditNotify := trace.SimpleServerTrace(context.Background(), "up_act_reserve_relation_channel_audit_notify")
				if err := s.ReserveRelationChannelAuditNotify(reserveRelationChannelAuditNotify, oldObj, newObj); err != nil {
					log.Errorc(reserveRelationChannelAuditNotify, l.UpActReserveRelationChannelAuditNotify+err.Error())
				}
				// 预约抽奖私信卡
				upActReserveRelationLotteryNotifyCard := trace.SimpleServerTrace(context.Background(), "up_act_reserve_relation_lottery_notify_create_card")
				if err := s.UpActReserveRelationLotteryNotifyCard(upActReserveRelationLotteryNotifyCard, oldObj, newObj); err != nil {
					log.Errorc(upActReserveRelationLotteryNotifyCard, l.UpActReserveRelationLotteryNotifyCard+err.Error())
				}
			}
			log.Info("consumeCanal key:%s partition:%d offset:%d table:%s", msg.Key, msg.Partition, msg.Offset, m.Table)
		}
	}
}

func (s *Service) likeList(c context.Context, sid int64, offset, limit, retryCnt int) (list []*l.Like, err error) {
	for i := 0; i < retryCnt; i++ {
		if list, err = s.dao.LikeList(c, sid, offset, limit); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) webDataList(c context.Context, vid int64, offset, limit, retryCnt int) (list []*l.WebData, err error) {
	for i := 0; i < retryCnt; i++ {
		if list, err = s.dao.WebDataList(c, vid, offset, limit); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) arts(c context.Context, aids []int64, retryCnt int) (arcs map[int64]*artmdl.Meta, err error) {
	var reply *artapi.ArticleMetasReply
	for i := 0; i < retryCnt; i++ {
		if reply, err = s.artClient.ArticleMetas(c, &artapi.ArticleMetasReq{Ids: aids}); err == nil {
			arcs = reply.Res
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) setSubjectStat(c context.Context, sid int64, stat *l.SubjectTotalStat, count, retryCnt int) (err error) {
	for i := 0; i < retryCnt; i++ {
		if err = s.dao.SetObjectStat(c, sid, stat, count); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

// Ping reports the heath of services.
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

// Close kafaka consumer close.
func (s *Service) Close() (err error) {
	defer s.waiter.Wait()
	s.lotteryRailgun.Close()
	s.closed = true
	s.cron.Stop()
	s.dao.Close()
	s.actSub.Close()
	//s.vipSub.Close()
	s.archiveSub.Close()
	s.bnjSub.Close()
	s.bnjAwardSub.Close()
	s.freeFlowSub.Close()

	s.gaiaRiskSub.Close()
	s.actPlatHistorySub.Close()
	s.actPlatHistorySpSub.Close()
	s.lotteryAddtimesSub.Close()
	s.archiveBinLogSub.Close()
	s.upReservePushPub.Close()

	CancelFunc4Worker()
	s.lotteryConsumewaiter.Wait()
	close(s.actPlatHistoryCh)
	for range s.actPlatHistoryWaitCh {
	}
	for range s.lotteryAddtimesWaitCh {
	}
	return
}

func (s *Service) createCron() {
	var err error
	if err = s.cron.AddFunc(s.c.Interval.GuessCron, s.guessRank); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.UpListHisCron, s.addUpListHis); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.EntCron, s.entRank); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.QuestionCron, s.questionBaseproc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.SubsRankCron, s.subsRankproc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.PoolCreatCron, s.poolCreateproc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.SteinListCron, s.steinListproc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.LoadGuessOidsCron, s.loadGuessOidsproc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.LoadStaffCron, s.loadStaffArc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.BnjReserveCron, s.bnjReserveproc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.SpringCron, s.loadRestartArc); err != nil {
		panic(err)
	}
	//if err = s.cron.AddFunc(s.c.Interval.SpringCron, s.loadYelAndGreenArc); err != nil {
	//	panic(err)
	//}
	if err = s.cron.AddFunc(s.c.Interval.AwardSubjectCron, s.loadAwardSubjects); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.ImageCron, s.loadImageLikes); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.SpringCron, s.loadMobileGameArc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.FactionCron, s.loadFactionLikes); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.SpringCron, s.loadStupidArc); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.StupidCron, s.loadStupidArcs); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.WxLottery.PageCron, s.wxLotteryLogPage); err != nil {
		panic(err)
	}
	if s.c.Bnj2020.Cron != "" {
		if err = s.cron.AddFunc(s.c.Bnj2020.Cron, s.cronInformationMessage); err != nil {
			panic(err)
		}
	}

	if err = s.cron.AddFunc(s.c.Interval.NewstarUpArcCron, s.NewstarArchiveTask); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Interval.NewstarFinishCron, s.FinishNewstar); err != nil {
		panic(err)
	}
	if s.c.Interval.ArticleDayCron != "" {
		if err = s.cron.AddFunc(s.c.Interval.ArticleDayCron, s.FinishArticleDay); err != nil {
			panic(err)
		}
	}
	if err = s.cron.AddFunc(s.c.Interval.NewstarIdentityCron, s.IdentityChecking); err != nil {
		panic(err)
	}
	if s.c.GameHoliday.SyncCron != "" {
		if err = s.cron.AddFunc(s.c.GameHoliday.SyncCron, s.GameHolidaySyncCounterFilter); err != nil {
			panic(err)
		}
	}
	if s.c.Interval.PushUpReserveVerifyCron != "" {
		if err = s.cron.AddFunc(s.c.Interval.PushUpReserveVerifyCron, s.upActReservePushVerify); err != nil {
			panic(err)
		}
	}
	//if s.c.Funny.SyncVideoCron != "" {
	//	if err = s.cron.AddFunc(s.c.Funny.SyncVideoCron, s.FunnySyncVideoData); err != nil {
	//		panic(err)
	//	}
	//}
	//if s.c.Funny.CaculatePartOne != "" {
	//	if err = s.cron.AddFunc(s.c.Funny.CaculatePartOne, s.CaculatePartOne); err != nil {
	//		panic(err)
	//	}
	//}
	//if s.c.Funny.CaculatePartTwo != "" {
	//	if err = s.cron.AddFunc(s.c.Funny.CaculatePartTwo, s.CaculatePartTwo); err != nil {
	//		panic(err)
	//	}
	//}
	//if s.c.Column.TriggerCron != "" {
	//	if err = s.cron.AddFunc(s.c.Column.TriggerCron, s.ColumnDataExport); err != nil {
	//		panic(err)
	//	}
	//}
	if s.c.S10Contribution.SyncDaySelectCron != "" {
		if err = s.cron.AddFunc(s.c.S10Contribution.SyncDaySelectCron, s.ContributionSyncCounterFilter); err != nil {
			panic(err)
		}
	}
	if s.c.S10Contribution.CalcSplitPeopleCron != "" {
		if err = s.cron.AddFunc(s.c.S10Contribution.CalcSplitPeopleCron, s.CalcUserContribution); err != nil {
			panic(err)
		}
	}
	if s.c.S10Contribution.UpDBCron != "" {
		if err = s.cron.AddFunc(s.c.S10Contribution.UpDBCron, s.UpUserContributionDB); err != nil {
			panic(err)
		}
	}
	if s.c.S10Contribution.TotalRankCron != "" {
		if err = s.cron.AddFunc(s.c.S10Contribution.TotalRankCron, s.ScoreTotalRank); err != nil {
			panic(err)
		}
	}
	if s.c.Selection.CalcAssistanceCron != "" {
		if err = s.cron.AddFunc(s.c.Selection.CalcAssistanceCron, s.SelAssistance); err != nil {
			panic(err)
		}
	}
	if s.c.Selection.VoteReportCron != "" {
		if err = s.cron.AddFunc(s.c.Selection.VoteReportCron, s.DayVoteReport); err != nil {
			panic(err)
		}
	}
	if s.c.Selection.ResetSelectionVoteCron != "" {
		if err = s.cron.AddFunc(s.c.Selection.ResetSelectionVoteCron, s.SetSelectionVoteCache); err != nil {
			panic(err)
		}
	}
	if s.c.College.RankCron != "" {
		if err = s.cron.AddFunc(s.c.College.RankCron, s.CollegeRank); err != nil {
			panic(err)
		}
	}
	if s.c.College.VersionCron != "" {
		if err = s.cron.AddFunc(s.c.College.VersionCron, s.CollegeVersion); err != nil {
			panic(err)
		}
	}
	if s.c.Share.ShareCron != "" {
		if err = s.cron.AddFunc(s.c.Share.ShareCron, s.ShareURLUpdate); err != nil {
			panic(err)
		}
	}
	if s.c.College.BonusCron != "" {
		if err = s.cron.AddFunc(s.c.College.BonusCron, s.CollegeBonus); err != nil {
			panic(err)
		}
	}
	if s.c.College.ScoreCron != "" {
		if err = s.cron.AddFunc(s.c.College.ScoreCron, s.CollegeScore); err != nil {
			panic(err)
		}
	}
	if s.c.Dubbing.RankCron != "" {
		if err = s.cron.AddFunc(s.c.Dubbing.RankCron, s.DubbingRank); err != nil {
			panic(err)
		}
	}
	if s.c.College.ScoreAutoCron != "" {
		if err = s.cron.AddFunc(s.c.College.ScoreAutoCron, s.CollegeScoreAuto); err != nil {
			panic(err)
		}
	}
	if s.c.S10Answer.HourCron != "" {
		if err = s.cron.AddFunc(s.c.S10Answer.HourCron, s.AnswerHour); err != nil {
			panic(err)
		}
	}
	if s.c.S10Answer.WeekCron != "" {
		if err = s.cron.AddFunc(s.c.S10Answer.WeekCron, s.AnswerWeek); err != nil {
			panic(err)
		}
	}
	if s.c.Rank.RankCron != "" {
		if err = s.cron.AddFunc(s.c.Rank.RankCron, s.CronStartJob); err != nil {
			panic(err)
		}
	}
	if s.c.Lottery.WinListCron != "" {
		if err = s.cron.AddFunc(s.c.Lottery.WinListCron, s.ClearLotteryWinList); err != nil {
			panic(err)
		}
	}
	if s.c.Interval.ActRelationInfoCron != "" {
		if err = s.cron.AddFunc(s.c.Interval.ActRelationInfoCron, s.CallInternalSyncActRelationInfoDB2Cache); err != nil {
			panic(err)
		}
	}
	if s.c.Interval.ActSubjectInfoCron != "" {
		if err = s.cron.AddFunc(s.c.Interval.ActSubjectInfoCron, s.CallInternalSyncActSubjectInfoDB2Cache); err != nil {
			panic(err)
		}
	}
	if s.c.Interval.ActSubjectReserveIDsInfoCron != "" {
		if err = s.cron.AddFunc(s.c.Interval.ActSubjectReserveIDsInfoCron, s.CallInternalSyncActSubjectReserveIDsInfoDB2Cache); err != nil {
			panic(err)
		}
	}
	if s.c.YellowAndGreen.YingYuanVoteCron != "" {
		if err = s.cron.AddFunc(s.c.YellowAndGreen.YingYuanVoteCron, s.YingYuanVote); err != nil {
			panic(err)
		}
	}
	if s.c.Knowledge.Cron != "" {
		if err = s.cron.AddFunc(s.c.Knowledge.Cron, s.ActKnowledge); err != nil {
			panic(err)
		}
	}
	// if s.c.Handwrite2021.Cron != "" {
	// 	if err = s.cron.AddFunc(s.c.Handwrite2021.Cron, s.Handwrite2021); err != nil {
	// 		panic(err)
	// 	}
	// }
	// if s.c.Handwrite2021.DataCron != "" {
	// 	if err = s.cron.AddFunc(s.c.Handwrite2021.DataCron, s.Handwrite2021Data); err != nil {
	// 		panic(err)
	// 	}
	// }
	if s.c.Rank.NewRankCron != "" {
		if err = s.cron.AddFunc(s.c.Rank.NewRankCron, s.SetRankLogCron); err != nil {
			panic(err)
		}
	}
	if s.c.ActDomainConfig.RunCron != "" {
		if err = s.cron.AddFunc(s.c.ActDomainConfig.RunCron, s.SyncActDomainCache); err != nil {
			panic(err)
		}
	}
	// fit健身打卡相关
	if s.c.FitJobConfig.RunFlushCron != "" {
		if err = s.cron.AddFunc(s.c.FitJobConfig.RunFlushCron, s.FitSvr.FlushPlanData); err != nil {
			panic(err)
		}
	}
	if s.c.FitJobConfig.RunTianMaCron != "" {
		if err = s.cron.AddFunc(s.c.FitJobConfig.RunTianMaCron, s.FitSvr.SendTianMaCard); err != nil {
			panic(err)
		}
	}
	if s.c.FitJobConfig.RunSetMemberCron != "" {
		if err = s.cron.AddFunc(s.c.FitJobConfig.RunSetMemberCron, s.FitSvr.SetMemberIntToRW); err != nil {
			panic(err)
		}
	}

	if s.c.GaoKaoAnswer.RunCron != "" {
		if err = s.cron.AddFunc(s.c.GaoKaoAnswer.RunCron, s.UpdateQuestionRecord); err != nil {
			panic(err)
		}
	}
	if s.c.KnowledgeTask.DelCron != "" {
		if err = s.cron.AddFunc(s.c.KnowledgeTask.DelCron, s.DeleteKnowledgeCalculate); err != nil {
			panic(err)
		}
	}
	/*if s.c.BwPark2021.RunCron != "" {
		if err = s.cron.AddFunc(s.c.BwPark2021.RunCron, s.bwParkStockSync); err != nil {
			panic(err)
		}

		if err = s.cron.AddFunc(s.c.BwPark2021.RunCron, s.bwBatchCacheTicketBind); err != nil {
			panic(err)
		}
	}*/

	if s.c.StockServerJobConf.RunCron != "" {
		if err = s.cron.AddFunc(s.c.StockServerJobConf.RunCron, s.StockServerSyncJob); err != nil {
			panic(err)
		}
	}
	s.cron.Start()
}

func (s *Service) loadLikeList(c context.Context, sid int64, retryCnt int) (list []*l.Like, err error) {
	for i := 0; i < retryCnt; i++ {
		if list, err = s.dao.LikeListState(c, sid); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (s *Service) loadLikesList(c context.Context, sids []int64, retryCnt int) (list []*l.Like, err error) {
	for i := 0; i < retryCnt; i++ {
		if list, err = s.dao.LikesListState(c, sids); err != nil {
			log.Error("loadLikesList s.dao.LikesListState(%v) error(%+v)", sids, err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	return
}

func (s *Service) loadRestartArc() {
	var (
		likes []*l.WebData
		err   error
		tmp   = make(map[int64]struct{})
	)
	if likes, err = s.webDataList(context.Background(), s.c.Restart2020.Vid, 0, _objectPieceSize, _retryTimes); err != nil {
		log.Error("loadRestartArc s.webDataList(%d,%d,%d) error(%+v)", s.c.Restart2020.Vid, 0, _objectPieceSize, err)
		return
	}
	for _, val := range likes {
		if val == nil || val.Data == "" {
			continue
		}
		aids := &l.AidsData{}
		if err = json.Unmarshal([]byte(val.Data), aids); err != nil {
			log.Error("loadRestartArc json.Unmarshal(%v) error(%+v)", val.Data, err)
			continue
		}
		for _, v := range strings.Split(aids.Aids, ",") {
			i, _ := strconv.ParseInt(v, 10, 64)
			if i > 0 {
				tmp[i] = struct{}{}
			}
		}
	}
	s.restartArc = tmp
}

func (s *Service) loadStaffArc() {
	data, err := s.loadActArc(s.c.Staff.Sid)
	if err != nil {
		return
	}
	s.staffArc = data
	log.Info("loadStaffArc success()")
}

func (s *Service) loadActArc(sid int64) (data map[int64]struct{}, err error) {
	var (
		tmp = make(map[int64]struct{})
		res []*l.Like
	)
	res, err = s.loadLikeList(context.Background(), sid, _retryTimes)
	if err != nil {
		log.Error("loadActArc Err %v", err)
		return
	}
	if len(res) == 0 {
		log.Warn("loadActArc no data, finish")
		return
	}
	for _, val := range res {
		if val != nil && val.State != -1 {
			tmp[val.Wid] = struct{}{}
		}
	}
	data = tmp
	return
}

//func (s *Service) loadYelAndGreenArc() {
//	var (
//		likes []*l.WebData
//		err   error
//		tmp   = make(map[int64]struct{})
//	)
//	if likes, err = s.webDataList(context.Background(), s.c.YelAndGreen.Vid, 0, _objectPieceSize, _retryTimes); err != nil {
//		log.Error("loadYelAndGreenArc s.webDataList(%d,%d,%d) error(%+v)", s.c.YelAndGreen.Vid, 0, _objectPieceSize, err)
//		return
//	}
//	for _, val := range likes {
//		if val == nil || val.Data == "" {
//			continue
//		}
//		aids := &l.AidsData{}
//		if err = json.Unmarshal([]byte(val.Data), aids); err != nil {
//			log.Error("loadYelAndGreenArc json.Unmarshal(%v) error(%+v)", val.Data, err)
//			continue
//		}
//		for _, v := range strings.Split(aids.Aids, ",") {
//			i, _ := strconv.ParseInt(v, 10, 64)
//			if i > 0 {
//				tmp[i] = struct{}{}
//			}
//		}
//	}
//	s.yelAndGreenArc = tmp
//}

// load mobile game act arc
func (s *Service) loadMobileGameArc() {
	var (
		likes []*l.WebData
		err   error
		tmp   = make(map[int64]struct{})
	)
	if likes, err = s.webDataList(context.Background(), s.c.MobileGame.Vid, 0, _objectPieceSize, _retryTimes); err != nil {
		log.Error("loadMobileGameArc s.webDataList(%d,%d,%d) error(%+v)", s.c.MobileGame.Vid, 0, _objectPieceSize, err)
		return
	}
	for _, val := range likes {
		if val == nil || val.Data == "" {
			continue
		}
		aids := &l.AidsData{}
		if err = json.Unmarshal([]byte(val.Data), aids); err != nil {
			log.Error("loadMobileGameArc json.Unmarshal(%v) error(%+v)", val.Data, err)
			continue
		}
		for _, v := range strings.Split(aids.Aids, ",") {
			i, _ := strconv.ParseInt(v, 10, 64)
			if i > 0 {
				tmp[i] = struct{}{}
			}
		}
	}
	s.mobileGameArc = tmp
}
