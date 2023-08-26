package article

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/smartystreets/goconvey/convey"
)

func TestGetBiliNoteContent(t *testing.T) {
	c := context.Background()
	convey.Convey("GetBiliNoteContent", t, func(ctx convey.C) {
		biliJson := `[{"insert":"来几句话\n转一下格式\n"}]`
		res, _, err := d.GetBiliNoteContent(c, biliJson, &note.NtPubMsg{Mid: int64(27515242)})
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(res, convey.ShouldNotBeNil)
		})
	})
}
