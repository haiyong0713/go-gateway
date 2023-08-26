package channel

import (
	"context"

	"go-common/library/log"

	channelgrpc "git.bilibili.co/bapis/bapis-go/community/interface/channel"
)

func (d *Dao) Infos(c context.Context, channelIDs []int64, mid int64) (res map[int64]*channelgrpc.Channel, err error) {
	if len(channelIDs) == 0 {
		return map[int64]*channelgrpc.Channel{}, nil
	}
	var (
		args  = &channelgrpc.InfosReq{Cids: channelIDs, Mid: mid}
		infos *channelgrpc.InfosReply
	)
	if infos, err = d.grpcClient.Infos(c, args); err != nil {
		log.Error("%v", err)
		return
	}
	res = infos.GetCidMap()
	return
}
