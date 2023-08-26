package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	xmodel "go-gateway/app/web-svr/activity/interface/model/guess"
	"go-gateway/app/web-svr/activity/job/model/guess"
	"go-gateway/app/web-svr/activity/job/tool"
)

type CompensationRepair struct {
	S10ContestIDList []int64 `form:"s10ContestIDList,split" validate:"required,dive,gt=0"`
	TableSuffixList  []int64 `form:"tableSuffixList,split"`
	OpDB             bool    `form:"opDB"`
}

type CompensationRepair4MID struct {
	S10ContestIDList []int64 `form:"s10ContestIDList,split" validate:"required,dive,gt=0"`
	MID              int64   `form:"mid" validate:"required,gt=0"`
	OpDB             bool    `form:"opDB"`
}

type OverIssueRepair struct {
	Path string `form:"path"`
	OpDB bool   `form:"opDB"`
}

type GuessRepair4MID struct {
	MID  int64 `form:"mid" validate:"required,gt=0"`
	OpDB bool  `form:"opDB"`
}

type GuessRepair4MainID struct {
	MainID int64 `form:"main_id" validate:"required,gt=0"`
	OpDB   bool  `form:"opDB"`
}

func (s *Service) GuessMainRepair(repair *GuessRepair4MainID) (m map[string]interface{}, err error) {
	ctx := context.Background()
	tableSuffixList := genTableSuffixList([]int64{})
	midList := make([]int64, 0)
	m = make(map[string]interface{}, 0)

	for _, v := range tableSuffixList {
		tmpList, tmpErr := s.guessDao.UnClearedMIDListByTableSuffixAndMainID(ctx, v, repair.MainID)
		if tmpErr != nil {
			err = tmpErr

			return
		}

		for _, v := range tmpList {
			midList = append(midList, v)
		}
	}

	m["main_id"] = repair.MainID
	if !repair.OpDB {
		m["mid_list"] = midList
		m["mid_list_len"] = len(midList)

		return
	}

	failedMIDList := make([]int64, 0)
	for _, v := range midList {
		repair4Mid := new(GuessRepair4MID)
		{
			repair4Mid.OpDB = true
			repair4Mid.MID = v
		}

		if _, tmpErr := s.UserGuessRepair(repair4Mid); tmpErr != nil {
			failedMIDList = append(failedMIDList, v)
		}
	}

	{
		m["mid_list_of_failed"] = failedMIDList
	}

	return
}

// Only for coins stats
func (s *Service) UserGuessRepair(repair *GuessRepair4MID) (m map[string]interface{}, err error) {
	ctx := context.Background()
	m = make(map[string]interface{}, 0)
	clearedGuessList, err := s.guessDao.AllClearedGuessList(ctx)
	if err != nil {
		m["error"] = err.Error()
		m["error_step"] = 1

		return
	}

	clearedGuessM := genGuessMap(clearedGuessList)
	resultOddsM, err := s.calculateResultOdds(clearedGuessM)
	if err != nil {
		m["error"] = err.Error()
		m["error_step"] = 2

		return
	}

	unClearedGuessList, err := s.guessDao.AllUnClearedGuessListByMid(ctx, repair.MID)
	if err != nil {
		m["error"] = err.Error()
		m["error_step"] = 3

		return
	}

	if len(unClearedGuessList) > 0 {
		guessDetail := make(map[int64]float64, 0)

		for _, v := range unClearedGuessList {
			if _, ok := clearedGuessM[v.MainID]; ok {
				if odds, ok := resultOddsM[genResultOddsKey(v.MainID, v.DetailID)]; ok {
					coins := calculateAddCoins(v, odds)
					incomeCount := coins - float64(v.Stake)
					guessDetail[v.ID] = incomeCount
				} else {
					guessDetail[v.ID] = 0
				}
			}
		}

		m["repair_list"] = guessDetail
		if len(guessDetail) > 0 && repair.OpDB {
			err = s.guessDao.RepairUserGuessLogs(context.Background(), repair.MID, guessDetail)
			if err == nil {
				var totalGuess, totalSuccess, totalIncome int64
				totalGuess, totalSuccess, totalIncome, err := s.guessDao.FetchUserGuessStats(ctx, repair.MID)
				if err == nil {
					err = s.guessDao.UpdateUserGuessStats(ctx, repair.MID, totalGuess, totalSuccess, totalIncome, 1)
					if err == nil {
						if cacheErr := s.guessDao.DeleteUserGuessLogCache(ctx, repair.MID, 1); cacheErr != nil {
							fmt.Println(
								fmt.Sprintf(
									"UserGuessRepair delete user guess log cache failed, mid: %v, err: %v",
									repair.MID,
									cacheErr))
						}
					}
				}
			}
		}
	}

	return
}

