package warden

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/log"
	xtime "go-common/library/time"
	rootsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdk "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/client"
	clientmd "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/client/metadata"
	def "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/defaults"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/request"
	"go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/sdkerr"
	"go-gateway/app/app-svr/app-gw/sdk/http-sdk/blademaster/ab"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	grpcmetadata "google.golang.org/grpc/metadata"
)

type GRPCClientRawError struct{}

func jsonify(in interface{}) string {
	out, _ := json.Marshal(in)
	return string(out)
}

type MethodOption struct {
	Method            string // pure method name
	BackupRetryOption BackupRetryOption
	ClientInfo        clientmd.ClientInfo
}

type ClientSDKConfig struct {
	AppID        string
	MethodOption []*MethodOption
	ClientInfo   clientmd.ClientInfo
	SDKConfig    sdk.Config
}

func (csc *ClientSDKConfig) MatchedMethodOption(method string) (*MethodOption, bool) {
	for _, mopt := range csc.MethodOption {
		if mopt.Method == method {
			return mopt, true
		}
	}
	return nil, false
}

func (csc *ClientSDKConfig) Init() error {
	if csc.AppID == "" {
		return errors.Errorf("empty appid")
	}
	for _, mopt := range csc.MethodOption {
		mopt.BackupRetryOption.forceBackupCondition = ab.ParseCondition(mopt.BackupRetryOption.ForceBackupCondition)
		if mopt.BackupRetryOption.forceBackupCondition == nil {
			mopt.BackupRetryOption.forceBackupCondition = ab.TRUE
		}
	}
	return nil
}

// ClientInterceptor is
type ClientInterceptor struct {
	sync.RWMutex
	cfg ClientSDKConfig

	client *client.Client
}

// New is
func New(cfg ClientSDKConfig) *ClientInterceptor {
	handlers := def.Handlers()
	handlers.Build.Clear()
	handlers.Retry.Clear()
	handlers.ValidateResponse.Clear()
	client := client.New(cfg.SDKConfig, handlers)
	clientInterceptor := &ClientInterceptor{
		client: client,
	}
	if err := clientInterceptor.ensureConfig(cfg); err != nil {
		panic(err)
	}
	return clientInterceptor
}

func (ci *ClientInterceptor) Reload(cfg ClientSDKConfig) error {
	if err := ci.ensureConfig(cfg); err != nil {
		return err
	}
	return nil
}

func (ci *ClientInterceptor) ensureConfig(cfg ClientSDKConfig) error {
	if err := cfg.Init(); err != nil {
		log.Error("Failed to load client sdk config: %+v", err)
		return err
	}
	log.Info("Parsed a new client sdk config: %s", jsonify(cfg))
	ci.Lock()
	defer ci.Unlock()
	ci.cfg = cfg
	log.Info("Succeeded to load dynamic sdk config: %q", jsonify(ci.cfg))
	return nil
}

func (ci *ClientInterceptor) dupConfig() ClientSDKConfig {
	ci.RLock()
	cfg := ci.cfg
	ci.RUnlock()
	return cfg
}

func SplitServiceMethod(serviceMethod string) (string, string, error) {
	if serviceMethod != "" && serviceMethod[0] == '/' {
		serviceMethod = serviceMethod[1:]
	}
	pos := strings.LastIndex(serviceMethod, "/")
	if pos == -1 {
		return "", "", errors.Errorf("malformed method name: %q", serviceMethod)
	}
	service := serviceMethod[:pos]
	method := serviceMethod[pos+1:]
	return service, method, nil
}

func unwrapSDKError(in error) error {
	nextErr := errors.Cause(in)
	ttl := 10
	for {
		// avoid infinity loop
		if ttl -= 1; ttl < 0 {
			log.Error("unwrap SDK error max times exceeded: %+v", in)
			break
		}
		batchErr, ok := nextErr.(sdkerr.BatchedErrors)
		if ok {
			origErrs := batchErr.OrigErrs()
			if len(origErrs) > 0 {
				nextErr = origErrs[0]
				continue
			}
		}
		break
	}
	return nextErr
}

