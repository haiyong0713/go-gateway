package service

import (
	"context"
	"encoding/json"
	"fmt"
	pb "go-gateway/app/web-svr/activity/interface/api"
	"sort"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/bws"
)

func (s *Service) CreateLotteryUsers() {
	// 先查有绑定的人
	var (
		id            int64
		users         []*bws.User
		achieves      map[int64]*bws.Achieve
		awards        []*bws.Award
		userAchieves  map[string][]*bws.UserAchieve
		userRank      []*bws.AchieveRank
		lotteryMidMap map[int64]struct{}
		rankMids      []int64
		rankMidMap    map[int64]int
		topMidMap     map[int64]struct{}
		err           error
	)
	c := context.Background()
	bid := s.c.Bws2019.Bid
	if bid == 0 || s.c.Bws2019.Limit == 0 {
		log.Error("createLotteryUsers conf(%v) error", s.c.Bws2019)
		return
	}
	if achieves, err = s.bws.Achievements(c, bid); err != nil {
		log.Error("createLotteryUsers s.bws.Achievements(%d) error(%v)", bid, err)
		return
	}
	if len(achieves) == 0 {
		log.Warn("createLotteryUsers len(achieves) == 0")
		return
	}
	log.Warn("createLotteryUsers achieve(%+v) data", achieves)
	if awards, err = s.bws.RechargeAward(c, bid); err != nil {
		log.Error("createLotteryUsers s.bws.RechargeAward bid(%d) error(%v)", bid, err)
		return
	} else if len(awards) == 0 {
		log.Error("createLotteryUsers no award")
		return
	}
	if rankMids, err = s.bws.AchieveRank(c, bid); err != nil {
		log.Error("createLotteryUsers s.bws.AchieveRank bid(%d) error(%v)", bid, err)
		return
	}
	if len(rankMids) == 0 {
		log.Error("createLotteryUsers rankMids len 0")
		return
	}
	rankMidMap = make(map[int64]int, len(rankMids))
	topMidMap = make(map[int64]struct{}, 3)
	for i, mid := range rankMids {
		if i < 3 {
			topMidMap[mid] = struct{}{}
		} else {
			rankMidMap[mid] = i + 1
		}
	}
	for {
		time.Sleep(500 * time.Millisecond)
		if users, err = s.bws.BindUsers(c, bid, id, s.c.Bws2019.Limit); err != nil {
			log.Error("createLotteryUsers s.bws.BindUsers bid(%d) id(%d) limit(%d) error(%v)", bid, id, s.c.Bws2019.Limit, err)
			return
		} else if len(users) == 0 {
			log.Warn("createLotteryUsers load data finish")
			break
		}
		keyMap := make(map[string]*bws.User)
		var keys []string
		for i, v := range users {
			keys = append(keys, v.Key)
			keyMap[v.Key] = v
			if i == len(users)-1 {
				id = v.ID
			}
		}
		if userAchieves, err = s.bws.UserAchieves(c, bid, keys); err != nil {
			log.Error("createLotteryUsers s.bws.UserAchieves bid(%d) keys(%v) error(%v)", bid, keys, err)
			return
		}
		for key, v := range userAchieves {
			if user, ok := keyMap[key]; !ok || user == nil {
				continue
			} else if _, isTop := topMidMap[user.Mid]; isTop {
				log.Warn("createLotteryUsers userRank skin user(%+v)", user)
				continue
			} else {
				var tmpPoint int64
				for _, userAchieve := range v {
					if achieve, ok := achieves[userAchieve.Aid]; ok {
						tmpPoint += achieve.AchievePoint
					}
				}
				userRank = append(userRank, &bws.AchieveRank{Mid: user.Mid, TotalPoint: tmpPoint})
			}
		}
	}
	sort.Slice(userRank, func(i, j int) bool {
		return userRank[i].TotalPoint > userRank[j].TotalPoint
	})
	var totalPoint int64
	for i, v := range userRank {
		if i < 19 {
			v.TotalPoint = v.TotalPoint * 50
		} else if i < 500 {
			v.TotalPoint = v.TotalPoint * 30
		} else if i < 1000 {
			v.TotalPoint = v.TotalPoint * 20
		}
		totalPoint += v.TotalPoint
	}
	log.Warn("createLotteryUsers total point (%d) len(userRank) %d", totalPoint, len(userRank))
	if totalPoint == 0 {
		log.Error("createLotteryUsers total point 0")
		return
	}
	// 按point点抽奖
	lotteryMidMap = make(map[int64]struct{})
	for _, v := range awards {
		if v.Amount <= 0 {
			log.Warn("createLotteryUsers awardID(%d) amount 0", v.ID)
			continue
		}
		var awardMids []*bws.LotteryUser
		for i := 0; i < v.Amount; i++ {
			offset := s.bwsLotteryRand.Int63n(totalPoint)
			var offsetPoint int64
			for i, ur := range userRank {
				offsetPoint += ur.TotalPoint
				if offsetPoint >= offset {
					midRank := -1
					if rank, ok := rankMidMap[ur.Mid]; ok {
						midRank = rank
					}
					awardMids = append(awardMids, &bws.LotteryUser{Bid: bid, Mid: ur.Mid, Rank: midRank})
					lotteryMidMap[ur.Mid] = struct{}{}
					if len(userRank) == i+1 {
						userRank = append(userRank[:i])
					} else {
						userRank = append(userRank[:i], userRank[i+1:]...)
					}
					totalPoint = totalPoint - ur.TotalPoint
					log.Warn("createLotteryUsers rand offsetPoint(%d) offset(%d) totalPoint(%d) len(userRank) (%d)", offsetPoint, offset, totalPoint, len(userRank))
					break
				}
			}
		}
		if err = s.bws.SetLotteryCache(c, s.c.Bws2019.Bid, v.ID, awardMids); err != nil {
			log.Error("createLotteryUsers s.bws.SetLotteryCache bid(%d) awardID(%d) awardMids(%v) error(%v)", s.c.Bws2019.Bid, v.ID, awardMids, err)
			return
		}
		log.Warn("createLotteryUsers save awardID(%d) mids(%v) success", v.ID, awardMids)
	}
	log.Warn("createLotteryUsers finish")
}

