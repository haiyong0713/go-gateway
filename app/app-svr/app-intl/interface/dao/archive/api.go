package archive

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"

	"go-gateway/app/app-svr/app-intl/interface/model/view"

	"github.com/pkg/errors"
)

const (
	_realteURL     = "/recsys/related"
	_commercialURL = "/x/internal/creative/arc/commercial"
	_relateRecURL  = "/recommand"
	_playURL       = "/playurl/batch"
)

// RelateAids get relate by aid
func (d *Dao) RelateAids(c context.Context, aid int64) (aids []int64, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("key", strconv.FormatInt(aid, 10))
	var res struct {
		Code int `json:"code"`
		Data []*struct {
			Value string `json:"value"`
		} `json:"data"`
	}
	if err = d.client.Get(c, d.realteURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.realteURL+"?"+params.Encode())
		return
	}
	if len(res.Data) != 0 && res.Data[0] != nil {
		if aids, err = xstr.SplitInts(res.Data[0].Value); err != nil {
			err = errors.Wrap(err, res.Data[0].Value)
		}
	}
	return
}

// NewRelateAids relate online recommend 在线实时推荐
func (d *Dao) NewRelateAids(c context.Context, aid, mid int64, build, autoplay int, buvid, from, trackid string, plat int8, zoneID int64) (res *view.RelateRes, returnCode string, err error) {
	ip := metadata.String(c, metadata.RemoteIP)
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("from", "2")
	params.Set("cmd", "related")
	params.Set("timeout", "200")
	params.Set("plat", strconv.Itoa(int(plat)))
	params.Set("build", strconv.Itoa(build))
	params.Set("buvid", buvid)
	params.Set("from_av", strconv.FormatInt(aid, 10))
	params.Set("request_cnt", "40")
	params.Set("source_page", from)
	params.Set("auto_play", strconv.Itoa(autoplay))
	params.Set("from_trackid", trackid)
	params.Set("need_dalao", "1")
	params.Set("zone_id", strconv.FormatInt(zoneID, 10))
	log.Info("dalaotest url(%s)", d.relateRecURL+"?"+params.Encode())
	if err = d.client.Get(c, d.relateRecURL, ip, params, &res); err != nil {
		returnCode = "500"
		return
	}
	returnCode = strconv.Itoa(res.Code)
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.relateRecURL+"?"+params.Encode())
		return
	}
	return
}
