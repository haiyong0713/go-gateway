package dao

import (
	"github.com/glycerine/goconvey/convey"
	"testing"
)

func Test_RemoveBatchesFromTodo(t *testing.T) {
	convey.Convey("RemoveBatchesFromTodo", t, func() {
		batchIDs := []int64{393}
		err := testD.RemoveBatchesFromTodo(batchIDs)
		convey.ShouldBeNil(err)
	})
}
