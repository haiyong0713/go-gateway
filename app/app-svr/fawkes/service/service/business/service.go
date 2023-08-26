package business

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/fawkes/service/conf"
	fkdao "go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	ossdao "go-gateway/app/app-svr/fawkes/service/dao/oss"
	"go-gateway/app/app-svr/fawkes/service/model"
	appmdl "go-gateway/app/app-svr/fawkes/service/model/app"
	bizapkmdl "go-gateway/app/app-svr/fawkes/service/model/bizapk"
	cdmdl "go-gateway/app/app-svr/fawkes/service/model/cd"
	"go-gateway/app/app-svr/fawkes/service/model/mod"
	tribemdl "go-gateway/app/app-svr/fawkes/service/model/tribe"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/robfig/cron"
)

// Service struct.
type Service struct {
	c                      *conf.Config
	fkDao                  *fkdao.Dao
	ossDao                 *ossdao.Dao
	configVersionCache     map[string]map[string]int64
	ffVersionCache         map[string]map[string]int64
	patchAllCache          map[string]map[string]*cdmdl.Patch
	packAllCache           map[string]map[int64][]*cdmdl.Pack
	bizApkListAllCache     map[int64]map[string]map[string][]*bizapkmdl.Apk
	tribeListAllCache      map[string]map[int64]map[string]map[string][]*tribemdl.TribeApk
	tribeHostRelationCache map[int64]int64
	upgradConfigAllCache   map[string]map[int64]*cdmdl.UpgradConfig
	versionAllCache        map[string]map[int64]*model.Version
	hotfixAllCache         map[string]map[int64][]*appmdl.HfUpgrade
	flowConfigAllCache     map[string]map[int64]*cdmdl.FlowConfig
	modCache               sync.Map // map[env][appKey]map[pool_name]map[module_name]
	md5Cache               sync.Map //key env_appkey value md5
	cron                   *cron.Cron
}

// New new service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:                      c,
		fkDao:                  fkdao.New(c),
		ossDao:                 ossdao.New(c),
		configVersionCache:     make(map[string]map[string]int64),
		ffVersionCache:         make(map[string]map[string]int64),
		patchAllCache:          make(map[string]map[string]*cdmdl.Patch),
		packAllCache:           make(map[string]map[int64][]*cdmdl.Pack),
		bizApkListAllCache:     make(map[int64]map[string]map[string][]*bizapkmdl.Apk),
		tribeListAllCache:      make(map[string]map[int64]map[string]map[string][]*tribemdl.TribeApk),
		tribeHostRelationCache: make(map[int64]int64),
		upgradConfigAllCache:   make(map[string]map[int64]*cdmdl.UpgradConfig),
		versionAllCache:        make(map[string]map[int64]*model.Version),
		hotfixAllCache:         make(map[string]map[int64][]*appmdl.HfUpgrade),
		flowConfigAllCache:     make(map[string]map[int64]*cdmdl.FlowConfig),
		cron:                   cron.New(),
	}
	s.loadVersion()
	if err := s.loadModModuleListAll(); err != nil {
		panic(fmt.Sprintf("loadModModuleListAll error:%+v", err))
	}
	s.loadPackAll()
	s.loadBizApkListAll()
	s.loadTribeListAll()
	s.loadUpgradConfigAll()
	s.loadVersionAll()
	s.loadHotfixAll()
	s.loadFlowConfigAll()

	_ = s.cron.AddFunc(s.c.Cron.LoadVersion, s.loadVersion)
	_ = s.cron.AddFunc(s.c.Cron.LoadModuleListAll, func() { _ = s.loadModModuleListAll() })
	_ = s.cron.AddFunc(s.c.Cron.LoadPackAll, s.loadPackAll)
	_ = s.cron.AddFunc(s.c.Cron.LoadBizApkListAll, s.loadBizApkListAll)
	_ = s.cron.AddFunc(s.c.Cron.LoadTribeListAll, s.loadTribeListAll)
	_ = s.cron.AddFunc(s.c.Cron.LoadUpgradConfigAll, s.loadUpgradConfigAll)
	_ = s.cron.AddFunc(s.c.Cron.LoadVersionAll, s.loadVersionAll)
	_ = s.cron.AddFunc(s.c.Cron.LoadHotfixAll, s.loadHotfixAll)
	_ = s.cron.AddFunc(s.c.Cron.LoadFlowConfigAll, s.loadFlowConfigAll)
	s.cron.Start()
	return
}

