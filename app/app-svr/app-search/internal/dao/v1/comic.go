package v1

import (
	"context"
	"net/http"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-common/library/utils/collection"
	"go-gateway/app/app-svr/app-search/internal/model/search"

	"github.com/pkg/errors"
)

const _comicInfos = "/twirp/comic.v0.Comic/GetComicInfos"

func (d *dao) GetComicInfos(ctx context.Context, ids []int64) (map[int64]*search.ComicInfo, error) {
	var (
		req *http.Request
		err error
	)
	reply := &struct {
		Code int64               `json:"code"`
		Msg  string              `json:"msg"`
		Data []*search.ComicInfo `json:"data"`
	}{}
	params := url.Values{}
	params.Set("ids", collection.JoinSliceInt(ids, ","))
	// new request
	if req, err = d.client.NewRequest("POST", d.comicInfos, metadata.String(ctx, metadata.RemoteIP), params); err != nil {
		return nil, err
	}
	if err = d.client.Do(ctx, req, reply); err != nil {
		return nil, err
	}
	if reply.Code != 0 {
		return nil, errors.Wrapf(ecode.New(int(reply.Code)), "reply code:%d, url:%s", reply.Code, d.main+"?"+params.Encode())
	}
	res := make(map[int64]*search.ComicInfo, len(reply.Data))
	for _, comic := range reply.Data {
		res[comic.ID] = comic
	}
	return res, nil
}
