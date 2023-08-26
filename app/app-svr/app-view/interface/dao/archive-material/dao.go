package archive_material

import (
	"context"
	"fmt"
	"go-common/library/ecode"

	"github.com/pkg/errors"

	"go-gateway/app/app-svr/app-view/interface/conf"

	api "git.bilibili.co/bapis/bapis-go/material/interface"
)

type Dao struct {
	archiveMaterial api.MaterialClient
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.archiveMaterial, err = api.NewClient(c.ArchiveMaterialClient); err != nil {
		panic(fmt.Sprintf("archiveMaterial NewClient not found err(%v)", err))
	}
	return
}

func (d *Dao) GetPlayerTag(ctx context.Context, aid int64) (*api.StoryPlayerRes, error) {
	req := &api.StoryPlayerReq{
		Avid: aid,
	}
	res, err := d.archiveMaterial.GetStoryPlayer(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "d.archiveMaterial.GetStoryPlayer is err:%+v", req)
	}
	if res == nil || res.Avid <= 0 {
		return nil, ecode.NothingFound
	}
	return res, nil
}
