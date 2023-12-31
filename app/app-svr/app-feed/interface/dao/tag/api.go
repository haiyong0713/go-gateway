package tag

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-feed/interface/model/tag"
)

const (
	_hot    = "/x/internal/tag/hots"
	_tags   = "/x/internal/tag/archive/multi/tags"
	_detail = "/x/internal/tag/detail"
)

// Hots.
func (d *Dao) Hots(c context.Context, mid int64, rid int64, now time.Time) (hs []*tag.Hot, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("rid", strconv.FormatInt(rid, 10))
	var res struct {
		Code int        `json:"code"`
		Data []*tag.Hot `json:"data"`
	}
	if err = d.client.Get(c, d.hot, "", params, &res); err != nil {
		return
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		err = errors.Wrap(code, d.hot+"?"+params.Encode())
		return
	}
	hs = res.Data
	return
}

func (d *Dao) Tags(c context.Context, mid int64, aids []int64, now time.Time) (tagm map[string][]*tag.Tag, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("aids", xstr.JoinInts(aids))
	var res struct {
		Code int                   `json:"code"`
		Data map[string][]*tag.Tag `json:"data"`
	}
	if err = d.client.Get(c, d.tags, "", params, &res); err != nil {
		return
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		err = errors.Wrap(code, d.tags+"?"+params.Encode())
		return
	}
	tagm = res.Data
	return
}

func (d *Dao) Detail(c context.Context, tagID int, pn, ps int, now time.Time) (arcids []int64, err error) {
	params := url.Values{}
	params.Set("tag_id", strconv.Itoa(tagID))
	params.Set("pn", strconv.Itoa(pn))
	params.Set("ps", strconv.Itoa(ps))
	var res struct {
		Code int `json:"code"`
		Data struct {
			News struct {
				Archives []struct {
					Aid int64 `json:"aid"`
				} `json:"archives"`
			} `json:"news"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.detail, "", params, &res); err != nil {
		return
	}
	code := ecode.Int(res.Code)
	if !code.Equal(ecode.OK) {
		err = errors.Wrap(code, d.detail+"?"+params.Encode())
		return
	}
	for _, arcs := range res.Data.News.Archives {
		arcids = append(arcids, arcs.Aid)
	}
	return
}
