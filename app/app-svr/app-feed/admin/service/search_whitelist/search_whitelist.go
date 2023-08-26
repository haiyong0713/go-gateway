package search_whitelist

import (
	"context"
	"fmt"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"

	"go-gateway/app/app-svr/app-feed/admin/model/common"
	model "go-gateway/app/app-svr/app-feed/admin/model/search_whitelist"
	Log "go-gateway/app/app-svr/app-feed/admin/util"
)

// 获取预览
func (s *Service) SearchWhiteListArchivePreview(c context.Context, avidList []int64) (previewList []*model.WhiteListArchiveItem, err error) {
	if len(avidList) < 1 {
		return
	}
	arcs, err := s.dao.Arcs(c, avidList)
	if err != nil {
		return
	}
	i := int64(0)
	for _, avid := range avidList {
		v, ok := arcs[avid]
		if ok {
			i += 1
			previewList = append(previewList, &model.WhiteListArchiveItem{
				Avid:  avid,
				Title: v.Title,
				Cover: v.Pic + "@216w.webp",
				Rank:  i,
			})
		}
	}
	return
}

// 审核 驳回 下线 删除
func (s *Service) SearchWhiteListOption(c context.Context, id int64, operation string, username string) (err error) {
	// 查出配置
	_, currentConfigList, err := s.dao.GetWhiteList(c, []int64{id}, "", 0, 0, nil, false, 1, 1)
	if err != nil {
		return
	}
	if len(currentConfigList) < 1 {
		err = ecode.Error(ecode.RequestErr, "配置不存在")
		return
	}
	currentConfig := currentConfigList[0]

	logItem := map[string]interface{}{
		"id":        id,
		"operation": operation,
	}
	obj := map[string]interface{}{
		"value": logItem,
		"id":    id,
	}
	if err = Log.AddLogs(common.LogSearchWhiteList, username, 0, 0, "SearchWhiteListOption", obj); err != nil {
		log.Error("search whitelist SearchWhiteListOption AddLog error(%v)", err)
		return
	}

	if operation == "delete" {
		err = s.dao.DeleteWhiteListItem(c, id, username)
		if err != nil {
			return
		}
		return
	}

	var statusMap = map[string]int{
		"pass1":   model.StatusWaitAudit2,
		"pass2":   model.StatusWaitOnline,
		"pass":    model.StatusWaitOnline,
		"online":  model.StatusOnline,
		"offline": model.StatusOffline,
		"reject":  model.StatusReject,
	}

	status, ok := statusMap[operation]
	if !ok {
		err = ecode.Error(ecode.RequestErr, "操作参数错误")
		return
	}

	err = s.dao.UpdateWhiteListStatus(c, []int64{id}, status, username)
	if err != nil {
		return
	}

	// 如果有关联的父配置
	if currentConfig.Pid > 0 {
		if status == model.StatusWaitOnline || status == model.StatusOnline {
			// 当前配置上线以后，把父配置删掉
			err = s.SearchWhiteListOption(c, currentConfig.Pid, "delete", "system")
			if err != nil {
				return
			}
		} else if status == model.StatusReject {
			// 当前配置审核被拒绝，把当前配置删除
			err = s.SearchWhiteListOption(c, id, "delete", "system")
			if err != nil {
				return
			}
			// 原始配置管理后台可见
			err = s.dao.UpdateWhiteList(c, currentConfig.Pid, 0, 0, 0, 0)
			if err != nil {
				return
			}
		}
	}
	return
}

// 添加
func (s *Service) SearchWhiteListAdd(c context.Context, searchWord []string, avidList []int64, stime, etime xtime.Time, username string, needCheck bool, pid int64) (err error) {
	if needCheck {
		var conflictList []string
		conflictList, err = s.dao.CheckConflict(c, 0, searchWord, stime, etime)
		if err != nil {
			return
		}
		if len(conflictList) > 0 {
			err = ecode.Error(ecode.RequestErr, fmt.Sprintf("搜索词重复：%v", conflictList))
			return
		}
	}
	logItem := map[string]interface{}{
		"search_word": searchWord,
		"avid_list":   avidList,
		"stime":       stime,
		"etime":       etime,
	}
	obj := map[string]interface{}{
		"value": logItem,
		"id":    0,
	}
	if err = Log.AddLogs(common.LogSearchWhiteList, username, 0, 0, "SearchWhiteListAdd", obj); err != nil {
		log.Error("search whitelist SearchWhiteListAdd AddLog error(%v)", err)
		return
	}
	// todo 后面会接分组
	var roleId int64 = 0
	id, err := s.dao.InsertWhiteList(c, stime, etime, username, roleId, pid)
	if err != nil {
		return
	}
	err = s.dao.InsertWhiteListQuery(c, searchWord, id, username)
	if err != nil {
		return
	}
	err = s.dao.InsertWhiteListArchive(c, avidList, id, username)
	if err != nil {
		return
	}
	return
}

