package kvstore

import (
	"context"
	"fmt"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/log"
	"go-common/library/sync/pipeline/fanout"

	"go-gateway/app/app-svr/distribution/distribution/internal/dao/kv"
	"go-gateway/app/app-svr/distribution/distribution/internal/distributionconst"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	"go-gateway/app/app-svr/distribution/distribution/internal/sessioncontext"

	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -struct_name=KVStore -batch=2 -max_group=20 -batch_err=break -nullcache=&DummyMessage{HasData:false} -check_null_code=$.HasData==false
	UserPreferenceBytes(ctx context.Context, keys []string) (map[string]*DummyMessage, error)
}

//go:generate kratos tool redisgen
type _redis interface {
	// redis: -struct_name=KVStore -key=StorePreferenceCacheKey -encode=json
	CacheUserPreferenceBytes(ctx context.Context, keys []string) (map[string]*DummyMessage, error)

	// redis: -struct_name=KVStore -key=StorePreferenceCacheKey -expire=d.cfg.PreferenceExpire -encode=json
	AddCacheUserPreferenceBytes(ctx context.Context, values map[string]*DummyMessage) error

	// redis: -struct_name=KVStore -key=StorePreferenceCacheKey
	DelCacheUserPreferenceBytes(ctx context.Context, keys []string) error
}

func StorePreferenceCacheKey(in string) string {
	return fmt.Sprintf("kvcache:%s", in)
}

type kvStoreConfig struct {
	PreferenceExpire int64
}

type KVStore struct {
	store *kv.Taishan
	redis *redis.Redis
	cache *fanout.Fanout
	cfg   kvStoreConfig
}

func New(store *kv.Taishan, redis *redis.Redis) *KVStore {
	cfg := kvStoreConfig{}
	ct := paladin.Map{}
	if err := paladin.Get("application.toml").Unmarshal(&ct); err != nil {
		panic(errors.Errorf("Failed to parse application.toml config: %+v", err))
	}
	if err := ct.Get("kvstore").UnmarshalTOML(&cfg); err != nil {
		panic(errors.Errorf("Failed to parse kvstore config: %+v", err))
	}
	if cfg.PreferenceExpire <= 0 {
		cfg.PreferenceExpire = 86400
	}
	return &KVStore{
		store: store,
		redis: redis,
		cache: fanout.New("kvstore-cache"),
		cfg:   cfg,
	}
}

func (kvs *KVStore) RawUserPreferenceBytes(ctx context.Context, keys []string) (map[string]*DummyMessage, error) {
	req := kvs.store.NewBatchGetReq(ctx, keys)
	reply, err := kvs.store.BatchGet(ctx, req)
	if err != nil {
		return nil, err
	}

	out := make(map[string]*DummyMessage, len(keys))
	for _, r := range reply.Records {
		if err := kv.CastTaishanError(r.Status); err != nil {
			log.Error("Failed to get preference from kv store: %+v: %+v", r, err)
			continue
		}
		ctr := &DummyMessage{}
		if err := ctr.Unmarshal(r.Columns[0].Value); err != nil {
			log.Error("Failed to unmarshal stored value as dummy message: %+v: %+v", r, err)
			continue
		}
		out[string(r.Key)] = ctr
	}
	return out, nil
}

func (kvs *KVStore) GetUserPreference(ctx context.Context, metas []*preferenceproto.PreferenceMeta) ([]*preferenceproto.Preference, error) {
	ssCtx, ok := sessioncontext.FromContext(ctx)
	if !ok {
		return nil, errors.Errorf("Session context is required")
	}
	keys := make([]string, 0, len(metas))
	keyToMeta := make(map[string]*preferenceproto.PreferenceMeta, len(metas))
	for _, m := range metas {
		keyBuilder := m.KeyBuilder()
		key := keyBuilder(ssCtx)
		keys = append(keys, key)
		keyToMeta[key] = m
	}
	preferenceBytes, err := kvs.UserPreferenceBytes(ctx, keys)
	if err != nil {
		return nil, err
	}

	out := make([]*preferenceproto.Preference, 0, len(metas))
	for key, value := range preferenceBytes {
		meta, ok := keyToMeta[key]
		if !ok {
			log.Error("Unexpected preference key: %+v", key)
			continue
		}
		ctr := dynamic.NewMessage(meta.ProtoDesc)
		if err := ctr.Unmarshal(value.Payload); err != nil {
			log.Error("Failed to unmarshal stored value as preference message: %+v", err)
			continue
		}
		out = append(out, &preferenceproto.Preference{
			Meta:    *meta,
			Message: ctr,
		})
	}

	return out, nil
}

func (kvs *KVStore) SetUserPreference(ctx context.Context, preferences []*preferenceproto.Preference) error {
	ssCtx, ok := sessioncontext.FromContext(ctx)
	if !ok {
		return errors.Errorf("Session context is required")
	}

	toPut := map[string][]byte{}
	toDelete := []string{}
	for _, p := range preferences {
		keyBuilder := p.Meta.KeyBuilder()
		key := keyBuilder(ssCtx)
		rawBytes, err := p.Message.Marshal()
		if err != nil {
			log.Error("Failed to marshal message: %+v", err)
			continue
		}
		toPut[key] = rawBytes
		toDelete = append(toDelete, key)
	}

	req := kvs.store.NewBatchPutReq(ctx, toPut, 0)
	reply, err := kvs.store.BatchPut(ctx, req)
	if err != nil {
		return err
	}
	if !reply.AllSucceed {
		for _, r := range reply.Records {
			if err := kv.CastTaishanError(r.Status); err != nil {
				log.Error("Failed to set preference to kv store: %+v: %+v", r, err)
				continue
			}
		}
	}
	_ = kvs.cache.Do(ctx, func(ctx context.Context) {
		if err := kvs.DelCacheUserPreferenceBytes(ctx, toDelete); err != nil {
			log.Error("Failed to delete preference cache: %+v: %+v", toDelete, err)
		}
	})
	return nil
}

func (*KVStore) Name() string {
	return distributionconst.DefaultStorageDriver
}
