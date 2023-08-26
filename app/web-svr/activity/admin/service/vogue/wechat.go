package vogue

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

const (
	_weChatHost        = "https://api.weixin.qq.com/cgi-bin"
	_alarmUrl          = "http://bap.bilibili.co/api/v1/message/add"
	_toShortPathAction = "long2short"
)

// alarmMsgPrefix
func alarmMsgPrefix(alarmTag string) string {
	log.Info(os.Getenv("DEPLOY_ENV"))
	if os.Getenv("DEPLOY_ENV") != "prod" {
		return "【测试环境】" + alarmTag
	}
	return "" + alarmTag
}

// weChatTokenUrl 获取微信access token接口url.
func weChatTokenUrl() string {
	return _weChatHost + "/token"
}

// weChatToShortPathUrl 获取微信长链接转短接口url.
func weChatToShortPathUrl(token string) string {
	return fmt.Sprintf(_weChatHost+"/shorturl?access_token=%s", token)
}

// WeChatHostMonitor 微信域名监测
func (s *Service) WeChatHostMonitor(c context.Context) {
	for {
		log.Info("WeChatHostMonitor start time(%v), env (%s)", time.Now(), os.Getenv("DEPLOY_ENV"))
		blocked, err := s.WeChatCheck(c, s.c.Alarm.WeChatShareHost)
		if err != nil {
			if ecode.Equal(ecode.Cause(err), ecode.AccessKeyErr) {
				_ = s.SendWeChatWorkMsg(c, alarmMsgPrefix(s.c.Alarm.AlarmTag)+"微信token获取失败")
			} else if ecode.Equal(ecode.Cause(err), ecode.LimitExceed) {
				_ = s.SendWeChatWorkMsg(c, alarmMsgPrefix(s.c.Alarm.AlarmTag)+"微信token调用次数限制")
			}
		}
		if blocked {
			_ = s.SendWeChatWorkMsg(c, alarmMsgPrefix(s.c.Alarm.AlarmTag)+"微信域名已被封禁")
		}

		time.Sleep(time.Duration(s.c.Alarm.WeChatMonitorTick))
	}
}

// WeChatBlockStatus 获取微信封禁状态
func (s *Service) WeChatBlockStatus(c context.Context, req *voguemdl.WeChatCheckReq) (resp *voguemdl.WeChatBlockStatusResp, err error) {
	resp = &voguemdl.WeChatBlockStatusResp{}
	if req.Refresh == 1 {
		_, err = s.WeChatCheck(c, s.c.Alarm.WeChatShareHost)
		if err != nil {
			log.Error("WeChatBlockStatus WeChatCheck err(%v), host(%s)", err, s.c.Alarm.WeChatShareHost)
			return
		}
	}
	resp.Blocked = s.weChatBlockStatus
	return
}

func (s *Service) SetWeChatBlockStatus(c context.Context, blocked bool) {
	s.weChatBlockStatus = blocked
	return
}

// WeChatToken 生成微信access token
func (s *Service) WeChatToken(c context.Context) (token string, err error) {
	var (
		req    *http.Request
		params = url.Values{}
		res    *struct {
			ErrCode     int    `json:"errcode"`
			ErrMsg      string `json:"errmsg"`
			AccessToken string `json:"access_token"`
			ExpiresIn   int64  `json:"expires_in"`
		}
	)
	params.Set("grant_type", "client_credential")
	params.Set("appid", s.c.Wechat.AppId)
	params.Set("secret", s.c.Wechat.Secret)
	if req, err = http.NewRequest("GET", weChatTokenUrl()+"?"+params.Encode(), nil); err != nil {
		log.Error("WeChatToken http.NewRequest err(%v), params(%v)", err, params)
		return
	}
	if err = s.httpClient.Do(c, req, &res); err != nil {
		log.Error("WeChatToken s.httpClient.Do err(%v), req(%v)", err, req)
		return
	}
	if res.ErrCode != ecode.OK.Code() {
		log.Error("WeChatToken errcode error(%d) msg(%s)", res.ErrCode, res.ErrMsg)
		err = ecode.RequestErr
		return
	}
	token = res.AccessToken
	return
}

