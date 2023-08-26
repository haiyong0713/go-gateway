package service

import (
	"testing"
	xtime "time"

	"go-common/library/database/orm"
	"go-common/library/time"

	"go-gateway/app/web-svr/esports/admin/component"
	"go-gateway/app/web-svr/esports/admin/conf"
	"go-gateway/app/web-svr/esports/admin/model"
)

func TestAutoSubscribeBiz(t *testing.T) {
	cfg := &orm.Config{
		DSN:         "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4",
		Active:      8,
		Idle:        2,
		IdleTimeout: time.Duration(10 * xtime.Second),
	}
	conf := &conf.Config{
		ORM: cfg,
	}
	component.InitRelations(conf)
	if err := component.GlobalDB.DB().Ping(); err != nil {
		t.Error(err)
		return
	}

	t.Run("test create auto sub table", testAutoSubscribe)
}

func testAutoSubscribe(t *testing.T) {
	tx := component.GlobalDB.Begin()

	// record season in auto subscribe season list
	autoSubSeason := model.AutoSubscribeSeason{
		SeasonID: 666,
	}
	err := tx.Create(&autoSubSeason).Error
	if err != nil {
		t.Error(err)
		if err = tx.Rollback().Error; err != nil {
			t.Error(err)
		}

		return
	}

	// create sub auto_subscribe detail table
	err = tx.Exec(model.GenAutoSubSeasonDetailSql(666)).Error
	if err != nil {
		t.Error(err, model.GenAutoSubSeasonDetailSql(666))
		if err = tx.Rollback().Error; err != nil {
			t.Error(err)
		}

		return
	}

	if err = tx.Commit().Error; err != nil {
		t.Error(err)
	}
}
