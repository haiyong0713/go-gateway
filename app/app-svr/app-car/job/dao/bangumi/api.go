package bangumi

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/job/model/bangumi"

	"github.com/pkg/errors"
)

const (
	_channelcontent          = "/ext/internal/archive/channel/content"
	_channelcontentchange    = "/ext/internal/archive/channel/content/change"
	_channelcontentoffshelve = "/ext/internal/archive/channel/content/offshelve"
)

func (d *Dao) ChannelContent(c context.Context, pn, ps, seasonType int, bsource string) ([]*bangumi.Content, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("page_no", strconv.Itoa(pn))
	params.Set("page_size", strconv.Itoa(ps))
	params.Set("season_type", strconv.Itoa(seasonType))
	params.Set("bsource", bsource)
	var res struct {
		Code   int                `json:"code"`
		Result []*bangumi.Content `json:"result"`
		Total  int                `json:"total"`
	}
	if err := d.client.Get(c, d.channelcontent, ip, params, &res); err != nil {
		return nil, err
	}
	// 10
	// 返回结果为空或者已被删除
	// 11
	// 该内容不在新番后台内
	if res.Code != ecode.OK.Code() && res.Code != 10 && res.Code != 11 {
		return nil, errors.Wrap(ecode.Int(res.Code), d.channelcontent+"?"+params.Encode())
	}
	return res.Result, nil
}

func (d *Dao) ChannelContentChange(c context.Context, pn, ps, seasonType int, bsource string, starttime, endtime time.Time) ([]*bangumi.Content, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("page_no", strconv.Itoa(pn))
	params.Set("page_size", strconv.Itoa(ps))
	params.Set("season_type", strconv.Itoa(seasonType))
	params.Set("bsource", bsource)
	params.Set("start_ts", strconv.FormatInt(starttime.Unix(), 10))
	params.Set("end_ts", strconv.FormatInt(endtime.Unix(), 10))
	var res struct {
		Code   int                `json:"code"`
		Result []*bangumi.Content `json:"result"`
		Total  int                `json:"total"`
	}
	if err := d.client.Get(c, d.channelcontentchange, ip, params, &res); err != nil {
		return nil, err
	}
	// 10
	// 返回结果为空或者已被删除
	// 11
	// 该内容不在新番后台内
	if res.Code != ecode.OK.Code() && res.Code != 10 && res.Code != 11 {
		return nil, errors.Wrap(ecode.Int(res.Code), d.channelcontentchange+"?"+params.Encode())
	}
	return res.Result, nil
}

func (d *Dao) ChannelContentoffshelve(c context.Context, pn, ps, seasonType int) ([]*bangumi.Offshelve, error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("page_no", strconv.Itoa(pn))
	params.Set("page_size", strconv.Itoa(ps))
	params.Set("season_type", strconv.Itoa(seasonType))
	var res struct {
		Code   int                  `json:"code"`
		Result []*bangumi.Offshelve `json:"result"`
	}
	if err := d.client.Get(c, d.channelcontentoffshelve, ip, params, &res); err != nil {
		return nil, err
	}
	// 10
	// 返回结果为空或者已被删除
	// 11
	// 该内容不在新番后台内
	if res.Code != ecode.OK.Code() && res.Code != 10 && res.Code != 11 {
		return nil, errors.Wrap(ecode.Int(res.Code), d.channelcontentoffshelve+"?"+params.Encode())
	}
	return res.Result, nil
}
