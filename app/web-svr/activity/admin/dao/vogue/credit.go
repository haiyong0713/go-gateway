package vogue

import (
	"context"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

const (
	// 用户接任务表
	_tableUserTask = "act_vogue_user_task"
	// 用户消耗积分表，表中goods = 0 的为虚拟商品 - 抽奖，goods > 0 的为实物商品 - 兑换
	_tableUserCost = "act_vogue_user_cost"
	// 用户邀请表
	_tableUserInvite = "act_vogue_user_invite"
)

func (d *Dao) CreditList(c context.Context, params *voguemdl.CreditSearch) (list []*voguemdl.CreditData, count int64, err error) {
	db := d.DB.Table(_tableUserTask)

	db = db.Select("act_vogue_user_task.*, act_vogue_goods.name AS goods_name, act_vogue_goods.score AS goods_score_setting").
		Joins("LEFT JOIN act_vogue_goods ON act_vogue_goods.id = act_vogue_user_task.goods")

	if params.Uid > 0 {
		db = db.Where("uid = ?", params.Uid)
	}
	if err = db.Count(&count).Error; err != nil {
		log.Error("[VogueCreditList] count (%v) error (%v)", params, err)
		return
	}
	if count == 0 {
		list = make([]*voguemdl.CreditData, 0)
		return
	}
	db = db.Order("id desc")
	if params.Ps > 0 {
		db = db.Limit(params.Ps).Offset((params.Pn - 1) * params.Ps)
	}
	if err = db.Find(&list).Error; err != nil {
		log.Error("[VogueCreditList] d.DB.Find, error(%v)", err)
	}
	return
}

// TaskUsers 获取参与任务的用户
func (d *Dao) TaskUsers(c context.Context, params *voguemdl.CreditSearch) (users []int64, count int64, err error) {
	var (
		list []*voguemdl.CreditData
	)

	db := d.DB.Table(_tableUserTask)
	db = db.Select("uid")

	if params.Uid > 0 {
		db = db.Where("uid = ?", params.Uid)
	}
	if err = db.Count(&count).Error; err != nil {
		log.Error("[TaskUsers] count (%v) error (%v)", params, err)
		return
	}
	if count == 0 {
		users = make([]int64, 0)
		return
	}
	db = db.Order("id desc")
	if params.Ps > 0 {
		db = db.Limit(params.Ps).Offset((params.Pn - 1) * params.Ps)
	}
	if err = db.Find(&list).Error; err != nil {
		log.Error("[TaskUsers] d.DB.Find, error(%v)", err)
		return
	}

	for _, item := range list {
		log.Info("%v", item)
		users = append(users, item.Uid)
	}

	log.Info("users: %v", users)

	return
}

// UsersTaskInfo 获取用户参与任务详情
func (d *Dao) UsersTaskInfo(c context.Context, mids []int64) (res map[int64]*voguemdl.CreditData, err error) {
	db := d.DB.Table(_tableUserTask)

	var list []*voguemdl.CreditData
	db = db.Select("act_vogue_user_task.*, act_vogue_goods.name AS goods_name, act_vogue_goods.score AS goods_score_setting").
		Joins("LEFT JOIN act_vogue_goods ON act_vogue_goods.id = act_vogue_user_task.goods").
		Where("act_vogue_user_task.uid IN (?)", mids)
	if err = db.Find(&list).Error; err != nil {
		log.Error("[UsersTaskInfo] d.DB.Find, error(%v)", err)
		return
	}

	res = make(map[int64]*voguemdl.CreditData)
	for _, item := range list {
		res[item.Uid] = item
	}

	return
}

// 邀请用户每日积分
func (d *Dao) CreditInviteUsersDailySum(c context.Context, mids []int64) (res map[int64]map[xtime.Time]int64, err error) {
	if len(mids) == 0 {
		return
	}
	db := d.DB.Table(_tableUserInvite)
	var list []*voguemdl.CreditUserInvite
	if err = db.Select("uid, score, ctime").Where("uid IN (?)", mids).Find(&list).Error; err != nil {
		log.Error("[VogueCreditInviteUsersSum] d.DB.Find, error(%v)", err)
		return
	}
	res = make(map[int64]map[xtime.Time]int64)
	for _, item := range list {
		if _, ok := res[item.Uid]; !ok {
			res[item.Uid] = make(map[xtime.Time]int64)
		}
		year, month, day := item.Ctime.Time().Date()
		today := xtime.Time(time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix())
		if _, ok := res[item.Uid][today]; !ok {
			res[item.Uid][today] = 0
		}
		res[item.Uid][today] += item.Score
	}
	return
}

// 消耗积分
func (d *Dao) CreditCostSum(c context.Context, mids []int64) (res map[int64]int64, err error) {
	if len(mids) == 0 {
		return
	}
	db := d.DB.Table(_tableUserCost)
	var list []*voguemdl.CreditUserCost
	if err = db.Select("mid, SUM(cost) AS cost").Where("mid IN (?)", mids).Group("mid").Find(&list).Error; err != nil {
		log.Error("[VogueCreditCostSum] d.DB.Find, error(%v)", err)
		return
	}
	res = make(map[int64]int64, len(list))
	for _, item := range list {
		res[item.Mid] = item.Cost
	}
	return
}

// 获取观看视频每日所得积分 - actPlatClient有不兼容更新，活动结束后下掉对应功能
func (d *Dao) CreditViewDayMap(c context.Context, mid int64) (resMap map[xtime.Time]int64, err error) {
	//var (
	//	start []byte
	//)
	resMap = make(map[xtime.Time]int64)
	//for {
	//	var resp *actPlat.GetCounterResResp
	//	if resp, err = d.actPlatClient.GetCounterRes(c, &actPlat.GetCounterResReq{
	//		Counter:     voguemdl.MethodView,
	//		Activity: d.c.VogueActivity.ActPlatActivity,
	//		Mid:      mid,
	//		Start:    start,
	//	}); err != nil {
	//		log.Error("[VogueCreditViewSum] d.actPlatClient.GetCounterRes, mid(%d) error(%v)", mid, err)
	//		return
	//	}
	//	log.Info("[VogueCreditViewSum] d.actPlatClient.GetCounterRes, mid(%d) resp(%v)", mid, resp)
	//	for _, n := range resp.CounterList {
	//		year, month, day := time.Unix(n.Time, 0).Date()
	//		today := xtime.Time(time.Date(year, month, day, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1).Unix())
	//		resMap[today] = n.Val
	//	}
	//	if len(resp.CounterList) < voguemdl.ActPlatGetLimit {
	//		break
	//	}
	//}
	return
}

// 邀请用户积分列表
func (d *Dao) CreditInviteList(c context.Context, params *voguemdl.CreditDetailSearch) (list []*voguemdl.CreditItem, err error) {
	db := d.DB.Table(_tableUserInvite)
	if err = db.Select("ctime, uid, score, mid AS friend, ? AS category, ? AS method, ? AS score_symbol", voguemdl.CategoryDeposit, voguemdl.MethodInvite, voguemdl.ScoreSymbolPositive).Where("uid = ?", params.Uid).Find(&list).Error; err != nil {
		log.Error("[VogueCreditInviteUsersSum] d.DB.Find, error(%v)", err)
		return
	}
	return
}

// 消耗积分列表
func (d *Dao) CreditCostList(c context.Context, params *voguemdl.CreditDetailSearch) (list []*voguemdl.CreditItem, err error) {
	db := d.DB.Table(_tableUserCost)
	if err = db.Select("ctime, mid AS uid, cost AS score, ? AS category, IF(goods = 0, ?, ?) AS method, ? AS score_symbol", voguemdl.CategoryWithdraw, voguemdl.MethodLottery, voguemdl.MethodPrize, voguemdl.ScoreSymbolNegtive).Where("mid = ?", params.Uid).Find(&list).Error; err != nil {
		log.Error("[VogueCreditInviteUsersSum] d.DB.Find, error(%v)", err)
		return
	}
	return
}

// 观看视频积分列表 - actPlatClient有不兼容更新，活动结束后下掉对应功能
func (d *Dao) CreditViewList(c context.Context, params *voguemdl.CreditDetailSearch) (list []*voguemdl.CreditItem, err error) {
	//var start []byte
	//for {
	//	var resp *actPlat.GetHistoryResp
	//	if resp, err = d.actPlatClient.GetHistory(c, &actPlat.GetHistoryReq{
	//		Counter:  voguemdl.MethodView,
	//		Activity: d.c.VogueActivity.ActPlatActivity,
	//		Mid:      params.Uid,
	//		Start:    start,
	//	}); err != nil {
	//		log.Error("[VogueCreditViewList] d.actPlatClient.GetHistory, params(%v) error(%v)", params, err)
	//		return
	//	}
	//	log.Info("[VogueCreditViewList] d.actPlatClient.GetHistory, resp(%v)", resp)
	//	for _, history := range resp.History {
	//		list = append(list, &voguemdl.CreditItem{
	//			Ctime:    xtime.Time(history.Time),
	//			Score:    history.Content.Count,
	//			Detail:   history.Content.Source,
	//			Category: voguemdl.CategoryDeposit,
	//			Method:   voguemdl.MethodView,
	//		})
	//	}
	//	if len(resp.History) < voguemdl.ActPlatGetLimit {
	//		break
	//	}
	//}
	return
}
