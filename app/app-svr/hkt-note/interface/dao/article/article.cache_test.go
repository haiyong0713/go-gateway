package article

import (
	"context"
	"testing"

	"go-gateway/app/app-svr/hkt-note/interface/model/article"
	notegrpc "go-gateway/app/app-svr/hkt-note/service/api"

	"github.com/smartystreets/goconvey/convey"
)

func TestAddCacheArtDetail(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheArtDetail", t, func(ctx convey.C) {
		val := &article.ArtDtlCache{
			Cvid:      123,
			NoteId:    1234,
			Mid:       1,
			PubStatus: 5,
			PubReason: "待审核",
			Oid:       10,
			OidType:   0,
			Pubtime:   0,
			Mtime:     0,
			Title:     "测试测试",
			Summary:   "",
			Deleted:   0,
		}
		err := d.AddCacheArtDetail(c, 123, "test", val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestRemCachesArtListUser(t *testing.T) {
	c := context.Background()
	convey.Convey("AddCacheArtDetail", t, func(ctx convey.C) {
		val := make(map[int64]*notegrpc.SimpleArticleCard)
		val[1] = &notegrpc.SimpleArticleCard{Cvid: 1, NoteId: 1}
		val[2] = &notegrpc.SimpleArticleCard{Cvid: 2, NoteId: 3}
		val[3] = &notegrpc.SimpleArticleCard{Cvid: 3, NoteId: 4}
		val[4] = &notegrpc.SimpleArticleCard{Cvid: 4, NoteId: 5}
		err := d.RemCachesArtListUser(c, 1, val)
		ctx.Convey("Then err should be nil.cids should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}
