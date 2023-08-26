package like

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/interface/model/like"

	api "git.bilibili.co/bapis/bapis-go/article/model"
	"github.com/pkg/errors"
)

const (
	_articleGiantURI   = "/x/internal/article/wenhao"
	_articleListURI    = "/x/internal/article/lists_info"
	_upArtListURI      = "/x/internal/article/up/lists"
	_articleGiantV4URI = "/x/internal/article/literaryGiant"
	_articleResetURI   = "/x/internal/article/literaryGiantDone"
)

// ArticleGiantV4 get article giant data.
func (d *Dao) ArticleGiantV4(c context.Context, mid int64) (data map[int64]*api.Meta, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int                 `json:"code"`
		Data map[int64]*api.Meta `json:"data"`
	}
	if err = d.client.Get(c, d.artGiantV4URL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.artGiantV4URL+"?"+params.Encode())
		return
	}
	data = res.Data
	return
}

func (d *Dao) ArticleGiantV4Reset(c context.Context, mid int64) (err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err = d.client.Get(c, d.artResetURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.artResetURL+"?"+params.Encode())
	}
	return
}

// ArticleGiant get article giant data.
func (d *Dao) ArticleGiant(c context.Context, mid int64) (data *like.ArticleGiant, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int                `json:"code"`
		Data *like.ArticleGiant `json:"data"`
	}
	if err = d.client.Get(c, d.articleGiantURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.articleGiantURL+"?"+params.Encode())
		return
	}
	data = res.Data
	return
}

// ArticleLists .
func (d *Dao) ArticleLists(c context.Context, ids []int64) (data []*like.ArticleList, err error) {
	params := url.Values{}
	params.Set("listIDs", xstr.JoinInts(ids))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Lists []*like.ArticleList `json:"lists"`
		}
	}
	if err = d.client.Get(c, d.articleListsURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.articleListsURL+"?"+params.Encode())
		return
	}
	data = res.Data.Lists
	return
}

// UpArtLists .
func (d *Dao) UpArtLists(c context.Context, mid int64) (list []*like.ArticleList, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data struct {
			Lists []*like.ArticleList `json:"lists"`
			Total int64               `json:"total"`
		}
	}
	if err = d.client.Get(c, d.upArtListURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.upArtListURL+"?"+params.Encode())
		return
	}
	list = res.Data.Lists
	return
}
