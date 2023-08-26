package view

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/component/tinker"
	"go-common/library/account/int64mid"
	"go-common/library/cache/redis"
	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/log/infoc.v2"
	"go-common/library/net/rpc/warden"
	"go-common/library/stat/prom"
	"go-common/library/sync/errgroup.v2"

	appresapi "go-gateway/app/app-svr/app-resource/interface/api/v1"
	"go-gateway/app/app-svr/app-view/interface/conf"
	abdao "go-gateway/app/app-svr/app-view/interface/dao/ab-play"
	accdao "go-gateway/app/app-svr/app-view/interface/dao/account"
	actdao "go-gateway/app/app-svr/app-view/interface/dao/act"
	addao "go-gateway/app/app-svr/app-view/interface/dao/ad"
	aidao "go-gateway/app/app-svr/app-view/interface/dao/ai"
	arcdao "go-gateway/app/app-svr/app-view/interface/dao/archive"
	aedao "go-gateway/app/app-svr/app-view/interface/dao/archive-extra"
	ahdao "go-gateway/app/app-svr/app-view/interface/dao/archive-honor"
	archiveMaterial "go-gateway/app/app-svr/app-view/interface/dao/archive-material"
	assdao "go-gateway/app/app-svr/app-view/interface/dao/assist"
	audiodao "go-gateway/app/app-svr/app-view/interface/dao/audio"
	bandao "go-gateway/app/app-svr/app-view/interface/dao/bangumi"
	"go-gateway/app/app-svr/app-view/interface/dao/bgroup"
	channeldao "go-gateway/app/app-svr/app-view/interface/dao/channel"
	"go-gateway/app/app-svr/app-view/interface/dao/checkin"
	coindao "go-gateway/app/app-svr/app-view/interface/dao/coin"
	confdao "go-gateway/app/app-svr/app-view/interface/dao/config"
	flowdao "go-gateway/app/app-svr/app-view/interface/dao/content-flow"
	contractdao "go-gateway/app/app-svr/app-view/interface/dao/contract"
	copyrightDao "go-gateway/app/app-svr/app-view/interface/dao/copyright"
	creativedao "go-gateway/app/app-svr/app-view/interface/dao/creative"
	creativeMaterial "go-gateway/app/app-svr/app-view/interface/dao/creative-material"
	creativeSpark "go-gateway/app/app-svr/app-view/interface/dao/creative-spark"
	"go-gateway/app/app-svr/app-view/interface/dao/delivery"
	dmdao "go-gateway/app/app-svr/app-view/interface/dao/dm"
	"go-gateway/app/app-svr/app-view/interface/dao/dynamic"
	elcdao "go-gateway/app/app-svr/app-view/interface/dao/elec"
	"go-gateway/app/app-svr/app-view/interface/dao/esports"
	favdao "go-gateway/app/app-svr/app-view/interface/dao/favorite"
	fkdao "go-gateway/app/app-svr/app-view/interface/dao/fawkes"
	gamedao "go-gateway/app/app-svr/app-view/interface/dao/game"
	garbdao "go-gateway/app/app-svr/app-view/interface/dao/garb"
	listenerdao "go-gateway/app/app-svr/app-view/interface/dao/listener"
	livedao "go-gateway/app/app-svr/app-view/interface/dao/live"
	locdao "go-gateway/app/app-svr/app-view/interface/dao/location"
	mngdao "go-gateway/app/app-svr/app-view/interface/dao/manager"
	musicdao "go-gateway/app/app-svr/app-view/interface/dao/music"
	natdao "go-gateway/app/app-svr/app-view/interface/dao/nat-page"
	"go-gateway/app/app-svr/app-view/interface/dao/notes"
	"go-gateway/app/app-svr/app-view/interface/dao/pgc"
	poDao "go-gateway/app/app-svr/app-view/interface/dao/player-online"
	psDao "go-gateway/app/app-svr/app-view/interface/dao/playurl"
	rcmddao "go-gateway/app/app-svr/app-view/interface/dao/recommend"
	rgndao "go-gateway/app/app-svr/app-view/interface/dao/region"
	reldao "go-gateway/app/app-svr/app-view/interface/dao/relation"
	"go-gateway/app/app-svr/app-view/interface/dao/reply"
	resDao "go-gateway/app/app-svr/app-view/interface/dao/resource"
	rscdao "go-gateway/app/app-svr/app-view/interface/dao/resource"
	searchdao "go-gateway/app/app-svr/app-view/interface/dao/search"
	seasondao "go-gateway/app/app-svr/app-view/interface/dao/season"
	sharedao "go-gateway/app/app-svr/app-view/interface/dao/share"
	silverdao "go-gateway/app/app-svr/app-view/interface/dao/silverbullet"
	steindao "go-gateway/app/app-svr/app-view/interface/dao/stein"
	thumbupdao "go-gateway/app/app-svr/app-view/interface/dao/thumbup"
	"go-gateway/app/app-svr/app-view/interface/dao/topic"
	"go-gateway/app/app-svr/app-view/interface/dao/trade"
	ugcpaydao "go-gateway/app/app-svr/app-view/interface/dao/ugcpay"
	"go-gateway/app/app-svr/app-view/interface/dao/vcloud"
	vudao "go-gateway/app/app-svr/app-view/interface/dao/videoup"
	"go-gateway/app/app-svr/app-view/interface/model"
	"go-gateway/app/app-svr/app-view/interface/model/region"
	"go-gateway/app/app-svr/app-view/interface/model/special"
	"go-gateway/app/app-svr/app-view/interface/model/view"
	arcApi "go-gateway/app/app-svr/archive/service/api"
	fkmdl "go-gateway/app/app-svr/fawkes/service/model/business"
	noteApi "go-gateway/app/app-svr/hkt-note/service/api"
	usApi "go-gateway/app/app-svr/ugc-season/service/api"

	dmapi "git.bilibili.co/bapis/bapis-go/community/interface/dm"
	replyapi "git.bilibili.co/bapis/bapis-go/community/interface/reply"
	mngApi "git.bilibili.co/bapis/bapis-go/manager/service/active"
	broadcast "git.bilibili.co/bapis/bapis-go/push/service/broadcast/v2"
	resApiV2 "git.bilibili.co/bapis/bapis-go/resource/service/v2"
	appfeaturegate "git.bilibili.co/evergarden/feature-gate/app-featuregate"

	"github.com/robfig/cron"
	"google.golang.org/grpc"
)

