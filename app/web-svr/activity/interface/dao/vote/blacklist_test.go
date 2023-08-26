package vote

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/api"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBlackList(t *testing.T) {
	activityId := int64(1)
	sourceGroupId := int64(1)
	ctx := context.Background()
	Convey("BlackList", t, func() {
		Convey("AddVoteActivityBlackList", func() {
			err := testDao.AddVoteActivityBlackList(ctx, &api.AddVoteActivityBlackListReq{
				ActivityId:    activityId,
				SourceGroupId: sourceGroupId,
				SourceItemId:  2000,
			})
			So(err, ShouldBeNil)
			err = testDao.AddVoteActivityBlackList(ctx, &api.AddVoteActivityBlackListReq{
				ActivityId:    activityId,
				SourceGroupId: sourceGroupId,
				SourceItemId:  2000,
			})
			So(err, ShouldBeNil)
		})
		Convey("BlackListCheck", func() {
			b, err := testDao.BlackListCheck(ctx, sourceGroupId, 2000)
			So(err, ShouldBeNil)
			So(b, ShouldBeTrue)
			b, err = testDao.BlackListCheck(ctx, sourceGroupId, 1999)
			So(err, ShouldBeNil)
			So(b, ShouldBeFalse)
		})
		//Convey("DelVoteActivityBlackList", func() {
		//	_, err := testDao.DelVoteActivityBlackList(ctx, &api.DelVoteActivityBlackListReq{
		//		ActivityId:   activityId,
		//		SourceType:   "TEST",
		//		SourceId:     1,
		//		SourceItemId: 2000,
		//	})
		//	So(err, ShouldBeNil)
		//})
		//Convey("BlackListCheck-2", func() {
		//	b, err := testDao.BlackListCheck(ctx, "TEST", activityId, 1, 2000)
		//	So(err, ShouldBeNil)
		//	So(b, ShouldBeFalse)
		//})
	})
}
