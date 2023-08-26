package show

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-common/library/railgun"
	"go-common/library/sync/errgroup"
	"go-gateway/app/app-svr/app-job/job/conf"
	arcdao "go-gateway/app/app-svr/app-job/job/dao/archive"
	"go-gateway/app/app-svr/app-job/job/dao/audit"
	bfsdao "go-gateway/app/app-svr/app-job/job/dao/bfs"
	"go-gateway/app/app-svr/app-job/job/dao/card"
	"go-gateway/app/app-svr/app-job/job/dao/favorite"
	feeddao "go-gateway/app/app-svr/app-job/job/dao/feed"
	"go-gateway/app/app-svr/app-job/job/dao/push"
	rcmddao "go-gateway/app/app-svr/app-job/job/dao/recommend"
	showdao "go-gateway/app/app-svr/app-job/job/dao/show"
	"go-gateway/app/app-svr/app-job/job/model"
	showmdl "go-gateway/app/app-svr/app-job/job/model/show"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	creativeAPI "git.bilibili.co/bapis/bapis-go/videoup/open/service"

	"github.com/robfig/cron"
)

// Service is show service.
type Service struct {
	// config
	c    *conf.Config
	cron *cron.Cron
	// dao
	dao     *showdao.Dao
	adao    *audit.Dao
	cdao    *card.Dao
	fdao    *feeddao.Dao
	favDao  *favorite.Dao
	pushDao *push.Dao
	// databus
	weeklySelSub    *databus.Databus
	selResBinlog    *databus.Databus
	archiveHonorPub *databus.Databus
	ottSeriesPub    *databus.Databus
	bfsDao          *bfsdao.Dao
	rcmdDao         *rcmddao.Dao
	arcDao          *arcdao.Dao
	// hotaid
	hotAidsCache map[string]map[int64]string
	// new aggregation
	waiter          sync.WaitGroup
	aggregationChan chan *showmdl.AggregationMsg
	aggregationSub  *databus.Databus
	// 分品类入口
	popularIconCache      map[int64]string
	creativeClient        creativeAPI.VideoUpOpenClient
	showCacheRailGun      *railgun.Railgun
	showRapidCacheRailGun *railgun.Railgun
	showGHCacheRailGun    *railgun.Railgun
}