func (s *Service) CreateSpecLottery() {
	if len(s.c.Bws2019.SpecBids) == 0 {
		log.Warn("createLotteryUsers no spec bids")
		return
	}
	var allRankUsers []*bws.LotteryUser
	ctx := context.Background()
	for _, bid := range s.c.Bws2019.SpecBids {
		rankMids, err := s.bws.AchieveRank(ctx, bid)
		if err != nil {
			log.Error("createLotteryUsers s.bws.AchieveRank bid(%d) error(%v)", bid, err)
			return
		}
		if len(rankMids) == 0 {
			log.Error("createLotteryUsers rankMids(%d) len 0", bid)
			return
		}
		var oneRankUsers []*bws.LotteryUser
		for i, mid := range rankMids {
			oneRankUsers = append(oneRankUsers, &bws.LotteryUser{Bid: bid, Mid: mid, Rank: i + 1})
		}
		allRankUsers = append(allRankUsers, oneRankUsers...)
	}
	if len(allRankUsers) == 0 {
		return
	}
	allLen := len(allRankUsers)
	offset := s.bwsLotteryRand.Intn(allLen)
	user := allRankUsers[offset]
	err := s.bws.SetSpecLotteryCache(ctx, s.c.Bws2019.Bid, user)
	if err != nil {
		log.Error("createSpecLottery s.bws.SetSpecLotteryCache error(%v)", err)
		return
	}
	log.Warn("createSpecLottery user(%v) finish", user)
}

func (s *Service) bwsVipCardproc() {
	defer s.waiter.Done()
	var (
		err error
	)
	if s.vipCardSub == nil {
		return
	}
	for {
		msg, ok := <-s.vipCardSub.Messages()
		if !ok {
			log.Info("databus:bwsVipCardproc Lottery-Award-T exit!")
			return
		}
		msg.Commit()
		m := &bws.VipCard{}
		if err = json.Unmarshal(msg.Value, m); err != nil {
			log.Error("bwsVipCardproc json.Unmarshal(%s) error(%+v)", msg.Value, err)
			continue
		}
		var isBwsCard bool
		for _, v := range s.c.Bws2019.VipCardCodeIDs {
			if m.BatchCodeID == v {
				isBwsCard = true
				break
			}
		}
		if !isBwsCard {
			log.Warn("bwsVipCardproc vip card(%d) not bws", m.BatchCodeID)
			continue
		}
		if isBwsCard && m.UseTime >= s.c.Bws2019.Stime && m.UseTime <= s.c.Bws2019.Etime {
			if err = s.bws.AddAchieve(context.Background(), m.Mid, s.c.Bws2019.Bid); err != nil {
				log.Error("bwsVipCardproc s.bws.AddAchieve(%d,%d) error(%+v)", m.Mid, s.c.Bws2019.Bid, err)
				continue
			}
		}
		log.Info("bwsVipCardproc success key:%s partition:%d offset:%d value:%s ", msg.Key, msg.Partition, msg.Offset, string(msg.Value))
	}
}

