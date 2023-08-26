package service

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/ecode"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/dao"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

// AddSeriesPointMatchConfig: 向阶段中添加积分赛配置
func (s *Service) AddSeriesPointMatchConfig(ctx context.Context, req *v1.SeriesPointMatchConfig) (res *v1.AddSeriesPointMatchConfigResp, err error) {
	res = &v1.AddSeriesPointMatchConfigResp{}
	err = s.innerSetSeriesPointMatchConfig(ctx, req, true)
	if err != nil {
		return
	}
	_, err = s.RefreshSeriesPointMatchInfo(ctx, &v1.RefreshSeriesPointMatchInfoReq{
		SeriesId: req.SeriesId,
	})
	return
}

// GetSeriesPointMatchConfig: 获取阶段下的添加积分赛配置
func (s *Service) GetSeriesPointMatchConfig(ctx context.Context, req *v1.GetSeriesPointMatchReq) (res *v1.SeriesPointMatchConfig, err error) {
	return s.dao.GetSeriesPointsMatchConfig(ctx, req.SeriesId)
}

// UpdateSeriesPointMatchConfig: 修改阶段中的积分赛配置
func (s *Service) UpdateSeriesPointMatchConfig(ctx context.Context, req *v1.SeriesPointMatchConfig) (res *v1.UpdateSeriesPointMatchResp, err error) {
	res = &v1.UpdateSeriesPointMatchResp{}
	err = s.innerSetSeriesPointMatchConfig(ctx, req, false)
	if err != nil {
		return
	}
	_, err = s.RefreshSeriesPointMatchInfo(ctx, &v1.RefreshSeriesPointMatchInfoReq{
		SeriesId: req.SeriesId,
	})
	return
}

