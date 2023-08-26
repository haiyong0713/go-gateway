package s10

import (
	"context"
	"sort"
	"time"

	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/model/s10"

	xtime "go-common/library/time"
)

func (s *Service) allGoodsProc() {
	for {
		s.allGoods()
		time.Sleep(3 * time.Second)
	}
}

func (s *Service) userLottery() {
	for {
		s.userLotteryInfo()
		time.Sleep(5 * time.Minute)
	}
}

func (s *Service) luckyUserMask(raw []*s10.UserLotteryInfo) []*s10.UserLotteryInfo {
	for _, v := range raw {
		tmp := []rune(v.Name)
		for i := range tmp {
			if i == 0 {
				continue
			}
			tmp[i] = '*'
		}
		v.Name = string(tmp)
	}
	return raw
}

func (s *Service) userLotteryInfo() {
	currTime := time.Now().Unix()
	ctx := context.Background()
	matchCfg := conf.LoadS10MatchesCfg()
	previous := int32(0)
	for _, v := range matchCfg.Matches {
		if currTime <= v.Lottery {
			break
		}
		previous = v.Robin
		_, ok := s.robinLotteryInfos.Load(v.Robin)
		if !ok {
			res, _ := s.dao.UsersLotteryInfoByRobin(ctx, v.Robin)
			if len(res) != 0 {
				res = s.luckyUserMask(res)
				s.robinLotteryInfos.Store(v.Robin, res)
			}
		}

	}
	res, _ := s.dao.UsersLotteryInfoByRobin(ctx, previous)
	if len(res) != 0 {
		res = s.luckyUserMask(res)
		s.robinLotteryInfos.Store(previous, res)
	}
}

func (s *Service) allGoods() {
	ctx := context.Background()
	currdate, err := s.timeToDate(time.Now().Unix())
	if err != nil {
		return
	}
	res, err := s.dao.AllGoods(ctx)
	if err != nil {
		return
	}

	roundGoods, err := s.dao.AllRobinGoods(ctx, xtime.Time(currdate))
	if err != nil {
		return
	}
	for robin, goodsList := range res {
		switch robin {
		case 0:
			res[0] = s.goodsSort(goodsList, roundGoods)
		default:
			sort.Slice(goodsList, func(i, j int) bool {
				return goodsList[i].Score > goodsList[j].Score
			})
		}
	}
	s.bonuses.Store(res)
	s.exchangeGoods(res, roundGoods, currdate)
}

func (s *Service) exchangeGoods(goodsMap map[int32][]*s10.Bonus, roundGoods map[int32]int32, currdate int64) {
	res := make(map[int32]*s10.Bonus, len(goodsMap)*5)
	for k, goods := range goodsMap {
		for _, v := range goods {
			if k != 0 {
				res[v.ID] = v
				continue
			}
			if v.RoundStock == 0 {
				v.IsRoundInfinite = true
			}
			if v.Stock == 0 {
				v.IsInfinite = true
				v.LeftTimes = 1<<31 - 1
			} else {
				if v.Send >= v.Stock {
					v.IsHaust = 1
				}
				v.LeftTimes = v.Stock - v.Send
			}
			if v.RoundExchangeTimes != 0 {
				v.IsRound = true
			}
			v.RoundSend = roundGoods[v.ID]
			v.CurrDate = currdate
			res[v.ID] = v
		}
	}
	s.goodsInfo.Store(res)
}

func (s *Service) goodsSort(goods []*s10.Bonus, robinGoods map[int32]int32) []*s10.Bonus {
	IsStockGoods := make([]*s10.Bonus, 0, len(goods))
	IsNotStockGoods := make([]*s10.Bonus, 0, len(goods))
	for _, v := range goods {
		v.RoundSend = robinGoods[v.ID]
		if v.Send >= v.Stock && v.Stock != 0 {
			IsNotStockGoods = append(IsNotStockGoods, v)
			continue
		}
		if v.RoundSend >= v.RoundStock && v.RoundStock != 0 {
			IsNotStockGoods = append(IsNotStockGoods, v)
			continue
		}
		IsStockGoods = append(IsStockGoods, v)
	}
	sort.Slice(IsStockGoods, func(i, j int) bool {
		return IsStockGoods[i].Rank < IsStockGoods[j].Rank ||
			(IsStockGoods[i].Rank == IsStockGoods[j].Rank && IsStockGoods[i].ID < IsStockGoods[j].ID)
	})
	sort.Slice(IsNotStockGoods, func(i, j int) bool {
		return IsNotStockGoods[i].Rank < IsNotStockGoods[j].Rank ||
			(IsNotStockGoods[i].Rank == IsNotStockGoods[j].Rank && IsNotStockGoods[i].ID < IsNotStockGoods[j].ID)
	})
	return append(IsStockGoods, IsNotStockGoods...)
}