func (s *Service) RepairSpecifiedUserGuess(repair *OverIssueRepair) error {
	if !tool.IsFileExists(repair.Path) {
		return errors.New("Specified file is not exists")
	}

	mainIDList := []int64{998, 999}
	list, err := s.guessDao.GuessMainList(context.Background(), mainIDList)
	if err != nil {
		return err
	}

	guessMap := genGuessMap(list)
	resultOddsM, err := s.calculateResultOdds(guessMap)
	if err != nil {
		return err
	}
	//resultOddsM := make(map[string]float64, 0)
	//{
	//  resultOddsM[genResultOddsKey(6666, 6666)] = 50
	//}

	go func() {
		csvFile, err := os.Open(repair.Path)
		if err != nil {
			return
		}

		nowTime := time.Now()
		notificationOfStart := fmt.Sprintf("收益bug修复开始 >>> %v", nowTime.Format("2006-01-02 15:04:05"))
		if bs, err := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, notificationOfStart, false); err == nil {
			_ = tool.SendCorpWeChatRobotAlarm(bs)
		}

		defer func() {
			csvFile.Close()

			notificationOfEnd := fmt.Sprintf("收益bug修复完成 >>> %v", time.Since(nowTime))
			if bs, err := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, notificationOfEnd, false); err == nil {
				_ = tool.SendCorpWeChatRobotAlarm(bs)
			}
		}()

		reader := csv.NewReader(csvFile)
		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				break
			}

			if len(record) != 12 {
				break
			}

			mid, _ := strconv.ParseInt(record[1], 10, 64)
			mainID, _ := strconv.ParseInt(record[2], 10, 64)

			if err = s.RepairOverIssueCoins(repair, mid, mainID, resultOddsM); err != nil {
				fmt.Println("RepairOverIssueCoins >>>", mid, mainID, err)
			}
		}
	}()

	time.Sleep(5 * time.Second)
	return nil
}

func (s *Service) RepairOverIssueCoins(repair *OverIssueRepair, mid, mainID int64, resultOddsMap map[string]float64) error {
	record, err := s.guessDao.UserGuessInfo(context.Background(), mid, mainID)
	if err != nil {
		return err
	}

	if d, ok := resultOddsMap[genResultOddsKey(mainID, record.DetailID)]; ok {
		coins := calculateAddCoins(record, d)
		incomeCount := coins - float64(record.Stake)
		diff := record.Income - int64(incomeCount*100)
		fmt.Println(
			fmt.Sprintf(
				"current_income: %v, new_income: %v, diff: %v",
				record.Income,
				incomeCount*100,
				diff))
		if diff > 0 && repair.OpDB {
			// update income user guess log
			err = s.guessDao.UpdateUserGuessLogCoins(context.Background(), record.ID, mid, incomeCount)
			if err != nil {
				return errors.New(fmt.Sprintf("RepairOverIssueCoinsFailed, UpdateUserLogCoins err: %v", err))
			}

			// incr -diff in user log
			err = s.guessDao.UpdateUserLogCoins(context.Background(), mid, -diff)
			if err != nil {
				return errors.New(fmt.Sprintf("RepairOverIssueCoinsFailed, UpdateUserLogCoins err: %v", err))
			}

			// delete user cache
			err = s.guessDao.DelUserCache(context.Background(), record.Mid, record.StakeType, 1, mainID)
			if err != nil {
				return errors.New(fmt.Sprintf("RepairOverIssueCoinsFailed, DelUserCache err: %v", err))
			}
		}
	}

	return nil
}

