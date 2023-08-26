package dao

import (
	"context"

	favmodel "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
)

const (
	TypeVideo = 2
)

func (d *dao) UserDefaultFavFolder(ctx context.Context, mid int64) (*favmodel.Folder, error) {
	reply, err := d.fav.UserFolder(ctx, &favgrpc.UserFolderReq{
		Typ:  TypeVideo,
		Mid:  mid,
		Vmid: mid,
		Fid:  0,
	})
	if err != nil {
		return nil, err
	}
	return reply.Res, nil
}
