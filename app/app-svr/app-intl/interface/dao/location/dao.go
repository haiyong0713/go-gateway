package location

import (
	"context"
	"strconv"

	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-intl/interface/conf"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
)

// Dao is location dao.
type Dao struct {
	locGRPC locgrpc.LocationClient
}

// New new a location dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.locGRPC, err = locgrpc.NewClient(c.LocationClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Info(c context.Context, ipaddr string) (info *locgrpc.InfoReply, err error) {
	if info, err = d.locGRPC.Info(c, &locgrpc.InfoReq{Addr: ipaddr}); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) AuthPIDs(c context.Context, pids, ipaddr string) (res map[string]*locgrpc.Auth, err error) {
	var auths *locgrpc.AuthPIDsReply
	if auths, err = d.locGRPC.AuthPIDs(c, &locgrpc.AuthPIDsReq{Pids: pids, IpAddr: ipaddr}); err != nil {
		log.Error("%v", err)
		return
	}
	res = make(map[string]*locgrpc.Auth)
	for pid, auth := range auths.Auths {
		p := strconv.FormatInt(pid, 10)
		res[p] = auth
	}
	return
}

func (d *Dao) Archive(c context.Context, aid, mid int64, ipaddr, cndip string) (auth *locgrpc.Auth, err error) {
	var archive *locgrpc.ArchiveReply
	if archive, err = d.locGRPC.Archive(c, &locgrpc.ArchiveReq{Aid: aid, Mid: mid, IpAddr: ipaddr, CdnAddr: cndip}); err != nil {
		log.Error("%v", err)
		return
	}
	if archive != nil {
		auth = archive.Auth
	}
	return
}

func (d *Dao) Info2(c context.Context) (*locgrpc.InfoComplete, error) {
	req := &locgrpc.AddrReq{Addr: metadata.String(c, metadata.RemoteIP)}
	reply, err := d.locGRPC.Info2(c, req)
	if err != nil {
		log.Error("d.locGRPC.Info2 err(%+v) req(%+v)", err, req)
		return nil, err
	}
	return reply.GetInfo(), nil
}
