package view

import (
	"context"

	"go-gateway/app/app-svr/app-view/interface/service/view/dependency"
)

type viewConfigKey struct{}

type viewConfig struct {
	dep             dependency.ViewDependency
	skipRelate      bool
	skipSpecialCell bool

	popupExp           bool
	autoSwindowExp     bool
	adTab              bool
	smallWindowExp     bool
	newSwindowExp      bool
	relatesBiserialExp bool
}

type ViewOption func(*viewConfig)

func SkipRelate(skip bool) ViewOption {
	return func(vc *viewConfig) {
		vc.skipRelate = skip
	}
}

func SkipSpecialCell(skip bool) ViewOption {
	return func(vc *viewConfig) {
		vc.skipSpecialCell = skip
	}
}

func WithPopupExp(exp bool) ViewOption {
	return func(vc *viewConfig) {
		vc.popupExp = exp
	}
}

func WithAutoSwindowExp(exp bool) ViewOption {
	return func(vc *viewConfig) {
		vc.autoSwindowExp = exp
	}
}

func WithSmallWindowExp(exp bool) ViewOption {
	return func(vc *viewConfig) {
		vc.smallWindowExp = exp
	}
}

func WithNewSwindowExp(exp bool) ViewOption {
	return func(vc *viewConfig) {
		vc.newSwindowExp = exp
	}
}

func WithRelatesBiserialExp(exp bool) ViewOption {
	return func(vc *viewConfig) {
		vc.relatesBiserialExp = exp
	}
}

func WithDependency(in dependency.ViewDependency) ViewOption {
	return func(vc *viewConfig) {
		vc.dep = in
	}
}

func WithAdTab(exp bool) ViewOption {
	return func(vc *viewConfig) {
		vc.adTab = exp
	}
}

func (vc *viewConfig) Apply(opts ...ViewOption) {
	for _, opt := range opts {
		opt(vc)
	}
}

func WithContext(ctx context.Context, cfg viewConfig) context.Context {
	return context.WithValue(ctx, viewConfigKey{}, cfg)
}

func FromContextOrCreate(ctx context.Context, create func() viewConfig) viewConfig {
	// retrive exist or create new
	vc, ok := ctx.Value(viewConfigKey{}).(viewConfig)
	if !ok {
		return create()
	}
	return vc
}

func (s *Service) defaultViewConfigCreater() func() viewConfig {
	return func() viewConfig {
		return viewConfig{
			dep: s.onDemandViewDependency(),
		}
	}
}
