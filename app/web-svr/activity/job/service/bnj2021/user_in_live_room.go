package bnj2021

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"go-common/library/log"

	bnjDao "go-gateway/app/web-svr/activity/job/dao/bnj"
	"go-gateway/app/web-svr/activity/job/model/bnj"
	"go-gateway/app/web-svr/activity/job/tool"

	"go-common/library/queue/databus"
)

const (
	bizNameOfUserInsertInLiveRoom = "bnj_live_duration_lottery_user_insert"
	metricKey4LiveUserSave        = "bnj_live_user_save"

	limitKey4LiveViewDuration = "live_view_duration"
)

var (
	UserInLiveRoomConsumerCfg *databus.Config
)

func InitBnjUserInLiveRoomConfig(cfg *databus.Config) {
	UserInLiveRoomConsumerCfg = cfg
}

func ASyncUserLogInLiveRoom(ctx context.Context) error {
	if UserInLiveRoomConsumerCfg == nil || UserInLiveRoomConsumerCfg.Topic == "" {
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

			consumer := databus.New(UserInLiveRoomConsumerCfg)
			tool.ResetCommonGauge(metricKey4LiveUserSave, 1)
			defer func() {
				tool.ResetCommonGauge(metricKey4LiveUserSave, 0)
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

					info := new(bnj.UserInLiveRoomFor2021)
					if err := json.Unmarshal(msg.Value, info); err == nil {
						if isBizLimitReachedByBiz(limitKey4LiveViewDuration, limitKey4LiveViewDuration) {
							tmpErr := bnjDao.InsertUnReceivedUserInfo(ctx, info)
							if tmpErr != nil {
								log.Error(
									"InsertUnReceivedUserInfo failed, err: %v, info: %v",
									tmpErr,
									string(msg.Value))
								tool.IncrCommonBizStatus(bizNameOfUserInsertInLiveRoom, tool.BizStatusOfFailed)
							} else {
								tool.IncrCommonBizStatus(bizNameOfUserInsertInLiveRoom, tool.BizStatusOfSucceed)
							}
						} else {
							go func(userInfo *bnj.UserInLiveRoomFor2021) {
								tmpErr := bnjDao.InsertUnReceivedUserInfo(ctx, userInfo)
								if tmpErr != nil {
									log.Error(
										"InsertUnReceivedUserInfo failed, err: %v, info: %v",
										tmpErr,
										string(msg.Value))
									tool.IncrCommonBizStatus(bizNameOfUserInsertInLiveRoom, tool.BizStatusOfFailed)
								} else {
									tool.IncrCommonBizStatus(bizNameOfUserInsertInLiveRoom, tool.BizStatusOfSucceed)
								}
							}(info.DeepCopy())
						}
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
