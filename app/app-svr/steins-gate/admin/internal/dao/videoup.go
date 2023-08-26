package dao

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go-common/library/xstr"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/net/metadata"

	"go-gateway/app/app-svr/steins-gate/admin/internal/model"

	"github.com/pkg/errors"
)

const (
	_videoUpViewURI = "/videoup/view"
	_dimensionURI   = "/v2/dash/hd/query"
)

// VideoUpView get video up view data.
func (d *Dao) VideoUpView(c context.Context, aid int64) (view *model.VideoUpView, err error) {
	params := url.Values{}
	params.Set("aid", strconv.FormatInt(aid, 10))
	var res struct {
		Code int                `json:"code"`
		Data *model.VideoUpView `json:"data"`
	}
	if err = d.client.Get(c, d.videoupURL, metadata.String(c, metadata.RemoteIP), params, &res); err != nil {
		return
	}
	if res.Code != ecode.OK.Code() {
		err = errors.Wrap(ecode.Int(res.Code), d.videoupURL+"?"+params.Encode())
		return
	}
	view = res.Data
	return
}

func (d *Dao) bvcSign(cid int64) (query string) {
	return d.bvcSignTool(fmt.Sprintf("%s=%d", "cid", cid))
}

//nolint:unused
func (d *Dao) bvcSignBatch(cids []int64) (query string) {
	return d.bvcSignTool(fmt.Sprintf("%s=%s", "cids", xstr.JoinInts(cids)))
}

func (d *Dao) bvcSignTool(cidQuery string) (query string) {
	kvs := []string{cidQuery, fmt.Sprintf("%s=%d", "timestamp", time.Now().Unix())}
	kvsStr := strings.Join(kvs, "&")
	mh := md5.Sum([]byte(kvsStr + "&key=" + d.c.Bvc.Key))
	sign := hex.EncodeToString(mh[:])
	return "?" + kvsStr + "&sign=" + sign

}
