package cartoon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-show/interface/conf"
	cmdl "go-gateway/app/app-svr/app-show/interface/model/cartoon"

	"github.com/pkg/errors"
)

const (
	_comicInfosURI  = "/twirp/comic.v0.Comic/GetComicInfos"
	_addFavoriteURI = "/twirp/bookshelf.v1.Bookshelf/AddFavorite"
	_delFavoriteURI = "/twirp/bookshelf.v1.Bookshelf/DeleteFavorite"
)

type Dao struct {
	c              *conf.Config
	client         *httpx.Client
	comicInfosURL  string
	addFavoriteURL string
	delFavoriteURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:              c,
		client:         httpx.NewClient(c.HTTPMangaCo),
		comicInfosURL:  c.Host.MangaCo + _comicInfosURI,
		addFavoriteURL: c.Host.MangaCo + _addFavoriteURI,
		delFavoriteURL: c.Host.MangaCo + _delFavoriteURI,
	}
	return
}

func (d *Dao) GetComicInfos(c context.Context, ids []int64, mid int64, from string) (map[int64]*cmdl.ComicItem, error) {
	p := struct {
		IDs  []int64 `json:"ids"`
		Mid  string  `json:"mid"`
		From string  `json:"from"`
	}{
		IDs:  ids,
		Mid:  fmt.Sprintf("%d", mid),
		From: from,
	}
	bs, _ := json.Marshal(p)
	payload := strings.NewReader(string(bs))
	req, err := http.NewRequest("POST", d.comicInfosURL, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json; charset=utf-8")
	var res struct {
		Code int               `json:"code"`
		Data []*cmdl.ComicItem `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.comicInfosURL+"?"+string(bs))
	}
	rly := make(map[int64]*cmdl.ComicItem)
	for _, v := range res.Data {
		if v == nil {
			continue
		}
		rly[v.ID] = v
	}
	return rly, nil
}

func (d *Dao) AddFavorite(c context.Context, ids []int64, mid int64) error {
	p := struct {
		ComicIDS string `json:"comic_ids"`
	}{
		ComicIDS: xstr.JoinInts(ids),
	}
	bs, _ := json.Marshal(p)
	payload := strings.NewReader(string(bs))
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	addURL := fmt.Sprintf("%s?%s", d.addFavoriteURL, params.Encode())
	req, err := http.NewRequest("POST", addURL, payload)
	if err != nil {
		log.Error("AddFavorite NewRequest(%s) error(%v)", addURL, err)
		return err
	}
	req.Header.Add("content-type", "application/json; charset=utf-8")
	var res struct {
		Code int `json:"code"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("AddFavorite  d.client.Do(%s,ids:%v) error(%v)", addURL, ids, err)
		return err
	}
	if res.Code != ecode.OK.Code() {
		log.Error("AddFavorite url(%s) ids(%+v) ecode(%d)", addURL+"&"+string(bs), ids, res.Code)
		return ecode.Int(res.Code)
	}
	return nil
}

func (d *Dao) DelFavorite(c context.Context, ids []int64, mid int64) error {
	p := struct {
		ComicIDS string `json:"comic_ids"`
	}{
		ComicIDS: xstr.JoinInts(ids),
	}
	bs, _ := json.Marshal(p)
	payload := strings.NewReader(string(bs))
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	delURL := fmt.Sprintf("%s?%s", d.delFavoriteURL, params.Encode())
	req, err := http.NewRequest("POST", delURL, payload)
	if err != nil {
		log.Error("DelFavorite  NewRequest(%s) error(%v)", delURL, err)
		return err
	}
	req.Header.Add("content-type", "application/json; charset=utf-8")
	var res struct {
		Code int `json:"code"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("DelFavorite  d.client.Do(%s,ids:%v) error(%v)", delURL, ids, err)
		return err
	}
	if res.Code != ecode.OK.Code() {
		log.Error("DelFavorite  ecode(%s) error(%d)", delURL+"&"+string(bs), res.Code)
		return ecode.Int(res.Code)
	}
	return nil
}
