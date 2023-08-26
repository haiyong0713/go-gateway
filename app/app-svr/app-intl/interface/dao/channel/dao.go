package channel

import (
	"context"

	"git.bilibili.co/bapis/bapis-go/community/interface/channel"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-intl/interface/conf"
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
