package param

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/dao/param"
	"go-gateway/app/app-svr/app-resource/interface/model"
	mparam "go-gateway/app/app-svr/app-resource/interface/model/param"

	farm "go-farm"

	"github.com/robfig/cron"
)

const (
	_initVersion = "param_version"
	_platKey     = "param_%d"
)

// Service param service.
type Service struct {
	c    *conf.Config
	dao  *param.Dao
	tick time.Duration
	// model param cache
	cache map[string][]*mparam.Param
	// cron
	cron *cron.Cron
}

// New new a param service.
func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   param.New(c),
		tick:  time.Duration(c.Tick),
		cache: map[string][]*mparam.Param{},
		// cron
		cron: cron.New(),
	}
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	s.load()
	if err := s.cron.AddFunc(s.c.Cron.LoadParam, s.load); err != nil {
		panic(err)
	}
}

// Param return param to string
func (s *Service) Param(plat int8, build int, ver string) (res map[string]string, version string, err error) {
	res, version, err = s.getCache(plat, build, ver)
	return
}

func (s *Service) getCache(plat int8, build int, ver string) (res map[string]string, version string, err error) {
	var (
		pk = fmt.Sprintf(_platKey, plat)
	)
	res = map[string]string{}
	for _, p := range s.cache[pk] {
		if model.InvalidBuild(build, p.Build, p.Condition) {
			continue
		}
		res[p.Name] = p.Value
	}
	if version = s.hash(res); version == ver {
		err = ecode.NotModified
		res = nil
	}
	return
}

func (s *Service) load() {
	log.Info("cronLog start param load")
	tmp, err := s.dao.All(context.TODO())
	if err != nil {
		log.Error("param s.dao.All() error(%v)", err)
		return
	}
	s.cache = tmp
	log.Info("param cacheproc success")
}

func (s *Service) hash(v map[string]string) string {
	bs, err := json.Marshal(v)
	if err != nil {
		log.Error("json.Marshal error(%v)", err)
		return _initVersion
	}
	return strconv.FormatUint(farm.Hash64(bs), 10)
}

// key get banner cache key.
//
//nolint:unused
func (s *Service) key(plat int8, build int) string {
	return fmt.Sprintf("%d_%d", plat, build)
}

// Close dao
func (s *Service) Close() {
	s.dao.Close()
}