// same as http sdk
func (ci *ClientInterceptor) newABEnvContext(ctx context.Context) context.Context {
	kv := make([]ab.KV, 0, 32)
	device, ok := device.FromContext(ctx)
	if ok {
		kv = append(kv,
			ab.KVString("sid", device.Sid),
			ab.KVString("buvid3", device.Buvid3),
			ab.KVInt("build", device.Build),
			ab.KVString("buvid", device.Buvid),
			ab.KVString("channel", device.Channel),
			ab.KVString("device", device.Device),
			ab.KVString("rawplatform", device.RawPlatform),
			ab.KVString("rawmobiapp", device.RawMobiApp),
			ab.KVString("model", device.Model),
			ab.KVString("brand", device.Brand),
			ab.KVString("osver", device.Osver),
			ab.KVString("useragent", device.UserAgent),
			ab.KVInt("plat", int64(device.Plat())),
			ab.KVBool("isandroid", device.IsAndroid()),
			ab.KVBool("isIOS", device.IsIOS()),
			ab.KVBool("isweb", device.IsWeb()),
			ab.KVBool("isoverseas", device.IsOverseas()),
			ab.KVString("mobiapp", device.MobiApp()),
			ab.KVString("mobiappbulechange", device.MobiAPPBuleChange()),
		)
	}
	// stolen form exp metadata
	if omd, ok := grpcmetadata.FromIncomingContext(ctx); ok {
		carrier := ab.GRPCCarrier(omd)
		_, expKV := carrier.Get()
		kv = append(kv, expKV...)
	}
	t := ab.New(kv...)
	return ab.NewContext(ctx, t)
}

func setClientInfo(r *request.Request, clientInfo *clientmd.ClientInfo) {
	if clientInfo.MaxRetries != nil {
		r.Retryer = client.DefaultRetryer{
			NumMaxRetries: int(rootsdk.Int64Value(clientInfo.MaxRetries)),
		}
	}
	if clientInfo.Timeout != 0 {
		// 配置时录入的是ms
		//nolint
		t := xtime.Duration(clientInfo.Timeout) * xtime.Duration(time.Millisecond)
		_, ctx, _ := t.Shrink(r.Context())
		r.SetContext(ctx)
	}
}

func (ci *ClientInterceptor) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(parentCtx context.Context, serviceMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx := ci.newABEnvContext(parentCtx)
		cfg := ci.dupConfig()
		operation := &request.Operation{
			AppID:       cfg.AppID,
			Method:      serviceMethod,
			CC:          cc,
			Invoker:     invoker,
			Opts:        opts,
			CallContext: ctx,
		}
		sdkReq := ci.client.NewRequest(operation, req, reply)
		sdkReq.SetContext(ctx)

		inheritClientInfo := func(src clientmd.ClientInfo, dst *clientmd.ClientInfo) {
			if dst.Timeout <= 0 {
				dst.Timeout = src.Timeout
			}
			if dst.MaxRetries == nil {
				dst.MaxRetries = src.MaxRetries
			}
		}
		// setup method options
		clientInfo := cfg.ClientInfo
		func() {
			_, method, err := SplitServiceMethod(serviceMethod)
			if err != nil {
				log.Error("Failed to parse service method: %q: %+v", serviceMethod, err)
				return
			}
			mopt, ok := cfg.MatchedMethodOption(method)
			if ok {
				setupBackupRetry(sdkReq, &mopt.BackupRetryOption)
				inheritClientInfo(mopt.ClientInfo, &clientInfo)
			}
		}()
		setClientInfo(sdkReq, &clientInfo)

		if err := sdkReq.Send(); err != nil {
			rawErr := unwrapSDKError(err)
			StoreGRPCRawError(ctx, rawErr)
			return rawErr
		}
		return nil
	}
}

const (
	_grpcErrorStorage = "_app_gw_sdk_grpc_raw_error_store_"
)

func SetupGRPCRawErrorStorage(ctx context.Context, errStore *error) context.Context {
	//nolint:staticcheck
	return context.WithValue(ctx, _grpcErrorStorage, errStore)
}

func StoreGRPCRawError(ctx context.Context, err error) {
	errStore, ok := ctx.Value(_grpcErrorStorage).(*error)
	if !ok {
		return
	}
	*errStore = err
}
