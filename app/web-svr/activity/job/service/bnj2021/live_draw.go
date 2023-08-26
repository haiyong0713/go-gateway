package bnj2021

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
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
	bizName4UpdateUserLiveCouponLotteryReward              = "bnj_user_coupon_reward_in_live"
	bizName4UpdateUserLiveCouponLotteryRewardOfBatchInsert = "bnj_user_coupon_reward_in_live_batch_insert"

	cacheKey4LiveARCouponIntoBackupMQ = "BNJ2021_live_AR_coupon_backup_%02d"

	metricKey4LiveARCoupon = "bnj_live_AR_coupon_draw"
)

var (
	UserLiveDrawCouponConsumerCfg *databus.Config
)

func InitBnjLiveDrawCouponConfig(cfg *databus.Config) {
	UserLiveDrawCouponConsumerCfg = cfg
}

func ASyncLiveARCouponFromBackupMQ(ctx context.Context) error {
	wg := new(sync.WaitGroup)

	for i := int64(0); i < 100; i++ {
		wg.Add(1)
		go func(index int64) {
			cacheKey := fmt.Sprintf(cacheKey4LiveARCouponIntoBackupMQ, index)
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
						_ = LiveARCouponBizByDrawCoupon(bs)
						tool.IncrBizCountAndLatency(srvName, metricKey4LiveARCoupon, startTime)
					}
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}
	return nil
}

func ASyncLiveARCouponBiz(ctx context.Context) error {
	if UserLiveDrawCouponConsumerCfg == nil || UserLiveDrawCouponConsumerCfg.Topic == "" {
		return nil
	}

	for {
		canRestart := true
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go func() {
			consumer := databus.New(UserLiveDrawCouponConsumerCfg)
			tool.ResetCommonGauge(metricKey4LiveARCoupon, 1)
			defer func() {
				tool.ResetCommonGauge(metricKey4LiveARCoupon, 0)
				_ = consumer.Close()
				wg.Done()
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

					startTime := time.Now()
					_ = LiveARCouponBizByDrawCoupon(msg.Value)
					tool.IncrBizCountAndLatency(srvName, metricKey4LiveARCoupon, startTime)
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

func ReIssueLiveARCouponByMidAndCoupon(mid, times int64) (err error) {
	coupon := new(bnj.UserARDrawCoupon)
	{
		coupon.MID = mid
		coupon.Coupon = times
	}

	list := genUserCouponLogByUserARDrawCoupon(coupon)
	err = bnjDao.BatchInsertLiveUserCouponLog(context.Background(), list)
	if err == nil {
		ARDrawAndUpdateReward(list)
	} else {
		log.Error("BatchInsertLiveUserCouponLog_reissue err: %v", err)
		tool.IncrCommonBizStatus(
			bizName4UpdateUserLiveCouponLotteryRewardOfBatchInsert,
			tool.BizStatusOfFailed)
	}

	return
}

func LiveARCouponBizByDrawCoupon(bs []byte) (err error) {
	coupon := new(bnj.UserARDrawCoupon)
	err = json.Unmarshal(bs, coupon)
	if err == nil {
		list := genUserCouponLogByUserARDrawCoupon(coupon)
		err = bnjDao.BatchInsertLiveUserCouponLog(context.Background(), list)
		if err == nil {
			ARDrawAndUpdateReward(list)
		} else {
			log.Error("BatchInsertLiveUserCouponLog err: %v", err)
			tool.IncrCommonBizStatus(
				bizName4UpdateUserLiveCouponLotteryRewardOfBatchInsert,
				tool.BizStatusOfFailed)
		}
	}

	return
}

func ARDrawAndUpdateReward(list []*bnj.UserCouponLogInLiveRoom) {
	for _, v := range list {
		rewardList := batchARDraw(v.MID, 1, drawType4LiveAR, 83, false)
		if len(rewardList) == 0 {
			continue
		}

		tmp := new(bnj.UserRewardInLiveRoom)
		{
			tmp.MID = v.MID
			tmp.No = v.No
			tmp.ReceiveUnix = v.ReceiveUnix
			tmp.Reward = rewardList[0]
			tmp.SceneID = bnj.SceneID4ARDraw
		}

		var (
			affectRows int64
			err        error
		)
		for i := 0; i < 3; i++ {
			affectRows, err = bnjDao.UpdateUserLiveCouponLotteryReward(context.Background(), tmp)
			if err == nil {
				break
			}
		}

		if err == nil && affectRows > 0 {
			err = bnjDao.RPushUserUnReceivedRewardInLiveDraw(context.Background(), tmp)
		}

		if err != nil {
			bs, _ := json.Marshal(tmp)
			log.Error("UpdateUserLiveCouponLotteryReward err: %v, info: %v", err, string(bs))
			tool.IncrCommonBizStatus(bizName4UpdateUserLiveCouponLotteryReward, tool.BizStatusOfFailed)
		}
	}
}

func genUserCouponLogByUserARDrawCoupon(coupon *bnj.UserARDrawCoupon) (list []*bnj.UserCouponLogInLiveRoom) {
	now := time.Now().UnixNano()
	list = make([]*bnj.UserCouponLogInLiveRoom, 0)
	for i := int64(1); i <= coupon.Coupon; i++ {
		tmp := new(bnj.UserCouponLogInLiveRoom)
		{
			tmp.MID = coupon.MID
			tmp.ReceiveUnix = now
			tmp.No = i
		}

		list = append(list, tmp)
	}

	return
}
