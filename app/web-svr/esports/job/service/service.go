package service

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/esports/job/component"
	"sync"
	"sync/atomic"
	"time"

	arcclient "git.bilibili.co/bapis/bapis-go/archive/service"
	"git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favclient "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"github.com/robfig/cron"
	"go-common/library/cache/memcache"
	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
	"go-gateway/app/web-svr/esports/job/conf"
	"go-gateway/app/web-svr/esports/job/dao"
	esportModel "go-gateway/app/web-svr/esports/job/model"
	mdlesp "go-gateway/app/web-svr/esports/job/model"
	"go-gateway/app/web-svr/esports/job/service/component_biz"
	"go-gateway/app/web-svr/esports/job/tool"
)

const (
	_favUsers    = 1000
	_tryTimes    = 3
	_defContest  = 0
	_linkinfo    = "点击前往直播间"
	_pushinfo    = "进入直播>>"
	_msgSize     = 500
	_tpMessage   = 0
	_tpPush      = 1
	_arcMaxLimit = 100
	_matchOpt    = "1005"
	_eventOpt    = "1006"
	_gamePid     = 4
)

// Service struct
type Service struct {
	c        *conf.Config
	dao      *dao.Dao
	matchIDs mdlesp.SyncMatch
	clientID string
	// cron
	cron       *cron.Cron
	mapSeason  map[int64]*esportModel.FtpSeason
	mapTeam    map[int64]*esportModel.FtpTeams
	mapMatchs  map[int64]*esportModel.FtpMatchs
	autoRules  *esportModel.AutoRule
	autoGames  map[string]*esportModel.BaseInfo
	autoMatchs map[int64]*esportModel.BaseInfo
	autoTeams  map[int64]*esportModel.Team
	// sub
	archiveNotifySub    *databus.Databus
	waiter              sync.WaitGroup
	gameTypeMap         map[int32]int32
	cache               *fanout.Fanout
	liveOffLineImageMap *atomic.Value
	liveImageCh         chan map[string]string
	autoTagRun          mdlesp.SyncAutoTag
	esportsBinlogSub    *databus.Databus
}

var (
	globalMemcache       *memcache.Memcache
	ctx4Worker           context.Context
	ctxCancelFunc4Worker context.CancelFunc
)

func init() {
	ctx4Worker, ctxCancelFunc4Worker = context.WithCancel(context.Background())
}

var localS *Service

// New init
func New(c *conf.Config) (s *Service) {
	if localS != nil {
		return localS
	}
	s = &Service{
		c:   c,
		dao: dao.New(c),
		matchIDs: mdlesp.SyncMatch{
			Data: make(map[int64]*mdlesp.ContestData),
		},
		cron:                cron.New(),
		archiveNotifySub:    initialize.NewDatabusV1(c.ArchiveNotifySub),
		esportsBinlogSub:    initialize.NewDatabusV1(c.EsportsBinlog),
		gameTypeMap:         make(map[int32]int32),
		cache:               fanout.New("cache"),
		liveOffLineImageMap: new(atomic.Value),
		liveImageCh:         make(chan map[string]string, 1),
		autoRules:           &esportModel.AutoRule{},
		autoGames:           make(map[string]*esportModel.BaseInfo),
		autoMatchs:          make(map[int64]*esportModel.BaseInfo),
		autoTeams:           make(map[int64]*esportModel.Team),
		autoTagRun:          mdlesp.SyncAutoTag{MaxNum: 1},
	}

	// init memcache conn
	globalMemcache = s.dao.GetMc()

	m := make(map[string]bool, 0)
	s.liveOffLineImageMap.Store(m)
	// init berserker configuration
	if err := tool.InitBerserker(c.Berserker); err != nil {
		panic(err)
	}

	conf.LoadSeasonNotifies(c.SeasonStatusNotifier)
	tool.UpdateCropWeChat(c.CorpWeChat)

	// archive consume
	s.waiter.Add(1)
	go s.arcConsumeproc(ctx4Worker)
	go s.matchs(ctx4Worker)
	//go s.contestsMsg()
	//go s.contestsPush()
	//go s.tunnelPushActiveEvent()
	// esports db canal
	//s.waiter.Add(1)
	//go s.consumeCanal()
	// 下掉雷达实时数据
	//go s.pushPoints()
	go s.WatchSeasonStatus(ctx4Worker)
	s.LoadLiveImageMap()
	s.gameTypeID()
	s.autoRule()
	s.autoGame()
	s.autoMatch()
	s.autoTeam()
	s.InitSyncS10RankingData()
	go s.setLiveMissImage(ctx4Worker)
	go s.createCron(ctx4Worker)
	//go WatchLOLTeams(ctx4Worker)
	//go WatchSeasonContests(ctx4Worker)
	//go WatchSeasonPosters(ctx4Worker)
	//go s.SyncScoreAnalysisBiz()
	go s.AsyncAutoSubscribe(context.Background())
	go s.ASyncResetMaxContestID(context.Background())
	// contest component
	go WatchGoingSeasonsComponent(ctx4Worker)
	go WatchSeasonContestComponent(ctx4Worker)
	go WatchSeasonContestBattleComponent(ctx4Worker)
	s.goroutineRegister(s.WatchGoingBattleSeasonsContestsTeams)
	s.goroutineRegister(RefreshAllContestStatusInfoLoop)
	s.goroutineRegister(s.RefreshOffLineImageLoop)
	go component_biz.HotDataHandler(ctx4Worker)
	// 只在线上执行，score接口
	if env.DeployEnv == env.DeployEnvProd {
		go s.ScoreLivePage()
	}
	go s.RefreshAllContestSeriesInfoLoop(ctx4Worker)

	//s.goroutineRegister(s.esportsBinlogConsumer)
	s.goroutineRegister(s.WatchActiveSeasonInfo)
	s.goroutineRegister(s.WatchGamesAllInfo)

	localS = s
	return s
}

