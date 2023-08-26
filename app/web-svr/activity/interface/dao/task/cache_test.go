package task

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/model/task"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestTasktaskKey(t *testing.T) {
	convey.Convey("taskKey", t, func(convCtx convey.C) {
		var (
			id = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := taskKey(id)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.Print(p1)
			})
		})
	})
}

func TestTasktaskIDsKey(t *testing.T) {
	convey.Convey("taskIDsKey", t, func(convCtx convey.C) {
		var (
			businessID = int64(0)
			foreignID  = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := taskIDsKey(businessID, foreignID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.Print(p1)
			})
		})
	})
}

func TestTasktaskRuleKey(t *testing.T) {
	convey.Convey("taskRuleKey", t, func(convCtx convey.C) {
		var (
			taskID = int64(3)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := taskRuleKey(taskID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.Print(p1)
			})
		})
	})
}

func TestTaskuserTaskFinKey(t *testing.T) {
	convey.Convey("userTaskFinKey", t, func(convCtx convey.C) {
		var (
			mid    = int64(0)
			taskID = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := userTaskFinKey(mid, taskID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.Print(p1)
			})
		})
	})
}

func TestTaskTaskRule(t *testing.T) {
	convey.Convey("TaskRule", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			taskID = int64(167)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.TaskRule(c, taskID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestTaskUserTaskState(t *testing.T) {
	convey.Convey("UserTaskState", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			t          = &task.Task{ID: 167, BusinessID: 1, ForeignID: 10535, Attribute: 7}
			tasks      = make(map[int64]*task.Task)
			mid        = int64(27515246)
			businessID = int64(1)
			foreignID  = int64(10401)
			nowTs      = int64(1565078400)
		)
		tasks[1] = t
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.UserTaskState(c, tasks, mid, businessID, foreignID, nowTs)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
