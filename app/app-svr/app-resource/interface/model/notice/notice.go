package notice

import xtime "go-common/library/time"

// Notice is notice type.
type Notice struct {
	ID        int        `json:"id,omitempty"`
	Title     string     `json:"title,omitempty"`
	Content   string     `json:"content,omitempty"`
	Start     xtime.Time `json:"start_time,omitempty"`
	End       xtime.Time `json:"end_time,omitempty"`
	URI       string     `json:"uri,omitempty"`
	Type      int        `json:"-"`
	Plat      int8       `json:"-"`
	Build     int        `json:"-"`
	Condition string     `json:"-"`
	Area      string     `json:"-"`
}

type PushDetail struct {
	ID                int64  `json:"id"`
	Title             string `json:"push_title"`
	Text              string `json:"push_text"`
	PackageUrl        string `json:"package_url"`
	PopupTitle        string `json:"popup_title"`
	AppName           string `json:"app_name"`
	AppCurrentVersion string `json:"app_current_version"`
	AppUpdateTime     string `json:"app_update_time"`
	AppDeveloper      string `json:"app_developer"`
	PermissionPurpose string `json:"permission_purpose"`
	PrivacyPolicy     string `json:"privacy_policy"`
	CrowedName        string `json:"-"`
	CrowedBusiness    string `json:"-"`
	ICON              string `json:"icon"`
	ApkSize           string `json:"apk_size"`
	JumpToAppStore    int64  `json:"jump_to_app_store"`
}
