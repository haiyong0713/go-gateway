package archive_extra

import (
	"context"
	"fmt"

	"go-gateway/app/app-svr/app-view/interface/conf"
	aeapi "go-gateway/app/app-svr/archive-extra/service/api"

	"github.com/pkg/errors"
)

// Dao is archive-honor dao
type Dao struct {
	// grpc
	aeClient aeapi.ArcExtraClient
}

// New is
func New(c *conf.Config) (d *Dao) {
	d = &Dao{}
	var err error
	if d.aeClient, err = aeapi.NewClient(c.ArchiveHonorClient); err != nil {
		panic(fmt.Sprintf("archive honor Client not found err(%v)", err))
	}
	return
}

// GetArchiveExtraValue 获取稿件额外信息
func (d *Dao) GetArchiveExtraValue(ctx context.Context, aid int64) (map[string]string, error) {
	req := &aeapi.GetArchiveExtraValueReq{Aid: aid}
	rep, err := d.aeClient.GetArchiveExtraValue(ctx, req)
	if err != nil {
		err = errors.Wrapf(err, "%v", req)
		return nil, err
	}

	return rep.GetExtraInfo(), nil
}
