package vote

import (
	"context"
	"encoding/json"
	"fmt"
	"go-gateway/app/web-svr/activity/ecode"
	riskModel "go-gateway/app/web-svr/activity/interface/model/risk"
	model "go-gateway/app/web-svr/activity/interface/model/vote"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestVote(t *testing.T) {
	activityId := int64(1)
	dataSourceGroupId := int64(1)
	mid := int64(123400)
	risk := &riskModel.Base{
		Buvid:      "",
		Origin:     "",
		Referer:    "",
		IP:         "",
		Ctime:      "",
		UserAgent:  "",
		Build:      "",
		Platform:   "",
		Action:     "",
		MID:        0,
		API:        "",
		EsTime:     0,
		ActivityID: activityId,
	}
	ctx := context.Background()
	Convey("Vote", t, func() {
		Convey("DoVote-Success-1", func() { //正常投票流程
			req := &model.DoVoteParams{
				ActivityId:        activityId,
				DataSourceGroupId: dataSourceGroupId,
				DataSourceItemId:  1888,
				Vote:              1,
			}
			err := testDao.DoVote(ctx, mid, risk, req)
			So(err, ShouldBeNil)
		})
		Convey("DoVote-Fail-1", func() { //黑名单校验
			req := &model.DoVoteParams{
				ActivityId:        activityId,
				DataSourceGroupId: dataSourceGroupId,
				DataSourceItemId:  2000,
				Vote:              1,
			}
			err := testDao.DoVote(ctx, mid, risk, req)
			So(err, ShouldEqual, ecode.ActivityVoteItemNotFound)
		})
		Convey("DoVote-Fail-2", func() { //单选项重复投票
			req := &model.DoVoteParams{
				ActivityId:        activityId,
				DataSourceGroupId: dataSourceGroupId,
				DataSourceItemId:  1888,
				Vote:              1,
			}
			err := testDao.DoVote(ctx, mid, risk, req)
			So(err, ShouldEqual, ecode.ActivityVoteItemVoted)
		})

		Convey("DoVote-Success-2", func() { //正常投票流程
			req := &model.DoVoteParams{
				ActivityId:        activityId,
				DataSourceGroupId: dataSourceGroupId,
				DataSourceItemId:  1889,
				Vote:              1,
			}
			err := testDao.DoVote(ctx, mid, risk, req)
			So(err, ShouldBeNil)
		})
		Convey("DoVote-Fail-3", func() { //投票次数不足
			req := &model.DoVoteParams{
				ActivityId:        activityId,
				DataSourceGroupId: dataSourceGroupId,
				DataSourceItemId:  1789,
				Vote:              1,
			}
			err := testDao.DoVote(ctx, mid, risk, req)
			So(err, ShouldEqual, ecode.ActivityVoteExceed)
		})

		Convey("RankAfterVote", func() { //重新查看排名
			Convey("RankAfterVote-RefreshVoteActivityRankExternal", func() {
				err := testDao.RefreshVoteActivityRankExternal(ctx, activityId)
				So(err, ShouldBeNil)
			})
			Convey("RankAfterVote-GetDSGRankExternal", func() {
				params := &model.InnerRankParams{
					Mid:               mid,
					ActivityId:        activityId,
					DataSourceGroupId: dataSourceGroupId,
					Version:           0,
					Pn:                1,
					Ps:                20,
				}
				res, err := testDao.GetDSGRankExternal(ctx, params)
				So(err, ShouldBeNil)
				bs, _ := json.Marshal(res)
				fmt.Printf("%v\n", string(bs))
			})
		})
		Convey("UndoVote-Success-1", func() { //取消投票
			req := &model.UndoVoteParams{
				ActivityId:        activityId,
				DataSourceGroupId: dataSourceGroupId,
				DataSourceItemId:  1889,
			}
			err := testDao.UndoVote(ctx, mid, req)
			So(err, ShouldBeNil)
		})

		Convey("UndoVote-Failed-1", func() { //再次取消
			req := &model.UndoVoteParams{
				ActivityId:        activityId,
				DataSourceGroupId: dataSourceGroupId,
				DataSourceItemId:  1889,
			}
			err := testDao.UndoVote(ctx, mid, req)
			So(err, ShouldEqual, ecode.ActivityVoteNoHistory)
		})

		Convey("RankAfterVoteUndo", func() { //重新刷新排名
			Convey("RankAfterVoteUndo-RefreshVoteActivityRankExternal", func() {
				err := testDao.RefreshVoteActivityRankExternal(ctx, activityId)
				So(err, ShouldBeNil)
			})
			Convey("RankAfterVoteUndo-GetDSGRankExternal", func() {
				params := &model.InnerRankParams{
					Mid:               mid,
					ActivityId:        activityId,
					DataSourceGroupId: dataSourceGroupId,
					Version:           0,
					Pn:                1,
					Ps:                20,
				}
				res, err := testDao.GetDSGRankExternal(ctx, params)
				So(err, ShouldBeNil)
				bs, _ := json.Marshal(res)
				fmt.Printf("%v\n", string(bs))
			})
		})

	})
}
