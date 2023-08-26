package search_whitelist

import (
	"context"
	"fmt"

	"go-common/library/log"
	xtime "go-common/library/time"

	api "git.bilibili.co/bapis/bapis-go/archive/service"

	model "go-gateway/app/app-svr/app-feed/admin/model/search_whitelist"
)

// 获取配置列表
func (d *Dao) GetWhiteList(c context.Context, idList []int64, cUser string, sTime, eTime xtime.Time, statusList []int, filterHidden bool, ps, pn int) (total int, resList []*model.WhiteListItem, err error) {
	query := d.DB.Model(&model.WhiteListItem{})
	if len(idList) > 0 {
		query = query.Where("id in (?)", idList)
	}
	if cUser != "" {
		query = query.Where("c_user = ?", cUser)
	}
	if sTime != 0 && eTime != 0 {
		query = query.Where("stime <= ?", eTime).Where("etime >= ?", sTime)
	}
	if len(statusList) > 0 {
		query = query.Where("status in (?)", statusList)
	}
	if filterHidden {
		query = query.Where("hidden = 0")
	}
	query = query.Where("is_deleted = 0")
	err = query.Count(&total).Error
	if err != nil {
		log.Error("GetWhiteList Count() error(%v)", err)
		return
	}
	err = query.Order("`id` desc").Offset(ps * (pn - 1)).Find(&resList).Error
	if err != nil {
		log.Error("GetWhiteList Find() error(%v)", err)
		return
	}
	return
}

// 批量获取多个pid下面的全部视频
func (d *Dao) GetWhiteListArchive(c context.Context, pidList []int64) (archiveMap map[int64][]*model.WhiteListArchiveItem, err error) {
	var (
		archiveList []*model.WhiteListArchiveItem
	)
	archiveMap = make(map[int64][]*model.WhiteListArchiveItem)
	for _, pid := range pidList {
		archiveMap[pid] = []*model.WhiteListArchiveItem{}
	}
	query := d.DB.Model(&model.WhiteListArchiveItem{}).
		Where("pid in (?)", pidList).
		Where("is_deleted = 0").
		Order("`pid` asc, `rank` asc")
	err = query.Find(&archiveList).Error
	if err != nil {
		return
	}
	for _, archive := range archiveList {
		archive.CardType = 4
		archiveMap[archive.Pid] = append(archiveMap[archive.Pid], archive)
	}
	return
}

// 批量获取多个pid下面的全部搜索词
func (d *Dao) GetWhiteListSearchWord(c context.Context, pidList []int64) (searchWordMap map[int64][]string, err error) {
	var (
		searchWordList []*model.WhiteListQueryItem
	)
	searchWordMap = make(map[int64][]string)
	for _, pid := range pidList {
		searchWordMap[pid] = []string{}
	}
	query := d.DB.Model(&model.WhiteListQueryItem{}).
		Where("pid in (?)", pidList).
		Where("is_deleted = 0").
		Order("`pid` asc, `id` asc")
	err = query.Find(&searchWordList).Error
	if err != nil {
		return
	}
	for _, searchWord := range searchWordList {
		searchWordMap[searchWord.Pid] = append(searchWordMap[searchWord.Pid], searchWord.SearchWord)
	}
	return
}

// 根据 search_word 搜索配置 id
func (d *Dao) SearchWhiteListIDByQuery(c context.Context, searchWord string) (idList []int64, err error) {
	var res []*model.WhiteListQueryItem
	query := d.DB.Model(&model.WhiteListQueryItem{}).
		Where("search_word = ?", searchWord).
		Where("is_deleted = 0")
	err = query.Find(&res).Error
	if err != nil {
		//nolint:govet
		log.Error("SearchWhiteListIDByQuery() Find() error(err)", err)
		return
	}
	for _, item := range res {
		idList = append(idList, item.Pid)
	}

	return
}

// 插入配置列表
func (d *Dao) InsertWhiteList(c context.Context, stime, etime xtime.Time, cUser string, roleId int64, pid int64) (id int64, err error) {
	newItem := &model.WhiteListItem{
		STime:  stime,
		ETime:  etime,
		CUser:  cUser,
		RoleId: roleId,
		Pid:    pid,
		Status: model.StatusWaitAudit1,
	}
	err = d.DB.Model(&model.WhiteListItem{}).Create(newItem).Error
	if err != nil {
		log.Error("InsertWhiteList Create() error(%v)", err)
		return
	}
	id = newItem.ID
	return
}

// 编辑配置列表
func (d *Dao) UpdateWhiteList(c context.Context, id int64, stime, etime xtime.Time, status, hidden int) (err error) {
	newItem := &model.WhiteListItem{
		STime:  stime,
		ETime:  etime,
		Status: status,
	}
	query := d.DB.Model(&model.WhiteListItem{}).Where("id = ?", id)
	err = query.Updates(newItem).Error
	if err != nil {
		log.Error("UpdateWhiteList Updates() error(%v)", err)
		return
	}
	err = query.Update("hidden", hidden).Error
	if err != nil {
		log.Error("UpdateWhiteList Update() error(%v)", err)
		return
	}
	return
}

