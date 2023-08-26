package task

import (
	"context"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestTaskRawTaskIDs(t *testing.T) {
	convey.Convey("RawTaskIDs", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			businessID = int64(1)
			foreignID  = int64(10401)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			ids, err := d.RawTaskIDs(c, businessID, foreignID)
			convCtx.Convey("Then err should be nil.ids should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(ids, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTaskRawTasks(t *testing.T) {
	convey.Convey("RawTasks", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{1}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawTasks(c, ids)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTaskRawTask(t *testing.T) {
	convey.Convey("RawTask", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			no, err := d.RawTask(c, id)
			convCtx.Convey("Then err should be nil.no should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(no, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTaskRawUserTaskState(t *testing.T) {
	convey.Convey("RawUserTaskState", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			taskIDs    = []int64{1}
			mid        = int64(88895359)
			businessID = int64(1)
			foreignID  = int64(10401)
			nowTs      = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawUserTaskState(c, taskIDs, mid, businessID, foreignID, nowTs)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTaskRawUserTaskLog(t *testing.T) {
	convey.Convey("RawUserTaskLog", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			mid    = int64(88895359)
			taskID = int64(3)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			list, err := d.RawUserTaskLog(c, mid, taskID, 1)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(list, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTaskRawTaskRule(t *testing.T) {
	convey.Convey("RawTaskRule", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			taskID = int64(4)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.RawTaskRule(c, taskID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTaskAddUserTaskState(t *testing.T) {
	convey.Convey("AddUserTaskState", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			mid        = int64(88895359)
			businessID = int64(1)
			taskID     = int64(1)
			foreignID  = int64(10401)
			round      = int64(0)
			count      = int64(0)
			finish     = int64(0)
			award      = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddUserTaskState(c, mid, businessID, taskID, foreignID, round, count, finish, award)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestTaskUpUserTaskState(t *testing.T) {
	convey.Convey("UpUserTaskState", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			mid    = int64(88895359)
			taskID = int64(1)
			round  = int64(0)
			count  = int64(0)
			finish = int64(0)
			award  = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.UpUserTaskState(c, mid, taskID, round, count, finish, award, 1)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestTaskUserTaskAward(t *testing.T) {
	convey.Convey("UserTaskAward", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			mid    = int64(88895359)
			taskID = int64(1)
			round  = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.UserTaskAward(c, mid, taskID, round, 1, 1)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestTaskAddUserTaskLog(t *testing.T) {
	convey.Convey("AddUserTaskLog", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			mid        = int64(88895359)
			businessID = int64(1)
			taskID     = int64(1)
			foreignID  = int64(10401)
			round      = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddUserTaskLog(c, mid, businessID, taskID, foreignID, round)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
