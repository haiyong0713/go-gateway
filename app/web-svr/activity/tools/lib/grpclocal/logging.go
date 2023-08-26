package grpclocal

import (
	"context"
	"fmt"
	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"time"
)

func ServerLogging(ctx context.Context, method, args string, action func() error) {
	startTime := time.Now()
	caller := "myself"
	zone := env.Zone
	var (
		buvid     string
		userAgent string
		remoteIP  string
		quota     float64
	)
	if d, ok := device.FromContext(ctx); ok {
		buvid = d.Buvid
		userAgent = d.UserAgent
	}
	if ip, ok := network.FromContext(ctx); ok {
		remoteIP = ip.RemoteIP
	}

	// call server handler
	err := action()

	// after server response
	code := ecode.Cause(err).Code()
	duration := time.Since(startTime)
	// monitor
	//_metricServerReqDur.Observe(int64(duration/time.Millisecond), info.FullMethod, caller)
	//_metricServerReqCodeTotal.Inc(info.FullMethod, caller, strconv.Itoa(code))

	mid := metadata.Int64(ctx, metadata.Mid)
	logFields := []log.D{
		log.KVString("user", caller),
		log.KVString("caller_zone", zone),
		log.KVString("ip", remoteIP),
		log.KVString("path", method),
		log.KVInt("ret", code),
		log.KVFloat64("ts", duration.Seconds()),
		log.KVFloat64("timeout_quota", quota),
		log.KVString("source", "grpc-access-log"),
		log.KVInt64("mid", mid),
		log.KVString("buvid", buvid),
		log.KVString("ua", userAgent),
	}
	// TODO: it will panic if someone remove String method from protobuf message struct that auto generate from protoc.
	logFields = append(logFields, log.KVString("args", args))
	if err != nil {
		logFields = append(logFields, log.KVString("error", err.Error()), log.KVString("stack", fmt.Sprintf("%+v", err)))
	}
	logFn(code, duration)(ctx, logFields...)
}

func logFn(code int, dt time.Duration) func(context.Context, ...log.D) {
	switch {
	case code < 0:
		return log.Errorv
	case dt >= time.Millisecond*500:
		// TODO: slowlog make it configurable.
		return log.Warnv
	case code > 0:
		return log.Warnv
	}
	return log.Infov
}
