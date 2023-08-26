package wishes_2021_spring

import (
	"strings"
)

type UserCommitListRequestInLive struct {
	LastID         int64  `form:"lastid" validate:"min=0"`
	ActivityUniqID string `form:"actid" validate:"required"`
	Order          string `form:"order" validate:"required"`
	Ps             int64  `form:"ps" validate:"min=10"`
	ActivityID     int64
}

type UserCommitListRespInLive struct {
	Total int64                    `json:"total"`
	Ps    int64                    `json:"ps"`
	Data  []map[string]interface{} `json:"list"`
}

type UserInfo struct {
	Mid            int64  `json:"mid"`
	Nickname       string `json:"nickname"`
	Face           string `json:"face"`
	Identification bool   `json:"identification"`
	Silence        bool   `json:"silence"`
	TelStatus      bool   `json:"tel_status"`
}

type AuditInfoText struct {
	Text string `json:"text"`
}
type AuditInfoImages struct {
	Url string `json:"url"`
}

type AuditInfo struct {
	Avid      int64  `json:"avid" validate:"min=1"`
	Activity  string `json:"activity" validate:"required"`
	Materials struct {
		User   UserInfo          `json:"user"`
		Text   []AuditInfoText   `json:"text"`
		Images []AuditInfoImages `json:"images"`
	} `json:"materials" validate:"required"`
	CTimeStr string `json:"ctime"`
	MtimeStr string `json:"mtime"`
}

type UserCommitManuscriptDB struct {
	Id         int64  `json:"id"`
	Mid        int64  `json:"mid"`
	ActivityId int64  `json:"activity_id"`
	Content    string `json:"content"`
	Bvid       string `json:"bvid"`
	CTimeStr   string `json:"ctime"`
	MtimeStr   string `json:"mtime"`
}

type UserCommit4AggregationWithUserInfo struct {
	*UserCommit4Aggregation
	UserInfo *UserInfo `json:"user_info"`
	LastId   int64     `json:"last_id"`
}

type UserCommit4Aggregation struct {
	Content   string                   `json:"saved_info"`
	ExtraList []map[string]interface{} `json:"video_posts"`
}

type CommonActivityConfig struct {
	ActivityID     int64    `toml:"id"`
	UniqID         string   `toml:"primary_key"`
	Remark         string   `toml:"remark"`
	MaxUploadTimes int64    `toml:"maxUpload"`
	StartTime      int64    `toml:"stime"`
	EndTime        int64    `toml:"etime"`
	AK             string   `toml:"appkey"`
	AS             string   `toml:"appsecret"`
	SendPropertys  []string `toml:"send_propertys"`
}

func NewUserCommit4Aggregation() (commit *UserCommit4Aggregation) {
	commit = new(UserCommit4Aggregation)
	{
		commit.ExtraList = make([]map[string]interface{}, 0)
	}

	return
}

func (req *UserCommitListRequestInLive) Validate() (isValid bool) {
	req.Order = strings.ToUpper(req.Order)
	if req.Order == "DESC" || req.Order == "ASC" {
		isValid = true

		return
	}

	return
}
