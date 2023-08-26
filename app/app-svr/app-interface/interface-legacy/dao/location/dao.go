package location

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"

	"github.com/pkg/errors"
)

// Dao is location dao.
type Dao struct {
	// rpc
	locGRPC locgrpc.LocationClient
}

// New new a location dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.locGRPC, err = locgrpc.NewClient(c.LocationGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Info(c context.Context, ipaddr string) (info *locgrpc.InfoReply, err error) {
	if info, err = d.locGRPC.Info(c, &locgrpc.InfoReq{Addr: ipaddr}); err != nil {
		log.Error("%+v", err)
	}
	return
}

// AuthPIDs get auth by pids.
func (d *Dao) AuthPIDs(c context.Context, pids, ipaddr string) (map[int64]*locgrpc.Auth, error) {
	reply, err := d.locGRPC.AuthPIDs(c, &locgrpc.AuthPIDsReq{Pids: pids, IpAddr: ipaddr})
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return reply.Auths, nil
}

func (d *Dao) Info2(ctx context.Context, addr string) (*locgrpc.InfoComplete, error) {
	req := &locgrpc.AddrReq{Addr: addr}
	reply, err := d.locGRPC.Info2(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", req)
	}
	return reply.GetInfo(), nil
}

func (d *Dao) ZoneLimitPolicies(ctx context.Context, req *locgrpc.ZoneLimitPoliciesReq) (*locgrpc.ZoneLimitPoliciesReply, error) {
	return d.locGRPC.ZoneLimitPolicies(ctx, req)

}
