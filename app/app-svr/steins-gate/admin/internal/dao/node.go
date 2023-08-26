package dao

import (
	"context"

	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/admin/internal/model"
	"go-gateway/app/app-svr/steins-gate/service/api"

	"github.com/pkg/errors"
)

// Node bts: -batch=50 -max_group=10 -batch_err=continue
func (d *Dao) Node(c context.Context, key int64) (res *api.GraphNode, err error) {
	var bvcRes *model.Dimension
	if res, err = d.RawNode(c, key); err != nil {
		err = errors.Wrapf(err, "NodeID %d", key)
		return
	}
	if res == nil {
		log.Warn("NodeID %d Not found", key)
		return
	}
	if bvcRes, err = d.BvcDimension(c, res.Cid); err != nil {
		err = errors.Wrapf(err, "NodeID %d", key)
		return
	}
	if bvcRes != nil {
		res.Height = bvcRes.Height
		res.Width = bvcRes.Width
	}
	return

}
