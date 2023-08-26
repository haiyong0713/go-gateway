package popups

import (
	"context"
	pb "go-gateway/app/app-svr/resource/service/api/v1"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPopupsGetMysqlPopUps(t *testing.T) {
	Convey("GetMysqlPopUps", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			ret, err := d.GetMysqlPopUps(c)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
			})
		})
	})
}

func TestPopupsCheckCrowd(t *testing.T) {
	Convey("checkCrowd", t, func() {
		var (
			req = &pb.PopUpsReq{
				Mid:  12834689,
				Plat: 1000,
			}
			crowd_base  = 1
			crowd_value = 50
		)
		Convey("When everything goes positive", func() {
			valid, err := d.CheckCrowd(req, crowd_base, int64(crowd_value))
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
			})
			Convey("Valid should be true", func() {
				So(valid, ShouldBeTrue)
			})
		})
	})
}

func TestPopupsGetEffectivePopUps(t *testing.T) {
	Convey("GetEffectivePopUps", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			ret, err := d.GetEffectivePopUps(c)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
			})
		})
	})
}

func TestPopupsUpdatePopUpsCache(t *testing.T) {
	Convey("UpdatePopUpsCache", t, func() {
		var (
			c = context.Background()
		)
		Convey("When everything goes positive", func() {
			err := d.UpdatePopUpsCache(c)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestPopupsFlushPopUpsCache(t *testing.T) {
	Convey("FlushPopUpsCache", t, func() {
		Convey("When everything goes positive", func() {
			d.FlushPopUpsCache()
			Convey("No return values", func() {
			})
		})
	})
}
