package rewards

import (
	"context"
	"testing"
	xtime "time"

	"go-common/library/database/sql"
	"go-common/library/time"
)

func TestNewCdKeyBiz(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = time.Duration(10 * xtime.Second)
		cfg.ExecTimeout = time.Duration(10 * xtime.Second)
		cfg.TranTimeout = time.Duration(10 * xtime.Second)
	}

	newDao := new(Dao)
	{
		newDao.db = sql.NewMySQL(cfg)
	}

	if err := newDao.db.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	d1, err := newDao.SendCdKey(context.Background(), 66666666, 666, "cdk_name_1", "1")
	if err != nil {
		t.Error("SendCdKey 1 err: ", err)

		return
	}

	t.Log("SendCdKey 1 cdkey: ", d1)

	d2, err := newDao.SendCdKey(context.Background(), 66666666, 666, "cdk_name_1", "1")
	if d1 != d2 {
		t.Error("d1 and  d2 should equal")

		return
	}

	t.Log("SendCdKey 2 cdkey: ", d2)

	d3, err := newDao.SendCdKey(context.Background(), 888888888, 666, "cdk_name_1", "1")
	if err != nil {
		t.Error("SendCdKey 3 err: ", err)

		return
	}

	t.Log("SendCdKey 3 cdkey: ", d3)
}
