package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-common/library/log"
	"go-common/library/net/trace"
	actapi "go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/conf"
	l "go-gateway/app/web-svr/activity/job/model/like"
	lmdl "go-gateway/app/web-svr/activity/job/model/like"
)

var lotteryWinCtx context.Context

func lotteryWinCtxInit() {
	lotteryWinCtx = trace.SimpleServerTrace(context.Background(), "lotteryWin")
}

const (
	_noLotteryLimit = 1
	_limitTypeHour  = 1
	_limitTypeDay   = 2
	_like           = 5
	_buy            = 6
)

func (s *Service) lotteryProc() {
	defer s.waiter.Done()
	var (
		ch = s.lotteryActionch
	)
	for {
		ms, ok := <-ch
		if !ok {
			log.Warn("lotteryProc s.lotteryActionch() quit")
			return
		}
		var (
			c        = context.Background()
			cfg      *conf.LotteryAddRule
			limitKey string
		)
		if cfg, ok = s.lotteryAdds[ms.MissionID]; !ok {
			continue
		}
		if cfg.NoLimit != _noLotteryLimit {
			// use mid and lottery id limit key
			limitKey = fmt.Sprintf("lott_ck_%d_%d", ms.Mid, cfg.LotteryID)
			switch cfg.LimitDuration {
			case _limitTypeHour: // hour
				limitKey += fmt.Sprintf("_%s", time.Now().Format("2006010215"))
			case _limitTypeDay: // day
				limitKey += fmt.Sprintf("_%s", time.Now().Format("20060102"))
			}
			if count, err := s.dao.RiGet(c, limitKey); err != nil {
				log.Error("lotteryProc s.dao.RsGet key(%s) error(%v)", limitKey, err)
				continue
			} else if count >= cfg.LimitTimes {
				log.Warn("lotteryProc limit time(%d) objID(%d) mid(%d) cfg(%+v)", count, ms.ObjID, ms.Mid, cfg)
				continue
			}
		}
		if ms.MissionID == s.c.Rule.DailyLikeSid { // 科学3分钟特殊活动
			if count, err := s.dao.StoryLikeSum(c, ms.MissionID, ms.Mid); err != nil {
				log.Error("lotteryProc science 3m s.dao.StoryLikeSum sid(%d) mid(%d) error(%v)", ms.MissionID, ms.Mid, err)
				continue
			} else if int(count) < cfg.LimitTimes {
				continue
			}
			checkKey := fmt.Sprintf("science_3m_%d_%d_%s", ms.Mid, ms.MissionID, time.Now().Format("20060102"))
			if check, err := s.dao.RsSetNX(context.Background(), checkKey, cfg.Expire); err != nil {
				log.Error("lotteryProc science 3m  s.dao.RsSetNX(%s) error(%+v)", checkKey, err)
				continue
			} else if !check {
				log.Warn("lotteryProc science 3m repeat (%s)", checkKey)
				continue
			}
		}
		if err := s.dao.AddLotteryTimes(c, cfg.LotteryID, ms.Mid); err != nil {
			log.Error("lotteryProc s.dao.AddLotteryTimes(%d,%d) error(%v)", cfg.LotteryID, ms.Mid, err)
			continue
		}
		if cfg.NoLimit != _noLotteryLimit {
			if _, err := s.dao.Incr(c, limitKey, cfg.Expire); err != nil {
				log.Error("lotteryProc s.dao.Incr(%s,%d) error(%v)", limitKey, cfg.Expire, err)
				continue
			}
		}
		log.Info("lotteryProc success sid(%d) mid(%d) objID(%d) lotteryID(%d)", ms.MissionID, ms.Mid, ms.ObjID, cfg.LotteryID)
	}
}

