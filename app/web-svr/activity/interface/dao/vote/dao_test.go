package vote

import (
	"context"
	"go-common/library/cache/redis"
	"go-common/library/container/pool"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	"go-common/library/log"
	xtime "go-common/library/time"
	"time"
)

var testDao *Dao

func init() {
	testDao = &Dao{}
	log.Init(&log.Config{
		Stdout: true,
	})
	testDao.redis = redis.NewRedis(&redis.Config{
		Config: &pool.Config{
			Active:      10,
			Idle:        5,
			IdleTimeout: xtime.Duration(time.Second),
			WaitTimeout: xtime.Duration(time.Second),
			Wait:        false,
		},
		Proto:        "tcp",
		Addr:         "127.0.0.1:6379",
		DialTimeout:  xtime.Duration(time.Second),
		ReadTimeout:  xtime.Duration(time.Second),
		WriteTimeout: xtime.Duration(time.Second),
	})
	testDao.db = sql.NewMySQL(&sql.Config{
		Idle:         5,
		Active:       10,
		Addr:         "127.0.0.1:3306",
		DSN:          "test:test@tcp(127.0.0.1:3306)/bilibili_lottery?timeout=5s&readTimeout=5s&writeTimeout=5s&parseTime=true&loc=Local&charset=utf8,utf8mb4",
		IdleTimeout:  xtime.Duration(time.Second),
		QueryTimeout: xtime.Duration(time.Second),
		ExecTimeout:  xtime.Duration(time.Second),
		TranTimeout:  xtime.Duration(time.Second),
	})
	//testDao.esClient = elastic.NewElastic(&elastic.Config{
	//	Host: "http://127.0.0.1:9200",
	//	HTTPClient: &httpx.ClientConfig{
	//		App: &httpx.App{
	//			Key:    "3c4e41f926e51656",
	//			Secret: "26a2095b60c24154521d24ae62b885bb",
	//		},
	//		Dial:    xtime.Duration(time.Second),
	//		Timeout: xtime.Duration(time.Second),
	//	},
	//})
	testDao.esClient = elastic.NewElastic(nil)
	testDao.voteRankZsetExpire = 103600
	testDao.dataSourceItemsInfoCacheExpire = 103600
	testDao.outdatedDataSourceItemsInfoCacheExpire = 103600
	testDao.activityCacheExpire = 103600
	testDao.realTimeVoteRankWithInfoExpire = 103600
	testDao.manualVoteRankWithInfoExpire = 103600
	testDao.onTimeVoteRankWithInfoExpire = 103600
	testDao.blackListCacheExpire = 103600
	testDao.userVoteCountExpire = 103600
	testDao.outdatedVoteRankWithInfoExpire = 103600
	testDao.adminVoteRankWithInfoExpire = 103600
	testDao.datasourceMap = map[string]DataSource{}
	ctx := context.Background()
	_, _ = testDao.redis.Do(ctx, "FLUSHALL")
	_, _ = testDao.db.Exec(ctx, "TRUNCATE TABLE act_vote_main;")
	_, _ = testDao.db.Exec(ctx, "TRUNCATE TABLE act_vote_data_sources_group;")
	_, _ = testDao.db.Exec(ctx, "TRUNCATE TABLE act_vote_data_source_items;")
	_, _ = testDao.db.Exec(ctx, "TRUNCATE TABLE act_vote_user_action_00;")
	_, _ = testDao.db.Exec(ctx, "TRUNCATE TABLE act_vote_user_summary_00;")
}
