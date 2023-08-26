package dao

import (
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin.v2"
)

var d *dao

func TestMain(m *testing.M) {
	flag.Set("conf", "../../configs")
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
