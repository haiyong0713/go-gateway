package archive

import (
	"context"

	arcmid "go-gateway/app/app-svr/archive/middleware"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

// ArcsPlayer  .
func (d *Dao) ArcsPlayer(c context.Context, aids []*arcgrpc.PlayAv) (map[int64]*arcgrpc.ArcPlayer, error) {
	batchArg, _ := arcmid.FromContext(c)
	arg := &arcgrpc.ArcsPlayerRequest{
		PlayAvs:      aids,
		BatchPlayArg: batchArg,
	}
	info, err := d.rpcClient.ArcsPlayer(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return nil, err
	}
	return info.GetArcsPlayer(), nil
}
