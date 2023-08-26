package task

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/interface/model/task"
	taskmdl "go-gateway/app/web-svr/activity/interface/model/task"

	"github.com/smartystreets/goconvey/convey"
)

func TestTasktaskRoundKey(t *testing.T) {
	convey.Convey("taskRoundKey", t, func(convCtx convey.C) {
		var (
			taskID = int64(0)
			round  = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := taskRoundKey(taskID, round)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.Print(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTaskuserTaskStateKey(t *testing.T) {
	convey.Convey("userTaskStateKey", t, func(convCtx convey.C) {
		var (
			mid        = int64(0)
			businessID = int64(0)
			foreignID  = int64(0)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			p1 := userTaskStateKey(mid, businessID, foreignID)
			convCtx.Convey("Then p1 should not be nil.", func(convCtx convey.C) {
				convCtx.Print(p1, convey.ShouldNotBeNil)
			})
		})
	})
}

func TestTaskCacheUserTaskState(t *testing.T) {
	convey.Convey("CacheUserTaskState", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			t1         = &task.Task{ID: 24, BusinessID: 1, ForeignID: 10554, Attribute: 21, CycleDuration: 0, Stime: 1570636800, Etime: 1575043200}
			t2         = &task.Task{ID: 25, BusinessID: 1, ForeignID: 10554, Attribute: 31, CycleDuration: 86400, Stime: 1570636800, Etime: 1575043200}
			tasks      = make(map[int64]*task.Task)
			mid        = int64(27515321)
			businessID = int64(1)
			foreignID  = int64(10554)
		)
		tasks[25] = t2
		tasks[24] = t1
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheUserTaskState(c, tasks, mid, businessID, foreignID)
			convCtx.Convey("Then err should be nil.list should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Println(res)
			})
		})
	})
}

func TestTaskAddCacheUserTaskState(t *testing.T) {
	convey.Convey("AddCacheUserTaskState", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			b          = &taskmdl.UserTask{ID: 1}
			missData   = map[string]*taskmdl.UserTask{"4_0": b}
			mid        = int64(27515246)
			businessID = int64(1)
			foreignID  = int64(10401)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheUserTaskState(c, missData, mid, businessID, foreignID)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestTaskSetCacheUserTaskState(t *testing.T) {
	convey.Convey("SetCacheUserTaskState", t, func(convCtx convey.C) {
		var (
			c          = context.Background()
			data       = &taskmdl.UserTask{}
			mid        = int64(27515246)
			businessID = int64(1)
			foreignID  = int64(10401)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.SetCacheUserTaskState(c, data, mid, businessID, foreignID)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetActityTaskCount(t *testing.T) {
	convey.Convey("GetActityTaskCount", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(56)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetActityTaskCount(c, id)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestGetActivityTaskMidStatus(t *testing.T) {
	convey.Convey("GetActivityTaskMidStatus", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(56)
			mid = int64(2)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			_, err := d.GetActivityTaskMidStatus(c, id, mid)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
