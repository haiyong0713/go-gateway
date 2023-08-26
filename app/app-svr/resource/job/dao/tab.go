package dao

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/resource/job/model/show"
	"time"
)

func (d *Dao) GetTabExts(time time.Time) (list []*show.TabExt, err error) {
	err = d.showDB.Model(show.TabExt{}).Where("`stime`<=? AND `etime`>=? AND `state`= 1", time, time).Find(&list).Error
	if err != nil {
		if err == ecode.NothingFound {
			return list, nil
		}
		log.Error("Dao:GetTabExts() error:%+v", err)
	}
	return list, err
}

func (d *Dao) GetTabLimits(tIDs []int64, tType int) (list map[int64][]*show.TabLimit, err error) {
	tabLimits := []show.TabLimit{}
	err = d.showDB.Model(show.TabLimit{}).Where("`t_id` in (?) and `type` = ? and `state` = 1", tIDs, tType).Find(&tabLimits).Error
	if err != nil {
		if err == ecode.NothingFound {
			return list, nil
		}
		log.Error("Dao:GetTabLimits() error:%+v", err)
		return list, err
	}
	if len(tabLimits) > 0 {
		list = make(map[int64][]*show.TabLimit)
		for _, tabLimit := range tabLimits {
			limit := tabLimit
			list[tabLimit.TID] = append(list[tabLimit.TID], &limit)
		}
	}
	return list, nil
}

func (d *Dao) SetTabExt2Cache(ctx context.Context, key string, tabExt map[string]*show.MenuExt) (err error) {
	conn := d.redisShow.Get(ctx)
	defer conn.Close()
	tabExtJson, err := json.Marshal(tabExt)
	if err != nil {
		log.Error("Dao:SetTabExt2Cache() json marshal fail:%+v,data:%+v", err, tabExt)
		return err
	}
	_, err = redis.String(conn.Do("SET", key, tabExtJson))
	if err != nil {
		log.Error("Dao:SetTabExt2Cache() set cache fail:%+v,key:%+v,data:%+v", err, key, tabExtJson)
		return err
	}
	return err
}
