package dao

import (
	"context"

	camp "git.bilibili.co/bapis/bapis-go/video/vod/playurlcamp"

	"github.com/pkg/errors"
)

func (d *dao) ProtobufPlayurl(ctx context.Context, in *camp.RequestMsg) (*camp.ResponseMsg, error) {
	reply, err := d.campCli.ProtobufPlayurl(ctx, in)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", in)
	}
	return reply, nil
}
