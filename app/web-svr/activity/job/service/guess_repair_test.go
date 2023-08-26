package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	xtime "time"

	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/sql"
	"go-common/library/time"

	"go-gateway/app/web-svr/activity/job/conf"
	"go-gateway/app/web-svr/activity/job/dao/guess"
)

func TestRepair(t *testing.T) {
	fmt.Println(decimal(float64(1)*1.52, _decimalIn))

	newCfg := &redis.Config{
		Name:  "local",
		Proto: "tcp",
		Addr:  "127.0.0.1:6379",
		Config: &pool.Config{
			IdleTimeout: time.Duration(10 * xtime.Second),
			Idle:        2,
			Active:      8,
		},
		WriteTimeout: time.Duration(10 * xtime.Second),
		ReadTimeout:  time.Duration(10 * xtime.Second),
		DialTimeout:  time.Duration(10 * xtime.Second),
	}

	sqlCfg := new(sql.Config)
	{
		sqlCfg.Addr = "127.0.0.1:3306"
		sqlCfg.DSN = "root:root@tcp(127.0.0.1:3306)/browser?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4"
		sqlCfg.QueryTimeout = time.Duration(10 * xtime.Second)
		sqlCfg.ExecTimeout = time.Duration(10 * xtime.Second)
		sqlCfg.TranTimeout = time.Duration(10 * xtime.Second)
		sqlCfg.Active = 8
		sqlCfg.Idle = 2
	}

	cfg := new(conf.Config)
	{
		tmpMysql := new(conf.MySQL)
		tmpMysql.Like = sqlCfg
		cfg.MySQL = tmpMysql

		cfg.Redis = new(conf.Redis)
		cfg.Redis.Config = newCfg
	}

	tmpService := new(Service)
	tmpService.guessDao = guess.New(cfg)

	repair := new(CompensationRepair4MID)
	{
		repair.MID = 608121144
		repair.S10ContestIDList = []int64{7017, 7440, 7441, 7442, 7443, 7444, 7445, 7446, 7447, 7448, 7449, 7450, 7451, 7452, 7453, 7454, 7567, 7568, 7455, 7456, 7457, 7458, 7582, 7581, 7459, 7460, 7461, 7462, 7463, 7464, 7465, 7466, 7467, 7468, 7469, 7470, 7471, 7472, 7473, 7474, 7475, 7476, 7477, 7478, 7479, 7480, 7481, 7482, 7483, 7484, 7485, 7486, 7487, 7488, 7489, 7490, 7491, 7492, 7493, 7494, 7495, 7496, 7497, 7498, 7499, 7500, 7501, 7502, 7503, 7504, 7505, 7506, 7507, 7508, 7509, 7510, 7511, 7512, 7513, 7514, 7515, 7516, 7517}
	}

	bs1, _ := json.Marshal(repair)
	fmt.Println(string(bs1))

	m := tmpService.CompensationRepair4MID(context.Background(), repair)
	bs, _ := json.Marshal(m)
	fmt.Println(string(bs))

	overIssueRepair := new(OverIssueRepair)
	{
		overIssueRepair.Path = "/Users/leelei/Downloads/repair_user_guess_status_test.csv"
		overIssueRepair.OpDB = true
	}
	err := tmpService.RepairSpecifiedUserGuess(overIssueRepair)
	fmt.Println(err)
}
