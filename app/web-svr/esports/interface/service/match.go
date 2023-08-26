package service

import (
	"context"
	"sort"
	"strconv"
	"time"

	livexroom "git.bilibili.co/bapis/bapis-go/live/xroom"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup"
	egV2 "go-common/library/sync/errgroup.v2"
	arcmdl "go-gateway/app/app-svr/archive/service/api"
	actmdl "go-gateway/app/web-svr/activity/interface/api"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/bvav"
	"go-gateway/app/web-svr/esports/interface/component"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
	"go-gateway/pkg/idsafe/bvid"
)

const (
	_gameNoSub = 6
	_gameSub   = 3
	_gameIn    = 5
	_gameLive  = 4
	_gameOver  = 1
	_caleDay   = 3
	_h5        = 1
	_typeMatch = "matchs"
	_typeGame  = "games"
	_typeTeam  = "teams"
	_typeYear  = "years"
	_typeTag   = "tags"
	_downline  = 0
	_guessOk   = 1
	regAVID    = `^http(s|)://www.bilibili.com/video/av([0-9]+).*$`
	regBVID    = `^*/video/([a-zA-Z1-9]+)(/)?(\?[\s\S]*)?$`
	_hotName   = "热门推荐"

	cacheKey4MaxContestID               = "contest:max_id"
	bizName4ContestDataPageOfResetCache = "contest_data_page"
)

const (
	ContestStatusNotStart = iota + 1
	ContestStatusOngoing
	ContestStatusEnd
)

var (
	_emptContest       = make([]*model.Contest, 0)
	_emptCalendar      = make([]*model.Calendar, 0)
	_emptFilter        = make([]*model.Filter, 0)
	_emptVideoList     = make([]*arcmdl.Arc, 0)
	_emptSeason        = make([]*model.Season, 0)
	_emptLdTeams       = make([]*model.LdTeam, 0)
	_emptContestDetail = make([]*model.ContestsData, 0)
	_emptSearchCard    = make([]*model.SearchRes, 0)
	_emptGameRank      = make([]*model.GameRank, 0)
	_emptSeasonRank    = make([]*model.SeasonRank, 0)

	maxContestID int64
)

func ASyncUpdateMaxContestID(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			var d int64
			if err := component.GlobalMemcached.Get(context.Background(), cacheKey4MaxContestID).Scan(&d); err == nil && d > 0 {
				maxContestID = d
			}
		case <-ctx.Done():
			return
		}
	}
}

// FilterMatch filter match.
func (s *Service) FilterMatch(c context.Context, p *model.ParamFilter) (rs map[string][]*model.Filter, err error) {
	var (
		tmpRs                map[string][]*model.Filter
		fm                   *model.FilterES
		fMap                 map[string]map[int64]*model.Filter
		matchs, games, teams []*model.Filter
	)
	isAll := p.Tid == 0 && p.Gid == 0 && p.Mid == 0 && p.Stime == ""
	if rs, err = s.dao.FMatCache(c); err != nil {
		err = nil
	}
	if isAll && len(rs) > 0 {
		return
	}
	matchs, games, teams = s.filterLeft()
	tmpRs = make(map[string][]*model.Filter, 3)
	tmpRs[_typeMatch] = matchs
	tmpRs[_typeGame] = games
	tmpRs[_typeTeam] = teams
	if fm, err = s.dao.FilterMatch(c, p); err != nil {
		log.Error("s.dao.FilterMatch error(%v)", err)
		return
	}
	fMap = s.filterMap(tmpRs)
	tmpRs = s.fmtES(fm, fMap)
	rs = make(map[string][]*model.Filter, 3)
	rs[_typeMatch] = tmpRs[_typeMatch]
	rs[_typeGame] = tmpRs[_typeGame]
	rs[_typeTeam] = tmpRs[_typeTeam]
	if isAll {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetFMatCache(c, rs)
		})
	} else {
		if len(rs[_typeMatch]) == 0 && len(rs[_typeGame]) == 0 && len(rs[_typeTeam]) == 0 {
			if tmpRs, err = s.dao.FMatCache(c); err != nil {
				err = nil
			}
			if len(tmpRs) > 0 {
				rs = tmpRs
			}
		}
	}
	return
}

func (s *Service) filterLeft() (matchs, games, teams []*model.Filter) {
	var (
		matchErr, gameErr, teamErr error
	)
	group := &errgroup.Group{}
	group.Go(func() error {
		if matchs, matchErr = s.dao.Matchs(context.Background()); matchErr != nil {
			log.Error("s.dao.Matchs error (%v)", matchErr)
		}
		return nil
	})
	group.Go(func() error {
		if games, gameErr = s.dao.Games(context.Background()); gameErr != nil {
			log.Error("s.dao.Games error (%v)", gameErr)
		}
		return nil
	})
	group.Go(func() error {
		if teams, teamErr = s.dao.Teams(context.Background()); teamErr != nil {
			log.Error("s.dao.Teams error (%v)", teamErr)
		}
		return nil
	})
	group.Wait()
	if len(matchs) == 0 {
		matchs = _emptFilter
	}
	if len(games) == 0 {
		games = _emptFilter
	}
	if len(teams) == 0 {
		teams = _emptFilter
	}
	return
}

// Calendar contest calendar count
func (s *Service) Calendar(c context.Context, p *model.ParamFilter) (rs []*model.Calendar, err error) {
	var fc map[string]int64
	before3 := time.Now().AddDate(0, 0, -_caleDay).Format("2006-01-02")
	after3 := time.Now().AddDate(0, 0, _caleDay).Format("2006-01-02")
	todayAll := p.Mid == 0 && p.Gid == 0 && p.Tid == 0 && p.Stime == before3 && p.Etime == after3
	if todayAll {
		if rs, err = s.dao.CalendarCache(c, p); err != nil {
			err = nil
		}
		if len(rs) > 0 {
			return
		}
	}
	if fc, err = s.dao.FilterCale(c, p); err != nil {
		log.Error("s.dao.FilterCale error(%v)", err)
		return
	}
	if len(fc) == 0 {
		rs = _emptCalendar
		return
	}
	for d, c := range fc {
		rs = append(rs, &model.Calendar{Stime: d, Count: c})
	}
	if todayAll {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetCalendarCache(c, p, rs)
		})
	}
	return
}

