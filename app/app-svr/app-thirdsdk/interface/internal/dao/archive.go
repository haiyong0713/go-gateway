package dao

import (
	"context"

	arc "go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

func (d *dao) Archive(ctx context.Context, aid int64) (*arc.Arc, error) {
	in := &arc.ArcRequest{Aid: aid}
	reply, err := d.arcCli.Arc(ctx, in)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", in)
	}
	return reply.GetArc(), nil
}
