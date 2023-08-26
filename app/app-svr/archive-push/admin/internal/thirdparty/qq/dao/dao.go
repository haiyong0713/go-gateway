package dao

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"go-common/library/conf/paladin"

	qqModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/qq/model"
)

type Dao struct {
	httpClient     *http.Client
	Cfg            *qqModel.Config
	TGLAccessToken string
}

func Init() (dao *Dao, cf func(), err error) {
	dao = &Dao{
		httpClient: NewHTTPClient(),
	}
	if err = paladin.Get("qq.toml").UnmarshalTOML(&dao.Cfg); err != nil {
		return
	}
	dao.RefreshAccessToken(0)
	ticker := dao.startAccessTokenTicker()
	cf = func() {
		dao.httpClient.CloseIdleConnections()
		ticker.Stop()
	}
	return
}

// NewHTTPClient new a http client.
func NewHTTPClient() (client *http.Client) {
	var (
		transport *http.Transport
		dialer    *net.Dialer
	)
	dialer = &net.Dialer{
		Timeout:   time.Second * 10,
		KeepAlive: time.Second * 20,
	}
	transport = &http.Transport{
		DialContext:     dialer.DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client = &http.Client{
		Transport: transport,
	}
	return
}
