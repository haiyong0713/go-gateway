package vogue

import (
	"context"
	"fmt"
	"math"
	"time"

	accountAPI "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/log"

	"go-gateway/app/web-svr/activity/ecode"
	model "go-gateway/app/web-svr/activity/interface/model/vogue"

	silver "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
)

func (s *Service) Prizes(c context.Context) (res []*model.Prize, err error) {
	var goods []*model.Goods
	if goods, err = s.dao.GoodsList(c); err != nil {
		log.Error("s.dao.GoodsList(%v)", err)
		return nil, err
	}
	res = make([]*model.Prize, 0, len(goods))
	for _, n := range goods {
		if n.AttrVal(model.GoodsAttrSellOut) == 1 || n.Stock-n.Send <= 0 {
			continue
		}
		p := &model.Prize{
			Id:      n.Id,
			Name:    n.Name,
			Type:    n.Type,
			Picture: n.Picture,
			Stock:   n.Stock - n.Send,
			Score:   n.Score,
			Want:    n.Want,
		}
		if p.Want <= 20 {
			p.Want = int64(time.Now().Nanosecond()/1000%20 + 5)
		}
		res = append(res, p)
	}
	return
}

func (s *Service) SelectPrizes(c context.Context, mid, PrizeId int64, Token string) (res *model.SelectPrize, err error) {
	if err = s.inActive(c); err != nil {
		return nil, err
	}
	if err = s.checkAccountInfo(c, mid); err != nil {
		return nil, err
	}
	// 邀请
	if Token != "" {
		var uid, score int64
		if uid, err = model.TokenEncode(Token); err != nil {
			log.Error("s.model.TokenEncode(%v)", err)
			return nil, err
		}
		if score, err = s.getInviteScore(c, uid, mid); err != nil {
			log.Error("s.getInviteScore(%v,%v,%v)", mid, uid, err)
			return nil, err
		}
		if _, err = s.dao.InsertInvite(c, uid, mid, score); err != nil {
			log.Error("s.dao.InsertInvite(%v,%v,%v)", mid, PrizeId, err)
			return nil, err
		}
		_ = s.dao.DelCacheTask(c, uid)
		_ = s.dao.DelCacheInviteList(c, uid, math.MaxInt32)
	}
	// 领取任务
	var (
		goods  *model.Goods
		affect int64
	)
	if goods, err = s.dao.Goods(c, PrizeId); err != nil {
		return
	}
	if goods == nil || goods.Id == 0 {
		err = ecode.ActivityGoodsNoFind
		return nil, err
	}
	if goods.Stock-goods.Send <= 0 || goods.AttrVal(model.GoodsAttrSellOut) == 1 {
		err = ecode.ActivityShortage
		return nil, err
	}
	//if _, err = client.ActPlatClient.AddFilterMemberInt(c, &actPlat.SetFilterMemberIntReq{
	//	Activity: model.ActPlatActivity,
	//	Counter:  model.ActPlatCounter,
	//	Filter:   model.ActPlatMidFilter,
	//	Values:   []*actPlat.FilterMemberInt{{Value: mid}},
	//}); err != nil {
	//	log.Error("s.actPlatClient.AddFilterMemberMidInt(%v,%v)", mid, err)
	//	return nil, err
	//}
	if affect, err = s.dao.InsertTask(c, mid, PrizeId); err != nil {
		log.Error("s.dao.InsertTask(%v,%v,%v)", mid, PrizeId, err)
		return nil, err
	}
	_ = s.dao.DelCacheTask(c, mid)
	if affect <= 0 {
		err = ecode.ActivitySelected
		return nil, err
	}
	if err = s.dao.GoodsAddWant(c, PrizeId); err != nil {
		log.Error("s.dao.GoodsAddWant(%v,%v,%v)", mid, PrizeId, err)
		err = nil
	}
	_ = s.dao.DelCacheGoods(c, PrizeId)
	_ = s.dao.DelCacheGoodsList(c)
	res = &model.SelectPrize{
		InitScore: 20,
	}
	return
}

