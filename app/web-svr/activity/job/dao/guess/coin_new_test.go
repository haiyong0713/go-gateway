package guess

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"go-common/library/database/sql"
	xtime "go-common/library/time"
)

var (
	newCoinDao = new(Dao)
)

// go test -v --count=1  coin_new.go coin_new_test.go coin_redis.go dao.go guess.go im_msg.go memcache.go  mysql.go redis.go
func TestNewCoinBiz(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = xtime.Duration(10 * time.Second)
		cfg.ExecTimeout = xtime.Duration(10 * time.Second)
		cfg.TranTimeout = xtime.Duration(10 * time.Second)
	}

	newCoinDao.db = sql.NewMySQL(cfg)
	if err := newCoinDao.db.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("test update user guess relations", UpdateUserGuessRelations)
	t.Run("all un_cleared guess list", allUnClearedGuessListByMID)
	t.Run("repair use guess log", RepairUserGuessLogs)
	t.Run("fetch user guess stats", FetchUserGuessStats)
	t.Run("update user guess stats", UpdateUserGuessStats)
}

func UpdateUserGuessStats(t *testing.T) {
	guess, success, income, err := newCoinDao.FetchUserGuessStats(context.Background(), 88888888)
	if err != nil {
		t.Error(err)

		return
	}

	err = newCoinDao.UpdateUserGuessStats(context.Background(), 88888888, guess, success, income, 1)
	if err != nil {
		t.Error(err)
	}
}

func FetchUserGuessStats(t *testing.T) {
	guess, success, income, err := newCoinDao.FetchUserGuessStats(context.Background(), 88888888)
	if err != nil {
		t.Error(err)

		return
	}

	fmt.Println(guess, success, income)
}

func RepairUserGuessLogs(t *testing.T) {
	m := make(map[int64]float64, 0)
	{
		m[80881] = 8.8
	}
	if err := newCoinDao.RepairUserGuessLogs(context.Background(), 88888888, m); err != nil {
		t.Error(err)
	}
}

func allUnClearedGuessListByMID(t *testing.T) {
	d, err := newCoinDao.AllUnClearedGuessListByMid(context.Background(), 88888888)
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(d)
	fmt.Println(string(bs))
}

func UpdateUserGuessRelations(t *testing.T) {
	if err := newCoinDao.UpdateUserGuessRelations(context.Background(), 88888888, 1, 80881, 2.3); err != nil {
		t.Error(err)
	}
}
