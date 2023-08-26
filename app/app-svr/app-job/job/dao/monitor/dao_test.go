package monitor

import (
	"flag"
	"path/filepath"

	"go-gateway/app/app-svr/app-job/job/conf"
)

var (
	d *Dao
)

func init() {
	dir, _ := filepath.Abs("../../cmd/app-job-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
}
