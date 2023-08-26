package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	gomd "go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	pb "go-gateway/app/app-svr/app-gw/peat-moss/api"
	"go-gateway/app/app-svr/app-gw/peat-moss/internal/service"
	gwsdk "go-gateway/app/app-svr/app-gw/sdk"
	sdkwarden "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"

	"github.com/realab/go-grpc-http1/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	gmd "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

//nolint:unused
var svc pb.AppGatewayGRPCProxyServer
var rawSvc *service.Service

type as404 struct{}

func (as404) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	http.Error(w, fmt.Sprintf("unable to handle request: %q", req.URL.String()), http.StatusBadRequest)
}

// New new a bm server.
func New(s pb.AppGatewayGRPCProxyServer) (*bm.Engine, func(), error) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	if err := paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return nil, nil, err
	}
	if err := ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return nil, nil, err
	}
	svc = s
	rawSvc = s.(*service.Service)
	engine := bm.DefaultServer(&cfg)
	pb.RegisterAppGatewayGRPCProxyBMServer(engine, s)
	initRouter(engine)
	grpcServer, err := newGRPC(s, engine)
	if err != nil {
		return nil, nil, err
	}
	grpcDowngradingHandler := server.CreateDowngradingHandler(
		grpcServer.Server(), as404{},
		server.SkipValidateMethod(true),
		server.PreferGRPCWeb(true),
		server.AlwaysTryGRPCWeb(true),
	)
	engine.NoRoute(func(ctx *bm.Context) {
		if unGzip(ctx.Request) {
			log.Infoc(ctx, "unzip: %v headers: %v", ctx.Request.URL, ctx.Request.Header)
		}
		grpcDowngradingHandler.ServeHTTP(ctx.Writer, ctx.Request)
	})
	if err := engine.Start(); err != nil {
		return nil, nil, err
	}
	if _, err := grpcServer.Start(); err != nil {
		return nil, nil, err
	}
	closeFn := func() {
		//nolint:errcheck
		engine.Shutdown(context.Background())
		//nolint:errcheck
		grpcServer.Shutdown(context.Background())
	}
	return engine, closeFn, nil
}

func initRouter(e *bm.Engine) {
	e.Ping(ping)
}

func ping(ctx *bm.Context) {
	if _, err := rawSvc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func PeatMossSLBRetryPreHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	errStore := new(error)
	ctx = sdkwarden.SetupGRPCRawErrorStorage(ctx, errStore)
	resp, err := handler(ctx, req)
	castStatusError := func() {
		if err == nil {
			return
		}
		if *errStore == nil {
			return
		}
		if !ecode.EqualError(ecode.ServerErr, err) {
			return
		}
		grpcStatus, ok := status.FromError(*errStore)
		if !ok {
			// warden client 只会返回 ecode
			// 如果该原始 error 是一个 grpc status code 且 ecode 为 -500 时，也许就应该直接返回 status error
			return
		}
		// 当前仅处理 DataLoss
		if grpcStatus.Code() == codes.DataLoss {
			log.Warn("Cast RPC error as gRPC status error: %+v to %d:%s", err, grpcStatus.Code(), grpcStatus.Message())
			err = *errStore
		}
	}
	castStatusError()
	return resp, err
}

func metadataFromHTTP(ctx context.Context, key string) string {
	md, ok := gmd.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	v := md.Get(key)
	if len(v) <= 0 {
		return ""
	}
	return v[0]
}

// WARN: this function is not thread safe
func duplicateMetadata(ctx context.Context, key string, value string) context.Context {
	mdV, ok := gmd.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	mdV.Set(key, value)

	gomdV, ok := gomd.FromContext(ctx)
	if !ok {
		return ctx
	}
	gomdV[key] = value
	return ctx
}

func castHTTPMetadataAsGRPCMetadata(ctx context.Context, srcKey, dstKey string) context.Context {
	value := metadataFromHTTP(ctx, srcKey)
	if value != "" {
		ctx = duplicateMetadata(ctx, dstKey, value)
	}
	return ctx
}

func PeatMossMetaHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = castHTTPMetadataAsGRPCMetadata(ctx, "x1-bilispy-color", gomd.Color)

	startAt := time.Now()
	resp, err := handler(ctx, req)
	//nolint:errcheck
	grpc.SetTrailer(ctx,
		gmd.Pairs(
			"x-peat-moss-upstream-service-time", strconv.FormatInt(int64(time.Since(startAt)/time.Millisecond), 10),
			"server", fmt.Sprintf("peat-moss app-gw-sdk/%s", gwsdk.SDKVersion),
		),
	)
	return resp, err
}

func newGRPC(svc pb.AppGatewayGRPCProxyServer, engine *bm.Engine) (ws *warden.Server, err error) {
	var (
		cfg warden.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	proxy, err := newProxyPass()
	if err != nil {
		return nil, err
	}
	handler, interceptorSetter := proxy.WrappedHandler()
	ws = warden.NewServer(&cfg, grpc.UnknownServiceHandler(handler))
	ws.Use(PeatMossMetaHandler)
	ws.Use(PeatMossSLBRetryPreHandler)
	interceptorSetter(ws.Interceptor)
	pb.RegisterAppGatewayGRPCProxyServer(ws.Server(), svc)
	proxy.SetupRouter(engine)
	return
}
