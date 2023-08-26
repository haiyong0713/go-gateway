package red

import (
	"context"
	"net/url"
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model"
	"go-gateway/app/app-svr/app-resource/interface/model/show"

	"github.com/pkg/errors"
)

// Dao is
type Dao struct {
	gameClient *bm.Client
}

// New red dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		gameClient: bm.NewClient(c.HTTPGame),
	}
	return
}

// RedDot func
func (d *Dao) RedDot(c context.Context, mid int64, redDotURL, platform, business string) (red *show.Red, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	if business == model.GotoGame {
		params.Set("platform", platform)
		params.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
	}
	var res struct {
		Code int       `json:"code"`
		Data *show.Red `json:"data"`
	}
	if err = d.gameClient.Get(c, redDotURL, "", params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), redDotURL+"?"+params.Encode())
		return
	}
	red = res.Data
	return
}
