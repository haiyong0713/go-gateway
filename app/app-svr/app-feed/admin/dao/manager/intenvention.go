package manager

import (
	"fmt"

	"go-gateway/app/app-svr/app-feed/admin/model/intervention"
)

// 新增干预记录
func (d *Dao) InsertDetail(detail *intervention.Detail) (id uint, err error) {
	create := &intervention.Detail{
		Avid:      detail.Avid,
		Title:     detail.Title,
		List:      detail.List,
		CreatedBy: detail.CreatedBy,
		StartTime: detail.StartTime,
		EndTime:   detail.EndTime,
	}
	err = d.DBShow.Table("intervention_details").Create(&create).Error
	id = create.ID
	return
}

// 修改干预记录详情
func (d *Dao) EditDetail(detail *intervention.Detail) (err error) {
	updates := &intervention.Detail{
		List:      detail.List,
		StartTime: detail.StartTime,
		EndTime:   detail.EndTime,
	}
	err = d.DBShow.Table("intervention_details").Model(&intervention.Detail{ID: detail.ID}).Updates(updates).Error
	return
}

// 修改干预记录状态
func (d *Dao) ChangeStatus(detail *intervention.Detail, newStatus int64) (err error) {
	err = d.DBShow.Table("intervention_details").Model(&detail).Update("online_status", newStatus).Error
	return
}

// 根据条件查询干预列表
func (d *Dao) List(filters *intervention.Detail, pageNum int) (result []intervention.Detail, err error) {
	pageSize := 20
	query := d.DBShow.Table("intervention_details")

	if filters.Avid != 0 {
		fmt.Println(filters.Avid)
		query = query.Where("avid = ?", filters.Avid)
	}
	if filters.Title != "" {
		fmt.Println(filters.Title)
		query = query.Where("title like ?", "%"+filters.Title+"%")
	}
	if filters.StartTime != 0 {
		fmt.Println(filters.StartTime)
		query = query.Where("start_time >= ?", filters.StartTime)
	}
	if filters.EndTime != 0 {
		fmt.Println(filters.EndTime)
		query = query.Where("end_time <= ?", filters.EndTime)
	}
	if filters.CreatedBy != "" {
		fmt.Println(filters.CreatedBy)
		query = query.Where("created_by = ?", filters.CreatedBy)
	}
	err = query.Order("id desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&result).Error
	return
}

// 根据条件查询干预列表总数
func (d *Dao) ListCount(filters *intervention.Detail) (count int, err error) {
	query := d.DBShow.Table("intervention_details")
	if filters.Avid != 0 {
		query = query.Where("avid = ?", filters.Avid)
	}
	if filters.Title != "" {
		query = query.Where("title like ?", "%"+filters.Title+"%")
	}
	if filters.StartTime != 0 {
		query = query.Where("start_time >= ?", filters.StartTime)
	}
	if filters.EndTime != 0 {
		query = query.Where("end_time <= ?", filters.EndTime)
	}
	if filters.CreatedBy != "" {
		query = query.Where("created_by = ?", filters.CreatedBy)
	}
	err = query.Count(&count).Error
	return count, err
}

// check 能否激活/加入记录，同一个avid是否和已激活的干预在生效时间上有overlap
func (d *Dao) CheckIntervention(detail *intervention.Detail) (count int, err error) {
	query := d.DBShow.Table("intervention_details")
	if detail.ID != 0 {
		query = query.Not("id = ?", detail.ID)
	}
	query = query.Not("online_status = 2")
	query = query.Where("avid = ?", detail.Avid)
	query = query.Where("start_time <= ? AND end_time >= ?", detail.EndTime, detail.StartTime)

	err = query.Count(&count).Error
	return
}

// 根据ID搜索干预
func (d *Dao) FindInterventionById(id uint) (result intervention.Detail, err error) {
	err = d.DBShow.Table("intervention_details").Where("id = ?", id).First(&result).Error
	return
}

// 插入一条操作日志
func (d *Dao) InsertOptLog(logDetail *intervention.OptLogDetail) (err error) {
	err = d.DBShow.Table("intervention_operations").Create(&logDetail).Error
	return
}

// 根据条件查询操作日志
func (d *Dao) OptLogList(filters *intervention.OptLogDetail, pageNum int) (result []intervention.OptLogDetail, err error) {
	pageSize := 20
	query := d.DBShow.Table("intervention_operations")
	if filters.InterventionId != 0 {
		query = query.Where("intervention_id = ?", filters.InterventionId)
	}
	if filters.Avid != 0 {
		query = query.Where("avid = ?", filters.Avid)
	}
	if filters.OpUser != "" {
		query = query.Where("op_user = ?", filters.OpUser)
	}
	err = query.Order("id desc").Offset((pageNum - 1) * pageSize).Limit(pageSize).Find(&result).Error
	return
}

// 操作日志的总量，分页用
func (d *Dao) OpLogListCount(filters *intervention.OptLogDetail) (count int, err error) {
	query := d.DBShow.Table("intervention_operations")
	if filters.Avid != 0 {
		query = query.Where("avid = ?", filters.Avid)
	}
	if filters.OpUser != "" {
		query = query.Where("op_user = ?", filters.OpUser)
	}
	err = query.Count(&count).Error
	return count, err
}
