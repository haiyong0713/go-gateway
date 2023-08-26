package playurl

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-view/interface/conf"

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

func (d *Dao) PlayOnlineTotal(ctx context.Context, aid, cid int64) (int64, bool) {
	res, err := d.psClient.PlayOnline(ctx, &psClient.PlayOnlineReq{Aid: aid, Cid: cid})
	if err != nil {
		log.Error("PlayOnline error(%+v) aid(%d), cid(%d)", err, aid, cid)
		return 0, false
	}
	if res.IsHide {
		return 0, false
	}
	return res.Count[_total], true
}
