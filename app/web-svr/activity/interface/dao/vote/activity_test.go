package vote

import (
	"context"
	"fmt"
	"go-gateway/app/web-svr/activity/interface/api"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_ActivityCURD(t *testing.T) {
	ctx := context.Background()
	var activityId int64
	var dsIdToDel int64
	Convey("Activity CURD", t, func() {
		Convey("AddActivity", func() {
			err := testDao.AddActivity(ctx, &api.AddVoteActivityReq{
				Name:      "test-1",
				StartTime: time.Now().Unix(),
				EndTime:   time.Now().Add(time.Hour * 24).Unix(),
				Creator:   "dujinyang",
			})
			So(err, ShouldBeNil)
		})
		Convey("AddActivity-Finished", func() {
			err := testDao.AddActivity(ctx, &api.AddVoteActivityReq{
				Name:      "blabla",
				StartTime: time.Now().Add(-time.Hour * 24).Unix(),
				EndTime:   time.Now().Add(-time.Hour).Unix(),
				Creator:   "dujinyang",
			})
			So(err, ShouldBeNil)
		})
		Convey("ListVoteActivityForRefresh", func() {
			res, err := testDao.ListVoteActivityForRefresh(ctx, &api.ListVoteActivityForRefreshReq{
				Type: api.ListVoteActivityForRefreshReqType_ListVoteActivityForRefreshReqTypeNotEnded,
			})
			So(err, ShouldBeNil)
			So(len(res.Activitys), ShouldEqual, 1)
			So(res.Activitys[0].Name, ShouldEqual, "test-1")
			So(res.Activitys[0].Creator, ShouldEqual, "dujinyang")
			activityId = res.Activitys[0].Id
		})
		Convey("ListActivity-no-filter", func() {
			res, err := testDao.ListActivity(ctx, &api.ListVoteActivityReq{
				Pn:      1,
				Ps:      20,
				Ongoing: 0,
				Keyword: "",
			})
			So(err, ShouldBeNil)
			So(len(res.Activitys), ShouldEqual, 2)
			So(res.Activitys[0].Name, ShouldEqual, "blabla")
		})
		Convey("ListActivity-ongoing", func() {
			res, err := testDao.ListActivity(ctx, &api.ListVoteActivityReq{
				Pn:      1,
				Ps:      20,
				Ongoing: 1,
				Keyword: "",
			})
			So(err, ShouldBeNil)
			So(len(res.Activitys), ShouldEqual, 1)
			So(res.Activitys[0].Name, ShouldEqual, "test-1")
		})
		Convey("ListActivity-not-ongoing", func() {
			res, err := testDao.ListActivity(ctx, &api.ListVoteActivityReq{
				Pn:      1,
				Ps:      20,
				Ongoing: 2,
				Keyword: "",
			})
			So(err, ShouldBeNil)
			So(len(res.Activitys), ShouldEqual, 1)
			So(res.Activitys[0].Name, ShouldEqual, "blabla")
		})
		Convey("ListActivity-not-ongoing-keyword", func() {
			res, err := testDao.ListActivity(ctx, &api.ListVoteActivityReq{
				Pn:      1,
				Ps:      20,
				Ongoing: 2,
				Keyword: "blabla",
			})
			So(err, ShouldBeNil)
			So(len(res.Activitys), ShouldEqual, 1)
			So(res.Activitys[0].Name, ShouldEqual, "blabla")
		})
		Convey("ListActivity-not-ongoing-keyword-2", func() {
			res, err := testDao.ListActivity(ctx, &api.ListVoteActivityReq{
				Pn:      1,
				Ps:      20,
				Ongoing: 2,
				Keyword: "2",
			})
			So(err, ShouldBeNil)
			So(len(res.Activitys), ShouldEqual, 1)
			So(res.Activitys[0].Name, ShouldEqual, "blabla")
		})
		Convey("UpdateActivityRule", func() {
			err := testDao.UpdateActivityRule(ctx, &api.UpdateVoteActivityRuleReq{
				ActivityId:           activityId,
				SingleDayLimit:       2,
				TotalLimit:           4,
				SingleOptionBehavior: int64(api.VoteSingleOptionBehavior_VoteSingleOptionBehaviorDayOnce),
				RiskControlRule:      "NULL",
				DisplayRiskVote:      true,
				DisplayVoteCount:     true,
				VoteUpdateRule:       int64(api.VoteCountUpdateRule_VoteCountUpdateRuleRealTime),
				VoteUpdateCron:       0,
			})
			So(err, ShouldBeNil)
		})
		Convey("AddActivityDataSourceGroup-1", func() {
			err := testDao.AddActivityDataSourceGroup(ctx, &api.AddVoteActivityDataSourceGroupReq{
				ActivityId: activityId,
				SourceType: "TEST",
				SourceId:   888,
			})
			So(err, ShouldBeNil)
		})
		Convey("AddActivityDataSourceGroup-2", func() {
			err := testDao.AddActivityDataSourceGroup(ctx, &api.AddVoteActivityDataSourceGroupReq{
				ActivityId: activityId,
				SourceType: "TEST",
				SourceId:   889,
			})
			So(err, ShouldBeNil)
		})
		Convey("AddActivityDataSourceGroup-3", func() {
			err := testDao.AddActivityDataSourceGroup(ctx, &api.AddVoteActivityDataSourceGroupReq{
				ActivityId: activityId,
				SourceType: "TEST",
				SourceId:   890,
			})
			So(err, ShouldBeNil)
		})
		Convey("UpdateActivityDataSourceGroup", func() {
			err := testDao.UpdateActivityDataSourceGroup(ctx, &api.UpdateVoteActivityDataSourceGroupReq{
				GroupId:    3,
				ActivityId: activityId,
				SourceType: "TEST",
				SourceId:   77,
			})
			So(err, ShouldBeNil)
		})
		Convey("ListActivityDataSourceGroups", func() {
			DSG, err := testDao.ListActivityDataSourceGroups(ctx, &api.ListVoteActivityDataSourceGroupsReq{ActivityId: activityId})
			So(err, ShouldBeNil)
			So(DSG.Groups, ShouldNotBeEmpty)
			So(len(DSG.Groups), ShouldEqual, 3)
			So(DSG.Groups[1].SourceId, ShouldEqual, 889)
			So(DSG.Groups[1].SourceType, ShouldEqual, "TEST")
			So(DSG.Groups[1].GroupId, ShouldNotEqual, int64(0))
			dsIdToDel = DSG.Groups[1].GroupId

		})
		Convey("DelActivityDataSourceGroup", func() {
			fmt.Println(dsIdToDel)
			err := testDao.DelActivityDataSourceGroup(ctx, &api.DelVoteActivityDataSourceGroupReq{
				ActivityId: activityId,
				GroupId:    dsIdToDel,
			})
			So(err, ShouldBeNil)
		})
		Convey("ListActivityDataSourceGroups-2", func() {
			DSG, err := testDao.ListActivityDataSourceGroups(ctx, &api.ListVoteActivityDataSourceGroupsReq{ActivityId: activityId})
			So(err, ShouldBeNil)
			So(DSG.Groups, ShouldNotBeEmpty)
			So(len(DSG.Groups), ShouldEqual, 2)
			So(DSG.Groups[0].SourceId, ShouldEqual, 888)
			So(DSG.Groups[0].SourceType, ShouldEqual, "TEST")
		})
		//Convey("DeleteActivity", func() {
		//	err := testDao.DelActivity(ctx, &api.DelVoteActivityReq{Id: activityId})
		//	So(err, ShouldBeNil)
		//})
		//Convey("RawActivitysAll-Fin", func() {
		//	res, err := testDao.ListVoteActivityForRefresh(ctx, &api.ListOngoingVoteActivityReq{})
		//	So(err, ShouldBeNil)
		//	So(res.Activitys, ShouldBeEmpty)
		//})
	})

}
