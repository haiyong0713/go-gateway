package pgc

import (
	"context"

	pgcFollowGrpc "git.bilibili.co/bapis/bapis-go/pgc/service/follow"
	"github.com/pkg/errors"
)

func (d *Dao) MyRelations(ctx context.Context, mid int64, withAnime, withCinema bool) (*pgcFollowGrpc.MyRelationsReply, error) {
	req := &pgcFollowGrpc.MyRelationsReq{
		Mid:        mid,
		WithAnime:  withAnime,
		WithCinema: withCinema,
	}
	ret, err := d.pgcFollowGRPC.MyRelations(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ret, err
}
