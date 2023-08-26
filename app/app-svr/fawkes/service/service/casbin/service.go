package casbin

import (
	"fmt"
	"sync"

	"go-common/library/database/orm"

	"go-gateway/app/app-svr/fawkes/service/conf"
	openmdl "go-gateway/app/app-svr/fawkes/service/model/open"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	casbinV2 "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v2"
	"github.com/jinzhu/gorm"
	"github.com/robfig/cron"
)

const casbinModel = `
[request_definition]
r = user_token, path, app_key

[policy_definition]
p = user_token, path, app_key

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.user_token == p.user_token && keyMatch(r.path, p.path) && appKeyMatch(r.app_key, p.app_key)
`

var (
	syncedEnforcer *casbinV2.SyncedEnforcer
	casbinDB       *gorm.DB
	once           sync.Once
)

func GetInstance() *casbinV2.SyncedEnforcer {
	once.Do(func() {
		syncedEnforcer = initCasbinEnforcer()
		reloadPolicy()
	})
	return syncedEnforcer
}

func initCasbinEnforcer() *casbinV2.SyncedEnforcer {
	casbinDB = orm.NewMySQL(conf.Conf.ORM)
	adapter, err := gormadapter.NewAdapterByDBUsePrefix(casbinDB, "auth_api_")
	if err != nil {
		panic(fmt.Errorf("new casbin adapter fail, err=%s", err))
	}
	m, err := model.NewModelFromString(casbinModel)
	if err != nil {
		panic(fmt.Errorf("new casbin model fail, err+%s", err))
	}
	syncedEnforcer, err = casbinV2.NewSyncedEnforcer(m, adapter)
	if err != nil {
		panic(fmt.Errorf("new casbin enforcer fail, err+%s", err))
	}
	syncedEnforcer.AddFunction("appKeyMatch", AppKeyMathFunc)
	/*	redisConf := conf.Conf.Redis.Fawkes
			w, _ := watcher.NewWatcher(redisConf.Addr, watcher.WatcherOptions{
				Channel:    "/casbin",
				IgnoreSelf: true,
			})
			err = syncedEnforcer.SetWatcher(w)
		if err != nil {
			panic(fmt.Errorf("casbin set watcher error err+%s", err))
		}*/
	return syncedEnforcer
}

func AppKeyMathFunc(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return AppKeyMath(name1, name2), nil
}

func AppKeyMath(key1 string, key2 string) bool {
	return key2 == openmdl.AnyAppKey || key1 == key2
}

func reloadPolicy() {
	c := cron.New()
	_ = c.AddFunc("@every 30s", func() {
		if syncedEnforcer != nil {
			err := syncedEnforcer.LoadPolicy()
			if err != nil {
				log.Error("%s", "casbin 策略刷新失败")
				return
			}
		}
	})
	c.Start()
}
