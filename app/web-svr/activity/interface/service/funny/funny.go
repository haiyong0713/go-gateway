package funny

import (
	"context"
	"fmt"
	accountapi "git.bilibili.co/bapis/bapis-go/account/service"
	actPlat "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/client"
	f "go-gateway/app/web-svr/activity/interface/model/funny"
	"time"
)

// 获取页面两个数字 和 是否已经增加过抽奖次数
func (s *Service) PageInfo(c context.Context) (*f.PageInfoReply, error) {
	task1, _ := s.funny.GetTask1Num(c)
	task2, _ := s.funny.GetTask2Num(c)
	return &f.PageInfoReply{Task1: task1, Task2: task2}, nil
}

// Likes 获取点赞数
func (s *Service) Likes(c context.Context, mid int64) (*f.LikesReply, error) {
	var (
		err    error
		count  int64
		status int //  0 无资格领取 1 未领取 2 已领取
	)

	if count, err = s.getMidLikes(c, mid); err != nil {
		log.Errorc(c, "api Likes getMidLikes (%d) error(%v)", mid, err)
		return nil, ecode.ActivityFunnyLikeGetErr
	}

	// 点赞数量超过10次 前端统一显示为10次
	if count > s.c.Funny.AwardLikeLimit {
		count = s.c.Funny.AwardLikeLimit
	}

	isAdd, _ := s.funny.GetUserTodayIsAdded(c, mid)

	// 未领取
	status = 1
	if count < s.c.Funny.AwardLikeLimit {
		status = 0
	}
	if isAdd == 1 {
		status = 2
	}

	return &f.LikesReply{Likes: count, Status: status}, nil
}

// getMidLikes 获取用户点赞信息
func (s *Service) getMidLikes(c context.Context, mid int64) (int64, error) {
	req := &actPlat.GetCounterResReq{
		Counter:  s.c.Funny.ActPlatCounter,
		Activity: s.c.Funny.ActPlatActivity,
		Mid:      mid,
		Time:     time.Now().Unix(),
	}
	resp, err := client.ActPlatClient.GetCounterRes(c, req)
	if err != nil {
		log.Errorc(c, "api getMidLikes.actPlatClient.GetCounterRes mid:%v Req(%v) Resp(%v) error(%v)", mid, req, resp, err)
		return 0, err
	}
	if resp == nil || len(resp.CounterList) != 1 {
		log.Errorc(c, "api getMidLikes.actPlatClient.GetCounterRes mid:%v Req(%v) Resp(%v) error(%v)", mid, req, resp, err)
		return 0, err
	}
	counter := resp.CounterList[0]
	return counter.Val, nil
}

// 增加抽奖次数
func (s *Service) AddLotteryTimes(c context.Context, mid int64) (*f.AddTimesReply, error) {
	// 账号信息验证
	if err := s.checkAccountInfo(c, mid); err != nil {
		log.Errorc(c, "api AddLotteryTimes checkAccountInfo error mid:%v err:%v", mid, err)
		return nil, err
	}
	// 获取当前点赞数
	like, err := s.getMidLikes(c, mid)
	if err != nil {
		log.Errorc(c, "api AddLotteryTimes getMidLikes mid:%v err :%v", mid, err)
		return nil, ecode.ActivityFunnyLikeGetErr
	}

	// 小于10次无法增加抽奖次数
	if like < s.c.Funny.AwardLikeLimit {
		log.Infoc(c, "api AddLotteryTimes mid: %v likes :%v", mid, like)
		return nil, ecode.ActivityFunnyLikeNotEnoughErr
	}
	// 生成订单号
	orderNo := s.getOrderNo(c, mid)
	// 增加获奖次数
	if err = s.like.AddLotteryTimes(c, s.c.Funny.Sid, mid, 0, 7, 1, fmt.Sprint(orderNo), false); err != nil {
		// 当日增加抽奖次数达到上限
		if err == ecode.ActivityLotteryAddTimesLimit {
			// 下方没有retry 所以查询到当日已经增加过 再次写缓存
			log.Infoc(c, "[api AddLotteryTimes already add times mid:%v]", mid)
			_ = s.funny.SetUserAddedTimes(c, mid)
			return nil, ecode.ActivityFunnyAddTimesLimit
		} else {
			log.Errorc(c, "[api AddLotteryTimes Err sid:%v mid:%v cid:0 actionType:7 num:0 orderNo:%v isOut:false err:%v]", s.c.Funny.Sid, mid, orderNo, err)
			return nil, ecode.ActivityFunnyAddTimesErr
		}
	}
	// 写缓存
	_ = s.funny.SetUserAddedTimes(c, mid)

	return nil, nil
}

// 账号检查
func (s *Service) checkAccountInfo(c context.Context, mid int64) (err error) {
	var profileReply *accountapi.ProfileReply
	if profileReply, err = s.accClient.Profile3(c, &accountapi.MidReq{
		Mid: mid,
	}); err != nil {
		log.Errorc(c, "api AddLotteryTimes checkAccountInfo accClient.Profile3(%v) error(%v)", mid, err)
		return nil
	}
	if profileReply.Profile.GetTelStatus() != 1 {
		return ecode.ActivityFunnyTelValid
	}
	if profileReply.Profile.GetSilence() == 1 {
		return ecode.ActivityFunnyBlocked
	}
	return
}

// 创建订单号
func (s *Service) getOrderNo(c context.Context, mid int64) string {
	return fmt.Sprintf("%d_%v_%v", mid, "funny", time.Now().Format("20060102"))
}
