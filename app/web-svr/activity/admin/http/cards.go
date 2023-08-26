package http

import (
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model"
	"go-gateway/app/web-svr/activity/admin/model/cards"
)

const (
	midEqual  = "mid = ?"
	ctimeDesc = "ctime desc"
)

func addCards(ctx *bm.Context) {
	v := new(struct {
		Name      string `form:"name" validate:"required"`
		LotteryID int64  `form:"lottery_id" validate:"required"`
		SID       string `form:"sid" validate:"required"`
		ReserveID int64  `form:"reserve_id" validate:"required"`
		CardsNum  int64  `form:"cards_num" validate:"required"`
		Cards     string `form:"cards" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	card := &cards.Cards{
		Name:      v.Name,
		LotteryID: v.LotteryID,
		SID:       v.SID,
		ReserveID: v.ReserveID,
		CardsNum:  v.CardsNum,
		Cards:     v.Cards,
	}

	ctx.JSON(nil, actSrv.AddCards(ctx, card))
	return
}

func editCards(ctx *bm.Context) {
	v := new(struct {
		ID        int64  `form:"id" validate:"required"`
		Name      string `form:"name" validate:"required"`
		LotteryID int64  `form:"lottery_id" validate:"required"`
		ReserveID int64  `form:"reserve_id" validate:"required"`
		CardsNum  int64  `form:"cards_num" validate:"required"`
		Cards     string `form:"cards" validate:"required"`
		SID       string `form:"sid" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	db := actSrv.DB
	db = db.Table("act_cards")
	var err error
	if err = db.Where("id=?", v.ID).Update(map[string]interface{}{
		"name":       v.Name,
		"lottery_id": v.LotteryID,
		"reserve_id": v.ReserveID,
		"cards_num":  v.CardsNum,
		"cards":      v.Cards,
		"sid":        v.SID,
	}).Error; err != nil {
		log.Errorc(ctx, " db.Model(&model.Cards{}).update()  error(%v)", err)
		err = ecode.Error(ecode.RequestErr, "数据更新失败")
		return
	}
	ctx.JSON(nil, err)

}

func getCardsByLotteryID(ctx *bm.Context) {
	v := new(struct {
		LotteryID int64 `form:"lottery_id" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	db := actSrv.DB
	db = db.Table("act_cards")
	var err error
	var card = new(cards.Cards)

	if err = db.Where("lottery_id = ?", v.LotteryID).First(card).Error; err != nil {
		log.Error(" s.DB.Where(lottery_id ,%d).First() error(%v)", v.LotteryID, err)
	}
	data := map[string]interface{}{
		"data": card,
	}
	ctx.JSON(data, err)

}
func cardsUser(ctx *bm.Context) {

	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}

	ctx.JSON(actSrv.CardsMid(ctx, v.Mid))
}

func cardsInviteLog(ctx *bm.Context) {
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
	db = db.Table("act_cards_relation")
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where(midEqual, v.Mid)
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).Order(ctimeDesc).
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

func cardsComposeCount(ctx *bm.Context) {
	v := new(struct {
		Activity string `form:"activity" validate:"required"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var (
		err   error
		count int64
	)

	db := actSrv.DB
	db = db.Table("act_compose_card_log")

	db = db.Where("activity = ?", v.Activity)
	if err = db.Model(&model.ActComposeCardLog{}).Count(&count).Error; err != nil {
		log.Errorc(ctx, "sp2021CardsNumsLog count error(%v)", err)
		ctx.JSON(nil, err)
		return
	}
	reply := &model.ActComposeLogReply{}

	reply.Count = count
	data := map[string]interface{}{
		"data": reply,
	}
	ctx.JSONMap(data, nil)
}

func cardsNumsLog(ctx *bm.Context) {
	v := new(struct {
		Mid int64 `form:"mid" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	var (
		err   error
		count int
		list  []*model.ActYouthCardsNums
	)

	db := actSrv.DB
	db = db.Table("act_youth_cards_nums")

	db = db.Where(midEqual, v.Mid)
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
	reply := &model.ActYouthCardsNumsReply{}

	reply.List = list
	data := map[string]interface{}{
		"data": reply,
	}
	ctx.JSONMap(data, nil)
}

func cardsComposeLog(ctx *bm.Context) {
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
	db = db.Table("act_compose_card_log")
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where(midEqual, v.Mid)
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).Order(ctimeDesc).
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

func cardsSendCardLog(ctx *bm.Context) {
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
	db = db.Table("act_send_card_log")
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where(midEqual, v.Mid)
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).Order(ctimeDesc).
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
