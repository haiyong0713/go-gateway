package like

import (
	"context"
	"strconv"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/lottery"
	riskmdl "go-gateway/app/web-svr/activity/interface/model/risk"

	passapi "git.bilibili.co/bapis/bapis-go/passport/service/user"
	spapi "git.bilibili.co/bapis/bapis-go/silverbullet/service/silverbullet-proxy"
)

const (
	_wxLotteryDoAPI            = "/x/activity/lottery/wx/do"
	_wxLotteryDoKey            = "wx_lot_do"
	_wxLotteryStrategy         = "applet_lottery"
	_wxLotteryDoExpire         = 1
	_wxLotteryNum              = 1
	_noticeType                = 10
	_notLottery                = 1
	_wxUserTypeNone            = 0
	_wxUserTypeNew             = 1
	_wxUserTypeNormal          = 2
	_wxUserTypeVip             = 3
	_wxLotteryPlayWindowNeed   = 0 // 未弹过，需要弹
	_wxLotteryPlayWindowUnNeed = 1 // 弹过，不需要弹
)

func (s *Service) WxDoLottery(ctx context.Context, mid, platform int64, buvid, ua, referer, origin, ip string, risk *riskmdl.Base) (*lottery.WxLotteryRes, error) {
	if err := s.checkTimeConf(ctx); err != nil {
		return nil, err
	}
	// 限速
	keyMid := _wxLotteryDoKey + strconv.FormatInt(mid, 10)
	if canLottery, err := s.dao.RsSetNX(ctx, keyMid, _wxLotteryDoExpire); err != nil || !canLottery {
		log.Error("WxDoLottery d.RsSetNX error(%+v)", err)
		return nil, ecode.ActivityInLottery
	}
	// 检验是否抽过
	lotteryLog, err := s.lottDao.WxLotteryLog(ctx, mid)
	if err != nil {
		log.Error("WxDoLottery s.lottDao.WxLotteryLog mid(%d) error(%+v)", mid, err)
		return nil, ecode.ActivityLotteryErr
	}
	if lotteryLog != nil {
		log.Warn("WxDoLottery mid:%d had lottery", mid)
		return nil, ecode.ActivityHadLottery
	}
	// 验证风控是否通过.
	riskReq := &spapi.RiskInfoReq{
		Mid:          mid,
		Api:          _wxLotteryDoAPI,
		StrategyName: []string{_wxLotteryStrategy},
		Ip:           metadata.String(ctx, metadata.RemoteIP),
		DeviceId:     buvid,
		Ua:           ua,
		Referer:      referer,
		Origin:       origin,
	}
	isAllowed := s.isAllowedByRisk(ctx, riskReq)
	if !isAllowed {
		log.Warn("WxDoLottery s.dao.isAllowedByRisk true req(%+v)", riskReq)
		return nil, ecode.ActivityLotteryUserUnusual
	}
	var logID int64
	if logID, err = s.lottDao.AddWxLotteryLog(ctx, mid, platform, lottery.HandleLotteryFrom(referer), buvid); err != nil {
		log.Error("WxDoLottery s.lottDao.AddWxLotteryLog mid(%d) error(%+v)", mid, err)
		return nil, ecode.ActivityLotteryErr
	}
	defer func() {
		s.cache.Do(ctx, func(ctx context.Context) {
			if _, e := s.lottDao.AddWxLotteryHis(ctx, mid, buvid); e != nil {
				log.Error("WxDoLottery s.lottDao.AddWxLotteryHis mid:%d buvid:%s error:%v", mid, buvid, e)
			}
			if e := s.lottDao.DelCacheWxLotteryLog(ctx, mid, buvid); e != nil {
				log.Error("WxDoLottery s.lottDao.DelCacheWxLotteryLog mid:%d error:%v", mid, e)
			}
		})
	}()

	res := new(lottery.WxLotteryRes)
	reply, err := s.lotterySvr.DoLottery(ctx, s.c.WxLottery.RewardSid, mid, risk, _wxLotteryNum, true, "")
	if err != nil {
		log.Error("WxDoLottery reward s.DoLottery mid:%d sid:%s error(%+v)", mid, s.c.WxLottery.RewardSid, err)
		if s.systemErr(err) {
			return nil, ecode.ActivityLotteryErr
		}
		return nil, err
	} else if len(reply) > 0 && reply[0] != nil {
		res = lottery.HandleRecordDetail(reply[0], s.c.WxLottery.RewardSid, _wxUserTypeNone, s.c.WxLottery.MoneyMap)
	}
	if err = s.lottDao.UpWxLotteryLog(ctx, mid, logID, res.Type, res.UserType, res.GiftID, res.Money, res.LotteryID, res.GiftName); err != nil {
		log.Error("WxDoLottery UpWxLotteryLog mid:%d logID:%d error(%v)", mid, logID, err)
	}
	res.JumpURL = s.c.WxLottery.DoJumpURL

	//判断新客
	checkNew, _ := s.checkThreeNewUser(ctx, mid, buvid)

	if checkNew {
		res.IsNew = 1
	}

	// 发送中奖私信.
	if res.GiftID > 0 {
		s.cache.Do(ctx, func(ctx context.Context) {
			notifyCode, ok := s.c.WxLottery.MessageMap[strconv.FormatInt(res.Type, 10)]
			if !ok || notifyCode == "" {
				log.Warn("WxDoLottery SendLetter notifyCode not found res:%+v", res)
				return
			}
			sendArg := &lottery.LetterParam{
				RecverIDs:  []uint64{uint64(mid)},
				SenderUID:  s.c.WxLottery.SenderUID,
				MsgType:    _noticeType,
				NotifyCode: notifyCode,
				Params:     res.GiftName,
			}
			s.lottDao.SendLetter(ctx, sendArg)
			return
		})
	}
	return res, nil
}

