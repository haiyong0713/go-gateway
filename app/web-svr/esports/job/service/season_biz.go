package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go-common/library/cache/memcache"

	actApi "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/esports/job/component"
	"go-gateway/app/web-svr/esports/job/conf"
	"go-gateway/app/web-svr/esports/job/dao"
	"go-gateway/app/web-svr/esports/job/model"
)

const (
	matchTyeTypeOfLOL = iota + 1

	moreStatusOfAnalysis   = "analysis"
	moreStatusOfSubscribe  = "subscribe"
	moreStatusOfPrediction = "prediction"
	moreStatusOfLive       = "live"
	moreStatusOfReplay     = "replay"
	moreStatusOfCollection = "collection"
	moreStatusOfEnd        = "end"

	moreDisplayOfAnalysis   = "比赛数据"
	moreDisplayOfSubscribe  = "订阅"
	moreDisplayOfPrediction = "预测"
	moreDisplayOfLive       = "直播中"
	moreDisplayOfReplay     = "回放"
	moreDisplayOfCollection = "集锦"
	moreDisplayOfEnd        = "已结束"

	clickStatusOfEnabled  = "enabled"
	clickStatusOfDisabled = "disabled"

	secondsOf10Minutes = 600
)

const (
	teamRegionIDOfNull = iota
	teamRegionIDOfChina
	teamRegionIDOfChinaTaiWan

	teamRegionDisplayOfNull        = "无"
	teamRegionDisplayOfChina       = "中国赛区"
	teamRegionDisplayOfChinaTaiWan = "中国台湾赛区"
)

var (
	teamsOfLOL         atomic.Value
	scoreTeamMapOfLOL  atomic.Value
	contestSeriesOfLOL atomic.Value

	globalContest4FrontendMap map[int64]model.Contest4Frontend
	globalContest2Tab         map[int64]*model.Contest2Tab
)

func init() {
	m := make(map[int64]*model.Team2Tab, 0)
	teamsOfLOL.Store(m)

	scoreTeamM := make(map[int64]*model.Team2Tab, 0)
	scoreTeamMapOfLOL.Store(scoreTeamM)

	series := make(map[int64]*model.ContestSeries, 0)
	contestSeriesOfLOL.Store(series)

	globalContest4FrontendMap = make(map[int64]model.Contest4Frontend, 0)
	globalContest2Tab = make(map[int64]*model.Contest2Tab, 0)
}

