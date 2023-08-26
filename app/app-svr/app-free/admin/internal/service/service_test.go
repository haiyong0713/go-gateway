package service

import (
	"flag"
	"os"
	"testing"

	"go-common/library/conf/paladin"
)

var (
	s *Service
)

func TestMain(m *testing.M) {
	flag.Set("conf", "../../configs")
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	s = New()
	os.Exit(m.Run())
}
