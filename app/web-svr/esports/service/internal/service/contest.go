package service

import (
	"context"
	"fmt"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/ecode"
	pb "go-gateway/app/web-svr/esports/service/api/v1"
	"go-gateway/app/web-svr/esports/service/internal/dao"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

const (
	_formatDate = "2006-01-02"
)

var _defaultChannelList = []int64{int64(1), int64(2)}

func (s *Service) getContestsModel(ctx context.Context, contestIds []int64, skipCache bool, skipMemory bool, onlyValid bool) (contestModelMaps map[int64]*model.ContestModel, err error) {
	contestModelMaps = make(map[int64]*model.ContestModel)
	missIds := contestIds
	if !skipMemory {
		contestModelMaps, missIds = s.getContestCacheFromMemory(contestIds)
	}
	if !skipCache {
		contestModelMapsFromCache, missIdsFromCache, errG := s.dao.GetContestsCache(ctx, missIds)
		if errG != nil {
			err = errG
			log.Errorc(ctx, "[Service][getContestsModel][GetContestsCache], err:%+v", err)
			return
		}
		for k, v := range contestModelMapsFromCache {
			contestModelMaps[k] = v
		}
		missIds = missIdsFromCache
	}
	if len(missIds) == 0 {
		contestModelMaps = filterValidContests(contestModelMaps, onlyValid)
		return
	}
	contestModels, err := s.dao.GetContestsByIds(ctx, missIds, onlyValid)
	if err != nil {
		log.Errorc(ctx, "[Service][getContestsModel][GetContestsByIds], err:%+v", err)
		return
	}
	for _, v := range contestModels {
		contestModelMaps[v.ID] = v
	}
	errG := s.dao.SetContestCache(ctx, contestModels)
	if errG != nil {
		log.Errorc(ctx, "[Service][getContestsModel][SetContestsCache], err:%+v", errG)
	}
	contestModelMaps = filterValidContests(contestModelMaps, onlyValid)
	return
}

func filterValidContests(contestModelMaps map[int64]*model.ContestModel, onlyValid bool) map[int64]*model.ContestModel {
	if !onlyValid || len(contestModelMaps) == 0 {
		return contestModelMaps
	}
	validContestsMap := make(map[int64]*model.ContestModel)
	for _, v := range contestModelMaps {
		if v.Status == model.FreezeTrue {
			continue
		}
		validContestsMap[v.ID] = v
	}
	return validContestsMap
}

func (s *Service) getContestCacheFromMemory(contestIds []int64) (contestModelMap map[int64]*model.ContestModel, missIds []int64) {
	contestModelMap = make(map[int64]*model.ContestModel)
	missIds = make([]int64, 0)
	for _, v := range contestIds {
		if cache, ok := s.activeContestsCacheMap.Get(v); ok {
			if contestCache, valid := cache.(*model.ContestModel); valid {
				contestModelMap[v] = contestCache
				continue
			}
		}
		missIds = append(missIds, v)
	}
	return
}

func (s *Service) SaveContest(ctx context.Context, req *pb.SaveContestReq) (res *pb.NoArgsResponse, err error) {
	res = &pb.NoArgsResponse{}
	if req == nil || req.Contest == nil {
		err = xecode.Errorf(xecode.RequestErr, "赛程信息为空")
		return
	}
	contest := req.GetContest()
	contestId := contest.ID
	// 参数校验
	err = s.contestParamsCheck(ctx, req)
	if err != nil {
		return
	}
	contestModel := s.contestModel2Internal(req.Contest)
	contestData := s.contestDatasModel2Internal(req.ContestData)

	// 新增赛程原子锁
	lockKey, value, err := s.contestUpdateLock(ctx, req.Contest, 0, 0)
	if err != nil {
		return
	}
	defer func() {
		_ = s.dao.RedisUnLock(ctx, lockKey, value)
	}()

	if contestId == 0 {
		err = s.addContest(ctx, contestModel, req.GameIds, req.TeamIds, contestData, req.AdId)
	} else {
		err = s.editContest(ctx, contestModel, req.GameIds, req.TeamIds, contestData)
	}
	if err != nil {
		return
	}
	// 缓存相关信息的构造
	err = s.clearCacheByContestId(ctx, contest.ID)
	return
}

func (s *Service) contestUpdateLock(ctx context.Context, contest *pb.ContestModel, retry int, internalMillSeconds int64) (lockKey string, value string, err error) {
	lockKey = formatContestEditLock(contest)
	value = s.dao.RedisUniqueValue()
	ttl := dao.ContestEditRedisLockTTL
	err = s.dao.RedisLock(ctx, lockKey, value, int64(ttl), retry, internalMillSeconds)
	if err != nil {
		log.Errorc(ctx, "[Service][SaveContest][RedisLock][Error], err:%+v", err)
		err = xecode.Errorf(xecode.RequestErr, "请求过于频繁，请稍后再试~")
		return
	}
	return
}

func (s *Service) clearCacheByContestId(ctx context.Context, contestId int64) (err error) {
	err = s.dao.DeleteContestCache(ctx, contestId)
	return
}

func (s *Service) addContest(
	ctx context.Context,
	contest *model.ContestModel,
	gameIds []int64,
	teamIds []int64,
	contestData []*model.ContestDataModel,
	adId int64,
) (err error) {
	err = s.dao.ContestAddTransaction(ctx, contest, gameIds, teamIds, contestData, adId)
	if err != nil {
		log.Errorc(ctx, "[Service][AddContest][ContestAddTransaction][Error], err:%+v", err)
		return
	}
	return
}

func (s *Service) editContest(
	ctx context.Context,
	contest *model.ContestModel,
	gameIds []int64,
	teamIds []int64,
	contestData []*model.ContestDataModel,
) (err error) {
	err = s.dao.ContestUpdateTransaction(ctx, contest, gameIds, teamIds, contestData)
	if err != nil {
		log.Errorc(ctx, "[Service][editContest][ContestUpdateTransaction][Error], err:%+v", err)
		return
	}
	return
}

func formatContestEditLock(contest *pb.ContestModel) string {
	return fmt.Sprintf(dao.ContestEditRedisLock, contest.Sid, contest.ID, contest.ExternalID)
}

func (s *Service) contestParamsCheck(ctx context.Context, req *pb.SaveContestReq) (err error) {
	contest := req.Contest
	// 时间校验
	if contest.Stime >= contest.Etime {
		err = xecode.Errorf(xecode.RequestErr, "比赛结束时间必须大于开始时间")
		return
	}
	// 直播间校验
	if contest.LiveRoom > 0 && !s.isLiveRoomValid(ctx, []int64{contest.LiveRoom}) {
		err = xecode.Errorf(ecode.EsportsMatchLiveInvalid, "赛程直播地址不存在")
		return
	}
	// 赛季校验
	seasonId := contest.Sid
	seasonInfo, err := s.getSeasonModel(ctx, seasonId, true, true)
	if err != nil {
		log.Errorc(ctx, "[Service][Contest][ContestParamsCheck][GetSeasonInfo][Error], err:%+v", err)
		err = xecode.Errorf(xecode.RequestErr, "赛季信息获取失败，请重试")
		return
	}
	if seasonInfo.SeasonType == int64(model.SeasonTypeEscape) && contest.Special != int64(model.ContestSpecial) {
		err = xecode.Errorf(xecode.RequestErr, "赛季的比赛类型为大逃杀类，添加赛程时只能添加特殊赛程")
		return
	}

	// 赛事校验
	matchId := contest.Mid
	if matchId == 0 {
		err = xecode.Errorf(xecode.RequestErr, "绑定的赛事为空，请确认参数")
		return
	}
	_, err = s.getMatchModel(ctx, matchId, true)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, err.Error())
		return
	}
	gameIds := req.GameIds
	if len(gameIds) == 0 {
		err = xecode.Errorf(xecode.RequestErr, "绑定的游戏列表为空")
		return
	}
	if len(gameIds) == 0 || len(gameIds) > 1 {
		err = xecode.Errorf(xecode.RequestErr, "一个赛程需要绑定一个游戏，且只能属于一个游戏")
		return
	}
	gamesInfo, err := s.getGamesModel(ctx, gameIds, true)
	if err != nil || len(gamesInfo) != len(gameIds) {
		err = xecode.Errorf(xecode.RequestErr, "绑定的游戏列表有误，请检查")
		return
	}
	// 根据时间计算比赛状态
	if contest.ContestStatus == model.ContestStatusInit {
		contest.ContestStatus = contestStatusCalculate(contest.Stime, contest.Etime)
	}
	return
}

