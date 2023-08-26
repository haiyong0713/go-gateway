package bml

import (
	"context"
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/interface/dao/bwsonline"
	"go-gateway/app/web-svr/activity/interface/model/bml"
	"go-gateway/app/web-svr/activity/interface/rewards"
	"go-gateway/app/web-svr/activity/interface/tool"
	"strings"
	"time"
)

func formatAnswer(old string) string {
	// 去掉空格
	old = strings.ReplaceAll(old, " ", "")
	// 去掉换行
	old = strings.ReplaceAll(old, "\n", "")
	// 全部转成小写字母
	return strings.ToLower(old)
}

func (s *Service) guessCommon(ctx context.Context, mid int64, answer string) (rewardConf *bml.RewardConf, err error) {
	ok := answer == formatAnswer(s.c.BMLGuessAct.CommonKey)
	log.Infoc(ctx, "start guess common compare user ok:%v , input:%v , right answer:%v ", ok, answer, formatAnswer(s.c.BMLGuessAct.CommonKey))
	if !ok {
		return nil, ecode.BMLGuessNotRightError
	}
	// 检查他有没有猜过
	var gList []*bml.GuessRecordItem
	if gList, err = s.MyGuessList(ctx, mid); err != nil {
		return
	}
	for _, v := range gList {
		if v.GuessType == bml.GuessTypeCommon || v.GuessType == bml.GuessTypeJoker {
			return nil, ecode.BMLRepeateGuessError
		}
	}
	return &bml.RewardConf{
		RewardId:      s.c.BMLGuessAct.Common30dayRewardId,
		RewardVersion: s.c.BMLGuessAct.Common30dayStockVersion,
		StockLimit:    s.c.BMLGuessAct.Common30dayStock,
	}, nil
}

func (s *Service) guessJoker(ctx context.Context, mid int64, answer string) (rewardConf *bml.RewardConf, err error) {
	log.Infoc(ctx, "start guess joker mid:%v , answer:%v", mid, answer)
	guessJkTimes, err2 := s.dao.IncrJokerKeyGuessTime(ctx, mid)
	if err2 == nil && guessJkTimes > s.c.BMLGuessAct.JokerKeyGuessMax && s.c.BMLGuessAct.JokerKeyGuessMax > 0 {
		log.Warnc(ctx, "BMLGuessJokerKeyOverLimitError guessJkTimes:%v , JokerKeyGuessMax:%v", guessJkTimes, s.c.BMLGuessAct.JokerKeyGuessMax)
		return nil, ecode.BMLGuessJokerKeyOverLimitError
	}
	md5Str := tool.MD5(answer)
	if !tool.InStrSlice(md5Str, s.c.BMLGuessAct.JokerKey) {
		return nil, ecode.BMLGuessNotRightError
	}
	if ts, err1 := s.dao.GetJokerKeyCache(ctx, answer); ts > 0 && err1 == nil {
		return nil, ecode.BMLGuessRepeateDrawError
	}

	var (
		preCheckFlag bool
		gList        []*bml.GuessRecordItem
	)
	if gList, err = s.MyGuessList(ctx, mid); err != nil {
		return
	}
	for _, v := range gList {
		if v.GuessType == bml.GuessTypeCommon {
			preCheckFlag = true
		}
		if v.GuessType == bml.GuessTypeJoker {
			return nil, ecode.BMLRepeateGuessError
		}
	}
	if !preCheckFlag {
		return nil, ecode.BMLCommonGuessNotcompleteError
	}
	return &bml.RewardConf{
		RewardId:      s.c.BMLGuessAct.JokerRewardId,
		RewardVersion: s.c.BMLGuessAct.JokerRewardStockVersion,
		StockLimit:    s.c.BMLGuessAct.JokerRewardStock,
	}, nil
}

func (s *Service) DoGuess(ctx context.Context, mid int64, GuessType int, answer string) (rewardConf *bml.RewardConf, err error) {
	nowTime := time.Now().Unix()
	if s.c.BMLGuessAct.ActBeginTime > nowTime {
		return nil, ecode.ActivityNotStart
	}
	if nowTime > s.c.BMLGuessAct.ActEndTime {
		return nil, ecode.ActivityOverEnd
	}

	if err = s.dao.FrequencyControl(ctx, mid, GuessType, s.c.BMLGuessAct.FrequencyControlTime); err != nil {
		return
	}
	// 频次控制
	switch GuessType {
	case bml.GuessTypeCommon:
		// 先看答案猜的对不对
		return s.guessCommon(ctx, mid, answer)
	case bml.GuessTypeJoker:
		return s.guessJoker(ctx, mid, answer)
	}
	return nil, ecode.SystemActivityParamsErr
}

