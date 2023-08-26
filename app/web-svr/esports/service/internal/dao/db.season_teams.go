package dao

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

const (
	_esSeasonTeam = "es_team_in_seasons"
)

func (d *dao) GetSeasonTeamsModel(ctx context.Context, seasonId int64) (seasonTeams []*model.SeasonTeamModel, err error) {
	seasonTeams = make([]*model.SeasonTeamModel, 0)
	if err = d.orm.Table(_esSeasonTeam).Where("sid = ?", seasonId).Find(&seasonTeams).Error; err != nil {
		log.Errorc(ctx, "[Dao][DB][getSeasonTeams][Error], err:%+v", err)
		return
	}
	return
}
