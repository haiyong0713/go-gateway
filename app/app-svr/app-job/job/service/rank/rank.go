package rank

import (
	"context"
	"fmt"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-job/job/conf"
	rankdao "go-gateway/app/app-svr/app-job/job/dao/rank"

	"github.com/robfig/cron"
)

const (
	_rankBangumi = "bangumi"
	_rankAll     = "all"
	_rankOrigin  = "origin"
)

var (
	// 番剧 动画，音乐，舞蹈，游戏，科技，娱乐，鬼畜，电影，时尚, 生活，连载番剧（二级分区），国漫，影视，纪录片，国创相关，数码，动物圈，汽车，运动
	_tids = []int{13, 1, 3, 129, 4, 36, 5, 119, 23, 155, 160, 11, 33, 167, 181, 177, 168, 188, 211, 217, 223, 234}
)

type Service struct {
	c     *conf.Config
	cron  *cron.Cron
	rankd *rankdao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		cron:  cron.New(),
		rankd: rankdao.New(c),
	}
	s.load()
	checkErr(s.cron.AddFunc("@every 3m", s.load)) // 间隔3分钟
	s.cron.Start()
	return
}

//nolint:gomnd
func (s *Service) load() {
	ctx := context.Background()
	for _, rid := range _tids {
		switch rid {
		case 33: // 新番排行版
			rankList, err := s.rankd.RankAppBangumi(ctx)
			if err != nil || len(rankList) < 5 {
				log.Error("s.rcmmnd.RankAppBangumi len lt 5 OR error(%v)", err)
				continue
			}
			if err = s.rankd.AddRankCache(ctx, _rankBangumi, 0, rankList); err != nil {
				log.Error("%+v", err)
				continue
			}
		default: // 分区排行榜
			rankList, err := s.rankd.RankAppRegion(ctx, rid)
			if err != nil || len(rankList) < 5 {
				log.Error("s.rcmmnd.RankAppRegion len lt 5 OR error(%v)", err)
				continue
			}
			if err = s.rankd.AddRankCache(ctx, _rankAll, rid, rankList); err != nil {
				log.Error("%+v", err)
				continue
			}
		}
	}
	// 全站排行版
	rankList, err := s.rankd.RankAppAll(ctx)
	if err != nil || len(rankList) < 5 {
		log.Error("s.rcmmnd.RankAppAll len lt 5 OR error(%v)", err)
	}
	if len(rankList) >= 5 {
		if err = s.rankd.AddRankCache(ctx, _rankAll, 0, rankList); err != nil {
			log.Error("%+v", err)
			return
		}
	}
	// 原创排行版
	rankList, err = s.rankd.RankAppOrigin(ctx)
	if err != nil || len(rankList) < 5 {
		log.Error("s.rcmmnd.RankAppAll len lt 5 OR error(%v)", err)
	}
	if len(rankList) >= 5 {
		if err = s.rankd.AddRankCache(ctx, _rankOrigin, 0, rankList); err != nil {
			log.Error("%+v", err)
			return
		}
	}
}

func checkErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("cron add func loadCache error(%+v)", err))
	}
}

func (s *Service) Close() {
	s.cron.Stop()
}
