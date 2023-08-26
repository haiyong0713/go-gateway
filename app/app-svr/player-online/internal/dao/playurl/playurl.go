package playurl

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/player-online/internal/conf"

	"go-common/library/log"

	psClient "git.bilibili.co/bapis/bapis-go/playurl/service"
)

const (
	_total = "total"
)

type Dao struct {
	c        *conf.Config
	psClient psClient.PlayURLClient
}

// New playurl dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.psClient, err = psClient.NewClient(nil); err != nil {
		panic(fmt.Sprintf("playurl-service NewClient error (%+v)", err))
	}
	return
}

func (d *Dao) PlayOnline(ctx context.Context, aid, cid int64) (int64, error) {
	res, err := d.psClient.PlayOnline(ctx, &psClient.PlayOnlineReq{Aid: aid, Cid: cid})
	if err != nil {
		log.Error("PlayOnline aid(%d), cid(%d) error(%+v)", aid, cid, err)
		return 0, err
	}
	if res.IsHide {
		return 0, nil
	}
	return res.Count[_total], nil
}
