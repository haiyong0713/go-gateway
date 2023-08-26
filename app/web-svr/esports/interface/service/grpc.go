package service

import (
	"context"
	"fmt"
	errgroup2 "go-common/library/sync/errgroup"
	api2 "go-gateway/app/web-svr/activity/interface/api"
	"strings"
	"time"

	"go-common/library/conf/env"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	errGroup "go-common/library/sync/errgroup.v2"
	"go-main/app/community/favorite/service/api"
	favmdl "go-main/app/community/favorite/service/model"

	pb "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/interface/dao"
	"go-gateway/app/web-svr/esports/interface/dao/match_component"
	"go-gateway/app/web-svr/esports/interface/model"
	"go-gateway/app/web-svr/esports/interface/tool"
)

var (
	_empContest = make([]*pb.Contest, 0)
	_empGame    = make([]*pb.Game, 0)

	specifiedContestMap = make(map[int64]*model.Contest, 0)
	specifiedTeamMap    = make(map[int64]*model.Team, 0)
	specifiedSeasonMap  = make(map[int64]*model.Season, 0)
	emptyAidList        = make([]int64, 0)
)

const (
	_maxRoom     = 10
	_imagePre    = "https://i0.hdslb.com"
	_imagePreUat = "https://uat-i0.hdslb.com"
	_contestType = 3

	link4ContestGuess  = "https://www.bilibili.com/h5/match/data/detail/%v"
	secondsOf10Minutes = 600

	clearCacheStatusOfAllSucceed     = 0
	clearCacheStatusOfPartialSucceed = 1
	clearCacheStatusOfAllFailed      = 2
	_olympicContestBase              = 10000000
)

// TODO:
func fetchContestIDList4MemoryCache() []int64 {
	list := make([]int64, 0)
	// TODO

	return list
}

func (s *Service) UpdateSeasonGuessVersion(ctx context.Context, req *pb.UpdateSeasonGuessVersionRequest) (resp *pb.UpdateSeasonGuessVersionReply, err error) {
	resp = new(pb.UpdateSeasonGuessVersionReply)
	{
		resp.Status = clearCacheStatusOfAllFailed
	}

	var seasonID int64
	seasonID, err = s.FetchSeasonIDByMatchID(ctx, req.MatchId)
	if err != nil {
		return
	}

	err = match_component.IncrSeasonGuessVersion(ctx, seasonID)
	if err == nil {
		resp.Status = clearCacheStatusOfAllSucceed
		for i := 0; i < 10; i++ {
			if tmpErr := match_component.DeleteSeasonGuessVersionBySeasonID(ctx, seasonID); tmpErr == nil {
				if tmpErr := dao.DeleteSeasonMatchIDListCacheBySeasonID(ctx, seasonID); tmpErr == nil {
					break
				}
			}
		}
	}

	return
}

func (s *Service) ClearUserGuessCache(ctx context.Context, req *pb.ClearUserGuessCacheRequest) (resp *pb.ClearUserGuessCacheReply, err error) {
	for i := 0; i < 3; i++ {
		err = s.DeleteUserSeasonGuessListByMatchID(ctx, req.Mid, req.MatchId)
	}

	resp = new(pb.ClearUserGuessCacheReply)
	{
		resp.Status = clearCacheStatusOfAllSucceed
	}

	if err != nil {
		resp.Status = clearCacheStatusOfAllFailed
	}

	return
}

func (s *Service) ASyncUpdateMemoryCache(contestIDList []int64) {
	ticker := time.NewTicker(3 * time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ticker.C:
			tmpContestIDList := contestIDList[:]
			if tmpList, err := s.dao.RecentContestIDList(context.Background()); err == nil && len(tmpList) > 0 {
				tmpContestIDList = append(tmpContestIDList, tmpList...)
			}

			if len(tmpContestIDList) > 0 {
				if d, err := s.dao.RawEpContests(context.Background(), tmpContestIDList); err == nil {
					specifiedContestMap = d
				}
			}

			if d, err := s.dao.FetchEffectiveTeamList(context.Background()); err == nil {
				specifiedTeamMap = d
			}

			if d, err := s.dao.FetchEffectiveSeasonList(context.Background()); err == nil {
				specifiedSeasonMap = d
			}
		}
	}
}

func fetchContestListFromMemoryCache(contestIDList []int64) (m map[int64]*model.Contest, missList []int64) {
	m = make(map[int64]*model.Contest, 0)
	missList = make([]int64, 0)

	for _, v := range contestIDList {
		if d, ok := specifiedContestMap[v]; ok {
			m[d.ID] = d.DeepCopy()
		} else {
			if v > 0 {
				missList = append(missList, v)
			}
		}
	}

	return
}

func fetchTeamListFromMemoryCache(teamIDList []int64) (m map[int64]*model.Team, missList []int64) {
	m = make(map[int64]*model.Team, 0)
	missList = make([]int64, 0)

	for _, v := range teamIDList {
		if d, ok := specifiedTeamMap[v]; ok {
			m[d.ID] = d.DeepCopy()
		} else {
			if v > 0 {
				missList = append(missList, v)
			}
		}
	}

	return
}