func contestStatusCalculate(startTime int64, endTime int64) (contestStatus int64) {
	now := time.Now().Unix()
	if now < startTime {
		return model.ContestStatusWaiting
	}
	if now >= startTime && now < endTime {
		return model.ContestStatusIng
	}
	return model.ContestStatusOver
}

func (s *Service) isLiveRoomValid(ctx context.Context, roomIDs []int64) (isValid bool) {
	isValid = false
	if len(roomIDs) == 0 {
		return
	}
	for _, v := range roomIDs {
		if v == 0 {
			return
		}
	}
	if d, err := s.dao.LiveRoomInfo(ctx, roomIDs); err == nil && len(d) == len(roomIDs) {
		isValid = true
	}
	return
}

// GetContestModel 获取赛程
func (s *Service) GetContestModel(ctx context.Context, req *pb.GetContestModelReq) (contestInfo *pb.ContestModel, err error) {
	contestInfo = &pb.ContestModel{}
	contestModel, err := s.dao.GetContestById(ctx, req.GetContestId(), true)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "赛程不存在或已被冻结")
		log.Errorc(ctx, "[Service][GetContestModel][getContestsModel][Error],req:%+v err:%+v", req, err)
		return
	}
	contestInfo = s.contestModel2External(contestModel)
	return
}

