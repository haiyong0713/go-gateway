package appstoreconnect

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"go-gateway/app/app-svr/fawkes/service/conf"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	goJwt "github.com/dgrijalva/jwt-go"
	"github.com/google/go-querystring/query"
)

// A Client manages communication with the AppStoreConnect API.
type Client struct {
	config   *conf.Config
	appsInfo sync.Map

	// HTTP client used to communicate with the API.
	client *http.Client

	Apps             *AppsService
	AppStoreVersions *AppStoreVersionsService
	BetaGroups       *BetaGroupsService
	Builds           *BuildsService
	Submissions      *SubmissionsService
}

// ErrorResponseError ...
type ErrorResponseError struct {
	Code   string      `json:"code,omitempty"`
	Status string      `json:"status,omitempty"`
	ID     string      `json:"id,omitempty"`
	Title  string      `json:"title,omitempty"`
	Detail string      `json:"detail,omitempty"`
	Source interface{} `json:"source,omitempty"`
}

// ErrorResponse ...
type ErrorResponse struct {
	Response *http.Response
	Errors   []ErrorResponseError `json:"errors,omitempty"`
}

// Error ...
func (r ErrorResponse) Error() string {
	m := fmt.Sprintf("%v %v: %d\n", r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode)
	var s string
	for _, err := range r.Errors {
		m += s + fmt.Sprintf("- %v %v", err.Title, err.Detail)
		s = "\n"
	}
	return m
}

func (c *Client) newJWT(appKey string) (jwtToken string, expTime int64, err error) {
	var (
		key *ecdsa.PrivateKey
	)
	info, err := loadAppInfo(appKey, &c.appsInfo)
	if err != nil {
		return
	}
	payload := goJwt.StandardClaims{
		Audience:  c.config.AppstoreConnect.Audience,
		Issuer:    info.IssuerID,
		ExpiresAt: time.Now().Unix() + c.config.AppstoreConnect.Expire,
	}
	token := goJwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": "ES256",
			"kid": info.KeyID,
		},
		Claims: payload,
		Method: goJwt.SigningMethodES256,
	}
	if key, err = parseP8PrivateKey(c.config.AppstoreConnect.KeyPath + "AuthKey_" + info.KeyID + ".p8"); err != nil {
		log.Error("newJWT: %v", err)
		return
	}
	expTime = payload.ExpiresAt
	if jwtToken, err = token.SignedString(key); err != nil {
		log.Error("newJWT: %v", err)
		return
	}
	return
}

func parseP8PrivateKey(path string) (pk *ecdsa.PrivateKey, err error) {
	var (
		rawByte []byte
		block   *pem.Block
	)
	if rawByte, err = ioutil.ReadFile(path); err != nil {
		log.Error("parseP8PrivateKey: %v", err)
		return nil, err
	}
	if block, _ = pem.Decode(rawByte); block == nil {
		log.Error("parseP8PrivateKey: not a pem")
		return nil, err
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Error("parseP8PrivateKey: %v", err)
		return nil, err
	}
	switch pk := key.(type) {
	case *ecdsa.PrivateKey:
		return pk, nil
	default:
		return nil, errors.New("token: AuthKey must be of type ecdsa.PrivateKey")
	}
}

// NewClient returns a new appstore connect API client.
func NewClient(conf *conf.Config) *Client {
	httpClient := http.DefaultClient
	client := &Client{
		client: httpClient,
		config: conf,
	}
	client.Apps = &AppsService{client: client}
	client.AppStoreVersions = &AppStoreVersionsService{client: client}
	client.BetaGroups = &BetaGroupsService{client: client}
	client.Builds = &BuildsService{client: client}
	client.Submissions = &SubmissionsService{client: client}
	return client
}

// RegisterApp register an app to appstoreconnect client
func (c *Client) RegisterApp(appKey, keyID, issuerID string) {
	appInfo := &AppInfo{
		IssuerID: issuerID,
		KeyID:    keyID,
		ExpireAt: 0,
	}
	c.appsInfo.Store(appKey, appInfo)
}