func (s *Service) CompensationRepair4MID(ctx context.Context, repair *CompensationRepair4MID) (m map[string]interface{}) {
	m = make(map[string]interface{}, 0)
	{
		m["s10ContestIDList"] = repair.S10ContestIDList
		m["mid"] = repair.MID
		m["OpDB"] = repair.OpDB
	}

	clearedGuessList, err := s.guessDao.AllClearedGuessList(ctx)
	if err != nil {
		m["error"] = err.Error()
		m["error_step"] = 1

		return
	}

	clearedGuessIDList := make([]int64, 0)
	for _, v := range clearedGuessList {
		clearedGuessIDList = append(clearedGuessIDList, v.ID)
	}

	m["clearedGuessList"] = clearedGuessIDList

	s10GuessList, err := s.guessDao.MainList(ctx, repair.S10ContestIDList)
	if err != nil {
		m["error"] = err.Error()
		m["error_step"] = 2

		return
	}

	s10GuessIDList := make([]int64, 0)
	for _, v := range s10GuessList {
		s10GuessIDList = append(s10GuessIDList, v.ID)
	}

	m["s10GuessIDList"] = s10GuessIDList

	clearedGuessM := genGuessMap(clearedGuessList)
	s10GuessM := genGuessMap(s10GuessList)
	resultOddsM, err := s.calculateResultOdds(clearedGuessM)
	if err != nil {
		m["error"] = err.Error()
		m["error_step"] = 3

		return
	}

	m["resultOddsM"] = resultOddsM

	list, err := s.guessDao.AllGuessListByMID(context.Background(), repair.MID)
	if err == nil {
		stats := s.RepairAndCompensation(repair.MID, clearedGuessM, s10GuessM, list, resultOddsM, repair.OpDB)
		for k, v := range stats {
			m[k] = v
		}
	} else {
		m["error"] = err.Error()
		m["error_step"] = 4
	}

	return
}

func (s *Service) CompensationRepair(ctx context.Context, repair *CompensationRepair) error {
	clearedGuessList, err := s.guessDao.AllClearedGuessList(ctx)
	if err != nil {
		return err
	}

	s10GuessList, err := s.guessDao.MainList(ctx, repair.S10ContestIDList)
	if err != nil {
		return err
	}

	clearedGuessM := genGuessMap(clearedGuessList)
	s10GuessM := genGuessMap(s10GuessList)
	tableSuffixList := genTableSuffixList(repair.TableSuffixList)

	resultOddsMap, err := s.calculateResultOdds(clearedGuessM)
	if err != nil {
		return err
	}

	go func() {
		nowTime := time.Now()
		notificationOfStart := fmt.Sprintf("补偿修复开始 >>> %v", nowTime.Format("2006-01-02 15:04:05"))
		if bs, err := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, notificationOfStart, false); err == nil {
			_ = tool.SendCorpWeChatRobotAlarm(bs)
		}

		defer func() {
			notificationOfEnd := fmt.Sprintf("补偿修复完成 >>> %v", time.Since(nowTime))
			if bs, err := tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfText, notificationOfEnd, false); err == nil {
				_ = tool.SendCorpWeChatRobotAlarm(bs)
			}
		}()

		for _, suffix := range tableSuffixList {
			if err := s.repairBySingleTable(context.Background(), suffix, clearedGuessM, s10GuessM, resultOddsMap, repair.OpDB); err != nil {
				fmt.Println("repairBySingleTableFailed", suffix, err)
			}
		}
	}()

	return nil
}

func genResultOddsKey(mainId, resultID int64) string {
	return fmt.Sprintf("%v_%v", mainId, resultID)
}

func (s *Service) calculateResultOdds(clearedGuessM map[int64]*xmodel.MainGuess) (m map[string]float64, err error) {
	m = make(map[string]float64)
	for _, v := range clearedGuessM {
		resultOdds, oddsErr := s.DetailOdds(v.ID, v.ResultID)
		if oddsErr != nil {
			err = oddsErr

			return
		}

		m[genResultOddsKey(v.ID, v.ResultID)] = resultOdds
	}

	return
}

func (s *Service) repairBySingleTable(ctx context.Context, suffix string, clearedGuessM,
	s10GuessM map[int64]*xmodel.MainGuess, resultOddsM map[string]float64, opDB bool) error {
	midList, err := s.guessDao.AllGuessedMidListByTableSuffix(ctx, suffix)
	if err != nil {
		return err
	}

	for _, v := range midList {
		if list, err := s.guessDao.AllGuessListByMID(context.Background(), v); err == nil {
			s.RepairAndCompensation(v, clearedGuessM, s10GuessM, list, resultOddsM, opDB)
		}
	}

	return nil
}

