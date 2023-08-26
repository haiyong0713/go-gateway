package duertv

import (
	"context"

	"go-common/library/log"
	"go-common/library/railgun"
	"go-gateway/app/app-svr/app-car/job/model/region"
)

func (s *Service) loadRegionCache() {
	regions, err := s.rg.AndroidAll(context.Background())
	if err != nil {
		log.Error("%+v", err)
		return
	}
	tmpRegion := map[int32]*region.Region{}
	tmp := map[int32]*region.Region{}
	// 先拿出所有分区
	for _, v := range regions {
		tmpRegion[v.Rid] = v
	}
	for _, v := range regions {
		// 只要二级分区，然后去找对应的一级分区
		if v.Reid != 0 {
			if reg, ok := tmpRegion[v.Reid]; ok {
				tmp[v.Rid] = reg
			}
		}
	}
	s.oneRegions = tmp
}

func (s *Service) initRegionRailGun(cronInputer *railgun.CronInputerConfig, cronProcessor *railgun.CronProcessorConfig, cfg *railgun.Config) {
	s.loadRegionCache()
	// 每5分钟跑一次
	inputer := railgun.NewCronInputer(cronInputer)
	processor := railgun.NewCronProcessor(cronProcessor, func(ctx context.Context) railgun.MsgPolicy {
		s.loadRegionCache()
		return railgun.MsgPolicyNormal
	})
	r := railgun.NewRailGun("车载分区缓存映射更新", cfg, inputer, processor)
	s.regionRailGun = r
}
