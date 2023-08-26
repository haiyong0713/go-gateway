package channel

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

// Dao is coin dao
type Dao struct {
	channelClient api.ChannelRPCClient
}

// New initial coin dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.channelClient, err = api.NewClient(c.ChannelClient); err != nil {
		panic(err)
	}
	return
}

// ResourceChannels .
func (d *Dao) ResourceChannels(c context.Context, aid, mid, ty int64) (res []*api.Channel, err error) {
	var reply *api.ResourceChannelsReply
	if reply, err = d.channelClient.ResourceChannels(c, &api.ResourceChannelsReq{Rid: aid, Mid: mid, Type: ty}); err != nil {
		log.Error("d.channelClient.ResourceChannels(%d,%d) error(%v)", aid, mid, err)
		return
	}
	if reply != nil {
		res = reply.Channels
	}
	return
}

// ResourceChannels .
func (d *Dao) ChannelHonor(c context.Context, aid int64) (*api.ResourceHonor, error) {
	reply, err := d.channelClient.ResourceHonor(c, &api.ResourceHonorReq{Rids: []int64{aid}})
	if err != nil {
		log.Error("d.channelClient.ResourceHonor(%d) error(%v)", aid, err)
		return nil, err
	}
	honorMap := reply.GetHonorMap()
	if res, ok := honorMap[aid]; ok {
		return res, nil
	}
	return nil, nil
}
