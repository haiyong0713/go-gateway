package dao

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strconv"
	"time"

	"go-common/library/conf/paladin"
	"go-common/library/ecode"

	"github.com/pkg/errors"
)

const (
	_timeout  = 800 * time.Millisecond
	_bucket   = "app-static"
	_bfsURL   = "/bfs/app-static/"
	_template = "%s\n%s\n\n%d\n"
	_method   = "PUT"
)

type BfsC struct {
	Key     string
	Secret  string
	Host    string
	Timeout int
}

type BfsConf struct {
	Bfs BfsC
}

type BfsClient struct {
	BfsC
	Client *http.Client
}

func authorize(key, secret, method, bucket string, expire int64) (authorization string) {
	var (
		content   string
		mac       hash.Hash
		signature string
	)
	content = fmt.Sprintf(_template, method, bucket, expire)
	mac = hmac.New(sha1.New, []byte(secret))
	mac.Write([]byte(content))
	signature = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	authorization = fmt.Sprintf("%s:%s:%d", key, signature, expire)
	return
}

func (b *BfsClient) New() error {
	conf := BfsConf{}
	if err := paladin.Get("bfs.toml").UnmarshalTOML(&conf); err != nil {
		panic(err)
	}
	b.Key = conf.Bfs.Key
	b.Secret = conf.Bfs.Secret
	b.Host = conf.Bfs.Host
	b.Timeout = conf.Bfs.Timeout
	b.Client = http.DefaultClient
	return nil
}

// Upload imgage upload .
func (b *BfsClient) Upload(c context.Context, fileType string, body io.Reader) (location string, err error) {
	req, err := http.NewRequest(_method, b.Host+_bfsURL, body)
	if err != nil {
		return
	}
	expire := time.Now().Unix()
	authorization := authorize(b.Key, b.Secret, _method, _bucket, expire)
	req.Header.Set("Host", b.Host+_bfsURL)
	req.Header.Add("Date", fmt.Sprint(expire))
	req.Header.Add("Authorization", authorization)
	req.Header.Add("Content-Type", fileType)
	// timeout
	c, cancel := context.WithTimeout(c, _timeout)
	req = req.WithContext(c)
	defer cancel()
	resp, err := b.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = errors.Wrap(ecode.Int(resp.StatusCode), req.URL.String())
		return
	}
	code := resp.Header.Get("Code")
	if code != strconv.Itoa(http.StatusOK) {
		err = errors.Wrap(ecode.String(code), req.URL.String())
		return
	}
	location = resp.Header.Get("Location")
	return
}

func (d *dao) BfsUpload(c context.Context, fileType string, body io.Reader) (location string, err error) {
	return d.bfs.Upload(c, fileType, body)
}
