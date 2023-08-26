package article

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/job/model/note"

	"github.com/smartystreets/goconvey/convey"
)

func TestCreateArticle(t *testing.T) {
	c := context.Background()
	convey.Convey("CreateArticle", t, func(ctx convey.C) {
		msg := &note.NtPubMsg{
			Mid:      27515242,
			NoteId:   10000,
			ContLen:  100,
			Title:    "3图模式🤮",
			Summary:  "撒旦访问如同4让它4 243唐仁4头若2她 ",
			Oid:      840016893,
			OidType:  0,
			ArcCover: "http://i1.hdslb.com/bfs/archive/6de59d2f7dc7273663363c1aa9acc2557cbf6597.jpg",
		}
		htmlCont := `<h1><span class="color-yellow-04"></span></h1>`
		imageUrls := []string{"//i0.hdslb.com/bfs/article/84ae95ebd02b02f13988b028a82d28d69c638f55.png", "//i0.hdslb.com/bfs/article/84ae95ebd02b02f13988b028a82d28d69c638f55.png", "//i0.hdslb.com/bfs/article/4e5318183ced1067167e5da9fd3f934142cde847.png"}
		argArticle := msg.ToArgArticle(0, htmlCont, imageUrls, 100)
		cvid, _, err := d.CreateArticle(c, argArticle, 1234)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(cvid, convey.ShouldBeGreaterThan, 0)
		})
	})
}

func TestEditArticle(t *testing.T) {
	c := context.Background()
	convey.Convey("EditArticle", t, func(ctx convey.C) {
		msg := &note.NtPubMsg{
			Mid:      27515242,
			NoteId:   10106064361488398,
			ContLen:  100,
			Title:    "coverAvid",
			Summary:  "123212345643撒旦访问如同4让它4 243唐仁4头若2她 ",
			Oid:      840122476,
			OidType:  0,
			ArcCover: "http://uat-i0.hdslb.com/bfs/archive/8470f18e415f8f9fae71b82e4e0f68635928a361.jpg",
		}
		htmlCont := `<h1><span class="color-yellow-04"></span></h1>`
		imageUrls := []string{"//uat-i0.hdslb.com/bfs/note/372f4ad9cb5e3bcd0361618ad1ef16663a255a72.jpg"}
		argArticle := msg.ToArgArticle(4421, htmlCont, imageUrls, 100)
		_, err := d.EditArticle(c, argArticle, 1234)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
