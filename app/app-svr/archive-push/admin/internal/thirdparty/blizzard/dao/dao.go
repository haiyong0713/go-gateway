package dao

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"go-common/library/conf/paladin"

	bzModel "go-gateway/app/app-svr/archive-push/admin/internal/thirdparty/blizzard/model"
)

type Dao struct {
	httpClient *http.Client
	Cfg        *bzModel.Config
}

func Init() (dao *Dao, cf func(), err error) {
	dao = &Dao{
		httpClient: NewHTTPClient(),
	}
	if err = paladin.Get("blizzard.toml").UnmarshalTOML(&dao.Cfg); err != nil {
		return
	}
	cf = func() {
		dao.httpClient.CloseIdleConnections()
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
