package fawkes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	xtime "go-common/library/time"

	mdl "go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
)

const logTitle = "FeedbackDao"
const tableName = "app_feedback"
const NilInt = -99

// FeedbackCount count feedback
func (d *Dao) FeedbackCount(c context.Context, appKey, versionCode, buvId, brand, model, osver, province, isp, description, remark, business, principal, crashReason, operator string, mid, status int64, robotKey string, isBug bool, startTime, endTime xtime.Time, createStart, createEnd xtime.Time) (count int, err error) {
	q := d.queryBy(appKey, versionCode, buvId, brand, model, osver, province, isp, description, remark, business, principal, crashReason, operator, mid, status, robotKey, isBug, startTime, endTime, createStart, createEnd)
	if err = q.Count(&count).Error; err != nil {
		log.Errorc(c, "%s\tFeedbackCount err: %+v", logTitle, err)
	}
	return
}

// FeedbackList list feed back
func (d *Dao) FeedbackList(c context.Context, appKey, versionCode, buvId, brand, model, osver, province, isp, description, remark, business, principal, crashReason, operator string, mid, status int64, robotKey string, isBug bool, startTime, endTime, createStart, createEnd xtime.Time, pn, ps int) (feedbackList []*mdl.FeedbackDB, err error) {
	if err = d.queryBy(appKey, versionCode, buvId, brand, model, osver, province, isp, description, remark, business, principal, crashReason, operator, mid, status, robotKey, isBug, startTime, endTime, createStart, createEnd).Order("id DESC").Offset((pn - 1) * ps).Limit(ps).Find(&feedbackList).Error; err != nil {
		log.Errorc(c, "%s\tfeedbackList error(%+v)", logTitle, err)
	}
	return
}

// FeedbackList alert 统计一定时间范围内部的数据作企业微信通知
func (d *Dao) FeedbackAlert(c context.Context, robotKey string) (feedbackList []*mdl.FeedbackDB, err error) {
	if err = d.ORMDB.Table(tableName).Where("wx_robots like ?", "%"+robotKey+"%").Where("ctime >= date(now()) AND ctime < DATE_ADD(date(now()),INTERVAL 1 DAY)").Find(&feedbackList).Error; err != nil {
		log.Errorc(c, "%s\tFeedbackAlert error(%+v)", logTitle, err)
	}
	return
}

func (d *Dao) FeedbackInsert(c context.Context, dbModel *mdl.FeedbackDB) (id int64, err error) {
	if err = d.ORMDB.Table(tableName).Create(&dbModel).Error; err != nil {
		log.Errorc(c, "%s\tinsert error: %+v", logTitle, err)
	}
	id = dbModel.ID
	return
}

func (d *Dao) FeedbackQueryByPk(c context.Context, id int64) (re mdl.FeedbackDB, err error) {
	if err = d.ORMDB.Table(tableName).First(&re, id).Error; err != nil {
		log.Errorc(c, "%s\t query by pk error: %+v", logTitle, err)
	}
	return
}

func (d *Dao) FeedbackDeleteByPk(c context.Context, id int64) (eff int64, err error) {
	exec := d.ORMDB.Table(tableName).Delete(&mdl.FeedbackDB{}, id)
	if err = exec.Error; err != nil {
		log.Errorc(c, "%s\tdelete err: %v", logTitle, err)
	} else {
		eff = exec.RowsAffected
	}
	return
}

func (d *Dao) CreateTapdBug(c context.Context, reqForm *mdl.FeedbackTapdBug, token string) (bugID string, err error) {
	var re struct {
		Msg  string `json:"message"`
		Code int    `json:"code"`
		Data string `json:"data"`
	}
	bytes, _ := json.Marshal(&reqForm)
	payload := strings.NewReader(string(bytes))
	req, _ := http.NewRequest("POST", "http://marthe.bilibili.co/ep/admin/marthe/v1/outer/tapd/bug/create", payload)
	req.Header.Add("content-type", "application/json; charset=utf-8")
	req.Header.Add("token", token)
	err = d.httpClient.Do(c, req, &re)
	if err != nil {
		log.Error("http create tapd false: %v", err)
		return
	}
	if re.Code != 0 {
		errStr := fmt.Sprintf("创建Tapd Bug失败-%v: %v", re.Code, re.Msg)
		err = errors.New(errStr)
		return
	}
	bugID = re.Data
	return
}

func (d *Dao) FeedbackUpdateByPk(c context.Context, dbModel *mdl.FeedbackDB) (effect int64, err error) {
	update := d.ORMDB.Table(tableName).Model(&mdl.FeedbackDB{}).Where("id = ?", dbModel.ID).Select("app_key", "principal", "crash_reason", "media_urls", "overview_img_url", "wx_robots", "description", "version_code", "buvid", "contact", "brand", "model", "osver", "isp", "province", "crash_time", "status", "editor", "bv", "remark", "business", "wx_robot_ids", "is_bug", "tapd_url", "send_to").Update(structs.Map(dbModel))
	if err = update.Error; err != nil {
		log.Errorc(c, "%s\tupdate by pk error: %+v", logTitle, err)
	} else {
		effect = update.RowsAffected
	}
	return
}

func (d *Dao) queryBy(appKey, versionCode, buvId, brand, model, osver, province, isp, description, remark, business, principal, crashReason, operator string, mid, status int64, robotKey string, isBug bool, startTime, endTime, createStart, createEnd xtime.Time) (query *gorm.DB) {
	query = d.ORMDB.Table(tableName)
	if len(appKey) != 0 {
		query = query.Where("app_key = ?", appKey)
	}
	if len(versionCode) != 0 {
		query = query.Where("version_code = ?", versionCode)
	}
	if len(buvId) != 0 {
		query = query.Where("buvid = ?", buvId)
	}
	if len(brand) != 0 {
		query = query.Where("brand = ?", brand)
	}
	if len(model) != 0 {
		query = query.Where("model = ?", model)
	}
	if len(osver) != 0 {
		query = query.Where("osver = ?", osver)
	}
	if len(province) != 0 {
		query = query.Where("province = ?", province)
	}
	if len(isp) != 0 {
		query = query.Where("isp = ?", isp)
	}
	if len(principal) != 0 {
		query = query.Where("principal = ?", principal)
	}
	if len(description) != 0 {
		query = query.Where("description like ?", "%"+description+"%")
	}
	if len(remark) != 0 {
		query = query.Where("remark like ?", "%"+remark+"%")
	}
	if len(business) != 0 {
		query = query.Where("business = ?", business)
	}
	if len(crashReason) != 0 {
		query = query.Where("crash_reason like ?", "%"+crashReason+"%")
	}
	if len(operator) != 0 {
		query = query.Where("operator = ?", operator)
	}
	if mid != 0 {
		query = query.Where("mid = ?", mid)
	}
	if status != NilInt {
		query = query.Where("status = ?", status)
	}
	if len(robotKey) != 0 {
		query = query.Where("FIND_IN_SET (?, wx_robot_ids)", robotKey)
	}
	if isBug {
		query = query.Where("is_bug = ?", isBug)
	}
	if startTime != 0 {
		query = query.Where("crash_time > ?", startTime)
	}
	if endTime != 0 {
		query = query.Where("crash_time < ?", endTime)
	}
	if createStart != 0 {
		query = query.Where("ctime > ?", createStart)
	}
	if createEnd != 0 {
		query = query.Where("ctime < ?", createEnd)
	}
	return
}
