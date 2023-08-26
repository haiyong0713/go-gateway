package grpc

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/service"

	v1 "go-gateway/app/web-svr/activity/interface/api"
	dao "go-gateway/app/web-svr/activity/interface/dao/newyear2021"
	"go-gateway/app/web-svr/activity/interface/service/newyear2021"
)

func (s *activityService) BNJARIncrCoupon(ctx context.Context, req *v1.BNJ2021ARCouponReq) (reply *v1.BNJ2021ARCouponReply, err error) {
	reply = new(v1.BNJ2021ARCouponReply)
	{
		reply.Coupon = req.Coupon
	}

	if req.Mid > 0 && req.Coupon > 0 {
		err = newyear2021.IncrARCoupon(ctx, req)
	}

	return
}

func (s *activityService) BNJARExchange(ctx context.Context, req *v1.BNJ2021ARExchangeReq) (reply *v1.BNJ2021ARExchangeReply, err error) {
	reply = new(v1.BNJ2021ARExchangeReply)
	reply, err = newyear2021.InsertARGameLog(ctx, req)

	return
}

func (s *activityService) BNJ2021ShareData(ctx context.Context, req *v1.BNJ2021ShareReq) (reply *v1.BNJ2021ShareReply, err error) {
	reply = new(v1.BNJ2021ShareReply)
	reward := newyear2021.LastARRewardByMID(ctx, req.Mid)
	reply.Coupon = reward.Coupon
	reply.Score = reward.Score

	return
}

func (s *activityService) BNJ2021LastLotteryData(ctx context.Context, req *v1.BNJ2021LastLotteryReq) (reply *v1.BNJ2021LastLotteryReply, err error) {
	reply = new(v1.BNJ2021LastLotteryReply)
	{
		reply.Name, err = newyear2021.LastDrawAwardByMID(ctx, req.Mid)
	}

	return
}

func (s *activityService) UpdateExamStats(ctx context.Context, req *v1.ExamStatsReq) (reply *v1.ExamStatsReply, err error) {
	_ = dao.MultiUpdateExamStats(ctx, req)
	reply = new(v1.ExamStatsReply)
	{
		reply.Message = "ok"
	}

	return
}

func (s *activityService) AppJumpUrl(ctx context.Context, req *v1.AppJumpReq) (reply *v1.AppJumpReply, err error) {
	reply = new(v1.AppJumpReply)
	switch req.BizType {
	case v1.AppJumpBizType_Type4Bnj2021AR:
		reply = newyear2021.GenAppRedirect4Gateway(ctx, req)
	case v1.AppJumpBizType_Type4Bnj2021TaskGame:
		reply = newyear2021.GenAppTaskGameRedirect4Gateway(ctx, req)
	}

	return
}

func (s *activityService) Bnj2021Lottery(ctx context.Context, req *v1.Bnj2021LotteryReq) (reply *v1.Bnj2021LotteryReply, err error) {
	reply = &v1.Bnj2021LotteryReply{
		List: make([]*v1.RewardsSendAwardReply, 0),
	}
	ctxForever := context.Background()
	list, err := service.NewYear2021Svc.DoLottery(ctxForever, req.Mid, req.Type, req.Count, service.LotterySvc, nil, req.ActivityId, req.NeedSend, req.Debug, req.UpdateCache, req.UpdateDb)
	for _, l := range list {
		reply.List = append(reply.List, &v1.RewardsSendAwardReply{
			ActivityId:   l.ActivityId,
			ActivityName: l.ActivityName,
			Type:         l.Type,
			Name:         l.Name,
			Icon:         l.Icon,
			ExtraInfo:    l.ExtraInfo,
			AwardId:      l.AwardId,
			Mid:          l.Mid,
		})
	}

	return
}
