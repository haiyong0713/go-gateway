package dao

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/space/admin/model"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaoSpaceUserTabAdd(t *testing.T) {
	Convey("SpaceUserTabAdd", t, func() {
		var (
			arg = &model.UserTabReq{}
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
			arg = &model.UserTabReq{}
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
			arg = &model.UserTabReq{}
		)
		Convey("When everything goes positive", func() {
			err := d.SpaceUserTabOnline(id, arg)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoSpaceUserTabDelete(t *testing.T) {
	Convey("SpaceUserTabDelete", t, func() {
		var (
			id = int64(0)
			t  = int(0)
		)
		Convey("When everything goes positive", func() {
			err := d.SpaceUserTabDelete(id, t)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestDaoSpaceUserTabList(t *testing.T) {
	Convey("SpaceUserTabList", t, func() {
		var (
			arg = &model.UserTabListReq{
				Mid: 27515233,
				Ps:  20,
				Pn:  1,
			}
		)
		Convey("When everything goes positive", func() {
			list, count, err := d.SpaceUserTabList(arg)
			Convey("Then err should be nil.list,count should not be nil.", func() {
				So(err, ShouldBeNil)
				So(count, ShouldNotBeNil)
				So(list, ShouldNotBeNil)
				fmt.Printf("list = %+v", list)
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
			arg = &model.UserTabReq{}
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

func TestDaoFindUserTabByTime(t *testing.T) {
	Convey("FindUserTabByTime", t, func() {
		var (
			state = int(0)
		)
		Convey("When everything goes positive", func() {
			ret, err := d.FindUserTabByTime(state)
			Convey("Then err should be nil.ret should not be nil.", func() {
				So(err, ShouldBeNil)
				So(ret, ShouldNotBeNil)
			})
		})
	})
}

func TestDaoOfflineUserTab(t *testing.T) {
	Convey("NoticeUserTab", t, func() {
		var (
			mid = int64(1111112136)
			//mid = int64(2062843265)
			pageId  = int64(5879)
			tabType = "napage"
		)
		Convey("When everything goes positive", func() {
			err := d.NoticeUserTab(mid, pageId, tabType)
			So(err, ShouldBeNil)
		})
	})
}
