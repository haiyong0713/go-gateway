package sdk

import (
	"context"
	"fmt"
	"hash/crc32"
	"strconv"
	"strings"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	featuregrpc "go-gateway/app/app-svr/feature/service/api"

	"google.golang.org/grpc"
)

const _default = "default"

type AbTestManual struct {
	Mid    int64
	Buvid  string
	Custom map[string]string
}

func abtestKey(key string) string {
	return fmt.Sprintf("abtest.%v", key)
}

func (f *Feature) abtestConfig() (err error) {
	var (
		tmp  *featuregrpc.ABTestReply
		args = &featuregrpc.ABTestReq{TreeId: f.TreeID}
	)
	if tmp, err = f.featureClient.ABTest(context.Background(), args); err != nil {
		return
	}
	var tmpCache = make(map[string]*featuregrpc.ABTestItem)
	for _, item := range tmp.AbtestItems {
		if item != nil && item.KeyName != "" {
			tmpCache[item.KeyName] = item
		}
	}
	f.abTestCache = tmpCache
	return
}

// GetABTest 获取版本控制结果
func GetABTest(c context.Context, key string) string {
	return metadata.String(c, abtestKey(key))
}

// ABTestHttp http的handler方法(不支持自定义模式)
func (f *Feature) ABTestHttp() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		var (
			mid   int64
			buvid string
		)
		if midInter, ok := ctx.Get("mid"); ok {
			mid = midInter.(int64)
		}
		buvid = ctx.Request.Header.Get("Buvid")
		f.abtest(ctx, mid, buvid, nil)
	}
}

// ABTestGRPC moss的handler方法(不支持自定义模式)
func (f *Feature) ABTestGRPC() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 用户信息
		au, _ := auth.FromContext(ctx)
		// 设备信息
		dev, ok := device.FromContext(ctx)
		if !ok {
			return
		}
		return handler(f.abtest(ctx, au.Mid, dev.Buvid, nil), req)
	}
}

// AbTestManual 手动方法 支持自定义模式
func (f *Feature) AbTestManual(c context.Context, req *AbTestManual) context.Context {
	return f.abtest(c, req.Mid, req.Buvid, req.Custom)
}

func (f *Feature) abtest(c context.Context, mid int64, buvid string, _ map[string]string) context.Context {
	if f == nil || f.abTestCache == nil {
		return c
	}
	for key, rule := range f.abTestCache {
		if md, ok := metadata.FromContext(c); ok {
			var baseValue string
			// 分组参数
			switch rule.AbType {
			case "mid":
				baseValue = strconv.FormatInt(mid, 10)
			case "buvid":
				baseValue = buvid
			default:
				continue
			}
			md[abtestKey(key)] = calRuleABtest(rule, baseValue)
		}
	}
	return c
}

func calRuleABtest(rule *featuregrpc.ABTestItem, baseValue string) string {
	if baseValue == "" {
		return _default
	}
	// flow校验
	for _, flow := range rule.Config {
		if flow.Start < 0 || flow.End > rule.Bucket {
			return _default
		}
	}
	var salt = "%s"
	if rule.Salt != "" {
		salt = rule.Salt
	}
	var hashValue = int32(crc32.ChecksumIEEE([]byte(fmt.Sprintf(salt, baseValue)))) % rule.Bucket
	if hashValue != 0 {
		for _, exp := range rule.Config {
			// 白名单逻辑
			for _, v := range strings.Split(exp.Whitelist, ",") {
				if v == baseValue {
					return exp.Group
				}
			}
			// 命中逻辑
			if hashValue >= exp.Start && hashValue <= exp.End {
				return exp.Group
			}
		}
	}
	return _default
}
