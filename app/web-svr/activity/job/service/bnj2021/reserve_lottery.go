package bnj2021

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/dao/bnj"
	bnjModel "go-gateway/app/web-svr/activity/job/model/bnj"
	"go-gateway/app/web-svr/activity/job/tool"

	"go-common/library/log"
)

const (
	bizNameOfResetReserveRewardLastID  = `bnj_reset_reserve_reward_lastID`
	bizNameOfReserveLotteryRPushInLive = `bnj_user_reserve_reward_in_live_rpush`
	bizNameOfReserveLotteryLogInsert   = `bnj_user_reserve_reward_log_insert`
	bizNameOfReserveRewardPay          = `bnj_user_reserve_reward_pay`

	ruleType4ReservedLottery        = 0
	ruleType4ReservedLiveLottery    = 2
	ruleType4ReservedLiveLotteryNew = 6
	metricKey4UserReserveBiz        = "bnj_user_reserve_biz"

	limitKey4ReserveLiveAward = "live_reserve_award"
)

var (
	currentUnix            int64
	payAwardIndex          int64
	bnj2021ReservedTotal   int64
	reserveRewardRuleM     map[int64]*bnjModel.ReserveRewardRuleFor2021
	reserveRewardRuleSyncM sync.Map
)

func init() {
	currentUnix = time.Now().Unix()
	reserveRewardRuleM = make(map[int64]*bnjModel.ReserveRewardRuleFor2021, 0)
	go aSyncUpdateCurrentUnix()
}

func aSyncUpdateCurrentUnix() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			currentUnix = time.Now().Unix()
			atomic.StoreInt64(&payAwardIndex, 0)
		}
	}
}

func ASyncReserveRewardRuleAndPay(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, v := range reserveRewardRuleM {
				if _, ok := reserveRewardRuleSyncM.Load(v.Count); !ok {
					if currentUnix >= v.EndTime {
						continue
					}

					go payRewardByRule(ctx, v.DeepCopy())
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func ASyncBnjReservedCount(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			UpdateBnjReservedCount(ctx)
		case <-ctx.Done():
			return nil
		}
	}
}

func UpdateBnjReservedCount(ctx context.Context) {
	if BnjRewardCfg.ReserveActivityID == 0 {
		return
	}

	req := new(api.GetReserveProgressReq)
	{
		rules := make([]*api.ReserveProgressRule, 0)
		rule := new(api.ReserveProgressRule)
		{
			rule.Dimension = 1
		}
		rules = append(rules, rule)

		req.Sid = BnjRewardCfg.ReserveActivityID
		req.Rules = rules
	}

	d, err := client.ActivityClient.GetReserveProgress(ctx, req)
	if err != nil || d == nil {
		return
	}

	for _, v := range d.Data {
		bnj2021ReservedTotal = v.Progress
	}
}

func payRewardByRule(ctx context.Context, rule *bnjModel.ReserveRewardRuleFor2021) {
	wg := new(sync.WaitGroup)
	startedAt := time.Now()
	reserveRewardRuleSyncM.Store(rule.Count, rule)
	tool.IncrCommonGauge(metricKey4UserReserveBiz)
	defer func() {
		tool.DecCommonGauge(metricKey4UserReserveBiz)
		reserveRewardRuleSyncM.Delete(rule.Count)
	}()

	wg.Add(1)
	suffix := fmt.Sprintf("%02d", BnjRewardCfg.ReserveActivityID%100)
	go func(s string) {
		defer func() {
			wg.Done()
		}()

		payRewardByRuleAndSuffix(ctx, rule, s)
	}(suffix)

	wg.Wait()

	bs, _ := json.Marshal(rule)
	log.Error("sendRewardByRule finished at %v, costs %v, rule(%v)", time.Now(), time.Since(startedAt), string(bs))
}

func payRewardByRuleAndSuffix(ctx context.Context, rule *bnjModel.ReserveRewardRuleFor2021,
	suffix string) {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			newRule, ok := reserveRewardRuleM[rule.Count]
			if !ok {
				return
			}

			if currentUnix < newRule.StartTime || bnj2021ReservedTotal < newRule.Count {
				break
			}

			if currentUnix >= newRule.EndTime {
				return
			}

			innerCtx := context.Background()
			lastID, err := bnj.FetchLastReceiveIDByCountAndSuffix(innerCtx, newRule.Count, suffix)
			if err != nil {
				break
			}

			if newRule.StartID != 0 && lastID == 0 {
				bs1, _ := json.Marshal(newRule)
				fmt.Println("payRewardByRuleAndSuffix_1: ", string(bs1), lastID)
				lastID = newRule.StartID
			}

			if !canSendByRuleAndRecord(newRule, lastID) {
				bs1, _ := json.Marshal(newRule)
				fmt.Println("payRewardByRuleAndSuffix_2: ", string(bs1), lastID)
				return
			}

			for {
				list, dbErr := bnj.FetchReservedMID(innerCtx, newRule.ActivityID, lastID, suffix)
				if dbErr != nil {
					fmt.Println("FetchReservedMID err: ", err)
					break
				}

				logKey := fmt.Sprintf("FetchReservedMID_%v", newRule.Count)
				if len(list) > 0 {
					lastID = list[len(list)-1].ID
					fmt.Println(logKey, list[0].ID, lastID, time.Now().Unix())
				} else {
					fmt.Println(logKey, " no enough records")
				}

				_ = reserveRewardPayBizByUserList(innerCtx, list, newRule, suffix)
				if !canSendByRuleAndRecord(newRule, lastID) {
					return
				}

				if len(list) < 1000 {
					break
				}
			}
		}
	}
}

