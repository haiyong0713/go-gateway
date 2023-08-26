package peak

//nolint:gosec
import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"go-common/library/database/orm"
	"go-common/library/log"
	"go-gateway/app/web-svr/appstatic/admin/conf"
)

const (
	_template = "%s\n%s\n\n%d\n"
	_method   = "PUT"
	_timeout  = 80000 * time.Millisecond
)

// Dao struct user of color egg Dao.
type Dao struct {
	DB     *gorm.DB
	c      *conf.Config
	client *http.Client
	bucket string
	bfsUrl string
	key    string
	secret string
}

// New create an instance of color egg Dao and return.
func New(c *conf.Config) (d *Dao) {
	d = &Dao{
		DB:     orm.NewMySQL(c.PeakDB),
		c:      c,
		client: http.DefaultClient,
		bucket: c.PeakBfs.Bucket,
		bfsUrl: c.PeakBfs.Addr,
		key:    c.PeakBfs.Key,
		secret: c.PeakBfs.Secret,
	}
	d.initORM()
	return
}

func (d *Dao) initORM() {
	d.DB.LogMode(true)
}

// Ping check connection of db , mc.
func (d *Dao) Ping(c context.Context) (err error) {
	if d.DB != nil {
		err = d.DB.DB().PingContext(c)
		return
	}
	return
}

// Close close connection of db , mc.
func (d *Dao) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}

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

func (d *Dao) Upload(c context.Context, fileType string, body io.Reader) (location string, err error) {
	req, err := http.NewRequest(_method, d.bfsUrl, body)
	if err != nil {
		log.Error("http.NewRequest error (%v) | fileType(%s) body(%v)", err, fileType, body)
		return
	}
	expire := time.Now().Unix()
	authorization := authorize(d.key, d.secret, _method, d.bucket, expire)
	log.Warn(authorization)
	req.Header.Set("Host", d.bfsUrl)
	req.Header.Add("Date", fmt.Sprint(expire))
	req.Header.Add("Authorization", authorization)
	req.Header.Add("Content-Type", fileType)
	c, cancel := context.WithTimeout(c, _timeout)
	req = req.WithContext(c)
	defer cancel()

	resp, err := d.client.Do(req)
	if err != nil {
		log.Error("d.Client.Do error(%v) | _url(%s) req(%v)", err, d.bfsUrl, req)
		err = fmt.Errorf("d.Client.Do error(%v) | _url(%s) req(%v)", err, d.bfsUrl, req)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error("Upload http.StatusCode nq http.StatusOK (%d) | url(%s)", resp.StatusCode, d.bfsUrl)
		err = fmt.Errorf("Upload http.StatusCode nq http.StatusOK (%d) | url(%s)", resp.StatusCode, d.bfsUrl)
		return
	}
	header := resp.Header
	code := header.Get("Code")
	if code != strconv.Itoa(http.StatusOK) {
		log.Error("strconv.Itoa err, code(%s) | url(%s)", code, d.bfsUrl)
		err = fmt.Errorf("strconv.Itoa err, code(%s) | url(%s)", code, d.bfsUrl)
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
