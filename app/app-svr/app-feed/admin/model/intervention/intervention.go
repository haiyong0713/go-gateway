package intervention

import "go-common/library/time"

type Detail struct {
	ID           uint   `json:"id" form:"id" gorm:"column:id"`
	Avid         int64  `json:"avid" form:"avid" gorm:"column:avid"`
	Bvid         string `json:"-" form:"-" gorm:"-"`
	Title        string `json:"title" form:"title" gorm:"column:title"`
	List         string `json:"list" form:"list" gorm:"column:list"`
	Pic          string `json:"pic" form:"pic" gorm:"-"`
	CreatedBy    string `json:"created_by" form:"created_by" gorm:"column:created_by"`
	StartTime    int64  `json:"start_time" form:"start_time" gorm:"column:start_time"`
	EndTime      int64  `json:"end_time" form:"end_time" gorm:"column:end_time"`
	OnlineStatus int64  `json:"online_status" form:"online_status" gorm:"column:online_status"`
}

type OptLogDetail struct {
	ID             uint      `json:"id" form:"id" gorm:"column:id"`
	Avid           int64     `json:"avid" form:"avid" gorm:"column:avid"`
	InterventionId uint      `json:"intervention_id" form:"intervention_id" gorm:"column:intervention_id"`
	OpUser         string    `json:"op_user" form:"op_user" gorm:"column:op_user"`
	OpType         uint      `json:"op_type" form:"op_type" gorm:"column:op_type"`
	MBefore        string    `json:"m_before" form:"m_before" gorm:"column:m_before"`
	MAfter         string    `json:"m_after" form:"m_after" gorm:"column:m_after"`
	Ctime          time.Time `json:"ctime" form:"ctime" gorm:"column:ctime"`
}
