package tool

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	RpcBizOfActPlatform = "act_platform"

	bizOpOfDBBackSource = "db_back_source"
	bizOpOfDBErr        = "db_err"
	bizOpOfDBNoResult   = "db_no_result"

	StatusOfSucceed = "succeed"
	StatusOfFailed  = "failed"
	StatusOfHit     = "hit"
	StatusOfMiss    = "miss"
)

var (
	metric4Breaker            *prometheus.CounterVec
	metric4Limiter            *prometheus.CounterVec
	Metric4RpcQps             *prometheus.CounterVec
	Metric4RpcCount           *prometheus.CounterVec
	Metric4RpcLatency         *prometheus.CounterVec
	Metric4PubDatabus         *prometheus.CounterVec
	Metric4FreeFlowPubDatabus *prometheus.CounterVec

	metric4CacheReset  *prometheus.CounterVec
	metric4MemoryCache *prometheus.CounterVec

	metric4DBRestoreBiz  *prometheus.CounterVec
	metric4BizStatus     *prometheus.CounterVec
	Metric4RewardSuccess *prometheus.CounterVec
	Metric4RewardFail    *prometheus.CounterVec

	metric4AsyncReserveCost  *prometheus.GaugeVec
	metric4AsyncReserveCount *prometheus.CounterVec

	metric4LotterGiftSendCount        *prometheus.CounterVec
	metric4LotterGiftProbabilityCount *prometheus.CounterVec
	metric4LotterGiftProbabilityAll   *prometheus.CounterVec
	metric4LotterGiftProbability      *prometheus.GaugeVec
	metric4LotterDo                   *prometheus.CounterVec

	metric4UpActReserveRelationInfo4LiveCached   *prometheus.CounterVec
	metric4UpActReserveRelationInfo4LiveNoCached *prometheus.CounterVec
	metric4UpActReserveRelationInfo4LiveNoAuth   *prometheus.CounterVec
)

func init() {
	metric4BizStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity_interface",
			Name:      "common_biz_status",
			Help:      "activity interface common biz status stats",
		},
		[]string{"biz_name", "status"})

	metric4DBRestoreBiz = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "db_restore_count_stats",
			Help:      "activity db restore count stats",
		},
		[]string{"biz_op", "biz_name"})

	metric4MemoryCache = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "memory_cache_stats",
			Help:      "activity memory cache stats",
		},
		[]string{"biz_name", "status"})
	metric4CacheReset = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "cache_reset_stats",
			Help:      "activity cache reset stats",
		},
		[]string{"biz_name", "status"})

	metric4Breaker = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "rpc_breaker_stats",
			Help:      "activity rpc breaker calling stats",
		},
		[]string{"biz_name"})
	metric4Limiter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "biz_limiter_stats",
			Help:      "activity biz limiter calling stats",
		},
		[]string{"biz_name"})
	Metric4RpcQps = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "rpc_calling_stats",
			Help:      "activity rpc calling stats",
		},
		[]string{"biz_name", "path", "code"})
	Metric4RpcCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "rpc_req_count_stats",
			Help:      "activity rpc req count stats",
		},
		[]string{"biz_name", "path"})
	Metric4RpcLatency = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "rpc_req_latency_stats",
			Help:      "activity rpc req latency stats",
		},
		[]string{"biz_name", "path"})
	Metric4PubDatabus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "pub_databus_stats",
			Help:      "activity pub databus stats",
		}, []string{"biz_name"})
	Metric4FreeFlowPubDatabus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "pub_freeflow_databus_stats",
			Help:      "activity freeflow pub databus stats",
		}, []string{"biz_name"})
	Metric4RewardSuccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "reward_award_send_success_count",
			Help:      "activity rewards award send success count",
		},
		[]string{"biz_name"})
	Metric4RewardFail = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "webSvr_activity",
			Name:      "reward_award_send_fail_count",
			Help:      "activity rewards award send fail count",
		},
		[]string{"biz_name", "stage"})

	metric4AsyncReserveCost = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "webSvr_activity",
		Name:      "reserve_async_sync_delay",
		Help:      "activity async reserve do delay time",
	}, []string{"sid", "state"})
	metric4AsyncReserveCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "webSvr_activity",
		Name:      "reserve_async_sync_count",
		Help:      "activity async reserve count",
	}, []string{"sid", "state"})

	metric4LotterGiftSendCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "webSvr_activity",
		Name:      "lottery_gift_send_count",
		Help:      "activity lottery gift send count",
	}, []string{"sid", "gift_id"})
	metric4LotterGiftProbability = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "webSvr_activity",
		Name:      "lottery_gift_probability",
		Help:      "activity gift probability",
	}, []string{"sid", "gift_id"})
	metric4LotterGiftProbabilityCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "webSvr_activity",
		Name:      "lottery_gift_probability_count",
		Help:      "activity gift probability count",
	}, []string{"sid", "gift_id"})
	metric4LotterGiftProbabilityAll = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "webSvr_activity",
		Name:      "lottery_gift_probability_all",
		Help:      "activity gift probability all",
	}, []string{"sid", "gift_id"})
	metric4LotterDo = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "webSvr_activity",
		Name:      "lottery_do",
		Help:      "activity lottery do",
	}, []string{"sid", "stage"})
	metric4UpActReserveRelationInfo4LiveCached = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "webSvr_activity",
		Name:      "up_act_reserve_relation_info_live_cached",
		Help:      "up act reserve relation info live cached",
	}, []string{"mid", "sid"})
	metric4UpActReserveRelationInfo4LiveNoCached = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "webSvr_activity",
		Name:      "up_act_reserve_relation_info_live_no_cached",
		Help:      "up act reserve relation info live no cached",
	}, []string{"mid", "sid"})
	metric4UpActReserveRelationInfo4LiveNoAuth = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "webSvr_activity",
		Name:      "up_act_reserve_relation_info_live_no_auth",
		Help:      "up act reserve relation info live no auth",
	}, []string{"mid"})

	prometheus.MustRegister(
		metric4BizStatus,
		metric4DBRestoreBiz,
		metric4MemoryCache,
		metric4CacheReset,
		metric4Breaker,
		metric4Limiter,
		Metric4RpcQps,
		Metric4PubDatabus,
		Metric4RpcCount,
		Metric4RpcLatency,
		Metric4RewardSuccess,
		Metric4RewardFail,
		Metric4FreeFlowPubDatabus,
		metric4AsyncReserveCost,
		metric4AsyncReserveCount,
		metric4LotterGiftSendCount,
		metric4LotterGiftProbability,
		metric4LotterGiftProbabilityCount,
		metric4LotterGiftProbabilityAll,
		metric4LotterDo,
		metric4UpActReserveRelationInfo4LiveCached,
		metric4UpActReserveRelationInfo4LiveNoCached,
		metric4UpActReserveRelationInfo4LiveNoAuth,
	)
}

