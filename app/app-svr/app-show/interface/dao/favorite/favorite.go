package favorite

import (
	"context"

	favgrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	"go-common/library/ecode"
	"go-common/library/log"
)

const (
	_popurlarFav        = 14
	_oidWeeklySelected  = 1
	_typeWeeklySelected = "weekly_selected"
)

func favStype(sType string) (oid int64) {
	switch sType {
	case _typeWeeklySelected:
		return _oidWeeklySelected
	}
	return
}

// FavAdd def.
func (d *Dao) FavAdd(ctx context.Context, mid int64, sType string) (err error) {
	oid := favStype(sType)
	if oid == 0 {
		return ecode.RequestErr
	}
	_, err = d.favClient.AddFav(ctx, &favgrpc.AddFavReq{
		Tp:  _popurlarFav,
		Mid: mid,
		Fid: 0,
		Oid: oid,
	})
	return
}

// FavDel deletes favorite from the default folder
func (d *Dao) FavDel(ctx context.Context, mid int64, sType string) (err error) {
	oid := favStype(sType)
	if oid == 0 {
		return ecode.RequestErr
	}
	_, err = d.favClient.DelFav(ctx, &favgrpc.DelFavReq{
		Tp:  _popurlarFav,
		Mid: mid,
		Fid: 0,
		Oid: oid,
	})
	return
}

// FavCheck returns whether the user has subscribed to the weekly-selected
func (d *Dao) FavCheck(ctx context.Context, mid int64, sType string) (fav bool, err error) {
	oid := favStype(sType)
	if oid == 0 {
		return false, ecode.RequestErr
	}
	var reply *favgrpc.IsFavoredReply
	if reply, err = d.favClient.IsFavored(ctx, &favgrpc.IsFavoredReq{
		Typ: _popurlarFav,
		Mid: mid,
		Oid: oid,
	}); err != nil {
		return
	}
	fav = reply.Faved
	return
}

// nolint:gomnd
func (d *Dao) Folders(c context.Context, folderIDs []int64, typ int32) (*favgrpc.FoldersReply, error) {
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
	if reply, err = d.favClient.Folders(c, req); err != nil {
		log.Error("d.favClient.Folders(%+v) error(%v)", req, err)
		return nil, err
	}
	return reply, nil
}
