package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model"

	"github.com/jinzhu/gorm"
)

const (
	_addPointSignSQL    = "INSERT INTO act_bws_point_sign(bid,pid,stime,etime,points,sign_points) VALUE %s"
	_addPointLevelSQL   = "INSERT INTO act_bws_points_level(bid,pid,`level`,points) VALUES %s"
	_addPointAwardSQL   = "INSERT INTO act_bws_points_award(bid,pl_id,`name`,icon,amount) VALUES %s"
	_addUserTokenSQL    = "INSERT INTO act_bws_users(mid,`key`,bid) VALUES %s"
	_addUserVipTokenSQL = "INSERT INTO act_bws_vip_token(`vip_key`,bid) VALUES %s"
)

// GetBid 获取当前bid
func (s *Service) GetBid(c context.Context) int64 {
	return s.c.Bws.Bid
}

// AddUserToken 导入用户和token
func (s *Service) AddUserToken(c context.Context, bid int64, mids []int64) (err error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "tx.Rollback()  %v", r)
			err = ecode.Error(ecode.RequestErr, "保存失败")
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	var userToken []*model.ActBwsUsers
	if err = s.DB.Where("mid in (?)", mids).Where("bid =?", bid).Find(&userToken).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "s.DB.Where(mid in %v).Find(),error(%v)", mids, err)
		return
	}
	userTokenMap := make(map[int64]struct{})
	for _, v := range userToken {
		userTokenMap[v.MID] = struct{}{}
	}
	newMids := make([]int64, 0)
	for _, v := range mids {
		if _, ok := userTokenMap[v]; ok {
			continue
		}
		newMids = append(newMids, v)
	}
	var (
		awardStr  []string
		awardArgs []interface{}
	)
	if len(newMids) > 0 {
		for _, v := range newMids {
			awardStr = append(awardStr, "(?,?,?)")
			awardArgs = append(awardArgs, v, createBwsKey(bid, v, time.Now().Unix()), bid)
		}
		if err = tx.Exec(fmt.Sprintf(_addUserTokenSQL, strings.Join(awardStr, ",")), awardArgs...).Error; err != nil {
			log.Errorc(c, "AddUserToken  Exec(%s) error(%v)", _addUserTokenSQL, err)
			return
		}
	}
	return nil

}

// AddVipUserToken 导入用户和token
func (s *Service) AddVipUserToken(c context.Context, bid int64, nums int) (err error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Errorc(c, "tx.Rollback()  %v", r)
			err = ecode.Error(ecode.RequestErr, "保存失败")
			return
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Errorc(c, "tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Errorc(c, "tx.Commit() error(%v)", err)
		}
	}()
	var (
		awardStr  []string
		awardArgs []interface{}
	)
	for i := 0; i < nums; i++ {
		index := i
		awardStr = append(awardStr, "(?,?)")
		awardArgs = append(awardArgs, createBwsKey(bid, int64(index), time.Now().Unix()), bid)
	}
	if err = tx.Exec(fmt.Sprintf(_addUserVipTokenSQL, strings.Join(awardStr, ",")), awardArgs...).Error; err != nil {
		log.Errorc(c, "addUserVipTokenSQL  Exec(%s) error(%v)", _addUserVipTokenSQL, err)
		return
	}
	return nil

}

func createBwsKey(bid, mid, ts int64) string {
	hasher := md5.New()
	key := fmt.Sprintf("%d_%d_%d_VMVFh6", bid, mid, ts)
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))[0:15]
}

