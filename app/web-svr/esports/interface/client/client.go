package client

import (
	"context"
	"fmt"
	"time"

	xecode "go-common/library/ecode"
	favclient "go-main/app/community/favorite/service/api"

	"go-gateway/app/web-svr/esports/interface/conf"
	"go-gateway/app/web-svr/esports/interface/tool"

	"github.com/sony/gobreaker"

	"google.golang.org/grpc"
)

type rpcCallingFunc func(ctx context.Context, req interface{}, opts ...grpc.CallOption) (interface{}, error)

const (
	breakerErrorOfOpen                    = "cb_open"
	breakerErrorOfHalfOpen4TooManyRequest = "cb_half_open_too_many_req"
)

var (
	favClient favclient.FavoriteClient
)

func New(cfg *conf.Config) (err error) {
	favClient, err = favclient.New(cfg.FavClient)

	return
}

func innerRpcCalling(ctx context.Context, bizName, path string, udf rpcCallingFunc, req interface{}, opts ...grpc.CallOption) (interface{}, error) {
	var srvErr error
	bizCode := "0"
	defer func() {
		tool.Metric4RpcQps.WithLabelValues(
			[]string{bizName, path, bizCode}...).Inc()
	}()

	cbKey := fmt.Sprintf("%v_%v", bizName, path)
	if cb, ok := tool.LoadCbByBizKey(cbKey); ok {
		result, cbErr := cb.Execute(func() (interface{}, error) {
			start := time.Now().UnixNano() / 1e6
			defer func() {
				end := time.Now().UnixNano() / 1e6
				latency := end - start
				tool.Metric4RpcCount.WithLabelValues(
					[]string{bizName, path}...).Inc()
				tool.Metric4RpcLatency.WithLabelValues(
					[]string{bizName, path}...).Add(float64(latency))
			}()

			resFromSvr, err := udf(ctx, req, opts...)
			srvErr = err
			if err != nil {
				bizCode = convertSrvErr2String(err)
				if !tool.CanBreakByCode(cbKey, bizCode) {
					err = nil
				}
			}

			return resFromSvr, err
		})

		switch cbErr {
		case gobreaker.ErrOpenState:
			bizCode = breakerErrorOfOpen
		case gobreaker.ErrTooManyRequests:
			bizCode = breakerErrorOfHalfOpen4TooManyRequest
		}

		if cbErr != nil {
			return nil, cbErr
		}

		if srvErr != nil {
			return nil, srvErr
		}

		return result, nil
	}

	res, err := udf(ctx, req, opts...)
	if err != nil {
		bizCode = convertSrvErr2String(err)
	}

	return res, err
}

func convertSrvErr2String(err error) string {
	code := xecode.Cause(err).Code()
	if code < 0 {
		code = -code
	}

	return fmt.Sprintf("%d", code)
}