var (
	_groupSpecial = int64(20)
)

const (
	_broadcastAppID = "push.service.broadcast"
	_maxRetryTime   = 10
)

func (s *Service) ActSeasonKey(plat int, seasonID int64) string {
	return fmt.Sprintf("%d_%d", plat, seasonID)
}

// Service is view service
type Service struct {
	c     *conf.Config
	pHit  *prom.Prom
	pMiss *prom.Prom
	prom  *prom.Prom
	// dao
	accDao              *accdao.Dao
	arcDao              *arcdao.Dao
	favDao              *favdao.Dao
	banDao              *bandao.Dao
	elcDao              *elcdao.Dao
	rgnDao              *rgndao.Dao
	liveDao             *livedao.Dao
	assDao              *assdao.Dao
	adDao               *addao.Dao
	rscDao              *rscdao.Dao
	relDao              *reldao.Dao
	coinDao             *coindao.Dao
	audioDao            *audiodao.Dao
	actDao              *actdao.Dao
	natDao              *natdao.Dao
	thumbupDao          *thumbupdao.Dao
	gameDao             *gamedao.Dao
	dmDao               *dmdao.Dao
	aiDao               *aidao.Dao
	creativeDao         *creativedao.Dao
	search              *searchdao.Dao
	ugcpayDao           *ugcpaydao.Dao
	locDao              *locdao.Dao
	rcmdDao             *rcmddao.Dao
	vuDao               *vudao.Dao
	fawkes              *fkdao.Dao
	steinDao            *steindao.Dao
	shareDao            *sharedao.Dao
	seasonDao           *seasondao.Dao
	ahDao               *ahdao.Dao
	channelDao          *channeldao.Dao
	confDao             *confdao.Dao
	silverDao           *silverdao.Dao
	vcloudDao           *vcloud.Dao
	mngDao              *mngdao.Dao
	dynamicDao          *dynamic.Dao
	replyDao            *reply.Dao
	flowDao             *flowdao.Dao
	dmClient            dmapi.DMClient
	replyClient         replyapi.ReplyInterfaceClient
	garbDao             *garbdao.Dao
	contractDao         *contractdao.Dao
	psDao               *psDao.Dao
	onlineClient        broadcast.BroadcastVideoAPIClient
	noteClient          noteApi.HktNoteClient
	bGroupDao           *bgroup.Dao
	copyright           *copyrightDao.Dao
	esportsDao          *esports.Dao
	topicDao            *topic.Dao
	listenerDao         *listenerdao.Dao
	notesDao            *notes.Dao
	poDao               *poDao.Dao
	musicDao            *musicdao.Dao
	pgcDao              *pgc.Dao
	deliveryDao         *delivery.Dao
	checkinDao          *checkin.Dao
	tradeDao            *trade.Dao
	archiveMaterialDao  *archiveMaterial.Dao
	creativeMaterialDao *creativeMaterial.Dao
	creativeSparkDao    *creativeSpark.Dao
	// region
	tick   time.Duration
	region map[int8]map[int]*region.Region
	// elec
	allowTypeIds map[int16]struct{}
	// chan
	inCh     chan interface{}
	dmRegion map[int16]struct{}
	// mamager cache
	specialCache map[int64]*special.Card
	specialMids  map[int64]struct{}
	// view relate game from AI
	RelateGameCache map[int64]int64
	// hot aids
	hotAids         map[int64]struct{}
	slideBackupAids []int64
	// fawkes version
	FawkesVersionCache map[string]map[string]*fkmdl.Version
	// infoc ad
	infocV2Log infoc.Infoc
	// cron
	cron        *cron.Cron
	chronosConf []*view.ChronosReply
	onlineRedis *redis.Pool
	// key为plat_seasonid
	bnjActSeason      map[string]*mngApi.CommonActivityResp
	bnjArcs           map[int64]*arcApi.ViewReply
	bnjSeasons        map[int64]*usApi.View
	tinker            *tinker.ABTest
	appResourceClient appresapi.AppResourceClient
	//online black list
	onlineBlackList  map[int64]struct{}
	resDao           *resDao.Dao
	VersionMapClient *int64mid.VersionMapClient
	chronosPkgInfo   map[string][]*view.PackageInfo
	// 竖屏切全屏进story黑名单,实验结束需删除
	abPlay *abdao.Dao
	// honor
	aeDao          *aedao.Dao
	AppFeatureGate appfeaturegate.AppFeatureGate
	// 二级分区id对应一级分区id
	ArchiveTypesMap     map[int32]int32
	inspirationMaterial map[int64]view.InspirationMaterial //map[topic_id]view.InspirationMaterial
}

