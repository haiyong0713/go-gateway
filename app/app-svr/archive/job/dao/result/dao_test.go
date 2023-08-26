package result

import (
	"flag"
	"os"
	"testing"

	"go-gateway/app/app-svr/archive/job/conf"

	"go-common/library/conf/paladin.v2"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.archive-job")
		flag.Set("conf_token", "MmQwIqWAyIaIu8CKb7MKcNSYlGGhoudN")
		flag.Set("tree_id", "2301")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	}
	if os.Getenv("UT_LOCAL_TEST") != "" {
		flag.Set("conf", "../../cmd/archive-job-test.toml")
	}
	flag.Parse()

	if err := paladin.Init(); err != nil {
		panic(err)
	}
	defer paladin.Close()
	cfg := &conf.Config{}
	if err := paladin.Get("archive-job.toml").UnmarshalTOML(&cfg); err != nil {
		panic(err)
	}
	if err := paladin.Watch("archive-job.toml", cfg); err != nil {
		panic(err)
	}

	d = New(cfg)
	m.Run()
	os.Exit(0)
}
