package fit

import (
	"context"
	"go-common/library/log"
	fit "go-gateway/app/web-svr/activity/admin/model/fit"
)

const (
	_addOnePlanSQL = "INSERT IGNORE INTO `act_fit_plan_config` (plan_title,plan_tags,bodan_id,plan_view,plan_danmaku,plan_fav,pic_cover,creator) VALUES(?,?,?,?,?,?,?,?)"
)

// AddOnePlan add a plan.
func (d *Dao) AddOnePlan(ctx context.Context, record *fit.PlanRecord) (int64, error) {
	res, err := d.db.Exec(ctx, _addOnePlanSQL,
		record.PlanTitle, record.PlanTags,
		record.BodanId, record.PlanView,
		record.PlanDanmaku, record.PlanFav,
		record.PicCover, record.Creator)
	if err != nil {
		log.Error("AddOnePlan:d.db.Exec error(%+v)", err)
		return 0, err
	}
	return res.LastInsertId()
}