func (s *Service) loadVersion() {
	log.Info("cronLog start loadVersion")
	var (
		tmpConfigVersion, tmpFFVersion map[string]map[string]int64
		tmpPatchAll                    map[string]map[string]*cdmdl.Patch
		err                            error
	)
	if tmpConfigVersion, err = s.fkDao.NewestConfigVersion(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	if tmpFFVersion, err = s.fkDao.NewestFFVersion(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	if tmpPatchAll, err = s.PatchAll(context.Background()); err != nil {
		log.Error("%v", err)
		return
	}
	s.configVersionCache = tmpConfigVersion
	s.ffVersionCache = tmpFFVersion
	s.patchAllCache = tmpPatchAll
}

func (s *Service) loadModModuleListAll() (err error) {
	log.Info("start loadModModuleListAll")
	ctx := context.Background()
	pool, err := s.fkDao.ModBusPoolList(ctx)
	if err != nil {
		return err
	}
	var poolIDs []int64
	for _, v := range pool {
		for _, p := range v {
			poolIDs = append(poolIDs, p.ID)
		}
	}
	if len(poolIDs) == 0 {
		return nil
	}
	module, err := s.fkDao.ModBusModuleList(ctx, poolIDs)
	if err != nil {
		return err
	}
	var g errgroup.Group
	envs := []mod.Env{mod.EnvProd, mod.EnvTest}
	for _, val := range envs {
		env := val
		g.Go(func(ctx context.Context) error {
			res, err := s.moduleList(ctx, pool, module, env)
			if err != nil {
				log.Error("日志告警 mod 全量缓存构建失败,env:%v,error:%+v", env, err)
				return err
			}
			s.modCache.Store(env, res)
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%v", err)
		return err
	}
	// 构建md5Cache
	s.modCache.Range(func(key, value interface{}) bool {
		k := key.(mod.Env)
		v := value.(map[string]map[string]map[string][]*mod.BusFile)
		var marshalErr error
		var bs []byte
		for appKey, val := range v {
			md5CacheKey := string(k) + "_" + appKey
			bs, marshalErr = json.Marshal(val)
			if marshalErr != nil {
				err = marshalErr
				return false
			}
			hm := md5.Sum(bs)
			newMd5Val := hex.EncodeToString(hm[:])
			s.md5Cache.Store(md5CacheKey, newMd5Val)
		}
		return true
	})
	return err
}

func (s *Service) loadPackAll() {
	log.Info("cronLog loadPackAll start")
	var (
		packAll map[string]map[int64][]*cdmdl.Pack
		err     error
	)
	if packAll, err = s.fkDao.PackAll(context.Background()); err != nil {
		log.Error("cronLog loadPackAll error %v", err)
	}
	s.packAllCache = packAll
}

func (s *Service) loadBizApkListAll() {
	log.Info("cronLog loadBizApkListAll start")
	var (
		bizApkListAll map[int64]map[string]map[string][]*bizapkmdl.Apk
		err           error
	)
	if bizApkListAll, err = s.BizApkListAll(context.Background()); err != nil {
		log.Error("cronLog loadBizApkListAll error %v", err)
	}
	s.bizApkListAllCache = bizApkListAll
}

func (s *Service) loadTribeListAll() {
	log.Info("cronLog loadTribeListAll start")
	var (
		tribeListAll      map[string]map[int64]map[string]map[string][]*tribemdl.TribeApk
		tribeHostRelation map[int64]int64
		err               error
	)
	if tribeListAll, tribeHostRelation, err = s.TribeListAll(context.Background()); err != nil {
		log.Error("cronLog TribeListAll error %v", err)
	}
	s.tribeListAllCache = tribeListAll
	s.tribeHostRelationCache = tribeHostRelation
}

func (s *Service) loadUpgradConfigAll() {
	log.Info("cronLog loadUpgradConfigAll start")
	var (
		upgradConfigAll map[string]map[int64]*cdmdl.UpgradConfig
		err             error
	)
	if upgradConfigAll, err = s.fkDao.UpgradConfigAll(context.Background()); err != nil {
		log.Error("cronLog loadUpgradConfigAll error %v", err)
	}
	s.upgradConfigAllCache = upgradConfigAll
}

func (s *Service) loadVersionAll() {
	log.Info("cronLog loadVersionAll start")
	var (
		versionAll map[string]map[int64]*model.Version
		err        error
	)
	if versionAll, err = s.fkDao.VersionAll(context.Background()); err != nil {
		log.Error("cronLog loadVersionAll error %v", err)
	}
	s.versionAllCache = versionAll
}

func (s *Service) loadHotfixAll() {
	log.Info("cronLog loadHotfixAll start")
	var (
		hotfixAll map[string]map[int64][]*appmdl.HfUpgrade
		err       error
	)
	if hotfixAll, err = s.HotfixAll(context.Background()); err != nil {
		log.Error("cronLog loadHotfixAll error %v", err)
	}
	s.hotfixAllCache = hotfixAll
}

func (s *Service) loadFlowConfigAll() {
	log.Info("cronLog loadFlowConfigAll start")
	var (
		flowConfigAll map[string]map[int64]*cdmdl.FlowConfig
		err           error
	)
	if flowConfigAll, err = s.fkDao.FlowConfigAll(context.Background()); err != nil {
		log.Error("cronLog loadFlowConfigAll error %v", err)
	}
	s.flowConfigAllCache = flowConfigAll
}

// Ping dao.
func (s *Service) Ping(c context.Context) (err error) {
	if err = s.fkDao.Ping(c); err != nil {
		log.Error("s.dao error(%v)", err)
	}
	return
}

// Close dao.
func (s *Service) Close() {
	s.cron.Stop()
	s.fkDao.Close()
}
