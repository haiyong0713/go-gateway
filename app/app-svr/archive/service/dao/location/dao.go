package location

import (
	"context"
	"go-common/component/metadata/network"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/archive/service/conf"

	"go-common/library/log"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
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

func (d *Dao) ArchiveAuthBatch(c context.Context, aids []int64) (map[int64]*locgrpc.Auth, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	nw, _ := network.FromContext(c)
	reply, err := d.locGRPC.ArchiveAuths(c, &locgrpc.ArchiveAuthsReq{Aids: aids, UserIp: ip, CdnIp: nw.WebcdnIP})
	if err != nil {
		log.Error("d.locGRPC.ArchiveAuths is err %v", err)
		return nil, err
	}
	return reply.GetAuths(), nil
}

func (d *Dao) Info2(c context.Context, ip string) (*locgrpc.InfoComplete, error) {
	req := &locgrpc.AddrReq{Addr: ip}
	reply, err := d.locGRPC.Info2Special(c, req)
	if err != nil {
		log.Error("d.locGRPC.Info2 error req(%+v) err(%+v)", req, err)
		return nil, err
	}
	return reply.GetInfo(), nil
}
