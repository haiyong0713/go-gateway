package fit

import (
	"context"
	"database/sql"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/fit"
)

const (
	_addOnePlanSQL  = "INSERT IGNORE INTO `act_fit_plan_config` (plan_title,plan_tags,bodan_id,plan_view,plan_danmaku,plan_fav,creator,pic_cover) VALUES(?,?,?,?,?,?,?,?)"
	_selPlanListSQL = "SELECT id,plan_title,plan_tags,bodan_id,plan_view,plan_danmaku,plan_fav,pic_cover,status,creator,ctime,mtime FROM `act_fit_plan_config` WHERE status = 1 limit ?,?"
	_selPlanById    = "SELECT id,plan_title,plan_tags,bodan_id,plan_view,plan_danmaku,plan_fav,pic_cover,status,creator,ctime,mtime FROM `act_fit_plan_config` WHERE id = ?"
)

// AddOnePlan add a plan.
func (d *dao) AddOnePlan(ctx context.Context, record fit.PlanRecord) (int64, error) {
	res, err := d.db.Exec(ctx, _addOnePlanSQL,
		record.PlanTitle, record.PlanTags,
		record.BodanId, record.PlanView,
		record.PlanDanmaku, record.PlanFav,
		record.Creator)

	if err != nil {
		log.Errorc(ctx, "AddOnePlan:d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.LastInsertId()
}

// GetPlanList get all plans.
func (d *dao) GetPlanList(ctx context.Context, offset, limit int) (res []*fit.PlanRecordRes, err error) {
	res = []*fit.PlanRecordRes{}

	rows, err := d.db.Query(ctx, _selPlanListSQL, offset, limit)
	if err != nil {
		log.Errorc(ctx, "GetPlanList:d.db.Query error.error detail is(%+v)", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		r := &fit.DBActFitPlanConfig{}
		err = rows.Scan(&r.ID, &r.PlanTitle, &r.PlanTags, &r.BodanId, &r.PlanView,
			&r.PlanDanmaku, &r.PlanFav, &r.PicCover, &r.Status, &r.Creator, &r.Ctime, &r.Mtime)
		if err != nil {
			log.Errorc(ctx, "GetPlanList:rows.Scan error.error detail is(%+v)", err)
			return
		}
		data := &fit.PlanRecordRes{
			ID:          r.ID,
			PlanTitle:   r.PlanTitle,
			PlanTags:    r.PlanTags,
			PlanView:    r.PlanView,
			PlanDanmaku: r.PlanDanmaku,
			PlanFav:     r.PlanFav,
			PicCover:    r.PicCover,
		}
		res = append(res, data)
	}
	if err = rows.Err(); err != nil {
		log.Errorc(ctx, "GetPlanList:rows.Err.error detail is(%+v)", err)
	}
	return
}

// GetPlanById
func (d *dao) GetPlanById(ctx context.Context, planId int64) (res *fit.DBActFitPlanConfig, err error) {
	res = new(fit.DBActFitPlanConfig)
	row := d.db.QueryRow(ctx, _selPlanById, planId)
	err = row.Scan(&res.ID, &res.PlanTitle, &res.PlanTags, &res.BodanId, &res.PlanView,
		&res.PlanDanmaku, &res.PlanFav, &res.PicCover, &res.Status, &res.Creator, &res.Ctime, &res.Mtime)
	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
			return
		}
		log.Errorc(ctx, "GetPlanById:row.Scan err, error(%v)", err)
	}
	return
}
