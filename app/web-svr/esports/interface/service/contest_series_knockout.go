package service

import (
	"context"
	"encoding/json"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/ecode"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/dao"
	"go-gateway/app/web-svr/esports/interface/model"
)

// AddSeriesKnockoutMatchConfig: 向阶段中添加淘汰赛配置
func (s *Service) AddSeriesKnockoutMatchConfig(ctx context.Context, req *v1.SeriesKnockoutMatchConfig) (res *v1.AddSeriesKnockoutMatchConfigResp, err error) {
	res = &v1.AddSeriesKnockoutMatchConfigResp{}
	var info *v1.SeriesKnockoutMatchInfo
	info, err = s.generatePreviewKnockoutMatchInfo(ctx, req)
	if err != nil {
		return
	}
	err = s.innerSetSeriesKnockoutMatchConfig(ctx, req, true)
	if err != nil {
		return
	}
	err = s.dao.SetSeriesKnockoutMatchInfo(ctx, info)
	return
}

// GetSeriesKnockoutMatchConfig: 获取阶段下的添加淘汰赛配置
func (s *Service) GetSeriesKnockoutMatchConfig(ctx context.Context, req *v1.GetSeriesKnockoutMatchConfigReq) (res *v1.SeriesKnockoutMatchConfig, err error) {
	return s.dao.GetSeriesKnockoutMatchConfig(ctx, req.SeriesId)
}

// UpdateSeriesKnockoutMatchConfig: 修改阶段中的淘汰赛配置
func (s *Service) UpdateSeriesKnockoutMatchConfig(ctx context.Context, req *v1.SeriesKnockoutMatchConfig) (res *v1.UpdateSeriesKnockoutMatchConfigResp, err error) {
	res = &v1.UpdateSeriesKnockoutMatchConfigResp{}
	var info *v1.SeriesKnockoutMatchInfo
	info, err = s.generatePreviewKnockoutMatchInfo(ctx, req)
	if err != nil {
		return
	}
	err = s.innerSetSeriesKnockoutMatchConfig(ctx, req, false)
	if err != nil {
		return
	}
	err = s.dao.SetSeriesKnockoutMatchInfo(ctx, info)
	return
}

// innerSetSeriesKnockoutMatchConfig: 向阶段中设置淘汰赛配置
func (s *Service) innerSetSeriesKnockoutMatchConfig(ctx context.Context, req *v1.SeriesKnockoutMatchConfig, shouldEmpty bool) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "innerSetSeriesKnockoutMatchConfig error: %v", err)
		}
	}()
	exist, empty, err := s.dao.IsSeriesExistsAndExtraConfigEmpty(ctx, req.SeriesId, dao.SeriesTypKnockout)
	if err != nil {
		return
	}
	if !exist {
		err = ecode.EsportsContestSeriesNotFound
		return
	}
	if empty != shouldEmpty {
		err = ecode.EsportsContestSeriesExtraConfigNotFound
		if shouldEmpty {
			err = ecode.EsportsContestSeriesExtraConfigFound
		}
		return
	}
	bs, err := json.Marshal(req)
	if err != nil {
		return
	}
	err = s.dao.SetSeriesExtraConfig(ctx, req.SeriesId, string(bs))
	return
}

func (s *Service) PreviewSeriesKnockoutMatchInfo(ctx context.Context, req *v1.SeriesKnockoutMatchConfig) (res *v1.SeriesKnockoutMatchInfo, err error) {
	return s.generatePreviewKnockoutMatchInfo(ctx, req)
}

func (s *Service) RefreshSeriesKnockoutMatchInfo(ctx context.Context, req *v1.RefreshSeriesKnockoutMatchInfoReq) (res *v1.SeriesKnockoutMatchInfo, err error) {
	res = &v1.SeriesKnockoutMatchInfo{}
	config, err := s.dao.GetSeriesKnockoutMatchConfig(ctx, req.SeriesId)
	if err != nil {
		log.Errorc(ctx, "RefreshSeriesKnockoutMatchInfo: s.dao.GetSeriesPointsMatchConfig error: %v", err)
		if err == ecode.EsportsContestSeriesExtraConfigNotFound {
			err = nil
			return
		}
		err = ecode.EsportsContestSeriesExtraConfigErr
		return
	}
	res, err = s.generatePreviewKnockoutMatchInfo(ctx, config)
	if err != nil {
		return
	}
	err = s.dao.SetSeriesKnockoutMatchInfo(ctx, res)
	return
}

