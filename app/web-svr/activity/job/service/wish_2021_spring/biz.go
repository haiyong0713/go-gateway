package wish_2021_spring

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
)

const (
	bizNameOfUserCommit = "grpc_CommonActivityUserCommit"

	cacheKey4UserCommit = "activity:common:user_commit:mq:%v"
)

func ASynCommonActivityUserCommitConsumeFromBackupMQ(ctx context.Context) {
	wg := new(sync.WaitGroup)

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go func(index int) {
			cacheKey := fmt.Sprintf(cacheKey4UserCommit, index)
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

						req := new(api.CommonActivityUserCommitReq)
						if tmpErr := json.Unmarshal(bs, req); tmpErr == nil {
							_, rpcErr := client.ActivityClient.CommonActivityUserCommit(context.Background(), req)
							if rpcErr != nil {
								tool.IncrCommonBizStatus(bizNameOfUserCommit, tool.BizStatusOfFailed)
							}
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	wg.Wait()
}