// WeChatCheck check 微信转链是否被封
func (s *Service) WeChatCheck(c context.Context, host string) (blocked bool, err error) {
	var (
		shorturl string
	)
	blocked = false
	accessToken, err := s.WeChatToken(c)
	log.Info("WeChatCheck AccessToken: %s", accessToken)
	if err != nil {
		log.Error("WeChatCheck AccessToken get failed, err(%v)", err)
		err = ecode.AccessKeyErr
		return
	}
	if shorturl, err = s.ToWeChatShortPath(c, host, accessToken); err != nil {
		log.Warn("WeChatCheck ToWeChatShortPath link(%s) token(%s) err(%v)", host, accessToken, err)
	}
	log.Info("WeChatCheck ShortUrl: %s", shorturl)
	if ecode.Equal(ecode.Cause(err), ecode.LimitExceed) {
		log.Warn(alarmMsgPrefix(s.c.Alarm.AlarmTag) + "微信token调用次数限制")
	}
	if blocked, err = s.GetContent(c, shorturl); err != nil {
		log.Warn("WeChatCheck GetContent shorturl(%s) err(%v)", shorturl, err)
	}
	log.Info("WeChatCheck blocked: %b", blocked)
	if blocked {
		s.SetWeChatBlockStatus(c, true)
		log.Warn("WeChatCheck GetContent shorturl(%s) err(%v)", shorturl, err)
		log.Warn(alarmMsgPrefix(s.c.Alarm.AlarmTag) + "微信域名已被封禁")
	} else {
		s.SetWeChatBlockStatus(c, false)
	}
	return
}

// ToWeChatShortPath
func (s *Service) ToWeChatShortPath(c context.Context, link, token string) (path string, err error) {
	var (
		result *struct {
			ShortUrl string `json:"short_url"`
			ErrCode  int    `json:"errcode"`
		}
	)
	params := map[string]string{
		"action":   _toShortPathAction,
		"long_url": link,
	}
	b, err := json.Marshal(params)
	if err != nil {
		log.Error("ToWeChatShortPath json.Marshal err(%v), params(%v)", err, params)
		return
	}
	req, err := http.NewRequest(http.MethodPost, weChatToShortPathUrl(token), bytes.NewReader(b))
	if err != nil {
		log.Error("ToWeChatShortPath http.NewRequest err(%v), params(%v)", err, params)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err = s.httpClient.Do(c, req, &result); err != nil {
		log.Error("ToWeChatShortPath s.httpClient.Do err(%v), params(%v)", err, params)
		return
	}
	if result.ErrCode != 0 {
		if result.ErrCode == 45009 {
			err = ecode.LimitExceed
		} else {
			err = ecode.AccessDenied
		}
		log.Error("s.ToWeChatShortPath res(%v)", result)
		return
	}
	path = result.ShortUrl
	return
}

// GetContent
func (s *Service) GetContent(c context.Context, path string) (blocked bool, err error) {
	var (
		content []byte
		req     *http.Request
	)
	if req, err = http.NewRequest(http.MethodGet, path, nil); err != nil {
		log.Error("GetContent path(%s) err(%v)", path, err)
		return
	}
	if content, err = s.httpClient.Raw(c, req); err != nil {
		log.Error("GetContent path(%s) error(%v)", path, err)
		return
	}
	if strings.Contains(string(content), "Tips") {
		blocked = true
	}
	return
}

// SendWeChatWorkMsg 发送企业微信消息.
func (s *Service) SendWeChatWorkMsg(c context.Context, msg string) (err error) {
	params := map[string]string{
		"content":   msg,
		"timestamp": strconv.FormatInt(time.Now().Unix(), 10),
		"token":     s.c.Alarm.WeChatToken,
		"type":      "wechat",
		"username":  s.c.Alarm.Username,
		"url":       "",
	}
	params["signature"] = s.sign(params)
	bs, err := json.Marshal(params)
	if err != nil {
		log.Error("SendWeChatWorkMsg json.Marshal err(%v), params(%v)", err, params)
		return
	}
	req, err := http.NewRequest(http.MethodPost, _alarmUrl, bytes.NewReader(bs))
	if err != nil {
		log.Error("SendWeChatWorkMsg http.NewRequest err(%v), params(%v)", err, params)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	res := voguemdl.WeChatResp{}
	if err = s.httpClient.Do(context.TODO(), req, &res); err != nil {
		log.Error("SendWeChatWorkMsg s.httpClient.Do err(%v), params(%v)", err, params)
		return
	}
	if res.Status != 0 {
		log.Error("SendWeChatWorkMsg response failed, res(%v), params(%v)", res, params)
	}
	return
}

// sign
func (s *Service) sign(params map[string]string) string {
	var keys []string
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
	io.WriteString(h, buf.String()+s.c.Alarm.WeChatSecret)
	return fmt.Sprintf("%x", h.Sum(nil))
}