func (s *Service) generatePreviewKnockoutMatchInfo(ctx context.Context, req *v1.SeriesKnockoutMatchConfig) (res *v1.SeriesKnockoutMatchInfo, err error) {
	contestInfos, err := s.dao.GetContestsBySeasonId(ctx, req.SeasonId)
	if err != nil {
		return
	}
	//get team infos
	teams, err := s.dao.GetTeamsInSeasonFromDB(ctx, []int64{req.SeasonId})
	if err != nil {
		return
	}
	teamIds := make([]int64, 0, len(teams))
	for _, t := range teams[req.SeasonId] {
		teamIds = append(teamIds, t.TeamId)
	}
	teamInfos, err := s.dao.RawEpTeams(ctx, teamIds)
	if err != nil {
		return
	}
	return s.innerGeneratePreviewKnockoutMatchInfo(ctx, req, teamInfos, contestInfos)
}

func (s *Service) innerGeneratePreviewKnockoutMatchInfo(ctx context.Context, req *v1.SeriesKnockoutMatchConfig,
	teamInfos map[int64]*model.Team,
	contestInfos map[int64]*model.Contest) (res *v1.SeriesKnockoutMatchInfo, err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "innerGeneratePreviewKnockoutMatchInfo error: %v", err)
		}
	}()
	//使用json Marshal/Unmarshal 实现deepCopy
	{
		var bs []byte
		bs, err = json.Marshal(req)
		if err != nil {
			return
		}
		err = json.Unmarshal(bs, &res)
		if err != nil {
			return
		}
		if res == nil {
			return
		}
	}

	err = res.Travel(func(m *v1.SeriesKnockoutContestInfoItem) (err error) {
		contest, ok := contestInfos[m.ContestId]
		if !ok {
			err = ecode.EsportsContestNotExist
			return err
		}
		m.ContestInfo = s.convertContest2Card(contest)
		m.ContestFreeze = contest.Status
		m.GameStage = contest.GameStage
		m.HomeTeamID = contest.HomeID
		m.AwayTeamID = contest.AwayID
		m.StartTime = contest.Stime
		m.EndTime = contest.Etime
		m.HomeTeamScore = contest.HomeScore
		m.AwayTeamScore = contest.AwayScore

		//根据比分判断胜负队伍ID,
		winTeamId := int64(0)
		if contest.ContestStatus == ContestStatusEnd {
			if contest.HomeScore > contest.AwayScore {
				winTeamId = contest.HomeID
			} else if contest.HomeScore < contest.AwayScore {
				winTeamId = contest.AwayID
			}
			//比分相同不处理, 保持winTeamId=0
		}
		m.WinTeamID = winTeamId

		homeTeam := teamInfos[m.HomeTeamID]

		if homeTeam != nil {
			m.HomeTeamName = homeTeam.Title
			m.HomeTeamLogo = homeTeam.Logo
		}
		m.ContestInfo.Home = convertTeam2Card(homeTeam)
		awayTeam := teamInfos[m.AwayTeamID]
		if awayTeam != nil {
			m.AwayTeamName = awayTeam.Title
			m.AwayTeamLogo = awayTeam.Logo
		}
		m.ContestInfo.Away = convertTeam2Card(awayTeam)
		return
	})
	res.RefreshTime = time.Now().Unix()
	return
}

func convertTeam2Card(team *model.Team) *v1.Team4FrontendComponent {
	if team == nil {
		return nil
	}
	return &v1.Team4FrontendComponent{
		ID:       team.ID,
		Icon:     team.Logo,
		Name:     team.Title,
		Wins:     0,
		Region:   genTeamRegionDisplayByRegionID(team.RegionID),
		RegionID: team.RegionID,
	}
}

func convertComponentTeam2Card(team *model.Team2TabComponent) *v1.Team4FrontendComponent {
	if team == nil {
		return nil
	}
	return &v1.Team4FrontendComponent{
		ID:       team.ID,
		Icon:     team.Logo,
		Name:     team.Title,
		Wins:     0,
		Region:   genTeamRegionDisplayByRegionID(team.RegionID),
		RegionID: team.RegionID,
	}
}

