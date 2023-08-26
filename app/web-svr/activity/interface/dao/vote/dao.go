package vote

import (
	"go-common/library/cache/redis"
	"go-common/library/database/elastic"
	"go-common/library/database/sql"
	"go-gateway/app/web-svr/activity/interface/component"
	"go-gateway/app/web-svr/activity/interface/conf"
	"time"
)

const (
	cacheVersionFormat = "0102150405"
)

type Dao struct {
	redis                                  *redis.Redis
	db                                     *sql.DB
	esClient                               *elastic.Elastic
	datasourceMap                          map[string]DataSource
	dataSourceItemsInfoCacheExpire         int64
	outdatedDataSourceItemsInfoCacheExpire int64
	voteRankZsetExpire                     int64
	realTimeVoteRankWithInfoExpire         int64
	manualVoteRankWithInfoExpire           int64
	onTimeVoteRankWithInfoExpire           int64
	adminVoteRankWithInfoExpire            int64
	activityCacheExpire                    int64
	blackListCacheExpire                   int64
	userVoteCountExpire                    int64
	outdatedVoteRankWithInfoExpire         int64
}

func New(c *conf.Config, dataSourceTypes map[string]DataSource) *Dao {
	d := &Dao{}
	d.redis = component.GlobalVoteRedis
	d.db = component.GlobalRewardsDB
	d.datasourceMap = make(map[string]DataSource)
	d.esClient = component.EsClient
	for typ, ds := range dataSourceTypes {
		tmpTyp := typ
		tmpDs := ds
		d.datasourceMap[tmpTyp] = tmpDs
	}
	//底层稿件信息缓存(正在进行中的活动): 每5min更新一次, 旧数据不会被使用.保留30min
	d.dataSourceItemsInfoCacheExpire = int64(time.Duration(c.Vote.DataSourceItemsInfoCacheExpire) / time.Second)

	//底层稿件信息缓存(结束90天内的活动): 每天更新一次, 旧数据不会被使用.保留2天
	d.outdatedDataSourceItemsInfoCacheExpire = int64(time.Duration(c.Vote.OutdatedDataSourceItemsInfoCacheExpire) / time.Second)

	//稿件票数zset: 实时更新,有回源逻辑, 保留90天
	d.voteRankZsetExpire = int64(time.Duration(c.Vote.VoteRankZsetExpire) / time.Second)

	//稿件投票排行榜(实时更新): 每30s刷新一次(version+1), 旧版本保留一小时
	d.realTimeVoteRankWithInfoExpire = int64(time.Duration(c.Vote.RealTimeVoteRankWithInfoExpire) / time.Second)

	//稿件投票排行榜(手动更新): 不自动更新, 旧版本保留15天
	d.manualVoteRankWithInfoExpire = int64(time.Duration(c.Vote.ManualVoteRankWithInfoExpire) / time.Second)

	//稿件投票排行榜(定时更新): 每天刷新一次, 旧版本保留7天
	d.onTimeVoteRankWithInfoExpire = int64(time.Duration(c.Vote.OnTimeVoteRankWithInfoExpire) / time.Second)

	//稿件投票排行榜(后台使用): 每5min刷新一次, 旧版本保留1小时
	d.adminVoteRankWithInfoExpire = int64(time.Duration(c.Vote.AdminVoteRankWithInfoExpire) / time.Second)

	//稿件投票排行榜(已结束90天内的活动的排行榜): 每天刷新一次, 旧版本保留7天
	d.outdatedVoteRankWithInfoExpire = int64(time.Duration(c.Vote.OutdatedVoteRankWithInfoExpire) / time.Second)

	//活动缓存: 有回源逻辑,保留3s
	d.activityCacheExpire = int64(time.Duration(c.Vote.ActivityCacheExpire) / time.Second)

	//黑名单缓存: 有回源逻辑, CURD时会被删除, 保留1天
	d.blackListCacheExpire = int64(time.Duration(c.Vote.BlackListCacheExpire) / time.Second)

	//用户投票数缓存: 有回源逻辑,保留3分钟
	d.userVoteCountExpire = int64(time.Duration(c.Vote.UserVoteCountExpire) / time.Second)
	return d
}