func fetchSeasonListFromMemoryCache(seasonIDList []int64) (m map[int64]*model.Season, missList []int64) {
	m = make(map[int64]*model.Season, 0)
	missList = make([]int64, 0)

	for _, v := range seasonIDList {
		if d, ok := specifiedSeasonMap[v]; ok {
			m[d.ID] = d.DeepCopy()
		} else {
			if v > 0 {
				missList = append(missList, v)
			}
		}
	}

	return
}

// LiveDelFav grpc del fav for live
func (s *Service) LiveDelFav(c context.Context, param *pb.FavRequest) (rs *pb.NoArgRequest, err error) {
	rs = &pb.NoArgRequest{}
	log.Info("LiveDelFav Request(%+v)", param)
	if err = s.DelFav(c, param.Mid, param.Cid); err != nil {
		log.Error("LiveDelFav Request(%v) Error(%v)", param, err)
		return
	}
	log.Info("LiveDelFav Request Success(%+v)", param)
	return rs, nil
}

// LiveAddFav grpc add fav for live
func (s *Service) LiveAddFav(c context.Context, param *pb.FavRequest) (rs *pb.NoArgRequest, err error) {
	log.Info("LiveAddFav Request(%+v)", param)
	rs = &pb.NoArgRequest{}
	if err = s.AddFav(c, param.Mid, param.Cid); err != nil {
		log.Error("LiveAddFav Request(%v) Error(%v)", param, err)
		return
	}
	log.Info("LiveAddFav Request Success(%+v)", param)
	return rs, nil
}

// LiveContests .
func (s *Service) LiveContests(c context.Context, param *pb.LiveContestsRequest) (rs *pb.LiveContestsReply, err error) {
	var (
		contests    map[int64]*model.Contest
		liveContest []*pb.Contest
		tmpRs       []*model.Contest
	)
	log.Info("LiveContests Request(%+v)", param)
	if len(param.Cids) == 0 {
		return &pb.LiveContestsReply{
			Contests: _empContest,
		}, xecode.NothingFound
	}
	rs = new(pb.LiveContestsReply)
	rs.Contests = _empContest
	// 奥林匹配赛程处理
	for _, cid := range param.Cids {
		if cid > _olympicContestBase {
			rs.Contests, err = s.OlympicContestsHandler(c, param.Cids)
			if err != nil {
				return
			}
			return
		}
	}
	p := &model.ParamContest{
		Cids: param.Cids,
		Mid:  param.Mid,
	}

	contestListFromMemoryCache, missList := fetchContestListFromMemoryCache(p.Cids)
	tool.Metric4MemoryCache.WithLabelValues([]string{"hit_contest"}...).Add(float64(len(contestListFromMemoryCache)))
	if len(missList) > 0 {
		tool.Metric4MemoryCache.WithLabelValues([]string{"miss_contest"}...).Add(float64(len(missList)))
		if contests, err = s.dao.ContestListByIDList(c, missList); err != nil {
			log.Error("LiveContests Request Mid(%d) Cids(%v) Error(%v)", param.Mid, param.Cids, err)
			return
		}

		for k, v := range contestListFromMemoryCache {
			contests[k] = v
		}
	} else {
		contests = contestListFromMemoryCache
	}

	for _, c := range contests {
		tmpRs = append(tmpRs, c)
	}
	liveContest = s.fmtRPCContest(c, p.Cids, tmpRs, p.Mid)
	if len(liveContest) == 0 {
		return &pb.LiveContestsReply{
			Contests: _empContest,
		}, xecode.NothingFound
	}
	rs = &pb.LiveContestsReply{
		Contests: liveContest,
	}
	return rs, nil
}

func (s *Service) OlympicContestsHandler(ctx context.Context, contestIds []int64) (contests []*pb.Contest, err error) {
	contests = make([]*pb.Contest, 0)
	if len(contestIds) == 0 {
		return
	}
	group, errCtx := errgroup2.WithContext(ctx)
	for _, cid := range contestIds {
		contestId := cid
		if contestId > _olympicContestBase {
			group.Go(func() (err error) {
				resp, err := s.actClient.GetOlympicContestDetail(errCtx, &api2.GetOlympicContestDetailReq{
					Id: contestId - _olympicContestBase,
				})
				if err != nil || resp == nil || resp.Id == 0 {
					return
				}
				contest := s.formatOlympicContest(resp)
				if contest != nil {
					contests = append(contests, contest)
				}
				return
			})
		}
	}
	if err = group.Wait(); err != nil {
		log.Errorc(ctx, "[OlympicContestsHandler][GetOlympicContestDetail][Error], err:%+v", err)
		return
	}
	return
}

