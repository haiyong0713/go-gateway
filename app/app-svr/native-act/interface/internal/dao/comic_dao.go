package dao

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-common/library/utils/collection"

	"go-gateway/app/app-svr/native-act/interface/internal/model"
)

const (
	_comicInfosUri  = "/twirp/comic.v0.Comic/GetComicInfos"
	_comicAddFavUri = "/twirp/bookshelf.v1.Bookshelf/AddFavorite"
	_comicDelFavUri = "/twirp/bookshelf.v1.Bookshelf/DeleteFavorite"
)

type comicDao struct {
	host       string
	httpClient *bm.Client
}

func (d *comicDao) GetComicInfos(ctx context.Context, ids []int64, mid int64) (map[int64]*model.ComicInfo, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("ids", collection.JoinSliceInt(ids, ","))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int                `json:"code"`
		Data []*model.ComicInfo `json:"data"`
	}
	if err := d.httpClient.Post(ctx, d.host+_comicInfosUri, ip, params, &res); err != nil {
		log.Error("Fail to request comic.GetComicInfos, params=%+v error=%+v", params.Encode(), err)
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.host+_comicInfosUri+"?"+params.Encode())
		log.Error("Fail to request comic.GetComicInfos, error=%+v", err)
		return nil, err
	}
	infos := make(map[int64]*model.ComicInfo, len(res.Data))
	for _, comic := range res.Data {
		if comic == nil {
			continue
		}
		infos[comic.ID] = comic
	}
	return infos, nil
}

func (d *comicDao) AddFavorite(ctx context.Context, ids []int64, mid int64) error {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("comic_ids", collection.JoinSliceInt(ids, ","))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err := d.httpClient.Post(ctx, d.host+_comicAddFavUri, ip, params, &res); err != nil {
		log.Error("Fail to request comic.AddFavorite, params=%+v error=%+v", params.Encode(), err)
		return err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.host+_comicAddFavUri+"?"+params.Encode())
		log.Error("Fail to request comic.AddFavorite, error=%+v", err)
		return err
	}
	return nil
}

func (d *comicDao) DelFavorite(ctx context.Context, ids []int64, mid int64) error {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	params.Set("comic_ids", collection.JoinSliceInt(ids, ","))
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	var res struct {
		Code int `json:"code"`
	}
	if err := d.httpClient.Post(ctx, d.host+_comicDelFavUri, ip, params, &res); err != nil {
		log.Error("Fail to request comic.DelFavorite, params=%+v error=%+v", params.Encode(), err)
		return err
	}
	if res.Code != ecode.OK.Code() {
		err := errors.Wrap(ecode.Int(res.Code), d.host+_comicDelFavUri+"?"+params.Encode())
		log.Error("Fail to request comic.DelFavorite, error=%+v", err)
		return err
	}
	return nil
}