func (s *Service) WxLotteryAward(ctx context.Context, mid int64, buvid string) (*lottery.WxAwardRes, error) {
	actErr := s.checkTimeConf(ctx)

	if mid <= 0 {
		his, err := s.lottDao.WxLotteryHisByBuvid(ctx, buvid)
		if err != nil {
			log.Error("WxLotteryAward s.lottDao.WxLotteryHisByBuvid buvid(%s) error(%+v)", buvid, err)
			return nil, ecode.ActivityLotteryErr
		}
		if his != nil && his.Mid > 0 {
			// buvid 抽过，播放窗口不弹窗
			return &lottery.WxAwardRes{PlayWindow: _wxLotteryPlayWindowUnNeed}, nil
		}
		if actErr != nil {
			// 活动未开始或结束
			return &lottery.WxAwardRes{PlayWindow: _wxLotteryPlayWindowUnNeed}, nil
		}
		return &lottery.WxAwardRes{NotLottery: _notLottery, PlayWindow: s.checkLotteryPlayWindow(ctx, mid, buvid)}, nil
	}
	lotteryLog, err := s.lottDao.WxLotteryLog(ctx, mid)
	if err != nil {
		log.Error("WxLotteryAward s.lottDao.WxLotteryLog mid(%d) error(%+v)", mid, err)
		return nil, ecode.ActivityLotteryErr
	}
	if lotteryLog != nil {
		var imgURL string
		if lotteryLog.LotteryID != "" {
			if lotteryLog.GiftID > 0 {
				if lotteryGift, err := s.lottDao.LotteryGift(ctx, lotteryLog.LotteryID); err != nil {
					log.Error("WxLotteryAward LotteryGift sid:%s error:%v", lotteryLog.LotteryID, err)
					err = nil
				} else if len(lotteryGift) > 0 {
					for _, v := range lotteryGift {
						if v != nil && v.ID == lotteryLog.GiftID {
							lotteryLog.GiftName = v.Name
							imgURL = v.ImgUrl
						}
					}
				}
			}
		}
		return &lottery.WxAwardRes{
			PlayWindow: _wxLotteryPlayWindowUnNeed,
			Mid:        lotteryLog.Mid,
			GiftID:     lotteryLog.GiftID,
			GiftName:   lotteryLog.GiftName,
			ImgURL:     imgURL,
			Type:       lotteryLog.GiftType,
			JumpURL:    s.c.WxLottery.JumpURLMap[strconv.FormatInt(lotteryLog.GiftType, 10)],
		}, nil
	}
	if actErr != nil {
		// 活动未开始或结束
		return &lottery.WxAwardRes{PlayWindow: _wxLotteryPlayWindowUnNeed}, nil
	}
	return &lottery.WxAwardRes{NotLottery: _notLottery, PlayWindow: s.checkLotteryPlayWindow(ctx, mid, buvid)}, nil
}

