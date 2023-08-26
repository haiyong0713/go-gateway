package sessioncontext

import (
	"context"

	"go-common/component/metadata/auth"
	"go-common/component/metadata/device"

	"google.golang.org/grpc"
)

type sessionContextKey struct{}

type SessionContext interface {
	Device() device.Device
	Mid() int64
	ExtraContext() map[string]string
	ExtraContextValue(key string) (string, bool)
}

func FromContext(ctx context.Context) (SessionContext, bool) {
	ssCtx, ok := ctx.Value(sessionContextKey{}).(SessionContext)
	return ssCtx, ok
}

func NewContext(ctx context.Context, s SessionContext) context.Context {
	ctx = context.WithValue(ctx, sessionContextKey{}, s)
	return ctx
}

type sessionContextImpl struct {
	mid          int64
	device       device.Device
	extraContext map[string]string
}

func (s *sessionContextImpl) Device() device.Device {
	return s.device
}

func (s *sessionContextImpl) Mid() int64 {
	return s.mid
}

func (s *sessionContextImpl) ExtraContext() map[string]string {
	return s.extraContext
}

func (s *sessionContextImpl) ExtraContextValue(key string) (string, bool) {
	v, ok := s.extraContext[key]
	return v, ok
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		s := &sessionContextImpl{}

		a, ok := auth.FromContext(ctx)
		if ok {
			s.mid = a.Mid
		}

		dev, ok := device.FromContext(ctx)
		if ok {
			s.device = dev
		}
		extraContext, ok := extractExtraContext(req)
		if ok {
			s.extraContext = extraContext
		}
		ctx = NewContext(ctx, s)
		return handler(ctx, req)
	}
}

type extraContextCarrier interface {
	GetExtraContext() map[string]string
}

func extractExtraContext(req interface{}) (map[string]string, bool) {
	carrier, ok := req.(extraContextCarrier)
	if !ok {
		return nil, false
	}
	return carrier.GetExtraContext(), true
}
