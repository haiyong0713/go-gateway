package show

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/show"

	"github.com/smartystreets/goconvey/convey"
)

func TestDao_SearchOgvAdd(t *testing.T) {
	convey.Convey("SearchOgvAdd", t, func(ctx convey.C) {
		//[{"plat":0,"build":521010,"conditions":"gt"},{"plat":1,"build":6500,"conditions":"gt"},{"plat":2,"build":12070,"conditions":"gt"}]
		var (
			param = &show.SearchOgvAP{
				Color:          1,
				Stime:          xtime.Time(time.Now().Unix()),
				Plat:           "[{\"plat\":0,\"build\":521010,\"conditions\":\"gt\"},{\"plat\":1,\"build\":6500,\"conditions\":\"gt\"}]",
				HdCover:        "http://bilibili/cover",
				HdTitle:        "标题",
				HdSubtitle:     "副标题",
				GameStatus:     1,
				GamePos:        1,
				GameValue:      "1",
				PgcPos:         2,
				PgcIds:         "34190",
				PgcMoreURL:     "http:pgc",
				MoreshowStatus: 1,
				MoreshowPos:    3,
				MoreshowValue:  "[{\"word\":\"word1\",\"type\":1,\"value\":\"value2\"},{\"word\":\"word2\",\"type\":1,\"value\":\"value2\"}]",
				Query:          "[{\"value\":\"test1sssss11aaa1a1\"}]",
				Person:         "person",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SearchOgvAdd(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_SearchOgvUpdate(t *testing.T) {
	convey.Convey("SearchOgvAdd", t, func(ctx convey.C) {
		//[{"plat":0,"build":521010,"conditions":"gt"},{"plat":1,"build":6500,"conditions":"gt"},{"plat":2,"build":12070,"conditions":"gt"}]
		var (
			param = &show.SearchOgvUP{
				ID:             1,
				Color:          2,
				Stime:          xtime.Time(time.Now().Unix()),
				Plat:           "[{\"plat\":0,\"build\":521010,\"conditions\":\"gt\"},{\"plat\":1,\"build\":6500,\"conditions\":\"gt\"}]",
				HdCover:        "http://bilibili/cover",
				HdTitle:        "标题",
				HdSubtitle:     "副标题",
				GameStatus:     1,
				GamePos:        1,
				GameValue:      "1",
				PgcPos:         2,
				PgcIds:         "34190",
				PgcMoreURL:     "http:pgc",
				MoreshowStatus: 1,
				MoreshowPos:    3,
				MoreshowValue:  "[{\"word\":\"word1\",\"type\":1,\"value\":\"value2\"}]",
				Query:          "[{\"value\":\"1test1ddddaaa11\"}]",
				Person:         "person",
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SearchOgvUpdate(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_SearchOgvOption(t *testing.T) {
	convey.Convey("SearchOgvAdd", t, func(ctx convey.C) {
		//[{"plat":0,"build":521010,"conditions":"gt"},{"plat":1,"build":6500,"conditions":"gt"},{"plat":2,"build":12070,"conditions":"gt"}]
		var (
			param = &show.SearchOgvOption{
				ID:    1,
				Check: 1,
			}
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SearchOgvOption(param)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_ValOgvQuery(t *testing.T) {
	convey.Convey("SearchOgvAdd", t, func(ctx convey.C) {
		query := "[{\"value\":\"我是新增的\"},{\"value\":\"test1111aaaa1\"},{\"value\":\"bbbbccccddd1\"}]"
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			_, err := d.ValOgvQuery(query)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestDao_SearchOgvFind(t *testing.T) {
	convey.Convey("SearchOgvFind", t, func(ctx convey.C) {
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			res, err := d.SearchOgvFind(72)
			ctx.Convey("Then err should be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
			fmt.Println("query string is ", res.QueryStr)
		})
	})
}
