package like

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/interface/model/like"

	"github.com/pkg/errors"
)

const (
	_showInfoURL     = "/router/rest"
	_showInfoMethod  = "taobao.film.bilishowinfo.get"
	_showInfoVersion = "2.0"
	_fateSwitchKey   = "fate_switch_total"
	_fateConfKey     = "fate_conf"
)

var fateReg = regexp.MustCompile(`itemprop="interactionCount" content="([\d]+)"`)

// TppShowInfo get tao bao film want and view total.
func (d *Dao) TppShowInfo(c context.Context, id int64) (totalView, wantCount int64, err error) {
	var req *http.Request
	params := url.Values{}
	params.Set("method", _showInfoMethod)
	params.Set("show_id", strconv.FormatInt(id, 10))
	fullURL := d.fateShowInfoURL + "?" + d.showSign(params)
	if req, err = http.NewRequest(http.MethodGet, fullURL, nil); err != nil {
		log.Error("ShowInfo http.NewRequest url(%s) error(%v)", fullURL, err)
		return
	}
	var res struct {
		FilmBilishowinfoGetResponse struct {
			ShowRelateVideoPvTotalCount int64  `json:"show_relate_video_pv_total_count"`
			WantCount                   int64  `json:"want_count"`
			RequestID                   string `json:"request_id"`
		} `json:"film_bilishowinfo_get_response"`
		ErrorResponse *struct {
			SubMsg  string `json:"sub_msg"`
			Code    int    `json:"code"`
			SubCode string `json:"sub_code"`
			Msg     string `json:"msg"`
		} `json:"error_response"`
	}
	if err = d.httpFate.Do(c, req, &res); err != nil {
		log.Error("TppShowInfo d.httpFate.Do uri(%s) error(%v)", fullURL, err)
		return
	}
	if res.ErrorResponse != nil {
		log.Error("TppShowInfo error res.ErrorResponse(%+v)", res.ErrorResponse)
		err = errors.Wrapf(ecode.Int(res.ErrorResponse.Code), "subcode:%s msg:%s", res.ErrorResponse.SubCode, res.ErrorResponse.Msg)
		return
	}
	totalView = res.FilmBilishowinfoGetResponse.ShowRelateVideoPvTotalCount
	wantCount = res.FilmBilishowinfoGetResponse.WantCount
	return
}

// QQShowInfo get qq total view.
func (d *Dao) QQShowInfo(c context.Context, url string) (total int64, err error) {
	var (
		req *http.Request
		bs  []byte
	)
	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		log.Error("QQShowInfo http.NewRequest url(%s) error(%v)", url, err)
		return
	}
	if bs, err = d.httpFate.Raw(c, req); err != nil {
		log.Error("QQShowInfo d.httpFate.Do uri(%s) error(%v)", url, err)
		return
	}
	match := fateReg.FindStringSubmatch(string(bs))
	if len(match) > 1 {
		totalStr := match[1]
		total, err = strconv.ParseInt(totalStr, 10, 64)
	} else {
		log.Error("QQShowInfo d.httpFate.Do uri(%s) nothing match", url)
		err = ecode.NothingFound
	}
	return
}

func (d *Dao) showSign(params url.Values) string {
	params.Set("app_key", d.c.Fate.AppKey)
	params.Set("timestamp", time.Now().Format("2006-01-02 15:04:05"))
	params.Set("format", "json")
	params.Set("v", _showInfoVersion)
	params.Set("sign_method", "md5")
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var buf strings.Builder
	for _, key := range keys {
		vs := params[key]
		buf.WriteString(key)
		for _, v := range vs {
			buf.WriteString(v)
		}
	}
	mh := md5.Sum([]byte(d.c.Fate.Secret + buf.String() + d.c.Fate.Secret))
	params.Set("sign", strings.ToUpper(hex.EncodeToString(mh[:])))
	uri := params.Encode()
	if strings.IndexByte(uri, '+') > -1 {
		uri = strings.Replace(uri, "+", "%20", -1)
	}
	return uri
}

// SetFateInfoCache .
func (d *Dao) SetFateInfoCache(c context.Context, key string, count int64) (err error) {
	var (
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if _, err = conn.Do("SET", key, count); err != nil {
		log.Error("SetFateInfoCache conn.Send(SET, %s, %s) error(%v)", key, string(bs), err)
	}
	return
}

// SetFateSwitchCache .
func (d *Dao) SetFateSwitchCache(c context.Context, data *like.FateSwitch) (err error) {
	var (
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		return
	}
	if _, err = conn.Do("SET", _fateSwitchKey, bs); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", _fateSwitchKey, string(bs), err)
	}
	return
}

// SetFateConfCache .
func (d *Dao) SetFateConfCache(c context.Context, data *like.FateConfData) (err error) {
	var (
		conn = d.redis.Get(c)
		bs   []byte
	)
	defer conn.Close()
	if bs, err = json.Marshal(data); err != nil {
		return
	}
	if _, err = conn.Do("SET", _fateConfKey, bs); err != nil {
		log.Error("conn.Send(SET, %s, %s) error(%v)", _fateConfKey, string(bs), err)
	}
	return
}
