package dao

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/siddontang/go-mysql/mysql"
	"github.com/smartystreets/goconvey/convey"
)

func TestDaoRemoveHvarRec(t *testing.T) {
	var (
		periodValid = int(90)
	)
	convey.Convey("RemoveHvarRec", t, func(ctx convey.C) {
		err := d.RemoveHvarRec(periodValid)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDaopickHvarRec(t *testing.T) {
	var (
		periodValid = int(90)
		total       = int64(0)
		expired     = time.Now().AddDate(0, 0, -periodValid)
	)
	convey.Convey("pickHvarRec", t, func(ctx convey.C) {
		for index := int64(0); index < _sharding; index++ {
			var count int64
			err := d.db.QueryRow(context.Background(), fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE mtime < \"%s\"", tableName(index), expired.Format(mysql.TimeFormat))).Scan(&count)
			if err != nil {
				convey.Println(err)
				return
			}
			total += count
			convey.Println("index ", index, " count ", count)
		}
		convey.Printf("Total %d", total)
	})
}

func TestDao_GetHvarLock(t *testing.T) {
	convey.Convey("RemoveHvarRec", t, func(ctx convey.C) {
		get, err := d.GetHvarLock(context.Background())
		convey.Println(get)
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestDao_DelHvarLock(t *testing.T) {
	convey.Convey("TestDao_DelHvarLock", t, func(ctx convey.C) {
		err := d.DelHvarLock(context.Background())
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
