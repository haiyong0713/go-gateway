package bfs

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	//nolint:gosec
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-feed/admin/conf"
)

const (
	_template = "%s\n%s\n\n%d\n"
	_method   = "PUT"
	_urlGif   = "/imageserver/image/check?%s"
)

// Dao is bfs dao.
type Dao struct {
	c            *conf.Config
	client       *http.Client
	bucket       string
	url          string
	key          string
	secret       string
	thumbnailUrl string
}

// New bfs dao.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		c: c,
		// http client
		client: &http.Client{
			Timeout: time.Duration(c.Bfs.Timeout),
		},
		bucket:       c.Bfs.Bucket,
		url:          c.Bfs.Addr,
		key:          c.Bfs.Key,
		secret:       c.Bfs.Secret,
		thumbnailUrl: c.Host.Thumbnail,
	}
	return
}

// Upload upload bfs.
func (d *Dao) Upload(c context.Context, fileType string, body io.Reader) (location string, err error) {
	req, err := http.NewRequest(_method, d.url, body)
	if err != nil {
		log.Error("http.NewRequest error (%v) | fileType(%s) body(%v)", err, fileType, body)
		return
	}
	expire := time.Now().Unix()
	authorization := authorize(d.key, d.secret, _method, d.bucket, expire)
	req.Header.Set("Host", d.url)
	req.Header.Add("Date", fmt.Sprint(expire))
	req.Header.Add("Authorization", authorization)
	req.Header.Add("Content-Type", fileType)
	// timeout
	c, cancel := context.WithTimeout(c, time.Duration(d.c.Bfs.Timeout))
	req = req.WithContext(c)
	defer cancel()
	resp, err := d.client.Do(req)

	//nolint:staticcheck,govet
	defer resp.Body.Close()

	if err != nil {
		log.Error("d.Client.Do error(%v) | _url(%s) req(%v)", err, d.url, req)
		err = fmt.Errorf("d.Client.Do error(%v) | _url(%s) req(%v)", err, d.url, req)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Error("Upload http.StatusCode nq http.StatusOK (%d) | url(%s)", resp.StatusCode, d.url)
		err = fmt.Errorf("Upload http.StatusCode nq http.StatusOK (%d) | url(%s)", resp.StatusCode, d.url)
		return
	}
	header := resp.Header
	code := header.Get("Code")
	if code != strconv.Itoa(http.StatusOK) {
		log.Error("strconv.Itoa err, code(%s) | url(%s)", code, d.url)
		err = fmt.Errorf("strconv.Itoa err, code(%s) | url(%s)", code, d.url)
		return
	}
	location = header.Get("Location")
	return
}

// authorize returns authorization for upload file to bfs
func authorize(key, secret, method, bucket string, expire int64) (authorization string) {
	var (
		content   string
		mac       hash.Hash
		signature string
		err       error
	)
	content = fmt.Sprintf(_template, method, bucket, expire)
	mac = hmac.New(sha1.New, []byte(secret))
	if _, err = mac.Write([]byte(content)); err != nil {
		return ""
	}
	signature = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	authorization = fmt.Sprintf("%s:%s:%d", key, signature, expire)
	return
}

// FileMd5 calculates the local file's md5 and store it in a file
func (d *Dao) FileMd5(content []byte) (md5Str string, err error) {
	md5hash := md5.New()
	if _, err = io.Copy(md5hash, bytes.NewReader(content)); err != nil {
		log.Error("FileMd5 is error (%v)", err)
		return
	}
	md5 := md5hash.Sum(nil)
	md5Str = hex.EncodeToString(md5[:])
	return
}

// ValidGif .
func (d *Dao) ValidGif(c context.Context, frame string, body []byte) (err error) {
	params := url.Values{}
	params.Set("frame", frame)
	urlStr := d.thumbnailUrl + fmt.Sprintf(_urlGif, params.Encode())
	bufferWriter := &bytes.Buffer{}
	formFileWriter := multipart.NewWriter(bufferWriter)
	formFile, err := formFileWriter.CreateFormFile("file", "file.gif")
	if err != nil {
		log.Error("ValidGif CreateFormFile error(%+v)", err)
		return
	}
	//nolint:errcheck
	formFile.Write(body)
	formFileWriter.Close()
	request, err := http.NewRequest("POST", urlStr, bufferWriter)
	if err != nil {
		log.Error("ValidGif NewRequest error(%+v)", err)
		return
	}
	request.Header.Add("Content-Type", formFileWriter.FormDataContentType())
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Error("ValidGif Do error(%+v)", err)
		return
	}
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("ValidGif ReadAll response(%+v) error(%+v)", response, err)
		return err
	}
	log.Info("ValidGif response(%+s)", string(content))
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("帧数不超过%s", frame)
	}
	return nil
}
