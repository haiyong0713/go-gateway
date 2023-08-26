package service

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/esports/interface/conf"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	egV2 "go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

const _seasonStartOK = 1

func (s *Service) GetTeamsInSeason(ctx context.Context, seasonId int64) (teams []*model.TeamInSeason, err error) {
	//Step 1: get from local memory.
	{
		var exist bool
		teams, exist = s.teamsInSeasonMap[seasonId]
		if exist {
			tool.Metric4MemoryCache.WithLabelValues("hit_team_in_season_memory").Inc()
			return teams, nil
		}
		tool.Metric4MemoryCache.WithLabelValues("miss_team_in_season_memory").Inc()
	}

	//Step 2: get from redis.
	{
		teams, err = s.dao.GetTeamsInSeasonFromCache(ctx, seasonId)
		if err == nil {
			tool.Metric4MemoryCache.WithLabelValues("hit_team_in_season_redis").Inc()
			return teams, nil
		}
		tool.Metric4MemoryCache.WithLabelValues("miss_team_in_season_redis").Inc()
		//try to set cache when get from db success
		defer func() {
			if err == nil {
				if setCacheErr := s.dao.AddTeamsInSeasonToCache(context.Background(), seasonId, teams); setCacheErr != nil {
					log.Errorc(context.Background(), "teams_in_season season_id(%v) cache update error: %v", seasonId, setCacheErr)
				}
			}
		}()
	}
	tool.AddDBBackSourceMetricsByKeyList("teams_in_season", []int64{seasonId})
	//Step 3: get from db
	{
		//query
		teamsInSeasonMap := make(map[int64] /*seasonId*/ []*model.TeamInSeason, 0)
		teamsInSeasonMap, err = s.dao.GetTeamsInSeasonFromDB(context.Background(), []int64{seasonId})
		if err != nil {
			log.Errorc(context.Background(), "teams_in_season season_id(%v) query in db error: %v", seasonId, err)
			tool.AddDBErrMetricsByKeyList("teams_in_season", []int64{seasonId})
			return nil, err
		}
		if len(teamsInSeasonMap) == 0 {
			teams = make([]*model.TeamInSeason, 0)
			return
		}
		teams = teamsInSeasonMap[seasonId]
		if teams == nil {
			err = fmt.Errorf("teams_in_season query success but not record found")
			log.Errorc(context.Background(), "%s", err)
		}

	}
	return
}

// AsyncUpdateOngoingSeasonTeamInMemoryCache: update all ongoing season's team info in memory
func (s *Service) AsyncUpdateOngoingSeasonTeamInMemoryCache() {
	ctx := context.Background()
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ongoingSeasonId, err := s.dao.GetOngoingSeasonIDFromDB(ctx)
			if err != nil {
				continue
			}
			if len(ongoingSeasonId) == 0 {
				continue
			}
			tmpTeamsInSeasonMap, err := s.dao.GetTeamsInSeasonFromDB(ctx, ongoingSeasonId)
			if len(tmpTeamsInSeasonMap) == 0 { //maybe db is down, skip update
				log.Errorc(ctx, "ongoingSeasonIds is not empty, but got empty teams from db, "+
					"maybe db is down, please check manually")
				tool.AddDBNoResultMetricsByKeyList("watch_team_in_season", []int64{0})
				return
			}
			s.teamsInSeasonMap = tmpTeamsInSeasonMap
			log.Infoc(ctx, "update teams_in_season for ongoing seasons[%v] success", ongoingSeasonId)
		case <-ctx.Done():
			return
		}
	}
}

// AsyncUpdateMatchSeasonInMemoryCache.
func (s *Service) AsyncUpdateMatchSeasonInMemoryCache(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tmpMatchSeasonsMap, err := s.dao.MatchSeasonsFromDB(ctx, s.c.GoingMatchs.MatchIDs)
			if err != nil {
				log.Errorc(ctx, "AsyncUpdateMatchSeasonInMemoryCache s.dao.GetTeamsInSeasonFromDB() error(%+v)", err)
				continue
			}
			if len(tmpMatchSeasonsMap) == 0 {
				log.Warnc(ctx, "AsyncUpdateMatchSeasonInMemoryCache tmpMatchSeasonsMap empty from db")
				tool.AddDBNoResultMetricsByKeyList("watch_match_season", []int64{0})
				return
			}
			s.matchSeasonMap = tmpMatchSeasonsMap
			log.Infoc(ctx, "update teams_in_season for ongoing seasons[%v] success", s.c.GoingMatchs.MatchIDs)
		case <-ctx.Done():
			return
		}
	}
}

