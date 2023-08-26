package model

type DolbyWhiteList struct {
	ID          int64  `json:"id" gorm:"id" form:"id"`
	Model       string `json:"model" gorm:"model" form:"model"`
	Brand       string `json:"brand" gorm:"brand" form:"brand"`
	BFSPath     string `json:"bfs_path" gorm:"bfs_path" form:"bfs_path"`
	BFSPathHash string `json:"bfs_path_hash" gorm:"bfs_path_hash" form:"-"`
}

func (DolbyWhiteList) TableName() string {
	return "dolby_vision_whitelist"
}

type QnBlackList struct {
	ID     int64  `json:"id" gorm:"id" form:"id"`
	Model  string `json:"model" gorm:"model" form:"model"`
	Brand  string `json:"brand" gorm:"brand" form:"brand"`
	QnList string `json:"qn_list" gorm:"qn_list" form:"qn_list"`
}

func (QnBlackList) TableName() string {
	return "qn_blacklist"
}

type LimitFreeInfo struct {
	ID        int64  `json:"id" gorm:"id" form:"id"`
	Aid       int64  `json:"aid" gorm:"aid" form:"aid"`
	Stime     int64  `json:"stime" gorm:"stime" form:"stime"`
	Etime     int64  `json:"etime" gorm:"etime" form:"etime"`
	LimitFree int64  `json:"limit_free" gorm:"limit_free" form:"limit_free"`
	Subtitle  string `json:"subtitle" gorm:"subtitle" form:"subtitle"`
	Remark    string `json:"remark" gorm:"remark" form:"remark"`
	State     int64  `json:"state" gorm:"state" form:"state"`
}

func (LimitFreeInfo) TableName() string {
	return "resolution_limit_free"
}