func (s *Service) fmtContest(c context.Context, contests []*model.Contest, mid int64) {
	var cids []int64
	if len(contests) == 0 {
		return
	}
	for _, contest := range contests {
		cids = append(cids, contest.ID)
	}
	favContest, _ := s.isFavs(c, mid, cids)
	for _, contest := range contests {
		contest.ContestFreeze = contest.Status
		if contest.ContestStatus == ContestStatusEnd {
			contest.GameState = _gameOver
		} else if contest.ContestStatus == ContestStatusOngoing {
			if contest.LiveRoom == 0 {
				contest.GameState = _gameIn
			} else {
				contest.GameState = _gameLive
			}
		} else if contest.LiveRoom > 0 {
			if v, ok := favContest[contest.ID]; ok && v && mid > 0 {
				contest.GameState = _gameSub
			} else {
				contest.GameState = _gameNoSub
			}
		}
		if v, ok := favContest[contest.ID]; ok && v && mid > 0 {
			contest.IsSub = 1
		}
	}
}

// ListContest contest list.
func (s *Service) ListContest(c context.Context, mid int64, p *model.ParamContest) (rs []*model.Contest, total int, err error) {
	var (
		tmpRs      []*model.Contest
		dbContests map[int64]*model.Contest
		cids       []int64
	)
	// get from cache.
	isImprove := p.Pn == 1 && p.Mid == 0 && p.Gid == 0 && p.Tid == 0 && len(p.Cids) == 0 && p.GState == "" && p.GsRecT == 0
	isFirst := isImprove && p.Stime == "" && p.Etime == "" && len(p.Sids) == 0 && p.Sort == 0
	isNoSeason := isImprove && p.Stime != "" && p.Etime != "" && len(p.Sids) == 0
	isSeason := isImprove && p.Stime != "" && p.Etime != "" && len(p.Sids) == 1
	if isNoSeason {
		if rs, total, err = s.dao.ImproveContestCache(c, p.Stime, p.Etime, p.Ps, p.Sort); err != nil {
			err = nil
		} else if len(rs) > 0 {
			s.fmtContest(c, rs, mid)
			return
		}
	} else if isSeason {
		if rs, total, err = s.dao.S9ContestCache(c, p.Sids[0], p.Stime, p.Etime, p.Ps, p.Sort); err != nil {
			err = nil
		} else if len(rs) > 0 {
			s.fmtContest(c, rs, mid)
			return
		}
	} else if isFirst {
		if rs, total, err = s.dao.ContestCache(c, p.Ps); err != nil {
			err = nil
		} else if len(rs) > 0 {
			s.fmtContest(c, rs, mid)
			return
		}
	}
	if cids, total, err = s.dao.SearchContest(c, p); err != nil {
		log.Error("s.dao.SearchContest error(%v)", err)
		err = nil
		if isSeason {
			rs = s.s9Rs(c, mid, p.Ps)
			return
		}
	}
	if total == 0 || len(cids) == 0 {
		rs = _emptContest
		return
	}
	if len(cids) > 0 {
		if dbContests, err = s.dao.EpContests(c, cids); err != nil {
			log.Error("s.dao.EpContests error(%v)", err)
			err = nil
			if isSeason {
				rs = s.s9Rs(c, mid, p.Ps)
				return
			}
			rs = _emptContest
			return
		}
	}
	for _, cid := range cids {
		if contest, ok := dbContests[cid]; ok {
			tmpRs = append(tmpRs, contest)
		}
	}
	rs = s.ContestInfo(c, cids, tmpRs, mid)
	if isNoSeason {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetImproveContestCache(c, p.Stime, p.Etime, p.Ps, p.Sort, rs, total)
		})
	} else if isSeason {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetS9ContestCache(c, p.Sids[0], p.Stime, p.Etime, p.Ps, p.Sort, rs, total)
		})
	} else if isFirst {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetContestCache(c, p.Ps, rs, total)
		})
	}
	return
}

func (s *Service) s9Rs(c context.Context, mid int64, ps int) (rs []*model.Contest) {
	var (
		tmpRs []*model.Contest
		cids  []int64
	)
	if len(s.s9Contests) == 0 {
		rs = _emptContest
		return
	}
	for index, contest := range s.s9Contests {
		if index == ps {
			break
		}
		tmpRs = append(tmpRs, contest)
		cids = append(cids, contest.ID)
	}
	rs = s.ContestInfo(c, cids, tmpRs, mid)
	if len(rs) == 0 {
		rs = _emptContest
	}
	log.Warn("s9Rs  sid(%d)  rs count(%d)", s.c.Rule.S9SwitchSID, len(rs))
	return
}