// RefreshContestStatusInfo .
func (s *Service) RefreshContestStatusInfo(ctx context.Context, req *pb.RefreshContestStatusInfoReq) (res *pb.NoArgsResponse, err error) {
	res = &pb.NoArgsResponse{}
	contestModel, err := s.GetContestModel(ctx, &pb.GetContestModelReq{ContestId: req.ContestId})
	if err != nil {
		return
	}
	if contestModel == nil || contestModel.Status == model.FreezeTrue {
		err = xecode.Errorf(xecode.RequestErr, "赛程不存在或已被冻结")
		return
	}
	// 新增赛程原子锁
	lockKey, value, err := s.contestUpdateLock(ctx, contestModel, 3, 100)
	if err != nil {
		return
	}
	defer func() {
		_ = s.dao.RedisUnLock(ctx, lockKey, value)
	}()
	newContestStatus := contestStatusCalculate(contestModel.Stime, contestModel.Etime)
	if newContestStatus == contestModel.ContestStatus {
		// warn日志
		log.Warnc(ctx, "[Service][RefreshContestStatusInfo][Warn], status no need to change, time:%d, contest:%+v, newContestStatus:%d",
			time.Now().Unix(),
			contestModel,
			newContestStatus,
		)
		return
	}
	err = s.dao.ContestContestStatusUpdate(ctx, contestModel.ID, newContestStatus)
	if err != nil {
		log.Errorc(ctx, "[Service][RefreshContestStatusInfo][ContestContestStatusUpdate][Error], contestID(%d) err:%+v", req.ContestId, err)
		return
	}
	return
}

func (s *Service) GetContestInfoListBySeason(ctx context.Context, req *pb.GetContestInfoListBySeasonReq) (res *pb.GetContestInfoListBySeasonResponse, err error) {
	res = &pb.GetContestInfoListBySeasonResponse{}
	arg := &pb.GetSeasonContestsReq{SeasonId: req.SeasonID}
	contestInfoList, err := s.GetSeasonContests(ctx, arg)
	if err != nil {
		log.Errorc(ctx, "[Service][GetContestInfoListBySeason][GetSeasonContests][Error], req(%+v) err:%+v", req, err)
		return
	}
	if contestInfoList == nil {
		log.Infoc(ctx, "[Service][GetContestInfoListBySeason][GetSeasonContests], req(%+v) contestInfoList nil", req)
		return
	}
	tmpMap := make(map[int64]*pb.SeasonContests, len(contestInfoList.Contests))
	for _, contestInfo := range contestInfoList.Contests {
		cardList := make([]*pb.ContestDetail, 0)
		if contestInfo.GetStime() == 0 {
			continue
		}
		dateStr := time.Unix(contestInfo.GetStime(), 0).Format(_formatDate)
		tmpDate, err := time.ParseInLocation(_formatDate, dateStr, time.Local)
		if err != nil {
			log.Errorc(ctx, "[Service][GetContestInfoListBySeason][ParseInLocation][Error], req(%+v) err:%+v", req, err)
			return nil, err
		}
		dateUnix := tmpDate.Unix()
		if d, ok := tmpMap[dateUnix]; ok {
			cardList = d.Contests
		}
		cardList = append(cardList, contestInfo)
		tmpMap[dateUnix] = &pb.SeasonContests{Contests: cardList}
	}
	res.ComponentContestList = tmpMap
	return
}

