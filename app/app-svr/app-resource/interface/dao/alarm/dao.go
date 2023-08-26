package alarm

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	httpx "go-common/library/net/http/blademaster"
	"go-gateway/app/app-svr/app-resource/interface/conf"
)

const (
	_bapURL = "/api/v1/message/add"
)

// Dao macross dao.
type Dao struct {
	// conf
	c *conf.Config
	// http client
	client *httpx.Client
	// url
	bapURL string
}

type wxParams struct {
	Username  string `json:"username"`
	Content   string `json:"content"`
	Token     string `json:"token"`
	Timestamp int64  `json:"timestamp"`
	Sign      string `json:"signature"`
}

type resp struct {
	Status int64  `json:"status"`
	Msg    string `json:"msg"`
}

// New dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c:      c,
		client: httpx.NewClient(c.HTTPClient),
		// url
		bapURL: c.Host.Bap + _bapURL,
	}
	return
}

// SendWeChart send wechart
func (d *Dao) SendWeChart(c context.Context, contents string, users []string) (err error) {
	params := url.Values{}
	params.Set("username", strings.Join(users, ","))
	params.Set("content", contents)
	params.Set("token", d.c.WeChant.Token)
	params.Set("timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	mh := md5.Sum([]byte(params.Encode() + d.c.WeChant.Secret))
	params.Set("signature", hex.EncodeToString(mh[:]))
	p := &wxParams{
		Username: params.Get("username"),
		Content:  params.Get("content"),
		Token:    params.Get("token"),
		Sign:     params.Get("signature"),
	}
	p.Timestamp, _ = strconv.ParseInt(params.Get("timestamp"), 10, 64)
	bs, _ := json.Marshal(p)
	payload := strings.NewReader(string(bs))
	req, _ := http.NewRequest("POST", d.bapURL, payload)
	req.Header.Add("content-type", "application/json; charset=utf-8")
	v := &resp{}
	if err = d.client.Do(c, req, v); err != nil {
		log.Error("s.client.Do error(%v)", err)
	}
	return
}
