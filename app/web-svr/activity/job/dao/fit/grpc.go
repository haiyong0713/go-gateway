package fit

import (
	"context"

	"go-common/library/log"
	favgrpc "go-main/app/community/favorite/service/api"
)

// Folders grpc
func (d *dao) Folders(c context.Context, folderIDs []int64, typ int32) (*favgrpc.FoldersReply, error) {
	var (
		err   error
		ids   []*favgrpc.FolderID
		reply *favgrpc.FoldersReply
	)
	for _, folderID := range folderIDs {
		fid := folderID / 100
		mid := folderID % 100
		ids = append(ids, &favgrpc.FolderID{Fid: fid, Mid: mid})
	}
	req := &favgrpc.FoldersReq{Typ: typ, Ids: ids}
	if reply, err = d.FavClient.Folders(c, req); err != nil {
		log.Errorc(c, "d.favClient.Folders(%+v) error(%v)", req, err)
		return nil, err
	}
	return reply, nil
}

// FavoritesAll 获取播单里详情
func (d *dao) FavoritesAll(c context.Context, tp int32, mid, uid, fid int64, pn, ps int32) (fav *favgrpc.FavoritesReply, err error) {
	fav = &favgrpc.FavoritesReply{}
	fav, err = d.FavClient.FavoritesAll(c, &favgrpc.FavoritesReq{
		Tp:  tp,
		Mid: mid, //访问者MID
		Uid: uid, //收藏夹创建者MID
		Fid: fid, //收藏夹ID /100
		Pn:  pn,
		Ps:  ps,
	})
	if err != nil {
		log.Errorc(c, "dao get FavoritesAll: error(%v)", err)
		return nil, err
	}
	return fav, nil
}
