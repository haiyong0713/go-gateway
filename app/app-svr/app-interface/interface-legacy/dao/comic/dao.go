package comic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
	"go-gateway/app/app-svr/app-interface/interface-legacy/model/comic"

	"github.com/pkg/errors"
)

// Dao comic dao.
type Dao struct {
	c      *conf.Config
	client *bm.Client
	// url
	upComic       string
	comics        string
	favComic      string
	favComicCount string
}

const (
	_upComic       = "/twirp/comic.v0.Comic/GetUserComics"
	_comics        = "/twirp/comic.v0.Comic/GetComicInfos"
	_favComic      = "/twirp/bookshelf.v1.Bookshelf/ListFavorite"
	_favComicCount = "/twirp/bookshelf.v1.Bookshelf/CountFavorite"
)

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:             c,
		client:        bm.NewClient(c.HTTPClient),
		upComic:       c.Host.Manga + _upComic,
		comics:        c.Host.Manga + _comics,
		favComic:      c.Host.Manga + _favComic,
		favComicCount: c.Host.Manga + _favComicCount,
	}
	return
}

func (d *Dao) Comics(c context.Context, ids []int64) (comics map[int64]*comic.Comic, err error) {
	type params struct {
		IDs []int64 `json:"ids"`
	}
	p := &params{
		IDs: ids,
	}
	bs, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", d.comics, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int            `json:"code"`
		Msg  string         `json:"msg"`
		Data []*comic.Comic `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.upComic)
		return
	}
	comics = make(map[int64]*comic.Comic)
	for _, comic := range res.Data {
		comics[comic.ID] = comic
	}
	return
}

func (d *Dao) UpComics(c context.Context, mid int64, pn, ps int) (comics *comic.Comics, err error) {
	type params struct {
		UID      string `json:"uid"`
		Page     int    `json:"page"`
		PageSize int    `json:"page_size"`
	}
	p := &params{
		UID:      strconv.FormatInt(mid, 10),
		Page:     pn,
		PageSize: ps,
	}
	bs, _ := json.Marshal(p)
	req, _ := http.NewRequest("POST", d.upComic, strings.NewReader(string(bs)))
	req.Header.Set("Content-Type", "application/json")
	var res struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *comic.Comics `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.upComic)
		return
	}
	comics = res.Data
	return
}

func (d *Dao) FavComics(c context.Context, mid int64, pn, ps int) (comics []*comic.FavComic, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("page_num", strconv.Itoa(pn))
	params.Set("page_size", strconv.Itoa(ps))
	paramStr := params.Encode()
	if strings.IndexByte(paramStr, '+') > -1 {
		paramStr = strings.Replace(paramStr, "+", "%20", -1)
	}
	var (
		buffer bytes.Buffer
		querry string
	)
	buffer.WriteString(paramStr)
	querry = buffer.String()
	req, _ := http.NewRequest("POST", d.favComic, strings.NewReader(querry))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	var res struct {
		Code int               `json:"code"`
		Msg  string            `json:"msg"`
		Data []*comic.FavComic `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.favComic)
		return
	}
	comics = res.Data
	return
}

func (d *Dao) FavComicsCount(c context.Context, mid int64) (count int, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	paramStr := params.Encode()
	if strings.IndexByte(paramStr, '+') > -1 {
		paramStr = strings.Replace(paramStr, "+", "%20", -1)
	}
	var (
		buffer bytes.Buffer
		querry string
	)
	buffer.WriteString(paramStr)
	querry = buffer.String()
	req, _ := http.NewRequest("POST", d.favComicCount, strings.NewReader(querry))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	var res struct {
		Code int                  `json:"code"`
		Msg  string               `json:"msg"`
		Data *comic.FavComicCount `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("%v", err)
		return
	}
	if res.Code != 0 {
		err = errors.Wrap(ecode.Int(res.Code), d.favComicCount)
		return
	}
	if res.Data != nil {
		count = res.Data.Count
	}
	return
}
