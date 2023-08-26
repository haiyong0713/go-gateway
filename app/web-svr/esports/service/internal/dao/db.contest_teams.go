package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"

	"github.com/jinzhu/gorm"
)

const (
	_contestFilter    = "contest_id = ?"
	_isDeletedFilter  = "is_deleted = ?"
	contestTeamsTable = "es_contest_teams"
)

func (d *dao) BatchAddTeams(ctx context.Context, tx *gorm.DB, contestId int64, teamIds []int64) (err error) {
	if len(teamIds) == 0 {
		return
	}

	var rowStrings []string
	param := make([]interface{}, 0)
	for _, v := range teamIds {
		rowStrings = append(rowStrings, "(?,?)")
		param = append(param, contestId, v)
	}
	sql := fmt.Sprintf(_batchAddTeamSql, strings.Join(rowStrings, ","))
	if err = tx.Table(contestTeamsTable).Model(&model.ContestTeam{}).Exec(sql, param...).Error; err != nil {
		log.Errorc(ctx, "[DB][BatchAddTeamsWhenUpdate][Error] err:(%+v)", err)
	}
	return
}

func (d *dao) contestTeamsUpdate(ctx context.Context, contest *model.ContestModel, teamIds []int64, tx *gorm.DB) (err error) {
	contestId := contest.ID
	contestTeams, err := d.GetTeamsOrderBySurvivalRank(ctx, contestId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "获取战队配置失败，请重试")
		return
	}
	// 如果有队伍变更, 则将已配置的生存排名顺序重建，重建规则：被保留的战队排名依次前移即可
	newTeamIdsMap := make(map[int64]bool)
	for _, teamId := range teamIds {
		newTeamIdsMap[teamId] = true
	}
	rebuildRank := false
	oldTeamIds := make([]int64, 0)
	waitForReBuildRankTeamsMap := make(map[int64]*model.ContestTeam)
	index := 0
	for _, contestTeam := range contestTeams {
		oldTeamIds = append(oldTeamIds, contestTeam.TeamId)
		if contestTeam.RankEditStatus != model.ContestTeamRankEditStatusOn {
			continue
		}
		rebuildRank = true
		contestTeam.SurvivalRank = int64(index + 1)
		index++
		waitForReBuildRankTeamsMap[contestTeam.TeamId] = contestTeam
	}
	// 队伍未更新的话不进行后续操作
	if int64SliceCompare(teamIds, oldTeamIds) {
		return
	}
	err = d.BatchDeleteTeamsByContestId(ctx, tx, contestId)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeamsUpdate][Do][Error], error:(%+v)", err)
		return
	}
	if rebuildRank {
		err = d.batchAddTeamsHandlerWhenUpdate(ctx, contest, teamIds, waitForReBuildRankTeamsMap, tx)
	} else {
		err = d.BatchAddTeams(ctx, tx, contestId, teamIds)
	}
	return
}

func (d *dao) batchAddTeamsHandlerWhenUpdate(ctx context.Context, contest *model.ContestModel, teamIds []int64, waitForReBuildRankTeamsMap map[int64]*model.ContestTeam, tx *gorm.DB) (err error) {
	contestId := contest.ID
	insertTeamInfos := make([]*model.ContestTeam, 0)
	for _, teamId := range teamIds {
		if v, isOk := waitForReBuildRankTeamsMap[teamId]; isOk {
			insertTeamInfos = append(insertTeamInfos, v)
		} else {
			insertTeamInfos = append(insertTeamInfos, d.contestTeamDefaultFormat(contestId, teamId))
		}
	}
	err = d.teamsScoreCalculate(ctx, contest.SeriesId, insertTeamInfos)
	if err != nil {
		log.Errorc(ctx, "[Service][ContestTeamsUpdate][teamsScoreCalculate][Error], error:(%+v)", err)
		return
	}
	err = d.BatchAddTeamsWhenUpdate(ctx, tx, insertTeamInfos)
	return
}

