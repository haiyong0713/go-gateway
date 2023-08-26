package service

import (
	"context"
	"fmt"
	"time"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/model"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
)

var (
	mapScoreRankingData  = map[string]*model.RankingData{}
	s10ContestMatchIDMap = map[int64]int64{}
	s10ScoreTeamRegionID = map[string]int{}
	currentRoundID       string
	s10InterventionData  = &model.S10RankingInterventionData{}
)

func (s *Service) tasksProgress(ctx context.Context, mid int64) ([]*api.TaskProgress, error) {
	reply, err := s.actClient.TasksProgress(ctx, &api.TasksProgressReq{Mid: mid})
	if err != nil {
		log.Errorc(ctx, "s.actClient.TasksProgress(%d) error(%v)", mid, err)
		return nil, err
	}
	return reply.Tasks, nil
}

func (s *Service) totalPoints(ctx context.Context, mid int64) (*model.UsesrPoint, error) {
	res := new(model.UsesrPoint)
	if mid <= 0 {
		return res, nil
	}
	res.IsLogin = true
	reply, err := s.actClient.TotalPoints(ctx, &api.TotalPointsdReq{Mid: mid})
	if err != nil {
		log.Errorc(ctx, "s.actClient.TotalPoints(%d) error(%v)", mid, err)
		return res, err
	}
	res.Points = reply.Total
	return res, nil
}

func (s *Service) TasksAndPoints(ctx context.Context, mid int64) (*model.Tasks, error) {
	var (
		err          error
		taskProgress []*api.TaskProgress
		totalPoint   *model.UsesrPoint
		banner       = s.LiveBanner()
		pointsAct    = s.PointsActBanner()
		pointsActWeb = s.PointsActWebBanner()
		webBanner    = s.LiveWebBanner()
		wg           = errgroup.WithContext(ctx)
	)
	wg.Go(func(ctx context.Context) (err error) {
		taskProgress, err = s.tasksProgress(ctx, mid)
		return
	})
	wg.Go(func(ctx context.Context) (err error) {
		totalPoint, err = s.totalPoints(ctx, mid)
		return
	})
	if err = wg.Wait(); err != nil {
		log.Errorc(ctx, "wg.Wait() error(%v)", err)
		code := xecode.Cause(err).Code()
		if code != 75505 && code != 75504 && code != 75787 && code != 75788 {
			err = ecode.ActivityTasksProgressGetFail
		}
		//if err != ecode.ActivityAppstoreNotStart && err != ecode.ActivityAppstoreEnd && err != ecode.ActivityPointGetFail && err != ecode.ActivityTasksProgressGetFail {
		//	err = ecode.ActivityTasksProgressGetFail
		//}
	}
	return &model.Tasks{
		User:          totalPoint,
		TasksProgress: taskProgress,
		Banner:        banner,
		PointsAct:     pointsAct,
		WebBanner:     webBanner,
		PointsActWeb:  pointsActWeb,
		SeasonID:      conf.LoadSeasonContestWatch().SeasonID,
	}, err
}

func (s *Service) S10RankingDataWatch() {
	for range time.Tick(time.Duration(s.c.RankingDataWatch.WatchDuration)) {
		s.S10RankingDataUpdate()
	}
}

