package dao

import (
	"context"
	"encoding/json"
	"testing"
	xtime "time"

	"go-common/library/database/sql"
	"go-common/library/time"

	xsql "go-gateway/app/web-svr/esports/job/sql"
)

// go test -v -count=1 v2_team_test.go v2_team.go v2_season.go
func TestVTeamBiz(t *testing.T) {
	dbCfg := new(sql.Config)
	{
		dbCfg.Addr = "127.0.0.1:3306"
		dbCfg.DSN = "root:root@tcp(127.0.0.1:3306)/esport?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		dbCfg.QueryTimeout = time.Duration(10 * xtime.Second)
		dbCfg.ExecTimeout = time.Duration(10 * xtime.Second)
		dbCfg.TranTimeout = time.Duration(10 * xtime.Second)
	}

	if err := xsql.InitByCfg(); err != nil {
		t.Error(err)

		return
	}

	m, err := FetchTeamInfoByLargeIDList(context.Background(), []int64{1255, 1256})
	if err != nil {
		t.Error(err)

		return
	}

	bs, _ := json.Marshal(m)
	t.Log(string(bs))
}