// AddPointSign .
func (s *Service) AddPointSign(c context.Context, arg *model.ActBwsPointArg) (err error) {
	if s.c.Bws.StartDay == 0 || s.c.Bws.EndDay == 0 || s.c.Bws.StartHour == 0 || s.c.Bws.EndHour == 0 || arg.TotalPoint == 0 || arg.SignPoint == 0 {
		err = ecode.RequestErr
		return
	}
	if s.c.Bws.EndDay < s.c.Bws.StartDay || s.c.Bws.EndHour < s.c.Bws.StartHour {
		err = ecode.RequestErr
		return
	}
	var (
		signs      []*model.ActBwsPointSign
		start, end time.Time
		rowStr     []string
		args       []interface{}
	)
	for i := s.c.Bws.StartDay; i <= s.c.Bws.EndDay; i++ {
		for j := s.c.Bws.StartHour; j <= s.c.Bws.EndHour; j++ {
			hourStr := strconv.FormatInt(i, 10) + " " + strconv.Itoa(j)
			if start, err = time.ParseInLocation("20060102 15:04:05", hourStr+":00:00", time.Local); err != nil {
				log.Error("time.Parse %s error(%v)", hourStr+":00:00", err)
				err = ecode.RequestErr
				return
			}
			if end, err = time.ParseInLocation("20060102 15:04:05", hourStr+":59:59", time.Local); err != nil {
				log.Error("time.Parse %s error(%v)", hourStr+":59:59", err)
				err = ecode.RequestErr
				return
			}
			sign := &model.ActBwsPointSign{
				Bid:        arg.BID,
				Stime:      start.Unix(),
				Etime:      end.Unix(),
				Points:     arg.TotalPoint,
				SignPoints: arg.SignPoint,
			}
			signs = append(signs, sign)
		}
	}
	tx := s.DB.Begin()
	if err = tx.Create(&arg.ActBwsPoint).Error; err != nil {
		log.Error("addBwsPoint(%v) error(%v)", arg.ActBwsPoint, err)
		tx.Rollback()
		return
	}
	for _, v := range signs {
		rowStr = append(rowStr, "(?,?,?,?,?,?)")
		args = append(args, v.Bid, arg.ID, v.Stime, v.Etime, v.Points, v.SignPoints)
	}
	if err = tx.Exec(fmt.Sprintf(_addPointSignSQL, strings.Join(rowStr, ",")), args...).Error; err != nil {
		log.Error("addBwsPoint(%v) tx.Exec error(%v)", _addPointSignSQL, err)
		tx.Rollback()
		return
	}
	if err = tx.Commit().Error; err != nil {
		log.Error("addBwsPoint(%v)  tx.Commit() error(%v)", arg, err)
	}
	return
}

// AddPointLevel add point level and award.
func (s *Service) AddPointLevel(c context.Context, pid int64, arg *model.SaveBwsPointLevelParam) (err error) {
	var (
		preLevels []*model.ActBwsPointLevel
		prePlIDs  []int64
		levels    []*model.ActBwsPointLevel
		addAward  bool
	)
	point := new(model.ActBwsPoint)
	if err = s.DB.Model(&model.ActBwsPoint{}).Where("id = ?", pid).First(point).Error; err != nil {
		log.Error("AddPointLevel pid(%d) error(%v)", pid, err)
		return
	}
	if err = s.DB.Where("pid=?", pid).Find(&preLevels).Error; err != nil {
		log.Error("AddPointLevel pid(%d) error(%v)", pid, err)
		return
	}
	for _, v := range preLevels {
		if v.ID > 0 {
			prePlIDs = append(prePlIDs, v.ID)
		}
	}
	tx := s.DB.Begin()
	if len(prePlIDs) > 0 {
		if err = tx.Model(&model.ActBwsPointLevel{}).Where("id IN(?)", prePlIDs).Update("is_delete", 1).Error; err != nil {
			log.Error("AddPointLevel Update level(%d) error(%v)", pid, err)
			tx.Rollback()
			return
		}
		if err = tx.Model(&model.ActBwsPointAward{}).Where("pl_id IN(?)", prePlIDs).Update("is_delete", 1).Error; err != nil {
			log.Error("AddPointLevel Update level(%d) error(%v)", pid, err)
			tx.Rollback()
			return
		}
	}
	rowStr := make([]string, 0, len(arg.Levels))
	args := make([]interface{}, 0)
	for _, level := range arg.Levels {
		rowStr = append(rowStr, "(?,?,?,?)")
		args = append(args, point.BID, pid, level.Level, level.Points)
		if len(level.Awards) > 0 {
			addAward = true
		}
	}
	if err = tx.Exec(fmt.Sprintf(_addPointLevelSQL, strings.Join(rowStr, ",")), args...).Error; err != nil {
		log.Error("AddPointLevel level Exec(%s) error(%v)", _addPointLevelSQL, err)
		tx.Rollback()
		return
	}
	if addAward {
		var (
			awardStr  []string
			awardArgs []interface{}
		)
		if err = tx.Where("pid=?", pid).Find(&levels).Error; err != nil {
			log.Error("AddPointLevel Find(%d) error(%v)", pid, err)
			tx.Rollback()
			return
		}
		levelMap := make(map[int]*model.ActBwsPointLevel, len(levels))
		for _, v := range levels {
			levelMap[v.Level] = v
		}
		for _, v := range arg.Levels {
			if len(v.Awards) > 0 {
				if item, ok := levelMap[v.Level]; ok {
					for _, award := range v.Awards {
						awardStr = append(awardStr, "(?,?,?,?,?)")
						awardArgs = append(awardArgs, point.BID, item.ID, award.Name, award.Icon, award.Amount)
					}
				}
			}
		}
		if err = tx.Exec(fmt.Sprintf(_addPointAwardSQL, strings.Join(awardStr, ",")), awardArgs...).Error; err != nil {
			log.Error("AddPointLevel award Exec(%s) error(%v)", _addPointAwardSQL, err)
			tx.Rollback()
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		log.Error("AddPointLevel tx.Commit() error(%v)", err)
	}
	return
}
