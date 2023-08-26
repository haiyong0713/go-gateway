package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/api"
	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/dao/match_component"
	"go-gateway/app/web-svr/esports/interface/model"
)

const (
	cacheKey4GuessListBySeasonID = "season:guess:list:%v:%v:%v"
)

func cacheKey4UserSeasonGuessList(ctx context.Context, mid, seasonID int64) (key string, err error) {
	var version int64
	if d, ok := HotSeasonInMemory[seasonID]; ok {
		version = d
	} else {
		version, err = match_component.FetchSeasonGuessVersionBySeasonID(ctx, seasonID)
	}

	if err != nil {
		return
	}

	key = fmt.Sprintf(cacheKey4GuessListBySeasonID, mid, seasonID, version)

	return
}

func (s *Service) FetchSeasonIDByMatchID(ctx context.Context, matchID int64) (seasonID int64, err error) {
	if d, ok := HotMatch2SeasonInmemory[matchID]; ok {
		seasonID = d

		return
	}

	tmpMatchMap, tmpErr := s.dao.ContestListByIDList(ctx, []int64{matchID})
	if tmpErr != nil {
		err = tmpErr

		return
	}

	if d, ok := tmpMatchMap[matchID]; ok {
		seasonID = d.Sid
	}

	return
}

func (s *Service) FetchUserSeasonGuessList(ctx context.Context, param *model.UserSeasonGuessReq) (resp *model.UserSeasonGuessResp, err error) {
	resp = new(model.UserSeasonGuessResp)
	{
		resp.Data = make([]*model.MatchGuess, 0)
	}
	list := make([]*model.MatchGuess, 0)
	list, err = s.FetchGuessListBySeasonID(ctx, param.MID, param.SeasonID)
	if err != nil {
		return
	}

	resp.PageStructure = model.NewPageStructure(param.PageSize, param.PageNum, int64(len(list)))
	if startIndex, endIndex, ok := resp.PageStructure.CalculateStartAndEndIndex(); ok {
		resp.Data = list[startIndex:endIndex]
		err = rebuildMatchGuessList(resp)
	}

	return
}

func rebuildMatchGuessList(resp *model.UserSeasonGuessResp) (err error) {
	if len(resp.Data) > 0 {
		matchIDList := make([]int64, 0)
		for _, v := range resp.Data {
			matchIDList = append(matchIDList, v.Oid)
		}

		for _, v := range resp.Data {
			if v.Income > 0 {
				v.Income = calculateCoins(v.Income)
			}
		}
	}

	return
}

func calculateCoins(coins float32) (newOne float32) {
	if coins > 0 {
		coinsWithNoDecimal, _ := strconv.ParseFloat(fmt.Sprintf("%.f", coins), 32)
		coinsWithNoDecimalOfNew := float32(coinsWithNoDecimal)
		if coins != coinsWithNoDecimalOfNew {
			tmp, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", coins), 32)
			newOne = float32(tmp)
		} else {
			newOne = coinsWithNoDecimalOfNew
		}
	}

	return
}

func (s *Service) FetchSeasonGuessSummary(ctx context.Context, param *model.GuessParams4V2) (d *model.SeasonGuessSummary, err error) {
	d = new(model.SeasonGuessSummary)
	list := make([]*model.MatchGuess, 0)
	list, err = s.FetchGuessListBySeasonID(ctx, param.MID, param.SeasonID)
	if err == nil {
		d.Total = int64(len(list))
		for _, v := range list {
			if v.Income > 0 {
				d.Wins++
				d.Coins = d.Coins + v.Income
			}
		}

		if d.Coins > 0 {
			d.Coins = calculateCoins(d.Coins)
		}
	}

	return
}

func (s *Service) FetchGuessListBySeasonID(ctx context.Context, mid, seasonID int64) (list []*model.MatchGuess, err error) {
	list, err = FetchGuessListBySeasonIDFromCache(ctx, mid, seasonID)
	if err != nil && err != memcache.ErrNotFound {
		return
	}

	if err == nil {
		return
	}

	if err == memcache.ErrNotFound {
		list, err = s.FetchGuessListBySeasonIDFromGRPC(ctx, mid, seasonID)
	}

	return
}

func (s *Service) DeleteUserSeasonGuessListByMatchID(ctx context.Context, mid, matchID int64) (err error) {
	var (
		seasonID int64
		cacheKey string
	)
	seasonID, err = s.FetchSeasonIDByMatchID(ctx, matchID)
	if err != nil || seasonID == 0 {
		return
	}

	cacheKey, err = cacheKey4UserSeasonGuessList(ctx, mid, seasonID)
	if err == nil {
		err = component.GlobalMemcached4UserGuess.Delete(ctx, cacheKey)
		if err == memcache.ErrNotFound {
			err = nil
		}
	}

	return
}

func DeleteUserGuessListBySeasonID(ctx context.Context, mid, seasonID int64) (err error) {
	var cacheKey string
	cacheKey, err = cacheKey4UserSeasonGuessList(ctx, mid, seasonID)
	err = component.GlobalMemcached.Delete(ctx, cacheKey)

	return
}

func FetchGuessListBySeasonIDFromCache(ctx context.Context, mid, seasonID int64) (list []*model.MatchGuess, err error) {
	list = make([]*model.MatchGuess, 0)
	var cacheKey string
	cacheKey, err = cacheKey4UserSeasonGuessList(ctx, mid, seasonID)
	if err == nil {
		err = component.GlobalMemcached4UserGuess.Get(ctx, cacheKey).Scan(&list)
	}

	return
}

