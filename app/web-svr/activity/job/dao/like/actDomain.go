package like

import (
	"context"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
)

// SyncActDomainCache ...
func (dao *Dao) SyncActDomainCache(c context.Context, syncNum int) (err error) {
	var res struct {
		Code int `json:"code"`
		Data struct {
			Rows int64 `json:"rows"`
		} `json:"data"`
	}
	params := url.Values{}
	params.Set("sync_num", fmt.Sprint(syncNum))
	if err = dao.httpClient.Post(c, dao.syncActDomainURL, "", params, &res); err != nil {
		log.Error("SyncActDomainCache:d.httpClient.Post sync sync_num(%d) error(%v)", syncNum, err)
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), dao.syncActDomainURL+"?"+params.Encode())
	}
	return
}
