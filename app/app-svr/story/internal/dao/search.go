package dao

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/story/internal/model"

	"github.com/pkg/errors"
)

const (
	_arcSearchURI    = "http://s.search.bilibili.co/space/search/v2"
	_arcSearchType   = "sub_video"
	_additionalRanks = "-6"
)

func (d *dao) ArcSpaceSearch(ctx context.Context, arg *model.ArcSearchParam) (*model.ArcSearchReply, int64, error) {
	ip := metadata.String(ctx, metadata.RemoteIP)
	params := url.Values{}
	if arg.AttrNot != 0 {
		params.Set("attr_not", strconv.FormatUint(arg.AttrNot, 10))
	}
	params.Set("search_type", _arcSearchType)
	params.Set("additional_ranks", _additionalRanks)
	if arg.Mid > 0 {
		params.Set("mid", strconv.FormatInt(arg.Mid, 10))
	}
	params.Set("page", strconv.FormatInt(arg.Pn, 10))
	params.Set("pagesize", strconv.FormatInt(arg.Ps, 10))
	params.Set("clientip", ip)
	if arg.Tid > 0 {
		params.Set("tid", strconv.FormatInt(arg.Tid, 10))
	}
	if arg.Order != "" {
		params.Set("order", arg.Order)
	}
	if arg.Keyword != "" {
		params.Set("keyword", arg.Keyword)
	}
	var res struct {
		Code   int                   `json:"code"`
		Total  int64                 `json:"total"`
		Result *model.ArcSearchReply `json:"result"`
	}
	if err := d.searchClient.Get(ctx, _arcSearchURI, ip, params, &res); err != nil {
		return nil, 0, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, 0, errors.Wrapf(ecode.Int(res.Code), "url: %q, params: %+v, code: %d", _arcSearchURI, arg, res.Code)
	}
	return res.Result, res.Total, nil
}
