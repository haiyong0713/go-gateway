package service

import (
	"context"
	"fmt"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/esports/ecode"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/model"
	"sort"
)

const (
	_grpcContestBatchSizeMax = 100
	_grpcMaxRoomSize         = 10
	_favStateAdd             = 1
	_favStateCancel          = 2
)

func (s *Service) GetContestInfo(ctx context.Context, req *pb.GetContestRequest) (response *pb.ContestInfo, err error) {
	response = new(pb.ContestInfo)
	if req.Cid == 0 {
		err = xecode.Errorf(xecode.RequestErr, "赛程id非法")
	}
	contestsMap, err := s.getContestsModel(ctx, []int64{req.Cid}, false, false, true)
	if err != nil {
		log.Errorc(ctx, "[GRPC][GetContestInfo][Error], err:%+v", err)
		return
	}
	contest, ok := contestsMap[req.Cid]
	if !ok {
		err = xecode.Errorf(xecode.RequestErr, "赛程不存在或已被冻结")
		return
	}
	// 获取战队信息
	teamIds := getUniqueTeamsByContests([]*model.ContestModel{contest})
	teamsInfoMap, err := s.getTeamsModel(ctx, teamIds, false, false)
	if err != nil {
		log.Errorc(ctx, "[GRPC][GetContestInfo][getTeamsModel][Error], err:%+v", err)
		return
	}
	seasonInfo, err := s.getSeasonModel(ctx, contest.Sid, false, false)
	if err != nil {
		log.Errorc(ctx, "[GRPC][GetContestInfo][getSeasonModel][Error], err:%+v", err)
		return
	}
	seasonMap := map[int64]*model.SeasonModel{contest.Sid: seasonInfo}
	guessCids := make([]int64, 0)
	if contest.GuessType == model.ContestGuessHasTrue {
		guessCids = append(guessCids, contest.ID)
	}
	subRelation, guessRelation, _ := s.getUserInfo(ctx, []int64{contest.ID}, guessCids, req.Mid)
	response.Contest = s.formatContestDetail(contest, teamsInfoMap, seasonMap, guessRelation, subRelation)
	return
}

func (s *Service) getUserInfo(ctx context.Context, subContestIds []int64, guessContestIds []int64, mid int64) (subRelation map[int64]bool, guessRelation map[int64]bool, err error) {
	subRelation = make(map[int64]bool)
	guessRelation = make(map[int64]bool)
	if mid <= 0 {
		return
	}
	eg := errgroup.WithContext(ctx)
	if len(subContestIds) > 0 {
		eg.Go(func(ctx context.Context) error {
			subRelationRes, errG := s.dao.GetSubscribeRelationByContests(ctx, subContestIds, mid)
			if errG != nil {
				log.Errorc(ctx, "[GRPC][getUserInfo][GetSubscribeRelationByContests][Error],err:%+v", errG)
			}
			subRelation = subRelationRes
			return errG
		})
	}
	if len(guessContestIds) > 0 {
		eg.Go(func(ctx context.Context) error {
			guessRelationRes, errG := s.dao.GetGuessDetail(ctx, guessContestIds, mid)
			if errG != nil {
				log.Errorc(ctx, "[GRPC][getUserInfo][GetSubscribeRelationByContests][Error],err:%+v", errG)
			}
			guessRelation = guessRelationRes
			return errG
		})
	}
	err = eg.Wait()
	if err != nil {
		log.Errorc(ctx, "[GRPC][ErrGroup][Error], err:%+v", err)
		return
	}
	return
}

func getUniqueTeamsByContests(contests []*model.ContestModel) (teamIds []int64) {
	teamIds = make([]int64, 0)
	teamIdsMap := make(map[int64]bool)
	for _, v := range contests {
		if v.HomeID != 0 && !mapKeyExist(v.HomeID, teamIdsMap) {
			teamIdsMap[v.HomeID] = true
		}
		if v.AwayID != 0 && !mapKeyExist(v.AwayID, teamIdsMap) {
			teamIdsMap[v.AwayID] = true
		}
		if v.SuccessTeam != 0 && !mapKeyExist(v.SuccessTeam, teamIdsMap) {
			teamIdsMap[v.SuccessTeam] = true
		}
	}
	for k := range teamIdsMap {
		teamIds = append(teamIds, k)
	}
	return
}

func mapKeyExist(key int64, mapEntries map[int64]bool) bool {
	exists := false
	_, ok := mapEntries[key]
	if ok {
		exists = true
	}
	return exists
}

