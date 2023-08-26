package archive

import (
	"context"
	"fmt"
	"math"
	"sync"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/interface/conf"
	iter "go-gateway/app/app-svr/app-car/interface/pkg"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	"go-gateway/app/app-svr/archive/service/api"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	flowctrlgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"
	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	c *conf.Config
	// grpc
	rpcClient         arcgrpc.ArchiveClient
	flowControlClient flowctrlgrpc.FlowControlClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	// grpc
	var err error
	if d.rpcClient, err = arcgrpc.NewClient(c.ArchiveGRPC); err != nil {
		panic(fmt.Sprintf("rpcClient NewClientt error (%+v)", err))
	}
	if d.flowControlClient, err = flowctrlgrpc.NewClient(c.FlowControlGRPC); err != nil {
		panic(fmt.Sprintf("flowControlClient NewClient error (%+v)", err))
	}
	return
}

// Archives get archive.
func (d *Dao) Archives(c context.Context, aids []int64) (map[int64]*api.Arc, error) {
	reply, err := d.rpcClient.Arcs(c, &arcgrpc.ArcsRequest{Aids: aids})
	if err != nil {
		return nil, err
	}
	return reply.GetArcs(), nil
}

func (d *Dao) Views(c context.Context, aids []int64) (map[int64]*arcgrpc.ViewReply, error) {
	arg := &arcgrpc.ViewsRequest{Aids: aids}
	viewReply, err := d.rpcClient.Views(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return viewReply.GetViews(), nil
}

func (d *Dao) View(c context.Context, aid int64) (*api.ViewReply, error) {
	arg := &api.ViewRequest{Aid: aid}
	reply, err := d.rpcClient.View(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply, nil
}

func (d *Dao) ViewsAll(c context.Context, aids []int64) (map[int64]*arcgrpc.ViewReply, error) {
	const (
		_max = 50
	)
	var (
		forNum     = int(math.Ceil(float64(len(aids)) / float64(_max)))
		mutex      = sync.Mutex{}
		start, end int
	)
	res := map[int64]*arcgrpc.ViewReply{}
	g := errgroup.WithContext(c)
	for i := 0; i < forNum; i++ {
		start = i * _max
		end = start + _max
		var (
			tmpaids []int64
		)
		if len(aids) >= end {
			tmpaids = aids[start:end]
		} else if len(aids) < end {
			tmpaids = aids[start:]
		} else if len(aids) < start {
			break
		}
		g.Go(func(cc context.Context) error {
			reply, err := d.Views(cc, tmpaids)
			if err != nil {
				log.Error("d.Views(%v) error(%v)", tmpaids, err)
				return err
			}
			if reply != nil {
				mutex.Lock()
				for aid, v := range reply {
					res[aid] = v
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) ArcsPlayer(c context.Context, aids []int64) (map[int64]*arcgrpc.ArcPlayer, error) {
	batchArg, _ := arcmid.FromContext(c)
	arg := &arcgrpc.ArcsPlayerRequest{
		BatchPlayArg: batchArg,
	}
	for _, aid := range aids {
		arg.PlayAvs = append(arg.PlayAvs, &api.PlayAv{Aid: aid})
	}
	info, err := d.rpcClient.ArcsPlayer(c, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return info.ArcsPlayer, nil
}

func (d *Dao) ArcsPlayerAll(c context.Context, aids []int64) (map[int64]*arcgrpc.ArcPlayer, error) {
	const (
		_max = 50
	)
	var (
		forNum     = int(math.Ceil(float64(len(aids)) / float64(_max)))
		mutex      = sync.Mutex{}
		start, end int
	)
	res := map[int64]*arcgrpc.ArcPlayer{}
	g := errgroup.WithContext(c)
	for i := 0; i < forNum; i++ {
		start = i * _max
		end = start + _max
		var (
			tmpaids []int64
		)
		if len(aids) >= end {
			tmpaids = aids[start:end]
		} else if len(aids) < end {
			tmpaids = aids[start:]
		} else if len(aids) < start {
			break
		}
		g.Go(func(cc context.Context) error {
			reply, err := d.ArcsPlayer(cc, tmpaids)
			if err != nil {
				log.Error("d.ArcsPlayer(%v) error(%v)", tmpaids, err)
				return err
			}
			if reply != nil {
				mutex.Lock()
				for aid, v := range reply {
					res[aid] = v
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) ArcsAll(c context.Context, aids []int64) (map[int64]*arcgrpc.Arc, error) {
	const (
		_max = 50
	)
	var (
		forNum     = int(math.Ceil(float64(len(aids)) / float64(_max)))
		mutex      = sync.Mutex{}
		start, end int
	)
	res := map[int64]*arcgrpc.Arc{}
	g := errgroup.WithContext(c)
	for i := 0; i < forNum; i++ {
		start = i * _max
		end = start + _max
		var (
			tmpaids []int64
		)
		if len(aids) >= end {
			tmpaids = aids[start:end]
		} else if len(aids) < end {
			tmpaids = aids[start:]
		} else if len(aids) < start {
			break
		}
		g.Go(func(cc context.Context) error {
			reply, err := d.Archives(cc, tmpaids)
			if err != nil {
				log.Error("d.ArcsPlayer(%v) error(%v)", tmpaids, err)
				return err
			}
			if reply != nil {
				mutex.Lock()
				for aid, v := range reply {
					res[aid] = v
				}
				mutex.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) Arc(c context.Context, aid int64) (*api.Arc, error) {
	arg := &api.ArcRequest{Aid: aid}
	reply, err := d.rpcClient.Arc(c, arg)
	if err != nil {
		log.Error("d.rpcClient.Arc(%v) error(%+v)", arg, err)
		return nil, err
	}
	if reply.Arc == nil {
		return nil, ecode.NothingFound
	}
	return reply.GetArc(), nil
}

func (d *Dao) SimpleArc(c context.Context, aid int64) (*api.SimpleArc, error) {
	arg := &api.SimpleArcRequest{Aid: aid}
	reply, err := d.rpcClient.SimpleArc(c, arg)
	if err != nil {
		log.Error("d.rpcClient.SimpleArc(%v) error(%+v)", arg, err)
		return nil, err
	}
	if reply.Arc == nil {
		return nil, ecode.NothingFound
	}
	return reply.GetArc(), nil
}

func (d *Dao) SimpleArcs(c context.Context, aids []int64) (map[int64]*api.SimpleArc, error) {
	arcs := make(map[int64]*api.SimpleArc)
	if len(aids) == 0 {
		return nil, errors.Wrapf(ecode.RequestErr, "SimpleArcs aids(%v)", aids)
	}
	for _, step := range iter.Steps(len(aids), 50) {
		curAids := aids[step.Head:step.Tail]
		req := &api.SimpleArcsRequest{
			Aids: curAids,
		}
		arcsReply, err := d.rpcClient.SimpleArcs(c, req)
		if err != nil {
			return nil, errors.Wrapf(err, "SimpleArcs aids(%v)", aids)
		}
		if arcsReply == nil {
			return nil, errors.Wrapf(ecode.NothingFound, "SimpleArcs aids(%v)", aids)
		}
		if arcsReply.Arcs != nil {
			for aid, view := range arcsReply.Arcs {
				arcs[aid] = view
			}
		}
	}
	return arcs, nil
}

func (d *Dao) Description(c context.Context, aid int64) (desc string, err error) {
	arg := &api.DescriptionRequest{Aid: aid}
	reply, err := d.rpcClient.Description(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return "", err
	}
	return reply.Desc, nil
}

func (d *Dao) ArcsPlayerV2(c context.Context, aids []*api.PlayAv, showPgcPlayurl bool, from string) (map[int64]*api.ArcPlayer, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*api.ArcPlayer)
	batchArg, _ := arcmid.FromContext(c)
	tmpBatchArg := &api.BatchPlayArg{}
	if batchArg != nil {
		*tmpBatchArg = *batchArg
		tmpBatchArg.ShowPgcPlayurl = showPgcPlayurl
		tmpBatchArg.From = from
	}
	for i := 0; i < len(aids); i += max50 {
		var partAids []*api.PlayAv
		if i+max50 > len(aids) {
			partAids = aids[i:]
		} else {
			partAids = aids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			arg := &api.ArcsPlayerRequest{
				PlayAvs:      partAids,
				BatchPlayArg: tmpBatchArg,
			}
			archives, err := d.rpcClient.ArcsPlayer(ctx, arg)
			if err != nil {
				log.Error("ArcsPlayerV2 partAids(%+v) err(%v)", partAids, err)
				return err
			}
			mu.Lock()
			for aid, arc := range archives.GetArcsPlayer() {
				if arc == nil {
					continue
				}
				if !arc.Arc.IsNormal() {
					continue
				}
				res[aid] = arc
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