func (s *Service) getInviteScore(c context.Context, uid, mid int64) (score int64, err error) {
	var (
		todayScore, start, end, limit int64
	)
	if err = s.checkAccountInfo(c, mid); err != nil {
		return 0, nil
	}
	if score, err = s.inviteScore(c); err != nil {
		log.Error("s.inviteScore(%v)", err)
		return 0, err
	}
	if start, end, err = s.secondDoubleTime(c); err != nil {
		log.Error("s.doubleTime(%v)", err)
		return 0, err
	}
	if limit, err = s.todayLimit(c); err != nil {
		log.Error("s.todayLimit(%v)", err)
		return 0, err
	}
	now := time.Now().Unix()
	if start < now && now < end {
		limit *= 2
		score *= 2
	}
	// 获取最高分
	if _, todayScore, err = s.getScore(c, uid); err != nil {
		log.Error("s.dao.getScore(%v,%v)", mid, err)
		return 0, err
	}
	max := limit - todayScore
	if score > max {
		score = max
	}
	return
}

func (s *Service) checkAccountInfo(c context.Context, mid int64) (err error) {
	var profileReply *accountAPI.ProfileReply
	if profileReply, err = s.accClient.Profile3(c, &accountAPI.MidReq{
		Mid: mid,
	}); err != nil {
		log.Error("accClient.Profile3(%v) error(%v)", mid, err)
		return nil
	}
	if profileReply.Profile.GetTelStatus() != 1 {
		return ecode.ActivityVogueTelValid
	}
	if profileReply.Profile.GetSilence() == 1 {
		return ecode.ActivityVogueBlocked
	}
	return
}

func (s *Service) Addtimes(c context.Context, mid int64, ip string) (res *model.AddRes, err error) {
	if err = s.inActive(c); err != nil {
		return
	}
	var (
		ok            bool
		orderNo, cost int64
	)
	if ok, err = s.dao.AddTimeLock(c, mid); err != nil {
		log.Error("s.dao.AddTimeLock(%v)", err)
	}
	if !ok {
		err = ecode.ActivityRapid
		return
	}
	defer s.dao.DelTimeLock(c, mid)
	if err = s.checkAccountInfo(c, mid); err != nil {
		return nil, err
	}
	if err = s.risk(c, mid, ip); err != nil {
		return nil, err
	}
	if cost, _, err = s.getScore(c, mid); err != nil {
		log.Error("s.getScore(%v)", err)
		return nil, err
	}
	if cost < s.c.Vogue.PrizeCost {
		err = ecode.ActivityInsufficient
		return
	}
	if orderNo, err = s.dao.InsertUserCost(c, mid, 0, 0); err != nil {
		log.Error("s.dao.InsertUserCost(%v)", err)
		return nil, err
	}
	if err = s.AddLotteryTimes(c, s.c.Vogue.Sid, mid, 0, 7, 0, fmt.Sprint(orderNo), false); err != nil {
		log.Error("s.dao.Addtimes(%v)", err)
		return nil, err
	}
	if _, err = s.dao.UpdateUserCost(c, orderNo, s.c.Vogue.PrizeCost); err != nil {
		log.Error("s.dao.UpdateUserCost(%v)", err)
		return nil, err
	}
	res = &model.AddRes{
		Score: cost - s.c.Vogue.PrizeCost,
	}
	return
}

func (s *Service) Exchange(c context.Context, mid int64, ip string) (err error) {
	if err = s.inActive(c); err != nil {
		return
	}
	var (
		state = 3
		task  *model.Task
		goods *model.Goods
		cost  int64
	)
	if task, err = s.dao.Task(c, mid); err != nil {
		log.Error("s.dao.Task(%v)", err)
		return err
	}
	if task.GoodsState != model.GoodsStateNormal {
		err = ecode.ActivityCashed
		return
	}
	if goods, err = s.dao.Goods(c, task.Goods); err != nil {
		log.Error("s.dao.Goods(%v)", err)
		return err
	}
	if goods.Stock-goods.Send <= 0 || goods.AttrVal(model.GoodsAttrSellOut) == 1 {
		err = ecode.ActivityShortage
		return err
	}
	if cost, _, err = s.getScore(c, mid); err != nil {
		log.Error("s.getScore(%v)", err)
		return err
	}
	if cost < goods.Score {
		err = ecode.ActivityInsufficient
		return
	}
	if err = s.checkAccountInfo(c, mid); err != nil {
		return err
	}
	if err = s.risk(c, mid, ip); err != nil {
		return err
	}
	var affect int64
	if affect, err = s.dao.GoodsAddSend(c, task.Goods); err != nil {
		log.Error("s.dao.GoodsAddSend(%v)", err)
		err = nil
	}
	if affect <= 0 {
		err = ecode.ActivityShortage
		return err
	}
	if _, err = s.dao.InsertUserCost(c, mid, goods.Score, goods.Id); err != nil {
		log.Error("s.dao.InsertUserCost(%v)", err)
		return err
	}
	if goods.AttrVal(model.GoodsAttrReal) == 1 {
		state = model.GoodsStateAddress
	}
	if _, err = s.dao.UpdateTask(c, mid, state); err != nil {
		return err
	}
	_ = s.dao.DelCacheTask(c, mid)
	_ = s.dao.DelCacheGoods(c, goods.Id)
	_ = s.dao.DelCacheGoodsList(c)
	return
}

