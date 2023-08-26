package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	errGroup "go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/component"
	"go-gateway/app/web-svr/esports/admin/model"
	v1 "go-gateway/app/web-svr/esports/interface/api/v1"

	accwarden "git.bilibili.co/bapis/bapis-go/account/service"
)

const (
	_tv                           = 1
	_live                         = 2
	_maxSeasonRank                = 20
	_insertTeamInSeasonIfNotExist = "INSERT INTO es_team_in_seasons (sid,tid) VALUES (?,?) ON DUPLICATE KEY UPDATE rank=rank"
)

var (
	_emptySeasonList = make([]*model.Season, 0)
	_emptyRankList   = make([]*model.SeasonInfo, 0)
)

// SeasonInfo .
func (s *Service) SeasonInfo(c context.Context, id int64) (data *model.SeasonInfo, err error) {
	var gameMap map[int64][]*model.Game
	season := new(model.Season)
	if err = s.dao.DB.Where("id=?", id).First(&season).Error; err != nil {
		log.Error("SeasonInfo Error (%v)", err)
		return
	}
	if gameMap, err = s.gameList(model.TypeSeason, []int64{id}); err != nil {
		return
	}
	if games, ok := gameMap[id]; ok {
		data = &model.SeasonInfo{Season: season, Games: games}
	} else {
		data = &model.SeasonInfo{Season: season, Games: _emptyGameList}
	}
	return
}

func (s *Service) BigFix(ctx context.Context, tp, sid int64) (err error) {
	var rly interface{}
	conn := component.GlobalAutoSubCache.Get(ctx)
	defer func() {
		_ = conn.Close()
	}()
	rkey := fmt.Sprintf("big_sid_%d", sid)
	if rly, err = conn.Do("SET", rkey, "1", "EX", s.c.Rule.MatchFixLimit, "NX"); err != nil {
		log.Error("conn.Do(GET key(%s)) error(%v)", rkey, err)
		err = fmt.Errorf("操作失败(%+v)", err)
		return
	}
	if err != nil {
		err = fmt.Errorf("数据更新失败，请稍后再操作(%+v)", err)
		return
	}
	if rly == nil {
		err = fmt.Errorf("数据在更新中，请稍等")
		return
	}
	// 因为修复需要很久
	go s.dao.FixBigUseJob(tp, sid)
	return
}

// SeasonList .
func (s *Service) SeasonList(c context.Context, mid, pn, ps, gid int64, title string) (list []*model.SeasonInfo, count int64, err error) {
	var (
		seasons   []*model.Season
		seasonIDs []int64
		gameMap   map[int64][]*model.Game
	)
	list = make([]*model.SeasonInfo, 0)
	source := s.dao.DB.Model(&model.Season{})
	if gid > 0 {
		filterByGameSeasonIds := make([]int64, 0)
		gidMaps := make([]*model.GIDMap, 0)
		if err = s.dao.DB.Model(&model.GIDMap{}).Where("is_deleted=?", _notDeleted).Where("type=?", model.TypeSeason).Where("gid=?", gid).Find(&gidMaps).Error; err != nil {
			log.Error("gameList gidMap Error (%v)", err)
			return
		}
		for _, gidMap := range gidMaps {
			filterByGameSeasonIds = append(filterByGameSeasonIds, gidMap.Oid)
		}
		source = source.Where("id IN(?)", filterByGameSeasonIds)
	}
	if mid > 0 {
		source = source.Where("mid=?", mid)
	}
	if title != "" {
		source = source.Where("title like ?", "%"+title+"%")
	}
	source.Count(&count)
	if err = source.Offset((pn - 1) * ps).Order("rank DESC,id ASC").Limit(ps).Find(&seasons).Error; err != nil {
		log.Error("SeasonList Error (%v)", err)
		return
	}
	if len(seasons) == 0 {
		return
	}
	for _, v := range seasons {
		seasonIDs = append(seasonIDs, v.ID)
	}
	if gameMap, err = s.gameList(model.TypeSeason, seasonIDs); err != nil {
		return
	}

	for _, v := range seasons {
		v.Platforms = s.platforms(v.SyncPlatform)
		if games, ok := gameMap[v.ID]; ok {
			list = append(list, &model.SeasonInfo{Season: v, Games: games})
		} else {
			list = append(list, &model.SeasonInfo{Season: v, Games: _emptyGameList})
		}
	}
	return
}

