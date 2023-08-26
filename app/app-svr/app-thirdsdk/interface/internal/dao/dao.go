package dao

import (
	"context"

	"go-common/library/conf/paladin.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-common/library/sync/pipeline/fanout"
	"go-gateway/app/app-svr/app-thirdsdk/interface/internal/model"
	arc "go-gateway/app/app-svr/archive/service/api"

	camp "git.bilibili.co/bapis/bapis-go/video/vod/playurlcamp"

	"github.com/google/wire"
	"github.com/pkg/errors"
)

var Provider = wire.NewSet(New)

// Dao dao interface
//
//go:generate kratos tool btsgen
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	ProtobufPlayurl(ctx context.Context, in *camp.RequestMsg) (*camp.ResponseMsg, error)
	Archive(ctx context.Context, aid int64) (*arc.Arc, error)
	UserBindSync(ctx context.Context, vendor string, param *model.UserBindParam) error
	ArcStatusSync(ctx context.Context, verdor string, param *model.ArcStatusParam) error
}

// dao dao.
type dao struct {
	cache         *fanout.Fanout
	campCli       camp.PlayurlServiceClient
	arcCli        arc.ArchiveClient
	httpMgr       *bm.Client
	userBindSync  string
	arcStatusSync string
}

// New new a dao and return.
func New() (d Dao, cf func(), err error) {
	return newDao()
}

func newDao() (d *dao, cf func(), err error) {
	type Host struct {
		Mgr string
	}
	var cfg struct {
		PlayURLCamp *warden.ClientConfig
		ArcGRPC     *warden.ClientConfig
		HTTPMgr     *bm.ClientConfig
		Host        *Host
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		cache:         fanout.New("cache"),
		userBindSync:  cfg.Host.Mgr + _userBindSyncURL,
		arcStatusSync: cfg.Host.Mgr + _arcStatusSyncURL,
	}
	if d.campCli, err = camp.NewClient(cfg.PlayURLCamp); err != nil {
		err = errors.WithStack(err)
		return
	}
	if d.arcCli, err = arc.NewClient(cfg.ArcGRPC); err != nil {
		err = errors.WithStack(err)
		return
	}
	d.httpMgr = bm.NewClient(cfg.HTTPMgr)
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
	_ = d.cache.Close()
}

// Ping ping the resource.
func (d *dao) Ping(_ context.Context) (err error) {
	return nil
}