// New new archive
// nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:      c,
		abPlay: abdao.New(c),
		pHit:   prom.CacheHit,
		pMiss:  prom.CacheMiss,
		prom:   prom.BusinessInfoCount,
		// dao
		accDao:              accdao.New(c),
		arcDao:              arcdao.New(c),
		favDao:              favdao.New(c),
		banDao:              bandao.New(c),
		elcDao:              elcdao.New(c),
		rgnDao:              rgndao.New(c),
		liveDao:             livedao.New(c),
		assDao:              assdao.New(c),
		adDao:               addao.New(c),
		aeDao:               aedao.New(c),
		rscDao:              rscdao.New(c),
		relDao:              reldao.New(c),
		coinDao:             coindao.New(c),
		audioDao:            audiodao.New(c),
		actDao:              actdao.New(c),
		natDao:              natdao.New(c),
		thumbupDao:          thumbupdao.New(c),
		gameDao:             gamedao.New(c),
		dmDao:               dmdao.New(c),
		aiDao:               aidao.New(c),
		creativeDao:         creativedao.New(c),
		search:              searchdao.New(c),
		ugcpayDao:           ugcpaydao.New(c),
		locDao:              locdao.New(c),
		rcmdDao:             rcmddao.New(c),
		vuDao:               vudao.New(c),
		fawkes:              fkdao.New(c),
		steinDao:            steindao.New(c),
		shareDao:            sharedao.New(c),
		seasonDao:           seasondao.New(c),
		ahDao:               ahdao.New(c),
		channelDao:          channeldao.New(c),
		confDao:             confdao.New(c),
		silverDao:           silverdao.New(c),
		vcloudDao:           vcloud.New(c),
		garbDao:             garbdao.New(c),
		mngDao:              mngdao.New(c),
		contractDao:         contractdao.New(c),
		dynamicDao:          dynamic.New(c),
		replyDao:            reply.New(c),
		flowDao:             flowdao.New(c),
		psDao:               psDao.New(c),
		bGroupDao:           bgroup.New(c),
		copyright:           copyrightDao.New(c),
		esportsDao:          esports.New(c),
		topicDao:            topic.New(c),
		listenerDao:         listenerdao.New(c),
		notesDao:            notes.New(c),
		musicDao:            musicdao.New(c),
		pgcDao:              pgc.New(c),
		deliveryDao:         delivery.New(c),
		checkinDao:          checkin.New(c),
		tradeDao:            trade.New(c),
		archiveMaterialDao:  archiveMaterial.New(c),
		creativeMaterialDao: creativeMaterial.New(c),
		creativeSparkDao:    creativeSpark.New(c),
		// region
		tick:   time.Duration(c.Tick),
		region: map[int8]map[int]*region.Region{},
		// chan
		inCh:         make(chan interface{}, 1024),
		allowTypeIds: map[int16]struct{}{},
		dmRegion:     map[int16]struct{}{},
		specialMids:  map[int64]struct{}{},
		// manager
		specialCache: map[int64]*special.Card{},
		// hot aids
		hotAids: map[int64]struct{}{},
		// fawkes
		FawkesVersionCache: make(map[string]map[string]*fkmdl.Version),
		chronosConf:        make([]*view.ChronosReply, 0), //初始化
		bnjActSeason:       make(map[string]*mngApi.CommonActivityResp),
		bnjArcs:            make(map[int64]*arcApi.ViewReply),
		bnjSeasons:         make(map[int64]*usApi.View),
		// cron
		cron:           cron.New(),
		onlineRedis:    redis.NewPool(c.Redis.OnlineRedis),
		resDao:         resDao.New(c),
		poDao:          poDao.New(c),
		AppFeatureGate: appfeaturegate.GetFeatureConf(),
	}
	// infoc ad
	var err error
	if s.infocV2Log, err = infoc.New(c.InfocV2); err != nil {
		panic(err)
	}
	s.tinker = tinker.Init(s.infocV2Log, nil)
	for _, id := range c.Custom.ElecShowTypeIDs {
		s.allowTypeIds[id] = struct{}{}
	}
	for _, id := range c.DMRegion {
		s.dmRegion[id] = struct{}{}
	}
	if s.dmClient, err = dmapi.NewClient(nil); err != nil {
		panic(fmt.Sprintf("env:%s no dm grpc newClient error(%v)", env.DeployEnv, err))
	}
	if s.replyClient, err = replyapi.NewClient(nil); err != nil {
		panic(fmt.Sprintf("env:%s no reply grpc newClient error(%v)", env.DeployEnv, err))
	}
	if s.noteClient, err = noteApi.NewClient(nil); err != nil {
		panic(fmt.Sprintf("env:%s no note grpc newClient error(%v)", env.DeployEnv, err))
	}
	if s.onlineClient, err = func(cfg *warden.ClientConfig, opts ...grpc.DialOption) (broadcast.BroadcastVideoAPIClient, error) {
		client := warden.NewClient(cfg, opts...)
		conn, err := client.Dial(context.Background(), "discovery://default/"+_broadcastAppID)
		if err != nil {
			return nil, err
		}
		return broadcast.NewBroadcastVideoAPIClient(conn), nil
	}(nil); err != nil {
		panic(fmt.Sprintf("env:%s no BroadcastVideoAPIClient grpc newClient error(%v)", env.DeployEnv, err))
	}
	if s.appResourceClient, err = appresapi.NewClient(nil); err != nil {
		panic(fmt.Sprintf("env:%s no app-resource grpc newClient error(%v)", env.DeployEnv, err))
	}
	if s.VersionMapClient, err = int64mid.NewVersionClient(nil); err != nil {
		panic(fmt.Sprintf("env:%s no NewVersionClient grpc newClient error(%v)", env.DeployEnv, err))
	}
	// load data
	s.loadRegion()
	s.loadManager()
	s.loadRelateGame()
	s.loadHotCache()
	s.loadFawkes()
	go s.infocproc()
	go s.tickproc()
	go s.hotCacheproc()
	go s.loadFawkesProc()
	s.initCron()
	s.cron.Start()
	return s
}

