package channel

import (
	"context"
	"sync"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-car/job/conf"

	api "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

const (
	_videoChannel = 3
)

type Dao struct {
	channelClient api.ChannelRPCClient
}

// New initial coin dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.channelClient, err = api.NewClient(nil); err != nil {
		panic(err)
	}
	return
}

// ResourceChannels .
func (d *Dao) ResourceChannels(c context.Context, aid int64) ([]*api.Channel, error) {
	reply, err := d.channelClient.ResourceChannels(c, &api.ResourceChannelsReq{Rid: aid, Type: _videoChannel})
	if err != nil {
		log.Error("d.channelClient.ResourceChannels(%d) error(%v)", aid, err)
		return nil, err
	}
	return reply.GetChannels(), nil
}

func (d *Dao) ResourceChannelsAll(c context.Context, aids []int64) (map[int64][]*api.Channel, error) {
	res := map[int64][]*api.Channel{}
	g := errgroup.WithContext(c)
	mu := sync.Mutex{}
	for _, v := range aids {
		aid := v
		g.Go(func(ctx context.Context) (err error) {
			reply, err := d.channelClient.ResourceChannels(ctx, &api.ResourceChannelsReq{Rid: aid, Type: _videoChannel})
			if err != nil {
				log.Error("ResourceChannelsAll d.channelClient.ResourceChannels(%d) error(%v)", aid, err)
				return err
			}
			mu.Lock()
			res[aid] = reply.GetChannels()
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return res, nil
}
