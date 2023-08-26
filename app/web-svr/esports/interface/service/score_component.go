package service

import (
	"context"
	"sort"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	errGroupV2 "go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

var (
	_emptyPlayerKdaRank = make([]*model.PlayerDataKdaRank, 0)
	_emptyPlayerMvpRank = make([]*model.PlayerDataMvpRank, 0)
	_emptyHero2Rank     = make([]*model.LolDataHero2, 0)

	seasonLolDataHero2Map  map[int64][]*model.LolDataHero2
	seasonLolDataPlayerMap map[int64][]*model.LolPlayer
	seasonLolDataTeamMap   map[int64][]*model.LolTeam
)

func init() {
	seasonLolDataPlayerMap = make(map[int64][]*model.LolPlayer)
	seasonLolDataTeamMap = make(map[int64][]*model.LolTeam)
	seasonLolDataHero2Map = make(map[int64][]*model.LolDataHero2)
}

func (s *Service) batchSetLolDataMemoryBySeason(ctx context.Context, season *model.ComponentSeason) {
	tmpPlayerMap := make(map[int64][]*model.LolPlayer)
	tmpHero2Map := make(map[int64][]*model.LolDataHero2)
	groupV2 := errGroupV2.WithContext(ctx)
	groupV2.Go(func(ctx context.Context) error {
		// lol player .
		tmpPlayer2List, err := s.FetchLolDataPlayer(ctx, season.LeidaSid)
		if err != nil {
			log.Errorc(ctx, "watchLolDataByGoingSeason s.FetchLolDataPlayer() leidaSID(%+v) error(%+v)", season.LeidaSid, err)
			return err
		}
		tmpPlayerMap[season.LeidaSid] = tmpPlayer2List
		return nil
	})
	groupV2.Go(func(ctx context.Context) error {
		// lol hero .
		tmpHero2List, err := s.FetchLolDataHero2(ctx, season.LeidaSid)
		if err != nil {
			log.Errorc(ctx, "watchLolDataByGoingSeason s.dao.FetchLolDataHero2() leidaSID(%+v) error(%+v)", season.LeidaSid, err)
			return err
		}
		tmpHero2Map[season.LeidaSid] = tmpHero2List
		return nil
	})
	if err := groupV2.Wait(); err != nil {
		log.Errorc(ctx, "watchLolDataByGoingSeason errGroup (%+v)", err)
		return
	}
	seasonLolDataPlayerMap = tmpPlayerMap
	seasonLolDataHero2Map = tmpHero2Map
}

func (s *Service) watchLolDataByGoingSeason(ctx context.Context) {
	groupV2 := errGroupV2.WithContext(ctx)
	groupV2.Go(func(ctx context.Context) (err error) {
		s.watchLolDataOtherByGoingSeason(ctx)
		return nil
	})
	groupV2.Go(func(ctx context.Context) (err error) {
		s.watchLolDataTeamByScoreSeason(ctx)
		return nil
	})
	if err := groupV2.Wait(); err != nil {
		log.Errorc(ctx, "RefreshLolDataCache errGroup (%+v)", err)
	}
}
func (s *Service) watchLolDataOtherByGoingSeason(ctx context.Context) {
	tmpCfg := conf.LoadSeasonContestComponentWatch()
	if tmpCfg == nil || !tmpCfg.CanWatch {
		return
	}
	for _, season := range goingSeasonsListGlobal {
		if season.LeidaSid == 0 {
			continue
		}
		s.batchSetLolDataMemoryBySeason(ctx, season)
	}
}

func (s *Service) watchLolDataTeamByScoreSeason(ctx context.Context) {
	lolGameID := s.mapGameDb[_lolType]
	tmpTeamMap := make(map[int64][]*model.LolTeam)
	if lolSeasons, err := s.dao.GameSeason(context.Background(), lolGameID); err != nil {
		log.Error("watchLolDataTeamByGoingSeason s.dao.GameSeason() LOL error(%+v)", err)
	} else {
		for _, season := range lolSeasons {
			if season.LeidaSID == 0 {
				continue
			}
			// lol team .
			tmpTeamList, err := s.FetchLolDataTeam(ctx, season.LeidaSID)
			if err != nil {
				log.Errorc(ctx, "watchLolDataTeamByGoingSeason s.FetchLolDataTeam() leidaSID(%+v) error(%+v)", season.LeidaSID, err)
				return
			}
			tmpTeamMap[season.LeidaSID] = tmpTeamList
		}
	}
	seasonLolDataTeamMap = tmpTeamMap
}

func (s *Service) getSeasonInfo(ctx context.Context, seasonID int64) (res *model.ComponentSeason, err error) {
	res = &model.ComponentSeason{}
	// 内存变量中获取.
	for _, season := range goingSeasonsListGlobal {
		if season.ID == seasonID {
			return season, nil
		}
	}
	// 先取cache 最后回源DB.
	var mapSeason map[int64]*model.Season
	if mapSeason, err = s.dao.SeasonListByIDList([]int64{seasonID}); err != nil {
		log.Errorc(ctx, "getSeasonInfo s.dao.SeasonListByIDList() seasonID(%+v) error(%+v)", seasonID, err)
		return
	}
	if season, ok := mapSeason[seasonID]; ok {
		res = &model.ComponentSeason{
			ID:       season.ID,
			LeidaSid: season.LeidaSID,
			Stime:    season.Stime,
			Etime:    season.Etime,
			Title:    season.Title,
		}
	}
	return
}

// PlayerKdaRank  player kda rank.
func (s *Service) PlayerKdaRank(ctx context.Context, param *model.ParamKdaRank) (res []*model.PlayerDataKdaRank, err error) {
	seasson, err := s.getSeasonInfo(ctx, param.SeasonID)
	if err != nil {
		log.Errorc(ctx, "PlayerKdaRank s.getSeasonInfo() sid(%d) error(%+v)", param.SeasonID, err)
		return
	}
	if seasson == nil {
		res = _emptyPlayerKdaRank
		return
	}
	if res, err = s.getPlayerKdaRank(ctx, seasson.LeidaSid, param); err != nil {
		log.Errorc(ctx, "PlayerKdaRank s.getPlayerKdaRank() LeidaSid(%d) error(%+v)", seasson.LeidaSid, err)
		return
	}
	return
}

func (s *Service) getPlayerKdaRank(ctx context.Context, leidaSID int64, param *model.ParamKdaRank) (res []*model.PlayerDataKdaRank, err error) {
	res = _emptyPlayerKdaRank
	if leidaSID == 0 {
		return
	}
	positionPlayer := make(map[int64]struct{}, 5)
	tmpPlayers, ok := seasonLolDataPlayerMap[leidaSID]
	if !ok {
		if tmpPlayers, err = s.FetchLolDataPlayer(ctx, leidaSID); err != nil {
			log.Errorc(ctx, "getPlayerKdaRank s.FetchLolDataPlayer() leidaSID(%d) error(%+v)", leidaSID, err)
			return
		}
	}
	if len(tmpPlayers) == 0 {
		res = _emptyPlayerKdaRank
		return
	}
	s.lolPlayerSort(_sortDESC, param.SortType, tmpPlayers)
	for _, playerData := range tmpPlayers {
		if _, ok := positionPlayer[playerData.PositionID]; !ok {
			positionPlayer[playerData.PositionID] = struct{}{}
			res = append(res, &model.PlayerDataKdaRank{
				PlayerDataRank: &model.PlayerDataRank{
					ID:         playerData.PlayerID,
					PlayerID:   playerData.PlayerID,
					PlayerName: playerData.Name,
					ImageURL:   playerData.ImageURL,
					TeamID:     playerData.TeamID,
					TeamName:   playerData.TeamAcronym,
					PositionID: playerData.PositionID,
					Position:   playerData.Position,
				},
				Kda: playerData.KDA,
			})
		}
	}
	if len(res) == 0 {
		res = _emptyPlayerKdaRank
		return
	}
	// 按位置从小到大排序.
	sort.SliceStable(res, func(i, j int) bool {
		return res[i].PositionID < res[j].PositionID
	})
	return
}

// PlayerMvpRank  player mvp rank.
func (s *Service) PlayerMvpRank(ctx context.Context, param *model.ParamMvpRank) (res []*model.PlayerDataMvpRank, err error) {
	season, err := s.getSeasonInfo(ctx, param.SeasonID)
	if err != nil {
		log.Errorc(ctx, "PlayerMvpRank s.getSeasonInfo() sid(%d) error(%+v)", param.SeasonID, err)
		return
	}
	if season == nil {
		res = _emptyPlayerMvpRank
		return
	}
	if res, err = s.getPlayerMvpRank(ctx, season.LeidaSid, param); err != nil {
		log.Errorc(ctx, "PlayerMvpRank s.getPlayerMvpRank() param(%+v) error(%+v)", param, err)
		return
	}
	return
}

func (s *Service) getPlayerMvpRank(ctx context.Context, leidaSID int64, param *model.ParamMvpRank) (res []*model.PlayerDataMvpRank, err error) {
	res = _emptyPlayerMvpRank
	if leidaSID == 0 {
		return
	}
	tmpPlayers, ok := seasonLolDataPlayerMap[leidaSID]
	if !ok {
		if tmpPlayers, err = s.FetchLolDataPlayer(ctx, leidaSID); err != nil {
			log.Errorc(ctx, "getPlayerMvpRank s.FetchLolDataPlayer() leidaSID(%d) error(%+v)", leidaSID, err)
			return
		}
	}
	if len(tmpPlayers) == 0 {
		res = _emptyPlayerMvpRank
		return
	}
	s.lolPlayerSort(_sortDESC, param.SortType, tmpPlayers)
	for rank, playerData := range tmpPlayers {
		if len(res) == param.Top {
			return
		}
		res = append(res, &model.PlayerDataMvpRank{
			PlayerDataRank: &model.PlayerDataRank{
				ID:         playerData.PlayerID,
				PlayerID:   playerData.PlayerID,
				PlayerName: playerData.Name,
				ImageURL:   playerData.ImageURL,
				TeamID:     playerData.TeamID,
				TeamName:   playerData.TeamAcronym,
				PositionID: playerData.PositionID,
				Position:   playerData.Position,
			},
			Mvp:  playerData.MVP,
			Rank: rank + 1,
		})
	}
	if len(res) == 0 {
		res = _emptyPlayerMvpRank
		return
	}
	return
}

// Hero2Rank  hero2 rank.
func (s *Service) Hero2Rank(ctx context.Context, param *model.ParamHero2Rank) (res []*model.LolDataHero2, err error) {
	res = _emptyHero2Rank
	season, err := s.getSeasonInfo(ctx, param.SeasonID)
	if err != nil {
		log.Errorc(ctx, "Hero2Rank s.getSeasonInfo() sid(%d) error(%+v)", param.SeasonID, err)
		return
	}
	// 内存中取 .
	seasonHero2s, ok := seasonLolDataHero2Map[season.LeidaSid]
	if !ok {
		seasonHero2s, err = s.FetchLolDataHero2(ctx, season.LeidaSid)
		if err != nil {
			log.Errorc(ctx, "Hero2Rank s.FetchLolDataHero2 sid(%d) error(%+v)", param.SeasonID, err)
			return
		}
	}
	sort.SliceStable(seasonHero2s, func(i, j int) bool {
		if seasonHero2s[i].AppearCount != seasonHero2s[j].AppearCount {
			return seasonHero2s[i].AppearCount > seasonHero2s[j].AppearCount
		}
		return seasonHero2s[i].HeroID < seasonHero2s[j].HeroID
	})
	for _, hero := range seasonHero2s {
		tmpHero := new(model.LolDataHero2)
		*tmpHero = *hero
		if len(res) == param.Top {
			return
		}
		if tmpHero.AppearCount > 0 {
			tmpHero.VictoryRate = tool.DecimalFloat(float64(tmpHero.VictoryCount)/float64(tmpHero.AppearCount), 2)
		}
		res = append(res, tmpHero)
	}
	return
}

func (s *Service) FetchLolDataHero2(ctx context.Context, leidaSID int64) (res []*model.LolDataHero2, err error) {
	if leidaSID == 0 {
		return
	}
	if res, err = s.dao.FetchLolDataHero2FromCache(ctx, leidaSID); err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "FetchLolDataHero2 s.dao.FetchLolDataHero2FromCache() leidaSID(%+v) error(%+v)", leidaSID, err)
		return
	}
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		res, err = s.dao.FetchLolDataHero2(ctx, leidaSID)
		if err != nil {
			log.Errorc(ctx, "FetchLolDataHero2 s.dao.FetchLolDataHero2() leidaSID(%+v) error(%+v)", leidaSID, err)
			return
		}
		if e := s.dao.FetchLolDataHero2ToCache(ctx, leidaSID, res, 600); e != nil {
			log.Errorc(ctx, "FetchLolDataHero2 s.dao.FetchLolDataHero2ToCache() leidaSID(%+v) error(%+v)", leidaSID, err)
		}
	}
	return
}

