package dao

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

const (
	matchTableName = "es_matchs"
)

func (d *dao) GetMatchModel(ctx context.Context, matchId int64) (matchModel *model.MatchModel, err error) {
	matchModel = new(model.MatchModel)
	if err = d.orm.Table(matchTableName).Where("status = ?", model.FreezeFalse).Where("id = ?", matchId).Find(&matchModel).Error; err != nil {
		log.Errorc(ctx, "[Dao][GetMatchModel][Orm][Error], err:%+v", err)
	}
	return
}
