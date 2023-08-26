package dao

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
)

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

// Send send message to phone
func (d *Dao) Send(c context.Context, msg string) (err error) {
	var (
		users  = d.conf.Rule.WeChantUsers
		token  = d.conf.Rule.WeChatToken
		secret = d.conf.Rule.WeChatSecret
		urls   = d.conf.Rule.WeChanURI
		params = url.Values{}
	)
	params.Set("username", users)
	params.Set("content", msg)
	params.Set("token", token)
	params.Set("timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	mh := md5.Sum([]byte(params.Encode() + secret))
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
	req, _ := http.NewRequest("POST", urls, payload)
	req.Header.Add("content-type", "application/json; charset=utf-8")
	v := &resp{}
	if err = d.httpR.Do(context.Background(), req, v); err != nil {
		log.Error("s.client.Do error(%v)", err)
	}
	return
}