// innerSetSeriesPointMatchConfig: 向阶段中设置积分赛配置
func (s *Service) innerSetSeriesPointMatchConfig(ctx context.Context, req *v1.SeriesPointMatchConfig, shouldEmpty bool) (err error) {
	defer func() {
		if err != nil {
			log.Errorc(ctx, "innerSetSeriesPointMatchConfig error: %v", err)
		}
	}()
	err = req.ManualValidate()
	if err != nil {
		return
	}
	exist, empty, err := s.dao.IsSeriesExistsAndExtraConfigEmpty(ctx, req.SeriesId, dao.SeriesTypPoint)
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

func (s *Service) PreviewSeriesPointMatchInfo(ctx context.Context, req *v1.SeriesPointMatchConfig) (res *v1.SeriesPointMatchInfo, err error) {
	return s.generatePreviewPointMatchInfo(ctx, req)
}

func (s *Service) RefreshSeriesPointMatchInfo(ctx context.Context, req *v1.RefreshSeriesPointMatchInfoReq) (res *v1.SeriesPointMatchInfo, err error) {
	res = &v1.SeriesPointMatchInfo{}
	config, err := s.dao.GetSeriesPointsMatchConfig(ctx, req.SeriesId)
	if err != nil {
		log.Errorc(ctx, "RefreshSeriesPointMatchInfo: s.dao.GetSeriesPointsMatchConfig error: %v", err)
		if err == ecode.EsportsContestSeriesExtraConfigNotFound {
			// 如果ExtraConfig的配置为空，则不处理后续流程
			err = nil
			return
		}
		err = ecode.EsportsContestSeriesExtraConfigErr
		return
	}
	res, err = s.generatePreviewPointMatchInfo(ctx, config)
	if err != nil {
		return
	}
	err = s.dao.SetSeriesPointMatchInfo(ctx, res)
	return
}

func (s *Service) generatePreviewPointMatchInfo(ctx context.Context, req *v1.SeriesPointMatchConfig) (res *v1.SeriesPointMatchInfo, err error) {
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
	contestInfos, err := s.dao.GetContestsBySeriesId(ctx, req.SeriesId)
	if err != nil {
		return
	}
	return s.innerGeneratePreviewPointMatchInfo(ctx, req, teamInfos, contestInfos)
}

func (s *Service) innerGeneratePreviewPointMatchInfo(ctx context.Context, req *v1.SeriesPointMatchConfig,
	teamInfos map[int64]*model.Team, contestInfos map[int64]*model.Contest) (res *v1.SeriesPointMatchInfo, err error) {
	teamConfigMap := make(map[int64]*v1.SeriesPointMatchTeamConfig, len(req.Teams))
	defer func() {
		if err != nil {
			log.Errorc(ctx, "innerGeneratePreviewPointMatchInfo error: %v", err)
		}
	}()
	groupOrder := make(map[string]int, 0)
	for _, g := range req.Teams {
		_, ok := groupOrder[g.Group]
		if ok {
			continue
		}
		groupOrder[g.Group] = len(groupOrder) + 1
	}
	res = &v1.SeriesPointMatchInfo{
		SeasonId:       req.SeasonId,
		SeriesId:       req.SeriesId,
		UseTeamGroup:   req.UseTeamGroup,
		UseSmallScore:  req.SmallScoreDecrLose != 0 || req.SmallScoreIncrWin != 0,
		TotalTeams:     make([]*v1.SeriesPointMatchTeamListItem, 0),
		GroupTeams:     make([]*v1.SeriesPointMatchGroupList, 0),
		GroupOutletNum: req.GroupOutletNum,
	}
	if len(teamInfos) == 0 {
		return
	}
	for _, teamConfig := range req.Teams {
		t := teamConfig
		teamConfigMap[t.Tid] = t
	}
	getTeamGroup := func(teamId int64) string {
		if !req.UseTeamGroup {
			return ""
		}
		c := teamConfigMap[teamId]
		if c == nil {
			return ""
		}
		return c.Group
	}

	getTeamPriority := func(teamId int64) int64 {
		c := teamConfigMap[teamId]
		if c == nil {
			return 0
		}
		return c.Priority
	}
	tmpTeamScoreMap := make(map[int64]*v1.SeriesPointMatchTeamListItem, len(teamInfos))
	for _, team := range teamInfos {
		if tool.Int64InSlice(team.ID, conf.Conf.SeriesIgnoreTeamsIDList) {
			continue
		}
		tmpTeamScoreMap[team.ID] = &v1.SeriesPointMatchTeamListItem{
			TeamId:     team.ID,
			LeidaTid:   team.LeidaTID,
			Group:      getTeamGroup(team.ID),
			TeamName:   team.Title,
			IconUrl:    team.Logo,
			Score:      0,
			SmallScore: 0,
		}
	}
	//计算战队胜负场数, 得分情况
	{
		for _, contest := range contestInfos {
			//判断赛程是否结束,未结束的不参与计算
			if contest.ContestStatus != ContestStatusEnd {
				continue
			}
			//比分相同, 不纳入计算范围
			if contest.HomeScore == contest.AwayScore {
				continue
			}
			//根据比分判断胜负队伍ID, 胜负队伍小场胜负数
			winTeamId := int64(0)
			winTeamScore := int64(0)
			loseTeamId := int64(0)
			loseTeamScore := int64(0)
			if contest.HomeScore > contest.AwayScore {
				winTeamId = contest.HomeID
				winTeamScore = contest.HomeScore
				loseTeamId = contest.AwayID
				loseTeamScore = contest.AwayScore
			} else {
				winTeamId = contest.AwayID
				winTeamScore = contest.AwayScore
				loseTeamId = contest.HomeID
				loseTeamScore = contest.HomeScore
			}
			//进行加减分,获胜/失败次数操作
			winTeam, ok := tmpTeamScoreMap[winTeamId]
			if ok {
				winTeam.WinTimes += 1
				winTeam.Score += req.ScoreIncrWin
				//根据小场胜利次数小分增加
				winTeam.SmallScore = winTeam.SmallScore + (winTeamScore * req.SmallScoreIncrWin)
				//根据小场失败次数计算小分, 失败队伍的小场获胜数即为胜利队伍的小场失败数
				winTeam.SmallScore = winTeam.SmallScore + (loseTeamScore * req.SmallScoreDecrLose)
			}
			loseTeam, ok := tmpTeamScoreMap[loseTeamId]
			if ok {
				loseTeam.LoseTimes += 1
				loseTeam.Score += req.ScoreDecrLose
				//根据小场胜利次数小分增加
				loseTeam.SmallScore = loseTeam.SmallScore + (loseTeamScore * req.SmallScoreIncrWin)
				//根据小场失败次数计算小分, 获胜队伍的小场获胜数即为失败队伍的小场失败数
				loseTeam.SmallScore = loseTeam.SmallScore + (winTeamScore * req.SmallScoreDecrLose)
			}
		}
		for _, teamScore := range tmpTeamScoreMap {
			t := teamScore
			res.TotalTeams = append(res.TotalTeams, t)
		}
	}

	/*
		计算战队排名,名次排列逻辑:
		首先按照胜场数量，从大到小排序；
		若胜场数量相同则按照积分，从大到小排序；
		若积分相同则按照 胜场减败场的数量，从大到小排序；
		若 胜场减败场的数量相同则按照小分，从大到小排序；
		若小分也相同则名次相同，按照权重，从大到小排序
		若权重也相同, 则根据teamId,从大到小排序
	*/
	{
		sort.Slice(res.TotalTeams, func(i, j int) bool {
			ti := res.TotalTeams[i]
			tj := res.TotalTeams[j]
			if ti.WinTimes == tj.WinTimes { //胜场相同, 进入积分判断
				if ti.Score == tj.Score { //胜场&积分相同, 进入胜场-败场判断
					tiNetWin := ti.WinTimes - ti.LoseTimes
					tjNetWin := tj.WinTimes - tj.LoseTimes
					if tiNetWin == tjNetWin { //胜场&积分&胜场-败场相同, 进入小分判断
						if ti.SmallScore == tj.SmallScore { //胜场&积分&胜场-败场相同&小分判断, 根据权重判断
							tiP := getTeamPriority(ti.TeamId)
							tjP := getTeamPriority(tj.TeamId)
							if tiP == tjP { //优先级也相同, 根据teamId倒序
								return ti.TeamId < tj.TeamId
							}
							return tiP > tjP
						}
						return ti.SmallScore > tj.SmallScore
					}
					return tiNetWin > tjNetWin
				}
				return ti.Score > tj.Score
			}
			return ti.WinTimes > tj.WinTimes
		})
	}

	//构造结果集, 并计算排名. 注意有排名并列的情况
	{
		isTeamInSameRank := func(t1, t2 *v1.SeriesPointMatchTeamListItem) bool {
			return t1.WinTimes == t2.WinTimes &&
				t1.Score == t2.Score &&
				t1.SmallScore == t2.SmallScore &&
				(t1.WinTimes-t1.LoseTimes) == (t2.WinTimes-t2.LoseTimes)
		}

		refreshTeamRank := func(list []*v1.SeriesPointMatchTeamListItem) {
			lastRank := int64(1)
			if len(list) == 0 {
				return
			}
			list[0].Rank = 1
			for i := 1; i < len(list); i++ {
				if isTeamInSameRank(list[i], list[i-1]) {
					list[i].Rank = lastRank
					continue
				}
				lastRank = int64(i) + 1
				list[i].Rank = lastRank
			}
		}
		//使用了分组功能,构造分组内数据,并刷新组内排名
		tmpGroupTeamMap := make(map[string][]*v1.SeriesPointMatchTeamListItem, 0)
		if req.UseTeamGroup {
			for _, team := range res.TotalTeams {
				group := team.Group
				if group == "" {
					continue
				}
				if tmpGroupTeamMap[group] == nil {
					tmpGroupTeamMap[group] = make([]*v1.SeriesPointMatchTeamListItem, 0)

				}
				tmpTeam := team
				tmpGroupTeamMap[group] = append(tmpGroupTeamMap[group], tmpTeam)
			}
			for group, teams := range tmpGroupTeamMap {
				tt := teams
				groupTeams := &v1.SeriesPointMatchGroupList{
					Name:       group,
					GroupTeams: tt,
				}
				refreshTeamRank(groupTeams.GroupTeams)
				res.GroupTeams = append(res.GroupTeams, groupTeams)
			}
			sort.Slice(res.GroupTeams, func(i, j int) bool {
				return groupOrder[res.GroupTeams[i].Name] < groupOrder[res.GroupTeams[j].Name]
			})
			//使用分组时, 无需返回total
			res.TotalTeams = nil
		}

		//未使用分组功能,刷新total的排名
		if !req.UseTeamGroup {
			refreshTeamRank(res.TotalTeams)
		}
	}
	res.RefreshTime = time.Now().Unix()

	return
}

func (s *Service) HttpSeriesPointMatchInfo(ctx context.Context, req *v1.GetSeriesPointMatchInfoReq) (res *model.SeriesPointMatchMore, err error) {
	res = &model.SeriesPointMatchMore{}
	pointMatchInfo, err := s.GetSeriesPointMatchInfo(ctx, req)
	if err != nil {
		return
	}
	if pointMatchInfo == nil {
		return
	}
	season, err := s.getSeasonInfo(ctx, pointMatchInfo.SeasonId)
	if err != nil {
		log.Errorc(ctx, "SeriesPointMatchInfoHttp s.getSeasonInfo() sid(%d) error(%+v)", pointMatchInfo.SeasonId, err)
		return
	}
	res = &model.SeriesPointMatchMore{
		SeriesPointMatchInfo: pointMatchInfo,
		Season:               season,
	}
	return
}

func (s *Service) GetSeriesPointMatchInfo(ctx context.Context, req *v1.GetSeriesPointMatchInfoReq) (res *v1.SeriesPointMatchInfo, err error) {
	var cacheFound bool
	res = &v1.SeriesPointMatchInfo{}
	res, cacheFound, err = s.dao.GetSeriesPointMatchInfo(ctx, req.SeriesId)
	if err != nil {
		return
	}
	if !cacheFound {
		//检查该阶段是否存在, 使用缓存避免DB穿透
		var seriesExists bool
		var okToRefreshing bool
		seriesExists, err = s.dao.IsSeriesExists(ctx, req.SeriesId, dao.SeriesTypPoint)
		if err != nil {
			log.Errorc(ctx, "GetSeriesPointMatchInfo: s.dao.IsSeriesExists error: %v", err)
			return
		}
		if !seriesExists {
			err = ecode.EsportsContestSeriesNotFound
			return
		}
		//尝试标记刷新中状态, 防止缓存失效时, 同时进行多个积分表刷新动作
		okToRefreshing, err = s.dao.MarkSeriesRefreshing(ctx, req.SeriesId, dao.SeriesTypPoint)
		if err != nil {
			return
		}
		if !okToRefreshing {
			//刷新中, 暂时返回错误
			err = ecode.EsportsContestSeriesNotFound
			return
		}
		//检查阶段的积分表配置是否存在
		_, err = s.dao.GetSeriesPointsMatchConfig(ctx, req.SeriesId)
		if err != nil {
			log.Errorc(ctx, "GetSeriesPointMatchInfo: s.dao.IsSeriesExtraConfigEmpty error: %v", err)
			err = ecode.EsportsContestSeriesExtraConfigErr
			return
		}
		res, err = s.RefreshSeriesPointMatchInfo(ctx, &v1.RefreshSeriesPointMatchInfoReq{SeriesId: req.SeriesId})
	}
	return
}

func (s *Service) IsSeriesPointMatchInfoGenerated(ctx context.Context, req *v1.IsSeriesPointMatchInfoGeneratedReq) (res *v1.IsSeriesPointMatchInfoGeneratedResp, err error) {
	res = &v1.IsSeriesPointMatchInfoGeneratedResp{}
	_, res.ViewGenerated, err = s.dao.GetSeriesPointMatchInfo(ctx, req.SeriesId)
	return
}
