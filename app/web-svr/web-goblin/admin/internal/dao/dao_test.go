package dao

import (
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin"
	"go-common/library/testing/lich"
)

var (
	d   Dao
	err error
)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../configs")
	flag.Set("f", "../../test/docker-compose.yaml")
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	if err := lich.Setup(); err != nil {
		panic(err)
	}
	defer lich.Teardown()
	d, err = NewDao()
	if code := m.Run(); code != 0 {
		lich.Teardown()
		os.Exit(code)
	}
}