func (s *Service) formatOlympicContest(olympicInfo *api2.GetOlympicContestDetailResp) (contest *pb.Contest) {
	if olympicInfo == nil || olympicInfo.Id == 0 {
		return nil
	}
	gameState := _gameNoSub
	if olympicInfo.ContestStatus == 2 {
		gameState = _gameIn
	} else if olympicInfo.ContestStatus == 3 {
		gameState = _gameOver
	}
	return &pb.Contest{
		ID:            _olympicContestBase + olympicInfo.Id,
		GameStage:     olympicInfo.GameStage,
		Stime:         olympicInfo.Stime,
		Etime:         olympicInfo.Stime,
		HomeID:        0,
		AwayID:        0,
		HomeScore:     olympicInfo.HomeScore,
		AwayScore:     olympicInfo.AwayScore,
		GameState:     int64(gameState),
		CollectionURL: olympicInfo.VideoUrl,
		DataType:      0,
		MatchID:       0,
		Season:        nil,
		HomeTeam: &pb.Team{
			ID:       0,
			Title:    olympicInfo.HomeTeamName,
			Logo:     formatTeamUrl(olympicInfo.HomeTeamUrl),
			LogoFull: formatTeamUrl(olympicInfo.HomeTeamUrl),
		},
		AwayTeam: &pb.Team{
			ID:       0,
			Title:    olympicInfo.AwayTeamName,
			Logo:     formatTeamUrl(olympicInfo.AwayTeamUrl),
			LogoFull: formatTeamUrl(olympicInfo.AwayTeamUrl),
		},
		SuccessTeaminfo: nil,
		GuessShow:       0,
		GameStage1:      "",
		GameStage2:      "",
		JumpURL:         "",
		CanGuess:        false,
		GuessLink:       "",
		ContestFreeze:   0,
		ContestStatus:   0,
		IsOlympic:       true,
		OlympicShowRule: olympicInfo.ShowRule,
	}
}

func formatTeamUrl(url string) (finalUrl string) {
	finalUrl = url
	if strings.HasPrefix(url, "//") {
		finalUrl = fmt.Sprintf("https:%s", finalUrl)
	}
	return
}

// OttContests .
func (s *Service) OttContests(ctx context.Context, param *pb.OttContestsRequest) (rs *pb.OttContestsReply, err error) {
	var (
		contests map[int64]*model.Contest
		tmpRs    []*model.Contest
	)
	rs = &pb.OttContestsReply{Contests: _empContest}
	if len(param.Cids) == 0 {
		err = xecode.RequestErr
		return
	}
	p := &model.ParamContest{
		Cids: param.Cids,
		Mid:  param.Mid,
	}
	if contests, err = s.dao.RawEpContests(ctx, p.Cids); err != nil {
		log.Error("OttContests s.dao.RawEpContests Mid(%d) Cids(%v) Error(%v)", param.Mid, param.Cids, err)
		return
	}
	if len(contests) == 0 {
		return
	}
	for _, contest := range contests {
		tmpRs = append(tmpRs, contest)
	}
	rs = &pb.OttContestsReply{
		Contests: s.fmtRPCContest(ctx, p.Cids, tmpRs, p.Mid),
	}
	return
}

