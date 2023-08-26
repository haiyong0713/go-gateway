package dao

import (
	"testing"

	"go-gateway/app/web-svr/space/admin/model"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoSysNotice(t *testing.T) {
	convey.Convey("SysNoticeAdd", t, func(ctx convey.C) {
		var (
			param = &model.SysNoticeList{
				Uid:    1,
				Scopes: []int{1, 2},
				Status: 0,
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			list, err := d.SysNotice(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(list, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSysNoticeAdd(t *testing.T) {
	convey.Convey("SysNoticeAdd", t, func(ctx convey.C) {
		var (
			param = &model.SysNoticeAdd{
				Content: "test",
				Url:     "url",
				Scopes:  "1,2",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SysNoticeAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSysNoticeUpdate(t *testing.T) {
	convey.Convey("SysNoticeUpdate", t, func(ctx convey.C) {
		var (
			param = &model.SysNoticeUp{
				ID:      1,
				Content: "aaa",
				Url:     "bbb",
				Scopes:  "1",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SysNoticeUpdate(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSysNoticeOpt(t *testing.T) {
	convey.Convey("SysNoticeOpt", t, func(ctx convey.C) {
		var (
			param = &model.SysNoticeOpt{
				ID:     1,
				Status: 1,
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SysNoticeOpt(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSysNoticeUidAdd(t *testing.T) {
	convey.Convey("SysNoticeUidAdd", t, func(ctx convey.C) {
		var (
			param = &model.SysNotUidAddDel{
				ID:   2,
				UIDs: []int64{1, 2, 3},
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SysNoticeUidAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSysNoticeUidFirst(t *testing.T) {
	convey.Convey("SysNoticeUidFirst", t, func(ctx convey.C) {
		var (
			//uid      = int64(10000)
			uid      = int64(1)
			scopes   = []int{1}
			noticeId = int64(158)
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			err := d.SysNoticeUidFirst(uid, noticeId, scopes)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSysNoticeUidFind(t *testing.T) {
	convey.Convey("SysNoticeUidFind", t, func(ctx convey.C) {
		var (
			param = &model.SysNoticeUidParam{}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			value, err := d.SysNoticeUidFind(param)
			ctx.Convey("Then err should be nil.value should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(value, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestDaoSysNoticeUidDel(t *testing.T) {
	convey.Convey("SysNoticeUidDel", t, func(ctx convey.C) {
		var (
			param = &model.SysNotUidAddDel{
				ID:   2,
				UIDs: []int64{1, 2},
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SysNoticeUidDel(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDaoSysNoticeInfo(t *testing.T) {
	convey.Convey("SysNoticeInfo", t, func(ctx convey.C) {
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			ret, err := d.SysNoticeInfo(157)
			ctx.Convey("Then ret should not be nil.", func(ctx convey.C) {
				ctx.So(ret, convey.ShouldNotBeNil)
			})
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
