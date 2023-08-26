package dao

import (
	"context"
	"encoding/json"
	"go-common/library/cache/redis"
	"go-common/library/ecode"
	"go-common/library/log"
	feedMenuModel "go-gateway/app/app-svr/app-feed/admin/model/menu"
	api "go-gateway/app/app-svr/resource/service/api/v1"
	"gorm.io/gorm"
	"time"
)

func (d *Dao) GetSkinExts(time time.Time) (list []*feedMenuModel.SkinExt, err error) {
	err = d.showDB.Model(feedMenuModel.SkinExt{}).Where("`stime`<=? AND `etime`>=? AND `state`= ?", time, time, api.SkinExtState_ONLINE).Find(&list).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return list, nil
		}
	}
	return list, err
}

func (d *Dao) GetSkinLimits(skinExtIDs []int64) (list map[int64][]*feedMenuModel.SkinLimit, err error) {
	skinLimits := make([]feedMenuModel.SkinLimit, 0)
	err = d.showDB.Model(feedMenuModel.SkinLimit{}).Where("`s_id` in (?) and `state` = ?", skinExtIDs, api.SkinLimitState_ONLINE).Find(&skinLimits).Error
	if err != nil {
		if err == ecode.NothingFound {
			return list, nil
		}
		log.Error("resource-job.Dao.GetSkinLimits Error (%+v)", err)
		return list, err
	}
	if len(skinLimits) > 0 {
		list = make(map[int64][]*feedMenuModel.SkinLimit)
		for _, skinLimit := range skinLimits {
			limit := skinLimit
			if _, ok := list[skinLimit.SID]; !ok {
				list[skinLimit.SID] = make([]*feedMenuModel.SkinLimit, 0)
			}
			list[skinLimit.SID] = append(list[skinLimit.SID], &limit)
		}
	}
	return list, nil
}

func (d *Dao) SetSkinInfo2Cache(ctx context.Context, key string, skinInfos []*api.SkinInfo) (err error) {
	conn := d.redisShow.Get(ctx)
	defer conn.Close()
	if len(skinInfos) > 0 {
		var skinInfoJSON []byte
		skinInfoJSONs := make([]string, 0)
		for _, skinInfo := range skinInfos {
			if skinInfoJSON, err = json.Marshal(skinInfo); err != nil {
				log.Error("resource-job.Dao.SetSkinExt2Cache json marshalling Error (%+v)", err)
				return err
			}
			skinInfoJSONs = append(skinInfoJSONs, string(skinInfoJSON))
		}
		log.Info("redis set %s:%v", key, skinInfoJSONs)
		if _, err = redis.String(conn.Do("SET", key, skinInfoJSONs)); err != nil {
			log.Error("resource-job.Dao.SetSkinExt2Cache redis setting Error (%+v)", err)
			return err
		}
	} else {
		if _, err = conn.Do("DEL", key); err != nil {
			log.Error("resource-job.Dao.SetSkinExt2Cache redis deleting Error (%+v)", err)
			return err
		}
	}
	return
}