func (s *Service) fmtRPCContest(c context.Context, cids []int64, tmpRs []*model.Contest, mid int64) (rs []*pb.Contest) {
	var (
		liveS                                               *pb.Season
		liveHomeTeam, liveAwayTeam, liveSuccessTeam         *pb.Team
		seasonLogo, homeLogo, awayLogo, successLogo, imgPre string
	)
	tmp := s.ContestInfo(c, cids, tmpRs, mid)
	for _, v := range tmp {
		if v == nil {
			continue
		}
		homeTeam := &model.Team{}
		awayTeam := &model.Team{}
		successTeam := &model.Team{}
		liveHomeTeam = &pb.Team{}
		liveAwayTeam = &pb.Team{}
		liveSuccessTeam = &pb.Team{}
		if s, ok := v.Season.(*model.Season); ok {
			v.LiveSeason = s
		} else {
			v.LiveSeason = &model.Season{}
		}
		if hTeam, ok := v.HomeTeam.(*model.Team); ok {
			homeTeam = hTeam
		}
		if aTeam, ok := v.AwayTeam.(*model.Team); ok {
			awayTeam = aTeam
		}
		if sTeam, ok := v.SuccessTeaminfo.(*model.Team); ok {
			successTeam = sTeam
		}
		if env.DeployEnv == env.DeployEnvUat {
			imgPre = _imagePreUat
		} else {
			imgPre = _imagePre
		}
		if v.LiveSeason != nil && v.LiveSeason.ID > 0 {
			if v.LiveSeason.Logo != "" {
				seasonLogo = imgPre + v.LiveSeason.Logo
			}
			liveS = &pb.Season{
				ID:           v.LiveSeason.ID,
				Mid:          v.LiveSeason.Mid,
				Title:        v.LiveSeason.Title,
				SubTitle:     v.LiveSeason.SubTitle,
				Stime:        v.LiveSeason.Stime,
				Etime:        v.LiveSeason.Etime,
				Sponsor:      v.LiveSeason.Sponsor,
				Logo:         v.LiveSeason.Logo,
				Dic:          v.LiveSeason.Dic,
				Status:       v.LiveSeason.Status,
				Rank:         v.LiveSeason.Rank,
				IsApp:        v.LiveSeason.IsApp,
				URL:          v.LiveSeason.URL,
				DataFocus:    v.LiveSeason.DataFocus,
				SearchImage:  v.LiveSeason.SearchImage,
				LogoFull:     seasonLogo,
				SyncPlatform: v.LiveSeason.SyncPlatform,
			}
		}
		if homeTeam != nil && homeTeam.ID > 0 {
			if homeTeam.Logo != "" {
				homeLogo = imgPre + homeTeam.Logo
			}
			liveHomeTeam = &pb.Team{
				ID:       homeTeam.ID,
				Title:    homeTeam.Title,
				SubTitle: homeTeam.SubTitle,
				ETitle:   homeTeam.ETitle,
				Area:     homeTeam.Area,
				Logo:     homeTeam.Logo,
				UID:      homeTeam.UID,
				Members:  homeTeam.Members,
				Dic:      homeTeam.Dic,
				TeamType: homeTeam.TeamType,
				LogoFull: homeLogo,
			}
		}
		if awayTeam != nil && awayTeam.ID > 0 {
			if awayTeam.Logo != "" {
				awayLogo = imgPre + awayTeam.Logo
			}
			liveAwayTeam = &pb.Team{
				ID:       awayTeam.ID,
				Title:    awayTeam.Title,
				SubTitle: awayTeam.SubTitle,
				ETitle:   awayTeam.ETitle,
				Area:     awayTeam.Area,
				Logo:     awayTeam.Logo,
				UID:      awayTeam.UID,
				Members:  awayTeam.Members,
				Dic:      awayTeam.Dic,
				TeamType: awayTeam.TeamType,
				LogoFull: awayLogo,
			}
		}
		if successTeam != nil && successTeam.ID > 0 {
			if successTeam.Logo != "" {
				successLogo = imgPre + successTeam.Logo
			}
			liveSuccessTeam = &pb.Team{
				ID:       successTeam.ID,
				Title:    successTeam.Title,
				SubTitle: successTeam.SubTitle,
				ETitle:   successTeam.ETitle,
				Area:     successTeam.Area,
				Logo:     successTeam.Logo,
				UID:      successTeam.UID,
				Members:  successTeam.Members,
				Dic:      successTeam.Dic,
				TeamType: successTeam.TeamType,
				LogoFull: successLogo,
			}
		}
		var jumpURL string
		if s.c.Rule.JumpURL != "" {
			jumpURL = fmt.Sprintf(s.c.Rule.JumpURL, v.ID)
		}
		liveC := &pb.Contest{
			ID:              v.ID,
			GameStage:       v.GameStage,
			Stime:           v.Stime,
			Etime:           v.Etime,
			HomeID:          v.HomeID,
			AwayID:          v.AwayID,
			HomeScore:       v.HomeScore,
			AwayScore:       v.AwayScore,
			LiveRoom:        v.LiveRoom,
			Aid:             v.Aid,
			Collection:      v.Collection,
			GameState:       v.GameState,
			Dic:             v.Dic,
			Status:          v.Status,
			Sid:             v.Sid,
			Mid:             v.Mid,
			Special:         int64(v.Special),
			SuccessTeam:     v.SuccessTeam,
			SpecialName:     v.SpecialName,
			SpecialTips:     v.SpecialTips,
			SpecialImage:    v.SpecialImage,
			Playback:        v.Playback,
			CollectionURL:   v.CollectionURL,
			LiveURL:         v.LiveURL,
			DataType:        v.DataType,
			MatchID:         v.MatchID,
			Season:          liveS,
			HomeTeam:        liveHomeTeam,
			AwayTeam:        liveAwayTeam,
			SuccessTeaminfo: liveSuccessTeam,
			GuessShow:       int64(v.GuessShow),
			GameStage1:      v.GameStage1,
			GameStage2:      v.GameStage2,
			JumpURL:         jumpURL,
			CanGuess:        false,
			GuessLink:       fmt.Sprintf(link4ContestGuess, v.ID),
			ContestFreeze:   v.Status,
			ContestStatus:   v.ContestStatus,
		}

		// can not guess if less than 600s
		if liveC.Stime-time.Now().Unix() > secondsOf10Minutes {
			if v.GuessType == 1 {
				liveC.CanGuess = true
			}
		}

		rs = append(rs, liveC)
	}
	return
}

// SubContestUsers .
func (s *Service) SubContestUsers(c context.Context, param *pb.SubContestsRequest) (rs *pb.FavedUsersReply, err error) {
	if param.Cid == 0 {
		return &pb.FavedUsersReply{}, xecode.RequestErr
	}
	rs = &pb.FavedUsersReply{}
	var (
		favRs   *api.FavedUsersReply
		rsUsers []*pb.User
	)
	if favRs, err = s.favClient.FavedUsers(context.Background(), &api.FavedUsersReq{Type: int32(favmdl.TypeEsports), Oid: param.Cid, Pn: param.Pn, Ps: param.Ps}); err != nil {
		log.Error("SubContestUsers s.favClient.FavedUsers  Request(%+v)", param)
		return
	}
	rs.Page = &pb.ModelPage{
		Count: favRs.Page.Count,
		Size_: favRs.Page.Size_,
		Num:   favRs.Page.Num,
	}
	for _, users := range favRs.User {
		rsUsers = append(rsUsers, &pb.User{
			Id:    users.Id,
			Oid:   users.Oid,
			Mid:   users.Mid,
			Typ:   users.Typ,
			State: users.State,
			Ctime: users.Ctime,
			Mtime: users.Mtime,
		})
	}
	rs.User = rsUsers
	return
}

