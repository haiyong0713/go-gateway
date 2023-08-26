package search

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/library/cache/memcache"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/sync/errgroup.v2"
	xtime "go-common/library/time"

	pb "go-gateway/app/app-svr/app-feed/admin/api/search"
	"go-gateway/app/app-svr/app-feed/admin/bvav"
	"go-gateway/app/app-svr/app-feed/admin/conf"
	accdao "go-gateway/app/app-svr/app-feed/admin/dao/account"
	arcdao "go-gateway/app/app-svr/app-feed/admin/dao/archive"
	"go-gateway/app/app-svr/app-feed/admin/dao/article"
	"go-gateway/app/app-svr/app-feed/admin/dao/game"
	"go-gateway/app/app-svr/app-feed/admin/dao/manager"
	pgcdao "go-gateway/app/app-svr/app-feed/admin/dao/pgc"
	"go-gateway/app/app-svr/app-feed/admin/dao/search"
	"go-gateway/app/app-svr/app-feed/admin/dao/show"
	"go-gateway/app/app-svr/app-feed/admin/dataplat"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	gameModel "go-gateway/app/app-svr/app-feed/admin/model/game"
	searchModel "go-gateway/app/app-svr/app-feed/admin/model/search"
	showModel "go-gateway/app/app-svr/app-feed/admin/model/show"

	accountGRPC "git.bilibili.co/bapis/bapis-go/account/service"
	relationGRPC "git.bilibili.co/bapis/bapis-go/account/service/relation"

	Log "go-gateway/app/app-svr/app-feed/admin/util"

	permit "go-common/library/net/http/blademaster/middleware/permit"
	permitPb "go-common/library/net/http/blademaster/middleware/permit/api"

	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
)

const (
	// permit consts, align with php codes
	APPLY_PRIV       = 3
	CHECK_PRIV       = 4
	APPLY_CHECK_PRIV = 5
)

var (
	ctx = context.TODO()
)

// Service is search service
type Service struct {
	dao           *search.Dao
	showDao       *show.Dao
	accDao        *accdao.Dao
	cron          *cron.Cron
	HotFre        string
	DarkFre       string
	RcmdFre       string
	BrandFre      string
	SidebarCroFre string
	GameFre       string
	ChannelFre    string
	gameDao       *game.Dao
	pgcDao        *pgcdao.Dao
	managerDao    *manager.Dao
	articleDao    *article.Dao
	arcDao        *arcdao.Dao
	c             *conf.Config
	// client
	auth           permitPb.PermitClient
	accClient      accountGRPC.AccountClient
	relationClient relationGRPC.RelationClient

	// cache
	RcmdAppCache        []*searchModel.SpreadConfig
	RcmdWebCache        []*searchModel.SpreadConfig
	RcmdCache           []*searchModel.SpreadConfig
	BrandBlacklistCache []*searchModel.BrandBlacklistItem
	GameCache           map[string]string
	EntryGameCache      map[string]*gameModel.EntryInfo
	ChannelIdCache      []int64
}

const (
	_HotPubState           = "tianma_search_hot_state"
	_HotPubValue           = "tianma_search_hot_value"
	_HotPubSearchState     = "tianma_search_hot_search_state"
	_DarkPubState          = "tianma_search_dark_state"
	_DarkPubValue          = "tianma_search_dark_value"
	_DarkPubSearchState    = "tianma_search_dark_search_state"
	_HotAutoPubState       = "tianma_search_auto_hot_state"
	_DarkAutoPubState      = "tianma_search_auto_dark_state"
	_LastSearchSyncValue   = "tianma_search_last_sync_state"
	_LastSearchOnlineValue = "tianma_search_last_online_state"
	_HotShowUnpub          = 0
	_HotShowPub            = 1
	_HotShowUnUp           = 2
	_DarkShowUnpub         = 0
	_DarkShowPub           = 1
	_DarkShowUnUp          = 2
	SearchInterFire        = 2 // 搜索干预 小火苗
	SearchInterNow         = 4 // 搜索干预 最新
	SearchInterHot         = 5 // 搜索干预 最热
	SearchInterLive        = 7 // 搜索干预 直播中
	HotwordFromDBAll       = 0 // 默认，现在和未来生效的干预
	HotwordFromDBOnline    = 1 // 正在生效的干预
	HotwordFromDBFuture    = 2 // 未来生效的干预
)

// New new a search service
func New(c *conf.Config) (s *Service) {
	s = &Service{
		dao:           search.New(c),
		showDao:       show.New(c),
		cron:          cron.New(),
		HotFre:        c.Cfg.HotCroFre,
		DarkFre:       c.Cfg.DarkCroFre,
		RcmdFre:       c.Cfg.RcmdCroFre,
		BrandFre:      c.Cfg.BrandCroFre,
		SidebarCroFre: c.Cfg.SidebarCroFre,
		GameFre:       c.Cfg.GameCroFre,
		ChannelFre:    c.Cfg.ChannelCroFre,
		accDao:        accdao.New(c),
		gameDao:       game.New(c),
		managerDao:    manager.New(c),
		articleDao:    article.New(c),
		arcDao:        arcdao.New(c),
		auth:          permit.New2(nil).PermitClient,
		c:             c,
	}
	var err error
	if s.accClient, err = accountGRPC.NewClient(nil); err != nil {
		panic(err)
	}
	if s.relationClient, err = relationGRPC.NewClient(nil); err != nil {
		panic(err)
	}

	s.pgcDao, err = pgcdao.New(c)
	if err != nil {
		panic(err)
	}

	//nolint:errcheck,biligowordcheck
	go s.StartCronTask()

	//nolint:biligowordcheck
	//go s.LoadRcmd()
	//nolint:biligowordcheck
	go s.LoadBrandBlacklist()
	//nolint:biligowordcheck
	go s.LoadGame()
	//nolint:biligowordcheck
	go s.LoadChannel()
	return
}

// CrontLoad search box history
func (s *Service) StartCronTask() (err error) {
	if err = s.cron.AddFunc(s.HotFre, s.LoadHot); err != nil {
		log.Error("searchSrv.StartCronTask AddFunc LoadHot error(%v)", err)
		panic(err)
	}

	if err = s.cron.AddFunc(s.DarkFre, s.LoadDark); err != nil {
		log.Error("searchSrv.StartCronTask AddFunc LoadDark error(%v)", err)
		panic(err)
	}

	if err = s.cron.AddFunc(s.RcmdFre, s.LoadRcmd); err != nil {
		log.Error("searchSrv.StartCronTask AddFunc LoadRcmd error(%v)", err)
		panic(err)
	}

	if err = s.cron.AddFunc(s.BrandFre, s.LoadBrandBlacklist); err != nil {
		log.Error("searchSrv.StartCronTask AddFunc LoadBrandBlacklist error(%v)", err)
		panic(err)
	}

	if err = s.cron.AddFunc(s.SidebarCroFre, s.RemoveExpiredSidebar); err != nil {
		log.Error("searchSrv.RemoveExpiredSidebar AddFunc LoadBrandBlacklist error(%v)", err)
		panic(err)
	}

	if err = s.cron.AddFunc(s.GameFre, s.LoadGame); err != nil {
		log.Error("searchSrv.StartCronTask AddFunc LoadGame error(%v)", err)
		panic(err)
	}

	if err = s.cron.AddFunc(s.ChannelFre, s.LoadChannel); err != nil {
		log.Error("searchSrv.StartCronTask AddFunc LoadChannel error(%v)", err)
		panic(err)
	}

	s.cron.Start()
	return
}

// LoadHot crontab auto load hot word
func (s *Service) LoadHot() {
	var (
		err    error
		status bool
	)
	timeTwelve := time.Now().Format("2006-01-02 ") + "12:00:00"
	timeTwelveStr, _ := s.parseTime(timeTwelve, "2006-01-02 15:04:05")
	timeZero := time.Now().Format("2006-01-02 ") + "00:00:00"
	timeZeroStr, _ := s.parseTime(timeZero, "2006-01-02 15:04:05")
	log.Info("searchSrv.LoadHot Auto LoadHot Start!")
	if time.Now().Unix() == timeZeroStr.Unix() {
		// 0点会自动发布一次数据
		if err = s.SetHotPub(ctx, "crontabLoadHot", 0); err != nil {
			log.Error("searchSrv.LoadHot SetHotPub error(%v)", err)
			return
		}
		log.Info("searchSrv.LoadHot Auto LoadHot Success! 00:00 clock")
	} else if time.Now().Unix() >= timeTwelveStr.Unix() {
		log.Info("searchSrv.LoadHot Auto LoadHot Time > (%v)", timeTwelveStr)
		if status, err = s.isTodayAutoPubHot(ctx); err != nil {
			log.Error("searchSrv.LoadHot isTodayAutoPubHot error(%v)", err)
			return
		}
		log.Info("searchSrv.LoadHot Auto LoadHot Publish Status = (%v)", status)
		if status {
			return
		}
		if err = s.SetHotPub(ctx, "crontabLoadHot", 0); err != nil {
			log.Error("searchSrv.LoadHot SetHotPub error(%v)", err)
			return
		}
		log.Info("searchSrv.LoadHot Auto LoadHot Success! more than 12:00 clock")
	}
}

// LoadDark crontab auto load dark word
func (s *Service) LoadDark() {
	var (
		status bool
		err    error
	)
	log.Info("searchSrv.LoadDark Auto LoadDark Start!")
	timeTwelve := time.Now().Format("2006-01-02 ") + "12:00:00"
	timeTwelveStr, _ := s.parseTime(timeTwelve, "2006-01-02 15:04:05")
	timeZero := time.Now().Format("2006-01-02 ") + "00:00:00"
	timeZeroStr, _ := s.parseTime(timeZero, "2006-01-02 15:04:05")
	log.Info("searchSrv.LoadDark Auto LoadDark Start!")
	if time.Now().Unix() == timeZeroStr.Unix() {
		// 0点会自动发布一次数据
		if err = s.SetDarkPub(ctx, "crontabLoadDark", 0); err != nil {
			log.Error("searchSrv.LoadDark SetDarkPub error(%v)", err)
			return
		}
		log.Info("searchSrv.LoadDark Auto LoadDark Success! 00:00 clock")
	} else if time.Now().Unix() >= timeTwelveStr.Unix() {
		log.Info("searchSrv.LoadDark Auto LoadDark Time > (%v)", timeTwelveStr)
		if status, err = s.isTodayAutoPubDark(ctx); err != nil {
			log.Error("searchSrv.LoadDark isTodayAutoPubDark error(%v)", err)
			return
		}
		log.Info("searchSrv.LoadDark Auto LoadDark Publish Status = (%v)", status)
		if status {
			return
		}
		if err = s.SetDarkPub(ctx, "crontabLoadDark", 0); err != nil {
			log.Error("searchSrv.LoadDark SetDarkPub error(%v)", err)
			return
		}
		log.Info("searchSrv.LoadDark Auto LoadDark Success!")
	}
}

