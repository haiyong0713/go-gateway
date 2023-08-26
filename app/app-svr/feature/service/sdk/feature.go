package sdk

import (
	"os"
	"strconv"
	"time"

	"go-common/library/net/rpc/warden"
	xtime "go-common/library/time"

	featuregrpc "go-gateway/app/app-svr/feature/service/api"

	"github.com/robfig/cron"
)

type Feature struct {
	TreeID              int64
	featureClient       featuregrpc.FeatureClient
	cron                *cron.Cron
	buildLimitCache     map[string]map[string][]*featuregrpc.BuildLimitConditions
	businessConfigCache map[string]string
	abTestCache         map[string]*featuregrpc.ABTestItem
}

type Config struct {
	TreeID string
}

type OriginResutl struct {
	MobiApp    string
	Device     string
	Build      int64
	BuildLimit bool
}

func New(c *Config) (f *Feature) {
	if c == nil {
		c = &Config{
			TreeID: os.Getenv("TREE_ID"),
		}
	}
	f = new(Feature)
	var err error
	if f.TreeID, err = strconv.ParseInt(c.TreeID, 10, 64); err != nil {
		panic(err)
	}
	if f.featureClient, err = featuregrpc.NewClient(&warden.ClientConfig{Timeout: xtime.Duration(time.Millisecond * 1000)}); err != nil {
		panic(err)
	}
	f.buildLimitCache = make(map[string]map[string][]*featuregrpc.BuildLimitConditions)
	f.businessConfigCache = make(map[string]string)
	f.abTestCache = make(map[string]*featuregrpc.ABTestItem)
	// 初始化缓存
	f.initCache()
	// 定时任务
	f.initCron()
	return f
}

// 获取对应配置
func (f *Feature) initCache() {
	var err error
	if err = f.buildLimitConfig(); err != nil {
		panic(err)
	}
	if err = f.loadBusinessConfig(); err != nil {
		panic(err)
	}
	if err = f.abtestConfig(); err != nil {
		panic(err)
	}
}

func (f *Feature) initCron() {
	f.cron = cron.New()
	// nolint:errcheck
	if err := f.cron.AddFunc("*/10 * * * * *", func() { f.buildLimitConfig() }); err != nil { // load Zone Idx & types
		panic(err)
	}
	// nolint:errcheck
	if err := f.cron.AddFunc("*/10 * * * * *", func() { f.loadBusinessConfig() }); err != nil { // load Zone Idx & types
		panic(err)
	}
	// nolint:errcheck
	if err := f.cron.AddFunc("*/10 * * * * *", func() { f.abtestConfig() }); err != nil { // load Zone Idx & types
		panic(err)
	}
	f.cron.Start()
}
