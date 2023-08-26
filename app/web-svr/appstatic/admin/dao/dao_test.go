package dao

import (
	"flag"
	"strings"

	"go-gateway/app/web-svr/appstatic/admin/conf"

	"go-common/library/conf/paladin.v2"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

var d *Dao

var _ = func() bool {
	Init()
	return true
}()

func Init() {
	flag.Set("conf", "../cmd/appstatic-admin-test.toml")
	cfg, err := confInit()
	if err != nil {
		panic(err)
	}
	d = New(cfg)
}

func confInit() (*conf.Config, error) {
	err := paladin.Init()
	if err != nil {
		return nil, err
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err = paladin.Get("appstatic-admin-test.toml").UnmarshalTOML(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	d.client.SetTransport(gock.DefaultTransport)
	return r
}

func WithDao(f func(d *Dao)) func() {
	return func() {
		Reset(func() {})
		f(d)
	}
}