func (s *Service) GetSeasonContests(ctx context.Context, req *pb.GetSeasonContestsReq) (res *pb.SeasonContests, err error) {
	res = &pb.SeasonContests{}
	if req == nil {
		err = xecode.Errorf(xecode.RequestErr, "参数为空")
		return
	}
	seasonId := req.SeasonId
	seasonInfo, err := s.getSeasonModel(ctx, seasonId, false, false)
	if err != nil {
		log.Errorc(ctx, "[Service][GetSeasonContests][GetSeasonInfo][Error], err:%+v", err)
		return
	}
	contestModels, err := s.getSeasonContests(ctx, seasonId, false)
	if err != nil {
		return
	}
	teamIds := getUniqueTeamsByContests(contestModels)
	teamsMap, err := s.getTeamsModel(ctx, teamIds, false, false)
	if err != nil {
		log.Errorc(ctx, "[Service][GetSeasonContests][getTeamsModel][Error], err:%+v", err)
		return
	}
	res = new(pb.SeasonContests)
	res.Contests = make([]*pb.ContestDetail, 0)
	seasonMap := map[int64]*model.SeasonModel{seasonId: seasonInfo}
	for _, contestModel := range contestModels {
		res.Contests = append(res.Contests, s.formatContestDetail(contestModel, teamsMap, seasonMap, nil, nil))
	}
	return
}

func (s *Service) formatContestDetail(
	contestModel *model.ContestModel,
	teamsMap map[int64]*model.TeamModel,
	seasonMap map[int64]*model.SeasonModel,
	guessRelations map[int64]bool,
	subscribeRelations map[int64]bool,
) *pb.ContestDetail {
	if contestModel == nil {
		return nil
	}
	season := seasonMap[contestModel.Sid]
	homeTeam := teamsMap[contestModel.HomeID]
	awayTeam := teamsMap[contestModel.AwayID]
	successTeam := teamsMap[contestModel.SuccessTeam]
	return &pb.ContestDetail{
		ID:              contestModel.ID,
		GameStage:       contestModel.GameStage,
		Stime:           contestModel.Stime,
		Etime:           contestModel.Etime,
		HomeID:          contestModel.HomeID,
		AwayID:          contestModel.AwayID,
		HomeScore:       contestModel.HomeScore,
		AwayScore:       contestModel.AwayScore,
		HomeTeam:        s.formatTeam(homeTeam),
		AwayTeam:        s.formatTeam(awayTeam),
		Sid:             contestModel.Sid,
		Season:          s.formatSeason(season),
		Mid:             contestModel.Mid,
		SeriesId:        contestModel.SeriesId,
		Series:          nil,
		LiveRoom:        contestModel.LiveRoom,
		Aid:             contestModel.Aid,
		Collection:      contestModel.Collection,
		Dic:             contestModel.Dic,
		Special:         contestModel.Special,
		SuccessTeam:     contestModel.SuccessTeam,
		SuccessTeamInfo: s.formatTeam(successTeam),
		SpecialName:     contestModel.SpecialName,
		SpecialTips:     contestModel.SpecialTips,
		SpecialImage:    contestModel.SpecialImage,
		Playback:        contestModel.Playback,
		CollectionURL:   contestModel.CollectionURL,
		LiveURL:         contestModel.LiveURL,
		DataType:        contestModel.DataType,
		MatchID:         contestModel.MatchID,
		GameStage1:      contestModel.GameStage1,
		GameStage2:      contestModel.GameStage2,
		JumpURL:         s.formatContestJumpUrl(contestModel.ID),
		GuessLink:       s.formatContestGuessUrl(contestModel.ID),
		ContestFrozen:   s.formatContestFrozen(contestModel),
		ContestStatus:   s.formatContestStatus(contestModel),
		IsGuessed:       s.formatIsGuessed(contestModel, guessRelations),
		IsSubscribed:    s.formatIsSubScribed(contestModel, subscribeRelations),
		GameId:          contestModel.GameId,
		Game:            s.formatContestGame(contestModel.GameId),
	}
}

