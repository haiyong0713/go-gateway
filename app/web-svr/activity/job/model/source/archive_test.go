package source

import (
	rankmdl "go-gateway/app/web-svr/activity/job/model/rank_v2"
	"math/rand"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestSort(t *testing.T) {

	convey.Convey("Sort", t, func(convCtx convey.C) {
		r := OidArchiveGroup{}
		r.TopLength = 10
		r.Data = append(r.Data, &ArchiveGroup{Score: 1},
			&ArchiveGroup{Score: 9},
			&ArchiveGroup{Score: 4},
			&ArchiveGroup{Score: 5},
			&ArchiveGroup{Score: 10, HistoryRank: 2},
			&ArchiveGroup{Score: 3},
			&ArchiveGroup{Score: 2},
			&ArchiveGroup{Score: 10, HistoryRank: 1},
			&ArchiveGroup{Score: 4},
			&ArchiveGroup{Score: 2},
		)
		rankmdl.Sort(&r)
		expected := OidArchiveGroup{}
		expected.Data = append(expected.Data, &ArchiveGroup{Score: 10, HistoryRank: 1},
			&ArchiveGroup{Score: 10, HistoryRank: 2},
			&ArchiveGroup{Score: 9},
			&ArchiveGroup{Score: 5},
			&ArchiveGroup{Score: 4},
			&ArchiveGroup{Score: 4},
			&ArchiveGroup{Score: 3},
			&ArchiveGroup{Score: 2},
			&ArchiveGroup{Score: 2},
			&ArchiveGroup{Score: 1},
		)
		for i := range r.Data {
			convCtx.So(r.Data[i].Score, convey.ShouldResemble, expected.Data[i].Score)
			convCtx.So(r.Data[i].HistoryRank, convey.ShouldResemble, expected.Data[i].HistoryRank)

		}

	})

	convey.Convey("Sort large nums", t, func(convCtx convey.C) {
		r := OidArchiveGroup{}
		r.TopLength = 100
		for i := 0; i < 5000; i++ {
			r.Data = append(r.Data, &ArchiveGroup{Score: int64(rand.Intn(1000000))})

		}
		var err error
		rankmdl.Sort(&r)
		var history = r.Data[0].Score
		for i := range r.Data {
			if r.Data[i].Score > history {
				convCtx.So(err, convey.ShouldNotBeNil)
			}
			history = r.Data[i].Score
		}

	})
}

func TestSortAdd(t *testing.T) {
	convey.Convey("Sort add", t, func(convCtx convey.C) {
		r := OidArchiveGroup{}
		remain := OidArchiveGroup{}
		r.TopLength = 10
		r.Data = append(r.Data, &ArchiveGroup{Score: 1},
			&ArchiveGroup{Score: 9},
			&ArchiveGroup{Score: 4},
			&ArchiveGroup{Score: 5},
			&ArchiveGroup{Score: 10},
			&ArchiveGroup{Score: 3},
			&ArchiveGroup{Score: 2},
			&ArchiveGroup{Score: 10},
			&ArchiveGroup{Score: 4},
			&ArchiveGroup{Score: 2},
		)
		remain.Data = append(remain.Data, &ArchiveGroup{Score: 91},
			&ArchiveGroup{Score: 10},
			&ArchiveGroup{Score: 44},
			&ArchiveGroup{Score: 33},
			&ArchiveGroup{Score: 422},
			&ArchiveGroup{Score: 412},
			&ArchiveGroup{Score: 55},
			&ArchiveGroup{Score: 123},
			&ArchiveGroup{Score: 55},
			&ArchiveGroup{Score: 66},
		)
		rankmdl.Sort(&r)
		rankmdl.Add(&r, &remain)
		expected := OidArchiveGroup{}
		expected.Data = append(expected.Data, &ArchiveGroup{Score: 422},
			&ArchiveGroup{Score: 412},
			&ArchiveGroup{Score: 123},
			&ArchiveGroup{Score: 91},
			&ArchiveGroup{Score: 66},
			&ArchiveGroup{Score: 55},
			&ArchiveGroup{Score: 55},
			&ArchiveGroup{Score: 44},
			&ArchiveGroup{Score: 33},
			&ArchiveGroup{Score: 10},
		)
		for i := range expected.Data {
			convCtx.So(r.Data[i].Score, convey.ShouldResemble, expected.Data[i].Score)
		}

	})
}
