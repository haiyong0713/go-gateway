package dao

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

const (
	_videoChannel = 3
)

// ResourceChannels .
func (d *dao) ResourceChannels(c context.Context, aids []int64, mid int64) (res map[int64][]*channelgrpc.Channel, err error) {
	var (
		g     = errgroup.WithContext(c)
		mutex = sync.Mutex{}
	)
	res = map[int64][]*channelgrpc.Channel{}
	for _, v := range aids {
		aid := v
		g.Go(func(ctx context.Context) (err error) {
			reply, err := d.channelClient.ResourceChannels(ctx, &channelgrpc.ResourceChannelsReq{Rid: aid, Mid: mid, Type: _videoChannel})
			if err != nil {
				log.Error("d.grpcClient.ResourceChannels(%d,%d) error(%v)", aid, mid, err)
				return
			}
			if reply != nil {
				mutex.Lock()
				res[aid] = reply.Channels
				mutex.Unlock()
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		log.Error("%+v", err)
	}
	return
}