func (s *Service) formatContestGame(gameId int64) (game *pb.GameDetail) {
	game = new(pb.GameDetail)
	if gameId == 0 {
		return
	}
	gamesMap := s.getGamesInfoFromLocal([]int64{gameId})
	if gameModel, ok := gamesMap[gameId]; ok {
		game = s.gameModel2ExternalDetail(gameModel)
	}
	return
}

func (s *Service) formatContestStatus(contestModel *model.ContestModel) (contestStatus pb.ContestStatusEnum) {
	contestStatus = pb.ContestStatusEnum_Waiting
	if contestModel.ContestStatus == model.ContestStatusIng {
		contestStatus = pb.ContestStatusEnum_Ing
	}
	if contestModel.ContestStatus == model.ContestStatusOver {
		contestStatus = pb.ContestStatusEnum_Over
	}
	return
}

func (s *Service) formatContestFrozen(contestModel *model.ContestModel) (contestFrozen pb.ContestFrozenEnum) {
	contestFrozen = pb.ContestFrozenEnum_True
	if contestModel.Status == model.FreezeFalse {
		contestFrozen = pb.ContestFrozenEnum_False
	}
	return
}

func (s *Service) formatIsGuessed(contest *model.ContestModel, guessRelation map[int64]bool) (isGuessed pb.GuessStatusEnum) {
	if contest.GuessType == model.ContestGuessHasTrue {
		isGuessed = pb.GuessStatusEnum_HasGuessNoGuessed
	} else {
		isGuessed = pb.GuessStatusEnum_HasNoGuess
		return
	}
	if !s.formatCanGuess(contest) {
		isGuessed = pb.GuessStatusEnum_HasGuessOverNoGuessed
		// 非未开始状态
		if guessRelation == nil {
			return
		}
		if v, ok := guessRelation[contest.ID]; ok && v {
			isGuessed = pb.GuessStatusEnum_HasGuessOverGuessed
		}
	} else {
		// 未开始状态
		if v, ok := guessRelation[contest.ID]; ok && v {
			isGuessed = pb.GuessStatusEnum_HasGuessGuessed
		}
	}
	return
}

func (s *Service) formatIsSubScribed(contest *model.ContestModel, subscribeRelation map[int64]bool) (isSubScribed pb.SubscribedStatusEnum) {
	// 可订阅的标记
	if contest.LiveRoom > 0 {
		isSubScribed = pb.SubscribedStatusEnum_CanSubNoSub
	} else {
		isSubScribed = pb.SubscribedStatusEnum_CanNotSub
		return
	}

	if contest.ContestStatus != model.ContestStatusWaiting {
		isSubScribed = pb.SubscribedStatusEnum_CanSubOverNoSub
		// 非未开始状态
		if subscribeRelation == nil {
			return
		}
		if v, ok := subscribeRelation[contest.ID]; ok && v {
			isSubScribed = pb.SubscribedStatusEnum_CanSubOverSubed
		}
	} else {
		// 未开始状态
		if v, ok := subscribeRelation[contest.ID]; ok && v {
			isSubScribed = pb.SubscribedStatusEnum_CanSubSubed
		}
	}
	return
}

func (s *Service) formatCanGuess(contestModel *model.ContestModel) bool {
	canGuess := false
	guessOverBeforeSTimeSeconds := 0
	if s.conf.Rule.GuessOverBeforeSTime != 0 {
		guessOverBeforeSTimeSeconds = s.conf.Rule.GuessOverBeforeSTime
	}
	if contestModel.GuessType == model.ContestGuessHasTrue &&
		contestModel.ContestStatus == model.ContestStatusWaiting &&
		contestModel.Stime-time.Now().Unix() > int64(guessOverBeforeSTimeSeconds) {
		canGuess = true
	}
	return canGuess
}

func (s *Service) formatContestJumpUrl(contestId int64) string {
	var jumpURL string
	if s.conf.Rule.ContestJumpURL != "" {
		jumpURL = fmt.Sprintf(s.conf.Rule.ContestJumpURL, contestId)
	}
	return jumpURL
}

func (s *Service) formatContestGuessUrl(contestId int64) string {
	var guessUrl string
	if s.conf.Rule.ContestGuessURL != "" {
		guessUrl = fmt.Sprintf(s.conf.Rule.ContestGuessURL, contestId)
	}
	return guessUrl
}

