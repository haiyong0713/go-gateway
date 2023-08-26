package dao

import (
	"github.com/glycerine/goconvey/convey"
	"testing"
)

func Test_GetArcByAID(t *testing.T) {
	convey.Convey("GetArcByAID", t, func() {
		aid := int64(840103236)
		res, err := testD.GetArcByAID(aid)
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(res)
	})
}