func (c *Client) checkJWT(appKey string) (err error) {
	now := time.Now().Unix()
	appInfo, err := loadAppInfo(appKey, &c.appsInfo)
	if err != nil {
		return
	}
	// 如果超过最后一次记录 jwt 的超时时间，则重新生成一个 jwt
	if appInfo.ExpireAt < now {
		var (
			token   string
			expTime int64
		)
		if token, expTime, err = c.newJWT(appKey); err != nil {
			log.Error("checkJWT: %v", err)
			return err
		}
		appInfo.Token = token
		appInfo.ExpireAt = expTime
		c.appsInfo.Store(appKey, appInfo)
	}
	return
}

// UploadIPA upload ipa to app store connect
func (c *Client) UploadIPA(appKey, ipaPath, ipaDescPath string) (out bytes.Buffer, err error) {
	if err = c.checkJWT(appKey); err != nil {
		log.Error("NewRequest: %v", err)
		return
	}
	appInfo, err := loadAppInfo(appKey, &c.appsInfo)
	if err != nil {
		return
	}
	if out, err = c.cmdUpload(appInfo.Token, ipaPath, ipaDescPath); err != nil {
		log.Error("cmdUpload: %v", err)
		return
	}
	return
}

func (c *Client) cmdUpload(JWT, ipaPath, ipaDescPath string) (out bytes.Buffer, err error) {
	var (
		errOut bytes.Buffer
	)
	outputPath := filepath.Dir(ipaPath) + "/upload_log.txt"
	// nolint:gosec
	cmd := exec.Command(c.config.AppstoreConnect.ITMSTransporter, "-m", "upload", "-jwt", "\""+JWT+"\"", "-v", "eXtreme", "-assetFile", ipaPath, "-assetDescription", ipaDescPath, "-o", outputPath)
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	log.Warn("iTMS Uploading: %v", ipaPath)
	if err = cmd.Run(); err != nil {
		log.Error("Command Run stdout=(%s) stderr=(%s) error(%v)", out.String(), errOut.String(), err)
		return
	}
	log.Error("iTMS Uploading Cmd Run Success! log: %v " + out.String())
	return
}

// NewRequest create a new request for appstore connect API client.
func (c *Client) NewRequest(appKey, method, path string, body interface{}) (req *http.Request, err error) {
	if err = c.checkJWT(appKey); err != nil {
		log.Error("NewRequest: %v", err)
		return
	}
	url := c.config.AppstoreConnect.BaseURL + path
	if req, err = http.NewRequest(method, url, nil); err != nil {
		log.Error("NewRequest: %v", err)
		return
	}
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete || method == http.MethodPatch {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			log.Error("NewRequest: %v", err)
			return nil, err
		}
		bodyReader := bytes.NewReader(bodyBytes)
		req.Body = ioutil.NopCloser(bodyReader)
		req.ContentLength = int64(bodyReader.Len())
		req.Header.Set("Content-Type", "application/json")
	}
	appInfo, err := loadAppInfo(appKey, &c.appsInfo)
	if err != nil {
		return
	}
	req.Header.Add("Authorization", "Bearer "+appInfo.Token)
	return
}

func checkResponse(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}
	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		_ = json.Unmarshal(data, errorResponse)
	}
	return errorResponse
}

// Do sends an API request and returns the API response.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return resp, err
	}
	if v != nil {
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}
	return resp, err
}

// addOptions adds the parameters in opt as URL query parameters to s. opt
// must be a struct whose fields may contain "url" tags.
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}
	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}
	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}
	u.RawQuery = qs.Encode()
	return u.String(), nil
}

func loadAppInfo(appKey string, appsInfo *sync.Map) (appInfo *AppInfo, err error) {
	load, ok := appsInfo.Load(appKey)
	if !ok {
		errMsg := fmt.Sprintf("appKey[%s] 不存在", appKey)
		log.Error(errMsg)
		err = errors.New(errMsg)
		return
	}
	appInfo = load.(*AppInfo)
	return
}
