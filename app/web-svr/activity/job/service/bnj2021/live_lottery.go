package bnj2021

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go-common/library/cache/redis"

	"go-gateway/app/web-svr/activity/job/component"
	bnjDao "go-gateway/app/web-svr/activity/job/dao/bnj"
	"go-gateway/app/web-svr/activity/job/model/bnj"
	"go-gateway/app/web-svr/activity/job/tool"

	"go-common/library/log"
	"go-common/library/queue/databus"
)

const (
	limitKeyOfGRPCLottery = "grpc_lottery"

	bizNameOfUserLottery4GRPC         = "bnj_user_lottery_grpc"
	bizNameOfUserLottery4UpdateStatus = "bnj_user_lottery_update_status"
	bizNameOfUpdateUserRewardInLive   = "bnj_update_user_reward_in_live"
	bizNameOfResetLastID              = "bnj_reset_user_lottery_lastID"

	cacheKey4BackupOfLiveAwardRec = "bnj2021_live_award_rec_%02d"

	metricKey4LiveAwardRec        = "bnj_live_award_rec"
	metricKey4LiveUserDraw        = "bnj_live_user_draw"
	metricKey4LiveUserDrawDefault = "bnj_live_user_draw_default"

	limitKey4LiveViewDurationDraw = "live_view_duration_draw"
)

var (
	liveDurationRuleM     map[int64]*bnj.LotteryRuleFor2021
	liveDurationRuleSyncM sync.Map

	UserLotteryReceiveConsumerCfg *databus.Config
)

func InitBnjLiveDrawRecConfig(cfg *databus.Config) {
	UserLotteryReceiveConsumerCfg = cfg
}