func (s *Service) LoadBrandBlacklist() {
	const MAX_PS = 1000
	var (
		total int
		tmp   *searchModel.BrandBlackListListResp
		resp  *searchModel.BrandBlackListListResp
		param = &searchModel.BrandBlacklistListReq{
			Pn: 1,
			Ps: MAX_PS,
		}
		err error
	)
	resp, err = s.BrandBlacklistList(context.Background(), param)
	if err != nil || resp.Page == nil || resp.Page.Total == 0 {
		return
	}

	cache := make([]*searchModel.BrandBlacklistItem, 0, total)
	//nolint:gosimple
	for _, item := range resp.List {
		cache = append(cache, item)
	}
	if resp.Page.Total >= MAX_PS {
		pages := int(math.Ceil(float64(total) / float64(MAX_PS)))
		for pn := 2; pn <= pages; pn++ {
			param.Pn = pn
			param.Ps = MAX_PS
			if tmp, err = s.BrandBlacklistList(context.Background(), param); err != nil {
				return
			}
			cache = append(cache, tmp.List...)
		}
	}

	s.BrandBlacklistCache = cache
}

func (s *Service) LoadGame() {
	SearchWebs := make([]*showModel.SearchWeb, 0)
	w := map[string]interface{}{
		"deleted":   common.NotDeleted,
		"check":     common.Pass,
		"card_type": common.WebSearchGame,
	}
	query := s.showDao.DB.Model(&showModel.SearchWeb{})
	if err := query.Where(w).Order("`id` DESC").Find(&SearchWebs).Error; err != nil {
		log.Error("searchSvc.OpenSearchWebList Find error(%v)", err)
		return
	}

	ctx := context.TODO()
	eg := errgroup.WithCancel(ctx)

	cache := make(map[string]string)
	lock := sync.RWMutex{}
	for _, v := range SearchWebs {
		var (
			gameId int64
			title  string
		)
		// 此处不同线程读取的v是pointer类型，且始终是同一个变量，遍历可能致使不同goroutine读取到相同的v
		cardValue := v.CardValue
		if val, ok := s.GameCache[cardValue]; ok {
			lock.Lock()
			cache[cardValue] = val
			lock.Unlock()
			continue
		}
		gameId, err := strconv.ParseInt(cardValue, 10, 64)
		if err != nil {
			log.Error("LoadGame ParseInt invalid gameId(%v), error(%v)", cardValue, err)
			continue
		}

		eg.Go(func(ctx context.Context) (gameErr error) {
			title, gameErr = s.gameDao.SearchGame(ctx, gameId)
			if gameErr != nil || title == "" {
				return nil
			}

			lock.Lock()
			cache[cardValue] = title
			lock.Unlock()

			return
		})
		time.Sleep(50 * time.Millisecond)
	}

	entryCache := make(map[string]*gameModel.EntryInfo)
	entryLock := sync.RWMutex{}
	for _, v := range SearchWebs {
		var (
			gameId int64
			info   *gameModel.EntryInfo
		)
		cardValue := v.CardValue
		if val, ok := s.EntryGameCache[cardValue]; ok {
			entryLock.Lock()
			entryCache[cardValue] = val
			entryLock.Unlock()
			continue
		}
		gameId, err := strconv.ParseInt(cardValue, 10, 64)
		if err != nil {
			log.Error("LoadGame ParseInt invalid gameId(%v), error(%v)", cardValue, err)
			continue
		}

		eg.Go(func(ctx context.Context) (gameErr error) {
			info, gameErr = s.gameDao.GameEntryInfo(ctx, gameId)
			if gameErr != nil || info == nil {
				return nil
			}

			entryLock.Lock()
			entryCache[cardValue] = info
			entryLock.Unlock()

			return
		})
		time.Sleep(50 * time.Millisecond)
	}

	if egErr := eg.Wait(); egErr != nil {
		log.Error("search LoadGame got error(%v)", egErr)
		return
	}
	s.GameCache = cache
	s.EntryGameCache = entryCache
}

func (s *Service) LoadChannel() {
	ids, _, err := s.OpenChannelIds(0, 0)
	if err != nil {
		log.Error("search LoadChannel got error(%v)", err)
		return
	}
	s.ChannelIdCache = ids
}

func (s *Service) LoadRcmd() {
	var (
		now = time.Now()
		c   = context.Background()
		eg  = errgroup.WithCancel(c)
	)

	eg.Go(func(ctx context.Context) (appErr error) {
		// load rcmd app cache (valid in 30 minutes)
		paramApp := &searchModel.RecomParam{
			Pn:      1,
			Ps:      1,
			StartTs: now.Unix(),
			EndTs:   now.Add(30 * time.Minute).Unix(),
			Plat:    PLAT_APP,
		}
		var ret []*searchModel.SpreadConfig
		if ret, appErr = s.batchOpenRcmd(ctx, paramApp); appErr != nil {
			log.Error("searchSvr.LoadRcmd rcmd app cache appErr(%v)", appErr)
			return
		}
		s.RcmdAppCache = ret
		return
	})

	eg.Go(func(ctx context.Context) (webErr error) {
		// load rcmd web cache (valid in 30 minutes)
		paramWeb := &searchModel.RecomParam{
			Pn:      1,
			Ps:      1,
			StartTs: now.Unix(),
			EndTs:   now.Add(30 * time.Minute).Unix(),
			Plat:    PLAT_WEB,
		}
		var ret []*searchModel.SpreadConfig
		if ret, webErr = s.batchOpenRcmd(ctx, paramWeb); webErr != nil {
			log.Error("searchSvr.LoadRcmd rcmd web cache webErr(%v)", webErr)
			return
		}
		s.RcmdWebCache = ret
		return
	})

	eg.Go(func(ctx context.Context) (rcmdErr error) {
		// load rcmd cache (valid now)
		param := &searchModel.RecomParam{Pn: 1, Ps: 1}
		var ret []*searchModel.SpreadConfig
		if ret, rcmdErr = s.batchOpenRcmd(ctx, param); rcmdErr != nil {
			log.Error("searchSvr.LoadRcmd rcmd cache rcmdErr(%v)", rcmdErr)
			return
		}
		s.RcmdCache = ret
		return
	})

	if err := eg.Wait(); err != nil {
		log.Error("service.LoadRcmd errgroup wait error(%v)", err)
	}
}

func (s *Service) batchOpenRcmd(c context.Context, param *searchModel.RecomParam) (ret []*searchModel.SpreadConfig, err error) {
	const MAX_PS = 1000
	var (
		total int
		tmp   *searchModel.RecomRes
	)
	if total, err = s.OpenRecommendCount(c, param); err != nil {
		return
	}

	pages := int(math.Ceil(float64(total) / float64(MAX_PS)))
	ret = make([]*searchModel.SpreadConfig, 0, total)
	for pn := 1; pn <= pages; pn++ {
		param.Pn = pn
		param.Ps = MAX_PS
		if tmp, err = s.OpenRecommend(c, param); err != nil {
			return
		}
		ret = append(ret, tmp.Item...)
	}

	return
}

// isTodayAutoPubHot is today publish hot word
func (s *Service) isTodayAutoPubHot(c context.Context) (status bool, err error) {
	var (
		flag bool
		date string
	)
	if flag, date, err = s.dao.GetSearchAuditStat(c, _HotAutoPubState); err != nil {
		log.Error("searchSrv.isTodayAutoPubHot GetPubState error(%v)", err)
		return
	}
	// 已发布 且是今天发布的数据 则证明今天发布过
	if flag && date == time.Now().Format("2006-01-02") {
		return true, nil
	}
	return
}

// isTodayAutoPubHot is today publish hot word
func (s *Service) isTodayAutoPubDark(c context.Context) (status bool, err error) {
	var (
		flag bool
		date string
	)
	if flag, date, err = s.dao.GetSearchAuditStat(c, _DarkAutoPubState); err != nil {
		log.Error("searchSrv.isTodayAutoPubDark GetPubState error(%v)", err)
		return
	}
	// 已发布 且是今天发布的数据 则证明今天发布过
	if flag && date == time.Now().Format("2006-01-02") {
		return true, nil
	}
	return
}

// parseTime parse string to unix timestamp
func (s *Service) parseTime(t string, timeLayout string) (theTime time.Time, err error) {
	// timeLayout := "2006-01-02 15:04:05"
	// timeLayout := "2006-01-02" //转化所需模板

	loc, _ := time.LoadLocation("Local") // 重要：获取时区
	// 使用模板在对应时区转化为time.time类型
	if theTime, err = time.ParseInLocation(timeLayout, t, loc); err != nil {
		log.Error("searchSrv.parseTime ParseInLocation(%v) error(%v)", t, err)
		return
	}
	return
}

// GetSearchValue 获取搜索的数据
func (s *Service) GetSearchValue(date string, blackSlice []string) (his []searchModel.History, err error) {
	if err = s.dao.DB.Model(&searchModel.History{}).
		Where("atime = ?", date).Where("searchword not in (?)", blackSlice).
		Where("deleted = ?", searchModel.NotDelete).
		Order("pv desc").
		Find(&his).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.GetSearchValue error(%v)", err)
		return
	}
	return
}

// GetSearHisValue 获取搜索热词的数据
func (s *Service) GetSearHisValue(blackSlice []string) (his []searchModel.History, err error) {
	var hisTmp searchModel.History
	if err = s.dao.DB.Model(&searchModel.History{}).
		Where("deleted = ?", searchModel.NotDelete).Order("atime desc").Limit(1).
		First(&hisTmp).Error; err != nil {
		log.Error("searchSrv.GetSearchHisValue Last Day error(%v)", err)
		return
	}
	dao := s.dao.DB.Model(&searchModel.History{}).
		Where("atime = ?", hisTmp.Atime)
	if len(blackSlice) > 0 {
		dao = dao.Where("searchword not in (?)", blackSlice)
	}
	if err = dao.Where("deleted = ?", searchModel.NotDelete).
		Order("pv desc").Find(&his).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.GetSearchHisValue error(%v)", err)
		return
	}
	return
}

