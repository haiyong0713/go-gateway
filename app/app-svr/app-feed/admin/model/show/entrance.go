package show

// EntranceCore .
type EntranceCore struct { // form, json, validate, gorm
	Title                   string `form:"title"        json:"title"        validate:"required"`
	Icon                    string `form:"icon"         json:"icon"         validate:"required"`
	RedirectUri             string `form:"redirect_uri" json:"redirect_uri" validate:"required"  gorm:"column:redirect_uri"`
	Rank                    int    `form:"rank"         json:"rank"`
	ModuleID                string `form:"module_id"    json:"module_id"    validate:"required"  gorm:"column:module_id"`
	Grey                    int    `form:"grey"         json:"grey"`
	RedDot                  int    `form:"red_dot"      json:"red_dot"      gorm:"column:red_dot"`
	RedDotText              string `form:"red_dot_text" json:"red_dot_text" gorm:"column:red_dot_text"`
	WhiteList               string `form:"white_list"   json:"white_list"   gorm:"column:white_list"`
	WhiteListBgroupBusiness string `form:"white_list_bgroup_business"   json:"white_list_bgroup_business"   gorm:"column:white_list_bgroup_business"`
	WhiteListBgroupName     string `form:"white_list_bgroup_name"   json:"white_list_bgroup_name"   gorm:"column:white_list_bgroup_name"`
	BlackList               string `form:"black_list"   json:"black_list"   gorm:"column:black_list"`
	BuildLimit              string `form:"build_limit"  json:"-"            validate:"required"  gorm:"column:build_limit"`
	ID                      int64  `form:"id"           json:"id" `
	State                   int    `json:"state"`
	TopPhoto                string `form:"top_photo"   json:"top_photo"   gorm:"column:top_photo"`
	ShareDesc               string `form:"share_desc"   json:"share_desc"   gorm:"column:share_desc"`
	ShareTitle              string `form:"share_title"   json:"share_title"   gorm:"column:share_title"`
	ShareSubTitle           string `form:"share_sub_title"   json:"share_sub_title"   gorm:"column:share_sub_title"`
	ShareIcon               string `form:"share_icon"   json:"share_icon"   gorm:"column:share_icon"`
}

// EntranceSave .
type EntranceSave struct {
	EntranceCore
	Version int
}

// EntranceList .
type EntranceList struct {
	EntranceCore
	BuildLimitSc []*VersionControl `json:"build_limit,omitempty"`
	VideoCount   int               `json:"video_count"`
}

// EntranceListRes .
type EntranceListRes struct {
	Items []*EntranceList `json:"items"`
	Pager PagerCfg        `json:"pager"`
}

// VersionControl .
type VersionControl struct {
	Plat           int    `json:"plat"`
	ConditionStart string `json:"condition_start"`
	BuildStart     int    `json:"build_start"`
	ConditionEnd   string `json:"condition_end"`
	BuildEnd       int    `json:"build_end"`
}

// PagerCfg .
type PagerCfg struct {
	Num   int `json:"num"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// ToEntranceMap .
func (v *EntranceSave) ToEntranceMap() (res map[string]interface{}) {
	res = make(map[string]interface{})
	res["title"] = v.Title
	res["icon"] = v.Icon
	res["redirect_uri"] = v.RedirectUri
	res["rank"] = v.Rank
	res["module_id"] = v.ModuleID
	res["grey"] = v.Grey
	res["red_dot"] = v.RedDot
	res["red_dot_text"] = v.RedDotText
	res["build_limit"] = v.BuildLimit
	res["white_list"] = v.WhiteList
	res["white_list_bgroup_business"] = v.WhiteListBgroupBusiness
	res["white_list_bgroup_name"] = v.WhiteListBgroupName
	res["black_list"] = v.BlackList
	res["share_desc"] = v.ShareDesc
	res["share_title"] = v.ShareTitle
	res["share_sub_title"] = v.ShareSubTitle
	res["share_icon"] = v.ShareIcon
	return
}

// TableName .
func (*EntranceSave) TableName() string {
	return "popular_top_entrance"
}

func (*EntranceList) TableName() string {
	return "popular_top_entrance"
}

// EntranceView .
type EntranceView struct {
	ID        int64    `json:"id"`
	HeadImage string   `json:"head_image"`
	Tags      []*Tag   `json:"tags"`
	Videos    []*Video `json:"videos"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Video struct {
	RID     int64  `json:"rid"`
	TagName string `json:"tag_name"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	State   int    `json:"state"`
	TagID   int64  `json:"tag_id"`
	BvID    string `json:"bvid"`
	TagIDs  []int64
}

type MiddleTopPhoto struct {
	ID           int64  `json:"id"`
	LocationId   int64  `json:"location_id"`
	LocationName string `json:"location_name"`
	TopPhoto     string `json:"top_photo"`
	BuildLimit   string `json:"build_limit"`
}
