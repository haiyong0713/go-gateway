package sdk

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go-common/component/metadata/device"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"

	featuregrpc "go-gateway/app/app-svr/feature/service/api"
	xmetric "go-gateway/app/app-svr/feature/service/model/metric"

	"google.golang.org/grpc"
)

func buildlimitKey(key string) string {
	return fmt.Sprintf("buildlimit.%v", key)
}

type BuildLimitManual struct {
	MobiApp string
	Device  string
	Build   int64
}

func (f *Feature) buildLimitConfig() (err error) {
	var (
		tmp  *featuregrpc.BuildLimitReply
		args = &featuregrpc.BuildLimitReq{TreeId: f.TreeID}
	)
	if tmp, err = f.featureClient.BuildLimit(context.Background(), args); err != nil {
		log.Error("featureSdkError buildlimit load cache err %v", err)
		xmetric.FeatureSdkError.Inc("buildlimit", "load_faild", "")
		return
	}
	// 原始配置重组
	var buildLimitCacheTmp = make(map[string]map[string][]*featuregrpc.BuildLimitConditions)
	for _, keyConfig := range tmp.Keys {
		keyName := keyConfig.KeyName
		for _, mobiAppConfig := range keyConfig.Plats {
			mobiAppName := mobiAppConfig.MobiApp
			buildLimitTmp, ok := buildLimitCacheTmp[keyName]
			if !ok {
				buildLimitTmp = make(map[string][]*featuregrpc.BuildLimitConditions)
				buildLimitCacheTmp[keyName] = buildLimitTmp
			}
			buildLimitTmp[mobiAppName] = mobiAppConfig.Conditions
		}
	}
	f.buildLimitCache = buildLimitCacheTmp
	return
}

func (f *Feature) BuildLimitHttp() bm.HandlerFunc {
	return func(ctx *bm.Context) {
		build, err := strconv.ParseInt(ctx.Request.Form.Get("build"), 10, 64)
		if err != nil {
			return
		}
		f.buildLimit(ctx, build, ctx.Request.Form.Get("mobi_app"), ctx.Request.Form.Get("device"))
	}
}

func (f *Feature) BuildLimitGRPC() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		dev, ok := device.FromContext(ctx)
		if !ok {
			return
		}
		return handler(f.buildLimit(ctx, dev.Build, dev.RawMobiApp, dev.Device), req)
	}
}

func (f *Feature) BuildLimitManual(c context.Context, req *BuildLimitManual) context.Context {
	return f.buildLimit(c, req.Build, req.MobiApp, req.Device)
}

func (f *Feature) buildLimit(c context.Context, build int64, mobiApp, device string) context.Context {
	if mobiApp == "iphone" {
		switch strings.TrimSpace(device) {
		case "pad":
			mobiApp = "iphone_pad"
		case "phone", "":
			mobiApp = "iphone"
		default:
			log.Warn("featureSdkError buildlimit unknown ios device: %v", device)
			mobiApp = "iphone_unknown"
		}
	}
	if f == nil || f.buildLimitCache == nil {
		return c
	}
	for key, v := range f.buildLimitCache {
		var (
			rule []*featuregrpc.BuildLimitConditions
			ok   bool
		)
		if rule, ok = v[mobiApp]; !ok || len(rule) == 0 {
			continue
		}
		if md, ok := metadata.FromContext(c); ok {
			md[buildlimitKey(key)] = calRule(rule, build)
		}
	}
	return c
}

// GetBuildLimit 白名单逻辑，key命中即为true，调用前务必确认原逻辑
func GetBuildLimit(c context.Context, key string, req *OriginResutl) bool {
	var res = metadata.Bool(c, buildlimitKey(key))
	if req != nil && res != req.BuildLimit {
		log.Error("featureSdkError buildlimit result wrong: key(%v) nw(%v) old(%+v)", key, res, req)
		xmetric.FeatureSdkError.Inc("buildlimit", "result_wrong", key)
		res = req.BuildLimit
	}
	return res
}

// nolint:gocognit
func calRule(verRules []*featuregrpc.BuildLimitConditions, build int64) bool {
	rMap := make(map[string][]int64)
	for _, vr := range verRules {
		rMap[vr.Op] = append(rMap[vr.Op], vr.Build)
	}
	// 不等于
	var result bool
	if neList, ok := rMap["ne"]; ok {
		if contains(neList, build) {
			return false
		}
		result = true
	}
	// 等于
	if eqList, ok := rMap["eq"]; ok {
		if contains(eqList, build) {
			return true
		}
		result = false
	}
	if len(rMap["lt"]) == 0 && len(rMap["le"]) == 0 && len(rMap["gt"]) == 0 && len(rMap["ge"]) == 0 {
		return result
	}
	// 区间起止
	var (
		leftPoint, rightPoint int64
		leftType              = "lt"
		rightType             = "gt"
	)
	// 大于
	for _, gt := range rMap["gt"] {
		if (rightPoint == 0 && gt > 0) || (gt < rightPoint) {
			rightPoint = gt
			rightType = "gt"
		}
	}
	// 大于等于
	for _, ge := range rMap["ge"] {
		if (rightPoint == 0 && ge > 0) || (ge < rightPoint) {
			rightPoint = ge
			rightType = "ge"
		}
	}
	// 小于
	for _, lt := range rMap["lt"] {
		if (leftPoint == 0 && lt > 0) || (lt > leftPoint) {
			leftPoint = lt
			leftType = "lt"
		}
	}
	// 小于等于
	for _, le := range rMap["le"] {
		if (leftPoint == 0 && le > 0) || (le > leftPoint) {
			leftPoint = le
			leftType = "le"
		}
	}
	// 类型判断
	var ruleType = "unintersection" // 无交集-unintersection;有交集:intersection
	if leftPoint == rightPoint {
		if (leftType == "le") && (rightType == "ge") { // 全集
			return true
		}
	}
	if leftPoint > rightPoint {
		ruleType = "intersection"
	}
	switch ruleType {
	case "intersection": // 与
		if hitRule(leftPoint, build, leftType) && hitRule(rightPoint, build, rightType) {
			return true
		}
	case "unintersection": // 或
		if hitRule(leftPoint, build, leftType) || hitRule(rightPoint, build, rightType) {
			return true
		}
	}
	return false
}

// contains 命中判断 等于true/不等于false
func contains(arr []int64, ele int64) bool {
	for _, a := range arr {
		if ele == a {
			return true
		}
	}
	return false
}

// hitRule 命中判断 命中true/未命中false
func hitRule(a int64, ele int64, op string) bool {
	switch op {
	case "gt":
		if ele <= a {
			return false
		}
	case "ge":
		if ele < a {
			return false
		}
	case "lt":
		if ele >= a {
			return false
		}
	case "le":
		if ele > a {
			return false
		}
	default:
		return false
	}
	return true
}
