package dao

import (
	"context"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/admin/model"
)

func (d *Dao) GetContest(ctx context.Context, contestId int64) (contest *model.Contest, err error) {
	contest = new(model.Contest)
	if err = d.DB.Where("id=?", contestId).First(&contest).Error; err != nil {
		log.Errorc(ctx, "[Dao][GetContest][Error], err:(%+v)", err)
		return
	}
	return
}
