package service

import (
	"encoding/json"
	"testing"
	xtime "time"

	"go-gateway/app/web-svr/esports/admin/component"
	"go-gateway/app/web-svr/esports/admin/dao"
	"go-gateway/app/web-svr/esports/admin/model"
)

var (
	migrationDao *dao.Dao
)

func TestMigrationBiz(t *testing.T) {
	migrationDao = new(dao.Dao)
	{
		migrationDao.DB = component.GlobalDB
	}

	list := make([]*model.Team, 0)
	if err := migrationDao.DB.Find(&list).Error; err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(list)
	t.Log(string(bs))

	teamIDList := []int64{1255, 1256}
	teamList := make([]*model.Team, 0)
	if err := migrationDao.DB.Where("id in (?)", teamIDList).Find(&teamList).Error; err != nil {
		t.Error(err)

		return
	}

	bs1, _ := json.Marshal(teamList)
	t.Log(string(bs1))

	rs := make([]*model.Contest, 0)
	if err := migrationDao.DB.Model(&model.Contest{}).Where("guess_type=1 AND stime >= ?", xtime.Now().Add(10*xtime.Minute).Unix()).Find(&rs).Error; err != nil {
		t.Error(err)

		return
	}

	bs2, _ := json.Marshal(rs)
	t.Log(string(bs2))
}
