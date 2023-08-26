package lol

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/model/lol"
)

var _emptyGuessList = make([]*lol.ContestDetail, 0)

const (
	_statusUnknown          = "unknown"
	_statusPred             = "failed"
	_statusPredSucc         = "succeed"
	_decimal        float64 = 100

	status4Settlement = "settlement"
)

const (
	settlementCode4Prepare = iota
	settlementCode4InProcess
	settlementCode4Cleared

	settlementStatus4Prepare   = "prepared"
	settlementStatus4InProcess = "in_process"
	settlementStatus4Cleared   = "cleared"
)

// UserWinCoinsV2 get user coins v2.
func (s *Service) UserWinCoinsV2(ctx context.Context, mid int64) (res *lol.PredictMsg, err error) {
	var userGuessList []*lol.UserGuessOid
	res = &lol.PredictMsg{
		Predicts:    0,
		PredictWins: 0,
		Coins:       0,
	}
	if userGuessList, err = s.userGuessOids(ctx, mid); err != nil {
		log.Errorc(ctx, "UserPredictListV2 s.userGuessOids mid(%d) error(%+v)", mid, err)
		err = ecode.ActGuessDataFail
		return
	}
	res.Predicts = len(userGuessList)
	if res.Predicts == 0 {
		return
	}
	for _, l := range userGuessList {
		if l.Income > 0 {
			income := l.Stake + (l.Income / _decimal)
			res.PredictWins++
			res.Coins = res.Coins + income
		}
	}
	if res.Coins > 0 {
		coinsWithNoDecimal, _ := strconv.ParseFloat(fmt.Sprintf("%.f", res.Coins), 64)
		if res.Coins != coinsWithNoDecimal {
			res.Coins, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", res.Coins), 64)
		} else {
			res.Coins = coinsWithNoDecimal
		}
	}
	return
}

// UserWinCoins get user coins.
func (s *Service) UserWinCoins(c context.Context, mid int64) (res *lol.PredictMsg, err error) {
	res = &lol.PredictMsg{
		Predicts:    0,
		PredictWins: 0,
		Coins:       0,
	}

	var lists []*lol.ContestDetail
	if ok, _ := s.dao.ExpireUserListCache(c, mid); !ok {
		return
	}
	if lists, err = s.dao.UserListCache(c, mid); err != nil {
		return
	}
	for _, l := range lists {
		if l.Coins > 0 {
			res.PredictWins++
			res.Coins = res.Coins + l.Coins
		}
		res.Predicts++
	}

	if res.Coins > 0 {
		coinsWithNoDecimal, _ := strconv.ParseFloat(fmt.Sprintf("%.f", res.Coins), 64)
		if res.Coins != coinsWithNoDecimal {
			res.Coins, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", res.Coins), 64)
		} else {
			res.Coins = coinsWithNoDecimal
		}
	}

	return
}

func (s *Service) userGuessOids(ctx context.Context, mid int64) (res []*lol.UserGuessOid, err error) {
	var (
		detailMap map[int64]string
		oidMap    map[int64]int64
		mainIDs   []int64
	)
	detailMap = make(map[int64]string, len(s.mainID)*2)
	oidMap = make(map[int64]int64, len(s.mainID))
	for mainID, details := range s.detailOptions {
		mainIDs = append(mainIDs, mainID)
		for _, detail := range details {
			detailMap[detail.DetailID] = detail.Option
			oidMap[mainID] = detail.Oid
		}
	}
	if len(mainIDs) == 0 {
		log.Infoc(ctx, "UserPredictListV2 userGuessOids mid(%d) ids is empty", mid)
		return
	}
	if res, err = s.dao.UserGuessOid(ctx, mid, mainIDs); err != nil {
		log.Errorc(ctx, "UserPredictListV2 userGuessOids s.dao.RawUserGuessOid mid(%d) mainIDs(%+v) error(%+v)", mid, mainIDs, err)
		err = ecode.ActGuessDataFail
		return
	}
	for _, resInfo := range res {
		option, ok := detailMap[resInfo.DetailID]
		if ok {
			resInfo.DetailName = option
		} else {
			log.Infoc(ctx, "UserPredictListV2 DetailMap mid(%d) resInfo.DetailID(%d) not ok", mid, resInfo.DetailID)
		}
		oidValue, okOid := oidMap[resInfo.MainID]
		if !okOid {
			log.Infoc(ctx, "UserPredictListV2 oidMap mid(%d) resInfo.DetailID(%d) not ok", mid, resInfo.DetailID)
		}
		resInfo.Oid = oidValue
	}
	return
}

