package seq

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-wall/job/conf"

	seqgrpc "git.bilibili.co/bapis/bapis-go/infra/service/sequence"
)

type Dao struct {
	c        *conf.Config
	seqGRPC  seqgrpc.SeqClient
	business int64
	token    string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:        c,
		business: c.Seq.BusinessID,
		token:    c.Seq.Token,
	}
	var err error
	if d.seqGRPC, err = seqgrpc.NewClient(c.SeqGRPC); err != nil {
		panic(err)
	}
	return
}

// SeqID is.
func (d *Dao) SeqID(ctx context.Context) (id int64, err error) {
	args := &seqgrpc.BusinessReq{
		Business: d.business,
		Token:    d.token,
	}
	reply, err := d.seqGRPC.SnowFlake(ctx, args)
	if err != nil {
		log.Error("%+v", err)
		return
	}
	id = reply.GetId()
	return
}