// 修改
func (s *Service) SearchWhiteListEdit(c context.Context, id int64, searchWord []string, avidList []int64, stime, etime xtime.Time, username string, needCheck bool) (err error) {
	if needCheck {
		var conflictList []string
		conflictList, err = s.dao.CheckConflict(c, id, searchWord, stime, etime)
		if err != nil {
			return
		}
		if len(conflictList) > 0 {
			err = ecode.Error(ecode.RequestErr, fmt.Sprintf("搜索词重复：%v", conflictList))
			return
		}
	}
	logItem := map[string]interface{}{
		"search_word": searchWord,
		"avid_list":   avidList,
		"stime":       stime,
		"etime":       etime,
		"id":          id,
	}
	obj := map[string]interface{}{
		"value": logItem,
		"id":    id,
	}
	if err = Log.AddLogs(common.LogSearchWhiteList, username, 0, 0, "SearchWhiteListEdit", obj); err != nil {
		log.Error("search whitelist SearchWhiteListEdit AddLog error(%v)", err)
		return
	}
	_, currentConfigList, err := s.dao.GetWhiteList(c, []int64{id}, "", 0, 0, nil, false, 1, 1)
	if err != nil {
		return
	}
	if len(currentConfigList) < 1 {
		err = ecode.Error(ecode.RequestErr, "配置不存在")
		return
	}
	currentConfig := currentConfigList[0]
	// 如果是待生效、生效中的配置，就新建一份配置，新老关联，老的隐藏
	if currentConfig.Status == model.StatusOnline || currentConfig.Status == model.StatusWaitOnline {
		err = s.SearchWhiteListAdd(c, searchWord, avidList, stime, etime, username, false, id)
		if err != nil {
			return
		}
		err = s.dao.UpdateWhiteList(c, id, 0, 0, 0, 1)
		if err != nil {
			return
		}
	} else {
		// 如果是其他未生效配置，直接变更
		err = s.dao.UpdateWhiteList(c, id, stime, etime, model.StatusWaitAudit1, 0)
		if err != nil {
			return
		}
		err = s.dao.DeleteWhiteListSearchWord(c, id, username)
		if err != nil {
			return
		}
		err = s.dao.InsertWhiteListQuery(c, searchWord, id, username)
		if err != nil {
			return
		}
		err = s.dao.DeleteWhiteListArchive(c, id, username)
		if err != nil {
			return
		}
		err = s.dao.InsertWhiteListArchive(c, avidList, id, username)
		if err != nil {
			return
		}
	}
	return
}

// 根据配置id获取下面全部视频
func (s *Service) SearchWhiteListArchiveList(c context.Context, id int64) (previewList []*model.WhiteListArchiveItem, err error) {
	archiveMap, err := s.dao.GetWhiteListArchive(c, []int64{id})
	if err != nil {
		return
	}

	archiveList, ok := archiveMap[id]
	if !ok {
		return
	}

	var avidList []int64
	for _, item := range archiveList {
		avidList = append(avidList, item.Avid)
	}
	arcs, err := s.dao.Arcs(c, avidList)
	if err != nil {
		return
	}

	previewList = archiveList

	for i, item := range previewList {
		arc, ok := arcs[item.Avid]
		if ok {
			previewList[i].Title = arc.Title
			previewList[i].Cover = arc.Pic + "@216w.webp"
		}
	}

	return
}