// StimeContests .
func (s *Service) StimeContests(c context.Context, param *pb.StimeContestsRequest) (rs *pb.LiveContestsReply, err error) {
	var (
		cids       []int64
		cData      []*model.Contest
		dbContests map[int64]*model.Contest
	)
	if len(param.Roomids) > _maxRoom {
		return &pb.LiveContestsReply{}, xecode.RequestErr
	}
	if cids, _, err = s.dao.SearchContestQuery(c, &model.ParamContest{Stime: param.Stime, Etime: param.Etime, Roomids: param.Roomids}); err != nil {
		log.Error("s.dao.SearchContestQuery Request(%+v) error(%v) ", param, err)
		return
	}
	if len(cids) == 0 {
		return &pb.LiveContestsReply{
			Contests: _empContest,
		}, nil
	}
	if dbContests, err = s.dao.EpContests(c, cids); err != nil {
		log.Error("s.dao.Contest error(%v)", err)
		return
	}
	for _, cid := range cids {
		if contest, ok := dbContests[cid]; ok {
			cData = append(cData, contest)
		}
	}
	liveContest := s.fmtRPCContest(c, cids, cData, param.Mid)
	if len(liveContest) == 0 {
		return &pb.LiveContestsReply{
			Contests: _empContest,
		}, nil
	}
	rs = &pb.LiveContestsReply{
		Contests: liveContest,
	}
	return rs, nil
}

// Games .
func (s *Service) Games(c context.Context, param *pb.GamesRequest) (rs *pb.GamesReply, err error) {
	var (
		games  map[int64]*pb.Game
		tmpRs  []*pb.Game
		imgPre string
	)
	rs = &pb.GamesReply{Games: _empGame}
	if len(param.Gids) == 0 {
		return
	}
	if games, err = s.dao.EpGames(c, param.Gids); err != nil {
		log.Error("Games Request Gids(%v) Error(%v)", param.Gids, err)
		return
	}
	for _, gid := range param.Gids {
		if game, ok := games[gid]; ok {
			if game.Logo != "" {
				if env.DeployEnv == env.DeployEnvUat {
					imgPre = _imagePreUat
				} else {
					imgPre = _imagePre
				}
				game.LogoFull = imgPre + game.Logo
			}
			tmpRs = append(tmpRs, game)
		}
	}
	if len(tmpRs) == 0 {
		return
	}
	rs = &pb.GamesReply{
		Games: tmpRs,
	}
	return rs, nil
}

// ContestList contest list.
func (s *Service) ContestList(c context.Context, param *pb.ContestListRequest) (rs *pb.ContestListReply, err error) {
	var (
		cids         []int64
		cData        []*model.Contest
		dbContests   map[int64]*model.Contest
		count        int
		sTime, eTime time.Time
	)
	rs = &pb.ContestListReply{
		Contests: _empContest,
		Page: &pb.ModelPage{
			Num:   int32(param.Pn),
			Size_: int32(param.Ps),
		}}
	if param.Stime != "" {
		if sTime, err = time.ParseInLocation("2006-01-02", param.Stime, time.Local); err != nil {
			err = xecode.RequestErr
			return
		}
		param.Stime = time.Unix(sTime.Unix(), 0).Format("2006-01-02") + " 00:00:00"
	}
	if param.Etime != "" {
		if eTime, err = time.ParseInLocation("2006-01-02", param.Etime, time.Local); err != nil {
			err = xecode.RequestErr
			return
		}
		param.Etime = time.Unix(eTime.Unix(), 0).Format("2006-01-02") + " 23:59:59"
	}
	isImprove := param.Pn == 1 && param.Stime != "" && param.Etime != "" && param.MatchId == 0 && param.Tid == 0 && len(param.Cids) == 0 && param.GuessType == 0
	isNoSeason := isImprove && len(param.Sids) == 0
	isSeason := isImprove && len(param.Sids) == 1
	if isNoSeason {
		if cList, total, e := s.dao.CacheNoSeasonCont(c, sTime.Unix(), eTime.Unix(), param.Ps, param.Sort); e == nil && len(cList) > 0 {
			s.fmtGrpcContest(c, cList, param.Mid)
			rs.Page.Count = int32(total)
			rs.Contests = cList
			return
		}
	} else if isSeason {
		if cList, total, e := s.dao.CacheSeasonCont(c, param.Sids[0], sTime.Unix(), eTime.Unix(), param.Ps, param.Sort); e == nil && len(cList) > 0 {
			s.fmtGrpcContest(c, cList, param.Mid)
			rs.Page.Count = int32(total)
			rs.Contests = cList
			return
		}
	}
	p := &model.ParamContest{
		Sort:   int(param.Sort),
		Mid:    param.MatchId,
		Tid:    param.Tid,
		Stime:  param.Stime,
		Etime:  param.Etime,
		Sids:   param.Sids,
		Cids:   param.Cids,
		GsType: int(param.GuessType),
		Pn:     int(param.Pn),
		Ps:     int(param.Ps)}
	if cids, count, err = s.dao.SearchContestQuery(c, p); err != nil {
		log.Error("s.dao.SearchContestQuery Request(%+v) error(%v) ", param, err)
		err = nil
		return
	}
	rs.Page.Count = int32(count)
	if len(cids) == 0 {
		return
	}
	if dbContests, err = s.dao.EpContests(c, cids); err != nil {
		log.Error("s.dao.Contest error(%v)", err)
		err = nil
		return
	}
	for _, cid := range cids {
		if contest, ok := dbContests[cid]; ok {
			cData = append(cData, contest)
		}
	}
	if len(cData) == 0 {
		return
	}
	rsContest := s.fmtRPCContest(c, cids, cData, param.Mid)
	rs.Contests = rsContest
	if isNoSeason {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCacheNoSeasonCont(c, sTime.Unix(), eTime.Unix(), param.Ps, param.Sort, rs.Contests, int(rs.Page.Count))
		})
	} else if isSeason {
		s.cache.Do(c, func(c context.Context) {
			s.dao.AddCacheSeasonCont(c, param.Sids[0], sTime.Unix(), eTime.Unix(), param.Ps, param.Sort, rs.Contests, int(rs.Page.Count))
		})
	}
	return
}

