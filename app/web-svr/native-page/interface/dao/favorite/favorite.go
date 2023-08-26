package favorite

import (
	"context"

	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-common/library/log"
)

func (d *Dao) Folders(c context.Context, folderIDs []int64, typ int32) (*favgrpc.FoldersReply, error) {
	var (
		err   error
		ids   []*favgrpc.FolderID
		reply *favgrpc.FoldersReply
	)
	hundred := int64(100)
	for _, folderID := range folderIDs {
		fid := folderID / hundred
		mid := folderID % hundred
		ids = append(ids, &favgrpc.FolderID{Fid: fid, Mid: mid})
	}
	req := &favgrpc.FoldersReq{Typ: typ, Ids: ids}
	if reply, err = d.favClient.Folders(c, req); err != nil {
		log.Error("d.favClient.Folders(%+v) error(%v)", req, err)
		return nil, err
	}
	return reply, nil
}
