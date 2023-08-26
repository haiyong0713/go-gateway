package dataplat

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	xtime "go-common/library/time"

	pkgerr "github.com/pkg/errors"
)

/*
  访问数据平台的http client，处理了签名、接口监控等
*/

// ClientConfig client config
type ClientConfig struct {
	Key         string
	Secret      string
	ClusterName string
	Dial        xtime.Duration
	Timeout     xtime.Duration
	KeepAlive   xtime.Duration
}

// New new client
func New(c *ClientConfig) *HttpClient {
	return &HttpClient{
		client: &http.Client{},
		conf:   c,
	}
}

// HttpClient http client
type HttpClient struct {
	client *http.Client
	conf   *ClientConfig
	Debug  bool
}

// Response response
// 对应返回值为"result"的接口
type Response struct {
	Code   int         `json:"code"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}

// 对应返回值为"results"的接口
type Response2 struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Results interface{} `json:"results"`
}

type ResponseHive struct {
	Code         int      `json:"code"`
	Msg          string   `json:"msg"`
	JobStatusUrl *string  `json:"jobStatusUrl"`
	JobStatusId  int      `json:"statusId"`
	StatusMsg    string   `json:"statusMsg"`
	HdfsPath     []string `json:"hdfsPath"`
}

type PostRequest struct {
	AppKey     string `json:"appKey"`
	Timestamp  string `json:"timestamp"`
	Version    string `json:"version"`
	SignMethod string `json:"signMethod"`
	Sign       string `json:"sign"`
	Query      string `json:"query"`
}

const (
	keyAppKey = "appKey"

	// 	keyAppID      = "apiId"
	keyTimeStamp  = "timestamp"
	keySign       = "sign"
	keySignMethod = "signMethod"
	keyVersion    = "version"
	// TimeStampFormat time format in second
	TimeStampFormat = "2006-01-02 15:04:05"
)

// Get issues a GET to the specified URL.
func (client *HttpClient) Get(c context.Context, uri string, params url.Values, res interface{}) (err error) {
	req, err := client.NewRequest(http.MethodGet, uri, params)
	if err != nil {
		return
	}
	return client.Do(c, req, res)
}

func (client *HttpClient) Post(c context.Context, uri string, params url.Values, res interface{}) (err error) {
	req, err := client.NewRequest(http.MethodPost, uri, params)
	if err != nil {
		return
	}
	return client.Do(c, req, res)
}

// NewRequest new http request with method, uri, ip, values and headers.
// TODO(zhoujiahui): param realIP should be removed later.
func (client *HttpClient) NewRequest(method, uri string, params url.Values) (req *http.Request, err error) {
	if params == nil {
		params = url.Values{}
	}
	if client.conf != nil && client.conf.ClusterName != "" {
		params.Add("clusterName", client.conf.ClusterName)
	}

	signStr, err := client.sign(params)
	if err != nil {
		err = pkgerr.Wrapf(err, "uri:%s,params:%v", uri, params)
		return
	}
	params.Add(keySign, signStr)

	if method == http.MethodGet {
		enc := params.Encode()
		ru := uri
		if enc != "" {
			ru = uri + "?" + enc
		}
		req, err = http.NewRequest(http.MethodGet, ru, nil)
		if err != nil {
			err = pkgerr.Wrapf(err, "method:%s,uri:%s", method, ru)
			return
		}
	} else {
		postRequest := &PostRequest{
			AppKey:     params.Get("appKey"),
			Timestamp:  params.Get("timestamp"),
			Version:    params.Get("version"),
			SignMethod: params.Get("signMethod"),
			Sign:       params.Get("sign"),
			Query:      params.Get("query"),
		}
		//nolint:ineffassign,staticcheck
		postData := []byte{}
		if postData, err = json.Marshal(postRequest); err != nil {
			log.Error("NewRequest json.Marshal input(%v) error(%v)", postRequest, err)
			return
		}

		req, err = http.NewRequest(http.MethodPost, uri, strings.NewReader(string(postData)))
		if err != nil {
			err = pkgerr.Wrapf(err, "method:%s,uri:%s,body:%s", method, uri, postData)
			return
		}
	}
	const (
		_contentType = "Content-Type"
		_urlencoded  = "application/json"
	)
	if method == http.MethodPost {
		req.Header.Set(_contentType, _urlencoded)
	}

	return
}

// Do sends an HTTP request and returns an HTTP json response.
func (client *HttpClient) Do(c context.Context, req *http.Request, res interface{}, v ...string) (err error) {
	var bs []byte
	if bs, err = client.Raw(c, req, v...); err != nil {
		log.Error("DataPlat request error(%v)", err)
		return
	}
	// fmt.Printf("%s %+v %T", bs ,res,res)
	if res != nil {
		if err = json.Unmarshal(bs, res); err != nil {
			err = pkgerr.Wrapf(err, "host:%s, url:%s, response:%s", req.URL.Host, realURL(req), string(bs))
		}
	}
	return
}

// Raw get from url
func (client *HttpClient) Raw(c context.Context, req *http.Request, v ...string) (bs []byte, err error) {
	var resp *http.Response
	var uri = fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.Host, req.URL.Path)

	var now = time.Now()
	var code string
	defer func() {
		bm.MetricClientReqDur.Observe(int64(time.Since(now)/time.Millisecond), uri)
		if code != "" {
			bm.MetricClientReqCodeTotal.Inc(uri, code)
		}
	}()
	req = req.WithContext(c)
	if resp, err = client.client.Do(req); err != nil {
		err = pkgerr.Wrapf(err, "host:%s, url:%s", req.URL.Host, realURL(req))
		code = "failed"
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		err = pkgerr.Errorf("incorrect http status:%d host:%s, url:%s", resp.StatusCode, req.URL.Host, realURL(req))
		code = strconv.Itoa(resp.StatusCode)
		return
	}
	if bs, err = readAll(resp.Body, 16*1024); err != nil {
		err = pkgerr.Wrapf(err, "host:%s, url:%s", req.URL.Host, realURL(req))
		return
	}
	if client.Debug {
		log.Info("reqeust: host:%s, url:%s, response body:%s", req.URL.Host, realURL(req), string(bs))
	}
	return
}

// sign calc appkey and appsecret sign.
// see http://info.bilibili.co/pages/viewpage.action?pageId=5410881#id-%E6%95%B0%E6%8D%AE%E7%9B%98%EF%BC%8D%E5%AE%89%E5%85%A8%E8%AE%A4%E8%AF%81-%E4%BA%8C%E7%AD%BE%E5%90%8D%E7%AE%97%E6%B3%95
func (client *HttpClient) sign(params url.Values) (sign string, err error) {
	key := client.conf.Key
	secret := client.conf.Secret
	if params == nil {
		params = url.Values{}
	}

	params.Set(keyAppKey, key)
	params.Set(keyVersion, "1.0")
	if params.Get(keyTimeStamp) == "" {
		params.Set(keyTimeStamp, time.Now().Format(TimeStampFormat))
	}
	params.Set(keySignMethod, "md5")

	var needSignParams = url.Values{}
	needSignParams.Add(keyAppKey, key)
	needSignParams.Add(keyTimeStamp, params.Get(keyTimeStamp))
	needSignParams.Add(keyVersion, params.Get(keyVersion))

	// tmp := params.Encode()
	var valueMap = map[string][]string(needSignParams)
	var buf bytes.Buffer
	// 开头与结尾加secret
	buf.Write([]byte(secret))
	keys := make([]string, 0, len(valueMap))
	for k := range valueMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := valueMap[k]
		prefix := k
		buf.WriteString(prefix)
		for _, v := range vs {
			buf.WriteString(v)
			break
		}
	}

	buf.Write([]byte(secret))
	var encoder = md5.New()
	if _, err = encoder.Write(buf.Bytes()); err != nil {
		return "", err
	}
	sign = fmt.Sprintf("%X", encoder.Sum(nil))
	return
}

// readAll reads from r until an error or EOF and returns the data it read
// from the internal buffer allocated with a specified capacity.
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}

// realUrl return url with http://host/params.
func realURL(req *http.Request) string {
	if req.Method == http.MethodGet {
		return req.URL.String()
	} else if req.Method == http.MethodPost {
		ru := req.URL.Path
		if req.Body != nil {
			rd, ok := req.Body.(io.Reader)
			if ok {
				buf := bytes.NewBuffer([]byte{})
				//nolint:errcheck
				buf.ReadFrom(rd)
				ru = ru + "?" + buf.String()
			}
		}
		return ru
	}
	return req.URL.Path
}

// SetTransport set client transport
func (client *HttpClient) SetTransport(t http.RoundTripper) {
	client.client.Transport = t
}
