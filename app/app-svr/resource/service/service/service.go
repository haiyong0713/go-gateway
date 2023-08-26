package service

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/xstr"

	"go-common/library/stat/prom"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	pb "go-gateway/app/app-svr/resource/service/api/v1"
	pb2 "go-gateway/app/app-svr/resource/service/api/v2"
	"go-gateway/app/app-svr/resource/service/conf"
	"go-gateway/app/app-svr/resource/service/dao/abtest"
	"go-gateway/app/app-svr/resource/service/dao/ads"
	"go-gateway/app/app-svr/resource/service/dao/alarm"
	cachedao "go-gateway/app/app-svr/resource/service/dao/cache"
	"go-gateway/app/app-svr/resource/service/dao/card"
	"go-gateway/app/app-svr/resource/service/dao/cpm"
	"go-gateway/app/app-svr/resource/service/dao/entry"
	"go-gateway/app/app-svr/resource/service/dao/manager"
	"go-gateway/app/app-svr/resource/service/dao/menu"
	"go-gateway/app/app-svr/resource/service/dao/player"
	"go-gateway/app/app-svr/resource/service/dao/popups"
	"go-gateway/app/app-svr/resource/service/dao/resolution"
	"go-gateway/app/app-svr/resource/service/dao/resource"
	"go-gateway/app/app-svr/resource/service/dao/show"
	"go-gateway/app/app-svr/resource/service/dao/ugctab"
	"go-gateway/app/app-svr/resource/service/model"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	article "git.bilibili.co/bapis/bapis-go/article/service"
	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	garb "git.bilibili.co/bapis/bapis-go/garb/service"
	hmtchannelgrpc "git.bilibili.co/bapis/bapis-go/hmt-channel/interface"
	display "git.bilibili.co/bapis/bapis-go/platform/interface/display"
	recommend "git.bilibili.co/bapis/bapis-go/recommend/service"
	vedio "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/robfig/cron"
)

const (
	_updateAct      = "update"
	_archiveTable   = "archive"
	_initSidebarKey = "sidebar_%d_%d_%s"
)

var (
	_emptyResources = make(map[int]*model.Resource)
)

