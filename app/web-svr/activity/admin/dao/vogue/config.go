package vogue

import (
	"context"
	"fmt"

	"go-common/library/log"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

const (
	_name            = "view"
	_activity        = "fashion_618"
	_vogueConfigList = "vogue_config_%s"
)

// 获取所有配置项
func (d *Dao) ConfigList(c context.Context) (list []*voguemdl.ConfigItem, err error) {
	if err = d.DB.Find(&list).Error; err != nil {
		log.Error("[ConfigList] d.DB.Find, error(%v)", err)
	}
	return
}

// 更改某一项配置
func (d *Dao) ModifyConfig(c context.Context, params *voguemdl.ConfigItem) (err error) {
	count := 0
	config := &voguemdl.ConfigItem{}
	if err = d.DB.Model(&config).Where("name = ?", params.Name).Count(&count).Error; err != nil {
		log.Error("[ModifyConfig] d.DB.Where, error(%v)", err)
		return
	}
	if count > 0 {
		if err = d.DB.Model(&config).Where("name = ?", params.Name).Updates(&params).Error; err != nil {
			log.Error("[ModifyConfig] d.DB.Update, error(%v)", err)
		}
		_ = d.DelCacheConfig(c, params.Name)
	} else {
		if err = d.DB.Model(&config).Create(&params).Error; err != nil {
			log.Error("[ModifyConfig] d.DB.Create, error(%v)", err)
		}
	}

	return
}

// 更改浏览的基础积分 view_score
func (d *Dao) ModifyViewScore(c context.Context, params *voguemdl.ConfigItem) (err error) {
	//var (
	//	res = &common.Counter{}
	//)
	//if res, err = d.actPlatClient.GetCounter(c, &actPlat.SimpleCounterReq{
	//	Name:     _name,
	//	Activity: _activity,
	//}); err != nil {
	//	log.Error("[VogueModifyViewScore] d.actPlatClient.GetCounter, params(%v) error(%v)", params, err)
	//	return
	//}

	//min, err := strconv.Atoi(params.Config)
	//if err != nil {
	//	log.Error("strconv.Atoi(params.Config) error(%v)", err)
	//	return
	//}
	//res.Content.RandRange.Min = int64(min)
	//res.Content.RandRange.Max = int64(min)
	//
	//if _, err = d.actPlatClient.UpdateCounter(c, &actPlat.FullCounterReq{
	//	Name:        res.Name,
	//	Activity:    res.Activity,
	//	DataSource:  res.DataSource,
	//	NeedHistory: res.NeedHistory,
	//	Content:     res.Content,
	//}); err != nil {
	//	log.Error("[VogueModifyViewScore] d.actPlatClient.UpdateCounter, params(%v) error(%v)", params, err)
	//	return
	//}

	return
}

// 更新活动的周期，起止时间
func (d *Dao) ModifyDuration(c context.Context, params *voguemdl.ConfigResponse) (err error) {
	//startTime, err := strconv.Atoi(params.ActStart)
	//endTime, err := strconv.Atoi(params.ActEnd)
	//if err != nil {
	//	log.Error("strconv.Atoi(params.ActStart/ActEnd) error(%v)", err)
	//	return
	//}
	//if _, err = d.actPlatClient.UpdateActivity(c, &actPlat.FullActivityReq{
	//	Name:        _activity,
	//	Description: _activity,
	//	Contact:     "chenliang02,daizhichen",
	//	StartTime:   xtime.Time(int64(startTime)),
	//	EndTime:     xtime.Time(int64(endTime)),
	//}); err != nil {
	//	log.Error("[VogueModifyViewScore] d.actPlatClient.UpdateActivity, params(%v) error(%v)", params, err)
	//	return
	//}
	return
}

// 更新暴击的积分的值
func (d *Dao) ChangeDoubleStatus(c context.Context, doubleOn int, viewScore int64, scoreList *voguemdl.CritList) (err error) {
	//var (
	//	res = &common.Counter{}
	//)
	//if res, err = d.actPlatClient.GetCounter(c, &actPlat.SimpleCounterReq{
	//	Name:     _name,
	//	Activity: _activity,
	//}); err != nil {
	//	log.Error("[VogueModifyActivityDuration] d.actPlatClient.GetCounter, error(%v)", err)
	//	return
	//}
	//log.Info("d.actPlatClient.GetCounter, %+v", res)
	//log.Info("double on: %d", doubleOn)
	//
	//critList := make([]*common.CounterCrit, len(*scoreList))
	//for i, v := range *scoreList {
	//	if doubleOn == 1 {
	//		// 更新counter为双倍
	//		critList[i] = &common.CounterCrit{
	//			CritCount: v.Num,
	//			RandRange: &common.IntRange{
	//				Min: v.Min * 2,
	//				Max: v.Max * 2,
	//			},
	//		}
	//	} else {
	//		// 更新counter为默认值
	//		critList[i] = &common.CounterCrit{
	//			CritCount: v.Num,
	//			RandRange: &common.IntRange{
	//				Min: v.Min,
	//				Max: v.Max,
	//			},
	//		}
	//	}
	//}
	//
	//res.Content.Crit = critList
	//
	//if doubleOn == 1 {
	//	// 更新基础值的counter为双倍
	//	res.Content.RandRange.Min = viewScore * 2
	//	res.Content.RandRange.Max = viewScore * 2
	//} else {
	//	// 更新基础值的counter为默认值
	//	res.Content.RandRange.Min = viewScore
	//	res.Content.RandRange.Max = viewScore
	//}
	//
	//if _, err = d.actPlatClient.UpdateCounter(c, &actPlat.FullCounterReq{
	//	Name:        res.Name,
	//	Activity:    res.Activity,
	//	DataSource:  res.DataSource,
	//	NeedHistory: res.NeedHistory,
	//	Content:     res.Content,
	//}); err != nil {
	//	log.Error("[VogueChangeDoubleStatus] d.actPlatClient.UpdateCounter, params(%v) error(%v)", doubleOn, err)
	//	return
	//}

	return
}

func keyVogueConfigList(name string) string {
	return fmt.Sprintf(_vogueConfigList, name)
}

// 删除redis的配置缓存
func (d *Dao) DelCacheConfig(c context.Context, name string) (err error) {
	conn := d.redis.Get(c)
	defer conn.Close()

	key := keyVogueConfigList(name)

	if _, err = conn.Do("DEL", key); err != nil {
		log.Error("DelCacheConfig(%d) error(%v)", key, err)
	}
	return
}
