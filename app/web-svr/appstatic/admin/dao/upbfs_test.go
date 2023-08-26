package dao

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestDaoUpload(t *testing.T) {
	var (
		c        = context.Background()
		fileType = "txt"
		timing   = time.Now().Unix()
		data     = []byte("123")
		fileName = "test.123"
		url      = fmt.Sprintf(d.c.Bfs.Host+_uploadURL, "app-static", fileName)
	)
	convey.Convey("UploadChronos", t, func(ctx convey.C) {
		ctx.Convey("everything is fine", func(ctx convey.C) {
			// httpMock(_method, url).Reply(200).SetHeader("Location", "test").SetHeader("code", "200").JSON(`{"code":200}`)
			location, err := d.Upload(c, "", fileType, timing, data)
			convey.Println("loc: ", location)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(location, convey.ShouldNotBeNil)
		})
		ctx.Convey("http code error", func(ctx convey.C) {
			httpMock(_method, url).Reply(-400)
			_, err := d.Upload(c, "", fileType, timing, data)
			ctx.So(err, convey.ShouldNotBeNil)
		})
		ctx.Convey("business code error", func(ctx convey.C) {
			httpMock(_method, url).Reply(200).JSON(`{"code":-400}`)
			_, err := d.Upload(c, fileName, fileType, timing, data)
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestDaoauthorize(t *testing.T) {
	var (
		key    = "key"
		secret = "secret"
		method = "put"
		bucket = "tv-cover"
		file   = "file"
		expire = int64(0)
	)
	convey.Convey("authorize", t, func(ctx convey.C) {
		authorization := authorize(key, secret, method, bucket, file, expire)
		ctx.Convey("Then authorization should not be nil.", func(ctx convey.C) {
			ctx.So(authorization, convey.ShouldNotBeNil)
		})
	})
}