func (s *Service) fmtGrpcContest(c context.Context, contests []*pb.Contest, mid int64) {
	var cids []int64
	if len(contests) == 0 {
		return
	}
	for _, contest := range contests {
		cids = append(cids, contest.ID)
	}
	favContest, _ := s.isFavs(c, mid, cids)
	for _, contest := range contests {
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
	}
}

// GameMap .
func (s *Service) GameMap(c context.Context, param *pb.GameMapRequest) (rs *pb.GameMapReply, err error) {
	var (
		gmap  map[int64]int64
		gids  []int64
		games map[int64]*pb.Game
	)
	rs = &pb.GameMapReply{
		Games: make(map[int64]*pb.Game),
	}
	if len(param.Cids) == 0 {
		return rs, xecode.RequestErr
	}
	if gmap, err = s.dao.EpGameMap(c, param.Cids, _contestType); err != nil {
		log.Error("GameMap s.dao.RawEpGameMap  cids(%v) tp(%d) error(%+v)", param.Cids, _contestType, err)
		return
	}
	if len(gmap) == 0 {
		return
	}
	for _, gid := range gmap {
		gids = append(gids, gid)
	}
	if games, err = s.dao.EpGames(c, gids); err != nil {
		log.Error("GameMap Request gids(%v) Error(%v)", gids, err)
		return
	}
	if len(games) == 0 {
		return
	}
	for cid, gid := range gmap {
		if game, ok := games[gid]; ok {
			if game.Logo != "" {
				game.LogoFull = _imagePre + game.Logo
			}
			rs.Games[cid] = game
		}
	}
	return
}

// RefreshContestDataPageCacheRequest .
func (s *Service) RefreshContestDataPageCache(c context.Context, param *pb.RefreshContestDataPageCacheRequest) (rs *pb.NoArgRequest, err error) {
	for _, cid := range param.Cids {
		var contestDataPage *model.ContestDataPage
		//get from db
		for i := 0; i <= 3; i++ {
			contestDataPage, err = s.getContestDataPageFromDB(c, cid)
			if err == nil {
				break
			}
		}
		//get from db fail, update metric to make alert.
		if err != nil {
			log.Errorc(c, "query contest_data_page id(%v) from db error: %v", cid, err)
			tool.Metric4CacheResetFailed.WithLabelValues([]string{bizName4ContestDataPageOfResetCache, tool.CacheOfRemote}...).Inc()
			continue
		}

		for i := 0; i <= 3; i++ {
			err = s.dao.AddCSingleDataV2(c, cid, contestDataPage)
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Errorc(c, "set contest_data_page id (%v) to cache error: %v", cid, err)
			tool.Metric4CacheResetFailed.WithLabelValues([]string{bizName4ContestDataPageOfResetCache, tool.CacheOfRemote}...).Inc()
		}
	}
	return &pb.NoArgRequest{}, nil
}

func (s *Service) ClearCache(ctx context.Context, req *pb.ClearCacheRequest) (reply *pb.ClearCacheReply, err error) {
	status := clearCacheStatusOfAllSucceed

	var resetErr error
	failedList := make([]int64, 0)
	if len(req.CacheKeys) > 0 {
		switch req.CacheType {
		case pb.ClearCacheType_CONTEST:
			failedList, resetErr = s.dao.ResetContestCacheByIDList(req.CacheKeys)
		case pb.ClearCacheType_SEASON:
			failedList, resetErr = s.dao.ResetSeasonCacheByIDList(req.CacheKeys)
		case pb.ClearCacheType_TEAM:
			failedList, resetErr = s.dao.ResetTeamCacheByIDList(req.CacheKeys)
		case pb.ClearCacheType_TEAMS_IN_SEASON:
			failedList, resetErr = s.dao.ResetTeamsInSeasonBySeasonIds(req.CacheKeys)
		}

	}

	if resetErr != nil {
		status = clearCacheStatusOfAllFailed
	} else {
		len4FailedList := len(failedList)
		if len4FailedList > 0 {
			status = clearCacheStatusOfPartialSucceed
			if len4FailedList == len(req.CacheKeys) {
				status = clearCacheStatusOfAllFailed
			}
		}
	}

	reply = new(pb.ClearCacheReply)
	{
		reply.Status = int64(status)
		reply.CacheType = req.CacheType
		reply.CacheKeys = req.CacheKeys
		reply.FailedKeys = failedList
	}

	return
}

