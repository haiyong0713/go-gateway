package service

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/job/service"
	"go-gateway/app/web-svr/esports/admin/client"
	v12 "go-gateway/app/web-svr/esports/service/api/v1"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/web-svr/esports/admin/model"
	"go-gateway/app/web-svr/esports/ecode"
	espclient "go-gateway/app/web-svr/esports/interface/api/v1"
)

var _emptyTeamList = make([]*model.Team, 0)

// TeamInfo .
func (s *Service) TeamInfo(c context.Context, id int64) (data *model.TeamInfo, err error) {
	var gameMap map[int64][]*model.Game
	team := new(model.Team)
	if err = s.dao.DB.Where("id=?", id).Where("is_deleted=?", _notDeleted).First(&team).Error; err != nil {
		log.Error("TeamInfo Error (%v)", err)
		return
	}
	if gameMap, err = s.gameList(model.TypeTeam, []int64{id}); err != nil {
		return
	}
	if games, ok := gameMap[id]; ok {
		data = &model.TeamInfo{Team: team, Games: games}
	} else {
		data = &model.TeamInfo{Team: team, Games: _emptyGameList}
	}
	return
}

// TeamList .
func (s *Service) TeamList(c context.Context, pn, ps int64, title string, status int) (list []*model.TeamInfo, count int64, err error) {
	var (
		teams   []*model.Team
		teamIDs []int64
		gameMap map[int64][]*model.Game
	)
	source := s.dao.DB.Model(&model.Team{})
	if status != _statusAll {
		source = source.Where("is_deleted=?", _notDeleted)
	}
	if title != "" {
		source = source.Where("title like ?", "%"+title+"%")
	}
	source.Count(&count)
	if err = source.Offset((pn - 1) * ps).Limit(ps).Find(&teams).Error; err != nil {
		log.Error("TeamList Error (%v)", err)
		return
	}
	if len(teams) == 0 {
		return
	}
	for _, v := range teams {
		teamIDs = append(teamIDs, v.ID)
	}
	if gameMap, err = s.gameList(model.TypeTeam, teamIDs); err != nil {
		return
	}
	for _, v := range teams {
		if games, ok := gameMap[v.ID]; ok {
			list = append(list, &model.TeamInfo{Team: v, Games: games})
		} else {
			list = append(list, &model.TeamInfo{Team: v, Games: _emptyGameList})
		}
	}
	return
}

