package rcmd

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/naming/discovery"
	bm "go-common/library/net/http/blademaster"
	"go-common/library/net/http/blademaster/resolver"
	"go-gateway/app/web-svr/web/interface/conf"
	"go-gateway/app/web-svr/web/interface/model/rcmd"

	"github.com/pkg/errors"
)

const _rcmd = "/recommand"

type Dao struct {
	c          *conf.Config
	rcmdURI    string
	rcmdClient *bm.Client
}

func New(c *conf.Config) *Dao {
	return &Dao{
		c:          c,
		rcmdURI:    c.Host.RcmdDiscovery + _rcmd,
		rcmdClient: bm.NewClient(c.HTTPClient.Rcmd, bm.SetResolver(resolver.New(nil, discovery.Builder()))),
	}
}

func (d *Dao) TopRcmd(ctx context.Context, mid int64, freshType, ps int, ip, buvid string, isFeed, freshIdx, freshIdx1h, yNum int64, feedVersion string) (data []*rcmd.AITopRcmd, useFeature json.RawMessage, code int, err error) {
	timeout := int64(time.Duration(d.c.Rcmd.Timeout) / time.Millisecond)
	params := url.Values{}
	params.Add("cmd", "web_pegasus")
	params.Add("timeout", strconv.FormatInt(timeout, 10))
	params.Add("mid", strconv.FormatInt(mid, 10))
	params.Add("request_cnt", strconv.Itoa(ps))
	params.Add("fresh_type", strconv.Itoa(freshType))
	params.Add("buvid", buvid)
	params.Add("is_feed", strconv.FormatInt(isFeed, 10))
	params.Add("fresh_idx", strconv.FormatInt(freshIdx, 10))
	params.Add("fresh_idx_1h", strconv.FormatInt(freshIdx1h, 10))
	params.Add("feed_version", feedVersion)
	params.Add("y_num", strconv.FormatInt(yNum, 10))
	var res struct {
		Code        int               `json:"code"`
		Data        []*rcmd.AITopRcmd `json:"data"`
		UserFeature json.RawMessage   `json:"user_feature"`
	}
	if err := d.rcmdClient.Get(ctx, d.rcmdURI, ip, params, &res); err != nil {
		return nil, nil, -500, errors.Wrap(err, d.rcmdURI+"?"+params.Encode())
	}
	if res.Code != 0 {
		return nil, nil, res.Code, errors.Wrap(ecode.Int(res.Code), d.rcmdURI+"?"+params.Encode())
	}
	return res.Data, res.UserFeature, 0, nil
}

func (d *Dao) TopFeedRcmd(ctx context.Context, req *rcmd.TopRcmdReq) (*rcmd.TopFeedRcmdRep, int, error) {
	timeout := int64(time.Duration(d.c.Rcmd.Timeout) / time.Millisecond)
	params := url.Values{}
	params.Add("cmd", "web_pegasus")
	params.Add("timeout", strconv.FormatInt(timeout, 10))
	params.Add("mid", strconv.FormatInt(req.Mid, 10))
	params.Add("request_cnt", strconv.Itoa(req.Ps))
	params.Add("fresh_type", strconv.Itoa(req.FreshType))
	params.Add("buvid", req.Buvid)
	params.Add("is_feed", strconv.Itoa(req.IsFeed))
	params.Add("fresh_idx", strconv.Itoa(req.FreshIdx))
	params.Add("fresh_idx_1h", strconv.Itoa(req.FreshIdx))
	params.Add("feed_version", req.FeedVersion)
	params.Add("y_num", strconv.Itoa(req.YNum))
	params.Add("homepage_ver", strconv.Itoa(req.HomepageVer))
	params.Add("fetch_row", strconv.Itoa(req.FetchRow))
	params.Add("brush", strconv.Itoa(req.Brush))
	params.Add("s_id", req.Sid)
	params.Add("country", req.Country)
	params.Add("province", req.Province)
	params.Add("city", req.City)
	params.Add("ip", req.Ip)
	ua := strings.Replace(req.UserAgent, " ", "", -1)
	params.Add("ua", ua)
	//params.Add("session", req.Session)
	var resource int = d.c.Rcmd.AdResource[strconv.Itoa(-1)]
	if r, ok := d.c.Rcmd.AdResource[strconv.Itoa(req.FreshType)]; ok {
		resource = r
	}
	params.Add("resource", strconv.Itoa(resource))
	var res *rcmd.TopFeedRcmdRep
	if err := d.rcmdClient.Get(ctx, d.rcmdURI, req.Ip, params, &res); err != nil {
		return nil, -500, errors.Wrap(err, d.rcmdURI+"?"+params.Encode())
	}
	if res.Code != 0 {
		return nil, res.Code, errors.Wrap(ecode.Int(res.Code), d.rcmdURI+"?"+params.Encode())
	}
	return res, 0, nil
}
