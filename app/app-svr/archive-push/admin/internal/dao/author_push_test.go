package dao

import (
	"github.com/glycerine/goconvey/convey"
	"testing"
)

func Test_GetActiveAuthorPushesByAuthorIDs(t *testing.T) {
	convey.Convey("GetActiveAuthorPushesByAuthorIDs", t, func() {
		authorIDs := []int64{17}
		res, err := testD.GetActiveAuthorPushesByAuthorIDs(authorIDs)
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(res)
	})
}
