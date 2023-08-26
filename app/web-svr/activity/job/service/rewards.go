package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-common/library/conf/env"
	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/trace"
	"go-common/library/rate/limit/quota"
	"go-gateway/app/web-svr/activity/interface/api"
	"go-gateway/app/web-svr/activity/job/client"
	"go-gateway/app/web-svr/activity/job/component"
	"go-gateway/app/web-svr/activity/job/model/rewards"
	"go-gateway/app/web-svr/activity/job/tool"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"go-common/library/cache/redis"
)

const waiterNameFmt = "%s.%s.main.web-svr.activity-job|Rewards|%s|total"
const waiterForUnknownType = "Default"
const cacheKey4LiveAwardSendingBackupMQ = "REWARD_SENDING_BACKOFF_%v_%02d"

const sql4RetryForInitState = `
select
  mid,
  unique_id,
  award_name,
  award_id,
  award_type
from
  rewards_award_record_%02d
where
  state = 0
  and ctime < date_sub(now(), interval 2 hour)
order by
  ctime asc`

var (
	Metric4AwardSendingQueueLen = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "webSvr_activity_job",
			Name:      "reward_award_sending_queue",
			Help:      "activity job award sending queue",
		},
		[]string{"typ"})
	Metric4AwardSendingDelay = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "webSvr_activity_job",
			Name:      "reward_award_sending_delay",
			Help:      "activity job award sending delay time",
		},
		[]string{"typ"})
)
var rewardsSendingLimiterMap sync.Map

func init() {
	prometheus.MustRegister(Metric4AwardSendingQueueLen, Metric4AwardSendingDelay)
}

func getWaiterName(typ string) string {
	return fmt.Sprintf(waiterNameFmt, env.DeployEnv, env.Zone, typ)
}

func backoffKey4AwardSending(typ string, index int64) string {
	return fmt.Sprintf(cacheKey4LiveAwardSendingBackupMQ, typ, index)
}

func waitSendingLimiter(typ string) {
	iFace, ok := rewardsSendingLimiterMap.Load(typ)
	var waiter quota.Waiter
	if ok {
		waiter = iFace.(quota.Waiter)
	} else {
		waiter = quota.NewWaiter(&quota.WaiterConfig{
			ID: getWaiterName(typ),
		})
		if waiter.UnknowResource() {
			waiter = quota.NewWaiter(&quota.WaiterConfig{
				ID: getWaiterName(waiterForUnknownType),
			})
		}
		rewardsSendingLimiterMap.Store(typ, waiter)
	}
	waiter.Wait()
}

func (s *Service) ASyncRewardsAwardSending(ctx context.Context) {
	allTypes, err := client.ActivityClient.RewardsListAwardType(ctx, &api.RewardsListAwardTypeReq{})
	if err != nil {
		panic(err)
	}
	for _, typ := range allTypes.Types {
		tmpTyp := typ
		go s.innerASyncRewardsAwardSendingMonitor(ctx, tmpTyp)
		for i := 0; i < 20; i++ {
			tmpI := i
			go s.innerASyncRewardsAwardSendingWorker(ctx, tmpI, tmpTyp)
		}
	}
	for i := 0; i < 100; i++ {
		tmpI := i
		go s.innerRewardsAwardRetry(ctx, tmpI)
	}
}

func (s *Service) innerASyncRewardsAwardSendingMonitor(ctx context.Context, typ string) {
	log.Infoc(ctx, "innerASyncRewardsAwardSendingMonitor %v is starting...", typ)
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var (
				total int64
				err   error
			)

			for i := int64(0); i < 20; i++ {
				var tmpLen int64
				tmpLen, err = redis.Int64(component.BackUpMQ.Do(ctx, "LLEN", backoffKey4AwardSending(typ, i)))
				if err != nil {
					break
				}

				total = total + tmpLen
			}

			if err == nil {
				Metric4AwardSendingQueueLen.WithLabelValues([]string{typ}...).Set(float64(total))
			}
		}
	}
}

