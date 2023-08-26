package archive

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-common/library/log"

	archive "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/app-svr/player-online/internal/conf"
)

type Dao struct {
	c         *conf.Config
	arcClient archive.ArchiveClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		c: c,
	}
	var err error
	if d.arcClient, err = archive.NewClient(c.ArchiveRpc); err != nil {
		panic(fmt.Sprintf("arcClient NewClient error (%+v)", err))
	}
	return d
}

func (d *Dao) SimpleArc(c context.Context, aid int64) (*archive.SimpleArc, error) {
	arg := &archive.SimpleArcRequest{Aid: aid}
	reply, err := d.arcClient.SimpleArc(c, arg)
	if err != nil {
		log.Error("d.SimpleArc arg(%v) error(%+v)", arg, err)
		return nil, err
	}
	if reply.Arc == nil {
		return nil, ecode.NothingFound
	}
	return reply.GetArc(), nil
}