// ContestInfo contest add  team season.
func (s *Service) ContestInfo(c context.Context, cids []int64, cData []*model.Contest, mid int64) (rs []*model.Contest) {
	defer func() {
		bvav.AvToBv(rs)
	}()
	var (
		mapTeam    = make(map[int64]*model.Team, 0)
		mapSeason  = make(map[int64]*model.Season, 0)
		tids, sids []int64
	)
	for _, c := range cData {
		tids = append(tids, c.HomeID)
		tids = append(tids, c.AwayID)
		tids = append(tids, c.SuccessTeam)
		sids = append(sids, c.Sid)
	}
	eg := egV2.WithContext(c)
	eg.Go(func(ctx context.Context) (err error) {
		if len(tids) > 0 {
			tmpTids := s.rmRepeat(tids)
			teamListFromMemoryCache, missList := fetchTeamListFromMemoryCache(tmpTids)
			tool.Metric4MemoryCache.WithLabelValues([]string{"hit_team"}...).Add(float64(len(teamListFromMemoryCache)))
			if len(missList) > 0 {
				tool.Metric4MemoryCache.WithLabelValues([]string{"miss_team"}...).Add(float64(len(missList)))
				if mapTeam, err = s.dao.TeamListByIDList(ctx, missList); err != nil {
					log.Error("ContestInfo.dao.Teams error(%v)", err)
					err = nil
				}
			}

			for k, v := range teamListFromMemoryCache {
				mapTeam[k] = v
			}
		}
		return
	})
	eg.Go(func(ctx context.Context) (err error) {
		if len(sids) > 0 {
			tmpSids := s.rmRepeat(sids)
			seasonListFromMemoryCache, missList := fetchSeasonListFromMemoryCache(tmpSids)
			tool.Metric4MemoryCache.WithLabelValues([]string{"hit_season"}...).Add(float64(len(seasonListFromMemoryCache)))
			if len(missList) > 0 {
				tool.Metric4MemoryCache.WithLabelValues([]string{"miss_season"}...).Add(float64(len(missList)))
				if mapSeason, err = s.dao.SeasonListByIDList(missList); err != nil {
					log.Error("ContestInfo.dao.EpSeasons error(%v)", err)
					err = nil
				}
			}

			for k, v := range seasonListFromMemoryCache {
				mapSeason[k] = v
			}
		}
		return
	})
	eg.Wait()
	favContest, _ := s.isFavs(c, mid, cids)
	contestIds := make([]int64, 0)
	for _, contest := range cData {
		if contest.GuessType != _guessOk {
			continue
		}
		contestIds = append(contestIds, contest.ID)
	}
	contestGuessMap := make(map[int64]bool, 0)
	if mid > 0 {
		contestGuessMap = s.fetchComponentContestGuessMap(c, mid, contestIds)
	}
	for _, contest := range cData {
		if contest == nil {
			continue
		}
		contest.ContestFreeze = contest.Status
		if v, ok := mapTeam[contest.HomeID]; ok && v != nil {
			contest.HomeTeam = v
		} else {
			contest.HomeTeam = struct{}{}
		}
		if v, ok := mapTeam[contest.AwayID]; ok && v != nil {
			contest.AwayTeam = v
		} else {
			contest.AwayTeam = struct{}{}
		}
		if v, ok := mapTeam[contest.SuccessTeam]; ok && v != nil {
			contest.SuccessTeaminfo = v
		} else {
			contest.SuccessTeaminfo = struct{}{}
		}
		if v, ok := mapSeason[contest.Sid]; ok && v != nil {
			s.ldSeasonGame.Lock()
			v.GameType = s.ldSeasonGame.Data[v.LeidaSID]
			s.ldSeasonGame.Unlock()
			contest.Season = v
		} else {
			contest.Season = struct{}{}
		}
		if contest.ContestStatus == ContestStatusEnd {
			contest.GameState = _gameOver
		} else if contest.ContestStatus == ContestStatusOngoing {
			if contest.LiveRoom == 0 {
				contest.GameState = _gameIn
			} else {
				contest.GameState = _gameLive
			}
		} else if contest.LiveRoom > 0 {
			if v, ok := favContest[contest.ID]; ok && v && mid > 0 {
				contest.GameState = _gameSub
			} else {
				contest.GameState = _gameNoSub
			}
		}
		if (contest.GuessType == _guessOk) && (s.validateGuess(contest)) {
			contest.GuessShow = _guessOk
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
		// 是否订阅赛程
		if v, ok := favContest[contest.ID]; ok && v {
			contest.IsSub = 1
		}
		if v, ok := contestGuessMap[contest.ID]; ok && v {
			contest.IsGuess = _isGuess
		}
		// 主队
		home, _ := mapTeam[contest.HomeID]
		away, _ := mapTeam[contest.AwayID]
		contest.Home = convertTeam2Card(home)
		contest.Away = convertTeam2Card(away)
		contest.Series = &v1.ContestSeriesComponent{
			ID:          contest.SeriesID,
			ParentTitle: "",
			ChildTitle:  "",
			StartTime:   0,
			EndTime:     0,
		}
		rs = append(rs, contest)
	}
	return
}

func (s *Service) validateGuess(contest *model.Contest) bool {
	return time.Now().Unix() < contest.Stime
}

func (s *Service) rmRepeat(p []int64) (rs []int64) {
	m := make(map[int64]struct{})
	for _, v := range p {
		if _, ok := m[v]; !ok {
			rs = append(rs, v)
			m[v] = struct{}{}
		}
	}
	return rs
}

// ListVideo video list.
func (s *Service) ListVideo(c context.Context, p *model.ParamVideo) (rs []*arcmdl.Arc, total int, err error) {
	var (
		vData     []*model.SearchVideo
		aids      []int64
		arcsReply *arcmdl.ArcsReply
	)
	isFirst := p.Mid == 0 && p.Gid == 0 && p.Tid == 0 && p.Year == 0 && p.Tag == 0 && p.Sort == 0 && p.Pn == 1
	if isFirst {
		// get from cache.
		if rs, total, err = s.dao.VideoCache(c, p.Ps); err != nil {
			err = nil
		} else if len(rs) > 0 {
			return
		}
	}
	if vData, total, err = s.dao.SearchVideo(c, p); err != nil {
		log.Error("s.dao.SearchVideo(%v) error(%v)", p, err)
		return
	}
	if total == 0 {
		rs = _emptVideoList
		return
	}
	for _, arc := range vData {
		aids = append(aids, arc.AID)
	}
	if arcsReply, err = s.arcClient.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
		log.Error("ListVideo s.arc.Archives3 error(%v)", err)
		return
	}
	for _, aid := range aids {
		if arc, ok := arcsReply.Arcs[aid]; ok && arc.IsNormal() {
			rs = append(rs, arc)
		}
	}
	if isFirst {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetVideoCache(c, p.Ps, rs, total)
		})
	}
	return
}