// HotwordFromDB 从DB中取有效数据
//
//nolint:gocognit
func (s *Service) HotwordFromDB(date string, timeType int) (resList []searchModel.Intervene, searchCount int, err error) {
	var (
		black               []searchModel.Black
		blackSlice          []string
		originalHistoryList []searchModel.History
		manualWordList      []searchModel.Intervene
		allResultList       []searchModel.Intervene
	)

	if black, err = s.BlackAll(); err != nil {
		log.Error("searchSrv.HotList Black error(%v)", err)
		return
	}
	for _, v := range black {
		blackSlice = append(blackSlice, v.Searchword)
	}

	if timeType == HotwordFromDBFuture {
		// 未来生效的干预，不需要返回线上搜索结果
		originalHistoryList = []searchModel.History{}
	} else {
		// 搜索推送过来的未干预的数据，当天的
		originalHistoryList, err = s.GetSearchValue(date, blackSlice)
		if err != nil {
			log.Error("searchSrv.HotwordFromDB error(%v)", err)
			return
		}
		searchCount = len(originalHistoryList)

		// 如果是取今天发布的数据 且没有取到 则以昨天的为准
		if time.Now().Format("2006-01-02") == date && len(originalHistoryList) == 0 {
			// 如果 当天的搜索热词暂未同步过来 则取昨天的搜索热词
			if originalHistoryList, err = s.GetSearHisValue(blackSlice); err != nil {
				log.Error("searchSrv.HotList GetHotPubLog error(%v)", err)
				return
			}
		}
	}

	// 未结束的运营干预词，生效中的，也就是从管理后台添加的，position!=-1 的
	query := s.dao.DB.Model(&searchModel.Intervene{})
	// 需要准确的当前时间
	dateStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
	switch timeType {
	case HotwordFromDBOnline:
		// 只返回正在生效的干预
		query = query.Where("etime >= ?", dateStr).
			Where("stime <= ?", dateStr).
			Where("rank >= ?", 1)
	case HotwordFromDBFuture:
		// 只返回未来生效的干预
		query = query.Where("stime >= ? OR rank < ?", dateStr, 1)
	default:
		// 返回以上两种全部的干预
		query = query.Where("etime >= ?", dateStr)
	}
	query = query.Where("searchword not in (?)", blackSlice).
		Where("deleted = ?", searchModel.NotDelete).
		Order("rank asc, mtime desc")
	if err = query.Find(&manualWordList).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.HotList Intervene error(%v)", err)
		return
	}

	var manualWordMap = make(map[string]searchModel.Intervene, len(manualWordList))
	listLength := 20
	if len(manualWordList) > listLength {
		listLength = len(manualWordList)
	}
	allResultList = make([]searchModel.Intervene, listLength+len(originalHistoryList))
	for i, v := range manualWordList {
		manualWordMap[v.Searchword] = v

		// 按rank顺序拼装人工干预结果
		if timeType == HotwordFromDBFuture {
			allResultList[i] = v
		} else {
			if v.Rank-1 < 0 {
				// 说明是待定

				//nolint:makezero
				allResultList = append(allResultList, v)
			} else {
				allResultList[v.Rank-1] = v
			}
		}
	}
	// todo ci ll yr
	var originalHistoryMap = make(map[string]searchModel.History, len(originalHistoryList))
	var originalHistoryListIndex = 0
	for _, v := range originalHistoryList {
		originalHistoryMap[v.Searchword] = v
	}

	// 搜索历史词，就是搜索推送过来的未干预数据，默认position为-1
	// 搜索的排序规则未知，大致按照pv降序
	// 用搜索历史次，按照pv降序，填补allResultList的空缺
	for i, v := range allResultList {
		if v.Searchword == "" {
			// 说明是空缺，说明没有干预，直接展示，添加搜索的热词数据的返回结果
			// 顺序找到下一个没有干预的词
			originalHistoryItem := searchModel.History{}
			for {
				if originalHistoryListIndex >= len(originalHistoryList) {
					// 接下来没有搜索推送过来的词了，终止
					originalHistoryItem = searchModel.History{}
					break
				}
				originalHistoryItem = originalHistoryList[originalHistoryListIndex]
				originalHistoryListIndex += 1
				// 检查看看想要添加的词，有没有已经在干预列表内
				if _, ok := manualWordMap[originalHistoryItem.Searchword]; ok {
					// 在干预列表内，继续寻找
					continue
				} else {
					// 找到了，继续进行
					break
				}
			}
			if originalHistoryItem.Searchword == "" {
				// 说明没找到可以填补空位的词
				continue
			}
			item := searchModel.Intervene{
				ID:         0, // 系统词统一不返回 id，防止和干预表的 id 混淆
				Searchword: originalHistoryItem.Searchword,
				Rank:       i + 1, // 填入实际展示位置
				Tag:        originalHistoryItem.Tag,
				Pv:         originalHistoryItem.Pv,
				Type:       1, // 搜索词默认为 1
				Uv:         originalHistoryItem.Uv,
				Click:      originalHistoryItem.Click,
				ShowWord:   originalHistoryItem.Searchword,
			}
			allResultList[i] = item
		} else {
			// 标记为正在干预
			allResultList[i].Intervene = 1
			// 有词，说明正在干预中的词，复制搜索的数据到干预的词中
			// todo ci llyr
			originalHistoryItem, ok := originalHistoryMap[v.Searchword]
			if ok {
				// 如果搜索历史里面有这个干预的词，复制搜索的数据到干预的词中
				allResultList[i].Pv = originalHistoryItem.Pv
				allResultList[i].Uv = originalHistoryItem.Uv
				allResultList[i].Click = originalHistoryItem.Click
			}
		}
	}

	// 清除所有空词
	for _, item := range allResultList {
		if item.Searchword != "" {
			resList = append(resList, item)
		}
	}

	return
}

// BlackAll black list
func (s *Service) BlackList(lp *searchModel.BlackListParam) (pager *searchModel.BlackListPager, err error) {
	pager = &searchModel.BlackListPager{
		Page: common.Page{
			Num:  lp.Pn,
			Size: lp.Ps,
		},
	}
	w := map[string]interface{}{
		"deleted": common.NotDeleted,
	}
	query := s.dao.DB.Model(&searchModel.Black{}).Where(w)
	if lp.Searchword != "" {
		query = query.Where("searchword like ?", "%"+lp.Searchword+"%")
	}
	if err = query.Count(&pager.Page.Total).Error; err != nil {
		log.Error("BlackAll count error(%v)", err)
		return
	}
	if err = query.Order("`id` DESC").Offset((lp.Pn - 1) * lp.Ps).Limit(lp.Ps).Find(&pager.Item).Error; err != nil {
		log.Error("BlackAll query error(%v)", err)
		return
	}
	return
}

// GetDarkValue 获取搜索热词的数据
func (s *Service) GetDarkValue(blackSlice []string) (his []searchModel.Dark, err error) {
	var darkTmp searchModel.Dark
	if err = s.dao.DB.Model(&searchModel.Dark{}).
		Where("deleted = ?", searchModel.NotDelete).Order("atime desc").Limit(1).
		First(&darkTmp).Error; err != nil {
		log.Error("searchSrv.GetDarkValue Last Day error(%v)", err)
		return
	}
	dao := s.dao.DB.Model(&searchModel.Dark{}).
		Where("atime = ?", darkTmp.Atime)
	if len(blackSlice) > 0 {
		dao = dao.Where("searchword not in (?)", blackSlice)
	}
	if err = dao.Where("deleted = ?", searchModel.NotDelete).Find(&his).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.GetDarkValue error(%v)", err)
		return
	}
	return
}

// DarkwordFromDB 从DB中取有效数据
func (s *Service) DarkwordFromDB(date string) (darkValue []searchModel.Dark, searchCount int, err error) {
	var (
		black      []searchModel.Black
		blackSlice []string
		dark       []searchModel.Dark
	)
	if black, err = s.BlackAll(); err != nil {
		log.Error("searchSrv.DarkwordFromDB BlackList error(%v)", err)
	}
	for _, v := range black {
		blackSlice = append(blackSlice, v.Searchword)
	}
	if err = s.dao.DB.Model(&searchModel.Dark{}).Where("deleted = ?", searchModel.NotDelete).
		Where("atime = ?", date).Where("deleted = ?", searchModel.NotDelete).Find(&dark).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.DarkwordFromDB Find error(%v)", err)
		return
	}
	searchCount = len(dark)
	if err = s.dao.DB.Model(&searchModel.Dark{}).Where("deleted = ?", searchModel.NotDelete).
		Where("atime = ?", date).Where("searchword not in (?)", blackSlice).
		Where("deleted = ?", searchModel.NotDelete).Find(&dark).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.DarkwordFromDB Find error(%v)", err)
		return
	}
	// 若搜索没有推黑马词过来 则以昨天的数据为准
	if time.Now().Format("2006-01-02") == date && len(dark) == 0 {
		if dark, err = s.GetDarkValue(blackSlice); err != nil {
			log.Error("searchSrv.DarkwordFromDB GetDarkPubLog error(%v)", err)
			return
		}
	}
	m := make(map[string]bool)
	for _, val := range dark {
		if _, ok := m[val.Searchword]; !ok {
			m[val.Searchword] = true
			darkValue = append(darkValue, val)
		}
	}
	return
}

// OpenHotList open hotword list
func (s *Service) OpenHotList(c *bm.Context) (hotout []searchModel.Intervene, err error) {
	var (
		hot []searchModel.Intervene
	)
	if hot, err = s.GetHotPub(c); err != nil {
		log.Error("searchSrv.OpenHotList GetHotPub error(%v)", err)
		return
	}
	cTime := time.Now().Unix()
	inter := map[string]bool{}
	for _, v := range hot {
		if v.Intervene == 1 && cTime >= v.Stime.Time().Unix() && cTime <= v.Etime.Time().Unix() {
			// 运营词
			inter[v.Searchword] = true
		}
	}
	for _, v := range hot {
		if v.Intervene == 0 {
			// 如果运营词已存在 则以运营词为准
			if _, flag := inter[v.Searchword]; flag {
				continue
			}
			// <1 是ai的数据 直接添加
			v.Rank = -1
			hotout = append(hotout, v)
		} else if cTime >= v.Stime.Time().Unix() && cTime <= v.Etime.Time().Unix() {
			hotout = append(hotout, v)
		}
	}

	// 记录上次AI同步的时间
	lastSync := searchModel.LastTimeSyncItem{
		LastTime: time.Now().Unix(),
	}
	itemJSON := &memcache.Item{
		Key:        _LastSearchSyncValue,
		Flags:      memcache.FlagJSON,
		Object:     lastSync,
		Expiration: 0,
	}
	if err = s.dao.MC.Set(c, itemJSON); err != nil {
		log.Error("searchSrv.OpenHotList s.dao.MC.Set error(%v)", err)
		err = nil
	}

	return
}

