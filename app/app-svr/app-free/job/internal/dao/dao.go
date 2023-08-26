package dao

import (
	"context"

	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-free/job/internal/model"

	location "git.bilibili.co/bapis/bapis-go/community/service/location"

	"github.com/google/wire"
	"github.com/pkg/errors"
)

var Provider = wire.NewSet(New, NewDB)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	AllFreeRecords(ctx context.Context) (res []*model.FreeRecord, err error)
	Info(c context.Context, ipaddr string) (res *location.InfoCompleteReply, err error)
}

// dao dao.
type dao struct {
	db *sql.DB
	// rpc
	locGRPC location.LocationClient
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// New new a dao and return.
func New(db *sql.DB) (d Dao, cf func(), err error) {
	return newDao(db)
}

func newDao(db *sql.DB) (d *dao, cf func(), err error) {
	d = &dao{
		db: db,
	}
	cf = d.Close
	var (
		grpc struct {
			LocationGRPC *warden.ClientConfig
		}
	)
	checkErr(paladin.Get("grpc.toml").UnmarshalTOML(&grpc))
	if d.locGRPC, err = location.NewClient(grpc.LocationGRPC); err != nil {
		panic(err)
	}
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

func (d *dao) AllFreeRecords(ctx context.Context) (res []*model.FreeRecord, err error) {
	return d.RawAllFreeRecords(ctx)
}

func (d *dao) Info(c context.Context, ipaddr string) (res *location.InfoCompleteReply, err error) {
	if res, err = d.locGRPC.Info2(c, &location.AddrReq{Addr: ipaddr}); err != nil {
		err = errors.Wrap(err, ipaddr)
	}
	return
}
