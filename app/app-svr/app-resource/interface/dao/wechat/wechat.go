package wechat

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"go-common/library/cache/credis"
	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/conf"
)

const (
	//nolint:gosec
	_wechatTokenURL  = "/cgi-bin/token"
	_wechatTicketURL = "/cgi-bin/ticket/getticket"
)

// Dao is wechat dao
type Dao struct {
	c               *conf.Config
	client          *httpx.Client
	wechatTokenURL  string
	wechatTicketURL string
	// redis
	redis      credis.Redis
	keyExpired int32
}

// New initial wechat dao
func New(c *conf.Config) *Dao {
	d := &Dao{
		c:               c,
		client:          httpx.NewClient(c.HTTPWechat),
		wechatTokenURL:  c.Host.Wechat + _wechatTokenURL,
		wechatTicketURL: c.Host.Wechat + _wechatTicketURL,
		// redis
		redis:      credis.NewRedis(c.Redis.Resource.Config),
		keyExpired: int32(time.Duration(c.Redis.Resource.Expire) / time.Second),
	}
	return d
}

func (d *Dao) WechatAuth(c context.Context, nonce, timestamp, currentUrl string) (string, error) {
	param := url.Values{}
	param.Set("nonce", nonce)
	param.Set("timestamp", timestamp)
	param.Set("url", currentUrl)
	param.Set("grant_type", d.c.WechatAuth.ClientCredential)
	param.Set("appid", d.c.WechatAuth.Appid)
	param.Set("secret", d.c.WechatAuth.Secret)
	var token struct {
		Token string `json:"access_token"`
	}
	if err := d.httpGet(c, d.wechatTokenURL, param, &token); err != nil {
		log.Error("%+v", err)
		return "", err
	}
	param.Set("type", "jsapi")
	param.Set("access_token", token.Token)
	var ticket struct {
		Errcode int    `json:"errcode"`
		EArrmsg string `json:"errmsg"`
		Ticket  string `json:"ticket"`
	}
	if err := d.httpGet(c, d.wechatTicketURL, param, &ticket); err != nil {
		log.Error("%+v", err)
		return "", err
	}
	return ticket.Ticket, nil
}

func (d *Dao) httpGet(c context.Context, urlStr string, params url.Values, res interface{}) error {
	ru := urlStr
	if params != nil {
		ru = urlStr + "?" + params.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, ru, nil)
	if err != nil {
		log.Error("httpGet url(%s) error(%v)", urlStr+"?"+params.Encode(), err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-BACKEND-BILI-REAL-IP", "")
	return d.client.Do(c, req, &res)
}
