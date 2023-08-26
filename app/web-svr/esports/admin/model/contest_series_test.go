package model

import (
	"testing"
	"time"

	"go-common/library/database/orm"
	"go-gateway/app/web-svr/esports/admin/component"
	"go-gateway/app/web-svr/esports/admin/conf"
)

func initDB() {
	cfg := new(orm.Config)
	{
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.Active = 5
		cfg.Idle = 2
		cfg.IdleTimeout = 1
	}
	conf := &conf.Config{
		ORM: cfg,
	}
	component.InitRelations(conf)

}

func TestContestSeriesBiz(t *testing.T) {
	initDB()
	t.Run("test insert biz", insertBiz)
	t.Run("test update biz", updateBiz)
	t.Run("test delete biz", deleteBiz)
	t.Run("test list biz", listBiz)
	t.Run("test count biz", countBiz)
}

func insertBiz(t *testing.T) {
	series := new(ContestSeries)
	{
		series.StartTime = time.Now().Unix()
		series.EndTime = time.Now().Unix() + 1000
		series.ParentTitle = "parent_title"
		series.ChildTitle = "child_title"
		series.SeasonID = 1
	}

	if err := series.Insert(); err != nil {
		t.Errorf("insert series occur err: %v", err)
	}
}

func updateBiz(t *testing.T) {
	series := new(ContestSeries)
	{
		series.ID = 1
		series.StartTime = time.Now().Unix()
		series.EndTime = time.Now().Unix() + 1000
		series.ParentTitle = "parent_title_new"
		series.ChildTitle = "child_title_new"
		series.SeasonID = 2
	}

	if err := series.Update(); err != nil {
		t.Errorf("update series occur err: %v", err)
	}
}

func deleteBiz(t *testing.T) {
	series := new(ContestSeries)
	{
		series.ID = 1
		series.StartTime = time.Now().Unix()
		series.EndTime = time.Now().Unix() + 1000
		series.ParentTitle = "parent_title_new"
		series.ChildTitle = "child_title_new"
	}

	if err := series.Delete(); err != nil {
		t.Errorf("delete series occur err: %v", err)
	}

	newOne, err := FindContestSeriesByID(1)
	if err != nil || newOne.ID != 1 {
		t.Errorf("delete series occur err: %v, data: %v", err, newOne)
	}

	_ = series.MarkAsNotDeleted()
}

func listBiz(t *testing.T) {
	_, err := ContestSeriesList(2, 10, 0)
	if err != nil {
		t.Errorf("fetch series list occur err: %v", err)
	}
}

func countBiz(t *testing.T) {
	count := ContestSeriesCount(2)
	if count != 1 {
		t.Errorf("contest series count is not %v", count)
	}
}
