package dao

import (
	"testing"

	"go-gateway/app/web-svr/space/job/internal/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_SendLetter(t *testing.T) {
	Convey("SendLetter", t, func(convCtx C) {
		Convey("When everything goes positive", func(convCtx C) {
			err := d.SendLetter(ctx, &model.LetterParam{
				RecverIDs: []uint64{27515257},
				SenderUID: 12076317,
				MsgType:   1,
				Content:   "你配置的手机端空间头图视频已失效，请进入空间重新设置",
			})
			So(err, ShouldBeNil)
		})
	})
}
