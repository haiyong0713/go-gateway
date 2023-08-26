package service

import (
	"context"
	"sort"
	"time"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	actpb "go-gateway/app/web-svr/activity/interface/api"
	xecode "go-gateway/app/web-svr/esports/ecode"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/model"
)

var (
	_emptyGame         = make([]*model.Filter, 0)
	_emptySeason       = make([]*model.Season, 0)
	_emptyCalendar     = make([]*model.Calendar, 0)
	_emptyGuessList    = make([]*actpb.GuessList, 0)
	_emptyContestList  = make([]*model.Contest, 0)
	_emptyGuessContest = make(map[int64]*model.Contest)
)

const (
	_guessBusID     = int64(actpb.GuessBusiness_esportsType)
	_guessStackType = int64(actpb.StakeType_coinType)
	_addFav         = 1
	_isGuess        = 1
	_detailPs       = 6
)

func (s *Service) GuessListByContestID(c context.Context, cid, mid int64) (list []*actpb.GuessList, err error) {
	req := &actpb.GuessListReq{
		Business: _guessBusID,
		Oid:      cid,
		Mid:      mid,
	}

	reply, reqErr := s.actClient.GuessList(c, req)
	if reqErr != nil {
		err = reqErr

		return
	}

	return reply.MatchGuess, nil
}

// GuessDetail 竞猜详情 首页
func (s *Service) GuessDetail(c context.Context, cid, mid int64) (guess *model.GuessDetail, err error) {
	var (
		dbContests map[int64]*model.Contest
		contest    *model.Contest
		ok         bool
		reply      *actpb.GuessListReply
	)
	tmpRs := &model.GuessDetail{
		Guess: _emptyGuessList,
		Stats: &model.GuessTeamStats{
			HomeStats: struct{}{},
			AwayStats: struct{}{},
		},
		Detail: _emptContestDetail,
	}
	group := &errgroup.Group{}
	group.Go(func(ctx context.Context) (contestErr error) {
		if dbContests, contestErr = s.dao.EpContests(c, []int64{cid}); contestErr != nil {
			log.Error("GuessDetails.EpContests cid(%d) error(%v)", cid, contestErr)
			err = nil
			guess = tmpRs
		}
		return nil
	})
	group.Go(func(ctx context.Context) (actGuessErr error) {
		req := &actpb.GuessListReq{
			Business: _guessBusID,
			Oid:      cid,
			Mid:      mid,
		}
		if reply, actGuessErr = s.actClient.GuessList(c, req); actGuessErr != nil {
			log.Error("GuessDetailValue actClient.GuessList Param(%+v) Error(%v)", req, err)
		}
		return nil
	})
	group.Wait()
	if contest, ok = dbContests[cid]; !ok {
		return nil, xecode.EsportsGuessNOTFound
	}
	if contest.GuessType == 0 {
		return nil, xecode.EsportsGuessNOTFound
	}
	defer func() {
		if reply != nil && len(reply.MatchGuess) > 0 {
			guess.Guess = reply.MatchGuess
		}
	}()
	//竞猜数据
	if guess, err = s.dao.GuessDetailCache(c, cid); err != nil || guess == nil {
		rs := s.ContestInfo(c, []int64{cid}, []*model.Contest{contest}, 0)
		if len(rs) == 0 {
			guess = tmpRs
			return
		}
		guess = s.GuessDetailValue(c, rs[0], mid)
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddGuessDetailCache(c, cid, guess)
		})
	}
	return
}

// removeCurrentCid remove current cid
func (s *Service) removeCurrentCid(cids []int64, cid int64) (res []int64) {
	for _, v := range cids {
		if v != cid {
			res = append(res, v)
		}
	}
	return res
}

