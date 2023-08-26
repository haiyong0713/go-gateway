package vogue

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go-common/library/log"

	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

// 获取全部配置
func (s *Service) ConfigList(c context.Context) (rsp map[string]string, err error) {
	var (
		list []*voguemdl.ConfigItem
	)
	if list, err = s.dao.ConfigList(c); err != nil {
		log.Error("[ConfigList] s.dao.ConfigList error(%v)", err)
		return
	}

	rsp = make(map[string]string)

	for _, config := range list {
		rsp[config.Name] = config.Config
	}

	return
}

func (s *Service) ConfigCreditLimit(c context.Context) (res *voguemdl.ConfigCreditLimit, err error) {
	var (
		listRes map[string]string
	)
	listRes, err = s.ConfigList(c)
	if err != nil {
		log.Error("[ConfigCreditLimit] s.dao.ConfigList error(%v)", err)
		return
	}
	res = &voguemdl.ConfigCreditLimit{}
	if res.DailyLimit, err = strconv.ParseInt(listRes["today_limit"], 10, 64); err != nil {
		log.Error("[ConfigCreditLimit] strconv.ParseInt today_limit error(%v), listRes(%v)", err, listRes)
		return
	}
	if res.ActDoubleStart, err = strconv.ParseInt(listRes["act_double_start"], 10, 64); err != nil {
		log.Error("[ConfigCreditLimit] strconv.ParseInt act_double_start error(%v)", err)
		return
	}
	if res.ActDoubleEnd, err = strconv.ParseInt(listRes["act_double_end"], 10, 64); err != nil {
		log.Error("[ConfigCreditLimit] strconv.ParseInt act_double_end error(%v)", err)
		return
	}
	if res.ActSecondDoubleStart, err = strconv.ParseInt(listRes["act_second_double_start"], 10, 64); err != nil {
		log.Error("[ConfigCreditLimit] strconv.ParseInt act_second_double_start error(%v)", err)
		return
	}
	if res.ActSecondDoubleEnd, err = strconv.ParseInt(listRes["act_second_double_end"], 10, 64); err != nil {
		log.Error("[ConfigCreditLimit] strconv.ParseInt act_second_double_end error(%v)", err)
		return
	}
	return

}

// 修改配置
func (s *Service) ModifyConfig(c context.Context, request map[string]string) (err error) {
	tx := s.dao.DB.Begin()

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit().Error
		return
	}()

	for k, v := range request {
		if err = s.dao.ModifyConfig(c, &voguemdl.ConfigItem{
			Name:   k,
			Config: v,
		}); err != nil {
			log.Error("s.dao.ModifyConfig(%v) error(%v)", request, err)
			return
		}
	}

	if err = s.dao.ModifyViewScore(c, &voguemdl.ConfigItem{
		Name:   "view_score",
		Config: request["view_score"],
	}); err != nil {
		log.Error("s.dao.ModifyViewScore(%v) error(%v)", request, err)
		return
	}

	if err = s.dao.ModifyDuration(c, &voguemdl.ConfigResponse{
		ActStart: request["act_start"],
		ActEnd:   request["act_end"],
	}); err != nil {
		return
	}

	return
}

// 双倍监控
func (s *Service) DoubleScoreMonitor(c context.Context) {
	for {
		log.Info("DoubleScoreMonitor start time(%v)", time.Now())

		configList, err := s.ConfigList(c)
		if err != nil {
			log.Error("ConfigList error(%v)", err)
			return
		}
		viewScore, err := strconv.Atoi(configList["view_score"])
		if err != nil {
			log.Error("strconv.Atoi(configList[view_score]) error(%v)", err)
			return
		}
		scoreList := &voguemdl.CritList{}
		err = json.Unmarshal([]byte(configList["score_list"]), scoreList)
		if err != nil {
			log.Error("json.Unmarshal([]byte(configList[score_list]) error(%v)", err)
			return
		}

		nextDoubleOn, err := s.CheckDoubleScore(c, configList)
		if err != nil {
			log.Error("CheckDoubleScore error(%v)", err)
			return
		}
		err = s.dao.ChangeDoubleStatus(c, nextDoubleOn, int64(viewScore), scoreList)
		if err != nil {
			log.Error("ChangeDoubleStatus error(%v)", err)
			return
		}

		time.Sleep(time.Duration(1) * time.Second)
	}
}

// 检测是否是双倍期间
func (s *Service) CheckDoubleScore(c context.Context, configList map[string]string) (doubleOn int, err error) {
	startTime, err := strconv.Atoi(configList["act_double_start"])
	endTime, err := strconv.Atoi(configList["act_double_end"])
	if err != nil {
		log.Error("strconv.Atoi(configList[act_double_start/end]) error(%v)", err)
		return
	}
	secondStartTime, err := strconv.Atoi(configList["act_second_double_start"])
	secondEndTime, err := strconv.Atoi(configList["act_second_double_end"])
	if err != nil {
		log.Error("strconv.Atoi(configList[act_second_double_start/end]) error(%v)", err)
		return
	}
	now := time.Now().Unix()
	if now >= int64(startTime) && now <= int64(endTime) {
		doubleOn = 1
	} else if now >= int64(secondStartTime) && now <= int64(secondEndTime) {
		doubleOn = 1
	} else {
		doubleOn = 0
	}
	return
}
