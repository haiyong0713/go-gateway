package dao

import (
	"context"

	arcgrpc "go-gateway/app/app-svr/archive/service/api"
)

// ArcDao is archive dao.
type arcDao struct {
	archive arcgrpc.ArchiveClient
}

func (d *arcDao) ArcsPlayer(ctx context.Context, req *arcgrpc.ArcsPlayerRequest) (map[int64]*arcgrpc.ArcPlayer, error) {
	info, err := d.archive.ArcsPlayer(ctx, req)
	if err != nil {
		return nil, err
	}
	return info.ArcsPlayer, nil
}
