package like

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/model/like"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLikekeyLikeCheck(t *testing.T) {
	Convey("keyLikeCheck", t, func() {
		var (
			mid = int64(0)
			sid = int64(0)
		)
		Convey("When everything goes positive", func() {
			p1 := keyLikeCheck(mid, sid)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestLikekeyActOnGoing(t *testing.T) {
	Convey("keyActOnGoing", t, func() {
		var (
			typeIds = []int64{}
		)
		Convey("When everything goes positive", func() {
			p1 := keyActOnGoing(typeIds)
			Convey("Then p1 should not be nil.", func() {
				So(p1, ShouldNotBeNil)
			})
		})
	})
}

func TestLikeCacheLikeCheck(t *testing.T) {
	Convey("CacheLikeCheck", t, func() {
		var (
			c   = context.Background()
			mid = int64(0)
			sid = int64(0)
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheLikeCheck(c, mid, sid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeAddCacheLikeCheck(t *testing.T) {
	Convey("AddCacheLikeCheck", t, func() {
		var (
			c    = context.Background()
			mid  = int64(0)
			data = &like.Item{}
			sid  = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheLikeCheck(c, mid, data, sid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeCacheActSubjectsOnGoing(t *testing.T) {
	Convey("CacheActSubjectsOnGoing", t, func() {
		var (
			c       = context.Background()
			typeIds = []int64{1, 4, 13}
		)
		Convey("When everything goes positive", func() {
			res, err := d.CacheActSubjectsOnGoing(c, typeIds)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeAddCacheActSubjectsOnGoing(t *testing.T) {
	Convey("AddCacheActSubjectsOnGoing", t, func() {
		var (
			c       = context.Background()
			typeIds = []int64{}
			list    = []int64{}
		)
		Convey("When everything goes positive", func() {
			err := d.AddCacheActSubjectsOnGoing(c, typeIds, list)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeCacheActivityArchives(t *testing.T) {
	Convey("CacheActivityArchives", t, func() {
		ctx := context.Background()
		sid := int64(10680)
		mid := int64(15555180)
		Convey("When everything goes positive", func() {
			res, err := d.CacheActivityArchives(ctx, sid, mid)
			Convey("Then err should be nil.res should not be nil.", func() {
				So(err, ShouldBeNil)
				Print(res)
			})
		})
	})
}

func TestLikeAddCacheActivityArchives(t *testing.T) {
	Convey("AddCacheActivityArchives", t, func() {
		ctx := context.Background()
		sid := int64(10680)
		mid := int64(15555180)
		arcs := []*like.Item{{ID: 1, Wid: 110}}
		Convey("When everything goes positive", func() {
			err := d.AddCacheActivityArchives(ctx, sid, arcs, mid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestLikeDelCacheActivityArchives(t *testing.T) {
	Convey("AddCacheActivityArchives", t, func() {
		ctx := context.Background()
		sid := int64(10680)
		mid := int64(15555180)
		Convey("When everything goes positive", func() {
			err := d.DelCacheActivityArchives(ctx, sid, mid)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