func (s *Service) innerASyncRewardsAwardSendingWorker(ctx context.Context, id int, typ string) {
	log.Infoc(ctx, "innerASyncRewardsAwardSendingWorker %v No %v is starting...", typ, id)
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				ctx := context.Background()
				bs, err := redis.Bytes(component.BackUpMQ.Do(ctx, "RPOP", backoffKey4AwardSending(typ, int64(id))))
				if err != nil {
					if err == redis.ErrNil {
						err = nil
					} else {
						log.Errorc(ctx, "LOG_ALERT: innerASyncRewardsAwardSendingWorker type: %v query redis error: %v", typ, err)
					}
					break
				}
				m := &rewards.AsyncSendingAwardInfo{}
				err = json.Unmarshal(bs, m)
				if err != nil {
					log.Errorc(ctx, "LOG_ALERT: innerASyncRewardsAwardSendingWorker typ: %v json.Unmarshal() error: %v, bs: %v", typ, err, bs)
					continue
				}
				waitSendingLimiter(typ)
				_, err = s.SendAward(m)
				Metric4AwardSendingDelay.WithLabelValues([]string{typ}...).Set(float64(m.SendTime))
			}
		}
	}
}

// innerRewardsAwardRetry: redis消息队列丢失情况处理
func (s *Service) innerRewardsAwardRetry(ctx context.Context, idx int) {
	log.Infoc(ctx, "innerRewardsAwardRetry %v is starting...", idx)
	ticker := time.NewTicker(15 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			//查询超过两个小时仍处于初始状态的奖励发放记录
			rows, err := component.GlobalRewardsDB.Query(ctx, fmt.Sprintf(sql4RetryForInitState, idx))
			if err == sql.ErrNoRows {
				err = nil
			}
			if err != nil {
				log.Errorc(ctx, "LOG_ALERT: innerRewardsAwardRetry query retry for init rows error: %v", err)
				continue
			}
			defer rows.Close()
			for rows.Next() {
				var typ string
				var name string
				m := &rewards.AsyncSendingAwardInfo{}
				if err := rows.Scan(&m.Mid, &m.UniqueId, &name, &m.AwardId, &typ); err != nil {
					log.Errorc(ctx, "LOG_ALERT: innerRewardsAwardRetry scan retry for init rows error: %v", err)
					continue
				}
				bs, err := json.Marshal(m)
				if err != nil {
					log.Errorc(ctx, "LOG_ALERT: json.Marshal error: %v", err)
					continue
				}
				log.Infoc(ctx, "innerRewardsAwardRetry retry for list: %v, name: %v, msg: %+v", backoffKey4AwardSending(typ, 1), name, m)
				for i := 0; i < 3; i++ {
					_, err = component.BackUpMQ.Do(ctx, "LPUSH", backoffKey4AwardSending(typ, 1), string(bs))
					if err == nil {
						break
					}
				}
			}
			if err = rows.Err(); err != nil {
				log.Errorc(ctx, "LOG_ALERT: innerRewardsAwardRetry query retry for init rows error: %v", err)
				continue
			}
		}
	}
}

func (s *Service) SendAward(m *rewards.AsyncSendingAwardInfo) (res *api.RewardsSendAwardReply, err error) {
	ctx := trace.NewContext(context.Background(), trace.New("SendAward"))
	req := &api.RewardsSendAwardReq{
		Mid:      m.Mid,
		UniqueId: m.UniqueId,
		Business: m.Business,
		AwardId:  m.AwardId,
		Sync:     true, //发奖的实际消费者必须使用同步发放. 否则会导致消息环(异步->异步->异步递归)
	}
	res, err = client.ActivityClient.RewardsSendAward(ctx, req)
	if ecode.Cause(err).Code() == 75971 { //奖励已发放
		err = nil
	}
	if err != nil {
		log.Errorc(ctx, "ASyncRewardsAwardSending call grpc failed, req(%+v), err(%v)", req, err)
		tool.IncrCommonBizStatus("ASyncRewardsAwardSending", tool.BizStatusOfFailed)
	}
	return
}
