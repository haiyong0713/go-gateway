package client

import (
	"context"
	"fmt"
	xtime "time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/tool"

	actPlatform "git.bilibili.co/bapis/bapis-go/platform/interface/act-plat-v2"

	"github.com/sony/gobreaker"
)

type actCallingFunc func(ctx context.Context, req interface{}) (interface{}, error)

const (
	bizKeyOfActPlatform             = "act_platform"
	Path4ActPlatformOfHistory       = "GetFormulaHistory"
	Path4ActPlatformOfCounterRes    = "GetCounterRes"
	Path4ActPlatformOfFormulaResult = "GetFormulaResult"

	codeStringOfBreaker = "cb_open"
)

type ActPlatActivityPointsReq struct {
	Points    int64  `json:"points"`
	Timestamp int64  `json:"timestamp"`
	Mid       int64  `json:"mid"`
	Source    int64  `json:"source"`
	Activity  string `json:"activity"`
	Business  string `json:"business"`
	Extra     string `json:"extra"`
}

func FetchActPlatformHistory(ctx context.Context, req interface{}) (interface{}, error) {
	return ActPlatClient.GetFormulaHistory(ctx, req.(*actPlatform.GetFormulaHistoryReq))
}

func FetchActPlatformCounterRes(ctx context.Context, req interface{}) (interface{}, error) {
	return ActPlatClient.GetCounterRes(ctx, req.(*actPlatform.GetCounterResReq))
}

func GetActPlatformCounterRes(ctx context.Context, mid, time int64, counter, act string) (int64, error) {
	req := &actPlatform.GetCounterResReq{
		Counter:  counter,
		Activity: act,
		Mid:      mid,
		Time:     time,
		Start:    nil,
	}
	reply, err := FetchResourceFromActPlatform(ctx, Path4ActPlatformOfCounterRes, FetchActPlatformCounterRes, req)
	if err != nil {
		log.Errorc(ctx, "GetCounterRes (mid:%d,time:%d,counter:%s,activity:%s) error:%v", mid, time, counter, act, err)
		return 0, err
	}
	var progress int64
	res := reply.(*actPlatform.GetCounterResResp)
	for _, v := range res.CounterList {
		if v.Val > 0 {
			progress += v.Val
		}
	}

	return progress, nil
}

func GetActPlatformCounterTotal(ctx context.Context, mid int64, counter, act string) (int64, error) {
	req := &actPlatform.GetTotalResReq{
		Counter:  counter,
		Activity: act,
		Mid:      mid,
	}
	reply, err := ActPlatClient.GetTotalRes(ctx, req)
	if err != nil {
		log.Errorc(ctx, "GetTotalRes (mid:%d,counter:%s,activity:%s) error:%v", mid, counter, act, err)
		return 0, err
	}

	return reply.Total, nil
}

func FetchResourceFromActPlatform(ctx context.Context, path string, udf actCallingFunc, req interface{}) (interface{}, error) {
	codeString := "0"
	defer func() {
		tool.Metric4RpcQps.WithLabelValues(
			[]string{
				tool.RpcBizOfActPlatform,
				path,
				codeString,
			}...).Inc()
	}()

	var srvErr error
	cbKey := fmt.Sprintf("%v_%v", bizKeyOfActPlatform, path)
	if cb, ok := tool.LoadCbByBizKey(cbKey); ok {
		result, cbErr := cb.Execute(func() (interface{}, error) {
			start := xtime.Now().UnixNano() / 1e6
			defer func() {
				end := xtime.Now().UnixNano() / 1e6
				latency := end - start
				tool.Metric4RpcCount.WithLabelValues(
					[]string{
						tool.RpcBizOfActPlatform,
						path,
					}...).Inc()
				tool.Metric4RpcLatency.WithLabelValues(
					[]string{
						tool.RpcBizOfActPlatform,
						path,
					}...).Add(float64(latency))
			}()

			resFromSvr, err := udf(ctx, req)
			srvErr = err
			if err != nil {
				code := xecode.Cause(err).Code()
				if code < 0 {
					code = -code
				}

				codeString = fmt.Sprintf("%d", code)
				if !tool.CanBreakerByCode(cbKey, codeString) {
					err = nil
				}
			}

			return resFromSvr, err
		})

		switch cbErr {
		case gobreaker.ErrOpenState, gobreaker.ErrTooManyRequests:
			codeString = codeStringOfBreaker
		}

		if cbErr != nil {
			return nil, cbErr
		}

		if srvErr != nil {
			return nil, srvErr
		}

		return result, nil
	}

	return udf(ctx, req)
}
