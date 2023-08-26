package favorite

import (
	"context"
	"net/url"
	"strconv"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-gateway/app/app-svr/app-car/interface/model/favorite"

	"github.com/pkg/errors"
)

const (
	_folderSpace = "/x/v3/fav/folder/space"
)

func (d *Dao) FolderSpace(c context.Context, mobiApp string, build int, accessKey, cookie, buvid, referer string, mid int64) ([]*favorite.Space, error) {
	var (
		ip = metadata.String(c, metadata.RemoteIP)
	)
	params := url.Values{}
	params.Set("access_key", accessKey)
	params.Set("up_mid", strconv.FormatInt(mid, 10))
	var res struct {
		Code int               `json:"code"`
		Data []*favorite.Space `json:"data"`
	}
	req, err := d.client.NewRequest("GET", d.folderSpace, ip, params)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Buvid", buvid)
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Referer", referer)
	if err = d.client.Do(c, req, &res); err != nil {
		return nil, err
	}
	if !ecode.EqualError(ecode.Int(res.Code), ecode.OK) {
		err := errors.Wrap(ecode.Int(res.Code), d.folderSpace+"?"+params.Encode())
		log.Error("%v", err)
		return nil, err
	}
	return res.Data, nil
}