func (s *Service) FetchLolDataPlayer(ctx context.Context, leidaSID int64) (res []*model.LolPlayer, err error) {
	if leidaSID == 0 {
		return
	}
	if res, err = s.dao.FetchLolDataPlayerFromCache(ctx, leidaSID); err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "FetchLolDataPlayer s.dao.FetchLolDataPlayerFromCache() leidaSID(%+v) error(%+v)", leidaSID, err)
		return
	}
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		res, err = s.dao.LolPlayers(ctx, leidaSID)
		if err != nil {
			log.Errorc(ctx, "FetchLolDataPlayer s.dao.LolPlayers() leidaSID(%+v) error(%+v)", leidaSID, err)
			return
		}
		if e := s.dao.FetchLolDataPlayerToCache(ctx, leidaSID, res, 600); e != nil {
			log.Errorc(ctx, "FetchLolDataPlayer s.dao.FetchLolDataPlayerToCache() leidaSID(%+v) error(%+v)", leidaSID, err)
		}
	}
	return
}

func (s *Service) FetchLolDataTeam(ctx context.Context, leidaSID int64) (res []*model.LolTeam, err error) {
	if leidaSID == 0 {
		return
	}
	if res, err = s.dao.FetchLolDataTeamFromCache(ctx, leidaSID); err != nil && err != memcache.ErrNotFound {
		log.Errorc(ctx, "FetchLolDataTeam s.dao.FetchLolDataTeamFromCache() leidaSID(%+v) error(%+v)", leidaSID, err)
		return
	}
	if err == nil {
		return
	}
	if err == memcache.ErrNotFound {
		res, err = s.dao.LolTeams(ctx, leidaSID)
		if err != nil {
			log.Errorc(ctx, "FetchLolDataTeam s.dao.LolTeams() leidaSID(%+v) error(%+v)", leidaSID, err)
			return
		}
		if e := s.dao.FetchLolDataTeamToCache(ctx, leidaSID, res, 600); e != nil {
			log.Errorc(ctx, "FetchLolDataTeam s.dao.FetchLolDataTeamToCache() leidaSID(%+v) error(%+v)", leidaSID, err)
		}
	}
	return
}

func (s *Service) RefreshLolDataCache(ctx context.Context, leidaSid int64) (err error) {
	groupV2 := errGroupV2.WithContext(ctx)
	groupV2.Go(func(ctx context.Context) error {
		return s.dao.DeleteLolDataPlayerCache(ctx, leidaSid)
	})
	groupV2.Go(func(ctx context.Context) error {
		return s.dao.DeleteLolDataTeamCache(ctx, leidaSid)
	})
	groupV2.Go(func(ctx context.Context) error {
		return s.dao.DeleteLolDataHero2Cache(ctx, leidaSid)
	})
	if err = groupV2.Wait(); err != nil {
		log.Errorc(ctx, "RefreshLolDataCache errGroup leidaSid(%d) (%+v)", leidaSid, err)
		return
	}
	return
}
