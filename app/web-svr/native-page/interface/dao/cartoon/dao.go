package cartoon

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go-common/library/ecode"
	httpx "go-common/library/net/http/blademaster"

	"go-gateway/app/web-svr/native-page/interface/conf"
	cmdl "go-gateway/app/web-svr/native-page/interface/model/cartoon"

	"github.com/pkg/errors"
)

const (
	_comicInfosURI = "/twirp/comic.v0.Comic/GetComicInfos"
)

type Dao struct {
	c             *conf.Config
	client        *httpx.Client
	ComicInfosURL string
}

func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:             c,
		client:        httpx.NewClient(c.HTTPMangaCo),
		ComicInfosURL: c.Host.MangaCo + _comicInfosURI,
	}
	return
}

func (d *Dao) GetComicInfos(c context.Context, ids []int64, mid int64) (map[int64]*cmdl.ComicItem, error) {
	p := struct {
		IDs []int64 `json:"ids"`
		Mid string  `json:"mid"`
	}{
		IDs: ids,
		Mid: fmt.Sprintf("%d", mid),
	}
	bs, _ := json.Marshal(p)
	payload := strings.NewReader(string(bs))
	req, err := http.NewRequest("POST", d.ComicInfosURL, payload)
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
		return nil, errors.Wrap(ecode.Int(res.Code), d.ComicInfosURL+"?"+string(bs))
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
