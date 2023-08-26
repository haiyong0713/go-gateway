package http

import (
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model"
)

func sp2021User(ctx *bm.Context) {

	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(actSrv.SpringFestival2021Mid(ctx, v.Mid))
}

func sp2021InviteLog(ctx *bm.Context) {
	v := new(struct {
		Mid  int64 `form:"mid" validate:"min=1"`
		Page int   `form:"pn" default:"1"`
		Size int   `form:"ps" default:"20"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var (
		err   error
		count int
		list  []*model.ActSpringRelation
	)

	db := actSrv.DB
	db = db.Table("act_spring_relation")
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("mid = ?", v.Mid)
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).Order("ctime desc").
		Find(&list).Error; err != nil {
		log.Errorc(ctx, "sp2021InviteLog(%d,%d) error(%v)", v.Page, v.Size, err)
		ctx.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActTask{}).Count(&count).Error; err != nil {
		log.Errorc(ctx, "sp2021InviteLog count error(%v)", err)
		ctx.JSON(nil, err)
		return
	}
	reply := &model.ActSpringRelationReply{}
	reply.Page = map[string]interface{}{
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	reply.List = list
	data := map[string]interface{}{
		"data": reply,
	}
	ctx.JSONMap(data, nil)
}

func sp2021CardsNumsLog(ctx *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var (
		err   error
		count int
		list  []*model.ActSpringCardsNums
	)

	db := actSrv.DB
	db = db.Table("act_spring_cards_nums")

	db = db.Where("mid = ?", v.Mid)
	if err = db.
		Find(&list).Error; err != nil {
		log.Errorc(ctx, "sp2021CardsNumsLog(%d) error(%v)", v.Mid, err)
		ctx.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActTask{}).Count(&count).Error; err != nil {
		log.Errorc(ctx, "sp2021CardsNumsLog count error(%v)", err)
		ctx.JSON(nil, err)
		return
	}
	reply := &model.ActSpringCardsNumsReply{}

	reply.List = list
	data := map[string]interface{}{
		"data": reply,
	}
	ctx.JSONMap(data, nil)
}

func sp2021ComposeLog(ctx *bm.Context) {
	v := new(struct {
		Mid  int64 `form:"mid" validate:"min=1"`
		Page int   `form:"pn" default:"1"`
		Size int   `form:"ps" default:"20"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var (
		err   error
		count int
		list  []*model.ActSpringComposeCardLog
	)

	db := actSrv.DB
	db = db.Table("act_spring_compose_card_log")
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("mid = ?", v.Mid)
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).Order("ctime desc").
		Find(&list).Error; err != nil {
		log.Errorc(ctx, "sp2021ComposeLog(%d,%d) error(%v)", v.Page, v.Size, err)
		ctx.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActTask{}).Count(&count).Error; err != nil {
		log.Errorc(ctx, "sp2021ComposeLog count error(%v)", err)
		ctx.JSON(nil, err)
		return
	}
	reply := &model.ActSpringComposeCardLogReply{}
	reply.Page = map[string]interface{}{
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	reply.List = list
	data := map[string]interface{}{
		"data": reply,
	}
	ctx.JSONMap(data, nil)
}

func sp2021SendCardLog(ctx *bm.Context) {
	v := new(struct {
		Mid  int64 `form:"mid" validate:"min=1"`
		Page int   `form:"pn" default:"1"`
		Size int   `form:"ps" default:"20"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var (
		err   error
		count int
		list  []*model.ActSpringSendCardLog
	)

	db := actSrv.DB
	db = db.Table("act_spring_send_card_log")
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("mid = ?", v.Mid)
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).Order("ctime desc").
		Find(&list).Error; err != nil {
		log.Errorc(ctx, "sp2021ComposeLog(%d,%d) error(%v)", v.Page, v.Size, err)
		ctx.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActTask{}).Count(&count).Error; err != nil {
		log.Errorc(ctx, "sp2021ComposeLog count error(%v)", err)
		ctx.JSON(nil, err)
		return
	}
	reply := &model.ActSpringSendCardLogReply{}
	reply.Page = map[string]interface{}{
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	reply.List = list
	data := map[string]interface{}{
		"data": reply,
	}
	ctx.JSONMap(data, nil)
}
