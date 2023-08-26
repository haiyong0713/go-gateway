package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_UploadFile(t *testing.T) {
	var (
		localPath      = "/tmp/esports"
		remoteDir      = "/open/esports"
		remoteFileName = "esports"
	)
	convey.Convey("UploadFile", t, func(ctx convey.C) {
		err := d.UploadFile(localPath, remoteDir, remoteFileName)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_FileMd5(t *testing.T) {
	var (
		localPath = "/tmp/esports"
		remoteDir = "/tmp/esports.md5"
	)
	convey.Convey("UploadFile", t, func(ctx convey.C) {
		err := d.FileMd5(localPath, remoteDir)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaoSeason(t *testing.T) {
	convey.Convey("Season", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.Season(c, 1, 10)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_SeasonCount(t *testing.T) {
	convey.Convey("Season", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.SeasonCount(c)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			fmt.Printf("SeasonCount is %d", res)
		})
	})
}

func TestDaoSeasonVa(t *testing.T) {
	convey.Convey("Season", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.Season(c, 1, 10)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(res, convey.ShouldNotBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_FtpTeamsCount(t *testing.T) {
	convey.Convey("Season", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.FtpTeamsCount(c)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			fmt.Printf("SeasonCount is %d", res)
		})
	})
}

func TestDao_FtpTeams(t *testing.T) {
	convey.Convey("Season", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.FtpTeams(c, 0, 10)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_FtpContests(t *testing.T) {
	convey.Convey("FtpContests", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.FtpContests(c, 0, 10)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_FtpContestsCount(t *testing.T) {
	convey.Convey("FtpContestsCount", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.FtpContestsCount(c)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_FtpMatchsCount(t *testing.T) {
	convey.Convey("FtpMatchsCount", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.FtpMatchsCount(c)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}

func TestDao_FtpMatchs(t *testing.T) {
	convey.Convey("FtpContests", t, func(ctx convey.C) {
		var (
			c = context.Background()
		)
		ctx.Convey("When everything goes positive", func(ctx convey.C) {
			res, err := d.FtpMatchs(c, 0, 10)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
