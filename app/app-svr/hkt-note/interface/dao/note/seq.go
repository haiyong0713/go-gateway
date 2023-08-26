package note

import (
	"context"

	seqgrpc "git.bilibili.co/bapis/bapis-go/infra/service/sequence"
)

func (d *Dao) SeqId(ctx context.Context) (int64, error) {
	args := &seqgrpc.BusinessReq{
		Business: d.c.NoteCfg.Seq.BusinessId,
		Token:    d.c.NoteCfg.Seq.Token,
	}
	reply, err := d.seqClient.SnowFlake(ctx, args)
	if err != nil {
		return 0, err
	}
	return reply.GetId(), nil
}
