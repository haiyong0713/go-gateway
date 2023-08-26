package task

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/admin/model/task"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTaskSaveTask(t *testing.T) {
	Convey("SaveTask", t, func() {
		var (
			c      = context.Background()
			addArg = task.AddArg{Name: "test", BusinessID: 1, ForeignID: 1001, Level: 1, PreTask: "3"}
			arg    = &task.SaveArg{ID: 2, AddArg: addArg}

			preData = &task.Item{Task: &task.Task{ID: 5}, Rule: &task.Rule{TaskID: 5, PreTask: "6"}}
		)
		Convey("When everything goes positive", func() {
			err := d.SaveTask(c, arg, preData)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestTaskAddAward(t *testing.T) {
	Convey("AddAward", t, func() {
		var (
			c      = context.Background()
			taskID = int64(1)
			mid    = int64(1)
			award  = int64(0)
		)
		Convey("When everything goes positive", func() {
			err := d.AddAward(c, taskID, mid, award)
			Convey("Then err should be nil.", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
