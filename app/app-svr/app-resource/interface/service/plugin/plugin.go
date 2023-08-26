package plugin

import (
	"context"
	"sort"
	"strings"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	pgdao "go-gateway/app/app-svr/app-resource/interface/dao/plugin"
	"go-gateway/app/app-svr/app-resource/interface/model/plugin"

	"github.com/robfig/cron"
)

type Service struct {
	c           *conf.Config
	pgDao       *pgdao.Dao
	tick        time.Duration
	pluginCache map[string][]*plugin.Plugin
	// cron
	cron *cron.Cron
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:           c,
		pgDao:       pgdao.New(c),
		tick:        time.Duration(c.Tick),
		pluginCache: map[string][]*plugin.Plugin{},
		// cron
		cron: cron.New(),
	}
	s.initCron()
	s.cron.Start()
	return
}

func (s *Service) initCron() {
	s.load()
	if err := s.cron.AddFunc(s.c.Cron.LoadPlugin, s.load); err != nil {
		panic(err)
	}
}

func (s *Service) Plugin(build, baseCode, seed int, name string) (pg *plugin.Plugin) {
	if build == 0 || seed == 0 || name == "" {
		return
	}
	if ps, ok := s.pluginCache[name]; ok {
		for _, p := range ps {
			if ((p.Policy == 1 && baseCode == p.BaseCode) || p.Policy == 2 && baseCode >= p.BaseCode) && seed%100 <= p.Coverage && build >= p.MinBuild && ((p.MaxBuild == 0) || (p.MaxBuild != 0 && build <= p.MaxBuild)) {
				pg = p
				break
			}
		}
	}
	return
}

// load cache data
func (s *Service) load() {
	log.Info("cronLog start plugin load")
	psm, err := s.pgDao.All(context.TODO())
	if err != nil {
		log.Error("s.pgDao.All() error(%v)", err)
		return
	}
	pgCache := make(map[string][]*plugin.Plugin, len(psm))
	for name, ps := range psm {
		sort.Sort(plugin.Plugins(ps))
		pgCache[name] = ps
	}
	s.pluginCache = pgCache
}

func (s *Service) Dependencies(ctx context.Context, param *plugin.TraceParam) (*plugin.DependenceReply, error) {
	edges, err := s.pgDao.Traces(ctx, param)
	if err != nil {
		log.Error("Failed to request trace: %+v", err)
		return nil, err
	}
	operationSet := sets.NewString()
	for _, edge := range edges {
		for _, operation := range edge.OperationTable {
			if operation.SrcOperation != param.Operation {
				continue
			}
			if operation.SrcOperation == operation.DstOperation {
				continue
			}
			if strings.Contains(operation.DstOperation, "HTTP") ||
				strings.Contains(operation.DstOperation, "Fanout") ||
				strings.Contains(operation.DstOperation, "Redis") ||
				strings.Contains(operation.DstOperation, "Memcache") {
				continue
			}
			operationSet.Insert(operation.DstOperation)
		}
	}
	out := &plugin.DependenceReply{
		OperationName: operationSet.List(),
	}
	return out, nil
}
