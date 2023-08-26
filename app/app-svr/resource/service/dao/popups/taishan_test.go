package popups

import (
	"context"
	"go-gateway/app/app-svr/resource/service/model"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPopupsnewGetReq(t *testing.T) {
	Convey("newGetReq", t, func() {
		var (
			Mid   = 123515
			Buvid = "CFDSF-FDFDF123-XZCVG"
			ID    = 10
			key   = model.PopUpsKey(int64(ID), int64(Mid), Buvid)
		)
		Convey("When everything goes positive", func() {
			p1 := d.newGetReq(key)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestPopupsPutReq(t *testing.T) {
	Convey("PutReq", t, func() {
		var (
			value = []byte("true")
			ttl   = uint32(0)
			Mid   = 6112731
			Buvid = "CFDSF-FDFDF123-XZCVG"
			ID    = 16
			key   = model.PopUpsKey(int64(ID), int64(Mid), Buvid)
		)
		Convey("When everything goes positive", func() {
			err := d.PutReq([]byte(key), value, ttl)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestPopupsgetFromTaishan(t *testing.T) {
	Convey("getFromTaishan", t, func() {
		var (
			c     = context.Background()
			Mid   = 123515
			Buvid = "CFDSF-FDFDF123-XZCVG"
			ID    = 10
			key   = model.PopUpsKey(int64(ID), int64(Mid), Buvid)
		)
		Convey("When everything goes positive", func() {
			p1, err := d.getFromTaishan(c, key)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestPopupsGetIsPopFromTaishan(t *testing.T) {
	Convey("GetIsPopFromTaishan", t, func() {
		var (
			c     = context.Background()
			Mid   = 6112731
			Buvid = "CFDSF-FDFDF123-XZCVG"
			ID    = 17
			key   = model.PopUpsKey(int64(ID), int64(Mid), Buvid)
		)
		Convey("When everything goes positive", func() {
			is_pop, err := d.GetIsPopFromTaishan(c, key)
			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("is_pop should be false", func() {
				So(is_pop, ShouldBeFalse)
			})
		})
	})
}

//func TestPopupsdelteTaishanKey(t *testing.T) {
//	Convey("delteTaishanKey", t, func() {
//		var (
//			c   = context.Background()
//			Mid = 27515397
//			Buvid = "CFDSF-FDFDF123-XZCVG"
//			ID = 16
//			key = model.PopUpsKey(int64(ID), int64(Mid), Buvid)
//		)
//		Convey("When everything goes positive", func() {
//			 err := d.delteTaishanKey(c, key)
//			Convey("Then err should be nil", func() {
//				So(err, ShouldBeNil)
//			})
//		})
//	})
//}