func NewV2(c *conf.Config) (s *Service) {
	s = &Service{
		c:   c,
		dao: dao.New(c),
		matchIDs: mdlesp.SyncMatch{
			Data: make(map[int64]*mdlesp.ContestData),
		},
		cron:                cron.New(),
		gameTypeMap:         make(map[int32]int32),
		cache:               fanout.New("cache"),
		liveOffLineImageMap: new(atomic.Value),
		liveImageCh:         make(chan map[string]string, 1),
		autoRules:           &esportModel.AutoRule{},
		autoGames:           make(map[string]*esportModel.BaseInfo),
		autoMatchs:          make(map[int64]*esportModel.BaseInfo),
		autoTeams:           make(map[int64]*esportModel.Team),
		autoTagRun:          mdlesp.SyncAutoTag{MaxNum: 1},
	}
	localS = s
	// 保留service的各种初始化操作
	initServiceConfig(s)
	// databus的多机器部署，仅可写集群部署
	s.multiActiveDeploy(s.c, initDatabus, "databus")
	// 部分databus消费者可多机器部署
	s.multiActiveDeploy(s.c, initDatabusSubCanMultiDeploy, "databus_sub_binlog")
	// 更新数据的ticker or cron ，仅可写集群部署
	s.multiActiveDeploy(s.c, updateTickerOrCron, "update_ticker_cron")
	allDeploy()
	return
}

func initServiceConfig(s *Service) {
	// init memcache conn
	globalMemcache = s.dao.GetMc()
	m := make(map[string]bool, 0)
	s.liveOffLineImageMap.Store(m)
	// init berserker configuration
	if err := tool.InitBerserker(s.c.Berserker); err != nil {
		panic(err)
	}
	conf.LoadSeasonNotifies(s.c.SeasonStatusNotifier)
	tool.UpdateCropWeChat(s.c.CorpWeChat)

	s.LoadLiveImageMap()
	s.gameTypeID()
	s.autoRule()
	s.autoGame()
	s.autoMatch()
	s.autoTeam()
	s.InitSyncS10RankingData()
}

func updateTickerOrCron(ctx context.Context) {
	go initialize.CallC(localS.matchs)
	go initialize.CallC(localS.WatchSeasonStatus)
	go initialize.CallC(localS.setLiveMissImage)
	go initialize.CallC(localS.createCron)
	go localS.AsyncAutoSubscribe(context.Background())
	go localS.ASyncResetMaxContestID(context.Background())
	localS.goroutineRegister(RefreshAllContestStatusInfoLoop)
	localS.goroutineRegister(localS.RefreshOffLineImageLoop)
	// 只在线上执行，score接口
	if env.DeployEnv == env.DeployEnvProd {
		go localS.ScoreLivePage()
	}
}