// FilterVideo filter video.
func (s *Service) FilterVideo(c context.Context, p *model.ParamFilter) (rs map[string][]*model.Filter, err error) {
	var (
		tmpRs                             map[string][]*model.Filter
		fv                                *model.FilterES
		fMap                              map[string]map[int64]*model.Filter
		matchs, games, teams, tags, years []*model.Filter
	)
	isAll := p.Year == 0 && p.Tag == 0 && p.Tid == 0 && p.Gid == 0 && p.Mid == 0
	if rs, err = s.dao.FVideoCache(c); err != nil {
		err = nil
	}
	if isAll && len(rs) > 0 {
		return
	}
	matchs, games, teams, tags, years = s.filterTop()
	tmpRs = make(map[string][]*model.Filter, 3)
	tmpRs[_typeMatch] = matchs
	tmpRs[_typeGame] = games
	tmpRs[_typeTeam] = teams
	tmpRs[_typeYear] = years
	tmpRs[_typeTag] = tags
	if fv, err = s.dao.FilterVideo(c, p); err != nil {
		log.Error("s.dao.FilterVideo error(%v)", err)
		return
	}
	fMap = s.filterMap(tmpRs)
	rs = s.fmtES(fv, fMap)
	if isAll {
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetFVideoCache(c, rs)
		})
	} else {
		if len(rs[_typeMatch]) == 0 && len(rs[_typeGame]) == 0 && len(rs[_typeTeam]) == 0 && len(rs[_typeYear]) == 0 && len(rs[_typeTag]) == 0 {
			if tmpRs, err = s.dao.FVideoCache(c); err != nil {
				err = nil
			}
			if len(tmpRs) > 0 {
				rs = tmpRs
			}
		}
	}
	return
}

func (s *Service) filterTop() (matchs, games, teams, tags, years []*model.Filter) {
	var (
		matchErr, gameErr, tagErr, yearErr error
	)
	group := &errgroup.Group{}
	group.Go(func() error {
		if matchs, matchErr = s.dao.Matchs(context.Background()); matchErr != nil {
			log.Error("s.dao.Matchs error (%v)", matchErr)
		}
		return nil
	})
	group.Go(func() error {
		if games, gameErr = s.dao.Games(context.Background()); gameErr != nil {
			log.Error("s.dao.Games error (%v)", gameErr)
		}
		return nil
	})
	group.Go(func() error {
		componentTeamMap := allTeamsOfComponent.Load().(map[int64]*model.Team2TabComponent)
		for _, team := range componentTeamMap {
			teams = append(teams, &model.Filter{
				ID:       team.ID,
				Title:    team.Title,
				SubTitle: team.SubTitle,
				Logo:     team.Logo,
			})
		}
		return nil
	})
	group.Go(func() error {
		if tags, tagErr = s.dao.Tags(context.Background()); tagErr != nil {
			log.Error("s.dao.tags error (%v)", tagErr)
		}
		return nil
	})
	group.Go(func() error {
		if years, yearErr = s.dao.Years(context.Background()); yearErr != nil {
			log.Error("s.dao.Years error (%v)", yearErr)
		}
		return nil
	})
	group.Wait()
	if len(matchs) == 0 {
		matchs = _emptFilter
	}
	if len(games) == 0 {
		games = _emptFilter
	}
	if len(teams) == 0 {
		teams = _emptFilter
	}
	if len(years) == 0 {
		years = _emptFilter
	}
	if len(tags) == 0 {
		tags = _emptFilter
	}
	return
}

func (s *Service) fmtES(fv *model.FilterES, fMap map[string]map[int64]*model.Filter) (rs map[string][]*model.Filter) {
	var (
		err                                      error
		intMid, intGid, intTeam, intTag, intYear int64
		matchs, games, teams, tags, years        []*model.Filter
	)
	group := &errgroup.Group{}
	group.Go(func() error {
		for _, midGroup := range fv.GroupByMatch {
			if intMid, err = strconv.ParseInt(midGroup.Key, 10, 64); err != nil {
				err = nil
				continue
			}
			if match, ok := fMap[_typeMatch][intMid]; ok {
				matchs = append(matchs, match)
			}
		}
		return nil
	})
	group.Go(func() error {
		for _, gidGroup := range fv.GroupByGid {
			if intGid, err = strconv.ParseInt(gidGroup.Key, 10, 64); err != nil {
				err = nil
				continue
			}
			if game, ok := fMap[_typeGame][intGid]; ok {
				games = append(games, game)
			}
		}
		return nil
	})
	group.Go(func() error {
		for _, teamGroup := range fv.GroupByTeam {
			if intTeam, err = strconv.ParseInt(teamGroup.Key, 10, 64); err != nil {
				err = nil
				continue
			}
			if team, ok := fMap[_typeTeam][intTeam]; ok {
				teams = append(teams, team)
			}
		}
		return nil
	})
	group.Go(func() error {
		for _, tagGroup := range fv.GroupByTag {
			if intTag, err = strconv.ParseInt(tagGroup.Key, 10, 64); err != nil {
				err = nil
				continue
			}
			if tag, ok := fMap[_typeTag][intTag]; ok {
				tags = append(tags, tag)
			}
		}
		return nil
	})
	group.Go(func() error {
		for _, yearGroup := range fv.GroupByYear {
			if intYear, err = strconv.ParseInt(yearGroup.Key, 10, 64); err != nil {
				err = nil
				continue
			}
			if year, ok := fMap[_typeYear][intYear]; ok {
				years = append(years, year)
			}
		}
		return nil
	})
	group.Wait()
	rs = make(map[string][]*model.Filter, 5)
	if len(matchs) == 0 {
		matchs = _emptFilter
	} else {
		sort.Slice(matchs, func(i, j int) bool {
			return matchs[i].Rank > matchs[j].Rank || (matchs[i].Rank == matchs[j].Rank && matchs[i].ID < matchs[j].ID)
		})
	}
	if len(games) == 0 {
		games = _emptFilter
	} else {
		sort.Slice(games, func(i, j int) bool { return games[i].ID < games[j].ID })
	}
	if len(teams) == 0 {
		teams = _emptFilter
	} else {
		sort.Slice(teams, func(i, j int) bool { return teams[i].ID < teams[j].ID })
	}
	if len(years) == 0 {
		years = _emptFilter
	} else {
		sort.Slice(years, func(i, j int) bool { return years[i].ID < years[j].ID })
	}
	if len(tags) == 0 {
		tags = _emptFilter
	} else {
		sort.Slice(tags, func(i, j int) bool { return tags[i].ID < tags[j].ID })
	}
	rs[_typeMatch] = matchs
	rs[_typeGame] = games
	rs[_typeTeam] = teams
	rs[_typeTag] = tags
	rs[_typeYear] = years
	return
}

