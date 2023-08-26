package dao

import (
	"context"
	"fmt"
	"testing"
	xtime "time"

	"go-common/library/database/sql"
	"go-common/library/time"
)

var (
	contestDao = new(Dao)
)

// go test -v --count=1 contest.go contest_test.go dao.go
func TestContestBiz(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = time.Duration(10 * xtime.Second)
		cfg.ExecTimeout = time.Duration(10 * xtime.Second)
		cfg.TranTimeout = time.Duration(10 * xtime.Second)
	}

	contestDao.db = sql.NewMySQL(cfg)
	if err := contestDao.db.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("max contestId biz", maxContestID)
}

func maxContestID(t *testing.T) {
	d, err := contestDao.FetchMaxContestID(context.Background())
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println("max contest id: ", d)
}
