package channel

import (
	"context"
	"sync"

	"go-common/library/log"

	"go-common/library/sync/errgroup.v2"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

type Dao struct {
	c          *conf.Config
	grpcClient channelgrpc.ChannelRPCClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.grpcClient, err = channelgrpc.NewClient(c.ChannelClient); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) ChannelInfos(c context.Context, tids []int64) (map[int64]*channelgrpc.Channel, error) {
	var max50 = 50
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	res := make(map[int64]*channelgrpc.Channel)
	for i := 0; i < len(tids); i += max50 {
		var partTids []int64
		if i+max50 > len(tids) {
			partTids = tids[i:]
		} else {
			partTids = tids[i : i+max50]
		}
		g.Go(func(ctx context.Context) (err error) {
			cis, err := d.ChannelInfosSlick(ctx, partTids)
			if err != nil {
				return err
			}
			mu.Lock()
			for tid, ci := range cis {
				res[tid] = ci
			}
			mu.Unlock()
			return
		})
	}
	if err := g.Wait(); err != nil {
		log.Error("ChannelInfos tids(%+v) eg.wait(%+v)", tids, err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) ChannelInfosSlick(c context.Context, tids []int64) (res map[int64]*channelgrpc.Channel, err error) {
	args := &channelgrpc.InfosReq{Cids: tids}
	var details *channelgrpc.InfosReply
	if details, err = d.grpcClient.Infos(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = details.GetCidMap()
	return
}