// New new a show service.
// nolint:biligowordcheck
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:            c,
		dao:          showdao.New(c),
		adao:         audit.New(c),
		cdao:         card.New(c),
		fdao:         feeddao.New(c),
		favDao:       favorite.New(c),
		pushDao:      push.New(c),
		cron:         cron.New(),
		bfsDao:       bfsdao.New(c),
		rcmdDao:      rcmddao.New(c),
		arcDao:       arcdao.New(c),
		hotAidsCache: make(map[string]map[int64]string),
		// new aggregation
		aggregationChan:  make(chan *showmdl.AggregationMsg, 10240),
		popularIconCache: make(map[int64]string),
	}
	var err error
	if s.creativeClient, err = creativeAPI.NewClient(c.CreativeClient); err != nil {
		panic("creativeGRPC not found!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	}
	if model.EnvRun() {
		// only YLF inits the databus and consume
		s.archiveHonorPub = databus.New(c.ArchiveHonorPub)
		s.ottSeriesPub = databus.New(c.OTTSeriesPub)
		// AI weekly selected data
		s.weeklySelSub = databus.New(c.AISelectedSub)
		s.waiter.Add(1)
		go s.weeklyInsertion() // subscribe to AI weekly selected data
		// aggregationSub
		s.aggregationSub = databus.New(c.AggregationSub)
		s.waiter.Add(1)
		go s.aggregationConsumeproc()
		s.waiter.Add(1)
		go s.aggregationproc()
		if s.c.Custom.RefreshSwitchOn {
			go s.refreshHonorLink(s.c.WeeklySel.MinSerieNumber, s.c.WeeklySel.MaxSerieNumber)
		}
	}
	s.selResBinlog = databus.New(c.SelResBinlogSub) // 订阅db的canal，保证db变化后才删缓存
	s.waiter.Add(1)
	go s.consumePopular() // 两边机房消费各自的canal
	s.initShowCacheRailGun()
	s.initGoodHistoryRailGun()
	s.initRapidCacheRailGun()
	s.initCron()
	s.cron.Start()
	return
}

// nolint:errcheck
func (s *Service) initCron() {
	if model.EnvRun() {
		s.pub()
		s.hotLabel()
		s.popularIcon()
		s.aggregationMaterial()
		checkErr(s.cron.AddFunc("@every 2m", s.pub))
		checkErr(s.cron.AddFunc("@every 1m", s.hotLabel))
		checkErr(s.cron.AddFunc("@every 1m", s.popularIcon))
		checkErr(s.cron.AddFunc("@every 10m", s.aggregationMaterial)) // 10分钟刷新一次热点对应物料
		checkErr(s.cron.AddFunc(s.c.WechatAlert.AI.Cron, func() { s.alertAI(_sTypeWeeklySelected) }))
		checkErr(s.cron.AddFunc(s.c.WechatAlert.Audit.Cron, func() { s.alertAuditor(_sTypeWeeklySelected) }))
	}
	checkErr(s.cron.AddFunc(s.c.Popular.PopularCardCron, s.loadPopularCard))
	checkErr(s.cron.AddFunc(s.c.Popular.PopularCardCron, s.loadCarPopularCard)) // 车载热门
	checkErr(s.cron.AddFunc("@every 10s", s.entrance))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// pub publish show data by timer.
func (s *Service) pub() {
	c := context.Background()
	now := time.Now()
	ps, err := s.dao.PTime(c, now)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	if len(ps) == 0 {
		return
	}
	tx, err := s.dao.BeginTran(c)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	for _, p := range ps {
		if err = s.dao.Pub(tx, p); err != nil {
			_ = tx.Rollback()
			log.Error("%+v", err)
			return
		}
	}
	if err = tx.Commit(); err != nil {
		log.Error("%+v", err)
		return
	}
	log.Info("show publish success plats(%v)", ps)
}

// nolint:gocognit
func (s *Service) hotLabel() {
	tmpHotAids, err := s.rcmdDao.Recommend(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	if len(tmpHotAids) == 0 {
		log.Error("hot archive empty")
		return
	}
	tmpHotAidsCache := make(map[string]map[int64]string)
	for _aid := range tmpHotAids {
		aid := _aid
		time.Sleep(time.Millisecond * 100)
		var arc *arcgrpc.Arc
		if arc, err = s.arcDao.Arc(context.Background(), aid); err != nil {
			log.Error("%v", err)
			continue
		}
		if arc == nil {
			log.Error("hotLabel arc(%d) nil", aid)
			continue
		}
		if arc.Pic == "" {
			log.Error("hotLabel arc(%d) cover empty", aid)
			continue
		}
		// NOTE: if true make diff, false force refresh.
		var isTopLeftExist, isTopRightExist, isBottomExist bool
		if s.c.HotLabels.IsDiff {
			for pos, hotLabel := range s.hotAidsCache {
				if hotLabel == nil {
					continue
				}
				switch pos {
				case "top_left":
					if pic, ok := hotLabel[aid]; ok && pic == arc.Pic {
						isTopLeftExist = true
						tl, ok := tmpHotAidsCache["top_left"]
						if !ok {
							tl = make(map[int64]string)
							tmpHotAidsCache["top_left"] = tl
						}
						tl[aid] = pic
					}
				case "top_right":
					if pic, ok := hotLabel[aid]; ok && pic == arc.Pic {
						isTopRightExist = true
						tr, ok := tmpHotAidsCache["top_right"]
						if !ok {
							tr = make(map[int64]string)
							tmpHotAidsCache["top_right"] = tr
						}
						tr[aid] = pic
					}
				case "bottom":
					if pic, ok := hotLabel[aid]; ok && pic == arc.Pic {
						isBottomExist = true
						b, ok := tmpHotAidsCache["bottom"]
						if !ok {
							b = make(map[int64]string)
							tmpHotAidsCache["bottom"] = b
						}
						b[aid] = pic
					}
				}
			}
		}
		if isTopLeftExist && isTopRightExist && isBottomExist {
			continue
		}
		// 1:1的居中裁剪(600*600)
		var resp *http.Response
		if resp, err = http.Get(fmt.Sprintf("%s@!popular-all-share11", arc.Pic)); err != nil {
			log.Error("%v", err)
			continue
		}
		defer resp.Body.Close()
		var (
			tailor11      = true
			h11, w11, v11 float64
		)
		w11, _ = strconv.ParseFloat(resp.Header.Get("o-width"), 64)
		h11, _ = strconv.ParseFloat(resp.Header.Get("o-height"), 64)
		if h11 != 0 {
			if v11, err = strconv.ParseFloat(fmt.Sprintf("%.1f", w11/h11), 64); err != nil || v11 != 1 {
				log.Error("hotLabel popular-all-share11 faild aid(%d) w(%v) h(%v) v(%v)", aid, w11, h11, v11)
				tailor11 = false
			}
		} else {
			tailor11 = false
		}
		var bs []byte
		if bs, err = ioutil.ReadAll(resp.Body); err != nil {
			log.Error("%v", err)
			continue
		}
		// 5:4的居中裁剪(750*600)
		var resp2 *http.Response
		if resp2, err = http.Get(fmt.Sprintf("%s@!popular-all-share54", arc.Pic)); err != nil {
			log.Error("%v", err)
			continue
		}
		defer resp2.Body.Close()
		var (
			tailor54      = true
			h54, w54, v54 float64
		)
		w54, _ = strconv.ParseFloat(resp2.Header.Get("o-width"), 64)
		h54, _ = strconv.ParseFloat(resp2.Header.Get("o-height"), 64)
		if h54 != 0 {
			if v54, err = strconv.ParseFloat(fmt.Sprintf("%.2f", w54/h54), 64); err != nil || v54 > 1.30 || v54 < 1.20 {
				log.Error("hotLabel popular-all-share54 faild aid(%d) w(%v) h(%v) v(%v)", aid, w54, h54, v54)
				tailor54 = false
			}
		} else {
			tailor54 = false
		}
		var bs2 []byte
		if bs2, err = ioutil.ReadAll(resp2.Body); err != nil {
			log.Error("%v", err)
			continue
		}
		fileName := strconv.FormatInt(aid, 10)
		filenameWithSuffix := path.Base(arc.Pic)
		fileSuffix := path.Ext(filenameWithSuffix)
		var mutex sync.Mutex
		g, ctx := errgroup.WithContext(context.Background())
		if s.c.HotLabels.TopLeft != nil && !isTopLeftExist && tailor11 {
			// 左上角水印用600*600居中裁剪图
			g.Go(func() (err error) {
				if _, err = s.bfsDao.Upload(ctx, s.c.HotLabels.Bucket, s.c.HotLabels.Dir, fileName, "", bs,
					s.c.HotLabels.TopLeft.WMKey, s.c.HotLabels.TopLeft.WMPaddingX, s.c.HotLabels.TopLeft.WMPaddingY,
					s.c.HotLabels.TopLeft.WMScale, s.c.HotLabels.TopLeft.WMPos, s.c.HotLabels.TopLeft.WMTransparency); err != nil {
					log.Error("%v", err)
					return nil
				}
				mutex.Lock()
				var (
					tmp map[int64]string
					ok  bool
				)
				if tmp, ok = tmpHotAidsCache["top_left"]; !ok {
					tmp = make(map[int64]string)
					tmpHotAidsCache["top_left"] = tmp
				}
				tmp[aid] = arc.Pic
				mutex.Unlock()
				return
			})
		}
		if s.c.HotLabels.TopRight != nil && !isTopRightExist && tailor54 {
			// 右上角水印用750*600居中裁剪图
			g.Go(func() (err error) {
				if _, err = s.bfsDao.Upload(ctx, s.c.HotLabels.Bucket, s.c.HotLabels.Dir, fmt.Sprintf("%s%s%s", fileName, s.c.HotLabels.TopRight.Suffix, fileSuffix), "", bs2,
					s.c.HotLabels.TopRight.WMKey, s.c.HotLabels.TopRight.WMPaddingX, s.c.HotLabels.TopRight.WMPaddingY*uint32(h54)/600,
					s.c.HotLabels.TopRight.WMScale, s.c.HotLabels.TopRight.WMPos, s.c.HotLabels.TopRight.WMTransparency); err != nil {
					log.Error("%v", err)
					return nil
				}
				mutex.Lock()
				var (
					tmp map[int64]string
					ok  bool
				)
				if tmp, ok = tmpHotAidsCache["top_right"]; !ok {
					tmp = make(map[int64]string)
					tmpHotAidsCache["top_right"] = tmp
				}
				tmp[aid] = arc.Pic
				mutex.Unlock()
				return
			})
		}
		if s.c.HotLabels.Bottom != nil && !isBottomExist && tailor11 {
			//底部水印用600*600居中裁剪图
			g.Go(func() (err error) {
				if _, err = s.bfsDao.Upload(ctx, s.c.HotLabels.Bucket, s.c.HotLabels.Dir, fmt.Sprintf("%s%s%s", fileName, s.c.HotLabels.Bottom.Suffix, fileSuffix), "", bs,
					s.c.HotLabels.Bottom.WMKey, s.c.HotLabels.Bottom.WMPaddingX, s.c.HotLabels.Bottom.WMPaddingY,
					s.c.HotLabels.Bottom.WMScale, s.c.HotLabels.Bottom.WMPos, s.c.HotLabels.Bottom.WMTransparency); err != nil {
					log.Error("%v", err)
					return nil
				}
				mutex.Lock()
				var (
					tmp map[int64]string
					ok  bool
				)
				if tmp, ok = tmpHotAidsCache["bottom"]; !ok {
					tmp = make(map[int64]string)
					tmpHotAidsCache["bottom"] = tmp
				}
				tmp[aid] = arc.Pic
				mutex.Unlock()
				return
			})
		}
		if err = g.Wait(); err != nil {
			log.Error("%v", err)
			continue
		}
	}
	s.hotAidsCache = tmpHotAidsCache
}

func (s *Service) popularIcon() {
	popularIcon, err := s.dao.AllEntranceIcon(context.Background())
	if err != nil {
		log.Error("%v", err)
		return
	}
	if popularIcon == nil {
		return
	}
	tmpPopularIcon := make(map[int64]string)
	for id, icon := range popularIcon {
		if iconCache, ok := s.popularIconCache[id]; ok && icon == iconCache {
			continue
		}
		var resp *http.Response
		if resp, err = http.Get(fmt.Sprintf("%s@120w_120h_1e_1c", icon)); err != nil {
			log.Error("%v", err)
			continue
		}
		defer resp.Body.Close()
		var bs []byte
		if bs, err = ioutil.ReadAll(resp.Body); err != nil {
			log.Error("%v", err)
			continue
		}
		filenameWithSuffix := path.Base(icon)
		fileSuffix := path.Ext(filenameWithSuffix)
		fileName := strings.TrimSuffix(filenameWithSuffix, fileSuffix)
		if _, err = s.bfsDao.Upload(context.Background(), s.c.HotLabels.Bucket, s.c.HotLabels.Dir, fmt.Sprintf("%s%s%s", fileName, s.c.HotLabels.Bottom.Suffix, fileSuffix), "", bs,
			s.c.HotLabels.Bottom.WMKey, s.c.HotLabels.Bottom.WMPaddingX, s.c.HotLabels.Bottom.WMPaddingY,
			s.c.HotLabels.Bottom.WMScale, s.c.HotLabels.Bottom.WMPos, s.c.HotLabels.Bottom.WMTransparency); err != nil {
			log.Error("%v", err)
			continue
		}
		tmpPopularIcon[id] = icon
		time.Sleep(time.Millisecond * 100)
	}
	s.popularIconCache = tmpPopularIcon
}

func (s *Service) entrance() {
	if err := s.dao.AddEntranceCache(context.Background()); err != nil {
		log.Error("%+v", err)
		return
	}
	log.Info("AddEntranceCache success")
}

func (s *Service) Ping(c context.Context) (err error) {
	return
}

func (s *Service) Close() {
	s.cron.Stop()
	if model.EnvRun() {
		s.archiveHonorPub.Close()
		s.aggregationSub.Close()
		s.weeklySelSub.Close()
	}
	s.selResBinlog.Close()
	s.waiter.Wait()
	s.cron.Stop()
}

// nolint:errcheck
func (s *Service) initShowCacheRailGun() {
	if err := s.loadShowCache(); err != nil {
		panic(fmt.Sprintf("loadShowCache error:%+v", err))
	}
	if err := s.loadShowTempCache(); err != nil {
		panic(fmt.Sprintf("loadShowTempCache error:%+v", err))
	}
	if err := s.loadArticleCardsCache(); err != nil {
		panic(fmt.Sprintf("loadArticleCardsCache error:%+v", err))
	}
	if err := s.loadCardSetCache(); err != nil {
		panic(fmt.Sprintf("loadCardSetCache error:%+v", err))
	}
	if err := s.loadEventTopicCache(); err != nil {
		panic(fmt.Sprintf("loadEventTopicCache error:%+v", err))
	}
	if err := s.loadCardCache(); err != nil {
		panic(fmt.Sprintf("loadCardCache error:%+v", err))
	}
	if err := s.loadAuditCache(); err != nil {
		panic(fmt.Sprintf("loadAuditCache error:%+v", err))
	}
	if err := s.loadColumnListCache(); err != nil {
		panic(fmt.Sprintf("loadColumnListCache error:%+v", err))
	}
	if err := s.loadColumnsCache(); err != nil {
		panic(fmt.Sprintf("loadColumnsCache error:%+v", err))
	}
	if err := s.loadNperCache(); err != nil {
		panic(fmt.Sprintf("loadNperCache error:%+v", err))
	}

	r := railgun.NewRailGun("loadShowCache", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "0 */3 * * * *"}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		s.loadShowCache()
		s.loadShowTempCache()
		s.loadArticleCardsCache()
		s.loadCardSetCache()
		s.loadEventTopicCache()
		s.loadCardCache()
		s.loadAuditCache()
		s.loadColumnListCache()
		s.loadColumnsCache()
		s.loadNperCache()
		return railgun.MsgPolicyNormal
	}))
	s.showCacheRailGun = r
	r.Start()
}

