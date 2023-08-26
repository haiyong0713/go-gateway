package dao

import (
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"

	"gopkg.in/h2non/gock.v1"
)

var d *dao

func TestMain(m *testing.M) {
	flag.Set("conf", "../../public")
	flag.Parse()
	var err error
	if err = paladin.Init(); err != nil {
		panic(err)
	}
	var cf func()
	if d, cf, err = newTestDao(); err != nil {
		panic(err)
	}
	d.httpClient.SetTransport(gock.DefaultTransport)
	ret := m.Run()
	cf()
	os.Exit(ret)
}