// AddTeam .
func (s *Service) AddTeam(c context.Context, param *model.Team, gids []int64) (err error) {
	var (
		games   []*model.Game
		gidMaps []*model.GIDMap
	)
	if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN (?)", gids).Find(&games).Error; err != nil {
		log.Error("AddTeam check game ids Error (%v)", err)
		return
	}
	if len(games) == 0 {
		log.Error("AddTeam games(%v) not found", gids)
		err = xecode.RequestErr
		return
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	reply := &model.Reply{
		Business: model.ReplyTeamBus,
	}
	if err = tx.Model(&model.Reply{}).Create(reply).Error; err != nil {
		log.Error("AddTeam dao.Reply Create(%+v) error(%v)", reply, err)
		err = tx.Rollback().Error
		return
	}
	param.ReplyID = reply.ID
	if err = tx.Model(&model.Team{}).Create(param).Error; err != nil {
		log.Error("AddTeam s.dao.DB.Model Create(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	for _, v := range games {
		gidMaps = append(gidMaps, &model.GIDMap{Type: model.TypeTeam, Oid: param.ID, Gid: v.ID})
	}
	sql, sqlParam := model.GidBatchAddSQL(gidMaps)
	if err = tx.Model(&model.GIDMap{}).Exec(sql, sqlParam...).Error; err != nil {
		log.Error("AddTeam s.dao.DB.Model Create(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	err = tx.Commit().Error
	if err == nil {
		go func() {
			_ = s.ClearESportCacheByType(espclient.ClearCacheType_TEAM, []int64{param.ID})
		}()
	}

	// register reply
	if err = s.dao.RegReply(c, param.ReplyID, param.Adid, model.ReplyID); err != nil {
		err = nil
	}
	go s.BatchRefreshContestDataPageCacheByTeamId(context.Background(), param.ID)
	return
}

// EditTeam .
func (s *Service) EditTeam(c context.Context, param *model.Team, gids []int64) (err error) {
	var (
		games                    []*model.Game
		preGidMaps, addGidMaps   []*model.GIDMap
		upGidMapAdd, upGidMapDel []int64
		contests                 []*model.Contest
	)
	preData := new(model.Team)
	if err = s.dao.DB.Where("id=?", param.ID).First(&preData).Error; err != nil {
		log.Error("EditTeam s.dao.DB.Where id(%d) error(%d)", param.ID, err)
		return
	}

	if contests, err = s.listContest(param.ID); err == nil && len(contests) > 0 {
		var cids []int64
		var contestMap = make(map[int64]*model.Contest)
		for _, c := range contests {
			cids = append(cids, c.ID)
			contestMap[c.ID] = c
		}
		log.Warn("poster contestIDs:%+v", cids)
		rep, err := s.espClient.LiveContests(context.Background(), &espclient.LiveContestsRequest{
			Cids: cids,
		})
		if err != nil {
			goto TEAMUPDATE
		}
		if rep != nil && len(rep.Contests) > 0 {
			eg := errgroup.WithContext(c)
			for _, c := range rep.Contests {
				if c.GameState <= 2 {
					continue
				}
				contest := contestMap[c.ID]
				eg.Go(func(c context.Context) error {
					if err = s.DrawPost(c, param, contest); err != nil {
						return err
					}
					return nil
				})
			}
			if err = eg.Wait(); err != nil {
				return err
			}
			if _, err = s.espClient.RefreshContestDataPageCache(c, &espclient.RefreshContestDataPageCacheRequest{
				Cids: cids,
			}); err != nil {
				return err
			}
		}
	}

TEAMUPDATE:
	if err = s.dao.DB.Model(&model.Game{}).Where("status=?", _statusOn).Where("id IN (?)", gids).Find(&games).Error; err != nil {
		log.Error("EditTeam check game ids Error (%v)", err)
		return
	}
	if len(games) == 0 {
		log.Error("EditTeam games(%v) not found", gids)
		err = xecode.RequestErr
		return
	}
	if err = s.dao.DB.Model(&model.GIDMap{}).Where("oid=?", param.ID).Where("type=?", model.TypeTeam).Find(&preGidMaps).Error; err != nil {
		log.Error("EditTeam games(%v) not found", gids)
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
			addGidMaps = append(addGidMaps, &model.GIDMap{Type: model.TypeTeam, Oid: param.ID, Gid: gid})
		}
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&model.Team{}).Save(param).Error; err != nil {
		log.Error("EditTeam Team Update(%+v) error(%v)", param, err)
		err = tx.Rollback().Error
		return
	}
	if len(upGidMapAdd) > 0 {
		if err = tx.Model(&model.GIDMap{}).Where("id IN (?)", upGidMapAdd).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("EditTeam GIDMap Add(%+v) error(%v)", upGidMapAdd, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(upGidMapDel) > 0 {
		if err = tx.Model(&model.GIDMap{}).Where("id IN (?)", upGidMapDel).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("EditTeam GIDMap Del(%+v) error(%v)", upGidMapDel, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(addGidMaps) > 0 {
		sql, sqlParam := model.GidBatchAddSQL(addGidMaps)
		if err = tx.Model(&model.GIDMap{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("EditTeam GIDMap Create(%+v) error(%v)", addGidMaps, err)
			err = tx.Rollback().Error
			return
		}
	}
	err = tx.Commit().Error
	if err == nil {
		s.ClearTeamCache(c, param.ID)
		go func() {
			_ = s.ClearESportCacheByType(espclient.ClearCacheType_TEAM, []int64{param.ID})
			relatedSeasons := make([]*model.TeamInSeason, 0)
			seasonIds := make([]int64, 0)
			if err := s.dao.DB.Where("tid=?", param.ID).Find(&relatedSeasons).Error; err != nil {
				log.Error("EditTeamRefresh cache failed: %v", err)
				return
			}
			for _, season := range relatedSeasons {
				seasonIds = append(seasonIds, season.Sid)
			}
			log.Info("EditTeamRefresh trigger for sid %v", seasonIds)
			_ = s.ClearESportCacheByType(espclient.ClearCacheType_TEAMS_IN_SEASON, seasonIds)
		}()
	}

	return
}

// ForbidTeam .
func (s *Service) ForbidTeam(c context.Context, id int64, state int) (err error) {
	preTeam := new(model.Team)
	if err = s.dao.DB.Where("id=?", id).First(&preTeam).Error; err != nil {
		log.Error("TeamForbid s.dao.DB.Where id(%d) error(%d)", id, err)
		return
	}
	if err = s.dao.DB.Model(&model.Team{}).Where("id=?", id).Update(map[string]int{"is_deleted": state}).Error; err != nil {
		log.Error("TeamForbid s.dao.DB.Model error(%v)", err)
	}
	s.ClearTeamCache(c, id)
	go s.BatchRefreshContestDataPageCacheByTeamId(context.Background(), id)
	return
}

func (s *Service) DrawPost(c context.Context, team *model.Team, contest *model.Contest) (err error) {
	var (
		templateID = 2
		materials  []*model.Material
	)
	// 头图
	if team.PictureUrl != "" {
		templateID = 1
		materials = append(materials, &model.Material{
			ID:   len(materials) + 1,
			Type: 1,
			Data: team.PictureUrl,
		})
	}

	// 比赛标题
	tmpSeason := new(model.Season)
	if err = s.dao.DB.Where("id=?", contest.Sid).First(&tmpSeason).Error; err == nil && tmpSeason.ID > 0 {
		materials = append(materials, &model.Material{
			ID:   len(materials) + 1,
			Type: 2,
			Data: fmt.Sprintf("%s %s", tmpSeason.Title, contest.GameStage),
		})
	} else {
		materials = append(materials, &model.Material{
			ID:   len(materials) + 1,
			Type: 2,
			Data: fmt.Sprintf("%s %s %s", contest.GameStage, contest.GameStage1, contest.GameStage2),
		})
	}

	// 比赛时间
	t := time.Unix(contest.Stime, 0)
	materials = append(materials, &model.Material{
		ID:   len(materials) + 1,
		Type: 2,
		Data: fmt.Sprintf("%d月%d日 %02d:%02d", t.Month(), t.Day(), t.Hour(), t.Minute()),
	})
	//比赛主队
	homeData := new(model.Team)
	if err = s.dao.DB.Where("id=?", contest.HomeID).First(&homeData).Error; err != nil {
		log.Error("s.dao.DB.Where id(%d) error(%+v)", contest.HomeID, err)
		err = ecode.EsportsDrawPost
		return
	}
	materials = append(materials, &model.Material{
		ID:   len(materials) + 1,
		Type: 1,
		Data: homeData.Logo,
	})
	materials = append(materials, &model.Material{
		ID:   len(materials) + 1,
		Type: 2,
		Data: homeData.Title,
	})
	// 比赛客队
	awayData := new(model.Team)
	if err = s.dao.DB.Where("id=?", contest.AwayID).First(&awayData).Error; err != nil {
		log.Error("s.dao.DB.Where id(%d) error(%+v)", contest.AwayID, err)
		err = ecode.EsportsDrawPost
		return
	}
	materials = append(materials, &model.Material{
		ID:   len(materials) + 1,
		Type: 1,
		Data: awayData.Logo,
	})
	materials = append(materials, &model.Material{
		ID:   len(materials) + 1,
		Type: 2,
		Data: awayData.Title,
	})
	err = s.drawPost(c, contest.ID, templateID, materials, team)
	return
}

func (s *Service) drawPost(c context.Context, cid int64, templateID int, materials []*model.Material, team *model.Team) (err error) {
	materials = append(materials, &model.Material{
		ID:   len(materials) + 1,
		Type: 2,
		Data: team.Title,
	})

	var picture string
	if picture, err = s.dao.DrawPost(c, cid, templateID, materials); err != nil {
		return
	}
	if err = s.dao.SavePost(c, cid, team.Title, picture); err != nil {
		return
	}
	return
}

func (s *Service) BatchRefreshContestDataPageCacheByTeamId(c context.Context, teamId int64) error {
	contests, err := s.listContest(teamId)
	if err != nil {
		return err
	}
	cids := make([]int64, 0, len(contests))
	for _, contest := range contests {
		cids = append(cids, contest.ID)
	}
	return s.BatchRefreshContestDataPageCache(c, cids)
}

func (s *Service) ClearTeamCache(ctx context.Context, teamId int64) {
	s.cache.Do(ctx, func(c context.Context) {
		_, _ = client.EsportsServiceClient.ClearTeamCache(service.Ctx4Worker, &v12.ClearTeamCacheReq{
			TeamId: teamId,
		})
	})
}
