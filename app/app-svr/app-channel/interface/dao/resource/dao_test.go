package resource

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-channel/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func ctx() context.Context {
	return context.Background()
}

func init() {
	dir, _ := filepath.Abs("../../cmd/app-channel-test.toml")
	flag.Set("conf", dir)
	d = New(conf.Conf)
	time.Sleep(time.Second)
}

func TestEntrancesIsHidden(t *testing.T) {
	Convey("EntrancesIsHidden", t, func() {
		_, err := d.EntrancesIsHidden(ctx(), []int64{1, 2}, 9999, 0, "xiaomi")
		So(err, ShouldBeNil)
	})
}
