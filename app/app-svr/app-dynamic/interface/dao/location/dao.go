package location

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	"go-common/library/sync/errgroup.v2"
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

func (d *Dao) InfoGRPC(c context.Context, ipaddr string) (info *locgrpc.InfoReply, err error) {
	if info, err = d.locGRPC.Info(c, &locgrpc.InfoReq{Addr: ipaddr}); err != nil {
		log.Error("%v", err)
	}
	return
}

func (d *Dao) IP2Location(c context.Context, ipaddrs map[string]struct{}) (map[string]*locgrpc.InfoComplete, error) {
	mu := sync.Mutex{}
	eg := errgroup.WithCancel(c)
	eg.GOMAXPROCS(20)
	ret := make(map[string]*locgrpc.InfoComplete)
	// 放进去一个空地址兜底
	ipaddrs[""] = struct{}{}
	for ip := range ipaddrs {
		ipaddr := ip
		eg.Go(func(ctx context.Context) error {
			resp, err := d.locGRPC.Info2Special(ctx, &locgrpc.AddrReq{Addr: ipaddr})
			if err != nil {
				return err
			}
			if resp.GetInfo() == nil {
				log.Warnc(ctx, "unexpected nil Info2Special.Info with addr(%s)", ipaddr)
				return nil
			}
			mu.Lock()
			ret[ipaddr] = resp.GetInfo()
			mu.Unlock()
			return nil
		})
	}
	return ret, eg.Wait()
}