// HotList hotword list
func (s *Service) HotList(c *bm.Context, t string) (hotout searchModel.HotwordOut, err error) {
	var (
		dateStamp  time.Time
		todayStamp time.Time
		flag       bool
		hot        []searchModel.Intervene
		date       string
	)
	defer func() {
		// todo: 好像前端也用不到 bvid
		if err == nil && len(hotout.Hotword) != 0 {
			for k, v := range hotout.Hotword {
				if v.GotoType == common.SeaHotGoToArch {
					if hotout.Hotword[k].BvID, err = bvav.AvStrToBvStr(v.GotoValue); err != nil {
						hotout.Hotword[k].BvID = err.Error()
						err = nil
						log.Error("searchSrv.HotList AvStrToBvStr error(%v)", err)
					}
				}
			}
		}
	}()
	if dateStamp, err = s.parseTime(t, "2006-01-02"); err != nil {
		return
	}
	today := time.Unix(time.Now().Unix(), 0).Format("2006-01-02")
	if todayStamp, err = s.parseTime(today, "2006-01-02"); err != nil {
		return
	}
	if flag, date, err = s.dao.GetSearchAuditStat(c, _HotPubState); err != nil {
		log.Error("searchSrv.HotList GetPublishCache error(%v)", err)
		return
	}
	// 过去的时间 则直接从日志中取数据
	// todo: 已经没有这个功能了
	if dateStamp.Unix() < todayStamp.Unix() {
		var logFlag bool
		if hotout.Hotword, logFlag, err = s.GetHotPubLog(t); err != nil {
			log.Error("searchSrv.HotList GetHotPubLog error(%v)", err)
			return
		}
		if logFlag {
			hotout.State = _HotShowPub
		} else {
			hotout.State = _HotShowUnpub
		}
		return
	}
	// 取出干预中的热词
	if hot, _, err = s.HotwordFromDB(t, HotwordFromDBAll); err != nil {
		log.Error("searchSrv.HotList HotwordFromDB error(%v)", err)
		return
	}
	hotout.Hotword = hot
	// 今天的数据
	// todo: 没必要判断了，肯定是今天的数据，而且没有上下线逻辑了
	if dateStamp.Unix() == todayStamp.Unix() {
		// 已发布 且是今天发布的数据
		if flag && date == time.Now().Format("2006-01-02") {
			// 2.判断发布的时候 是否有搜索数据过来
			var pubStatus bool
			if pubStatus, _, err = s.dao.GetSearchAuditStat(c, _HotPubSearchState); err != nil {
				log.Error("searchSrv.SetHotPub SetSearchPubStat error(%v)", err)
				return
			}
			if pubStatus {
				// 发布的时候 有搜索的数据 提示上线
				hotout.State = _HotShowPub
			} else {
				// 发布的是 没有搜索的数据 提示 未更新
				hotout.State = _HotShowUnUp
			}
			return
		}
		// 未发布
		hotout.State = _HotShowUnpub
		return
	}
	// 未来的数据 都是未发布的
	hotout.State = _HotShowUnpub
	return
}

// DarkList darkword list
func (s *Service) DarkList(c *bm.Context, t string) (darkout searchModel.DarkwordOut, err error) {
	var (
		dateStamp  time.Time
		todayStamp time.Time
		flag       bool
		// flagAuto   bool
		dark []searchModel.Dark
		date string
	)
	if dateStamp, err = s.parseTime(t, "2006-01-02"); err != nil {
		return
	}
	today := time.Unix(time.Now().Unix(), 0).Format("2006-01-02")
	if todayStamp, err = s.parseTime(today, "2006-01-02"); err != nil {
		return
	}
	if flag, date, err = s.dao.GetSearchAuditStat(c, _DarkPubState); err != nil {
		log.Error("searchSrv.DarkList GetPublishCache error(%v)", err)
		return
	}
	// 过去的时间 则直接从日志中取数据
	if dateStamp.Unix() < todayStamp.Unix() {
		var logFlag bool
		if darkout.Darkword, logFlag, err = s.GetDarkPubLog(t); err != nil {
			log.Error("searchSrv.HotList GetHotPubLog error(%v)", err)
			return
		}
		if logFlag {
			darkout.State = _DarkShowPub
		} else {
			darkout.State = _DarkShowUnpub
		}
		return
	}
	// 公共的逻辑
	if dark, _, err = s.DarkwordFromDB(t); err != nil {
		log.Error("searchSrv.DarkList HotwordFromDB error(%v)", err)
		return
	}
	darkout.Darkword = dark
	// 今天的数据
	if dateStamp.Unix() == todayStamp.Unix() {
		// 已发布 且是今天发布的数据 则直接取缓存数据
		if flag && date == time.Now().Format("2006-01-02") {
			// 判断发布的时候 是否有搜索数据过来
			var pubStatus bool
			if pubStatus, _, err = s.dao.GetSearchAuditStat(c, _DarkPubSearchState); err != nil {
				log.Error("searchSrv.DarkList GetPubState error(%v)", err)
				return
			}
			if pubStatus {
				// 发布的时候 有搜索的数据 提示上线
				darkout.State = _DarkShowPub
			} else {
				// 发布的是 没有搜索的数据 提示 未更新
				darkout.State = _DarkShowUnUp
			}
			return
		}
		// 未更新
		darkout.State = _DarkShowUnpub
		return
	}
	// 未来的数据 都是未发布的
	darkout.State = _HotShowUnpub
	return
}

// BlackAll all black list
func (s *Service) BlackAll() (black []searchModel.Black, err error) {
	if err = s.dao.DB.Model(&searchModel.Black{}).
		Where("deleted = ?", searchModel.NotDelete).Find(&black).Error; err != nil {
		log.Error("BlackAll.History Index error(%v)", err)
		return
	}
	return
}

// DelBlack add black
func (s *Service) DelBlack(c *bm.Context, id int, person string, uid int64) (err error) {
	var (
		black searchModel.Black
	)
	// 根据id查找热词
	if err = s.dao.DB.Model(&searchModel.Black{}).
		Where("id = ?", id).First(&black).Error; err != nil {
		log.Error("searchSrv.DelBlack Black First error(%v)", err)
		return
	}
	if err = s.dao.DB.Model(&searchModel.Black{}).
		Where("id = ?", id).Update("deleted", searchModel.Delete).Error; err != nil {
		log.Error("searchSrv.DelBlack Update error(%v)", err)
		return
	}
	// 更新AI热词为删除状态
	if err = s.dao.DB.Model(&searchModel.History{}).
		Where("searchword = ?", black.Searchword).Update("deleted", searchModel.Delete).Error; err != nil {
		log.Error("searchSrv.DelBlack Update History error(%v)", err)
		return
	}
	// 更新运营热词为删除状态
	if err = s.dao.DB.Model(&searchModel.Intervene{}).
		Where("searchword = ?", black.Searchword).Update("deleted", searchModel.Delete).Error; err != nil {
		log.Error("searchSrv.DelBlack Update error(%v)", err)
		return
	}
	// 更新黑马词为删除状态
	if err = s.dao.DB.Model(&searchModel.Dark{}).
		Where("searchword = ?", black.Searchword).Update("deleted", searchModel.Delete).Error; err != nil {
		log.Error("searchSrv.DelBlack Update error(%v)", err)
		return
	}
	// 设置黑名单之后 立即发布新数据
	if err = s.SetHotPub(c, person, uid); err != nil {
		return
	}
	if err = s.SetDarkPub(c, person, uid); err != nil {
		return
	}
	obj := map[string]interface{}{
		"id": id,
	}
	if err = Log.AddLog(searchModel.Business, person, uid, int64(id), searchModel.ActionDelBlack, obj); err != nil {
		log.Error("searchSrv.DelBlack AddLog error(%v)", err)
		return
	}
	return
}

// AddBlack add black
func (s *Service) AddBlack(c *bm.Context, black string, person string, uid int64) (err error) {
	var (
		word searchModel.Black
	)
	if err = s.dao.DB.Model(&searchModel.Black{}).
		Where("deleted = ?", searchModel.NotDelete).Where("searchword = ?", black).
		First(&word).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.AddBlack get First error(%v)", err)
		return
	}
	if err != gorm.ErrRecordNotFound {
		err = fmt.Errorf("黑名单已存在")
		return
	}
	w := searchModel.AddBlack{
		Searchword: black,
	}
	if err = s.dao.DB.Model(&searchModel.Black{}).
		Create(w).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.AddBlack Create error(%v)", err)
		return
	}
	// 设置黑名单之后 立即发布新数据
	if err = s.SetHotPub(c, person, uid); err != nil {
		return
	}
	if err = s.SetDarkPub(c, person, uid); err != nil {
		return
	}
	obj := map[string]interface{}{
		"blackword": word,
	}
	if err = Log.AddLog(searchModel.Business, person, uid, 0, searchModel.ActionAddBlack, obj); err != nil {
		log.Error("searchSrv.AddBlack AddLog error(%v)", err)
		return
	}
	return
}

// checkBlack checkout blacklist
func (s *Service) checkBlack(word string) (state bool, err error) {
	var (
		black searchModel.Black
	)
	if err = s.dao.DB.Model(&searchModel.Black{}).
		Where("deleted = ?", searchModel.NotDelete).Where("searchword = ?", word).
		First(&black).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.checkBlack get First error(%v)", err)
		return
	}
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return true, nil
}

// checkInter checkout intervene
func (s *Service) checkInter(word string, id int) (state bool, err error) {
	var (
		intervene searchModel.Intervene
	)
	dataStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
	query := s.dao.DB.Model(&searchModel.Intervene{}).
		Where("searchword = ?", word)
	if id != 0 {
		query = query.Where("id != ?", id)
	}
	// 取未删除且结束时间大于当前时间的词
	query = query.Where("deleted = ?", searchModel.NotDelete).Where("etime > ?", dataStr)
	if err = query.First(&intervene).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.checkInter get First error(%v)", err)
		return
	}
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return true, nil
}

// checkTimeConflict checkout intervene time conflict
func (s *Service) checkTimeConflict(i searchModel.InterveneAdd, id int) (state bool, conflictList []searchModel.Intervene, err error) {
	var (
		c          int
		black      []searchModel.Black
		blackSlice []string
	)

	if i.Rank < 1 {
		// rank < 1 表示位置待定
		return
	}

	if black, err = s.BlackAll(); err != nil {
		log.Error("searchSrv.HotList Black error(%v)", err)
	}
	for _, v := range black {
		blackSlice = append(blackSlice, v.Searchword)
	}
	dateStr := time.Now().Format("2006-01-02 15:04:05")
	query := s.dao.DB.Model(&searchModel.Intervene{}).
		Where("rank = ?", i.Rank).
		Where("stime < ?", i.Etime).
		Where("etime > ?", i.Stime).
		Where("etime >= ?", dateStr).
		Where("searchword not in (?)", blackSlice).
		Where("deleted = ?", searchModel.NotDelete)
	if id != 0 {
		query = query.Where("id != ?", id)
	}
	if err = query.Find(&conflictList).Error; err != nil {
		log.Error("searchSrv.checkTimeConflict Find error(%v)", err)
		return
	}
	c = len(conflictList)
	if c > 0 {
		state = true
		return
	}
	return
}