func (s *Service) SendReward(ctx context.Context, mid int64, guessType int, answer string, rewardConf *bml.RewardConf) (gresult *bml.GuessResult, err error) {
	var uniqIds []string
	if rewardConf.RewardId <= 0 {
		return nil, ecode.SystemActivityParamsErr
	}
	// 0、预检查库存
	var stocksMap map[int64]int
	if stocksMap, err = s.bwsdao.GetGiftStocks(ctx, s.c.BMLGuessAct.ActName, []int64{rewardConf.RewardId}); err != nil {
		return
	}
	log.Infoc(ctx, "SendReward stocksMap:%v", stocksMap)
	if stocksMap[rewardConf.RewardId] <= 0 {
		return nil, ecode.BMLGuessRewardSendOutError
	}
	// 1、扣库存
	uniqIds, err = s.bwsdao.ConsumerStock(ctx, s.c.BMLGuessAct.ActName, rewardConf.RewardId, rewardConf.RewardVersion, 1)
	if len(uniqIds) < 1 {
		return nil, ecode.BMLGuessRewardSendOutError
	}
	record := &bml.GuessOrderRecord{
		Mid:         mid,
		RewardId:    rewardConf.RewardId,
		GuessType:   guessType,
		GuessAnswer: answer,
		OrderNo:     uniqIds[0],
		State:       bml.GuessOrderRecordStateComplete,
	}
	var serialNoMap map[string]int
	if serialNoMap, err = bwsonline.ParseSerialNo(uniqIds); err == nil && len(serialNoMap) > 0 {
		if serialNoMap[record.OrderNo] <= 0 || serialNoMap[record.OrderNo] > rewardConf.StockLimit {
			log.Errorc(ctx, "invalid stock serial no:%v , record.OrderNo:%v", serialNoMap, record.OrderNo)
			return nil, ecode.BMLGuessRewardSendOutError
		}
	}

	// 2、写中奖记录
	// GuessTypeCommon 答案是唯一的，为了满足唯一键，构造一个不重复的值
	if guessType == bml.GuessTypeCommon {
		record.GuessAnswer = fmt.Sprintf("%d-%d-%v", mid, guessType, answer)
		record.State = bml.GuessOrderRecordStateInit
	}
	var lastInsertId int64
	if lastInsertId, err = s.dao.AddGuessOrderRecord(ctx, record); err != nil || lastInsertId <= 0 {
		log.Infoc(ctx, "AddGuessOrderRecord lastInsertId:%v , failed:%+v", lastInsertId, err)
		return
	}
	if guessType == bml.GuessTypeJoker {
		if err2 := s.dao.CacheJokerKey(ctx, answer, 86400*30); err != nil {
			log.Errorc(ctx, "CacheJokerKey answer:%v , err:%v", answer, err2)
		}
		return &bml.GuessResult{
			GuessType: bml.GuessTypeJoker,
			IsRight:   true,
		}, err
	}

	_ = s.cache.SyncDo(ctx, func(ctx context.Context) {
		// 3、操作发放奖励
		for i := 0; i < 3; i++ {
			var info *api.RewardsSendAwardReply
			info, err = rewards.Client.SendAwardByIdAsync(ctx, mid, fmt.Sprintf("%v-%v", record.OrderNo, mid), s.c.BMLGuessAct.ActName, rewardConf.RewardId, true, true)
			if err == nil {
				break
			}
			log.Errorc(ctx, "SendAwardByIdAsync RewardsSendAwardReply:%v loop:%v  err:%+v", info, i, err)
		}
		if err != nil {
			return
		}
		// 4、更新到账记录
		var ef int64
		ef, err = s.dao.UpdateGuessOrderById(ctx, lastInsertId)
		log.Infoc(ctx, "UpdateGuessOrderById , lastInsertId:%v , effect_rows:%v , err:%v", lastInsertId, ef, err)
	})

	return &bml.GuessResult{
		GuessType: guessType,
		IsRight:   true,
	}, err
}

func (s *Service) CacheUserAnswerRecord(ctx context.Context, item *bml.GuessRecordItem, mid int64) (err error) {
	return s.dao.CacheUserAnswerRecord(ctx, item, mid, 3*(s.c.BMLGuessAct.ActEndTime-s.c.BMLGuessAct.ActBeginTime))
}

// ReservedList 获取我的预约列表
func (s *Service) MyGuessList(ctx context.Context, mid int64) (rList []*bml.GuessRecordItem, err error) {
	var guessOrderList []*bml.GuessOrderRecord
	if guessOrderList, err = s.dao.RawGuessOrderListByMid(ctx, mid); err != nil {
		return
	}

	if len(guessOrderList) <= 0 {
		var (
			item *bml.GuessRecordItem
			err2 error
		)
		item, err2 = s.dao.GetUserAnswerRecord(ctx, mid, bml.GuessTypeCommon)
		if err2 == nil && item != nil && item.GuessType > 0 {
			return []*bml.GuessRecordItem{item}, nil
		}
	}

	for _, v := range guessOrderList {
		rList = append(rList, &bml.GuessRecordItem{
			GuessType: v.GuessType,
			GuessTime: v.Ctime.Time().Unix(),
			RewardId:  v.RewardId,
		})
	}
	return
}
