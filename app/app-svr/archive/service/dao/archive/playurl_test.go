package archive

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestArchivePGCPlayerInfos(t *testing.T) {
	var (
		c        = context.TODO()
		aids     = []int64{10111001}
		platform = "iphone"
		ip       = "121.31.246.238"
		session  = ""
		fnval    = int64(16)
		fnver    = int64(0)
	)
	convey.Convey("PGCPlayerInfos", t, func(ctx convey.C) {
		pm, err := d.PGCPlayerInfos(c, aids, platform, ip, session, fnval, fnver)
		ctx.Convey("Then err should be nil.pm should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(pm, convey.ShouldNotBeNil)
			p, _ := json.Marshal(pm)
			convey.Printf("%s", p)
		})
	})
}