// AddInter add intervene word
func (s *Service) AddInter(c *bm.Context, v searchModel.InterveneAdd, person string, uid int64) (err error) {
	var (
		state            bool
		conflictWordList []searchModel.Intervene
	)
	if state, err = s.checkBlack(v.Searchword); err != nil {
		log.Error("searchSrv.addInter checkBlack error(%v)", err)
		return
	}
	if state {
		err = fmt.Errorf("所添加的词在黑名单中已存在")
		return
	}
	if state, err = s.checkInter(v.Searchword, 0); err != nil {
		log.Error("searchSrv.addInter checkBlack error(%v)", err)
		return
	}
	if state {
		err = fmt.Errorf("当前搜索词已存在生效中或者待生效的干预")
		return
	}
	if state, conflictWordList, err = s.checkTimeConflict(v, 0); err != nil {
		log.Error("searchSrv.addInter checkTimeConflict error(%v)", err)
		return
	}
	if state {
		conflictWord := conflictWordList[0]
		err = fmt.Errorf("相同时间内，该位置已存在搜索词[%v]", conflictWord.Searchword)
		return
	}
	// 兼容老数据 小火苗类型数据 type为 hot
	if v.Type == SearchInterFire {
		v.Tag = "hot"
	}
	if err = s.dao.DB.Model(&searchModel.InterveneAdd{}).Create(&v).Error; err != nil {
		log.Error("searchSrv.AddIntervene Create error(%v)", err)
		return
	}
	if err = s.dao.SetSearchAuditStat(c, _HotPubState, false); err != nil {
		log.Error("searchSrv.DelBlack SetPubStat error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"value": v,
	}
	if err = Log.AddLog(searchModel.Business, person, uid, 0, searchModel.ActionAddInter, obj); err != nil {
		log.Error("searchSrv.addInter AddLog error(%v)", err)
		return
	}

	s.PubHotwordImmediately(ctx, "AddInter", 0)

	return
}

// UpdateInter update intervene word
//
//nolint:gocognit
func (s *Service) UpdateInter(c *bm.Context, v searchModel.InterveneAdd, id int, person string, uid int64) (err error) {
	tx := s.dao.DB.Begin()
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit().Error
		//nolint:gosimple
		return
	}()

	// 先查出我的位置
	currentConfig := &searchModel.Intervene{}
	if id != 0 {
		if err = tx.Model(&searchModel.Intervene{}).Where("id = ?", id).Find(currentConfig).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				err = ecode.Error(ecode.RequestErr, "配置不存在，请刷新页面，重新编辑")
			} else {
				log.Error("searchSrv.UpdateInter findCurrentPosition error(%v)", err)
			}
			return
		}
		// 防止编辑过期的配置
		if currentConfig.Deleted > 0 {
			err = ecode.Error(ecode.RequestErr, "配置不存在，请刷新页面，重新编辑")
			return
		}
	}

	// 是否是系统词创建
	isFromSys := false
	if id == 0 {
		isFromSys = true
		// id==0，说明是从系统词创建
		// 取出所有正在生效和未来生效的干预热词
		// 如果当前更改的词存在，那么就继续根据id更新
		// 如果不存在，就新创建一个词，然后继续
		var hotList []searchModel.Intervene
		if hotList, _, err = s.HotwordFromDB(time.Now().Format("2006-01-01"), HotwordFromDBAll); err != nil {
			log.Error("s.HotwordFromDB error(%v)", err)
			return
		}
		hotListWordMap := map[string]searchModel.Intervene{}
		for _, hotItem := range hotList {
			if hotItem.Intervene == 1 {
				hotListWordMap[hotItem.Searchword] = hotItem
			}
		}

		if _, ok := hotListWordMap[v.Searchword]; ok {
			err = fmt.Errorf("当前干预词已经在干预中或待生效")
			return
		}

		// 如果不存在，创建一条对应位置的干预
		if err = tx.Model(&searchModel.InterveneAdd{}).Create(&v).Error; err != nil {
			log.Error("searchSrv.UpdateInter Create error(%v)", err)
			return
		}
		id = v.ID
		if err = s.dao.SetSearchAuditStat(c, _HotPubState, false); err != nil {
			log.Error("searchSrv.UpdateInter SetPubStat error(%v)", err)
		}
	}

	var (
		state        bool
		conflictList []searchModel.Intervene
	)
	if state, err = s.checkBlack(v.Searchword); err != nil {
		log.Error("searchSrv.UpdateInter checkBlack error(%v)", err)
		return
	}
	if state {
		err = fmt.Errorf("所添加的词在黑名单中已存在")
		return
	}

	// 检查是否有和现在的词位置和时间冲突
	if state, conflictList, err = s.checkTimeConflict(v, id); err != nil {
		log.Error("searchSrv.UpdateInter checkTimeConflict error(%v)", err)
		return
	}

	if state {
		// 需要和冲突的位置交换

		myRank := currentConfig.Rank
		if isFromSys {
			if v.OldRank > 0 {
				myRank = v.OldRank
			} else {
				err = fmt.Errorf("缺少当前词的原始位置old_position")
				return
			}
		}

		var conflictIdList []int
		for _, conflictItem := range conflictList {
			// 检查当前配置是否生效中
			currentTime := time.Now().Unix()
			currentIsOnline := currentConfig.Stime.Time().Unix() <= currentTime && currentConfig.Etime.Time().Unix() >= currentTime
			conflictIsOnline := conflictItem.Stime.Time().Unix() <= currentTime && conflictItem.Etime.Time().Unix() >= currentTime
			// 如果生效中，那么就不能和待生效的有冲突
			// 如果待生效，那么就不能和生效中的有冲突
			if currentIsOnline != conflictIsOnline {
				if conflictIsOnline {
					err = fmt.Errorf("当前编辑的配置和生效中的配置[%v]有冲突，请修改时间或者位置", conflictItem.Searchword)
				} else {
					err = fmt.Errorf("当前编辑的配置和待生效的配置[%v]有冲突，请修改时间或者位置", conflictItem.Searchword)
				}
				return
			} else {
				conflictId := conflictItem.ID
				conflictIdList = append(conflictIdList, conflictId)
			}

			// 可达，数据报表需要区分同一天的不同位次，所以每次更改都要先删除，再新增
			// 这里我先新增，后删除，新增同样的配置，更改为我的位置
			if err = tx.Model(&searchModel.InterveneAdd{}).Create(&searchModel.InterveneAdd{
				Searchword: conflictItem.Searchword,
				Rank:       myRank,
				Tag:        conflictItem.Tag,
				Stime:      conflictItem.Stime,
				Etime:      conflictItem.Etime,
				Type:       conflictItem.Type,
				Image:      conflictItem.Image,
				GotoType:   conflictItem.GotoType,
				GotoValue:  conflictItem.GotoValue,
				ShowWord:   conflictItem.ShowWord,
				ResourceId: conflictItem.ResourceId,
			}).Error; err != nil {
				log.Error("searchSrv.UpdateInter createNewConflictItem error(%v)", err)
				return
			}
		}
		// 把原本冲突的位置配置全部删除
		if err = tx.Model(&searchModel.Intervene{}).
			Where("id in (?)", conflictIdList).
			Update("deleted", searchModel.Delete).Error; err != nil {
			log.Error("searchSrv.UpdateInter deleteConflictPosition params(%v) error(%v)", conflictIdList, err)
			return
		}
	}
	// 兼容老数据 小火苗类型数据 type为 hot
	if v.Type == SearchInterFire {
		v.Tag = "hot"
	}
	if !isFromSys {
		// 如果是已有干预词，先新建，再把老的删除
		if err = tx.Model(&searchModel.InterveneAdd{}).
			Create(&searchModel.InterveneAdd{
				Searchword: v.Searchword,
				Rank:       v.Rank,
				Tag:        v.Tag,
				Stime:      v.Stime,
				Etime:      v.Etime,
				Type:       v.Type,
				Image:      v.Image,
				GotoType:   v.GotoType,
				GotoValue:  v.GotoValue,
				ShowWord:   v.ShowWord,
				ResourceId: v.ResourceId,
			}).Error; err != nil {
			log.Error("searchSrv.UpdateInter Create error(%v)", err)
			return
		}
		if err = tx.Model(&searchModel.Intervene{}).
			Where("id = ?", id).Update("deleted", searchModel.Delete).Error; err != nil {
			log.Error("searchSrv.UpdateInter Delete error(%v)", err)
			return
		}
	}
	// 如果是系统词，本次新建就是最新数据，不需要更新

	if err = s.dao.SetSearchAuditStat(c, _HotPubState, false); err != nil {
		log.Error("searchSrv.DelBlack SetPubStat error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"value": v,
		"id":    id,
	}
	if err = Log.AddLog(searchModel.Business, person, uid, int64(id), searchModel.ActionUpdateInter, obj); err != nil {
		log.Error("searchSrv.UpdateInter AddLog error(%v)", err)
		return
	}

	s.PubHotwordImmediately(ctx, "UpdateInter", 0)

	return
}