func ASyncLiveAwardRecFromBackupMQ(ctx context.Context) error {
	wg := new(sync.WaitGroup)

	for i := int64(0); i < 100; i++ {
		wg.Add(1)
		go func(index int64) {
			cacheKey := fmt.Sprintf(cacheKey4BackupOfLiveAwardRec, index)
			ticker := time.NewTicker(5 * time.Second)
			defer func() {
				wg.Done()
				ticker.Stop()
			}()

			for {
				select {
				case <-ticker.C:
					for {
						bs, err := redis.Bytes(component.BackUpMQ.Do(context.Background(), "RPOP", cacheKey))
						if err != nil {
							break
						}

						startTime := time.Now()
						_ = LiveRewardReceiveBiz(context.Background(), bs)
						tool.IncrBizCountAndLatency(srvName, metricKey4LiveAwardRec, startTime)
					}
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}
	return nil
}

func ASyncLiveRewardReceiveBiz(ctx context.Context) error {
	if UserLotteryReceiveConsumerCfg == nil || UserLotteryReceiveConsumerCfg.Topic == "" {
		return nil
	}

	for {
		canRestart := true
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
			}()

			consumer := databus.New(UserLotteryReceiveConsumerCfg)
			tool.ResetCommonGauge(metricKey4LiveAwardRec, 1)
			defer func() {
				tool.ResetCommonGauge(metricKey4LiveAwardRec, 0)
				_ = consumer.Close()
			}()

			for {
				select {
				case <-ctx.Done():
					canRestart = false

					return
				case msg, ok := <-consumer.Messages():
					if !ok {
						return
					}

					if isBizLimitReachedByBiz(metricKey4LiveAwardRec, metricKey4LiveAwardRec) {
						startTime := time.Now()
						_ = LiveRewardReceiveBiz(context.Background(), msg.Value)
						tool.IncrBizCountAndLatency(srvName, metricKey4LiveAwardRec, startTime)
					} else {
						go func(bs []byte) {
							startTime := time.Now()
							_ = LiveRewardReceiveBiz(context.Background(), bs)
							tool.IncrBizCountAndLatency(srvName, metricKey4LiveAwardRec, startTime)
						}(msg.Value)
					}
					_ = msg.Commit()
				}
			}
		}()

		wg.Wait()

		if !canRestart {
			return nil
		}

		// avoid frequently restart
		time.Sleep(time.Second * 5)
	}
}

func LiveRewardReceiveBiz(ctx context.Context, bs []byte) (err error) {
	reward := new(bnj.UserRewardInLiveRoom)
	if jsonErr := json.Unmarshal(bs, reward); jsonErr == nil {
		var affectRows int64

		switch reward.SceneID {
		case bnj.SceneID4ARDraw:
			affectRows, err = bnjDao.MarkBnjLiveUserCouponLotteryReceived(ctx, reward)
		case bnj.SceneID4LiveView:
			affectRows, err = bnjDao.MarkBnjLiveUserLotteryReceived(ctx, reward.MID, reward.Duration)
		case bnj.SceneID4Reserve:
			affectRows, err = bnjDao.UpdateReserveDrawRewardAsReceived(ctx, reward.MID)
		}

		if err == nil && affectRows > 0 {
			var uniqueID string
			for {
				uniqueID = genUniqueIDByUnixTime()
				if uniqueID != "" {
					break
				}

				time.Sleep(50 * time.Millisecond)
			}

			tool.IncrCommonBizStatus(bizNameOfUserLottery4UpdateStatus, tool.BizStatusOfSucceed)
			_ = PayDrawReward(
				reward.MID,
				reward.Reward.AwardId,
				businessOfPayReward,
				uniqueID)
		} else if err != nil {
			log.Error("LiveRewardReceiveBiz failed, err: %v, info: %v", err, string(bs))
			tool.IncrCommonBizStatus(bizNameOfUserLottery4UpdateStatus, tool.BizStatusOfFailed)
		}
	}

	return
}

func StartBnjLotteryBizByRules(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, v := range liveDurationRuleM {
				now := time.Now().Unix()
				if now < v.StartTime || now >= v.EndTime {
					continue
				}

				if _, ok := liveDurationRuleSyncM.Load(v.Duration); !ok {
					tmpRule := v.DeepCopy()
					go startBnjLotteryBiz(ctx, tmpRule)
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func startBnjLotteryBiz(ctx context.Context, rule *bnj.LotteryRuleFor2021) {
	startTime := time.Now()
	wg := new(sync.WaitGroup)
	liveDurationRuleSyncM.Store(rule.Duration, rule)
	defer func() {
		liveDurationRuleSyncM.Delete(rule.Duration)
	}()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		suffix := fmt.Sprintf("%02d", i%100)
		go func(s string) {
			startBnjLotteryBizBySuffix(ctx, wg, rule, s)
		}(suffix)
	}

	wg.Wait()
	bs, _ := json.Marshal(rule)
	logStr := fmt.Sprintf(
		"BnjLotteryBiz rule(%v) done, startAt: %v, endAt: %v, costs: %v",
		string(bs),
		startTime,
		time.Now(),
		time.Since(startTime))
	fmt.Println(logStr)
}

func startBnjLotteryBizBySuffix(ctx context.Context, wg *sync.WaitGroup,
	rule *bnj.LotteryRuleFor2021, suffix string) {
	ticker := time.NewTicker(time.Second)
	tool.IncrCommonGauge(metricKey4LiveUserDraw)
	defer func() {
		tool.DecCommonGauge(metricKey4LiveUserDraw)
		ticker.Stop()
		wg.Done()
	}()

	lastID, err := bnjDao.FetchLastReceiveIDByDurationAndSuffix(ctx, rule.Duration, suffix)
	if err != nil {
		fmt.Println("FetchLastReceiveIDByDurationAndSuffix err:", rule.Duration, suffix, err, time.Now())

		return
	}

	for {
		select {
		case <-ticker.C:
			if !isRuleEffective(rule) {
				return
			}

			lastID, _ = bnjLotteryByLastReceivedID(ctx, rule, suffix, lastID)
		case <-ctx.Done():
			return
		}
	}
}

func isRuleEffective(rule *bnj.LotteryRuleFor2021) (effective bool) {
	_, effective = liveDurationRuleM[rule.Duration]
	if effective {
		now := time.Now().Unix()
		if now < rule.StartTime || now >= rule.EndTime {
			effective = false
		}
	}

	return
}

func bnjLotteryByLastReceivedID(ctx context.Context, rule *bnj.LotteryRuleFor2021, suffix string, lastID int64) (
	newLastID int64, err error) {
	list := make([]*bnj.UserInLiveRoomFor2021, 0)
	list, err = bnjDao.FetchBnjUnReceivedUserList(
		context.Background(),
		suffix,
		lastID,
		rule.Duration,
		1000)
	if err != nil {
		return
	}

	for _, v := range list {
		if !isRuleEffective(rule) {
			return
		}

		var defaultAward bool
		if isBizLimitReachedByBiz(limitKey4LiveViewDurationDraw, limitKey4LiveViewDurationDraw) {
			defaultAward = true
		}

		startTime := time.Now()
		_ = drawInLiveRoom(ctx, v, rule, suffix, defaultAward)
		if defaultAward {
			tool.IncrBizCountAndLatency(srvName, metricKey4LiveUserDrawDefault, startTime)
		} else {
			tool.IncrBizCountAndLatency(srvName, metricKey4LiveUserDraw, startTime)
		}

		newLastID = v.ID
	}

	return
}

func drawInLiveRoom(ctx context.Context, user *bnj.UserInLiveRoomFor2021, rule *bnj.LotteryRuleFor2021,
	suffix string, defaultAward bool) (err error) {
	var (
		affectRows int64
		bizErr     error
	)

	rewardList := batchARDraw(user.MID, 1, drawType4LiveDuration, 82, defaultAward)
	if len(rewardList) == 0 {
		return
	}

	bs, _ := json.Marshal(rewardList[0])
	affectRows, bizErr = bnjDao.UpdateUserRewardInLive(ctx, user.MID, rule.Duration, string(bs))
	if bizErr == nil && affectRows > 0 {
		reward := new(bnj.UserRewardInLiveRoom)
		{
			reward.SceneID = bnj.SceneID4LiveView
			reward.MID = user.MID
			reward.Duration = rule.Duration
			reward.Reward = rewardList[0]
		}

		bizErr = bnjDao.RPushUserUnReceivedRewardInLiveDraw(ctx, reward)
	}

	if bizErr != nil {
		log.Errorc(
			ctx,
			"UpdateUserRewardInLive mid(%v) duration(%v) reward(%v), err(%v)",
			user.MID,
			rule.Duration,
			string(bs),
			bizErr)

		tool.IncrCommonBizStatus(bizNameOfUpdateUserRewardInLive, tool.BizStatusOfFailed)
	} else {
		tool.IncrCommonBizStatus(bizNameOfUpdateUserRewardInLive, tool.BizStatusOfSucceed)
	}

	if err := bnjDao.UpdateLastIDByDurationAndSuffix(ctx, rule.Duration, user.ID, suffix); err != nil {
		tool.IncrCommonBizStatus(bizNameOfResetLastID, tool.BizStatusOfFailed)
	}

	return
}

func genUniqueIDByUnixTime() (uniqueID string) {
	for {
		prefix := currentUnix - 1600000000
		suffix := atomic.AddInt64(&payAwardIndex, 1)
		if suffix > 9999999 {
			return
		}

		uniqueID = fmt.Sprintf("%v_%v", prefix, suffix)
		break
	}

	return
}