// AsyncUpdateGoingSeasonsInfoMemoryCache.
func (s *Service) AsyncUpdateGoingSeasonsInfoMemoryCache(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			tmpSeasonsMap, err := s.dao.RawFetchSeasonsInfoMap(ctx, s.c.GoingMatchs.GoingSeasons)
			if err != nil {
				log.Errorc(ctx, "AsyncUpdateGoingSeasonsInfoMemoryCache s.dao.GetTeamsInSeasonFromDB() error(%+v)", err)
				continue
			}
			s.seasonMap = tmpSeasonsMap
			log.Infoc(ctx, "AsyncUpdateGoingSeasonsInfoMemoryCache update going seasons count(%d) success", len(tmpSeasonsMap))
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) rebuildReserveMap(ctx context.Context) {
	res := make(map[int64]int64, 0)
	for strSid, reserveID := range s.c.GoingMatchs.ReserveMap {
		if intSid, err := strconv.ParseInt(strSid, 10, 64); err != nil {
			panic(err)
		} else {
			res[reserveID] = intSid
		}
	}
	s.reserveMap = res
}

func (s *Service) MatchSeasonsInfo(ctx context.Context, mid int64, p *model.ParamMatchSeasons) (res []*model.MatchSeason, err error) {
	var (
		seasonsMap   map[int64]*model.MatchSeason
		seasonSubMap map[int64]bool
	)
	eg := egV2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if seasonsMap, err = s.fetchMatchSeasons(ctx, p.MatchID); err != nil {
			log.Errorc(ctx, "MatchSeasonsInfo  s.fetchMatchSeasons() param(%+v) error(%+v)", p, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if mid > 0 {
			if seasonSubMap, err = s.seasonSubMap(ctx, mid, p.SeasonIDs); err != nil {
				log.Errorc(ctx, "MatchSeasonsInfo  s.seasonSubMap() param(%+v) error(%+v)", p, err)
				return err
			}
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	res = rebuildMatchSeasonSubscribe(p.SeasonIDs, seasonsMap, seasonSubMap)
	return
}

func rebuildMatchSeasonSubscribe(SeasonIDs []int64, seasonsMap map[int64]*model.MatchSeason, seasonSubMap map[int64]bool) (res []*model.MatchSeason) {
	res = make([]*model.MatchSeason, 0)
	nowTime := time.Now().Unix()
	for _, seasonID := range SeasonIDs {
		tmpSeason := &model.MatchSeason{}
		if season, ok := seasonsMap[seasonID]; ok {
			*tmpSeason = *season
		} else {
			continue // 赛事下不存在该赛季.
		}
		if isSub, ok := seasonSubMap[seasonID]; ok {
			tmpSeason.IsSub = isSub
		}
		if nowTime >= tmpSeason.Stime {
			tmpSeason.StartSeason = _seasonStartOK
		}
		res = append(res, tmpSeason)
	}
	return
}

func (s *Service) fetchMatchSeasons(ctx context.Context, matchID int64) (res map[int64]*model.MatchSeason, err error) {
	var (
		matchSeasons []*model.MatchSeason
		exist        bool
	)
	res = make(map[int64]*model.MatchSeason, 0)
	matchSeasons, exist = s.matchSeasonMap[matchID]
	if !exist { // 不存在.
		if matchSeasons, err = s.dao.FetchSeasonsByMatchId(ctx, matchID); err != nil {
			log.Errorc(ctx, "MatchSeasonsInfo fetchMatchSeasons s.dao.SeasonsByMatchIdFromDB() matchID(%d) error(%+v)", matchID, err)
			return
		}
		tool.Metric4MemoryCache.WithLabelValues([]string{"miss_match_seasons"}...).Inc()
	} else {
		tool.Metric4MemoryCache.WithLabelValues([]string{"hit_match_seasons"}...).Inc()
	}
	tmpRes := make(map[int64]*model.MatchSeason, len(matchSeasons))
	for _, season := range matchSeasons {
		tmpSeason := new(model.MatchSeason)
		*tmpSeason = *season
		tmpRes[tmpSeason.SeasonID] = tmpSeason
	}
	if len(tmpRes) > 0 {
		res = tmpRes
	}
	return
}

func (s *Service) seasonSubMap(ctx context.Context, mid int64, seasonIDs []int64) (resMap map[int64]bool, err error) {
	resMap = make(map[int64]bool, 0)
	subSeasonIDs := s.haveSubSeasonIDs(seasonIDs)
	if len(subSeasonIDs) > 0 {
		reserveReply, e := s.actClient.ReserveFollowings(ctx, &api.ReserveFollowingsReq{Mid: mid, Sids: subSeasonIDs})
		if e != nil {
			log.Errorc(ctx, "MatchSeasonsInfo seasonSubMap s.haveSubSeasonIDs() mid(%d) seasonIDs(%+v) error(%+v)", mid, seasonIDs, e)
			return
		}
		for reserveID, reserve := range reserveReply.List {
			if seasonID, ok := s.reserveMap[reserveID]; ok {
				resMap[seasonID] = reserve.IsFollow
			}
		}
	}
	return
}

func (s *Service) haveSubSeasonIDs(seasonIDs []int64) (res []int64) {
	res = make([]int64, 0)
	if len(s.c.GoingMatchs.ReserveMap) == 0 {
		return
	}
	for _, sid := range seasonIDs {
		if reserveID, ok := s.c.GoingMatchs.ReserveMap[strconv.FormatInt(sid, 10)]; ok {
			res = append(res, reserveID)
		}
	}
	return
}

func (s *Service) DelMatchSeasonsCache(ctx context.Context, matchID int64) (err error) {
	if err = retry.WithAttempts(ctx, "match_seasons_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return s.dao.DelCacheSeasonsByMatchId(ctx, matchID)
	}); err != nil {
		log.Errorc(ctx, "MatchSeasonsInfo DelMatchSeasonsCache s.dao.DelCacheSeasonsByMatchId() matchID(%d) error(%+v)", matchID, err)
		return err
	}
	return
}

func (s *Service) DelSeasonInfoCache(ctx context.Context, sid int64) (err error) {
	if err = retry.WithAttempts(ctx, "season_info_del_cache", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		return s.dao.DelCacheSeasonInfoByID(ctx, sid)
	}); err != nil {
		log.Errorc(ctx, "DelSeasonInfoCache s.dao.DelCacheSeasonInfoByID() sid(%d) error(%+v)", sid, err)
		return err
	}
	return
}

func (s *Service) SeasonTeamsInfo(ctx context.Context, mid int64, p *model.ParamSeasonTeams) (res []*model.SeasonTeam, err error) {
	var (
		tmpTeams, newTeams []*model.TeamInSeason
		teamIDs            []int64
		teamSubMap         map[int64]bool
	)
	// 赛季下的战队
	if tmpTeams, err = s.GetTeamsInSeason(ctx, p.SeasonID); err != nil {
		log.Errorc(ctx, "SeasonTeamsInfo s.GetTeamsInSeason() sid(%d) error(%+v)", p.SeasonID, err)
		return
	}
	if len(tmpTeams) == 0 {
		res = make([]*model.SeasonTeam, 0)
		return
	}
	for _, team := range tmpTeams {
		if tool.Int64InSlice(team.TeamId, conf.Conf.SeriesIgnoreTeamsIDList) {
			continue
		}
		tmpTeam := new(model.TeamInSeason)
		*tmpTeam = *team
		newTeams = append(newTeams, tmpTeam)
		teamIDs = append(teamIDs, team.TeamId)
	}
	req := &model.AutoSubRequest{
		SeasonID:   p.SeasonID,
		TeamIDList: teamIDs,
	}
	// 是否订阅战队.
	if mid > 0 {
		if teamSubMap, err = s.fetchAutoSubStatus(ctx, mid, req); err != nil {
			log.Errorc(ctx, "SeasonTeamsInfo s.fetchAutoSubStatus() seasonID(%d) req(%+v) error(%+v)", p.SeasonID, req, err)
			return
		}
	}
	for _, team := range newTeams {
		var isSub bool
		if len(teamSubMap) > 0 {
			isSub = teamSubMap[team.TeamId]
		}
		tmpTeam := &model.SeasonTeam{
			TeamInSeason: team,
			IsSub:        isSub,
		}
		res = append(res, tmpTeam)
	}
	return
}

func (s *Service) fetchBatchSeasons(ctx context.Context, sids []int64) (res map[int64]*model.MatchSeason, err error) {
	var (
		missList      []int64
		missSeasonMap map[int64]*model.MatchSeason
	)
	if res, missList, err = s.seasonsInfoFromMemory(ctx, sids); err != nil {
		log.Errorc(ctx, "BatchSeasonsInfo fetchBatchSeasons s.dao.SeasonsByMatchIdFromDB() sids(%d) error(%+v)", sids, err)
		return
	}
	tool.Metric4MemoryCache.WithLabelValues([]string{"hit_batch_seasons"}...).Add(float64(len(res)))
	if len(missList) == 0 {
		return
	}
	tool.Metric4MemoryCache.WithLabelValues([]string{"miss_batch_seasons"}...).Add(float64(len(missList)))
	if missSeasonMap, err = s.dao.FetchSeasonsInfoMap(ctx, sids); err != nil {
		log.Errorc(ctx, "BatchSeasonsInfo fetchBatchSeasons s.dao.SeasonsByMatchIdFromDB() noSids(%d) error(%+v)", missList, err)
		return
	}
	for sid, season := range missSeasonMap {
		res[sid] = season
	}
	return
}

func (s *Service) seasonsInfoFromMemory(ctx context.Context, sids []int64) (haveSeason map[int64]*model.MatchSeason, noSids []int64, err error) {
	haveSeason = make(map[int64]*model.MatchSeason, 0)
	for _, sid := range sids {
		if season, ok := s.seasonMap[sid]; ok {
			haveSeason[sid] = season
		} else {
			noSids = append(noSids, sid)
		}
	}
	return
}

func (s *Service) BatchSeasonsInfo(ctx context.Context, mid int64, p *model.ParamSeasonsInfo) (res []*model.MatchSeason, err error) {
	var (
		seasonsMap   map[int64]*model.MatchSeason
		seasonSubMap map[int64]bool
	)
	eg := egV2.WithContext(ctx)
	eg.Go(func(ctx context.Context) (err error) {
		if seasonsMap, err = s.fetchBatchSeasons(ctx, p.SeasonIDs); err != nil {
			log.Errorc(ctx, "BatchSeasonsInfo  s.fetchMatchSeasons() param(%+v) error(%+v)", p, err)
			return err
		}
		return nil
	})
	eg.Go(func(ctx context.Context) (err error) {
		if mid > 0 {
			if seasonSubMap, err = s.seasonSubMap(ctx, mid, p.SeasonIDs); err != nil {
				log.Errorc(ctx, "BatchSeasonsInfo  s.seasonSubMap() param(%+v) error(%+v)", p, err)
				return err
			}
		}
		return nil
	})
	if err = eg.Wait(); err != nil {
		return
	}
	tmpSeasonsMap := deepCopySeasonsMap(seasonsMap)
	res = rebuildMatchSeasonSubscribe(p.SeasonIDs, tmpSeasonsMap, seasonSubMap)
	return
}

func deepCopySeasonsMap(seasonsAll map[int64]*model.MatchSeason) map[int64]*model.MatchSeason {
	tmpRes := make(map[int64]*model.MatchSeason, 0)
	for sid, season := range seasonsAll {
		tmpSeason := new(model.MatchSeason)
		*tmpSeason = *season
		tmpRes[sid] = tmpSeason
	}
	return tmpRes
}
