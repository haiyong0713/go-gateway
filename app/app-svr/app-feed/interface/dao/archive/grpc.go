package archive

import (
	"context"
	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

func (d *Dao) ArcsPlayer(ctx context.Context, aids []int64, from string) (res map[int64]*arcgrpc.ArcPlayer, err error) {
	batchArg, _ := arcmid.FromContext(ctx)
	duplicateBatchArg := *batchArg
	duplicateBatchArg.From = from
	playAvs := make([]*arcgrpc.PlayAv, 0, len(aids))
	for _, aid := range aids {
		item := &arcgrpc.PlayAv{
			Aid: aid,
		}
		playAvs = append(playAvs, item)
	}
	arg := &arcgrpc.ArcsPlayerRequest{
		PlayAvs:      playAvs,
		BatchPlayArg: &duplicateBatchArg,
	}
	info, err := d.rpcClient.ArcsPlayer(ctx, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.ArcsPlayer
	return
}
