package dao

import (
	"context"
	"time"

	"go-gateway/app/app-svr/newmont/service/api"
	secmdl "go-gateway/app/app-svr/newmont/service/internal/model/section"

	"go-common/library/database/sql"
	bm "go-common/library/net/http/blademaster"

	"go-common/library/conf/paladin.v2"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New, NewDB)

// dao dao.
type dao struct {
	db         *sql.DB
	sectionDao *sectionDao
}

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	CreateSectionDao() *sectionDao
}

type sectionDao struct {
	httpClient *bm.Client
	db         *sql.DB
}

type SectionDao interface {
	WhiteCheck(c context.Context, checkURL string, mid int64, buvid string) (bool, error)
	RedDot(c context.Context, mid int64, redDotURL string) (bool, error)
	FetchDynamicConf(c context.Context, checkURL string, mid int64, buvid string) (*secmdl.DynamicConf, error)
	EffectUrl(c context.Context, mid int64, checkURL string) (bool, error)
	SidebarLimit(ctx context.Context) (map[int64][]*secmdl.SideBarLimit, error)
	SideBar(ctx context.Context, now time.Time) ([]*secmdl.SideBar, error)
	SidebarLang(ctx context.Context) (map[int64]string, error)
	SideBarModules(ctx context.Context, moduleType int32) (sm map[int32][]*secmdl.ModuleInfo, err error)
	Icons(c context.Context, startTime, endTime time.Time) (map[int64]*api.MngIcon, error)
	HiddenLimits(c context.Context) (map[int64][]*api.HiddenLimit, error)
	Hiddens(c context.Context, now time.Time) ([]*api.Hidden, error)
}

// New new a dao and return.
func New(db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(db)
}

func newDao(db *sql.DB) (d *dao, cf func(), err error) {
	var cfg struct {
		Client *bm.ClientConfig
	}
	if err = paladin.Get("http.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		db:         db,
		sectionDao: newSectionDao(cfg.Client, db),
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
	d.db.Close()
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return nil
}

func (d *dao) CreateSectionDao() *sectionDao {
	return d.sectionDao
}

func newSectionDao(clientConfig *bm.ClientConfig, db *sql.DB) *sectionDao {
	return &sectionDao{
		httpClient: bm.NewClient(clientConfig),
		db:         db,
	}
}
