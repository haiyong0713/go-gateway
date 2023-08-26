package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	coinapi "git.bilibili.co/bapis/bapis-go/community/service/coin"
	"go-common/library/log"
	lmdl "go-gateway/app/web-svr/activity/job/model/like"
	"go-main/app/account/usersuit/service/api"
)

const _coinAwardReason = "黄绿合战抽奖"

func (s *Service) awardproc() {
	defer s.waiter.Done()
	var (
		err error
	)
	if s.lotteryAwardSub == nil {
		return
	}
	for {
		msg, ok := <-s.lotteryAwardSub.Messages()
		if !ok {
			log.Info("databus:awardproc Lottery-Award-T exit!")
			return
		}
		msg.Commit()
		m := &lmdl.LotteryAward{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("awardproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		if m.Mid == 0 || m.Lid == 0 {
			continue
		}
		switch m.Lid {
		case s.c.YeGr.LotteryID:
			if _, ok := s.suitAwardIDs[m.AwardID]; ok {
				limitKey := fmt.Sprintf("yegl_suit_%d", m.Mid)
				if limitStr, err := s.dao.RsGet(context.Background(), limitKey); err != nil {
					log.Error("awardproc s.dao.RsGet error(%v)", err)
				} else {
					limit, _ := strconv.Atoi(limitStr)
					if limit > s.c.YeGr.SuitLimit {
						continue
					}
				}
				req := &api.GrantByMidsReq{
					Mids:   []int64{m.Mid},
					Pid:    s.c.YeGr.SuitID,
					Expire: s.c.YeGr.SuitExpire,
				}
				if _, err := s.suitClient.GrantByMids(context.Background(), req); err != nil {
					log.Error("awardproc s.suitClient.GrantByMids req(%+v) error(%v)", req, err)
				} else {
					// set expire 30 day
					s.dao.Incr(context.Background(), limitKey, 2592000)
				}
			} else if m.AwardID == s.c.YeGr.CoinOneAwardID {
				if _, err := s.coinClient.ModifyCoins(context.Background(), &coinapi.ModifyCoinsReq{Mid: m.Mid, Count: 1, Reason: _coinAwardReason}); err != nil {
					log.Error("awardproc s.coinClient.ModifyCoins mid(%d) error(%v)", m.Mid, err)
				}
			} else if m.AwardID == s.c.YeGr.CoinTwoAwardID {
				if _, err := s.coinClient.ModifyCoins(context.Background(), &coinapi.ModifyCoinsReq{Mid: m.Mid, Count: 2, Reason: _coinAwardReason}); err != nil {
					log.Error("awardproc s.coinClient.ModifyCoins mid(%d) error(%v)", m.Mid, err)
				}
			}
		}
		log.Info("awardproc success key:%s partition:%d offset:%d value:%s ", msg.Key, msg.Partition, msg.Offset, msg.Value)
	}
}

func (s *Service) actLikeLotteryProc(msg json.RawMessage) {
	var act = new(lmdl.Action)
	if err := json.Unmarshal(msg, act); err != nil {
		log.Error("yeGrLotteryProc json.Unmarshal(%s) error(%v)", msg, err)
		return
	}
	for _, sid := range s.c.Rule.LikeAddLotterySids {
		if act.Sid == sid && act.Mid > 0 {
			msg := &lmdl.LotteryMsg{MissionID: act.Sid, Mid: act.Mid, ObjID: act.Lid}
			s.lotteryActionch <- msg
			log.Info("actLikeLotteryProc add lotteryActionch(%+v)", msg)
			break
		}
	}
}