func (s *Service) platforms(value int64) (rs string) {
	rsTv := value & _tv
	if rsTv > 0 {
		rs = strconv.Itoa(_tv)
	}
	rsLive := value & _live
	if rsLive > 0 {
		if rsTv > 0 {
			rs += ","
		}
		rs += strconv.Itoa(_live)
	}
	return rs
}
func (s *Service) checkoutSendUID(ctx context.Context, sendUID int64) error {
	if sendUID == 0 {
		return nil
	}
	var ip = metadata.String(ctx, metadata.RemoteIP)
	arg := &accwarden.MidReq{Mid: sendUID, RealIp: ip}
	infoReply, err := s.accClient.Info3(ctx, arg)
	if err != nil {
		log.Error("checkoutSendUID 账号Infos:grpc错误 s.accClient.Info mid(%+v) error(%v)", sendUID, err)
		err = xecode.Errorf(xecode.RequestErr, "私信通知卡片发送账号不正确")
		return err
	}
	if infoReply == nil || infoReply.Info.Mid == 0 {
		err = xecode.Errorf(xecode.RequestErr, "私信通知卡片发送账号不正确")
		return err
	}
	return nil
}

// AddSeason .
func (s *Service) AddSeason(c context.Context, param *model.Season, gids []int64) (err error) {
	var (
		games   []*model.Game
		gidMaps []*model.GIDMap
		types   []int64
	)
	if param.Platforms != "" {
		if types, err = xstr.SplitInts(param.Platforms); err != nil {
			err = xecode.RequestErr
			log.Error("AddSeason check Platforms (%s) Error (%v)", param.Platforms, err)
			return
		}
	}
	if err = s.checkoutSendUID(c, param.MessageSenduid); err != nil {
		return
	}
	for _, tp := range types {
		param.SyncPlatform += tp
	}
	// TODO check name exist
	if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN (?)", gids).Find(&games).Error; err != nil {
		log.Error("AddSeason check game ids Error (%v)", err)
		return
	}
	if len(games) == 0 {
		log.Error("AddSeason games(%v) not found", gids)
		err = xecode.RequestErr
		return
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&model.Season{}).Create(param).Error; err != nil {
		log.Error("AddSeason s.dao.DB.Model Create(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	for _, v := range games {
		gidMaps = append(gidMaps, &model.GIDMap{Type: model.TypeSeason, Oid: param.ID, Gid: v.ID})
	}
	sql, sqlParam := model.GidBatchAddSQL(gidMaps)
	if err = tx.Model(&model.GIDMap{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Error("AddSeason s.dao.DB.Model Create(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}

	// record season in auto subscribe season list
	//autoSubSeason := model.AutoSubscribeSeason{
	//	SeasonID: param.ID,
	//}
	//err = tx.Create(&autoSubSeason).Error
	//if err != nil {
	//	return tx.Rollback().Error
	//}

	// create sub auto_subscribe detail table
	//err = tx.Exec(model.GenAutoSubSeasonDetailSql(param.ID)).Error
	//if err != nil {
	//	return tx.Rollback().Error
	//}

	if err := tx.Commit().Error; err != nil {
		log.Errorc(c, "AddSeason tx.Commit error(%v)", err)
		return err
	}

	go func() {
		_ = s.BatchRefreshContestDataPageCacheBySeasonId(context.Background(), param.ID)
		_ = s.ClearESportCacheByType(v1.ClearCacheType_SEASON, []int64{param.ID})
	}()

	s.cache.Do(c, func(c context.Context) {
		if e := s.ClearMatchSeasonsCacheByGRPC(&v1.ClearMatchSeasonsCacheRequest{MatchID: param.Mid, SeasonID: param.ID}); e != nil {
			log.Error("MatchSeasonsInfo ClearComponentContestCacheGRPC matchID(%+v) seasonID(%d) error(%+v)", param.Mid, param.ID, err)
		}
	})
	return
}

// EditSeason .
func (s *Service) EditSeason(c context.Context, param *model.Season, gids []int64) (err error) {
	var (
		games                    []*model.Game
		preGidMaps, addGidMaps   []*model.GIDMap
		upGidMapAdd, upGidMapDel []int64
		types                    []int64
	)
	if param.Platforms != "" {
		if types, err = xstr.SplitInts(param.Platforms); err != nil {
			log.Error("EditSeason check Platforms (%s) Error (%v)", param.Platforms, err)
			err = xecode.RequestErr
			return
		}
	}
	if err = s.checkoutSendUID(c, param.MessageSenduid); err != nil {
		return
	}
	for _, tp := range types {
		param.SyncPlatform += tp
	}
	preData := new(model.Season)
	if err = s.dao.DB.Where("id=?", param.ID).First(&preData).Error; err != nil {
		log.Error("EditSeason s.dao.DB.Where id(%d) error(%d)", param.ID, err)
		return
	}
	if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN (?)", gids).Find(&games).Error; err != nil {
		log.Error("EditSeason check game ids Error (%v)", err)
		return
	}
	if len(games) == 0 {
		log.Error("EditSeason games(%v) not found", gids)
		err = xecode.RequestErr
		return
	}
	if err = s.dao.DB.Model(&model.GIDMap{}).Where("oid=?", param.ID).Where("type=?", model.TypeSeason).Find(&preGidMaps).Error; err != nil {
		log.Error("EditSeason games(%v) not found", gids)
		return
	}
	gidsMap := make(map[int64]int64, len(gids))
	preGidsMap := make(map[int64]int64, len(preGidMaps))
	for _, v := range gids {
		gidsMap[v] = v
	}
	for _, v := range preGidMaps {
		preGidsMap[v.Gid] = v.Gid
		if _, ok := gidsMap[v.Gid]; ok {
			if v.IsDeleted == 1 {
				upGidMapAdd = append(upGidMapAdd, v.ID)
			}
		} else {
			upGidMapDel = append(upGidMapDel, v.ID)
		}
	}
	for _, gid := range gids {
		if _, ok := preGidsMap[gid]; !ok {
			addGidMaps = append(addGidMaps, &model.GIDMap{Type: model.TypeSeason, Oid: param.ID, Gid: gid})
		}
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&model.Season{}).Save(param).Error; err != nil {
		log.Error("EditSeason Match Update(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	if len(upGidMapAdd) > 0 {
		if err = tx.Model(&model.GIDMap{}).Where("id IN (?)", upGidMapAdd).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("EditSeason GIDMap Add(%+v) error(%v)", upGidMapAdd, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(upGidMapDel) > 0 {
		if err = tx.Model(&model.GIDMap{}).Where("id IN (?)", upGidMapDel).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("EditSeason GIDMap Del(%+v) error(%v)", upGidMapDel, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(addGidMaps) > 0 {
		sql, sqlParam := model.GidBatchAddSQL(addGidMaps)
		if err = tx.Model(&model.GIDMap{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("EditSeason GIDMap Create(%+v) error(%v)", addGidMaps, err)
			err = tx.Rollback().Error
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Errorc(c, "EditSeason tx.Commit error(%v)", err)
		return err
	}
	s.ClearSeasonCache(c, param.ID)
	go func() {
		_ = s.BatchRefreshContestDataPageCacheBySeasonId(context.Background(), param.ID)
		_ = s.ClearESportCacheByType(v1.ClearCacheType_SEASON, []int64{param.ID})
	}()

	s.cache.Do(c, func(c context.Context) {
		if e := s.ClearMatchSeasonsCacheByGRPC(&v1.ClearMatchSeasonsCacheRequest{MatchID: param.Mid, SeasonID: param.ID}); e != nil {
			log.Error("MatchSeasonsInfo ClearComponentContestCacheGRPC matchID(%+v) seasonID(%d) error(%+v)", param.Mid, param.ID, err)
		}
	})
	return
}

// ForbidSeason .
func (s *Service) ForbidSeason(c context.Context, id int64, state int) (err error) {
	preSeason := new(model.Season)
	if err = s.dao.DB.Where("id=?", id).First(&preSeason).Error; err != nil {
		log.Errorc(c, "SeasonForbid s.dao.DB.Where id(%d) error(%d)", id, err)
		return
	}
	if err = s.dao.DB.Model(&model.Season{}).Where("id=?", id).Update(map[string]int{"status": state}).Error; err != nil {
		log.Errorc(c, "SeasonForbid s.dao.DB.Model error(%v)", err)
		return err
	}
	go s.BatchRefreshContestDataPageCacheBySeasonId(context.Background(), id)

	s.ClearSeasonCache(c, id)
	s.cache.Do(c, func(c context.Context) {
		if e := s.ClearMatchSeasonsCacheByGRPC(&v1.ClearMatchSeasonsCacheRequest{MatchID: preSeason.Mid, SeasonID: preSeason.ID}); e != nil {
			log.Error("MatchSeasonsInfo ClearComponentContestCacheGRPC matchID(%+v) seasonID(%d) error(%+v)", preSeason.Mid, preSeason.ID, err)
		}
	})
	return
}

// ForbidRankSeason .
func (s *Service) ForbidRankSeason(c context.Context, id int64, state int) (err error) {
	preSeason := new(model.SeasonRank)
	if err = s.dao.DB.Where("id=?", id).First(&preSeason).Error; err != nil {
		log.Error("ForbidRankSeason s.dao.DB.Where id(%d) error(%d)", id, err)
		return
	}
	if err = s.dao.DB.Model(&model.SeasonRank{}).Where("id=?", id).Update(map[string]int{"is_deleted": state}).Error; err != nil {
		log.Errorc(c, "ForbidRankSeason s.dao.DB.Model error(%v)", err)
		return err
	}
	return
}

// RankInfo .
func (s *Service) RankInfo(c context.Context, id int64) (data *model.SeasonRank, err error) {
	var gameName, seasonName string
	data = new(model.SeasonRank)
	if err = s.dao.DB.Where("is_deleted=?", _notDeleted).Where("id=?", id).First(&data).Error; err != nil {
		log.Error("SeasonInfo Error (%v)", err)
		return
	}
	group := errGroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		game := new(model.Game)
		if err = s.dao.DB.Where("id=?", data.Gid).First(&game).Error; err != nil {
			log.Error("RankInfo Error (%v)", err)
		}
		if game != nil && game.Title != "" {
			gameName = game.Title
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		season := new(model.Season)
		if err = s.dao.DB.Where("id=?", data.Sid).First(&season).Error; err != nil {
			log.Error("RankInfo Error (%v)", err)
			return err
		}
		if season != nil && season.Title != "" {
			seasonName = season.Title
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		return
	}
	data.GameName = gameName
	data.SeasonName = seasonName
	return
}

// RankList .
func (s *Service) RankList(c context.Context, gid int64) (list []*model.SeasonInfo, count int64, err error) {
	var (
		SeasonRanks []*model.SeasonRank
		seasonIDs   []int64
		seasons     []*model.Season
		gameMap     map[int64][]*model.Game
		rankMap     map[int64]*model.SeasonRank
	)
	source := s.dao.DB.Model(&model.SeasonRank{})
	source = source.Where("is_deleted=?", _notDeleted)
	source = source.Where("gid=?", gid)
	source.Count(&count)
	if err = source.Find(&SeasonRanks).Error; err != nil {
		log.Error("RankList Error (%v)", err)
		return
	}
	if len(SeasonRanks) == 0 {
		list = _emptyRankList
		return
	}
	rankMap = make(map[int64]*model.SeasonRank, len(SeasonRanks))
	for _, v := range SeasonRanks {
		seasonIDs = append(seasonIDs, v.Sid)
		rankMap[v.Sid] = v
	}
	seasonIDs = unique(seasonIDs)
	group := errGroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		sourceSeason := s.dao.DB.Model(&model.Season{})
		if err = sourceSeason.Where("id IN(?)", seasonIDs).Find(&seasons).Error; err != nil {
			log.Error("RankList Error (%v)", err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if gameMap, err = s.gameList(model.TypeSeason, seasonIDs); err != nil {
			return err
		}
		return nil
	})
	if err = group.Wait(); err != nil {
		return
	}
	for _, v := range seasons {
		var (
			seasonRank int64
			rankID     int64
		)
		sid := v.ID
		if sRank, ok := rankMap[sid]; ok {
			seasonRank = sRank.Rank
			rankID = sRank.ID
		}
		v.Platforms = s.platforms(v.SyncPlatform)
		v.ID = rankID
		if games, ok := gameMap[sid]; ok {
			list = append(list, &model.SeasonInfo{Season: v, Games: games, SeasonRank: seasonRank, RankID: rankID})
		} else {
			list = append(list, &model.SeasonInfo{Season: v, Games: _emptyGameList, SeasonRank: seasonRank, RankID: rankID})
		}
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].SeasonRank != list[j].SeasonRank {
			return list[i].SeasonRank > list[j].SeasonRank
		}
		return list[i].RankID > list[j].RankID
	})
	return
}

// AddSeasonRank .
func (s *Service) AddSeasonRank(c context.Context, param *model.SeasonRank) (err error) {
	var datas []*model.SeasonRank
	if err = s.dao.DB.Where("is_deleted=?", _notDeleted).Find(&datas).Error; err != nil {
		log.Error("AddSeasonRank s.dao.DB.Find Error (%v)", err)
		return
	}
	if len(datas) > _maxSeasonRank {
		return fmt.Errorf("最多添加20个赛季")
	}
	for _, v := range datas {
		if v.Gid == param.Gid && v.Sid == param.Sid {
			return fmt.Errorf("赛季已存在")
		}
	}
	if err = s.dao.DB.Model(&model.SeasonRank{}).Create(param).Error; err != nil {
		log.Error("AddSeasonRank s.dao.DB.Model Create(%+v) error(%v)", param, err)
	}
	return
}

// EditSeasonRank .
func (s *Service) EditSeasonRank(c context.Context, param *model.SeasonRank) (err error) {
	if param.ID <= 0 {
		return fmt.Errorf("id不存在")
	}
	//check name not repeat.
	preData := new(model.SeasonRank)
	s.dao.DB.Where("id != ?", param.ID).Where("gid=?", param.Gid).Where("sid=?", param.Sid).Where("is_deleted=?", _notDeleted).First(&preData)
	if preData.ID > 0 {
		return fmt.Errorf("赛季已存在")
	}
	if err = s.dao.DB.Model(&model.SeasonRank{}).Update(param).Error; err != nil {
		log.Error("EditSeasonRank s.dao.DB.Model Update(%+v) error(%v)", param, err)
	}
	return
}

func (s *Service) BatchRefreshContestDataPageCacheBySeasonId(c context.Context, seasonId int64) error {
	contests, err := s.listContestBySeason(seasonId)
	if err != nil {
		return err
	}
	cids := make([]int64, 0, len(contests))
	for _, contest := range contests {
		cids = append(cids, contest.ID)
	}
	return s.BatchRefreshContestDataPageCache(c, cids)
}

func (s *Service) getTeamAndSeasonById(teamId, seasonId int64) (*model.Team, *model.Season, error) {
	//make sure team exist.
	team := &model.Team{ID: teamId}
	if err := s.dao.DB.First(&team).Error; err != nil {
		return nil, nil, fmt.Errorf("get team by id %v error: %v", teamId, err)
	}
	//make sure season exist.
	season := &model.Season{ID: seasonId}
	if err := s.dao.DB.First(&season).Error; err != nil {
		return nil, nil, fmt.Errorf("get season by id %v error: %v", seasonId, err)
	}
	return team, season, nil
}

// AddTeamToSeason .
func (s *Service) AddTeamToSeason(c context.Context, teamId, seasonId, rank int64) error {
	if teamId <= 0 || seasonId <= 0 {
		return fmt.Errorf("team id or season id illegal")
	}

	_, _, err := s.getTeamAndSeasonById(teamId, seasonId)
	if err != nil {
		return err
	}
	teamInSeason := model.NewTeamInSeason(seasonId, teamId, rank)
	if !s.dao.DB.Find(&model.TeamInSeason{
		Sid: seasonId,
		Tid: teamId,
	}).RecordNotFound() {
		return fmt.Errorf("赛季下已存在该队伍")
	}
	if err := s.dao.DB.Create(&teamInSeason).Error; err != nil {
		return err
	}
	go s.espClient.ClearCache(context.Background(), &v1.ClearCacheRequest{
		CacheType: v1.ClearCacheType_TEAMS_IN_SEASON,
		CacheKeys: []int64{seasonId},
	})
	return nil
}

// ListTeamInSeason .
func (s *Service) ListTeamInSeason(c context.Context, seasonId int64) ([]*model.TeamInSeasonResponse, error) {
	res := make([]*model.TeamInSeasonResponse, 0)
	if seasonId <= 0 {
		return res, fmt.Errorf("season id illegal")
	}
	teamsInSeason := make([]*model.TeamInSeason, 0)
	if err := s.dao.DB.Model(model.TeamInSeason{}).Where("sid=?", seasonId).Find(&teamsInSeason).Error; err != nil {
		return res, err
	}
	if len(teamsInSeason) == 0 {
		return res, nil
	}
	teamSeasonMap := make(map[int64]*model.TeamInSeason, len(teamsInSeason))
	tids := make([]int64, 0, len(teamsInSeason))
	for _, teamInSeason := range teamsInSeason {
		teamSeasonMap[teamInSeason.Tid] = teamInSeason
		tids = append(tids, teamInSeason.Tid)
	}
	teams := make([]*model.Team, 0)
	if err := s.dao.DB.Where("id in (?)", tids).Find(&teams).Error; err != nil {
		return res, err
	}
	for _, team := range teams {
		res = append(res, &model.TeamInSeasonResponse{
			Team: team,
			Rank: teamSeasonMap[team.ID].Rank,
		})
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Rank > res[j].Rank {
			return true
		} else if res[i].Rank == res[j].Rank && res[i].ID < res[j].ID {
			return true
		}
		return false
	})
	return res, nil
}

// RemoveTeamFromSeason .
func (s *Service) RemoveTeamFromSeason(c context.Context, teamId, seasonId int64) error {
	if teamId <= 0 || seasonId <= 0 {
		return fmt.Errorf("team id or season id illegal")
	}

	_, _, err := s.getTeamAndSeasonById(teamId, seasonId)
	if err != nil {
		return err
	}
	tmpTeamInSeason := model.NewTeamInSeason(seasonId, teamId, 0)
	//api is delete, so this teamInSeason must exist in db.
	if err := s.dao.DB.Where("tid=?", teamId).Where("sid=?", seasonId).Find(&tmpTeamInSeason).Error; err != nil {
		return fmt.Errorf("no such releation teamId=%v, seasonId=%v", teamId, seasonId)
	}
	teamInSeason := model.NewTeamInSeason(seasonId, teamId, 0)
	if err := s.dao.DB.Delete(&teamInSeason).Error; err != nil {
		return err
	}
	go s.espClient.ClearCache(context.Background(), &v1.ClearCacheRequest{
		CacheType: v1.ClearCacheType_TEAMS_IN_SEASON,
		CacheKeys: []int64{seasonId},
	})
	return nil
}

// UpdateTeamInSeason .
func (s *Service) UpdateTeamInSeason(c context.Context, teamId, seasonId, rank int64) error {
	if teamId <= 0 || seasonId <= 0 {
		return fmt.Errorf("team id or season id illegal")
	}

	_, _, err := s.getTeamAndSeasonById(teamId, seasonId)
	if err != nil {
		return err
	}
	tmpTeamInSeason := model.NewTeamInSeason(seasonId, teamId, 0)
	//api is update, so this teamInSeason must exist in db.
	if err := s.dao.DB.Where("tid=?", teamId).Where("sid=?", seasonId).Find(&tmpTeamInSeason).Error; err != nil {
		return fmt.Errorf("no such releation teamId=%v, seasonId=%v", teamId, seasonId)
	}
	//update in db
	teamInSeason := model.NewTeamInSeason(seasonId, teamId, rank)
	if err := s.dao.DB.Where("tid=?", teamId).Where("sid=?", seasonId).Save(&teamInSeason).Error; err != nil {
		return err
	}
	go s.espClient.ClearCache(context.Background(), &v1.ClearCacheRequest{
		CacheType: v1.ClearCacheType_TEAMS_IN_SEASON,
		CacheKeys: []int64{seasonId},
	})
	return nil
}

// RebuildTeamInSeason .
func (s *Service) RebuildTeamInSeason(c context.Context) error {
	contests := make([]*model.Contest, 0)
	if err := s.dao.DB.Find(&contests).Error; err != nil {
		return err
	}
	for _, contest := range contests {
		if err := s.dao.DB.Exec(_insertTeamInSeasonIfNotExist, contest.Sid, contest.HomeID).Error; err != nil {
			log.Errorc(c, "rebuild teams_in_season fail: %v", err)
			return err
		}
		if err := s.dao.DB.Exec(_insertTeamInSeasonIfNotExist, contest.Sid, contest.AwayID).Error; err != nil {
			log.Errorc(c, "rebuild teams_in_season fail: %v", err)
			return err
		}
	}
	log.Infoc(c, "rebuild teams_in_season success")
	return nil
}

// RebuildTeamInSeason .
func (s *Service) RebuildTeamInSeasonInBackground(c context.Context) (string, error) {
	go s.RebuildTeamInSeason(context.Background())
	return "trigger rebuild for teams_in_season success, please check log", nil
}

func (s *Service) ClearSeasonCache(ctx context.Context, seasonId int64) {
	//s.cache.Do(ctx, func(c context.Context) {
	//	_, _ = client.EsportsServiceClient.ClearSeasonCache(service.Ctx4Worker, &v12.ClearSeasonCacheReq{
	//		SeasonId: seasonId,
	//	})
	//})
}