// 按时间筛选赛程列表
func (s *Service) GetContestsByTime(ctx context.Context, req *pb.GetTimeContestsRequest) (response *pb.GetTimeContestsResponse, err error) {
	response = new(pb.GetTimeContestsResponse)
	roomIds := req.RoomIds
	if len(roomIds) > _grpcMaxRoomSize {
		err = xecode.Errorf(xecode.RequestErr, fmt.Sprintf("筛选房间数不可超过%d个", _grpcMaxRoomSize))
		return
	}
	response.Contests = make([]*pb.ContestDetail, 0)
	contestIds, _, cache, err := s.dao.GetContestsCacheOrEs(ctx, &model.ContestsQueryParamsModel{
		MatchId:     req.MatchId,
		Gid:         req.GameId,
		Tid:         req.TeamId,
		Stime:       req.Stime,
		Etime:       req.Etime,
		Sids:        nil,
		RoomIds:     req.GetRoomIds(),
		Sort:        int(req.TimeSort),
		ContestIds:  nil,
		CursorPage:  true,
		Cursor:      req.Cursor,
		CursorSize:  int(req.CursorSize),
		Channel:     req.Channel,
		Debug:       s.conf.Rule.EsSearchQueryDebug,
		NeedInvalid: req.NeedInvalid,
	})
	if err != nil {
		log.Errorc(ctx, "[GRPC][SearchContestsByTime][Error], err:%+v", err)
		return
	}
	response.Cache = cache
	if len(contestIds) == 0 {
		return
	}
	res, err := s.GetContests(ctx, &pb.GetContestsRequest{
		Mid:         req.Mid,
		Cids:        contestIds,
		NeedInvalid: req.NeedInvalid,
	})
	if err != nil {
		log.Errorc(ctx, "[GRPC][GetContests][Error], err:%+v", err)
		return
	}
	response.Contests = res.GetContests()
	response.Cursor = 0
	if len(response.Contests) > 0 {
		response.Cursor = req.Cursor + 1
	}
	return
}

