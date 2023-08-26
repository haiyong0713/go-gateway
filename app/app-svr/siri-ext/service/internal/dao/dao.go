package dao

import (
	"context"

	"go-common/library/conf/paladin.v2"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/siri-ext/service/internal/model"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	favmodel "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	uparcapi "git.bilibili.co/bapis/bapis-go/up-archive/service"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	toviewsgrpc "git.bilibili.co/bapis/bapis-go/community/service/toview"

	"github.com/google/wire"
)

var Provider = wire.NewSet(New)

type host struct {
	Search string
}

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) error
	Suggest3(ctx context.Context, arg *model.SearchSuggestReq) (*model.Suggest3, error)
	AccountInfo3(ctx context.Context, mid int64) (*accgrpc.Info, error)
	ArcPassed(ctx context.Context, mid int64) (*uparcapi.ArcPassedReply, error)
	UserDefaultFavFolder(ctx context.Context, mid int64) (*favmodel.Folder, error)
	UserToViewsIsEmpty(ctx context.Context, mid int64) bool
}

// dao dao.
type dao struct {
	locgrpc    locgrpc.LocationClient
	httpClient *bm.Client
	host       host
	account    accgrpc.AccountClient
	uparc      uparcapi.UpArchiveClient
	fav        favgrpc.FavoriteClient
	toviews    toviewsgrpc.ToViewsClient
}

// New new a dao and return.
func New() (d Dao, cf func(), err error) {
	return newDao()
}

func newDao() (d *dao, cf func(), err error) {
	var cfg struct {
		HTTPSearch   *bm.ClientConfig
		LocationGRPC *warden.ClientConfig
		AccountGRPC  *warden.ClientConfig
		UpArcGRPC    *warden.ClientConfig
		FavGRPC      *warden.ClientConfig
		ToViewsGRPC  *warden.ClientConfig
		Host         host
	}
	if err = paladin.Get("application.toml").UnmarshalTOML(&cfg); err != nil {
		return
	}
	d = &dao{
		host:       cfg.Host,
		httpClient: bm.NewClient(cfg.HTTPSearch),
	}
	d.locgrpc, err = locgrpc.NewClient(cfg.LocationGRPC)
	if err != nil {
		return
	}
	d.account, err = accgrpc.NewClient(cfg.AccountGRPC)
	if err != nil {
		return
	}
	d.uparc, err = uparcapi.NewClient(cfg.UpArcGRPC)
	if err != nil {
		return
	}
	d.fav, err = favgrpc.NewClient(cfg.FavGRPC)
	if err != nil {
		return
	}
	d.toviews, err = toviewsgrpc.NewClient(cfg.ToViewsGRPC)
	if err != nil {
		return
	}
	cf = d.Close
	return
}

// Close close the resource.
func (d *dao) Close() {
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) error {
	return nil
}
