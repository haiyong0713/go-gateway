package article

import (
	"flag"
	"os"
	"testing"

	"go-common/library/net/trace"
	"go-gateway/app/app-svr/hkt-note/service/conf"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	Init()
	os.Exit(m.Run())
}

func Init() {
	flag.Set("app_id", "main.app-svr.hkt-note-service")
	flag.Set("conf_token", "07c1826c1f39df02a1411cdd6f455879")
	flag.Set("tree_id", "15326")
	flag.Set("conf_version", "docker-1")
	flag.Set("deploy_env", "uat")
	flag.Set("conf_host", "config.bilibili.co")
	flag.Set("conf_path", "/tmp")
	flag.Set("region", "sh")
	flag.Set("zone", "sh001")
	flag.Set("conf", "../../cmd/hkt-note-service.toml")
	trace.Init(conf.Conf.Tracer)
	d = New(conf.Conf)
}