func (s *Service) filterMap(f map[string][]*model.Filter) (rs map[string]map[int64]*model.Filter) {
	var (
		match, game, team, tag, year                *model.Filter
		mapMatch, mapGame, mapTeam, mapTag, mapYear map[int64]*model.Filter
	)
	group := &errgroup.Group{}
	group.Go(func() error {
		mapMatch = make(map[int64]*model.Filter, len(f[_typeMatch]))
		for _, match = range f[_typeMatch] {
			if match != nil {
				mapMatch[match.ID] = match
			}
		}
		return nil
	})
	group.Go(func() error {
		mapGame = make(map[int64]*model.Filter, len(f[_typeGame]))
		for _, game = range f[_typeGame] {
			if game != nil {
				mapGame[game.ID] = game
			}
		}
		return nil
	})
	group.Go(func() error {
		mapTeam = make(map[int64]*model.Filter, len(f[_typeTeam]))
		for _, team = range f[_typeTeam] {
			if team != nil {
				mapTeam[team.ID] = team
			}
		}
		return nil
	})
	group.Go(func() error {
		mapTag = make(map[int64]*model.Filter, len(f[_typeTag]))
		for _, tag = range f[_typeTag] {
			if tag != nil {
				mapTag[tag.ID] = tag
			}
		}
		return nil
	})
	group.Go(func() error {
		mapYear = make(map[int64]*model.Filter, len(f[_typeYear]))
		for _, year = range f[_typeYear] {
			if year != nil {
				mapYear[year.ID] = year
			}
		}
		return nil
	})
	group.Wait()
	rs = make(map[string]map[int64]*model.Filter, 5)
	rs[_typeMatch] = mapMatch
	rs[_typeGame] = mapGame
	rs[_typeTeam] = mapTeam
	rs[_typeTag] = mapTag
	rs[_typeYear] = mapYear
	return
}

// Season season list.
func (s *Service) Season(c context.Context, p *model.ParamSeason) (rs []*model.Season, count int, err error) {
	var (
		seasons []*model.Season
		start   = (p.Pn - 1) * p.Ps
		end     = start + p.Ps - 1
	)
	if rs, count, err = s.dao.SeasonCache(c, start, end); err != nil || len(rs) == 0 {
		err = nil
		if seasons, err = s.dao.Season(c); err != nil {
			log.Error("s.dao.Season error(%v)", err)
			return
		}
		count = len(seasons)
		if count == 0 || count < start {
			rs = _emptSeason
			return
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetSeasonCache(c, seasons, count)
		})
		if count > end+1 {
			rs = seasons[start : end+1]
		} else {
			rs = seasons[start:]
		}
	}
	return
}

// AppSeason  app season list.
func (s *Service) AppSeason(c context.Context, p *model.ParamSeason) (rs []*model.Season, count int, err error) {
	var (
		seasons []*model.Season
		start   = (p.Pn - 1) * p.Ps
		end     = start + p.Ps - 1
	)
	if rs, count, err = s.dao.SeasonMCache(c, start, end); err != nil || len(rs) == 0 {
		err = nil
		if seasons, err = s.dao.AppSeason(c); err != nil {
			log.Error("s.dao.AppSeason error(%v)", err)
			return
		}
		count = len(seasons)
		if count == 0 || count < start {
			rs = _emptSeason
			return
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.SetSeasonMCache(c, seasons, count)
		})
		if count > end+1 {
			rs = seasons[start : end+1]
		} else {
			rs = seasons[start:]
		}
	}
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].Rank > rs[j].Rank || (rs[i].Rank == rs[j].Rank && rs[i].Stime > rs[j].Stime)
	})
	return
}

// GameRank seasonrank game list.
func (s *Service) GameRank(c context.Context) (rs []*model.GameRank, err error) {
	var (
		rankGids []int64
		isHot    bool
		tmpRs    []*model.GameRank
	)
	defer func() {
		if len(rankGids) == 0 {
			tmpRs = _emptGameRank
			return
		}
		if isHot {
			rs = append(rs, &model.GameRank{
				Title:    _hotName,
				SubTitle: _hotName,
			})
		}
		rs = append(rs, tmpRs...)
	}()
	if rankGids, err = s.dao.SeasonGames(c); err != nil {
		log.Error("GameRank s.dao.SeasonGames error(%+v)", err)
		return
	}
	if len(rankGids) == 0 {
		return
	}
	for _, id := range rankGids {
		if id == 0 {
			isHot = true
			break
		}
	}
	if tmpRs, err = s.dao.H5Games(c); err != nil {
		log.Error("s.dao.GameRank error(%v+)", err)
		err = nil
	}
	if len(tmpRs) > 0 {
		return
	}
	if tmpRs, err = s.dao.RawH5Games(c, rankGids); err != nil {
		log.Error("Games s.dao.H5Games Gids(%v) Error(%v)", rankGids, err)
		return
	}
	if len(tmpRs) > 0 {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddH5Games(c, tmpRs)
		})
	}
	return
}

