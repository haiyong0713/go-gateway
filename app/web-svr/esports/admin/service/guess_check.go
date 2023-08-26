package service

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/admin/client"
	"go-gateway/app/web-svr/esports/admin/model"
	v1 "go-gateway/app/web-svr/esports/service/api/v1"
	"sort"
	"time"
)

const (
	_defaultDateTemplate = "20060102"
	_defaultTimeTemplate = "2006-01-02 15:04:05"
	_guessBusID          = int64(api.GuessBusiness_esportsType)
	_defaultNoGuess      = "无竞猜"
)

var guessStatusMapping map[int64]string = map[int64]string{
	0: "未结算",
	1: "结算中",
	2: "已结算",
	3: "已流局",
}

var joinGuessStatusMapping map[int64]string = map[int64]string{
	0: "未参与",
	1: "已参与",
}

// GetValidSeasonsByDate 获取日期格式如："2021-01-07"当天的比赛赛季
func (s *Service) GetValidSeasonsByDate(ctx context.Context, date string) (seasonForSelect *model.SeasonSelect, err error) {
	beginTime, err := formatTimeByDate(date)
	if err != nil {
		return
	}
	endTime := beginTime + 86399
	resp, err := client.EsportsServiceClient.GetSeasonByTime(ctx, &v1.GetSeasonByTimeReq{
		BeginTime: beginTime,
		EndTime:   endTime,
	})
	if err != nil {
		return
	}
	seasons := resp.Seasons
	seasonForSelect = new(model.SeasonSelect)
	seasonForSelect.Options = make([]*model.SeasonOption, 0)
	for _, season := range seasons {
		seasonForSelect.Options = append(seasonForSelect.Options,
			&model.SeasonOption{
				Label: season.Title,
				Value: season.ID,
			},
		)
	}
	return
}

func (s *Service) GetSeasonDateContestGuess(ctx context.Context, date string, seasonId int64, mid int64, skipNotJoin bool) (tableList *model.GuessContestTable, err error) {
	tableList = new(model.GuessContestTable)
	beginTime, err := formatTimeByDate(date)
	if err != nil {
		return
	}
	endTime := beginTime + 86399
	resp, err := client.EsportsServiceClient.GetSeasonContests(ctx, &v1.GetSeasonContestsReq{
		SeasonId: seasonId,
	})
	if err != nil {
		return
	}
	guessContests := make([]int64, 0)
	if resp == nil || len(resp.Contests) == 0 {
		return
	}
	contests := make([]*v1.ContestDetail, 0)
	for _, contest := range resp.Contests {
		if contest.Stime < beginTime || contest.Stime > endTime {
			continue
		}
		contests = append(contests, contest)
		if contest.IsGuessed != v1.GuessStatusEnum_HasNoGuess {
			guessContests = append(guessContests, contest.ID)
		}
	}
	sort.Slice(contests, func(i, j int) bool {
		return contests[i].Stime < contests[j].Stime
	})

	// 获取竞猜详情
	guessMap, userGroups, err := s.getGuessInfo(ctx, guessContests, mid)
	if err != nil {
		return
	}

	mainResMap := make(map[int64]*api.GuessUserGroup)
	for _, userGroup := range userGroups {
		mainResMap[userGroup.MainID] = userGroup
	}
	outputContests := make([]*model.GuessContest, 0)
	for _, contest := range contests {
		contestGuessInfo, ok := guessMap[contest.ID]
		if !ok {
			if !skipNotJoin {
				outputContests = append(outputContests, formatDefaultContestGuess(contest, nil, mid))
			}
			continue
		}
		for _, guessInfo := range contestGuessInfo.MatchGuess {
			if _, hit := mainResMap[guessInfo.Id]; !hit {
				if !skipNotJoin {
					outputContests = append(outputContests, formatDefaultContestGuess(contest, guessInfo, mid))
				}
				continue
			}
			outputContests = append(outputContests,
				&model.GuessContest{
					Id:                    contest.ID,
					Mid:                   mid,
					ContestInfo:           formatContestInfo(contest),
					GuessInfo:             formatGuessInfo(guessInfo),
					GuessStatus:           getGuessStatus(mainResMap[guessInfo.Id]),
					JoinStatus:            joinGuessStatusMapping[guessInfo.IsGuess],
					ResultOption:          guessInfo.RightOption,
					JoinOption:            mainResMap[guessInfo.Id].Option,
					SettlementStatus:      getGuessStatus(mainResMap[guessInfo.Id]),
					SettlementStatusCache: formatSettlementStatusCache(mainResMap[guessInfo.Id]),
					JoinNum:               formatJoinNum(mainResMap[guessInfo.Id]),
					Income:                formatIncome(mainResMap[guessInfo.Id]),
				})
		}
	}
	tableList.Items = outputContests
	return
}

