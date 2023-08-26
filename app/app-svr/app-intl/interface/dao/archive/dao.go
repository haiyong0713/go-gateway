package archive

import (
	"context"
	"fmt"
	"sync"

	"go-common/library/log"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"

	"go-common/library/sync/errgroup.v2"

	hisgrpc "git.bilibili.co/bapis/bapis-go/community/interface/history"
	"go-gateway/app/app-svr/app-intl/interface/conf"
	"go-gateway/app/app-svr/app-intl/interface/model/view"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

const (
	_maxAids = 100
)

// Dao is archive dao.
type Dao struct {
	// http client
	client        *bm.Client
	realteURL     string
	commercialURL string
	relateRecURL  string
	playURL       string
	//grpc
	rpcClient arcgrpc.ArchiveClient
	hisClient hisgrpc.HistoryClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		// http client
		client:        bm.NewClient(c.HTTPWrite, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
		realteURL:     c.Host.Data + _realteURL,
		commercialURL: c.Host.APICo + _commercialURL,
		relateRecURL:  c.HostDiscovery.Data + _relateRecURL,
		playURL:       c.Host.Bvcvod + _playURL,
	}
	//grpc
	var err error
	if d.rpcClient, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.hisClient, err = hisgrpc.NewClient(c.HistoryGRPC); err != nil {
		panic(fmt.Sprintf("his rpcClient NewClientt error (%+v)", err))
	}
	return
}

// Ping ping check memcache connection
func (d *Dao) Ping(c context.Context) (err error) {
	return nil
}

// Archives multi get archives.
func (d *Dao) Archives(c context.Context, aids []int64) (am map[int64]*api.Arc, err error) {
	if len(aids) == 0 {
		return
	}
	am = map[int64]*api.Arc{}
	mutexArcs := sync.Mutex{}
	g := errgroup.WithContext(c)
	if arcsLen := len(aids); arcsLen > 0 {
		for i := 0; i < arcsLen; i += _maxAids {
			var partAids []int64
			if i+_maxAids > arcsLen {
				partAids = aids[i:]
			} else {
				partAids = aids[i : i+_maxAids]
			}
			g.Go(func(ctx context.Context) (err error) {
				var (
					tmpRes *api.ArcsReply
					arg    = &api.ArcsRequest{Aids: partAids}
				)
				if tmpRes, err = d.rpcClient.Arcs(ctx, arg); err != nil {
					log.Error("Archives aids(%v) d.rpcClient.Arcs error(%v)", aids, err)
					return
				}
				if len(tmpRes.GetArcs()) > 0 {
					mutexArcs.Lock()
					for aid, stat := range tmpRes.GetArcs() {
						am[aid] = stat
					}
					mutexArcs.Unlock()
				}
				return
			})
		}
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
		return
	}
	return am, err
}

// Archive get archive mc->rpc.
func (d *Dao) Archive(c context.Context, aid int64) (a *api.Arc, err error) {
	reply, err := d.rpcClient.Arc(c, &arcgrpc.ArcRequest{Aid: aid})
	if err != nil {
		return nil, err
	}
	return reply.GetArc(), nil
}

// Progress is  archive plays progress .
func (d *Dao) Progress(c context.Context, aid, mid int64) (h *view.History, err error) {
	arg := &hisgrpc.ProgressReq{Mid: mid, Aids: []int64{aid}}
	his, err := d.hisClient.Progress(c, arg)
	if err != nil {
		log.Error("d.hisRPC.Progress(%v) error(%v)", arg, err)
		return
	}
	if his != nil {
		if resVal, ok := his.Res[aid]; ok && resVal != nil {
			h = &view.History{Cid: resVal.Cid, Progress: resVal.Pro}
		}
	}
	return
}