// UpdateSearch update search hot tag
func (s *Service) UpdateSearch(c *bm.Context, tag string, id int, person string, uid int64) (err error) {
	if err = s.dao.DB.Model(&searchModel.History{}).
		Where("id = ?", id).Update("tag", tag).Error; err != nil {
		log.Error("searchSrv.UpdateSearch Update error(%v)", err)
		return
	}
	if err = s.dao.SetSearchAuditStat(c, _HotPubState, false); err != nil {
		log.Error("searchSrv.DelBlack SetPubStat error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"value": tag,
		"id":    id,
	}
	if err = Log.AddLog(searchModel.Business, person, uid, int64(id), searchModel.ActionUpdateSearch, obj); err != nil {
		log.Error("searchSrv.UpdateSearch AddLog error(%v)", err)
		return
	}

	s.PubHotwordImmediately(ctx, "UpdateSearch", 0)

	return
}

// DeleteHot delete hot word
func (s *Service) DeleteHot(c context.Context, id int, t uint8, person string, uid int64) (err error) {
	if t == searchModel.HotAI {
		// 删除AI热词
		// todo: 界面上已经没有这个功能了
		if err = s.dao.DB.Model(&searchModel.History{}).
			Where("id = ?", id).Update("deleted", searchModel.Delete).Error; err != nil {
			log.Error("searchSrv.DeleteHot Update AI error(%v)", err)
			return
		}
	} else if t == searchModel.HotOpe {
		// 删除运营热词
		if err = s.dao.DB.Model(&searchModel.Intervene{}).
			Where("id = ?", id).Update("deleted", searchModel.Delete).Error; err != nil {
			log.Error("searchSrv.DeleteHot Update Operate error(%v)", err)
			return
		}
	}
	obj := map[string]interface{}{
		"type": t,
		"id":   id,
	}
	if err = Log.AddLog(searchModel.Business, person, uid, int64(id), searchModel.ActionDeleteHot, obj); err != nil {
		log.Error("searchSrv.DeleteHot AddLog error(%v)", err)
		return
	}

	s.PubHotwordImmediately(ctx, "DeleteHot", 0)

	return
}

// DeleteDark delete dark word
func (s *Service) DeleteDark(c context.Context, id int, person string, uid int64) (err error) {
	if err = s.dao.DB.Model(&searchModel.Dark{}).
		Where("id = ?", id).Update("deleted", searchModel.Delete).Error; err != nil {
		log.Error("searchSrv.DeleteDark Update error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"id": id,
	}
	if err = Log.AddLog(searchModel.Business, person, uid, int64(id), searchModel.ActionDeleteDark, obj); err != nil {
		log.Error("searchSrv.DeleteDark AddLog error(%v)", err)
		return
	}

	s.PubHotwordImmediately(ctx, "DeleteDark", 0)

	return
}

// OpenAddDarkword open api for search add dark word
func (s *Service) OpenAddDarkword(c context.Context, values searchModel.OpenDark) (err error) {
	if err = s.dao.DB.Model(&searchModel.Dark{}).
		Where("atime = ?", values.Date).Update("deleted", searchModel.Delete).Error; err != nil {
		log.Error("searchSrv.OpenAddDarkword Update error(%v)", err)
		return
	}
	for _, v := range values.Values {
		dark := searchModel.Dark{
			Searchword: v.Searchword,
			PV:         v.PV,
			UV:         v.UV,
			Click:      v.Click,
			Atime:      values.Date,
		}
		if err = s.dao.DB.Model(&searchModel.Dark{}).Create(&dark).Error; err != nil {
			log.Error("searchSrv.OpenAddDarkword Create error(%v)", err)
			return
		}
	}
	// 如果有黑马词同步过来 则更新发布状态为false
	if err = s.dao.SetSearchAuditStat(c, _DarkAutoPubState, false); err != nil {
		log.Error("searchSrv.DelBlack SetPubStat error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"value": values,
	}
	if err = Log.AddLog(searchModel.Business, "SEARCH", 0, 0, searchModel.ActionOpenAddDark, obj); err != nil {
		log.Error("searchSrv.OpenAddDarkword AddLog error(%v)", err)
		return
	}

	s.PubHotwordImmediately(ctx, "OpenAddDarkword", 0)

	return
}

// OpenAddHotword open api for search add hotword
func (s *Service) OpenAddHotword(c context.Context, values searchModel.OpenHot) (err error) {
	if err = s.dao.DB.Model(&searchModel.Hot{}).Where("atime = ?", values.Date).Delete(&searchModel.Hot{}).Error; err != nil {
		log.Error("searchSrv.OpenAddHotword Delete error(%v)", err)
		return
	}
	for _, v := range values.Values {
		hot := searchModel.Hot{
			Searchword: v.Searchword,
			PV:         v.PV,
			UV:         v.UV,
			Click:      v.Click,
			Atime:      values.Date,
		}
		if err = s.dao.DB.Model(&searchModel.Hot{}).Create(&hot).Error; err != nil {
			log.Error("searchSrv.OpenAddHotword Create error(%v)", err)
			return
		}
	}
	// 如果有搜索热词同步过来 则更新自动发布状态为false
	if err = s.dao.SetSearchAuditStat(c, _HotAutoPubState, false); err != nil {
		log.Error("searchSrv.DelBlack SetPubStat error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"value": values,
	}
	if err = Log.AddLog(searchModel.Business, "SEARCH", 0, 0, searchModel.ActionOpenAddHot, obj); err != nil {
		log.Error("searchSrv.OpenAddHotword AddLog error(%v)", err)
		return
	}

	s.PubHotwordImmediately(ctx, "OpenAddHotword", 0)

	return
}

// 获取上次ai同步的时间
func (s *Service) GetLastSyncTime(c context.Context) (lastTime searchModel.LastTimeSyncItem, err error) {
	if err = s.dao.MC.Get(c, _LastSearchSyncValue).Scan(&lastTime); err != nil {
		if err == memcache.ErrNotFound {
			return searchModel.LastTimeSyncItem{
				LastTime: 0,
			}, nil
		}
		return
	}
	return
}

// 获取最近一次上线的时间
func (s *Service) GetLastOnlineTime(c context.Context) (lastTime searchModel.LastTimeSyncItem, err error) {
	if err = s.dao.MC.Get(c, _LastSearchOnlineValue).Scan(&lastTime); err != nil {
		if err == memcache.ErrNotFound {
			return searchModel.LastTimeSyncItem{
				LastTime: 0,
			}, nil
		}
		return
	}
	return
}

// GetHotPub get hotword publish from mc
func (s *Service) GetHotPub(c context.Context) (hot []searchModel.Intervene, err error) {
	if err = s.dao.MC.Get(c, _HotPubValue).Scan(&hot); err != nil {
		if err == memcache.ErrNotFound {
			return nil, nil
		}
		return
	}
	return
}

// GetDarkPub get darkword publish from mc
func (s *Service) GetDarkPub(c context.Context) (dark []searchModel.Dark, err error) {
	if err = s.dao.MC.Get(c, _DarkPubValue).Scan(&dark); err != nil {
		if err == memcache.ErrNotFound {
			return nil, nil
		}
		return
	}
	return
}

// 编辑过后，立即发布配置
func (s *Service) PubHotwordImmediately(c context.Context, person string, uid int64) {
	// 记录上次AI同步的时间
	lastSync := searchModel.LastTimeSyncItem{
		LastTime: time.Now().Unix(),
	}
	lastOnlineJSON := &memcache.Item{
		Key:        _LastSearchOnlineValue,
		Flags:      memcache.FlagJSON,
		Object:     lastSync,
		Expiration: 0,
	}
	_ = s.dao.MC.Set(c, lastOnlineJSON)

	//nolint:biligowordcheck
	go func() {
		if err := s.SetHotPub(c, person, uid); err != nil {
			log.Error("PubHotwordImmediately error in go routine error(%v) person(%v)", err, person)
			return
		}
		log.Info("PubHotwordImmediately Success!")
	}()
}

// SetHotPub set hotword publish to mc
func (s *Service) SetHotPub(c context.Context, person string, uid int64) (err error) {
	var (
		hot         []searchModel.Intervene
		searchCount int
	)
	// 只能发布当天的数据
	// 从DB中取今天的数据
	if hot, searchCount, err = s.HotwordFromDB(time.Now().Format("2006-01-02"), HotwordFromDBAll); err != nil {
		log.Error("searchSrv.SetHoGetSearHisValuetPub HotwordFromDB error(%v)", err)
		return
	}
	itemJSON := &memcache.Item{
		Key:        _HotPubValue,
		Flags:      memcache.FlagJSON,
		Object:     hot,
		Expiration: 0,
	}
	if err = s.dao.MC.Set(c, itemJSON); err != nil {
		log.Error("searchSrv.SetHotPub conn.Set error(%v)", err)
		return
	}
	if searchCount == 0 {
		// 证明搜索没有推数据过来 设置搜索的数据为假
		if err = s.dao.SetSearchAuditStat(c, _HotPubSearchState, false); err != nil {
			log.Error("searchSrv.SetHotPub SetSearchPubStat error(%v)", err)
			return
		}
	} else {
		// 证明搜索有推数据过来 设置搜索的数据为真
		if err = s.dao.SetSearchAuditStat(c, _HotPubSearchState, true); err != nil {
			log.Error("searchSrv.SetHotPub SetSearchPubStat error(%v)", err)
			return
		}
	}
	// 设置自动发布状态为true
	if err = s.dao.SetSearchAuditStat(c, _HotAutoPubState, true); err != nil {
		log.Error("searchSrv.SetHotPub SetPubStat error(%v)", err)
		return
	}
	// 设置运营发布状态为true
	if err = s.dao.SetSearchAuditStat(c, _HotPubState, true); err != nil {
		log.Error("searchSrv.SetHotPub SetPubStat error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"value": hot,
	}
	if err = Log.AddLog(searchModel.Business, person, uid, 0, searchModel.ActionPublishHot, obj); err != nil {
		log.Error("searchSrv.SetHotPub AddLog error(%v)", err)
		return
	}
	if err = s.HotPubLog(hot); err != nil {
		log.Error("searchSrv.SetHotPub HotPubLog error(%v)", err)
		return
	}

	// 记录上次AI同步的时间
	lastSync := searchModel.LastTimeSyncItem{
		LastTime: time.Now().Unix(),
	}
	lastOnlineJSON := &memcache.Item{
		Key:        _LastSearchOnlineValue,
		Flags:      memcache.FlagJSON,
		Object:     lastSync,
		Expiration: 0,
	}
	if err = s.dao.MC.Set(c, lastOnlineJSON); err != nil {
		log.Error("searchSrv.SetHotPub s.dao.MC.Set error(%v)", err)
		err = nil
	}

	return
}

// HotPubLog hotword publish log
func (s *Service) HotPubLog(hot []searchModel.Intervene) (err error) {
	t := time.Now().Unix()
	for _, v := range hot {
		w := searchModel.HotPubLog{
			Searchword: v.Searchword,
			Position:   v.Rank,
			Pv:         v.Pv,
			Tag:        v.Tag,
			Stime:      v.Stime,
			Etime:      v.Etime,
			Atime:      time.Now().Format("2006-01-02"),
			Groupid:    t,
		}
		if err = s.dao.DB.Model(&searchModel.HotPubLog{}).Create(&w).Error; err != nil {
			log.Error("searchSrv.DarkPubLog Create error(%v)", err)
			return
		}
	}
	return
}

// GetHotPubLog get hotword publish log
func (s *Service) GetHotPubLog(date string) (hotout []searchModel.Intervene, pub bool, err error) {
	var (
		// hotout searchModel.HotwordOut
		logs []searchModel.HotPubLog
	)
	l := searchModel.HotPubLog{}
	if err = s.dao.DB.Model(&searchModel.HotPubLog{}).Where("atime = ?", date).Order("groupid desc").
		First(&l).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.GetHotPubLog First error(%v)", err)
		return
	}
	// 证明没有发布过
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	// 取最大的groupid的值
	if err = s.dao.DB.Model(&searchModel.HotPubLog{}).Where("groupid = ?", l.Groupid).
		Find(&logs).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.GetHotPubLog Find error(%v)", err)
		return
	}
	for _, v := range logs {
		a := searchModel.Intervene{
			Searchword: v.Searchword,
			Rank:       v.Position,
			Pv:         v.Pv,
			Tag:        v.Tag,
			Stime:      v.Stime,
			Etime:      v.Etime,
		}
		hotout = append(hotout, a)
	}
	return hotout, true, nil
}

// GetDarkPubLog get darkword publish log
func (s *Service) GetDarkPubLog(date string) (darkout []searchModel.Dark, pub bool, err error) {
	var (
		// hotout searchModel.HotwordOut
		logs []searchModel.DarkPubLog
	)
	l := searchModel.DarkPubLog{}
	if err = s.dao.DB.Model(&searchModel.DarkPubLog{}).Where("atime = ?", date).Order("groupid desc").
		First(&l).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.GetDarkPubLog First error(%v)", err)
		return
	}
	// 证明没有发布过
	if err == gorm.ErrRecordNotFound {
		return nil, false, nil
	}
	// 取最大的groupid的值
	if err = s.dao.DB.Model(&searchModel.DarkPubLog{}).Where("groupid = ?", l.Groupid).
		Find(&logs).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("searchSrv.GetDarkPubLog Find error(%v)", err)
		return
	}
	for _, v := range logs {
		a := searchModel.Dark{
			Searchword: v.Searchword,
			PV:         v.Pv,
		}
		darkout = append(darkout, a)
	}
	return darkout, true, nil
}

// SetDarkPub set darkword to mc
func (s *Service) SetDarkPub(c context.Context, person string, uid int64) (err error) {
	var (
		dark        []searchModel.Dark
		searchCount int
	)
	// 只能发布当天的数据
	// 从DB中取今天的数据
	if dark, searchCount, err = s.DarkwordFromDB(time.Now().Format("2006-01-02")); err != nil {
		log.Error("searchSrv.SetHotPub HotwordFromDB error(%v)", err)
		return
	}
	itemJSON := &memcache.Item{
		Key:        _DarkPubValue,
		Flags:      memcache.FlagJSON,
		Object:     dark,
		Expiration: 0,
	}
	if err = s.dao.MC.Set(c, itemJSON); err != nil {
		log.Error("searchSrv.SetHotPub conn.Set error(%v)", err)
		return
	}
	if searchCount == 0 {
		// 证明搜索没有推数据过来 设置搜索的数据为假
		if err = s.dao.SetSearchAuditStat(c, _DarkPubSearchState, false); err != nil {
			log.Error("searchSrv.SetDarkPub SetSearchPubStat error(%v)", err)
			return
		}
	} else {
		// 证明搜索有推数据过来 设置搜索的数据为真
		if err = s.dao.SetSearchAuditStat(c, _DarkPubSearchState, true); err != nil {
			log.Error("searchSrv.SetDarkPub SetSearchPubStat error(%v)", err)
			return
		}
	}
	if err = s.dao.SetSearchAuditStat(c, _DarkPubState, true); err != nil {
		log.Error("searchSrv.SetHotPub SetPubStat error(%v)", err)
		return
	}
	if err = s.dao.SetSearchAuditStat(c, _DarkAutoPubState, true); err != nil {
		log.Error("searchSrv.SetHotPub SetPubStat error(%v)", err)
		return
	}
	obj := map[string]interface{}{
		"value": dark,
	}
	if err = Log.AddLog(searchModel.Business, person, uid, 0, searchModel.ActionPublishDark, obj); err != nil {
		log.Error("searchSrv.SetDarkPub AddLog error(%v)", err)
		return
	}
	if err = s.DarkPubLog(dark); err != nil {
		log.Error("searchSrv.SetDarkPub DarkPubLog error(%v)", err)
		return
	}
	return
}

// DarkPubLog get darkword publish log
func (s *Service) DarkPubLog(dark []searchModel.Dark) (err error) {
	t := time.Now().Unix()
	for _, v := range dark {
		w := searchModel.DarkPubLog{
			Searchword: v.Searchword,
			Pv:         v.PV,
			Atime:      time.Now().Format("2006-01-02"),
			Groupid:    t,
		}
		if err = s.dao.DB.Model(&searchModel.DarkPubLog{}).Create(&w).Error; err != nil {
			log.Error("searchSrv.DarkPubLog Create error(%v)", err)
			return
		}
	}
	return
}

// SearchInterHistory 搜索干预的历史数据
func (s *Service) SearchInterHistory(param *searchModel.InterHisParam) (pager *searchModel.InterHisListPager, err error) {
	var (
		blackSlice []string
		black      []searchModel.Black
	)
	pager = &searchModel.InterHisListPager{
		Page: common.Page{
			Num:  param.Pn,
			Size: param.Ps,
		},
	}
	if black, err = s.BlackAll(); err != nil {
		log.Error("SearchInterHistory BlackAll error(%v)", err)
		return
	}
	for _, v := range black {
		blackSlice = append(blackSlice, v.Searchword)
	}
	query := s.dao.DB.Model(&searchModel.Intervene{})
	if param.Date != "" {
		sTime := param.Date + " 00:00:00"
		eTime := param.Date + " 23:59:59"
		query = query.Where("stime >= ?", sTime).Where("stime <= ?", eTime)
	}
	if param.Searchword != "" {
		query = query.Where("searchword like ?", "%"+param.Searchword+"%")
	}
	query = query.Where("searchword not in (?)", blackSlice)
	if err = query.Count(&pager.Page.Total).Error; err != nil {
		log.Error("SearchInterHistory count error(%v)", err)
		return
	}
	// 未结束的运营干预词
	if err = query.Order("`id` DESC").Offset((param.Pn - 1) * param.Ps).Limit(param.Ps).
		Find(&pager.Item).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Error("SearchInterHistory error(%v)", err)
		return
	}
	for _, v := range pager.Item {
		if v.ShowWord == "" {
			v.ShowWord = v.Searchword
		}
	}
	return
}

