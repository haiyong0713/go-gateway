package dao

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/distribution/distribution/internal/dao/kv"
	"go-gateway/app/app-svr/distribution/distribution/internal/distributionconst"
	"go-gateway/app/app-svr/distribution/distribution/internal/storagedriver"
	"go-gateway/app/app-svr/distribution/distribution/internal/storagedriver/experimentalflag"
	"go-gateway/app/app-svr/distribution/distribution/internal/storagedriver/kvstore"
	"go-gateway/app/app-svr/distribution/distribution/internal/storagedriver/multipletusflag"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/net/rpc/warden"
	"go-common/library/sync/pipeline/fanout"

	parabox "git.bilibili.co/bapis/bapis-go/community/interface/parabox"
	tus "git.bilibili.co/bapis/bapis-go/datacenter/service/titan"
	"github.com/google/wire"
	"github.com/pkg/errors"
)

var Provider = wire.NewSet(New, NewRedis, kv.NewKV)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
}

// dao dao.
type dao struct {
	redis   *redis.Redis
	kv      *kv.Taishan
	cache   *fanout.Fanout
	parabox parabox.ParaboxClient
	tus     tus.TitanUserServerClient
}

// New new a dao and return.
func New(r *redis.Redis, kv *kv.Taishan) (d Dao, cf func(), err error) {
	return newDao(r, kv)
}

func newDao(r *redis.Redis, kv *kv.Taishan) (d *dao, cf func(), err error) {
	cfg := struct {
		Parabox *warden.ClientConfig
		Tus     *warden.ClientConfig
	}{}
	if err := paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, nil, errors.Errorf("Failed to parse application.toml config: %+v", err)
	}

	d = &dao{
		redis: r,
		kv:    kv,
		cache: fanout.New("cache"),
	}

	parabox, err := parabox.NewClient(cfg.Parabox)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	d.parabox = parabox

	tus, err := newTusClint(cfg.Tus)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	d.tus = tus

	kvDriver := kvstore.New(kv, r)
	storagedriver.Register(kvDriver)

	expDriver := experimentalflag.New(parabox)
	storagedriver.Register(expDriver)

	tusmultipleDriver := multipletusflag.New(tus, kv)
	storagedriver.Register(tusmultipleDriver)

	cf = d.Close
	return
}

func (d *dao) Name() string {
	return distributionconst.DefaultStorageDriver
}

// Close close the resource.
func (d *dao) Close() {
	d.cache.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}

func newTusClint(cfg *warden.ClientConfig) (tus.TitanUserServerClient, error) {
	client := warden.NewClient(cfg)
	conn, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", "datacenter.titan.tus"))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return tus.NewTitanUserServerClient(conn), nil
}