// nolint:errcheck
func (s *Service) initGoodHistoryRailGun() {
	if err := s.loadGoodHistory(); err != nil {
		panic(fmt.Sprintf("loadGoodHistory error:%+v", err))
	}
	r := railgun.NewRailGun("loadGoodHistory", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "*/10 * * * * *"}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		s.loadGoodHistory()
		return railgun.MsgPolicyNormal
	}))
	s.showGHCacheRailGun = r
	r.Start()
}

// nolint:errcheck
func (s *Service) initRapidCacheRailGun() {
	if err := s.loadLiveCards(); err != nil {
		panic(fmt.Sprintf("loadLiveCards error:%+v", err))
	}
	if err := s.loadLargeCards(); err != nil {
		panic(fmt.Sprintf("loadLargeCards error:%+v", err))
	}
	if err := s.loadMidTopPhoto(); err != nil {
		panic(fmt.Sprintf("loadMidTopPhoto error:%+v", err))
	}
	if err := s.loadPopEntrances(); err != nil {
		panic(fmt.Sprintf("loadPopEntrances error:%+v", err))
	}
	r := railgun.NewRailGun("loadRapidShowCache", nil, railgun.NewCronInputer(&railgun.CronInputerConfig{Spec: "*/3 * * * * *"}), railgun.NewCronProcessor(nil, func(ctx context.Context) railgun.MsgPolicy {
		s.loadLiveCards()
		s.loadLargeCards()
		s.loadMidTopPhoto()
		s.loadPopEntrances()
		return railgun.MsgPolicyNormal
	}))
	s.showRapidCacheRailGun = r
	r.Start()
}
