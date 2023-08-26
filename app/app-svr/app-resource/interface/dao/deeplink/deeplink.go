package deeplink

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/taishan"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	model "go-gateway/app/app-svr/app-resource/interface/model/deeplink"

	"github.com/pkg/errors"
)

const (
	_huaweiLinkPath = "/ddl/deeplink"
	_huaweiOKCode   = 0
)

var ErrTaishanResultNil = errors.New("taishan Record.Columns is nil")

type Dao struct {
	client  *httpx.Client
	conf    *conf.Config
	taishan taishan.TaishanProxyClient
}

func New(c *conf.Config) (d *Dao) {
	var err error
	d = &Dao{
		client: httpx.NewClient(c.HTTPHuawei),
		conf:   c,
	}
	if d.taishan, err = taishan.NewClient(c.TaishanRPC); err != nil {
		panic(err)
	}
	return
}

func (d *Dao) DeepLinkHW(c context.Context, request *model.HWDeeplinkReq) (string, error) {
	bytesData, err := json.Marshal(request)
	if err != nil {
		return "", errors.WithStack(err)
	}
	requestUrl := d.conf.Host.Huawei + _huaweiLinkPath
	req, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewReader(bytesData))
	if err != nil {
		return "", errors.WithStack(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", authorizeHW(request.Bundle, d.conf.HuaweiSecretKey, time.Now().Unix()))
	var res struct {
		Status   int    `json:"status"`
		Msg      string `json:"message"`
		DeepLink string `json:"deepLink"`
	}
	if err = d.client.Do(c, req, &res); err != nil {
		return "", errors.Errorf("deepLinkHW POST request failed: url(%s), body(%s), err(%+v)", requestUrl, string(bytesData), err)
	}
	if res.Status != _huaweiOKCode {
		return "", errors.Errorf("invalid response(%+v), url(%s), body(%s)", res, requestUrl, string(bytesData))
	}
	deeplink, err := url.QueryUnescape(res.DeepLink)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return deeplink, nil
}

func authorizeHW(bundle, secretKey string, nonce int64) string {
	key := fmt.Sprintf("%s:%s", bundle, secretKey)
	data := fmt.Sprintf("%d:%s", nonce, url.QueryEscape(_huaweiLinkPath))
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	signature := strings.ToLower(hex.EncodeToString(mac.Sum(nil)))
	return fmt.Sprintf("Digest username=%s,nonce=%d,response=%s,algorithm=HmacSHA256", bundle, nonce, signature)
}

func checkTaishanRecordErr(r *taishan.Record) error {
	if r == nil || r.Status == nil || r.Status.ErrNo == 404 || r.Columns == nil || len(r.Columns) == 0 || r.Columns[0] == nil || r.Columns[0].Value == nil {
		return ErrTaishanResultNil
	}
	return nil
}

func (d *Dao) deepLinkAIGetGroup(ctx context.Context, buvid string, innerType int64) (*model.AiDeeplinkGroupRsp, error) {
	req := &taishan.GetReq{
		Table: d.conf.DeeplinkTaishanCfg.UserTable,
		Auth: &taishan.Auth{
			Token: d.conf.DeeplinkTaishanCfg.UserTableToken,
		},
		Record: &taishan.Record{Key: []byte(fmt.Sprintf("%s_%d", buvid, innerType))},
	}
	resp, err := d.taishan.Get(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "deepLinkAIGetGroup d.taishan.Get error, key=%+v", req.Record.GetKey())
	}
	if err = checkTaishanRecordErr(resp.Record); err != nil {
		return nil, err
	}
	res := &model.AiDeeplinkGroupRsp{}
	if err = json.Unmarshal(resp.Record.Columns[0].Value, &res); err != nil {
		return nil, errors.Wrapf(err, "deeplink Unmarshal error key=%s", buvid)
	}
	return res, nil
}

//nolint:ineffassign
func (d *Dao) DeepLinkAI(ctx context.Context, buvid string, meta *model.AiDeeplinkMaterial) (*model.AiDeeplink, error) {
	if !d.conf.DeeplinkTaishanCfg.Open {
		return nil, nil
	}
	// 第一段先查用户表分组
	groupRsp, err := d.deepLinkAIGetGroup(ctx, buvid, meta.InnerType)
	if err != nil {
		log.Warn("d.deepLinkAIGetGroup failed to get group buvid=%s, err=%+v", buvid, err)
		err = nil
	}
	if groupRsp != nil {
		meta.AbId = groupRsp.AbId
	}
	// 如果用户表查询失败，则走在线分组
	if meta.AbId == "" {
		meta.AbId = resolveDeeplinkMetaAbIdOnline(buvid)
	}
	// 第二段查稿件表的deeplink
	key := taishanDeepLinkAIKey(meta)
	req := &taishan.GetReq{
		Table: d.conf.DeeplinkTaishanCfg.ArchiveTable,
		Auth: &taishan.Auth{
			Token: d.conf.DeeplinkTaishanCfg.ArchiveTableToken,
		},
		Record: &taishan.Record{Key: key},
	}
	resp, err := d.taishan.Get(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "d.taishan.Get error, key=%s", key)
	}
	if err = checkTaishanRecordErr(resp.Record); err != nil {
		return nil, err
	}
	res := &model.AiDeeplink{}
	if err = json.Unmarshal(resp.Record.Columns[0].Value, &res); err != nil {
		return nil, errors.Wrapf(err, "deeplink Unmarshal error key=%s", key)
	}
	return res, nil
}

func resolveDeeplinkMetaAbIdOnline(buvid string) string {
	// 如果用户表查询失败，则走在线分组
	key := fmt.Sprintf("%s_yuzhuang", buvid)
	h := md5.New()
	_, _ = h.Write([]byte(key))
	b, err := strconv.ParseUint(hex.EncodeToString(h.Sum(nil))[18:], 16, 64)
	if err != nil {
		log.Error("resolveDeeplinkMetaAbId strconv.ParseUint failed buvid=%s", buvid)
	}
	return fmt.Sprintf("yuz_%d", b%10)
}

func taishanDeepLinkAIKey(meta *model.AiDeeplinkMaterial) []byte {
	return []byte(fmt.Sprintf("%s-%s-%d-%s-%s", meta.AbId, meta.InnerId, meta.InnerType, meta.SourceName, meta.AccountId))
}