func (s *Service) S10RankingDataUpdate() {
	ctx := context.Background()
	tmpMap := map[string]*model.RankingData{}
	var roundList []string
	if err := component.GlobalMemcached.Get(ctx, s.c.RankingDataWatch.RoundIDListCacheKey).Scan(&roundList); err != nil {
		log.Errorc(ctx, "S10RankingDataUpdate component.GlobalMemcached.Get(%s) error(%v)", s.c.RankingDataWatch.RoundIDListCacheKey, err)
		return
	}
	for _, roundID := range roundList {
		tmpOne := new(model.RankingData)
		if err := component.GlobalMemcached.Get(ctx, fmt.Sprint(s.c.RankingDataWatch.RoundDataCacheKeyPre, roundID)).Scan(&tmpOne); err != nil {
			log.Errorc(ctx, "S10RankingDataUpdate component.GlobalMemcached.Get(%s) error(%v)", fmt.Sprint(s.c.RankingDataWatch.RoundDataCacheKeyPre, roundID), err)
			return
		}
		tmpMap[roundID] = tmpOne
	}
	var tmpRoundID string
	if err := component.GlobalMemcached.Get(ctx, s.c.RankingDataWatch.CurrentRoundIDCacheKey).Scan(&tmpRoundID); err != nil {
		log.Errorc(ctx, "S10RankingDataUpdate component.GlobalMemcached.Get(%s) error(%v)", s.c.RankingDataWatch.CurrentRoundIDCacheKey, err)
		return
	}
	currentRoundID = tmpRoundID
	mapScoreRankingData = tmpMap
	tmpInterventionData := new(model.S10RankingInterventionData)
	if err := component.GlobalMemcached.Get(ctx, s.c.RankingDataWatch.InterventionCacheKey).Scan(&tmpInterventionData); err != nil {
		log.Errorc(ctx, "S10RankingDataUpdate: globalMemcache.Get(%s) err[%v]", s.c.RankingDataWatch.InterventionCacheKey, err)
	} else {
		s10InterventionData = tmpInterventionData
	}
	idMap := make(map[int64]int64)
	if err := component.GlobalMemcached.Get(ctx, s.c.SeasonContestWatch.ContestMatchIDMapCacheKey).Scan(&idMap); err != nil {
		log.Errorc(ctx, "S10RankingDataUpdate: globalMemcache.Get(%s) err[%v]", s.c.SeasonContestWatch.ContestMatchIDMapCacheKey, err)
		return
	}
	s10ContestMatchIDMap = idMap
	teamMap := make(map[string]*model.Team2Tab)
	regionIDMap := make(map[string]int)
	if err := component.GlobalMemcached.Get(ctx, s.c.SeasonContestWatch.TeamScoreMapCacheKey).Scan(&teamMap); err != nil {
		log.Errorc(ctx, "S10RankingDataUpdate: globalMemcache.Get(%s) err[%v]", s.c.SeasonContestWatch.TeamScoreMapCacheKey, err)
		return
	}
	for id, team := range teamMap {
		regionIDMap[id] = team.RegionID
	}
	s10ScoreTeamRegionID = regionIDMap
}

func (s *Service) S10RankingData(ctx context.Context, roundID string, needPrevious bool, from string, eliminate int) (r *model.RankingDataRet, err error) {
	if roundID == "" {
		roundID = currentRoundID
	}
	var data *model.RankingData
	var ok bool
	if data, ok = mapScoreRankingData[roundID]; !ok {
		return
	}
	r = new(model.RankingDataRet)
	r.RankingData = data
	r.Matches = s10ContestMatchIDMap
	r.TeamRegion = s10ScoreTeamRegionID
	r.RoundID = roundID
	r.EliminateNum = s10InterventionData.EliminateNum
	r.PromoteNum = s10InterventionData.PromoteNum
	r.FinalEliminateNum = s10InterventionData.FinalEliminateNum
	r.FinalPromoteNum = s10InterventionData.FinalPromoteNum
	// 聚合页入围赛数据补充
	if needPrevious && s10InterventionData != nil {
		previousRoundID := s10InterventionData.FinalistRound
		if previousRoundID != "" && previousRoundID != roundID {
			if data, ok = mapScoreRankingData[previousRoundID]; ok {
				r.Previous = &model.RankingDataPrevious{
					RankingData: data,
				}
				// 聚合页需要的入围赛数据兜底图处理
				if len(s10InterventionData.RoundInfo) > 0 {
					for _, round := range s10InterventionData.RoundInfo {
						if round.RoundID == previousRoundID {
							r.Previous.Picture = round.WebPic
							break
						}
					}
				}
			}
		}
	}
	// 当前阶段描述文案计算
	if from == "ugc-tab" || from == "live-tab" {
		if r.RankingData.Stage == 1 {
			r.Description = s.c.RankingDataWatch.Description.Finalist
		} else {
			r.Description = s.c.RankingDataWatch.Description.Final
		}
	} else {
		if r.RankingData.Stage == 1 {
			if from == "live-web" {
				// 直播web入围赛可以手动切换阶段，需要特殊处理判断文案
				if eliminate == 1 {
					r.Description = s.c.RankingDataWatch.Description.FinalistEliminate
				} else {
					r.Description = s.c.RankingDataWatch.Description.FinalistPoint
				}
			} else {
				if len(r.RankingData.Tree) > 0 {
					r.Description = s.c.RankingDataWatch.Description.FinalistEliminate
				} else {
					r.Description = s.c.RankingDataWatch.Description.FinalistPoint
				}
			}
		} else {
			if len(r.RankingData.Tree) > 0 {
				r.Description = s.c.RankingDataWatch.Description.FinalEliminate
			} else {
				r.Description = s.c.RankingDataWatch.Description.FinalPoint
			}
		}
	}
	// 积分兜底图处理
	if len(s10InterventionData.RoundInfo) > 0 {
		for _, round := range s10InterventionData.RoundInfo {
			if round.RoundID == roundID {
				if from == "ugc-tab" || from == "live-tab" {
					r.Picture = round.H5Pic
				} else {
					r.Picture = round.WebPic
				}
				break
			}
		}
	}
	return
}
