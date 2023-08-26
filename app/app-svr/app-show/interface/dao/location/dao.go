package location

import (
	"context"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-show/interface/conf"

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
	if d.locGRPC, err = locgrpc.NewClient(c.LocationGRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) Info(c context.Context, ipaddr string) (info *locgrpc.InfoReply, err error) {
	if info, err = d.locGRPC.Info(c, &locgrpc.InfoReq{Addr: ipaddr}); err != nil {
		log.Error("d.locGRPC.Info %v", err)
	}
	return
}

func (d *Dao) AuthPIDs(c context.Context, pids, ipaddr string) (res map[string]*locgrpc.Auth, err error) {
	var auths *locgrpc.AuthPIDsReply
	if auths, err = d.locGRPC.AuthPIDs(c, &locgrpc.AuthPIDsReq{Pids: pids, IpAddr: ipaddr}); err != nil {
		log.Error("%v", err)
		return
	}
	if auths == nil {
		return
	}
	res = make(map[string]*locgrpc.Auth)
	for pid, auth := range auths.Auths {
		p := strconv.FormatInt(pid, 10)
		res[p] = auth
	}
	return
}
