package dao

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"go-common/library/cache/memcache"
	"go-common/library/log"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/admin/component"
	"go-gateway/app/web-svr/esports/admin/model"
	cacheModel "go-gateway/app/web-svr/esports/common/cache/model"
	model2 "go-gateway/app/web-svr/esports/interface/model"
	"strings"
)

var (
	_batchAddTeamSql           = "insert into es_contest_teams (`contest_id`, `team_id`) values %s"
	_batchAddTeamWhenUpdateSql = "insert into es_contest_teams" +
		" (`contest_id`, `team_id`, `survival_rank`, `kill_number`, `score`, `rank_edit_status`) " +
		"values %s"

	_contestFilter   = "contest_id = ?"
	_isDeletedFilter = "is_deleted = ?"
)

func (d *Dao) BatchAddTeams(ctx context.Context, tx *gorm.DB, contestId int64, teamIds []int64) (err error) {
	if len(teamIds) == 0 {
		log.Warnc(ctx, "[DB][BatchAddTeams][Error], err: teamIds Empty, contestId:%d", contestId)
		return
	}
	var rowStrings []string
	param := make([]interface{}, 0)
	for _, v := range teamIds {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, contestId, v)
	}
	sql := fmt.Sprintf(_batchAddTeamSql, strings.Join(rowStrings, ","))
	db := d.choseDB(tx)
	if err = db.Model(&model.ContestTeam{}).Exec(sql, param...).Error; err != nil {
		log.Errorc(ctx, "[DB][BatchAddTeamsWhenUpdate][Error] err:(%+v)", err)
	}
	return
}

func (d *Dao) BatchAddTeamsWhenUpdate(ctx context.Context, tx *gorm.DB, contestTeams []*model.ContestTeam) (err error) {
	if len(contestTeams) == 0 {
		return
	}
	var rowStrings []string
	param := make([]interface{}, 0)
	for _, v := range contestTeams {
		rowStrings = append(rowStrings, "(?,?,?,?,?,?)")
		param = append(param, v.ContestId, v.TeamId, v.SurvivalRank, v.KillNumber, v.Score, v.RankEditStatus)
	}
	sql := fmt.Sprintf(_batchAddTeamWhenUpdateSql, strings.Join(rowStrings, ","))
	db := d.choseDB(tx)
	if err = db.Model(&model.ContestTeam{}).Exec(sql, param...).Error; err != nil {
		log.Errorc(ctx, "[Dao][DB][BatchAddTeamsWhenUpdate][Error] err:(%+v)", err)
	}
	return
}

func (d *Dao) BatchDeleteTeamsByContestId(ctx context.Context, tx *gorm.DB, contestId int64) (err error) {
	db := d.choseDB(tx)
	err = db.Model(&model.ContestTeam{}).
		Where(_contestFilter, contestId).
		Where(_isDeletedFilter, model.ContestTeamNotDeleted).
		Updates(map[string]interface{}{"is_deleted": model.ContestTeamDeleted}).Error
	if err != nil {
		log.Errorc(ctx, "[DB][BatchDeleteTeams][Error] err:(%+v)", err)
		return
	}
	return
}

func (d *Dao) BatchDeleteTeamsByIds(ctx context.Context, tx *gorm.DB, ids []int64) (err error) {
	idsStr := xstr.JoinInts(ids)
	db := d.choseDB(tx)
	err = db.Model(&model.ContestTeam{}).
		Where("ids in (?)", idsStr).
		Updates(map[string]interface{}{"is_deleted": model.ContestTeamDeleted}).Error
	if err != nil {
		log.Errorc(ctx, "[DB][BatchDeleteTeams][Error] err:(%+v)", err)
		return
	}
	return
}

func (d *Dao) GetTeamList(ctx context.Context, contestId int64) (contestTeams []*model.ContestTeam, err error) {
	err = d.DB.Model(&model.ContestTeam{}).
		Where(_contestFilter, contestId).
		Where(_isDeletedFilter, model.ContestTeamNotDeleted).
		Order("id ASC").
		Find(&contestTeams).Error
	if err != nil {
		log.Errorc(ctx, "[DB][GetTeamList][Error], err:(%+v)", err)
		return
	}
	return
}

func (d *Dao) GetTeamsOrderBySurvivalRank(ctx context.Context, contestId int64) (contestTeams []*model.ContestTeam, err error) {
	err = d.DB.Model(&model.ContestTeam{}).
		Where(_contestFilter, contestId).
		Where(_isDeletedFilter, model.ContestTeamNotDeleted).
		Order("rank_edit_status DESC").
		Order("survival_rank ASC").
		Find(&contestTeams).Error
	if err != nil {
		log.Errorc(ctx, "[DB][GetTeamList][Error], err:(%+v)", err)
		return
	}
	return
}

func (d *Dao) choseDB(tx *gorm.DB) *gorm.DB {
	db := d.DB
	if tx != nil {
		return tx
	}
	return db
}

func (d *Dao) RebuildTeamsMcCacheByContestTeamInfos(ctx context.Context, contestId int64, contestTeams []*model2.ContestTeamInfo) (err error) {
	teamsScoreInfo := make([]*model2.ContestTeamScoreInfo, 0)
	for _, v := range contestTeams {
		teamsScoreInfo = append(teamsScoreInfo, &model2.ContestTeamScoreInfo{
			TeamId:         v.TeamId,
			Score:          v.ScoreInfo.Score,
			KillNumber:     v.ScoreInfo.KillNumber,
			SurvivalRank:   v.ScoreInfo.SurvivalRank,
			SeasonTeamRank: v.ScoreInfo.SeasonTeamRank,
			Rank:           v.ScoreInfo.Rank,
		})
	}
	cacheKey := cacheModel.GetContestTeamsCacheKey(contestId)
	value := cacheModel.ContestTeamsMcCache{
		ContestId:   contestId,
		Teams:       teamsScoreInfo,
		BuildMethod: "admin",
	}
	item := &memcache.Item{
		Key:        cacheKey,
		Object:     value,
		Expiration: cacheModel.GetContestTeamsCacheKeyTtlSecond(),
		Flags:      memcache.FlagJSON}
	err = component.GlobalMemcached.Set(ctx, item)
	if err != nil {
		log.Errorc(ctx, "[Dao][ContestTeams][[SetCache][ErrorGroup][Error], key:(%s), err(%+v)", cacheKey, err)
	}
	return
}