// GameSeason game season list.
func (s *Service) GameSeason(c context.Context, gid int64) (rs []*model.SeasonRank, err error) {
	var (
		season map[int64]*model.Season
		sids   []int64
		tmpRs  []*model.SeasonRank
	)
	if tmpRs, err = s.dao.CacheSeasonRank(c, gid); err != nil {
		log.Error("s.dao.CacheSeasonRank gid(%d) error(%+v)", gid, err)
		err = nil
	}
	if tmpRs, err = s.dao.SeasonRank(c, gid); err != nil {
		log.Error("s.dao.SeasonRank gid(%d) error(%+v)", gid, err)
		return
	}
	rs = _emptSeasonRank
	if len(tmpRs) == 0 {
		return
	}
	s.cache.Do(c, func(c context.Context) {
		s.dao.AddCacheSeasonRank(c, gid, tmpRs)
	})
	for _, seasonRank := range tmpRs {
		sids = append(sids, seasonRank.Sid)
	}
	if season, err = s.dao.EpSeasons(c, sids); err != nil {
		log.Error("s.dao.EpSeasons gid(%d) sids(%+v) error(%+v)", gid, sids, err)
		return
	}
	count := 0
	for _, seasonRank := range tmpRs {
		if count >= 10 {
			break
		}
		s, ok := season[seasonRank.Sid]
		if !ok {
			continue
		}
		rs = append(rs, &model.SeasonRank{
			ID:       seasonRank.ID,
			Sid:      seasonRank.Sid,
			Rank:     seasonRank.Rank,
			Title:    s.Title,
			SubTitle: s.SubTitle,
		})
		count++
	}
	return
}

// Contest contest data.
func (s *Service) Contest(c context.Context, mid, cid, platform int64) (res *model.ContestDataPage, err error) {
	if maxContestID != 0 && cid > maxContestID {
		err = xecode.RequestErr

		return
	}

	var (
		contest             *model.Contest
		contestData         []*model.ContestsData
		teams               map[int64]*model.Team
		season              map[int64]*model.Season
		mapArc              map[int64]*arcmdl.Arc
		teamErr, contestErr error
	)
	if res, err = s.dao.GetCSingleData(c, cid); err != nil || res == nil {
		err = nil
		res = &model.ContestDataPage{}
		group, errCtx := errgroup.WithContext(c)
		group.Go(func() error {
			if contest, contestErr = s.dao.Contest(errCtx, cid); contestErr != nil {
				log.Error("SingleData.dao.Contest error(%v)", teamErr)
			}
			return contestErr
		})
		group.Go(func() error {
			if contestData, _ = s.dao.ContestData(errCtx, cid); err != nil {
				log.Error("SingleData.dao.ContestData error(%v)", teamErr)
			}
			return nil
		})
		err = group.Wait()
		if err != nil {
			return
		}
		if contest.ID == 0 {
			err = xecode.NothingFound
			return
		}
		if len(contestData) == 0 {
			contestData = _emptContestDetail
		}
		if teams, err = s.dao.EpTeams(c, []int64{contest.HomeID, contest.AwayID}); err != nil {
			log.Error("SingleData.dao.Teams error(%v)", err)
			err = nil
		}
		if season, err = s.dao.EpSeasons(c, []int64{contest.Sid}); err != nil {
			log.Error("SingleData.dao.EpSeasons error(%v)", err)
			err = nil
		}
		s.ContestInfos(contest, teams, season)
		res.Contest = contest
		res.Detail = contestData
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCSingleData(c, cid, res)
		})
	}
	if res.Contest != nil {
		if resp, err := s.liveClient.GetRoomPlayInfo(c, &livexroom.GetRoomPlayReq{RoomId: res.Contest.LiveRoom, Attrs: []string{"status", "show"}}); err == nil {
			if resp.Info != nil {
				if resp.Info.Status != nil {
					res.Contest.LiveStatus = resp.Info.Status.LiveStatus
				}
				if resp.Info.Show != nil {
					res.Contest.LivePopular = resp.Info.Show.PopularityCount
					res.Contest.LiveCover = resp.Info.Show.Cover
					res.Contest.LiveTitle = resp.Info.Show.Title
				}
			}
		}
	}
	if res.Contest != nil && len(res.Detail) > 0 {
		if platform == _h5 {
			mapArc = s.matchDArc(c, res.Detail)
		}
		mapDetail := s.gameStatus(c, res.Contest, res.Detail)
		for _, data := range res.Detail {
			if g, ok := mapDetail[data.PointData]; ok {
				data.GameStatus = g
				if platform == _h5 {
					aid := s.detailAid(data.URL)
					if aid > 0 {
						if arc, ok := mapArc[aid]; ok {
							data.Aid = arc.Aid
							data.Pic = arc.Pic
							data.View = arc.Stat.View
							data.Danmaku = arc.Stat.Danmaku
							data.Duration = arc.Duration
						}
					}
				}
			}
		}
	}
	tmp := []*model.Contest{res.Contest}
	s.fmtContest(c, tmp, mid)
	return
}

func (s *Service) gameStatus(c context.Context, contest *model.Contest, detail []*model.ContestsData) (rs map[int64]int64) {
	var (
		games   []*model.LolGame
		owGames []*model.OwGame
		rsGame  interface{}
		gameMap map[int64]*model.Game
		err     error
		ok      bool
	)
	if contest == nil {
		return
	}
	if rsGame, err = s.ldGame(c, contest.MatchID, contest.DataType); err != nil || rsGame == nil {
		return
	}
	if games, ok = rsGame.([]*model.LolGame); ok {
		gameMap = make(map[int64]*model.Game, len(games))
		for _, game := range games {
			gameMap[game.GameID] = &model.Game{GameID: game.GameID, Finished: game.Finished, BeginAt: game.BeginAt}
		}
	} else if owGames, ok = rsGame.([]*model.OwGame); ok {
		gameMap = make(map[int64]*model.Game, len(games))
		for _, game := range owGames {
			gameMap[game.GameID] = &model.Game{GameID: game.GameID, Finished: game.Finished, BeginAt: game.BeginAt}
		}
	}
	if len(gameMap) > 0 {
		rs = make(map[int64]int64, len(detail))
		for _, data := range detail {
			if g, ok := gameMap[data.PointData]; ok {
				if g.Finished == 1 || (contest.Etime > 0 && time.Now().Unix() > contest.Etime) {
					rs[data.PointData] = 1
				} else if g.Finished == 0 && g.BeginAt != "" {
					rs[data.PointData] = 2
				}
			}
		}
	}
	return
}