// checkLotteryPlayWindow 检查播放弹窗的状态，0未弹过，1弹过
func (s *Service) checkLotteryPlayWindow(ctx context.Context, mid int64, buvid string) int {
	if mid > 0 {
		if res, err := s.lottDao.GetLotteryPlayWindowMid(ctx, mid); err == nil && res {
			return _wxLotteryPlayWindowUnNeed
		}
		return _wxLotteryPlayWindowNeed
	}
	if res, err := s.lottDao.GetLotteryPlayWindowBuvid(ctx, buvid); err == nil && res {
		return _wxLotteryPlayWindowUnNeed
	}
	return _wxLotteryPlayWindowNeed
}

func (s *Service) WxLotteryPlayWindow(ctx context.Context, mid int64, buvid string) (*lottery.WxPlayWindowRes, error) {
	if err := s.checkTimeConf(ctx); err != nil {
		return nil, err
	}
	s.lottDao.AddLotteryPlayWindowBuvid(ctx, buvid, s.c.WxLottery.PlayWindowDuration)
	if mid > 0 {
		s.lottDao.AddLotteryPlayWindowMid(ctx, mid, s.c.WxLottery.PlayWindowDuration)
	}
	return &lottery.WxPlayWindowRes{}, nil
}

func (s *Service) WxLotteryAwardRedDot(ctx context.Context, mid int64) (bool, string, error) {
	// 不检查活动时间，查询中奖记录
	lotteryLog, err := s.lottDao.WxLotteryLog(ctx, mid)
	if err != nil {
		log.Errorc(ctx, "WxLotteryAwardRedDot mid(%d) error(%+v)", mid, err)
		return false, "", err
	}
	// 没有中奖的
	if lotteryLog == nil || lotteryLog.GiftID == 0 {
		return false, "", nil
	}
	// 中奖了，超过14天，不弹窗提示
	if (time.Now().Unix() - int64(lotteryLog.Ctime)) > 14*24*3600 {
		return false, "", nil
	}

	redDot, err := s.lottDao.CacheWxRedDot(ctx, mid)
	if err != nil {
		return false, "", err
	}
	var alertURL string
	if redDot {
		alertURL = s.c.WxLottery.AlertURL
		s.cache.Do(ctx, func(ctx context.Context) {
			s.lottDao.ExpireCacheWxRedDot(ctx, mid)
		})
	}
	return redDot, alertURL, err
}

func (s *Service) WxLotteryGifts(_ context.Context) (*lottery.WxLotteryGiftRes, error) {
	return s.wxLotteryGift, nil
}

func (s *Service) checkTimeConf(_ context.Context) error {
	now := time.Now()
	if now.Unix() < s.c.WxLottery.Stime.Unix() {
		return ecode.ActivityNotStart
	}
	if now.Unix() > s.c.WxLottery.Etime.Unix() {
		return ecode.ActivityOverEnd
	}
	return nil
}

func (s *Service) isAllowedByRisk(ctx context.Context, req *spapi.RiskInfoReq) bool {
	reply, err := s.silverClient.RiskInfo(ctx, req)
	if err != nil {
		return true
	}
	riskInfo, ok := reply.Infos[req.StrategyName[0]]
	if !ok {
		return true
	}
	if riskInfo.Level != 0 {
		return false
	}
	return true
}

// checkThreeNewUser 判断是否是三新用户
func (s *Service) checkThreeNewUser(ctx context.Context, mid int64, buvid string) (bool, error) {
	res, err := s.passportClient.CheckFreshUser(ctx, &passapi.CheckFreshUserReq{Mid: mid, Buvid: buvid, Period: s.c.WxLottery.BuvidPeriod})
	if err != nil {
		log.Errorc(ctx, "checkThreeNewUser s.passportClient.CheckFreshUser(%d) error(%v)", mid, err)
		return false, err
	}
	if res != nil {
		return res.IsNew, nil
	}
	return false, nil
}

func (s *Service) systemErr(err error) bool {
	return xecode.EqualError(xecode.ServerErr, err) || xecode.EqualError(xecode.Deadline, err)
}

//func (s *Service) loadWxLotteryGift() {
//	res, err := s.dao.SourceItem(context.Background(), s.c.WxLottery.Vid)
//	if err != nil {
//		log.Error("loadWxLotteryGift s.dao.SourceItem(%d) error(%v)", s.c.Eleven.ArcVid, err)
//		return
//	}
//	tmp := new(lottery.WxLotteryGiftRes)
//	if err = json.Unmarshal(res, &tmp); err != nil {
//		log.Error("loadWxLotteryGift json.Unmarshal(%s) error(%v)", res, err)
//		return
//	} else {
//		s.wxLotteryGift = tmp
//	}
//	log.Info("loadWxLotteryGift success")
//}
