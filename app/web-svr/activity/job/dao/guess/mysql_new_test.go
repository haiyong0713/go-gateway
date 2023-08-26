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
	dao = new(Dao)
)

// go test -v mysql_new_test.go dao.go im_msg.go guess.go mysql.go
func TestGuessMysql(t *testing.T) {
	cfg := new(sql.Config)
	{
		cfg.Addr = "127.0.0.1:3306"
		cfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		cfg.QueryTimeout = xtime.Duration(10 * time.Second)
		cfg.ExecTimeout = xtime.Duration(10 * time.Second)
		cfg.TranTimeout = xtime.Duration(10 * time.Second)
	}

	dao.db = sql.NewMySQL(cfg)
	if err := dao.db.Ping(context.Background()); err != nil {
		t.Error(err)

		return
	}

	t.Run("un cleared mid list by mainID", unClearedMIDListByMainID)
	t.Run("db query with no error", GuessFinishByLimit)
	t.Run("single table guess record", SingleTableGuessRecord)
	t.Run("ResetUserLog", ResetUserLog)
	t.Run("UpdateUserGuess", UpdateUserGuess)
}

func unClearedMIDListByMainID(t *testing.T) {
	list, err := dao.UnClearedMIDListByTableSuffixAndMainID(context.Background(), "88", 8888)
	if err != nil {
		t.Error(err)

		return
	}

	t.Log(list)
}

func UpdateUserGuess(t *testing.T) {
	if err := dao.UpdateUserGuess(context.Background(), 80881, 88, 100); err != nil {
		t.Error(err)
	}
}

func ResetUserLog(t *testing.T) {
	if err := dao.ResetUserLog(context.Background(), 88888888, 9, 100, 3, 1); err != nil {
		t.Error(err)
	}
}

func GuessFinishByLimit(t *testing.T) {
	if _, err := dao.GuessFinishByLimit(context.Background(), 88, 8888, 100); err != nil {
		t.Error(err)
	}
}

func SingleTableGuessRecord(t *testing.T) {
	d, err := dao.SingleTableGuessRecord(context.Background(), []int64{88}, 8888)
	if err != nil {
		t.Error(err)
	} else {
		bs, _ := json.Marshal(d)
		fmt.Println(string(bs))
	}
}
