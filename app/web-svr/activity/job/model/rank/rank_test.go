package rank

import (
	"math/rand"
	"testing"

	"github.com/glycerine/goconvey/convey"
)

type RankTest struct {
	Score int64
}
type RankBatchTest struct {
	Data      []RankTest
	TopLength int
}

// Len 返回长度
func (r *RankBatchTest) Len() int {
	return len(r.Data)
}

// TopLen 返回top的长度
func (r *RankBatchTest) TopLen() int {
	if r.Len() < r.TopLength {
		return r.Len()
	}
	return r.TopLength
}

// Less 比较
func (r *RankBatchTest) Less(i, j int) bool {
	return (*r).Data[i].Score > (*r).Data[j].Score
}

// Swap 交换
func (r *RankBatchTest) Swap(i, j int) {
	(*r).Data[i], (*r).Data[j] = (*r).Data[j], (*r).Data[i]
}

// Cut 切除
func (r *RankBatchTest) Cut(len int) {
	(*r).Data = (*r).Data[:len]
}

// Append 添加
func (r *RankBatchTest) Append(remainData Interface) {
	remain := remainData.(*RankBatchTest)
	pRemain := *remain
	for i := range pRemain.Data {
		(*r).Data = append((*r).Data, pRemain.Data[i])
	}
}
func TestSort(t *testing.T) {

	convey.Convey("Sort", t, func(convCtx convey.C) {
		r := RankBatchTest{}
		r.TopLength = 10
		r.Data = append(r.Data, RankTest{Score: 1},
			RankTest{Score: 9},
			RankTest{Score: 4},
			RankTest{Score: 5},
			RankTest{Score: 10},
			RankTest{Score: 3},
			RankTest{Score: 2},
			RankTest{Score: 10},
			RankTest{Score: 4},
			RankTest{Score: 2},
		)
		Sort(&r)
		expected := RankBatchTest{}
		expected.Data = append(expected.Data, RankTest{Score: 10},
			RankTest{Score: 10},
			RankTest{Score: 9},
			RankTest{Score: 5},
			RankTest{Score: 4},
			RankTest{Score: 4},
			RankTest{Score: 3},
			RankTest{Score: 2},
			RankTest{Score: 2},
			RankTest{Score: 1},
		)
		for i := range r.Data {
			convCtx.So(r.Data[i].Score, convey.ShouldResemble, expected.Data[i].Score)

		}

	})

	convey.Convey("Sort large nums", t, func(convCtx convey.C) {
		r := RankBatchTest{}
		r.TopLength = 100
		for i := 0; i < 5000; i++ {
			r.Data = append(r.Data, RankTest{Score: int64(rand.Intn(1000000))})

		}
		var err error
		Sort(&r)
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
		r := RankBatchTest{}
		remain := RankBatchTest{}
		r.TopLength = 10
		r.Data = append(r.Data, RankTest{Score: 1},
			RankTest{Score: 9},
			RankTest{Score: 4},
			RankTest{Score: 5},
			RankTest{Score: 10},
			RankTest{Score: 3},
			RankTest{Score: 2},
			RankTest{Score: 10},
			RankTest{Score: 4},
			RankTest{Score: 2},
		)
		remain.Data = append(remain.Data, RankTest{Score: 91},
			RankTest{Score: 10},
			RankTest{Score: 44},
			RankTest{Score: 33},
			RankTest{Score: 422},
			RankTest{Score: 412},
			RankTest{Score: 55},
			RankTest{Score: 123},
			RankTest{Score: 55},
			RankTest{Score: 66},
		)
		Sort(&r)
		Add(&r, &remain)
		expected := RankBatchTest{}
		expected.Data = append(expected.Data, RankTest{Score: 422},
			RankTest{Score: 412},
			RankTest{Score: 123},
			RankTest{Score: 91},
			RankTest{Score: 66},
			RankTest{Score: 55},
			RankTest{Score: 55},
			RankTest{Score: 44},
			RankTest{Score: 33},
			RankTest{Score: 10},
		)
		for i := range expected.Data {
			convCtx.So(r.Data[i].Score, convey.ShouldResemble, expected.Data[i].Score)
		}

	})
}
