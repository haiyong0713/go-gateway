package bnj2021

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go-common/library/cache/redis"

	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/tool"

	"go-common/library/log"
	"go-common/library/queue/databus"
)

const (
	bizNameOfARExchangeGRPC = "bnj_AR_coupon_exchange_grpc"
	metricKey4ARExchange    = "bnj_AR_coupon_exchange"

	cacheKey4BackupOfARExchange = "bnj2021_AR_exchange_%02d"
)

func InitARRewardConsumerCfg(cfg *databus.Config) {
	ARRewardConsumerCfg = cfg
}

func ASyncARExchangeFromBackupMQ(ctx context.Context) error {
	wg := new(sync.WaitGroup)

	for i := int64(0); i < 100; i++ {
		wg.Add(1)
		go func(index int64) {
			cacheKey := fmt.Sprintf(cacheKey4BackupOfARExchange, index)
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

						arExchange(bs)
					}
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}

func ASyncARRewardConsumer(ctx context.Context) error {
	if ARRewardConsumerCfg == nil || ARRewardConsumerCfg.Topic == "" {
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

			consumer := databus.New(ARRewardConsumerCfg)
			tool.ResetCommonGauge(metricKey4ARExchange, 1)
			defer func() {
				tool.ResetCommonGauge(metricKey4ARExchange, 0)
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

					arExchange(msg.Value)
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

func arExchange(data []byte) {
	startTime := time.Now()
	req := new(api.BNJ2021ARExchangeReq)
	if err := json.Unmarshal(data, req); err == nil {
		if _, err := client.ActivityClient.BNJARExchange(context.Background(), req); err != nil {
			bs, _ := json.Marshal(req)
			log.Error("BNJ_AR exchange failed, req(%v), err(%v)", string(bs), err)
			tool.IncrCommonBizStatus(bizNameOfARExchangeGRPC, tool.BizStatusOfFailed)
		} else {
			tool.IncrBizCountAndLatency(srvName, metricKey4ARExchange, startTime)
		}
	}
}
