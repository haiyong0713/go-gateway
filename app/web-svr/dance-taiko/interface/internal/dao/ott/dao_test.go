package ott

import (
	"context"
	"flag"
	"go-common/library/conf/paladin"
	"os"
	"testing"
)

var d *dao
var ctx = context.Background()

func TestMain(m *testing.M) {
	flag.Set("conf", "../../../configs")
	flag.Set("f", "../../test/docker-compose.yaml")
	flag.Parse()
	var err error
	if err = paladin.Init(); err != nil {
		panic(err)
	}
	var cf func()
	if d, cf, err = newTestDao(); err != nil {
		panic(err)
	}
	ret := m.Run()
	cf()
	os.Exit(ret)
}