// GuessMoreMatch 竞猜详情 更多比赛
func (s *Service) GuessMoreMatch(c context.Context, homeID, awayID, sid, cid int64) (res *model.MoreShow, err error) {
	var (
		recentHomeContest, recentAwayContest []*model.Contest
		recentHomeCIDs, recentAwayCIDs       []int64
		recentAwayRes, recentHomeRes         map[int64]*model.Contest
	)
	res = &model.MoreShow{}
	now := time.Now().Unix()
	group := &errgroup.Group{}
	group.Go(func(ctx context.Context) error {
		pRecentHome := &model.ParamContest{
			Tid:    homeID,
			GsRecT: now,
			Ps:     _detailPs,
			Sort:   1,
		}
		if sid != 0 {
			pRecentHome.Sids = []int64{sid}
		}
		if recentHomeCIDs, _, err = s.dao.SearchContestQuery(c, pRecentHome); err != nil {
			log.Error("GuessDetailValue SearchContestQuery Param(%+v) Error(%v)", pRecentHome, err)
			return nil
		}
		recentHomeCIDs = s.removeCurrentCid(recentHomeCIDs, cid)
		if len(recentHomeCIDs) > 0 {
			var (
				tmpRs []*model.Contest
			)
			if recentHomeRes, err = s.dao.EpContests(c, recentHomeCIDs); err != nil {
				log.Error("GuessDetailValue EpContests ID(%v) error(%v)", recentHomeCIDs, err)
				return nil
			}
			for _, cid := range recentHomeCIDs {
				if contest, ok := recentHomeRes[cid]; ok {
					tmpRs = append(tmpRs, contest)
				}
			}
			recentHomeContest = s.ContestInfo(c, recentHomeCIDs, tmpRs, 0)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		pRecentAway := &model.ParamContest{
			Tid:    awayID,
			GsRecT: now,
			Ps:     _detailPs,
			Sort:   1,
		}
		if sid != 0 {
			pRecentAway.Sids = []int64{sid}
		}
		if recentAwayCIDs, _, err = s.dao.SearchContestQuery(c, pRecentAway); err != nil {
			log.Error("GuessDetailValue SearchContestQuery Param(%+v) Error(%v)", pRecentAway, err)
			return nil
		}
		recentAwayCIDs = s.removeCurrentCid(recentAwayCIDs, cid)
		if len(recentAwayCIDs) > 0 {
			var (
				tmpRs []*model.Contest
			)
			if recentAwayRes, err = s.dao.EpContests(c, recentAwayCIDs); err != nil {
				log.Error("GuessDetailValue  EpContests ID(%d) error(%v)", recentAwayCIDs, err)
				return nil
			}
			for _, cid := range recentAwayCIDs {
				if contest, ok := recentAwayRes[cid]; ok {
					tmpRs = append(tmpRs, contest)
				}
			}
			recentAwayContest = s.ContestInfo(c, recentAwayCIDs, tmpRs, 0)
		}
		return nil
	})
	group.Wait()
	if len(recentHomeRes) > 0 {
		res.Home = recentHomeContest
	}
	if len(recentAwayRes) > 0 {
		res.Away = recentAwayContest
	}
	return
}

// GuessDetailValue .
func (s *Service) GuessDetailValue(c context.Context, contest *model.Contest, mid int64) (guess *model.GuessDetail) {
	var (
		contestData []*model.ContestsData
		err         error
	)
	season := &model.Season{}
	guess = &model.GuessDetail{
		Contest: contest,
		Guess:   _emptyGuessList,
		Stats: &model.GuessTeamStats{
			HomeStats: struct{}{},
			AwayStats: struct{}{},
		},
		Detail: _emptContestDetail,
	}

	if s, ok := contest.Season.(*model.Season); ok {
		season = s
	}
	group := &errgroup.Group{}
	group.Go(func(ctx context.Context) error {
		var (
			homeTeam *model.Team
		)
		if v, ok := contest.HomeTeam.(*model.Team); ok {
			homeTeam = v
		}
		if homeTeam != nil {
			pStatHome := &model.ParamSpecial{
				ID:       homeTeam.LeidaTID,
				LeidaSID: season.LeidaSID,
				Tp:       season.GameType,
				Recent:   0,
			}
			log.Info("GuessDetailValue  pStatHome param(%+v)", pStatHome)
			guess.Stats.HomeStats = s.teamStats(c, pStatHome)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		var (
			awayTeam *model.Team
		)
		if v, ok := contest.AwayTeam.(*model.Team); ok {
			awayTeam = v
		}
		if awayTeam != nil {
			pStatAway := &model.ParamSpecial{
				ID:       awayTeam.LeidaTID,
				LeidaSID: season.LeidaSID,
				Tp:       season.GameType,
				Recent:   0,
			}
			log.Info("GuessDetailValue  pStatAway param(%+v)", pStatAway)
			guess.Stats.AwayStats = s.teamStats(c, pStatAway)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if contestData, _ = s.dao.ContestData(ctx, contest.ID); err != nil {
			log.Error("GuessDetailValue ContestData error(%v)", contest.ID)
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		return
	}
	if len(contestData) > 0 {
		guess.Detail = contestData
	}
	return
}

// GuessDetailCoin 竞猜详情 用户投币记录
func (s *Service) GuessDetailCoin(c context.Context, mID, mainID int64) (reply *actpb.GuessUserGroup, err error) {
	req := &actpb.UserGuessGroupReq{
		Mid:      mID,
		Business: _guessBusID,
		MainId:   mainID,
	}
	if reply, err = s.actClient.UserGuessGroup(c, req); err != nil {
		log.Error("GuessDetailCoin actClient.UserGuessGroup Param(%+v) Error(%v)", req, err)
		return
	}
	return
}

// AddGuessDetail add guess
func (s *Service) AddGuessDetail(c context.Context, param *model.AddGuessParam) (err error) {
	var (
		contest  *model.Contest
		contests map[int64]*model.Contest
		ok       bool
	)
	if contests, err = s.dao.EpContests(c, []int64{param.OID}); err != nil {
		log.Error("AddGuessDetail EpContests Param(%+v) Error(%v)", param, err)
		return
	}
	if len(contests) == 0 {
		return xecode.EsportsGuessNOTFound
	}
	if contest, ok = contests[param.OID]; !ok || contest == nil {
		return xecode.EsportsGuessNOTFound
	}
	if contest.GuessType == 0 {
		return xecode.EsportsGuessNOTFound
	}
	//提前十分钟结束竞猜
	if (time.Now().Unix() + 10*60) > contest.Stime {
		return xecode.EsportsGuessEndErr
	}
	req := &actpb.GuessUserAddReq{
		Mid:       param.MID,
		MainID:    param.MainID,
		DetailID:  param.DetailID,
		StakeType: _guessStackType,
		Stake:     param.Count,
	}
	if _, err = s.actClient.GuessUserAdd(c, req); err != nil {
		log.Error("AddGuessDetail actClient.GuessUserAdd Param(%+v) Error(%v)", req, err)
		return err
	}

	if tmpErr := s.DeleteUserSeasonGuessListByMatchID(c, param.MID, param.OID); tmpErr != nil {
		log.Errorc(c, "DeleteUserGuessListBySeasonID failed: %v, mid: %v, matchID: %v", tmpErr, param.MID, param.OID)
	}

	if err = s.dao.DelGuessDetailCache(c, param.OID); err != nil {
		log.Error("AddGuessDetail DelGuessDetailCache Param(%+v) Error(%v)", req, err)
	}
	if param.IsFav == _addFav {
		s.cache.Do(c, func(ctx context.Context) {
			if err = s.AddFav(context.Background(), param.MID, param.OID); err != nil {
				log.Error("AddGuessDetail AddFav Param(%+v) Error(%v)", param, err)
			}
		})
	}
	return nil
}

// GuessCollCalendar guess collection calendar 竞猜合集页 日历模块
func (s *Service) GuessCollCalendar(c context.Context, p *model.ParamContest) (res []*model.Calendar, err error) {
	var (
		cids       []int64
		dbContests map[int64]*model.Contest
	)
	p.GsType = _isGuess
	if cids, _, err = s.dao.SearchContestQuery(c, p); err != nil {
		log.Error("s.dao.SearchContest error(%v)", err)
	}
	if len(cids) == 0 {
		res = _emptyCalendar
		return
	}
	if dbContests, err = s.dao.EpContests(c, cids); err != nil {
		log.Error("s.dao.Contest error(%v)", err)
		return
	}
	timeMap := make(map[string]int64)
	for _, v := range dbContests {
		s := time.Unix(v.Stime, 0).Format("2006-01-02")
		timeMap[s]++
	}
	for k, v := range timeMap {
		calendar := &model.Calendar{
			Stime: k,
			Count: v,
		}
		res = append(res, calendar)
	}
	return
}

// GuessCollGS guess season and game 竞猜合集页 游戏 赛季
func (s *Service) GuessCollGS(c context.Context, gid int64) (res *model.GuessCollection, err error) {
	var (
		guessGameIDs, guessSeasonIDs                    []int64
		gameErr, guessGameErr, guessSeasonErr, calenErr error
		games, gamesRes                                 []*model.Filter
		gamesMap                                        map[int64]*model.Filter
		seasonMap                                       map[int64]*model.Season
		season                                          []*model.Season
		calendar                                        []*model.Calendar
	)
	addCache := true
	if res, err = s.dao.GuessCollecCache(c, gid); err != nil {
		addCache = false
		err = nil
		log.Error("Service.GuessCollGS.GuessCollecCache error(%v)", guessGameErr)
	}
	if res != nil {
		return
	}
	res = &model.GuessCollection{}
	group := &errgroup.Group{}
	group.Go(func(ctx context.Context) error {
		if guessGameIDs, guessGameErr = s.dao.GuessCollGame(ctx); guessGameErr != nil {
			log.Error("Service.GuessCollGS error(%v)", guessGameErr)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if games, gameErr = s.dao.Games(ctx); gameErr != nil {
			log.Error("Service.Games error(%v)", gameErr)
			return nil
		}
		gamesMap = make(map[int64]*model.Filter)
		for _, v := range games {
			gamesMap[v.ID] = v
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if guessSeasonIDs, guessSeasonErr = s.dao.GuessCollSeason(ctx, gid); guessSeasonErr != nil {
			log.Error("Service.GuessCollSeason error(%v)", guessSeasonErr)
			return nil
		}
		var (
			sidUnique []int64
		)
		sidMap := make(map[int64]bool)
		for _, v := range guessSeasonIDs {
			if sidMap[v] {
				continue
			}
			sidUnique = append(sidUnique, v)
			sidMap[v] = true
		}
		if len(guessSeasonIDs) == 0 {
			return nil
		}
		if seasonMap, err = s.dao.EpSeasons(c, sidUnique); err != nil {
			err = nil
			seasonMap = make(map[int64]*model.Season)
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if calendar, calenErr = s.dao.GuessCCalen(ctx); calenErr != nil {
			log.Error("Service.GuessCCalen error (%v)", guessSeasonErr)
		}
		return nil
	})
	group.Wait()
	if len(gamesMap) > 0 && len(guessGameIDs) > 0 {
		for _, id := range guessGameIDs {
			if v, ok := gamesMap[id]; ok {
				gamesRes = append(gamesRes, v)
			}
		}
	}
	for _, v := range seasonMap {
		season = append(season, v)
	}
	if len(season) == 0 {
		season = _emptySeason
	}
	if len(gamesRes) == 0 {
		gamesRes = _emptyGame
	}
	if len(calendar) == 0 {
		calendar = _emptyCalendar
	}
	res.Season = season
	res.Game = gamesRes
	if !addCache {
		return
	}
	s.cache.Do(c, func(ctx context.Context) {
		s.dao.AddGuessCollecCache(ctx, gid, res)
	})
	return res, nil
}

// GuessCollQues guess collection contest 竞猜合集页 竞猜问题
func (s *Service) GuessCollQues(c context.Context, p *model.ParamContest, mid int64) (res []*model.GuessCollQues, total int, err error) {
	var (
		cids                 []int64
		dbContests           map[int64]*model.Contest
		contests             []*model.Contest
		guessLists           *actpb.GuessListsReply
		guessQues            map[int64]*actpb.GuessListReply
		contestErr, guessErr error
	)
	res = make([]*model.GuessCollQues, 0)
	group := &errgroup.Group{}
	p.GsType = _isGuess
	if p.Stime == "" {
		p.Stime = time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05")
	}
	if cids, total, err = s.dao.SearchContestQuery(c, p); err != nil {
		log.Error("s.GuessCollQues.SearchContestQuery Param(%v) error(%v)", p, err)
		err = nil
		return
	}
	if total == 0 || len(cids) == 0 {
		return
	}
	group.Go(func(ctx context.Context) error {
		var (
			tmpRs []*model.Contest
		)
		if dbContests, contestErr = s.dao.EpContests(c, cids); contestErr != nil {
			log.Error("s.GuessCollQues.EpContests Param(%v) error(%v)", cids, contestErr)
			contestErr = nil
			return nil
		}
		for _, cid := range cids {
			if contest, ok := dbContests[cid]; ok {
				tmpRs = append(tmpRs, contest)
			}
		}
		contests = s.ContestInfo(c, cids, tmpRs, 0)
		return nil
	})
	group.Go(func(ctx context.Context) error {
		req := &actpb.GuessListsReq{
			Business: _guessBusID,
			Oids:     cids,
			Mid:      mid,
		}
		if guessLists, guessErr = s.actClient.GuessLists(c, req); guessErr != nil {
			log.Error("s.GuessCollQues.GuessLists Param(%v) error(%v)", req, guessErr)
			guessErr = nil
		}
		return guessErr
	})
	group.Wait()
	if guessLists != nil {
		guessQues = guessLists.MatchGuesses
	}
	if len(dbContests) == 0 {
		return
	}
	for _, contest := range contests {
		if contest == nil {
			continue
		}
		if contest.Stime < time.Now().Unix() {
			continue
		}
		q := &model.GuessCollQues{
			Contest:   contest,
			Questions: _emptyGuessList,
		}
		if ques, ok := guessQues[contest.ID]; ok {
			q.Questions = ques.MatchGuess
		}
		res = append(res, q)
	}
	return
}

// GuessCollStatis .
func (s *Service) GuessCollStatis(c context.Context, mid int64) (res *model.GuessCollUser, err error) {
	res = &model.GuessCollUser{}
	req := &actpb.UserGuessDataReq{
		Mid:       mid,
		StakeType: _guessStackType,
		Business:  _guessBusID,
	}
	if res.GuessData, err = s.actClient.UserGuessData(c, req); err != nil {
		log.Error("GuessCollStatis UserGuessList Param(%v) error(%v)", req, err)
		return nil, err
	}
	return
}

// GuessCollRecord guess collection record 竞猜合集页 用户竞猜记录
func (s *Service) GuessCollRecord(c context.Context, param *model.GuessCollRecoParam) (res *model.GuessCollRecoRes, err error) {
	var (
		guessList  *actpb.UserGuessListReply
		cids       []int64
		tmpRs      []*model.Contest
		guessMap   map[int64][]*actpb.GuessUserGroup
		contestMap map[int64]*model.Contest
	)
	res = &model.GuessCollRecoRes{
		Page: &actpb.PageInfo{
			Pn:    param.Pn,
			Ps:    param.Ps,
			Total: 0,
		},
	}
	req := &actpb.UserGuessListReq{
		Business: _guessBusID,
		Mid:      param.Mid,
		Ps:       param.Ps,
		Pn:       param.Pn,
		Status:   param.Type,
	}
	if guessList, err = s.actClient.UserGuessList(c, req); err != nil {
		log.Error("GuessCollRecord.UserGuessList Param(%v) error(%v)", req, err)
		return
	}
	if guessList != nil {
		guessMap = make(map[int64][]*actpb.GuessUserGroup)
		for _, v := range guessList.UserGroup {
			cids = append(cids, v.Oid)
			guessMap[v.Oid] = append(guessMap[v.Oid], v)
		}
		var (
			dbContests map[int64]*model.Contest
			records    []*model.GuessCollReco
		)
		if dbContests, err = s.dao.EpContests(c, cids); err != nil {
			log.Error("GuessCollRecord.EpContests Param(%v) error(%v)", cids, err)
			return
		}
		for _, cid := range cids {
			if contest, ok := dbContests[cid]; ok {
				tmpRs = append(tmpRs, contest)
			}
		}
		contests := s.ContestInfo(c, cids, tmpRs, 0)
		contestMap = make(map[int64]*model.Contest)
		for _, contest := range contests {
			contestMap[contest.ID] = contest
		}
		for k, v := range guessMap {
			if contest, ok := contestMap[k]; ok {
				record := &model.GuessCollReco{
					Contest:     contest,
					Guess:       v,
					ContestRank: contest.Etime,
					ContestID:   contest.ID,
				}
				records = append(records, record)
			}
		}
		sort.Slice(records, func(i, j int) bool {
			if records[i].ContestRank != records[j].ContestRank {
				return records[i].ContestRank > records[j].ContestRank
			}
			return records[i].ContestID > records[j].ContestID
		})
		res.Page = guessList.Page
		res.GuessCollReco = records
	}
	return
}

// GuessTeamRecent .
func (s *Service) GuessTeamRecent(c context.Context, param *model.ParamEsGuess) (res []*model.Contest, err error) {
	var allTeamContests []*model.Contest
	res = _emptContest
	if allTeamContests, err = fetchHomeAwayContestList(c, param); err != nil {
		log.Errorc(c, "GuessTeamRecent fetchHomeAwayContestList param(%+v) error(%+v)", param, err)
		return
	}
	count := len(allTeamContests)
	if count == 0 {
		return
	}
	seasonsMap := s.GetAllSeasonsOfComponent(c)
	allTeamMap := s.GetAllTeamsOfComponent(c)
	for index, contest := range allTeamContests {
		if index == param.Ps {
			return
		}
		contest.ContestFreeze = contest.Status
		contest.Season = struct{}{}
		contest.HomeTeam = struct{}{}
		contest.AwayTeam = struct{}{}
		if seasonInfo, ok := seasonsMap[contest.Sid]; ok {
			contest.Season = seasonInfo
		}
		if homeTeam, ok := allTeamMap[contest.HomeID]; ok {
			contest.HomeTeam = homeTeam
		}
		if awayTeam, ok := allTeamMap[contest.AwayID]; ok {
			contest.AwayTeam = awayTeam
		}
		contest.StartTime = contest.Stime
		// 赛程结束时间
		contest.EndTime = contest.Etime
		// 赛程比赛阶段
		contest.Title = contest.GameStage
		// 回播房间号url
		contest.PlayBackV2 = contest.Playback
		// 赛季id
		contest.SeasonID = contest.Sid
		// 主队
		home, _ := allTeamMap[contest.HomeID]
		away, _ := allTeamMap[contest.AwayID]
		contest.Home = convertComponentTeam2Card(home)
		contest.Away = convertComponentTeam2Card(away)
		contest.Series = &v1.ContestSeriesComponent{
			ID:          contest.SeriesID,
			ParentTitle: "",
			ChildTitle:  "",
			StartTime:   0,
			EndTime:     0,
		}
		res = append(res, contest)
	}
	return
}

// GuessMatchRecord guess match record 单场比赛竞猜记录
func (s *Service) GuessMatchRecord(c context.Context, mid, oid int64) (res *model.GuessMatchReco, err error) {
	var (
		guessList *actpb.UserGuessMatchReply
	)
	res = &model.GuessMatchReco{
		Guess: make([]*actpb.GuessUserGroup, 0),
	}
	req := &actpb.UserGuessMatchReq{
		Business: _guessBusID,
		Mid:      mid,
		Oid:      oid,
	}
	if guessList, err = s.actClient.UserGuessMatch(c, req); err != nil {
		log.Error("GuessMatchRecord.UserGuessMatch Param(%v) error(%v)", req, err)
		return res, nil
	}
	if guessList != nil {
		res.Guess = guessList.UserGroup
	}
	return
}