func (s *Service) vipLotteryproc() {
	defer s.waiter.Done()
	if s.vipLotterySub == nil {
		return
	}
	for {
		msg, ok := <-s.vipLotterySub.Messages()
		if !ok {
			log.Info("databus:vipLotteryproc VipLotteryTimes-T exit!")
			return
		}
		msg.Commit()
		m := &lmdl.VipLottery{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("vipLotteryproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		if m.Mid <= 0 || m.Ctime <= 0 {
			continue
		}
		log.Info("vipLotteryproc success mid(%d) act_token(%s) ctime(%d)", m.Mid, m.ActToken, m.Ctime)
	}
}

func (s *Service) ottVipLotteryproc() {
	defer s.waiter.Done()
	if s.ottVipLotterySub == nil {
		return
	}
	for {
		msg, ok := <-s.ottVipLotterySub.Messages()
		if !ok {
			log.Info("databus:ottVipLotteryproc TvipOrderSucc-T exit!")
			return
		}
		msg.Commit()
		m := &lmdl.OttVipLottery{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("ottVipLotteryproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		if m.Mid <= 0 || m.Ctime <= 0 {
			continue
		}
		lotteryID, ok := s.vipLotteryIDs[m.ActivityPlatformID]
		if !ok || lotteryID == 0 {
			log.Warn("ottVipLotteryproc not found appID(%v)", m)
			continue
		}
		checkKey := fmt.Sprintf("ottvip_a_t_%d_%s_%d", m.Mid, m.OrderNo, lotteryID)
		if check, err := s.dao.RsSetNX(context.Background(), checkKey, s.c.Rule.VipLotteryExpire); err != nil {
			log.Error("ottVipLotteryproc s.dao.RsSetNX(%s) error(%+v)", msg.Value, err)
			continue
		} else if !check {
			log.Warn("ottVipLotteryproc repeat (%s)", checkKey)
			continue
		}
		if err := s.retryAddLotteryTimes(context.Background(), lotteryID, m.Mid, _retryTimes); err != nil {
			log.Error("ottVipLotteryproc s.retryAddLotteryTimes(%d,%d) error(%v)", lotteryID, m.Mid, err)
			continue
		}
		log.Info("ottVipLotteryproc success data(%+v)", m)
	}
}

func (s *Service) customizeLotteryproc() {
	defer s.waiter.Done()
	if s.customizeLotterySub == nil {
		return
	}
	for {
		msg, ok := <-s.customizeLotterySub.Messages()
		if !ok {
			log.Info("databus:customizeLotteryproc TvipOrderSucc-T exit!")
			return
		}
		msg.Commit()
		m := &lmdl.CustomizeLottery{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("customizeLotteryproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		if m.Mid <= 0 || m.Ctime <= 0 {
			continue
		}
		var ltype int
		switch m.Type {
		case lmdl.CustomizeVip:
			ltype = lmdl.LotteryCustomizeType
		case lmdl.BuyVip:
			ltype = lmdl.LotteryVip
		case lmdl.Ogv:
			ltype = lmdl.LotteryOgvType
		default:
			log.Warn("customizeLotteryproc type(%s) undefined", m.Type)
		}
		s.goAddLotteryTimesByType(context.Background(), m.ActToken, m.Mid, ltype, m.OrderNo)
		log.Info("customizeLotteryproc success data(%+v)", m)
	}
}

func (s *Service) retryAddLotteryTimes(c context.Context, lotteryID, mid int64, retryCnt int) (err error) {
	for i := 0; i < retryCnt; i++ {
		if err = s.dao.AddLotteryTimes(c, lotteryID, mid); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return
}

// goAddLotteryTimes  .
func (s *Service) goAddLotteryTimesByType(c context.Context, sid string, mid int64, ltype int, orderNo string) (err error) {
	addTimes, ok1 := s.lotteryTypeAddTimes[ltype]
	if !ok1 {
		log.Errorc(c, "upAddLotteryTimes listmap not found type(%d)", ltype)
		return
	}
	mapAddTimes := make(map[string]*l.Lottery)
	for _, v := range addTimes {
		mapAddTimes[v.Info] = v
	}
	info, ok := mapAddTimes[sid]
	if !ok || info == nil {
		log.Infoc(c, "upAddLotteryTimes empty sid(%s) mid(%d) type(%d) ", sid, mid, ltype)
		return
	}
	if ok {
		if err = s.dao.GoAddLotteryTimes(c, info.Sid, info.ID, mid, ltype, orderNo); err != nil {
			log.Errorc(c, "s.dao.GoAddLotteryTimes(%v,%d,%d) error(%+v)", info, mid, ltype, err)
			return
		}
	}

	log.Infoc(c, "goAddLotteryTimes success s.dao.GoAddLotteryTimes(%s,%d)", sid, mid)
	return
}

// ClearLotteryWinList ...
func (s *Service) ClearLotteryWinList() {
	lotteryWinCtxInit()
	allLottery, err := s.dao.RawLotteryAllList(lotteryWinCtx)
	if err != nil {
		log.Errorc(lotteryWinCtx, " s.dao.RawLotteryAllList err(%v)", err)
		return
	}
	if allLottery != nil {
		for _, v := range allLottery {
			_, err := s.actGRPC.LotteryWinList(lotteryWinCtx, &actapi.LotteryWinListReq{
				Sid:       v.Sid,
				Num:       10,
				NeedCache: false,
			})
			if err != nil {
				log.Errorc(lotteryWinCtx, "s.actGRPC.LotteryWinList err(%v)", err)
			}
		}
	}
}
