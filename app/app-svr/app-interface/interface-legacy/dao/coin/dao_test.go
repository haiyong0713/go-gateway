package coin

import (
	"context"
	"flag"
	"path/filepath"
	"testing"

	"go-gateway/app/app-svr/app-interface/interface-legacy/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func init() {
	dir, _ := filepath.Abs("../../cmd/app-interface-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
}

func Test_CoinList(t *testing.T) {
	Convey("should get Archives", t, func() {
		_, _, err := d.CoinList(context.Background(), 1, 3, 4)
		So(err, ShouldBeNil)
	})
}