// Service define resource service
type Service struct {
	c             *conf.Config
	cpm           *cpm.Dao
	abtest        *abtest.Dao
	res           *resource.Dao
	ads           *ads.Dao
	alarmDao      *alarm.Dao
	show          *show.Dao
	cacheDao      *cachedao.Dao
	manager       *manager.Dao
	cardDao       *card.Dao
	menuDao       *menu.Dao
	resolutionDao *resolution.Dao
	// location rpc
	locGRPC  locgrpc.LocationClient
	garbGRPC garb.GarbClient
	// web cache
	resCache            []*model.Resource           // => resource
	asgCache            []*model.Assignment         // => assignments
	resCacheMap         map[int]*model.Resource     // resID => resource
	asgCacheMap         map[int][]*model.Assignment // resID => [ => assignments]
	defBannerCache      *model.Assignment
	videoAdsAPPCache    map[int8]map[int8]map[int8]map[string]*model.VideoAD // plat => [ => adsType] => [ => adsTarget] => aid(seasonId or typeId)
	missch              chan interface{}
	typeList            map[string]string
	resArchiveWarnCache map[int64][]*model.ResWarnInfo
	resURLWarnCache     map[string][]*model.ResWarnInfo
	posCache            map[int][]int
	// app cache
	bannerCache         map[int8]map[int][]*model.Banner
	categoryBannerCache map[int8]map[int][]*model.Banner
	bossBannerCache     map[int8]map[int][]*model.Banner
	bannerHashCache     map[int]string
	bannerLimitCache    map[int]int
	allBannerCache      map[int64]*model.Banner
	indexIcon           map[int][]*model.IndexIcon
	playIcon            *model.PlayerIcon
	playIconType        map[int32]*model.PlayerIcon
	playIconTag         map[int64]*model.PlayerIcon
	playIconArchive     map[int64]*model.PlayerIcon
	playIconPgc         map[int64]*model.PlayerIcon
	cardCache           map[int8]*model.Head
	sideBarCache        []*model.SideBar
	sideBarLimitCache   map[int64][]*model.SideBarLimit
	mineModuleCache     map[int32][]*model.ModuleInfo
	entryModuleCache    map[int32][]*model.ModuleInfo
	sideBarByModule     map[string][]*model.SideBar
	// 天马运营tab
	activemCache     map[int64][]*pb.Active
	activcovermCache map[int64]string
	menuCache        []*model.Menu
	// feed pos rec cache
	feedPosRecCache map[int64]*pb.CardPosRec
	// live
	cmtbox map[int64]*model.Cmtbox
	// abtest
	abTestCache map[string]*model.AbTest
	// PasterAIDCache
	PasterAIDCache []map[int64]int64
	// database
	archiveSub      *databus.Databus
	closeSub        bool
	closeMonitorURL bool
	// waiter
	waiter sync.WaitGroup
	// lock
	abTestLock sync.Mutex
	//pgc special cards
	specialCache map[int64]*pb.SpecialReply
	//pgc relate cards
	relateCache map[int64]*model.Relate
	//pgc id relate relate card id
	relatePgcMapCache map[int64]int64
	// audit
	auditCache map[string][]int
	// web rcmd
	webRcmd        []*pb.WebRcmd
	webRcmdCard    []*pb.WebRcmdCard
	webSpecialCard []*pb2.WebSpecialCard
	// app特殊卡
	appSpecialCard []*pb2.AppSpecialCard
	// 物料缓存
	materialMapCache map[int64]*pb2.Material
	// app特殊卡map
	appSpecialCardMap        map[string]*pb2.AppSpecialCard
	appRcmdRelatePgcCache    map[int64]*model.AppRcmd
	appRcmdRelatePgcMapCache map[int64]int64
	// 特殊卡缓存 map
	specailCardMap map[int64]*pb2.AppSpecialCard
	// activity
	actgrpc  actgrpc.ActivityClient
	dySearch []*model.DySeach
	// custom config
	//customConfigStore map[int32]map[int64]*model.CustomConfig
	//searchOgvCache .
	searchOgvCache map[int64][]int64
	// archive grpc
	arcGRPC arcgrpc.ArchiveClient
	// Article grpc
	artGRPC article.ArticleGRPCClient
	// video grpc
	vedioGRPC   vedio.VideoUpOpenClient
	HiddenCache []*pb.HiddenInfo
	// icon cache
	IconCache    map[int64]*pb.MngIcon
	PreIconCache map[int64]*pb.MngIcon
	// display service
	displayGRPC display.DisplayClient
	// params cache
	paramsCache []*pb.Param
	// cron
	cron *cron.Cron
	// app-entry-s10
	entry          *entry.Dao
	entriesInCache []*effectiveEntry
	// ugc tab
	ugctab *ugctab.Dao
	// player-panel-s10
	player        *player.Dao
	panelsInCache []*effectivePanel
	// tab_ext
	tabExt sync.Map
	// pop-ups
	popups                    *popups.Dao
	cacheFlushFrontPageTicker *time.Ticker
	// bw-list
	bwListSceneTokenCache map[string]*model.BWListWithGroup
	bwListGroupTokenCache map[string]*model.BWListWithGroup
	//recommend rpc
	recommendClient recommend.RecommendClient
	// 运营配置的个性化图标
	// key:id value:icon
	operationIcon map[int64]string
	// prom
	infoProm *prom.Prom
	// hmt channel grpc
	hmtChannelClient hmtchannelgrpc.ChannelRPCClient
	limitFreeOnline  []*model.LimitFreeInfo
}