func fetchContestIDListBySeason(ctx context.Context, seasonID int64) (contestIDList []int64, err error) {
	contestIDList = make([]int64, 0)
	tmpMatchIDList, ok := HotMatchInMemory[seasonID]
	if !ok {
		var tmpRes []*pb.ContestCardComponent
		if tmpRes, err = fetchComponentContestListAll(ctx, seasonID); err != nil {
			log.Errorc(ctx, "FetchGuessListBySeasonIDFromGRPC  fetchComponentContestListAll seasonID(%d) error(%+v)", seasonID, err)
			return
		}
		for _, contest := range tmpRes {
			if contest.GuessType == 0 {
				continue
			}
			contestIDList = append(contestIDList, contest.ID)
		}
	} else {
		contestIDList = tmpMatchIDList
	}
	return
}

func (s *Service) FetchGuessListBySeasonIDFromGRPC(ctx context.Context, mid, seasonID int64) (list []*model.MatchGuess, err error) {
	list = make([]*model.MatchGuess, 0)
	matchIDList := make([]int64, 0)
	if matchIDList, err = fetchContestIDListBySeason(ctx, seasonID); err != nil {
		log.Errorc(ctx, "FetchGuessListBySeasonIDFromGRPC fetchContestIDListBySeason() seasonID(%d) error(%+v)", seasonID, err)
		return
	}
	if len(matchIDList) == 0 {
		return
	}
	list, err = s.FetchGuessListByMatchListFromGRPC(ctx, mid, matchIDList)
	if err != nil {
		return
	}
	if len(list) > 1 {
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].Ctime.Time().Equal(list[j].Ctime.Time()) {
				return list[i].Id > list[j].Id
			}

			return list[i].Ctime.Time().After(list[j].Ctime.Time())
		})
	}
	var cacheKey string
	cacheKey, err = cacheKey4UserSeasonGuessList(ctx, mid, seasonID)
	if err != nil {
		return
	}
	item := &memcache.Item{
		Key:        cacheKey,
		Object:     list,
		Expiration: int32(60),
		Flags:      memcache.FlagJSON,
	}
	_ = component.GlobalMemcached4UserGuess.Set(ctx, item)
	return
}

func (s *Service) FetchGuessListByMatchListFromGRPC(ctx context.Context, mid int64, matchIDList []int64) (list []*model.MatchGuess, err error) {
	list = make([]*model.MatchGuess, 0)
	if len(matchIDList) == 0 {
		return
	}

	limit := 100
	listLen := len(matchIDList)
	lenAfterSplit := listLen / limit
	if d := listLen % limit; d > 0 {
		lenAfterSplit++
	}
	for i := 0; i < lenAfterSplit; i++ {
		startIndex := limit * i
		endIndex := startIndex + limit
		if endIndex > listLen {
			endIndex = listLen
		}

		req := new(api.UserGuessMatchsReq)
		{
			req.Mid = mid
			req.Business = 1
			req.Oids = matchIDList[startIndex:endIndex]
			req.Pn = 1
			req.Ps = 100
		}
		tmpResp, tmpErr := s.actClient.UserGuessMatchs(ctx, req)
		if tmpErr != nil {
			err = tmpErr

			return
		}

		matchMap := make(map[int64]*model.Contest, 0)
		teamMap := make(map[int64]*model.Team, 0)
		matchMap, teamMap, err = s.FetchMatchTeamsInfo(ctx, matchIDList[startIndex:endIndex])
		if err != nil {
			return
		}

		if tmpResp.UserGroup != nil && len(tmpResp.UserGroup) > 0 {
			for _, v := range tmpResp.UserGroup {
				if match, ok := matchMap[v.Oid]; ok {
					if home, ok := teamMap[match.HomeID]; ok {
						if away, ok := teamMap[match.AwayID]; ok {
							tmpMatchGuess := new(model.MatchGuess)
							{
								tmpMatchGuess.GuessUserGroup = v
								tmpMatchGuess.HomeTeam = home.Convert2SimplifyEdition()
								tmpMatchGuess.AwayTeam = away.Convert2SimplifyEdition()
								tmpMatchGuess.GameStage = match.GameStage
							}

							list = append(list, tmpMatchGuess)

							continue
						}
					}
				}

				log.Errorc(ctx, "FetchGuessListByMatchListFromGRPC: generate MatchGuess not expected(%v)", v.Oid)
			}
		}
	}

	return
}

func (s *Service) FetchMatchTeamsInfo(ctx context.Context, matchIDList []int64) (matchM map[int64]*model.Contest,
	teamM map[int64]*model.Team, err error) {
	matchM, err = s.dao.ContestListByIDList(ctx, matchIDList)
	if err != nil {
		return
	}

	teamIDList := make([]int64, 0)
	teamIDMap := make(map[int64]int64, 0)
	for _, v := range matchM {
		if _, ok := teamIDMap[v.HomeID]; !ok {
			teamIDMap[v.HomeID] = 1
			teamIDList = append(teamIDList, v.HomeID)
		}

		if _, ok := teamIDMap[v.AwayID]; !ok {
			teamIDMap[v.AwayID] = 1
			teamIDList = append(teamIDList, v.AwayID)
		}
	}

	teamM, err = s.dao.TeamListByIDList(ctx, teamIDList)

	return
}
