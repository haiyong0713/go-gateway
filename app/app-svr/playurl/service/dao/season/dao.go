package season

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/playurl/service/conf"
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
func (d *Dao) Season(c context.Context, seasonID int64) (*api.Season, error) {
	var (
		req = &api.SeasonRequest{SeasonID: seasonID}
	)
	reply, err := d.seasonClient.Season(c, req)
	if err != nil {
		err = errors.Wrapf(err, "%+v", req)
		return nil, err
	}
	return reply.GetSeason(), nil
}
