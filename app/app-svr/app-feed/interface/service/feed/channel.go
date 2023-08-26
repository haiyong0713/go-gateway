package feed

import (
	"context"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
	"go-common/library/log"
)

func (s *Service) Channels(c context.Context, channelIDs []int64, mid int64) (res map[int64]*channelgrpc.ChannelCard, err error) {
	var chls []*channelgrpc.ChannelCard
	if chls, err = s.channelDao.Channels(c, channelIDs, mid); err != nil {
		log.Error("%v", err)
		return
	}
	res = make(map[int64]*channelgrpc.ChannelCard)
	for _, chl := range chls {
		if chl != nil {
			res[chl.ChannelId] = chl
		}
	}
	return
}
