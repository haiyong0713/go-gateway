package music

import (
	"context"
	"net/url"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"

	xmetric "go-gateway/app/app-svr/app-dynamic/interface/model/metric"
	"go-gateway/app/app-svr/app-dynamic/interface/model/music"

	"github.com/pkg/errors"
)

const _audioDetail = "/audio/music-service-c/news/detail"

func (d *Dao) AudioDetail(c context.Context, ids []int64) (map[int64]*music.MusicResItem, error) {
	params := url.Values{}
	params.Set("rids", xstr.JoinInts(ids))
	audioDetailURL := d.c.Hosts.ApiCo + _audioDetail
	var ret struct {
		Code int                           `json:"code"`
		Msg  string                        `json:"msg"`
		Data map[int64]*music.MusicResItem `json:"data"`
	}
	if err := d.client.Get(c, audioDetailURL, "", params, &ret); err != nil {
		xmetric.DyanmicItemAPI.Inc(audioDetailURL, "request_error")
		log.Error("AudioDetail http GET(%s) failed, params:(%s), error(%+v)", audioDetailURL, params.Encode(), err)
		return nil, err
	}
	if ret.Code != 0 {
		log.Error("AudioDetail http GET(%s) failed, params:(%s), code: %v, msg: %v", audioDetailURL, params.Encode(), ret.Code, ret.Msg)
		xmetric.DyanmicItemAPI.Inc(audioDetailURL, "reply_code_error")
		err := errors.Wrapf(ecode.Int(ret.Code), "AudioDetail url(%v) code(%v) msg(%v)", audioDetailURL, ret.Code, ret.Msg)
		return nil, err
	}
	return ret.Data, nil
}
