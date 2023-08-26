package result

import (
	"context"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/archive/job/model/archive"

	"github.com/smartystreets/goconvey/convey"
)

func TestTxDelStaff(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(4052032)
	)
	convey.Convey("TxDelStaff", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			tx, err := d.BeginTran(c)
			ctx.So(err, convey.ShouldBeNil)
			err = d.TxDelStaff(tx, aid)
			ctx.So(err, convey.ShouldBeNil)
			err = tx.Commit()
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestTxAddStaff(t *testing.T) {
	var (
		c     = context.TODO()
		aid   = int64(4052032)
		staff []*archive.Staff
	)
	convey.Convey("TxAddStaff", t, func(ctx convey.C) {
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			staff = append(staff, &archive.Staff{Aid: aid, Mid: 3333, Title: "哈哈", Ctime: "2018-11-28T16:50:14+08:00", Attribute: 1})
			staff = append(staff, &archive.Staff{Aid: aid, Mid: 4444, Title: "2223", Ctime: "2018-11-28T16:50:14+08:00"})
			tx, err := d.BeginTran(c)
			ctx.So(err, convey.ShouldBeNil)
			err = d.TxAddStaff(tx, staff)
			ctx.So(err, convey.ShouldBeNil)
			err = tx.Commit()
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_UpPassed(t *testing.T) {
	convey.Convey("UpPassed", t, func(ctx convey.C) {
		_, err := d.UpPassed(context.TODO(), 1684013)
		ctx.So(err, convey.ShouldBeNil)
	})
}

func Test_Archive(t *testing.T) {
	convey.Convey("Archive", t, func(ctx convey.C) {
		arc, ip, err := d.RawArc(context.TODO(), 520090699)
		fmt.Printf("arc(%+v), ip(%s)", arc, string(ip))
		ctx.So(err, convey.ShouldBeNil)
	})
}

func Test_TxAddArchive(t *testing.T) {
	convey.Convey("TxAddArchive", t, func(ctx convey.C) {
		tx, err := d.BeginTran(context.TODO())
		ctx.So(err, convey.ShouldBeNil)
		_, err = d.TxAddArchive(tx, &archive.Archive{}, &archive.Addit{}, 0, 0, "", "")
		ctx.So(err, convey.ShouldBeNil)
		err = tx.Commit()
		ctx.So(err, convey.ShouldBeNil)
	})
}

func Test_TxUpArchive(t *testing.T) {
	convey.Convey("TxUpArchive", t, func(ctx convey.C) {
		tx, err := d.BeginTran(context.TODO())
		ctx.So(err, convey.ShouldBeNil)
		_, err = d.TxUpArchive(tx, &archive.Archive{ID: 0}, &archive.Addit{}, 0, 0, "", "")
		ctx.So(err, convey.ShouldBeNil)
		err = tx.Commit()
		ctx.So(err, convey.ShouldBeNil)
	})
}

func TestUpArcSID(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(1)
		sid = int64(1)
	)
	convey.Convey("TxUpArcSID", t, func(ctx convey.C) {
		err := d.UpArcSID(c, sid, aid)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDelArcSID(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(1)
	)
	convey.Convey("TxUpArcSID", t, func(ctx convey.C) {
		err := d.DelArcSID(c, 0, aid)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_TxAddVideo(t *testing.T) {
	convey.Convey("Archive", t, func() {
		tx, err := d.BeginTran(context.TODO())
		convey.So(err, convey.ShouldBeNil)
		_, err = d.TxAddVideo(tx, &archive.Video{Aid: 1, Cid: 1}, "")
		convey.So(err, convey.ShouldBeNil)
		err = tx.Commit()
		convey.So(err, convey.ShouldBeNil)
	})
}

func Test_TxDelVideoByCid(t *testing.T) {
	convey.Convey("TxDelVideoByCid", t, func() {
		tx, err := d.BeginTran(context.TODO())
		convey.So(err, convey.ShouldBeNil)
		_, err = d.TxDelVideoByCid(tx, 1, 1)
		convey.So(err, convey.ShouldBeNil)
		err = tx.Commit()
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestResultTxSortVideos(t *testing.T) {
	var (
		aid = int64(10098500)
		cid = int64(8940666)
	)
	tx, err := d.BeginTran(context.Background())
	if err != nil {
		convey.Print(err)
		return
	}
	convey.Convey("TxSortVideos", t, func(ctx convey.C) {
		rows, err := d.TxSortVideos(tx, aid, cid)
		convey.Convey("Then err should be nil.rows should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(rows, convey.ShouldNotBeNil)
		})
		tx.Commit()
	})
}

func TestResultTxStickVideo(t *testing.T) {
	var (
		aid = int64(10098500)
		cid = int64(8940666)
	)
	tx, err := d.BeginTran(context.Background())
	if err != nil {
		convey.Println(err)
		return
	}
	convey.Convey("TxStickVideo", t, func(ctx convey.C) {
		rows, err := d.TxStickVideo(tx, aid, cid)
		ctx.Convey("Then err should be nil.rows should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(rows, convey.ShouldNotBeNil)
		})
		tx.Commit()
	})
}

func TestResultTxUpArcFirstCID(t *testing.T) {
	var (
		aid = int64(10098500)
		cid = int64(8940666)
	)
	tx, err := d.BeginTran(context.Background())
	if err != nil {
		convey.Println(err)
		return
	}
	convey.Convey("TxUpArcFirstCID", t, func(ctx convey.C) {
		rows, err := d.TxUpArcFirstCID(tx, aid, cid)
		convey.Println(rows)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
		tx.Commit()
	})
}

func TestCheckVideoShot(t *testing.T) {
	var (
		cid = int64(1)
		c   = context.Background()
	)
	convey.Convey("TestCheckVideoShot", t, func(ctx convey.C) {
		err := d.CheckVideoShot(c, cid, 0)
		convey.Println(err)
	})
}
