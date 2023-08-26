package college

import "time"

// College .
type College struct {
	ID              int64     `form:"id" json:"id"`
	ProvinceID      int64     `form:"province_id" json:"province_id"`
	Province        string    `form:"province" json:"province"`
	CollegeName     string    `form:"college_name" json:"college_name"`
	Initial         string    `form:"initial" json:"initial"`
	TagID           int64     `form:"tag_id" json:"tag_id"`
	ProvinceInitial string    `form:"province_initial" json:"province_initial"`
	White           string    `form:"white" json:"white"`
	Mid             int64     `form:"mid" json:"mid"`
	RelationMid     string    `form:"relation_mid" json:"relation_mid"`
	Score           int64     `form:"score" json:"score"`
	State           int64     `form:"state" default:"1" json:"state"`
	Ctime           time.Time `form:"ctime"  json:"ctime"`
	Mtime           time.Time `form:"mtime"  json:"mtime" `
}

// Reply ...
type Reply struct {
	List []*College             `json:"list"`
	Page map[string]interface{} `json:"page"`
}

// AidReply ...
type AidReply struct {
	List []*AIDList             `json:"list"`
	Page map[string]interface{} `json:"page"`
}

// AIDList ...
type AIDList struct {
	ID    int64     `form:"id" json:"id"`
	Aid   int64     `form:"aid" json:"aid"`
	Score int64     `form:"score" json:"score"`
	State int       `form:"state" json:"state" default:"1"`
	Ctime time.Time `form:"ctime"  json:"ctime"`
	Mtime time.Time `form:"mtime"  json:"mtime" `
}

// TableName ...
func (*College) TableName() string {
	return "act_college"
}

// TableName ...
func (*AIDList) TableName() string {
	return "act_college_aid"
}
