package show

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/service/model"
)

func (d *Dao) GetTabExtFromCache(ctx context.Context, key string) (tabExt map[string]*model.MenuExt, err error) {
	conn := d.redis.Conn(ctx)
	defer conn.Close()
	tabExtJson, err := redis.String(conn.Do("GET", key))
	if err != nil {
		if err != redis.ErrNil {
			log.Error("Dao:GetTabExtFromCache() get cache fail:%+v,key:%+v", err, key)
		}
		return
	}
	if tabExtJson == "" {
		return
	}
	err = json.Unmarshal([]byte(tabExtJson), &tabExt)
	//log.Warn("tabExt:%+v", tabExt)
	if err != nil {
		log.Error("Dao:GetTabExtFromCache() unmarshal fail:%+v,data:%+v", err, tabExtJson)
	}
	return
}
