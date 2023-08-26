package dao

import (
	"context"

	seasonGRPC "git.bilibili.co/bapis/bapis-go/ugc-season/service"

	"github.com/pkg/errors"
)

func (d *Dao) SeasonView(ctx context.Context, seasonID int64) (*seasonGRPC.View, error) {
	req := &seasonGRPC.ViewRequest{SeasonID: seasonID}
	reply, err := d.seasonClient.View(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "%+v", req)
	}
	return reply.GetView(), nil
}
