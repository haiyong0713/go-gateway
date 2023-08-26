package dao

import (
	"context"
	"go-gateway/app/web-svr/esports/service/internal/model"

	"go-common/library/log"
)

const (
	_teamTableName = "es_teams"
)

func (d *dao) GetTeamsByIds(ctx context.Context, teamIds []int64) (teamModels []*model.TeamModel, err error) {
	teamModels = make([]*model.TeamModel, 0)
	if err = d.orm.Table(_teamTableName).Where("id in (?)", teamIds).Find(&teamModels).Error; err != nil {
		log.Errorc(ctx, "[Dao][GetTeamsByIds][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) getTeamsMapByID(ctx context.Context, teamIDs []int64) (teamsMap map[int64]*model.TeamModel, err error) {
	teams := make([]*model.TeamModel, 0)
	if err = d.orm.Table(_teamTableName).Model(&model.TeamModel{}).Where("id IN (?)", teamIDs).Find(&teams).Error; err != nil {
		log.Errorc(ctx, "getTeamsMapByID teamIDs(%+v) Error (%v)", teamIDs, err)
		return
	}
	teamsMap = make(map[int64]*model.TeamModel, len(teams))
	for _, v := range teams {
		teamsMap[v.ID] = v
	}
	return
}
