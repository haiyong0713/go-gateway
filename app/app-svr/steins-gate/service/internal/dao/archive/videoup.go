package archive

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/steins-gate/service/internal/model"

	"github.com/pkg/errors"
)

// VideoUpView get video up view data.
func (d *Dao) VideoUpView(c context.Context, aid int64) (view *model.VideoUpView, err error) {
	params := url.Values{}
	params.Set("aid", strconv.FormatInt(aid, 10))
	var res struct {
		Code int                `json:"code"`
		Data *model.VideoUpView `json:"data"`
	}
	if err = d.httpVideoClient.Get(c, d.videoUpViewURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.videoUpViewURL+"?"+params.Encode())
		return
	}
	view = res.Data
	return

}
