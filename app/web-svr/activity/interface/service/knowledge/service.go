package knowledge

import (
	"context"
	"time"

	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/dao/knowledge"
	model "go-gateway/app/web-svr/activity/interface/model/knowledge"
)

var (
	goingKnowledgeConfigMap map[int64]*model.KnowConfig
)

const (
	bvidJsonUrl = "http://activity.hdslb.com/blackboard/static/jsonlist/231/rI9eRj4RQv.json"
)

func init() {
	goingKnowledgeConfigMap = make(map[int64]*model.KnowConfig)
}

type Service struct {
	c                *conf.Config
	dao              *knowledge.Dao
	cache            *fanout.Fanout
	KnowledgeBvidMap map[string]struct{}
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   knowledge.New(c),
		cache: fanout.New("like_service_cache", fanout.Worker(1), fanout.Buffer(1024)),
	}
	s.KnowledgeBvidMap = make(map[string]struct{})
	ctx := context.Background()
	s.initKnowledgeList(ctx)
	s.fetchKnowledgeConfig()
	go s.updateKnowledgeListloop()
	return s
}

func (s *Service) GetBvidMap() map[string]struct{} {
	return s.KnowledgeBvidMap
}

func (s *Service) updateKnowledgeListloop() {
	ctx := context.Background()
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		err := s.initKnowledgeList(ctx)
		if err != nil {
			continue
		}
	}
}

func (s *Service) initKnowledgeList(ctx context.Context) error {
	bvidPeriod, err := s.dao.BvidList(ctx, bvidJsonUrl, time.Now().Unix())
	if err != nil {
		log.Errorc(ctx, "s.dao.OgvLink err(%v)", err)
		return err
	}
	bvidMap := make(map[string]struct{})
	if bvidPeriod != nil && len(bvidPeriod.PeriodList) > 0 {
		for _, v := range bvidPeriod.PeriodList {
			if len(v.WeekList) > 0 {
				for _, w := range v.WeekList {
					if len(w.GoldList) > 0 {
						for _, g := range w.GoldList {
							bvid := g
							bvidMap[bvid] = struct{}{}
						}
					}
					if len(w.HorseList) > 0 {
						for _, h := range w.HorseList {
							bvid := h
							bvidMap[bvid] = struct{}{}
						}
					}
				}
			}
			if len(v.SuperList) > 0 {
				for _, s := range v.SuperList {
					bvid := s
					bvidMap[bvid] = struct{}{}
				}
			}
		}
	}
	s.KnowledgeBvidMap = bvidMap
	return nil
}

func (s *Service) fetchKnowledgeConfig() {
	knowConfigs, err := s.dao.RawFetchKnowledgeConfigs(context.Background())
	if err != nil {
		panic(err)
	}
	goingKnowledgeConfigMap = knowConfigs
}