// ComponentSeasonContestList .
func (s *Service) ComponentSeasonContestList(ctx context.Context, param *pb.ComponentSeasonContestListRequest) (rs *pb.ComponentSeasonContestListReply, err error) {
	rs = &pb.ComponentSeasonContestListReply{
		ComponentContestList: make(map[int64]*pb.ContestCardComponentList),
	}
	var componentContests map[int64][]*pb.ContestCardComponent
	if componentContests, err = fetchComponentContestListBySeasonID(context.Background(), param.Sid); err != nil {
		log.Errorc(ctx, "contest component ComponentSeasonContestList fetchComponentContestListBySeasonID() sid(%d) error(%+v)", param.Sid, err)
		return
	}
	tmpComponentContestList := make(map[int64]*pb.ContestCardComponentList, len(componentContests))
	for startDate, list := range componentContests {
		tmpComponentContestList[startDate] = &pb.ContestCardComponentList{List: list}
	}
	rs.ComponentContestList = tmpComponentContestList
	return
}

// ComponentSeasonContestBattle.
func (s *Service) ComponentSeasonContestBattle(ctx context.Context, param *pb.ComponentSeasonContestBattleRequest) (rs *pb.ComponentSeasonContestBattleReply, err error) {
	rs = &pb.ComponentSeasonContestBattleReply{
		ComponentContestBattle: make(map[int64]*pb.ContestBattleCardComponentList),
	}
	var componentContestsBattle map[int64][]*pb.ContestBattleCardComponent
	if componentContestsBattle, err = fetchComponentContestBattleBySeasonID(context.Background(), param.Sid); err != nil {
		log.Errorc(ctx, "contest component ComponentSeasonContestList fetchComponentContestListBySeasonID() sid(%d) error(%+v)", param.Sid, err)
		return
	}
	tmpComponentContestList := make(map[int64]*pb.ContestBattleCardComponentList, len(componentContestsBattle))
	for startDate, list := range componentContestsBattle {
		tmpComponentContestList[startDate] = &pb.ContestBattleCardComponentList{List: list}
	}
	rs.ComponentContestBattle = tmpComponentContestList
	return
}

// ClearComponentContestCache .
func (s *Service) ClearComponentContestCache(ctx context.Context, param *pb.ClearComponentContestCacheRequest) (res *pb.NoArgRequest, err error) {
	res = &pb.NoArgRequest{}
	egV2 := errGroup.WithContext(ctx)
	egV2.Go(func(ctx context.Context) error {
		// 删除赛季下赛程缓存.
		if param.ContestID > 0 {
			return match_component.FetchContestsBySeasonDeleteCache(ctx, param.SeasonID)
		}
		return nil
	})
	egV2.Go(func(ctx context.Context) error {
		// 删除赛季下赛程吃鸡类比赛缓存.
		if param.ContestID > 0 {
			return match_component.FetchContestBattleBySeasonDeleteCache(ctx, param.SeasonID)
		}
		return nil
	})
	egV2.Go(func(ctx context.Context) error {
		// 删除两队最近交锋赛程列表.
		if param.ContestID > 0 {
			return match_component.FetchHomeAwayContestsDeleteCache(ctx, param.ContestHome, param.ContestAway)
		}
		return nil
	})
	egV2.Go(func(ctx context.Context) error {
		// 删除赛程阶段缓存.
		if param.SeriesID > 0 {
			return match_component.DelContestSeriesCacheKey(ctx, param.SeriesID)
		}
		return nil
	})
	egV2.Go(func(ctx context.Context) error {
		// 删除赛程卡缓存.
		if isGoingSeason(param.SeasonID) { // 忽略进行中的赛季.
			return nil
		}
		return match_component.FetchContestCardListDeleteCache(ctx, param.SeasonID)
	})
	egV2.Go(func(ctx context.Context) error {
		// 删除赛程卡吃鸡类比赛缓存.
		if isGoingBattleSeason(param.SeasonID) { // 忽略进行中的吃鸡类比赛赛季.
			return nil
		}
		return match_component.FetchContestBattleCardListDeleteCache(ctx, param.SeasonID)
	})
	egV2.Go(func(ctx context.Context) error {
		// 删除赛程阶段积分表/树状图缓存.
		return s.dao.DelSeriesExtraInfo(ctx, param.SeriesID)

	})
	if err = egV2.Wait(); err != nil {
		log.Errorc(ctx, "contest component ClearComponentContestCache param(%+v) error(%+v)", param, err)
	}
	return
}

// ClearMatchSeasonsCache.
func (s *Service) ClearMatchSeasonsCache(ctx context.Context, param *pb.ClearMatchSeasonsCacheRequest) (res *pb.NoArgRequest, err error) {
	res = &pb.NoArgRequest{}
	egV2 := errGroup.WithContext(ctx)
	egV2.Go(func(ctx context.Context) error {
		if err = s.DelMatchSeasonsCache(ctx, param.MatchID); err != nil {
			log.Errorc(ctx, "ClearMatchSeasonsCache s.dao.DelCacheSeasonsByMatchId() matchID(%d) error(%+v)", param.MatchID, err)
			return err
		}
		return nil
	})
	egV2.Go(func(ctx context.Context) error {
		if err = s.DelSeasonInfoCache(ctx, param.SeasonID); err != nil {
			log.Errorc(ctx, "ClearMatchSeasonsCache s.dao.DelSeasonInfoCache() seasonID(%d) error(%+v)", param.SeasonID, err)
			return err
		}
		return nil
	})
	if err = egV2.Wait(); err != nil {
		log.Errorc(ctx, "ClearMatchSeasonsCache egV2.Wait() param(%+v) error(%+v)", param, err)
	}
	return
}

