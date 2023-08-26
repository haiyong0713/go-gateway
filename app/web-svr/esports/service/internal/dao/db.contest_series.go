package dao

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/service/internal/model"
)

const (
	tableOfContestSeries = "contest_series"
	sqlWhereDeleted      = "is_deleted = ?"
)

func (d *dao) GetScoreRuleConfigBySeriesId(ctx context.Context, seriesId int64) (contestSeries *model.ContestSeriesByScoreRule, err error) {
	contestSeries = new(model.ContestSeriesByScoreRule)
	err = d.orm.Table(tableOfContestSeries).Model(model.ContestSeriesByScoreRule{}).Where("id = ? AND is_deleted = ?", seriesId, model.IsDeletedFalse).Find(&contestSeries).Error
	if err != nil {
		log.Errorc(ctx, "[DB][ScoreRuleConfigGet][Error], err:(%+v)", err)
		return
	}
	return
}

func (d *dao) GetSeriesById(ctx context.Context, seriesId int64) (contestSeriesModel *model.ContestSeriesModel, err error) {
	contestSeriesModel = new(model.ContestSeriesModel)
	if err = d.orm.Table(tableOfContestSeries).Where("id = ?", seriesId).
		Where(sqlWhereDeleted, model.IsDeletedFalse).
		Find(&contestSeriesModel).Error; err != nil {
		log.Errorc(ctx, "[DB][GetSeriesById][Error], err:(%+v)", err)
	}
	return
}

func (d *dao) GetSeriesBySeasonId(ctx context.Context, seasonId int64) (contestSeriesModels []*model.ContestSeriesModel, err error) {
	contestSeriesModels = make([]*model.ContestSeriesModel, 0)
	if err = d.orm.Table(tableOfContestSeries).Where("season_id = ?", seasonId).
		Where(sqlWhereDeleted, model.IsDeletedFalse).
		Find(&contestSeriesModels).Error; err != nil {
		log.Errorc(ctx, "[DB][GetSeriesBySeasonId][Error], err:(%+v)", err)
		return
	}
	return
}

func (d *dao) GetSeriesByIds(ctx context.Context, seriesIds []int64) (contestSeriesModelMap map[int64]*model.ContestSeriesModel, err error) {
	contestSeriesModelMap = make(map[int64]*model.ContestSeriesModel)
	contestSeriesModels := make([]*model.ContestSeriesModel, 0)
	if err = d.orm.Table(tableOfContestSeries).Where("id in (?)", seriesIds).
		Where(sqlWhereDeleted, model.IsDeletedFalse).
		Find(&contestSeriesModels).Error; err != nil {
		log.Errorc(ctx, "[DB][GetSeriesBySeasonId][Error], err:(%+v)", err)
		return
	}
	for _, v := range contestSeriesModels {
		contestSeriesModelMap[v.ID] = v
	}
	return
}
