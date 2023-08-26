package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	gometadata "go-common/library/net/metadata"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-gw/gateway/api"
	"go-gateway/app/app-svr/app-gw/sdk"
	sdkwarden "go-gateway/app/app-svr/app-gw/sdk/grpc-sdk/warden"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func jsonify(in interface{}) string {
	out, _ := json.Marshal(in)
	return string(out)
}

type clientSet struct {
	conn *grpc.ClientConn
	sdk  *sdkwarden.ClientInterceptor
}

// ProxyPass is
type ProxyPass struct {
	configLock sync.RWMutex
	cfg        Config

	clientLock sync.RWMutex
	client     map[string]*clientSet

	colorClientLock sync.RWMutex
	colorClient     map[string]map[string]*clientSet
}

func New(cfg Config) *ProxyPass {
	return &ProxyPass{
		cfg:         cfg,
		client:      map[string]*clientSet{},
		colorClient: map[string]map[string]*clientSet{},
	}
}

type dummyMessage struct {
	payload []byte
}

func (dm *dummyMessage) Reset()                   { dm.payload = dm.payload[:0] }
func (dm *dummyMessage) String() string           { return fmt.Sprintf("%q", dm.payload) }
func (dm *dummyMessage) ProtoMessage()            {}
func (dm *dummyMessage) Marshal() ([]byte, error) { return dm.payload, nil }
func (dm *dummyMessage) Unmarshal(in []byte) error {
	dm.payload = append(dm.payload[:0], in...)
	return nil
}

func (p *ProxyPass) Reload(cfg Config) error {
	return p.ensureConfig(cfg)
}

func (p *ProxyPass) ensureConfig(cfg Config) error {
	for _, ds := range cfg.DynService {
		if err := ds.Init(); err != nil {
			log.Error("Failed to load dynamic service: %+v", err)
			return err
		}
	}
	p.configLock.Lock()
	defer p.configLock.Unlock()
	p.cfg = cfg

	p.clientLock.Lock()
	defer p.clientLock.Unlock()
	for _, ds := range p.cfg.DynService {
		target := ds.ResolvableTarget()
		cs, ok := p.client[target]
		if !ok {
			continue
		}
		if err := cs.sdk.Reload(ds.ClientSDKConfig); err != nil {
			continue
		}
	}

	log.Info("Succeeded to load dynamic service: %s", jsonify(p.cfg))
	return nil
}

func (p *ProxyPass) dial(ctx context.Context, meta *ServiceMeta) (*clientSet, error) {
	target := meta.ResolvableTarget()
	clientCfg := meta.FixedClientConfig()

	clientSDK := sdkwarden.New(meta.ClientSDKConfig)
	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(clientSDK.UnaryClientInterceptor()),
	}
	client := warden.NewClient(clientCfg, opts...)
	dialed, err := client.Dial(ctx, target)
	if err != nil {
		return nil, err
	}
	p.clientLock.Lock()
	defer p.clientLock.Unlock()
	cs, ok := p.client[target]
	if ok {
		log.Warn("Already has established connection for service %s, will give up the new one.", target)
		dialed.Close()
		return cs, nil
	}
	newCS := &clientSet{
		conn: dialed,
		sdk:  clientSDK,
	}
	p.client[target] = newCS
	return newCS, nil
}

func (p *ProxyPass) dialWithColor(ctx context.Context, color string, meta *ServiceMeta) (*clientSet, error) {
	target := meta.ResolvableTarget()
	clientCfg := meta.FixedClientConfig()

	clientSDK := sdkwarden.New(meta.ClientSDKConfig)
	opts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(clientSDK.UnaryClientInterceptor()),
	}
	client := warden.NewClient(clientCfg, opts...)
	dialed, err := client.Dial(ctx, target)
	if err != nil {
		return nil, err
	}
	p.colorClientLock.Lock()
	defer p.colorClientLock.Unlock()
	cs, ok := p.colorClient[target][color]
	if ok {
		log.Warn("Already has established color %q connection for service %s, will give up the new one.", color, target)
		dialed.Close()
		return cs, nil
	}
	newCS := &clientSet{
		conn: dialed,
		sdk:  clientSDK,
	}
	if _, ok := p.colorClient[target]; !ok {
		p.colorClient[target] = map[string]*clientSet{}
	}
	p.colorClient[target][color] = newCS
	return newCS, nil
}

