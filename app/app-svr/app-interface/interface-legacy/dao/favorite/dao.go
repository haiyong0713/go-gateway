package favorite

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/time"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/favorite"

	favmdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favclient "git.bilibili.co/bapis/bapis-go/community/service/favorite"

	"github.com/pkg/errors"
)

const (
	_folder      = "/x/internal/v2/fav/folder"
	_folderVideo = "/x/internal/v2/fav/video"
)

// Dao is favorite dao
type Dao struct {
	client     *httpx.Client
	favor      string
	favorVideo string
	// rpc
	favClient favclient.FavoriteClient
}

// New initial favorite dao
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		client:     httpx.NewClient(c.HTTPClient),
		favor:      c.Host.APICo + _folder,
		favorVideo: c.Host.APICo + _folderVideo,
	}
	var err error
	if d.favClient, err = favclient.NewClient(c.FavClient); err != nil {
		panic(err)
	}
	return
}

// Folders get favorite floders from api.
func (d *Dao) Folders(c context.Context, mid, vmid int64, mobiApp string, build int, mediaList bool) (fs []*favorite.Folder, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("vmid", strconv.FormatInt(vmid, 10))
	params.Set("mobi_app", mobiApp)
	// params.Set("build", strconv.Itoa(build))
	if mediaList {
		params.Set("medialist", "1")
	}
	var res struct {
		Code int                `json:"code"`
		Data []*favorite.Folder `json:"data"`
	}
	if err = d.client.Get(c, d.favor, ip, params, &res); err != nil {
		return
	}
	b, _ := json.Marshal(&res)
	log.Info("Folders url(%s) response(%s)", d.favor+"?"+params.Encode(), b)
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.favor+"?"+params.Encode())
		return
	}
	fs = res.Data
	return
}

// FolderVideo get favorite floders from UGC api.
func (d *Dao) FolderVideo(c context.Context, accessKey, actionKey, device, mobiApp, platform, keyword, order string, build, tid, pn, ps int, mid, fid, vmid int64) (fav *favorite.Video, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("access_key", accessKey)
	params.Set("actionKey", actionKey)
	params.Set("build", strconv.Itoa(build))
	params.Set("device", device)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("fid", strconv.FormatInt(fid, 10))
	params.Set("tid", strconv.Itoa(tid))
	params.Set("keyword", keyword)
	params.Set("order", order)
	params.Set("pn", strconv.Itoa(pn))
	params.Set("ps", strconv.Itoa(ps))
	params.Set("mobi_app", mobiApp)
	params.Set("platform", platform)
	params.Set("vmid", strconv.FormatInt(vmid, 10))
	var res struct {
		Code int             `json:"code"`
		Data *favorite.Video `json:"data"`
	}
	if err = d.client.Get(c, d.favorVideo, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.favorVideo+"?"+params.Encode())
		return
	}
	fav = res.Data
	return
}

// UserFolders is get folder.
func (d *Dao) UserFolders(c context.Context, typ int8, mid, vmid int64) (res []*favmdl.Folder, err error) {
	var userFolders *favclient.UserFoldersReply
	if userFolders, err = d.favClient.UserFolders(c, &favclient.UserFoldersReq{Typ: int32(typ), Mid: mid, Vmid: vmid}); err != nil {
		log.Error("d.favRPC.UserFolder error(%+v)", err)
		return
	}
	res = userFolders.GetRes()
	return
}

// FavoritesRPC favorites list.
func (d *Dao) FavoritesRPC(c context.Context, typ int8, mid, vmid, fid int64) (favs *favorite.Favorites, err error) {
	arg := &favclient.FavoritesReq{Tp: int32(typ), Mid: mid, Uid: vmid, Fid: fid}
	var reply *favclient.FavoritesReply
	if reply, err = d.favClient.FavoritesAll(c, arg); err != nil {
		log.Error("d.favClient.FavoritesAll(%+v) error(%v)", arg, err)
		return
	}
	favs = &favorite.Favorites{}
	favs.Page.Count = int(reply.Res.Page.Count)
	favs.Page.Num = int(reply.Res.Page.Num)
	favs.Page.Size = int(reply.Res.Page.Size_)
	for _, data := range reply.Res.List {
		favs.List = append(favs.List, &favorite.Favorite{
			ID:    data.Id,
			Oid:   data.Oid,
			Mid:   data.Mid,
			Fid:   data.Fid,
			Type:  int8(data.Type),
			State: int8(data.State),
			CTime: time.Time(data.Ctime),
			MTime: time.Time(data.Mtime),
		})
	}
	return
}

// UserFavs user favorite count.
func (d *Dao) UserFavs(c context.Context, types []int32, mid int64) (res map[int32]int64, err error) {
	var (
		req      = &favclient.UserFavsReq{Types: types, Mid: mid}
		userFavs *favclient.UserFavsReply
	)
	if userFavs, err = d.favClient.UserFavs(c, req); err != nil {
		err = errors.Wrapf(err, "%v", req)
		return
	}
	if userFavs != nil {
		res = userFavs.Favs
	}
	return
}

func (d *Dao) IsFavVideos(ctx context.Context, mid int64, aids []int64) (map[int64]int8, error) {
	const _typeVideo = 2

	reply, err := d.favClient.IsFavoreds(ctx, &favclient.IsFavoredsReq{
		Typ:  _typeVideo,
		Mid:  mid,
		Oids: aids,
	})
	if err != nil {
		return nil, err
	}
	res := make(map[int64]int8)
	for k, v := range reply.Faveds {
		if v {
			res[k] = 1
		}
	}
	return res, nil
}

func (d *Dao) LastFavTime(ctx context.Context, mid int64) (int64, error) {
	reply, err := d.favClient.LastFavTime(ctx, &favclient.LastFavTimeReq{Mid: mid, Types: []int32{2, 11}})
	if err != nil {
		return 0, errors.Wrapf(err, "LastFavTime error mid(%d)", mid)
	}
	return reply.FavTime, nil
}
