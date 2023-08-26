package location

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/location"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

// Dao is location dao.
type Dao struct {
	// grpc
	locGRPC locgrpc.LocationClient
}

// New new a location dao.
func New(c *conf.Config) (d *Dao) {
	g, err := locgrpc.NewClient(c.LocationGRPC)
	if err != nil {
		panic(err)
	}
	d = &Dao{
		// grpc
		locGRPC: g,
	}
	return
}

// Info get ipinfo.
func (d *Dao) Info(c context.Context, ipaddr string) (info *location.Info, err error) {
	var ir *locgrpc.InfoReply
	if ir, err = d.locGRPC.Info(c, &locgrpc.InfoReq{Addr: ipaddr}); err != nil {
		return
	}
	info = &location.Info{
		Addr:        ir.Addr,
		Country:     ir.Country,
		Province:    ir.Province,
		City:        ir.City,
		ISP:         ir.Isp,
		ZoneId:      ir.ZoneId,
		CountryCode: ir.CountryCode,
	}
	return
}

// AuthPIDs get auth by pids.
func (d *Dao) AuthPIDs(c context.Context, pids, ipaddr string) (auths map[int64]*locgrpc.Auth, err error) {
	reply, err := d.locGRPC.AuthPIDs(c, &locgrpc.AuthPIDsReq{Pids: pids, IpAddr: ipaddr})
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return reply.Auths, nil
}

// RawAuthPIDs get auth by pids.
func (d *Dao) RawAuthPIDs(c context.Context, req *locgrpc.AuthPIDsReq) (map[int64]*locgrpc.Auth, error) {
	reply, err := d.locGRPC.AuthPIDs(c, req)
	if err != nil {
		log.Error("%v", err)
		return nil, err
	}
	return reply.Auths, nil
}

func (d *Dao) ZoneLimitPolicies(ctx context.Context, req *locgrpc.ZoneLimitPoliciesReq) (*locgrpc.ZoneLimitPoliciesReply, error) {
	return d.locGRPC.ZoneLimitPolicies(ctx, req)
}

// InfoComplete get ipinfo.
func (d *Dao) InfoComplete(c context.Context, ipaddr string) (info *locgrpc.InfoComplete, err error) {
	if reply, err := d.locGRPC.InfoComplete(c, &locgrpc.InfoCompleteReq{Addr: ipaddr}); err == nil && reply != nil {
		info = reply.Info
	}
	return
}

// ZlimitInfo.
func (d *Dao) ZlimitInfo(c context.Context, ipaddr []string, gids []int64) (*locgrpc.ZlimitInfoReply, error) {
	rly, err := d.locGRPC.ZlimitInfo2(c, &locgrpc.ZlimitInfoReq{Gids: gids, Addrs: ipaddr})
	if err != nil {
		log.Error("d.locGRPC.ZlimitInfo2(%v,%v) error(%v)", ipaddr, gids, err)
		return nil, err
	}
	return rly, nil
}
