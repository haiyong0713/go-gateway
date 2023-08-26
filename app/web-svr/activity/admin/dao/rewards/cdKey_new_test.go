package rewards

import (
	"context"
	"testing"
	xtime "time"

	"go-common/library/database/sql"
	"go-common/library/time"
)

// go test -v -count=1 cdKey_new_test.go cdkey.go dao.go
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

	keys := make([]string, 0)
	{
		keys = append(keys, "1")
		keys = append(keys, "2")
		keys = append(keys, "3")
		keys = append(keys, "4")
		keys = append(keys, "5")
		keys = append(keys, "6")
		keys = append(keys, "7")
		keys = append(keys, "8")
	}
	if err := newDao.UploadCdKey(context.Background(), 1000, "cdk_name_1", keys); err != nil {
		t.Error("cdk_name_1: ", err)

		return
	}

	if err := newDao.UploadCdKey(context.Background(), 1000, "cdk_name_2", keys); err != nil {
		t.Error("cdk_name_2: ", err)

		return
	}
}
