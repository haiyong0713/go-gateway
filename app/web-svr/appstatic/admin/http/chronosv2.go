package http

//nolint:gosec
import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"strconv"
	"time"

	"go-common/library/ecode"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/pkg/errors"
)

var privateKey = []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDBHQnPuW5erII+g0D0gDW4ipKEcGi5r8uiKijn/Ju6lizi3JEM
oXFmAzXwmXj5AQtOT3K0LuqRFufqWRRneGjf+eXSCrE3Xseuf1JuZ7oktfJDam0G
ItEjKm17Eb3z930AahYr8ScLgL77Keu4WXRZ58NBI9l+wJJvCwXC4VHAwwIDAQAB
AoGAYMp9MHBwsWMlpM+ErwfT5TsPVPJCi09hcVZQSnaCV3MN7GdBDGOewtK5Jm7G
A4hScl2/0C/zweUJOJyNbY8cgMaOOmHOVfkEnx8Ux+3xdstvT0XUi5BmNM09KzST
kGXbVt5q+bvAnbB9NTP/h9b4D2d5rUDZ41rFuEwXPF1uiYECQQD0YE476fSztljX
vGNYQkEs+5ELd/MhPluixGl51HHBnnTgx2IdRmpB2DGv52PaOk4Ocwbv5PVrhjLQ
C1Yinc+3AkEAykyFjZ6Mji4zVaF15z4+FpH75riWBcyY/dtBVpTNNCgkFnLi9rYo
5+dduvJIGfmHbQHByihghTv4QdcR7iB/VQJBALl25ake8/H4MBD7DsKK9f/3pKr5
i/Hs64rqWcp2aycw5S864sGpETeLppoDmIqkuVzJ+7fRIllKbgHquKJo9p0CQGvY
o5I+JfxeUOujqfFfU0ZBCSOU4BWzXxRmYMzBgyv9AlAdazXPIruOsn9JTnradgH8
38zf/aTJta2T9HEYTgkCQFOO413/bWSt1Kv86A1l9p5b0hnd2Ho9HiV8+Cu1/3B2
72/z6PCB4cR557apaU6PZdabwHtPi1VMIA6rYu0m03c=
-----END RSA PRIVATE KEY-----
`)

func saveChronosV2App(c *bm.Context) {
	params := &model.AppInfo{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.SaveAppInfo(c, params))
}

func deleteChronosV2App(c *bm.Context) {
	params := &struct {
		AppKey string `form:"app_key"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.DeleteAppInfo(c, params.AppKey))
}

func showChronosV2AppList(c *bm.Context) {
	c.JSON(apsSvc.ShowAppInfoList(c))
}

func showChronosV2AppDetail(c *bm.Context) {
	params := &struct {
		AppKey string `form:"app_key"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(apsSvc.ShowAppInfoDetail(c, params.AppKey))
}

func saveChronosV2Service(c *bm.Context) {
	params := &model.ServiceInfo{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.SaveServiceInfo(c, params))
}

func deleteChronosV2Service(c *bm.Context) {
	params := &struct {
		ServiceKey string `form:"service_key"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.DeleteServiceInfo(c, params.ServiceKey))
}

func showChronosV2ServiceList(c *bm.Context) {
	c.JSON(apsSvc.ShowServiceInfoList(c))
}

func showChronosV2ServiceDetail(c *bm.Context) {
	params := &struct {
		ServiceKey string `form:"service_key"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(apsSvc.ShowServiceInfoDetail(c, params.ServiceKey))
}

func saveChronosV2Package(c *bm.Context) {
	params := &model.PackageInfo{}
	if err := c.Bind(params); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		c.Abort()
		return
	}
	reply, err := apsSvc.SavePackageToAudit(c, params, userName)
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
	bs, err := json.Marshal(reply)
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
	c.JSON(reply, nil)
}

func deleteChronosV2Package(c *bm.Context) {
	params := &struct {
		ID int64 `form:"id"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		c.Abort()
		return
	}
	reply, err := apsSvc.DeletePackageToAudit(c, params.ID, userName)
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
	bs, err := json.Marshal(reply)
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
	c.JSON(reply, nil)
}

func rankChronosV2Package(c *bm.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	_ = c.Request.Body.Close()
	params := &struct {
		PackageIDRank map[int64]int64 `json:"package_id_rank"`
	}{}
	if err = json.Unmarshal(body, params); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	username, _ := c.Get("username")
	userName, ok := username.(string)
	if !ok || userName == "" {
		c.JSON(nil, ecode.NoLogin)
		c.Abort()
		return
	}
	reply, err := apsSvc.RankPackageToAudit(c, params.PackageIDRank, userName)
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
	bs, err := json.Marshal(reply)
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bs))
	c.JSON(reply, nil)
}

func showChronosV2Package(c *bm.Context) {
	params := &struct {
		AppKey     string `form:"app_key"`
		ServiceKey string `form:"service_key"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(apsSvc.ShowPackageInfoList(c, params.AppKey, params.ServiceKey))
}

func showChronosV2PackageDetail(c *bm.Context) {
	params := &struct {
		UUID string `form:"uuid"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(apsSvc.ShowPackageInfoDetail(c, params.UUID))
}

func approved(c *bm.Context) {
	params := &model.PackageOpReply{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.AuditApproved(c, params.AuditID))
}

func reject(c *bm.Context) {
	params := &model.PackageOpReply{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(nil, apsSvc.AuditReject(c, params.AuditID))
}

func auditList(c *bm.Context) {
	params := &struct {
		AppKey     string `form:"app_key"`
		ServiceKey string `form:"service_key"`
	}{}
	if err := c.Bind(params); err != nil {
		return
	}
	c.JSON(apsSvc.AuditList(c, params.AppKey, params.ServiceKey))
}

func batchSaveChronosV2Packages(c *bm.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	_ = c.Request.Body.Close()
	params := &struct {
		PackageList []*model.PackageInfo `json:"package_list"`
		AppKey      string               `json:"app_key"`
		ServiceKey  string               `json:"service_key"`
	}{}
	if err = json.Unmarshal(body, params); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, apsSvc.BatchSavePackage(c, params.PackageList, params.AppKey, params.ServiceKey))
}

func chronosRsaVerify(c *bm.Context) {
	err := func() error {
		encryptMsg := c.Request.Header.Get("chronos-token")
		decodeMsg, err := base64.StdEncoding.DecodeString(encryptMsg)
		if err != nil {
			return err
		}
		startTime, err := rsaDecrypt(decodeMsg, privateKey)
		if err != nil {
			return err
		}
		startTimeStamp, err := strconv.ParseInt(string(startTime), 10, 64)
		if err != nil {
			return err
		}
		minute := int64(60)
		if time.Now().Unix()-startTimeStamp > minute {
			return ecode.RequestErr
		}
		return nil
	}()
	if err != nil {
		c.JSON(nil, err)
		c.Abort()
		return
	}
}

func rsaDecrypt(encryptMsg []byte, privateKey []byte) ([]byte, error) {
	blockPub, _ := pem.Decode(privateKey)
	if blockPub == nil {
		return nil, errors.New("blockPub空")
	}
	priv, err := x509.ParsePKCS1PrivateKey(blockPub.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptOAEP(sha1.New(), rand.Reader, priv, encryptMsg, nil)
}
