package v1

import (
	"context"

	ugcSeasonGrpc "go-gateway/app/app-svr/ugc-season/service/api"

	"github.com/pkg/errors"
)

func (d *dao) SeasonView(ctx context.Context, req *ugcSeasonGrpc.ViewRequest) (*ugcSeasonGrpc.ViewReply, error) {
	reply, err := d.ugcSeasonClient.View(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "req=%+v", req)
	}
	return reply, nil
}
