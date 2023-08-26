package resource

import (
	"context"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	"github.com/pkg/errors"
)

func (s *Service) ResourceChannels(c context.Context, mid, aid, typ int64) (*channelgrpc.ResourceChannelsReply, error) {
	var (
		err error
		req = &channelgrpc.ResourceChannelsReq{
			Rid:  aid,
			Mid:  mid,
			Type: typ,
		}
		reply *channelgrpc.ResourceChannelsReply
	)
	if reply, err = s.channelClient.ResourceChannels(c, req); err != nil {
		err = errors.Wrapf(err, "s.channelClient.ResourceChannels(%+v)", req)
		return nil, err
	}
	return reply, nil
}