func (p *ProxyPass) colorConn(ctx context.Context, color string, meta *ServiceMeta) (*clientSet, error) {
	target := meta.ResolvableTarget()
	p.colorClientLock.RLock()
	cs, ok := p.colorClient[target][color]
	p.colorClientLock.RUnlock()
	if ok {
		return cs, nil
	}
	log.Warn("No exist service %s with color %q client set, will dial on demand", target, color)
	return p.dialWithColor(ctx, color, meta)
}

func (p *ProxyPass) conn(ctx context.Context, meta *ServiceMeta) (*clientSet, error) {
	color := gometadata.String(ctx, gometadata.Color)
	if color != "" {
		return p.colorConn(ctx, color, meta)
	}

	target := meta.ResolvableTarget()
	p.clientLock.RLock()
	cs, ok := p.client[target]
	p.clientLock.RUnlock()
	if ok {
		return cs, nil
	}
	log.Warn("No exist service %s client set, will dial on demand", target)
	cs, err := p.dial(ctx, meta)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

// MatchLongestPath is
func (p *ProxyPass) MatchLongestPath(serviceMethod string) (*ServiceMeta, bool) {
	matched := []*ServiceMeta{}
	p.configLock.RLock()
	for _, p := range p.cfg.DynService {
		if !p.matcher.Match(serviceMethod) {
			continue
		}
		matched = append(matched, p)
	}
	p.configLock.RUnlock()

	if len(matched) <= 0 {
		return nil, false
	}

	sort.Slice(matched, func(i, j int) bool {
		l, r := matched[i].matcher, matched[j].matcher
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

// The `parentCtx` is used to tracing the RPC call.
// The `proxyCtx` is only used to store values in proxy stage.
func (p *ProxyPass) handleProxy(parentCtx context.Context, proxyCtx *Context) (interface{}, error) {
	cs, err := p.conn(parentCtx, proxyCtx.serviceMeta)
	if err != nil {
		return nil, err
	}

	inMD, _ := metadata.FromIncomingContext(parentCtx)
	invokeCtx := metadata.NewOutgoingContext(parentCtx, inMD)

	stream := proxyCtx.serverStream
	header, trailer := metadata.MD{}, metadata.MD{}
	callOpts := []grpc.CallOption{
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	}
	reply := &dummyMessage{}
	if err := cs.conn.Invoke(invokeCtx, proxyCtx.serviceMethod, proxyCtx.Req(), reply, callOpts...); err != nil {
		err = errors.Cause(err)
		return nil, err
	}
	if err := stream.SendHeader(header); err != nil {
		return nil, errors.WithStack(err)
	}
	stream.SetTrailer(trailer)
	if err := stream.SendMsg(reply); err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

func (p *ProxyPass) WrappedHandler() (grpc.StreamHandler, func(grpc.UnaryServerInterceptor)) {
	interceptor := func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
	streamHandler := func(srv interface{}, stream grpc.ServerStream) error {
		serviceMethod, ok := grpc.MethodFromServerStream(stream)
		if !ok {
			return status.Errorf(codes.Internal, "failed to get method from stream")
		}
		service, _, err := sdkwarden.SplitServiceMethod(serviceMethod)
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, err.Error())
		}
		serviceMeta, ok := p.MatchLongestPath(serviceMethod)
		if !ok {
			return status.Errorf(codes.FailedPrecondition, "unrecognized service: %q", service)
		}

		proxyCtx := &Context{
			Context:       stream.Context(),
			srv:           srv,
			serverStream:  stream,
			serviceMethod: serviceMethod,
			serviceMeta:   serviceMeta,
		}
		if err := stream.RecvMsg(proxyCtx.Req()); err != nil {
			return errors.WithStack(err)
		}
		serverInfo := &grpc.UnaryServerInfo{
			Server:     srv,
			FullMethod: serviceMethod,
		}
		unaryHandler := func(parentCtx context.Context, _ interface{}) (interface{}, error) {
			return p.handleProxy(parentCtx, proxyCtx)
		}
		if _, err := interceptor(proxyCtx, proxyCtx.Req(), serverInfo, unaryHandler); err != nil {
			return err
		}
		return nil
	}
	interceptorSetter := func(inputInterceptor grpc.UnaryServerInterceptor) {
		interceptor = inputInterceptor
	}
	return streamHandler, interceptorSetter
}

func (p *ProxyPass) configs(ctx *bm.Context) {
	res, err := p.Configs()
	if err != nil {
		ctx.String(400, "%v", err)
		return
	}
	ctx.Bytes(200, "text/plain", res)
}

func (p *ProxyPass) Configs() ([]byte, error) {
	tomlBuf := bytes.Buffer{}
	encoder := toml.NewEncoder(&tomlBuf)
	raw := struct {
		ProxyConfig Config
	}{
		ProxyConfig: p.dupConfig(),
	}
	if err := encoder.Encode(raw); err != nil {
		log.Error("Failed to encode proxy config as toml: %+v: %+v", p.cfg, err)
		return nil, err
	}
	return tomlBuf.Bytes(), nil
}

func (p *ProxyPass) dupConfig() Config {
	p.configLock.Lock()
	cfg := p.cfg
	p.configLock.Unlock()
	return cfg
}

func (p *ProxyPass) profile(ctx *bm.Context) {
	ctx.JSON(p.Profile(), nil)
}

type GatewayProfile struct {
	GatewayVersion string `json:"gateway_version"`
	SDKVersion     string `json:"sdk_version"`
	ConfigDigest   string `json:"config_digest"`
}

func (p *ProxyPass) Profile() *GatewayProfile {
	return &GatewayProfile{
		GatewayVersion: api.GatewayVersion, // todo sdk理论上不该引用 gateway 的包
		SDKVersion:     sdk.SDKVersion,
		ConfigDigest:   p.getCfgDigest(),
	}
}

func (p *ProxyPass) getCfgDigest() string {
	cfg := p.dupConfig()
	return cfg.Digest()
}

func (p *ProxyPass) reloadCfg(ctx *bm.Context) {
	req := &struct {
		Digest         string `form:"digest" validate:"required"`
		Content        string `form:"content" validate:"required"`
		OriginalDigest string `form:"original_digest" validate:"required"`
	}{}
	if err := ctx.Bind(req); err != nil {
		return
	}

	raw := struct {
		ProxyConfig *Config
	}{}
	if _, err := toml.Decode(req.Content, &raw); err != nil {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("invalid sdk builder config: %+v", err)))
		return
	}
	if raw.ProxyConfig == nil {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, "empty proxy config"))
		return
	}

	cfgDigest := raw.ProxyConfig.Digest()
	if cfgDigest != req.Digest {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("mismatched digest: %q != %q", cfgDigest, req.Digest)))
		return
	}
	if err := p.CasReload(*raw.ProxyConfig, req.OriginalDigest); err != nil {
		ctx.JSON(nil, ecode.Error(ecode.RequestErr, fmt.Sprintf("failed to load provided config: %+v", err)))
		return
	}

	reply := struct {
		Loaded string `json:"loaded"`
		Digest string `json:"digest"`
	}{}
	loaded := p.dupConfig()
	buf := &bytes.Buffer{}
	//nolint:errcheck
	toml.NewEncoder(buf).Encode(loaded)
	reply.Loaded = buf.String()
	reply.Digest = loaded.Digest()
	ctx.JSON(reply, nil)
}

func (p *ProxyPass) CasReload(cfg Config, originalDigest string) error {
	for _, ds := range cfg.DynService {
		if err := ds.Init(); err != nil {
			log.Error("Failed to load dynamic service: %+v", err)
			return err
		}
		log.Info("Parsed a new proxy rule: %+v", ds)
	}
	p.configLock.Lock()
	defer p.configLock.Unlock()
	if p.cfg.Digest() != originalDigest {
		return errors.Errorf("invalid original digest: %s, %s", p.cfg.Digest(), originalDigest)
	}
	p.cfg = cfg

	p.clientLock.Lock()
	defer p.clientLock.Unlock()
	for _, ds := range p.cfg.DynService {
		target := ds.ResolvableTarget()
		cs, ok := p.client[target]
		if !ok {
			continue
		}
		if err := cs.sdk.Reload(ds.ClientSDKConfig); err != nil {
			continue
		}
	}
	log.Info("Succeeded to reload proxy config: %+v", p.cfg)
	return nil
}

func (p *ProxyPass) SetupRouter(e *bm.Engine) {
	e.GET("/_/grpc-configs.toml", p.configs)
	e.GET("/_/grpc-profile", p.profile)
	e.POST("/_/grpc-reload", p.reloadCfg)
}
