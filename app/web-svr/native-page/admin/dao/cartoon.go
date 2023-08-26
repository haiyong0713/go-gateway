package dao

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-common/library/ecode"

	"go-gateway/app/web-svr/native-page/admin/model"

	"github.com/pkg/errors"
)

const (
	_comicInfosURI = "/twirp/comic.v0.Comic/GetComicInfos"
)

func (d *Dao) GetComicInfos(c context.Context, ids []int64) (map[int64]*model.ComicItem, error) {
	p := struct {
		IDs []int64 `json:"ids"`
	}{
		IDs: ids,
	}
	bs, _ := json.Marshal(p)
	payload := strings.NewReader(string(bs))
	req, err := http.NewRequest("POST", d.ComicInfosURL, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("content-type", "application/json; charset=utf-8")
	var res struct {
		Code int                `json:"code"`
		Data []*model.ComicItem `json:"data"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, errors.Wrap(ecode.Int(res.Code), d.ComicInfosURL+"?"+string(bs))
	}
	rly := make(map[int64]*model.ComicItem)
	for _, v := range res.Data {
		if v == nil {
			continue
		}
		rly[v.ID] = v
	}
	return rly, nil
}
