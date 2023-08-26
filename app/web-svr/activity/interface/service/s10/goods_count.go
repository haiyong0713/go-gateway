package s10

import (
	"context"

	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/s10"

	"go-common/library/log"
	xtime "go-common/library/time"
)

func (s *Service) goodsRestCountRebuild(ctx context.Context, gid int32) error {
	_, load := s.singleFlightForRestCount.LoadOrStore(gid, struct{}{})
	if load {
		return ecode.ActivityExchangePointFail
	}
	defer s.singleFlightForRestCount.Delete(gid)
	stock, send, _, err := s.dao.GoodsByID(ctx, gid)
	if err != nil {
		return ecode.ActivityExchangePointFail
	}
	rest := int64(stock - send - 1)
	var flag bool
	if rest < 0 {
		rest = 0
		flag = true
	}
	err = s.dao.AddRestCountByGoodsCache(ctx, gid, rest)
	if err != nil || flag {
		log.Errorc(ctx, "s10 goodsRestCountRebuild err:%v flag:%v", err, flag)
		return ecode.ActivityExchangePointFail
	}
	return nil
}

func (s *Service) goodsRoundRestCountRebuild(ctx context.Context, gid int32, currTime xtime.Time) error {
	_, load := s.singleFlightForRoundRestCount.LoadOrStore(gid, struct{}{})
	if load {
		return ecode.ActivityExchangePointFail
	}
	defer s.singleFlightForRoundRestCount.Delete(gid)
	roundSend, roundStock, exist, err := s.dao.RoundGoodsByID(ctx, gid, currTime)
	if err != nil {
		return err
	}
	if !exist {
		info, ok := s.goodsInfo.Load().(map[int32]*s10.Bonus)
		if !ok {
			return ecode.ActivityExchangePointFail
		}
		if v, ok := info[gid]; ok {
			roundStock = v.RoundStock
		}
		if _, err = s.dao.AddGoodsRoundSendCount(ctx, gid, roundStock, currTime); err != nil {
			return err
		}
	}
	flag := false
	roundRest := int64(roundStock - roundSend - 1)
	if roundRest < 0 {
		roundRest = 0
		flag = true
	}
	err = s.dao.AddRoundRestCountByGoodsCache(ctx, gid, roundRest, currTime)
	if err != nil || flag {
		log.Errorc(ctx, "s10 goodsRoundRestCountRebuild err:%v flag:%v", err, flag)
		return ecode.ActivityExchangePointFail
	}
	return nil
}
