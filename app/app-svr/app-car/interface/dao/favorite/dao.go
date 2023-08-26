package favorite

import (
	"context"
	"fmt"

	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-car/interface/conf"

	favorite "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favoritegrpc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	toviewgrpc "git.bilibili.co/bapis/bapis-go/community/service/toview"
)

const (
	// 2 视频
	_favTypeVedio = 2
)

type Dao struct {
	client      *httpx.Client
	folderSpace string
	// grpc
	toviewClient   toviewgrpc.ToViewsClient
	favoriteClient favoritegrpc.FavoriteClient
}

func New(c *conf.Config) *Dao {
	d := &Dao{
		client:      httpx.NewClient(c.HTTPClient),
		folderSpace: c.Host.APICo + _folderSpace,
	}
	var err error
	if d.toviewClient, err = toviewgrpc.NewClient(nil); err != nil {
		panic(fmt.Sprintf("toviewgrpc NewClientt error (%+v)", err))
	}
	if d.favoriteClient, err = favoritegrpc.NewClient(nil); err != nil {
		panic(fmt.Sprintf("favoritegrpc NewClientt error (%+v)", err))
	}
	return d
}

func (d *Dao) UserToViews(ctx context.Context, mid int64, pn, ps int) ([]*toviewgrpc.ToView, error) {
	reply, err := d.toviewClient.UserToViews(ctx, &toviewgrpc.UserToViewsReq{Mid: mid, Pn: int32(pn), Ps: int32(ps), BusinessId: 1})
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetToviews(), nil
}

func (d *Dao) FavoritesAll(ctx context.Context, mid, vmid, favid int64, favType, pn, ps int) (*favoritegrpc.ModelFavorites, error) {
	arg := &favoritegrpc.FavoritesReq{
		Tp:  int32(favType),
		Mid: mid,
		Uid: vmid,
		Fid: favid,
		Pn:  int32(pn),
		Ps:  int32(ps),
	}
	reply, err := d.favoriteClient.FavoritesAll(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetRes(), nil
}

func (d *Dao) UserFolders(ctx context.Context, mid, oid int64, favType int) ([]*favorite.Folder, error) {
	arg := &favoritegrpc.UserFoldersReq{
		Typ: int32(favType),
		Mid: mid,
		Oid: oid,
	}
	reply, err := d.favoriteClient.UserFolders(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetRes(), nil
}

func (d *Dao) Folders(ctx context.Context, ids []int64, favType int, mid int64) ([]*favorite.Folder, error) {
	var favIds []*favoritegrpc.FolderID
	for _, id := range ids {
		// nolint:gomnd
		fid := id / 100
		// nolint:gomnd
		favmid := id % 100
		favIds = append(favIds, &favoritegrpc.FolderID{Fid: fid, Mid: favmid})
	}
	arg := &favoritegrpc.FoldersReq{
		Ids: favIds,
		Mid: mid,
		Typ: int32(favType),
	}
	reply, err := d.favoriteClient.Folders(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return nil, err
	}
	return reply.GetRes(), nil
}

func (d *Dao) FavAddFolders(ctx context.Context, mobiApp, device, platform string, fids []int64, aid, mid int64) error {
	arg := &favoritegrpc.FavAddFoldersReq{
		Typ:      _favTypeVedio,
		Otype:    _favTypeVedio,
		Fids:     fids,
		Oid:      aid,
		Mid:      mid,
		MobiApp:  mobiApp,
		Device:   device,
		Platform: platform,
	}
	if _, err := d.favoriteClient.FavAddFolders(ctx, arg); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

func (d *Dao) FavDelFolders(ctx context.Context, mobiApp, device, platform string, fids []int64, aid, mid int64) error {
	arg := &favoritegrpc.FavDelFoldersReq{
		Typ:      _favTypeVedio,
		Otype:    _favTypeVedio,
		Fids:     fids,
		Oid:      aid,
		Mid:      mid,
		MobiApp:  mobiApp,
		Device:   device,
		Platform: platform,
	}
	if _, err := d.favoriteClient.FavDelFolders(ctx, arg); err != nil {
		log.Error("%+v", err)
		return err
	}
	return nil
}

func (d *Dao) AddFolder(ctx context.Context, mid int64, name, desc string, public int) (int64, error) {
	arg := &favoritegrpc.AddFolderReq{
		Typ:         _favTypeVedio,
		Mid:         mid,
		Name:        name,
		Description: desc,
		Public:      int32(public),
	}
	reply, err := d.favoriteClient.AddFolder(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return 0, err
	}
	return reply.GetFid(), nil
}

func (d *Dao) IsFavored(ctx context.Context, mid, aid int64) bool {
	arg := &favoritegrpc.IsFavoredReq{
		Typ: _favTypeVedio,
		Oid: aid,
		Mid: mid,
	}
	reply, err := d.favoriteClient.IsFavored(ctx, arg)
	if err != nil {
		log.Error("%+v", err)
		return false
	}
	return reply.GetFaved()
}