// SubContestUsersV2 .
func (s *Service) SubContestUsersV2(ctx context.Context, param *pb.SubContestUsersV2Request) (rs *pb.SubContestUsersV2Reply, err error) {
	rs = &pb.SubContestUsersV2Reply{}
	if param.Cid == 0 {
		err = xecode.RequestErr
		return
	}
	var (
		favRs   *api.SubscribersReply
		rsUsers []*pb.User
	)
	arg := &api.SubscribersReq{Type: int32(favmdl.TypeEsports), Oid: param.Cid, Cursor: param.Cursor, Size_: param.CursorSize}
	if favRs, err = s.favClient.Subscribers(ctx, arg); err != nil {
		log.Errorc(ctx, "SubContestUserV2 s.favClient.Subscribers  param(%+v) arg(%+v) error(%+v)", param, arg, err)
		return
	}
	rs.Cursor = favRs.Cursor
	if len(favRs.User) == 0 {
		rs.User = make([]*pb.User, 0)
		return
	}
	for _, users := range favRs.User {
		rsUsers = append(rsUsers, &pb.User{
			Id:    users.Id,
			Oid:   users.Oid,
			Mid:   users.Mid,
			Typ:   users.Typ,
			State: users.State,
			Ctime: users.Ctime,
			Mtime: users.Mtime,
		})
	}
	rs.User = rsUsers
	return
}

// VideoListFilter .
func (s *Service) VideoListFilter(ctx context.Context, param *pb.VideoListFilterRequest) (rs *pb.VideoListFilterReply, err error) {
	rs = &pb.VideoListFilterReply{}
	arg := &model.ParamFilter{
		Gid:  param.GameId,
		Mid:  param.MatchId,
		Year: param.YearId,
	}
	filterRes, e := s.FilterVideo(ctx, arg)
	if e != nil {
		err = e
		log.Errorc(ctx, "VideoListFilter s.FilterVideo() arg(%+v) error(%+v)", arg, err)
		return
	}
	for filterType, filterList := range filterRes {
		switch filterType {
		case _typeGame:
			rs.Games = formatVideoListFilter(filterList)
		case _typeMatch:
			rs.Matchs = formatVideoListFilter(filterList)
		case _typeYear:
			rs.Years = formatVideoListFilter(filterList)
		default:
		}
	}
	return
}

func formatVideoListFilter(filterList []*model.Filter) (res *pb.VideoListFilterItemList) {
	tmpList := make([]*pb.VideoListFilterItem, 0)
	res = &pb.VideoListFilterItemList{List: tmpList}
	for _, item := range filterList {
		tmpList = append(tmpList, &pb.VideoListFilterItem{
			ID:       item.ID,
			Title:    item.Title,
			SubTitle: item.SubTitle,
		})
	}
	res.List = tmpList
	return
}

// ClearTopicVideoListCache .
func (s *Service) ClearTopicVideoListCache(ctx context.Context, param *pb.ClearTopicVideoListRequest) (res *pb.NoArgRequest, err error) {
	res = &pb.NoArgRequest{}
	if err = s.dao.DelVideoListCacheKey(ctx, param.ID); err != nil {
		log.Errorc(ctx, "DelVideoListCacheKey s.dao.DelVideoListCache() id(%d) error(%+v)", param.ID, err)
	}
	return
}

// EsTopicVideoList .
func (s *Service) EsTopicVideoList(ctx context.Context, param *pb.EsTopicVideoListRequest) (rs *pb.EsTopicVideoListReply, err error) {
	var (
		searchList []*model.SearchVideo
		total      int
		searchAids []int64
	)
	rs = &pb.EsTopicVideoListReply{}
	searchParam := &model.ParamVideo{
		Gid:  param.GameId,
		Mid:  param.MatchId,
		Year: param.YearId,
		Pn:   int(param.Pn),
		Ps:   int(param.Ps),
	}
	if searchList, total, err = s.dao.SearchVideo(ctx, searchParam); err != nil {
		log.Errorc(ctx, "EsTopicVideoList s.dao.SearchVideo(%v) error(%v)", searchParam, err)
		return
	}
	if total > 0 {
		for _, searchValue := range searchList {
			searchAids = append(searchAids, searchValue.AID)
		}
	} else {
		searchAids = emptyAidList
	}
	rs = &pb.EsTopicVideoListReply{
		SearchAids: searchAids,
		Page: &pb.ModelPage{
			Num:   int32(param.Pn),
			Size_: int32(param.Ps),
			Count: int32(total),
		},
	}
	return
}

// RefreshLolData .
func (s *Service) RefreshLolData(ctx context.Context, param *pb.RefreshLolDataRequest) (res *pb.NoArgRequest, err error) {
	res = &pb.NoArgRequest{}
	if err = s.RefreshLolDataCache(ctx, param.LeidaSid); err != nil {
		log.Errorc(ctx, "RefreshLolData s.RefreshLolDataCache() param.LeidaSid(%d) error(%+v)", param.LeidaSid, err)
		return
	}
	return
}
