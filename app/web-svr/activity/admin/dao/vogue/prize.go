package vogue

import (
	"context"

	"go-common/library/log"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

const (
	_tablePrizes = "act_vogue_user_task"
)

func (d *Dao) PrizesList(c context.Context, params *voguemdl.PrizeSearch) (list []*voguemdl.PrizeData, count int64, err error) {
	db := d.DB.Table(_tablePrizes)
	db = db.Select("act_vogue_user_task.*, act_vogue_goods.name AS goods_name, act_vogue_goods.attr AS goods_attr, act_vogue_goods.score AS goods_score_setting, act_vogue_user_cost.cost AS goods_score").
		Joins("LEFT JOIN act_vogue_goods ON act_vogue_goods.id = act_vogue_user_task.goods").
		Joins("LEFT JOIN act_vogue_user_cost ON act_vogue_user_task.uid = act_vogue_user_cost.mid AND act_vogue_user_task.goods = act_vogue_user_cost.goods")

	db = db.Where("act_vogue_user_task.goods_state != ?", voguemdl.UserTaskStatusInProgress)
	if params.Uid > 0 {
		db = db.Where("uid = ?", params.Uid)
	}
	if err = db.Count(&count).Error; err != nil {
		log.Error("[VoguePrizesList] count (%v) error (%v)", params, err)
		return
	}
	if count == 0 {
		list = make([]*voguemdl.PrizeData, 0)
		return
	}
	db = db.Order("id desc")
	if params.Ps > 0 {
		db = db.Limit(params.Ps).Offset((params.Pn - 1) * params.Ps)
	}
	if err = db.Find(&list).Error; err != nil {
		log.Error("[VoguePrizesList] d.DB.Find, error(%v)", err)
	}
	return
}
