package push

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"go-gateway/app/app-svr/app-feed/admin/model/selected"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"go-common/library/log"
	"go-common/library/xstr"
)

const (
	_pinkVersion     = 1
	_linkTypeBrowser = 7
)

type _response struct {
	Code int `json:"code"`
	Data int `json:"data"`
}

// NoticeUser pushs the notification to users.
func (d *Dao) NoticeUser(mids []int64, uuid string, serie *selected.Serie) (err error) {
	if serie == nil {
		log.Error("NoticeUser serie is nil mids(%v), uuid(%s)", mids, uuid)
		return
	}
	var (
		cfg       = d.c.WeeklySelected.Push
		clientCfg = d.c.HTTPClient.Push
	)
	if serie.ShareSubtitle == "" {
		serie.PushTitle = cfg.Title
	}
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	//nolint:errcheck
	w.WriteField("app_id", strconv.Itoa(_pinkVersion))
	//nolint:errcheck
	w.WriteField("business_id", cfg.BusinessID)
	//nolint:errcheck
	w.WriteField("alert_title", serie.PushTitle)
	//nolint:errcheck
	w.WriteField("alert_body", serie.PushSubtitle)
	//nolint:errcheck
	w.WriteField("mids", xstr.JoinInts(mids))
	//nolint:errcheck
	w.WriteField("link_type", fmt.Sprintf("%d", _linkTypeBrowser))
	//nolint:errcheck
	w.WriteField("link_value", cfg.Link)
	//nolint:errcheck
	w.WriteField("uuid", uuid)
	w.Close()
	query := map[string]string{
		"ts":     strconv.FormatInt(time.Now().Unix(), 10),
		"appkey": clientCfg.Key,
	}
	query["sign"] = signature(query, clientCfg.Secret)
	url := fmt.Sprintf("%s?ts=%s&appkey=%s&sign=%s", d.pushURL, query["ts"], query["appkey"], query["sign"])
	req, err := http.NewRequest(http.MethodPost, url, buf)
	log.Info("[SeriePublish] Push Body %s", buf.String())
	if err != nil {
		log.Error("http.NewRequest(%s) error(%v)", url, err)
		return
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("token=%s", cfg.Token))
	res := &_response{}
	if err = d.client.Do(context.TODO(), req, &res); err != nil {
		log.Error("httpClient.Do() error(%v)", err)
		return
	}
	if res.Code != 0 || res.Data == 0 {
		log.Error("push failed mids_total(%d) body(%s) response(%+v)", len(mids), serie.PushSubtitle, res)
	} else {
		log.Info("push success mids_total(%d) body(%s) response(%+v)", len(mids), serie.PushSubtitle, res)
	}
	return
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
	//nolint:errcheck
	io.WriteString(h, buf.String()+secret)
	return fmt.Sprintf("%x", h.Sum(nil))
}
