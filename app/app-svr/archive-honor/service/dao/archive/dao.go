package archive

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/archive-honor/service/conf"
	"go-gateway/app/app-svr/archive/service/api"

	"github.com/pkg/errors"
)

// Dao is archive-honor dao
type Dao struct {
	c         *conf.Config
	arcClient api.ArchiveClient
}

// New new a Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
	}
	var err error
	if d.arcClient, err = api.NewClient(c.ArcClient); err != nil {
		panic(fmt.Sprintf("NewArchiveClient err(%+v)", err))
	}
	return
}

// Arc get Archive info
func (d *Dao) Arc(c context.Context, aid int64) (*api.Arc, error) {
	ArcReq := &api.ArcRequest{Aid: aid}
	ArcReply, err := d.arcClient.Arc(c, ArcReq)
	if err != nil {
		return nil, err
	}
	if ArcReply == nil || ArcReply.Arc == nil {
		return nil, errors.New("no arc replay")
	}
	return ArcReply.Arc, nil
}
