package menu

import (
	"go-common/library/ecode"
	"go-common/library/time"

	locgrpc "git.bilibili.co/bapis/bapis-go/community/service/location"
	locadmingrpc "git.bilibili.co/bapis/bapis-go/platform/admin/location"
)

const (
	// Skin Location 主题配置区域限制
	SkinLocationBusinessSource        = "skin"
	SkinLocationPolicyGroupTypeSkin   = int32(locadmingrpc.SKIN)
	SkinLocationPolicyGroupNamePrefix = "运营主题策略组"
	SkinLocationPolicyGroupRemark     = "运营主题地区限制策略"
	SkinLocationPolicyPlayAuth        = int32(locgrpc.Status_Allow)
	SkinLocationPolicyDownAuth        = int32(locgrpc.StatusDown_AllowDown)
)

// SkinListReply .
type SkinListReply struct {
	List []*SkinList `json:"list"`
	Page *Page       `json:"page"`
}

// SkinSaveReply .
type SkinSaveReply struct {
	ID int64 `json:"id,omitempty"`
}

// SkinReply .
type SkinReply struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

// SkinSaveParam .
type SkinSaveParam struct {
	ID             int64     `form:"id" default:"0" validate:"min=0"`
	SkinID         int64     `form:"skin_id" validate:"min=1"`
	SkinName       string    `form:"skin_name" validate:"min=1"`
	Attribute      int64     `form:"attribute"`
	Stime          time.Time `form:"stime" validate:"required"`
	Etime          time.Time `form:"etime" validate:"required"`
	Limit          string    `form:"limit" validate:"required"`
	AreaIDs        string    `form:"area_ids" validate:"required"`
	UserScopeType  string    `form:"user_scope_type"`
	UserScopeValue string    `form:"user_scope_value"`
	DressUpType    string    `form:"dress_up_type"`  //主动装扮类型 默认：全部人群下发,only_pure:仅纯色用户,mid_scope:人群包
	DressUpValue   string    `form:"dress_up_value"` //人群包id
}

// SkinList .
type SkinList struct {
	*SkinExt
	Image string       `json:"image"`
	Limit []*SkinLimit `json:"limit"`
}

type SkinListParam struct {
	SID int64 `form:"s_id" validate:"min=0"`
	Pn  int   `form:"pn" default:"1" validate:"min=1"`
	Ps  int   `form:"ps" default:"15" validate:"min=1"`
}

// SkinExt .
type SkinExt struct {
	ID                    int64     `gorm:"column:id" json:"id"`
	SkinID                int64     `gorm:"column:skin_id" json:"skin_id"`
	SkinName              string    `gorm:"column:skin_name" json:"skin_name"`
	Attribute             int64     `gorm:"column:attribute" json:"attribute"`
	State                 int       `gorm:"column:state" json:"state"`
	Ctime                 time.Time `gorm:"column:ctime" json:"ctime"`
	Mtime                 time.Time `gorm:"column:mtime" json:"mtime"`
	Stime                 time.Time `gorm:"column:stime" json:"stime"`
	Etime                 time.Time `gorm:"column:etime" json:"etime"`
	Operator              string    `gorm:"column:operator" json:"operator"`
	LocationPolicyGroupID int64     `gorm:"column:location_policy_gid" json:"location_policy_gid"`
	UserScopeType         string    `gorm:"column:user_scope_type" json:"user_scope_type"`
	UserScopeValue        string    `gorm:"column:user_scope_value" json:"user_scope_value"`
	DressUpType           string    `gorm:"column:dress_up_type" json:"dress_up_type"`
	DressUpValue          string    `gorm:"column:dress_up_value" json:"dress_up_value"`
}

// AttrVal get attr val by bit.
func (a *SkinExt) AttrVal(bit uint) int64 {
	return (a.Attribute >> bit) & int64(1)
}

// TableName .
func (a SkinExt) TableName() string {
	return "skin_ext"
}

// SkinLimit .
type SkinLimit struct {
	ID         int64     `gorm:"column:id" json:"id"`
	SID        int64     `gorm:"column:s_id" json:"s_id"`
	Plat       int8      `gorm:"column:plat" json:"plat"`
	Build      int64     `gorm:"column:build" json:"build"`
	Conditions string    `gorm:"column:conditions" json:"conditions"`
	State      int       `gorm:"column:state" json:"state"`
	Ctime      time.Time `gorm:"column:ctime" json:"ctime"`
	Mtime      time.Time `gorm:"column:mtime" json:"mtime"`
}

// TableName .
func (a SkinLimit) TableName() string {
	return "skin_limit"
}

// SkinBuildLimit .
type SkinBuildLimit struct {
	Plat       int8   `json:"plat"`
	Build      int64  `json:"build"`
	Conditions string `json:"conditions"`
}

func (lt SkinBuildLimit) ValidateParam() (err error) {
	if lt.Build < 0 {
		err = ecode.RequestErr
		return
	}
	if lt.Plat != PlatAndroid && lt.Plat != PlatIPhone {
		err = ecode.RequestErr
		return
	}
	if lt.Conditions != ConditionsLt && lt.Conditions != ConditionsGt && lt.Conditions != ConditionsEq && lt.Conditions != ConditionsNe {
		err = ecode.RequestErr
	}
	return
}

// SearchReply .
type SkinSearchReply struct {
	Total int64
	List  []*SkinExt
}