// 运营词完全重新排序
//
//nolint:gocognit
func (s *Service) HotSort(c *bm.Context, configList []*searchModel.SortConfigItem, person string, uid int64) (err error) {
	tx := s.dao.DB.Begin()
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit().Error
		//nolint:gosimple
		return
	}()

	// 先取出来当前生效的全部干预
	var hotList []searchModel.Intervene
	if hotList, _, err = s.HotwordFromDB(time.Now().Format("2006-01-01"), HotwordFromDBOnline); err != nil {
		log.Error("s.HotwordFromDB error(%v)", err)
		return
	}
	hotListWordMap := map[string]searchModel.Intervene{}
	hotListIndexMap := map[int]searchModel.Intervene{}
	for _, hotItem := range hotList {
		if hotItem.Intervene == 1 {
			hotListWordMap[hotItem.Searchword] = hotItem
			hotListIndexMap[hotItem.Rank] = hotItem
		}
	}

	for _, configItem := range configList {
		// 检查是否是想要干预，还是自然
		if configItem.Intervene == 1 {
			// 干预
			// 检查是否是已经存在干预
			if v, ok := hotListWordMap[configItem.Searchword]; ok {
				// 如果存在，对已有干预进行位置变更
				var oldConfigItem = &searchModel.InterveneAdd{}
				if err = tx.Model(&searchModel.InterveneAdd{}).
					Where("id = ?", v.ID).
					Find(&oldConfigItem).Error; err != nil {
					log.Error("searchSrv.HotSort FindOldConfigItem error(%v)", err)
					return
				}

				// 如果新老位置一样，那就跳过
				if configItem.Position == oldConfigItem.Rank {
					continue
				}

				// 删除之前的配置
				if err = tx.Model(&searchModel.Intervene{}).
					Where("searchword = ? OR rank = ? ", configItem.Searchword, configItem.Position).
					Update("deleted", searchModel.Delete).Error; err != nil {
					log.Error("searchSrv.HotSort Delete error(%v)", err)
					return
				}

				// 不一样，需要对已有干预进行变更
				// 先新增一个一样的配置，更新这个新建的位置，然后删除老的
				if err = tx.Model(&searchModel.InterveneAdd{}).Create(&searchModel.InterveneAdd{
					Searchword: oldConfigItem.Searchword,
					Rank:       configItem.Position,
					Tag:        oldConfigItem.Tag,
					Stime:      oldConfigItem.Stime,
					Etime:      oldConfigItem.Etime,
					Type:       oldConfigItem.Type,
					Image:      oldConfigItem.Image,
					GotoType:   oldConfigItem.GotoType,
					GotoValue:  oldConfigItem.GotoValue,
					ShowWord:   oldConfigItem.ShowWord,
					ResourceId: oldConfigItem.ResourceId,
				}).Error; err != nil {
					log.Error("searchSrv.HotSort Create error(%v)", err)
					return
				}

				if err = s.dao.SetSearchAuditStat(c, _HotPubState, false); err != nil {
					log.Error("searchSrv.HotSort SetPubStat error(%v)", err)
					return
				}
			} else {
				// 检查目标位置，是不是已经有干预词，如果有，就把已有的删除
				if _, ok := hotListIndexMap[configItem.Position]; ok && configItem.Position > 0 {
					if err = tx.Model(&searchModel.Intervene{}).
						Where("rank = ?", configItem.Position).
						Update("deleted", searchModel.Delete).Error; err != nil {
						log.Error("searchSrv.HotSort Update Operate error(%v)", err)
						return
					}
				}

				// 如果不存在，创建一条对应位置的干预，24小时有效时间
				if err = tx.Model(&searchModel.InterveneAdd{}).Create(&searchModel.InterveneAdd{
					Rank:       configItem.Position,
					ShowWord:   configItem.Searchword,
					Searchword: configItem.Searchword,
					Type:       1,
					Stime:      xtime.Time(time.Now().Unix()),
					Etime:      xtime.Time(time.Now().Unix() + 86400),
					ResourceId: configItem.ResourceId,
				}).Error; err != nil {
					log.Error("searchSrv.HotSort Create error(%v)", err)
					return
				}

				if err = s.dao.SetSearchAuditStat(c, _HotPubState, false); err != nil {
					log.Error("searchSrv.HotSort SetPubStat error(%v)", err)
					return
				}
			}
		} else {
			// 不想要该位置进行人工干预
			// 检查是否是已经存在干预
			if v, ok := hotListWordMap[configItem.Searchword]; ok {
				// 如果存在，删除这条干预
				if err = tx.Model(&searchModel.Intervene{}).
					Where("id = ?", v.ID).Update("deleted", searchModel.Delete).Error; err != nil {
					log.Error("searchSrv.HotSort Update Operate error(%v)", err)
					return
				}
			}
		}
	}

	obj := map[string]interface{}{
		"config_list": configList,
	}
	// todo: oid=0 可以吗？
	if err = Log.AddLog(searchModel.Business, person, uid, 0, searchModel.HotSort, obj); err != nil {
		log.Error("searchSrv.HotSort AddLog error(%v)", err)
		return
	}

	s.PubHotwordImmediately(ctx, "HotSort", 0)

	return
}

