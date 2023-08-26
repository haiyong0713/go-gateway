package comic

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"

	comicmdl "go-gateway/app/app-svr/app-dynamic/interface/model/comic"
	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"

	"go-gateway/app/app-svr/app-dynamic/interface/conf"

	"github.com/pkg/errors"
)

type Dao struct {
	c *conf.Config
	// http client
	client *bm.Client
	// domain
	additional        string
	batchGetInfo      string
	batchListFavorite string
	batchDelFavs      string
	batchIsFav        string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:                 c,
		client:            bm.NewClient(c.HTTPClient),
		additional:        c.Hosts.Comic + _additional,
		batchGetInfo:      c.Hosts.Comic + _batchGetInfo,
		batchListFavorite: c.Hosts.Comic + _batchListFavorite,
		batchDelFavs:      c.Hosts.Comic + _batchDelFavs,
		batchIsFav:        c.Hosts.Comic + _batchIsFav,
	}
	return
}

const (
	_additional        = "/twirp/comic.v0.Comic/GetComicInfos"
	_batchGetInfo      = "/twirp/comic.v0.Dynamic/BatchGetInfo"
	_batchListFavorite = "/twirp/bookshelf.v0.Bookshelf/ListFavorite"
	_batchDelFavs      = "/twirp/bookshelf.v0.Bookshelf/DelFavs"
	_batchIsFav        = "/twirp/bookshelf.v0.Bookshelf/IsFav"
)

func (d *Dao) Comics(c context.Context, mid int64, ids []int64) (comics map[int64]*comicmdl.Comic, err error) {
	type params struct {
		IDs []int64 `json:"ids"`
		Mid string  `json:"mid"`
	}
	p := &params{
		IDs: ids,
		Mid: strconv.FormatInt(mid, 10),
	}
	bs, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", d.additional, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int               `json:"code"`
		Msg  string            `json:"msg"`
		Data []*comicmdl.Comic `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		xmetric.DyanmicItemAPI.Inc(d.additional, "request_error")
		log.Error("%v", err)
		return
	}
	if res.Code != 0 {
		xmetric.DyanmicItemAPI.Inc(d.additional, "reply_code_error")
		err = errors.Wrap(ecode.Int(res.Code), d.additional)
		return
	}
	comics = make(map[int64]*comicmdl.Comic)
	for _, comic := range res.Data {
		comics[comic.ID] = comic
	}
	return
}

func (d *Dao) BatchInfo(c context.Context, mid int64, ids []int64) (map[int64]*comicmdl.Batch, error) {
	type params struct {
		IDs []int64 `json:"ids"`
		Mid string  `json:"mid"`
	}
	p := &params{
		IDs: ids,
		Mid: strconv.FormatInt(mid, 10),
	}
	bs, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", d.batchGetInfo, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Comics map[int64]*comicmdl.Batch `json:"comics"`
		} `json:"data"`
	}
	if err := d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.Wrap(ecode.Int(res.Code), d.batchGetInfo)
	}
	return res.Data.Comics, nil
}

func (d *Dao) ListFavorite(c context.Context, mid int64) ([]int64, error) {
	type params struct {
		Mid string `json:"mid"`
	}
	p := &params{
		Mid: strconv.FormatInt(mid, 10),
	}
	bs, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", d.batchListFavorite, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			ComicIDs []int64 `json:"comic_ids"`
		} `json:"data"`
	}
	if err := d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.Wrap(ecode.Int(res.Code), d.batchListFavorite)
	}
	return res.Data.ComicIDs, nil
}

func (d *Dao) DelFavs(c context.Context, mid, cid int64) error {
	type params struct {
		Mid  string  `json:"mid"`
		Cids []int64 `json:"cids"`
	}
	p := &params{
		Mid:  strconv.FormatInt(mid, 10),
		Cids: []int64{cid},
	}
	bs, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", d.batchDelFavs, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return err
	}
	if res.Code != 0 {
		return errors.Wrap(ecode.Int(res.Code), d.batchDelFavs)
	}
	return nil
}

func (d *Dao) IsFav(c context.Context, mid int64, cids []int64) (map[int64]bool, error) {
	type params struct {
		Mid  string  `json:"mid"`
		Cids []int64 `json:"cids"`
	}
	p := &params{
		Mid:  strconv.FormatInt(mid, 10),
		Cids: cids,
	}
	bs, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", d.batchIsFav, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int `json:"code"`
		Data struct {
			Info map[int64]bool `json:"info"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	if err := d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.Wrap(ecode.Int(res.Code), d.batchIsFav)
	}
	return res.Data.Info, nil
}