func canSendByRuleAndRecord(rule *bnjModel.ReserveRewardRuleFor2021, lastID int64) (can bool) {
	if rule.EndID > 0 && lastID >= rule.EndID {
		return
	}

	can = true

	return
}

func reserveRewardPayBizByUserList(ctx context.Context, list []*bnjModel.ReservedUser,
	newRule *bnjModel.ReserveRewardRuleFor2021, suffix string) (err error) {
	switch newRule.Count {
	case ruleType4ReservedLiveLottery, ruleType4ReservedLiveLotteryNew:
		if len(list) == 0 {
			return
		}

		var (
			lastID  int64
			lastMid int64
			wg      sync.WaitGroup
		)
		for _, v := range list {
			lastID = v.ID
			lastMid = v.MID
			wg.Add(1)
			go func(tmp *bnjModel.ReservedUser) {
				defer func() {
					wg.Done()
				}()

				if !canSendByRuleAndRecord(newRule, tmp.ID) {
					return
				}

				waitBizLimit(limitKey4ReserveLiveAward, limitKey4ReserveLiveAward)
				startTime := time.Now()
				reserveRewardPayBiz(ctx, tmp, newRule, suffix, false)
				tool.IncrBizCountAndLatency(srvName, fmt.Sprintf("%v_%v", metricKey4UserReserveBiz, newRule.Count), startTime)
			}(v.DeepCopy())
		}
		wg.Wait()
		fmt.Println("ResetBnj2021ReserveRewardLastRecID_concurrency: ", lastID, lastMid, newRule.Count)
		if tmpErr := bnj.ResetBnj2021ReserveRewardLastRecID(ctx, lastID, newRule.Count, suffix); tmpErr != nil {
			fmt.Println("ResetBnj2021ReserveRewardLastRecID_concurrency_1: ", lastID, lastMid, newRule.Count, tmpErr)
			tool.IncrCommonBizStatus(bizNameOfResetReserveRewardLastID, tool.BizStatusOfFailed)
		}
	default:
		for _, v := range list {
			if !canSendByRuleAndRecord(newRule, v.ID) {
				return
			}

			startTime := time.Now()
			reserveRewardPayBiz(ctx, v, newRule, suffix, true)
			tool.IncrBizCountAndLatency(srvName, fmt.Sprintf("%v_%v", metricKey4UserReserveBiz, newRule.Count), startTime)
		}
	}

	return
}

