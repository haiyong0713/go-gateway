package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"go-common/library/cache/redis"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/xstr"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"
	"go-gateway/app/web-svr/web/ecode"
	"go-gateway/app/web-svr/web/interface/model"

	"github.com/pkg/errors"
)

const (
	_arcRecommendURI  = "/recommand"
	_keyArcAppeal     = "arc_appeal_%d_%d"
	_relatedCmd       = "web_related"
	_relatedNeedOpera = "1"
	_webNeedRmRepeat  = "1"
)

// ArcReport add archive report
func (d *Dao) ArcReport(c context.Context, mid, aid, tp int64, reason, pics string) (err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("aid", strconv.FormatInt(aid, 10))
	params.Set("type", strconv.FormatInt(tp, 10))
	params.Set("reason", reason)
	params.Set("pics", pics)
	var res struct {
		Code int `json:"code"`
	}
	if err = d.httpW.Post(c, d.arcReportURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != xecode.OK.Code() {
		log.Error("archive report(%s) param(%v) ecode err(%d)", d.arcReportURL, params, res.Code)
		err = xecode.Int(res.Code)
	}
	return
}

// ArcAppeal add archive appeal.
func (d *Dao) ArcAppeal(c context.Context, mid int64, data map[string]string, business int) (err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	for name, value := range data {
		params.Set(name, value)
	}
	params.Set("mid", strconv.FormatInt(mid, 10))
	params.Set("business", strconv.Itoa(business))
	if v, ok := data["attach"]; ok && v != "" {
		params.Set("attachments", v)
	}
	var res struct {
		Code int `json:"code"`
	}
	if err = d.httpW.Post(c, d.arcAppealURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != xecode.OK.Code() {
		log.Error("archive report(%s) ecode err(%d)", d.arcAppealURL+"?"+params.Encode(), res.Code)
		err = xecode.Int(res.Code)
	}
	return
}

// AppealTags get appeal tags.
func (d *Dao) AppealTags(c context.Context, business int) (rs json.RawMessage, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("business", strconv.Itoa(business))
	var res struct {
		Code int             `json:"code"`
		Data json.RawMessage `json:"data"`
	}
	if err = d.httpR.Get(c, d.appealTagsURL, ip, params, &res); err != nil {
		return
	}
	if res.Code != xecode.OK.Code() {
		log.Error("archive report(%s) param(%v) ecode err(%d)", d.arcReportURL, params, res.Code)
		err = xecode.Int(res.Code)
	}
	rs = res.Data
	return
}

// RelatedAids get related aids from bigdata
func (d *Dao) RelatedAids(c context.Context, aid int64) (aids []int64, err error) {
	var (
		params = url.Values{}
		ip     = metadata.String(c, metadata.RemoteIP)
	)
	params.Set("key", strconv.FormatInt(aid, 10))
	var res struct {
		Code int `json:"code"`
		Data []*struct {
			Value string `json:"value"`
		} `json:"data"`
	}
	if err = d.httpR.Get(c, d.relatedURL, ip, params, &res); err != nil {
		log.Error("relate url(%s) error(%v) ", d.relatedURL+"?"+params.Encode(), err)
		return
	}
	if res.Code != 0 {
		log.Error("url(%s) res code(%d) or res.result(%v)", d.relatedURL+"?"+params.Encode(), res.Code, res.Data)
		err = xecode.Int(res.Code)
		return
	}
	if res.Data == nil {
		err = nil
		return
	}
	if len(res.Data) > 0 {
		if aids, err = xstr.SplitInts(res.Data[0].Value); err != nil {
			log.Error("relate aids url(%s) value(%s) error(%v)", d.relatedURL+"?"+params.Encode(), res.Data[0].Value, err)
		}
	}
	return
}

func keyArcAppealLimit(mid, aid int64) string {
	return fmt.Sprintf(_keyArcAppeal, mid, aid)
}

// SetArcAppealCache set arc appeal cache.
func (d *Dao) SetArcAppealCache(c context.Context, mid, aid int64) (err error) {
	key := keyArcAppealLimit(mid, aid)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	if err = conn.Send("SET", key, "1"); err != nil {
		log.Error("SetArcAppealCache conn.Send(SET, %s) error(%v)", key, err)
		return
	}
	if err = conn.Send("EXPIRE", key, d.redisAppealLimitExpire); err != nil {
		log.Error("SetArcAppealCache conn.Send(Expire, %s, %d) error(%v)", key, d.redisAppealLimitExpire, err)
		return
	}
	if err = conn.Flush(); err != nil {
		log.Error("SetArcAppealCache conn.Flush error(%v)", err)
		return
	}
	for i := 0; i < 2; i++ {
		if _, err = conn.Receive(); err != nil {
			log.Error("SetArcAppealCache conn.Receive() error(%v)", err)
			return
		}
	}
	return
}

// ArcAppealCache get arc appeal cache.
func (d *Dao) ArcAppealCache(c context.Context, mid, aid int64) (err error) {
	key := keyArcAppealLimit(mid, aid)
	conn := d.redisBak.Get(c)
	defer conn.Close()
	if _, err = redis.Bytes(conn.Do("GET", key)); err != nil {
		if err == redis.ErrNil {
			err = nil
			return
		}
		log.Error("ArcAppealCache conn.Do(GET, %s) error(%v)", key, err)
	}
	err = ecode.ArcAppealLimit
	return
}

func (d *Dao) Arcs(c context.Context, aids []int64) (res map[int64]*arcgrpc.Arc, err error) {
	arg := &arcgrpc.ArcsRequest{
		Aids: aids,
	}
	info, err := d.arcClient.Arcs(c, arg)
	if err != nil {
		err = errors.Wrapf(err, "%v", arg)
		return
	}
	res = info.Arcs
	return
}

func (d *Dao) ArcRecommends(ctx context.Context, aid, mid int64, needOperation, webRmRepeat, inActivity bool, buvid string) ([]*model.ArcRecommend, error) {
	if mid <= 0 && buvid == "" {
		return nil, errors.WithMessage(xecode.NothingFound, "no mid and buvid")
	}
	params := url.Values{}
	params.Set("cmd", _relatedCmd)
	params.Set("timeout", strconv.FormatInt(int64(time.Duration(d.c.HTTPClient.Read.Timeout)/time.Millisecond), 10))
	params.Set("from_av", strconv.FormatInt(aid, 10))
	if needOperation {
		params.Set("need_operation", _relatedNeedOpera)
	}
	if webRmRepeat {
		params.Set("web_rm_repeat", _webNeedRmRepeat)
	}
	if inActivity {
		params.Set("in_activity", "1")
	}
	if mid > 0 {
		params.Set("mid", strconv.FormatInt(mid, 10))
	}
	params.Set("buvid", buvid)
	var res struct {
		Code int                   `json:"code"`
		Data []*model.ArcRecommend `json:"data"`
	}
	if err := d.httpR.Get(ctx, d.arcRecommendURL, metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, errors.Wrapf(err, "d.httpR.Get aid:%d mid:%d", aid, mid)
	}
	if res.Code != 0 {
		if res.Code == -1 || res.Code == -3 || res.Code == -4 || res.Code == -5 || res.Code == -6 {
			log.Warn("ArcRecommends code %d aid:%d mid:%d", res.Code, aid, mid)
			return []*model.ArcRecommend{}, nil
		}
		return nil, errors.Wrapf(xecode.Int(res.Code), "d.httpR.Get code aid:%d mid:%d", aid, mid)
	}
	return res.Data, nil
}

func (d *Dao) ArcRedirectUrls(c context.Context, aids []int64) (map[int64]*arcgrpc.RedirectPolicy, error) {
	req := &arcgrpc.ArcsRedirectPolicyRequest{
		Aids: aids,
	}
	res, err := d.arcClient.ArcsRedirectPolicy(c, req)
	if err != nil {
		return nil, err
	}
	return res.GetRedirectPolicy(), nil
}
