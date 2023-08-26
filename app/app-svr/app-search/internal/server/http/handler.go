package http

import (
	"go-common/library/conf/env"
	bm "go-common/library/net/http/blademaster"
)

func newSearchDegradeHandler(args ...string) bm.HandlerFunc {
	if env.DeployEnv != env.DeployEnvProd {
		// 非生产环境不做降级处理
		return func(ctx *bm.Context) {}
	}
	return newSearchCacheSvr.Cache(deg.Args(args...), nil)
}

func hotSearchDegradeHandler(args ...string) bm.HandlerFunc {
	if env.DeployEnv != env.DeployEnvProd {
		// 非生产环境不做降级处理
		return func(ctx *bm.Context) {}
	}
	return hotSearchCacheSvr.Cache(deg.Args(args...), nil)
}
