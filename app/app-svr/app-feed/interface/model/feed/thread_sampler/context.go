package thread_sampler

import (
	"context"
)

type isSampledKey struct{}

func NewContext(ctx context.Context, isSampled bool) context.Context {
	return context.WithValue(ctx, isSampledKey{}, isSampled)
}

func FromContext(ctx context.Context) (isSampled, ok bool) {
	isSampled, ok = ctx.Value(isSampledKey{}).(bool)
	return
}
