package space

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-gateway/app/app-svr/app-job/job/model/space"

	"github.com/pkg/errors"
)

const (
	_audioList = "/audio/music-service-c/songs/internal/uppersongs-page"
	_upComic   = "/twirp/comic.v0.Comic/GetUserComics"
)

func (d *Dao) AudioList(c context.Context, vmid int64, pn, ps int, ip string) (aus []*space.Audio, hasNext bool, nextPage int, err error) {
	params := url.Values{}
	params.Set("uid", strconv.FormatInt(vmid, 10))
	params.Set("pageIndex", strconv.Itoa(pn))
	params.Set("pageSize", strconv.Itoa(ps))
	var res struct {
		Code int `json:"code"`
		Data *struct {
			NextPage    int            `json:"nextPage"`
			HasNextPage bool           `json:"hasNextPage"`
			List        []*space.Audio `json:"list"`
		} `json:"data"`
	}
	if err = d.clientAsyn.Get(c, d.audioList, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(err, d.audioList+"?"+params.Encode())
		return
	}
	if res.Data != nil {
		aus = res.Data.List
		hasNext = res.Data.HasNextPage
		nextPage = res.Data.NextPage
	}
	return
}
