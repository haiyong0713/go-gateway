package fit

import (
	"context"
	xsql "database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/fit"
)

const (
	fitPlanTableName = "act_fit_plan_config"
)

const (
	_selPlanListSQL = "SELECT id,plan_title,plan_tags,bodan_id,plan_view,plan_danmaku,plan_fav,pic_cover,status,creator,ctime,mtime FROM `act_fit_plan_config` WHERE status = 1 limit ?,?"
	_updatePlanById = "UPDATE act_fit_plan_config SET plan_view = ? , plan_danmaku = ? WHERE id = ?"
)

// GetPlanList get all plans.
func (d *dao) GetPlanList(ctx context.Context, offset, limit int) (res []*fit.PlanRecordRes, err error) {
	res = []*fit.PlanRecordRes{}

	rows, err := d.db.Query(ctx, _selPlanListSQL, offset, limit)
	if err != nil {
		log.Errorc(ctx, "Job GetPlanList:d.db.Query error.error detail is(%+v)", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		r := &fit.DBActFitPlanConfig{}
		err = rows.Scan(&r.ID, &r.PlanTitle, &r.PlanTags, &r.BodanId, &r.PlanView,
			&r.PlanDanmaku, &r.PlanFav, &r.PicCover, &r.Status, &r.Creator, &r.Ctime, &r.Mtime)
		if err != nil {
			log.Errorc(ctx, "Job GetPlanList:rows.Scan error.error detail is(%+v)", err)
			return
		}
		data := &fit.PlanRecordRes{
			ID:      r.ID,
			BodanId: r.BodanId,
		}
		res = append(res, data)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(ctx, "Job GetPlanList:rows.Err.error detail is(%+v)", err)
	}
	return
}

// UpdateOnePlanById 更新plan
func (d *dao) UpdateOnePlanById(c context.Context, planId int64, views int32, danmaku int32) (affected int64, err error) {
	var (
		res xsql.Result
	)
	if res, err = d.db.Exec(c, _updatePlanById, views, danmaku, planId); err != nil {
		log.Errorc(c, "UpdateOnePlanById:db.Exec error is :(%v).", err)
		return
	}
	return res.RowsAffected()
}
