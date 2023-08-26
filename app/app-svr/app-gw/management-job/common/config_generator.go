package common

import (
	"net/url"
	"regexp"
	"sort"
	"strings"

	"go-common/library/log"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	sdk "go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/client/metadata"
)

func IsMatchByPattern(pattern string, api string) bool {
	if !strings.HasPrefix(pattern, "~ ") {
		return false
	}
	rawExp := strings.TrimPrefix(pattern, "~ ")
	reg, err := regexp.Compile(rawExp)
	if err != nil {
		return false
	}
	return reg.MatchString(api)
}

func IsMatchByPrefix(prefix string, api string) bool {
	if !strings.HasPrefix(api, "/") {
		return false
	}
	return strings.HasPrefix(api, prefix)
}

func IsExactlyMatch(pattern, api string) bool {
	if !strings.HasPrefix(pattern, "= ") {
		return false
	}
	extPath := strings.TrimPrefix(pattern, "= ")
	return extPath == api
}

func EnabledBreakerAPI(in []*pb.BreakerAPI) []*pb.BreakerAPI {
	out := []*pb.BreakerAPI{}
	for _, v := range in {
		if !v.Enable {
			continue
		}
		out = append(out, v)
	}
	return out
}

func EnabledQuotaMethod(in []*pb.QuotaMethod) []*pb.QuotaMethod {
	out := []*pb.QuotaMethod{}
	for _, v := range in {
		if !v.Enable {
			continue
		}
		out = append(out, v)
	}
	return out
}

func EnabledDynPath(in []*pb.DynPath) []*pb.DynPath {
	filtered := make([]*pb.DynPath, 0, len(in))
	for _, v := range in {
		if v.ClientInfo == nil {
			continue
		}
		if !v.Enable {
			continue
		}
		filtered = append(filtered, v)
	}
	return filtered
}

func SetBreakerAction(in *pb.BreakerAction, pm *sdk.PathMeta) {
	switch action := in.Action.(type) {
	case *pb.BreakerAction_Null:
		pm.BackupRetryOption.BackupAction = ""
	case *pb.BreakerAction_Ecode:
		pm.BackupRetryOption.BackupAction = "ecode"
		pm.BackupRetryOption.BackupECode = action.Ecode.Ecode
	case *pb.BreakerAction_Placeholder:
		pm.BackupRetryOption.BackupAction = "placeholder"
		pm.BackupRetryOption.BackupPlaceholder = action.Placeholder.Data
	case *pb.BreakerAction_DirectlyBackup:
		pm.BackupRetryOption.BackupAction = "directly_backup"
		pm.BackupRetryOption.BackupURL = action.DirectlyBackup.BackupUrl
	case *pb.BreakerAction_RetryBackup:
		pm.BackupRetryOption.BackupAction = "retry_backup"
		pm.BackupRetryOption.BackupURL = action.RetryBackup.BackupUrl
	default:
		pm.BackupRetryOption.BackupAction = ""
		log.Warn("Unrecognized backup action: %+v", in)
	}
}

func BuildPathMetaByDynPath(paths []*pb.DynPath) []*sdk.PathMeta {
	out := make([]*sdk.PathMeta, 0, len(paths))
	for _, v := range paths {
		pm := &sdk.PathMeta{Pattern: v.Pattern}
		if v.ClientInfo != nil {
			pm.ClientInfo = metadata.ClientInfo{
				AppID:      v.ClientInfo.AppId,
				Endpoint:   v.ClientInfo.Endpoint,
				MaxRetries: v.ClientInfo.MaxRetries,
				Timeout:    v.ClientInfo.Timeout,
			}
		}
		out = append(out, pm)
	}
	return out
}

type PathMetaMethod func(pms []*sdk.PathMeta) ([]*sdk.PathMeta, error)

func RunProcess(pm []*sdk.PathMeta, methods ...PathMetaMethod) ([]*sdk.PathMeta, error) {
	for _, method := range methods {
		p, err := method(pm)
		if err != nil {
			return nil, err
		}
		pm = p
	}
	sort.SliceStable(pm, func(i, j int) bool {
		return pm[i].Pattern < pm[j].Pattern
	})
	return pm, nil
}

