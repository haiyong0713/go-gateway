package location

import (
	"context"

	"go-common/library/conf/paladin"
	"go-common/library/net/rpc/warden"

	location "git.bilibili.co/bapis/bapis-go/community/service/location"

	"github.com/pkg/errors"
)

// Dao dao interface
type Dao interface {
	Close()
	Ping(ctx context.Context) (err error)
	Infos(c context.Context, ips []string) (res map[string]*location.InfoComplete, err error)
	Info(c context.Context, ipaddr string) (res *location.InfoCompleteReply, err error)
}

// Dao is location dao.
type dao struct {
	// rpc
	locGRPC location.LocationClient
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// New new a location dao.
func New() Dao {
	var (
		grpc struct {
			LocationGRPC *warden.ClientConfig
		}
	)
	checkErr(paladin.Get("grpc.toml").UnmarshalTOML(&grpc))
	d := &dao{}
	var err error
	if d.locGRPC, err = location.NewClient(grpc.LocationGRPC); err != nil {
		panic(err)
	}
	return d
}

// Close close the resource.
func (d *dao) Close() {
}

// Ping ping the resource.
func (d *dao) Ping(ctx context.Context) (err error) {
	return
}

func (d *dao) Infos(c context.Context, ips []string) (res map[string]*location.InfoComplete, err error) {
	reply, err := d.locGRPC.Infos2(c, &location.AddrsReq{Addrs: ips})
	if err != nil {
		err = errors.Wrapf(err, "%v", ips)
		return
	}
	res = reply.Infos
	return
}

func (d *dao) Info(c context.Context, ipaddr string) (res *location.InfoCompleteReply, err error) {
	if res, err = d.locGRPC.Info2(c, &location.AddrReq{Addr: ipaddr}); err != nil {
		err = errors.Wrap(err, ipaddr)
	}
	return
}