// 获取在线的搜索热词的前20名
func (s *Service) HotTop(c *bm.Context) (hotout searchModel.HotwordOut, err error) {
	type onlineSearchItem struct {
		SearchWord string `json:"keyword"`
		Position   int    `json:"pos"`
	}
	type stat struct {
		Pv    int
		Click int
	}
	var (
		hot             []searchModel.Intervene
		searchOnlineRes struct {
			Code int                 `json:"code"`
			List []*onlineSearchItem `json:"list"`
		}
	)

	// 上次搜索同步的时间
	lastSyncTime, _ := s.GetLastSyncTime(c)
	hotout.LastSyncTime = lastSyncTime.LastTime

	// 上次上线时间
	lastOnlineTime, _ := s.GetLastOnlineTime(c)
	hotout.LastOnlineTime = lastOnlineTime.LastTime

	// 同步5分钟后，认为已经生效
	if lastOnlineTime.LastTime+s.c.TimeGap.Hotword < lastSyncTime.LastTime {
		hotout.State = 1
	}

	// 取出干预中的热词
	today := time.Unix(time.Now().Unix(), 0).Format("2006-01-02")
	if hot, _, err = s.HotwordFromDB(today, HotwordFromDBOnline); err != nil {
		log.Error("searchSrv.HotTop HotwordFromDB error(%v)", err)
		return
	}
	// 取前20
	hot = hot[0:int(math.Min(float64(len(hot)), 20))]

	// 同步搜索的线上数据
	if err = s.dao.Client.Get(ctx, s.dao.SearchOnlineURL, "", nil, &searchOnlineRes); err != nil {
		log.Error("[HotTop] d.client.Get() url(%s) error(%v)", s.dao.SearchOnlineURL, err)
		err = nil
	}

	// 将线上数据的位置和top20混合
	searchOnlineResMap := map[string]*onlineSearchItem{}
	for _, v := range searchOnlineRes.List {
		searchOnlineResMap[v.SearchWord] = v
	}

	// 实时数据接口有查询频率限制,必须要将20个词一起查,这里构造一个形为{'a','b','c'}的字串切片作为搜索词放进接口中
	worldlist := []string{}
	for _, v := range hot {
		worldlist = append(worldlist, v.Searchword)
	}
	statLiveItem, _ := s.GetStatisticsLive(c, worldlist)

	// 将对应的数据根据搜索词累加到statLive中.
	statLive := make(map[string]stat)
	for _, v := range statLiveItem {
		statLive[v.SearchWord] = stat{
			Pv:    v.Pv + statLive[v.SearchWord].Pv,
			Click: v.Click + statLive[v.SearchWord].Click,
		}
	}
	// 将top20的数据更新为实时数据
	for i, v := range hot {
		onlineSearchItem, ok := searchOnlineResMap[v.Searchword]
		if ok {
			hot[i].OnlinePosition = onlineSearchItem.Position
		}
		hot[i].Pv = statLive[hot[i].Searchword].Pv
		hot[i].Click = statLive[hot[i].Searchword].Click
	}

	hotout.Hotword = hot
	return
}

// 获取搜索热词的预定池
func (s *Service) HotPending(c *bm.Context) (hotout searchModel.HotwordOut, err error) {
	var (
		hot []searchModel.Intervene
	)

	// 上次搜索同步的时间
	lastSyncTime, _ := s.GetLastSyncTime(c)
	hotout.LastSyncTime = lastSyncTime.LastTime

	// 上次上线时间
	lastOnlineTime, _ := s.GetLastOnlineTime(c)
	hotout.LastOnlineTime = lastOnlineTime.LastTime

	if lastOnlineTime.LastTime < lastSyncTime.LastTime {
		hotout.State = 1
	}

	// 取出干预中的热词
	today := time.Unix(time.Now().Unix(), 0).Format("2006-01-02")
	if hot, _, err = s.HotwordFromDB(today, HotwordFromDBFuture); err != nil {
		log.Error("searchSrv.HotPending HotwordFromDB error(%v)", err)
		return
	}
	hotout.Hotword = hot
	return
}

// 查询曝光数据(历史数据)
func (s *Service) GetStatistics(c *bm.Context, searchWord, startTime, endTime string) (statisticsRes []*searchModel.StaticticsListItem, err error) {
	type platDataItem struct {
		LogDate    string `json:"log_date"`
		Rank       int    `json:"rank"`
		Pv         int    `json:"show_pv"`
		Uv         int    `json:"show_uv"`
		ClickRate  string `json:"discover_pv_clickrate"`
		SearchWord string `json:"keywords"`
		STime      string `json:"up_time"`
		ETime      string `json:"down_time"`
	}
	var (
		platDataList []*platDataItem
	)

	query := &dataplat.Query{}
	query.Select("log_date,rank,show_pv,show_uv,discover_pv_clickrate,keywords,up_time,down_time").
		Where(
			dataplat.ConditionMapType{"log_date": dataplat.ConditionLte(endTime)},
			dataplat.ConditionMapType{"log_date": dataplat.ConditionGte(startTime)},
			dataplat.ConditionMapType{"keywords": dataplat.ConditionIn(searchWord)},
		).Limit(30, 0).Order("log_date desc")
	if err = s.dao.CallDataAPI(c, s.dao.SearchStatisticsURL, query, &platDataList); err != nil {
		log.Error("[HotTop] s.dao.CallDataAPI() url(%s) error(%v)", s.dao.SearchStatisticsURL, err)
		return
	}

	for _, item := range platDataList {
		statisticsRes = append(statisticsRes, &searchModel.StaticticsListItem{
			SearchWord: item.SearchWord,
			Rank:       item.Rank,
			LogDate:    item.LogDate,
			Pv:         item.Pv,
			Uv:         item.Uv,
			ClickRate:  item.ClickRate,
			STime:      item.STime,
			ETime:      item.ETime,
		})
	}

	return
}

// 查询实时曝光数据(实时数据)
func (s *Service) GetStatisticsLive(c *bm.Context, searchWord []string) (res []searchModel.StaticticsLiveListItem, err error) {
	// ods_s_hot_search_stat_rt
	// 为每个搜索词加上单引号
	for i, v := range searchWord {
		searchWord[i] = "'" + v + "'"
	}
	// 由于接口要求,这里需要以原始的方式手动拼接字符串
	now := time.Now()
	_time := fmt.Sprintf("'%d-%d-%d'", now.Year(), now.Month(), now.Day())
	fmt.Println(_time)
	query := "select keywords,maxMerge(click_cnt)as cnt ,maxMerge(show_pv)as pv, pos from bili_main.ods_s_hot_search_stat_rt_view_dist "
	query += "WHERE keywords in (" + strings.Join(searchWord, ",") + ") " + " AND ctime >= " + _time
	query += "GROUP BY keywords,pos "
	query += "ORDER BY pos asc "
	// 由于接口要求,接受参数也只能使用原始的方法
	platDataList := [][]string{}
	if err = s.dao.CallDataAPI_normal(c, s.dao.SearchStatisticsURLLive, query, &platDataList); err != nil {
		log.Error("[HotTop] s.dao.CallDataAPI_normal() url(%s) error(%v)", s.dao.SearchStatisticsURLLive, err)
		return
	}
	// 将结果添加到返回结果中
	for i := 0; i < len(platDataList); i++ {
		searchWord := platDataList[i][0]
		click, _ := strconv.Atoi(platDataList[i][1])
		pv, _ := strconv.Atoi(platDataList[i][2])
		pos, _ := strconv.Atoi(platDataList[i][3])

		// pv有可能为0,此时我们将click作为点击率返回.
		clickrate := float64(click)
		if pv != 0 {
			clickrate = float64(click) / float64(pv)
		}
		// 构造百分比形式的点击率
		clickratestring := fmt.Sprintf("%.2f%%", clickrate*100)

		StaticticsLiveItem := searchModel.StaticticsLiveListItem{
			SearchWord: searchWord,
			Click:      click,
			Pv:         pv,
			Clickrate:  clickratestring,
			Pos:        pos,
		}
		res = append(res, StaticticsLiveItem)
	}
	return
}

// BatchOptResultSpread 批量审核搜索结果干预数据
func (s *Service) BatchOptResultSpread(c *bm.Context, req *pb.BatchOptResultSpreadReq) (resp *pb.BatchOptResultSpreadResp, err error) {
	resp = &pb.BatchOptResultSpreadResp{}
	if resp.InvalidIds, err = s.checkBatchOptIds(c, req); err != nil {
		log.Errorc(c, "s.BatchOptResultSpread checkBatchOptIds req (%+v) err(%v)", req, err)
		return
	}
	if len(resp.InvalidIds) > 0 {
		log.Errorc(c, "s.BatchOptResultSpread found invalid ids: %+v", resp.InvalidIds)
		return
	}
	if err = s.dao.SearchOptSpreadConfig(c, req.SpreadIds, req.Option); err != nil {
		log.Errorc(c, "s.BatchOptResultSpread call dao.SearchOptSpreadConfig ids(%+v) err(%v)", req.SpreadIds, err)
		return
	}

	//TODO: add action logs
	return
}

func (s *Service) checkBatchOptIds(c *bm.Context, req *pb.BatchOptResultSpreadReq) (invalidIds []*pb.BatchInvalidItem, err error) {
	var (
		configs map[int64]*searchModel.SpreadConfig
	)
	if configs, err = s.dao.SearchSpreadConfigQueryById(c, req.SpreadIds); err != nil {
		return
	}

	invalidIds = make([]*pb.BatchInvalidItem, 0, len(req.SpreadIds))
	for _, id := range req.SpreadIds {
		var ok bool
		var conf *searchModel.SpreadConfig
		var invalidItem = &pb.BatchInvalidItem{Id: id}

		if conf, ok = configs[id]; !ok {
			invalidItem.Msg = "搜索干预ID不存在"
			invalidIds = append(invalidIds, invalidItem)
		} else {
			switch req.Option {
			case common.OptionBatchPass, common.OptionBatchReject:
				if conf.Check != common.Verify {
					invalidItem.Msg = "搜索干预不是待审核状态"
					invalidIds = append(invalidIds, invalidItem)
				}
			case common.OptionBatchHidden:
				if conf.Check != common.Pass && conf.Check != common.Valid {
					invalidItem.Msg = "搜索干预不是已通过/已生效状态"
					invalidIds = append(invalidIds, invalidItem)
					continue
				}

				applyPerm := s.checkPermit(c, req.Uname, APPLY_PRIV)
				checkPerm := s.checkPermit(c, req.Uname, CHECK_PRIV)
				if applyPerm && checkPerm {
					//是超级管理员直接放过
					continue
				} else {
					//否则需要检查是不是申请人
					if !applyPerm {
						invalidItem.Msg = "无上下线权限,您可能不是申请人"
						invalidIds = append(invalidIds, invalidItem)
						continue
					}
					//检查卡片是不是自己创建的
					if conf.OperatorName != req.Uname {
						invalidItem.Msg = "卡片申请人只能上下线自己创建的卡片"
						invalidIds = append(invalidIds, invalidItem)
						continue
					}
				}
			}
		}
	}
	return
}

// checkPermit 权限检查 same with php generalRightTyping
func (s *Service) checkPermit(c *bm.Context, uname string, checkType int) (canAccess bool) {
	var applyPerm, checkPerm bool
	applyResp, err := s.auth.Permit(c, &permitPb.PermitReq{Username: uname, Permission: "SEARCH_APPLY_PRIV"})
	if err != nil {
		return
	}
	applyPerm = applyResp.CanAccess

	chechResp, err := s.auth.Permit(c, &permitPb.PermitReq{Username: uname, Permission: "SEARCH_CHECK_PRIV"})
	if err != nil {
		return
	}
	checkPerm = chechResp.CanAccess

	switch checkType {
	// 单独验证审核人权限或申请人权限，如编辑，审核等
	case APPLY_PRIV:
		return applyPerm
	case CHECK_PRIV:
		return checkPerm
	// 有审核人权限或者申请人权限即可访问，如显示、添加等
	case APPLY_CHECK_PRIV:
		return applyPerm || checkPerm
	default:
		return applyPerm || checkPerm
	}
	//nolint:govet
	return
}

func (s *Service) RemoveExpiredSidebar() {
	if err := s.showDao.DisableExpiredSidebar(context.Background()); err != nil {
		log.Error("【Feed-admin报警】移除过期sidebar失败：%s", err)
	}
}
