package dao

import (
	"context"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/service/internal/model"

	"go-common/library/log"
)

const (
	_seasonTableName = "es_seasons"
)

func (d *dao) GetSeasonByID(ctx context.Context, id int64) (season *model.SeasonModel, err error) {
	season = new(model.SeasonModel)
	if err = d.orm.Table(_seasonTableName).Where("id=?", id).First(&season).Error; err != nil {
		log.Errorc(ctx, "[getSeasonByID][First][Error], seasonID:%d err:%+v", id, err)
		return
	}
	return
}

func (d *dao) GetSeasonsByIDs(ctx context.Context, ids []int64) (seasons []*model.SeasonModel, err error) {
	seasons = make([]*model.SeasonModel, 0)
	if err = d.orm.Table(_seasonTableName).Where("id in (?)", ids).Find(&seasons).Error; err != nil {
		log.Errorc(ctx, "[GetSeasonsByIDs][Find][Error], seasonIDs:%s err:%+v", xstr.JoinInts(ids), err)
		return
	}
	return
}

func (d *dao) GetSeasonsBySETime(ctx context.Context, startTime int64, endTime int64) (seasons []*model.SeasonModel, err error) {
	seasons = make([]*model.SeasonModel, 0)
	err = d.orm.Table(_seasonTableName).Where("stime <= ?", startTime).Where("etime >= ?", endTime).Find(&seasons).Error
	if err != nil {
		log.Errorc(ctx, "[GetSeasonsBySETime][Find][Error], err:%+v", err)
		return
	}
	return
}
