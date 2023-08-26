package service

import "context"

type expConfigKey struct{}

type expConfig struct {
	isNewDevice    bool
	padIsNewDevice bool
}

func WithContext(ctx context.Context, cfg expConfig) context.Context {
	return context.WithValue(ctx, expConfigKey{}, cfg)
}

func GetExpConfigFromContext(ctx context.Context) expConfig {
	if ec, ok := ctx.Value(expConfigKey{}).(expConfig); ok {
		return ec
	}
	return expConfig{}
}

type ExpOption func(*expConfig)

func (ec *expConfig) Apply(opts ...ExpOption) {
	for _, opt := range opts {
		opt(ec)
	}
}

func WithBackgroundExp(exp bool) ExpOption {
	return func(ec *expConfig) {
		ec.isNewDevice = exp
	}
}

func WithBackgroundExpForPad(exp bool) ExpOption {
	return func(ec *expConfig) {
		ec.padIsNewDevice = exp
	}
}