// tickproc tick load cache.
func (s *Service) tickproc() {
	for {
		time.Sleep(s.tick)
		s.loadRegion()
		s.loadManager()
		s.loadRelateGame()
	}
}

func (s *Service) loadRegion() {
	res, err := s.rgnDao.Seconds(context.TODO())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.region = res
}

func (s *Service) getCardWithRetry() (card []*resApiV2.AppSpecialCard, err error) {
	for i := 0; i < _maxRetryTime; i++ {
		card, err = s.rscDao.GetSpecialCard(context.TODO())
		if err != nil {
			log.Error("getCardWithRetry s.loadManager() error(%+v)", err)
			continue
		}
		return
	}
	log.Error("getCardWithRetry exceed max try time s.loadManager() error(%+v)", err)
	return
}

func (s *Service) loadManager() {
	card, err := s.getCardWithRetry()
	if err != nil {
		log.Error("%+v", err)
		return
	}
	sp := make(map[int64]*special.Card)
	for _, v := range card {
		sp[v.Id] = &special.Card{
			ID:       v.Id,
			Title:    v.Title,
			Desc:     v.Desc,
			Cover:    v.Cover,
			ReType:   int(v.ReType),
			ReValue:  v.ReValue,
			Badge:    v.Corner,
			GifCover: v.Gifcover,
			Url:      v.Url,
		}
	}
	s.specialCache = sp
	midsM, err := s.creativeDao.Special(context.Background(), _groupSpecial)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	log.Info("load special mids(%+v)", midsM)
	s.specialMids = midsM
}