func (s *Service) convertContest2Card(contest *model.Contest) *v1.ContestCardComponent {
	if contest == nil {
		return nil
	}
	return &v1.ContestCardComponent{
		ID:            contest.ID,
		StartTime:     contest.Stime,
		EndTime:       contest.Etime,
		Title:         "",
		Status:        "",
		CollectionURL: contest.CollectionURL,
		LiveRoom:      contest.LiveRoom,
		PlayBack:      contest.Playback,
		DataType:      contest.DataType,
		MatchID:       contest.MatchID,
		SeasonID:      contest.Sid,
		GuessType:     int64(contest.GuessType),
		SeriesID:      contest.SeriesID,
		IsSub:         0,
		IsGuess:       0,
		Home:          nil,
		Away:          nil,
		Series:        nil,
		ContestStatus: contest.ContestStatus,
		ContestFreeze: contest.Status,
		GameState:     contest.GameState,
		GuessShow:     int64(contest.GuessShow),
		HomeScore:     contest.HomeScore,
		AwayScore:     contest.AwayScore,
	}
}
func (s *Service) GetSeriesKnockoutMatchInfoHttp(ctx context.Context, mid int64, req *v1.GetSeriesKnockoutMatchInfoReq) (res *v1.SeriesKnockoutMatchInfo, err error) {
	res, err = s.GetSeriesKnockoutMatchInfo(ctx, req)
	if mid == 0 {
		return
	}
	contestIDList4FavComponent, contestIDList4GuessComponent := getFavAndGuessContestIds(res)
	subscribeMap := s.fetchFavoriteMap(ctx, mid, contestIDList4FavComponent)
	guessMap := s.fetchComponentContestGuessMap(ctx, mid, contestIDList4GuessComponent)
	if len(subscribeMap) > 0 || len(guessMap) > 0 {
		err = res.Travel(func(m *v1.SeriesKnockoutContestInfoItem) (err error) {
			if m != nil && m.ContestInfo != nil {
				if d, ok := subscribeMap[m.ContestInfo.ID]; ok && d {
					m.ContestInfo.IsSub = _haveSubscribe
				}
				if d, ok := guessMap[m.ContestInfo.ID]; ok && d {
					m.ContestInfo.IsGuess = _haveGuess
				}
			}
			return
		})
	}
	return
}

func getFavAndGuessContestIds(res *v1.SeriesKnockoutMatchInfo) (contestIDList4FavComponent, contestIDList4GuessComponent []int64) {
	res.Travel(func(m *v1.SeriesKnockoutContestInfoItem) (err error) {
		if m != nil && m.ContestInfo != nil {
			contestIDList4FavComponent = append(contestIDList4FavComponent, m.ContestInfo.ID)
			if m.ContestInfo.GuessType == 1 {
				contestIDList4GuessComponent = append(contestIDList4GuessComponent, m.ContestInfo.ID)
			}
		}
		return
	})
	return
}

func (s *Service) GetSeriesKnockoutMatchInfo(ctx context.Context, req *v1.GetSeriesKnockoutMatchInfoReq) (res *v1.SeriesKnockoutMatchInfo, err error) {
	var cacheFound bool
	res = &v1.SeriesKnockoutMatchInfo{}
	res, cacheFound, err = s.dao.GetSeriesKnockoutMatchInfo(ctx, req.SeriesId)
	if err != nil {
		return
	}
	if !cacheFound {
		//检查该阶段是否存在, 使用缓存避免DB穿透
		var seriesExists bool
		var okToRefreshing bool
		seriesExists, err = s.dao.IsSeriesExists(ctx, req.SeriesId, dao.SeriesTypKnockout)
		if err != nil {
			log.Errorc(ctx, "GetSeriesKnockoutMatchInfo: s.dao.IsSeriesExists error: %v", err)
			return
		}
		if !seriesExists {
			err = ecode.EsportsContestSeriesNotFound
			return
		}
		//尝试标记刷新中状态, 防止缓存失效时, 同时进行多个积分表刷新动作
		okToRefreshing, err = s.dao.MarkSeriesRefreshing(ctx, req.SeriesId, dao.SeriesTypKnockout)
		if err != nil {
			return
		}
		if !okToRefreshing {
			//刷新中, 暂时返回错误
			err = ecode.EsportsContestSeriesNotFound
			return
		}
		res, err = s.RefreshSeriesKnockoutMatchInfo(ctx, &v1.RefreshSeriesKnockoutMatchInfoReq{SeriesId: req.SeriesId})
	}
	if res != nil {
		res.ToBeDeterminedTeamIds = conf.Conf.SeriesIgnoreTeamsIDList
	}
	return
}

func (s *Service) IsSeriesKnockoutMatchInfoGenerated(ctx context.Context, req *v1.IsSeriesKnockoutMatchInfoGeneratedReq) (res *v1.IsSeriesKnockoutMatchInfoGeneratedResp, err error) {
	res = &v1.IsSeriesKnockoutMatchInfoGeneratedResp{}
	_, res.ViewGenerated, err = s.dao.GetSeriesKnockoutMatchInfo(ctx, req.SeriesId)
	return
}
