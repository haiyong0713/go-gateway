package season

import (
	"context"
	"fmt"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-view/interface/conf"
	"go-gateway/app/app-svr/ugc-season/service/api"

	"github.com/pkg/errors"
)

// Dao is archive dao.
type Dao struct {
	seasonClient api.UGCSeasonClient
}

// New new a archive dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.seasonClient, err = api.NewClient(c.UGCSeasonClient); err != nil {
		panic(fmt.Sprintf("ugc-season NewClient not found err(%v)", err))
	}
	return
}

// Season def.
func (d *Dao) Season(c context.Context, seasonID int64) (season *api.View, err error) {
	var (
		req   = &api.ViewRequest{SeasonID: seasonID}
		reply *api.ViewReply
	)
	if reply, err = d.seasonClient.View(c, req); err != nil {
		err = errors.Wrapf(err, "%+v", req)
		return
	}
	season = reply.View
	return
}

func (d *Dao) SeasonInfo(c context.Context, seasonID int64) (*api.SeasonReply, error) {
	var (
		req   = &api.SeasonRequest{SeasonID: seasonID}
		reply *api.SeasonReply
		err   error
	)
	if reply, err = d.seasonClient.Season(c, req); err != nil {
		return nil, errors.Wrapf(ecode.Error(ecode.String(err.Error()), "服务开小差了，请稍后重试~"), "req:%+v err:%+v", req, err)
	}
	return reply, nil
}