func (s *Service) loadRelateGame() {
	g, err := s.aiDao.Av2Game(context.TODO())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.RelateGameCache = g
}

func (s Service) relateGame(aid int64) (id int64) {
	id = s.RelateGameCache[aid]
	return
}

func (s *Service) loadHotCache() {
	tmp, unsorted, err := s.rcmdDao.Recommend(context.TODO())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	s.hotAids = tmp
	s.slideBackupAids = unsorted
}

func (s *Service) hotCacheproc() {
	for {
		time.Sleep(time.Duration(s.c.Custom.HotAidsTick))
		s.loadHotCache()
	}
}

func (s *Service) loadFawkes() {
	fv, err := s.fawkes.FawkesVersion(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	if len(fv) > 0 {
		s.FawkesVersionCache = fv
	}
}

func (s *Service) loadFawkesProc() {
	for {
		time.Sleep(time.Duration(s.c.FawkesTick))
		s.loadFawkes()
	}
}

func (s *Service) initCron() {
	s.loadChronos()
	s.loadCommonActivities()
	s.loadOnlineBlackList()
	s.loadArchiveTypes()

	var err error
	if err = s.cron.AddFunc(s.c.Cron.LoadChronos, s.loadChronos); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadCommonActivities, s.loadCommonActivities); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadOnlineManagerConfig, s.loadOnlineBlackList); err != nil {
		panic(err)
	}
	s.loadChronosPackageInfo()
	if err = s.cron.AddFunc(s.c.Cron.LoadChronos, s.loadChronosPackageInfo); err != nil {
		panic(err)
	}
	if err = s.cron.AddFunc(s.c.Cron.LoadArchiveTypes, s.loadArchiveTypes); err != nil {
		panic(err)
	}
	//灵感话题
	if err = s.cron.AddFunc(s.c.Cron.LoadInspirationTopic, s.loadCreativeSparkMaterial); err != nil {
		panic(err)
	}
}

func (s *Service) loadArchiveTypes() {
	res, err := s.arcDao.Types(context.Background())
	if err != nil {
		log.Error("s.arcDao.Types error(%+v)", err)
		return
	}
	if res == nil || len(res.Types) == 0 {
		log.Error("s.arcDao.Types error res is empty")
		return
	}
	types := make(map[int32]int32, len(res.Types))
	for i, tp := range res.Types {
		types[i] = tp.Pid
	}

	s.ArchiveTypesMap = types
	log.Info("loadArchiveTypes types(%v)", s.ArchiveTypesMap)
}

func (s *Service) loadCreativeSparkMaterial() {
	res, err := s.creativeSparkDao.GetInspirationTopics(context.Background())
	if err != nil {
		log.Error("s.creativeSparkDao.GetInspirationTopics err %+v", err)
		return
	}
	tmp := make(map[int64]view.InspirationMaterial)
	for _, v := range res {
		tmp[v.TopicId] = view.InspirationMaterial{
			Title:         v.Title,
			Url:           v.Url,
			InspirationId: v.InspirationId,
		}
	}
	s.inspirationMaterial = tmp
}