func (d *dao) teamsScoreCalculate(ctx context.Context, contestSeriesId int64, contestTeams []*model.ContestTeam) (err error) {
	scoreRules, err := d.GetScoreRules(ctx, contestSeriesId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "获取赛程阶段信息失败, 请重试")
		return
	}
	for _, contestTeam := range contestTeams {
		survivalScore := int64(0)
		if contestTeam.SurvivalRank != 0 && contestTeam.SurvivalRank <= int64(len(scoreRules.RankScores)) {
			survivalScore = scoreRules.RankScores[contestTeam.SurvivalRank-1]
		}
		contestTeam.Score = survivalScore + contestTeam.KillNumber*scoreRules.KillScore
	}
	return
}

func (d *dao) GetScoreRules(ctx context.Context, seriesId int64) (scoreRules *model.PUBGContestSeriesScoreRule, err error) {
	contestSeries, err := d.GetScoreRuleConfigBySeriesId(ctx, seriesId)
	if err != nil {
		err = xecode.Errorf(xecode.RequestErr, "获取赛程阶段信息失败, 请重试")
		return
	}
	scoreRules = &model.PUBGContestSeriesScoreRule{
		KillScore:  0,
		RankScores: make([]int64, 0),
	}
	if contestSeries.ScoreRuleConfig == "" || contestSeries.ScoreRuleConfig == "null" {
		log.Errorc(ctx, "[Service][ContestSeries][GetScoreRules][ExtraConfig][Empty],contestSeriesId:%d", seriesId)
		return
	}
	if err = json.Unmarshal([]byte(contestSeries.ScoreRuleConfig), &scoreRules); err != nil {
		log.Errorc(ctx, "[Service][ContestSeries][GetScoreRules][ExtraConfig][Unmarshal][Error],err:%+v", err)
		err = xecode.Errorf(xecode.RequestErr, "获取赛程阶段积分规则失败，请重试")
	}
	return
}

func (d *dao) contestTeamDefaultFormat(contestId int64, teamId int64) (contestTeam *model.ContestTeam) {
	return &model.ContestTeam{
		ContestId:      contestId,
		TeamId:         teamId,
		SurvivalRank:   0,
		KillNumber:     0,
		Score:          0,
		RankEditStatus: model.ContestTeamRankEditStatusOff,
	}
}

func (d *dao) BatchAddTeamsWhenUpdate(ctx context.Context, tx *gorm.DB, contestTeams []*model.ContestTeam) (err error) {
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
	if err = tx.Table(contestTeamsTable).Model(&model.ContestTeam{}).Exec(sql, param...).Error; err != nil {
		log.Errorc(ctx, "[Dao][DB][BatchAddTeamsWhenUpdate][Error] err:(%+v)", err)
	}
	return
}

func (d *dao) GetTeamsOrderBySurvivalRank(ctx context.Context, contestId int64) (contestTeams []*model.ContestTeam, err error) {
	err = d.orm.Table(contestTeamsTable).Model(&model.ContestTeam{}).
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

func int64SliceCompare(slice1 []int64, slice2 []int64) bool {
	int64SliceSort(slice1)
	int64SliceSort(slice2)
	if len(slice1) != len(slice2) {
		return false
	}
	for index, v := range slice1 {
		if v != slice2[index] {
			return false
		}
	}
	return true
}

func int64SliceSort(slice []int64) {
	sort.Slice(slice, func(i, j int) bool {
		return slice[i] < slice[j]
	})
}

func (d *dao) BatchDeleteTeamsByContestId(ctx context.Context, tx *gorm.DB, contestId int64) (err error) {
	err = tx.Table(contestTeamsTable).Model(&model.ContestTeam{}).
		Where(_contestFilter, contestId).
		Where(_isDeletedFilter, model.ContestTeamNotDeleted).
		Updates(map[string]interface{}{"is_deleted": model.ContestTeamDeleted}).Error
	if err != nil {
		log.Errorc(ctx, "[DB][BatchDeleteTeams][Error] err:(%+v)", err)
		return
	}
	return
}

func (d *dao) BatchDeleteTeamsByIds(ctx context.Context, tx *gorm.DB, ids []int64) (err error) {
	err = tx.Table(contestTeamsTable).Model(&model.ContestTeam{}).
		Where("ids in (?)", ids).
		Updates(map[string]interface{}{"is_deleted": model.ContestTeamDeleted}).Error
	if err != nil {
		log.Errorc(ctx, "[DB][BatchDeleteTeams][Error] err:(%+v)", err)
		return
	}
	return
}
