package dao

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

const (
	gameTableName = "es_games"
)

func (d *dao) GetGamesByIds(ctx context.Context, gameIds []int64) (gameModels []*model.GameModel, err error) {
	gameModels = make([]*model.GameModel, 0)
	if err = d.orm.Table(gameTableName).Where("id in (?)", gameIds).Find(&gameModels).Error; err != nil {
		log.Errorc(ctx, "[Dao][GetGamesByIds][Error], err:%+v", err)
		return
	}
	return
}

func (d *dao) GetAllGames(ctx context.Context) (gameModels []*model.GameModel, err error) {
	gameModels = make([]*model.GameModel, 0)
	if err = d.orm.Table(gameTableName).Find(&gameModels).Error; err != nil {
		log.Errorc(ctx, "[Dao][GetAllGames][Error], err:%+v", err)
		return
	}
	return
}
