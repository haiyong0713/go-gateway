package dao

import (
	"context"

	"go-common/library/conf/paladin.v2"
	"go-common/library/database/sql"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-gw/baas/api"
	"go-gateway/app/app-svr/app-gw/baas/internal/model"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	TxBegin(ctx context.Context) (*sql.Tx, error)

	ModelAll(ctx context.Context, param *api.ModelListRequest) ([]*api.MapperModel, error)
	ModelAllCount(ctx context.Context, treeID int64) (int64, error)
	ModelByName(ctx context.Context, modelName string, treeID int64) (*api.MapperModel, error)
	ModelFieldByName(ctx context.Context, modelName string) ([]*api.MapperModelField, error)
	ModelField(ctx context.Context) ([]*api.MapperModelField, error)
	Transact(c context.Context, txFunc func(tx *sql.Tx) error) (err error)
	TxInsertModel(tx *sql.Tx, modelName, desc string, treeID int64) error
	TxInsertModelField(tx *sql.Tx, field *api.ModelField) error
	AddModelField(ctx context.Context, field *api.AddModelFieldRequest) error
	UpdateModelField(ctx context.Context, field *api.UpdateModelFieldRequest) error
	DelModelField(ctx context.Context, id int64) error
	ModelFieldRule(ctx context.Context) (map[string]*api.MapperModelFieldRule, error)
	AddFieldRule(ctx context.Context, param []*model.ItemFieldRule) error
	UpdateFieldRule(ctx context.Context, param *api.UpdateModelFieldRuleRequest) error

	ExportList(ctx context.Context) ([]*api.BaasExport, error)
	AddExport(ctx context.Context, param *api.AddExportRequest) error
	UpdateExport(ctx context.Context, export *api.UpdateExportRequest) error

	ImportByExportIds(ctx context.Context, ids []int64) (map[int64][]*api.BaasImport, error)
	AddImport(ctx context.Context, param *api.AddImportRequest) error
	UpdateImport(ctx context.Context, param *api.UpdateImportRequest) error
	ImportAll(ctx context.Context) (map[int64][]*api.ImportItem, error)

	FetchRoleTree(ctx context.Context, username, cookie string) ([]*model.Node, error)

	RawHttpImpl(ctx context.Context, uri string) ([]byte, error)
}

// dao dao.
type dao struct {
	db    *sql.DB
	cache *fanout.Fanout
	http  *bm.Client
	// http host
	Hosts struct {
		Easyst string
	}
}

// New new a dao and return.
func New(db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(db)
}

// Close close the resource.
// Close close the resource.
func (d *dao) Close() {
	d.db.Close()
	d.cache.Close()
}

func newDao(db *sql.DB) (d *dao, cf func(), err error) {
	var cfg struct {
		HTTPClient *bm.ClientConfig
	}
	if err = paladin.Get("http.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	resolver := resolver.New(nil, discovery.Builder())
	d = &dao{
		db:    db,
		cache: fanout.New("cache"),
		http:  bm.NewClient(cfg.HTTPClient, bm.SetResolver(resolver)),
	}
	ac := &paladin.TOML{}
	if err = paladin.Watch("application.toml", ac); err != nil {
		return
	}
	if err := ac.Get("hosts").UnmarshalTOML(&d.Hosts); err != nil {
		panic(err)
	}
	cf = d.Close
	return
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) error {
	return nil
}

func (d *dao) TxBegin(ctx context.Context) (*sql.Tx, error) {
	return d.db.Begin(ctx)
}