// 获取白名单配置列表
func (s *Service) SearchWhiteList(c context.Context, status int, stime, etime xtime.Time, cUser, searchWord string, filterHidden bool, ps, pn int) (total int, resList []*model.WhiteListItemWithQueryAndArchive, err error) {
	idList, err := s.dao.SearchWhiteListIDByQuery(c, searchWord)
	if err != nil {
		return
	}

	statusList := []int{status}
	if status == model.StatusDefault {
		if stime != 0 && etime != 0 {
			statusList = []int{model.StatusOnline, model.StatusWaitOnline}
		} else {
			statusList = []int{}
		}
	}

	total, configList, err := s.dao.GetWhiteList(c, idList, cUser, stime, etime, statusList, filterHidden, ps, pn)
	if err != nil {
		return
	}

	if len(configList) < 1 {
		return
	}

	idList = []int64{}
	for _, item := range configList {
		idList = append(idList, item.ID)
	}

	queryMap, err := s.dao.GetWhiteListSearchWord(c, idList)
	if err != nil {
		return
	}

	archiveMap, err := s.dao.GetWhiteListArchive(c, idList)
	if err != nil {
		return
	}

	for _, whiteListItem := range configList {
		item := &model.WhiteListItemWithQueryAndArchive{
			WhiteListItem: whiteListItem,
		}
		if v, ok := queryMap[whiteListItem.ID]; ok {
			item.SearchWord = v
		}
		if v, ok := archiveMap[whiteListItem.ID]; ok {
			item.Archive = v
		}
		resList = append(resList, item)
	}
	return
}

// 定时更新配置的状态
func (s *Service) JobUpdateState() {
	for {
		s.AtomUpdateState()
		time.Sleep(5 * time.Second)
	}
}

// 更新配置状态的方法
func (s *Service) AtomUpdateState() {
	ctx := context.Background()
	currentTime := xtime.Time(time.Now().Unix())
	// 查出所有正在生效的
	// 超过结束时间的，全都设置成失效
	_, onlineList, err := s.dao.GetWhiteList(ctx, nil, "", 0, 0, []int{model.StatusOnline}, false, 9999, 1)
	if err != nil {
		log.Error("AtomUpdateState() s.dao.GetWhiteList() error(%v)", err)
		//nolint:ineffassign
		err = nil
	}
	var timeoutIdList []int64
	for _, v := range onlineList {
		if currentTime > v.ETime {
			timeoutIdList = append(timeoutIdList, v.ID)
		}
	}
	err = s.dao.UpdateWhiteListStatus(ctx, timeoutIdList, model.StatusOffline, "system")
	if err != nil {
		log.Error("AtomUpdateState() s.dao.UpdateWhiteListStatus() error(%v)", err)
		//nolint:ineffassign
		err = nil
	}

	// 查出所有待生效的
	// 超过开始时间，没超过结束时间的，全都设置成生效中
	_, waitOnlineList, err := s.dao.GetWhiteList(ctx, nil, "", 0, 0, []int{model.StatusWaitOnline}, false, 9999, 1)
	if err != nil {
		log.Error("AtomUpdateState() s.dao.GetWhiteList() error(%v)", err)
		//nolint:ineffassign
		err = nil
	}
	var waitOnlineIdList []int64
	for _, v := range waitOnlineList {
		if currentTime > v.STime && currentTime < v.ETime {
			waitOnlineIdList = append(waitOnlineIdList, v.ID)
		}
	}
	err = s.dao.UpdateWhiteListStatus(ctx, waitOnlineIdList, model.StatusOnline, "system")
	if err != nil {
		log.Error("AtomUpdateState() s.dao.UpdateWhiteListStatus() error(%v)", err)
		//nolint:ineffassign
		err = nil
	}

	// 查出所有待处理的，全都设置成待一审
	_, defaultList, err := s.dao.GetWhiteList(ctx, nil, "", 0, 0, []int{model.StatusDefault}, false, 9999, 1)
	if err != nil {
		log.Error("AtomUpdateState() s.dao.GetWhiteList() error(%v)", err)
		//nolint:ineffassign
		err = nil
	}
	var defaultIdList []int64
	for _, v := range defaultList {
		defaultIdList = append(defaultIdList, v.ID)
	}
	err = s.dao.UpdateWhiteListStatus(ctx, defaultIdList, model.StatusWaitAudit1, "system")
	if err != nil {
		log.Error("AtomUpdateState() s.dao.UpdateWhiteListStatus() error(%v)", err)
		//nolint:ineffassign
		err = nil
	}
}
