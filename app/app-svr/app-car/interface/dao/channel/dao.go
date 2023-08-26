package channel

import (
	"context"
	"fmt"

	changrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	"go-gateway/app/app-svr/app-car/interface/conf"

	"github.com/pkg/errors"
)

const (
	TypWeb = 1
)

// Dao channel dao.
type Dao struct {
	c        *conf.Config
	chClient changrpc.ChannelRPCClient
}

// New new web channel  dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.chClient, err = changrpc.NewClient(c.ChannelClient); err != nil {
		panic(fmt.Sprintf("New ChannelRPCClient error (%+v)", err))
	}
	return
}

func (d *Dao) ResourceList(ctx context.Context, req *changrpc.ResourceListReq) (*changrpc.ResourceListReply, error) {
	req.Typ = TypWeb
	//和ChannelResourceList的区别是啥？
	reply, err := d.chClient.ResourceList(ctx, req)
	if err != nil {
		err = errors.Wrapf(err, "d.chClient.ResourceList(%+v)", req)
		return nil, err
	}
	return reply, nil
}
