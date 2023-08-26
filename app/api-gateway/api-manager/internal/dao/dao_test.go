package dao

import (
	"context"
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin"
	"go-common/library/testing/dockertest"
)

var d *dao
var ctx = context.Background()

func TestMain(m *testing.M) {
	flag.Set("conf", "../../configs")
	flag.Parse()
	dockertest.Run("../../test/docker-compose.yaml")
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
