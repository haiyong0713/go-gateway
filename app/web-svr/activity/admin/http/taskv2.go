package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model"
)

func taskv2List(ctx *bm.Context) {
	var (
		err   error
		count int
		list  []*model.ActTask
	)
	v := new(struct {
		ActivityID int64 `form:"activity_id"`
		Page       int   `form:"pn" default:"1"`
		Size       int   `form:"ps" default:"20"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	db := actSrv.DB
	db = db.Table("act_task")
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	if v.ActivityID > 0 {
		db = db.Where("activity_id = ?", v.ActivityID)
	}
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).Order("order_id asc").
		Find(&list).Error; err != nil {
		log.Errorc(ctx, "taskv2List(%d,%d) error(%v)", v.Page, v.Size, err)
		ctx.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActTask{}).Count(&count).Error; err != nil {
		log.Errorc(ctx, "taskv2List count error(%v)", err)
		ctx.JSON(nil, err)
		return
	}

	data := map[string]interface{}{
		"data":  list,
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	ctx.JSONMap(data, nil)
}

// saveTaskV2 存储
func saveTaskV2(c *bm.Context) {
	var (
		request = &model.ActTask{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(nil, taskv2Srv.TaskInsertOrUpdate(c, request))
}
