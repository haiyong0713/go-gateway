package rank

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-common/library/net/trace"

	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
	"time"
)

var rankJobCtx context.Context

var cronJobMap map[string]*rankmdl.Rank

func init() {
	if cronJobMap == nil {
		cronJobMap = make(map[string]*rankmdl.Rank)
	}
	rankJobCtx = trace.SimpleServerTrace(context.Background(), "rankJob")
}

// getRankOnline 获取排行配置
func (s *Service) getRankOnline(c context.Context) (rank []*rankmdl.Rank, err error) {
	now := time.Now()
	return s.rankDao.GetRankConfigOnline(c, now)
}

// Job ...
type Job struct {
	Func          func(rank *rankmdl.Rank, rankAttribute uint)
	Rank          *rankmdl.Rank
	RankAttribute uint
}

// Run ...
func (f Job) Run() {
	f.Func(f.Rank, f.RankAttribute)
}

func (s *Service) getCronMapName(rank *rankmdl.Rank, attributeType uint) string {
	return fmt.Sprintf("%d_%d", rank.ID, attributeType)
}

// RankJob 排行榜job
func (s *Service) RankJob() {

	allRank, err := s.getRankOnline(rankJobCtx)
	if err != nil {
		log.Errorc(rankJobCtx, "s.getRankOnline err(%v)", err)
		return
	}
	if allRank != nil {
		for _, v := range allRank {
			var err error
			crons := v.GetStatisticsCron()

			if crons != nil {
				for _, cron := range crons {
					job := Job{
						RankAttribute: cron.Type,
						Func:          s.Rank,
						Rank:          v,
					}
					if _, ok := cronJobMap[s.getCronMapName(v, cron.Type)]; !ok {
						log.Infoc(rankJobCtx, "add func map key (%s)", s.getCronMapName(v, cron.Type))
						if err = s.cron.AddJob(cron.Cron, job); err != nil {
							panic(err)
						}
						cronJobMap[s.getCronMapName(v, cron.Type)] = v
					} else {
						log.Infoc(rankJobCtx, "add func map key (%s) already", s.getCronMapName(v, cron.Type))
					}

				}
			}
		}
		s.cron.Start()

	}

}

// GetRankCronMap 获得运行中的脚本map
func (s *Service) GetRankCronMap() map[string]*rankmdl.Rank {
	return cronJobMap
}
