package rank

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

func TestArchiveStatBatchScore(t *testing.T) {
	convey.Convey("ArchiveStatBatch Score", t, func(convCtx convey.C) {
		var a = ArchiveStatMap{}
		archiveBatch := []*ArchiveStat{
			{
				View:    333233,
				Coin:    8832,
				Videos:  3,
				Reply:   1231,
				Danmaku: 993,
				Fav:     9832,
				Like:    9312,
			},
			{
				View:    412,
				Coin:    34,
				Videos:  2,
				Reply:   32,
				Danmaku: 63,
				Fav:     42,
				Like:    66,
			},
		}
		a[100082] = archiveBatch
		res := *a.Score(testScore)
		expected := MidScoreMap{}
		expected[100082] = &MidScore{
			Score: 424478,
		}
		convCtx.So(res[100082].Score, convey.ShouldEqual, expected[100082].Score)

	})
}

func testScore(arc *ArchiveStat) int64 {
	return getPlayScore(arc) + getQualityScore(arc) + getTopicScore(arc)

}

func getPlayScore(arc *ArchiveStat) int64 {
	videos := float64(arc.Videos)
	views := float64(arc.View)
	pRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", (4/(videos+3))), 64)
	aRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", ((300000+views)/(2*views))), 64)
	if aRevise > 1 {
		aRevise = 1
	}
	return int64(math.Floor(views*pRevise*aRevise + 0.5))
}

func getQualityScore(arc *ArchiveStat) int64 {
	like := float64(arc.Like)
	coin := float64(arc.Coin)
	fav := float64(arc.Fav)
	views := float64(arc.View)
	quality := like*5 + coin*10 + fav*20
	bRevise, _ := strconv.ParseFloat(fmt.Sprintf("%.3f", ((like*5+coin*10+fav*20)/(views+like*5+coin*10+fav*20))), 64)
	return int64(math.Floor(quality*bRevise + 0.5))
}

func getTopicScore(arc *ArchiveStat) int64 {
	return int64((arc.Danmaku + arc.Reply)) * 20
}
