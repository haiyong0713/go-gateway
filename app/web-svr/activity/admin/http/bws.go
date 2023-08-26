package http

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model"
)

const _signType = 7

// addBws 增加赛程
func addBws(c *bm.Context) {
	arg := new(model.ActBws)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Create(arg).Error; err != nil {
		log.Error("addBws(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// saveBws 存储赛程
func saveBws(c *bm.Context) {
	arg := new(model.ActBws)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Model(&model.ActBws{ID: arg.ID}).Update(arg).Error; err != nil {
		log.Error("saveBws(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// bwsInfo 赛程信息
func bwsInfo(c *bm.Context) {
	arg := new(model.ActBws)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.ID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err := actSrv.DB.First(arg, arg.ID).Error; err != nil {
		log.Error("bwsInfo(%d) error(%v)", arg.ID, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(arg, nil)
}

// bwsList 比赛对象列表
func bwsList(c *bm.Context) {
	var (
		err   error
		count int
		list  []*model.ActBws
	)
	v := new(struct {
		Del  int8 `form:"del" default:"0"`
		Page int  `form:"pn" default:"1"`
		Size int  `form:"ps" default:"20"`
	})
	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	db := actSrv.DB
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("del = ?", v.Del)
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).
		Find(&list).Error; err != nil {
		log.Error("bwsList(%d,%d) error(%v)", v.Page, v.Size, err)
		c.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActBws{}).Count(&count).Error; err != nil {
		log.Error("bwsList count error(%v)", err)
		c.JSON(nil, err)
		return
	}

	data := map[string]interface{}{
		"data":  list,
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	c.JSONMap(data, nil)
}

// addBwsAchievement 增加赛程
func addBwsAchievement(c *bm.Context) {
	arg := new(model.ActBwsAchievement)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Create(arg).Error; err != nil {
		log.Error("addBws(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// saveBwsAchievement 存储赛程
func saveBwsAchievement(c *bm.Context) {
	arg := new(model.ActBwsAchievement)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Model(&model.ActBwsAchievement{ID: arg.ID}).Update(arg).Error; err != nil {
		log.Error("saveBwsAchievement(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// actBwsAchievement 赛程信息
func bwsAchievement(c *bm.Context) {
	arg := new(model.ActBwsAchievement)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.ID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err := actSrv.DB.First(arg, arg.ID).Error; err != nil {
		log.Error("bwsAchievement(%d) error(%v)", arg.ID, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(arg, nil)
}

// bwsList 比赛对象列表
func bwsAchievements(c *bm.Context) {
	var (
		err   error
		count int
		list  []*model.ActBwsAchievement
	)
	v := new(struct {
		BID  int64 `form:"bid" default:"0"`
		Del  int8  `form:"del" default:"0"`
		Page int   `form:"pn" default:"1"`
		Size int   `form:"ps" default:"20"`
	})
	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	db := actSrv.DB
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("del = ?", v.Del)
	if v.BID != 0 {
		db = db.Where("bid = ?", v.BID)
	}
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).
		Find(&list).Error; err != nil {
		log.Error("bwsAchievements(%d,%d) error(%v)", v.Page, v.Size, err)
		c.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActBwsAchievement{}).Count(&count).Error; err != nil {
		log.Error("bwsAchievements count error(%v)", err)
		c.JSON(nil, err)
		return
	}

	data := map[string]interface{}{
		"data":  list,
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	c.JSONMap(data, nil)
}

// addBwsField 增加赛程
func addBwsField(c *bm.Context) {
	arg := new(model.ActBwsField)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Create(arg).Error; err != nil {
		log.Error("addBwsFieldws(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// saveBwsField 存储赛程
func saveBwsField(c *bm.Context) {
	arg := new(model.ActBwsField)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Model(&model.ActBwsField{ID: arg.ID}).Update(arg).Error; err != nil {
		log.Error("saveBwsField(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// actBwsAchievement 赛程信息
func bwsField(c *bm.Context) {
	arg := new(model.ActBwsField)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.ID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err := actSrv.DB.First(arg, arg.ID).Error; err != nil {
		log.Error("bwsField(%d) error(%v)", arg.ID, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(arg, nil)
}

// bwsList 比赛对象列表
func bwsFields(c *bm.Context) {
	var (
		err   error
		count int
		list  []*model.ActBwsField
	)
	v := new(struct {
		BID  int64 `form:"bid" default:"0"`
		Del  int8  `form:"del" default:"0"`
		Page int   `form:"pn" default:"1"`
		Size int   `form:"ps" default:"20"`
	})
	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	db := actSrv.DB
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("del = ?", v.Del)
	if v.BID != 0 {
		db = db.Where("bid = ?", v.BID)
	}
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).
		Find(&list).Error; err != nil {
		log.Error("bwsFields(%d,%d) error(%v)", v.Page, v.Size, err)
		c.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActBwsField{}).Count(&count).Error; err != nil {
		log.Error("bwsFields count error(%v)", err)
		c.JSON(nil, err)
		return
	}

	data := map[string]interface{}{
		"data":  list,
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	c.JSONMap(data, nil)
}

// addBwsPoint
func addBwsPoint(c *bm.Context) {
	arg := new(model.ActBwsPointArg)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.LockType == _signType {
		c.JSON(nil, actSrv.AddPointSign(c, arg))
		return
	}
	if err := actSrv.DB.Create(&arg.ActBwsPoint).Error; err != nil {
		log.Error("addBwsPoint(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// saveBwsPoint 存储赛程
func saveBwsPoint(c *bm.Context) {
	arg := new(model.ActBwsPointArg)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.LockType == _signType {
		tx := actSrv.DB.Begin()
		if err := tx.Model(&model.ActBwsPoint{ID: arg.ID}).Update(&arg.ActBwsPoint).Error; err != nil {
			log.Error("saveBwsPoint(%v) error(%v)", arg, err)
			tx.Rollback()
			c.JSON(nil, err)
			return
		}
		if err := tx.Model(&model.ActBwsPointSign{}).Where("pid = ?", arg.ID).Update(map[string]int64{"points": arg.TotalPoint, "sign_points": arg.SignPoint}).Error; err != nil {
			log.Error("saveBwsPoint sign(%v) error(%v)", arg, err)
			tx.Rollback()
			c.JSON(nil, err)
			return
		}
		if err := tx.Commit().Error; err != nil {
			c.JSON(nil, err)
			return
		}
	} else {
		if err := actSrv.DB.Model(&model.ActBwsPoint{ID: arg.ID}).Update(&arg.ActBwsPoint).Error; err != nil {
			log.Error("saveBwsPoint(%v) error(%v)", arg, err)
			c.JSON(nil, err)
			return
		}
	}
	c.JSON(nil, nil)
}

// actBwsAchievement 赛程信息
func bwsPoint(c *bm.Context) {
	arg := new(model.ActBwsPoint)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.ID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err := actSrv.DB.First(arg, arg.ID).Error; err != nil {
		log.Error("bwsPoint(%d) error(%v)", arg.ID, err)
		c.JSON(nil, err)
		return
	}
	if arg.LockType == _signType {
		sign := new(model.ActBwsPointSign)
		if err := actSrv.DB.Where("pid=?", arg.ID).Find(sign).Error; err != nil {
			log.Error("bwsPoint sign(%d) error(%v)", arg.ID, err)
		} else {
			c.JSON(&model.ActBwsPointResult{ActBwsPoint: arg, TotalPoint: sign.Points, ProvidePoints: sign.ProvidePoints, SignPoint: sign.SignPoints}, nil)
			return
		}
	}
	c.JSON(arg, nil)
}

// bwsList 比赛对象列表
func bwsPoints(c *bm.Context) {
	var (
		err      error
		count    int
		list     []*model.ActBwsPoint
		pids     []int64
		plids    []int64
		levels   []*model.ActBwsPointLevel
		awards   []*model.ActBwsPointAward
		res      []*model.ActBwsPointResult
		levelRes map[int64][]*model.ActBwsPointLevelResult
		awardRes map[int64][]*model.ActBwsPointAward
	)
	v := new(struct {
		FID      int64 `form:"fid" default:"0"`
		BID      int64 `form:"bid" default:"0"`
		LockType int64 `form:"lock_type" default:"lock_type"`
		Del      int8  `form:"del" default:"0"`
		Page     int   `form:"pn" default:"1"`
		Size     int   `form:"ps" default:"20"`
	})
	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	db := actSrv.DB
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("del = ?", v.Del)
	if v.BID != 0 {
		db = db.Where("bid = ?", v.BID)
	}
	if v.FID != 0 {
		db = db.Where("fid = ?", v.FID)
	}
	if v.LockType != 0 {
		db = db.Where("lock_type = ?", v.LockType)
	}
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).
		Find(&list).Error; err != nil {
		log.Error("bwsPoints(%d,%d) error(%v)", v.Page, v.Size, err)
		c.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActBwsPoint{}).Count(&count).Error; err != nil {
		log.Error("bwsPoints count error(%v)", err)
		c.JSON(nil, err)
		return
	}
	for _, v := range list {
		pids = append(pids, v.ID)
	}
	if len(pids) > 0 {
		if err = actSrv.DB.Where("pid IN (?)", pids).Where("is_delete=?", 0).Find(&levels).Error; err != nil {
			log.Error("bwsPoints level(%v) error(%v)", pids, err)
			err = nil
		} else {
			levelRes = make(map[int64][]*model.ActBwsPointLevelResult)
			for _, v := range levels {
				plids = append(plids, v.ID)
				levelRes[v.Pid] = append(levelRes[v.Pid], &model.ActBwsPointLevelResult{ActBwsPointLevel: v})
			}
			if len(plids) > 0 {
				if err = actSrv.DB.Where("pl_id IN (?)", plids).Where("is_delete=?", 0).Find(&awards).Error; err != nil {
					log.Error("bwsPoints award(%v) error(%v)", plids, err)
					err = nil
				} else {
					awardRes = make(map[int64][]*model.ActBwsPointAward)
					for _, v := range awards {
						awardRes[v.PlID] = append(awardRes[v.PlID], v)
					}
				}
			}
		}
	}
	for _, v := range list {
		item := &model.ActBwsPointResult{ActBwsPoint: v}
		if level, ok := levelRes[v.ID]; ok {
			item.Level = level
			for _, v := range item.Level {
				if award, ok := awardRes[v.ID]; ok {
					v.Awards = award
				}
			}
		}
		res = append(res, item)
	}
	data := map[string]interface{}{
		"data":  res,
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	c.JSONMap(data, nil)
}

func saveBwsPointLevel(c *bm.Context) {
	v := new(struct {
		Pid  int64  `form:"pid" validate:"min=1"`
		Data string `form:"data" validate:"min=1"`
	})
	if err := c.Bind(v); err != nil {
		return
	}
	param := new(model.SaveBwsPointLevelParam)
	if err := json.Unmarshal([]byte(v.Data), &param); err != nil {
		log.Warn("saveBwsPointLevel json.Unmarshal(%s) error(%v)", v.Data, err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if len(param.Levels) == 0 {
		log.Warn("saveBwsPointLevel param.Levels == 0")
		c.JSON(nil, ecode.RequestErr)
		return
	}
	levelIntMap := make(map[int]struct{}, len(param.Levels))
	for _, v := range param.Levels {
		levelIntMap[v.Level] = struct{}{}
	}
	if len(levelIntMap) < len(param.Levels) {
		log.Warn("saveBwsPointLevel param level repeat")
		c.JSON(nil, ecode.RequestErr)
		return
	}
	c.JSON(nil, actSrv.AddPointLevel(c, v.Pid, param))
}

// addBwsUserAchievement 保存用户
func addBwsUserAchievement(c *bm.Context) {
	arg := new(model.ActBwsUserAchievement)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Create(arg).Error; err != nil {
		log.Error("addBwsUserAchievement(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// saveBwsUserAchievement 保存用户成就
func saveBwsUserAchievement(c *bm.Context) {
	arg := new(model.ActBwsUserAchievement)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Model(&model.ActBwsUserAchievement{ID: arg.ID}).Update(arg).Error; err != nil {
		log.Error("saveBwsUserAchievement(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// bwsUserAchievement 用户成就信息
func bwsUserAchievement(c *bm.Context) {
	arg := new(model.ActBwsUserAchievement)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.ID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err := actSrv.DB.First(arg, arg.ID).Error; err != nil {
		log.Error("bwsUserAchievement(%d) error(%v)", arg.ID, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(arg, nil)
}

// bwsUserAchievements 用户成就列表
func bwsUserAchievements(c *bm.Context) {
	var (
		err   error
		count int
		list  []*model.ActBwsUserAchievement
	)
	v := new(struct {
		MID  int64 `form:"mid" default:"0"`
		AID  int64 `form:"aid" default:"0"`
		BID  int64 `form:"bid" default:"0"`
		Del  int8  `form:"del" default:"0"`
		Page int   `form:"pn" default:"1"`
		Size int   `form:"ps" default:"20"`
	})
	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	db := actSrv.DB
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("del = ?", v.Del)
	if v.BID != 0 {
		db = db.Where("bid = ?", v.BID)
	}
	if v.AID != 0 {
		db = db.Where("aid = ?", v.AID)
	}
	if v.MID != 0 {
		db = db.Where("mid = ?", v.MID)
	}
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).
		Find(&list).Error; err != nil {
		log.Error("bwsUserAchievements(%d,%d) error(%v)", v.Page, v.Size, err)
		c.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActBwsUserAchievement{}).Count(&count).Error; err != nil {
		log.Error("bwsUserAchievements count error(%v)", err)
		c.JSON(nil, err)
		return
	}

	data := map[string]interface{}{
		"data":  list,
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	c.JSONMap(data, nil)
}

// addBwsUserPoint 保存用户
func addBwsUserPoint(c *bm.Context) {
	arg := new(model.ActBwsUserPoint)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Create(arg).Error; err != nil {
		log.Error("addBwsUserPoint(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// saveBwsUserAchievement 保存用户成就
func saveBwsUserPoint(c *bm.Context) {
	arg := new(model.ActBwsUserPoint)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Model(&model.ActBwsUserPoint{ID: arg.ID}).Update(arg).Error; err != nil {
		log.Error("saveBwsUserPoint(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// bwsUserPoint 用户点数信息
func bwsUserPoint(c *bm.Context) {
	arg := new(model.ActBwsUserPoint)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.ID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err := actSrv.DB.First(arg, arg.ID).Error; err != nil {
		log.Error("bwsUserPoint(%d) error(%v)", arg.ID, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(arg, nil)
}

// bwsUserPoints 用户点数列表
func bwsUserPoints(c *bm.Context) {
	var (
		err   error
		count int
		list  []*model.ActBwsUserPoint
	)
	v := new(struct {
		MID      int64  `form:"mid" default:"0"`
		KEY      string `form:"key" default:""`
		LockType int64  `form:"lock_type" default:"lock_type"`
		BID      int64  `form:"bid" default:"0"`
		Del      int8   `form:"del" default:"0"`
		Page     int    `form:"pn" default:"1"`
		Size     int    `form:"ps" default:"20"`
	})
	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	db := actSrv.DB
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("del = ?", v.Del)
	if v.BID != 0 {
		db = db.Where("bid = ?", v.BID)
	}
	if v.KEY != "" {
		db = db.Where("key = ?", v.KEY)
	}
	if v.MID != 0 {
		db = db.Where("mid = ?", v.MID)
	}
	if v.LockType != 0 {
		db = db.Where("lock_type = ?", v.LockType)
	}
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).
		Find(&list).Error; err != nil {
		log.Error("bwsUserPoints(%d,%d) error(%v)", v.Page, v.Size, err)
		c.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActBwsUserPoint{}).Count(&count).Error; err != nil {
		log.Error("bwsUserPoints count error(%v)", err)
		c.JSON(nil, err)
		return
	}

	data := map[string]interface{}{
		"data":  list,
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	c.JSONMap(data, nil)
}

// addBwsUser 添加用户
func addBwsUser(c *bm.Context) {
	arg := new(model.ActBwsUser)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Create(arg).Error; err != nil {
		log.Error("addBwsUserPoint(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// saveBwsUserAchievement 保存用户成就
func saveBwsUser(c *bm.Context) {
	arg := new(model.ActBwsUser)
	if err := c.Bind(arg); err != nil {
		return
	}
	if err := actSrv.DB.Model(&model.ActBwsUser{ID: arg.ID}).Update(arg).Error; err != nil {
		log.Error("saveBwsUserPoint(%v) error(%v)", arg, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, nil)
}

// bwsUserPoint 用户点数信息
func bwsUser(c *bm.Context) {
	arg := new(model.ActBwsUser)
	if err := c.Bind(arg); err != nil {
		return
	}
	if arg.ID == 0 {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if err := actSrv.DB.First(arg, arg.ID).Error; err != nil {
		log.Error("bwsUserPoint(%d) error(%v)", arg.ID, err)
		c.JSON(nil, err)
		return
	}
	c.JSON(arg, nil)
}

func bwsUsersImport(c *blademaster.Context) {
	var (
		err  error
		data []byte
	)
	v := new(struct {
		BID int64 `form:"bid"`
	})

	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if v.BID == 0 {
		v.BID = actSrv.GetBid(c)
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorc(c, "importDetailCSV upload err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Errorc(c, "importDetailCSV ioutil.ReadAll err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Errorc(c, "importDetailCSV r.ReadAll() err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var args []int64
Loop:
	for _, row := range records {
		// import csv state online
		var mid int64
		for field, value := range row {
			value = strings.TrimSpace(value)
			switch field {
			case 0:
				if value == "" {
					log.Warn("importDetailCSV name mid(%s)", value)
					continue Loop
				}
				mid, err = strconv.ParseInt(value, 10, 64)
				if err != nil {
					// continue Loop
					log.Errorc(c, "records error (%v)", err)
					continue
				}

			}

		}
		args = append(args, mid)
	}
	if len(args) == 0 {
		log.Errorc(c, "importDetailCSV args no after filter")
		c.JSON(nil, ecode.RequestErr)
		return
	}
	go func() {
		ctx := context.Background()
		start := 0
		length := 1000
		for {
			end := start + length
			if len(args) < end {
				end = len(args)
			}
			actSrv.AddUserToken(ctx, v.BID, args[start:end])
			start = start + length
			if end == len(args) {
				break
			}
		}
	}()
	c.JSON(nil, nil)
}

func bwsUsersVipImport(c *blademaster.Context) {
	v := new(struct {
		BID  int64 `form:"bid"`
		Nums int   `form:"nums"`
	})

	if err := c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	if v.BID == 0 {
		v.BID = actSrv.GetBid(c)
	}
	go func() {
		ctx := context.Background()
		actSrv.AddVipUserToken(ctx, v.BID, v.Nums)
	}()
	c.JSON(nil, nil)
}

// bwsUserPoints 用户点数列表
func bwsUsers(c *bm.Context) {
	var (
		err   error
		count int
		list  []*model.ActBwsUser
	)
	v := new(struct {
		MID  int64  `form:"mid" default:"0"`
		KEY  string `form:"key" default:""`
		BID  int64  `form:"bid" default:"0"`
		Del  int8   `form:"del" default:"0"`
		Page int    `form:"pn" default:"1"`
		Size int    `form:"ps" default:"20"`
	})
	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	db := actSrv.DB
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	db = db.Where("del = ?", v.Del)
	if v.BID != 0 {
		db = db.Where("bid = ?", v.BID)
	}
	if v.KEY != "" {
		db = db.Where("key = ?", v.KEY)
	}
	if v.MID != 0 {
		db = db.Where("mid = ?", v.MID)
	}
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).
		Find(&list).Error; err != nil {
		log.Error("bwsUsers(%d,%d) error(%v)", v.Page, v.Size, err)
		c.JSON(nil, err)
		return
	}
	if err = db.Model(&model.ActBwsUser{}).Count(&count).Error; err != nil {
		log.Error("bwsUsers count error(%v)", err)
		c.JSON(nil, err)
		return
	}

	data := map[string]interface{}{
		"data":  list,
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}
	c.JSONMap(data, nil)
}

func bwsTasks(ctx *bm.Context) {
	var count int64
	if err := actSrv.DB.Model(&model.ActBwsTask{}).Count(&count).Error; err != nil {
		log.Error("bwsTasks count error(%v)", err)
		ctx.JSON(nil, err)
		return
	}
	var list []*model.ActBwsTask
	if err := actSrv.DB.Find(&list).Error; err != nil {
		log.Error("bwsTasks error(%v)", err)
		ctx.JSON(nil, err)
		return
	}
	data := map[string]interface{}{
		"data":  list,
		"total": count,
	}
	ctx.JSONMap(data, nil)
}

func bwsTaskAdd(ctx *bm.Context) {
	v := new(model.AddActBwsTaskArg)
	if err := ctx.Bind(v); err != nil {
		return
	}
	taskAdd := &model.ActBwsTask{
		Title:       v.Title,
		Cate:        v.Cate,
		FinishCount: v.FinishCount,
		RuleIDs:     v.RuleIDs,
		OrderNum:    v.OrderNum,
		State:       v.State,
	}
	if err := actSrv.DB.Model(&model.ActBwsTask{}).Create(taskAdd).Error; err != nil {
		log.Error("bwsTaskAdd s.dao.DB.Model Create(%+v) error(%v)", v, err)
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, nil)
}

func bwsTaskEdit(ctx *bm.Context) {
	v := new(model.SaveActBwsTaskArg)
	if err := ctx.Bind(v); err != nil {
		return
	}
	upData := map[string]interface{}{
		"title":        v.Title,
		"cate":         v.Cate,
		"finish_count": v.FinishCount,
		"rule_ids":     v.RuleIDs,
		"order_num":    v.OrderNum,
		"state":        v.State,
	}
	if err := actSrv.DB.Model(&model.ActBwsTask{}).Where("id=?", v.ID).Update(upData).Error; err != nil {
		log.Error("bwsTaskEdit s.dao.DB.Model Update(%+v) error(%v)", v, err)
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, nil)
}

func bwsTaskDel(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	if err := actSrv.DB.Model(&model.ActBwsTask{}).Where("id=?", v.ID).Update(map[string]interface{}{"state": 0}).Error; err != nil {
		log.Error("bwsTaskDel id:%d error:%v", v.ID, err)
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, nil)
}

func bwsAwards(ctx *bm.Context) {
	var count int64
	if err := actSrv.DB.Model(&model.ActBwsTask{}).Count(&count).Error; err != nil {
		log.Error("bwsAwards count error(%v)", err)
		ctx.JSON(nil, err)
		return
	}
	var list []*model.ActBwsAward
	if err := actSrv.DB.Find(&list).Error; err != nil {
		log.Error("bwsAwards error(%v)", err)
		ctx.JSON(nil, err)
		return
	}
	data := map[string]interface{}{
		"data":  list,
		"total": count,
	}
	ctx.JSONMap(data, nil)
}

func bwsAwardAdd(ctx *bm.Context) {
	v := new(model.AddActBwsAwardArg)
	if err := ctx.Bind(v); err != nil {
		return
	}
	awardAdd := &model.ActBwsAward{
		Title:    v.Title,
		Image:    v.Image,
		Intro:    v.Intro,
		Cate:     v.Cate,
		IsOnline: v.IsOnline,
		Stock:    v.Stock,
		OwnerMid: v.OwnerMid,
		Stage:    v.Stage,
		State:    v.State,
	}
	if err := actSrv.DB.Model(&model.ActBwsAward{}).Create(awardAdd).Error; err != nil {
		log.Error("bwsAwardAdd s.dao.DB.Model Create(%+v) error(%v)", v, err)
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, nil)
}

func bwsAwardEdit(ctx *bm.Context) {
	v := new(model.SaveActBwsAwardArg)
	if err := ctx.Bind(v); err != nil {
		return
	}
	upData := map[string]interface{}{
		"title":     v.Title,
		"image":     v.Image,
		"intro":     v.Intro,
		"cate":      v.Cate,
		"is_online": v.IsOnline,
		"stock":     v.Stock,
		"owner_mid": v.OwnerMid,
		"stage":     v.Stage,
		"state":     v.State,
	}
	if err := actSrv.DB.Model(&model.ActBwsAward{}).Where("id=?", v.ID).Update(upData).Error; err != nil {
		log.Error("bwsAwardEdit s.dao.DB.Model Update(%+v) error(%v)", v, err)
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, nil)
}

func bwsAwardDel(ctx *bm.Context) {
	v := new(struct {
		ID int64 `form:"id" validate:"min=1"`
	})
	if err := ctx.Bind(v); err != nil {
		return
	}
	if err := actSrv.DB.Model(&model.ActBwsAward{}).Where("id=?", v.ID).Update(map[string]interface{}{"state": 0}).Error; err != nil {
		log.Error("bwsAwardDel id:%d error:%v", v.ID, err)
		ctx.JSON(nil, err)
		return
	}
	ctx.JSON(nil, nil)
}