// 警告: 所有新的周期性设定统一使用cron
// New return service object
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                   c,
		cpm:                 cpm.New(c),
		abtest:              abtest.New(c),
		res:                 resource.New(c),
		cacheDao:            cachedao.New(c),
		ads:                 ads.New(c),
		alarmDao:            alarm.New(c),
		show:                show.New(c),
		manager:             manager.New(c),
		cardDao:             card.New(c),
		menuDao:             menu.New(c),
		resolutionDao:       resolution.New(c),
		resCacheMap:         make(map[int]*model.Resource),
		asgCacheMap:         make(map[int][]*model.Assignment),
		videoAdsAPPCache:    make(map[int8]map[int8]map[int8]map[string]*model.VideoAD),
		missch:              make(chan interface{}, 10240),
		typeList:            make(map[string]string),
		resArchiveWarnCache: make(map[int64][]*model.ResWarnInfo),
		resURLWarnCache:     make(map[string][]*model.ResWarnInfo),
		bannerHashCache:     make(map[int]string),
		indexIcon:           make(map[int][]*model.IndexIcon),
		bannerCache:         make(map[int8]map[int][]*model.Banner),
		categoryBannerCache: make(map[int8]map[int][]*model.Banner),
		bossBannerCache:     make(map[int8]map[int][]*model.Banner),
		allBannerCache:      make(map[int64]*model.Banner),
		posCache:            make(map[int][]int),
		cardCache:           make(map[int8]*model.Head),
		cmtbox:              make(map[int64]*model.Cmtbox),
		sideBarLimitCache:   make(map[int64][]*model.SideBarLimit),
		abTestCache:         make(map[string]*model.AbTest),
		specialCache:        make(map[int64]*pb.SpecialReply),
		relateCache:         make(map[int64]*model.Relate),
		relatePgcMapCache:   make(map[int64]int64),
		auditCache:          make(map[string][]int),
		playIconType:        make(map[int32]*model.PlayerIcon),
		playIconTag:         make(map[int64]*model.PlayerIcon),
		playIconArchive:     make(map[int64]*model.PlayerIcon),
		//customConfigStore:   make(map[int32]map[int64]*model.CustomConfig),
		searchOgvCache:  make(map[int64][]int64),
		feedPosRecCache: make(map[int64]*pb.CardPosRec),
		operationIcon:   make(map[int64]string),
		cron:            cron.New(),
		// app-entry-s10
		entry: entry.New(c),
		// s10-ugc-tab
		ugctab: ugctab.New(c),
		// player-customized-panel
		player: player.New(c),
		// popups
		popups: popups.New(c),
		// bw-list
		bwListSceneTokenCache: make(map[string]*model.BWListWithGroup),
		bwListGroupTokenCache: make(map[string]*model.BWListWithGroup),
		// specail-card
		specailCardMap: make(map[int64]*pb2.AppSpecialCard),
		infoProm:       prom.BusinessInfoCount,
	}
	var err error
	if s.displayGRPC, err = display.NewClient(c.DisplayGRPC); err != nil {
		panic(err)
	}
	if s.locGRPC, err = locgrpc.NewClient(c.LocationGRPC); err != nil {
		panic(err)
	}
	if s.arcGRPC, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(err)
	}
	if s.artGRPC, err = article.NewClient(c.ArticleGRPC); err != nil {
		panic(err)
	}
	if s.vedioGRPC, err = vedio.NewClient(c.VedioGRPC); err != nil {
		panic(err)
	}
	if s.actgrpc, err = actgrpc.NewClient(c.ActivityGRPC); err != nil {
		panic(err)
	}
	if s.garbGRPC, err = garb.NewClient(c.GarbGRPC); err != nil {
		panic(err)
	}
	if s.recommendClient, err = recommend.NewClient(c.RecommendGRPC); err != nil {
		panic(err)
	}
	if s.hmtChannelClient, err = hmtchannelgrpc.NewClient(c.HmtChannelGRPC); err != nil {
		panic(err)
	}
	// cron
	s.initCron()
	s.cron.Start()

	// 在loadproc中 -- start
	if err := s.loadResWithRetry(); err != nil {
		panic(err)
	}
	if err := s.loadVideoAds(); err != nil {
		panic(err)
	}
	if err := s.loadBannerCahce(); err != nil {
		panic(err)
	}
	//nolint:errcheck
	s.loadTypeList()
	s.loadPlayIcon()
	s.loadCardCache()
	s.loadRelateCache()
	s.loadAudit()
	s.loadWebRcmd()
	s.loadAppRcmd()
	s.loadDySearch()
	s.loadSearchOgvConfig()
	s.loadHiddenCache()
	s.loadSideBarCache()

	// 在loadproc中 -- end
	//nolint:biligowordcheck
	go s.loadproc()

	s.loadCmtbox()
	//nolint:biligowordcheck
	go s.loadCmtboxproc()

	//nolint:biligowordcheck
	go s.cacheproc()

	if s.c.MonitorArchive {
		s.archiveSub = databus.New(c.ArchiveSub)
		s.waiter.Add(1)
		//nolint:biligowordcheck
		go s.arcConsume()
	}
	if s.c.MonitorURL {
		//nolint:biligowordcheck
		go s.checkResURL()
	}

	// 以下服务使用各自的cron
	// 多态入口
	s.startFetchEntryData()
	// ugctab
	s.FlushCache()
	// player
	s.startFetchPlayerData()
	// frontpage
	s.startCacheFlushFrontPageProc() // 定时刷新版头缓存
	return
}

