package dao

import (
	"context"
	pb "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoRawUserTab(t *testing.T) {
	Convey("RawUserTab", t, func() {
		var (
			c   = context.Background()
			req = &pb.UserTabReq{}
		)
		Convey("When everything goes positive", func() {
			res, err := d.RawUserTab(c, req)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSpaceUserTabAdd(t *testing.T) {
	Convey("SpaceUserTabAdd", t, func() {
		var (
			arg = &model.UserTab{}
		)
		Convey("When everything goes positive", func() {
			err := d.SpaceUserTabAdd(arg)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoSpaceUserTabModify(t *testing.T) {
	Convey("SpaceUserTabModify", t, func() {
		var (
			arg = &model.UserTab{}
		)
		Convey("When everything goes positive", func() {
			err := d.SpaceUserTabModify(arg)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoSpaceUserTabOnline(t *testing.T) {
	Convey("SpaceUserTabOnline", t, func() {
		var (
			id  = int64(0)
			arg = &model.UserTab{}
		)
		Convey("When everything goes positive", func() {
			err := d.SpaceUserTabOnline(id, arg)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoValidUserTabFindByMid(t *testing.T) {
	Convey("ValidUserTabFindByMid", t, func() {
		var (
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			ret, err := d.ValidUserTabFindByMid(mid)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoEtimeUserTabFindByMid(t *testing.T) {
	Convey("EtimeUserTabFindByMid", t, func() {
		var (
			arg = &model.UserTab{}
		)
		Convey("When everything goes positive", func() {
			ret, err := d.EtimeUserTabFindByMid(arg)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSpaceUserTabFindById(t *testing.T) {
	Convey("SpaceUserTabFindById", t, func() {
		var (
			id = int64(0)
		)
		Convey("When everything goes positive", func() {
			ret, err := d.SpaceUserTabFindById(id)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoMidInfoReply(t *testing.T) {
	Convey("MidInfoReply", t, func() {
		var (
			c   = context.Background()
			mid = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.MidInfoReply(c, mid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				So(res, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoCheckNaPage(t *testing.T) {
	Convey("测试Native ID", t, func() {
		Convey("一个存在的native id", func() {
			var (
				pid = int64(123)
			)
			err := d.CheckNaPage(pid)
			Convey("应该成功", func() {
				So(err, ShouldBeNil)
			})
		})
		Convey("一个不存在的native id", func() {
			var (
				pid = int64(51260)
			)
			err := d.CheckNaPage(pid)
			Convey("应该出错", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
