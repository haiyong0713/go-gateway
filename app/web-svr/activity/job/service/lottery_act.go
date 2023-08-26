package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/log"
	pb "go-gateway/app/web-svr/activity/interface/api"
	l "go-gateway/app/web-svr/activity/job/model/like"
	"time"

	lmdl "go-gateway/app/web-svr/activity/job/model/like"
)

func (s *Service) actPlatHistoryMsgToCh() {
	defer func() {
		s.waiter.Done()
	}()
	if s.actPlatHistorySub == nil {
		return
	}
	for {
		msg, ok := <-s.actPlatHistorySub.Messages()
		if !ok {
			s.lotteryConsumewaiter.Done()
			log.Info("actPlatHistoryMsgToCh databus:actPlatHistorySub ActPlatHistory-T exit!")
			return
		}
		msg.Commit()
		m := &lmdl.ActPlatHistoryMsg{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("actPlatHistoryMsgToCh json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		log.Info("actPlatHistoryMsgToCh databus:actPlatHistorySub ActPlatHistory data(%+v)", m)

		if m.MID <= 0 || m.Diff <= 0 {
			continue
		}
		addTimes, ok1 := s.lotteryTypeAddTimes[lmdl.LotteryAct]
		if !ok1 {
			continue
		}
		mapAddTimes := make(map[string]*l.Lottery)
		for _, v := range addTimes {
			mapAddTimes[v.Info] = v
		}
		if _, ok2 := mapAddTimes[fmt.Sprintf("%s.%s", m.Activity, m.Counter)]; !ok2 {
			continue
		}
		s.actPlatHistoryCh <- m
		log.Info("actPlatHistoryMsgToCh success data(%+v)", m)
	}
}

func (s *Service) updateLotteryTimesLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.getLotteryTimes()
	}
}

func (s *Service) getLotteryTimes() {
	c := context.Background()
	res := make(map[int][]*l.Lottery)
	actTypeList := make([]int, 0)
	actTypeList = append(actTypeList, lmdl.LotteryArcType, lmdl.LotteryCustomizeType, lmdl.LotteryVip, lmdl.LotteryAct, lmdl.LotteryOgvType, lmdl.LotteryActPointType)
	// 获取有效抽奖
	allLottery, err := s.dao.RawLotteryAllList(c)
	if err != nil {
		log.Errorc(c, "s.dao.RawLotteryAllList error(%+v)", err)
		return
	}
	mapAllLottery := make(map[string]struct{})
	if allLottery != nil {
		for _, v := range allLottery {
			mapAllLottery[v.Sid] = struct{}{}
		}
	}
	for _, v := range actTypeList {
		// 查缓存获取抽奖活动id
		list, err := s.dao.RawLotteryAddTimesList(c, v)
		newList := make([]*l.Lottery, 0)
		if err != nil {
			log.Errorc(c, "s.dao.LotteryList error(%+v)", err)
			return
		}
		if list != nil {
			for _, times := range list {
				if _, ok := mapAllLottery[times.Sid]; ok {
					newList = append(newList, times)
				}
			}
		}

		res[v] = newList
	}
	s.lotteryTypeAddTimes = res
}

// LotteryAddTimes ...
func (s *Service) LotteryAddTimes(c context.Context, actType int) ([]*l.Lottery, error) {
	addTimes, ok1 := s.lotteryTypeAddTimes[actType]
	if ok1 {
		return addTimes, nil
	}
	return nil, nil
}

// ActTolotteryAddTimes ...
func (s *Service) ActTolotteryAddTimes() {
	defer func() {
		s.waiter.Done()
		close(s.actPlatHistoryWaitCh)
	}()

	c := context.Background()
	var err error
	for {
		m, ok := <-s.actPlatHistoryCh
		if !ok {
			break
		}
		addTimes, ok1 := s.lotteryTypeAddTimes[lmdl.LotteryAct]
		if !ok1 {
			continue
		}
		mapAddTimes := make(map[string]*l.Lottery)
		for _, v := range addTimes {
			mapAddTimes[v.Info] = v
		}
		if add, ok2 := mapAddTimes[fmt.Sprintf("%s.%s", m.Activity, m.Counter)]; ok2 {
			if m.Diff > 0 {
				for i := 0; i < int(m.Diff); i++ {
					index := i
					orderNo := fmt.Sprintf("%d_%d_%d_%d", m.MID, add.ID, m.TimeStamp, index)
					err = s.dao.LotteryAddTimesPub(c, m.MID, &lmdl.LotteryAddTimesMsg{
						MID:        m.MID,
						SID:        add.Sid,
						CID:        add.ID,
						ActionType: lmdl.LotteryAct,
						OrderNo:    orderNo,
					})
					if err != nil {
						log.Errorc(c, "ActTolotteryAddTimes err s.dao.LotteryAddTimesPub mid(%d) sid(%s) cid(%d) orderNo(%s) err(%v)", m.MID, add.Sid, add.ID, orderNo, err)
					} else {
						log.Infoc(c, "ActTolotteryAddTimes success s.dao.LotteryAddTimesPub mid(%d) sid(%s) cid(%d) orderNo(%s)", m.MID, add.Sid, add.ID, orderNo)
					}
				}
			}
		}
	}
}

func (s *Service) lotteryAddtimesMsgToCh() {
	defer func() {
		s.waiter.Done()
		close(s.lotteryAddtimesCh)
	}()
	if s.lotteryAddtimesSub == nil {
		return
	}
	for {
		msg, ok := <-s.lotteryAddtimesSub.Messages()
		if !ok {
			log.Info("databus:lotteryAddtimesMsgToCh VipLotteryTimes-T exit!")
			return
		}
		msg.Commit()
		m := &lmdl.LotteryAddTimesMsg{}
		if err := json.Unmarshal(msg.Value, m); err != nil {
			log.Error("lotteryAddtimesMsgToCh json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		s.lotteryAddtimesCh <- m
		log.Info("lotteryAddtimesMsgToCh success mid(%d)  sid(%s) cid(%d) orderNo(%s)", m.MID, m.SID, m.CID, m.OrderNo)
	}
}

// ActTolotteryAddTimes ...
func (s *Service) lotteryAddTimesConsume() {
	defer func() {
		s.waiter.Done()
		close(s.lotteryAddtimesWaitCh)
	}()
	c := context.Background()
	for {
		m, ok := <-s.lotteryAddtimesCh
		if !ok {
			break
		}
		_, err := s.actGRPC.LotteryAddTimes(c, &pb.LotteryAddTimesReq{
			Mid:        m.MID,
			Sid:        m.SID,
			ActionType: int64(m.ActionType),
			OrderNo:    m.OrderNo,
			Cid:        m.CID,
		})
		if err != nil {
			log.Errorc(c, "lotteryAddTimesConsume s.actGRPC.LotteryAddTimes err(%v) data(%+v)", err, m)
		} else {
			log.Infoc(c, "lotteryAddTimesConsume success s.actGRPC.LotteryAddTimes data(%+v)", m)
		}

	}
}