func (s *Service) initCron() {
	var err error
	s.loadIconCache()
	if err = s.cron.AddFunc(s.c.Cron.LoadIconCache, s.loadIconCache); err != nil {
		panic(err)
	}
	s.loadParamCache()
	if err = s.cron.AddFunc(s.c.Cron.LoadParamCache, s.loadParamCache); err != nil {
		panic(err)
	}
	s.loadPosRec()
	if err = s.cron.AddFunc(s.c.Cron.FeedPosRecCache, s.loadPosRec); err != nil {
		panic(err)
	}
	s.loadActiveCache()
	if err = s.cron.AddFunc(s.c.Cron.LoadCardCache, s.loadActiveCache); err != nil {
		panic(err)
	}
	s.loadAppMenuCache()
	if err = s.cron.AddFunc(s.c.Cron.LoadCardCache, s.loadAppMenuCache); err != nil {
		panic(err)
	}
	s.loadTabExt()
	if err = s.cron.AddFunc(s.c.Cron.LoadTabExtCache, s.loadTabExt); err != nil {
		panic(err)
	}
	s.loadSpecialCache()
	if err = s.cron.AddFunc(s.c.Cron.LoadCardCache, s.loadSpecialCache); err != nil {
		panic(err)
	}
	s.loadBWListWithGroupCache()
	if err = s.cron.AddFunc(s.c.Cron.LoadBWListCache, s.loadBWListWithGroupCache); err != nil {
		panic(err)
	}
	s.loadMaterialCache()
	if err = s.cron.AddFunc(s.c.Cron.LoadMaterialCache, s.loadMaterialCache); err != nil {
		panic(err)
	}
	s.loadSideBarOperationIcon()
	if err = s.cron.AddFunc("@every 5s", s.loadSideBarOperationIcon); err != nil {
		panic(err)
	}
	s.LoadSpecialCardCache()
	if err = s.cron.AddFunc(s.c.Cron.LoadSpecialCardCache, s.LoadSpecialCardCache); err != nil {
		panic(err)
	}
	s.loadLimitFreeOnline()
	if err := s.cron.AddFunc("@every 20s", s.loadLimitFreeOnline); err != nil {
		panic(err)
	}
}

// loadproc is a routine load ads to cache
func (s *Service) loadproc() {
	for {
		time.Sleep(time.Duration(s.c.Reload.Ad))
		//nolint:errcheck
		s.loadResWithRetry()
		//nolint:errcheck
		s.loadVideoAds()
		//nolint:errcheck
		s.loadBannerCahce()
		//nolint:errcheck
		s.loadTypeList()

		s.loadPlayIcon()
		s.loadCardCache()
		s.loadRelateCache()
		s.loadAudit()
		s.loadWebRcmd()
		s.loadAppRcmd()
		s.loadDySearch()
		s.loadSearchOgvConfig()
		s.loadHiddenCache()
		s.loadSideBarCache()
	}
}

func (s *Service) loadSideBarOperationIcon() {
	res, err := s.manager.OpIconList(context.Background())
	if err != nil {
		log.Error("s.manager.IconList error(%+v)", err)
		return
	}
	tempOperationIcon := make(map[int64]string, len(res.Icon))
	for _, v := range res.Icon {
		tempOperationIcon[int64(v.Id)] = v.Picture
	}
	s.operationIcon = tempOperationIcon
	log.Info("loadSiderBarOperationIcon success")
}

// loadCmtboxproc is a routine load cmtbox to cache
func (s *Service) loadCmtboxproc() {
	for {
		time.Sleep(time.Second * 10)
		s.loadCmtbox()
	}
}

func (s *Service) loadTypeList() (err error) {
	var (
		args         = &arcgrpc.NoArgRequest{}
		typesTmp     *arcgrpc.TypesReply
		tmpTypeList  map[int32]*arcgrpc.Tp
		tmpTypeList2 = make(map[string]string)
	)
	if typesTmp, err = s.arcGRPC.Types(context.Background(), args); err != nil {
		log.Error("s.arcGRPC.Types() error(%v)", err)
		return
	}
	if tmpTypeList = typesTmp.GetTypes(); len(tmpTypeList) == 0 {
		log.Error("typelist len is zero")
		return
	}
	for tid, typeInfo := range tmpTypeList {
		if typeInfo == nil {
			continue
		}
		tidStr := strconv.Itoa(int(tid))
		pidStr := strconv.Itoa(int(typeInfo.Pid))
		tmpTypeList2[tidStr] = pidStr
	}
	s.typeList = tmpTypeList2
	return
}

