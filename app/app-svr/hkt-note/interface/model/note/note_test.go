package note

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestToContentLen(t *testing.T) {
	convey.Convey("ToContentLen", t, func(convCtx convey.C) {
		data := `[{"test":"111"},{"insert":{}},{"insert":"111"},{"insert":{}},{"insert":"999"},{"insert":{"imageUpload":{"url":"http://uat-api.bilibili.com/x/note/image?image_id=31","status":"done","width":310}}},{"insert":{"tag":{"cid":10263970,"status":0,"index":1,"seconds":0,"cidCount":1,"key":"1611745525472","title":"123机器人_bilibili"}}},{"insert":"111\n\n"},{"insert":{}},{"insert":"\n\n\n999\n"},{"insert":{"imageUpload":{"url":"http://uat-api.bilibili.com/x/note/image?image_id=31","status":"done","width":310}}},{"insert":{"tag":{"cid":10263970,"status":0,"index":1,"seconds":0,"cidCount":1,"key":"1611745525472","title":"123机器人_bilibili"}}},{"insert":"\n999\n"},{"insert":{"imageUpload":{"url":"//uat-api.bilibili.com/x/note/image?image_id=31","status":"done","width":310}}},{"insert":"\n"},{"insert":{"tag":{"cid":10263970,"status":0,"index":1,"seconds":0,"cidCount":1,"key":"1611745525472","title":"123机器人_bilibili"}}},{"insert":"\n\n"}]`
		count := ToContentLen(data)
		convCtx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(count, convey.ShouldBeGreaterThan, 0)
		})
	})
}
