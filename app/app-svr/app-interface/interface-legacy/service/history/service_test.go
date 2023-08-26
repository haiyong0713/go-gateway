package history

import (
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"
)

var (
	s *Service
)

func TestMain(m *testing.M) {
	flag.Set("app_id", "main.app-svr.app-interface")
	flag.Set("conf_token", "1mWvdEwZHmCYGoXJCVIdszBOPVdtpXb3")
	flag.Set("tree_id", "2688")
	flag.Set("conf_version", "docker-1")
	flag.Set("deploy_env", "uat")
	flag.Set("conf_host", "config.bilibili.co")
	flag.Set("conf_path", "/tmp")
	flag.Set("region", "sh")
	flag.Set("zone", "sh001")
	flag.Set("conf", "/Users/zqq/go/src/go-gateway/app/app-svr/app-interface/interface-legacy/cmd/app-interface-test.toml")
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	s = New(conf.Conf)
	os.Exit(m.Run())
	// time.Sleep(time.Second)
}

func WithService(f func(s *Service)) func() {
	return func() {
		f(s)
	}
}
