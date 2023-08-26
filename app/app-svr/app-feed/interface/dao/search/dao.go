package search

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-feed/interface/conf"
	model "go-gateway/app/app-svr/app-feed/interface/model/search"

	"github.com/pkg/errors"
)

const (
	_arcSearchURI    = "/space/search/v2"
	_arcSearchType   = "sub_video"
	_additionalRanks = "-6"
)

// Dao is
type Dao struct {
	client       *bm.Client
	arcSearchURL string
}

// New is
func New(c *conf.Config) *Dao {
	d := &Dao{
		client:       bm.NewClient(c.HTTPSearch),
		arcSearchURL: c.Host.Search + _arcSearchURI,
	}
	return d
}

// ArcSpaceSearch is
func (d *Dao) ArcSpaceSearch(ctx context.Context, arg *model.ArcSearchParam) (*model.ArcSearchReply, int64, error) {
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
	if err := d.client.Get(ctx, d.arcSearchURL, ip, params, &res); err != nil {
		return nil, 0, err
	}
	if res.Code != ecode.OK.Code() {
		return nil, 0, errors.Wrapf(ecode.Int(res.Code), "url: %q, params: %+v, code: %d", d.arcSearchURL, arg, res.Code)
	}
	return res.Result, res.Total, nil
}