//nolint:gocognit
func (s *Service) loadPlayIcon() {
	var (
		pis                []*model.PlayerIcon
		err                error
		overallTmp         *model.PlayerIcon
		playIconTypeTmp    = make(map[int32]*model.PlayerIcon)
		playIconTagTmp     = make(map[int64]*model.PlayerIcon)
		playIconArchiveTmp = make(map[int64]*model.PlayerIcon)
		playIconPgcTmp     = make(map[int64]*model.PlayerIcon)
	)
	if pis, err = s.res.PlayerIcon(context.TODO()); err != nil {
		log.Error("s.res.PlayerIcon() error(%v)", err)
		return
	}
	for _, pi := range pis {
		switch pi.Type {
		case model.PlayIconOverall: // 全局
			if overallTmp == nil || overallTmp.MTime < pi.MTime {
				overallTmp = pi
			}
		case model.PlayIconType: // 分区
			if typeIds, err := xstr.SplitInts(pi.TypeValue); err == nil {
				for _, typeId := range typeIds {
					if pt, ok := playIconTypeTmp[int32(typeId)]; !ok || pt.MTime < pi.MTime {
						playIconTypeTmp[int32(typeId)] = pi
					}
				}
			}
		case model.PlayIconTag: // tag
			if tagIds, err := xstr.SplitInts(pi.TypeValue); err == nil {
				for _, tagId := range tagIds {
					if pt, ok := playIconTagTmp[tagId]; !ok || pt.MTime < pi.MTime {
						playIconTagTmp[tagId] = pi
					}
				}
			}
		case model.PlayIconArchive: // 稿件
			if aids, err := xstr.SplitInts(pi.TypeValue); err == nil {
				for _, aid := range aids {
					if pt, ok := playIconArchiveTmp[aid]; !ok || pt.MTime < pi.MTime {
						playIconArchiveTmp[aid] = pi
					}
				}
			}
		case model.PlayIconPgc:
			if seasonIDs, err := xstr.SplitInts(pi.TypeValue); err == nil {
				for _, sid := range seasonIDs {
					if pt, ok := playIconPgcTmp[sid]; !ok || pt.MTime < pi.MTime {
						playIconPgcTmp[sid] = pi
					}
				}
			}
		}
	}
	s.playIcon = overallTmp
	s.playIconType = playIconTypeTmp
	s.playIconTag = playIconTagTmp
	s.playIconArchive = playIconArchiveTmp
	s.playIconPgc = playIconPgcTmp
}

func (s *Service) loadCmtbox() {
	var (
		cmtbox map[int64]*model.Cmtbox
		err    error
	)
	if cmtbox, err = s.res.Cmtbox(context.TODO()); err != nil {
		log.Error("s.res.Cmtbox() error(%v)", err)
		return
	}
	s.cmtbox = cmtbox
}

func (s *Service) loadAudit() {
	var (
		at  map[string][]int
		err error
	)
	if at, err = s.show.Audit(context.TODO()); err != nil {
		log.Error("s.show.Audit error(%v)", err)
		return
	}
	s.auditCache = at
}

func (s *Service) loadDySearch() {
	var (
		err error
	)
	if s.dySearch, err = s.res.DySearch(context.TODO()); err != nil {
		log.Error("loadDySearch error(%v)", err)
		return
	}
}

func (s *Service) loadWebRcmd() {
	var (
		rcmdTmp        []*pb.WebRcmd
		rcmdCardTmp    []*pb.WebRcmdCard
		specialCardTmp []*pb2.WebSpecialCard
		err            error
	)
	// rcmd
	if rcmdTmp, err = s.show.WebRcmd(context.TODO()); err != nil {
		log.Error("s.show.WebRcmd error(%v)", err)
		return
	}
	s.webRcmd = rcmdTmp
	// web special card
	if specialCardTmp, err = s.show.WebSpecialCard(context.TODO()); err != nil {
		log.Error("s.show.WebSpecialCard error(%v)", err)
		return
	}
	s.webSpecialCard = specialCardTmp
	// rcmd_card
	if len(specialCardTmp) > 0 {
		rcmdCardTmp = make([]*pb.WebRcmdCard, len(specialCardTmp))
		for idx, card := range specialCardTmp {
			rcmdCardTmp[idx] = &pb.WebRcmdCard{
				ID:      card.Id,
				Type:    card.Type,
				Title:   card.Title,
				Desc:    card.Desc,
				Cover:   card.Cover,
				ReType:  card.ReType,
				ReValue: card.ReValue,
			}
		}
	} else {
		if rcmdCardTmp, err = s.show.WebRcmdCard(context.TODO()); err != nil {
			log.Error("s.show.WebRcmdCard error(%v)", err)
			return
		}
	}
	s.webRcmdCard = rcmdCardTmp
}

