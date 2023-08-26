package up

import xtime "go-common/library/time"

// SearchReply .
type SearchReply struct {
	Count int64   `json:"count"`
	Pn    int     `json:"pn"`
	Ps    int     `json:"ps"`
	List  []*Item `json:"list"`
}

type Item struct {
	*UpAct
	Name string `json:"name"`
}

type UpAct struct {
	ID        int64      `json:"id" gorm:"column:id"`
	Mid       int64      `json:"Mid" gorm:"column:mid"`
	Title     string     `json:"title" gorm:"column:title"`
	Statement string     `json:"statement" gorm:"column:statement"`
	Stime     xtime.Time `json:"stime" time_format:"2006-01-02 15:04:05" gorm:"column:stime"`
	Etime     xtime.Time `json:"etime" time_format:"2006-01-02 15:04:05" gorm:"column:etime"`
	Aid       int64      `json:"aid" gorm:"column:aid"`
	State     int        `json:"state" gorm:"column:state"`
	Offline   int        `json:"offline" gorm:"column:offline"`
}

type ListReply struct {
	Count int64    `json:"count"`
	List  []*UpAct `json:"list"`
}

// TableName UpAct def.
func (UpAct) TableName() string {
	return "act_up"
}
