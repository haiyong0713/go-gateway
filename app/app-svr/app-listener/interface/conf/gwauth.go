package conf

import (
	"context"
	"fmt"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// 简单的基于app key鉴权
// 用于提供给内部服务的API作为中间件使用
type AppAuth struct {
	AuthInfo map[string]*AppAuthInfo
}

type AppAuthInfo struct {
	AppName string
	AppKey  string
}

var DefaultAppAuthMetaKeys = []string{
	"x-bili-internal-gw-auth",
}

// 根据 grpc metadata中字段进行验证
// 默认取 "x-bili-internal-gw-auth" 的值
// 例如 x-bili-internal-gw-auth: "appName appKey"
func (ai *AppAuth) authGW(md metadata.MD, info *grpc.UnaryServerInfo, metaKey ...string) (bool, []string) {
	if ai == nil || ai.AuthInfo == nil {
		log.Warn("AppAuthGW: empty config. default allow passing for %q", info.FullMethod)
		// 无配置 默认允许
		return true, nil
	}
	if md == nil || len(md) == 0 {
		return false, nil
	}
	keys := DefaultAppAuthMetaKeys
	if len(metaKey) > 0 {
		keys = append(keys, metaKey...)
	}
	var appName, appKey string
	ret := make([]string, 0, len(keys))
	for _, k := range keys {
		if v := md.Get(k); len(v) > 0 {
			ret = append(ret, fmt.Sprintf("%s:(%s)", k, v[0]))
			idx := strings.Index(v[0], " ")
			if idx == -1 || idx+1 >= len(v[0]) {
				continue
			}
			appName = v[0][0:idx]
			appKey = v[0][idx+1:]
			if au, ok := ai.AuthInfo[appName]; ok && au != nil {
				if au.AppName == appName && au.AppKey == appKey {
					return true, ret
				}
			}
		}
	}
	return false, ret
}

// 简单的基于app key鉴权
// 用于提供给内部服务的API作为中间件使用
func (ai *AppAuth) UnaryServerInterceptor(metaKey ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, _ := metadata.FromIncomingContext(ctx)
		if ok, kvs := ai.authGW(md, info, metaKey...); !ok {
			if len(kvs) > 0 {
				log.Warnc(ctx, "AppAuthGW: rejected req with wrong keys: %v", kvs)
			} else {
				log.Warnc(ctx, "AppAuthGW: rejected req without required keys: %+v", md)
			}
			return nil, ecode.Unauthorized
		}
		return handler(ctx, req)
	}
}
