package service

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-gateway/app/web-svr/esports/job/conf"
)

var (
	scorePathOfBigData4Team   = "b/data_team_list.php"
	scorePathOfBigData4Player = "b/data_player2_list.php"
	scorePathOfBigData4Hero   = "b/data_hero_list.php"

	scoreKeyOfTournamentID = "tournamentID"
)

func (s *Service) SyncScoreAnalysisBiz() {
	cfg := conf.LoadScoreAnalysisCfg()
	if cfg == nil || cfg.TournamentID <= 0 {
		return
	}

	interval := cfg.Interval
	if interval <= 0 {
		interval = 60
	}

	// always init restart chan!!!
	conf.ScoreAnalysisRestartChan = make(chan int, 1)
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().Unix()
			if now >= cfg.EndTime {
				return
			}

			if now < cfg.StartTime {
				continue
			}

			if !conf.IsScoreAnalysisEnabled() {
				continue
			}

			s.syncScoreAnalysis(context.Background(), cfg.TournamentID)
		case _, ok := <-conf.ScoreAnalysisRestartChan:
			if !ok {
				go s.SyncScoreAnalysisBiz()

				return
			}
		}
	}
}

func (s *Service) syncScoreAnalysis(ctx context.Context, tournamentID int64) {
	//s.fetchTeamAnalysis(ctx, tournamentID)
	//s.fetchPlayerAnalysis(ctx, tournamentID)
	//s.fetchHeroAnalysis(ctx, tournamentID)
}

func genUrlValuesByTournamentID(tournamentID int64) url.Values {
	urlValues := url.Values{}
	{
		urlValues.Add(scoreKeyOfTournamentID, strconv.FormatInt(tournamentID, 10))
	}

	return urlValues
}

//func (s *Service) fetchTeamAnalysis(ctx context.Context, tournamentID int64) {
//    values := genUrlValuesByTournamentID(tournamentID)
//    res := struct {
//        Data    struct {
//            List []*model.ScoreOriginTeamAnalysis `json:"list"`
//        } `json:"data"`
//    }{}
//
//    if err := s.getScoreData(ctx, scorePathOfBigData4Team, values, &res); err == nil {
//        cacheList := make([]*model.ScoreTeamAnalysis, 0)
//        distinctM := make(map[int64]bool, 0)
//
//        for _, v := range res.Data.List {
//            if d, err := v.Convert2TeamAnalysis(ctx); err == nil {
//                if _, ok := distinctM[d.TeamID]; !ok {
//                    cacheList = append(cacheList, d)
//                    distinctM[d.TeamID] = true
//                }
//
//                if err := d.InsertUpdate(ctx); err != nil {
//                    fmt.Println("TeamAnalysis InsertUpdate err: ", err)
//                }
//            }
//        }
//
//        if cfg := conf.LoadScoreAnalysisCfg(); cfg != nil && cfg.CacheKey4Team != "" {
//            item := &memcache.Item{Key: cfg.CacheKey4Team, Object: cacheList, Expiration: cfg.Expiration, Flags: memcache.FlagJSON}
//            if err := globalMemcache.Set(ctx, item); err != nil {
//                fmt.Println("ScoreTeamAnalysis >>> globalMemcache.Set: ", cfg.CacheKey4Team, err)
//                // TODO
//            }
//        }
//    }
//}

//func (s *Service) fetchPlayerAnalysis(ctx context.Context, tournamentID int64) {
//    values := genUrlValuesByTournamentID(tournamentID)
//    res := struct {
//        Data    struct {
//            List []*model.ScoreOriginPlayerAnalysis `json:"list"`
//        } `json:"data"`
//    }{}
//
//    if err := s.getScoreData(ctx, scorePathOfBigData4Player, values, &res); err == nil {
//        cacheList := make([]*model.ScorePlayerAnalysis, 0)
//        distinctM := make(map[int64]bool, 0)
//
//        for _, v := range res.Data.List {
//            if d, err := v.Convert2PlayerAnalysis(ctx); err == nil {
//                if _, ok := distinctM[d.PlayerID]; !ok {
//                    cacheList = append(cacheList, d)
//                    distinctM[d.PlayerID] = true
//                }
//
//                if err := d.InsertUpdate(ctx); err != nil {
//                    fmt.Println("PlayerAnalysis InsertUpdate err: ", err)
//                }
//            }
//        }
//
//        if cfg := conf.LoadScoreAnalysisCfg(); cfg != nil && cfg.CacheKey4Player != "" {
//            item := &memcache.Item{Key: cfg.CacheKey4Player, Object: cacheList, Expiration: cfg.Expiration, Flags: memcache.FlagJSON}
//            if err := globalMemcache.Set(ctx, item); err != nil {
//                fmt.Println("ScorePlayAnalysis >>> globalMemcache.Set: ", cfg.CacheKey4Player, err)
//                // TODO
//            }
//        }
//    }
//}

//func (s *Service) fetchHeroAnalysis(ctx context.Context, tournamentID int64) {
//    values := genUrlValuesByTournamentID(tournamentID)
//    res := struct {
//        Data    struct {
//            List []*model.ScoreOriginHeroAnalysis `json:"list"`
//        } `json:"data"`
//    }{}
//
//    if err := s.getScoreData(ctx, scorePathOfBigData4Hero, values, &res); err == nil {
//        cacheList := make([]*model.ScoreHeroAnalysis, 0)
//        distinctM := make(map[int64]bool, 0)
//
//        for _, v := range res.Data.List {
//            if d, err := v.Convert2HeroAnalysis(ctx); err == nil {
//                if _, ok := distinctM[d.HeroID]; !ok {
//                    cacheList = append(cacheList, d)
//                    distinctM[d.HeroID] = true
//                }
//
//                if err := d.InsertUpdate(ctx); err != nil {
//                   fmt.Println("HeroAnalysis InsertUpdate err: ", err)
//                }
//            }
//        }
//
//        if cfg := conf.LoadScoreAnalysisCfg(); cfg != nil && cfg.CacheKey4Hero != "" {
//            item := &memcache.Item{Key: cfg.CacheKey4Hero, Object: cacheList, Expiration: cfg.Expiration, Flags: memcache.FlagJSON}
//            if err := globalMemcache.Set(ctx, item); err != nil {
//                fmt.Println("ScoreHeroAnalysis >>> globalMemcache.Set: ", cfg.CacheKey4Hero, err)
//                // TODO
//            }
//        }
//    }
//}
