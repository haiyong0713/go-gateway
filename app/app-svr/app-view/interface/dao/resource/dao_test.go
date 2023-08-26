package resource

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"testing"

	"go-gateway/app/app-svr/app-view/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	d *Dao
)

func init() {
	dir, _ := filepath.Abs("../../cmd/app-view-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	d = New(conf.Conf)
}

func Test_Paster(t *testing.T) {
	Convey("should get banner", t, func() {
		_, err := d.Paster(context.Background(), 1, 2, "", "", "")
		So(err, ShouldBeNil)
	})
}

func Test_PlayerIcon(t *testing.T) {
	Convey("should get player icon", t, func() {
		_, err := d.PlayerIcon(context.Background(), 0, 0, []int64{}, 0, true)
		So(err, ShouldBeNil)
	})
}

func TestDao_GetSpecialCard(t *testing.T) {
	Convey("should get special card", t, func() {
		resp, err := d.GetSpecialCard(context.Background())
		fmt.Println(resp)
		So(err, ShouldBeNil)
	})
}