func (s *Service) loadOnlineBlackList() {
	res, err := s.resDao.FetchAllOnlineBlackList(context.Background())
	if err != nil {
		log.Error("s.resDao.FetchAllOnlineBlackList error(%+v)", err)
		return
	}
	s.onlineBlackList = res
	log.Info("loadOnlineBlackList(%v)", s.onlineBlackList)
}

func (s *Service) loadCommonActivities() {
	eg := errgroup.WithContext(context.Background())
	eg.Go(func(ctx context.Context) error {
		acts, err := s.mngDao.CommonActivities(ctx)
		memData := make(map[string]*mngApi.CommonActivityResp)
		if err != nil {
			if ecode.EqualError(ecode.NothingFound, err) {
				s.bnjActSeason = memData
				log.Warn("ActivitySeason loadCommonActivities nothing found")
			} else {
				log.Error("活动页告警 ActivitySeason loadCommonActivities err(%+v)", err)
			}
			return nil
		}
		if len(acts.GetActivities()) == 0 {
			s.bnjActSeason = memData
			log.Warn("活动页告警 ActivitySeason loadCommonActivities len(acts.Activities)=0")
			return nil
		}
		dd, _ := json.Marshal(acts.GetActivities())
		log.Warn("ActivitySeason loadCommonActivities data(%s)", dd)
		for seasonID, act := range acts.GetActivities() {
			if act.GetActivePlay() == nil {
				continue
			}
			if act.GetHdPlay() != nil {
				memData[s.ActSeasonKey(model.PlatActSeasonHD, seasonID)] = act
			}
			if act.GetAppPlay() != nil {
				memData[s.ActSeasonKey(model.PlatActSeasonApp, seasonID)] = act
			}
		}
		s.bnjActSeason = memData
		return nil
	})
	if len(s.c.ActivitySeason.Aids) > 0 {
		eg.Go(func(ctx context.Context) error {
			arcs, err := s.arcDao.Views(ctx, s.c.ActivitySeason.Aids)
			if err != nil {
				log.Error("活动页告警 ActivitySeason loadCommonActivities Views aids(%+v) error(%+v)", s.c.ActivitySeason.Aids, err)
				return nil
			}
			s.bnjArcs = arcs
			return nil
		})
	} else {
		s.bnjArcs = make(map[int64]*arcApi.ViewReply)
	}
	if s.c.ActivitySeason.Sid > 0 {
		eg.Go(func(ctx context.Context) error {
			var tmpSeason = make(map[int64]*usApi.View)
			seasons, err := s.seasonDao.Season(ctx, s.c.ActivitySeason.Sid)
			if err != nil {
				log.Error("活动页告警 ActivitySeason loadCommonActivities Season sid(%d) error(%+v)", s.c.ActivitySeason.Sid, err)
				return nil
			}
			tmpSeason[s.c.ActivitySeason.Sid] = seasons
			s.bnjSeasons = tmpSeason
			return nil
		})
	} else {
		s.bnjSeasons = make(map[int64]*usApi.View)
	}
	if err := eg.Wait(); err != nil {
		log.Error("ActivitySeason eg.wait() err(%+v)", err)
	}
}

func (s *Service) loadChronos() {
	rly, e := s.arcDao.PlayerRules(context.Background())
	if e != nil {
		log.Error("loadChronos s.arcDao.PlayerRule error(%v)", e)
		return
	}
	tmp := make([]*view.ChronosReply, 0)
	for _, v := range rly {
		if v == nil {
			continue
		}
		t := view.FormatPlayRule(v)
		if t != nil { // t==nil 不满足check条件
			tmp = append(tmp, t)
		}
	}
	s.chronosConf = tmp
	// 查询问题时需要知道当时内存的配置信息
	tmpStr, _ := json.Marshal(tmp)
	log.Info("loadChronos success %s", tmpStr)
}

func (s *Service) loadChronosPackageInfo() {
	res, err := s.arcDao.ChronosPkgInfo(context.Background())
	if err != nil {
		log.Error("loadChronosPackageInfo error(%v)", err)
		return
	}
	s.chronosPkgInfo = res
	resStr, _ := json.Marshal(res)
	log.Info("loadChronosPackageInfo success %s", resStr)
}

// Close Service
func (s *Service) Close() {
	if s.infocV2Log != nil {
		s.infocV2Log.Close()
	}
	if s.tinker != nil {
		s.tinker.Close()
	}
	s.cron.Stop()
}
