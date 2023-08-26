package vote

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/ecode"
	"go-gateway/app/web-svr/activity/interface/api"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRankExternal(t *testing.T) {
	activityId := int64(1)
	sourceGroupId := int64(1)
	mid := int64(123400)
	ctx := context.Background()
	Convey("Rank", t, func() {
		Convey("RefreshVoteActivityRankExternal", func() {
			err := testDao.RefreshVoteActivityRankExternal(ctx, activityId)
			So(err, ShouldBeNil)
		})
		Convey("GetDSGRankExternal", func() {
			params := &model.InnerRankParams{
				Mid:               mid,
				ActivityId:        activityId,
				DataSourceGroupId: sourceGroupId,
				Version:           0,
				Pn:                1,
				Ps:                5,
			}
			res, err := testDao.GetDSGRankExternal(ctx, params)
			So(err, ShouldBeNil)
			bs, _ := json.Marshal(res)
			fmt.Printf("%v\n", string(bs))
		})
		Convey("GetDSGRankExternal-not-rand", func() {
			params := &model.InnerRankParams{
				Mid:               mid,
				ActivityId:        activityId,
				DataSourceGroupId: sourceGroupId,
				Version:           0,
				Pn:                1,
				Ps:                20,
			}
			res, err := testDao.GetDSGRankExternalOrder(ctx, false, params)
			So(err, ShouldBeNil)
			bs, _ := json.Marshal(res)
			fmt.Printf("%v\n", string(bs))
		})
		Convey("GetDSGRankExternal-rand", func() {
			params := &model.InnerRankParams{
				Mid:               mid,
				ActivityId:        activityId,
				DataSourceGroupId: sourceGroupId,
				Version:           0,
				Pn:                1,
				Ps:                5,
			}
			res, err := testDao.GetDSGRankExternalOrder(ctx, true, params)
			So(err, ShouldBeNil)
			bs, _ := json.Marshal(res)
			fmt.Printf("%v\n", string(bs))
		})
	})
}

func TestRankInternal(t *testing.T) {
	activityId := int64(1)
	sourceGroupId := int64(1)
	ctx := context.Background()
	Convey("Rank", t, func() {
		Convey("RefreshVoteActivityRankInternal", func() {
			err := testDao.RefreshVoteActivityRankInternal(ctx, activityId)
			So(err, ShouldBeNil)
		})
		Convey("GetDSGRankInternal", func() {
			res, err := testDao.GetDSGRankInternal(ctx, &api.GetVoteActivityRankInternalReq{
				SourceGroupId: sourceGroupId,
				Pn:            1,
				Ps:            30,
			})
			So(err, ShouldBeNil)
			bs, _ := json.Marshal(res)
			fmt.Printf("%v\n", string(bs))
		})
		Convey("GetDSGRankInternal-1", func() {
			_, err := testDao.GetDSGRankInternal(ctx, &api.GetVoteActivityRankInternalReq{
				SourceGroupId: 88,
				Pn:            1,
				Ps:            30,
			})
			So(err, ShouldEqual, ecode.ActivityVoteDSGNotFound)
		})
	})
}

func TestRankSearch(t *testing.T) {
	sourceGroupId := int64(1)
	mid := int64(123400)
	ctx := context.Background()
	Convey("Rank", t, func() {
		Convey("Search", func() {
			res, err := testDao.Search(ctx, sourceGroupId, mid, "test-1892", 10)
			So(err, ShouldBeNil)
			bs, _ := json.Marshal(res)
			fmt.Println(string(bs))
		})
	})
}
