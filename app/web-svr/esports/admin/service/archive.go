package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	accwarden "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/model"
	"go-gateway/pkg/idsafe/bvid"

	errGroup "go-common/library/sync/errgroup.v2"

	"github.com/jinzhu/gorm"
)

const (
	_typeAdd  = "add"
	_typeEdit = "edit"
)

var (
	_emptyArcList = make([]*model.ArcResult, 0)
	_emptyRule    = make([]string, 0)
)

// ArcList archive list.
func (s *Service) ArcList(c context.Context, arg *model.ArcListParam) (arcs []*model.ArcResult, total int, err error) {
	var (
		list                                        []*model.SearchArc
		gids, tids, matchIDs, teamIDs, mids, arcIDs []int64
		hitRules                                    map[int64]*model.ArchiveRule
		hitTagMap, hitKeyMap                        map[int64]string
		hitMidMap                                   map[int64]int64
		games                                       []*model.Game
		tags                                        []*model.Tag
		matchs                                      []*model.Match
		teams                                       []*model.Team
		gameMap                                     map[int64]*model.Game
		tagMap                                      map[int64]*model.Tag
		matchMap                                    map[int64]*model.Match
		teamMap                                     map[int64]*model.Team
		infosReply                                  *accwarden.InfosReply
		ip                                          = metadata.String(c, metadata.RemoteIP)
	)
	if list, total, err = s.dao.SearchArc(c, arg); err != nil {
		return
	}
	if len(list) == 0 {
		arcs = _emptyArcList
		return
	}
	for _, arc := range list {
		arcIDs = append(arcIDs, arc.ID)
		if len(arc.Gid) > 0 {
			gids = append(gids, arc.Gid...)
		}
		if len(arc.Tags) > 0 {
			tids = append(tids, arc.Tags...)
		}
		if len(arc.Matchs) > 0 {
			matchIDs = append(matchIDs, arc.Matchs...)
		}
		if len(arc.Teams) > 0 {
			teamIDs = append(teamIDs, arc.Teams...)
		}
		if arc.Mid > 0 {
			mids = append(mids, arc.Mid)
		}
	}
	group := errGroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if hitRules, hitTagMap, hitKeyMap, hitMidMap, err = s.haveRules(arcIDs); err != nil {
			log.Error("ArcList s.haveRules arcsID(%+v) error(%v)", arcIDs, err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if len(gids) > 0 {
			if err = s.dao.DB.Model(&model.Game{}).Where("id IN (?)", unique(gids)).Find(&games).Error; err != nil {
				log.Error("ArcList Game gids(%+v) error(%v)", gids, err)
				return err
			}
			if gl := len(games); gl > 0 {
				gameMap = make(map[int64]*model.Game, gl)
				for _, v := range games {
					gameMap[v.ID] = v
				}
			}
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if len(tids) > 0 {
			if err = s.dao.DB.Model(&model.Tag{}).Where("id IN (?)", unique(tids)).Find(&tags).Error; err != nil {
				log.Error("ArcList Tag tids(%+v) error(%v)", tids, err)
				return err
			}
			if tl := len(tags); tl > 0 {
				tagMap = make(map[int64]*model.Tag, tl)
				for _, v := range tags {
					tagMap[v.ID] = v
				}
			}
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if len(matchIDs) > 0 {
			if err = s.dao.DB.Model(&model.Match{}).Where("id IN (?)", unique(matchIDs)).Find(&matchs).Error; err != nil {
				log.Error("ArcList Match ids(%+v) error(%v)", tids, err)
				return err
			}
			if tl := len(matchs); tl > 0 {
				matchMap = make(map[int64]*model.Match, tl)
				for _, v := range matchs {
					matchMap[v.ID] = v
				}
			}
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if len(teamIDs) > 0 {
			if err = s.dao.DB.Model(&model.Team{}).Where("id IN (?)", unique(teamIDs)).Find(&teams).Error; err != nil {
				log.Error("ArcList Team ids(%+v) error(%v)", tids, err)
				return err
			}
			if tl := len(teams); tl > 0 {
				teamMap = make(map[int64]*model.Team, tl)
				for _, v := range teams {
					teamMap[v.ID] = v
				}
			}
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if len(mids) > 0 {
			if infosReply, err = s.accClient.Infos3(c, &accwarden.MidsReq{Mids: unique(mids), RealIp: ip}); err != nil {
				log.Error("账号Infos3:grpc错误 s.accClient.Infos3 error(%v)", err)
				return err
			}
		}
		return nil
	})
	err = group.Wait()
	var (
		bvidStr string
		bvidErr error
	)
	for _, v := range list {
		if bvidStr, bvidErr = bvid.AvToBv(v.Aid); bvidErr != nil {
			log.Error("ArcList.AvToBv(%d) error(%v)", v.Aid, bvidErr)
			continue
		}
		arcRes := &model.ArcResult{
			ID:           v.ID,
			Aid:          v.Aid,
			BvID:         bvidStr,
			TypeID:       v.TypeID,
			Title:        v.Title,
			State:        v.State,
			Mid:          v.Mid,
			Years:        v.Year,
			Source:       v.Source,
			Ctime:        v.Ctime,
			RuleTags:     _emptyRule,
			RuleKeywords: _emptyRule,
		}
		// 命中白名单规则
		if upMid, ok := hitMidMap[v.ID]; ok {
			arcRes.RuleMid = upMid
			//TODO es接入后删除
			arcRes.Source = 1
		}
		// 是否命中自动导入规则
		if hitRule, ok := hitRules[v.ID]; ok {
			for _, id := range hitRule.HitTags {
				// 根据id获取tag
				if name, tagOk := hitTagMap[id]; tagOk {
					arcRes.RuleTags = append(arcRes.RuleTags, name)
					//TODO es接入后删除
					arcRes.Source = 1
				}
			}
			for _, id := range hitRule.HitKeys {
				// 根据id获取关键词
				if name, keyOk := hitKeyMap[id]; keyOk {
					arcRes.RuleKeywords = append(arcRes.RuleKeywords, name)
					//TODO es接入后删除
					arcRes.Source = 1
				}
			}
		}
		if infosReply != nil && len(infosReply.Infos) > 0 {
			if info, ok := infosReply.Infos[v.Mid]; ok && info != nil {
				arcRes.Uname = info.Name
			}
		}
		if len(gameMap) > 0 {
			for _, gid := range v.Gid {
				if game, ok := gameMap[gid]; ok {
					arcRes.Games = append(arcRes.Games, game)
				}
			}
		}
		if len(tagMap) > 0 {
			for _, tid := range v.Tags {
				if tag, ok := tagMap[tid]; ok {
					arcRes.Tags = append(arcRes.Tags, tag)
				}
			}
		}
		if len(matchMap) > 0 {
			for _, tid := range v.Matchs {
				if match, ok := matchMap[tid]; ok {
					arcRes.Matchs = append(arcRes.Matchs, match)
				}
			}
		}
		if len(teamMap) > 0 {
			for _, tid := range v.Teams {
				if team, ok := teamMap[tid]; ok {
					arcRes.Teams = append(arcRes.Teams, team)
				}
			}
		}
		if len(arcRes.Games) == 0 {
			arcRes.Games = _emptyGameList
		}
		if len(arcRes.Tags) == 0 {
			arcRes.Tags = _emptyTagList
		}
		if len(arcRes.Matchs) == 0 {
			arcRes.Matchs = _emptyMatchList
		}
		if len(arcRes.Teams) == 0 {
			arcRes.Teams = _emptyTeamList
		}
		arcs = append(arcs, arcRes)
	}
	return
}

func (s *Service) haveRules(arcIDs []int64) (hitRules map[int64]*model.ArchiveRule, hitTagMap, hitKeyMap map[int64]string, hitMidMap map[int64]int64, err error) {
	var (
		hitTags, hitKeys []int64
		arcHits          []*model.EsArchiveHit
		ruleTags         []*model.EsArchiveTag
		ruleKeys         []*model.EsArchiveKeyword
	)
	if err = s.dao.DB.Model(&model.EsArchiveHit{}).Where("arcs_id IN (?) AND is_deleted=0", unique(arcIDs)).Find(&arcHits).Error; err != nil {
		log.Error("ArcList ArcHit arcsID(%+v) error(%v)", arcIDs, err)
		return
	}
	hitCounts := len(arcHits)
	if hitCounts > 0 {
		hitRules = make(map[int64]*model.ArchiveRule, hitCounts)
		hitMidMap = make(map[int64]int64, hitCounts)
		for _, hit := range arcHits {
			if hit.WhiteMid > 0 {
				hitMidMap[hit.ArcsID] = hit.WhiteMid
			}
			hitRule := &model.ArchiveRule{}
			if hit.TagIDs != "" {
				if tids, e := xstr.SplitInts(hit.TagIDs); e != nil {
					log.Error("ArcList Tag IDs(%s) xstr.SplitInts error(%v) ", hit.TagIDs, e)
				} else {
					hitTags = append(hitTags, tids...)
					hitRule.HitTags = append(hitRule.HitTags, tids...)
				}
			}
			if hit.KeywordIDs != "" {
				if kids, e := xstr.SplitInts(hit.KeywordIDs); e != nil {
					log.Error("ArcList Keyword IDs(%s) xstr.SplitInts error(%v) ", hit.KeywordIDs, e)
				} else {
					hitKeys = append(hitKeys, kids...)
					hitRule.HitKeys = append(hitRule.HitKeys, kids...)
				}
			}
			hitRules[hit.ArcsID] = hitRule
		}
		if len(hitTags) > 0 {
			if err = s.dao.DB.Model(&model.EsArchiveTag{}).Where("is_deleted = ?", _notDeleted).Where("id IN (?)", unique(hitTags)).Find(&ruleTags).Error; err != nil {
				log.Error("ArcList EsArchiveTag ids(%+v) error(%v)", hitTags, err)
				return
			}
			hitTagMap = make(map[int64]string)
			for _, v := range ruleTags {
				hitTagMap[v.ID] = v.Tag
			}
		}
		if len(hitKeys) > 0 {
			if err = s.dao.DB.Model(&model.EsArchiveKeyword{}).Where("is_deleted = ?", _notDeleted).Where("id IN (?)", unique(hitKeys)).Find(&ruleKeys).Error; err != nil {
				log.Error("ArcList EsArchiveKeyword ids(%+v) error(%v)", hitKeys, err)
				return
			}
			hitKeyMap = make(map[int64]string)
			for _, v := range ruleKeys {
				hitKeyMap[v.ID] = v.Keyword
			}
		}
	}
	return
}

// EditArc edit archive.
func (s *Service) EditArc(c context.Context, arg *model.ArcImportParam) (err error) {
	var preArc *model.Arc
	if err = s.dao.DB.Model(&model.Arc{}).Where("aid = ?", arg.Aid).Where("is_deleted=?", _notDeleted).First(&preArc).Error; err != nil {
		log.Error("EditArc check aid Error (%v)", err)
		return
	}
	var data *model.ArcRelation
	if data, err = s.arcRelationChanges(arg, _typeEdit); err != nil {
		return
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	if err = upArcRelation(tx, data); err != nil {
		err = tx.Rollback().Error
		return
	}
	err = tx.Commit().Error
	return
}

// BatchAddArc batch add archive.
func (s *Service) BatchAddArc(c context.Context, param *model.ArcAddParam) (err error) {
	var (
		arcs                           []*model.Arc
		upAddAids, addAids, changeAids []int64
	)
	if err = s.dao.DB.Model(&model.Arc{}).Where("aid IN (?)", param.Aids).Find(&arcs).Error; err != nil {
		log.Error("BatchAddArc check aids Error (%v)", err)
		return
	}
	if len(arcs) > 0 {
		arcMap := make(map[int64]*model.Arc, len(param.Aids))
		for _, v := range arcs {
			arcMap[v.Aid] = v
		}
		for _, aid := range param.Aids {
			if arc, ok := arcMap[aid]; ok {
				if arc.IsDeleted == _deleted {
					upAddAids = append(upAddAids, arc.ID)
					changeAids = append(changeAids, arc.Aid)
				}
			} else {
				addAids = append(addAids, aid)
				changeAids = append(changeAids, aid)
			}
		}
	} else {
		addAids = param.Aids
		changeAids = param.Aids
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	if len(addAids) > 0 {
		sql, sqlParam := model.ArcBatchAddSQL(addAids)
		if err = tx.Model(&model.Arc{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("BatchAddArc Arc tx.Model Exec(%+v) error(%v)", addAids, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(changeAids) > 0 {
		arcRelation := new(model.ArcRelation)
		for _, aid := range changeAids {
			var data *model.ArcRelation
			arg := &model.ArcImportParam{
				Aid:      aid,
				Gids:     param.Gids,
				MatchIDs: param.MatchIDs,
				TagIDs:   param.TagIDs,
				TeamIDs:  param.TeamIDs,
				Years:    param.Years,
			}
			if data, err = s.arcRelationChanges(arg, _typeAdd); err != nil {
				return
			}
			arcRelation.UpAddGids = append(arcRelation.UpAddGids, data.UpAddGids...)
			arcRelation.UpDelGids = append(arcRelation.UpDelGids, data.UpDelGids...)
			arcRelation.AddGids = append(arcRelation.AddGids, data.AddGids...)
			arcRelation.UpAddMatchs = append(arcRelation.UpAddMatchs, data.UpAddMatchs...)
			arcRelation.UpDelMatchs = append(arcRelation.UpDelMatchs, data.UpDelMatchs...)
			arcRelation.AddMatchs = append(arcRelation.AddMatchs, data.AddMatchs...)
			arcRelation.UpAddTags = append(arcRelation.UpAddTags, data.UpAddTags...)
			arcRelation.UpDelTags = append(arcRelation.UpDelTags, data.UpDelTags...)
			arcRelation.AddTags = append(arcRelation.AddTags, data.AddTags...)
			arcRelation.UpAddTeams = append(arcRelation.UpAddTeams, data.UpAddTeams...)
			arcRelation.UpDelTeams = append(arcRelation.UpDelTeams, data.UpDelTeams...)
			arcRelation.AddTeams = append(arcRelation.AddTeams, data.AddTeams...)
			arcRelation.UpAddYears = append(arcRelation.UpAddYears, data.UpAddYears...)
			arcRelation.UpDelYears = append(arcRelation.UpDelYears, data.UpDelYears...)
			arcRelation.AddYears = append(arcRelation.AddYears, data.AddYears...)
		}
		if err = upArcRelation(tx, arcRelation); err != nil {
			err = tx.Rollback().Error
			return
		}
	}
	if len(upAddAids) > 0 {
		if err = tx.Model(&model.Arc{}).Where("id IN (?)", upAddAids).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("BatchAddArc Save(%+v) error(%v)", upAddAids, err)
			err = tx.Rollback().Error
			return
		}
	}
	err = tx.Commit().Error
	return
}

// BatchEditArc batch add arc.
func (s *Service) BatchEditArc(c context.Context, param *model.ArcAddParam) (err error) {
	var (
		arcs       []*model.Arc
		changeAids []int64
	)
	if err = s.dao.DB.Model(&model.Arc{}).Where("aid IN (?)", param.Aids).Find(&arcs).Error; err != nil {
		log.Error("BatchEditArc check aids Error (%v)", err)
		return
	}
	if len(arcs) > 0 {
		arcMap := make(map[int64]*model.Arc, len(param.Aids))
		for _, v := range arcs {
			arcMap[v.Aid] = v
		}
		for _, aid := range param.Aids {
			if arc, ok := arcMap[aid]; ok {
				if arc.IsDeleted == _notDeleted || arc.IsDeleted == _checkPass {
					changeAids = append(changeAids, arc.Aid)
				}
			}
		}
	}
	if len(changeAids) == 0 {
		err = fmt.Errorf("稿件状态未改变")
		return
	}
	arcRelation := new(model.ArcRelation)
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	for _, aid := range changeAids {
		var data *model.ArcRelation
		arg := &model.ArcImportParam{
			Aid:      aid,
			Gids:     param.Gids,
			MatchIDs: param.MatchIDs,
			TagIDs:   param.TagIDs,
			TeamIDs:  param.TeamIDs,
			Years:    param.Years,
		}
		if data, err = s.arcRelationChanges(arg, _typeEdit); err != nil {
			return
		}
		arcRelation.UpAddGids = append(arcRelation.UpAddGids, data.UpAddGids...)
		arcRelation.UpDelGids = append(arcRelation.UpDelGids, data.UpDelGids...)
		arcRelation.AddGids = append(arcRelation.AddGids, data.AddGids...)
		arcRelation.UpAddMatchs = append(arcRelation.UpAddMatchs, data.UpAddMatchs...)
		arcRelation.UpDelMatchs = append(arcRelation.UpDelMatchs, data.UpDelMatchs...)
		arcRelation.AddMatchs = append(arcRelation.AddMatchs, data.AddMatchs...)
		arcRelation.UpAddTags = append(arcRelation.UpAddTags, data.UpAddTags...)
		arcRelation.UpDelTags = append(arcRelation.UpDelTags, data.UpDelTags...)
		arcRelation.AddTags = append(arcRelation.AddTags, data.AddTags...)
		arcRelation.UpAddTeams = append(arcRelation.UpAddTeams, data.UpAddTeams...)
		arcRelation.UpDelTeams = append(arcRelation.UpDelTeams, data.UpDelTeams...)
		arcRelation.AddTeams = append(arcRelation.AddTeams, data.AddTeams...)
		arcRelation.UpAddYears = append(arcRelation.UpAddYears, data.UpAddYears...)
		arcRelation.UpDelYears = append(arcRelation.UpDelYears, data.UpDelYears...)
		arcRelation.AddYears = append(arcRelation.AddYears, data.AddYears...)
	}
	if err = upArcRelation(tx, arcRelation); err != nil {
		err = tx.Rollback().Error
		return
	}
	err = tx.Commit().Error
	return
}

// ArcImportCSV archive import.
func (s *Service) ArcImportCSV(c context.Context, list []*model.ArcImportParam) (err error) {
	var (
		aids, addAids, changeAids, saveAids []int64
		arcs                                []*model.Arc
	)
	listMap := make(map[int64]*model.ArcImportParam, len(aids))
	for _, v := range list {
		aids = append(aids, v.Aid)
		listMap[v.Aid] = v
	}
	if err = s.dao.DB.Model(&model.Arc{}).Where("aid IN (?)", aids).Find(&arcs).Error; err != nil {
		log.Error("arcImport check aids Error (%v)", err)
		return
	}
	if len(arcs) > 0 {
		arcMap := make(map[int64]*model.Arc, len(aids))
		for _, v := range arcs {
			arcMap[v.Aid] = v
		}
		for _, aid := range aids {
			if arc, ok := arcMap[aid]; ok {
				if arc.IsDeleted == _deleted {
					saveAids = append(saveAids, arc.ID)
					changeAids = append(changeAids, arc.Aid)
				}
			} else {
				addAids = append(addAids, aid)
				changeAids = append(changeAids, aid)
			}
		}
	} else {
		addAids = aids
		changeAids = aids
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	if len(addAids) > 0 {
		sql, sqlParam := model.ArcBatchAddSQL(addAids)
		if err = tx.Model(&model.Arc{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("arcImport Arc tx.Model Exec(%+v) error(%v)", addAids, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(saveAids) > 0 {
		if err = tx.Model(&model.Arc{}).Where("id IN (?)", saveAids).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("arcImport Save(%+v) error(%v)", saveAids, err)
			err = tx.Rollback().Error
			return
		}
	}
	if len(changeAids) > 0 {
		arcRelation := new(model.ArcRelation)
		for _, aid := range changeAids {
			if arc, ok := listMap[aid]; ok {
				var data *model.ArcRelation
				arg := &model.ArcImportParam{
					Aid:      aid,
					Gids:     arc.Gids,
					MatchIDs: arc.MatchIDs,
					TagIDs:   arc.TagIDs,
					TeamIDs:  arc.TeamIDs,
					Years:    arc.Years,
				}
				if data, err = s.arcRelationChanges(arg, _typeAdd); err != nil {
					return
				}
				arcRelation.UpAddGids = append(arcRelation.UpAddGids, data.UpAddGids...)
				arcRelation.UpDelGids = append(arcRelation.UpDelGids, data.UpDelGids...)
				arcRelation.AddGids = append(arcRelation.AddGids, data.AddGids...)
				arcRelation.UpAddMatchs = append(arcRelation.UpAddMatchs, data.UpAddMatchs...)
				arcRelation.UpDelMatchs = append(arcRelation.UpDelMatchs, data.UpDelMatchs...)
				arcRelation.AddMatchs = append(arcRelation.AddMatchs, data.AddMatchs...)
				arcRelation.UpAddTags = append(arcRelation.UpAddTags, data.UpAddTags...)
				arcRelation.UpDelTags = append(arcRelation.UpDelTags, data.UpDelTags...)
				arcRelation.AddTags = append(arcRelation.AddTags, data.AddTags...)
				arcRelation.UpAddTeams = append(arcRelation.UpAddTeams, data.UpAddTeams...)
				arcRelation.UpDelTeams = append(arcRelation.UpDelTeams, data.UpDelTeams...)
				arcRelation.AddTeams = append(arcRelation.AddTeams, data.AddTeams...)
				arcRelation.UpAddYears = append(arcRelation.UpAddYears, data.UpAddYears...)
				arcRelation.UpDelYears = append(arcRelation.UpDelYears, data.UpDelYears...)
				arcRelation.AddYears = append(arcRelation.AddYears, data.AddYears...)
			}
		}
		if err = upArcRelation(tx, arcRelation); err != nil {
			err = tx.Rollback().Error
			return
		}
	}
	err = tx.Commit().Error
	return
}

func (s *Service) arcRelationChanges(arg *model.ArcImportParam, typ string) (data *model.ArcRelation, err error) {
	data = new(model.ArcRelation)
	// add game map
	if len(arg.Gids) > 0 {
		var gidMaps []*model.GIDMap
		if err = s.dao.DB.Model(&model.GIDMap{}).Where("oid=?", arg.Aid).Where("gid IN (?)", arg.Gids).Where("type=?", model.TypeArc).Find(&gidMaps).Error; err != nil {
			log.Error("arcRelationChanges GIDMap s.dao.DB.Model MatchMap(%+v) error(%v)", arg.Gids, err)
			return
		}
		if len(gidMaps) > 0 {
			gidMap := make(map[int64]*model.GIDMap, len(gidMaps))
			for _, v := range gidMaps {
				gidMap[v.Gid] = v
			}
			for _, gid := range arg.Gids {
				if gidItem, ok := gidMap[gid]; ok {
					if gidItem.IsDeleted == _deleted {
						data.UpAddGids = append(data.UpAddGids, gidItem.ID)
					}
				} else {
					data.AddGids = append(data.AddGids, &model.GIDMap{Type: model.TypeArc, Gid: gid, Oid: arg.Aid})
				}
			}
		} else {
			for _, gid := range arg.Gids {
				data.AddGids = append(data.AddGids, &model.GIDMap{Type: model.TypeArc, Gid: gid, Oid: arg.Aid})
			}
		}
	}
	// add match map
	if len(arg.MatchIDs) > 0 {
		var matchs []*model.MatchMap
		if err = s.dao.DB.Model(&model.MatchMap{}).Where("aid=?", arg.Aid).Where("mid IN (?)", arg.MatchIDs).Find(&matchs).Error; err != nil {
			log.Error("arcRelationChanges Arc s.dao.DB.Model MatchMap(%+v) error(%v)", arg.MatchIDs, err)
			return
		}
		if len(matchs) > 0 {
			matchMap := make(map[int64]*model.MatchMap, len(matchs))
			for _, v := range matchs {
				matchMap[v.Mid] = v
			}
			for _, mid := range arg.MatchIDs {
				if match, ok := matchMap[mid]; ok {
					if match.IsDeleted == _deleted {
						data.UpAddMatchs = append(data.UpAddMatchs, match.ID)
					}
				} else {
					data.AddMatchs = append(data.AddMatchs, &model.MatchMap{Mid: mid, Aid: arg.Aid})
				}
			}
		} else {
			for _, mid := range arg.MatchIDs {
				data.AddMatchs = append(data.AddMatchs, &model.MatchMap{Mid: mid, Aid: arg.Aid})
			}
		}
	}
	// add tags map
	if len(arg.TagIDs) > 0 {
		var tags []*model.TagMap
		if err = s.dao.DB.Model(&model.TagMap{}).Where("aid=?", arg.Aid).Where("tid IN (?)", arg.TagIDs).Find(&tags).Error; err != nil {
			log.Error("arcRelationChanges Arc s.dao.DB.Model TagMap(%+v) error(%v)", arg.TagIDs, err)
			return
		}
		if len(tags) > 0 {
			tagMap := make(map[int64]*model.TagMap, len(tags))
			for _, v := range tags {
				tagMap[v.Tid] = v
			}
			for _, tid := range arg.TagIDs {
				if tag, ok := tagMap[tid]; ok {
					if tag.IsDeleted == _deleted {
						data.UpAddTags = append(data.UpAddTags, tag.ID)
					}
				} else {
					data.AddTags = append(data.AddTags, &model.TagMap{Tid: tid, Aid: arg.Aid})
				}
			}
		} else {
			for _, tid := range arg.TagIDs {
				data.AddTags = append(data.AddTags, &model.TagMap{Tid: tid, Aid: arg.Aid})
			}
		}
	}
	// add teams map
	if len(arg.TeamIDs) > 0 {
		var teams []*model.TeamMap
		if err = s.dao.DB.Model(&model.TeamMap{}).Where("aid=?", arg.Aid).Where("tid IN (?)", arg.TeamIDs).Find(&teams).Error; err != nil {
			log.Error("arcRelationChanges Arc s.dao.DB.Model TeamMap(%+v) error(%v)", arg.TeamIDs, err)
			return
		}
		if len(teams) > 0 {
			teamMap := make(map[int64]*model.TeamMap, len(teams))
			for _, v := range teams {
				teamMap[v.Tid] = v
			}
			for _, teamID := range arg.TeamIDs {
				if team, ok := teamMap[teamID]; ok {
					if team.IsDeleted == _deleted {
						data.UpAddTeams = append(data.UpAddTeams, team.ID)
					}
				} else {
					data.AddTeams = append(data.AddTeams, &model.TeamMap{Tid: teamID, Aid: arg.Aid})
				}
			}
		} else {
			for _, teamID := range arg.TeamIDs {
				data.AddTeams = append(data.AddTeams, &model.TeamMap{Tid: teamID, Aid: arg.Aid})
			}
		}
	}
	// add year map
	if len(arg.Years) > 0 {
		var yearMaps []*model.YearMap
		if err = s.dao.DB.Model(&model.YearMap{}).Where("aid=?", arg.Aid).Where("year IN (?)", arg.Years).Find(&yearMaps).Error; err != nil {
			log.Error("arcRelationChanges Arc s.dao.DB.Model YearMap(%+v) error(%v)", arg.Years, err)
			return
		}
		if len(yearMaps) > 0 {
			yearMap := make(map[int64]*model.YearMap, len(yearMaps))
			for _, v := range yearMaps {
				yearMap[v.Year] = v
			}
			for _, yearID := range arg.Years {
				if year, ok := yearMap[yearID]; ok {
					if year.IsDeleted == _deleted {
						data.UpAddYears = append(data.UpAddYears, year.ID)
					}
				} else {
					data.AddYears = append(data.AddYears, &model.YearMap{Year: yearID, Aid: arg.Aid})
				}
			}
		} else {
			for _, yearID := range arg.Years {
				data.AddYears = append(data.AddYears, &model.YearMap{Year: yearID, Aid: arg.Aid})
			}
		}
	}
	if typ == _typeEdit {
		var (
			gidMaps []*model.GIDMap
			matchs  []*model.MatchMap
			tags    []*model.TagMap
			teams   []*model.TeamMap
			years   []*model.YearMap
		)
		// gid map
		if err = s.dao.DB.Model(&model.GIDMap{}).Where("oid=?", arg.Aid).Where("type=?", model.TypeArc).Where("is_deleted=?", _notDeleted).Find(&gidMaps).Error; err != nil {
			log.Error("arcRelationChanges GIDMap s.dao.DB.Model MatchMap(%+v) error(%v)", arg.Gids, err)
			return
		}
		if len(gidMaps) > 0 {
			if len(arg.Gids) > 0 {
				gidMap := make(map[int64]int64, len(arg.Gids))
				for _, gid := range arg.Gids {
					gidMap[gid] = gid
				}
				for _, v := range gidMaps {
					if _, ok := gidMap[v.Gid]; !ok {
						data.UpDelGids = append(data.UpDelGids, v.ID)
					}
				}
			} else {
				for _, v := range gidMaps {
					data.UpDelGids = append(data.UpDelGids, v.ID)
				}
			}
		}
		// match
		if err = s.dao.DB.Model(&model.MatchMap{}).Where("aid=?", arg.Aid).Where("is_deleted=?", _notDeleted).Find(&matchs).Error; err != nil {
			log.Error("arcRelationChanges Arc s.dao.DB.Model MatchMap(%+v) error(%v)", arg.MatchIDs, err)
			return
		}
		if len(matchs) > 0 {
			if len(arg.MatchIDs) > 0 {
				matchIDMap := make(map[int64]int64, len(arg.MatchIDs))
				for _, id := range arg.MatchIDs {
					matchIDMap[id] = id
				}
				for _, v := range matchs {
					if _, ok := matchIDMap[v.Mid]; !ok {
						data.UpDelMatchs = append(data.UpDelMatchs, v.ID)
					}
				}
			} else {
				for _, v := range matchs {
					data.UpDelMatchs = append(data.UpDelMatchs, v.ID)
				}
			}
		}
		// tag
		if err = s.dao.DB.Model(&model.TagMap{}).Where("aid=?", arg.Aid).Where("is_deleted=?", _notDeleted).Find(&tags).Error; err != nil {
			log.Error("arcRelationChanges Arc s.dao.DB.Model TagMap(%+v) error(%v)", arg.TagIDs, err)
			return
		}
		if len(tags) > 0 {
			if len(arg.TagIDs) > 0 {
				tagIDMap := make(map[int64]int64, len(arg.TagIDs))
				for _, id := range arg.TagIDs {
					tagIDMap[id] = id
				}
				for _, v := range tags {
					if _, ok := tagIDMap[v.Tid]; !ok {
						data.UpDelTags = append(data.UpDelTags, v.ID)
					}
				}
			} else {
				for _, v := range tags {
					data.UpDelTags = append(data.UpDelTags, v.ID)
				}
			}
		}
		// team
		if err = s.dao.DB.Model(&model.TeamMap{}).Where("aid=?", arg.Aid).Where("is_deleted=?", _notDeleted).Find(&teams).Error; err != nil {
			log.Error("arcRelationChanges Arc s.dao.DB.Model MatchMap(%+v) error(%v)", arg.MatchIDs, err)
			return
		}
		if len(teams) > 0 {
			if len(arg.TeamIDs) > 0 {
				teamIDMap := make(map[int64]int64, len(arg.TeamIDs))
				for _, id := range arg.TeamIDs {
					teamIDMap[id] = id
				}
				for _, v := range teams {
					if _, ok := teamIDMap[v.Tid]; !ok {
						data.UpDelTeams = append(data.UpDelTeams, v.ID)
					}
				}
			} else {
				for _, v := range teams {
					data.UpDelTeams = append(data.UpDelTeams, v.ID)
				}
			}
		}
		// year
		if err = s.dao.DB.Model(&model.YearMap{}).Where("aid=?", arg.Aid).Where("is_deleted=?", _notDeleted).Find(&years).Error; err != nil {
			log.Error("arcRelationChanges Arc s.dao.DB.Model MatchMap(%+v) error(%v)", arg.MatchIDs, err)
			return
		}
		if len(years) > 0 {
			if len(arg.Years) > 0 {
				yearMap := make(map[int64]int64, len(arg.Years))
				for _, id := range arg.Years {
					yearMap[id] = id
				}
				for _, v := range years {
					if _, ok := yearMap[v.Year]; !ok {
						data.UpDelYears = append(data.UpDelYears, v.ID)
					}
				}
			} else {
				for _, v := range years {
					data.UpDelYears = append(data.UpDelYears, v.ID)
				}
			}
		}
	}
	return
}

// BatchDelArc batch del archive.
func (s *Service) BatchDelArc(c context.Context, aids []int64) (err error) {
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&model.Arc{}).Where("aid IN (?)", aids).Update(map[string]int{"is_deleted": _deleted}).Error; err != nil {
		log.Error("BatchDelArc Arc s.dao.DB.Model Update(%+v) error(%v)", aids, err)
		err = tx.Rollback().Error
		return
	}
	if err = tx.Model(&model.GIDMap{}).Where("oid IN (?)", aids).Update(map[string]int{"is_deleted": _deleted}).Error; err != nil {
		log.Error("BatchDelArc GIDMap s.dao.DB.Model Update(%+v) error(%v)", aids, err)
		err = tx.Rollback().Error
		return
	}
	err = tx.Commit().Error
	return
}

// BatchPassArc batch pass archive.
func (s *Service) BatchPassArc(c context.Context, ids []int64) (err error) {
	var (
		arcs       []*model.Arc
		noPassAids []string
	)
	if err = s.dao.DB.Model(&model.Arc{}).Where("id IN (?)", unique(ids)).Find(&arcs).Error; err != nil {
		log.Error("BatchPassArc Arcs aids(%+v) error(%v)", ids, err)
		return
	}
	if len(arcs) == 0 {
		return
	}
	for _, arc := range arcs {
		if arc.IsDeleted == _checkNopass {
			noPassAids = append(noPassAids, strconv.FormatInt(arc.Aid, 10))
		}
	}
	if len(noPassAids) > 0 {
		strAids := strings.Join(noPassAids, ",")
		return fmt.Errorf("操作失败，该视频（avid：" + strAids + "）当前状态为“审核不通过”，如需添加，请使用手动添加的方式")
	}
	if err = s.dao.DB.Model(&model.Arc{}).Where("id IN (?)", ids).Update(map[string]int{"is_deleted": _checkPass}).Error; err != nil {
		log.Error("BatchPassArc Arc s.dao.DB.Model Update(%+v) error(%v)", ids, err)
	}
	return
}

// BatchNopassArc batch no pass archive.
func (s *Service) BatchNopassArc(c context.Context, ids []int64) (err error) {
	var (
		arcs     []*model.Arc
		passAids []string
		aids     []int64
	)
	if err = s.dao.DB.Model(&model.Arc{}).Where("id IN (?)", unique(ids)).Find(&arcs).Error; err != nil {
		log.Error("BatchPassArc Arcs aids(%+v) error(%v)", ids, err)
		return
	}
	if len(arcs) == 0 {
		return
	}
	for _, arc := range arcs {
		if arc.IsDeleted == _checkPass {
			passAids = append(passAids, strconv.FormatInt(arc.Aid, 10))
		}
		aids = append(aids, arc.Aid)
	}
	if len(passAids) > 0 {
		strAids := strings.Join(passAids, ",")
		return fmt.Errorf("操作失败，该视频（avid：" + strAids + "）当前状态为“审核通过”，如需更改，请手动修改单个视频状态")
	}
	tx := s.dao.DB.Begin()
	if err = tx.Error; err != nil {
		log.Error("s.dao.DB.Begin error(%v)", err)
		return
	}
	if err = tx.Model(&model.Arc{}).Where("id IN (?)", ids).Update(map[string]int{"is_deleted": _checkNopass}).Error; err != nil {
		log.Error("BatchNopassArc Arc s.dao.DB.Model Update(%+v) error(%v)", aids, err)
		err = tx.Rollback().Error
		return
	}
	if err = tx.Model(&model.GIDMap{}).Where("oid IN (?)", aids).Update(map[string]int{"is_deleted": _deleted}).Error; err != nil {
		log.Error("BatchNopassArc GIDMap s.dao.DB.Model Update(%+v) error(%v)", aids, err)
		err = tx.Rollback().Error
		return
	}
	err = tx.Commit().Error
	return
}

func upArcRelation(tx *gorm.DB, data *model.ArcRelation) (err error) {
	if len(data.AddGids) > 0 {
		sql, sqlParam := model.GidBatchAddSQL(data.AddGids)
		if err = tx.Model(&model.GIDMap{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("upArcRelation GIDMap tx.Model Exec(%+v) error(%v)", data.AddGids, err)
			return
		}
	}
	if len(data.UpAddGids) > 0 {
		if err = tx.Model(&model.GIDMap{}).Where("id IN (?)", data.UpAddGids).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("upArcRelation Tag tx.Model Updates(%+v) error(%v)", data.UpAddGids, err)
			return
		}
	}
	if len(data.UpDelGids) > 0 {
		if err = tx.Model(&model.GIDMap{}).Where("id IN (?)", data.UpDelGids).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("upArcRelation Tag tx.Model Updates(%+v) error(%v)", data.UpDelGids, err)
			return
		}
	}
	if len(data.AddMatchs) > 0 {
		sql, sqlParam := model.BatchAddMachMapSQL(data.AddMatchs)
		if err = tx.Model(&model.MatchMap{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("upArcRelation Match tx.Model Exec(%+v) error(%v)", data.AddMatchs, err)
			return
		}
	}
	if len(data.UpAddMatchs) > 0 {
		if err = tx.Model(&model.MatchMap{}).Where("id IN (?)", data.UpAddMatchs).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("upArcRelation Match tx.Model Updates(%+v) error(%v)", data.UpAddMatchs, err)
			return
		}
	}
	if len(data.UpDelMatchs) > 0 {
		if err = tx.Model(&model.MatchMap{}).Where("id IN (?)", data.UpDelMatchs).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("upArcRelation Match tx.Model Updates(%+v) error(%v)", data.UpDelMatchs, err)
			return
		}
	}
	if len(data.AddTags) > 0 {
		sql, sqlParam := model.BatchAddTagMapSQL(data.AddTags)
		if err = tx.Model(&model.TagMap{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("upArcRelation Tag tx.Model Exec(%+v) error(%v)", data.AddTags, err)
			return
		}
	}
	if len(data.UpAddTags) > 0 {
		if err = tx.Model(&model.TagMap{}).Where("id IN (?)", data.UpAddTags).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("upArcRelation Tag tx.Model Updates(%+v) error(%v)", data.UpAddTags, err)
			return
		}
	}
	if len(data.UpDelTags) > 0 {
		if err = tx.Model(&model.TagMap{}).Where("id IN (?)", data.UpDelTags).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("upArcRelation Tag tx.Model Updates(%+v) error(%v)", data.UpDelTags, err)
			return
		}
	}
	if len(data.AddTeams) > 0 {
		sql, sqlParam := model.BatchAddTeamMapSQL(data.AddTeams)
		if err = tx.Model(&model.TeamMap{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("upArcRelation Team tx.Model Exec(%+v) error(%v)", data.AddTeams, err)
			return
		}
	}
	if len(data.UpAddTeams) > 0 {
		if err = tx.Model(&model.TeamMap{}).Where("id IN (?)", data.UpAddTeams).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("upArcRelation Team tx.Model Updates(%+v) error(%v)", data.UpAddTags, err)
			return
		}
	}
	if len(data.UpDelTeams) > 0 {
		if err = tx.Model(&model.TeamMap{}).Where("id IN (?)", data.UpDelTeams).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("upArcRelation Team tx.Model Updates(%+v) error(%v)", data.UpDelTags, err)
			return
		}
	}
	if len(data.AddYears) > 0 {
		sql, sqlParam := model.BatchAddYearMapSQL(data.AddYears)
		if err = tx.Model(&model.YearMap{}).Exec(sql, sqlParam...).Error; err != nil {
			log.Error("upArcRelation Year tx.Model Exec(%+v) error(%v)", data.AddYears, err)
			return
		}
	}
	if len(data.UpAddYears) > 0 {
		if err = tx.Model(&model.YearMap{}).Where("id IN (?)", data.UpAddYears).Updates(map[string]interface{}{"is_deleted": _notDeleted}).Error; err != nil {
			log.Error("upArcRelation Year tx.Model Updates(%+v) error(%v)", data.UpAddTags, err)
			return
		}
	}
	if len(data.UpDelYears) > 0 {
		if err = tx.Model(&model.YearMap{}).Where("id IN (?)", data.UpDelYears).Updates(map[string]interface{}{"is_deleted": _deleted}).Error; err != nil {
			log.Error("upArcRelation Year tx.Model Updates(%+v) error(%v)", data.UpDelTags, err)
			return
		}
	}
	return
}