func allDeploy() {
	// contest component

	go WatchGoingSeasonsComponent(ctx4Worker)
	go WatchSeasonContestComponent(ctx4Worker)
	go WatchSeasonContestBattleComponent(ctx4Worker)
	localS.goroutineRegister(localS.WatchGoingBattleSeasonsContestsTeams)
	go component_biz.HotDataHandler(ctx4Worker)
	go localS.RefreshAllContestSeriesInfoLoop(ctx4Worker)
	localS.goroutineRegister(localS.WatchActiveSeasonInfo)
	localS.goroutineRegister(localS.WatchGamesAllInfo)
}

func initDatabusSubCanMultiDeploy(ctx context.Context) {
	log.Infoc(ctx, "[Service][Init][initDatabus][Begin]")
	log.Infoc(ctx, "[Service][Init][initDatabus][End]")
}

func initDatabus(ctx context.Context) {
	log.Infoc(ctx, "[Service][Init][initDatabus][Beign]")
	initDatabusSub(ctx)
	initDataBusPub(ctx)
	log.Infoc(ctx, "[Service][Init][initDatabus][End]")
	localS.waiter.Add(1)
	// databus协程启动
	go initialize.CallC(localS.arcConsumeproc)
	go initialize.CallC(localS.esportsBinlogConsumer)
}

func initDatabusSub(ctx context.Context) {
	log.Infoc(ctx, "[Service][Init][initDatabusSub][Beign]")
	localS.archiveNotifySub = initialize.NewDatabusV1(localS.c.ArchiveNotifySub)
	localS.esportsBinlogSub = initialize.NewDatabusV1(localS.c.EsportsBinlog)
	log.Infoc(ctx, "[Service][Init][initDatabus][End]")
}

func initDataBusPub(ctx context.Context) {

}

func (s *Service) multiActiveDeploy(conf *conf.Config, f func(ctx context.Context), configKey string) {
	if conf.DeployOption == nil || !conf.DeployOption.WriteDeploy {
		return
	}
	if configKey == "" {
		return
	}
	if len(conf.DeployOption.Switch) == 0 {
		return
	}
	if !conf.DeployOption.Switch[configKey] {
		return
	}
	f(ctx4Worker)
}

func (s *Service) goroutineRegister(f func(ctx context.Context)) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("[Async][GoRoutine][Run][Panic][Recover]err:(%+v)", err)
				return
			}
		}()
		f(context.Background())
	}()
}

func (s *Service) ftpDataCron() {
	stime := time.Now().Unix()
	log.Info("ftpDataCron group(%d) start(%s)", stime, time.Now().Format("2006-01-02 15:04:05"))
	s.LoadSeasons()
	s.LoadTeams()
	s.LoadMatchs()
	s.LoadContests()
	s.FtpUpload()
	log.Info("ftpDataCron group(%d) end(%s)", stime, time.Now().Format("2006-01-02 15:04:05"))
}

func (s *Service) createCron(ctx context.Context) (err error) {
	_ = s.cron.AddFunc(s.c.Berserker.CronSpec, s.startArchiveScoreWorker)
	s.cron.AddFunc(s.c.Leidata.After.GameSleepCron, s.matchGameCron)
	s.cron.AddFunc(s.c.Leidata.After.BigDataCron, s.bigDataCron)
	s.cron.AddFunc(s.c.Leidata.After.InfoDataCron, s.infoDataCron)
	s.cron.AddFunc(s.c.Search.FtpDataCron, s.ftpDataCron)
	s.cron.AddFunc(s.c.Interval.AutoArcRuleCron, s.autoRule)
	s.cron.AddFunc(s.c.Interval.AutoArcRuleCron, s.autoGame)
	s.cron.AddFunc(s.c.Interval.AutoArcRuleCron, s.autoMatch)
	s.cron.AddFunc(s.c.Interval.AutoArcRuleCron, s.autoTeam)
	s.cron.AddFunc(s.c.RankingDataWatch.Cron, func() {
		s.SyncS10RankingData()
	})
	s.cron.AddFunc(s.c.Interval.AutoArcPassCron, s.NewAutoCheckPass)
	if err = s.cron.AddFunc("@every 10s", s.LoadLiveImageMap); err != nil {
		panic(err)
	}
	s.cron.Start()
	return
}

