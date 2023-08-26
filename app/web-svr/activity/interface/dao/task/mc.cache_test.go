package task

import (
	"context"
	"go-gateway/app/web-svr/activity/interface/model/task"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestTaskCacheTask(t *testing.T) {
	convey.Convey("CacheTask", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(1)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheTask(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestTaskAddCacheTask(t *testing.T) {
	convey.Convey("AddCacheTask", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(1)
			val = &task.Task{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheTask(c, id, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestTaskDelCacheTask(t *testing.T) {
	convey.Convey("DelCacheTask", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(10555)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.DelCacheTask(c, id)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestTaskCacheTasks(t *testing.T) {
	convey.Convey("CacheTasks", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			ids = []int64{166, 167, 168}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheTasks(c, ids)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestTaskAddCacheTasks(t *testing.T) {
	convey.Convey("AddCacheTasks", t, func(convCtx convey.C) {
		var (
			c      = context.Background()
			values = make(map[int64]*task.Task)
			t      = &task.Task{ID: 1, BusinessID: 1}
		)
		values[1] = t
		values[2] = t
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheTasks(c, values)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestTaskCacheTaskIDs(t *testing.T) {
	convey.Convey("CacheTaskIDs", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			id        = int64(3)
			foreignID = int64(10401)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheTaskIDs(c, id, foreignID)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestTaskAddCacheTaskIDs(t *testing.T) {
	convey.Convey("AddCacheTaskIDs", t, func(convCtx convey.C) {
		var (
			c         = context.Background()
			id        = int64(1)
			val       = []int64{1, 2}
			foreignID = int64(10401)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheTaskIDs(c, id, val, foreignID)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}

func TestTaskCacheTaskRule(t *testing.T) {
	convey.Convey("CacheTaskRule", t, func(convCtx convey.C) {
		var (
			c  = context.Background()
			id = int64(4)
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			res, err := d.CacheTaskRule(c, id)
			convCtx.Convey("Then err should be nil.res should not be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
				convCtx.Print(res)
			})
		})
	})
}

func TestTaskAddCacheTaskRule(t *testing.T) {
	convey.Convey("AddCacheTaskRule", t, func(convCtx convey.C) {
		var (
			c   = context.Background()
			id  = int64(4)
			val = &task.TaskRule{}
		)
		convCtx.Convey("When everything goes positive", func(convCtx convey.C) {
			err := d.AddCacheTaskRule(c, id, val)
			convCtx.Convey("Then err should be nil.", func(convCtx convey.C) {
				convCtx.So(err, convey.ShouldBeNil)
			})
		})
	})
}