func PathMetaAppendBreakerAPIs(bas []*pb.BreakerAPI) PathMetaMethod {
	return func(pms []*sdk.PathMeta) ([]*sdk.PathMeta, error) {
		pmMap, err := CopyPathMeta(pms)
		if err != nil {
			return nil, err
		}
		for _, ba := range bas {
			u := &url.URL{Path: ba.Api}
			pm, ok := MatchLongestPath(u, pmMap)
			if !ok {
				log.Warn("No matched longest path: %+v", ba)
				continue
			}
			switch {
			case IsExactlyMatch(pm.Pattern, ba.Api):
				pmMap[pm.Pattern].BackupRetryOption = sdk.BackupRetryOption{
					ForceBackupCondition: ba.Condition,
					Ratio:                ba.Ratio,
				}
				SetBreakerAction(ba.Action, pmMap[pm.Pattern])
			case IsMatchByPattern(pm.Pattern, ba.Api), IsMatchByPrefix(pm.Pattern, ba.Api):
				pathMeta := new(sdk.PathMeta)
				*pathMeta = *pm
				pathMeta.BackupRetryOption = sdk.BackupRetryOption{
					ForceBackupCondition: ba.Condition,
					Ratio:                ba.Ratio,
				}
				SetBreakerAction(ba.Action, pathMeta)
				pathMeta.Pattern = "= " + ba.Api
				if err := pathMeta.InitStatic(); err != nil {
					log.Error("Failed to init derived path meta: %+v: %+v", pathMeta, err)
					continue
				}
				pmMap[pathMeta.Pattern] = pathMeta
			default:
				log.Error("Failed to match path meta: %+v: %+v: %+v", ba, pm, pmMap)
			}
		}
		out := make([]*sdk.PathMeta, 0, len(pmMap))
		for _, v := range pmMap {
			out = append(out, v)
		}
		return out, nil
	}
}

func PathMetaAppendRateLimiter(quotaMethods []*pb.QuotaMethod) PathMetaMethod {
	return func(pms []*sdk.PathMeta) ([]*sdk.PathMeta, error) {
		pmMap, err := CopyPathMeta(pms)
		if err != nil {
			return nil, err
		}
		for _, qm := range quotaMethods {
			u := &url.URL{Path: qm.Api}
			pm, ok := MatchLongestPath(u, pmMap)
			if !ok {
				log.Warn("No matched longest path: %+v", qm)
				continue
			}
			switch {
			case IsExactlyMatch(pm.Pattern, qm.Api):
				pmMap[pm.Pattern].RateLimiterOption = sdk.RateLimiterOption{Rule: qm.Rule}
				pmMap[pm.Pattern].RateLimiterOptions = append(pmMap[pm.Pattern].RateLimiterOptions,
					&sdk.RateLimiterOption{Rule: qm.Rule})
			case IsMatchByPattern(pm.Pattern, qm.Api), IsMatchByPrefix(pm.Pattern, qm.Api):
				pathMeta := new(sdk.PathMeta)
				*pathMeta = *pm
				pathMeta.RateLimiterOption = sdk.RateLimiterOption{Rule: qm.Rule}
				pathMeta.RateLimiterOptions = append(pathMeta.RateLimiterOptions, &sdk.RateLimiterOption{Rule: qm.Rule})
				pathMeta.Pattern = "= " + qm.Api
				if err := pathMeta.InitStatic(); err != nil {
					log.Error("Failed to init derived path meta: %+v: %+v", pathMeta, err)
					continue
				}
				pmMap[pathMeta.Pattern] = pathMeta

			default:
				log.Error("Failed to match path meta: %+v: %+v: %+v", qm, pm, pmMap)
			}
		}
		out := make([]*sdk.PathMeta, 0, len(pmMap))
		for _, v := range pmMap {
			out = append(out, v)
		}
		return out, nil
	}
}

func CopyPathMeta(req []*sdk.PathMeta) (map[string]*sdk.PathMeta, error) {
	out := make(map[string]*sdk.PathMeta)
	for _, v := range req {
		if err := v.InitStatic(); err != nil {
			return nil, err
		}
		out[v.Pattern] = v
	}
	return out, nil
}

func MatchLongestPath(url *url.URL, pm map[string]*sdk.PathMeta) (*sdk.PathMeta, bool) {
	matched := []*sdk.PathMeta{}
	for _, p := range pm {
		if !p.GetMatcher().Match(url) {
			continue
		}
		matched = append(matched, p)
	}

	if len(matched) <= 0 {
		return nil, false
	}

	sort.Slice(matched, func(i, j int) bool {
		l, r := matched[i].GetMatcher(), matched[j].GetMatcher()
		if l.Priority() < r.Priority() {
			return true
		}
		if l.Priority() == r.Priority() {
			return l.Len() > r.Len()
		}
		return false
	})
	return matched[0], true
}