func (s *Service) ldGame(c context.Context, matchID, dataType int64) (rs interface{}, err error) {
	switch dataType {
	case _lolType:
		if rs, err = s.dao.LolGames(c, matchID); err != nil {
			log.Error("s.ldGame lol  matchID(%d) error(%+v)", matchID, err)
		}
	case _dotaType:
		if rs, err = s.dao.DotaGames(c, matchID); err != nil {
			log.Error("s.ldGame dota  matchID(%d) error(%+v)", matchID, err)
		}
	case _owType:
		if rs, err = s.dao.OwGames(c, matchID); err != nil {
			log.Error("s.ldGame overwatch  matchID(%d) error(%+v)", matchID, err)
		}
	}
	return
}

func (s *Service) matchDArc(c context.Context, detail []*model.ContestsData) (rs map[int64]*arcmdl.Arc) {
	var (
		aids      []int64
		arcsReply *arcmdl.ArcsReply
		err       error
	)
	for _, d := range detail {
		aid := s.detailAid(d.URL)
		if aid > 0 {
			aids = append(aids, aid)
		}
	}
	count := len(aids)
	if count > 0 {
		if arcsReply, err = s.arcClient.Arcs(c, &arcmdl.ArcsRequest{Aids: aids}); err != nil {
			log.Error("matchDArc s.arc.Archives3 error(%v)", err)
			return
		}
		rs = arcsReply.Arcs
	}
	return
}

func (s *Service) detailAid(url string) (rs int64) {
	var err error
	paramsBv := s.regBv.FindStringSubmatch(url)
	if len(paramsBv) > 1 {
		bv := paramsBv[1]
		if rs, err = bvid.BvToAv(bv); err != nil {
			log.Error("detailAid bvid.BvToAv url(%s) error(%+v)", url, err)
		}
	}
	if rs > 0 {
		return
	}
	params := s.regAv.FindStringSubmatch(url)
	if len(params) > 2 {
		if rs, err = strconv.ParseInt(params[2], 10, 64); err != nil {
			log.Error("detailAid  strconv.ParseInt url(%s) error(%+v)", url, err)
		}
	}
	return
}

// ContestInfos contest infos.
func (s *Service) ContestInfos(contest *model.Contest, teams map[int64]*model.Team, season map[int64]*model.Season) {
	bvav.AvToBv([]*model.Contest{contest})
	contest.ContestFreeze = contest.Status
	if homeTeam, ok := teams[contest.HomeID]; ok {
		contest.Home = convertTeam2Card(homeTeam)
		contest.HomeTeam = homeTeam
	} else {
		contest.HomeTeam = struct{}{}
	}
	if awayTeam, ok := teams[contest.AwayID]; ok {
		contest.AwayTeam = awayTeam
		contest.Away = convertTeam2Card(awayTeam)
	} else {
		contest.AwayTeam = struct{}{}
	}
	if sea, ok := season[contest.Sid]; ok {
		s.ldSeasonGame.Lock()
		sea.GameType = s.ldSeasonGame.Data[sea.LeidaSID]
		s.ldSeasonGame.Unlock()
		contest.Season = sea
	} else {
		contest.Season = struct{}{}
	}
	if (contest.GuessType == _guessOk) && (s.validateGuess(contest)) {
		contest.GuessShow = _guessOk
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
	contest.Series = &v1.ContestSeriesComponent{
		ID:          contest.SeriesID,
		ParentTitle: "",
		ChildTitle:  "",
		StartTime:   0,
		EndTime:     0,
	}
}

// Recent contest recents.
func (s *Service) Recent(c context.Context, mid int64, param *model.ParamCDRecent) (res []*model.Contest, err error) {
	//if maxContestID != 0 && param.CID > maxContestID {
	//	err = xecode.RequestErr
	//
	//	return
	//}

	var (
		teams  map[int64]*model.Team
		season map[int64]*model.Season
	)
	if res, err = s.dao.GetCRecent(c, param); err != nil || len(res) == 0 {
		err = nil
		if res, err = s.dao.ContestRecent(c, param.HomeID, param.AwayID, param.CID, param.Ps); err != nil {
			log.Error("ContestRecent.dao.ContestRecent error(%v)", err)
			return
		}
		if len(res) == 0 {
			res = _emptContest
			return
		}
		for _, contest := range res {
			if teams, err = s.dao.EpTeams(c, []int64{contest.HomeID, contest.AwayID}); err != nil {
				log.Error("SingleData.dao.Teams error(%v)", err)
				err = nil
			}
			if season, err = s.dao.EpSeasons(c, []int64{contest.Sid}); err != nil {
				log.Error("SingleData.dao.EpSeasons error(%v)", err)
				err = nil
			}
			s.ContestInfos(contest, teams, season)
			contest.SuccessTeaminfo = struct{}{}
		}
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCRecent(c, param, res)
		})
	}
	s.fmtContest(c, res, mid)
	return
}