func (s *Service) bwParkStockSync() {
	ctx := context.Background()
	nowTime := time.Now().Unix()
	var (
		reserveList *pb.BwParkBeginReserveResp
		stockResp   *pb.SyncGiftStockResp
		err         error
	)
	reserveList, err = s.actGRPC.BwParkBeginReserveList(ctx, &pb.BwParkBeginReserveReq{
		BeginTime: nowTime,
		EndTime:   nowTime,
	})
	if err != nil || reserveList == nil || reserveList.ReserveList == nil {
		log.Infoc(ctx, "bwParkStockSync BwParkBeginReserveList err:%+v , reserveList:%v", err, *reserveList)
	}

	for _, v := range reserveList.ReserveList {
		stockReq := &pb.GiftStockReq{
			SID:     fmt.Sprintf("%v", pb.ActInterReserveTicketType_StandardTicket2021),
			GiftID:  v.ID,
			GiftVer: v.Ctime.Time().Unix(),
			GiftNum: s.c.BwPark2021.SyncNum,
		}
		if v.ReserveBeginTime <= nowTime && v.ReserveEndTime > nowTime && v.StandardTicketNum > 0 {
			stockResp, err = s.actGRPC.SyncGiftStockInCache(ctx, stockReq)
			if stockResp != nil {
				log.Infoc(ctx, "bwParkStockSync SyncGiftStockInCache standard stock param:%v , Resp:%v , err:%+v", *stockReq, *stockResp, err)
			}
		}

		if v.VipReserveBeginTime <= nowTime && v.VipReserveEndTime > nowTime && v.VipTicketNum > 0 {
			stockReq.SID = fmt.Sprintf("%v", pb.ActInterReserveTicketType_VipTicket2021)
			stockResp, err = s.actGRPC.SyncGiftStockInCache(ctx, stockReq)
			if stockResp != nil {
				log.Infoc(ctx, "bwParkStockSync SyncGiftStockInCache vip stock param:%v , Resp:%v , err:%+v", *stockReq, *stockResp, err)
			}
		}
	}
	return
}

func (s *Service) bwBatchCacheTicketBind() {
	ctx := context.Background()
	var (
		cacheRecordId, recordId int64
		err                     error
	)
	if cacheRecordId, err = s.bws.GetMaxSyncBindRecordId(ctx, s.c.BwPark2021.Bid); err != nil {
		log.Errorc(ctx, "bwBatchCacheTicketBind GetMaxSyncBindRecordId  err:%+v", err)
		return
	}
	recordId = cacheRecordId
	if recordId <= 0 {
		recordId = 1
	}
	for i := 0; i < 50; i++ {
		var replay *pb.BatchCacheBindRecordsResp
		replay, err = s.actGRPC.BatchCacheBindRecords(ctx, &pb.BatchCacheBindRecordsReq{
			Limit:      s.c.BwPark2021.SyncNum,
			StartIndex: recordId,
		})

		log.Infoc(ctx, "bwBatchCacheTicketBind BatchCacheBindRecords replay_len:%v , err:%v", len(replay.RecordIds), err)

		if err != nil {
			log.Errorc(ctx, "bwBatchCacheTicketBind err:%+v", err)
			return
		}

		for _, v := range replay.RecordIds {
			if v > recordId {
				recordId = v
			}
		}

		if len(replay.RecordIds) < int(s.c.BwPark2021.SyncNum) {
			break
		}
	}
	if recordId != cacheRecordId {
		err = s.bws.SetMaxSyncBindRecordId(ctx, recordId, s.c.BwPark2021.Bid)
		log.Infoc(ctx, "bwBatchCacheTicketBind SetMaxSyncBindRecordId recordId:%v , cacheRecordId:%v, err:%v", recordId, cacheRecordId, err)
	}
}
