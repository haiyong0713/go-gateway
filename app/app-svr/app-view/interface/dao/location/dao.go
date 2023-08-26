package location

import (
	"context"
	"go-common/component/metadata/network"
	"strconv"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-view/interface/conf"

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

func (d *Dao) AuthPIDs(c context.Context, pids, ipaddr string) (res map[string]*locgrpc.Auth, err error) {
	reply, err := d.locGRPC.AuthPIDs(c, &locgrpc.AuthPIDsReq{Pids: pids, IpAddr: ipaddr})
	if err != nil {
		log.Error("%v", err)
		return
	}
	res = make(map[string]*locgrpc.Auth, len(reply.Auths))
	for pid, auth := range reply.Auths {
		p := strconv.FormatInt(pid, 10)
		res[p] = auth
	}
	return
}

// Archive get auth by aid.
func (d *Dao) Archive(c context.Context, aid, mid int64, ipaddr, cdnip string) (auth *locgrpc.Auth, err error) {
	reply, err := d.locGRPC.Archive(c, &locgrpc.ArchiveReq{Aid: aid, Mid: mid, IpAddr: ipaddr, CdnAddr: cdnip})
	if err != nil {
		log.Error("%v", err)
		return
	}
	auth = reply.Auth
	return
}

// Archive get auth by aid.
func (d *Dao) Info2(c context.Context) (*locgrpc.InfoComplete, error) {
	req := &locgrpc.AddrReq{Addr: metadata.String(c, metadata.RemoteIP)}
	reply, err := d.locGRPC.Info2(c, req)
	if err != nil {
		log.Error("d.locGRPC.Info2 err(%+v) req(%+v)", err, req)
		return nil, err
	}
	return reply.GetInfo(), nil
}

func (d *Dao) GetGroups(c context.Context, groupId []int64) (map[int64]*locgrpc.Auth, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	nw, _ := network.FromContext(c)
	req := &locgrpc.GroupsReq{
		Gids:    groupId,
		IpAddr:  ip,
		CdnAddr: nw.WebcdnIP,
	}
	reply, err := d.locGRPC.Groups(c, req)
	if err != nil {
		return nil, err
	}
	return reply.GetAuths(), nil
}