// 批量插入搜索词
func (d *Dao) InsertWhiteListQuery(c context.Context, searchWordList []string, pid int64, username string) (err error) {
	sqlStr := "insert into search_whitelist_query(search_word, pid) values "
	valueStr := ""
	for i, searchWord := range searchWordList {
		if i == 0 {
			valueStr += fmt.Sprintf("('%v', '%v')", searchWord, pid)
		} else {
			valueStr += ", " + fmt.Sprintf("('%v', '%v')", searchWord, pid)
		}
	}
	sqlStr += valueStr
	err = d.DB.Exec(sqlStr).Error
	if err != nil {
		log.Error("InsertWhiteListQuery DB.Exec() error(%v)", err)
		return
	}
	return
}

// 批量插入稿件
func (d *Dao) InsertWhiteListArchive(c context.Context, avidList []int64, pid int64, username string) (err error) {
	if len(avidList) < 1 {
		return
	}
	sqlStr := "insert into search_whitelist_archive(`avid`, `pid`, `rank`) values "
	valueStr := ""
	curRank := 1
	for i, avid := range avidList {
		if i == 0 {
			valueStr += fmt.Sprintf("('%v', '%v', '%v')", avid, pid, curRank)
		} else {
			valueStr += ", " + fmt.Sprintf("('%v', '%v', '%v')", avid, pid, curRank)
		}
		curRank += 1
	}
	sqlStr += valueStr
	err = d.DB.Exec(sqlStr).Error
	if err != nil {
		log.Error("InsertWhiteListArchive DB.Exec() error(%v)", err)
		return
	}
	return
}

// 批量变更配置状态
func (d *Dao) UpdateWhiteListStatus(c context.Context, pidList []int64, status int, username string) (err error) {
	if len(pidList) < 1 {
		return
	}
	query := d.DB.Model(&model.WhiteListItem{}).
		Where("id in (?)", pidList)
	err = query.Update("status", status).Error
	if err != nil {
		log.Error("UpdateWhiteListStatus Update() error(%v)", err)
		return
	}
	return
}

// 删除某一个配置
func (d *Dao) DeleteWhiteListItem(c context.Context, pid int64, username string) (err error) {
	if pid == 0 {
		return
	}
	query := d.DB.Model(&model.WhiteListItem{}).
		Where("id = ?", pid)
	err = query.Update("is_deleted", 1).Error
	if err != nil {
		log.Error("DeleteWhiteListItem Update() error(%v)", err)
		return
	}
	return
}

// 删除某个pid下面的全部搜索词
func (d *Dao) DeleteWhiteListSearchWord(c context.Context, pid int64, username string) (err error) {
	if pid == 0 {
		return
	}
	query := d.DB.Model(&model.WhiteListQueryItem{}).
		Where("pid = ?", pid)
	err = query.Update("is_deleted", 1).Error
	if err != nil {
		log.Error("DeleteWhiteListSearchWord Update() error(%v)", err)
		return
	}
	return
}

// 批量删除关联稿件
func (d *Dao) DeleteWhiteListArchive(c context.Context, pid int64, username string) (err error) {
	if pid == 0 {
		return
	}
	query := d.DB.Model(&model.WhiteListArchiveItem{}).
		Where("pid = ?", pid)
	err = query.Update("is_deleted", 1).Error
	if err != nil {
		log.Error("DeleteWhiteListArchive Update() error(%v)", err)
		return
	}
	return
}

// 检查冲突
func (d *Dao) CheckConflict(c context.Context, id int64, searchWord []string, stime, etime xtime.Time) (conflictList []string, err error) {
	var (
		queryList              []model.WhiteListQueryItem
		queryPidList           []int64
		queryPidSearchWordsMap = map[int64]string{}
		configList             []model.WhiteListItem
	)
	query1 := d.DB.Model(&model.WhiteListQueryItem{}).
		Where("search_word in (?)", searchWord).
		Where("pid != ?", id).
		Where("is_deleted = 0")
	err = query1.Find(&queryList).Error
	if err != nil {
		log.Error("CheckConflict Find() error(%v)", err)
		return
	}

	for _, item := range queryList {
		queryPidList = append(queryPidList, item.Pid)
		queryPidSearchWordsMap[item.Pid] = item.SearchWord
	}

	query2 := d.DB.Model(&model.WhiteListItem{}).
		Where("id in (?)", queryPidList).
		Where("stime < ?", etime).
		Where("etime < ?", stime).
		Where("is_deleted = 0")
	err = query2.Find(&configList).Error
	if err != nil {
		log.Error("CheckConflict Find() error(%v)", err)
		return
	}
	for _, item := range configList {
		conflictList = append(conflictList, queryPidSearchWordsMap[item.ID])
	}

	return
}

// Arcs gets archives
func (d *Dao) Arcs(c context.Context, aids []int64) (res map[int64]*api.Arc, err error) {
	var (
		arg   = &api.ArcsRequest{Aids: aids}
		reply *api.ArcsReply
	)
	if reply, err = d.arcClient.Arcs(c, arg); err != nil {
		log.Error("d.arcRPC.Archive3(%v) error(%+v)", arg, err)
		return
	}
	res = reply.Arcs
	return
}