func IncrCommonBizStatus(bizName, status string) {
	metric4BizStatus.WithLabelValues([]string{bizName, status}...).Inc()
}

func IncrCacheResetMetric(bizName, status string) {
	metric4CacheReset.WithLabelValues([]string{bizName, status}...).Inc()
}

func IncrMemoryCacheHitOrMissMetric(bizName, status string) {
	metric4MemoryCache.WithLabelValues([]string{bizName, status}...).Inc()
}

func AddDBBackSourceMetrics(bizName string) {
	metric4DBRestoreBiz.WithLabelValues([]string{bizOpOfDBBackSource, bizName}...).Inc()
}

func AddDBErrMetrics(bizName string) {
	metric4DBRestoreBiz.WithLabelValues([]string{bizOpOfDBErr, bizName}...).Inc()
}

func AddDBNoResultMetrics(bizName string) {
	metric4DBRestoreBiz.WithLabelValues([]string{bizOpOfDBNoResult, bizName}...).Inc()
}

func IncrAsyncReserveCount(sid int64, state int32) {
	metric4AsyncReserveCount.WithLabelValues([]string{fmt.Sprint(sid), fmt.Sprint(state)}...).Inc()
}

func SetAsyncReserveDelay(sid int64, state int32, start int64) {
	delay := time.Now().UnixNano() - start
	if delay > 0 && start > 1e10 {
		metric4AsyncReserveCost.WithLabelValues([]string{fmt.Sprint(sid), fmt.Sprint(state)}...).Set(float64(delay) / 1e6)
	}
}

func IncrLotterySendGiftCount(sid, giftID int64, send int64) {
	metric4LotterGiftSendCount.WithLabelValues([]string{fmt.Sprint(sid), fmt.Sprint(giftID)}...).Add(float64(send))
}

func IncrLotteryDoCount(sid int64, stage string, nums int) {
	metric4LotterDo.WithLabelValues([]string{fmt.Sprint(sid), fmt.Sprint(stage)}...).Add(float64(nums))
}

func IncrLotterySendGiftProbability(sid, giftID int64, probability float64) {
	metric4LotterGiftProbability.WithLabelValues([]string{fmt.Sprint(sid), fmt.Sprint(giftID)}...).Set(probability)
	metric4LotterGiftProbabilityAll.WithLabelValues([]string{fmt.Sprint(sid), fmt.Sprint(giftID)}...).Add(probability)
	metric4LotterGiftProbabilityCount.WithLabelValues([]string{fmt.Sprint(sid), fmt.Sprint(giftID)}...).Inc()
}

func IncUpActReserveRelationInfo4LiveCached(sid int64) {
	metric4UpActReserveRelationInfo4LiveCached.WithLabelValues([]string{fmt.Sprint(sid)}...).Inc()
}

func IncUpActReserveRelationInfo4LiveNoCached(sid int64) {
	metric4UpActReserveRelationInfo4LiveNoCached.WithLabelValues([]string{fmt.Sprint(sid)}...).Inc()
}

func IncUpActReserveRelationInfo4LiveNoAuth(mid int64) {
	metric4UpActReserveRelationInfo4LiveNoAuth.WithLabelValues([]string{fmt.Sprint(mid)}...).Inc()
}