func reserveRewardPayBiz(ctx context.Context, info *bnjModel.ReservedUser,
	newRule *bnjModel.ReserveRewardRuleFor2021, suffix string, updateLastID bool) {
	var (
		rewardStr string
		reward    *api.RewardsSendAwardReply
	)
	receivedStatus := bnjModel.RewardTypeOfNotReceived

	switch newRule.Count {
	case ruleType4ReservedLottery:
		// Do nothing now
	case ruleType4ReservedLiveLottery:
		rewardList := batchARDraw(info.MID, 1, drawType4Reserve, 81, false)
		if len(rewardList) == 0 {
			return
		}

		reward = rewardList[0]
		bs, _ := json.Marshal(reward)
		rewardStr = string(bs)
	case ruleType4ReservedLiveLotteryNew:
		rewardList := batchARDraw(info.MID, 1, drawType4Reserve, 81, false)
		if len(rewardList) == 0 {
			return
		}

		reward = rewardList[0]
		bs, _ := json.Marshal(reward)
		rewardStr = string(bs)
	default:
		rewardStr = strconv.FormatInt(newRule.RewardID, 10)
		receivedStatus = bnjModel.RewardTypeOfReceived
	}

	var upsertErr error
	if rewardStr != "" {
		tmpCount := newRule.Count
		if tmpCount == ruleType4ReservedLiveLotteryNew {
			tmpCount = ruleType4ReservedLiveLottery
		}

		for i := 0; i < 10; i++ {
			time.Sleep(time.Duration(50*i) * time.Millisecond)
			_, upsertErr = bnj.UpsertReserveRewardLog(
				ctx,
				info.MID,
				tmpCount,
				int64(receivedStatus),
				rewardStr)
			if upsertErr == nil {
				break
			}
		}
	}

	if upsertErr != nil {
		tool.IncrCommonBizStatus(bizNameOfReserveLotteryLogInsert, tool.BizStatusOfFailed)

		return
	}

	switch newRule.Count {
	case ruleType4ReservedLottery:
		req := new(api.BNJ2021ARCouponReq)
		{
			req.Mid = info.MID
			req.Coupon = 10
		}
		_, _ = client.ActivityClient.BNJARIncrCoupon(ctx, req)
	case ruleType4ReservedLiveLottery:
		unReceiveReward := new(bnjModel.UserRewardInLiveRoom)
		{
			unReceiveReward.Reward = reward
			unReceiveReward.SceneID = bnjModel.SceneID4Reserve
			unReceiveReward.MID = info.MID
		}

		cacheErr := bnj.RPushUserUnReceivedRewardInLiveDraw(ctx, unReceiveReward)
		if cacheErr != nil {
			tool.IncrCommonBizStatus(bizNameOfReserveLotteryRPushInLive, tool.BizStatusOfFailed)
		} else {
			_ = PayDrawReward(
				info.MID,
				newRule.RewardID,
				businessOfPayReward,
				fmt.Sprintf("rd_%v_%v", info.MID, newRule.Count))
		}
	case ruleType4ReservedLiveLotteryNew:
		unReceiveReward := new(bnjModel.UserRewardInLiveRoom)
		{
			unReceiveReward.Reward = reward
			unReceiveReward.SceneID = bnjModel.SceneID4Reserve
			unReceiveReward.MID = info.MID
		}

		cacheErr := bnj.RPushUserUnReceivedRewardInLiveDraw(ctx, unReceiveReward)
		if cacheErr != nil {
			tool.IncrCommonBizStatus(bizNameOfReserveLotteryRPushInLive, tool.BizStatusOfFailed)
		} else {
			_ = PayDrawReward(
				info.MID,
				newRule.RewardID,
				businessOfPayReward,
				fmt.Sprintf("rd_%v_%v", info.MID, ruleType4ReservedLiveLottery))
		}
	default:
		_ = PayDrawReward(
			info.MID,
			newRule.RewardID,
			businessOfPayReward,
			fmt.Sprintf("rd_%v_%v", info.MID, newRule.Count))
	}

	if !updateLastID {
		return
	}

	if tmpErr := bnj.ResetBnj2021ReserveRewardLastRecID(ctx, info.ID, newRule.Count, suffix); tmpErr != nil {
		fmt.Println("ResetBnj2021ReserveRewardLastRecID: ", info.ID, info.MID, newRule.Count)
		tool.IncrCommonBizStatus(bizNameOfResetReserveRewardLastID, tool.BizStatusOfFailed)
	}
}

func ReSendReserveLiveAward(ctx context.Context, mid, awardID int64) (err error) {
	rewardList := batchARDraw(mid, 1, drawType4Reserve, 81, false)
	if len(rewardList) == 0 {
		return
	}

	reward := rewardList[0]
	bs, _ := json.Marshal(reward)
	rewardStr := string(bs)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(50*i) * time.Millisecond)
		_, err = bnj.UpsertReserveRewardLog(
			ctx,
			mid,
			2,
			int64(bnjModel.RewardTypeOfNotReceived),
			rewardStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	unReceiveReward := new(bnjModel.UserRewardInLiveRoom)
	{
		unReceiveReward.Reward = reward
		unReceiveReward.SceneID = bnjModel.SceneID4Reserve
		unReceiveReward.MID = mid
	}

	err = bnj.RPushUserUnReceivedRewardInLiveDraw(ctx, unReceiveReward)
	if err == nil {
		_ = PayDrawReward(
			mid,
			awardID,
			businessOfPayReward,
			fmt.Sprintf("rd_%v_%v", mid, 2))
	}

	return
}
