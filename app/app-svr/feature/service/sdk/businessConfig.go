package sdk

import (
	"context"
	"fmt"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	featuregrpc "go-gateway/app/app-svr/feature/service/api"
	xmetric "go-gateway/app/app-svr/feature/service/model/metric"

	"google.golang.org/grpc"
)

func businessConfigKey(key string) string {
	return fmt.Sprintf("businessConfig.%v", key)
}

func (f *Feature) loadBusinessConfig() error {
	var args = &featuregrpc.BusinessConfigReq{TreeId: f.TreeID}
	tmp, err := f.featureClient.BusinessConfig(context.Background(), args)
	if err != nil {
		log.Error("featureSdkError businessConfig load cache err %v", err)
		xmetric.FeatureSdkError.Inc("businessConfig", "load_faild", "")
		return err
	}
	var cacheTmp = make(map[string]string)
	for keyname, businessConfig := range tmp.GetBusinessConfigs() {
		cacheTmp[keyname] = businessConfig.Config
	}
	f.businessConfigCache = cacheTmp
	return err
}

func (f *Feature) BusinessConfigHttp() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		f.formBusinessConfig(ctx)
	}
}

func (f *Feature) BusinessConfigGRPC() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return handler(f.formBusinessConfig(ctx), req)
	}
}

func (f *Feature) BusinessConfigManual(c context.Context, req *BuildLimitManual) context.Context {
	return f.formBusinessConfig(c)
}

func (f *Feature) formBusinessConfig(c context.Context) context.Context {
	for key, businessConfig := range f.businessConfigCache {
		if md, ok := metadata.FromContext(c); ok {
			md[businessConfigKey(key)] = businessConfig
		}
	}
	return c
}

func GetBusinessConfig(c context.Context, key string) string {
	var res = metadata.String(c, businessConfigKey(key))
	if res == "" {
		log.Error("featureSdkError businessConfig result wrong: key(%v), res(%+v)", key, res)
		xmetric.FeatureSdkError.Inc("businessConfig", "result_wrong", key)
	}
	return res
}

func (f *Feature) BusinessConfig(key string) string {
	for k, businessConfig := range f.businessConfigCache {
		if k == key {
			return businessConfig
		}
	}
	xmetric.FeatureSdkError.Inc("businessConfig", "result_wrong", key)
	return ""
}
