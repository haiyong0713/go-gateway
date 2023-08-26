package dao

import (
	"context"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favoritegrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
)

const (
	_folderDivision = 100
)

type favoriteDao struct {
	client favoritegrpc.FavoriteClient
}

func (d *favoriteDao) Folders(c context.Context, folderIDs []int64, typ int32) (map[int64]*favmdl.Folder, error) {
	reqIDs := make([]*favoritegrpc.FolderID, 0, len(folderIDs))
	for _, folderID := range folderIDs {
		fid := folderID / _folderDivision
		mid := folderID % _folderDivision
		reqIDs = append(reqIDs, &favoritegrpc.FolderID{Fid: fid, Mid: mid})
	}
	req := &favoritegrpc.FoldersReq{Typ: typ, Ids: reqIDs}
	rly, err := d.client.Folders(c, req)
	if err != nil {
		return nil, err
	}
	list := make(map[int64]*favmdl.Folder, len(rly.Res))
	for _, folder := range rly.Res {
		list[folder.Mlid] = folder
	}
	return list, nil
}
