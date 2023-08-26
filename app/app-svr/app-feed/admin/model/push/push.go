package push

type PushDetail struct {
	ID                int64  `form:"id" gorm:"column:id" json:"id"`
	Title             string `form:"push_title" validate:"required" gorm:"column:title" json:"push_title"`
	Text              string `form:"push_text" validate:"required" gorm:"column:text" json:"push_text"`
	PackageUrl        string `form:"package_url" validate:"required" gorm:"column:package_url" json:"package_url"`
	STime             int64  `form:"stime" validate:"required" gorm:"column:stime" json:"stime"`
	ETime             int64  `form:"etime" validate:"required" gorm:"column:etime" json:"etime"`
	PopupTitle        string `form:"popup_title" validate:"required" gorm:"column:popup_title" json:"popup_title"`
	AppName           string `form:"app_name" validate:"required" gorm:"column:app_name" json:"app_name"`
	AppCurrentVersion string `form:"app_current_version" validate:"required" gorm:"column:app_current_version" json:"app_current_version"`
	AppUpdateTime     string `form:"app_update_time" validate:"required" gorm:"column:app_update_time" json:"app_update_time"`
	AppDeveloper      string `form:"app_developer" validate:"required" gorm:"column:app_developer" json:"app_developer"`
	PermissionPurpose string `form:"permission_purpose" validate:"required" gorm:"column:permission_purpose" json:"permission_purpose"`
	PrivacyPolicy     string `form:"privacy_policy" validate:"required" gorm:"column:privacy_policy" json:"privacy_policy"`
	CrowedName        string `form:"crowed_name" validate:"required" gorm:"column:crowed_name" json:"crowed_name"`
	CrowedBusiness    string `form:"crowed_business" validate:"required" gorm:"column:crowed_business" json:"crowed_business"`
	ICON              string `form:"icon" validate:"required" gorm:"column:icon" json:"icon"`
	ApkSize           string `form:"apk_size" validate:"required" gorm:"column:apk_size" json:"apk_size"`
	JumpToAppStore    int64  `form:"jump_to_app_store" gorm:"column:jump_to_app_store" json:"jump_to_app_store"`
}

func (p PushDetail) TableName() string {
	return "package_push"
}

func (p *PushDetail) IsUpdateOp() bool {
	return p.ID > 0
}
