package s10

import (
	"context"
	"time"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/s10"

	"go-common/library/log"
)

func (s *Service) MatchesCategories(ctx context.Context, mid int64) (*s10.MatchCategories, error) {
	if !s.whiteCheck(mid) {
		return new(s10.MatchCategories), nil
	}
	matchesCfg := conf.LoadS10MatchesCfg()
	bonuses := s.bonuses.Load().(map[int32][]*s10.Bonus)
	currentTime := time.Now().Unix()
	res := &s10.MatchCategories{
		IsLogin:     mid > 0,
		CurrentTime: currentTime,
		Matches:     make([]*s10.MatchItem, 0, len(matchesCfg.Matches)),
	}
	otherActivity := new(s10.Other)
	if len(matchesCfg.OtherActivity) > 0 {
		otherActivity = matchesCfg.OtherActivity[currentTime%int64(len(matchesCfg.OtherActivity))]
	}
	nextvisual := 0
	for _, match := range matchesCfg.Matches {

		if match.Start <= currentTime && currentTime < match.End {
			res.Ongoing = match.Robin
		}
		userLotteryInfoRep, _ := s.robinLotteryInfos.Load(match.Robin)
		userLotteryInfo, _ := userLotteryInfoRep.([]*s10.UserLotteryInfo)
		if currentTime <= match.Lottery {
			userLotteryInfo = nil
		}
		tmp := &s10.MatchItem{
			IsLottery:        currentTime > match.Lottery,
			Match:            match,
			Other:            otherActivity,
			UserLotteryInfos: userLotteryInfo,
		}
		switch {
		case currentTime >= match.Start:
			tmp.Bonuses = bonuses[match.Robin]
			res.CurrentRobin = match.Robin
		case currentTime < match.Start && nextvisual == 0:
			tmp.Bonuses = bonuses[match.Robin]
			nextvisual += 1
		default:
			tmp.Bonuses = matchesCfg.DefaultBonuses
		}
		res.Matches = append(res.Matches, tmp)
	}
	if mid <= 0 || res.CurrentRobin == 0 {
		return res, nil
	}
	var err error
	res.Lottery, err = s.userLotteryByMatch(ctx, mid, currentTime, res.CurrentRobin)
	if err != nil {
		res.IsDegrade = true
	}
	return res, nil
}

func (s *Service) userLotteryByMatch(ctx context.Context, mid, currtime int64, robin int32) (map[int32]*s10.MatchUser, error) {
	matchCfg := conf.LoadS10MatchesCfg()
	expireTime := make(map[int32]int64, len(matchCfg.Matches))
	lotteryTime := make(map[int32]int64, len(matchCfg.Matches))
	for _, v := range matchCfg.Matches {
		expireTime[v.Robin] = v.LotteryExpire
		lotteryTime[v.Robin] = v.Lottery
	}
	if _, ok := expireTime[robin]; !ok {
		return make(map[int32]*s10.MatchUser, 1), nil
	}
	userLottery, err := s.dao.LotteryCache(ctx, mid)
	if err != nil {
		return nil, err
	}
	if len(userLottery) == 1 {
		for key := range userLottery {
			if key == s10.S10LotterySentinels {
				return nil, nil
			}
		}
	}
	_, ok := userLottery[robin]
	if userLottery == nil || !ok {
		userLottery, err = s.lotteryGoods(ctx, mid)
		if err != nil {
			return nil, err
		}
		resCache := make(map[int32]*s10.MatchUser, len(userLottery)+1)
		for k, v := range userLottery {
			tmp := new(s10.MatchUser)
			*tmp = *v
			if v.Lucky != nil {
				tmp.Lucky = new(s10.Lucky)
				*tmp.Lucky = *v.Lucky
			}

			resCache[k] = tmp
		}
		if len(resCache) == 0 {
			resCache[s10.S10LotterySentinels] = new(s10.MatchUser)
		}
		if err = cache.Do(context.Background(), func(ctx context.Context) {
			s.dao.AddLotteryCache(ctx, mid, resCache)
		}); err != nil {
			log.Errorc(ctx, "s.cache.Do() error(%v)", err)
		}
	}
	goodsMap := s.goodsInfo.Load().(map[int32]*s10.Bonus)
	for robin, v := range userLottery {
		v.ExpireTime = expireTime[robin]
		if v.Lucky == nil {
			continue
		}
		if currtime <= lotteryTime[robin] {
			v.Lucky = nil
			v.IsRecieve = false
			continue
		}
		if goods, ok := goodsMap[v.Lucky.Gid]; ok {
			v.Lucky.Name = goods.Name
			v.Lucky.Type = goods.Type
			v.Lucky.Figture = goods.Figure
			v.Lucky.Desc = goods.Desc
		}
	}
	return userLottery, err
}

func (s *Service) UserLotteryByRobin(ctx context.Context, mid int64, robin int32) (map[int32]*s10.MatchUser, error) {
	if !s.whiteCheck(mid) {
		return make(map[int32]*s10.MatchUser), nil
	}
	currtime := time.Now().Unix()
	if err := s.s10GoodsTimePeriod(currtime); err != nil {
		return nil, err
	}
	matchesCfg := conf.LoadS10MatchesCfg()
	flag := false
	for _, v := range matchesCfg.Matches {
		if v.Robin == robin && currtime > v.Lottery {
			flag = true
			break
		}
	}
	if !flag {
		return make(map[int32]*s10.MatchUser), nil
	}
	return s.userLotteryByMatch(ctx, mid, currtime, robin)
}

func (s *Service) lotteryGoods(ctx context.Context, mid int64) (map[int32]*s10.MatchUser, error) {
	matchesCfg := conf.LoadS10MatchesCfg()
	robins := make([]int64, 0, len(matchesCfg.Matches))
	currtime := time.Now().Unix()
	for _, v := range matchesCfg.Matches {
		if currtime < v.Start {
			continue
		}
		robins = append(robins, int64(v.Robin))
	}
	if len(robins) == 0 {
		return nil, nil
	}
	var (
		err         error
		lotteryInfo []int32
	)
	if s.splitTab {
		lotteryInfo, err = s.dao.UserLotteryInfoSub(ctx, mid, robins)
	} else {
		lotteryInfo, err = s.dao.UserLotteryInfo(ctx, mid, robins)
	}
	lotteryMap := make(map[int32]struct{}, len(lotteryInfo))
	for _, v := range lotteryInfo {
		lotteryMap[v] = struct{}{}
	}
	if err != nil {
		return nil, err
	}
	resMap, err := s.dao.UserLottery(ctx, mid)
	if err != nil {
		return nil, err
	}
	res := make(map[int32]*s10.MatchUser, len(lotteryInfo))
	for _, robin := range robins {
		tmpRobin := int32(robin)
		tmp, _ := resMap[tmpRobin]
		_, ok := lotteryMap[tmpRobin]
		robinLottery := &s10.MatchUser{IsLottery: ok || tmp != nil, Lucky: tmp, IsRecieve: tmp != nil && tmp.State > 0}
		res[tmpRobin] = robinLottery
	}
	return res, nil
}