func WatchLOLTeams(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchTeamsByMap(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func WatchSeasonPosters(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchSeasonPostersByMap(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func WatchSeasonContests(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			watchSeasonByMap(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func watchTeamsByMap(ctx context.Context) {
	wg := new(sync.WaitGroup)
	for _, v := range conf.LoadSeasonWatchMap() {
		wg.Add(1)
		go func(s *conf.SeasonContestWatch) {
			defer func() {
				wg.Done()
			}()

			updateTeamsByWatchedSeason(ctx, s)
		}(v)
	}

	wg.Wait()
}

func updateTeamsByWatchedSeason(ctx context.Context, s *conf.SeasonContestWatch) {
	if !s.CanWatch() {
		return
	}

	switch s.MatchType {
	case matchTyeTypeOfLOL:
		handleLOLTeamsBySeasonID(ctx, s)
	}
}

func handleLOLTeamsBySeasonID(ctx context.Context, s *conf.SeasonContestWatch) {
	teamIDList, err := dao.SeasonTeamIDList(ctx, s.SeasonID)
	if err != nil {
		return
	}

	if len(teamIDList) > 0 {
		if teamList, err := dao.FetchTeamsByIDs(ctx, teamIDList); err == nil {
			rebuildAndStoreTeams(s, teamList)
		}
	}
}

func rebuildAndStoreTeams(s *conf.SeasonContestWatch, list []*model.Team2Tab) {
	newTeamList := make(map[int64]*model.Team2Tab, 0)
	newScoreTeamList := make(map[int64]*model.Team2Tab, 0)
	teamsOfLastCheckPoint := teamsOfLOL.Load().(map[int64]*model.Team2Tab)
	for k, v := range teamsOfLastCheckPoint {
		tmpTeam := new(model.Team2Tab)
		*tmpTeam = *v
		newTeamList[k] = tmpTeam
	}

	for _, v := range list {
		if v.ScoreTeamID > 0 {
			tmpTeam := new(model.Team2Tab)
			*tmpTeam = *v
			newScoreTeamList[v.ScoreTeamID] = tmpTeam
		}
	}

	for _, team := range list {
		newTeamList[team.ID] = team
	}

	teamsOfLOL.Store(newTeamList)
	scoreTeamMapOfLOL.Store(newScoreTeamList)

	if len(newScoreTeamList) > 0 {
		item := &memcache.Item{Key: s.TeamScoreMapCacheKey, Object: newScoreTeamList, Expiration: s.ExpiredDuration, Flags: memcache.FlagJSON}
		if err := globalMemcache.Set(ctx, item); err != nil {
			fmt.Println("TeamScoreMapCacheKey >>> globalMemcache.Set: ", s.TeamScoreMapCacheKey, err)
			// TODO
		}
	}
}

func watchSeasonPostersByMap(ctx context.Context) {
	wg := new(sync.WaitGroup)
	for _, v := range conf.LoadSeasonWatchMap() {
		wg.Add(1)
		go func(s *conf.SeasonContestWatch) {
			defer func() {
				wg.Done()
			}()

			updatePostersByWatchedSeason(ctx, s)
		}(v)
	}

	wg.Wait()
}

func updatePostersByWatchedSeason(ctx context.Context, s *conf.SeasonContestWatch) {
	if !s.CanWatch() {
		return
	}

	switch s.MatchType {
	case matchTyeTypeOfLOL:
		handleLOLPoster(ctx, s)
	}
}

func handleLOLPoster(ctx context.Context, s *conf.SeasonContestWatch) {
	list, err := dao.S10Poster(ctx)
	if err != nil || len(list) == 0 {
		return
	}

	for _, v := range list {
		tmpContest := model.Contest4Frontend{}
		if d, ok := globalContest4FrontendMap[v.ContestID]; ok {
			tmpContest = d
		}
		v.Contest = tmpContest

		if d, ok := globalContest2Tab[v.ContestID]; ok {
			v.More = genMoreListByStatus(d, d.CalculateStatus(), true)
		}
	}

	newPosterList := new(model.PosterList4S10)
	{
		newPosterList.List = list
		newPosterList.UpdateAt = time.Now().Unix()
	}
	item := &memcache.Item{Key: s.CacheKey4PosterList, Object: newPosterList, Expiration: s.ExpiredDuration, Flags: memcache.FlagJSON}
	if err := globalMemcache.Set(ctx, item); err != nil {
		fmt.Println("handleLOLPoster >>> globalMemcache.Set: ", s.CacheKey4PosterList, err)
		// TODO
	}
}

func watchSeasonByMap(ctx context.Context) {
	wg := new(sync.WaitGroup)
	for _, v := range conf.LoadSeasonWatchMap() {
		wg.Add(1)
		go func(s *conf.SeasonContestWatch) {
			defer func() {
				wg.Done()
			}()

			updateContestsByWatchedSeason(ctx, s)
		}(v)
	}

	wg.Wait()
}

func updateContestsByWatchedSeason(ctx context.Context, s *conf.SeasonContestWatch) {
	if !s.CanWatch() {
		return
	}

	switch s.MatchType {
	case matchTyeTypeOfLOL:
		handleLOLContestsBySeasonID(ctx, s)
	}
}

func handleLOLContestsBySeasonID(ctx context.Context, s *conf.SeasonContestWatch) {
	season, err := dao.SeasonByID(ctx, s.SeasonID)
	if err != nil || season == nil {
		return
	}

	contestList, err := dao.FetchContestsBySeasonID(ctx, s.SeasonID, s.FetchAll)
	if err != nil || len(contestList) == 0 {
		return
	}

	contestIDList := make([]int64, 0)
	for _, v := range contestList {
		contestIDList = append(contestIDList, v.ID)
	}

	updateContestAvCIDListCache(ctx, s, contestIDList)
	updateContestIDListCache(ctx, s, contestIDList)
	updateContestMatchIDListCache(ctx, s, contestList)
	updateContestSeriesMapCache(ctx, s, contestList)

	teamMap := teamsOfLOL.Load().(map[int64]*model.Team2Tab)
	if teamMap == nil || len(teamMap) == 0 {
		return
	}

	cards := generateLOLContestList4Frontend(contestList, teamMap)
	if len(cards) > 0 {
		item := &memcache.Item{Key: s.ContestListCacheKey, Object: cards, Expiration: s.ExpiredDuration, Flags: memcache.FlagJSON}
		if err := globalMemcache.Set(ctx, item); err != nil {
			fmt.Println("handleLOLContestsBySeasonID >>> globalMemcache.Set: ", s.ContestListCacheKey, err)
			// TODO
		}
	}
}

func updateContestMatchIDListCache(ctx context.Context, s *conf.SeasonContestWatch, contestList []*model.Contest2Tab) {
	contestMatchIDM := make(map[int64]int64, 0)
	for _, v := range contestList {
		if v.MatchID > 0 {
			contestMatchIDM[v.MatchID] = v.ID
		}
	}

	if len(contestMatchIDM) > 0 {
		item := &memcache.Item{Key: s.ContestMatchIDMapCacheKey, Object: contestMatchIDM, Expiration: s.ExpiredDuration, Flags: memcache.FlagJSON}
		if err := globalMemcache.Set(ctx, item); err != nil {
			fmt.Println("updateContestMatchIDListCache >>> globalMemcache.Set: ", s.ContestMatchIDMapCacheKey, err)
			// TODO
		}
	}
}

func updateContestSeriesMapCache(ctx context.Context, s *conf.SeasonContestWatch, contestList []*model.Contest2Tab) {
	idList := make([]int64, 0)
	for _, v := range contestList {
		if v.SeriesID > 0 {
			idList = append(idList, v.SeriesID)
		}
	}

	if m, err := dao.FetchContestSeriesList(ctx, idList); err == nil && len(m) > 0 {
		contestSeriesOfLOL.Store(m)
		item := &memcache.Item{Key: s.ContestSeriesMapCacheKey, Object: m, Expiration: s.ExpiredDuration, Flags: memcache.FlagJSON}
		if err := globalMemcache.Set(ctx, item); err != nil {
			fmt.Println("updateContestSeriesMapCache >>> globalMemcache.Set: ", s.ContestSeriesMapCacheKey, err)
			// TODO
		}
	}
}

func updateContestAvCIDListCache(ctx context.Context, s *conf.SeasonContestWatch, contestIDList []int64) {
	if m, err := dao.AvCIDMap(ctx, contestIDList); err == nil && len(m) > 0 {
		item := &memcache.Item{Key: s.ContestAvCIDListCacheKey, Object: m, Expiration: s.ExpiredDuration, Flags: memcache.FlagJSON}
		if err := globalMemcache.Set(ctx, item); err != nil {
			fmt.Println("updateContestAvCIDListCache >>> globalMemcache.Set: ", s.ContestIDListCacheKey, err)
			// TODO
		}
	}
}

func updateContestIDListCache(ctx context.Context, s *conf.SeasonContestWatch, contestIDList []int64) {
	if len(contestIDList) > 0 {
		item := &memcache.Item{Key: s.ContestIDListCacheKey, Object: contestIDList, Expiration: s.ExpiredDuration, Flags: memcache.FlagJSON}
		if err := globalMemcache.Set(ctx, item); err != nil {
			fmt.Println("updateContestIDListCache >>> globalMemcache.Set: ", s.ContestIDListCacheKey, err)
			// TODO
		}
	}
}

func generateLOLContestList4Frontend(contestList []*model.Contest2Tab, teamMap map[int64]*model.Team2Tab) map[int64][]*model.ContestCard {
	contestCardList := make(map[int64][]*model.ContestCard, 0)
	contestM := make(map[int64]model.Contest4Frontend, 0)
	contest2TabM := make(map[int64]*model.Contest2Tab, 0)
	for _, v := range contestList {
		cardList := make([]*model.ContestCard, 0)
		dateUnix := v.StimeDate
		if d, ok := contestCardList[dateUnix]; ok {
			cardList = d
		}

		teamLID := v.HomeID
		teamRID := v.AwayID
		if teamL, ok := teamMap[teamLID]; ok {
			if teamR, ok := teamMap[teamRID]; ok {
				newCard := genContestCardByContestTabAndTwoTeam(v, teamL, teamR)
				cardList = append(cardList, newCard)

				contestCardList[v.StimeDate] = cardList
				contestM[v.ID] = newCard.Contest

				tmpContest2Tab := new(model.Contest2Tab)
				*tmpContest2Tab = *v
				contest2TabM[v.ID] = tmpContest2Tab
			}
		}
	}

	globalContest4FrontendMap = contestM
	globalContest2Tab = contest2TabM

	return contestCardList
}

func genContestSeriesBySeriesID(id int64) (series *model.ContestSeries) {
	series = new(model.ContestSeries)
	if id == 0 {
		return
	}

	if m := contestSeriesOfLOL.Load().(map[int64]*model.ContestSeries); m != nil && len(m) > 0 {
		if d, ok := m[id]; ok {
			series = d
		}
	}

	return
}

func genContestCardByContestTabAndTwoTeam(contest *model.Contest2Tab, teamLeft, teamRight *model.Team2Tab) *model.ContestCard {
	contestCard := new(model.ContestCard)
	{
		contestCard.Contest.ID = contest.ID
		contestCard.Contest.Title = contest.GameStage
		contestCard.Contest.StartTime = contest.Stime
		contestCard.Contest.EndTime = contest.Etime
		contestCard.Contest.Status = contest.CalculateStatus()
		contestCard.Contest.Series = genContestSeriesBySeriesID(contest.SeriesID)
		contestCard.Contest.SeriesID = contest.SeriesID

		home := model.Team4Frontend{}
		{
			home.ID = teamLeft.ID
			home.Icon = teamLeft.Logo
			home.Name = teamLeft.Title
			home.RegionID = teamLeft.RegionID
			home.Region = genTeamRegionDisplayByRegionID(teamLeft.RegionID)
			home.Wins = contest.HomeScore
		}

		away := model.Team4Frontend{}
		{
			away.ID = teamRight.ID
			away.Icon = teamRight.Logo
			away.Name = teamRight.Title
			away.RegionID = teamRight.RegionID
			away.Region = genTeamRegionDisplayByRegionID(teamRight.RegionID)
			away.Wins = contest.AwayScore
		}

		contestCard.Contest.Home = home
		contestCard.Contest.Away = away

		contestCard.More = genMoreListByStatus(contest, contest.CalculateStatus(), false)
	}

	return contestCard
}

func genContestMore4Predict(contestID int64) (more *model.ContestMore) {
	ctx := context.Background()
	req := &actApi.GuessListReq{
		Business: int64(actApi.GuessBusiness_esportsType),
		Oid:      contestID,
	}

	if resp, err := component.ActivityClient.GuessList(ctx, req); err == nil && len(resp.MatchGuess) > 0 {
		more = genMoreByStatus(moreStatusOfPrediction, "", clickStatusOfEnabled)
	}

	return
}

func genMoreListByStatus(contest *model.Contest2Tab, status string, inPoster bool) []*model.ContestMore {
	list := make([]*model.ContestMore, 0)
	switch status {
	case model.ContestStatusOfNotStart:
		if contest.LiveRoom > 0 {
			sub := genMoreByStatus(moreStatusOfSubscribe, "", clickStatusOfEnabled)
			list = append(list, sub)
		}

		if contest.Stime-time.Now().Unix() > secondsOf10Minutes {
			predict := genContestMore4Predict(contest.ID)
			if predict != nil && predict.OnClick != "" {
				list = append(list, predict)
			}
		}
	case model.ContestStatusOfOngoing:
		if contest.LiveRoom > 0 {
			live := genMoreByStatus(moreStatusOfLive, strconv.FormatInt(contest.LiveRoom, 10), clickStatusOfEnabled)
			list = append(list, live)
		}
	case model.ContestStatusOfEnd:
		if contest.PlayBack == "" && contest.CollectionUrl == "" {
			end := genMoreByStatus(moreStatusOfEnd, "", clickStatusOfDisabled)
			list = append(list, end)
		} else {
			if contest.PlayBack != "" {
				replay := genMoreByStatus(moreStatusOfReplay, contest.PlayBack, clickStatusOfEnabled)
				list = append(list, replay)
			}

			if contest.CollectionUrl != "" {
				collection := genMoreByStatus(moreStatusOfCollection, contest.CollectionUrl, clickStatusOfEnabled)
				list = append(list, collection)
			}
		}

		if inPoster && contest.MatchID > 0 {
			analysis := genMoreByStatus(moreStatusOfAnalysis, fmt.Sprintf("%v", contest.MatchID), clickStatusOfEnabled)
			list = append(list, analysis)
		}
	}

	return list
}

func genMoreByStatus(status, link, enabled string) (more *model.ContestMore) {
	more = new(model.ContestMore)
	{
		more.Link = link
		more.OnClick = enabled
	}

	switch status {
	case moreStatusOfAnalysis:
		more.Title = moreDisplayOfAnalysis
		more.Status = moreStatusOfAnalysis
	case moreStatusOfPrediction:
		more.Title = moreDisplayOfPrediction
		more.Status = moreStatusOfPrediction
	case moreStatusOfSubscribe:
		more.Title = moreDisplayOfSubscribe
		more.Status = moreStatusOfSubscribe
	case moreStatusOfCollection:
		more.Title = moreDisplayOfCollection
		more.Status = moreStatusOfCollection
	case moreStatusOfLive:
		more.Title = moreDisplayOfLive
		more.Status = moreStatusOfLive
	case moreStatusOfReplay:
		more.Title = moreDisplayOfReplay
		more.Status = moreStatusOfReplay
	case moreStatusOfEnd:
		more.Title = moreDisplayOfEnd
		more.Status = moreStatusOfEnd
	}

	return
}

func genTeamRegionDisplayByRegionID(regionID int) (display string) {
	display = teamRegionDisplayOfNull

	switch regionID {
	case teamRegionIDOfNull:
		display = teamRegionDisplayOfNull
	case teamRegionIDOfChina:
		display = teamRegionDisplayOfChina
	case teamRegionIDOfChinaTaiWan:
		display = teamRegionDisplayOfChinaTaiWan
	}

	return
}
