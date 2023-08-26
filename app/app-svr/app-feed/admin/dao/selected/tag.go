package selected

import (
	"context"
)

type MidRes struct {
	Mid int64
}

var _subscriberTable = "tag_subscriber"

func (d *Dao) GetSubscriberCnt(_ context.Context, tagID int64, calDate string) (cnt int, err error) {
	err = d.TagDB.Table(_subscriberTable).
		Where("cal_date = ?", calDate).Where("tid = ?", tagID).Where("state = 0").
		Count(&cnt).Error
	return cnt, err
}

func (d *Dao) GetSubscriberDetail(_ context.Context, tagID int64, offset int, limit int, calDate string) (mids []int64, err error) {
	var midRes []MidRes
	err = d.TagDB.Table(_subscriberTable).
		Where("cal_date = ?", calDate).Where("tid = ?", tagID).Where("state = 0").
		Order("mid ASC").
		Offset(offset).Limit(limit).
		Find(&midRes).Error

	if err != nil {
		return mids, err
	}

	for _, item := range midRes {
		mids = append(mids, item.Mid)
	}
	return mids, nil
}
