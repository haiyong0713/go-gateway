package service

import (
	"github.com/glycerine/goconvey/convey"
	"testing"
)

func Test_CheckIfAuthorInWhiteList(t *testing.T) {
	convey.Convey("CheckIfAuthorInWhiteList", t, func() {
		res, err := testService.CheckIfAuthorInWhiteList(1, 11)
		convey.ShouldBeNil(err)
		convey.ShouldNotBeNil(res)
	})
}
