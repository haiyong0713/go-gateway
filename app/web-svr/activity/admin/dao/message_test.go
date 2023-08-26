package dao

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/admin/model"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDaogetMsgKey(t *testing.T) {
	Convey("getMsgKey", t, func() {
		var (
			c = context.Background()
			l = &model.LetterParam{RecverIDs: []uint64{15555180}, SenderUID: 37090048, MsgType: 1, Content: "test"}
		)
		Convey("When everything goes positive", func() {
			res, code, err := d.getMsgKey(c, l)
			Convey("Then err should be nil.res,code should not be nil.", func() {
				So(err, ShouldBeNil)
				Println(code)
				Println(res)
			})
		})
	})
}

func TestDaoSendLetter(t *testing.T) {
	Convey("SendLetter", t, func() {
		var (
			c = context.Background()
			l = &model.LetterParam{RecverIDs: []uint64{15555180}, SenderUID: 37090048, MsgType: 1, Content: "test"}
		)
		Convey("When everything goes positive", func() {
			code, err := d.SendLetter(c, l)
			Convey("Then err should be nil.code should not be nil.", func() {
				So(err, ShouldBeNil)
				Println(code)
			})
		})
	})
}
