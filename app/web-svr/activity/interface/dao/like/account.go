package like

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"

	"github.com/pkg/errors"
)

const (
	_checkTelURI = "/x/internal/passport-user/tel/is_new"
)

type PhoneTel struct {
	IsNew bool `json:"is_new"`
}

// CheckPhone .
func (d *Dao) CheckTel(c context.Context, mid int64) (isNew bool, err error) {
	params := url.Values{}
	params.Set("mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int `json:"code"`
		Data *PhoneTel
	}
	if err = d.client.Get(c, d.checkTelURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		err = errors.Wrapf(err, "d.client.Get(%s)", d.checkTelURL+"?"+params.Encode())
		return
	}
	if res.Code != ecode.OK.Code() {
		err = ecode.Int(res.Code)
		return
	}
	if res.Data != nil {
		isNew = res.Data.IsNew
		log.Info("User Tel Check success, mid(%d) isNew(%t)", mid, isNew)
	}
	return
}
