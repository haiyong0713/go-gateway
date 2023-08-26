package dao

import (
	"context"
	"testing"

	"go-common/library/database/taishan"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaowrapError(t *testing.T) {
	Convey("wrapError", t, func() {
		var (
			reply statusGetter
		)
		Convey("When everything goes positive", func() {
			err := wrapError(reply)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoNewKV(t *testing.T) {
	Convey("NewKV", t, func() {
		Convey("When everything goes positive", func() {
			p1, p2, err := NewKV()
			Convey("Then err should be nil.p1,p2 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p2, ShouldNotBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoNewTaishanError(t *testing.T) {
	Convey("NewTaishanError", t, func() {
		var (
			status = &taishan.Status{}
		)
		Convey("When everything goes positive", func() {
			err := NewTaishanError(status)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoNewGetReq(t *testing.T) {
	Convey("NewGetReq", t, func() {
		var (
			key = []byte("")
		)
		Convey("When everything goes positive", func() {
			p1 := d.taishan.NewGetReq(key)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoGet(t *testing.T) {
	Convey("Get", t, func() {
		var (
			ctx = context.Background()
			req = &taishan.GetReq{}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.taishan.Get(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoNewPutReq(t *testing.T) {
	Convey("NewPutReq", t, func() {
		var (
			key   = []byte("")
			value = []byte("")
			ttl   = uint32(0)
		)
		Convey("When everything goes positive", func() {
			p1 := d.taishan.NewPutReq(key, value, ttl)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoPut(t *testing.T) {
	Convey("Put", t, func() {
		var (
			ctx = context.Background()
			req = &taishan.PutReq{}
		)
		Convey("When everything goes positive", func() {
			err := d.taishan.Put(ctx, req)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoNewScanReq(t *testing.T) {
	Convey("NewScanReq", t, func() {
		var (
			start = []byte("")
			end   = []byte("")
			limit = uint32(0)
		)
		Convey("When everything goes positive", func() {
			p1 := d.taishan.NewScanReq(start, end, limit)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoScan(t *testing.T) {
	Convey("Scan", t, func() {
		var (
			ctx = context.Background()
			req = &taishan.ScanReq{}
		)
		Convey("When everything goes positive", func() {
			p1, err := d.taishan.Scan(ctx, req)
			Convey("Then err should be nil.p1 should not be nil.", func() {
				So(err, ShouldBeNil)
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoNewCASReq(t *testing.T) {
	Convey("NewCASReq", t, func() {
		var (
			key  = []byte("")
			oldV = []byte("")
			newV = []byte("")
		)
		Convey("When everything goes positive", func() {
			p1 := d.taishan.NewCASReq(key, oldV, newV)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoCAS(t *testing.T) {
	Convey("CAS", t, func() {
		var (
			ctx = context.Background()
			req = &taishan.CasReq{}
		)
		Convey("When everything goes positive", func() {
			err := d.taishan.CAS(ctx, req)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoNewDelReq(t *testing.T) {
	Convey("NewDelReq", t, func() {
		var (
			key = []byte("")
		)
		Convey("When everything goes positive", func() {
			p1 := d.taishan.NewDelReq(key)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoDel(t *testing.T) {
	Convey("Del", t, func() {
		var (
			ctx = context.Background()
			req = &taishan.DelReq{}
		)
		Convey("When everything goes positive", func() {
			err := d.taishan.Del(ctx, req)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
