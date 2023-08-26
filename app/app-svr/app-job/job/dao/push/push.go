package push

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	"go-gateway/app/app-svr/app-job/job/model"
	"go-gateway/app/app-svr/app-job/job/model/resource"

	"github.com/pkg/errors"
)

const (
	_bcResourceTargetPath = "/bilibili.broadcast.message.main.Resource/TopActivity"
	_topActivityMng       = 2
	_mobiApp              = "mobi_app"
	_build                = "build"
	_device               = "device"
)

type _response struct {
	Code int `json:"code"`
	Data int `json:"data"`
}

func signature(params map[string]string, secret string) string {
	keys := []string{}
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buf := bytes.Buffer{}
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(k) + "=")
		buf.WriteString(url.QueryEscape(params[k]))
	}
	h := md5.New()
	if _, err := io.WriteString(h, buf.String()+secret); err != nil {
		log.Error("signature io.WriteString error(%+v)", err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// nolint:errcheck
func (d *Dao) Push(c context.Context, params *model.PushParam, token string) (err error) {
	if params == nil {
		log.Error("Push params(%v) nil", params)
		return
	}
	var (
		ms  = xstr.JoinInts(params.MIDs)
		bvs = strings.Join(params.Buvids, ",")
	)
	if ms == "" && bvs == "" {
		log.Error("Push mids(%v) && buvids(%v) all empty", params.MIDs, params.Buvids)
		return
	}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	w.WriteField("app_id", strconv.FormatInt(params.AppID, 10))
	w.WriteField("business_id", strconv.FormatInt(params.BusinessID, 10))
	if params.AlertTitle != "" {
		w.WriteField("alert_title", params.AlertTitle)
	}
	if params.AlertBody != "" {
		w.WriteField("alert_body", params.AlertBody)
	}
	// 服务端只支持mid和buvid二选一,buvid优先
	if bvs != "" {
		w.WriteField("buvids", bvs)
	} else if ms != "" {
		w.WriteField("mids", ms)
	}
	w.WriteField("link_type", strconv.FormatInt(params.LinkType, 10))
	w.WriteField("link_value", params.LinkValue)
	w.WriteField("uuid", params.UUID)
	w.WriteField("pass_through", strconv.Itoa(params.PassThrough))
	w.Close()
	query := map[string]string{
		"ts":     strconv.FormatInt(time.Now().Unix(), 10),
		"appkey": d.c.HTTPClient.Key,
	}
	query["sign"] = signature(query, d.c.HTTPClient.Secret)
	url := fmt.Sprintf("%s?ts=%s&appkey=%s&sign=%s", d.pushURL, query["ts"], query["appkey"], query["sign"])
	log.Info("Push url %v", url)
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		log.Error("%v", err)
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("token=%s", token))
	res := &_response{}
	if err = d.client.Do(c, req, &res); err != nil {
		log.Error("httpClient.Do() error(%v)", err)
		return
	}
	if res.Code != 0 || res.Data == 0 {
		err = errors.Wrap(ecode.Int(res.Code), url)
	}
	return
}

// PushEntry is s10 push
func (d *Dao) PushEntry(c context.Context, entryMsg *resource.EntryMsg) {
	eg := errgroup.WithContext(c)
	for _, p := range entryMsg.Plat {
		if p == nil {
			continue
		}
		ma, dv := model.PlatToMobiApp(p.Plat)
		if ma == "" {
			continue
		}
		pp := &resource.PlatLimit{}
		*pp = *p
		eg.Go(func(ctx context.Context) error {
			return d.BroadcastEntry(ctx, entryMsg, pp, ma, dv)
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("broadcast PushEntry eg err(%+v)", err)
	}
}