func (s *Service) RepairAndCompensation(mid int64, clearedGuessM, s10GuessM map[int64]*xmodel.MainGuess,
	guessList []*guess.GuessUser, resultOddsM map[string]float64, opDB bool) (m map[string]interface{}) {
	var addCoinsOfTotal, addCoinsOfS10 float64
	var winCount, loseCount, totalStake, totalIncome int64
	s10GuessDetail := make(map[int64]float64, 0)
	updateGuessMap := make(map[int64]float64, 0)
	m = make(map[string]interface{}, 0)

	for _, v := range guessList {
		totalIncome = totalIncome + v.Income
		totalStake = totalStake + v.Stake
		if _, ok := s10GuessM[v.MainID]; ok {
			if main, ok := clearedGuessM[v.MainID]; ok {
				s10GuessDetail[main.Oid] = 0
			}
		}

		if v.Status == 0 {
			if main, ok := clearedGuessM[v.MainID]; ok {
				if v.DetailID == main.ResultID {
					if _, ok := s10GuessM[v.MainID]; ok {
						winCount++
						if d, ok := resultOddsM[genResultOddsKey(v.MainID, v.DetailID)]; ok {
							coins := calculateAddCoins(v, d)
							s10GuessDetail[main.Oid] = coins
							addCoinsOfS10 = addCoinsOfS10 + coins
							addCoinsOfTotal = addCoinsOfTotal + coins

							updateGuessMap[v.ID] = coins

							fmt.Println(fmt.Sprintf("CompensationRepair, %v, %v, %v, %v, %v", mid, v.MainID, main.Oid, main.ResultID, coins))
						}
					}
				} else {
					if _, ok := s10GuessM[v.MainID]; ok {
						loseCount++
					}
				}
			}
		} else {
			if main, ok := clearedGuessM[v.MainID]; ok {
				if v.DetailID == main.ResultID {
					winCount++

					if d, ok := resultOddsM[genResultOddsKey(v.MainID, v.DetailID)]; ok {
						coins := calculateAddCoins(v, d)
						addCoinsOfTotal = addCoinsOfTotal + coins
					}

					if _, ok := s10GuessM[v.MainID]; ok {
						if d, ok := resultOddsM[genResultOddsKey(v.MainID, v.DetailID)]; ok {
							coins := calculateAddCoins(v, d)
							s10GuessDetail[main.Oid] = coins
							addCoinsOfS10 = addCoinsOfS10 + coins
						}
					}
				} else {
					loseCount++
				}
			}
		}
	}

	m["totalIncome"] = totalIncome
	m["totalStake"] = totalStake
	m["s10_guess_detail"] = s10GuessDetail
	m["s10_add_coins"] = addCoinsOfS10
	m["update_guess_map"] = updateGuessMap
	m["guess_count"] = len(guessList)
	m["win_count"] = winCount
	m["lose_count"] = loseCount
	m["add_coins_total"] = addCoinsOfTotal

	if opDB {
		if err := s.guessDao.BatchAddUserListCache(context.Background(), mid, s10GuessDetail); err != nil {
			bs, _ := json.Marshal(s10GuessDetail)
			fmt.Println("BatchAddUserListCacheFailed", mid, string(bs), err)
		}

		if err := s.guessDao.ResetUserCoinCache(context.Background(), mid, addCoinsOfS10); err != nil {
			fmt.Println("ResetUserCoinCacheFailed", mid, addCoinsOfS10, err)
		}

		if len(updateGuessMap) > 0 {
			if _, err := s.guessDao.UpUser(context.Background(), mid, updateGuessMap); err != nil {
				bs, _ := json.Marshal(updateGuessMap)
				fmt.Println("UpdateUserGuessFailed", mid, string(bs), err)
			}
		}

		if err := s.guessDao.ResetUserLog(
			context.Background(),
			mid,
			int64(len(guessList)),
			totalIncome,
			winCount,
			1); err != nil {
			fmt.Println("ResetUserLogFailed", mid, int64(len(guessList)), totalIncome, winCount)
		}
	}

	return m
}

func genGuessMap(list []*xmodel.MainGuess) map[int64]*xmodel.MainGuess {
	m := make(map[int64]*xmodel.MainGuess)
	for _, v := range list {
		tmp := new(xmodel.MainGuess)
		*tmp = *v
		m[tmp.ID] = tmp
	}

	return m
}

func genTableSuffix(mid int64) string {
	return fmt.Sprintf("%02d", mid%100)
}

func genTableSuffixList(suffixList []int64) []string {
	list := make([]string, 0)
	if len(suffixList) > 0 {
		for _, v := range suffixList {
			list = append(list, genTableSuffix(v))
		}
	} else {
		for i := 0; i < 100; i++ {
			list = append(list, genTableSuffix(int64(i)))
		}
	}

	return list
}