func (s *Service) loadAppRcmd() {
	var (
		appSpecialCardTmp []*pb2.AppSpecialCard
		err               error
	)

	// app special card
	if appSpecialCardTmp, err = s.show.AppSpecialCard(context.TODO()); err != nil {
		log.Error("s.show.loadAppRcmd error(%v)", err)
		return
	}
	s.appSpecialCard = appSpecialCardTmp
	log.Info("loadappSpecialCard success")

	appSpecialCardMapTmp := make(map[string]*pb2.AppSpecialCard)
	for _, v := range appSpecialCardTmp {
		appSpecialCardMapTmp[strconv.FormatInt(v.Id, 10)] = v
	}
	s.appSpecialCardMap = appSpecialCardMapTmp
	log.Info("loadAppSpecialCardMap success")

	appRcmdList, err := s.show.AppRcmdRelatePgc(context.Background(), time.Now())
	if err != nil {
		log.Error("s.shao.loadAppRcmdRelatePgc error(%v)", err)
		return
	}
	appRcmdRelatePgcMapCacheTmp := make(map[int64]int64)
	appRcmdRelatePgcCacheTmp := make(map[int64]*model.AppRcmd)
	for _, r := range appRcmdList {
		var pgcIDs []int64
		if pgcIDs, err = xstr.SplitInts(r.PgcIds); err != nil {
			log.Error("xstr.SplitInts(%s) error(%v)", r.PgcIds, err)
			continue
		}

		for _, pgcID := range pgcIDs {
			appRcmdRelatePgcMapCacheTmp[pgcID] = r.ID
		}
		appRcmdRelatePgcCacheTmp[r.ID] = r
	}
	s.appRcmdRelatePgcMapCache = appRcmdRelatePgcMapCacheTmp
	s.appRcmdRelatePgcCache = appRcmdRelatePgcCacheTmp
	log.Info("appRcmdRelatePgcCache success %d", len(appRcmdRelatePgcCacheTmp))
}

func (s *Service) checkResURL() {
	for {
		time.Sleep(time.Duration(s.c.Reload.Ad))
		if s.closeMonitorURL {
			return
		}
		for url, resURL := range s.resURLWarnCache {
			s.alarmDao.CheckURL(url, resURL)
			time.Sleep(time.Duration(s.c.SpLimit))
		}
	}
}

// arcConsume consumer archive
func (s *Service) arcConsume() {
	defer s.waiter.Done()
	var (
		msgs = s.archiveSub.Messages()
		err  error
	)
	for {
		msg, ok := <-msgs
		if !ok {
			log.Error("s.archiveSub.Message closed")
			return
		}
		if s.closeSub {
			return
		}
		if err = msg.Commit(); err != nil {
			log.Error("arcConsume commit(%s) error(%v)", msg.Value, err)
		}

		m := &model.Message{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		if m.Table == _archiveTable {
			s.arcChan(m.Action, m.New, m.Old)
		}
	}
}

// Ping ping service
func (s *Service) Ping(c context.Context) (err error) {
	return s.res.Ping(c)
}

// Close close service
func (s *Service) Close() {
	if s.c.MonitorArchive {
		s.closeSub = true
		time.Sleep(2 * time.Second)
		s.archiveSub.Close()
		s.waiter.Wait()
	}
	s.cron.Stop()
	s.res.Close()
}

// Monitor for monitorURL
func (s *Service) Monitor(c context.Context) {
	s.closeMonitorURL = true
}

func (s *Service) addCache(d interface{}) {
	// asynchronous add rules to redis
	select {
	case s.missch <- d:
	default:
		log.Warn("cacheproc chan full")
	}
}

// cacheproc is a routine for add rules into redis.
func (s *Service) cacheproc() {
	for {
		d := <-s.missch
		//nolint:gosimple
		switch d.(type) {
		case map[string]map[int64]int64:
			v := d.(map[string]map[int64]int64)
			if err := s.ads.AddBuvidCount(context.TODO(), v); err != nil {
				log.Error("s.ads.AddBuvidCount(%v) error(%+v)", v, err)
			}
		default:
			log.Warn("cacheproc can't process the type")
		}
	}
}
