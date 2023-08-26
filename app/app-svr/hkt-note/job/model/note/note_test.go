package note

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestToDtlCache(t *testing.T) {
	convey.Convey("ToDtlCache", t, func(convCtx convey.C) {
		req := &NtDetailDB{
			NoteId:   1,
			Mid:      1,
			Aid:      0,
			NoteIdx:  0,
			NoteSize: 0,
			Title:    "",
			Summary:  "",
			Deleted:  0,
			Ctime:    "",
			Mtime:    "2020-08-14 17:59:14",
		}
		res := req.ToDtlCache()
		fmt.Println(res.Mtime)
	})
}

func TestToBody(t *testing.T) {
	convey.Convey("ToBody", t, func(convCtx convey.C) {
		data := "[{\"insert\":\"有总管\n都抖音就无总包\"},{\"insert\":{\"tag\":{\"cid\":245829118,\"status\":0,\"index\":2,\"seconds\":645,\"cidCount\":12,\"key\":\"1602813729629\",\"title\":\"02 02-施工管理1\"}}},{\"insert\":\"\\n\"},{\"insert\":123456789},{\"insert\":{\"test\":{\"cid\":245829118,\"status\":0,\"index\":2,\"seconds\":645,\"cidCount\":12,\"key\":\"1602813729629\",\"title\":\"02 02-施工管理1\"}}},{\"insert\":\"我就测\\n测不说话\"}]"
		res := ToBody(data)
		fmt.Println(res)
	})
}

func TestReplaceSensitive(t *testing.T) {
	convey.Convey("Replace", t, func(convCtx convey.C) {
		all := `[{"insert":{"imageUpload":{"url":"//uat-api.bilibili.com/x/note/image?location=/bfs/note/74c0d065f2d12ba1142512225bc0c8d3796cd908.png","status":"done","width":315}}},{"insert":"嘿嘿嘿\n\n\n"},{"insert":{}},{"insert":{"video":"https://player.bilibili.com/player.html?bvid=BV14y4y1M7cC"}},{"attributes":{"link":"https://bilibili.com"},"insert":"https://b23.tv/I17Vvi0"},{"insert":"\n"}]`
		res, err := ReplaceAndFilter(all, []string{"三十三", "阿富汗汗"})
		fmt.Println(res)
		convCtx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
