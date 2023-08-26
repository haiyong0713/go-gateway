package like

import (
	"context"
	"go-gateway/app/web-svr/activity/job/model/task"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// //2,27515430
func TestDoTask(t *testing.T) {
	convey.Convey("do task", t, func(ctx convey.C) {
		var (
			c            = context.Background()
			taskID int64 = 2
			mid    int64 = 15555180
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.DoTask(c, taskID, mid)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestGetChildTask(t *testing.T) {
	convey.Convey("GetChildTask", t, func(ctx convey.C) {
		var (
			c            = context.Background()
			taskID int64 = 56
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			child, err := d.GetChildTask(c, taskID)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(child, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestSetTaskCount(t *testing.T) {
	convey.Convey("SetTaskCount", t, func(ctx convey.C) {
		var (
			c            = context.Background()
			taskID int64 = 56
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.SetTaskCount(c, taskID, 1111)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})

}

func TestActivityTaskMidStatus(t *testing.T) {
	convey.Convey("ActivityTaskMidStatus", t, func(ctx convey.C) {
		var (
			c             = context.Background()
			taskID  int64 = 56
			midRule map[int64][]*task.MidRule
		)
		midRule = make(map[int64][]*task.MidRule)
		midRule[1] = []*task.MidRule{
			{
				Object: 1,
				MID:    1,
				State:  1,
			},
			{
				Object: 2,
				MID:    1,
				State:  1,
			},
		}
		midRule[2] = []*task.MidRule{
			{
				Object: 1,
				MID:    1,
				State:  1,
			},
			{
				Object: 2,
				MID:    1,
				State:  0,
			},
		}
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			err := d.ActivityTaskMidStatus(c, taskID, midRule)
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetActivityTaskMidStatus(t *testing.T) {
	convey.Convey("GetActivityTaskMidStatus", t, func(ctx convey.C) {
		var (
			c            = context.Background()
			taskID int64 = 256
		)
		ctx.Convey("When everything gose positive", func(ctx convey.C) {
			task, err := d.GetActivityTaskMidStatus(c, taskID, []int64{1, 2})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(task, convey.ShouldNotBeNil)
			})
		})
		ctx.Convey("has no result ", func(ctx convey.C) {
			task, err := d.GetActivityTaskMidStatus(c, taskID, []int64{4})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(task, convey.ShouldNotBeNil)
			})
		})
		ctx.Convey("no mid ", func(ctx convey.C) {
			task, err := d.GetActivityTaskMidStatus(c, 1, []int64{})
			ctx.Convey("Then err should be nil.res should not be nil.", func(ctx convey.C) {
				ctx.So(err, convey.ShouldBeNil)
				ctx.So(task, convey.ShouldNotBeNil)
				ctx.So(len(task), convey.ShouldEqual, 0)
			})
		})
	})
}