func (s *Service) Address(c context.Context, mid int64, addressId int64, ip string) (err error) {
	if err = s.inActive(c); err != nil {
		return
	}
	var (
		task  *model.Task
		goods *model.Goods
	)
	if task, err = s.dao.Task(c, mid); err != nil {
		log.Error("s.dao.Task(%v)", err)
		return err
	}
	if task.GoodsAddress != 0 {
		err = ecode.ActivityAddrHasAdd
		return err
	}
	if task.GoodsState != model.GoodsStateAddress {
		err = ecode.ActivityTaskPreNotCheck
		return err
	}
	if goods, err = s.dao.Goods(c, task.Goods); err != nil {
		log.Error("s.dao.Goods(%v)", err)
		return err
	}
	if goods.AttrVal(model.GoodsAttrReal) == 0 {
		err = ecode.ActivityAddrNotNeed
		return err
	}
	if err = s.checkAccountInfo(c, mid); err != nil {
		return err
	}
	if err = s.risk(c, mid, ip); err != nil {
		return err
	}
	if _, err = s.dao.UpdateTaskAddress(c, mid, model.GoodsStateShipping, addressId); err != nil {
		return err
	}
	_ = s.dao.DelCacheTask(c, mid)
	_ = s.dao.DelCacheGoods(c, goods.Id)
	_ = s.dao.DelCacheGoodsList(c)
	return
}

func (s *Service) InviteList(c context.Context, mid, id int64) (res []*model.InviteItem, err error) {
	if id <= 0 {
		id = math.MaxInt32
	}
	data, err := s.dao.InviteList(c, mid, id)
	if err != nil {
		log.Error("s.dao.InviteList(%v)", err)
		return nil, err
	}
	res = make([]*model.InviteItem, 0, len(data))
	if len(data) <= 0 {
		return
	}
	var mids []int64
	for _, n := range data {
		mids = append(mids, n.Mid)
	}
	user, err := s.accClient.Infos3(c, &accountAPI.MidsReq{
		Mids: mids,
	})
	if err != nil {
		log.Error("s.accClient.Infos3(%v)", err)
		return nil, err
	}
	var picture string
	for _, n := range data {
		picture = "http://i0.hdslb.com/bfs/face/member/noface.jpg"
		if userInfo, ok := user.GetInfos()[n.Mid]; ok {
			picture = userInfo.Face
		}
		item := &model.InviteItem{
			Id:      n.Id,
			Uid:     n.Uid,
			Mid:     n.Mid,
			Picture: picture,
			Score:   n.Score,
		}
		res = append(res, item)
	}
	return
}

// 风控黑名单
func (s *Service) risk(c context.Context, mid int64, ip string) (err error) {
	var riskinfo *silver.RiskInfoReply
	if riskinfo, err = s.silverBulletClient.RiskInfo(c, &silver.RiskInfoReq{
		StrategyName: []string{"618_room_of_requirement"},
		Mid:          mid,
		Ip:           ip,
	}); err != nil {
		log.Error("s.silverBulletClient.RiskInfo(%v)", err)
	} else {
		if risk, ok := riskinfo.GetInfos()["618_room_of_requirement"]; ok && risk.Score >= 100 {
			err = ecode.ActivityVogueNotAward
			return err
		}
	}
	return
}