func (s *Service) formatSeason(seasonModel *model.SeasonModel) *pb.SeasonDetail {
	if seasonModel == nil {
		return nil
	}
	return &pb.SeasonDetail{
		ID:           seasonModel.ID,
		Mid:          seasonModel.Mid,
		Title:        seasonModel.Title,
		SubTitle:     seasonModel.SubTitle,
		Stime:        seasonModel.Stime,
		Etime:        seasonModel.Etime,
		Sponsor:      seasonModel.Sponsor,
		Logo:         seasonModel.Logo,
		Dic:          seasonModel.Dic,
		Status:       seasonModel.Status,
		Rank:         seasonModel.Rank,
		IsApp:        seasonModel.IsApp,
		URL:          seasonModel.URL,
		DataFocus:    seasonModel.DataFocus,
		FocusURL:     seasonModel.FocusURL,
		SearchImage:  seasonModel.SearchImage,
		LogoFull:     s.formatFullLogoPath(seasonModel.Logo),
		SyncPlatform: seasonModel.SyncPlatform,
		Channel:      s.formatSeasonChannel(seasonModel),
	}
}

func (s *Service) formatSeasonChannel(seasonModel *model.SeasonModel) (channel []int64) {
	channel = make([]int64, 0)
	for _, channelId := range _defaultChannelList {
		if channelId&seasonModel.SyncPlatform > 0 {
			channel = append(channel, channelId)
		}
	}
	return
}

func (s *Service) formatTeam(teamModel *model.TeamModel) *pb.TeamDetail {
	if teamModel == nil {
		return nil
	}
	return &pb.TeamDetail{
		ID:       teamModel.ID,
		Title:    teamModel.Title,
		SubTitle: teamModel.SubTitle,
		ETitle:   teamModel.ETitle,
		Area:     teamModel.Area,
		Logo:     teamModel.Logo,
		Uid:      teamModel.UID,
		Members:  teamModel.Members,
		Dic:      teamModel.Dic,
		TeamType: 0,
		LogoFull: s.formatFullLogoPath(teamModel.Logo),
		RegionId: teamModel.RegionId,
	}
}

func (s *Service) getSeasonContests(ctx context.Context, seasonId int64, skipCache bool) (contestModels []*model.ContestModel, err error) {
	contestModels = make([]*model.ContestModel, 0)
	if !skipCache {
		contestModels = s.getSeasonContestsFromCache(seasonId)
		if contestModels != nil {
			return
		}
	}
	contestIds, err := s.getSeasonContestIds(ctx, seasonId, false)
	if err != nil {
		return
	}
	contestModelsMap, err := s.getContestsModel(ctx, contestIds, skipCache, skipCache, true)
	if err != nil {
		return
	}
	for _, v := range contestModelsMap {
		contestModels = append(contestModels, v)
	}
	return
}

func (s *Service) GetContestGameModel(ctx context.Context, req *pb.GetContestGameReq) (response *pb.GameModel, err error) {
	response = new(pb.GameModel)
	if req == nil || req.ID == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	gameModel, err := s.dao.GetContestGameById(ctx, req.ID)
	if err != nil {
		log.Errorc(ctx, "[Service][GetContestGame][Error],err:%+v", err)
		return
	}
	response = s.gameModel2External(gameModel)
	return
}

func (s *Service) GetContestGameDetail(ctx context.Context, req *pb.GetContestGameReq) (response *pb.GameDetail, err error) {
	response = new(pb.GameDetail)
	if req == nil || req.ID == 0 {
		err = xecode.Errorf(xecode.RequestErr, "参数错误")
		return
	}
	gameModel, err := s.dao.GetContestGameById(ctx, req.ID)
	if err != nil {
		log.Errorc(ctx, "[Service][GetContestGame][Error],err:%+v", err)
		return
	}
	response = s.gameModel2ExternalDetail(gameModel)
	return
}

func (s *Service) GetSeasonByTime(ctx context.Context, in *pb.GetSeasonByTimeReq) (resp *pb.GetSeasonByTimeResponse, err error) {
	resp = new(pb.GetSeasonByTimeResponse)
	seasonIds, err := s.dao.GetDistinctSeasonByTime(ctx, in.BeginTime, in.EndTime)
	if err != nil || len(seasonIds) == 0 {
		return
	}
	seasons, err := s.dao.GetSeasonsByIDs(ctx, seasonIds)
	if err != nil {
		return
	}
	resp.Seasons = make([]*pb.SeasonDetail, 0)
	for _, season := range seasons {
		resp.Seasons = append(resp.Seasons, s.formatSeason(season))
	}
	return
}