func (s *Service) contestsMsg() {
	var (
		err      error
		contests []*mdlesp.Contest
	)
	for {
		stime := time.Now().Add(time.Duration(s.c.Rule.Before))
		etime := stime.Add(time.Duration(s.c.Rule.SleepInterval))
		if contests, err = s.contests(context.Background(), stime.Unix(), etime.Unix()); err != nil {
			log.Error("contestsMsg contests stime(%d) etime(%d) error(%+v)", stime.Unix(), etime.Unix(), err)
		} else {
			for _, contest := range contests {
				tmpContest := contest
				go s.sendContests(tmpContest)
			}
		}
		time.Sleep(time.Duration(s.c.Rule.SleepInterval))
	}
}

func (s *Service) contestsPush() {
	var (
		err      error
		contests []*mdlesp.Contest
	)
	for {
		stime := time.Now()
		etime := stime.Add(time.Second * 2)
		if contests, err = s.contests(context.Background(), stime.Unix(), etime.Unix()); err != nil {
			log.Error("contestsPush contests stime(%d) etime(%d) error(%+v)", stime.Unix(), etime.Unix(), err)
		} else {
			for _, contest := range contests {
				tmpContest := contest
				go s.pubContests(tmpContest)
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func (s *Service) contests(c context.Context, stime, etime int64) (res []*mdlesp.Contest, err error) {
	for i := 0; i < _tryTimes; i++ {
		if res, err = s.dao.Contests(c, stime, etime); err == nil {
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
	if err != nil {
		log.Error("s.dao.Contests error(%v)", err)
	}
	return
}

// Ping Service
func (s *Service) Ping(c context.Context) (err error) {
	return s.dao.Ping(c)
}

// Close Service
func (s *Service) Close() {
	ctxCancelFunc4Worker()
	_ = globalMemcache.Close()
	s.dao.Close()
	if s.archiveNotifySub != nil {
		s.archiveNotifySub.Close()
	}
	if s.esportsBinlogSub != nil {
		s.esportsBinlogSub.Close()
	}
	s.waiter.Wait()
}

func (s *Service) sendContests(contest *mdlesp.Contest) (err error) {
	var (
		mids []int64
		msg  string
	)
	link := fmt.Sprintf("#{%s}{\"https://live.bilibili.com/%d\"}", _linkinfo, contest.LiveRoom)
	if mids, msg, err = s.midsParams(contest, link, _tpMessage); err != nil {
		log.Error("sendContests s.midsParams(%+v) mids_total(%d) error(%v)", contest, len(mids), err)
		return
	}
	s.dao.Batch(mids, msg, contest, _msgSize, s.dao.SendMessage)
	return
}

func (s *Service) pubContests(contest *mdlesp.Contest) (err error) {
	var (
		mids []int64
		msg  string
	)
	if mids, msg, err = s.midsParams(contest, _pushinfo, _tpPush); err != nil {
		log.Error("pubContests s.midsParams(%+v) mids_total(%d) error(%v)", contest, len(mids), err)
		return
	}
	s.dao.Batch(mids, msg, contest, s.c.Push.PartSize, s.dao.NoticeUser)
	return
}

func (s *Service) midsParams(contest *mdlesp.Contest, link string, tp int) (mids []int64, msg string, err error) {
	var (
		userList       *favclient.SubscribersReply
		teams          []*mdlesp.Team
		homeID, awayID int64
		tMap           map[int64]string
		cursor         int64
		c              = context.Background()
	)
	ms := make(map[int64]struct{}, 1000)
	for {
		if userList, err = s.favUsers(c, contest.ID, cursor); err != nil || userList == nil || len(userList.User) == 0 {
			log.Errorc(c, "midsParams s.favUsers contestID(%v) cursor(%d) error(%+v)", contest.ID, cursor, err)
			err = nil
			break
		}
		cursor = userList.Cursor
		for _, user := range userList.User {
			if _, ok := ms[user.Mid]; ok {
				continue
			}
			ms[user.Mid] = struct{}{}
			mids = append(mids, user.Mid)
		}
	}
	if len(mids) == 0 {
		err = ecode.RequestErr
		return
	}
	tm := time.Unix(contest.Stime, 0)
	stime := tm.Format("2006-01-02 15:04:05")
	if contest.Special == _defContest {
		homeID = contest.HomeID
		awayID = contest.AwayID
		if teams, err = s.dao.Teams(c, homeID, awayID); err != nil || len(teams) == 0 {
			log.Errorc(c, "midsParams  s.dao.Teams homeID(%d) awayID(%d) error(%v)", homeID, awayID, err)
			return
		}
		tMap = make(map[int64]string, 2)
		for _, temp := range teams {
			tMap[temp.ID] = temp.Title
		}
		if tp == _tpMessage {
			msg = fmt.Sprintf(s.c.Rule.AlertBodyDefault, contest.SeasonTitle, stime, tMap[contest.HomeID], tMap[contest.AwayID], link)
		} else if tp == _tpPush {
			msg = fmt.Sprintf(s.c.Push.BodyDefault, contest.SeasonTitle, tMap[contest.HomeID], tMap[contest.AwayID], link)
		}
	} else {
		if tp == _tpMessage {
			msg = fmt.Sprintf(s.c.Rule.AlertBodySpecial, contest.SeasonTitle, stime, contest.SpecialName, link)
		} else if tp == _tpPush {
			msg = fmt.Sprintf(s.c.Push.BodySpecial, contest.SeasonTitle, contest.SpecialName, link)
		}
	}
	count := len(mids)
	log.Info("midsParams get contest cid(%d) users number(%d)", contest.ID, count)
	return
}

func (s *Service) favUsers(c context.Context, cid int64, cursor int64) (res *favclient.SubscribersReply, err error) {
	for i := 0; i < _tryTimes; i++ {
		if res, err = component.FavClient.Subscribers(c, &favclient.SubscribersReq{Type: model.TypeEsports, Oid: cid, Cursor: cursor, Size_: _favUsers}); err == nil {
			break
		}
		time.Sleep(time.Second * 3)
	}
	if err != nil {
		log.Errorc(c, "favUsers s.favClient.Subscribers cid(%d) cursor(%d) error(%v)", cid, cursor, err)
	}
	return
}

func (s *Service) arcScore() {
	var (
		id int64
		c  = context.Background()
	)
	for {
		av, err := s.dao.Arcs(c, id, _arcMaxLimit)
		if err != nil {
			log.Error("ArcScore  s.dao.Arcs ID(%d) Limit(%d) error(%v)", id, _arcMaxLimit, err)
			id = id + int64(_arcMaxLimit)
			time.Sleep(time.Second)
			continue
		}
		if len(av) == 0 {
			id = 0
			time.Sleep(time.Duration(s.c.Rule.ScoreSleep))
			continue
		}
		go s.upArcScore(c, av)
		id = av[len(av)-1].ID
		time.Sleep(time.Second)
	}
}

func (s *Service) upArcScore(c context.Context, partArcs []*mdlesp.Arc) (err error) {
	var (
		partAids  []int64
		arcsReply *arcclient.ArcsReply
	)
	for _, arc := range partArcs {
		partAids = append(partAids, arc.Aid)
	}
	if len(partAids) == 0 {
		return
	}
	if arcsReply, err = component.ArcClient.Arcs(c, &arcclient.ArcsRequest{Aids: partAids}); err != nil || arcsReply == nil {
		log.Error("upArcScore  s.arcClient.Arcs(%v) error(%v)", partAids, err)
		return
	}
	if len(arcsReply.Arcs) > 0 {
		if err = s.dao.UpArcScore(c, partArcs, arcsReply.Arcs); err != nil {
			log.Error("upArcScore  s.dao.UpArcScore arcs(%+v) error(%v)", arcsReply, err)
		}
	}
	return
}

func (s *Service) ArchivesScoreSync(ctx context.Context) (err error) {
	log.Infoc(ctx, "[Service][ArchivesScoreSync][Begin]")
	err = s.cache.Do(ctx, func(ctx context.Context) {
		s.startArchiveScoreWorker()
		log.Infoc(ctx, "[Service][ArchivesScoreSync][End]")
	})
	if err != nil {
		log.Infoc(ctx, "[Service][ArchivesScoreSync][Begin]Error, err:(%+v)", err)
	}
	return
}
