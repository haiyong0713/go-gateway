package dao

import (
	"context"
	"fmt"
	"time"

	"go-common/library/cache/redis"
	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/net/rpc/warden"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/distribution/distribution/admin/internal/dao/kv"
	abm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/abtest"
	ac "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/actionlog"
	"go-gateway/app/app-svr/distribution/distribution/admin/internal/model/rename"
	tusm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tus"
	tmm "go-gateway/app/app-svr/distribution/distribution/admin/internal/model/tusmultiple"
	vcm "go-gateway/app/app-svr/distribution/distribution/model/tusmultipleversion"

	tus "git.bilibili.co/bapis/bapis-go/datacenter/service/titan"

	"github.com/google/wire"
	"github.com/pkg/errors"
)

var Provider = wire.NewSet(New, NewRedis, kv.NewKV)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	FetchConfigsFromTaishan(ctx context.Context, keys []string) (map[string][]byte, error)
	SaveABTestConfigs(ctx context.Context, details []*abm.Detail) error
	FetchAbtestExpID(ctx context.Context, expValue string) (int64, error)
	BatchFetchAbtestExpID(ctx context.Context, expValues []string) (map[string][]int64, error)
	FetchAbtestExpInfo(ctx context.Context, expID string) (*abm.Infos, error)
	BatchFetchAbtestExpInfo(ctx context.Context, req map[string]int64) ([]*abm.Infos, error)
	FetchAbtestGroupIDWithName(ctx context.Context, expID string) (map[string]string, error)
	SaveTusConfigs(ctx context.Context, details []*tusm.Detail) error
	BatchFetchTusInfos(ctx context.Context, tusValue []string) ([]*tusm.Info, error)
	SaveMultipleTusConfigs(ctx context.Context, details []*tmm.Detail, fieldName, configVersion string) error
	LogAction(ctx context.Context, param *ac.Log) (*ac.LogManagers, error)
	CreateRenameDao() RenameDao
	CreateTusEditDao() TusEditDao
	CreateTusMultipleVersionDao() TusMultipleVersionDao
}

// dao dao.
type dao struct {
	redis                 *redis.Redis
	kv                    *kv.Taishan
	demoExpire            int32
	bmClient              *bm.Client
	abtestHost            string
	tusHost               string
	renameDao             *renameDao
	tusEditDao            *tusEditDao
	tusMultipleVersionDao *tusMultipleVersionDao
	actionHost            string
}

type RenameDao interface {
	Rename(ctx context.Context, in *rename.Rename) error
	FetchRenameInfo(ctx context.Context, id string) (*rename.Rename, error)
	BatchFetchRenameInfo(ctx context.Context, ids []string) (map[string]*rename.Rename, error)
}

type renameDao struct {
	kv *kv.Taishan
}

type TusEditDao interface {
	FetchTargetTusValue(ctx context.Context, tusValues []string, mid int64) (tusValue string, err error)
	MigrateTusValueWithMids(ctx context.Context, tusValue string, mids map[int64]string) error
	MigrateTusValueToDefaultWithMids(ctx context.Context, tusValue string, mids map[int64]string) error
	PutinTusValueWithMids(ctx context.Context, tusValue string, mids map[int64]string) error
}

type tusEditDao struct {
	tus    tus.TitanUserServerClient
	crowed tus.CrowdClient
}

type TusMultipleVersionDao interface {
	FetchVersionManager(ctx context.Context, fieldName string) (*vcm.ConfigVersionManager, error)
	EditVersions(ctx context.Context, in *vcm.ConfigVersionManager) error
	BatchFetchVersionManager(ctx context.Context, fieldNames []string) ([]*vcm.ConfigVersionManager, error)
	DeleteVersionConfig(ctx context.Context, fieldName string, versionInfo *vcm.VersionInfo) error
}

type tusMultipleVersionDao struct {
	kv *kv.Taishan
}

// New new a dao and return.
func New(r *redis.Redis, kv *kv.Taishan) (d Dao, cf func(), err error) {
	return newDao(r, kv)
}

func newDao(r *redis.Redis, kv *kv.Taishan) (d *dao, cf func(), err error) {
	var cfg struct {
		DemoExpire xtime.Duration
		HTTP       *bm.ClientConfig
		ABTestHost string
		TusHost    string
		TusCfg     *warden.ClientConfig
		ActionHost string
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		redis:                 r,
		kv:                    kv,
		demoExpire:            int32(time.Duration(cfg.DemoExpire) / time.Second),
		bmClient:              bm.NewClient(cfg.HTTP, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		abtestHost:            cfg.ABTestHost,
		tusHost:               cfg.TusHost,
		renameDao:             newRenameDao(kv),
		tusEditDao:            newTusEditDao(cfg.TusCfg),
		tusMultipleVersionDao: newTusMultipleVersionDao(kv),
		actionHost:            cfg.ActionHost,
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}

func (d *dao) FetchConfigsFromTaishan(ctx context.Context, keys []string) (map[string][]byte, error) {
	req := d.kv.NewBatchGetReq(ctx, keys)
	reply, err := d.kv.BatchGet(ctx, req)
	if err != nil {
		return nil, err
	}
	if !reply.AllSucceed {
		return nil, errors.Wrapf(ecode.ServerErr, "Failed to fetch all configs from taishan")
	}
	configsWithKey := make(map[string][]byte, len(reply.Records))
	for _, v := range reply.Records {
		configsWithKey[string(v.Key)] = v.Columns[0].Value
	}
	return configsWithKey, nil
}

func newRenameDao(kv *kv.Taishan) *renameDao {
	return &renameDao{kv: kv}
}

func newTusEditDao(cfg *warden.ClientConfig) *tusEditDao {
	tusclient, err := func(cfg *warden.ClientConfig) (tus.TitanUserServerClient, error) {
		client := warden.NewClient(cfg)
		conn, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", "datacenter.titan.tus"))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return tus.NewTitanUserServerClient(conn), nil
	}(cfg)
	if err != nil {
		panic(err)
	}
	crowedClient, err := func(cfg *warden.ClientConfig) (tus.CrowdClient, error) {
		client := warden.NewClient(cfg)
		conn, err := client.Dial(context.Background(), fmt.Sprintf("discovery://default/%s", "datacenter.titan.core"))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return tus.NewCrowdClient(conn), nil
	}(cfg)
	if err != nil {
		panic(err)
	}
	return &tusEditDao{tus: tusclient, crowed: crowedClient}
}

func newTusMultipleVersionDao(kv *kv.Taishan) *tusMultipleVersionDao {
	return &tusMultipleVersionDao{kv: kv}
}

func (d *dao) CreateRenameDao() RenameDao {
	return d.renameDao
}

func (d *dao) CreateTusEditDao() TusEditDao {
	return d.tusEditDao
}

func (d *dao) CreateTusMultipleVersionDao() TusMultipleVersionDao {
	return d.tusMultipleVersionDao
}