// Intervene search .
func (s *Service) Intervene(c context.Context, pn, ps int64) (res []*model.SearchRes, count int64, err error) {
	var (
		mainIDs, allIDs []int64
		mdRs            map[int64]*model.SearchRes
	)
	if allIDs, err = s.dao.SearchMainIDs(c); err != nil {
		log.Error("s.dao.SearchMainIDs error(%+v)", err)
		return
	}
	start := (pn - 1) * ps
	end := start + ps - 1
	count = int64(len(allIDs))
	if count == 0 || count < start {
		res = _emptSearchCard
		return
	}
	if count > end+1 {
		mainIDs = allIDs[start : end+1]
	} else {
		mainIDs = allIDs[start:]
	}
	if mdRs, err = s.dao.SearchMD(c, mainIDs); err != nil {
		log.Error("s.dao.RawMDSearch mainIDS(%+v) error(%+v)", mainIDs, err)
		return
	}
	for _, id := range mainIDs {
		if md, ok := mdRs[id]; ok {
			res = append(res, md)
		}
	}
	return
}

// getContestDataPage: get ContestDataPage from redis and db.
func (s *Service) getContestDataPage(c context.Context, cid int64) (*model.ContestDataPage, error) {
	//try to get from redis.
	contestDataPage, err := s.dao.GetCSingleDataV2(c, cid)
	if err != nil || contestDataPage == nil {
		err = nil
		//try to get from db
		contestDataPage, err = s.getContestDataPageFromDB(c, cid)
		if err != nil {
			return nil, err
		}
		//get from db success, restore it to redis.
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCSingleDataV2(c, cid, contestDataPage)
		})
	}
	return contestDataPage, nil
}

// Contest contest data.
func (s *Service) ContestWithMatchRecord(c context.Context, mid, cid, platform int64) (res *model.ContestDataPageWithMatchRecord, err error) {
	mapArc := make(map[int64]*arcmdl.Arc, 0)
	res = model.NewContestDataPageWithMatchRecord()
	contestDataPage, err := s.getContestDataPage(c, cid)
	if err != nil {
		return res, err
	}
	res.Contest = contestDataPage.Contest
	res.Detail = contestDataPage.Detail
	if res.Contest != nil {
		if resp, err := s.liveClient.GetRoomPlayInfo(c, &livexroom.GetRoomPlayReq{RoomId: res.Contest.LiveRoom, Attrs: []string{"status", "show"}}); err == nil {
			if resp.Info != nil {
				if resp.Info.Status != nil {
					res.Contest.LiveStatus = resp.Info.Status.LiveStatus
				}
				if resp.Info.Show != nil {
					res.Contest.LivePopular = resp.Info.Show.PopularityCount
					res.Contest.LiveCover = resp.Info.Show.Cover
					res.Contest.LiveTitle = resp.Info.Show.Title
				}
			}
		}
	}
	if res.Contest != nil && len(res.Detail) > 0 {
		if platform == _h5 {
			mapArc = s.matchDArc(c, res.Detail)
		}
		mapDetail := s.gameStatus(c, res.Contest, res.Detail)
		for _, data := range res.Detail {
			if g, ok := mapDetail[data.PointData]; ok {
				data.GameStatus = g
				if platform == _h5 {
					aid := s.detailAid(data.URL)
					if aid > 0 {
						if arc, ok := mapArc[aid]; ok {
							data.Aid = arc.Aid
							data.Pic = arc.Pic
							data.View = arc.Stat.View
							data.Danmaku = arc.Stat.Danmaku
							data.Duration = arc.Duration
						}
					}
				}
			}
		}
	}
	tmp := []*model.Contest{res.Contest}
	s.fmtContest(c, tmp, mid)
	//竞猜记录
	//only call GuessMatchRecord when user is login and this contest is config guess
	res.Guess = make([]*actmdl.GuessUserGroup, 0)
	if res.Contest.GuessType != 0 && mid != 0 {
		guessMatchRecord, err := s.GuessMatchRecord(c, mid, cid)
		if err != nil {
			log.Errorc(c, "GuessMatchRecord Param(%v) error(%v)", cid, err)
			return res, err
		}
		res.Guess = guessMatchRecord.Guess
	}

	return res, nil
}

func (s *Service) getContestDataPageFromDB(c context.Context, cid int64) (contestDataPage *model.ContestDataPage, err error) {
	var (
		contest                             *model.Contest
		contestData                         = make([]*model.ContestsData, 0)
		teams                               = make(map[int64]*model.Team, 0)
		season                              = make(map[int64]*model.Season, 0)
		teamErr, contestErr, contestDataErr error
	)
	contestDataPage = &model.ContestDataPage{}
	group, errCtx := errgroup.WithContext(c)
	group.Go(func() error {
		if contest, contestErr = s.dao.Contest(errCtx, cid); contestErr != nil {
			log.Errorc(c, "SingleData.dao.Contest error(%v)", teamErr)
		}
		return contestErr
	})
	group.Go(func() error {
		if contestData, contestDataErr = s.dao.ContestData(errCtx, cid); contestDataErr != nil {
			log.Errorc(c, "SingleData.dao.ContestData error(%v)", contestDataErr)
		}
		return contestDataErr
	})
	defer func() {
		tool.AddDBBackSourceMetricsByKeyList(bizName4ContestDataPageOfResetCache, []int64{cid})
		if err != nil {
			tool.AddDBErrMetricsByKeyList(bizName4ContestDataPageOfResetCache, []int64{cid})
		}
	}()
	err = group.Wait()
	if err != nil {
		return
	}
	if contest.ID == 0 {
		err = xecode.NothingFound
		return
	}
	if len(contestData) == 0 {
		contestData = _emptContestDetail
	}
	if teams, err = s.dao.EpTeams(c, []int64{contest.HomeID, contest.AwayID}); err != nil {
		log.Errorc(c, "SingleData.dao.Teams error(%v)", err)
		err = nil
	}
	if season, err = s.dao.EpSeasons(c, []int64{contest.Sid}); err != nil {
		log.Errorc(c, "SingleData.dao.EpSeasons error(%v)", err)
		err = nil
	}
	s.ContestInfos(contest, teams, season)
	contestDataPage.Contest = contest
	contestDataPage.Detail = contestData
	return
}