func (s *Service) contestModel2Internal(fromModel *pb.ContestModel) (toModel *model.ContestModel) {
	if fromModel == nil {
		return nil
	}
	return &model.ContestModel{
		ID:            fromModel.ID,
		GameStage:     fromModel.GameStage,
		Stime:         fromModel.Stime,
		Etime:         fromModel.Etime,
		HomeID:        fromModel.HomeID,
		AwayID:        fromModel.AwayID,
		HomeScore:     fromModel.HomeScore,
		AwayScore:     fromModel.AwayScore,
		LiveRoom:      fromModel.LiveRoom,
		Aid:           fromModel.Aid,
		Collection:    fromModel.Collection,
		Dic:           fromModel.Dic,
		Status:        fromModel.Status,
		Sid:           fromModel.Sid,
		Mid:           fromModel.Mid,
		Special:       fromModel.Special,
		SuccessTeam:   fromModel.SuccessTeam,
		SpecialName:   fromModel.SpecialName,
		SpecialTips:   fromModel.SpecialTips,
		SpecialImage:  fromModel.SpecialImage,
		Playback:      fromModel.Playback,
		CollectionURL: fromModel.CollectionURL,
		LiveURL:       fromModel.LiveURL,
		DataType:      fromModel.DataType,
		MatchID:       fromModel.MatchID,
		GuessType:     fromModel.GuessType,
		GameStage1:    fromModel.GameStage1,
		GameStage2:    fromModel.GameStage2,
		SeriesId:      fromModel.SeriesId,
		PushSwitch:    fromModel.PushSwitch,
		ActivePush:    fromModel.ActivePush,
		ContestStatus: fromModel.ContestStatus,
		ExternalID:    fromModel.ExternalID,
	}
}

func (s *Service) contestModel2External(fromModel *model.ContestModel) (toModel *pb.ContestModel) {
	if fromModel == nil {
		return nil
	}
	return &pb.ContestModel{
		ID:            fromModel.ID,
		GameStage:     fromModel.GameStage,
		Stime:         fromModel.Stime,
		Etime:         fromModel.Etime,
		HomeID:        fromModel.HomeID,
		AwayID:        fromModel.AwayID,
		HomeScore:     fromModel.HomeScore,
		AwayScore:     fromModel.AwayScore,
		LiveRoom:      fromModel.LiveRoom,
		Aid:           fromModel.Aid,
		Collection:    fromModel.Collection,
		Dic:           fromModel.Dic,
		Status:        fromModel.Status,
		Sid:           fromModel.Sid,
		Mid:           fromModel.Mid,
		Special:       fromModel.Special,
		SuccessTeam:   fromModel.SuccessTeam,
		SpecialName:   fromModel.SpecialName,
		SpecialTips:   fromModel.SpecialTips,
		SpecialImage:  fromModel.SpecialImage,
		Playback:      fromModel.Playback,
		CollectionURL: fromModel.CollectionURL,
		LiveURL:       fromModel.LiveURL,
		DataType:      fromModel.DataType,
		MatchID:       fromModel.MatchID,
		GuessType:     fromModel.GuessType,
		GameStage1:    fromModel.GameStage1,
		GameStage2:    fromModel.GameStage2,
		SeriesId:      fromModel.SeriesId,
		PushSwitch:    fromModel.PushSwitch,
		ActivePush:    fromModel.ActivePush,
		ContestStatus: fromModel.ContestStatus,
		ExternalID:    fromModel.ExternalID,
	}
}

func (s *Service) contestDatasModel2Internal(fromModel []*pb.ContestDataModel) (toModel []*model.ContestDataModel) {
	toModel = make([]*model.ContestDataModel, 0)
	for _, v := range fromModel {
		single := s.contestDataModel2External(v)
		toModel = append(toModel, single)
	}
	return
}

func (s *Service) contestDataModel2External(froModel *pb.ContestDataModel) *model.ContestDataModel {
	return &model.ContestDataModel{
		ID:        froModel.ID,
		Cid:       froModel.Cid,
		Url:       froModel.Url,
		PointData: froModel.PointData,
		AvCid:     froModel.AvCid,
	}
}
