package note

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/interface/model/note"

	"github.com/smartystreets/goconvey/convey"
)

func TestUpContent(t *testing.T) {
	c := context.Background()
	convey.Convey("UpContent", t, func(ctx convey.C) {
		req := &note.NtContent{
			NoteId:  1,
			Content: "434245324524543434343",
			Tag:     "10208596-1-200,10208596-1-300,10208596-1-400",
		}
		err := d.UpContent(c, req)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
