package toview

import (
	"context"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	api "git.bilibili.co/bapis/bapis-go/community/service/toview"

	"github.com/pkg/errors"
)

type Dao struct {
	toViewClient api.ToViewsClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{}
	var err error
	if d.toViewClient, err = api.NewClient(c.ToViewGRPC); err != nil {
		panic(err)
	}
	return d
}

func (d *Dao) LastToViewTime(ctx context.Context, mid int64) (int64, error) {
	reply, err := d.toViewClient.UserToViews(ctx, &api.UserToViewsReq{Mid: mid, BusinessId: 1, Pn: 1, Ps: 1})
	if err != nil {
		return 0, errors.Wrapf(err, "d.toViewClient.UserToViews error mid(%v)", mid)
	}
	if len(reply.Toviews) == 0 || reply.Toviews[0] == nil {
		return 0, ecode.NothingFound
	}
	return reply.Toviews[0].Unix, nil
}