func (s *Service) UserPredictListV2(ctx context.Context, mid int64, pn, ps int) (res []*lol.ContestDetail, total int, err error) {
	var (
		userGuessList []*lol.UserGuessOid
		start         = (pn - 1) * ps
		end           = start + ps - 1
		allRecord     []*lol.ContestDetail
	)
	res = _emptyGuessList
	if userGuessList, err = s.userGuessOids(ctx, mid); err != nil {
		log.Errorc(ctx, "UserPredictListV2 s.userGuessOids mid(%d) error(%+v)", mid, err)
		err = ecode.ActGuessDataFail
		return
	}
	total = len(userGuessList)
	if total == 0 || total < start {
		return
	}
	for _, l := range userGuessList {
		var income float64
		detail, ok := s.contestDetail[l.Oid]
		if !ok {
			continue
		}
		if l.Income > 0 {
			income = l.Stake + (l.Income / _decimal)
		}
		item := &lol.ContestDetail{
			ContestID: l.Oid,
			Timestamp: detail.Contest.StartTime,
			Title:     detail.Contest.Title,
			Home:      detail.Contest.Home,
			Away:      detail.Contest.Away,
			Status:    detail.Contest.Status,
			Coins:     income,
			Stake:     l.Stake,
			MyGuess:   l.DetailName,
		}
		switch {
		case l.Income > 0:
			item.Predict = _statusPredSucc
		case l.Income == 0 && detail.Contest.Status == "end" && l.Status == 1:
			item.Predict = _statusPred
		default:
			item.Predict = _statusUnknown
		}
		if (detail.Contest.Status == "end" || l.Income > 0) && l.Status == 1 {
			if item.Home.Wins > item.Away.Wins {
				item.Win = item.Home.Name
			} else {
				item.Win = item.Away.Name
			}
		}
		item.Settlement = settlementStatus4Cleared
		if d, ok := unSettlementContestIDList[item.ContestID]; ok {
			switch d {
			case settlementCode4Prepare:
				item.Settlement = settlementStatus4Prepare
			case settlementCode4InProcess:
				item.Settlement = settlementStatus4InProcess
			}
		}
		allRecord = append(allRecord, item)
	}
	if len(allRecord) == 0 {
		log.Infoc(ctx, "UserPredictListV2 mid(%d) allRecord empty", mid)
		return
	}
	sort.Slice(allRecord, func(i, j int) bool {
		return allRecord[i].Timestamp > allRecord[j].Timestamp
	})
	if total > end+1 {
		res = allRecord[start : end+1]
	} else {
		res = allRecord[start:]
	}
	return
}

// UserPredictList get user predict list.
func (s *Service) UserPredictList(c context.Context, mid int64) (res []*lol.ContestDetail, err error) {
	var lists []*lol.ContestDetail
	if ok, _ := s.dao.ExpireUserListCache(c, mid); !ok {
		return
	}
	if lists, err = s.dao.UserListCache(c, mid); err != nil || len(lists) == 0 {
		return
	}
	for _, l := range lists {
		detail, ok := s.contestDetail[l.ContestID]
		if !ok {
			continue
		}
		item := &lol.ContestDetail{
			ContestID: l.ContestID,
			Timestamp: detail.Contest.StartTime,
			Title:     detail.Contest.Title,
			Home:      detail.Contest.Home,
			Away:      detail.Contest.Away,
			Status:    detail.Contest.Status,
			Coins:     l.Coins,
		}
		switch {
		case l.Coins > 0:
			item.Predict = _statusPredSucc
		case l.Coins == 0 && detail.Contest.Status == "end":
			item.Predict = _statusPred
		default:
			item.Predict = _statusUnknown
		}
		if detail.Contest.Status == "end" || l.Coins > 0 {
			if item.Home.Wins > item.Away.Wins {
				item.Win = item.Home.Name
			} else {
				item.Win = item.Away.Name
			}
		}

		item.Settlement = settlementStatus4Cleared
		if d, ok := unSettlementContestIDList[item.ContestID]; ok {
			switch d {
			case settlementCode4Prepare:
				item.Settlement = settlementStatus4Prepare
			case settlementCode4InProcess:
				item.Settlement = settlementStatus4InProcess
			}
		}

		res = append(res, item)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].Timestamp > res[j].Timestamp
	})

	return
}

// HasUserPredict judge if user has predict those contests.
func (s *Service) HasUserPredict(c context.Context, mid int64, contestIDs []int64) (res map[int64]bool, err error) {
	var ok bool
	if ok, err = s.dao.ExpireUserListCache(c, mid); !ok {
		res = make(map[int64]bool)
		return
	}
	res, err = s.dao.ExistUserListCache(c, mid, contestIDs)
	return
}