func (s *Service) getGuessInfo(ctx context.Context, contestIds []int64, mid int64) (guessList map[int64]*api.GuessListReply, guessGroups []*api.GuessUserGroup, err error) {
	guessList = make(map[int64]*api.GuessListReply)
	guessGroups = make([]*api.GuessUserGroup, 0)
	if mid == 0 || len(contestIds) == 0 {
		return
	}
	reply, err := client.ActivityServiceClient.GuessLists(ctx, &api.GuessListsReq{
		Business: _guessBusID,
		Oids:     contestIds,
		Mid:      mid,
	})
	if err != nil {
		return
	}
	// 获取用户竞猜记录
	groupRes, err := client.ActivityServiceClient.UserGuessMatchs(ctx, &api.UserGuessMatchsReq{
		Business: _guessBusID,
		Mid:      mid,
		Oids:     contestIds,
		Ps:       100,
		Pn:       1,
	})
	if err != nil {
		return
	}
	guessList = reply.MatchGuesses
	guessGroups = groupRes.UserGroup
	return
}

func formatJoinNum(guessGroup *api.GuessUserGroup) (joinNum string) {
	joinNum = ""
	if guessGroup == nil {
		return
	}
	return fmt.Sprintf("%d%s", guessGroup.Stake, "硬币")
}

func formatIncome(guessGroup *api.GuessUserGroup) (income string) {
	income = ""
	if guessGroup == nil {
		return
	}
	income = fmt.Sprintf("%.1f%s", guessGroup.Income, "硬币")
	return
}

func formatSettlementStatusCache(guessGroup *api.GuessUserGroup) (statusStr string) {
	statusStr = guessStatusMapping[0]
	if guessGroup == nil {
		return
	}
	if guessGroup.IsDeleted == 1 {
		statusStr = guessStatusMapping[3]
		return
	}
	if guessGroup.Status == 1 {
		statusStr = guessStatusMapping[2]
		return
	}
	return
}

func formatDefaultContestGuess(contest *v1.ContestDetail, guessInfo *api.GuessList, mid int64) (defaultContest *model.GuessContest) {
	return &model.GuessContest{
		Id:                    contest.ID,
		Mid:                   mid,
		ContestInfo:           formatContestInfo(contest),
		GuessInfo:             formatGuessInfo(guessInfo),
		GuessStatus:           getDefaultGuessStatus(guessInfo),
		JoinStatus:            joinGuessStatusMapping[0],
		ResultOption:          formatRightOption(guessInfo),
		JoinOption:            "",
		SettlementStatus:      "",
		SettlementStatusCache: "",
		JoinNum:               "",
		Income:                "",
	}
}

func getDefaultGuessStatus(guessInfo *api.GuessList) (guessStatus string) {
	guessStatus = "未配置"
	if guessInfo != nil {
		if guessInfo.RightOption == "" {
			guessStatus = guessStatusMapping[0]
		} else {
			guessStatus = guessStatusMapping[2]
		}
	}
	return
}

func formatRightOption(guessInfo *api.GuessList) (option string) {
	option = ""
	if guessInfo != nil {
		option = guessInfo.RightOption
	}
	return
}

func formatGuessInfo(guessBaseInfo *api.GuessList) (guessInfo string) {
	guessInfo = _defaultNoGuess
	if guessBaseInfo == nil {
		return
	}
	guessInfo = fmt.Sprintf("%s (mainId:%d)", guessBaseInfo.Title, guessBaseInfo.Id)
	for index, option := range guessBaseInfo.Details {
		guessInfo = fmt.Sprintf("%s\n%s", guessInfo, fmt.Sprintf("选项%d：%s", index+1, option.Option))
	}
	return guessInfo
}

func formatContestInfo(contest *v1.ContestDetail) (contestInfo string) {
	startTime := time.Unix(contest.Stime, 0).Format(_defaultTimeTemplate)
	contestInfo = fmt.Sprintf("比赛开始时间：%s\n", startTime)
	if contest.HomeTeam != nil && contest.AwayTeam != nil {
		contestInfo = fmt.Sprintf("%s比赛信息：%s VS", contestInfo, contest.HomeTeam.Title)
	}
	if contest.AwayTeam != nil {
		contestInfo = fmt.Sprintf("%s %s", contestInfo, contest.AwayTeam.Title)
	}
	var contestStatus string
	switch contest.ContestStatus {
	case v1.ContestStatusEnum_Ing:
		contestStatus = "进行中"
	case v1.ContestStatusEnum_Over:
		contestStatus = "已结束"
	default:
		contestStatus = "未开始"
	}
	contestInfo = fmt.Sprintf("%s\n比赛状态：%s", contestInfo, contestStatus)
	return
}

func formatTimeByDate(date string) (timestamp int64, err error) {
	dateTime, err := time.ParseInLocation(_defaultDateTemplate, date, time.Local)
	if err != nil {
		return
	}
	timestamp = dateTime.Unix()
	return
}

func getGuessStatus(guessGroup *api.GuessUserGroup) (statusStr string) {
	statusStr = guessStatusMapping[0]
	if guessGroup == nil {
		return
	}
	if guessGroup.IsDeleted == 1 {
		statusStr = guessStatusMapping[3]
		return
	}
	if guessGroup.ResultId != 0 {
		statusStr = guessStatusMapping[2]
		return
	}
	return
}
