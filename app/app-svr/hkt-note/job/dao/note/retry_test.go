package note

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/smartystreets/goconvey/convey"
)

func TestCacheRetry(t *testing.T) {
	c := context.Background()
	convey.Convey("DelNoteDetail", t, func(ctx convey.C) {
		res, err := d.CacheRetry(c, note.KeyRetryDBDelCont)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(len(res), convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestAddCacheRetry(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheRetry", t, func(ctx convey.C) {
		req := &note.NtAddMsg{
			Aid:      1,
			Mid:      1,
			NoteId:   2,
			Title:    "title",
			Summary:  "summary",
			NoteSize: 0,
			Mtime:    0,
		}
		jsonBody, _ := json.Marshal(req)
		d.AddCacheRetry(c, note.KeyRetryDBDetail, string(jsonBody), time.Now().Unix())
	})
}