// 赛程数据
func (s *Service) GetContests(ctx context.Context, req *pb.GetContestsRequest) (response *pb.ContestsResponse, err error) {
	response = new(pb.ContestsResponse)
	response.Contests = make([]*pb.ContestDetail, 0)
	if req == nil || req.Cids == nil || len(req.Cids) == 0 || len(req.Cids) > _grpcContestBatchSizeMax {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	contestsMap, err := s.getContestsModel(ctx, req.Cids, false, false, !req.NeedInvalid)
	if err != nil {
		log.Errorc(ctx, "[GRPC][GetContestInfo][Error], err:%+v", err)
		return
	}
	contests := make([]*model.ContestModel, 0)
	guessCids := make([]int64, 0)
	for _, v := range req.Cids {
		if contestsMap[v] == nil {
			continue
		}
		contests = append(contests, contestsMap[v])
		if contestsMap[v].GuessType == model.ContestGuessHasTrue {
			guessCids = append(guessCids, v)
		}
	}
	// 获取战队信息
	teamIds := getUniqueTeamsByContests(contests)
	teamsInfoMap, err := s.getTeamsModel(ctx, teamIds, false, false)
	if err != nil {
		log.Errorc(ctx, "[GRPC][GetContestInfo][getTeamsModel][Error], err:%+v", err)
		return
	}
	seasonIds := getUniqueSeasonId(contests)
	seasonInfo, err := s.getSeasonsModel(ctx, seasonIds, false)
	if err != nil {
		log.Errorc(ctx, "[GRPC][GetContestInfo][getSeasonModel][Error], err:%+v", err)
		return
	}
	subRelation, guessRelation, _ := s.getUserInfo(ctx, req.Cids, guessCids, req.Mid)
	for _, v := range contests {
		response.Contests = append(response.Contests, s.formatContestDetail(v, teamsInfoMap, seasonInfo, guessRelation, subRelation))
	}
	return
}

func getUniqueSeasonId(contests []*model.ContestModel) []int64 {
	seasonIds := make([]int64, 0)
	mapEntries := make(map[int64]bool)
	for _, v := range contests {
		if v.Sid == 0 || mapKeyExist(v.Sid, mapEntries) {
			continue
		}
		seasonIds = append(seasonIds, v.Sid)
	}
	return seasonIds
}

// 添加预约
func (s *Service) AddContestFav(ctx context.Context, req *pb.FavRequest) (response *pb.NoArgsResponse, err error) {
	response = &pb.NoArgsResponse{}
	err = s.favParamsCheck(ctx, req)
	if err != nil {
		return
	}
	if err = s.dao.AddFav(ctx, req.Cid, req.Mid); err != nil {
		log.Error("[GRPC][AddContestFav][Error] LiveAddFav Request(%v) Error(%v)", req, err)
		return
	}
	// 投递消息
	err = s.fanout.Do(ctx, func(ctx context.Context) {
		_ = s.dao.BGroupDataBusPub(ctx, req.Mid, req.Cid, _favStateAdd)
	})
	if err != nil {
		log.Error("[GRPC][AddContestFav][Fanout][Error] LiveAddFav Request(%v) Error(%v)", req, err)
		err = xecode.Errorf(xecode.RequestErr, "订阅失败, 请重试~")
		return
	}
	return
}

// 删除预约
func (s *Service) DelContestFav(ctx context.Context, req *pb.FavRequest) (response *pb.NoArgsResponse, err error) {
	response = &pb.NoArgsResponse{}
	err = s.favParamsCheck(ctx, req)
	if err != nil {
		return
	}
	if err = s.dao.DelFav(ctx, req.Cid, req.Mid); err != nil {
		log.Error("[GRPC[DelContestFav][Fanout]][Error] LiveAddFav Request(%v) Error(%v)", req, err)
		return
	}
	// 投递消息
	err = s.fanout.Do(ctx, func(ctx context.Context) {
		_ = s.dao.BGroupDataBusPub(ctx, req.Mid, req.Cid, _favStateCancel)
	})
	if err != nil {
		log.Error("[GRPC][AddContestFav][Error] LiveAddFav Request(%v) Error(%v)", req, err)
		err = xecode.Errorf(xecode.RequestErr, "取消订阅失败, 请重试~")
		return
	}
	return
}

func (s *Service) favParamsCheck(ctx context.Context, req *pb.FavRequest) (err error) {
	if req == nil || req.Mid == 0 || req.Cid == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	contestsMap, err := s.getContestsModel(ctx, []int64{req.Cid}, false, false, true)
	if err != nil {
		return
	}
	if contestsMap == nil || contestsMap[req.Cid] == nil {
		err = ecode.EsportsContestNotExist
		return
	}
	contest := contestsMap[req.Cid]
	if contest.LiveRoom <= 0 {
		err = ecode.EsportsContestFavNot
		return
	}
	if contest.ContestStatus == model.ContestStatusOver {
		err = ecode.EsportsContestEnd
		return
	}
	if contest.ContestStatus == model.ContestStatusIng {
		err = ecode.EsportsContestStart
		return
	}
	if contest.ContestStatus != model.ContestStatusWaiting {
		err = ecode.EsportsContestFavNot
		return
	}
	return
}

// 赛程订阅用户列表新接口
func (s *Service) GetContestSubscribers(ctx context.Context, req *pb.GetSubscribersRequest) (response *pb.ContestSubscribers, err error) {
	response = new(pb.ContestSubscribers)
	if req == nil || req.Cid == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	response, err = s.dao.GetSubscriberByContestId(ctx, req.Cid, req.Cursor, req.CursorSize)
	if err != nil {
		log.Errorc(ctx, "[GRPC][GetSubscriberByContestId][Error], err:%+v", err)
	}
	return
}

// 获取赛事下的所有游戏
func (s *Service) GetGames(ctx context.Context, req *pb.GetGamesRequest) (response *pb.GamesResponse, err error) {
	response = new(pb.GamesResponse)
	response.Games = make([]*pb.GameDetail, 0)
	var gameModelsMap map[int64]*model.GameModel
	if len(s.gamesCacheMap) == 0 {
		gameModelsMap, err = s.dao.GetAllGamesCache(ctx)
		if err != nil {
			log.Errorc(ctx, "[GRPC][GetGames][GetAllGamesCache][Error], err:%+v", err)
			return
		}
	} else {
		gameModelsMap = s.gamesCacheMap
	}
	if len(req.GameIds) != 0 {
		for _, gameId := range req.GameIds {
			if gameInfo, ok := gameModelsMap[gameId]; ok {
				response.Games = append(response.Games, s.gameModel2ExternalDetail(gameInfo))
			}
		}
		return
	}
	games := make([]*model.GameModel, 0)
	for _, gameModel := range gameModelsMap {
		games = append(games, gameModel)
	}
	sort.Slice(games, func(i, j int) bool {
		return games[i].Rank > games[j].Rank
	})
	for _, game := range games {
		response.Games = append(response.Games, s.gameModel2ExternalDetail(game))
	}
	return
}

// 获取战队详情
func (s *Service) GetTeams(ctx context.Context, req *pb.GetTeamsRequest) (response *pb.TeamsResponse, err error) {
	response = new(pb.TeamsResponse)
	response.Teams = make([]*pb.TeamDetail, 0)
	if req == nil || req.TeamIds == nil || len(req.TeamIds) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "请求参数错误")
		return
	}
	teamsModelMap, err := s.getTeamsModel(ctx, req.TeamIds, false, false)
	if err != nil {
		log.Errorc(ctx, "[GRPC][GetTeams][getTeamsModel][Error], err:%+v", err)
		return
	}
	for _, teamModel := range teamsModelMap {
		response.Teams = append(response.Teams, s.teamModel2ExternalInfo(teamModel))
	}
	return
}

func (s *Service) formatFullLogoPath(imgPath string) string {
	return fmt.Sprintf("%s%s", s.conf.Rule.LogoPathPre, imgPath)
}
