package common

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"go-common/library/log"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"
	sdkwarden "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden/server"
)

type ServiceMetaMethod func(sms []*sdkwarden.ServiceMeta) ([]*sdkwarden.ServiceMeta, error)

func BuildServiceMetaByDynPath(paths []*pb.DynPath) []*sdkwarden.ServiceMeta {
	out := make([]*sdkwarden.ServiceMeta, 0, len(paths))
	for _, v := range paths {
		sm := &sdkwarden.ServiceMeta{
			Pattern:     formatPattern(v),
			Target:      appIdFromDiscovery(v),
			ServiceName: identifier(v),
		}
		sm.ClientSDKConfig.AppID = appIdFromDiscovery(v)
		sm.ClientSDKConfig.ClientInfo.MaxRetries = v.ClientInfo.MaxRetries
		sm.ClientSDKConfig.ClientInfo.Timeout = v.ClientInfo.Timeout
		out = append(out, sm)
	}
	return out
}

// 生成适合peat moss的pattern
func formatPattern(req *pb.DynPath) string {
	if strings.HasPrefix(req.Pattern, "/") {
		// 精确匹配
		return fmt.Sprintf("= %s", req.Pattern)
	} else if strings.HasPrefix(req.Pattern, "~") {
		// 正则
		return fmt.Sprintf("~ %s", strings.TrimLeft(req.Pattern, "~ "))
	}
	// 前缀匹配
	return fmt.Sprintf("/%s", req.Pattern)
}

func appIdFromDiscovery(req *pb.DynPath) string {
	p, _ := url.Parse(req.ClientInfo.Endpoint)
	return p.Host
}

func identifier(req *pb.DynPath) string {
	if req.ClientInfo.AppId != "" {
		return req.ClientInfo.AppId
	}
	return req.ClientInfo.Endpoint
}

func RunGRPCProcess(sm []*sdkwarden.ServiceMeta, methods ...ServiceMetaMethod) ([]*sdkwarden.ServiceMeta, error) {
	for _, method := range methods {
		s, err := method(sm)
		if err != nil {
			return nil, err
		}
		sm = s
	}
	sort.SliceStable(sm, func(i, j int) bool {
		return sm[i].ServiceName < sm[j].ServiceName
	})
	return sm, nil
}

func ServiceMetaAppendBreakerAPIs(bas []*pb.BreakerAPI) ServiceMetaMethod {
	setMethodOption := func(in *pb.BreakerAction, mo *warden.MethodOption) {
		switch action := in.Action.(type) {
		case *pb.BreakerAction_Null:
			mo.BackupRetryOption.BackupAction = ""
		case *pb.BreakerAction_Ecode:
			mo.BackupRetryOption.BackupAction = "ecode"
			mo.BackupRetryOption.BackupECode = int(action.Ecode.Ecode)
		case *pb.BreakerAction_Placeholder:
			mo.BackupRetryOption.BackupAction = "placeholder"
			mo.BackupRetryOption.BackupPlaceholder = action.Placeholder.Data
		case *pb.BreakerAction_DirectlyBackup:
			mo.BackupRetryOption.BackupAction = "directly_backup"
			mo.BackupRetryOption.BackupTarget = action.DirectlyBackup.BackupUrl
		case *pb.BreakerAction_RetryBackup:
			mo.BackupRetryOption.BackupAction = "retry_backup"
			mo.BackupRetryOption.BackupTarget = action.RetryBackup.BackupUrl
		default:
			mo.BackupRetryOption.BackupAction = ""
			log.Warn("Unrecognized backup action: %+v", in)
		}
	}
	return func(sms []*sdkwarden.ServiceMeta) ([]*sdkwarden.ServiceMeta, error) {
		smMap, err := CopyServiceMeta(sms)
		if err != nil {
			return nil, err
		}
		for _, ba := range bas {
			serviceName, method, err := warden.SplitServiceMethod(ba.Api)
			if err != nil {
				log.Error("Failed to split service method: %+v", err)
				continue
			}
			sm, ok := smMap[serviceName]
			if !ok {
				log.Warn("No matched service meta: %s, %+v", serviceName, ba)
				continue
			}
			methodMap := CopyMethodOption(sm.ClientSDKConfig.MethodOption)
			methodOption, ok := methodMap[method]
			if !ok {
				methodOption := &warden.MethodOption{
					Method: method,
					BackupRetryOption: warden.BackupRetryOption{
						Ratio:                ba.Ratio,
						ForceBackupCondition: ba.Condition,
					},
				}
				setMethodOption(ba.Action, methodOption)
				sm.ClientSDKConfig.MethodOption = append(sm.ClientSDKConfig.MethodOption, methodOption)
				continue
			}
			methodOption.BackupRetryOption = warden.BackupRetryOption{
				Ratio:                ba.Ratio,
				ForceBackupCondition: ba.Condition,
			}
			setMethodOption(ba.Action, methodOption)
		}
		out := make([]*sdkwarden.ServiceMeta, 0, len(smMap))
		for _, v := range smMap {
			out = append(out, v)
		}
		return out, nil
	}
}

func CopyMethodOption(in []*warden.MethodOption) map[string]*warden.MethodOption {
	out := make(map[string]*warden.MethodOption, len(in))
	for _, method := range in {
		out[method.Method] = method
	}
	return out
}

func CopyServiceMeta(req []*sdkwarden.ServiceMeta) (map[string]*sdkwarden.ServiceMeta, error) {
	out := make(map[string]*sdkwarden.ServiceMeta)
	for _, v := range req {
		out[v.ServiceName] = v
	}
	return out, nil
}
