package web

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
)

func TestUploadBFS(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("UploadBFS", t, func(ctx convey.C) {
		lastDay := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		pushFileName := fmt.Sprintf("push_arc_%s.json", lastDay)
		content := "{}"
		path, err := d.UploadBFS(c, pushFileName, []byte(content))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%s", path)
		})
	})
}

func TestReadURLContent(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("ReadURLContent", t, func(ctx convey.C) {
		outURL := "http://uat-i0.hdslb.com/bfs/active/outarc/push_arc_2020-07-13.json"
		fmt.Printf("timeout:%v", d.c.Rule.ReadTimeout)
		path, err := d.ReadURLContent(c, outURL)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.Printf("%s", path)
		})
	})
}
