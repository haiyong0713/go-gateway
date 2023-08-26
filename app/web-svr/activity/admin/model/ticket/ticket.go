package ticket

import "go-common/library/time"

type ReqTicketCreate struct {
	Num int `json:"num" form:"num" validate:"required"`
}

type Ticket struct {
	ID     int64     `json:"id" gorm:"column:id"`
	Ticket string    `json:"ticket" gorm:"column:ticket"`
	State  uint8     `json:"state" gorm:"column:state"`
	Ctime  time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime  time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

func (Ticket) TableName() string {
	return "act_electronic_ticket"
}
