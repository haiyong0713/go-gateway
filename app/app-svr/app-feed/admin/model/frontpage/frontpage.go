package frontpage

import (
	xtime "go-common/library/time"
)

const (
	DefaultConfigID   = 1
	DefaultContractID = "frontpage"

	ActionOnline = "online"
	ActionHidden = "hidden"
	ActionDelete = "delete"
)

// Config assignment
type Config struct {
	ID               int64      `json:"id" gorm:"column:id"`
	ConfigName       string     `json:"name" gorm:"column:config_name"`
	ContractID       string     `json:"contract_id" gorm:"column:contract_id"`
	ResourceID       int64      `json:"resource_id" gorm:"column:resource_id"`
	Pic              string     `json:"pic" gorm:"column:pic"`
	LitPic           string     `json:"litpic" gorm:"column:litpic"`
	URL              string     `json:"url" gorm:"column:url"`
	Rule             string     `json:"rule" gorm:"column:rule"`
	Weight           int64      `json:"weight" gorm:"column:weight"`
	Agency           string     `json:"agency" gorm:"column:agency"`
	Price            float32    `json:"price" gorm:"column:price"`
	State            int        `json:"state" gorm:"column:state"`
	Atype            int8       `json:"atype" gorm:"column:atype"`
	STime            xtime.Time `json:"stime" gorm:"column:stime"`
	ETime            xtime.Time `json:"etime" gorm:"column:etime"`
	CTime            xtime.Time `json:"ctime" gorm:"column:ctime"`
	CUser            string     `json:"cuser" gorm:"column:cuser"`
	MTime            xtime.Time `json:"mtime" gorm:"column:mtime"`
	MUser            string     `json:"muser" gorm:"column:muser"`
	IsSplitLayer     int        `json:"is_split_layer" gorm:"column:is_split_layer"`
	SplitLayer       string     `json:"split_layer" gorm:"column:split_layer"`
	LocPolicyGroupID int64      `json:"loc_policy_group_id" gorm:"column:loc_policy_group_id"`
	Position         int64      `json:"pos" gorm:"-"`
	Auto             int        `json:"auto" gorm:"-"`
}

func (t *Config) TableName() string {
	return "frontpage"
}

type FrontpagesForFE struct {
	DefaultConfig *Config   `json:"default"`
	OnlineConfigs []*Config `json:"online"`
	HiddenConfigs []*Config `json:"hidden"`
}

type ConfigRule struct {
	IsCover int `json:"is_cover"`
	Style   int `json:"style"`
}

// User struct info of table user
type User struct {
	ID           int64      `json:"id" gorm:"column:id"`
	Username     string     `json:"username" gorm:"column:username"`
	Nickname     string     `json:"nickname" gorm:"column:nickname"`
	Email        string     `json:"email" gorm:"column:email"`
	Phone        int64      `json:"phone" gorm:"column:phone"`
	DepartmentID int        `json:"department_id" gorm:"column:department_id"`
	State        int        `json:"state" gorm:"column:state"`
	WXID         string     `json:"wx_id" gorm:"column:wx_id"`
	Ctime        xtime.Time `json:"ctime" gorm:"column:ctime"`
	Mtime        xtime.Time `json:"mtime" gorm:"column:mtime"`
}

// TableName return table name
func (a *User) TableName() string {
	return "user"
}
