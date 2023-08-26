package adresource

import (
	"context"
	"strconv"
	"sync"

	"go-common/library/log"

	"github.com/pkg/errors"
)

var (
	globalResourceIDStore = &resourceIDStore{
		store: map[tScene]tResourceID{},
	}
)

type Option func(*calcResourceIDConfig)

type calcResourceIDConfig struct {
	panicOnNoScene bool
}

func (cfg *calcResourceIDConfig) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}

func PanicOnNoScene(toPanic bool) Option {
	return func(cfg *calcResourceIDConfig) {
		cfg.panicOnNoScene = toPanic
	}
}

type resourceIDStore struct {
	lock  sync.RWMutex
	store map[tScene]tResourceID
}

func (r *resourceIDStore) Register(scene tScene, id tResourceID) {
	r.lock.Lock()
	defer r.lock.Unlock()
	if _, ok := r.store[scene]; ok {
		panic(errors.Errorf("duplicated resource scene detected: %q", scene))
	}
	r.store[scene] = id
}

func (r *resourceIDStore) CalcResourceID(ctx context.Context, scene tScene, opts ...Option) (tResourceID, bool) {
	cfg := &calcResourceIDConfig{}
	cfg.Apply(opts...)

	r.lock.RLock()
	defer r.lock.RUnlock()
	id, ok := r.store[scene]
	StatSceneResourceID.Inc(string(scene), strconv.FormatInt(int64(id), 10))
	if !ok {
		if cfg.panicOnNoScene {
			panic(errors.Errorf("Failed to get resource with scene: %q", scene))
		}
		log.Warn("Failed to get resource with scene: %q", scene)
		return EmptyResourceID, false
	}
	return id, true
}

func Register(scene tScene, id tResourceID) {
	globalResourceIDStore.Register(scene, id)
}

func CalcResourceID(ctx context.Context, scene tScene, opts ...Option) (tResourceID, bool) {
	defaultOpts := []Option{
		PanicOnNoScene(false),
	}
	runtimeOpts := append(defaultOpts, opts...)
	id, ok := globalResourceIDStore.CalcResourceID(ctx, scene, runtimeOpts...)
	return id, ok
}
